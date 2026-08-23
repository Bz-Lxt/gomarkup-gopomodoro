# Design Spec · Mini Pomodoro Scrum

## 美学方向

**「终端任务控制室」**：深墨绿底、酸性黄绿描边、等宽倒计时。不是通用紫色仪表盘，而是独立开发者夜里赶里程碑时的座舱。

可被记住的一件事：全屏番茄钟上那组 IBM Plex Mono 的巨大剩余秒数，外圈酸性进度环随状态变色。

## 色板

| Token | Hex | 用途 |
|---|---|---|
| bg | `#070b09` | 页面底 |
| panel | `#101714` | 卡片/列 |
| line | `#24352c` | 描边 |
| acid | `#c6ff3d` | 主强调、进行中 |
| cyan | `#3ee0c4` | 链接、理想线 |
| amber | `#ffb020` | 暂停、风险 |
| rose | `#ff5d73` | 放弃、逾期 |
| fog | `#8aa396` | 次级文字 |
| paper | `#e8f5e4` | 主文字 |

## 字体

- Display / 倒计时：`IBM Plex Mono`
- UI：`Syne`（标题）+ `IBM Plex Sans`（正文）
- 禁止 Inter / Roboto / Space Grotesk / 系统默认栈作为主字体

## 组件

- **MilestoneRail**：左侧轨道，当前项酸性左边框，进度条 + 风险色点
- **Kanban**：四列等宽，卡片可拖拽；🍅 `consumed/estimated`
- **FocusStage**：全屏倒计时、环形 SVG、白噪音三档
- **BurndownStage**：ECharts 双折线，理想线虚线 cyan，真实线实线 acid，范围变更 markPoint
- **Toast / Modal**：自定义，不用 `alert/confirm`

## 动效

| 状态 | 动效 |
|---|---|
| start | 圆环加速点亮 + 数字缩放 |
| pause | 饱和度降低，呼吸光 |
| complete | 酸性闪烁 + 粒子爆发文案「SETTLED」 |
| abort | 轻微抖动 + 玫瑰色消退 |

## 响应式

- ≥1280：三栏（轨 + 看板 + 可选侧）
- ≥768：轨可折叠
- <768：单列，看板横向滚动，倒计时仍全屏

## 四态

Loading 骨架 / Empty 空轨道文案 / Error 可关闭条（5s） / Success Toast
