package timeutil

import "time"

func StartOfWeek(t time.Time) time.Time {
	t = StartOfDay(t)
	wd := int(t.Weekday())
	if wd == 0 {
		wd = 7
	}
	return t.AddDate(0, 0, -(wd - 1))
}

func EndOfWeek(t time.Time) time.Time {
	return StartOfWeek(t).AddDate(0, 0, 7).Add(-time.Nanosecond)
}

func StartOfMonth(t time.Time) time.Time {
	t = t.In(Beijing)
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, Beijing)
}

func WeekIndex(start, at time.Time) int {
	s := StartOfWeek(start)
	a := StartOfWeek(at)
	d := int(a.Sub(s).Hours() / 24 / 7)
	if d < 0 {
		return 0
	}
	return d
}

func ClampDate(at, start, due time.Time) time.Time {
	at = StartOfDay(at)
	if at.Before(StartOfDay(start)) {
		return StartOfDay(start)
	}
	if at.After(StartOfDay(due)) {
		return StartOfDay(due)
	}
	return at
}

func EachDay(start, due time.Time, fn func(time.Time)) {
	for d := StartOfDay(start); !d.After(StartOfDay(due)); d = d.Add(24 * time.Hour) {
		fn(d)
	}
}
