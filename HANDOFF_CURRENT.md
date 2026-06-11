# MiMo Desktop — 上下文交接文档

> 日期：2026-06-10
> 版本：v0.6.0-dev（未提交，基于 eae742e）
> 技术栈：Wails v2 + Go 1.26 + React 19 + TypeScript + Zustand + Tailwind CSS
> 验证状态：npx tsc --noEmit 零错误 | go vet -tags wails 零错误
> Git 远程：origin -> https://github.com/lizhongru/mimo-desktop.git（未 push）
> 未提交变更：18 个文件，+296/-164 行

---

## 一、本次完成的工作（v0.6.0）

### 1.1 文件上传后端对接（全链路打通）
- internal/llm/message.go — 新增 Attachment 结构体和 Message.Attachments 字段
- internal/llm/openai.go — openAIMessage.Content 改为 json.RawMessage；新增 uildOpenAIContent() 将附件转为 OpenAI 多模态 content 数组（图片→image_url，文件→文本描述）；response 解析兼容 content 数组
- internal/agent/agent.go — Chat() 和 ChatStream() 签名新增 ttachments []llm.Attachment 参数
- desktop/app_chat.go — SendMessage() 新增 ttachmentsJSON string 参数，JSON 反序列化后传给 Agent
- 前端 ChatInput / WelcomeView / AppLayout / App.tsx — onSend 类型统一改为 (message, attachments?) 签名，附件 JSON 序列化后传给 Wails

### 1.2 拖拽上传优化
- **去掉全屏 overlay** — AppLayout 的全屏拖拽 overlay（dragenter/leave 计数器）在 WebView2 复杂子元素下不稳定，彻底移除
- **输入框区域拖拽** — ChatInput / WelcomeView 各自的输入框区域独立处理拖拽，用 ref 计数器精确控制
- **拖拽高亮** — 拖入时输入框 accent 边框 + 微光 + "松手即可添加"提示
- **自定义光标** — globals.css 新增 .drag-active（暗色文件图标）和 .drag-drop-zone（accent 色文件图标），用 SVG data URI
- **外层拦截** — AppLayout 外层 div 加 onDragOver={preventDefault} + onDrop={preventDefault}，阻止 WebView2 默认打开文件行为
- **子组件去掉 stopPropagation** — ChatInput / WelcomeView 的 handleDrop / handleDragOver 不再 stopPropagation，让事件正常冒泡

### 1.3 ModelReasoningPicker 重新设计
- **主面板（点击触发）**：推理级别 segmented control（⚡低/⚖️中/🧠高）+ 当前模型行
- **模型列表（点击展开）**：点击当前模型行 → 模型列表从左侧滑出，只显示显示名，不显示 key
- **hover 改 click**：hover 展开在 WebView2 中不可靠，改为点击切换
- **挂载时刷新**：useEffect(() => refreshModels(), []) 确保拿到最新模型列表
- ChatInput + WelcomeView 两处同步

### 1.4 设置页模型管理同步
- settingsStore.ts 新增 efreshModels() — 从后端 GetConfig 重新拉取模型列表，更新 models、_modelsMap、currentModel/currentModelKey
- SettingsPage.tsx — 在 onAdd / onUpdate / onRemove / onSetDefault 4 个操作成功后都调用 efreshModels()

### 1.5 侧栏会话分组
- 三组分组：**当前工作区** / **"对话"（无工作区）** / **其他工作区**
- sessionsWithWorkspace Set 标记：用户主动选了工作区的 session id 才归入工作区组
- 新建对话时清除 selectedWorkspace，确保新对话在"对话"组
- 三组都可以独立展开/收起

### 1.6 UserProfileFooter 改造
- 去掉工作区路径显示，改为 MiMo User 占位 + "点击打开设置"
- 去掉 workingDir 参数依赖，为后续用户登录留位置

### 1.7 外层拖拽光标
- 拖拽到非输入框区域时 drag-active CSS 类覆盖光标为暗色文件图标（不显示"复制"）
- 输入框区域用 drag-drop-zone 显示 accent 色文件图标

---

## 二、变更文件清单

| 文件 | 变更内容 |
|------|----------|
| internal/llm/message.go | +Attachment 类型，Message.Attachments 字段 |
| internal/llm/openai.go | Content→json.RawMessage，buildOpenAIContent()，response 解析兼容 |
| internal/agent/agent.go | Chat/ChatStream 签名加 attachments |
| desktop/app_chat.go | SendMessage 加 attachmentsJSON 参数 |
| desktop/app_session.go | SaveSessionFromFrontend 加 workingDir 参数 |
| desktop/frontend/src/App.tsx | SendMessage 声明更新，handleSend 传附件，handleSelectWorkspace 标记 session，handleNewChat 清除 selectedWorkspace |
| desktop/frontend/src/components/chat/ChatInput.tsx | ModelReasoningPicker 重写（click 展开），onSend 传附件，输入框拖拽高亮 |
| desktop/frontend/src/components/layout/AppLayout.tsx | 去掉全屏拖拽 overlay，外层 preventDefault + 自定义光标类 |
| desktop/frontend/src/components/layout/LeftSidebar.tsx | 三组分组 + sessionsWithWorkspace，UserProfileFooter 改用户信息占位 |
| desktop/frontend/src/components/settings/SettingsPage.tsx | onAdd/onUpdate/onRemove/onSetDefault 后调 refreshModels |
| desktop/frontend/src/components/welcome/WelcomeView.tsx | ModelReasoningPicker 同步，onSend 传附件，输入框拖拽高亮 |
| desktop/frontend/src/hooks/useAgent.ts | SaveSessionFromFrontend 传 selectedWorkspace |
| desktop/frontend/src/lib/i18n.ts | +search_models, no_models_found, back, conversations, click_to_settings |
| desktop/frontend/src/stores/sessionStore.ts | +selectedWorkspace, sessionsWithWorkspace, markSessionHasWorkspace |
| desktop/frontend/src/stores/settingsStore.ts | +refreshModels() |
| desktop/frontend/src/styles/globals.css | +drag-active, drag-drop-zone 自定义光标 |
| desktop/frontend/src/wails/wailsjs/go/desktop/App.d.ts | 类型声明同步 |
| desktop/frontend/src/wails/wailsjs/go/desktop/App.js | JS 绑定同步 |

---

## 三、当前已知问题

### 3.1 未测试功能
- **文件上传实际发送**：附件数据已能传到后端 Agent，但未实际测试 LLM 多模态请求是否正确（需要支持 vision 的模型）
- **Agent 工具闭环**：桌面端实际运行验证未完成
- **MCP 工具调用**：未测试
- **拖拽上传实际效果**：自定义光标 SVG 在不同 DPI 下可能需要调整

### 3.2 潜在风险
- **_modelsMap 用 s any 存储**：不在 Zustand 接口定义中，依赖 getState() 读取，Zustand 订阅机制不覆盖此字段
- **CRLF 行尾问题**：Windows 环境下 Go 文件和 TS 文件行尾不一致，导致 String.Replace 经常静默失败，后续修改文件建议用行操作或 [System.IO.File]::ReadAllText + 确认替换结果
- **selectedWorkspace 没持久化**：刷新页面后丢失，session 重新归组可能不一致（但 ListSessions 拉取后 sessionsWithWorkspace 也丢失）
- **sessionsWithWorkspace 没持久化**：同上，页面刷新后 Set 为空，所有 session 会被归入"对话"组（除了有后端 workingDir 的）
- **UserProfileFooter 的 MiMo User 是硬编码**：后续需要接入真实用户数据
- **ModelReasoningPicker modelsMap 一次性读取**：用 getState() 而非订阅，如果 store 异步更新了 _modelsMap，组件不会自动重渲染（但 models 数组的订阅会触发重渲染，间接刷新）

### 3.3 代码质量问题
- **WelcomeView 和 ChatInput 各有一份 ModelReasoningPicker 副本**：应提取为共享组件
- **WelcomeView 和 ChatInput 各有一份 MiniDropdown / Dropdown 副本**：同上
- **globals.css 里 	("conversations") 在 getDirName 中调用**：getDirName 是纯函数但依赖 i18n，如果语言切换时组件没重渲染，显示可能不一致

---

## 四、下一步建议

### 高优先级
1. **持久化 sessionsWithWorkspace**：存入 localStorage 或后端，避免刷新后分组丢失
2. **Agent 工具闭环验证**：发送需要工具调用的消息验证端到端
3. **文件上传实际测试**：用支持 vision 的模型测试图片上传

### 中优先级
4. **提取共享组件**：ModelReasoningPicker、Dropdown、MiniDropdown 提取为独立组件
5. **MCP 工具测试**
6. **代码分割**：设置页/工具面板懒加载
7. **React ErrorBoundary**
8. **MarkdownRenderer 代码块复制按钮**

### 低优先级
9. **上下文压缩中间进度**
10. **ToolCallCard 结果语法高亮**
11. **用户登录/信息存储**

---

## 五、环境与启动

```powershell
cd "D:\works\study\mimo cli"
wails dev -tags wails          # 开发模式（需重启加载 Go 变更）
cd desktop/frontend && npx tsc --noEmit  # TS 检查
go vet -tags wails ./...       # Go 检查
```