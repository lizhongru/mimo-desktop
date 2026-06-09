# MiMo Desktop — 上下文交接文档

> 日期：2026-06-09
> 版本：v0.5.0-dev
> 技术栈：Wails v2 + Go 1.26 + React 19 + TypeScript + Zustand + Tailwind CSS
> 验证状态：npx tsc --noEmit ✅ 零错误 | go vet -tags wails ✅ 零错误

---

## 一、项目概况

MiMo Desktop 是基于 Wails 的 AI 聊天桌面客户端。

- **入口**: wails_main.go
- **前端**: desktop/frontend/（React 19 + Zustand + Tailwind CSS）
- **后端**: desktop/（Go Wails binding）+ internal/（Agent + LLM + 配置）
- **配置文件**: `~/.mimo/config.yaml`
- **启动**: `wails dev -tags wails`（需 MSYS2 gcc）
- **TS 检查**: `cd desktop/frontend && npx tsc --noEmit`

---

## 二、本轮已完成的工作

### 2.1 后端修复

| 修改 | 文件 | 说明 |
|------|------|------|
| ReasoningLevel 生效 | internal/agent/agent.go | 全部 6 处 ChatRequest 传递 ReasoningEffort |
| 模型获取超时 | desktop/app_config.go | ListRemoteModelsWithConfig 加 15s context.WithTimeout |
| SaveSession 错误处理 | desktop/app_session.go | os.Getwd() 不再忽略错误 |
| confirmChan 超时保护 | desktop/app.go | 60s select timeout |

### 2.2 前端核心修复

| 修改 | 文件 | 说明 |
|------|------|------|
| useAnimatedOpen 初始值 | desktop/frontend/src/lib/useAnimatedOpen.ts | ref 追踪状态转换 |
| fade-in 动画 | tailwind.config.ts | translate(-50%) 改为 translateY |
| modal-overlay CSS | src/styles/globals.css | 新增 modal-overlay/modal-dialog CSS transition（用户参考样式） |
| View Transitions | src/styles/globals.css | vt-circle-reveal 关键帧 + prefers-reduced-motion |
| 文字色变量 | src/styles/globals.css | 暗色主题更亮（#ffffff），浅色主题更深（#0a0a12） |
| 主题防闪烁 | index.html | 阻塞脚本读 localStorage 同步设置 class |
| 主题切换动画 | src/lib/theme-transition.ts | **新建** 圆形 clip-path 扩散动画（VT API + rAF 降级） |

### 2.3 组件改造

| 组件 | 改动 |
|------|------|
| LeftSidebar.tsx | 新增 UserProfileFooter 子组件（菜单 + 面板）；Props 加 onExportSession；面板接入 ShortcutsPanel/HelpLogPanel/AboutPanel |
| SettingsPage.tsx | 改用 modal-overlay/modal-dialog 模式；移除 useAnimatedOpen JS 控制 |
| ToolsViewer.tsx | 同上改用 modal-overlay |
| ConfirmDialog.tsx | 同上改用 modal-overlay |
| ChatInput.tsx | ModelReasoningPicker 双层面板（推理+模型列表）；文件上传按钮 + 拖拽 + 附件预览；权限下拉用 permission_label |
| WelcomeView.tsx | 同步 ModelReasoningPicker 双层面板；文件上传按钮 + 附件预览 |
| StatusBar.tsx | 移除右下角模型名显示；移除工具调用状态显示 |
| AppLayout.tsx | 全界面拖拽覆盖层（"松手即可添加"） |
| MessageBubble.tsx | 右键菜单（复制/重新生成/删除） |
| App.tsx | regenerate 事件监听 |

### 2.4 i18n

新增 key：shortcuts, help_log, about, shortcut_*, help_log_title, about_*, copy_text, regenerate, delete_message, permission_label, attach_file, drop_to_add 等。

### 2.5 新建组件

- `src/components/common/ShortcutsPanel.tsx` — 快捷键面板
- `src/components/common/HelpLogPanel.tsx` — 帮助日志面板
- `src/components/common/AboutPanel.tsx` — 关于面板
- `src/lib/theme-transition.ts` — 主题切换动画工具

---

## 三、当前已知问题

### 3.1 需要实测确认

- **拖拽上传实际效果**：overlay 提示文案已修复（"松手即可添加"），但 AppLayout 的 onDragLeave 逻辑可能在某些边界情况下提前关闭 overlay，需要实测
- **模型名显示**：✅ 已修复 — 新增 currentModelKey 区分 key 和显示名，setCurrentModel 发送正确的 key 到后端，列表勾选标记已修复
- **文件上传实际发送**：附件只是前端预览，实际 sendMessage 调用时没有把附件数据传给后端，需要实现后端接收

### 3.2 Go 后端潜在风险

- **registerConfirmCallback**：guardrail 回调本身没有超时保护（已有 60s channel 超时，但回调执行仍可能阻塞）
- **Agent 工具闭环**：桌面端实际运行验证未完成
- **MCP 工具调用**：未测试
- **上下文压缩**：进度更新只有开始和完成，缺少中间状态

### 3.3 前端潜在风险

- **ModelManager 中 modelSearch 状态**：resetForm 时被重置但 remoteModels 不会重置
- **setCurrentModel 与 SetDefaultModel 的 key/value 不一致**：✅ 已修复 — 新增 currentModelKey + _modelsMap，setCurrentModel 接受 key 并自动解析显示名
- **拖拽文件到 WelcomeView**：AppLayout 的 overlay 只在 chat view（messages.length > 0）时显示，但已改为始终显示。不过 WelcomeView 内部没有 onDrop 处理器来接收文件

---

## 四、下一步建议

### 高优先级

1. ~~**修复模型 key/value 映射**~~ ✅ 已完成 — settingsStore 新增 currentModelKey + _modelsMap
2. **文件上传后端对接**：附件数据需要随 sendMessage 传递到后端 Agent
3. **拖拽到 WelcomeView**：WelcomeView 需要接收 AppLayout 传递的拖拽文件
4. **实际运行测试**：启动 wails dev 测试完整流程

### 中优先级

5. **Agent 工具闭环验证**：发送需要工具调用的消息
6. **MCP 工具测试**：配置 MCP 服务器验证
7. **代码分割**：设置页/工具面板懒加载
8. **错误边界**：添加 React ErrorBoundary

### 低优先级

9. **ModelManager 的 modelSearch 状态一致性**
10. **上下文压缩中间进度**
11. **移动端适配（如果需要）**

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
| src/components/chat/ChatInput.tsx | 输入框 + ModelReasoningPicker + 文件上传 |
| src/components/chat/StatusBar.tsx | 状态栏 |
| src/components/welcome/WelcomeView.tsx | 欢迎页 + 文件上传 |
| src/components/settings/SettingsPage.tsx | 设置页（modal-overlay） |
| src/components/settings/ModelManager.tsx | 模型管理 |
| src/components/common/ShortcutsPanel.tsx | 快捷键面板 |
| src/components/common/HelpLogPanel.tsx | 帮助日志面板 |
| src/components/common/AboutPanel.tsx | 关于面板 |
| src/components/confirm/ConfirmDialog.tsx | 安全确认对话框 |
| src/components/chat/MessageBubble.tsx | 消息气泡 + 右键菜单 |
| src/stores/settingsStore.ts | 设置状态（theme/language/model/reasoning） |
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
```
