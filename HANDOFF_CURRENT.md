# MiMo Desktop - HANDOFF (2026-06-16)

> 工作区：`D:\works\study\mimo cli`
> 分支：`master` | 远端：`origin/master`
> 状态：本文件记录截至 2026-06-16 的当前开发状态；最新提交请以 `git log --oneline -5` 为准。

---

## 1. 对话规则

- 所有回复必须用中文。
- 每次回复开头称呼用户为“哥哥”。
- 不要使用 `git reset --hard`、`git checkout --` 等可能丢失改动的命令。
- 未提交改动默认是工作成果，不要回退。

---

## 2. 最近提交基线

```text
18294d0 docs: 更新 HANDOFF 交接文档 — 消息操作栏、流式指示器、会话切换修复
84bf652 feat: 消息操作栏增强、侧边栏流式指示器、会话切换bug修复
a99e00f fix: use smooth scroll for session load
0e9c09e fix: reliable scroll-to-bottom on session load using containerRef + double-rAF
c20f0c1 fix: scroll to bottom after loading session messages
```

---

## 3. 本轮完成内容

### 3.1 会话切换与后台流式修复
- 根因：前端 `chatStore` 原本只有一套全局聊天/流式状态，当前展示会话和后台仍在生成的原会话会互相污染。
- `chatStore` 新增后台流式缓冲：`backgroundSessionId/backgroundMessages/backgroundThinking/backgroundDelta/backgroundToolCalls`。
- `App.tsx` 切换会话时，如果当前会话仍在流式，会先把流式状态转入后台缓冲，再加载目标会话。
- `useAgent.ts` 的 `DELTA/THINKING/TOOL_CALL/TOOL_RESULT` 会根据后台会话状态写入后台缓冲，不再写入当前前台会话。
- `CHAT_DONE/CHAT_ERROR/CHAT_CANCELLED` 支持后台会话收尾并回写原会话，避免 A 后台结束时出现在 B 对话中。
- 新增 `sessionSnapshots` 快照缓存，解决完成后异步保存与快速切回之间的竞态；切回时优先展示前端快照，后端保存成功后清理快照。

### 3.2 `firstMessage` 持久化
- `internal/session/store.go` 新增 `sessions.first_message` 迁移，并回填历史会话的首条用户消息。
- `Session` / `SessionDTO` / Wails models 同步增加 `firstMessage` 字段。
- `SaveSession` 保存时保留首条用户消息，不被后续 `lastMessage` 覆盖。
- `LeftSidebar` 标题优先显示 `firstMessage`，回退 `lastMessage`。

### 3.3 文本选择与“添加到对话”
- 聊天消息区允许文本选中，外层 `select-none` 不再阻止消息正文选择。
- 选中文字后显示原生 DOM 浮层按钮 `添加到对话`，避免 React 重渲染打断原生选区。
- 增加自绘选区高亮覆盖层，避免 WebView 在按钮出现后隐藏原生选区视觉。
- 点击 `添加到对话` 后，将选中文本按引用块格式追加到底部输入框，并聚焦到输入框末尾。
- `MessageBubble` 在存在有效文本选区时冻结 hover 操作栏，避免复制/重新生成按钮出现导致选区消失。

### 3.4 会话加载滚动优化
- 历史消息加载从逐条 `addRestoredMessage` 改为一次性 `replaceMessages`。
- 大批量历史加载时直接跳到底部；只有直播/流式新增内容时才平滑滚动。
- 解决进入长历史会话时从顶部慢慢滚到底部的问题。

### 3.5 思考中 DNA 动画
- `ThinkingBlock` 的 live 状态改为横向 DNA 双螺旋动画。
- live 状态不再显示“思考中”文字和省略号，只展示动画。
- DNA 动画使用青蓝/紫色渐变，并支持 `prefers-reduced-motion`。
- 历史 thinking 内容仍保留 `Brain` 图标、预览和展开逻辑。

### 3.6 Markdown 代码显示优化
- 行内代码从黄色字 + 灰底改为主题适配的 `mimo-inline-code`。
- fenced code block 增加 `CodeBlock` 包装组件，右上角 hover 显示复制图标，复制成功显示绿色 Check。
- 代码块背景、边框、复制按钮背景新增亮/暗主题变量。
- 使用 `customStyle` / `codeTagProps` 强制覆盖 `SyntaxHighlighter` 内部背景，修复代码块外圈黑底问题。

### 3.7 清理与文档
- 清理 `App.tsx` 中不再使用的事件 selector 和空 `useEffect`。
- 更新 `docs/workspace-architecture.md` 中的 `SessionDTO` 字段说明。
- 更新 Wails TypeScript model 中的 `SessionDTO.firstMessage`。

---

## 4. 修改文件清单

| 文件 | 改动说明 |
|------|---------|
| `desktop/frontend/src/App.tsx` | 会话切换后台流式缓冲、批量恢复消息、快照优先加载、firstMessage 写入 |
| `desktop/frontend/src/hooks/useAgent.ts` | 后台流式事件分流、后台收尾保存、会话快照保存竞态修复 |
| `desktop/frontend/src/stores/chatStore.ts` | 后台流式缓冲、会话快照、批量替换消息 |
| `desktop/frontend/src/stores/sessionStore.ts` | `firstMessage` 补全与更新签名 |
| `desktop/frontend/src/components/layout/LeftSidebar.tsx` | 会话标题优先使用 `firstMessage` |
| `desktop/frontend/src/components/chat/MessageList.tsx` | 选区浮层、自绘高亮、历史加载直接到底部 |
| `desktop/frontend/src/components/chat/ChatInput.tsx` | 监听选区追加事件，将片段加入输入框 |
| `desktop/frontend/src/components/chat/MessageBubble.tsx` | 选区存在时冻结 hover 操作栏 |
| `desktop/frontend/src/components/chat/ThinkingBlock.tsx` | 横向 DNA live thinking 动画 |
| `desktop/frontend/src/components/chat/MarkdownRenderer.tsx` | 代码块复制按钮、行内代码/代码块 class 调整 |
| `desktop/frontend/src/styles/globals.css` | 选择区、DNA 动画、代码主题变量和代码块背景 |
| `desktop/app_session.go` | `SessionDTO.firstMessage` 映射 |
| `internal/session/store.go` | `first_message` 数据库迁移、查询与保存逻辑 |
| `desktop/frontend/src/wails/wailsjs/go/models.ts` | Wails model 增加 `firstMessage` |
| `docs/workspace-architecture.md` | 文档同步 `SessionDTO.firstMessage` |

---

## 5. 验证状态

已执行并通过：

```text
cd desktop/frontend; npm run build
cd D:\works\study\mimo cli; go test ./desktop/... ./internal/session/... -count=1
```

已知构建输出仍有历史警告：
- `syntax-highlighter` chunk 约 660KB，大于 Vite 默认建议值。
- Node ESM warning 仍存在，不影响本次功能验证。

---

## 6. 用户已手测/反馈状态

- 会话后台流式不再串到当前会话。
- 进入长历史会话已能直接到底部。
- 代码块黑色背景问题已针对截图修复。
- 文本选择/添加到对话已多轮修正：使用原生按钮 + 自绘高亮 + 冻结 hover 操作栏。

---

## 7. 已知潜在问题

1. **Vite chunk 大小警告**：`syntax-highlighter` 仍是最大包，建议后续动态导入或拆分。
2. **项目历史 TS 类型问题**：全量 `tsc` 仍可能存在 `window.go` 等历史声明问题；本轮以 `vite build` 验证为准。
3. **文本选择交互依赖 WebView selection 行为**：当前已用自绘高亮兜底，但仍建议继续手测复杂 Markdown/代码块中的选择体验。
4. **流式后台保存依赖前端 `SaveSessionFromFrontend`**：已加前端快照兜底，但后续可考虑后端直接按 sessionID 保存，进一步降低前端职责。

---

## 8. 建议下一步

### 高优先级
1. **Skill 系统对接**：Dream/Distill 产出落盘为 `.mimo/skills/` 下的 skill 文件。
2. **输入历史记录完善**：当前 `ChatInput` 有基础上下键历史，可继续做跨会话/持久化历史。
3. **Token 预算警告**：状态栏提示上下文接近上限。

### 中优先级
4. **Vite 构建优化**：`syntax-highlighter` 动态导入或 manualChunks。
5. **MCP 工具在 actor 可用性验证**：实际跑一轮 actor 工具调用。
6. **清理历史 TS 类型错误**：补 `window.go` / Wails runtime 类型声明。

### 低优先级
7. **PDF 预览**：FilePreviewModal 支持 PDF 渲染。
8. **代码块体验增强**：显示语言标签、复制按钮常显开关、代码换行切换。
