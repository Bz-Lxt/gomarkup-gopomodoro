package burndown

import (
	"testing"

	"gopomodoro/internal/model"
)

func TestRemainingBoundaries(t *testing.T) {
	over := model.Task{EstimatedPomodoros: 3, ConsumedPomodoros: 5, KanbanColumn: model.ColInProgress}
	early := model.Task{EstimatedPomodoros: 8, ConsumedPomodoros: 2, KanbanColumn: model.ColDone}
	scope := model.Task{EstimatedPomodoros: 5, ConsumedPomodoros: 0, KanbanColumn: model.ColTodo}
	cross := model.Task{EstimatedPomodoros: 4, ConsumedPomodoros: 1, KanbanColumn: model.ColInProgress}

	if RemainingOfTasks([]model.Task{over}) != 0 {
		t.Fatal("over-estimate must clamp to 0")
	}
	if RemainingOfTasks([]model.Task{early}) != 0 {
		t.Fatal("done task remaining is 0 even if under-consumed")
	}
	before := RemainingOfTasks([]model.Task{cross})
	afterAdd := RemainingOfTasks([]model.Task{cross, scope})
	if afterAdd-before != 5 {
		t.Fatalf("scope add should lift remaining by 5, got %d -> %d", before, afterAdd)
	}
	est, doneEst, unfin, rem := RemainingBreakdown([]model.Task{over, early, scope, cross})
	if est != 3+8+5+4 {
		t.Fatalf("estimated %d", est)
	}
	if doneEst != 8 {
		t.Fatalf("done estimated %d", doneEst)
	}
	if unfin != 3+0+1 { // over clamped to 3
		t.Fatalf("unfinished consumed %d", unfin)
	}
	if rem != RemainingOfTasks([]model.Task{over, early, scope, cross}) {
		t.Fatalf("algebra mismatch %d vs %d", rem, RemainingOfTasks([]model.Task{over, early, scope, cross}))
	}
}
