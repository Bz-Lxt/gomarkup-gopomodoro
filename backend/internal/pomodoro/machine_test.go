package pomodoro

import (
	"testing"
	"time"

	"gopomodoro/internal/httpx"
	"gopomodoro/internal/model"
)

func sess(state model.PomodoroState) *model.PomodoroSession {
	now := time.Date(2026, 8, 23, 16, 0, 0, 0, time.UTC)
	end := now.Add(25 * time.Minute)
	s := &model.PomodoroSession{
		State: state, FocusDurationMS: 25 * 60 * 1000, Version: 1,
	}
	if state != model.StateIdle {
		s.StartedAt = &now
		s.ExpectedEndAt = &end
	}
	return s
}

func TestLegalTransitions(t *testing.T) {
	now := time.Date(2026, 8, 23, 16, 0, 0, 0, time.UTC)
	cases := []struct {
		from model.PomodoroState
		cmd  Command
		to   model.PomodoroState
	}{
		{model.StateIdle, CmdStart, model.StateRunning},
		{model.StateRunning, CmdPause, model.StatePaused},
		{model.StatePaused, CmdResume, model.StateRunning},
		{model.StateRunning, CmdAbort, model.StateAborted},
		{model.StatePaused, CmdAbort, model.StateAborted},
		{model.StateRunning, CmdTick, model.StateCompleted},
	}
	for _, tc := range cases {
		got, err := Apply(sess(tc.from), tc.cmd, now, "")
		if err != nil {
			t.Fatalf("%s + %s: %v", tc.from, tc.cmd, err)
		}
		if got.State != tc.to {
			t.Fatalf("%s + %s => %s, want %s", tc.from, tc.cmd, got.State, tc.to)
		}
	}
}

func TestForbiddenTransitions(t *testing.T) {
	now := time.Date(2026, 8, 23, 16, 0, 0, 0, time.UTC)
	bads := []struct {
		from model.PomodoroState
		cmd  Command
	}{
		{model.StateIdle, CmdTick},
		{model.StateIdle, CmdPause},
		{model.StateIdle, CmdAbort},
		{model.StatePaused, CmdTick},
		{model.StateCompleted, CmdStart},
		{model.StateAborted, CmdStart},
		{model.StateCompleted, CmdResume},
		{model.StateRunning, CmdStart},
	}
	for _, tc := range bads {
		_, err := Apply(sess(tc.from), tc.cmd, now, "")
		if err == nil {
			t.Fatalf("expected reject %s + %s", tc.from, tc.cmd)
		}
		ae, ok := httpx.IsAppError(err)
		if !ok || ae.Code != "E_INVALID_TRANSITION" {
			t.Fatalf("want E_INVALID_TRANSITION, got %v", err)
		}
	}
}

func TestPauseFreezesRemaining(t *testing.T) {
	now := time.Date(2026, 8, 23, 16, 0, 0, 0, time.UTC)
	s, err := Apply(sess(model.StateIdle), CmdStart, now, "")
	if err != nil {
		t.Fatal(err)
	}
	later := now.Add(5 * time.Minute)
	paused, err := Apply(s, CmdPause, later, "")
	if err != nil {
		t.Fatal(err)
	}
	if paused.RemainingMS(later.Add(10*time.Minute)) != paused.RemainingMS(later) {
		t.Fatal("paused remaining must freeze")
	}
	resumed, err := Apply(paused, CmdResume, later.Add(2*time.Minute), "")
	if err != nil {
		t.Fatal(err)
	}
	if resumed.State != model.StateRunning {
		t.Fatal(resumed.State)
	}
	if resumed.PausedAccumulatedMS < 2*60*1000 {
		t.Fatalf("paused_accumulated %d", resumed.PausedAccumulatedMS)
	}
}

func TestMustLegalTable(t *testing.T) {
	if err := MustLegalTable(); err != nil {
		t.Fatal(err)
	}
}

func TestFakeClockAdvanceCompletes(t *testing.T) {
	clk := NewFakeClock(time.Date(2026, 8, 23, 16, 0, 0, 0, time.UTC))
	done := make(chan struct{}, 1)
	clk.AfterFunc(25*time.Minute, func() { done <- struct{}{} })
	clk.Advance(24 * time.Minute)
	select {
	case <-done:
		t.Fatal("should not fire yet")
	default:
	}
	clk.Advance(time.Minute)
	select {
	case <-done:
	default:
		t.Fatal("timer should fire")
	}
}
