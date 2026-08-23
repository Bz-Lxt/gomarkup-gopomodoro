package pomodoro

import (
	"sync"
	"testing"
	"time"
)

// TestSweepDoesNotHoldRegistryWhileLockingSession documents the deadlock
// that caused nginx 504: sweep used to RLock registry then Lock session,
// while Command held session then Lock registry in drop().
func TestSweepSnapshotPattern(t *testing.T) {
	var registry, session sync.RWMutex
	done := make(chan struct{})
	go func() {
		registry.RLock()
		lives := []int{1}
		registry.RUnlock()
		for range lives {
			session.Lock()
			session.Unlock()
		}
		close(done)
	}()
	go func() {
		session.Lock()
		registry.Lock()
		registry.Unlock()
		session.Unlock()
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("lock order still deadlocks")
	}
}
