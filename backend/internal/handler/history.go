package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gopomodoro/internal/auth"
	"gopomodoro/internal/httpx"
	"gopomodoro/internal/store"
	"gopomodoro/internal/timeutil"
)

func ListSessions(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		from, _ := timeutil.ParseDate(c.Query("from"))
		to, _ := timeutil.ParseDate(c.Query("to"))
		list, err := d.DB.ListSessions(c.Request.Context(), store.SessionFilter{
			UserID: auth.UserID(c),
			State:  c.Query("state"),
			From:   from,
			To:     to,
			Limit:  limit,
		})
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		httpx.OK(c, list)
	}
}

func DailyStats(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := timeutil.Now()
		from := timeutil.StartOfDay(now).AddDate(0, 0, -13)
		rows, err := d.DB.DailyCompleted(c.Request.Context(), auth.UserID(c), from, now.Add(24*0))
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		httpx.OK(c, rows)
	}
}
