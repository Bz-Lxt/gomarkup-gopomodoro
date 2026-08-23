package ws_test

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"gopomodoro/internal/model"
	"gopomodoro/internal/pomodoro"
	"gopomodoro/internal/ws"
)

func TestHubConcurrentBroadcastAndRemove(t *testing.T) {
	h := ws.NewHub()
	userID := uuid.New()
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 2000; i++ {
		client := h.Add(userID)
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			h.Remove(client)
			runtime.Gosched()
		}()
	}

	now := time.Now()
	end := now.Add(time.Minute)
	view := pomodoro.SessionView{Session: &model.PomodoroSession{
		ID: userID, State: model.StateRunning, ExpectedEndAt: &end, FocusDurationMS: 60_000,
	}}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 100; i++ {
			h.BroadcastSession(userID, view)
			runtime.Gosched()
		}
	}()
	close(start)
	wg.Wait()
}
