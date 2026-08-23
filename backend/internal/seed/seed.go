package seed

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gopomodoro/internal/auth"
	"gopomodoro/internal/burndown"
	"gopomodoro/internal/logger"
	"gopomodoro/internal/model"
	"gopomodoro/internal/store"
	"gopomodoro/internal/timeutil"
)

const (
	DemoEmail    = "geek@gopomodoro.dev"
	DemoPassword = "pomodoro123"
	DemoName     = "极客番茄"
)

func Run(ctx context.Context, db *store.DB) error {
	n, err := db.CountUsers(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		logger.L().Info("seed skipped, users exist")
		return nil
	}
	hash, err := auth.HashPassword(DemoPassword)
	if err != nil {
		return err
	}
	now := timeutil.Now()
	user := &model.User{
		ID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Email: DemoEmail, PasswordHash: hash, DisplayName: DemoName,
		Timezone: "Asia/Shanghai", CreatedAt: now,
	}
	if err := db.CreateUser(ctx, user); err != nil {
		return err
	}
	proj := &model.Project{
		ID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		UserID: user.ID, Name: "GoGo Gateway",
		Description: "独立开发者的 V1.0 核心网关与观测里程碑",
		CreatedAt: now, UpdatedAt: now,
	}
	if err := db.CreateProject(ctx, proj); err != nil {
		return err
	}

	ms1 := &model.Milestone{
		ID: uuid.MustParse("33333333-3333-3333-3333-333333333331"),
		ProjectID: proj.ID, Title: "8 月底完成 V1.0 核心网关开发",
		StartDate: timeutil.Date(2026, 8, 1, 0, 0, 0),
		DueDate:   timeutil.Date(2026, 8, 31, 0, 0, 0),
		BaselinePoints: 36, Status: model.MSActive, CreatedAt: now, UpdatedAt: now,
	}
	ms2 := &model.Milestone{
		ID: uuid.MustParse("33333333-3333-3333-3333-333333333332"),
		ProjectID: proj.ID, Title: "9 月中完成观测与稳定性",
		StartDate: timeutil.Date(2026, 8, 10, 0, 0, 0),
		DueDate:   timeutil.Date(2026, 9, 15, 0, 0, 0),
		BaselinePoints: 22, Status: model.MSActive, CreatedAt: now, UpdatedAt: now,
	}
	if err := db.CreateMilestone(ctx, ms1); err != nil {
		return err
	}
	if err := db.CreateMilestone(ctx, ms2); err != nil {
		return err
	}

	type spec struct {
		title string
		est   int
		cons  int
		col   model.KanbanColumn
		ms    *model.Milestone
		order int
	}
	specs := []spec{
		{"接入层路由与 TLS 终止", 5, 3, model.ColInProgress, ms1, 0},
		{"JWT 鉴权中间件", 3, 3, model.ColDone, ms1, 1},
		{"限流与熔断", 4, 1, model.ColTodo, ms1, 2},
		{"配置热更新", 3, 0, model.ColTodo, ms1, 3},
		{"gRPC 协议转换", 5, 0, model.ColBacklog, ms1, 4},
		{"上游健康检查", 3, 2, model.ColInProgress, ms1, 5},
		{"请求日志采样", 2, 2, model.ColDone, ms1, 6},
		{"OpenTelemetry 埋点", 4, 1, model.ColTodo, ms2, 0},
		{"Prometheus 指标导出", 3, 0, model.ColTodo, ms2, 1},
		{"告警规则与值班手册", 3, 0, model.ColBacklog, ms2, 2},
		{"混沌演练脚本", 4, 0, model.ColBacklog, ms2, 3},
		{"压测基线与容量报告", 4, 0, model.ColBacklog, ms2, 4},
		{"文档站点与示例", 2, 0, model.ColBacklog, ms2, 5},
	}
	created := make([]model.Task, 0, len(specs))
	for i, s := range specs {
		mid := s.ms.ID
		t := &model.Task{
			ID: uuid.New(), ProjectID: proj.ID, MilestoneID: &mid,
			Title: s.title, EstimatedPomodoros: s.est, ConsumedPomodoros: s.cons,
			KanbanColumn: s.col, SortOrder: s.order, CreatedAt: now, UpdatedAt: now,
		}
		if err := db.CreateTask(ctx, t); err != nil {
			return err
		}
		created = append(created, *t)
		_ = i
	}
	if err := insertHistory(ctx, db, user.ID, created); err != nil {
		return err
	}

	// Historical burndown for ms1: remaining declines from 36 toward current.
	tasks1, err := db.ListTasksByMilestone(ctx, ms1.ID)
	if err != nil {
		return err
	}
	current := burndown.RemainingOfTasks(tasks1)
	days := []struct {
		offset int
		remain int
		et     model.BurndownEventType
	}{
		{0, 36, model.EventSnapshot},
		{3, 33, model.EventPomodoroCompleted},
		{6, 30, model.EventPomodoroCompleted},
		{9, 28, model.EventTaskDone},
		{12, 26, model.EventPomodoroCompleted},
		{15, 24, model.EventScopeChange},
		{18, 22, model.EventPomodoroCompleted},
		{21, current, model.EventSnapshot},
	}
	for i, d := range days {
		at := timeutil.Date(2026, 8, 1, 10, 0, 0).Add(time.Duration(d.offset) * 24 * time.Hour)
		if at.After(now) {
			at = now.Add(-time.Duration(len(days)-i) * time.Hour)
		}
		pt := &model.BurndownPoint{
			ID: uuid.New(), MilestoneID: ms1.ID, RecordedAt: at,
			RemainingPoints: d.remain,
			IdealPoints:     burndown.IdealAt(ms1.BaselinePoints, ms1.StartDate, ms1.DueDate, at),
			EventType:       d.et,
			EventID:         "seed-ms1-" + uuid.NewString(),
		}
		if err := db.InsertBurndownPoint(ctx, pt); err != nil {
			return err
		}
	}
	tasks2, _ := db.ListTasksByMilestone(ctx, ms2.ID)
	pt2 := &model.BurndownPoint{
		ID: uuid.New(), MilestoneID: ms2.ID, RecordedAt: ms2.StartDate,
		RemainingPoints: burndown.RemainingOfTasks(tasks2),
		IdealPoints:     float64(ms2.BaselinePoints),
		EventType:       model.EventSnapshot,
		EventID:         "seed-ms2-start",
	}
	if err := db.InsertBurndownPoint(ctx, pt2); err != nil {
		return err
	}
	logger.L().Info("seed inserted", "user", DemoEmail, "tasks", len(specs))
	return nil
}
