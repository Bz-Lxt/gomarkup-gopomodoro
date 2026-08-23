package pomodoro

import (
	"fmt"
	"time"

	"gopomodoro/internal/httpx"
	"gopomodoro/internal/model"
)

type Command string

const (
	CmdStart  Command = "start"
	CmdPause  Command = "pause"
	CmdResume Command = "resume"
	CmdAbort  Command = "abort"
	CmdTick   Command = "tick" // server-only: natural completion
)

type Transition struct {
	From model.PomodoroState
	To   model.PomodoroState
	Cmd  Command
}

var legal = map[Transition]struct{}{
	{model.StateIdle, model.StateRunning, CmdStart}:         {},
	{model.StateRunning, model.StatePaused, CmdPause}:       {},
	{model.StatePaused, model.StateRunning, CmdResume}:      {},
	{model.StateRunning, model.StateAborted, CmdAbort}:      {},
	{model.StatePaused, model.StateAborted, CmdAbort}:       {},
	{model.StateRunning, model.StateCompleted, CmdTick}:     {},
}

func CanTransit(from, to model.PomodoroState, cmd Command) bool {
	_, ok := legal[Transition{from, to, cmd}]
	return ok
}

func Illegal(from model.PomodoroState, cmd Command) error {
	return httpx.ErrInvalidTransition.WithDetails(map[string]any{
		"from":    from,
		"command": cmd,
		"hint":    "allowed: idle→running(start), running→paused(pause), paused→running(resume), running|paused→aborted(abort), running→completed(tick)",
	})
}

// Apply mutates a copy of the session according to cmd. Caller holds the session lock.
func Apply(s *model.PomodoroSession, cmd Command, now time.Time, reason string) (*model.PomodoroSession, error) {
	next := *s
	var to model.PomodoroState
	switch cmd {
	case CmdStart:
		to = model.StateRunning
		if !CanTransit(s.State, to, cmd) {
			return nil, Illegal(s.State, cmd)
		}
		end := now.Add(time.Duration(s.FocusDurationMS) * time.Millisecond)
		next.State = to
		next.StartedAt = &now
		next.PausedAt = nil
		next.PausedAccumulatedMS = 0
		next.ExpectedEndAt = &end
		next.EndedAt = nil
		next.AbortReason = ""
	case CmdPause:
		to = model.StatePaused
		if !CanTransit(s.State, to, cmd) {
			return nil, Illegal(s.State, cmd)
		}
		next.State = to
		p := now
		next.PausedAt = &p
	case CmdResume:
		to = model.StateRunning
		if !CanTransit(s.State, to, cmd) {
			return nil, Illegal(s.State, cmd)
		}
		left := s.RemainingMS(now)
		end := now.Add(time.Duration(left) * time.Millisecond)
		if s.PausedAt != nil {
			next.PausedAccumulatedMS += now.Sub(*s.PausedAt).Milliseconds()
		}
		next.State = to
		next.PausedAt = nil
		next.ExpectedEndAt = &end
	case CmdAbort:
		to = model.StateAborted
		if !CanTransit(s.State, to, cmd) {
			return nil, Illegal(s.State, cmd)
		}
		next.State = to
		e := now
		next.EndedAt = &e
		if reason == "" {
			reason = string(model.AbortUser)
		}
		next.AbortReason = reason
		next.ExpectedEndAt = s.ExpectedEndAt
	case CmdTick:
		to = model.StateCompleted
		if !CanTransit(s.State, to, cmd) {
			return nil, Illegal(s.State, cmd)
		}
		next.State = to
		e := now
		next.EndedAt = &e
		next.PausedAt = nil
	default:
		return nil, Illegal(s.State, cmd)
	}
	next.UpdatedAt = now
	next.Version = s.Version + 1
	return &next, nil
}

func MustLegalTable() error {
	// Guard against accidental table edits: these four must stay forbidden.
	forbidden := []Transition{
		{model.StateIdle, model.StateCompleted, CmdTick},
		{model.StatePaused, model.StateCompleted, CmdTick},
		{model.StateCompleted, model.StateRunning, CmdStart},
		{model.StateAborted, model.StateRunning, CmdStart},
		{model.StateIdle, model.StateCompleted, CmdStart},
	}
	for _, t := range forbidden {
		if CanTransit(t.From, t.To, t.Cmd) {
			return fmt.Errorf("forbidden transition leaked into table: %+v", t)
		}
	}
	return nil
}
