# MiMo CLI — 上下文交接文档

> 日期：2026-06-07
> 对话目的：Bug 修复 + MCP 功能增强 + UI 优化

---

## 一、本次对话完成的工作

### Bug 修复（5 个）

| # | Bug | 修复方案 | 状态 |
|---|-----|---------|------|
| 1 | **MCP 工具不被 LLM 调用** | `runStreamLoop` 中 `finishReason=="stop"` 提前返回时忽略了同时存在的 `toolCalls`。修改为只在 `len(toolCalls)==0` 时才提前返回 | ✅ |
| 2 | **MCP `ListTools` 返回空** | `sendRequest` 读到了 `notifications/initialized` 通知消息（id=0），被当成响应。改为循环读取直到收到 `id!=0` 的真正响应 | ✅ |
| 3 | **MCP 并发请求响应错位** | `sendRequest` 只在 Send 时加锁，Receive 在锁外。改为锁住整个 Send+Receive 周期 + 60s 超时 | ✅ |
| 4 | **Resize 后流式内容丢失** | `WindowSizeMsg` 返回 `nil` 导致 stream 监听中断。改为 busy 时返回 `m.listenStream()` | ✅ |
| 5 | **Resize 后工具调用记录消失** | `rebuildMessages()` 不重建 `toolLines`。**已知问题，未修复** | ❌ |

### 功能增强（6 个）

| # | 功能 | 说明 | 状态 |
|---|------|------|------|
| 1 | **MCP 包全局安装** | npm 包从当前目录 `node_modules/` 改为安装到 `~/.mimo/mcp/`，不再污染项目目录 | ✅ |
| 2 | **MCP filesystem 目录权限** | `ConnectServer` 检测 filesystem 类型且 args 只有入口文件时，自动追加当前工作目录 | ✅ |
| 3 | **MCP 向导全链路可选择** | 步骤 1(名称) 可回车跳过用默认名、步骤 2(类型) ↑↓ 选择、步骤 3(包名) ↑↓ 推荐列表、步骤 4(目录) ↑↓ 选择 | ✅ |
| 4 | **输入组件升级** | 主输入从 `textinput` 改为 `textarea`，支持 Ctrl+J 换行多行输入 | ✅ |
| 5 | **多行消息渲染** | `renderUserText` 每行作为独立元素追加到 `m.messages`，不再用 `\n` 拼接 | ✅ |
| 6 | **Busy 状态工具显示优化** | `renderMessages` 只显示当前正在执行的工具（最后一个），不再刷屏全部历史 | ✅ |

### 文件修改清单

| 文件 | 修改内容 |
|------|---------|
| `cmd/tui_model.go` | 输入组件 textarea 化；renderUserText 多行修复；WindowSizeMsg stream 监听修复；MCP 向导全链路选择；inputReset/inputValue/inputFocus 辅助函数 |
| `cmd/tui_view.go` | View() 根据模式选择 textarea 或 cmdInput 渲染 |
| `internal/agent/agent.go` | `runStreamLoop` 修复 finishReason=="stop" 时 toolCalls 被丢弃的 bug |
| `internal/mcp/client.go` | `sendRequest` 锁住整个周期 + 跳过通知消息(id=0) + 60s 超时 |
| `internal/mcp/manager.go` | `ConnectServer` filesystem 自动追加当前目录 |

---

## 二、已知问题和待修复

### 高优先级

1. **Resize 后工具调用记录消失（问题 1）**
   - `rebuildMessages()` 的 assistant 分支不重建 `toolLines`
   - 修复方案：在 `rebuildMessages` 的 assistant case 中，根据 `showToolCalls` 和 `msg.toolLines` 重建工具调用显示
   - 用户已确认要先测试再修复

2. **多行消息截断（问题 2）**
   - 已改为 `renderUserText` 逐行追加，待用户实测确认是否修复

### 中优先级

3. **IME 拼音不跟随光标** — Bubbletea alt-screen 底层限制，无法修复
4. **文本无法直接选中复制** — alt-screen 限制，变通：`/copy` 或 Shift+鼠标
5. **MCP 工具优先级** — LLM 倾向使用内置工具而非 MCP 工具，已通过系统提示词优化

### 低优先级

6. **tui_model.go 代码质量** — 用 `gofmt` 统一格式化
7. **chatMessage 序列化兼容性** — 新增 `toolLines` 字段，旧 session 文件可能不兼容

---

## 三、下一步计划

### 立即要做

1. **修复 resize 后工具调用消失**（问题 1）
   - 在 `rebuildMessages()` 的 assistant case 中加入 `toolLines` 重建
   - 需判断 `showToolCalls` 开关

2. **验证多行消息修复**（问题 2）
   - 用户实测 Ctrl+J 换行输入 + 发送后显示

3. **MCP 工具全面测试**
   - 测试 filesystem 服务器的 14 个工具（read_text_file, write_file, list_directory 等）
   - 验证工具调用结果正确显示

### 后续优化

4. **ReAct 规划引擎**（第三阶段最后一项）
5. **单元测试补充** — MCP 客户端、工具适配器、TUI 组件
6. **第四阶段生态功能** — 技能/插件/记忆/沙箱系统

---

## 四、关键架构变化

### 输入组件双轨制

```
tuiModel {
    cmdInput textinput.Model  // 向导/命名/选择模式（单行）
    ta       textarea.Model   // 普通消息输入（多行，Ctrl+J 换行，Enter 发送）
}

辅助函数（自动路由到正确的输入组件）：
  inputReset()
  inputValue()
  inputFocus()
  inputSetValue(v)
```

### MCP 安装流程

```
/mcp add
  步骤 1: 名称（直接回车用默认名 "mcp-server"）
  步骤 2: 类型（↑↓ 选择: npm 包 / 本地命令 / 远程 URL）
  步骤 3: 包名（↑↓ 推荐列表 或 直接输入包名）
  步骤 4: 目录权限（↑↓: 当前目录 / 用户目录 / 自定义输入）— filesystem 专属
  步骤 5: 自定义路径 — 选了"自定义"才到这步

安装位置：~/.mimo/mcp/（全局目录）
Config 存储：args 只存入口文件路径（不含目录）
ConnectServer：filesystem 类型自动追加当前工作目录
```

### MCP 消息处理

```
sendRequest 流程：
  1. 加锁
  2. Send 请求
  3. 循环 Receive，跳过 id=0 的通知消息
  4. 返回第一个 id!=0 的响应
  5. 60s 超时保护
  6. 解锁
```

---

## 五、配置文件状态

```yaml
# ~/.mimo/config.yaml
default_model: mimo
models:
    mimo:
        api_base: https://token-plan-cn.xiaomimimo.com/v1
        api_key: tp-c2r6d15dvo4my8nupmylrmvoxcqkp1hh7vgqsnavctmuplub
        model: mimo-v2.5-pro
        max_tokens: 128000
agent:
    max_iterations: 50
    planning_mode: react
mcp:
    servers: {}  # 用户清空了，需要重新添加
context:
    max_tokens: 1000000
    ignore_patterns: [node_modules, .git, __pycache__, dist, build, .venv, vendor, .gocache, .gomod, .mimo]
```

---

## 六、构建命令

```powershell
$env:GOPROXY="https://goproxy.cn,direct"
$env:GOCACHE="D:\works\study\mimo cli\.gocache"
$env:GOMODCACHE="D:\works\study\mimo cli\.gomod"
cd "D:\works\study\mimo cli"
go build -o bin\mimo.exe ./main.go
```

---

## 七、快捷键参考

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

---

## 八、待验证清单

- [ ] 多行消息（Ctrl+J）发送后完整显示
- [ ] Resize 后工具调用记录保留（待修复）
- [ ] MCP 工具调用正常（需重新 `/mcp add`）
- [ ] 所有内置命令正常工作
- [ ] 会话保存和恢复
- [ ] 压缩功能
