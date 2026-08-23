package model

import "testing"

func TestTaskRemainingClamp(t *testing.T) {
	over := Task{EstimatedPomodoros: 2, ConsumedPomodoros: 9, KanbanColumn: ColTodo}
	if over.Remaining() != 0 {
		t.Fatal(over.Remaining())
	}
	done := Task{EstimatedPomodoros: 9, ConsumedPomodoros: 1, KanbanColumn: ColDone}
	if done.Remaining() != 0 {
		t.Fatal("done")
	}
	open := Task{EstimatedPomodoros: 5, ConsumedPomodoros: 2, KanbanColumn: ColInProgress}
	if open.Remaining() != 3 {
		t.Fatal(open.Remaining())
	}
}
