package report

import (
	"testing"

	"gopomodoro/internal/model"
	"gopomodoro/internal/timeutil"
)

func TestTripleOnTrack(t *testing.T) {
	now := timeutil.Date(2026, 8, 23, 0, 0, 0)
	due := timeutil.Date(2026, 9, 15, 0, 0, 0)
	f := Triple(10, 2, due, now)
	if !f.OnTrack {
		t.Fatalf("%+v", f)
	}
	if f.LikelyDate == "" || f.OptimisticDate == "" {
		t.Fatal("dates")
	}
}

func TestTripleZeroVelocity(t *testing.T) {
	now := timeutil.Date(2026, 8, 23, 0, 0, 0)
	due := timeutil.Date(2026, 8, 30, 0, 0, 0)
	f := Triple(8, 0, due, now)
	if f.OnTrack {
		t.Fatal("zero velocity cannot be on track")
	}
}

func TestSummarizeAndPressure(t *testing.T) {
	tasks := []model.Task{
		{EstimatedPomodoros: 5, ConsumedPomodoros: 2, KanbanColumn: model.ColInProgress},
		{EstimatedPomodoros: 3, ConsumedPomodoros: 3, KanbanColumn: model.ColDone},
	}
	s := SummarizeTasks(tasks)
	if s["remaining"] != 3 {
		t.Fatalf("%v", s)
	}
	start := timeutil.Date(2026, 8, 1, 0, 0, 0)
	due := timeutil.Date(2026, 8, 31, 0, 0, 0)
	now := timeutil.Date(2026, 8, 1, 0, 0, 0)
	if ScopePressure(10, 10, start, due, now) != 1 {
		t.Fatal("start pressure should be 1")
	}
}
