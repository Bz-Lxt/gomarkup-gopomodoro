# API · Mini Pomodoro Scrum

Base URL：`/api/v1`  
时区：全部时间字段为 GMT+8（`yyyy-MM-dd` / `yyyy-MM-dd HH:mm:ss`）  
鉴权：`Authorization: Bearer <token>`，WebSocket 可用 `?token=`

## 统一信封

```json
{ "ok": true, "data": {} }
{ "ok": false, "error": { "code": "E_VALIDATION", "message": "...", "details": {} } }
```

## 错误码表

| Code | HTTP | 含义 |
|---|---|---|
| E_VALIDATION | 400 | 入站字段缺失、类型或边界错误 |
| E_UNAUTHORIZED | 401 | 未登录或 JWT 无效 |
| E_FORBIDDEN | 403 | 越权或测试接口未开启 |
| E_NOT_FOUND | 404 | 资源不存在 |
| E_CONFLICT | 409 | 邮箱占用等通用冲突 |
| E_INVALID_TRANSITION | 409 | 非法番茄钟状态迁移 |
| E_SESSION_BUSY | 409 | 该用户已有 Running/Paused 会话 |
| E_OPTIMISTIC_LOCK | 409 | version 冲突，请重试 |
| E_INTERNAL | 500 | 未分类内部错误 |

---

## POST /auth/register

请求：

```json
{ "email": "dev@example.com", "password": "password1", "display_name": "Dev" }
```

响应 `201`：`{ "user": { "id": "...", "email": "..." }, "auth": { "token": "...", "expires_at": "2026-08-24 16:00:00" } }`

## POST /auth/login

请求：`{ "email": "geek@gopomodoro.dev", "password": "pomodoro123" }`  
响应同注册。

## GET /me

响应：当前用户（不含密码哈希）。

## GET /health 与 GET /api/v1/health

探针。后者返回 `{ "status":"ok", "time":"...", "live": 0, "timezone":"Asia/Shanghai" }`。

---

## 项目

- `GET /projects` 列表
- `POST /projects` `{ "name":"GoGo Gateway", "description":"..." }`
- `GET /projects/:id`
- `PATCH /projects/:id` `{ "name":"...", "archived": false }`

## 里程碑

- `GET /projects/:id/milestones`
- `POST /projects/:id/milestones`

```json
{
  "title": "8 月底完成 V1.0 核心网关开发",
  "start_date": "2026-08-01",
  "due_date": "2026-08-31",
  "baseline_points": 36
}
```

- `GET /milestones/:id`
- `GET /milestones/:id/burndown?granularity=day|week`
- `GET /milestones/:id/metrics`

燃尽响应示例：

```json
{
  "milestone_id": "...",
  "baseline_points": 36,
  "remaining_points": 22,
  "ideal": { "x": ["2026-08-01"], "y": [36] },
  "actual": { "x": ["2026-08-01 10:00:00"], "y": [36] },
  "scope_marks": [{ "at": "2026-08-16 10:00:00", "remaining": 24 }]
}
```

## 任务

- `GET /projects/:id/tasks?milestone_id=`
- `POST /projects/:id/tasks` `{ "title":"限流", "estimated_pomodoros": 4, "kanban_column":"todo", "milestone_id":"..." }`
- `PATCH /tasks/:id`
- `DELETE /tasks/:id`
- `POST /tasks/reorder` `{ "items": [{ "id":"...", "kanban_column":"done", "sort_order": 0 }] }`

`kanban_column` ∈ `backlog|todo|in_progress|done`

## 番茄钟

- `POST /pomodoros` `{ "task_id": "..." }` → 创建并 `start`
- `GET /pomodoros/active`
- `GET /pomodoros/:id`
- `POST /pomodoros/:id/pause`
- `POST /pomodoros/:id/resume`
- `POST /pomodoros/:id/abort` `{ "reason": "user" }`
- `POST /pomodoros/:id/test-complete` **仅** `ALLOW_TEST_COMPLETE=true` 时可用，模拟服务端 tick

会话视图：

```json
{
  "session": {
    "id": "...",
    "state": "running",
    "focus_duration_ms": 1500000,
    "resume_token": "ab12...",
    "version": 2
  },
  "remaining_ms": 1492000,
  "grace_left_s": 0
}
```

非法迁移示例：对 `idle` 调 `test-complete` → `409 E_INVALID_TRANSITION`。

## WebSocket `/ws?token=`

入站：`hello`（可带 `resume_token`）、`subscribe`（`milestone_id`）、`ping`  
出站：`session.state`、`session.tick`、`burndown.update`、`grace`、`pong`、`error`  
心跳：服务端每 15s ping，45s 无 pong 断开；断连后 120s 宽限期，超时 `Aborted(network_timeout)`。
