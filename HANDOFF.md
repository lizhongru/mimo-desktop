# MiMo CLI — 上下文交接文档

> 日期：2026-06-06
> 对话目的：第二阶段功能开发 + 第三阶段部分功能

---

## 一、本次对话完成的工作

### 第二阶段（全部完成）

| # | 功能 | 状态 | 说明 |
|---|------|------|------|
| 8 | `mimo run` 增强 | ✅ | 管道输入(stdin)、`--output json/markdown`、退出码 0/1/2 |
| 9 | AGENT.md 解析 | ✅ | 结构化提取 sections、`.mimo/rules.md` 注入、Git 信息自动注入 |
| 10 | 多行输入 | ❌ 放弃 | Windows 终端检测不到 Shift+Enter，已清理代码 |
| 11 | Git 工具补全 | ✅ | status 分类显示、diff 截断+diffstat、log 详细格式、commit 输出摘要 |
| 12 | 主题切换 | ✅ | dark/light 两套配色、`/theme` 上下键交互选择、配置持久化 |

### 第三阶段（2/4 完成）

| # | 功能 | 状态 | 说明 |
|---|------|------|------|
| 13 | 完整工具集 | ✅ | 从 13 个增加到 25 个工具 |
| 14 | SQLite 会话持久化 | ✅ | `~/.mimo/sessions.db`，纯 Go SQLite（modernc.org/sqlite） |
| 15 | ReAct 规划引擎 | 待做 | 需要改 agent 核心循环，任务分解+执行计划展示 |
| 16 | MCP 客户端 | 待做 | 连接 MCP 服务器 |

### 额外修复和优化

- **规则注入**：`.mimo/rules.md` 内容注入到每条用户消息前面（解决模型忽略系统提示词末尾规则的问题）
- **命令列表**：固定 5 个高度 + 上下滚动 + 单列带描述
- **命令历史**：所有输入（包括 `/` 命令）都保存到历史
- **排队消息**：移到输入框上方、状态栏下方，上下有空行
- **`/confirm` 命令**：重新启用安全确认（按 `a` 全部确认后可用）
- **搜索引擎**：从 DuckDuckGo 换成 Bing（国内可用）
- **`/ml` 多行模式**：已移除（Windows 不支持）

---

## 二、新增工具清单（10 个）

| 工具 | 文件 | 功能 | 安全等级 |
|------|------|------|---------|
| `file_diff` | `file_diff.go` | 两文件逐行对比 | LOW |
| `clipboard` | `clipboard.go` | 剪贴板读写 | LOW |
| `process` | `process.go` | 进程列表/终止 | HIGH |
| `env` | `env.go` | 环境变量读取/列表 | LOW |
| `dependency` | `dependency.go` | 包管理器（npm/pip/go/cargo 自动检测） | MEDIUM |
| `http_request` | `http_request.go` | HTTP 请求 | MEDIUM |
| `web_search` | `web_search.go` | Bing 搜索 | LOW |
| `file_delete` | `file_delete.go` | 文件删除（确认机制 + 保护目录） | HIGH |
| `dir_create` | `dir_create.go` | 目录创建（在 file_delete.go 中） | LOW |
| `json_query` | `json_query.go` | JSON/YAML 点路径查询 | LOW |

---

## 三、已知问题和潜在风险

### 已知问题

1. **confirmAll 持久化**：按 `a` 全部确认后，`confirmAll=true` 一直生效到会话结束。已加 `/confirm` 命令重置，但用户可能不知道。

2. **搜索 HTML 解析脆弱**：Bing 搜索结果的 HTML 解析依赖 `b_algo` class 名称，如果 Bing 改版会失效。有 fallback 逻辑兜底。

3. **tui_model.go 过大**：已有 ~700+ 行，包含所有业务逻辑。建议后续拆分为 `commands.go`、`session.go`、`helpers.go`。

4. **tui_view.go 多次重写**：本次对话中 View 函数被多次修改，可能存在冗余代码。

5. **SQLite 依赖体积**：`modernc.org/sqlite` 是纯 Go 实现，编译后二进制会增大。

6. **剪贴板 Windows 实现**：clipboard.go 中 Windows 的 `Set-Clipboard` 实现有重复的 cmd 赋值，可能不工作。

7. **sessionStore 生命周期**：sessionStore 在 `interactive.go` 中 defer Close()，但 mimo 是长期运行的，defer 只在退出时执行，这没问题，但要注意不要在其他地方关闭它。

### 潜在风险

8. **无单元测试**：新增的工具和 session store 都没有单元测试。`internal/agent/agent_test.go` 存在但可能过时。

9. **`go.sum` 可能过期**：添加了 `modernc.org/sqlite` 依赖后，`go.sum` 可能需要更新。

10. **AGENT.md 解析中英文混合**：`parseAgentMD` 函数用中文关键词匹配 section（如"代理规则"、"禁止操作"），纯英文 AGENT.md 可能无法正确分类。

---

## 四、下一步计划

### 立即要做

1. **ReAct 规划引擎**（第三阶段 3/4）
   - 需要修改 `internal/agent/agent.go` 的核心循环
   - 实现任务分解：将复杂任务拆解为子步骤
   - 执行计划展示：在 TUI 中显示当前计划和进度
   - 支持 Plan-Execute 和 ReAct 两种模式
   - 参考 init.md 中的规划引擎设计

2. **MCP 客户端**（第三阶段 4/4）
   - 实现 MCP 协议客户端
   - 支持 stdio 和 SSE 传输
   - 自动发现和注册 MCP 工具
   - 配置文件中添加 mcp.servers 配置

### 后续优化

3. **tui_model.go 拆分重构**
4. **单元测试补充**
5. **剪贴板 Windows 实现修复**

---

## 五、关键文件清单

| 文件 | 职责 | 行数 |
|------|------|------|
| `cmd/tui_model.go` | Model 定义 + Update + 所有业务逻辑 | ~700+ |
| `cmd/tui_view.go` | View 渲染 | ~90 |
| `cmd/tui_styles.go` | 样式定义（支持 dark/light 主题） | ~170 |
| `cmd/tui_messages.go` | 消息类型桥接 | ~40 |
| `cmd/interactive.go` | 启动入口，组装依赖 | ~100 |
| `cmd/run.go` | `mimo run` 单次执行 | ~200 |
| `cmd/root.go` | 根命令 | ~40 |
| `internal/agent/agent.go` | Agent 核心：ReAct 循环 | ~800 |
| `internal/session/store.go` | SQLite 会话存储 | ~230 |
| `internal/context/manager.go` | 上下文管理 + 系统提示词 | ~300 |
| `internal/tools/*.go` | 25 个工具 | 各 ~100 |
| `internal/config/schema.go` | 配置定义 | ~130 |
| `internal/llm/openai.go` | OpenAI 兼容 LLM | ~300 |
| `internal/llm/anthropic.go` | Anthropic LLM | ~350 |
| `internal/safety/guardrail.go` | 安全护栏 | ~170 |

---

## 六、构建命令

```powershell
$env:GOPROXY="https://goproxy.cn,direct"
$env:GOCACHE="D:\works\study\mimo cli\.gocache"
$env:GOMODCACHE="D:\works\study\mimo cli\.gomod"
cd "D:\works\study\mimo cli"
go build -o bin\mimo.exe ./main.go
```

## 七、测试清单文件

- [TEST_CHECKLIST.md](TEST_CHECKLIST.md) — 第二阶段测试清单
- [TEST_CHECKLIST_PHASE3.md](TEST_CHECKLIST_PHASE3.md) — 第三阶段测试清单
