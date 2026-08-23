package burndown

import (
	"testing"

	"gopomodoro/internal/timeutil"
)

func TestIdealEndpoints(t *testing.T) {
	start := timeutil.Date(2026, 8, 1, 0, 0, 0)
	due := timeutil.Date(2026, 8, 31, 0, 0, 0)
	if got := IdealAt(36, start, due, start); got != 36 {
		t.Fatalf("start ideal %v", got)
	}
	if got := IdealAt(36, start, due, due); got != 0 {
		t.Fatalf("due ideal %v", got)
	}
	mid := timeutil.Date(2026, 8, 16, 0, 0, 0)
	got := IdealAt(36, start, due, mid)
	if got <= 0 || got >= 36 {
		t.Fatalf("mid ideal should be between, got %v", got)
	}
}

func TestIdealSeriesLength(t *testing.T) {
	start := timeutil.Date(2026, 8, 1, 0, 0, 0)
	due := timeutil.Date(2026, 8, 3, 0, 0, 0)
	s := IdealSeries(10, start, due)
	if len(s) != 3 {
		t.Fatalf("len %d", len(s))
	}
}
