package report

import (
	"math"
	"time"

	"gopomodoro/internal/burndown"
	"gopomodoro/internal/model"
	"gopomodoro/internal/timeutil"
)

type Forecast struct {
	Remaining       int     `json:"remaining"`
	Velocity        float64 `json:"velocity"`
	OptimisticDate  string  `json:"optimistic_date"`
	LikelyDate      string  `json:"likely_date"`
	PessimisticDate string  `json:"pessimistic_date"`
	OnTrack         bool    `json:"on_track"`
	SlackDays       int     `json:"slack_days"`
}

// Triple computes P50 / optimistic / pessimistic finish dates from remaining
// points and a week-window velocity. Optimistic assumes +30% velocity;
// pessimistic assumes -30%.
func Triple(remaining int, velocity float64, due, now time.Time) Forecast {
	f := Forecast{Remaining: remaining, Velocity: velocity}
	if remaining <= 0 {
		d := timeutil.FormatDate(now)
		f.OptimisticDate, f.LikelyDate, f.PessimisticDate = d, d, d
		f.OnTrack = true
		return f
	}
	f.LikelyDate = burndown.PredictDoneOn(remaining, velocity, now)
	f.OptimisticDate = burndown.PredictDoneOn(remaining, velocity*1.3, now)
	f.PessimisticDate = burndown.PredictDoneOn(remaining, velocity*0.7, now)
	left := timeutil.DaysBetween(now, due)
	need := 0
	if velocity > 0 {
		need = int(math.Ceil(float64(remaining) / velocity))
	} else {
		need = left + 30
	}
	f.SlackDays = left - need
	f.OnTrack = f.SlackDays >= 0 && velocity > 0
	return f
}

func ScopePressure(baseline, remaining int, start, due, now time.Time) float64 {
	ideal := burndown.IdealAt(baseline, start, due, now)
	if ideal <= 0 {
		if remaining <= 0 {
			return 0
		}
		return 2
	}
	return float64(remaining) / ideal
}

func SummarizeTasks(tasks []model.Task) map[string]int {
	out := map[string]int{
		"count": 0, "estimated": 0, "consumed": 0, "remaining": 0,
		"done": 0, "in_progress": 0, "todo": 0, "backlog": 0,
	}
	for _, t := range tasks {
		out["count"]++
		out["estimated"] += t.EstimatedPomodoros
		out["consumed"] += t.ConsumedPomodoros
		out["remaining"] += t.Remaining()
		switch t.KanbanColumn {
		case model.ColDone:
			out["done"]++
		case model.ColInProgress:
			out["in_progress"]++
		case model.ColTodo:
			out["todo"]++
		default:
			out["backlog"]++
		}
	}
	return out
}
