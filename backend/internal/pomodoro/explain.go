package pomodoro

import (
	"fmt"
	"strings"

	"gopomodoro/internal/model"
)

// Explain returns a stable, client-displayable reason for why a command
// is accepted or rejected from the given state.
func Explain(from model.PomodoroState, cmd Command) string {
	if CanTransit(from, targetOf(cmd), cmd) {
		return fmt.Sprintf("允许：%s --%s--> %s", from, cmd, targetOf(cmd))
	}
	switch {
	case from == model.StateIdle && cmd == CmdTick:
		return "禁止：空闲会话不能直接结算，必须先 start 并走完倒计时"
	case from == model.StatePaused && cmd == CmdTick:
		return "禁止：暂停中不能自然完成，须 resume 后再由服务端 tick"
	case from.Terminal() && cmd == CmdStart:
		return "禁止：终态会话不可复活，请开新的番茄钟"
	case from == model.StateIdle && (cmd == CmdPause || cmd == CmdResume || cmd == CmdAbort):
		return "禁止：尚未开始的会话没有可暂停/恢复/放弃的运行态"
	case from == model.StateRunning && cmd == CmdStart:
		return "禁止：已在专注中，不能重复 start"
	case from == model.StatePaused && cmd == CmdPause:
		return "禁止：已暂停"
	default:
		return fmt.Sprintf("禁止：%s 状态下不接受 %s", from, cmd)
	}
}

func targetOf(cmd Command) model.PomodoroState {
	switch cmd {
	case CmdStart, CmdResume:
		return model.StateRunning
	case CmdPause:
		return model.StatePaused
	case CmdAbort:
		return model.StateAborted
	case CmdTick:
		return model.StateCompleted
	default:
		return model.StateIdle
	}
}

// LegalCommands lists commands a client may send from the current state.
func LegalCommands(from model.PomodoroState) []Command {
	var out []Command
	for _, cmd := range []Command{CmdStart, CmdPause, CmdResume, CmdAbort, CmdTick} {
		if CanTransit(from, targetOf(cmd), cmd) {
			if cmd == CmdTick {
				continue // never expose tick to ordinary clients
			}
			out = append(out, cmd)
		}
	}
	return out
}

func FormatLegal(from model.PomodoroState) string {
	cmds := LegalCommands(from)
	if len(cmds) == 0 {
		return "终态，无后续命令"
	}
	parts := make([]string, 0, len(cmds))
	for _, c := range cmds {
		parts = append(parts, string(c))
	}
	return strings.Join(parts, ",")
}
