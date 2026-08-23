package report

import (
	"time"

	"gopomodoro/internal/burndown"
	"gopomodoro/internal/model"
	"gopomodoro/internal/timeutil"
)

type SeriesPoint struct {
	Day       string  `json:"day"`
	Ideal     float64 `json:"ideal"`
	Actual    *int    `json:"actual,omitempty"`
	IsScope   bool    `json:"is_scope"`
}

// MergeDaySeries aligns irregular actual samples onto the civil-day ideal axis.
// Last observation wins for a given day; missing days carry forward.
func MergeDaySeries(baseline int, start, due time.Time, actual []model.BurndownPoint) []SeriesPoint {
	ideal := burndown.IdealSeries(baseline, start, due)
	byDay := map[string]model.BurndownPoint{}
	for _, p := range actual {
		key := timeutil.FormatDate(p.RecordedAt)
		prev, ok := byDay[key]
		if !ok || p.RecordedAt.After(prev.RecordedAt) {
			byDay[key] = p
		}
	}
	out := make([]SeriesPoint, 0, len(ideal))
	var carry *int
	for _, p := range ideal {
		key := timeutil.FormatDate(p.Day)
		sp := SeriesPoint{Day: key, Ideal: p.Ideal}
		if a, ok := byDay[key]; ok {
			v := a.RemainingPoints
			carry = &v
			sp.IsScope = a.EventType == model.EventScopeChange
		}
		sp.Actual = carry
		out = append(out, sp)
	}
	return out
}

func WeekBuckets(points []SeriesPoint) []SeriesPoint {
	if len(points) == 0 {
		return nil
	}
	var out []SeriesPoint
	var acc *SeriesPoint
	for _, p := range points {
		day, _ := timeutil.ParseDate(p.Day)
		if acc == nil || day.Weekday() == time.Monday {
			cp := p
			acc = &cp
			out = append(out, cp)
			continue
		}
		out[len(out)-1] = p
	}
	return out
}
