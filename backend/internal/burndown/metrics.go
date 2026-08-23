package burndown

import (
	"math"
	"time"

	"gopomodoro/internal/model"
	"gopomodoro/internal/timeutil"
)

func AbortRate(completed, aborted int) float64 {
	den := completed + aborted
	if den == 0 {
		return 0
	}
	return float64(aborted) / float64(den)
}

func AvgDailyVelocity(completed int, windowDays int) float64 {
	if windowDays < 1 {
		windowDays = 1
	}
	return float64(completed) / float64(windowDays)
}

func PredictDoneOn(remaining int, velocity float64, from time.Time) string {
	if remaining <= 0 {
		return timeutil.FormatDate(from)
	}
	if velocity <= 0 {
		return ""
	}
	days := int(math.Ceil(float64(remaining) / velocity))
	if days < 0 {
		days = 0
	}
	return timeutil.FormatDate(timeutil.StartOfDay(from).AddDate(0, 0, days))
}

func Risk(remaining int, due time.Time, velocity float64, now time.Time) string {
	if remaining <= 0 {
		return "done"
	}
	leftDays := timeutil.DaysBetween(now, due)
	if velocity <= 0 {
		if timeutil.StartOfDay(now).After(timeutil.StartOfDay(due)) {
			return "overdue"
		}
		return "unknown"
	}
	need := int(math.Ceil(float64(remaining) / velocity))
	if timeutil.StartOfDay(now).After(timeutil.StartOfDay(due)) {
		return "overdue"
	}
	if need > leftDays {
		return "at_risk"
	}
	if need+2 >= leftDays {
		return "tight"
	}
	return "on_track"
}

func Assemble(m *model.EfficiencyMetrics, remaining int, now time.Time) {
	weekDays := 7
	m.AbortRate = AbortRate(m.WeekCompleted+m.TodayCompleted, m.WeekAborted+m.TodayAborted)
	// Prefer week window for velocity; fall back to today.
	completed := m.WeekCompleted
	if completed == 0 {
		completed = m.TodayCompleted
		weekDays = 1
	}
	m.AvgDailyVelocity = AvgDailyVelocity(completed, weekDays)
	m.PredictedDoneOn = PredictDoneOn(remaining, m.AvgDailyVelocity, now)
}
