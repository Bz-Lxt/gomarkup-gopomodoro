package pomodoro

import (
	"testing"
	"time"

	"gopomodoro/internal/model"
)

// TestWorkdayScenario walks a typical independent-dev day:
// start → pause (meeting) → resume → natural complete path via remaining=0.
func TestWorkdayScenario(t *testing.T) {
	now := time.Date(2026, 8, 23, 9, 0, 0, 0, time.UTC)
	s := &model.PomodoroSession{State: model.StateIdle, FocusDurationMS: 25 * 60 * 1000, Version: 1}
	s, err := Apply(s, CmdStart, now, "")
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(10 * time.Minute)
	s, err = Apply(s, CmdPause, now, "")
	if err != nil {
		t.Fatal(err)
	}
	frozen := s.RemainingMS(now)
	now = now.Add(15 * time.Minute)
	if s.RemainingMS(now) != frozen {
		t.Fatal("meeting pause leaked time")
	}
	s, err = Apply(s, CmdResume, now, "")
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Duration(s.RemainingMS(now)) * time.Millisecond)
	s, err = Apply(s, CmdTick, now, "")
	if err != nil {
		t.Fatal(err)
	}
	if s.State != model.StateCompleted {
		t.Fatal(s.State)
	}
	if s.PausedAccumulatedMS < 15*60*1000 {
		t.Fatalf("paused %d", s.PausedAccumulatedMS)
	}
}

func TestTwoAbortsNeverComplete(t *testing.T) {
	now := time.Date(2026, 8, 23, 9, 0, 0, 0, time.UTC)
	s, err := Apply(sess(model.StateIdle), CmdStart, now, "")
	if err != nil {
		t.Fatal(err)
	}
	s, err = Apply(s, CmdAbort, now.Add(time.Minute), string(model.AbortNetworkTimeout))
	if err != nil {
		t.Fatal(err)
	}
	if s.AbortReason != string(model.AbortNetworkTimeout) {
		t.Fatal(s.AbortReason)
	}
	if _, err := Apply(s, CmdTick, now.Add(2*time.Minute), ""); err == nil {
		t.Fatal("aborted cannot complete")
	}
}
