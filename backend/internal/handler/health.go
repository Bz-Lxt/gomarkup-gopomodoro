package handler

import (
	"github.com/gin-gonic/gin"
	"gopomodoro/internal/httpx"
	"gopomodoro/internal/timeutil"
)

func Health(d *Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := d.DB.SQL.PingContext(c.Request.Context()); err != nil {
			httpx.Fail(c, httpx.ErrInternal)
			return
		}
		httpx.OK(c, gin.H{
			"status":    "ok",
			"time":      timeutil.FormatDateTime(timeutil.Now()),
			"live":      d.Registry.LiveCount(),
			"timezone":  "Asia/Shanghai",
		})
	}
}

func Livez(c *gin.Context) {
	c.String(200, "ok")
}
