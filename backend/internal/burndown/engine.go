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

type Broadcaster interface {
	BroadcastBurndown(milestoneID model.ID, point model.BurndownPoint)
}

type Engine struct {
	DB    *store.DB
	Clock func() time.Time
	Hub   Broadcaster
}

func NewEngine(db *store.DB, hub Broadcaster) *Engine {
	return &Engine{DB: db, Clock: timeutil.Now, Hub: hub}
}

func (e *Engine) Handle(ctx context.Context, ev model.DomainEvent) error {
	claimed, err := e.DB.TryClaimEvent(ctx, ev.ID, string(ev.Type), e.Clock())
	if err != nil {
		return err
	}
	if !claimed {
		logger.L().Info("event already processed", "event_id", ev.ID)
		return nil
	}
	if ev.MilestoneID == nil {
		return nil
	}
	ms, err := e.DB.MilestoneByID(ctx, *ev.MilestoneID)
	if err != nil {
		return err
	}
	tasks, err := e.DB.ListTasksByMilestone(ctx, ms.ID)
	if err != nil {
		return err
	}
	remaining := RemainingOfTasks(tasks)
	now := e.Clock()
	ideal := IdealAt(ms.BaselinePoints, ms.StartDate, ms.DueDate, now)

	et := model.EventSnapshot
	switch ev.Type {
	case model.DomPomodoroCompleted:
		et = model.EventPomodoroCompleted
	case model.DomTaskDone:
		et = model.EventTaskDone
	case model.DomScopeChanged:
		et = model.EventScopeChange
		if delta, ok := ev.Payload["delta"].(int); ok && delta != 0 {
			_ = e.DB.InsertScopeChange(ctx, &model.ScopeChangeLog{
				ID: uuid.New(), MilestoneID: ms.ID, DeltaPoints: delta,
				Reason: payloadString(ev.Payload, "reason"), OccurredAt: now,
			})
		}
	}

	pt := model.BurndownPoint{
		ID:              uuid.New(),
		MilestoneID:     ms.ID,
		RecordedAt:      now,
		RemainingPoints: remaining,
		IdealPoints:     ideal,
		EventType:       et,
		EventID:         ev.ID,
	}
	if err := e.DB.InsertBurndownPoint(ctx, &pt); err != nil {
		return err
	}
	if e.Hub != nil {
		e.Hub.BroadcastBurndown(ms.ID, pt)
	}
	logger.L().Info("burndown updated", "milestone", ms.ID, "remaining", remaining, "event", ev.Type)
	return nil
}

func (e *Engine) Chart(ctx context.Context, ms *model.Milestone, gran string) (map[string]any, error) {
	from := timeutil.StartOfDay(ms.StartDate)
	to := timeutil.EndOfDay(ms.DueDate)
	now := e.Clock()
	if now.After(to) {
		to = timeutil.EndOfDay(now)
	}
	points, err := e.DB.ListBurndownPoints(ctx, ms.ID, from, to)
	if err != nil {
		return nil, err
	}
	tasks, err := e.DB.ListTasksByMilestone(ctx, ms.ID)
	if err != nil {
		return nil, err
	}
	remaining := RemainingOfTasks(tasks)
	ideal := IdealSeries(ms.BaselinePoints, ms.StartDate, ms.DueDate)

	idealX := make([]string, 0, len(ideal))
	idealY := make([]float64, 0, len(ideal))
	for _, p := range ideal {
		if gran == "week" && p.Day.Weekday() != time.Monday && !p.Day.Equal(timeutil.StartOfDay(ms.StartDate)) && !p.Day.Equal(timeutil.StartOfDay(ms.DueDate)) {
			continue
		}
		idealX = append(idealX, timeutil.FormatDate(p.Day))
		idealY = append(idealY, p.Ideal)
	}
	realX := make([]string, 0, len(points))
	realY := make([]int, 0, len(points))
	marks := make([]map[string]any, 0)
	for _, p := range points {
		realX = append(realX, timeutil.FormatDateTime(p.RecordedAt))
		realY = append(realY, p.RemainingPoints)
		if p.EventType == model.EventScopeChange {
			marks = append(marks, map[string]any{
				"at": timeutil.FormatDateTime(p.RecordedAt), "remaining": p.RemainingPoints,
			})
		}
	}
	return map[string]any{
		"milestone_id":     ms.ID,
		"baseline_points":  ms.BaselinePoints,
		"remaining_points": remaining,
		"ideal":            map[string]any{"x": idealX, "y": idealY},
		"actual":           map[string]any{"x": realX, "y": realY},
		"scope_marks":      marks,
		"granularity":      gran,
	}, nil
}

func (e *Engine) Metrics(ctx context.Context, userID, milestoneID model.ID) (*model.EfficiencyMetrics, error) {
	now := e.Clock()
	today := timeutil.StartOfDay(now)
	week := today.AddDate(0, 0, -6)
	m := &model.EfficiencyMetrics{}
	var err error
	m.TodayCompleted, err = e.DB.CountSessionsSince(ctx, userID, string(model.StateCompleted), today)
	if err != nil {
		return nil, err
	}
	m.TodayAborted, err = e.DB.CountSessionsSince(ctx, userID, string(model.StateAborted), today)
	if err != nil {
		return nil, err
	}
	m.WeekCompleted, err = e.DB.CountSessionsSince(ctx, userID, string(model.StateCompleted), week)
	if err != nil {
		return nil, err
	}
	m.WeekAborted, err = e.DB.CountSessionsSince(ctx, userID, string(model.StateAborted), week)
	if err != nil {
		return nil, err
	}
	tasks, err := e.DB.ListTasksByMilestone(ctx, milestoneID)
	if err != nil {
		return nil, err
	}
	Assemble(m, RemainingOfTasks(tasks), now)
	return m, nil
}

func payloadString(p map[string]any, k string) string {
	if p == nil {
		return ""
	}
	if s, ok := p[k].(string); ok {
		return s
	}
	return ""
}
