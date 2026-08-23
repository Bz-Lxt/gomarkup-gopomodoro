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

type projectReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Archived    *bool  `json:"archived"`
}

func ListProjects(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		all := c.Query("archived") == "true"
		list, err := d.DB.ListProjects(c.Request.Context(), auth.UserID(c), all)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		if list == nil {
			list = []model.Project{}
		}
		httpx.OK(c, list)
	}
}

func CreateProject(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req projectReq
		if err := c.ShouldBindJSON(&req); err != nil {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		name, err := validate.RequiredString("name", req.Name, 1, 80)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		now := timeutil.Now()
		p := &model.Project{
			ID: uuid.New(), UserID: auth.UserID(c), Name: name,
			Description: req.Description, CreatedAt: now, UpdatedAt: now,
		}
		if err := d.DB.CreateProject(c.Request.Context(), p); err != nil {
			httpx.Fail(c, err)
			return
		}
		httpx.Created(c, p)
	}
}

func UpdateProject(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		p, err := d.DB.ProjectByID(c.Request.Context(), auth.UserID(c), id)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		var req projectReq
		if err := c.ShouldBindJSON(&req); err != nil {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		if req.Name != "" {
			name, err := validate.RequiredString("name", req.Name, 1, 80)
			if err != nil {
				httpx.Fail(c, err)
				return
			}
			p.Name = name
		}
		if req.Description != "" {
			p.Description = req.Description
		}
		if req.Archived != nil {
			p.Archived = *req.Archived
		}
		p.UpdatedAt = timeutil.Now()
		if err := d.DB.UpdateProject(c.Request.Context(), p); err != nil {
			httpx.Fail(c, err)
			return
		}
		httpx.OK(c, p)
	}
}

func GetProject(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		p, err := d.DB.ProjectByID(c.Request.Context(), auth.UserID(c), id)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		httpx.OK(c, p)
	}
}
