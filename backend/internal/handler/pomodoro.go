package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopomodoro/internal/auth"
	"gopomodoro/internal/httpx"
	"gopomodoro/internal/pomodoro"
)

type startReq struct {
	TaskID string `json:"task_id"`
}

func StartPomodoro(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req startReq
		if err := c.ShouldBindJSON(&req); err != nil {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		tid, err := uuid.Parse(req.TaskID)
		if err != nil {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		if _, err := d.DB.TaskOwnedBy(c.Request.Context(), auth.UserID(c), tid); err != nil {
			httpx.Fail(c, err)
			return
		}
		view, err := d.Registry.Start(c.Request.Context(), auth.UserID(c), tid)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		httpx.Created(c, view)
	}
}

func command(d *Deps, cmd pomodoro.Command) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		reason := ""
		if cmd == pomodoro.CmdAbort {
			var body struct {
				Reason string `json:"reason"`
			}
			_ = c.ShouldBindJSON(&body)
			reason = body.Reason
		}
		view, err := d.Registry.Command(c.Request.Context(), auth.UserID(c), id, cmd, reason)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		httpx.OK(c, view)
	}
}

func PausePomodoro(d *Deps) gin.HandlerFunc  { return command(d, pomodoro.CmdPause) }
func ResumePomodoro(d *Deps) gin.HandlerFunc { return command(d, pomodoro.CmdResume) }
func AbortPomodoro(d *Deps) gin.HandlerFunc  { return command(d, pomodoro.CmdAbort) }

func ActivePomodoro(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		view := d.Registry.Active(auth.UserID(c))
		if view == nil {
			httpx.OK(c, gin.H{"session": nil})
			return
		}
		httpx.OK(c, view)
	}
}

func GetPomodoro(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		if view := d.Registry.Get(id); view != nil {
			if view.Session.UserID != auth.UserID(c) {
				httpx.Fail(c, httpx.ErrForbidden)
				return
			}
			httpx.OK(c, view)
			return
		}
		s, err := d.DB.SessionByID(c.Request.Context(), id)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		if s.UserID != auth.UserID(c) {
			httpx.Fail(c, httpx.ErrForbidden)
			return
		}
		httpx.OK(c, pomodoro.SessionView{Session: s, RemainingMS: s.RemainingMS(d.Registry.Clock.Now())})
	}
}

func TestComplete(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !d.Cfg.AllowTestComplete {
			httpx.Fail(c, httpx.ErrForbidden)
			return
		}
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		view, err := d.Registry.Command(c.Request.Context(), auth.UserID(c), id, pomodoro.CmdTick, "")
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		httpx.OK(c, view)
	}
}
