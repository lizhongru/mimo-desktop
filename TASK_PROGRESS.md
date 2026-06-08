# MiMo CLI - 消息队列与 UI 优化任务记录

> 日期：2026-06-06
> 状态：进行中

---

## 一、已完成任务

### 1. 消息队列系统
- [x] `tuiModel` 新增 `messageQueue []string` 字段
- [x] busy 时 Enter 键消息入队，不再阻塞输入
- [x] `agentDoneMsg` / `agentErrMsg` 完成后自动发送队列首条消息
- [x] Esc 取消时清空队列
- [x] 队列消息在分隔线上方显示（带"排队中"标记）
- [x] 自动发送队列消息时保留当前输入框内容

### 2. 输入框行为
- [x] busy 时输入框保持可用（移除 `if m.busy { return m, nil }` 拦截）
- [x] 移除输入框下方的"处理中..."状态行

### 3. 取消与错误提示
- [x] Esc 取消后，goroutine 返回的 `context.Canceled` 错误不再泄露
- [x] 使用 `cancelPending` 标记，显示 `「问题...」已终止` 格式
- [x] `agentDoneMsg` 中也检查 `cancelPending`（修复竞态问题）

### 4. 滚动行为
- [x] 流式输出时默认跟随底部
- [x] 用户 PgUp / 鼠标滚轮上滚后停止跟随
- [x] 滚回底部后恢复跟随
- [x] 新消息发送时重置跟随状态

### 5. 用户昵称系统
- [x] Config 新增 `UserName` 字段（`~/.mimo/config.yaml`）
- [x] 首次启动引导用户设置昵称
- [x] 所有 `❯ You` 替换为 `❯ 昵称`
- [x] `/name` 命令运行中修改昵称

### 6. 状态栏优化
- [x] 显示格式：`● mimo   (↑ 1.2k · 5.4s)`
- [x] 从第一句对话开始一直显示
- [x] token 和时间实时更新（`spinnerFrame` 驱动 View 刷新）
- [x] 处理中时 `↑` 箭头高亮显示

### 7. 每条回复统计
- [x] 每条回复下方显示：`● ↑ prompt · ↓ completion · time: 5.4s · tools: 3`
- [x] prompt/completion 分开显示
- [x] token 为 0 时自动隐藏该项
- [x] 窗口大小改变时 rebuild 正确显示

### 8. Token 计算修复
- [x] `usageMsg` 改为替换而非累加（API 返回累计值）
- [x] 新增 `msgTokens` 字段用增量计算单轮消耗
- [x] `msgPromptTokens` / `msgCompletionTokens` 分开追踪
- [x] token 为 0 时用 `len(response)/2` 粗略估算

### 9. 压缩进度显示
- [x] 压缩过程中显示动态进度条 + 旋转动画 + 已用时
- [x] 状态栏显示 `compressing... Ns`
- [x] 压缩完成后显示详细统计（进度条 100%、token 节省量、百分比、耗时）
- [x] `/compress` 强制压缩，跳过 75% 阈值检查

### 10. 命令补全系统
- [x] 输入 `/` 触发命令补全
- [x] 上下键选择补全项
- [x] Tab 循环切换
- [x] Enter/y 确认选中

### 11. 命令历史持久化
- [x] 启动从 `~/.mimo/history` 加载
- [x] 退出自动保存，最多 100 条
- [x] 上下键回溯历史

### 12. 会话恢复
- [x] 退出时保存 JSON + Markdown 到 `~/.mimo/sessions/`
- [x] `/resume` 交互式选择历史会话（上下键、显示最后提问作为标题）
- [x] 恢复后 Agent 上下文完整还原，可继续对话
- [x] 恢复后继承 sessionId，继续对话更新同一文件
- [x] 自定义 JSON 序列化（chatMessage 未导出字段修复）

### 13. 多模型切换
- [x] `/model` 列出所有可用模型
- [x] `/model 名称` 切换模型，写入配置

### 14. 对话导出
- [x] `/export` 导出对话到 `~/.mimo/exports/export_时间戳.md`

### 15. Token 预算提示
- [x] 超过 96k（75% of 128k）时状态栏变色 + ⚠ 警告

### 16. 退出后会话保存
- [x] 退出时静默保存到文件
- [x] 不显示对话历史、不显示退出确认

### 17. UI 渲染
- [x] alt-screen 模式，UI 稳定不重复
- [x] 旋转动画 MiMo 前缀
- [x] 欢迎横幅显示版本和模型

### 18. 复制功能
- [x] `/copy` 复制最后一条回复
- [x] `/copy N` 复制第 N 条
- [x] `/copy all` 复制全部
- [x] `/copy user` 复制最后一条用户消息
- [x] PowerShell Set-Clipboard（UTF-8 兼容）

### 19. mimo run 增强
- [x] 管道输入：stdin 不是终端时自动读取
- [x] 输出格式化：`--output json` / `--output markdown`
- [x] 语义化退出码：0=成功, 1=错误, 2=需人工审核
- [x] JSON 输出包含任务、响应、耗时、token、退出码

### 20. AGENT.md 结构化解析
- [x] 解析 AGENT.md 的 ## 和 ### 标题分节
- [x] 自动识别：代理规则、禁止操作、编码规范、常用命令、技术栈、项目结构
- [x] 代理规则和禁止操作以醒目标题注入系统提示词
- [x] 支持 `.mimo/rules.md` 额外规则文件
- [x] Git 信息（分支、状态）自动注入上下文
- [x] 规则注入到每条用户消息（解决模型忽略系统提示词末尾规则的问题）

### 21. 多行输入
- [x] Shift+Enter 切换单行/多行输入模式
- [x] 多行模式下 Enter 添加换行，Ctrl+Enter 发送
- [x] 单行模式下 Enter 直接发送
- [x] 底部提示显示当前模式的快捷键

### 22. Git 工具输出格式化
- [x] `git_status`：分类显示 Staged/Unstaged/Untracked 文件
- [x] `git_diff`：添加 diffstat 摘要 + 内容截断（防上下文爆炸）
- [x] `git_log`：详细格式（hash | date | author | message）
- [x] `git_commit`：提交后显示新提交信息

### 23. 主题切换
- [x] dark/light 两套配色方案
- [x] `/theme` 命令查看/切换主题
- [x] 主题设置持久化到 `~/.mimo/config.yaml`
- [x] 启动时自动加载保存的主题
- [x] `/copy` 复制最后一条回复
- [x] `/copy N` 复制第 N 条
- [x] `/copy all` 复制全部
- [x] `/copy user` 复制最后一条用户消息
- [x] PowerShell Set-Clipboard（UTF-8 兼容）
- [x] 命令列表固定高度（5个）+ 上下滚动
- [x] 命令历史记录所有输入（包括 / 命令）
- [x] 主题选择改为上下键交互式选择

---

## 二、已修复 Bug（共 12 个）

| # | Bug | 文件 | 修复 |
|---|-----|------|------|
| 1 | Token 统计可能为 0 | `tui_model.go` | `len(response)/2` 粗略估算 |
| 2 | 排队消息 chatMessage 未记录 | `tui_model.go` | 入队时写入 + sendMessage 去重 |
| 3 | cancelPending 竞态 | `tui_model.go` | agentDoneMsg 检查 cancelPending |
| 4 | file_read offset 行号错误 | `tools/file_read.go` | 行号从 `offset+i+1` 开始 |
| 5 | Anthropic stream 缩进混乱 | `llm/anthropic.go` | case 对齐 |
| 6 | OpenAI 流式工具调用缺 index | `llm/openai.go` | openAIToolCall 加 Index 字段 |
| 7 | temperature:0 无法设置 | `message.go` + providers | Temperature/TopP 改 *float64 |
| 8 | toolDefinitions 类型断言 panic | `agent/agent.go` | 用 ,ok 断言 |
| 9 | 压缩不生效（75% 阈值） | `agent/agent.go` | 新增 compressContextForce 跳过阈值 |
| 10 | 重复 MiMo 前缀 | `tui_model.go` | sendMessage 移除静态前缀 |
| 11 | JSON 序列化空对象 | `chat_message_json.go` | 自定义 MarshalJSON/UnmarshalJSON |
| 12 | 排队消息 busyStart 覆盖 | 预期行为 | 每条消息独立计时 |

---

## 三、当前已知问题

### 问题 1：IME 拼音不跟随光标
- **现象**：中文输入法拼音候选框固定在输入框左侧，不跟随光标
- **原因**：Bubbletea alt-screen 模式下，终端 IME 无法追踪光标位置
- **状态**：这是 Bubbletea 在 Windows 上的底层限制，非代码 bug
- **影响**：中文输入体验差，但不影响功能

### 问题 2：文本无法直接选中复制
- **现象**：alt-screen 模式下鼠标无法直接选中文本
- **原因**：alt-screen 使用独立屏幕缓冲区
- **变通**：`/copy` 命令，或 Windows Terminal 中按住 Shift+鼠标选中
- **状态**：已尝试 inline 模式但终端缩放时 UI 重复错乱，回退到 alt-screen

### 问题 3：inline 模式终端缩放 UI 错乱
- **现象**：不用 alt-screen 时，拖动终端边缘改变宽度会导致状态栏、分隔线重复出现
- **原因**：Bubbletea inline 渲染器在宽度变化时无法正确计算上一帧行数
- **状态**：这是终端 inline 渲染的通病，Claude Code CLI 也有同样问题
- **结论**：保持 alt-screen 模式，牺牲 IME 和文本选中换取 UI 稳定

### 问题 4：排队消息的 busyStart 覆盖
- **现象**：自动发送队列消息时 `sendMessage` 重置 `busyStart`
- **影响**：状态栏时间重新开始计时（预期行为，每条消息独立计时）
- **状态**：不修复

---

## 四、代码架构

### 关键文件
| 文件 | 职责 |
|------|------|
| `cmd/tui_model.go` | Model 定义 + Update 逻辑 + 所有业务逻辑（~640 行） |
| `cmd/tui_view.go` | View 渲染 + 状态栏 + 欢迎横幅 |
| `cmd/tui_styles.go` | Lipgloss 样式定义 |
| `cmd/tui_messages.go` | 桥接消息类型（12 种） |
| `cmd/chat_message_json.go` | chatMessage 自定义 JSON 序列化 |
| `cmd/interactive.go` | 启动入口，组装依赖，注册回调 |
| `cmd/run.go` | `mimo run` 单次执行模式 |
| `internal/agent/agent.go` | Agent 核心：ReAct 循环、并行工具、上下文压缩 |
| `internal/llm/gateway.go` | LLM 网关，自动路由 Anthropic/OpenAI |
| `internal/llm/anthropic.go` | Anthropic 流式/非流式实现 |
| `internal/llm/openai.go` | OpenAI 流式/非流式实现 |
| `internal/tools/*.go` | 13 个工具（shell/file/git/search/web 等） |
| `internal/safety/guardrail.go` | 安全防护（已接入 Agent） |
| `internal/safety/classifier.go` | 操作分类 |
| `internal/config/config.go` | 配置加载/合并/保存 |
| `internal/context/manager.go` | 上下文管理 + 系统提示词构建 |
| `internal/backup/backup.go` | 文件备份管理 |
| `internal/ignore/ignore.go` | 忽略规则 |

### 数据流
```
用户输入 → handleKey → sendMessage → go runAgent
                                         ↓
                                    agent.ChatStream
                                         ↓
                              callbacks → streamChan → listenStream → Update
                                                                         ↓
                                                            thinkingMsg / deltaMsg / toolCallMsg / ...
                                                                         ↓
                                                              agentDoneMsg → finalizeResponse
```

### 新增字段清单
```go
// tuiModel 字段
messageQueue     []string     // 消息队列
cancelPending    bool         // 取消标记
userScrolledUp   bool         // 用户手动滚动
spinnerFrame     int          // 动画帧计数
busyStart        time.Time    // 本轮开始时间
busyEnd          time.Time    // 本轮结束时间
msgStartTokens   int          // 本轮起始 token
msgToolCalls     int          // 本轮工具调用次数
msgTokens        int          // 本轮 token 消耗
msgPromptTokens  int          // 本轮 prompt token
msgCompletionTokens int       // 本轮 completion token
userName         string       // 用户昵称
nameSetup        bool         // 昵称设置中
sessionId        string       // 会话 ID（用于文件保存）
multilineInput  textarea.Model // 多行输入
useTextarea     bool           // 是否使用多行模式
resuming         bool         // 会话恢复选择中
resumeFiles      []string     // 可恢复的会话文件列表
resumeIdx        int          // 当前选中的会话索引
compressing      bool         // 上下文压缩中
compressStart    time.Time    // 压缩开始时间

// Config 新增字段
UserName string `yaml:"user_name"`

// Agent 新增方法
LoadMessages(msgs []llm.Message)  // 加载历史消息
CompressContext(ctx) (before, after, err)  // 手动压缩
```

### 消息类型（12 种）
```go
thinkingMsg    // 思考过程增量
deltaMsg       // 流式文本增量
toolCallMsg    // 工具调用
toolResultMsg  // 工具结果
agentDoneMsg   // Agent 完成
agentErrMsg    // Agent 错误
usageMsg       // Token 用量
compressingMsg // 压缩开始
compressDoneMsg // 压缩完成
confirmMsg     // 安全确认
spinnerTickMsg // 旋转动画刷新
```

### 可用命令
```
/copy [N|all|user]  - 复制响应到剪贴板
/compress           - 压缩上下文
/rollback [N]       - 回滚文件到备份
/clear              - 清空屏幕
/help               - 显示帮助
/name <昵称>        - 修改昵称
/model [名称]       - 查看/切换模型
/export             - 导出对话记录
/resume             - 恢复历史会话
/exit               - 退出程序
```

---

## 五、v0.1.0 MVP 完成度

| MVP 目标 | 状态 |
|---------|------|
| 项目骨架（Cobra + BubbleTea） | ✅ |
| 基础交互模式 | ✅ |
| LLM API 对接（OpenAI + Anthropic） | ✅ |
| 流式输出渲染 | ✅ |
| 基础工具（shell/file/git/search/web 等 13 个） | ✅ |
| 安全确认机制 | ✅ |
| 配置文件系统（深度合并） | ✅ |
| `mimo run` 单次执行 | ✅ |

**超出 MVP 的功能**：上下文压缩、备份管理、忽略规则、并行工具、取消操作、消息队列、命令补全、用户昵称、旋转动画、压缩进度条、会话恢复、多模型切换、对话导出、Token 预算提示、命令历史持久化

---

## 六、下一步计划

### 第一阶段：小功能（待做）
| # | 功能 | 状态 |
|---|------|------|
| 1 | 命令历史持久化 | ✅ 已完成 |
| 2 | /name 命令 | ✅ 已完成 |
| 3 | Token 预算提示 | ✅ 已完成 |
| 4 | prompt/completion 分开显示 | ✅ 已完成 |
| 5 | /export 命令 | ✅ 已完成 |
| 6 | 多模型配置 | ✅ 已完成 |

### 第二阶段：中等功能（待做）
| # | 功能 | 说明 |
|---|------|------|
| 7 | 会话恢复 | ✅ 已完成 |
| 8 | `mimo run` 增强 | ✅ 已完成 |
| 9 | AGENT.md 解析 | ✅ 已完成 |
| 10 | 多行输入 | ❌ 已放弃（Windows 终端限制） |
| 11 | Git 工具补全 | ✅ 已完成 |
| 12 | 主题切换 | ✅ 已完成 |

### 第三阶段：大功能（待做）
| # | 功能 | 说明 |
|---|------|------|
| 13 | 完整工具集 | 补全 20+ 工具 |
| 14 | ReAct 规划引擎 | 任务分解、执行计划展示 |
| 15 | 会话持久化框架 | SQLite 存储 |
| 16 | MCP 客户端 | 连接 MCP 服务器 |

### 第四阶段：生态功能（待做）
| # | 功能 | 说明 |
|---|------|------|
| 17 | 技能系统 | SKILL.md 加载/执行/市场 |
| 18 | 插件系统 | 插件注册工具/技能/钩子 |
| 19 | 记忆系统 | 向量化存储 |
| 20 | 沙箱执行 | Docker 隔离 |
| 21 | MCP 服务器 | 暴露能力给其他工具 |
| 22 | 多代理协作 | 角色分工 |

---

## 七、技术债与注意事项

1. **tui_model.go 已有 ~640 行**：建议后续重构时拆分（commands.go、session.go、helpers.go）
2. **inline 模式已放弃**：alt-screen 是当前唯一稳定的渲染方案，IME 和文本选中是已知限制
3. **chatMessage JSON 序列化**：通过 `chat_message_json.go` 单独文件实现，避免修改整个结构体
4. **PowerShell 转义问题**：修改 Go 代码时避免在 PowerShell 中使用反引号，优先用 Node.js 或 apply_patch
5. **Go 代理配置**：编译需要 `$env:GOPROXY="https://goproxy.cn,direct"`，否则下载依赖超时
6. **构建缓存**：`$env:GOCACHE` 和 `$env:GOMODCACHE` 需指向可写目录
7. **session 恢复后 sessionId 继承**：确保恢复的会话继续更新同一文件