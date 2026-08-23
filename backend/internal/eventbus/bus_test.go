package eventbus

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"gopomodoro/internal/model"
)

func TestPublishIdempotentHandler(t *testing.T) {
	var n atomic.Int32
	seen := map[string]struct{}{}
	b := New(8, 1, func(ev model.DomainEvent) error {
		if _, ok := seen[ev.ID]; ok {
			return nil
		}
		seen[ev.ID] = struct{}{}
		n.Add(1)
		return nil
	})
	ev := model.DomainEvent{ID: "e1", Type: model.DomPomodoroCompleted}
	b.Publish(ev)
	b.Publish(ev)
	time.Sleep(50 * time.Millisecond)
	b.Close(context.Background())
	if n.Load() != 1 {
		t.Fatalf("got %d", n.Load())
	}
}
