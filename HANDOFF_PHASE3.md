# MiMo CLI — 第三阶段上下文交接文档

> 日期：2026-06-07
> 对话目的：第三阶段功能开发（ReAct 规划引擎）

---

## 一、第三阶段完成情况

### 已完成

| # | 功能 | 状态 | 说明 |
|---|------|------|------|
| 13 | 完整工具集 | ✅ | 25 个工具 |
| 14 | SQLite 会话持久化 | ✅ | `~/.mimo/sessions.db` |
| 15 | ReAct 规划引擎 | ✅ | 详见下方 |

### 待做

| # | 功能 | 说明 |
|---|------|------|
| 16 | MCP 客户端 | 连接 MCP 服务器 |

---

## 二、ReAct 规划引擎详细说明

### 新增文件

| 文件 | 行数 | 职责 |
|------|------|------|
| `internal/agent/planner.go` | ~230 | 规划引擎核心，Plan/PlanStep 数据结构 |

### 修改文件

| 文件 | 修改内容 |
|------|----------|
| `internal/agent/agent.go` | 添加 Planner、planningMode、isComplexTask、runPlanExecuteLoop、回调函数 |
| `cmd/tui_model.go` | 添加 /plan 命令、planSelecting、thinking 保存、实时内容显示 |
| `cmd/tui_view.go` | 状态栏显示规划模式 |
| `cmd/tui_messages.go` | 添加 planGeneratedMsg、planStepStartMsg、planStepDoneMsg、planningMsg |
| `cmd/interactive.go` | 设置规划模式 |
| `internal/config/schema.go` | 添加 PlanningMode 字段 |

### 核心功能

1. **三种规划模式**
   - `react`: ReAct 循环（默认，适合大多数任务）
   - `plan-execute`: 先生成计划再执行（适合复杂任务）
   - `auto`: 自动判断任务复杂度

2. **isComplexTask 判断逻辑**
   - 简单任务（问候、单文件操作）→ 不走计划
   - 复杂任务（项目、系统、多步骤）→ 走计划

3. **实时反馈**
   - thinking 内容实时显示
   - 计划生成和执行进度
   - 文件内容实时预览
   - /plan 命令交互式切换

---

## 三、已知问题和 Bug

### 已修复

1. ✅ planningMsg 覆盖 rawThinking → 改为追加
2. ✅ thinking 显示逻辑混乱 → 区分 planning/thinking 显示
3. ✅ thinking 内容不保存 → chatMessage 添加 thinking 字段
4. ✅ 窗口大小变化丢失 thinking → rebuildMessages 重建
5. ✅ 计划执行结果不入历史 → 添加到 a.messages
6. ✅ isComplexTask 误判 → 优化判断逻辑
7. ✅ 命令列表无序 → 按字母排序

### 潜在问题

1. **GeneratePlan 是同步调用**
   - 没有流式输出
   - 用户等待时只看到 "planning... Xs"
   - 后续可改为流式调用

2. **isComplexTask 仍有误判可能**
   - 基于关键词匹配，不够智能
   - 后续可用 LLM 判断任务复杂度

3. **Plan-Execute 模式未充分测试**
   - 当前默认使用 react 模式
   - plan-execute 模式的错误处理需要验证

4. **tui_model.go 过大**
   - 已有 800+ 行
   - 建议拆分为 commands.go、plan.go、helpers.go

5. **无单元测试**
   - planner.go 和 agent.go 的新功能没有测试

---

## 四、配置说明

### 当前推荐配置

```yaml
# ~/.mimo/config.yaml
models:
    mimo:
        api_base: https://token-plan-cn.xiaomimimo.com/anthropic
        api_key: tp-xxx
        model: mimo-v2.5-pro
        max_tokens: 128000      # 模型最大输出
        temperature: 0.3
        top_p: 0.95             # 修复了之前的 0 值

context:
    max_tokens: 1000000         # 上下文窗口（1M）

agent:
    planning_mode: react        # 推荐使用 react 模式
    max_iterations: 25
    max_parallel_tools: 5
```

### /plan 命令

```
/plan                  # 交互式选择
/plan react            # 直接切换
/plan plan-execute     # 直接切换
/plan auto             # 直接切换
```

---

## 五、关键文件清单

| 文件 | 职责 | 行数 |
|------|------|------|
| `cmd/tui_model.go` | Model 定义 + Update + 业务逻辑 | ~800+ |
| `cmd/tui_view.go` | View 渲染 + 状态栏 | ~100 |
| `cmd/tui_messages.go` | 消息类型桥接 | ~70 |
| `cmd/tui_styles.go` | 样式定义（dark/light） | ~170 |
| `cmd/interactive.go` | 启动入口，组装依赖 | ~120 |
| `internal/agent/agent.go` | Agent 核心 + ReAct 循环 + Plan-Execute | ~650 |
| `internal/agent/planner.go` | 规划引擎核心 | ~230 |
| `internal/tools/file_write.go` | 文件写入工具 | ~120 |
| `internal/config/schema.go` | 配置定义 | ~140 |

---

## 六、下一步计划

### 立即要做

1. **MCP 客户端**（第三阶段 4/4）
   - 实现 MCP 协议客户端
   - 支持 stdio 和 SSE 传输
   - 自动发现和注册 MCP 工具
   - 配置文件中添加 mcp.servers 配置

### 后续优化

2. **tui_model.go 拆分重构**
   - 拆分为 commands.go、plan.go、helpers.go

3. **单元测试补充**
   - planner.go 测试
   - agent.go 新功能测试

4. **GeneratePlan 流式化**
   - 改为流式调用，实时显示计划生成过程

5. **isComplexTask 优化**
   - 用 LLM 判断任务复杂度

---

## 七、构建命令

```powershell
$env:GOPROXY="https://goproxy.cn,direct"
$env:GOCACHE="D:\works\study\mimo cli\.gocache"
$env:GOMODCACHE="D:\works\study\mimo cli\.gomod"
cd "D:\works\study\mimo cli"
go build -o bin\mimo.exe ./main.go
```

---

## 八、测试清单

### ReAct 规划引擎测试

- [ ] 输入"你好" → 秒回，不走计划
- [ ] 输入"创建一个 HTML 文件" → 直接执行，不走计划
- [ ] 输入"搭建博客系统" → react 模式下直接执行
- [ ] /plan 命令 → 交互式选择
- [ ] /plan react → 切换成功
- [ ] 状态栏显示当前模式
- [ ] 调整窗口大小 → thinking 内容保留
- [ ] 文件写入 → 实时显示内容预览

### 功能测试

- [ ] /help → 命令列表按字母排序
- [ ] /theme → 主题切换正常
- [ ] /model → 模型切换正常
- [ ] /compress → 上下文压缩正常
- [ ] /export → 导出对话正常
- [ ] /resume → 恢复会话正常
