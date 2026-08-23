package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopomodoro/internal/auth"
	"gopomodoro/internal/burndown"
	"gopomodoro/internal/httpx"
	"gopomodoro/internal/model"
	"gopomodoro/internal/timeutil"
	"gopomodoro/internal/validate"
)

type milestoneReq struct {
	Title          string `json:"title"`
	StartDate      string `json:"start_date"`
	DueDate        string `json:"due_date"`
	BaselinePoints int    `json:"baseline_points"`
	Status         string `json:"status"`
}

func ListMilestones(d *Deps) gin.HandlerFunc {
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
		list, err := d.DB.ListMilestones(c.Request.Context(), pid)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		now := timeutil.Now()
		for i := range list {
			tasks, _ := d.DB.ListTasksByMilestone(c.Request.Context(), list[i].ID)
			list[i].RemainingPoints = burndown.RemainingOfTasks(tasks)
			list[i].Risk = burndown.Risk(list[i].RemainingPoints, list[i].DueDate, 2, now)
		}
		if list == nil {
			list = []model.Milestone{}
		}
		httpx.OK(c, list)
	}
}

func CreateMilestone(d *Deps) gin.HandlerFunc {
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
		var req milestoneReq
		if err := c.ShouldBindJSON(&req); err != nil {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		title, err := validate.RequiredString("title", req.Title, 1, 120)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		start, err := timeutil.ParseDate(req.StartDate)
		if err != nil {
			httpx.Fail(c, validate.Fail("start_date", "格式须为 yyyy-MM-dd"))
			return
		}
		due, err := timeutil.ParseDate(req.DueDate)
		if err != nil {
			httpx.Fail(c, validate.Fail("due_date", "格式须为 yyyy-MM-dd"))
			return
		}
		if due.Before(start) {
			httpx.Fail(c, validate.Fail("due_date", "不得早于 start_date"))
			return
		}
		bp, err := validate.NonNegInt("baseline_points", req.BaselinePoints, 10000)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		st := model.MSActive
		if req.Status != "" {
			st = model.MilestoneStatus(req.Status)
			if !st.Valid() {
				httpx.Fail(c, validate.Fail("status", "非法状态"))
				return
			}
		}
		now := timeutil.Now()
		m := &model.Milestone{
			ID: uuid.New(), ProjectID: pid, Title: title,
			StartDate: start, DueDate: due, BaselinePoints: bp, Status: st,
			CreatedAt: now, UpdatedAt: now,
		}
		if err := d.DB.CreateMilestone(c.Request.Context(), m); err != nil {
			httpx.Fail(c, err)
			return
		}
		httpx.Created(c, m)
	}
}

func GetMilestone(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		m, err := d.DB.MilestoneOwnedBy(c.Request.Context(), auth.UserID(c), id)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		tasks, _ := d.DB.ListTasksByMilestone(c.Request.Context(), m.ID)
		m.RemainingPoints = burndown.RemainingOfTasks(tasks)
		m.Risk = burndown.Risk(m.RemainingPoints, m.DueDate, 2, timeutil.Now())
		httpx.OK(c, m)
	}
}
