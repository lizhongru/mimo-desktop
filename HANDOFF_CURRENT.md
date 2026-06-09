# MiMo Desktop — 上下文交接文档

> 日期：2026-06-10
> 版本：v0.5.0-dev
> 最新 commit：eae742e (master)
> 技术栈：Wails v2 + Go 1.26 + React 19 + TypeScript + Zustand + Tailwind CSS
> 验证状态：npx tsc --noEmit 零错误 | go vet -tags wails 零错误
> Git 远程：origin -> https://github.com/lizhongru/mimo-desktop.git (尚未 push)

---

## 一、项目概况

MiMo Desktop 是基于 Wails 的 AI 聊天桌面客户端。

- **入口**: wails_main.go
- **前端**: desktop/frontend/（React 19 + Zustand + Tailwind CSS）
- **后端**: desktop/（Go Wails binding）+ internal/（Agent + LLM + 配置）
- **配置文件**: `~/.mimo/config.yaml`
- **启动**: `wails dev -tags wails`（需 MSYS2 gcc）
- **TS 检查**: `cd desktop/frontend && npx tsc --noEmit`
- **Go 检查**: `go vet -tags wails ./...`

---

## 二、v0.5.0 已完成的工作

### 2.1 ModelReasoningPicker 重新设计（ChatInput + WelcomeView 同步）
- **触发按钮**：显示 model 显示名（如 `mimo-v2.5-pro`），激活态 accent 高亮边框 + 箭头旋转动画
- **主面板（第一层）**：推理 segmented control（⚡低 / ⚖️中 / 🧠高，带滑动指示器动画 200ms ease-out）+ 分割线 + 当前模型行（可点击展开第二层）
- **模型面板（第二层）**：从主面板侧边滑出（slide-right），模型列表 + 勾选标记
- **智能方向**：根据触发按钮右侧剩余空间（< 460px）自动判断面板在左侧还是右侧展开
- **面板宽度**：主面板 260px，模型面板 180px

### 2.2 模型 key/value 映射修复（settingsStore）
- **根因**：`currentModel` 存的是 model 显示名（如 `mimo-v2.5-pro`），`models` 数组存的是 config key（如 `mimo`），导致勾选标记永远不匹配，SetDefaultModel 发送错误值
- **修复**：新增 `currentModelKey` 字段 + `_modelsMap` 内部映射；`setCurrentModel(key)` 接受 key，内部解析显示名存 `currentModel`，发送 key 到后端

### 2.3 LeftSidebar UserProfileFooter
- 底部 Footer 改造为可展开菜单
- **设置** / **快捷键** / **帮助日志** / **关于** 四个入口
- **语言切换**：鼠标悬停展开右侧子面板
- **主题切换**：暗色/亮色 toggle 开关（带动画）
- 接入 ShortcutsPanel / HelpLogPanel / AboutPanel 弹窗

### 2.4 WelcomeView 工作区选择
- WorkspacePicker 组件：点击展开下拉（最近工作区 + 浏览文件夹）
- 调用 `SelectDirectory` 后端接口

### 2.5 后端修复（v0.3.0 已有）
- ReasoningLevel 生效：全部 6 处 ChatRequest 传递 ReasoningEffort
- 模型获取超时：ListRemoteModelsWithConfig 加 15s context.WithTimeout
- SaveSession 错误处理：os.Getwd() 不再忽略错误
- confirmChan 超时保护：60s select timeout

### 2.6 前端核心修复（v0.3.0 已有）
- useAnimatedOpen ref 追踪状态转换
- modal-overlay/modal-dialog CSS transition
- 主题防闪烁（index.html 阻塞脚本）
- 主题切换动画（theme-transition.ts 圆形 clip-path 扩散）
- View Transitions + prefers-reduced-motion

### 2.7 新建组件/工具
- `src/components/common/ShortcutsPanel.tsx` — 快捷键面板
- `src/components/common/HelpLogPanel.tsx` — 帮助日志面板
- `src/components/common/AboutPanel.tsx` — 关于面板
- `src/components/settings/ModelManager.tsx` — 模型管理（完整 CRUD + 远程模型获取）
- `src/lib/theme-transition.ts` — 主题切换动画工具
- `src/lib/useAnimatedOpen.ts` — 动画 hook

### 2.8 i18n
- 200+ 条翻译 key，覆盖 zh/en
- 新增：shortcuts, help_log, about, reasoning_*, permission_label, attach_file, drop_to_add 等

---

## 三、当前已知问题（需要实测确认）

### 3.1 未测试功能
- **拖拽上传**：AppLayout overlay 提示文案已修复（"松手即可添加"），但 onDragLeave 逻辑可能在边界情况下提前关闭 overlay
- **文件上传实际发送**：附件只是前端预览，sendMessage 调用时没有把附件数据传给后端，**需要实现后端接收**
- **Agent 工具闭环**：桌面端实际运行验证未完成
- **MCP 工具调用**：未测试

### 3.2 前端潜在风险
- **ModelManager modelSearch 状态**：resetForm 时被重置但 remoteModels 不会重置
- **上下文压缩**：进度更新只有开始和完成，缺少中间状态

### 3.3 后端潜在风险
- **registerConfirmCallback**：guardrail 回调本身没有超时保护（已有 60s channel 超时，但回调执行仍可能阻塞）
- **工具调用结果渲染**：ToolCallCard 的 result 只是纯文本 pre，没有语法高亮

---

## 四、下一步建议

### 高优先级
1. **文件上传后端对接**：附件数据需要随 sendMessage 传递到后端 Agent
2. **实际运行测试**：启动 `wails dev -tags wails` 测试完整流程（模型切换、推理级别、工具调用）
3. **Agent 工具闭环验证**：发送需要工具调用的消息验证端到端

### 中优先级
4. **MCP 工具测试**：配置 MCP 服务器验证
5. **代码分割**：设置页/工具面板懒加载
6. **错误边界**：添加 React ErrorBoundary
7. **MarkdownRenderer 代码块复制按钮**：SyntaxHighlighter 没有复制功能

### 低优先级
8. **上下文压缩中间进度**
9. **ToolCallCard 结果语法高亮**
10. **移动端适配（如果需要）**

---

## 五、关键文件清单

| 文件 | 职责 |
|------|------|
| wails_main.go | 桌面端入口 |
| desktop/app.go | App 主结构体 + 初始化 |
| desktop/app_chat.go | 聊天/导出/压缩/GetModelName |
| desktop/app_config.go | 配置读写（SetTheme/SetLanguage/SetDefaultModel/ListRemoteModels 等） |
| desktop/app_session.go | 会话管理 |
| desktop/app_tools.go | 工具/MCP 查询 |
| internal/config/schema.go | 配置结构体（DefaultModel, Models map） |
| internal/config/config.go | 配置加载/保存/合并 |
| internal/llm/gateway.go | LLM Gateway（SetCurrentModel/GetCurrentModel） |
| internal/llm/openai.go | OpenAI provider |
| internal/agent/agent.go | Agent 核心 |
| src/App.tsx | React 入口 + initFromConfig |
| src/components/layout/AppLayout.tsx | 主布局 + 拖拽 overlay |
| src/components/layout/LeftSidebar.tsx | 侧栏 + UserProfileFooter |
| src/components/chat/ChatInput.tsx | 输入框 + ModelReasoningPicker + 文件上传 + Dropdown |
| src/components/chat/StatusBar.tsx | 状态栏 |
| src/components/welcome/WelcomeView.tsx | 欢迎页 + WorkspacePicker + 文件上传 |
| src/components/settings/SettingsPage.tsx | 设置页（modal-overlay） |
| src/components/settings/ModelManager.tsx | 模型管理（CRUD + 远程获取） |
| src/components/common/ShortcutsPanel.tsx | 快捷键面板 |
| src/components/common/HelpLogPanel.tsx | 帮助日志面板 |
| src/components/common/AboutPanel.tsx | 关于面板 |
| src/components/confirm/ConfirmDialog.tsx | 安全确认对话框 |
| src/components/chat/MessageBubble.tsx | 消息气泡 + 右键菜单 |
| src/stores/settingsStore.ts | 设置状态（theme/language/currentModel/currentModelKey/reasoning） |
| src/stores/chatStore.ts | 聊天状态 + deleteMessage |
| src/stores/sessionStore.ts | 会话状态 |
| src/lib/i18n.ts | 国际化（200+ 条翻译） |
| src/lib/useAnimatedOpen.ts | 动画 hook |
| src/lib/theme-transition.ts | 主题切换动画 |
| src/styles/globals.css | CSS 变量 + modal + View Transitions |
| tailwind.config.ts | Tailwind 动画配置 |
| index.html | 阻塞脚本防主题闪烁 |

---

## 六、环境与启动

```powershell
$env:PATH = "C:\msys64\mingw64\bin;..." # MSYS2 gcc
cd "D:\works\study\mimo cli"
wails dev -tags wails          # 开发模式
wails build -tags wails        # 生产打包
cd desktop/frontend && npx tsc --noEmit  # TS 检查
go vet -tags wails ./...       # Go 检查
git push origin master         # 推送（当前未推送）
```

---

## 七、可清理的无用文件/目录

以下目录在 .gitignore 中，不会被提交，但占用磁盘空间（约 1.1GB）：

| 目录 | 大小 | 说明 |
|------|------|------|
| .gocache/ | ~483 MB | Go 编译缓存 |
| .gomod/ | ~583 MB | Go 模块缓存 |
| bin/ | ~17 MB | 编译产物 (mimo-desktop.exe) |
| build/bin/ | ~45 MB | Wails 构建产物 |
| .mimo/backups/ | < 1 MB | 配置备份 |

清理命令：`Remove-Item -Recurse -Force bin,build,.gocache,.gomod,.mimo/backups`
