package model

import "github.com/google/uuid"

type ID = uuid.UUID

func NewID() ID { return uuid.New() }

func ParseID(s string) (ID, error) { return uuid.Parse(s) }

type KanbanColumn string

const (
	ColBacklog    KanbanColumn = "backlog"
	ColTodo       KanbanColumn = "todo"
	ColInProgress KanbanColumn = "in_progress"
	ColDone       KanbanColumn = "done"
)

func (c KanbanColumn) Valid() bool {
	switch c {
	case ColBacklog, ColTodo, ColInProgress, ColDone:
		return true
	}
	return false
}

func (c KanbanColumn) IsDone() bool { return c == ColDone }

type MilestoneStatus string

const (
	MSPlanning MilestoneStatus = "planning"
	MSActive   MilestoneStatus = "active"
	MSDone     MilestoneStatus = "done"
)

func (s MilestoneStatus) Valid() bool {
	switch s {
	case MSPlanning, MSActive, MSDone:
		return true
	}
	return false
}

type PomodoroState string

const (
	StateIdle      PomodoroState = "idle"
	StateRunning   PomodoroState = "running"
	StatePaused    PomodoroState = "paused"
	StateCompleted PomodoroState = "completed"
	StateAborted   PomodoroState = "aborted"
)

func (s PomodoroState) Terminal() bool {
	return s == StateCompleted || s == StateAborted
}

func (s PomodoroState) Active() bool {
	return s == StateRunning || s == StatePaused
}

type AbortReason string

const (
	AbortUser           AbortReason = "user"
	AbortNetworkTimeout AbortReason = "network_timeout"
)

type BurndownEventType string

const (
	EventPomodoroCompleted BurndownEventType = "pomodoro_completed"
	EventTaskDone          BurndownEventType = "task_done"
	EventScopeChange       BurndownEventType = "scope_change"
	EventSnapshot          BurndownEventType = "snapshot"
)

type DomainEventType string

const (
	DomPomodoroCompleted DomainEventType = "pomodoro.completed"
	DomPomodoroAborted   DomainEventType = "pomodoro.aborted"
	DomTaskDone          DomainEventType = "task.done"
	DomScopeChanged      DomainEventType = "scope.changed"
)
