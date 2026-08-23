# QA Record

## Round 1 · 2026-08-23 16:40 (GMT+8)

**Cost**: ¥0（无计费 API，全程 Mock/离线）

### 执行环境

- 主栈：`docker compose up --build -d`（frontend :35172 / backend :35173 / postgres :35174）
- 单元测试：`cd backend && go test ./...`（宿主机与镜像同源）
- API 冒烟：`API_BASE=http://127.0.0.1:35173/api/v1 python3 tests/api_smoke.py`
- 浏览器手跑：登录 → 看板 → 燃尽指标 → 番茄钟 start（FOCUS LOCKED 24:49）

### 结果

| 检查 | 结果 | 备注 |
|---|---|---|
| Docker Build | PASS | backend golang:1.25-alpine；frontend 多阶段 nginx |
| Health Check | PASS | `/health` 返回 `ok`；`/api/v1/health` 带 GMT+8 |
| Auth + Seed | PASS | 1 项目 / 2 里程碑 / 13 任务 |
| Pomodoro start | PASS | remaining_ms ≈ 1500000 |
| E_SESSION_BUSY | PASS | 409 |
| E_INVALID_TRANSITION | PASS | aborted 后 resume → 409 |
| Burndown | PASS | ideal/actual 双序列 |
| go test ./... | PASS | 状态机穷尽矩阵 6 合法 / 19 非法 |
| 浏览器看板 | PASS | 四列 + 🍅 消耗/估点 |
| 浏览器燃尽 | PASS | 今日 0 / 本周 5 / 废弃率 55% / 预测 2026-09-12 |
| 浏览器番茄钟 | PASS | 开始后 FOCUS LOCKED |
| Playwright 容器 | PENDING | qa 镜像拉取 Playwright 约 700MB，本轮未完成 |

### 缺陷

1. **P0 死锁（已修）**：`sweepGrace` 持 `Registry.mu` 再锁 session，`Command.drop` 持 session 再锁 `Registry.mu`，WS 断连后 HTTP 经 nginx 出现 504。已改为先拷贝 live 切片再释放 registry 锁；终态 `drop` 在释放 session 锁之后执行。待 Round 2 验证 rebuild 后 abort 不再挂起。

## Round 2 · 2026-08-23 16:55 (GMT+8)

**Cost**: ¥0

对 Round 1 缺陷 #1 只验证既定方案，未更换修复方向。热替换 linux/arm64 二进制并 `docker restart backend` 后：

```
start running remain 1499545
pause paused
abort aborted
active {'session': None}
```

前端反代 `/api/v1/health` 恢复 200。Playwright 容器因镜像体积在本轮机器高负载下超时，E2E 规格已落盘 `tests/e2e_flow.spec.ts`，关键路径由浏览器手跑 + API 冒烟覆盖。

`[PASS] 死锁修复后 start/pause/abort`
`[PASS] Frontend proxy health`
