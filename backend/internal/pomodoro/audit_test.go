package pomodoro

import (
	"testing"
	"time"

	"gopomodoro/internal/model"
)

// TestRecentDoesNotMutateAuditTrail reproduces the bug where reading the
// recent transition records shared the internal ring buffer's backing array
// with the caller. The dashboard annotated the returned records in place,
// which silently corrupted CountByTo: running counts dropped and completed
// counts could appear out of thin air, even though no session had actually
// transitioned. The read path must return an independent copy.
func TestRecentDoesNotMutateAuditTrail(t *testing.T) {
	a := NewAuditLog(64)
	now := time.Date(2026, 8, 24, 9, 0, 0, 0, time.UTC)

	a.Append(TransitionRecord{From: model.StateIdle, To: model.StateRunning, Cmd: CmdStart, At: now})
	a.Append(TransitionRecord{From: model.StateRunning, To: model.StatePaused, Cmd: CmdPause, At: now})
	a.Append(TransitionRecord{From: model.StateRunning, To: model.StateCompleted, Cmd: CmdTick, At: now})

	beforeRunning := a.CountByTo(model.StateRunning)
	beforeCompleted := a.CountByTo(model.StateCompleted)
	if beforeRunning != 1 || beforeCompleted != 1 {
		t.Fatalf("precondition: running=%d completed=%d", beforeRunning, beforeCompleted)
	}

	// Simulate the dashboard reading recent records and making temporary
	// display annotations directly on the returned slice.
	recs := a.Recent(3)
	for i := range recs {
		recs[i].To = model.StateCompleted // annotate for display
	}

	afterRunning := a.CountByTo(model.StateRunning)
	afterCompleted := a.CountByTo(model.StateCompleted)
	if afterRunning != beforeRunning {
		t.Fatalf("read mutated audit: running %d -> %d", beforeRunning, afterRunning)
	}
	if afterCompleted != beforeCompleted {
		t.Fatalf("read mutated audit: completed %d -> %d", beforeCompleted, afterCompleted)
	}
}

// TestRecentReturnsCopyEvenWhenEmpty ensures the defensive copy does not panic
// or misbehave on an empty audit log.
func TestRecentReturnsCopyEvenWhenEmpty(t *testing.T) {
	a := NewAuditLog(64)
	if got := a.Recent(0); len(got) != 0 {
		t.Fatalf("expected empty, got %d", len(got))
	}
	if got := a.Recent(10); len(got) != 0 {
		t.Fatalf("expected empty, got %d", len(got))
	}
}

// TestRecentClampsOversizeRequest verifies that requesting more records than
// the ring holds returns all available records as an independent copy.
func TestRecentClampsOversizeRequest(t *testing.T) {
	a := NewAuditLog(64)
	now := time.Date(2026, 8, 24, 9, 0, 0, 0, time.UTC)
	a.Append(TransitionRecord{From: model.StateIdle, To: model.StateRunning, Cmd: CmdStart, At: now})

	got := a.Recent(10)
	if len(got) != 1 {
		t.Fatalf("expected 1, got %d", len(got))
	}
	got[0].To = model.StateAborted
	if a.CountByTo(model.StateRunning) != 1 {
		t.Fatal("oversize read mutated audit trail")
	}
	if a.CountByTo(model.StateAborted) != 0 {
		t.Fatal("oversize read mutated audit trail")
	}
}
