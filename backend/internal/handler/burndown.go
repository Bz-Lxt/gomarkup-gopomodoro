package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopomodoro/internal/auth"
	"gopomodoro/internal/burndown"
	"gopomodoro/internal/httpx"
	"gopomodoro/internal/report"
	"gopomodoro/internal/timeutil"
	"gopomodoro/internal/validate"
)

func GetBurndown(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		ms, err := d.DB.MilestoneOwnedBy(c.Request.Context(), auth.UserID(c), id)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		gran, err := validate.Granularity(c.DefaultQuery("granularity", "day"))
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		chart, err := d.Engine.Chart(c.Request.Context(), ms, gran)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		httpx.OK(c, chart)
	}
}

func GetMetrics(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			httpx.Fail(c, httpx.ErrValidation)
			return
		}
		ms, err := d.DB.MilestoneOwnedBy(c.Request.Context(), auth.UserID(c), id)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		m, err := d.Engine.Metrics(c.Request.Context(), auth.UserID(c), id)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		tasks, _ := d.DB.ListTasksByMilestone(c.Request.Context(), id)
		remain := burndown.RemainingOfTasks(tasks)
		fc := report.Triple(remain, m.AvgDailyVelocity, ms.DueDate, timeutil.Now())
		httpx.OK(c, gin.H{
			"today_completed":    m.TodayCompleted,
			"week_completed":     m.WeekCompleted,
			"today_aborted":      m.TodayAborted,
			"week_aborted":       m.WeekAborted,
			"abort_rate":         m.AbortRate,
			"avg_daily_velocity": m.AvgDailyVelocity,
			"predicted_done_on":  m.PredictedDoneOn,
			"forecast":           fc,
			"commentary":         report.Commentary(fc, m.AbortRate, remain),
		})
	}
}
