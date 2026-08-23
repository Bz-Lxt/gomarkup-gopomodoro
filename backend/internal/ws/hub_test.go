package ws

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gopomodoro/internal/model"
	"gopomodoro/internal/pomodoro"
)

func TestHubBroadcastDoesNotBlock(t *testing.T) {
	h := NewHub()
	c := h.Add(uuid.New())
	// fill buffer
	for i := 0; i < 40; i++ {
		h.trySend(c, Outbound{Type: TypeTick})
	}
	uid := uuid.New()
	c2 := h.Add(uid)
	now := time.Now()
	end := now.Add(time.Minute)
	h.BroadcastSession(uid, pomodoro.SessionView{Session: &model.PomodoroSession{
		ID: uid, State: model.StateRunning, ExpectedEndAt: &end, FocusDurationMS: 1500000,
	}})
	select {
	case msg := <-c2.send:
		if msg.Type != TypeSession && msg.Type != TypeTick {
			t.Fatalf("%s", msg.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
	h.Remove(c)
	h.Remove(c2)
}

func TestSubscribeMilestone(t *testing.T) {
	h := NewHub()
	c := h.Add(uuid.New())
	mid := uuid.New()
	h.Subscribe(c, mid)
	h.BroadcastBurndown(mid, model.BurndownPoint{MilestoneID: mid, RemainingPoints: 3})
	select {
	case msg := <-c.send:
		if msg.Type != TypeBurndown {
			t.Fatal(msg.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("no burndown")
	}
}
