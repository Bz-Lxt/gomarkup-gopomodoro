package pomodoro

import (
	"sync"
	"time"

	"gopomodoro/internal/model"
)

// TransitionRecord is an in-process audit trail used by the live dashboard
// and by tests that assert no illegal command was ever accepted.
type TransitionRecord struct {
	SessionID model.ID
	UserID    model.ID
	From      model.PomodoroState
	To        model.PomodoroState
	Cmd       Command
	At        time.Time
	Reason    string
}

type AuditLog struct {
	mu   sync.RWMutex
	ring []TransitionRecord
	cap  int
}

func NewAuditLog(n int) *AuditLog {
	if n < 32 {
		n = 256
	}
	return &AuditLog{ring: make([]TransitionRecord, 0, n), cap: n}
}

func (a *AuditLog) Append(r TransitionRecord) {
	if a == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.ring) >= a.cap {
		a.ring = a.ring[1:]
	}
	a.ring = append(a.ring, r)
}

func (a *AuditLog) Recent(n int) []TransitionRecord {
	if a == nil {
		return nil
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	if n <= 0 || n > len(a.ring) {
		n = len(a.ring)
	}
	return a.ring[len(a.ring)-n:]
}

func (a *AuditLog) CountByTo(state model.PomodoroState) int {
	if a == nil {
		return 0
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	n := 0
	for _, r := range a.ring {
		if r.To == state {
			n++
		}
	}
	return n
}
