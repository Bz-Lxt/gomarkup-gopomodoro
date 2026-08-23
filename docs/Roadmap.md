# Roadmap · Mini Pomodoro Scrum

> **文档权威性**：本文件定义 **WHEN**（分期与构建顺序）。`docs/Requirements.md` 定义 **WHAT**。
> **规模**：Tier 2（预估 10k–12.5k LoC）→ MVP / V1 / V2 边界强制。
> **构建顺序裁定**：Logic-First（交换 SOP Phase 2 与 Phase 3）。

---

## Phase Order Decision

**裁定：Logic-First。**

燃尽图大屏的坐标系、理想线公式、真实线跳变与 `scope_change` 打点全部由领域模型派生；沉浸式番茄钟的倒计时权威、宽限期文案与可点按钮集合由五态状态机派生。先画 UI 只能对着臆想协议空转。因此先冻结后端契约（状态机、事件、WS 协议、REST），再让前端按契约渲染。

---

## 目录结构（冻结）

```
GoPomodoro/
├── backend/                 # Go 模块化单体
├── frontend-user/           # Vue 3 用户端（唯一前端）
├── tests/                   # API smoke + Playwright E2E
├── docs/
├── docker-compose.yml
└── README.md                # /deploy 阶段补齐七段式
```

不交付 `frontend-admin` / `frontend-mp`：本产品无管理后台与微信小程序需求。

---

## MVP（本轮 /auto 必须交付）

| ID | 任务 | 负责 | 验收 |
|---|---|---|---|
| M1 | Git 初始化 + `.gitignore` + Compose 随机端口 | Architect | `git status` 干净可跟踪 |
| M2 | PostgreSQL 显式迁移 + 种子数据 | Logic | ≥1 项目 / 2 里程碑 / 12 任务 / 历史燃尽点 |
| M3 | 用户注册登录 JWT | Logic | 种子账号可登录 |
| M4 | 项目 / 里程碑 / 任务 CRUD + 四列看板拖拽 API | Logic | 乐观锁/顺序更新 |
| M5 | 番茄钟五态状态机 + 会话级锁 + version 乐观锁 | Logic | 非法迁移 100% 拒绝 |
| M6 | 内存 Registry + 启动重建 + 120s 宽限期 | Logic | 重启后 Running/Paused 100% 恢复 |
| M7 | WebSocket 心跳 / 续连 / 倒计时推送 | Logic | 刷新误差 ≤ 1s |
| M8 | 燃尽事件总线 + 幂等聚合 + 时序落库 | Logic | P95 ≤ 500ms，重复投递不双扣 |
| M9 | REST 接线 + 统一错误码 + 结构化日志 | Logic | `docs/API.md` |
| M10 | Docker 多阶段 + 反向代理一键启动 | Logic | `localhost:35172` 可访问 |

## V1（本轮一并交付，属于必交付）

| ID | 任务 | 负责 | 验收 |
|---|---|---|---|
| V1-1 | 看板 / 番茄钟 / 燃尽三大页面 | UI | 深色极客主题，四态齐全 |
| V1-2 | 白噪音 ≥3 种（Web Audio 本地合成）+ 音量 | UI | 无外部音频依赖 |
| V1-3 | 启动/暂停/完成/放弃独立动效 | UI | 完成有庆祝反馈 |
| V1-4 | 效率指标卡（今日/本周/废弃率/预测完成日） | Logic+UI | 数据来自聚合引擎 |
| V1-5 | Playwright E2E 全链路 + API smoke | QA | Mock/离线，Cost ¥0 |
| V1-6 | API 文档含示例与错误码表 | Logic | `docs/API.md` |

## V2（明确不交付，仅预留扩展点）

- 团队多人协作与权限
- 自定义看板列
- Redis Pub/Sub 横向扩展
- 数据导出与周报

扩展点：`eventbus` 接口可替换为 Redis；`kanban_column` 为枚举而非自由字符串。

---

## 端口分配（Dev · 已探测空闲）

| 服务 | 主机端口 | 容器端口 |
|---|---|---|
| frontend-user (nginx) | **35172** | 80 |
| backend (gin) | **35173** | 8080 |
| postgres | **35174** | 5432 |

`/deploy` 阶段再改为 8081+。

---

## 模块划分（后端）

```
cmd/server            入口、优雅退出、Registry 重建
internal/config       环境配置
internal/logger       slog 统一日志
internal/timeutil     GMT+8
internal/httpx        统一响应 / 错误码 / 中间件
internal/auth         JWT / 密码 / 鉴权
internal/model        领域实体
internal/store        PostgreSQL 仓储 + 迁移锁
internal/pomodoro     状态机 / Registry / 宽限期 / Clock
internal/burndown     理想线 / 剩余点 / 指标
internal/eventbus     进程内事件 + worker
internal/ws           Hub / 协议 / Hijacker 安全
internal/handler      HTTP 适配
internal/seed         幂等种子
internal/validate     入站校验
```

---

## 进度

- [x] Phase 1 Architect
- [x] Phase 3 Logic（因 Logic-First 提前）
- [x] Phase 2 UI
- [x] Phase 4 QA
- [x] Phase 5 Audit + `/learn`
