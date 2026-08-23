package handler

import (
	"github.com/gin-gonic/gin"
	"gopomodoro/internal/auth"
	"gopomodoro/internal/httpx"
	"gopomodoro/internal/ws"
)

func NewRouter(d *Deps, wsh *ws.Handler) *gin.Engine {
	if d.Cfg.Production() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(httpx.Recovery(), httpx.RequestID(), httpx.AccessLog(), httpx.CORS())

	r.GET("/health", func(c *gin.Context) { c.String(200, "ok") })
	r.GET("/api/v1/health", Health(d))

	v1 := r.Group("/api/v1")
	v1.POST("/auth/register", Register(d))
	v1.POST("/auth/login", Login(d))

	authed := v1.Group("")
	authed.Use(auth.Middleware(d.Cfg.JWTSecret))
	authed.GET("/me", Me(d))
	authed.GET("/projects", ListProjects(d))
	authed.POST("/projects", CreateProject(d))
	authed.GET("/projects/:id", GetProject(d))
	authed.PATCH("/projects/:id", UpdateProject(d))
	authed.GET("/projects/:id/milestones", ListMilestones(d))
	authed.POST("/projects/:id/milestones", CreateMilestone(d))
	authed.GET("/milestones/:id", GetMilestone(d))
	authed.GET("/milestones/:id/burndown", GetBurndown(d))
	authed.GET("/milestones/:id/metrics", GetMetrics(d))
	authed.GET("/projects/:id/tasks", ListTasks(d))
	authed.POST("/projects/:id/tasks", CreateTask(d))
	authed.PATCH("/tasks/:id", UpdateTask(d))
	authed.DELETE("/tasks/:id", DeleteTask(d))
	authed.POST("/tasks/reorder", ReorderTasks(d))
	authed.POST("/pomodoros", StartPomodoro(d))
	authed.GET("/pomodoros/active", ActivePomodoro(d))
	authed.GET("/pomodoros", ListSessions(d))
	authed.GET("/stats/daily", DailyStats(d))
	authed.GET("/pomodoros/:id", GetPomodoro(d))
	authed.POST("/pomodoros/:id/pause", PausePomodoro(d))
	authed.POST("/pomodoros/:id/resume", ResumePomodoro(d))
	authed.POST("/pomodoros/:id/abort", AbortPomodoro(d))
	authed.POST("/pomodoros/:id/test-complete", TestComplete(d))

	r.GET("/ws", auth.Middleware(d.Cfg.JWTSecret), wsh.Serve)
	return r
}
