package handler

import "gopomodoro/internal/model"

type AuthResponse struct {
	User any `json:"user"`
	Auth any `json:"auth"`
}

type ReorderRequest struct {
	Items []reorderItem `json:"items"`
}

func taskPublic(t *model.Task) map[string]any {
	return map[string]any{
		"id":                   t.ID,
		"project_id":           t.ProjectID,
		"milestone_id":         t.MilestoneID,
		"title":                t.Title,
		"estimated_pomodoros":  t.EstimatedPomodoros,
		"consumed_pomodoros":   t.ConsumedPomodoros,
		"remaining_pomodoros":  t.Remaining(),
		"kanban_column":        t.KanbanColumn,
		"sort_order":           t.SortOrder,
	}
}

func tasksPublic(list []model.Task) []map[string]any {
	out := make([]map[string]any, 0, len(list))
	for i := range list {
		out = append(out, taskPublic(&list[i]))
	}
	return out
}
