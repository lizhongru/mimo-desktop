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

## 2. Git 提交历史（最近 10 条）

```text
156a930 feat: actor streaming output
3c4eefc feat: task panel event timeline on expand
cf713b8 perf: code splitting — main bundle 1.2MB → 460KB
3277501 feat: file preview modal, JSON tree, code block modernization
86f2302 docs: update HANDOFF for new session
aebc0c5 chore: regenerate wails bindings for task methods
98de7ae feat: tree-style task IDs, rename/archive/progress
76e0720 feat: memory config now affects runtime behavior
e5ad79a feat: real LLM execution for sub-agents (actors)
5c7dda8 feat: workspace file tree, welcome-first new chat, sidebar styling
```

---

## 3. 本轮完成的功能（4 个提交）

### 3.1 文件预览弹窗 + JSON 树 + 代码块现代化（3277501）
- **FilePreviewModal.tsx**（新建）：独立全屏弹窗预览，不挤压聊天区域。
  - 暗色半透明遮罩 + backdrop-blur，适配亮/暗主题。
  - 支持最大化/还原、Esc 关闭、换行切换、在资源管理器中打开。
  - 图片查看器带缩放控制。
  - 文字可选中（`select-text` + CSS 规则覆盖外层 `select-none`）。
- **JsonTree.tsx**（新建/重写）：
  - 左侧固定 `w-5` gutter 专门放折叠箭头，右侧 JSON 内容不受干扰。
  - 配色：字符串绿色、数字橙色、布尔紫色、键名蓝色。
  - 无复制按钮，纯选中模式。
- **CodeBlock.tsx**（新建）：移除复制按钮，现代化字体和间距。
- **MarkdownPreview.tsx**（新建）：暗色主题 Markdown 渲染。
- **RightSidebar.tsx**（重写）：
  - 文件树始终显示，不再被预览替换。
  - 预览加载/错误改为底部浮动 toast，不遮挡文件树。
  - 文件夹展开状态在预览操作中保持不变。

### 3.2 Code splitting（cf713b8）
- `vite.config.ts`：`manualChunks` 拆分 syntax-highlighter、vendor-react、vendor-state、vendor-icons。
- `AppLayout.tsx`：6 个面板组件改为 `React.lazy`（SettingsPage、ToolsViewer、MemoryPanelModal、CheckpointPanelModal、TaskPanelModal、ActorPanelModal）。
- **效果**：主包 1,212KB → 460KB（-62%），gzip 392KB → 137KB。

### 3.3 任务面板事件历史（3c4eefc）
- `TaskPanel.tsx`：展开任务时懒加载事件（`TaskGetEvents` API），首次展开才请求。
- 左侧垂直时间线 + 彩色圆点标记事件类型（created/started/completed/blocked/progress/renamed 等）。
- 显示时间戳和事件摘要。

### 3.4 Actor 流式输出（156a930）
- `internal/actor/actor.go`：`Executor` 接口新增 `SetStreamCallback`，`Registry` 透传回调。
- `desktop/app_actor_executor.go`：`llmExecutor` 改用 `ChatStream` 替代 `Chat`，流式推送文本 delta 和工具调用通知。
- `desktop/app_actor.go`：`initActorRegistry` 注册流式回调，通过 `runtime.EventsEmit` 推送 `actor:delta` 事件。
- `desktop/events.go` + `lib/events.ts`：新增 `EventActorDelta` / `ACTOR_DELTA` 事件。
- `ActorPanel.tsx`：监听流式事件，running 状态下实时显示输出 + 闪烁光标。

---


### 3.5 会话导出增强 — Markdown 导出（当前提交）
- **LeftSidebar.tsx**：会话右键菜单新增"导出对话"选项（Download 图标）。
  - ContextMenu 组件新增 onExport prop，仅在有 sessionId 时显示。
  - LeftSidebar 函数解构 onExportSession 并传入 ContextMenu。
  - 点击后调用已有的 handleExportSession → 后端 ExportChat 导出为 .md 文件。
- 后端 desktop/app_chat.go 的 ExportChat 方法已支持 Markdown 格式（标题、角色 emoji、时间戳）。
- 导出流程：加载会话消息 → 转为 ExportMessage → 调用系统保存对话框 → 写入 .md 文件。

---

## 4. 验证状态

- `cd desktop/frontend; vite build` — ✓ built in ~5-10s，主包 460KB
- `go build ./desktop/... ./internal/...` — success
- `go vet ./desktop/... ./internal/...` — success
- `go test ./desktop/... ./internal/...` — all pass

---

## 5. 已知潜在问题

1. **Vite 大 chunk warning**：syntax-highlighter chunk 660KB，可进一步按语言拆分。不阻塞。
2. **react-markdown + react-syntax-highlighter 重复导入**：MarkdownPreview 和 CodeBlock 都导入了 syntax-highlighter，可能造成重复 chunk。考虑统一。
3. **actor 流式输出缓冲**：`streamRef` 在组件卸载后可能仍更新 state（内存泄漏风险）。建议加 unmount 清理。
4. **TaskGetEvents 无缓存**：每次展开都请求后端，可加 localStorage 或 SWR 缓存。
5. **FilePreviewModal 动画**：退出动画用 setTimeout 200ms，如果 onClose 回调中有状态更新可能闪烁。
6. **JsonTree 深层性能**：level < 3 默认展开，超大 JSON（1000+ keys）可能卡顿。考虑虚拟化或限制展开数量。
7. **Wails 自动生成文件**：每次 `wails generate module` 会改 `App.d.ts/App.js/models.ts`，需手动提交。
8. **前端 `tsc --noEmit` 噪声**：历史窗口类型定义不完整，不阻塞功能。
9. **Node.js 版本**：当前 22.3.0，Vite 7 要求 22.12+，有 warning 但不影响构建。

---

## 6. 建议下一步（按优先级）

### 高优先级
1. **Skill 系统对接**：Dream/Distill 产出应落盘为 `.mimo/skills/` 下的 skill 文件。
2. **MCP 工具在 actor 可用**：确认 `llmExecutor` 是否包含 MCP 工具定义，当前 `e.tools.Definitions()` 应已包含，需实际测试验证。
3. ~~**会话导出增强**~~ ✅ 已完成：左侧栏右键菜单可导出 Markdown。

### 中优先级
4. **输入历史记录**：上下箭头翻阅历史消息。
5. **消息右键菜单**：复制/重新生成/删除。
6. **Token 预算警告**：状态栏变色提示上下文接近上限。

### 低优先级
7. **PDF 预览**：FilePreviewModal 支持 PDF 渲染。
8. **文件变更 diff 预览**：在预览弹窗里做 diff 视图。
9. **系统托盘支持**。
10. **多窗口/多标签**。

---

## 7. 关键文件索引

### 后端
- `desktop/app.go` — 主 Wails 绑定对象
- `desktop/app_files.go` — 文件树 + 文件预览
- `desktop/app_actor.go` — Actor Wails 绑定 + 流式回调注册
- `desktop/app_actor_executor.go` — LLM 执行器（流式 ChatStream）
- `desktop/app_task.go` — Task Wails 绑定（含 GetEvents）
- `desktop/events.go` — 事件名常量（含 EventActorDelta）
- `internal/actor/actor.go` — Actor 生命周期 + Executor 接口（含 StreamCallback）
- `internal/agent/agent.go` — 核心 Agent
- `internal/llm/gateway.go` — LLM 网关

### 前端
- `App.tsx` — 全局入口 + 事件绑定
- `components/layout/AppLayout.tsx` — 主布局 + 6 个 lazy 面板
- `components/layout/RightSidebar.tsx` — 文件树 + toast 提示
- `components/file/FilePreviewModal.tsx` — 文件预览弹窗
- `components/file/JsonTree.tsx` — JSON 折叠树（左侧 gutter）
- `components/file/CodeBlock.tsx` — 代码块
- `components/file/MarkdownPreview.tsx` — Markdown 渲染
- `components/task/TaskPanel.tsx` — 任务面板（含事件时间线）
- `components/actor/ActorPanel.tsx` — 子智能体面板（含流式输出）
- `lib/events.ts` — 事件常量（含 ACTOR_DELTA）
- `vite.config.ts` — Code splitting 配置

---

## 8. 构建命令

```powershell
cd desktop/frontend; npm install
cd desktop/frontend; npx vite.cmd build
go build ./desktop/... ./internal/...
go test ./desktop/... ./internal/...
go vet ./desktop/... ./internal/...
```

---

## 9. AGENTS.md 规则

- 所有回答必须用中文。
- 每次回复开头称呼用户为"哥哥"。
