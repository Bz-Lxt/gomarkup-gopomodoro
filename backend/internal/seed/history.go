package seed

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gopomodoro/internal/model"
	"gopomodoro/internal/store"
	"gopomodoro/internal/timeutil"
)

type histSpec struct {
	dayOffset int
	hour      int
	completed bool
	taskIdx   int
}

var demoHistory = []histSpec{
	{21, 9, true, 1}, {21, 10, true, 1}, {21, 14, false, 0},
	{20, 9, true, 6}, {20, 11, true, 6},
	{18, 9, true, 0}, {18, 10, true, 0}, {18, 15, true, 5},
	{17, 9, false, 2}, {17, 10, true, 2},
	{16, 9, true, 0}, {16, 11, true, 5}, {16, 16, true, 5},
	{15, 9, true, 1}, {14, 10, true, 6},
	{13, 9, true, 7}, {13, 14, false, 7},
	{12, 9, true, 0}, {12, 10, true, 0}, {11, 9, true, 5},
	{10, 9, true, 8}, {9, 10, true, 0}, {8, 9, true, 2},
	{7, 9, true, 5}, {6, 11, false, 3}, {5, 9, true, 0},
	{4, 9, true, 5}, {3, 10, true, 7}, {2, 9, true, 0},
	{1, 9, true, 0}, {1, 15, false, 2},
}

func insertHistory(ctx context.Context, db *store.DB, userID model.ID, tasks []model.Task) error {
	now := timeutil.Now()
	for i, h := range demoHistory {
		if h.taskIdx >= len(tasks) {
			h.taskIdx = 0
		}
		task := tasks[h.taskIdx]
		start := timeutil.StartOfDay(now).AddDate(0, 0, -h.dayOffset).Add(time.Duration(h.hour) * time.Hour)
		end := start.Add(25 * time.Minute)
		state := model.StateCompleted
		reason := ""
		if !h.completed {
			state = model.StateAborted
			reason = string(model.AbortUser)
			end = start.Add(8 * time.Minute)
		}
		s := &model.PomodoroSession{
			ID: uuid.New(), UserID: userID, TaskID: task.ID, State: state,
			FocusDurationMS: 25 * 60 * 1000, StartedAt: &start, ExpectedEndAt: &end,
			EndedAt: &end, AbortReason: reason, ResumeToken: "seed-" + uuid.NewString(),
			Version: 2, CreatedAt: start, UpdatedAt: end,
		}
		if err := db.InsertSession(ctx, s); err != nil {
			return err
		}
		_ = i
	}
	return nil
}
