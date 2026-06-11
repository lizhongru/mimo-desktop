# MiMo Desktop — 当前上下文交接

> 日期：2026-06-11
> 工作区：`D:\works\study\mimo cli`
> 状态：未提交；仓库里有大量历史 dirty 变更，下一位 agent 不要随手 reset/revert。

---

## 1. 当前任务背景

用户最初在排查「会话按工作区/项目分区」相关 bug。

已经确认过的一轮用户测试结果：

1. 不关联项目时，对话进入 `DEFAULT`。
2. 创建 A 项目后，对话进入 A 项目。
3. 创建 B 项目后，对话进入 B 项目。

这说明「新建会话归属分区」主逻辑已经基本正常。

之后用户又反馈过这些问题：

- 点击项目/DEFAULT 中的对话，无法查看详情继续对话。
- 点击项目 A 展开时，项目 B 也同时展开。
- 希望 `DEFAULT` 放在项目下方，并改为中文显示。
- 选择工作区后，让模型查看当前工作区内容时，读取的文件夹路径不对。
- 最新需求：欢迎页模型选择器和聊天页不一致，需要统一；聊天页需要支持手动选择文件上传或选择文件夹。

---

## 2. 最新已完成工作

### 2.1 欢迎页/聊天页模型选择器统一

新增共享组件：

- `desktop/frontend/src/components/chat/ModelReasoningPicker.tsx`

当前 `WelcomeView` 和 `ChatInput` 都使用这个共享组件，模型选择和推理强度切换行为一致。

### 2.2 前端文件/文件夹附件能力

新增附件工具：

- `desktop/frontend/src/lib/attachments.ts`

修改：

- `desktop/frontend/src/components/chat/ChatInput.tsx`
- `desktop/frontend/src/components/welcome/WelcomeView.tsx`

当前能力：

- 聊天页支持手动选择文件。
- 聊天页支持选择文件夹。
- 欢迎页也保持相同入口。
- 文件夹选择通过 `webkitdirectory/directory`。
- 文件夹内文件会保留 `webkitRelativePath` 相对路径。
- 普通文件选择不再限制扩展名，代码文件也能直接选。
- 附件在输入框上方以 chip 形式展示，可移除。
- 文件/文件夹按钮增加了 `aria-label`。

### 2.3 i18n 文案补齐

修改：

- `desktop/frontend/src/lib/i18n.ts`

新增/修正：

- `attach_file`: 中文为「添加文件」
- `attach_folder`: 中文为「添加文件夹」
- 还补了部分之前工作区/侧栏相关文案 key。

### 2.4 后端附件内容链路补强

修改：

- `internal/llm/openai.go`
- `internal/llm/anthropic.go`

新增测试：

- `internal/llm/openai_test.go`
- `internal/llm/anthropic_test.go`

当前能力：

- OpenAI-compatible 路径：
  - 图片附件转为 `image_url`。
  - 文本/代码类附件会从 `dataUrl` 解码，把真实文件内容放入模型输入。
  - 非文本/非图片文件只给模型文件名和 MIME 类型提示。
- Anthropic-compatible 路径：
  - 图片附件转为 Anthropic image/source block。
  - 文本/代码类附件同样会解码并放入 text block。
- 文本附件内容有最大截断：`200000` 字符，避免超大文件直接打爆上下文。

### 2.5 Wails 绑定/前后端签名

已有链路：

- `desktop/frontend/src/App.tsx`
  - `handleSend(message, attachments?)`
  - 附件序列化为 JSON 后传给 `SendMessage`
- `desktop/app_chat.go`
  - `SendMessage(message, attachmentsJSON)`
  - 解析 JSON 后传给 Agent
- `internal/agent/agent.go`
  - `Chat/ChatStream` 已接收 `[]llm.Attachment`
- `internal/llm/message.go`
  - `Message.Attachments`
  - `Attachment{Name, Type, DataURL}`

---

## 3. 已验证

最后一轮验证已通过：

```powershell
cd "D:\works\study\mimo cli\desktop\frontend"
npx tsc --noEmit
npm run build
```

```powershell
cd "D:\works\study\mimo cli"
go test ./internal/llm ./internal/agent ./desktop
git diff --check
```

结果：

- TypeScript 类型检查通过。
- Vite 生产构建通过。
- 相关 Go 测试通过。
- `git diff --check` 通过。
- `npm run build` 仍有既有的大 chunk warning 和 Node ESM warning，不是阻断错误。

浏览器 DOM 验证也做过，目标：`http://127.0.0.1:5173/`

确认：

- 欢迎页存在「添加文件」「添加文件夹」「当前模型」。
- 文件 input 不再有 `accept` 限制。
- 文件夹 input 有 `webkitdirectory` 和 `directory`。
- 模型选择面板能打开，能看到低/中/高推理。

注意：浏览器验证是 Vite 前端 DOM 验证，完整 Wails Go bridge 行为仍需要桌面应用里测。

---

## 4. 下一步建议

优先让用户在桌面 Wails 应用里实测：

1. 欢迎页模型选择器是否和聊天页一致。
2. 欢迎页切换模型后，进入聊天页是否保持同一模型。
3. 聊天页手动选择单个文件、多个文件是否能正常作为附件发送。
4. 聊天页选择文件夹后，模型是否能看到文件夹内文本/代码文件内容和相对路径。
5. 上传 `.md/.txt/.json/.ts/.go` 等文本/代码文件，让模型复述内容，确认不是只看到文件名。
6. 上传图片，确认支持 vision 的模型能看到图片。
7. 切换不同模型供应商，确认 OpenAI-compatible / Anthropic-compatible 的附件行为一致。
8. 回归工作区路径 bug：选择工作区后，让模型查看当前工作区文件，确认读取的是用户选择的目录。

如果用户测试失败，优先检查：

- `desktop/app_chat.go` 是否把当前 session 的 working dir 写入 context。
- `internal/tools/workdir.go` 和各工具是否从 context 读取工作目录。
- `App.tsx` 的 `selectedWorkspace` / `currentSessionId` 是否在发送前正确绑定。
- 左侧栏加载 session 后是否恢复 `workspaceId` 和消息。

---

## 5. 潜在问题/风险

### 5.1 仓库状态

- 当前工作区有很多历史未提交/未跟踪文件。
- 不要使用 `git reset --hard`、`git checkout --` 之类命令。
- 不要 revert 与当前任务无关的文件。

### 5.2 文件上传限制

- 文本/代码文件会解码进入上下文。
- 图片会作为图片附件传给模型。
- PDF、压缩包、Office 文档等二进制文件目前不会解析内容，只会传文件名和类型提示。
- 文件夹如果很大，前端会把选中的文件都读成 data URL，可能导致内存和请求体很大；后续可以加数量/大小限制和提示。

### 5.3 UI/UX

- 文件/文件夹按钮目前是 icon-only + title/aria-label。
- 上传后的消息列表目前不一定会把附件显示在已发送消息气泡里，主要是在发送前展示 chip。
- 如果用户期望「只上传附件不输入文字也能发送」，当前发送按钮仍主要依赖文本输入，需要再改。

### 5.4 模型选择器

- `ModelReasoningPicker` 依赖 `settingsStore.refreshModels()` 拉取模型。
- `_modelsMap` 仍是 Zustand store 里的非接口字段，用 `as any` 读取；可工作，但类型上不够干净。

### 5.5 Wails/Vite 验证差异

- Vite 浏览器里没有真实 `window.go`，只能做前端 DOM/交互结构验证。
- 真正的 session 创建、消息发送、工作区绑定、工具读目录，必须在 Wails 桌面应用里测试。

---

## 6. 关键文件清单

前端：

- `desktop/frontend/src/components/chat/ModelReasoningPicker.tsx`
- `desktop/frontend/src/lib/attachments.ts`
- `desktop/frontend/src/components/chat/ChatInput.tsx`
- `desktop/frontend/src/components/welcome/WelcomeView.tsx`
- `desktop/frontend/src/App.tsx`
- `desktop/frontend/src/components/layout/AppLayout.tsx`
- `desktop/frontend/src/components/layout/LeftSidebar.tsx`
- `desktop/frontend/src/stores/sessionStore.ts`
- `desktop/frontend/src/stores/settingsStore.ts`
- `desktop/frontend/src/lib/i18n.ts`

后端：

- `desktop/app_chat.go`
- `desktop/app_session.go`
- `internal/agent/agent.go`
- `internal/llm/message.go`
- `internal/llm/openai.go`
- `internal/llm/anthropic.go`
- `internal/tools/workdir.go`
- `internal/tools/*`

测试：

- `internal/llm/openai_test.go`
- `internal/llm/anthropic_test.go`
- `internal/session/store_test.go`
- `internal/session/workspace_test.go`
- `internal/tools/workdir_test.go`

---

## 7. 启动/检查命令

```powershell
cd "D:\works\study\mimo cli"
wails dev -tags wails
```

```powershell
cd "D:\works\study\mimo cli\desktop\frontend"
npx tsc --noEmit
npm run build
```

```powershell
cd "D:\works\study\mimo cli"
go test ./internal/llm ./internal/agent ./desktop
git diff --check
```
