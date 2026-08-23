package model

import "time"

type User struct {
	ID           ID        `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	DisplayName  string    `json:"display_name"`
	Timezone     string    `json:"timezone"`
	CreatedAt    time.Time `json:"created_at"`
}

type Project struct {
	ID          ID        `json:"id"`
	UserID      ID        `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Archived    bool      `json:"archived"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Milestone struct {
	ID             ID               `json:"id"`
	ProjectID      ID               `json:"project_id"`
	Title          string           `json:"title"`
	StartDate      time.Time        `json:"start_date"`
	DueDate        time.Time        `json:"due_date"`
	BaselinePoints int              `json:"baseline_points"`
	Status         MilestoneStatus  `json:"status"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	RemainingPoints int             `json:"remaining_points,omitempty"`
	Risk           string           `json:"risk,omitempty"`
}

type Task struct {
	ID                  ID           `json:"id"`
	ProjectID           ID           `json:"project_id"`
	MilestoneID         *ID          `json:"milestone_id"`
	Title               string       `json:"title"`
	EstimatedPomodoros  int          `json:"estimated_pomodoros"`
	ConsumedPomodoros   int          `json:"consumed_pomodoros"`
	KanbanColumn        KanbanColumn `json:"kanban_column"`
	SortOrder           int          `json:"sort_order"`
	CreatedAt           time.Time    `json:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at"`
}

func (t Task) Remaining() int {
	if t.KanbanColumn.IsDone() {
		return 0
	}
	r := t.EstimatedPomodoros - t.ConsumedPomodoros
	if r < 0 {
		return 0
	}
	return r
}

type PomodoroSession struct {
	ID                   ID            `json:"id"`
	UserID               ID            `json:"user_id"`
	TaskID               ID            `json:"task_id"`
	State                PomodoroState `json:"state"`
	FocusDurationMS      int64         `json:"focus_duration_ms"`
	StartedAt            *time.Time    `json:"started_at"`
	PausedAt             *time.Time    `json:"paused_at"`
	PausedAccumulatedMS  int64         `json:"paused_accumulated_ms"`
	ExpectedEndAt        *time.Time    `json:"expected_end_at"`
	EndedAt              *time.Time    `json:"ended_at"`
	AbortReason          string        `json:"abort_reason,omitempty"`
	ResumeToken          string        `json:"resume_token"`
	Version              int64         `json:"version"`
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`
}

func (s *PomodoroSession) RemainingMS(now time.Time) int64 {
	switch s.State {
	case StateIdle:
		return s.FocusDurationMS
	case StatePaused:
		if s.ExpectedEndAt == nil || s.PausedAt == nil {
			return s.FocusDurationMS
		}
		left := s.ExpectedEndAt.Sub(*s.PausedAt).Milliseconds()
		if left < 0 {
			return 0
		}
		return left
	case StateRunning:
		if s.ExpectedEndAt == nil {
			return s.FocusDurationMS
		}
		left := s.ExpectedEndAt.Sub(now).Milliseconds()
		if left < 0 {
			return 0
		}
		return left
	default:
		return 0
	}
}

type BurndownPoint struct {
	ID              ID                 `json:"id"`
	MilestoneID     ID                 `json:"milestone_id"`
	RecordedAt      time.Time          `json:"recorded_at"`
	RemainingPoints int                `json:"remaining_points"`
	IdealPoints     float64            `json:"ideal_points"`
	EventType       BurndownEventType  `json:"event_type"`
	EventID         string             `json:"event_id"`
}

type ScopeChangeLog struct {
	ID          ID        `json:"id"`
	MilestoneID ID        `json:"milestone_id"`
	DeltaPoints int       `json:"delta_points"`
	Reason      string    `json:"reason"`
	OccurredAt  time.Time `json:"occurred_at"`
}

type DomainEvent struct {
	ID          string          `json:"id"`
	Type        DomainEventType `json:"type"`
	UserID      ID              `json:"user_id"`
	ProjectID   ID              `json:"project_id"`
	MilestoneID *ID             `json:"milestone_id,omitempty"`
	TaskID      *ID             `json:"task_id,omitempty"`
	SessionID   *ID             `json:"session_id,omitempty"`
	Payload     map[string]any  `json:"payload,omitempty"`
	OccurredAt  time.Time       `json:"occurred_at"`
}

type EfficiencyMetrics struct {
	TodayCompleted   int     `json:"today_completed"`
	WeekCompleted    int     `json:"week_completed"`
	TodayAborted     int     `json:"today_aborted"`
	WeekAborted      int     `json:"week_aborted"`
	AbortRate        float64 `json:"abort_rate"`
	AvgDailyVelocity float64 `json:"avg_daily_velocity"`
	PredictedDoneOn  string  `json:"predicted_done_on,omitempty"`
}
