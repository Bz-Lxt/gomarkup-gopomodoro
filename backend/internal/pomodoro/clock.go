package pomodoro

import "time"

// Clock is injectable so unit tests can drive the state machine without sleeping.
type Clock interface {
	Now() time.Time
	AfterFunc(d time.Duration, f func()) Timer
}

type Timer interface {
	Stop() bool
}

type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

func (RealClock) AfterFunc(d time.Duration, f func()) Timer {
	return time.AfterFunc(d, f)
}

type FakeClock struct {
	now    time.Time
	timers []*fakeTimer
}

func NewFakeClock(now time.Time) *FakeClock {
	return &FakeClock{now: now}
}

func (c *FakeClock) Now() time.Time { return c.now }

func (c *FakeClock) AfterFunc(d time.Duration, f func()) Timer {
	t := &fakeTimer{deadline: c.now.Add(d), fn: f}
	c.timers = append(c.timers, t)
	return t
}

func (c *FakeClock) Advance(d time.Duration) {
	c.now = c.now.Add(d)
	alive := c.timers[:0]
	for _, t := range c.timers {
		if t.stopped {
			continue
		}
		if !t.deadline.After(c.now) {
			t.stopped = true
			t.fn()
			continue
		}
		alive = append(alive, t)
	}
	c.timers = alive
}

type fakeTimer struct {
	deadline time.Time
	fn       func()
	stopped  bool
}

func (t *fakeTimer) Stop() bool {
	if t.stopped {
		return false
	}
	t.stopped = true
	return true
}
