# Requirements · Mini Pomodoro Scrum

> **文档权威性**：本文件定义 **WHAT**（做什么、验收线）。`docs/Roadmap.md` 定义 **WHEN**（分期顺序）。
> **状态**：`FROZEN`（需求已冻结）
> **冻结时间**：2026-08-23 15:54 (GMT+8)
> **规模判定**：Tier 2（10k–40k LoC）→ 强制分期路线图
> **时区基准**：全系统统一 `Asia/Shanghai` (GMT+8)，容器与数据库均需显式设置 `TZ`

---

## 1. 业务目标（一句话）

把「待办任务」量化为**番茄钟点数**，通过真实专注结算驱动里程碑燃尽图实时波动，让独立开发者/小团队用可度量的数据看清自己的产能。

**核心价值链**：`估点 → 专注 → 结算 → 聚合 → 燃尽可视化 → 效率复盘`

---

## 2. 目标用户与场景

| 角色 | 场景 |
|---|---|
| 独立开发者 | 单人给自己排 V1.0 里程碑，每天靠番茄钟推进，看燃尽线判断是否会延期 |
| 自由职业者 | 多项目并行，用番茄钟数据量化投入产出，作为报价依据 |
| 小团队（≤10 人） | 共享看板与里程碑，团队级燃尽图 |

**MVP 范围**：单用户 + 多项目。团队协作列入 V1。

---

## 3. 核心领域模型与统一口径（裁定项）

### 3.1 计量单位（C4 裁定）

- **1 Point（点数）= 1 个成功结算（Completed）的番茄钟**
- 默认番茄钟时长 **25 分钟**（`focus_duration` 可配置，默认 1500s）
- 任务的 `estimated_pomodoros` 即该任务的**估点**
- 里程碑总点数 = 其下所有关联任务估点之和

### 3.2 实体清单

| 实体 | 关键字段 | 约束 |
|---|---|---|
| `User` | id, email, password_hash, timezone | MVP 单用户可注册登录 |
| `Project` | id, user_id, name, description, archived | 任务与里程碑的容器 |
| `Milestone` | id, project_id, title, **start_date**, **due_date**, **baseline_points**, status | C3：三字段缺一不可，`baseline_points` 为创建/启动时冻结的基线 |
| `Task` | id, project_id, milestone_id(nullable), title, `estimated_pomodoros`, `consumed_pomodoros`, kanban_column, order, status | 卡片上显示 `consumed / estimated 🍅` |
| `PomodoroSession` | id, task_id, state, **started_at**, `paused_accumulated_ms`, `expected_end_at`, ended_at, abort_reason | 状态机载体 |
| `BurndownPoint` | id, milestone_id, `recorded_at`, remaining_points, ideal_points, event_type | 时序表，燃尽图数据源 |
| `ScopeChangeLog` | id, milestone_id, delta_points, reason, occurred_at | C7：范围变更打点 |

### 3.3 看板列（Kanban Column）

固定四列：`Backlog` → `Todo` → `In Progress` → `Done`（MVP 不支持自定义列，V2 再开放）

---

## 4. 番茄钟状态机（硬性规格）

### 4.1 状态与合法迁移表

```
        ┌──────── start ────────┐
        │                       ▼
     [Idle] ──────────────► [Running] ◄──── resume ──── [Paused]
                              │  │  │                      │
                              │  │  └──── pause ───────────┘
                     timeout  │  └──── abort ──────► [Aborted]
                              ▼                          ▲
                        [Completed]         grace_expired │
                                                          │
                                            [Paused] ─ abort ─┘
```

**合法迁移矩阵**（任何未列出的迁移必须拒绝并返回 `409 Conflict` + 错误码 `E_INVALID_TRANSITION`）：

| From | To | 触发 |
|---|---|---|
| Idle | Running | `start` |
| Running | Paused | `pause` |
| Running | Completed | 倒计时自然归零（**仅服务端可触发**） |
| Running | Aborted | `abort` / 宽限期超时 |
| Paused | Running | `resume` |
| Paused | Aborted | `abort` / 宽限期超时 |
| Completed | — | 终态，不可变更 |
| Aborted | — | 终态，不可变更 |

**明确禁止**：`Idle → Completed`、`Completed → Running`、`Aborted → Running`、`Paused → Completed`。

### 4.2 关键裁定

- **C1（计时权威）**：客户端倒计时**仅为展示**。唯一权威是服务端 `started_at` + `focus_duration - paused_accumulated_ms`。内存 Registry 仅作热态加速，进程重启后必须能从数据库时间戳重建所有 Running/Paused 会话。
- **C2（异常流产）**：WebSocket 断开 **不等于** 放弃。断连进入 `grace period = 120s`：
  - 120s 内重连（携带 session resume token）→ 恢复原状态，倒计时连续；
  - 超过 120s 未重连 → 服务端将会话置为 `Aborted`，`abort_reason = network_timeout`；
  - 服务端计时器为权威，宽限期判定由后台巡检 goroutine 执行。
- **单活跃会话约束**：同一 User 同时最多存在 1 个 `Running`/`Paused` 会话。发起第二个必须返回 `409` + `E_SESSION_BUSY`。
- **并发安全**：状态迁移必须在会话级互斥锁内完成，且数据库层用乐观锁（`version` 字段）或条件更新兜底，保证 WS 与 HTTP 双通道并发下不出现双写。
- **C5（Aborted 不计入燃尽）**：只有 `Completed` 的番茄钟才 `consumed_pomodoros += 1` 并触发燃尽扣减。`Aborted` 仅记入效率指标（专注废弃率）。

---

## 5. 燃尽图聚合引擎

### 5.1 理想燃尽线（C3）

对里程碑 `[start_date, due_date]` 区间，按自然日线性插值：

```
ideal(d) = baseline_points × (1 - elapsed_days(d) / total_days)
```

- `total_days = due_date - start_date`（至少 1 天）
- 端点：`ideal(start_date) = baseline_points`，`ideal(due_date) = 0`

### 5.2 真实燃尽线

```
remaining(t) = Σ(estimated_pomodoros) - Σ(已完成任务的 estimated) - Σ(未完成任务已消耗且不超估的 consumed)
```

**扣减规则**：
- 番茄钟 `Completed` → 该任务剩余点数 `-1`（下限 0，超估不产生负值）
- 任务标记 `Done` → 该任务剩余点数直接归零（无论实际消耗多少）
- 任务被删除/解绑里程碑 → 记 `ScopeChangeLog`，剩余点数按 delta 调整

### 5.3 C7（范围变更）

中途新增任务或上调估点 → 真实燃尽线**允许上翘**，并在该时间点写入 `event_type = scope_change` 的燃尽点，前端 ECharts 以 `markPoint` 标注，不做数据平滑掩盖。

### 5.4 事件驱动要求

- 结算动作（番茄钟完成 / 任务完成 / 范围变更）必须**异步**触发领域事件，不阻塞 HTTP 响应
- 事件消费者重算里程碑剩余点数，写入 `burndown_points`（含 GMT+8 时间戳）
- **幂等性**：同一事件 ID 重复投递不得产生重复扣减
- 聚合结果通过 WebSocket 广播给订阅该里程碑的客户端

---

## 6. 前端功能需求

### 6.1 敏捷看板与里程碑视图（主页面）

- 左侧栏：里程碑列表，每项显示 `标题 / 截止日 / 剩余点数 / 进度条 / 延期风险标识`
- 中部：四列 Kanban，支持拖拽改列与排序（乐观更新 + 失败回滚）
- 卡片：标题、`3/5 🍅` 形式的估点与消耗、所属里程碑色标、一键「开始专注」
- 点击里程碑可筛选看板卡片

### 6.2 极客沉浸式番茄钟

- 全屏沉浸模式：大号倒计时数字 + 环形进度
- 控制：开始 / 暂停 / 继续 / 放弃（放弃需二次确认）
- 白噪音：≥3 种本地音频（雨声 / 咖啡馆 / 白噪），可切换、可调音量、静音
- 动效：启动、暂停、完成、放弃各有独立过渡动画；完成时有明确庆祝反馈
- **刷新页面后倒计时必须无缝续接**（误差 ≤ 1s）
- 断连时 UI 明确显示「连接中断，剩余 XXs 内重连可恢复」

### 6.3 动态燃尽图大屏

- ECharts 双折线：理想线（虚线/低饱和）+ 真实线（实线/高亮）
- 支持切换里程碑、时间粒度（日/周）
- WebSocket 推送后**增量更新**曲线，无整体重绘闪烁
- 附属指标卡：今日番茄数、本周番茄数、平均每日产能、专注废弃率、预测完成日

### 6.4 视觉标准（Redline 2）

- 深色「极客」主题为主，Tailwind 设计系统，等宽字体用于倒计时数字
- 响应式：≥1280px 桌面为主战场，≥768px 平板可用，移动端至少不破版
- 完整反馈态：Loading / Empty / Error / Success 四态齐全，禁止「工程师 UI」

---

## 7. 推荐技术栈（最终由 Chief Architect 确认）

| 层 | 选型 | 理由 |
|---|---|---|
| 后端 | **Go 1.23** + Gin + gorilla/websocket | 用户指定 Go；WS 生态成熟 |
| 数据库 | **PostgreSQL 16** | 时序燃尽数据 + 事务一致性 |
| 迁移 | 显式 SQL 迁移（goose/golang-migrate） | 禁止 AutoMigrate 裸奔 |
| 前端 | **Vue 3 + TypeScript**（C6 裁定） | 用户给出二选一，Composition API 与 ECharts 状态同步更契合 |
| 样式 | TailwindCSS | 用户指定 |
| 图表 | ECharts 5 | 用户指定 |
| 事件总线 | Go channel + worker pool（进程内） | 单实例交付足够；跨实例 Redis Pub/Sub 列入 V2 |
| 容器 | Docker Compose（多阶段构建，ARM64 + AMD64） | Redline 1 |

**明确排除（防范围漂移）**：不做支付、不做第三方登录、不做移动 App、不做 AI 建议、不做多租户 SaaS 计费。

---

## 8. 验收基线（可度量，非散文）

### 8.1 硬性门槛

| # | 指标 | 阈值 |
|---|---|---|
| A1 | `docker compose up --build -d` 一键启动 | 零手动步骤，前端可从 `localhost` 访问 |
| A2 | 跨平台镜像 | `linux/arm64` 与 `linux/amd64` 均可 pull/build |
| A3 | 数据库迁移 | 首次启动自动完成，含种子数据（≥1 项目 / 2 里程碑 / 12 任务 / 若干历史燃尽点） |

### 8.2 状态机与实时性

| # | 指标 | 阈值 |
|---|---|---|
| B1 | 非法状态迁移拒绝率 | **100%**，返回 `409` + `E_INVALID_TRANSITION` |
| B2 | 刷新页面倒计时误差 | **≤ 1s** |
| B3 | 服务重启后 Running/Paused 会话恢复率 | **100%** |
| B4 | WS 心跳间隔 / 超时判定 | 15s ping，45s 无 pong 判定断连 |
| B5 | 断连宽限期 | 120s，超期自动 `Aborted(network_timeout)` |
| B6 | 并发活跃会话 | **≥ 500** 个 Running 会话稳定运行，无 goroutine 泄漏、无 data race |
| B7 | `go test -race` | **零 race 报告** |

### 8.3 聚合引擎

| # | 指标 | 阈值 |
|---|---|---|
| C1 | 番茄钟结算 → 燃尽点落库延迟 | **P95 ≤ 500ms** |
| C2 | 事件重复投递 | 幂等，不产生重复扣减 |
| C3 | 燃尽计算正确性 | 单元测试覆盖超估、提前完成、范围变更、跨天 4 类边界 |

### 8.4 接口与前端

| # | 指标 | 阈值 |
|---|---|---|
| D1 | REST API P95 响应 | **< 200ms**（本地 Docker，种子数据规模） |
| D2 | 前端首屏可交互 | **< 2s** |
| D3 | WS 推送到 UI 更新 | **< 300ms** |

### 8.5 工程质量（对齐 global.md）

| # | 指标 | 阈值 |
|---|---|---|
| E1 | 统一 Logger | 全项目禁止裸 `fmt.Println` / `console.log`，须有 level 控制，生产屏蔽 debug |
| E2 | 反序列化校验 | 所有外部输入（HTTP body / WS 消息 / 配置）必须校验字段存在性、类型、边界，不依赖调用处 |
| E3 | API 文档 | 独立 `docs/API.md`，含每端点请求/响应示例 + 参数类型 + **错误码表** |
| E4 | 后端测试覆盖 | 整体 **≥ 60%**；**状态机模块 ≥ 90%**；聚合引擎 ≥ 80% |
| E5 | E2E | Playwright 覆盖：创建里程碑→建任务→跑番茄钟→刷新续接→完成结算→燃尽线更新 |
| E6 | Go 文件数 / 代码量 | **≥ 32 个 Go 文件**，后端 6000–8500 行（不含生成代码与 vendor） |
| E7 | 时区 | 所有落库与展示时间为 GMT+8，容器设置 `TZ=Asia/Shanghai` |

---

## 9. 冲突裁定汇总（PM 冻结，Phase 2–5 不得推翻）

| ID | 冲突 | 裁定 |
|---|---|---|
| C1 | 内存计时 vs 刷新不丢失 | 服务端时间戳为权威，内存仅热态缓存，可从库重建 |
| C2 | 断网即流产 vs 刷新也断 WS | 120s 宽限期 + resume token；超期才 Aborted |
| C3 | 理想线缺起点 | 里程碑强制 `start_date` + `due_date` + `baseline_points` |
| C4 | 「番茄数」vs「点数」 | 1 点 = 1 个 Completed 番茄钟 |
| C5 | Aborted 是否计燃尽 | 不计燃尽，只计废弃率 |
| C6 | Vue vs React | **Vue 3 + TypeScript** |
| C7 | 中途加任务导致跳变 | 允许真实线上翘 + `scope_change` 打点标注 |

---

## 10. 分期边界（详细排期见 Roadmap）

| 阶段 | 内容 | 是否 MVP 必交付 |
|---|---|---|
| **MVP** | 用户认证、项目/里程碑/任务 CRUD、四列看板拖拽、番茄钟五态状态机、WS 心跳与刷新续接、燃尽事件引擎、双线燃尽图、Docker 一键启动 | ✅ 必须 |
| **V1** | 白噪音多音源与音量、沉浸式动效打磨、效率指标卡（废弃率/预测完成日）、E2E 全链路、API 文档 | ✅ 必须 |
| **V2** | 团队多人协作、自定义看板列、Redis Pub/Sub 横向扩展、数据导出、周报 | ❌ 不交付（仅预留扩展点） |

---

## 11. 未采用项与理由（防漂移记录）

- **Redis 强依赖**：单实例交付不需要，进程内事件总线足够，避免无谓运维复杂度。
- **微服务拆分**：10k LoC 规模用模块化单体（Modular Monolith）更合理。
- **移动端原生 App**：用户未要求，响应式 Web 已覆盖。
