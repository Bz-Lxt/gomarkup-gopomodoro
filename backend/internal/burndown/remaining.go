package burndown

import "gopomodoro/internal/model"

// RemainingOfTasks implements the frozen remaining-points formula:
//
//	remaining = Σ estimated − Σ(done.estimated) − Σ(min(consumed, estimated) of unfinished)
//
// which is equivalent to summing Task.Remaining().
func RemainingOfTasks(tasks []model.Task) int {
	total := 0
	for _, t := range tasks {
		total += t.Remaining()
	}
	return total
}

// RemainingBreakdown is used by unit tests to assert the algebraic form.
func RemainingBreakdown(tasks []model.Task) (estimated, doneEstimated, unfinishedConsumed, remaining int) {
	for _, t := range tasks {
		estimated += t.EstimatedPomodoros
		if t.KanbanColumn.IsDone() {
			doneEstimated += t.EstimatedPomodoros
			continue
		}
		c := t.ConsumedPomodoros
		if c > t.EstimatedPomodoros {
			c = t.EstimatedPomodoros
		}
		if c < 0 {
			c = 0
		}
		unfinishedConsumed += c
	}
	remaining = estimated - doneEstimated - unfinishedConsumed
	if remaining < 0 {
		remaining = 0
	}
	return
}
