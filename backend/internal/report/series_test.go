package report

import (
	"testing"

	"gopomodoro/internal/model"
	"gopomodoro/internal/timeutil"
)

func TestMergeCarryForward(t *testing.T) {
	start := timeutil.Date(2026, 8, 1, 0, 0, 0)
	due := timeutil.Date(2026, 8, 4, 0, 0, 0)
	actual := []model.BurndownPoint{{
		RecordedAt: timeutil.Date(2026, 8, 2, 12, 0, 0), RemainingPoints: 7,
		EventType: model.EventPomodoroCompleted,
	}}
	s := MergeDaySeries(10, start, due, actual)
	if len(s) != 4 {
		t.Fatalf("len %d", len(s))
	}
	if s[0].Actual != nil {
		t.Fatal("day0 should have no actual yet")
	}
	if s[1].Actual == nil || *s[1].Actual != 7 {
		t.Fatal("day1")
	}
	if s[3].Actual == nil || *s[3].Actual != 7 {
		t.Fatal("carry")
	}
}
