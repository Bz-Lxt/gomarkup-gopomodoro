package eventbus

import (
	"context"
	"sync"

	"gopomodoro/internal/logger"
	"gopomodoro/internal/model"
)

type Bus struct {
	ch      chan model.DomainEvent
	handler Handler
	wg      sync.WaitGroup
	closed  chan struct{}
	once    sync.Once
}

func New(buffer int, workers int, h Handler) *Bus {
	if buffer < 16 {
		buffer = 64
	}
	if workers < 1 {
		workers = 2
	}
	b := &Bus{
		ch:      make(chan model.DomainEvent, buffer),
		handler: h,
		closed:  make(chan struct{}),
	}
	for i := 0; i < workers; i++ {
		b.wg.Add(1)
		go b.worker(i)
	}
	return b
}

func (b *Bus) Publish(ev model.DomainEvent) {
	select {
	case <-b.closed:
		logger.L().Error("event dropped, bus closed", "event_id", ev.ID)
	case b.ch <- ev:
	}
}

func (b *Bus) worker(id int) {
	defer b.wg.Done()
	for ev := range b.ch {
		if err := b.handler(ev); err != nil {
			logger.L().Error("event handler failed", "worker", id, "event_id", ev.ID, "err", err)
		}
	}
}

func (b *Bus) Close(ctx context.Context) {
	b.once.Do(func() {
		close(b.closed)
		close(b.ch)
	})
	done := make(chan struct{})
	go func() {
		b.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		logger.L().Error("event bus close timeout")
	}
}
