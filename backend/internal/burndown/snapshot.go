package burndown

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gopomodoro/internal/logger"
	"gopomodoro/internal/model"
	"gopomodoro/internal/store"
	"gopomodoro/internal/timeutil"
)

// Snapshotter writes one daily remaining-point snapshot per active milestone
// so the real line has a vertex even on days without completions.
type Snapshotter struct {
	DB       *store.DB
	Engine   *Engine
	Interval time.Duration
	stop     chan struct{}
}

func NewSnapshotter(db *store.DB, engine *Engine) *Snapshotter {
	return &Snapshotter{DB: db, Engine: engine, Interval: time.Hour, stop: make(chan struct{})}
}

func (s *Snapshotter) Start() {
	go s.loop()
}

func (s *Snapshotter) Stop() {
	select {
	case <-s.stop:
	default:
		close(s.stop)
	}
}

func (s *Snapshotter) loop() {
	t := time.NewTicker(s.Interval)
	defer t.Stop()
	s.tick(context.Background())
	for {
		select {
		case <-s.stop:
			return
		case <-t.C:
			s.tick(context.Background())
		}
	}
}

func (s *Snapshotter) tick(ctx context.Context) {
	// Walk every project owner via a lightweight distinct query on milestones.
	rows, err := s.DB.SQL.QueryContext(ctx, `SELECT id, project_id, title, start_date, due_date, baseline_points, status, created_at, updated_at FROM milestones WHERE status <> 'done'`)
	if err != nil {
		logger.L().Error("snapshot list milestones", "err", err)
		return
	}
	defer rows.Close()
	now := timeutil.Now()
	dayKey := timeutil.FormatDate(now)
	for rows.Next() {
		m, err := scanMS(rows)
		if err != nil {
			continue
		}
		eventID := "snap:" + m.ID.String() + ":" + dayKey
		claimed, err := s.DB.TryClaimEvent(ctx, eventID, string(model.EventSnapshot), now)
		if err != nil || !claimed {
			continue
		}
		tasks, err := s.DB.ListTasksByMilestone(ctx, m.ID)
		if err != nil {
			continue
		}
		pt := model.BurndownPoint{
			ID: uuid.New(), MilestoneID: m.ID, RecordedAt: now,
			RemainingPoints: RemainingOfTasks(tasks),
			IdealPoints:     IdealAt(m.BaselinePoints, m.StartDate, m.DueDate, now),
			EventType:       model.EventSnapshot, EventID: eventID,
		}
		if err := s.DB.InsertBurndownPoint(ctx, &pt); err != nil {
			logger.L().Error("snapshot insert", "err", err)
		}
	}
}

func scanMS(rows interface{ Scan(...any) error }) (*model.Milestone, error) {
	var m model.Milestone
	if err := rows.Scan(&m.ID, &m.ProjectID, &m.Title, &m.StartDate, &m.DueDate, &m.BaselinePoints, &m.Status, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return nil, err
	}
	m.StartDate = timeutil.StartOfDay(m.StartDate)
	m.DueDate = timeutil.StartOfDay(m.DueDate)
	return &m, nil
}
