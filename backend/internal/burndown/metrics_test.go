package burndown

import (
	"testing"

	"gopomodoro/internal/timeutil"
)

func TestAbortRateAndPredict(t *testing.T) {
	if AbortRate(8, 2) != 0.2 {
		t.Fatalf("rate %v", AbortRate(8, 2))
	}
	if AbortRate(0, 0) != 0 {
		t.Fatal("empty rate")
	}
	now := timeutil.Date(2026, 8, 23, 0, 0, 0)
	pred := PredictDoneOn(10, 2, now)
	if pred != "2026-08-28" {
		t.Fatalf("predict %s", pred)
	}
	if PredictDoneOn(5, 0, now) != "" {
		t.Fatal("zero velocity should be empty")
	}
	if Risk(0, now, 1, now) != "done" {
		t.Fatal("done risk")
	}
	if Risk(100, timeutil.Date(2026, 8, 20, 0, 0, 0), 1, now) != "overdue" {
		t.Fatal("overdue")
	}
}
