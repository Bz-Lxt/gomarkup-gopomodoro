package pomodoro

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/google/uuid"
	"gopomodoro/internal/eventbus"
	"gopomodoro/internal/httpx"
	"gopomodoro/internal/logger"
	"gopomodoro/internal/model"
	"gopomodoro/internal/store"
	"gopomodoro/internal/timeutil"
)

type SessionView struct {
	Session     *model.PomodoroSession `json:"session"`
	RemainingMS int64                  `json:"remaining_ms"`
	GraceLeftS  int                    `json:"grace_left_s"`
}

type Live struct {
	mu             sync.Mutex
	session        *model.PomodoroSession
	timer          Timer
	disconnectedAt *time.Time
	conns          int
}

type HubSink interface {
	BroadcastSession(userID model.ID, view SessionView)
	BroadcastGrace(userID model.ID, leftS int)
}

type Registry struct {
	mu          sync.RWMutex
	lives       map[model.ID]*Live // keyed by session id
	byUser      map[model.ID]model.ID
	DB          *store.DB
	Clock       Clock
	Focus       time.Duration
	Grace       time.Duration
	Bus         *eventbus.Bus
	Hub         HubSink
	sweepEvery  time.Duration
	stopSweep   chan struct{}
	sweepOnce   sync.Once
}

func NewRegistry(db *store.DB, clock Clock, focus, grace time.Duration, bus *eventbus.Bus) *Registry {
	if clock == nil {
		clock = RealClock{}
	}
	return &Registry{
		lives:      make(map[model.ID]*Live),
		byUser:     make(map[model.ID]model.ID),
		DB:         db,
		Clock:      clock,
		Focus:      focus,
		Grace:      grace,
		Bus:        bus,
		sweepEvery: time.Second,
		stopSweep:  make(chan struct{}),
	}
}

func (r *Registry) StartSweep() {
	go r.sweepLoop()
}

func (r *Registry) StopSweep() {
	r.sweepOnce.Do(func() { close(r.stopSweep) })
}

func (r *Registry) Rebuild(ctx context.Context) error {
	list, err := r.DB.ListLiveSessions(ctx)
	if err != nil {
		return err
	}
	now := r.Clock.Now()
	for i := range list {
		s := list[i]
		if s.State == model.StateRunning && s.ExpectedEndAt != nil && !s.ExpectedEndAt.After(now) {
			if err := r.completeLocked(ctx, &s, now); err != nil {
				logger.L().Error("rebuild complete failed", "session", s.ID, "err", err)
			}
			continue
		}
		live := &Live{session: clone(&s)}
		r.mu.Lock()
		r.lives[s.ID] = live
		r.byUser[s.UserID] = s.ID
		r.mu.Unlock()
		if s.State == model.StateRunning {
			r.armTimer(live)
		}
	}
	logger.L().Info("registry rebuilt", "live", len(list))
	return nil
}

func (r *Registry) Start(ctx context.Context, userID, taskID model.ID) (*SessionView, error) {
	if existing := r.activeOf(userID); existing != nil {
		return nil, httpx.ErrSessionBusy
	}
	dbActive, err := r.DB.ActiveSessionByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if dbActive != nil {
		return nil, httpx.ErrSessionBusy
	}
	now := r.Clock.Now()
	idle := &model.PomodoroSession{
		ID:              uuid.New(),
		UserID:          userID,
		TaskID:          taskID,
		State:           model.StateIdle,
		FocusDurationMS: r.Focus.Milliseconds(),
		ResumeToken:     newToken(),
		Version:         1,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := r.DB.InsertSession(ctx, idle); err != nil {
		return nil, err
	}
	next, err := Apply(idle, CmdStart, now, "")
	if err != nil {
		return nil, err
	}
	if err := r.DB.UpdateSessionOptimistic(ctx, next, idle.Version); err != nil {
		return nil, err
	}
	live := &Live{session: clone(next)}
	r.mu.Lock()
	r.lives[next.ID] = live
	r.byUser[userID] = next.ID
	r.mu.Unlock()
	r.armTimer(live)
	view := r.viewOf(live)
	r.emitSession(next.UserID, *view)
	return view, nil
}

func (r *Registry) Command(ctx context.Context, userID, sessionID model.ID, cmd Command, reason string) (*SessionView, error) {
	live := r.live(sessionID)
	if live == nil {
		s, err := r.DB.SessionByID(ctx, sessionID)
		if err != nil {
			return nil, err
		}
		if s.UserID != userID {
			return nil, httpx.ErrForbidden
		}
		return r.applyDB(ctx, s, cmd, reason)
	}
	live.mu.Lock()
	if live.session.UserID != userID {
		live.mu.Unlock()
		return nil, httpx.ErrForbidden
	}
	prev := live.session.Version
	now := r.Clock.Now()
	next, err := Apply(live.session, cmd, now, reason)
	if err != nil {
		live.mu.Unlock()
		return nil, err
	}
	if err := r.DB.UpdateSessionOptimistic(ctx, next, prev); err != nil {
		live.mu.Unlock()
		return nil, err
	}
	if live.timer != nil {
		live.timer.Stop()
		live.timer = nil
	}
	live.session = clone(next)
	live.disconnectedAt = nil
	if next.State == model.StateRunning {
		r.armTimerLocked(live)
	}
	terminal := next.State == model.StateAborted || next.State == model.StateCompleted
	live.mu.Unlock()
	if next.State == model.StateAborted {
		r.drop(next.ID, next.UserID)
		r.publishAbort(next)
	}
	if next.State == model.StateCompleted {
		if err := r.DB.IncrementConsumed(ctx, next.TaskID); err != nil {
			return nil, err
		}
		r.drop(next.ID, next.UserID)
		r.publishComplete(ctx, next)
	}
	_ = terminal
	view := SessionView{Session: clone(next), RemainingMS: next.RemainingMS(now)}
	r.emitSession(next.UserID, view)
	return &view, nil
}

func (r *Registry) applyDB(ctx context.Context, s *model.PomodoroSession, cmd Command, reason string) (*SessionView, error) {
	now := r.Clock.Now()
	next, err := Apply(s, cmd, now, reason)
	if err != nil {
		return nil, err
	}
	if err := r.DB.UpdateSessionOptimistic(ctx, next, s.Version); err != nil {
		return nil, err
	}
	if next.State.Active() {
		live := &Live{session: clone(next)}
		r.mu.Lock()
		r.lives[next.ID] = live
		r.byUser[next.UserID] = next.ID
		r.mu.Unlock()
		if next.State == model.StateRunning {
			r.armTimer(live)
		}
	}
	if next.State == model.StateAborted {
		r.publishAbort(next)
	}
	if next.State == model.StateCompleted {
		if err := r.DB.IncrementConsumed(ctx, next.TaskID); err != nil {
			return nil, err
		}
		r.drop(next.ID, next.UserID)
		r.publishComplete(ctx, next)
	}
	view := SessionView{Session: next, RemainingMS: next.RemainingMS(now)}
	r.emitSession(next.UserID, view)
	return &view, nil
}

func (r *Registry) Active(userID model.ID) *SessionView {
	live := r.activeOf(userID)
	if live == nil {
		return nil
	}
	return r.viewOf(live)
}

func (r *Registry) Get(sessionID model.ID) *SessionView {
	live := r.live(sessionID)
	if live == nil {
		return nil
	}
	return r.viewOf(live)
}

func (r *Registry) MarkConnected(sessionID model.ID) {
	live := r.live(sessionID)
	if live == nil {
		return
	}
	live.mu.Lock()
	live.conns++
	live.disconnectedAt = nil
	live.mu.Unlock()
}

func (r *Registry) MarkDisconnected(sessionID model.ID) {
	live := r.live(sessionID)
	if live == nil {
		return
	}
	live.mu.Lock()
	if live.conns > 0 {
		live.conns--
	}
	if live.conns == 0 && live.session.State.Active() {
		t := r.Clock.Now()
		live.disconnectedAt = &t
	}
	live.mu.Unlock()
}

func (r *Registry) ResumeByToken(ctx context.Context, userID model.ID, token string) (*SessionView, error) {
	if token == "" {
		if v := r.Active(userID); v != nil {
			return v, nil
		}
		s, err := r.DB.ActiveSessionByUser(ctx, userID)
		if err != nil || s == nil {
			return nil, err
		}
		return &SessionView{Session: s, RemainingMS: s.RemainingMS(r.Clock.Now())}, nil
	}
	s, err := r.DB.SessionByResumeToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if s.UserID != userID {
		return nil, httpx.ErrForbidden
	}
	if v := r.Get(s.ID); v != nil {
		r.MarkConnected(s.ID)
		return v, nil
	}
	return &SessionView{Session: s, RemainingMS: s.RemainingMS(r.Clock.Now())}, nil
}

func (r *Registry) forceComplete(sessionID model.ID) {
	live := r.live(sessionID)
	if live == nil {
		return
	}
	ctx := context.Background()
	live.mu.Lock()
	if live.session.State != model.StateRunning {
		live.mu.Unlock()
		return
	}
	snap := clone(live.session)
	live.mu.Unlock()
	now := r.Clock.Now()
	if err := r.completeLocked(ctx, snap, now); err != nil {
		logger.L().Error("tick complete failed", "session", sessionID, "err", err)
		return
	}
}

func (r *Registry) completeLocked(ctx context.Context, s *model.PomodoroSession, now time.Time) error {
	next, err := Apply(s, CmdTick, now, "")
	if err != nil {
		return err
	}
	if err := r.DB.UpdateSessionOptimistic(ctx, next, s.Version); err != nil {
		return err
	}
	if err := r.DB.IncrementConsumed(ctx, next.TaskID); err != nil {
		return err
	}
	r.drop(next.ID, next.UserID)
	r.publishComplete(ctx, next)
	view := SessionView{Session: next, RemainingMS: 0}
	r.emitSession(next.UserID, view)
	return nil
}

func (r *Registry) publishComplete(ctx context.Context, s *model.PomodoroSession) {
	task, err := r.DB.TaskByID(ctx, s.TaskID)
	if err != nil {
		logger.L().Error("load task after complete", "err", err)
		return
	}
	ev := model.DomainEvent{
		ID: "pomo:" + s.ID.String() + ":completed",
		Type: model.DomPomodoroCompleted, UserID: s.UserID, ProjectID: task.ProjectID,
		MilestoneID: task.MilestoneID, TaskID: &s.TaskID, SessionID: &s.ID,
		OccurredAt: timeutil.Now(),
	}
	if r.Bus != nil {
		r.Bus.Publish(ev)
	}
}

func (r *Registry) publishAbort(s *model.PomodoroSession) {
	if r.Bus == nil {
		return
	}
	r.Bus.Publish(model.DomainEvent{
		ID: "pomo:" + s.ID.String() + ":aborted",
		Type: model.DomPomodoroAborted, UserID: s.UserID, SessionID: &s.ID,
		OccurredAt: timeutil.Now(),
	})
}

func (r *Registry) PublishScope(userID, projectID model.ID, milestoneID *model.ID, taskID *model.ID, delta int, reason string) {
	if r.Bus == nil || milestoneID == nil {
		return
	}
	r.Bus.Publish(model.DomainEvent{
		ID: "scope:" + uuid.NewString(),
		Type: model.DomScopeChanged, UserID: userID, ProjectID: projectID,
		MilestoneID: milestoneID, TaskID: taskID,
		Payload: map[string]any{"delta": delta, "reason": reason},
		OccurredAt: timeutil.Now(),
	})
}

func (r *Registry) PublishTaskDone(userID, projectID model.ID, milestoneID *model.ID, taskID model.ID) {
	if r.Bus == nil || milestoneID == nil {
		return
	}
	r.Bus.Publish(model.DomainEvent{
		ID: "task:" + taskID.String() + ":done",
		Type: model.DomTaskDone, UserID: userID, ProjectID: projectID,
		MilestoneID: milestoneID, TaskID: &taskID,
		OccurredAt: timeutil.Now(),
	})
}

func (r *Registry) armTimer(live *Live) {
	live.mu.Lock()
	defer live.mu.Unlock()
	r.armTimerLocked(live)
}

func (r *Registry) armTimerLocked(live *Live) {
	if live.timer != nil {
		live.timer.Stop()
	}
	left := live.session.RemainingMS(r.Clock.Now())
	if left < 0 {
		left = 0
	}
	id := live.session.ID
	live.timer = r.Clock.AfterFunc(time.Duration(left)*time.Millisecond, func() {
		r.forceComplete(id)
	})
}

func (r *Registry) sweepLoop() {
	t := time.NewTicker(r.sweepEvery)
	defer t.Stop()
	for {
		select {
		case <-r.stopSweep:
			return
		case <-t.C:
			r.sweepGrace()
		}
	}
}

func (r *Registry) sweepGrace() {
	now := r.Clock.Now()
	type hit struct{ id, user model.ID }
	var expired []hit
	r.mu.RLock()
	snapshot := make([]*Live, 0, len(r.lives))
	for _, live := range r.lives {
		snapshot = append(snapshot, live)
	}
	r.mu.RUnlock()
	for _, live := range snapshot {
		live.mu.Lock()
		id := live.session.ID
		if live.disconnectedAt != nil && now.Sub(*live.disconnectedAt) >= r.Grace && live.session.State.Active() {
			expired = append(expired, hit{id, live.session.UserID})
		} else if live.disconnectedAt != nil && r.Hub != nil {
			left := int((r.Grace - now.Sub(*live.disconnectedAt)).Seconds())
			if left < 0 {
				left = 0
			}
			uid := live.session.UserID
			live.mu.Unlock()
			r.Hub.BroadcastGrace(uid, left)
			continue
		}
		live.mu.Unlock()
	}
	for _, h := range expired {
		ctx := context.Background()
		if _, err := r.Command(ctx, h.user, h.id, CmdAbort, string(model.AbortNetworkTimeout)); err != nil {
			logger.L().Error("grace abort failed", "session", h.id, "err", err)
		} else {
			logger.L().Info("session aborted by grace", "session", h.id)
		}
	}
}

func (r *Registry) live(id model.ID) *Live {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lives[id]
}

func (r *Registry) activeOf(userID model.ID) *Live {
	r.mu.RLock()
	id, ok := r.byUser[userID]
	r.mu.RUnlock()
	if !ok {
		return nil
	}
	return r.live(id)
}

func (r *Registry) drop(sessionID, userID model.ID) {
	r.mu.Lock()
	delete(r.lives, sessionID)
	if r.byUser[userID] == sessionID {
		delete(r.byUser, userID)
	}
	r.mu.Unlock()
}

func (r *Registry) viewOf(live *Live) *SessionView {
	live.mu.Lock()
	defer live.mu.Unlock()
	now := r.Clock.Now()
	grace := 0
	if live.disconnectedAt != nil {
		left := int((r.Grace - now.Sub(*live.disconnectedAt)).Seconds())
		if left < 0 {
			left = 0
		}
		grace = left
	}
	return &SessionView{Session: clone(live.session), RemainingMS: live.session.RemainingMS(now), GraceLeftS: grace}
}

func (r *Registry) emitSession(userID model.ID, view SessionView) {
	if r.Hub != nil {
		r.Hub.BroadcastSession(userID, view)
	}
}

func (r *Registry) LiveCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.lives)
}

func clone(s *model.PomodoroSession) *model.PomodoroSession {
	if s == nil {
		return nil
	}
	cp := *s
	return &cp
}

func newToken() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
