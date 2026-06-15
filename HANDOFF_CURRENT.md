# MiMo Desktop - HANDOFF (2026-06-15)

> 工作区：`D:\works\study\mimo cli`
> 分支：`master` | 远端：`origin/master`
> 状态：工作区干净，与远端同步。

---

## 1. 对话规则

- 所有回复用中文，包括思考。
- 每次回复开头称呼用户为"哥哥"。
- 不要用 git reset --hard 、git checkout -- 等会丢改动的命令。
- 未提交改动默认是工作成果，不回退。

---

## 2. Git 提交历史

```text
aebc0c5 chore: regenerate wails bindings for task methods
98de7ae feat: tree-style task IDs, rename/archive/progress
76e0720 feat: memory config now affects runtime behavior
e5ad79a feat: real LLM execution for sub-agents (actors)
5c7dda8 feat: workspace file tree, welcome-first new chat, sidebar styling
0d072fc feat: refine sidebar conversations
aef153e feat: enforce persisted permission rules
7b75b2f feat: persist advanced settings
0e55a2e docs: define advanced settings persistence
0bdaebc feat: auto checkpoint saved chat sessions
```

---

## 3. 本轮完成的功能

### 3.1 任务面板 rename/archive/progress（本轮新增）
- `TaskPanel.tsx` 新增 4 个 state（renamingTaskId/renameValue/progressTaskId/progressValue）。
- 新增 3 个 handler：`handleRenameTask`、`handleArchiveTask`、`handleProgressTask`。
- 新增 3 个操作按钮：重命名（紫色）、进度（青色）、归档（灰色）。
- 新增 2 个 inline 编辑器：重命名输入框、进度说明输入框，支持 Enter 提交和取消。
- `STATUS_COLORS` 新增 `archived: "text-gray-500"`，`STATUS_ICONS` 新增 `archived: Archive`。
- `i18n.ts` 新增 6 个翻译键：`task_rename/task_archive/task_progress/task_progress_placeholder/task_rename_placeholder/task_unblock`。

### 3.2 文件预览增强 — 图片支持（本轮新增）
- `app_files.go`：`FilePreview` 结构体新增 `IsImage`/`Mime` 字段。
- 新增 `isImageExtension` 和 `mimeByExtension` 辅助函数，支持 png/jpg/gif/webp/bmp/ico/svg。
- `ReadFilePreview` 对图片文件返回 base64 编码内容 + MIME 类型。
- `RightSidebar.tsx` 新增图片预览区域，使用 `data:` URL 渲染 base64 图片。

### 3.3 文件预览增强 — Markdown 渲染（本轮新增）
- 新建 `components/file/MarkdownPreview.tsx`，基于 `react-markdown`。
- 使用 Tailwind prose 类 + 自定义 `[&_xxx]` 选择器实现深色主题适配。
- 支持 h1-h3、段落、代码块、行内代码、列表、链接、引用、表格、图片、分隔线。
- `RightSidebar.tsx` 中 `language === "markdown"` 时自动使用 MarkdownPreview。

### 3.4 文件预览增强 — JSON 折叠树（本轮新增）
- 新建 `components/file/JsonTree.tsx`，支持对象/数组递归渲染。
- 默认展开前 2 层，深层可点击折叠/展开。
- 字符串绿色、数字琥珀色、布尔紫色、null 灰色，key 蓝色。
- `RightSidebar.tsx` 中 `language === "json"` 时尝试 JSON.parse，成功用 JsonTree，失败回退 CodeBlock。

### 3.5 文件预览能力（上轮完成）
- `desktop/app_files.go` 新增 `ReadFilePreview`，支持目录、文本文件、二进制文件识别。
- `RightSidebar.tsx` 新增 `PreviewView`，支持返回文件树、行号展示、Reveal in explorer。

### 3.6 文件树真实数据（5c7dda8）
- `desktop/app_files.go`：`ListWorkspaceFiles` 和 `ListDirChildren`，尊重 `.mimoignore`。
- `FileTree.tsx` 从 mock 改为懒加载真实目录数据。
- `RightSidebar.tsx` 新增 Activity / Files 双 Tab。

### 3.7 新建对话跳欢迎页（5c7dda8）
- `handleNewChat` 不再创建后端 session，只清空状态回欢迎页。
- 实际 session 在 `handleSend` 发第一条消息时懒创建。

### 3.8 子智能体真实 LLM 执行（e5ad79a）
- `internal/actor/actor.go` 新增 `Executor` 接口 + `NewRegistryWithExecutor`。
- `desktop/app_actor_executor.go` 实现完整 tool-use 循环（最多 10 轮）。
- `app.go` 启动时注入真实 executor，无 LLM 时回退 mock。

### 3.9 Memory 配置生效（76e0720）
- `Config{CCIndex, SearchScoreFloor}` 传入 `NewServiceWithConfig`。
- `Reconcile` 检查 `CCIndex=false` 时只做 prune。
- `ftsSearch` 用 `SearchScoreFloor` 替代硬编码 0.15。

### 3.10 任务 ID 语义化（98de7ae）
- ID 从 `T{ns}` 改为 `T1/T1.1` 树状编号。
- 新增 Rename/Archive/Progress + `TaskArchived` 状态。
- Desktop 层暴露 `TaskRename/TaskArchive/TaskProgress`。

---

## 4. 验证状态

本轮完成后，已通过以下校验：

- `cd desktop/frontend; npm run build` — 2948 modules, build success
- `go build ./desktop/... ./internal/...` — success
- `go vet ./desktop/... ./internal/...` — success
- `go test ./desktop/... ./internal/...` — all pass (desktop 2.4s)

注意：`npx tsc --noEmit` 仍存在历史窗口类型噪声，不阻塞功能交付。

---

## 5. 已知潜在问题

1. **Vite 大 chunk warning**：`index.js` 超 500KB，需 code-split。Node ESM warning 仍在。不影响功能。
2. **Wails 自动生成文件**：每次 `wails generate module` 会改 `App.d.ts/App.js/models.ts`，需手动提交。
3. **actor 测试覆盖不足**：现有测试只覆盖 mock registry，未覆盖 `llmExecutor`。建议补 mock gateway 集成测试。
4. **任务树 ID 与旧数据混合**：旧 `T{ns}` 格式仍存在 DB，新任务从 `T1` 开始。同 session 新旧混合，前端显示可能不一致。非阻塞。
5. **Memory CCIndex=false UX**：关闭后 Search 仍工作（搜已有索引），但 Reconcile 不再索引新文件。用户可能误解。需 UI 说明。
6. **前端 `tsc --noEmit` 噪声仍在**：建议后续统一收敛全局 Window 类型定义。
7. **react-markdown 新增依赖**：`package-lock.json` 和 `node_modules` 会更新，需提交。

---

## 6. 建议下一步

### 前端体验
1. **文件预览进一步增强**：PDF 预览、视频缩略图、代码折叠。
2. **Code splitting**：处理 Vite 大 chunk warning。
3. **任务面板事件历史**：点击任务展开时加载 `TaskGetEvents` 显示事件时间线。

### 后端功能
4. **Actor 流式输出**：`llmExecutor` 用非流式 Chat，改 `ChatStream` + 事件推送。
5. **Skill 系统对接**：Dream/Distill 产出应落盘为 skill 文件。
6. **MCP 工具在 actor 可用**：确认 `llmExecutor` 是否包含 MCP 工具。
7. **会话导出增强**：支持 Markdown 格式导出。

---

## 7. 关键文件索引

### 后端
- `desktop/app.go` — 主 Wails 绑定对象
- `desktop/app_files.go` — 文件树接口 + 文件预览接口（含图片 base64）
- `desktop/app_actor.go` — Actor Wails 绑定
- `desktop/app_actor_executor.go` — LLM 执行器
- `desktop/app_memory.go` — Memory Wails 绑定
- `desktop/app_task.go` — Task Wails 绑定（含 Rename/Archive/Progress）
- `internal/actor/actor.go` — Actor 生命周期 + Executor 接口
- `internal/agent/agent.go` — 核心 Agent
- `internal/memory/service.go` — Memory 检索 + 索引
- `internal/task/registry.go` — 任务持久化 + 树状 ID
- `internal/tools/registry.go` — 工具注册中心
- `internal/llm/gateway.go` — LLM 网关

### 前端
- `App.tsx` — 全局入口 + 事件绑定
- `components/layout/AppLayout.tsx` — 主布局
- `components/layout/LeftSidebar.tsx` — 左侧栏
- `components/layout/RightSidebar.tsx` — 右侧栏（Activity + Files + File Preview）
- `components/file/FileTree.tsx` — 文件树
- `components/file/MarkdownPreview.tsx` — Markdown 渲染组件（新增）
- `components/file/JsonTree.tsx` — JSON 折叠树组件（新增）
- `components/task/TaskPanel.tsx` — 任务面板（含 rename/archive/progress）
- `stores/sessionStore.ts` — Session 状态
- `stores/chatStore.ts` — Chat 状态
- `lib/i18n.ts` — 国际化（含新增任务操作翻译键）

---

## 8. 构建命令

```powershell
cd desktop/frontend; npm install
cd desktop/frontend; npm run build
wails generate module
go build ./desktop/... ./internal/...
go test ./desktop/... ./internal/...
go vet ./desktop/... ./internal/...
git -c http.sslBackend=schannel -c http.version=HTTP/1.1 push origin master
```
