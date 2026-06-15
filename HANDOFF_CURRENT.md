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
84bf652 feat: 消息操作栏增强、侧边栏流式指示器、会话切换bug修复
156a930 feat: actor streaming output
3c4eefc feat: task panel event timeline on expand
cf713b8 perf: code splitting — main bundle 1.2MB → 460KB
3277501 feat: file preview modal, JSON tree, code block modernization
86f2302 docs: update HANDOFF for new session
aebc0c5 chore: regenerate wails bindings for task methods
98de7ae feat: tree-style task IDs, rename/archive/progress
76e0720 feat: memory config now affects runtime behavior
e5ad79a feat: real LLM execution for sub-agents (actors)
```

---

## 3. 本轮完成的功能（84bf652）

### 3.1 消息操作栏增强
- 右键菜单改为底部悬浮操作栏（胶囊型，hover 显示）
- 复制按钮：点击后绿色 Check 图标 + "Copied" 文字，1.6s 自动恢复
- 重新生成按钮：点击后旋转动画 + "Regenerating..." 提示
- 用户消息只保留复制，助手消息保留复制 + 重新生成
- 统一使用 CSS 变量适配亮暗主题

### 3.2 侧边栏流式指示器
- 正在生成的会话，左上角消息图标变为旋转 Loader2（主题色 `--color-accent`）
- 对话结束后自动恢复为普通 MessageSquare 图标
- 依赖 `sessionStore.streamingSessionId` 驱动
- `App.tsx` 中 `handleSend` 发送时设置 `setStreamingSessionId`
- `useAgent.ts` 中 `CHAT_DONE/ERROR/CANCELLED` 清除 `streamingSessionId`

### 3.3 会话标题规则
- 标题以第一次用户提问为准，不随后续对话变化
- `sessionStore` 新增 `firstMessage` 字段
- 新建会话时同时写入 `firstMessage` 和 `lastMessage`
- `setSessions` 和 `addSession` 自动补全 `firstMessage`

### 3.4 思考中组件灵动化
- 实时思考时：Brain → Sparkles 星光图标 + 脉冲呼吸光圈 + 自身闪烁
- 动态省略号：`思考中...` 循环跳动（500ms 间隔）
- 无内容时也显示"思考中..."动画
- `ThinkingBlock` 新增 `isLive` prop
- `MessageList` 流式区域使用 `<ThinkingBlock isLive />`

### 3.5 会话切换 Bug 修复
- **根因**：流式回调不检查当前 UI 展示的会话，切换后继续往新会话写
- **修复**：
  - `chatStore` 新增 `activeSessionId`，发送消息时记录
  - `useAgent.ts` 的 DELTA/THINKING/TOOL_CALL/TOOL_RESULT 加会话 ID 校验
  - `CHAT_DONE` 完成时如果会话已切换，静默保存到原会话
  - `CHAT_ERROR` / `CHAT_CANCELLED` 同理处理
  - 删除 `App.tsx` 中重复的 `window.runtime.EventsOn` 事件监听
  - `handleLoadSession` 切换会话时，如果另一个会话正在流式，不清空消息
  - `resetStreamState` 同时清除 `activeSessionId`

### 3.6 ToolCallCard JSON 解析兜底
- 加固 `parseArgs` 函数，处理空字符串、非对象 JSON、数组等情况

---

## 4. 修改的文件清单

| 文件 | 改动说明 |
|------|---------|
| `desktop/frontend/src/App.tsx` | 删除重复事件监听、handleLoadSession 流式保护、setStreamingSessionId |
| `desktop/frontend/src/components/chat/MessageBubble.tsx` | 底部操作栏、复制/重新生成反馈、用户消息只复制 |
| `desktop/frontend/src/components/chat/MessageList.tsx` | 主题适配、去掉固定显示逻辑 |
| `desktop/frontend/src/components/chat/ThinkingBlock.tsx` | 灵动思考动画、isLive 模式 |
| `desktop/frontend/src/components/chat/ToolCallCard.tsx` | JSON 解析兜底 |
| `desktop/frontend/src/components/layout/LeftSidebar.tsx` | 侧边栏流式指示器、firstMessage 标题 |
| `desktop/frontend/src/hooks/useAgent.ts` | 会话 ID 校验、CHAT_DONE/ERROR/CANCELLED 处理 |
| `desktop/frontend/src/stores/chatStore.ts` | activeSessionId 追踪、resetStreamState 清理 |
| `desktop/frontend/src/stores/sessionStore.ts` | firstMessage 字段、setSessions 补全 |

---

## 5. 验证状态

- `cd desktop/frontend; vite build` — ✓ built in ~5s，主包 468KB
- 项目有历史 TS 类型错误（`window.go` 未声明等），不影响构建和运行

---

## 6. 已知潜在问题

1. **项目有大量历史 TS 类型错误**（`window.go` 未声明、Wails EventCallback 类型不匹配等），不是本次引入，`tsc` 检查报 120+ 错误
2. **Vite 构建有 chunk 大小警告**（syntax-highlighter 660KB、index 468KB），建议 manualChunks 优化
3. **`firstMessage` 只在前端 store 里**，没有持久化到后端会话数据，重启后旧会话的 firstMessage 会回退为 lastMessage
4. **流式切换场景下原会话保存依赖 `SaveSessionFromFrontend`**，如果后端有并发问题可能丢失数据
5. **`App.tsx` 中的 `handleDone` 仍然是死代码**（事件监听已删除），可以清理
6. **`workspaceIdFromDir` 函数已定义但未使用**，可以清理

---

## 7. 建议下一步（按优先级）

### 高优先级
1. **手动测试流式切换场景**：发送消息 → 切换会话 → 确认原会话内容正确保存、新会话不受影响
2. **清理 `App.tsx` 中的死代码**（handleDone、workspaceIdFromDir 等）
3. **`firstMessage` 持久化到后端**，确保重启后标题不丢
4. **Skill 系统对接**：Dream/Distill 产出应落盘为 `.mimo/skills/` 下的 skill 文件

### 中优先级
5. **输入历史记录**：上下箭头翻阅历史消息
6. **Token 预算警告**：状态栏变色提示上下文接近上限
7. **MCP 工具在 actor 可用**：实际测试验证

### 低优先级
8. **Vite 构建优化**：syntax-highlighter 动态导入、vendor 拆分
9. **清理历史 TS 类型错误**（工作量较大）
10. **PDF 预览**：FilePreviewModal 支持 PDF 渲染