package pomodoro_test

import (
	"testing"
	"time"

	"gopomodoro/internal/model"
	"gopomodoro/internal/pomodoro"
)

func TestRecentAuditRecordsAreIsolated(t *testing.T) {
	log := pomodoro.NewAuditLog(32)
	log.Append(pomodoro.TransitionRecord{
		From:      model.StateIdle,
		To:        model.StateRunning,
		Cmd:       pomodoro.CmdStart,
		At:        time.Date(2026, 8, 24, 9, 0, 0, 0, time.UTC),
	})

	recent := log.Recent(1)
	recent[0].To = model.StateCompleted

	if got := log.CountByTo(model.StateRunning); got != 1 {
		t.Fatalf("editing returned records changed running transition count to %d, want 1", got)
	}
	if got := log.CountByTo(model.StateCompleted); got != 0 {
		t.Fatalf("editing returned records changed completed transition count to %d, want 0", got)
	}
}
