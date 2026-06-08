# MiMo Desktop — 项目状态报告

> 更新日期：2026-06-08
> 版本：v0.3.0-dev
> 技术栈：Wails v2 + Go 1.26 + React 19 + TypeScript + Zustand + Tailwind CSS

---

## 一、当前开发进度

### 1.1 整体架构

```
┌─────────────────────────────────────────────┐
│           React Frontend (TypeScript)        │
│  AppLayout / ChatInput / MessageList / ...   │
│  Zustand Stores / useAgent Hook / i18n       │
├──────────────── Wails Bridge ────────────────┤
│              Go Backend (desktop/)            │
│  app.go / app_chat.go / app_session.go / ... │
├─────────────────────────────────────────────┤
│           internal/ 核心引擎                  │
│  agent / llm / tools / safety / mcp / ...    │
└─────────────────────────────────────────────┘
```

### 1.2 已完成的功能模块

#### Go 后端 (`desktop/`)

| 文件 | 功能 | 状态 |
|------|------|------|
| `app.go` | App 初始化、基础设施组装、窗口控制、安全确认桥接、11 个 Agent 回调注册 | ✅ 完整 |
| `app_chat.go` | `SendMessage` 流式对话、`CancelOperation`、安全确认响应、`CompressContext`、`GetModelName` | ✅ 完整 |
| `app_config.go` | `GetConfig` DTO、`SetTheme/Language/DefaultModel/SafetyLevel/PlanningMode`、`AddModel/RemoveModel` | ✅ 完整 |
| `app_session.go` | `ListSessions/LoadSession/CreateNewSession/DeleteSession/RenameSession/SaveSessionFromFrontend` | ✅ 完整 |
| `app_tools.go` | `GetWorkingDir`、`GetTools`（含 MCP 标记）、`GetMCPServers` | ✅ 完整 |
| `events.go` | 15 个事件名常量 | ✅ 完整 |
| `wails_main.go` | Wails 入口，1200x800 无边框窗口，暗色主题 | ✅ 完整 |

#### React 前端 (`desktop/frontend/src/`)

| 模块 | 文件 | 功能 | 状态 |
|------|------|------|------|
| **状态管理** | `stores/chatStore.ts` | 消息列表、流式状态、工具调用、确认弹窗 | ✅ |
| | `stores/sessionStore.ts` | 会话列表、当前会话 ID、侧栏开关 | ✅ |
| | `stores/settingsStore.ts` | 主题/语言/字号、从后端配置初始化 | ✅ |
| | `stores/activityStore.ts` | 活动日志、文件变更、计划步骤 | ✅ |
| **事件系统** | `hooks/useAgent.ts` | 监听 16 个 Wails 事件，分发到 stores | ✅ |
| | `lib/events.ts` | `EVENTS` 常量 + `EventsOn/EventsOff` 封装 | ✅ |
| **聊天组件** | `chat/MessageList.tsx` | 消息列表、流式渲染、自动滚动 | ✅ |
| | `chat/MessageBubble.tsx` | 消息气泡（用户/AI）、thinking 折叠、工具卡片 | ✅ |
| | `chat/ChatInput.tsx` | textarea 自适应高度、Enter 发送、流式取消 | ✅ |
| | `chat/StatusBar.tsx` | 底部状态栏：thinking/工具执行/模型名 | ✅ |
| | `chat/MarkdownRenderer.tsx` | GFM Markdown 渲染 + 代码高亮 | ✅ |
| | `chat/ThinkingBlock.tsx` | 思考过程折叠块 | ✅ |
| | `chat/ToolCallCard.tsx` | 工具调用卡片（参数/结果可展开） | ✅ |
| **布局** | `layout/AppLayout.tsx` | 三栏布局 + 无边框标题栏 + 窗口控制 | ✅ |
| | `layout/LeftSidebar.tsx` | 会话列表（按工作区分组）、右键菜单、批量管理 | ✅ |
| | `layout/RightSidebar.tsx` | 活动日志、计划进度 | ✅ |
| **其他** | `confirm/ConfirmDialog.tsx` | 安全确认弹窗（approve/deny/approve-all） | ✅ |
| | `common/ToolsViewer.tsx` | 工具查看器弹窗 | ✅ |
| | `settings/SettingsPage.tsx` | 设置页（通用/工具 MCP/模型/帮助 4 个 tab） | ✅ |
| | `lib/types.ts` | TypeScript 类型定义 | ✅ |
| | `lib/i18n.ts` | 中英双语国际化（100+ 翻译 key） | ✅ |

#### internal/ 核心引擎（与 CLI 共享）

| 包 | 功能 | 状态 |
|---|------|------|
| `internal/agent/` | ReAct 循环、流式对话、工具执行、并行调用、上下文压缩、11 种回调 | ✅ |
| `internal/llm/` | 多 LLM 提供商（Anthropic/OpenAI/MiMo/DeepSeek）、流式/非流式、Gateway 路由 | ✅ |
| `internal/tools/` | 25 个内置工具 + Registry | ✅ |
| `internal/safety/` | 三级安全策略（lockdown/confirm/auto）、操作分类、审计日志 | ✅ |
| `internal/mcp/` | MCP 协议客户端、stdio/SSE 传输、自动工具注册 | ✅ |
| `internal/session/` | SQLite 会话持久化（`~/.mimo/sessions.db`） | ✅ |
| `internal/config/` | 多层 YAML 配置（系统/用户/项目）深度合并 | ✅ |
| `internal/context/` | 系统提示词构建、上下文管理、规则注入 | ✅ |
| `internal/backup/` | 文件写入前自动备份 | ✅ |
| `internal/ignore/` | `.mimoignore` 忽略规则 | ✅ |

### 1.3 前端与后端 API 对照

前端 `App.tsx` 声明了 30 个 `window.go.desktop.App.*` 方法，**全部在 Go 后端有对应实现，无缺失**。

---

## 二、已知 Bug

### 2.1 高优先级

| # | Bug | 位置 | 影响 | 修复方案 |
|---|-----|------|------|---------|
| 1 | **`RespondToConfirm` 可能永久阻塞** | `app_chat.go:67` | 用户关闭确认弹窗时，goroutine 永久阻塞在 channel 上，内存泄漏 | 加 `select` + `time.After` 超时（30s），超时自动拒绝 |
| 2 | **`SendMessage` 使用 `context.Background()`** | `app_chat.go:32` | `CancelOperation` 只设内部标志，无法通过 context 传递取消信号，长任务取消可能延迟 | 改用 `context.WithCancel`，`Cancel()` 同时 cancel context |
| 3 | **主题闪烁** | `settingsStore.ts:49-50` | 模块加载时硬设 `dark` class，`initFromConfig` 后可能修正，首次渲染闪烁 | 延迟 class 添加到 `initFromConfig` 中执行 |

### 2.2 中优先级

| # | Bug | 位置 | 影响 | 修复方案 |
|---|-----|------|------|---------|
| 4 | **`RespondToConfirmAll` 缺少并发保护** | `app_chat.go:71` | `confirmAll` 字段读写无锁 | 使用 `a.mu` 保护或改用 `atomic.Bool` |
| 5 | **`currentSessionID` 缺少并发保护** | `app_session.go` | `CreateNewSession`/`SaveCurrentSession` 读写无锁 | 统一使用 `a.mu` 保护 |
| 6 | **`ListSessions` 重复调用** | `App.tsx:104` + `LeftSidebar.tsx:219` | mount 时调用两次，limit 不同（30 vs 50），存在竞态 | 只在 App.tsx 中调用一次，LeftSidebar 通过 store 读取 |
| 7 | **`OpenInExplorer` 忽略错误** | `app.go:233` | `os.Getwd()` 失败时 `path` 为空串，打开不可预期位置 | 检查 error，失败时 return |
| 8 | **`prose` 类不生效** | `MarkdownRenderer.tsx` | 使用了 `prose prose-sm prose-invert` 但未安装 `@tailwindcss/typography` 插件 | `npm install @tailwindcss/typography` 并配置 tailwind.config |

### 2.3 低优先级

| # | Bug | 位置 | 影响 |
|---|-----|------|------|
| 9 | `GetVersion` 硬编码版本号 `"0.1.0"` | `app_chat.go:84` | 版本号不随构建更新 |
| 10 | `GetConfig` 中 `Safety.Permission` 映射自 `Agent.Permission` | `app_config.go:60` | 可能是字段映射错误 |
| 11 | `MCP 工具名` 用 `__` 判断 MCP 归属 | `app_tools.go:34-39` + `ToolCallCard.tsx:38` | 工具名含 `__` 时误判 |
| 12 | `ToolCallEvent.status` 缺少 `"error"` | `types.ts` | 与 `ActivityEntry.status` 语义不一致 |
| 13 | `SyntaxHighlighter` 亮色模式仍用暗色主题 | `MarkdownRenderer.tsx` | 主题切换时代码块颜色不跟随 |

---

## 三、缺失功能

### 3.1 核心功能缺失

| # | 功能 | 说明 | 优先级 |
|---|------|------|--------|
| 1 | **Agent 工具执行闭环未贯通** | MCP 工具调用集成完成，但 Agent loop 中工具执行 → 结果返回 → 继续推理的完整流程未在桌面端验证 | P0 |
| 2 | **上下文压缩 UI 反馈** | Go 侧 `CompressContext` 已实现，前端 `agent:compressing` 事件已监听，但无进度条/完成提示 UI | P1 |
| 3 | **对话导出** | CLI 有 `/export` 命令，桌面端无对应功能 | P2 |
| 4 | **多行输入** | 前端 textarea 支持 Shift+Enter 换行，但未与 Go 侧的多行消息处理打通验证 | P2 |

### 3.2 UI 功能缺失

| # | 功能 | 说明 | 优先级 |
|---|------|------|--------|
| 5 | **`pinnedDirs` 不持久化** | 置顶的工作区目录存在 `useState`，刷新后丢失 | P2 |
| 6 | **Token 预算警告** | CLI 有 96k token 变色警告，桌面端无对应 | P2 |
| 7 | **输入框无历史记录** | CLI 有上下键回溯历史，桌面端 textarea 无此功能 | P3 |
| 8 | **欢迎横幅** | CLI 有版本/模型欢迎横幅，桌面端只有空态 "MiMo" | P3 |
| 9 | **`EVENTS.CHAT_START` 未监听** | `events.ts` 定义了但 `useAgent.ts` 未使用，无法在 UI 上反映对话开始状态 | P3 |

### 3.3 i18n 遗漏

| 位置 | 硬编码文本 | 应改为 |
|------|-----------|--------|
| `AppLayout.tsx:103` | `"MiMo"` 标题 | `t("app_name")` |
| `LeftSidebar.tsx:59` | `"其他"` | `t("other_projects")` |
| `AppLayout.tsx:122` | `t("tools")` | key 不在 `TranslationKey` 中，需补充 |
| `SettingsPage.tsx:322,672` | `"MiMo Desktop v0.1.0"` | 从后端 `GetVersion()` 获取 |

---

## 四、代码质量问题

### 4.1 死代码

| 文件 | 代码 | 说明 |
|------|------|------|
| `app_session.go:102-110` | `SaveCurrentSession()` | 空壳 stub，前端从未调用 |
| `app_chat.go:100-106` | `toJSON()` | 辅助函数，无任何调用 |
| `desktop/main.go` | 整个文件 | 只有 package 声明 |
| `events.ts:47-50` | `EventsOff()` | 已导出但从未调用 |
| `package.json` | `clsx` 依赖 | 未在任何源文件中导入 |

### 4.2 重复代码

| 代码 | 位置 | 建议 |
|------|------|------|
| `SafetyBadge` 组件 | `ToolsViewer.tsx` + `SettingsPage.tsx` | 提取到 `common/SafetyBadge.tsx` |
| `ToolInfo` / `MCPServerInfo` 接口 | `ToolsViewer.tsx` + `SettingsPage.tsx` | 提取到 `lib/types.ts` |

### 4.3 类型安全问题

| 位置 | 问题 | 建议 |
|------|------|------|
| `App.tsx` | `(window as Record<string, unknown>).__currentSessionId` | 用 zustand store 管理 |
| `useAgent.ts` | 大量 `args[0] as string` 强制断言 | 加运行时类型守卫 |
| `types.ts` | `ToolCallEvent` 缺少 `id` 字段 | ✅ 已修复（加了 `id` 字段） |

### 4.4 硬编码值汇总

| 值 | 出现次数 | 建议 |
|----|---------|------|
| 版本号 `"v0.1.0"` | 3 处 | 用 `ldflags` 注入或从后端获取 |
| 会话列表上限 30/50 | 2 处（冲突） | 统一为常量 |
| 各种截断长度（35/120/200/500） | 6 处 | 提取为配置常量 |
| 侧栏宽度 260/320px | 3 处 | 提取为 CSS 变量或常量 |

---

## 五、优化建议

### 5.1 性能优化

| # | 优化项 | 说明 |
|---|--------|------|
| 1 | **代码分割** | 前端 JS bundle 1MB+，可用 `React.lazy` + `import()` 按路由分割（设置页、工具查看器等非首屏组件） |
| 2 | **自动滚动优化** | 大量消息时 `scrollIntoView({ behavior: "smooth" })` 可能卡顿，改用 `scrollTop = scrollHeight` |
| 3 | **消息虚拟列表** | 消息数 >100 时考虑用虚拟滚动（`react-window` 或 `@tanstack/virtual`） |
| 4 | **Markdown 渲染优化** | 长消息的 Markdown 渲染可用 `React.memo` + `useMemo` 避免不必要的重渲染 |

### 5.2 体验优化

| # | 优化项 | 说明 |
|---|--------|------|
| 5 | **流式期间允许复制** | ✅ 已修复（`readOnly` 替代 `disabled`） |
| 6 | **消息右键菜单** | 复制文本、复制为 Markdown、重新生成、删除 |
| 7 | **拖拽调整侧栏宽度** | 固定宽度不够灵活 |
| 8 | **文件变更预览** | 工具执行文件操作时，在右侧栏显示 diff 预览 |
| 9 | **键盘导航** | `Ctrl+Shift+N` 新窗口、`Ctrl+W` 关闭会话、`Ctrl+1-9` 切换会话 |
| 10 | **系统托盘** | 最小化到托盘，保持后台运行 |

### 5.3 架构优化

| # | 优化项 | 说明 |
|---|--------|------|
| 11 | **统一版本管理** | 用 `ldflags` 在构建时注入版本号到 Go，前端通过 `GetVersion()` 获取 |
| 12 | **错误边界** | React 组件加 `ErrorBoundary`，避免白屏崩溃 |
| 13 | **事件类型化** | `EventsOn` 回调参数从 `unknown[]` 改为泛型，编译期类型检查 |
| 14 | **安全确认超时** | `confirmChan` 加 30s 超时，防止 goroutine 泄漏 |
| 15 | **context 取消** | `SendMessage` 用 `context.WithCancel`，`CancelOperation` 同时 cancel context |

---

## 六、未来开发路线

### Phase 1: 稳定化（当前 → 1 周）

- [ ] 修复高优先级 Bug（#1-#3）
- [ ] 贯通 Agent 工具执行闭环完整测试
- [ ] 安装 `@tailwindcss/typography` 修复 Markdown 排版
- [ ] 统一 `ListSessions` 调用，消除重复请求
- [ ] 清理死代码（`SaveCurrentSession` stub、`toJSON`、空 `main.go`）

### Phase 2: 功能补全（1-2 周）

- [ ] 上下文压缩 UI 反馈（进度条 + 完成统计）
- [ ] 对话导出功能（Markdown/JSON）
- [ ] Token 预算警告（状态栏变色）
- [ ] 消息右键菜单（复制/重新生成/删除）
- [ ] 提取公共组件（`SafetyBadge`、`ToolInfo` 接口）
- [ ] 版本号通过 `ldflags` 注入

### Phase 3: 体验提升（2-4 周）

- [ ] 代码分割（设置页/工具查看器 lazy load）
- [ ] 消息虚拟列表
- [ ] 拖拽调整侧栏宽度
- [ ] 文件变更 diff 预览
- [ ] 输入历史记录
- [ ] 系统托盘支持
- [ ] React ErrorBoundary

### Phase 4: 高级功能（1-2 月）

- [ ] 多窗口/多标签
- [ ] 插件系统 UI
- [ ] 向量化记忆搜索
- [ ] 沙箱执行（Docker 隔离）
- [ ] MCP 服务器管理 UI（添加/删除/重启）
- [ ] 协作模式（共享会话）

---

## 七、打包流程

### 7.1 环境准备

```bash
# 前置依赖
# - Go 1.26+
# - Node.js 18+
# - Wails CLI v2
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 首次安装前端依赖
cd desktop/frontend && npm install
```

### 7.2 开发模式

```bash
cd "D:\works\study\mimo cli"
wails dev -tags wails
```

- 自动启动 Vite dev server（hot reload）
- 前端修改实时生效
- Go 修改自动重编译
- 默认端口 http://localhost:5173（自动代理到 Wails）

### 7.3 生产打包

```bash
cd "D:\works\study\mimo cli"
wails build -tags wails
```

自动执行：
1. 生成 Wails JS bindings（`desktop/frontend/src/wails/`）
2. 安装前端依赖（`npm install`）
3. 编译前端（`npm run build` → `desktop/frontend/dist/`）
4. 编译 Go 代码（`go build -tags wails`）
5. 打包为独立 exe

输出：`build/bin/mimo-desktop.exe`（约 17MB）

### 7.4 Go 代理配置（国内网络）

```bash
$env:GOPROXY="https://goproxy.cn,direct"
$env:GOCACHE="D:\works\study\mimo cli\.gocache"
$env:GOMODCACHE="D:\works\study\mimo cli\.gomod"
```

### 7.5 纯 Go 编译（不打包前端）

```bash
# 仅编译 Go 代码，需先手动 build 前端
cd desktop/frontend && npm run build
cd "D:\works\study\mimo cli"
go build -tags wails -o bin/mimo-desktop.exe .
```

### 7.6 CLI 版本编译（不带 wails）

```bash
go build -o bin/mimo.exe ./main.go
```

---

## 八、项目目录结构

```
mimo cli/
├── main.go                    # CLI 入口（//go:build !wails）
├── wails_main.go              # Desktop 入口（//go:build wails）
├── wails.json                 # Wails 配置
├── go.mod / go.sum
├── cmd/                       # CLI 命令（Cobra + Bubbletea TUI）
│   ├── root.go / run.go / interactive.go / version.go
│   └── tui_model.go / tui_view.go / tui_styles.go / ...
├── internal/                  # 核心引擎（CLI 和 Desktop 共享）
│   ├── agent/                 # Agent 循环（ReAct/Plan-Execute）
│   ├── llm/                   # LLM 提供商抽象
│   ├── tools/                 # 25 个内置工具
│   ├── safety/                # 安全护栏
│   ├── mcp/                   # MCP 协议客户端
│   ├── session/               # SQLite 会话存储
│   ├── config/                # 配置管理
│   ├── context/               # 上下文管理
│   ├── backup/                # 文件备份
│   └── ignore/                # 忽略规则
├── desktop/                   # Desktop 专用代码
│   ├── app.go                 # App 主结构体 + 初始化
│   ├── app_chat.go            # 聊天相关方法
│   ├── app_config.go          # 配置读写方法
│   ├── app_session.go         # 会话管理方法
│   ├── app_tools.go           # 工具/MCP 查询方法
│   ├── events.go              # 事件名常量
│   ├── main.go                # 空文件（仅 package 声明）
│   └── frontend/              # React 前端
│       ├── package.json
│       ├── vite.config.ts
│       ├── tailwind.config.ts
│       └── src/
│           ├── App.tsx
│           ├── main.tsx
│           ├── components/    # 14 个组件
│           ├── stores/        # 4 个 Zustand stores
│           ├── hooks/         # useAgent
│           └── lib/           # types / events / i18n
├── build/                     # Wails 打包输出
│   └── bin/mimo-desktop.exe
└── *.md                       # 文档（HANDOFF / TASK_PROGRESS / ...）
```

---

## 九、关键配置文件

| 文件 | 说明 |
|------|------|
| `~/.mimo/config.yaml` | 用户配置（API Key、模型、安全策略） |
| `~/.mimo/sessions.db` | SQLite 会话数据库 |
| `~/.mimo/history` | CLI 命令历史 |
| `~/.mimo/mcp/` | MCP 包全局安装目录 |
| `.mimo/rules.md` | 项目级规则文件 |
| `.mimoignore` | 忽略规则 |
| `AGENT.md` | 代理指令文件 |
