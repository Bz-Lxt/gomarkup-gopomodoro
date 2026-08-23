package burndown

import (
	"time"

	"gopomodoro/internal/timeutil"
)

// IdealAt returns the linear interpolation of remaining points on the ideal line.
func IdealAt(baseline int, start, due, at time.Time) float64 {
	if baseline <= 0 {
		return 0
	}
	total := timeutil.DaysBetween(start, due)
	elapsed := timeutil.ElapsedDays(start, at)
	if elapsed <= 0 {
		return float64(baseline)
	}
	if elapsed >= total {
		return 0
	}
	return float64(baseline) * (1 - float64(elapsed)/float64(total))
}

// IdealSeries emits one point per civil day from start to due inclusive.
func IdealSeries(baseline int, start, due time.Time) []struct {
	Day   time.Time
	Ideal float64
} {
	start = timeutil.StartOfDay(start)
	due = timeutil.StartOfDay(due)
	if due.Before(start) {
		due = start
	}
	var out []struct {
		Day   time.Time
		Ideal float64
	}
	for d := start; !d.After(due); d = d.Add(24 * time.Hour) {
		out = append(out, struct {
			Day   time.Time
			Ideal float64
		}{Day: d, Ideal: IdealAt(baseline, start, due, d)})
	}
	return out
}
