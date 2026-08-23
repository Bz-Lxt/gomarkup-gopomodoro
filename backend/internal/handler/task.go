package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopomodoro/internal/auth"
	"gopomodoro/internal/httpx"
	"gopomodoro/internal/model"
	"gopomodoro/internal/timeutil"
	"gopomodoro/internal/validate"
)

type taskReq struct {
	Title              string  `json:"title"`
	EstimatedPomodoros int     `json:"estimated_pomodoros"`
	KanbanColumn       string  `json:"kanban_column"`
	MilestoneID        *string `json:"milestone_id"`
	SortOrder          *int    `json:"sort_order"`
}

type reorderItem struct {
	ID           string `json:"id"`
	KanbanColumn string `json:"kanban_column"`
	SortOrder    int    `json:"sort_order"`
}

func ListTasks(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		pid, err := uuid.Parse(c.Param("id"))
		if err != nil {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		if _, err := d.DB.ProjectByID(c.Request.Context(), auth.UserID(c), pid); err != nil {
			httpx.Fail(c, err)
			return
		}
		var mid *model.ID
		if raw := c.Query("milestone_id"); raw != "" {
			id, err := uuid.Parse(raw)
			if err != nil {
				httpx.Fail(c, httpx.ErrValidation)
				return
			}
			mid = &id
		}
		list, err := d.DB.ListTasks(c.Request.Context(), pid, mid)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		if list == nil {
			list = []model.Task{}
		}
		httpx.OK(c, list)
	}
}

func CreateTask(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		pid, err := uuid.Parse(c.Param("id"))
		if err != nil {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		if _, err := d.DB.ProjectByID(c.Request.Context(), auth.UserID(c), pid); err != nil {
			httpx.Fail(c, err)
			return
		}
		var req taskReq
		if err := c.ShouldBindJSON(&req); err != nil {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		title, err := validate.RequiredString("title", req.Title, 1, 160)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		est, err := validate.PositiveInt("estimated_pomodoros", req.EstimatedPomodoros, 99)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		col := model.ColBacklog
		if req.KanbanColumn != "" {
			col, err = validate.Column(req.KanbanColumn)
			if err != nil {
				httpx.Fail(c, err)
				return
			}
		}
		var mid *model.ID
		if req.MilestoneID != nil && *req.MilestoneID != "" {
			id, err := uuid.Parse(*req.MilestoneID)
			if err != nil {
				httpx.Fail(c, validate.Fail("milestone_id", "非法 UUID"))
				return
			}
			if _, err := d.DB.MilestoneOwnedBy(c.Request.Context(), auth.UserID(c), id); err != nil {
				httpx.Fail(c, err)
				return
			}
			mid = &id
		}
		now := timeutil.Now()
		t := &model.Task{
			ID: uuid.New(), ProjectID: pid, MilestoneID: mid, Title: title,
			EstimatedPomodoros: est, KanbanColumn: col, CreatedAt: now, UpdatedAt: now,
		}
		if err := d.DB.CreateTask(c.Request.Context(), t); err != nil {
			httpx.Fail(c, err)
			return
		}
		if mid != nil {
			d.Registry.PublishScope(auth.UserID(c), pid, mid, &t.ID, est, "task_added")
		}
		httpx.Created(c, t)
	}
}

func UpdateTask(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		t, err := d.DB.TaskOwnedBy(c.Request.Context(), auth.UserID(c), id)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		var req taskReq
		if err := c.ShouldBindJSON(&req); err != nil {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		oldRemain := t.Remaining()
		oldMid := t.MilestoneID
		wasDone := t.KanbanColumn.IsDone()
		if req.Title != "" {
			title, err := validate.RequiredString("title", req.Title, 1, 160)
			if err != nil {
				httpx.Fail(c, err)
				return
			}
			t.Title = title
		}
		if req.EstimatedPomodoros > 0 {
			est, err := validate.PositiveInt("estimated_pomodoros", req.EstimatedPomodoros, 99)
			if err != nil {
				httpx.Fail(c, err)
				return
			}
			t.EstimatedPomodoros = est
		}
		if req.KanbanColumn != "" {
			col, err := validate.Column(req.KanbanColumn)
			if err != nil {
				httpx.Fail(c, err)
				return
			}
			t.KanbanColumn = col
		}
		if req.SortOrder != nil {
			t.SortOrder = *req.SortOrder
		}
		if req.MilestoneID != nil {
			if *req.MilestoneID == "" {
				t.MilestoneID = nil
			} else {
				mid, err := uuid.Parse(*req.MilestoneID)
				if err != nil {
					httpx.Fail(c, validate.Fail("milestone_id", "非法 UUID"))
					return
				}
				t.MilestoneID = &mid
			}
		}
		t.UpdatedAt = timeutil.Now()
		if err := d.DB.UpdateTask(c.Request.Context(), t); err != nil {
			httpx.Fail(c, err)
			return
		}
		uid := auth.UserID(c)
		if !wasDone && t.KanbanColumn.IsDone() {
			d.Registry.PublishTaskDone(uid, t.ProjectID, t.MilestoneID, t.ID)
		} else {
			delta := t.Remaining() - oldRemain
			if oldMid != nil && (t.MilestoneID == nil || *t.MilestoneID != *oldMid) {
				d.Registry.PublishScope(uid, t.ProjectID, oldMid, &t.ID, -oldRemain, "task_unbound")
				if t.MilestoneID != nil {
					d.Registry.PublishScope(uid, t.ProjectID, t.MilestoneID, &t.ID, t.Remaining(), "task_bound")
				}
			} else if delta != 0 && t.MilestoneID != nil {
				d.Registry.PublishScope(uid, t.ProjectID, t.MilestoneID, &t.ID, delta, "estimate_changed")
			}
		}
		httpx.OK(c, t)
	}
}

func DeleteTask(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		t, err := d.DB.TaskOwnedBy(c.Request.Context(), auth.UserID(c), id)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		if err := d.DB.DeleteTask(c.Request.Context(), id); err != nil {
			httpx.Fail(c, err)
			return
		}
		if t.MilestoneID != nil {
			d.Registry.PublishScope(auth.UserID(c), t.ProjectID, t.MilestoneID, &t.ID, -t.Remaining(), "task_deleted")
		}
		httpx.OK(c, gin.H{"deleted": true})
	}
}

func ReorderTasks(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Items []reorderItem `json:"items"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || len(req.Items) == 0 {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		uid := auth.UserID(c)
		updates := make([]struct {
			ID     uuid.UUID
			Column model.KanbanColumn
			Order  int
		}, 0, len(req.Items))
		var doneTasks []model.Task
		for _, it := range req.Items {
			id, err := uuid.Parse(it.ID)
			if err != nil {
				httpx.Fail(c, httpx.ErrValidation)
				return
			}
			col, err := validate.Column(it.KanbanColumn)
			if err != nil {
				httpx.Fail(c, err)
				return
			}
			t, err := d.DB.TaskOwnedBy(c.Request.Context(), uid, id)
			if err != nil {
				httpx.Fail(c, err)
				return
			}
			if !t.KanbanColumn.IsDone() && col.IsDone() {
				doneTasks = append(doneTasks, *t)
			}
			updates = append(updates, struct {
				ID     uuid.UUID
				Column model.KanbanColumn
				Order  int
			}{id, col, it.SortOrder})
		}
		if err := d.DB.ReorderTasks(c.Request.Context(), updates); err != nil {
			httpx.Fail(c, err)
			return
		}
		for _, t := range doneTasks {
			d.Registry.PublishTaskDone(uid, t.ProjectID, t.MilestoneID, t.ID)
		}
		httpx.OK(c, gin.H{"updated": len(updates)})
	}
}
