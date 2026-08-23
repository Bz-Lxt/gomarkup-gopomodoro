package pomodoro

import (
	"testing"
	"time"

	"gopomodoro/internal/httpx"
	"gopomodoro/internal/model"
)

func TestExhaustiveMatrix(t *testing.T) {
	states := []model.PomodoroState{
		model.StateIdle, model.StateRunning, model.StatePaused, model.StateCompleted, model.StateAborted,
	}
	cmds := []Command{CmdStart, CmdPause, CmdResume, CmdAbort, CmdTick}
	now := time.Date(2026, 8, 23, 17, 0, 0, 0, time.UTC)
	allowed := 0
	denied := 0
	for _, st := range states {
		for _, cmd := range cmds {
			_, err := Apply(sess(st), cmd, now, "test")
			legal := CanTransit(st, targetOf(cmd), cmd)
			if legal && err != nil {
				t.Fatalf("legal %s+%s rejected: %v", st, cmd, err)
			}
			if !legal && err == nil {
				t.Fatalf("illegal %s+%s accepted", st, cmd)
			}
			if !legal {
				ae, ok := httpx.IsAppError(err)
				if !ok || ae.Code != "E_INVALID_TRANSITION" {
					t.Fatalf("illegal %s+%s wrong error %v", st, cmd, err)
				}
				denied++
			} else {
				allowed++
			}
			_ = Explain(st, cmd)
		}
	}
	if allowed != 6 {
		t.Fatalf("expected 6 legal cells, got %d", allowed)
	}
	if denied != 19 {
		t.Fatalf("expected 19 illegal cells, got %d", denied)
	}
}

func TestLegalCommandsHideTick(t *testing.T) {
	cmds := LegalCommands(model.StateRunning)
	for _, c := range cmds {
		if c == CmdTick {
			t.Fatal("tick must not be client-visible")
		}
	}
	if FormatLegal(model.StateCompleted) != "终态，无后续命令" {
		t.Fatal(FormatLegal(model.StateCompleted))
	}
}

func TestExplainIdleTick(t *testing.T) {
	s := Explain(model.StateIdle, CmdTick)
	if s == "" || s[:6] != "禁止" && s[:6] != "禁止：" {
		if Explain(model.StateIdle, CmdTick) == "" {
			t.Fatal("empty")
		}
	}
}
