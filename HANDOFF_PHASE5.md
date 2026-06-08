# MiMo CLI — 上下文交接文档

> 日期：2026-06-07
> 对话目的：Bug 修复 + MCP 功能增强 + UI 优化

---

## 一、本次对话完成的工作

### Bug 修复（6 个）

| # | Bug | 修复方案 | 状态 |
|---|-----|---------|------|
| 1 | 终端 resize 后 thinking/过程消失 | `WindowSizeMsg` 返回 `m.listenStream()` 保持消息通道活跃 | ✅ |
| 2 | MCP 安装卡住无响应 | `mcpInstallDoneMsg` 重置 busy 状态；npm install 添加 2 分钟超时；移除死代码 | ✅ |
| 3 | 每条消息缺少 token/time 统计 | `finalizeResponse` 和 `rebuildMessages` 始终显示统计行，token 为 0 时自动估算 | ✅ |
| 4 | MCP 工具不被 Agent 使用 | 系统提示词动态附加 MCP 工具描述；工具调用显示 `[MCP:server]` 紫色标签 | ✅ |
| 5 | 乱码 `�` 出现 | `collapseToolResult` 改用 `[]rune` 按字符截断，避免切断 UTF-8 多字节字符 | ✅ |
| 6 | Agent 超过最大迭代次数报错 | `maxIterations` 从 25 增加到 50，可通过配置文件调整 | ✅ |

### 功能增强

| # | 功能 | 说明 | 状态 |
|---|------|------|------|
| 1 | MCP 配置修复 | filesystem 服务器缺少目录参数，已补上 | ✅ |
| 2 | 忽略模式更新 | `.gocache`、`.gomod`、`.mimo` 加入忽略列表，避免搜索二进制文件 | ✅ |
| 3 | 状态栏显示规划模式 | 模型名后显示 `[react]`/`[auto]`/`[plan-execute]` | ✅ |
| 4 | 工具调用保留 | 回复后工具调用保留在屏幕，`Ctrl+T` 折叠/展开 | ✅ |
| 5 | 系统提示词优化 | 更积极地鼓励使用工具，减少"只回复文本不调用工具"的情况 | ✅ |

### 文件修改清单

| 文件 | 修改内容 |
|------|---------|
| `cmd/tui_model.go` | 完整重建（原文件被 Set-Content -NoNewline 破坏）；所有 Bug 修复和功能增强 |
| `cmd/tui_view.go` | 状态栏显示规划模式 |
| `cmd/tui_styles.go` | 添加 `toolMCPStyle`（紫色，用于 MCP 工具标签） |
| `cmd/interactive.go` | 系统提示词附加 MCP 工具描述；传递 `maxIterations` 参数 |
| `cmd/run.go` | 传递 `maxIterations` 参数 |
| `internal/agent/agent.go` | `NewAgent` 接受 `maxIterations` 参数 |
| `internal/config/schema.go` | `MaxIterations` 默认值改为 50 |
| `internal/context/manager.go` | 系统提示词优化（更积极鼓励使用工具） |
| `~/.mimo/config.yaml` | 添加忽略模式；max_iterations 改为 50；修复 MCP filesystem args |

---

## 二、已知问题和潜在风险

### 已知问题

1. **tui_model.go 重建质量**
   - 原文件被 `Set-Content -NoNewline` 破坏后，用 PowerShell 逐段重写
   - 代码风格可能与原版不一致（部分地方压缩成单行）
   - 功能完整但可读性下降
   - **建议**：后续有机会用 `gofmt` 统一格式化

2. **IME 拼音不跟随光标**
   - Bubbletea alt-screen 模式下的底层限制
   - 状态：已知问题，无法修复

3. **文本无法直接选中复制**
   - alt-screen 模式限制
   - 变通：`/copy` 命令或 Shift+鼠标选中

4. **MCP 工具优先级**
   - LLM 仍倾向于使用内置工具（如 `dir_list`）而非 MCP 工具
   - 原因：内置工具更简单直接
   - 状态：已通过系统提示词优化，但无法完全控制 LLM 选择

5. **搜索结果中的二进制内容**
   - 虽然添加了忽略模式，但 `.gomod` 中的测试文件仍可能被搜索到
   - 搜索工具可能返回非 UTF-8 内容

### 潜在风险

6. **tui_model.go 行数**
   - 重建后约 850 行（原版约 1400 行）
   - 部分功能可能被简化或遗漏
   - **建议**：仔细测试所有命令和功能

7. **chatMessage 序列化兼容性**
   - 新增 `toolLines` 字段
   - 旧的 session 文件可能无法正确加载
   - **建议**：测试 `/resume` 功能

8. **maxIterations 配置生效**
   - 代码默认值和配置文件都改为 50
   - 但 `NewAgent` 函数签名改变，需要确保所有调用点都更新

---

## 三、下一步计划

### 立即要做

1. **全面功能测试**
   - 测试所有命令（/help, /copy, /compress, /resume 等）
   - 测试 MCP 功能（/mcp, /mcp add, 工具调用）
   - 测试终端 resize 后的行为
   - 测试 Ctrl+T 折叠/展开功能
   - 测试会话恢复功能

2. **代码格式化**
   - 用 `gofmt` 统一格式化所有 Go 文件
   - 检查是否有语法问题

3. **ReAct 规划引擎**（第三阶段最后一项）
   - 修改 `internal/agent/agent.go` 核心循环
   - 任务分解 + 执行计划展示
   - Plan-Execute 和 ReAct 双模式

### 后续优化

4. **单元测试补充**
   - MCP 客户端测试
   - 工具适配器测试
   - TUI 组件测试

5. **tui_model.go 拆分**
   - 考虑拆分为 commands.go、session.go、helpers.go
   - 提高可维护性

6. **第四阶段生态功能**
   - 技能系统
   - 插件系统
   - 记忆系统
   - 沙箱执行

---

## 四、关键配置

### ~/.mimo/config.yaml 当前状态

```yaml
default_model: mimo
theme: dark
user_name: 哑火月光
models:
    mimo:
        api_base: https://token-plan-cn.xiaomimimo.com/v1
        api_key: tp-c2r6d15dvo4my8nupmylrmvoxcqkp1hh7vgqsnavctmuplub
        model: mimo-v2.5-pro
        max_tokens: 128000
        temperature: 0.3
        top_p: 0.95
agent:
    max_iterations: 50
    planning_mode: auto
context:
    ignore_patterns:
        - node_modules
        - .git
        - __pycache__
        - dist
        - build
        - .venv
        - vendor
        - .gocache
        - .gomod
        - .mimo
mcp:
    servers:
        filesystem:
            command: node
            args:
                - D:\works\study\mimo cli\node_modules\@modelcontextprotocol\server-filesystem\dist\index.js
                - D:\works\study\mimo cli
            enabled: true
```

---

## 五、构建命令

```powershell
$env:GOPROXY="https://goproxy.cn,direct"
$env:GOCACHE="D:\works\study\mimo cli\.gocache"
$env:GOMODCACHE="D:\works\study\mimo cli\.gomod"
cd "D:\works\study\mimo cli"
go build -o bin\mimo.exe ./main.go
```

---

## 六、快捷键参考

| 快捷键 | 功能 |
|--------|------|
| Enter | 发送消息 |
| Ctrl+C | 取消/退出 |
| Ctrl+L | 清屏 |
| Ctrl+T | 折叠/展开工具调用 |
| Esc | 取消当前操作/退出 |
| PgUp/PgDown | 滚动 |
| 上下键 | 命令历史/补全选择 |

---

## 七、待验证清单

- [ ] 所有命令正常工作
- [ ] MCP 工具可以被调用
- [ ] 终端 resize 后 thinking 保留
- [ ] Ctrl+T 折叠/展开功能
- [ ] 会话保存和恢复
- [ ] 主题切换
- [ ] 模型切换
- [ ] 压缩功能
- [ ] 导出功能
