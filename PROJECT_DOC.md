# MiMo CLI — 项目综合文档

> 最后更新：2026-06-08
> 合并自 README、HANDOFF_PHASE2~7、TASK_PROGRESS、TEST_CHECKLIST、UI_FIX_PLAN

---

## 一、项目概述

MiMo CLI 是一款基于大语言模型的代理式命令行开发工具，同时提供 Wails v2 桌面客户端。通过自然语言描述，能够自主完成代码编写、文件操作、项目构建、测试调试等开发全流程任务。

### 核心特性

| 特性 | 描述 |
|------|------|
| 智能代理 | 基于 MiMo 大模型，理解自然语言指令并自主执行任务 |
| 丰富工具集 | 25 个内置工具，覆盖文件操作、Git、网络、数据库等场景 |
| 多模型支持 | 支持 MiMo、OpenAI、Anthropic、DeepSeek 等多模型热切换 |
| 安全可靠 | 分级安全策略 + 沙箱执行 + 操作审计 |
| 高效执行 | 并行任务执行 + 流式输出 + 智能上下文管理 |
| 可扩展 | 技能系统 + 插件架构 + MCP 协议支持 |

### 系统架构

```
┌──────────────────────────────────────────────────────────┐
│  Interface Layer (CLI TUI / Desktop Wails / IDE)         │
│          ↓                                               │
│  Core Agent Engine                                       │
│    ├── Intent Parser（意图解析）                          │
│    ├── Planner Engine（规划引擎：ReAct / Plan-Execute）   │
│    ├── Executor Engine（执行引擎：并行工具调用）           │
│    └── Observer Engine（观察引擎：上下文压缩）             │
│          ↓                                               │
│  Infrastructure Layer                                    │
│    ├── Tool System（25 个内置工具 + MCP 外部工具）         │
│    ├── Context Manager（系统提示词 + 规则注入）            │
│    ├── Safety Guardrail（三级安全策略）                    │
│    └── Session Store（SQLite 持久化）                     │
└──────────────────────────────────────────────────────────┘
```

---

## 二、开发历程

### Phase 1：MVP 基础（已完成）

| 功能 | 状态 |
|------|------|
| Cobra + Bubbletea 项目骨架 | ✅ |
| 基础交互模式（alt-screen） | ✅ |
| LLM API 对接（OpenAI + Anthropic） | ✅ |
| 流式输出渲染 | ✅ |
| 13 个基础工具（shell/file/git/search/web） | ✅ |
| 安全确认机制 | ✅ |
| 配置文件系统（多层 YAML 深度合并） | ✅ |
| `mimo run` 单次执行 | ✅ |
| 消息队列（busy 时排队发送） | ✅ |
| 用户昵称系统 | ✅ |
| Token 统计与预算警告（96k 变色） | ✅ |
| 会话恢复（/resume） | ✅ |
| 对话导出（/export） | ✅ |
| 命令补全与历史持久化 | ✅ |
| 多模型切换（/model） | ✅ |
| 压缩进度条与动画 | ✅ |

### Phase 2：功能增强（已完成）

| 功能 | 说明 |
|------|------|
| `mimo run` 增强 | 管道输入、`--output json/markdown`、退出码 0/1/2 |
| AGENT.md 解析 | 结构化提取 sections、`.mimo/rules.md` 注入、Git 信息自动注入 |
| Git 工具补全 | status 分类显示、diff 截断+diffstat、log 详细格式、commit 输出摘要 |
| 主题切换 | dark/light 两套配色、`/theme` 上下键交互选择、配置持久化 |

### Phase 3：核心引擎（已完成）

| 功能 | 说明 |
|------|------|
| 完整工具集 | 从 13 个增加到 25 个工具 |
| SQLite 会话持久化 | `~/.mimo/sessions.db`，纯 Go SQLite（modernc.org/sqlite） |
| ReAct 规划引擎 | 三种模式（react/plan-execute/auto）、任务复杂度自动判断 |
| MCP 客户端 | JSON-RPC 2.0、stdio/SSE 传输、自动工具注册、交互式向导添加 |

### Phase 4：桌面客户端（v0.3.0-dev，进行中）

| 功能 | 说明 |
|------|------|
| Wails v2 + React 19 桌面应用 | 无边框窗口、三栏布局、流式聊天 |
| 14 个前端组件 | MessageList/ChatInput/SettingsPage 等 |
| 4 个 Zustand stores | chat/session/settings/activity |
| 中英双语 i18n | 100+ 翻译 key |
| 安全确认弹窗 | approve/deny/approve-all |

---

## 三、工具清单（25 个）

| 工具 | 功能 | 安全等级 |
|------|------|----------|
| `shell` | 执行 Shell 命令 | HIGH |
| `file_read` | 读取文件内容 | LOW |
| `file_write` | 写入/创建文件 | MEDIUM |
| `file_edit` | 精确编辑文件 | MEDIUM |
| `file_delete` | 删除文件（确认机制 + 保护目录） | HIGH |
| `file_diff` | 两文件逐行对比 | LOW |
| `dir_list` | 列出目录内容 | LOW |
| `search` | 项目内文本搜索（ripgrep） | LOW |
| `glob` | 文件模式匹配 | LOW |
| `git` | Git 操作全套 | MEDIUM |
| `web_fetch` | 抓取网页内容 | LOW |
| `web_search` | Bing 搜索 | LOW |
| `http_request` | 发送 HTTP 请求 | MEDIUM |
| `json_query` | JSON/YAML 点路径查询 | LOW |
| `clipboard` | 剪贴板读写 | LOW |
| `process` | 进程列表/终止 | HIGH |
| `env` | 环境变量读取/列表 | LOW |
| `dependency` | 包管理器（npm/pip/go/cargo 自动检测） | MEDIUM |

---

## 四、ReAct 规划引擎

### 三种模式

| 模式 | 说明 | 适用场景 |
|------|------|---------|
| `react` | ReAct 循环（默认） | 大多数任务 |
| `plan-execute` | 先生成计划再执行 | 复杂多步骤任务 |
| `auto` | 自动判断任务复杂度 | 不确定时 |

### 任务复杂度判断

- **简单任务**（问候、单文件操作）→ 直接执行，不走计划
- **复杂任务**（项目搭建、系统级操作、多步骤）→ 生成计划后执行

### 切换命令

```
/plan                  # 交互式选择
/plan react            # 直接切换
/plan plan-execute     # 直接切换
/plan auto             # 直接切换
```

---

## 五、MCP 协议支持

### 配置方式

```yaml
# ~/.mimo/config.yaml
mcp:
  servers:
    filesystem:
      command: node
      args: ["path/to/server-filesystem/dist/index.js", "workdir"]
      enabled: true
```

### 使用命令

| 命令 | 功能 |
|------|------|
| `/mcp` | 查看 MCP 服务器状态 |
| `/mcp add` | 添加新服务器（交互式向导，上下键选择推荐列表） |
| `/mcp remove <名称>` | 移除服务器 |

### 架构

| 文件 | 职责 |
|------|------|
| `internal/mcp/protocol.go` | JSON-RPC 2.0 消息类型和 MCP 协议定义 |
| `internal/mcp/transport.go` | 传输层接口 |
| `internal/mcp/transport_stdio.go` | Stdio 传输（子进程通信） |
| `internal/mcp/transport_sse.go` | SSE 传输（HTTP 远程通信） |
| `internal/mcp/client.go` | MCP 客户端（initialize、tools/list、tools/call） |
| `internal/mcp/mcp_tool.go` | MCP 工具适配器（实现 BaseTool 接口） |
| `internal/mcp/manager.go` | MCP 管理器（多服务器连接管理） |

---

## 六、配置说明

### 配置层级（优先级从高到低）

1. 命令行参数
2. 项目级配置 `./mimo.yaml`
3. 用户级配置 `~/.mimo/config.yaml`
4. 系统级配置 `/etc/mimo/config.yaml`
5. 内置默认值

### 推荐配置示例

```yaml
# ~/.mimo/config.yaml
default_model: mimo
language: zh-CN
theme: dark
user_name: 你的昵称

models:
  mimo:
    api_base: https://api.mimo.xiaomi.com/v1
    api_key: tp-xxx
    model: mimo-v2.5-pro
    max_tokens: 128000
    temperature: 0.3
    top_p: 0.95

safety:
  level: confirm
  backup_before_write: true
  audit_log: ~/.mimo/audit.log
  blocked_commands: [sudo, "chmod 777"]
  protected_files: [.env, "*.pem", "*.key", "id_rsa*"]
  protected_branches: [main, master, "release/*"]

agent:
  max_iterations: 50
  planning_mode: react
  permission: exec
  show_token_usage: true

context:
  max_tokens: 1000000
  ignore_patterns: [node_modules, .git, __pycache__, dist, build, .venv, vendor, .gocache, .gomod, .mimo]
```

---

## 七、CLI 快捷键与命令

### 快捷键

| 快捷键 | 功能 |
|--------|------|
| Enter | 发送消息 |
| Ctrl+J | 换行（多行输入） |
| Ctrl+C | 取消/退出 |
| Ctrl+L | 清屏 |
| Ctrl+T | 折叠/展开工具调用 |
| Esc | 取消当前操作/退出 |
| PgUp/PgDown | 滚动 |
| 上下键 | 命令历史/补全选择 |

### 斜杠命令

| 命令 | 功能 |
|------|------|
| `/copy [N\|all\|user]` | 复制响应到剪贴板 |
| `/compress` | 压缩上下文 |
| `/rollback [N]` | 回滚文件到备份 |
| `/clear` | 清空屏幕 |
| `/help` | 显示帮助 |
| `/name <昵称>` | 修改昵称 |
| `/model [名称]` | 查看/切换模型 |
| `/export` | 导出对话记录 |
| `/resume` | 恢复历史会话 |
| `/plan [模式]` | 切换规划模式 |
| `/theme [主题]` | 切换主题 |
| `/mcp` | MCP 服务器管理 |
| `/confirm` | 重置安全确认 |
| `/exit` | 退出程序 |

---

## 八、关键文件清单

### CLI 核心

| 文件 | 职责 | 行数 |
|------|------|------|
| `cmd/tui_model.go` | TUI 核心（Model + Update + 命令 + 辅助） | ~1400 |
| `cmd/tui_view.go` | View 渲染 + 状态栏 | ~130 |
| `cmd/tui_styles.go` | 样式定义（dark/light 主题） | ~170 |
| `cmd/tui_messages.go` | 消息类型桥接 | ~70 |
| `cmd/chat_message_json.go` | chatMessage 自定义 JSON 序列化 | ~60 |
| `cmd/interactive.go` | 启动入口，组装依赖 | ~200 |
| `cmd/run.go` | `mimo run` 单次执行 | ~200 |

### 内部引擎

| 文件 | 职责 | 行数 |
|------|------|------|
| `internal/agent/agent.go` | Agent 核心：ReAct 循环 + Plan-Execute + 并行工具 | ~650 |
| `internal/agent/planner.go` | 规划引擎核心 | ~230 |
| `internal/llm/gateway.go` | LLM 网关，自动路由 Anthropic/OpenAI | ~200 |
| `internal/llm/anthropic.go` | Anthropic 流式/非流式 | ~350 |
| `internal/llm/openai.go` | OpenAI 流式/非流式 | ~300 |
| `internal/tools/*.go` | 25 个工具 | 各 ~100 |
| `internal/safety/guardrail.go` | 安全护栏 | ~170 |
| `internal/safety/classifier.go` | 操作分类 | ~100 |
| `internal/mcp/*.go` | MCP 客户端（7 个文件） | ~770 |
| `internal/session/store.go` | SQLite 会话存储 | ~230 |
| `internal/config/schema.go` | 配置定义 | ~140 |
| `internal/config/config.go` | 配置加载/合并/保存 | ~120 |
| `internal/context/manager.go` | 上下文管理 + 系统提示词 | ~300 |

### Desktop 桌面端

| 文件 | 职责 |
|------|------|
| `desktop/app.go` | App 主结构体 + 初始化 |
| `desktop/app_chat.go` | 聊天相关方法 |
| `desktop/app_config.go` | 配置读写方法 |
| `desktop/app_session.go` | 会话管理方法 |
| `desktop/app_tools.go` | 工具/MCP 查询方法 |
| `desktop/events.go` | 事件名常量 |
| `desktop/frontend/src/` | React 前端（14 组件 + 4 stores + i18n） |

---

## 九、已修复 Bug 汇总

### CLI Bug（Phase 1~7，共 18 个）

| # | Bug | 修复 |
|---|-----|------|
| 1 | Token 统计可能为 0 | `len(response)/2` 粗略估算 |
| 2 | 排队消息 chatMessage 未记录 | 入队时写入 + sendMessage 去重 |
| 3 | cancelPending 竞态 | agentDoneMsg 检查 cancelPending |
| 4 | file_read offset 行号错误 | 行号从 `offset+i+1` 开始 |
| 5 | Anthropic stream 缩进混乱 | case 对齐 |
| 6 | OpenAI 流式工具调用缺 index | openAIToolCall 加 Index 字段 |
| 7 | temperature:0 无法设置 | Temperature/TopP 改 *float64 |
| 8 | toolDefinitions 类型断言 panic | 用 ,ok 断言 |
| 9 | 压缩不生效 | 新增 compressContextForce 跳过阈值 |
| 10 | JSON 序列化空对象 | 自定义 MarshalJSON/UnmarshalJSON |
| 11 | 终端 resize 后 thinking 消失 | WindowSizeMsg 返回 listenStream() |
| 12 | MCP 安装卡住无响应 | 2 分钟超时 + busy 状态重置 |
| 13 | MCP 工具不被 Agent 使用 | 系统提示词动态附加 MCP 工具描述 |
| 14 | 乱码 `�` 出现 | collapseToolResult 改用 []rune 截断 |
| 15 | MCP ListTools 返回空 | 循环读取跳过 id=0 通知消息 |
| 16 | MCP 并发请求响应错位 | 锁住整个 Send+Receive 周期 + 60s 超时 |
| 17 | Resize 后工具调用记录消失 | rebuildMessages 追加 toolLines 渲染 |
| 18 | 会话序列化丢失 thinking/toolLines | JSON 序列化 + SQLite migration |

### Desktop Bug（Phase 8，共 13 个）

| # | Bug | 修复 |
|---|-----|------|
| 1 | RespondToConfirm 永久阻塞 | select + 30s 超时 |
| 2 | SendMessage 用 context.Background() | 改用 context.WithCancel |
| 3 | 主题闪烁 | 删除模块级 dark class，由 initFromConfig 管理 |
| 4 | RespondToConfirmAll 无并发保护 | a.mu.Lock() 保护 |
| 5 | currentSessionID 无并发保护 | 统一加锁 |
| 6 | ListSessions 重复调用 | 删除 LeftSidebar 中的调用 |
| 7 | OpenInExplorer 忽略错误 | 检查 os.Getwd() 失败 |
| 8 | prose 类不生效 | 安装 @tailwindcss/typography |
| 9 | GetVersion 硬编码 | ldflags 变量注入 |
| 10 | Safety.Permission 映射 | 添加注释说明 |
| 11 | MCP 工具名 __ 误判 | strings.Index 替代逐字符扫描 |
| 12 | ToolCallEvent 缺 error 状态 | 类型加 "error" |
| 13 | 亮色模式代码块暗色主题 | 动态切换 oneDark/oneLight |

---

## 十、已知问题与技术债

### CLI 已知问题

| 问题 | 说明 | 状态 |
|------|------|------|
| IME 拼音不跟随光标 | Bubbletea alt-screen 底层限制 | 无法修复 |
| 文本无法直接选中复制 | alt-screen 限制，变通：`/copy` 或 Shift+鼠标 | 已知限制 |
| tui_model.go 过大 | ~1400 行，建议拆分 commands.go/plan.go/helpers.go | 待重构 |
| wrapText 中文换行 | 中文无空格不换行 + len() 宽度计算错误 | 待修复 |
| isComplexTask 误判 | 基于关键词匹配，不够智能 | 可用 LLM 判断 |
| GeneratePlan 非流式 | 用户等待时只看到 "planning... Xs" | 可改为流式 |

### Desktop 已知问题

| 问题 | 说明 | 状态 |
|------|------|------|
| TS 类型安全 | `(window as Record<string, unknown>)` 强制断言 | 待修 |
| i18n 遗漏 | 4 处硬编码文本未用 t() | 待修 |
| 死代码 | SaveCurrentSession stub、toJSON、空 main.go | 待清理 |
| 重复代码 | SafetyBadge/ToolInfo 接口在多文件重复 | 待提取 |

### 潜在风险

- 无单元测试（新增工具和 session store 均无测试）
- SQLite ALTER TABLE 容错（忽略列已存在错误，有意为之）
- MCP 工具优先级（LLM 倾向使用内置工具，已通过系统提示词优化）
- chatMessage 序列化兼容性（新增字段，旧 session 可能不兼容）

---

## 十一、构建与打包

### Go 代理配置（国内网络）

```powershell
$env:GOPROXY="https://goproxy.cn,direct"
$env:GOCACHE="D:\works\study\mimo cli\.gocache"
$env:GOMODCACHE="D:\works\study\mimo cli\.gomod"
```

### CLI 编译

```powershell
cd "D:\works\study\mimo cli"
go build -o bin\mimo.exe ./main.go
```

### Desktop 开发模式

```powershell
cd "D:\works\study\mimo cli"
wails dev -tags wails
```

### Desktop 生产打包

```powershell
cd "D:\works\study\mimo cli"
wails build -tags wails
# 输出：build/bin/mimo-desktop.exe（约 17MB）
```

---

## 十二、测试清单

### CLI 核心功能测试

- [ ] `mimo run "hello"` 正常输出
- [ ] `echo "what is 1+1" | mimo run` 管道输入
- [ ] `mimo run "hello" --output json` JSON 输出
- [ ] `mimo run "hello" --output markdown` Markdown 输出
- [ ] 退出码 0/1/2 正确
- [ ] `/theme` 主题切换 + 持久化
- [ ] `/model` 模型切换
- [ ] `/resume` 会话恢复
- [ ] `/export` 对话导出
- [ ] `/compress` 上下文压缩
- [ ] `/copy` 复制功能
- [ ] `/plan` 规划模式切换
- [ ] 状态栏显示 token/时间/规划模式
- [ ] Token 超 96k 时状态栏变色警告

### MCP 功能测试

- [ ] `/mcp` 显示服务器状态
- [ ] `/mcp add` 交互式添加
- [ ] `/mcp remove` 移除服务器
- [ ] MCP 工具被 Agent 正确调用
- [ ] 文件系统 MCP 工具读写正常

### Desktop 功能测试

- [ ] 流式聊天正常
- [ ] 安全确认弹窗（approve/deny/approve-all）
- [ ] 会话列表加载/切换/删除
- [ ] 设置页（主题/语言/模型/安全级别）
- [ ] 工具查看器弹窗
- [ ] 快捷键（Ctrl+B/I/N/Esc）
- [ ] 窗口控制（最小化/最大化/关闭）

---

## 十三、未来路线图

### 短期（稳定化）

- [ ] 修复 wrapText 中文换行
- [ ] 补充单元测试
- [ ] tui_model.go 拆分重构
- [ ] Desktop 高优 Bug 修复

### 中期（功能补全）

- [ ] Desktop 上下文压缩 UI 反馈
- [ ] Desktop 对话导出功能
- [ ] Desktop 消息右键菜单
- [ ] GeneratePlan 流式化

### 长期（生态建设）

- [ ] 技能系统（SKILL.md 加载/执行/市场）
- [ ] 插件系统（插件注册工具/技能/钩子）
- [ ] 记忆系统（向量化存储）
- [ ] 沙箱执行（Docker 隔离）
- [ ] MCP 服务器暴露（给其他工具使用）
- [ ] 多代理协作（角色分工）
- [ ] Desktop 多窗口/多标签
- [ ] Desktop 系统托盘
