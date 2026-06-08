# MiMo CLI — 项目初始化文档 (INIT.md)

> **项目代号**: `mimo-cli`
> **版本**: v0.1.0 (初始化)
> **定位**: 基于 MiMo 大模型的代理式命令行开发工具

---

## 一、项目愿景

打造一款**安全、智能、可扩展**的 AI 代理命令行工具。用户通过自然语言描述目标，工具自主完成代码编写、文件操作、项目构建、测试调试等开发全流程任务。

### 核心差异化

| 特性       | 现有工具         | MiMo CLI 目标           |
| ---------- | ---------------- | ----------------------- |
| 模型绑定   | 多数绑定单一模型 | 多模型热切换，MiMo 优先 |
| 本地化     | 英文为主         | 中英双语原生支持        |
| 安全机制   | 基础确认         | 分级沙箱 + 行为审计     |
| 技能生态   | 封闭             | 开放技能市场            |
| 上下文感知 | 项目级           | 组织级知识库接入        |
| 协作能力   | 单用户           | 多人协同代理            |

---

## 二、系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                     MiMo CLI Architecture                    │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────┐ │
│  │ Terminal  │  │ VS Code  │  │ JetBrains│  │   Web UI   │ │
│  │   TUI     │  │ Extension│  │  Plugin  │  │  (future)  │ │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘  └─────┬──────┘ │
│        │             │             │              │         │
│  ┌─────▼─────────────▼─────────────▼──────────────▼──────┐ │
│  │                   Interface Layer                      │ │
│  │         (Protocol Adapter / Event Bus)                 │ │
│  └─────────────────────┬─────────────────────────────────┘ │
│                        │                                    │
│  ┌─────────────────────▼─────────────────────────────────┐ │
│  │                   Core Agent Engine                    │ │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ │ │
│  │  │ Intent   │ │ Planner  │ │Executor  │ │ Observer │ │ │
│  │  │ Parser   │ │  Engine  │ │  Engine  │ │  Engine  │ │ │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘ │ │
│  └────────┬────────────┬────────────┬────────────┬───────┘ │
│           │            │            │            │          │
│  ┌────────▼───┐ ┌──────▼──────┐ ┌──▼────────┐ ┌▼────────┐│
│  │  Context   │ │    Tool     │ │  Safety   │ │ Memory  ││
│  │  Manager   │ │   System    │ │ Guardrail │ │ System  ││
│  └────────────┘ └─────────────┘ └───────────┘ └─────────┘│
│           │            │            │            │          │
│  ┌────────▼────────────▼────────────▼────────────▼───────┐ │
│  │                   Infrastructure Layer                 │ │
│  │  ┌─────────┐ ┌──────────┐ ┌─────────┐ ┌───────────┐  │ │
│  │  │   LLM   │ │  Sandbox │ │  File   │ │   Auth    │  │ │
│  │  │ Gateway │ │  Engine  │ │  Index  │ │  Service  │  │ │
│  │  └─────────┘ └──────────┘ └─────────┘ └───────────┘  │ │
│  └───────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

---

## 三、功能模块清单

### 模块 1：Core Agent Engine — 核心代理引擎

#### 1.1 意图解析器（Intent Parser）

```
职责: 将用户自然语言输入转化为结构化任务描述
```

- 多语言意图识别（中/英/日）
- 意图分类：
  - `CODE_GEN` — 代码生成
  - `FILE_OPS` — 文件操作
  - `DEBUG` — 调试排错
  - `REFACTOR` — 代码重构
  - `PROJECT_INIT` — 项目初始化
  - `TEST` — 测试编写与执行
  - `DEPLOY` — 部署操作
  - `QUERY` — 代码库查询
  - `EXPLAIN` — 代码解释
  - `REVIEW` — 代码审查
  - `CUSTOM` — 自定义技能触发
- 上下文补全（当指令模糊时主动追问）
- 指令消歧（同一指令多种理解时提供选项）

#### 1.2 规划引擎（Planner Engine）

```
职责: 将任务分解为可执行的步骤序列
```

**支持 4 种规划模式：**

| 模式              | 适用场景           | 特征                 |
| ----------------- | ------------------ | -------------------- |
| `ReAct`           | 探索性任务、调试   | 思考→行动→观察 循环  |
| `Plan-Execute`    | 明确的工程任务     | 先出计划，确认后执行 |
| `Tree-of-Thought` | 复杂决策           | 多路径并行探索       |
| `Reflexion`       | 需要自我纠错的任务 | 执行→反思→改进循环   |

**规划器功能：**
- 步骤依赖关系分析（DAG 建模）
- 并行步骤自动识别与并发执行
- 步骤优先级排序
- 失败回滚策略预定义
- 步骤资源预估（时间/Token/计算）
- 用户中断后的计划恢复
- 计划版本管理（回退到历史计划）

#### 1.3 执行引擎（Executor Engine）

```
职责: 逐步执行计划中的操作
```

- 顺序执行 / 并行执行 / 条件分支执行
- 每步执行前二次安全确认（可配置）
- 执行状态实时推送到终端
- 单步超时控制（默认 120s，可配置）
- 执行失败自动重试（指数退避，最多 3 次）
- 执行日志持久化
- 支持暂停/恢复执行
- 支持跳过当前步骤
- 支持手动接管（代理执行到一半转为手动）

#### 1.4 观察引擎（Observer Engine）

```
职责: 分析执行结果，决定下一步行动
```

- 输出截断与摘要（防止上下文溢出）
- 错误模式识别与自动修复建议
- 执行结果评分（成功/部分成功/失败）
- 异常检测（输出是否符合预期）
- 迭代次数上限控制（防止死循环，默认 25 轮）

---

### 模块 2：Tool System — 工具系统

#### 2.1 内置工具集

| 工具           | 功能                           | 安全等级 |
| -------------- | ------------------------------ | -------- |
| `shell`        | 执行 Shell 命令                | HIGH     |
| `file_read`    | 读取文件内容                   | LOW      |
| `file_write`   | 写入/创建文件                  | MEDIUM   |
| `file_edit`    | 精确编辑文件（行级）           | MEDIUM   |
| `file_delete`  | 删除文件（移入回收站）         | HIGH     |
| `dir_list`     | 列出目录内容                   | LOW      |
| `dir_create`   | 创建目录                       | LOW      |
| `search`       | 项目内文本搜索（grep/ripgrep） | LOW      |
| `git`          | Git 操作全套                   | MEDIUM   |
| `web_fetch`    | 抓取网页内容                   | LOW      |
| `web_search`   | 网络搜索                       | LOW      |
| `http_request` | 发送 HTTP 请求                 | MEDIUM   |
| `json_query`   | JSON/YAML 解析查询             | LOW      |
| `diff`         | 文件差异对比                   | LOW      |
| `docker`       | Docker 容器操作                | HIGH     |
| `database`     | 数据库查询（只读）             | MEDIUM   |
| `screenshot`   | 网页/应用截图                  | LOW      |
| `clipboard`    | 剪贴板读写                     | LOW      |
| `process`      | 进程管理（查看/终止）          | HIGH     |
| `env`          | 环境变量管理                   | MEDIUM   |
| `dependency`   | 包管理器操作                   | MEDIUM   |

#### 2.2 工具定义协议

```yaml
tool:
  name: "shell"
  description: "Execute shell commands"
  version: "1.0.0"
  safety_level: "HIGH"
  parameters:
    command:
      type: "string"
      description: "Shell command to execute"
      required: true
    timeout:
      type: "integer"
      description: "Timeout in seconds"
      default: 120
    working_dir:
      type: "string"
      description: "Working directory"
      default: "."
  returns:
    stdout: "string"
    stderr: "string"
    exit_code: "integer"
    duration: "float"
  confirm_before_execute: true
  dangerous_patterns:
    - "rm -rf /"
    - "mkfs"
    - "dd if="
    - "> /dev/"
    - ":(){ :|:& };:"
  blocked_patterns:
    - "curl.*|.*sh"
    - "wget.*|.*bash"
```

#### 2.3 自定义工具开发

```python
# tools/custom/my_tool.py
from mimo_cli.tools import BaseTool, tool_config

@tool_config(
    name="my_custom_tool",
    description="自定义工具描述",
    safety_level="MEDIUM"
)
class MyCustomTool(BaseTool):
    
    async def execute(self, params: dict, context: dict) -> dict:
        """
        params: LLM 传入的参数
        context: 当前执行上下文
        """
        # 业务逻辑
        result = await self._do_something(params)
        return {"output": result}
    
    def validate_params(self, params: dict) -> bool:
        """参数校验"""
        return "required_field" in params
```

---

### 模块 3：Skill System — 技能系统

#### 3.1 技能结构

```
~/.mimo/skills/
├── builtin/                    # 内置技能
│   ├── scaffold/
│   │   ├── SKILL.md
│   │   └── templates/
│   ├── git-workflow/
│   │   └── SKILL.md
│   ├── test-gen/
│   │   └── SKILL.md
│   ├── doc-gen/
│   │   └── SKILL.md
│   ├── refactor/
│   │   └── SKILL.md
│   ├── security-audit/
│   │   └── SKILL.md
│   ├── performance-profile/
│   │   └── SKILL.md
│   └── db-migration/
│       └── SKILL.md
├── community/                  # 社区技能（从市场安装）
│   ├── k8s-deploy/
│   ├── aws-lambda/
│   └── react-native/
└── custom/                     # 用户自定义技能
    └── my-deploy/
        └── SKILL.md
```

#### 3.2 SKILL.md 规范

```yaml
---
name: "scaffold"
display_name: "项目脚手架"
description: "快速生成项目结构和基础配置"
version: "2.1.0"
author: "MiMo Team"
tags: ["project", "init", "template"]
compatibility: ">=0.1.0"
triggers:
  - pattern: "创建.*项目"
    intent: "PROJECT_INIT"
  - pattern: "scaffold|boilerplate|template"
    intent: "PROJECT_INIT"
parameters:
  project_type:
    type: "enum"
    values: ["python", "node", "go", "rust", "java", "react", "vue", "nextjs"]
    required: true
  project_name:
    type: "string"
    required: true
dependencies: []
permissions:
  - "file_write"
  - "dir_create"
  - "shell"
---

## 行为指令

当用户触发此技能时：

### 步骤 1: 确认参数
询问用户项目名称、技术栈、是否需要以下可选功能：
- [ ] Docker 配置
- [ ] CI/CD 配置
- [ ] 代码规范（ESLint/Prettier/Black）
- [ ] 测试框架
- [ ] 文档模板

### 步骤 2: 生成项目
根据用户选择，使用 templates/ 目录下的模板生成项目结构。

### 步骤 3: 初始化依赖
执行包管理器的安装命令。

### 步骤 4: 初始化 Git
```bash
git init
git add .
git commit -m "chore: initial project scaffold"
```

### 步骤 5: 输出摘要
展示项目结构树和后续建议操作。
```

#### 3.3 技能市场

- 技能发布：`mimo skill publish ./my-skill`
- 技能安装：`mimo skill install @community/k8s-deploy`
- 技能搜索：`mimo skill search "docker deploy"`
- 技能评分与评论
- 技能版本管理与自动更新
- 技能安全扫描（发布前自动审查）

---

### 模块 4：Context Manager — 上下文管理

#### 4.1 多层级上下文

```
┌─────────────────────────────────────┐
│ Layer 5: Organization Knowledge     │  组织知识库（企业版）
├─────────────────────────────────────┤
│ Layer 4: Session History            │  当前会话历史
├─────────────────────────────────────┤
│ Layer 3: User Preferences           │  用户偏好与规则
├─────────────────────────────────────┤
│ Layer 2: Project Context            │  项目结构与配置
├─────────────────────────────────────┤
│ Layer 1: AGENT.md Rules             │  项目级代理指令
└─────────────────────────────────────┘
```

#### 4.2 上下文收集清单

| 来源 | 内容 | 优先级 |
|------|------|--------|
| `AGENT.md` | 项目级代理指令和规则 | P0 |
| `README.md` | 项目说明 | P0 |
| `package.json` / `pyproject.toml` / `go.mod` 等 | 依赖与元信息 | P0 |
| `.gitignore` | 忽略规则 | P1 |
| 目录树 | 项目结构（深度 3 层，自动忽略 node_modules 等） | P1 |
| Git 状态 | 当前分支、最近提交、未暂存变更 | P1 |
| `.env.example` | 环境变量模板 | P2 |
| `Makefile` / `Taskfile` | 可用命令 | P2 |
| CI 配置 | `.github/workflows/`、Jenkinsfile 等 | P2 |
| `tsconfig.json` / `eslint.config` 等 | 编码规范配置 | P2 |
| 依赖锁文件 | `package-lock.json`、`poetry.lock` 等 | P3 |
| `.editorconfig` | 编辑器配置 | P3 |

#### 4.3 上下文压缩策略

当上下文超过 Token 限制时的分层压缩：

```
策略 1: 目录树裁剪    — 去除深层嵌套，保留结构概览
策略 2: 文件内容摘要   — 大文件只保留前 N 行 + 函数签名
策略 3: 历史消息滑窗   — 保留最近 N 轮，早期对话压缩为摘要
策略 4: 智能相关性排序 — 根据当前任务只保留相关文件上下文
策略 5: 分块检索       — 向量化存储，按需检索相关片段
```

#### 4.4 AGENT.md 规范

```markdown
# AGENT.md — 项目代理配置

## 项目概述
这是一个基于 FastAPI 的后端服务项目。

## 技术栈
- Python 3.11
- FastAPI + SQLAlchemy + Alembic
- PostgreSQL 15
- Redis 7

## 编码规范
- 使用 type hints
- 函数/变量命名: snake_case
- 类命名: PascalCase
- 所有公共函数必须有 docstring
- 错误处理使用自定义异常类

## 项目结构
- `app/api/` — API 路由
- `app/core/` — 核心配置
- `app/models/` — 数据模型
- `app/services/` — 业务逻辑
- `app/utils/` — 工具函数
- `tests/` — 测试文件

## 常用命令
- 启动: `uvicorn app.main:app --reload`
- 测试: `pytest tests/ -v`
- 迁移: `alembic upgrade head`
- 代码检查: `ruff check .`

## 代理规则
- 修改数据库模型后，必须生成并执行迁移
- 新增 API 端点必须编写对应测试
- 不要直接修改 main.py，通过路由注册
- 提交前确保所有测试通过

## 禁止操作
- 不要修改 .env 文件
- 不要删除 alembic/versions/ 下的迁移文件
- 不要修改 CI 配置
```

---

### 模块 5：LLM Gateway — 模型网关

#### 5.1 多模型支持

```yaml
# ~/.mimo/config.yaml
models:
  default: "mimo"
  
  providers:
    mimo:
      api_base: "https://api.mimo.xiaomi.com/v1"
      api_key: "${MIMO_API_KEY}"
      model: "mimo-v2"
      max_tokens: 32768
      temperature: 0.3
      
    deepseek:
      api_base: "https://api.deepseek.com/v1"
      api_key: "${DEEPSEEK_API_KEY}"
      model: "deepseek-coder"
      max_tokens: 32768
      
    openai:
      api_base: "https://api.openai.com/v1"
      api_key: "${OPENAI_API_KEY}"
      model: "gpt-4o"
      max_tokens: 32768
      
    anthropic:
      api_base: "https://api.anthropic.com"
      api_key: "${ANTHROPIC_API_KEY}"
      model: "claude-sonnet-4-20250514"
      max_tokens: 32768
      
    ollama:
      api_base: "http://localhost:11434/v1"
      model: "codellama:34b"
      max_tokens: 8192
      
    # 自定义 OpenAI 兼容端点
    custom:
      api_base: "http://my-server:8000/v1"
      api_key: "${CUSTOM_API_KEY}"
      model: "my-model"
      max_tokens: 16384
```

#### 5.2 模型路由策略

```yaml
routing:
  # 不同任务类型使用不同模型
  strategy: "task-based"  # task-based | round-robin | cost-optimized | latency-optimized
  
  rules:
    - task: "CODE_GEN"
      model: "mimo"
      fallback: "deepseek"
    - task: "EXPLAIN"
      model: "mimo"
      fallback: "openai"
    - task: "REVIEW"
      model: "mimo"
      fallback: "anthropic"
    - task: "SIMPLE_QUERY"
      model: "ollama"  # 简单任务用本地模型节省成本
```

#### 5.3 流式输出

- SSE (Server-Sent Events) 流式传输
- 支持中途取消（Ctrl+C 优雅中断）
- Token 使用量实时统计显示
- 流式 Markdown 渲染（代码块高亮）
- 思考过程可视化（如果模型支持 thinking/reasoning）

---

### 模块 6：Safety Guardrail — 安全护栏

#### 6.1 三级安全策略

```
┌─────────────────────────────────────────────┐
│ Level 3: LOCKDOWN（锁定模式）                │
│ - 仅允许只读操作                             │
│ - 适用于生产环境 / 敏感项目                   │
├─────────────────────────────────────────────┤
│ Level 2: CONFIRM（确认模式）【默认】          │
│ - 危险操作需用户确认                         │
│ - 文件写入前显示 diff                        │
│ - Shell 命令执行前显示预览                    │
├─────────────────────────────────────────────┤
│ Level 1: AUTO（自动模式）                    │
│ - 自动执行所有操作                            │
│ - 仅保留不可逆操作的确认                      │
│ - 适用于可信环境/CI 场景                      │
└─────────────────────────────────────────────┘
```

#### 6.2 危险操作分类

| 等级       | 操作类型                                   | 处理方式                  |
| ---------- | ------------------------------------------ | ------------------------- |
| `CRITICAL` | `rm -rf /`、格式化磁块、写入系统文件       | 始终拦截                  |
| `HIGH`     | 删除文件、`git push --force`、修改系统配置 | 强制确认 + 延迟执行（5s） |
| `MEDIUM`   | 写入文件、执行未知脚本、网络请求           | 确认后执行                |
| `LOW`      | 读取文件、查看 Git 状态、目录列表          | 直接执行                  |

#### 6.3 安全功能清单

- [ ] 命令黑名单 / 正则拦截
- [ ] 文件路径白名单（限制可操作目录）
- [ ] 文件写入前自动备份到 `.mimo/backups/`
- [ ] Git 操作保护分支列表（`main`, `master`, `release/*`）
- [ ] 网络请求域名白名单
- [ ] 敏感文件检测（`.env`、密钥文件、证书）自动拒绝读取/写入
- [ ] 代码注入检测（防止生成含恶意代码的内容）
- [ ] 操作审计日志（所有操作记录到 `~/.mimo/audit.log`）
- [ ] 磁盘写入量限制（单次会话最大写入量）
- [ ] Token 消耗预算限制（防止意外大量消耗）

---

### 模块 7：Memory System — 记忆系统

#### 7.1 三层记忆架构

```
短期记忆 (Working Memory)
├── 当前会话对话历史
├── 当前任务执行状态
└── 临时文件/变量缓存
    生命周期: 会话内

中期记忆 (Session Memory)
├── 项目知识图谱
├── 用户编码偏好
├── 历史任务与结果
└── 学习到的项目模式
    生命周期: 跨会话持久化

长期记忆 (Knowledge Base)
├── 用户规则库 (rules.md)
├── 代码片段库
├── 错误模式库
└── 解决方案库
    生命周期: 永久（可手动清理）
```

#### 7.2 记忆功能

- 会话历史持久化（`~/.mimo/sessions/`）
- 会话恢复（`mimo resume <session-id>`）
- 会话列表查看（`mimo sessions list`）
- 用户规则学习（多次纠正后自动提取规则）
- 项目模式识别（识别项目的编码风格和惯例）
- 向量化存储（用于语义检索历史上下文）
- 记忆导出/导入（团队共享知识）

---

### 模块 8：TUI Interface — 终端界面

#### 8.1 交互模式

```bash
# 单次模式（适合脚本/CI）
mimo run "创建一个 FastAPI 用户管理 API"

# 交互模式（默认）
mimo

# 管道模式
cat error.log | mimo "分析这个错误日志"

# 续接上次会话
mimo resume

# 指定技能
mimo --skill scaffold "创建 Python 项目 my-app"

# 指定模型
mimo --model deepseek "重构这个函数"

# 指定安全级别
mimo --safety lockdown "查看项目结构"
```

#### 8.2 界面组件

```
┌─ MiMo CLI v0.1.0 ───────────────────────────── Model: mimo-v2 ─┐
│                                                                │
│  [user] 创建一个 REST API，用户 CRUD + JWT 认证                  │
│                                                                  │
│  [mimo] 理解你的需求。我来规划一下执行步骤：                      │
│                                                                  │
│  ┌─ 执行计划 ──────────────────────────────────────────────────┐ │
│  │  ✅ 1. 创建项目结构                      [complete] 2.3s    │ │
│  │  ✅ 2. 安装依赖                          [complete] 8.1s    │ │
│  │  🔄 3. 编写数据模型                      [running...]       │ │
│  │  ⬜ 4. 编写 API 路由                     [pending]          │ │
│  │  ⬜ 5. 实现 JWT 认证                     [pending]          │ │
│  │  ⬜ 6. 编写测试                          [pending]          │ │
│  │  ⬜ 7. 运行测试验证                      [pending]          │ │
│  └──────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  ┌─ 文件变更 ──────────────────────────────────────────────────┐ │
│  │  + app/models/user.py                                      │ │
│  │  ~ app/main.py                     (3 lines changed)        │ │
│  │  + app/schemas/user.py                                     │ │
│  └──────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  ──────────────────────────────────────────────────────────────  │
│  Tokens: 2,341 in / 892 out  │  Cost: ¥0.05  │  Time: 12.4s    │
└──────────────────────────────────────────────────────────────────┘
```

#### 8.3 TUI 功能

- 键盘快捷键：
  - `Ctrl+C` — 取消当前操作
  - `Ctrl+L` — 清屏
  - `Ctrl+R` — 搜索历史命令
  - `Ctrl+D` — 接受代理建议
  - `Tab` — 自动补全
  - `↑/↓` — 浏览历史
  - `Ctrl+P` — 命令面板
- 代码语法高亮
- Markdown 实时渲染
- 表格对齐显示
- 进度条与 spinner 动画
- 主题切换（暗色/亮色/自定义）
- 分屏模式（同时查看文件和对话）
- Vim 模式支持

---

### 模块 9：Project Awareness — 项目感知

#### 9.1 智能项目检测

```python
PROJECT_SIGNATURES = {
    "python":     ["pyproject.toml", "setup.py", "requirements.txt", "Pipfile"],
    "node":       ["package.json", "yarn.lock", "pnpm-lock.yaml"],
    "go":         ["go.mod", "go.sum"],
    "rust":       ["Cargo.toml"],
    "java":       ["pom.xml", "build.gradle", "build.gradle.kts"],
    "dotnet":     ["*.csproj", "*.sln"],
    "ruby":       ["Gemfile"],
    "php":        ["composer.json"],
    "swift":      ["Package.swift", "*.xcodeproj"],
    "kotlin":     ["build.gradle.kts"],
    "flutter":    ["pubspec.yaml"],
    "terraform":  ["*.tf"],
    "docker":     ["Dockerfile", "docker-compose.yml"],
}
```

#### 9.2 自动配置推断

- 检测到 `package.json` → 读取 scripts 字段，了解可用命令
- 检测到 `.eslintrc` → 了解代码风格规范
- 检测到 `pytest.ini` → 了解测试配置
- 检测到 `Makefile` → 了解构建命令
- 检测到 `.github/workflows/` → 了解 CI 流程

#### 9.3 框架特化

| 框架             | 特化能力                           |
| ---------------- | ---------------------------------- |
| React / Next.js  | 组件生成、路由理解、状态管理       |
| Vue / Nuxt       | SFC 生成、组合式 API               |
| Django / FastAPI | Model/View/Serializer 生成         |
| Spring Boot      | Controller/Service/Repository 生成 |
| Express / NestJS | 中间件、路由、DTO 生成             |
| Flutter          | Widget 生成、状态管理              |

---

### 模块 10：Git Integration — Git 深度集成

#### 10.1 Git 操作清单

```bash
# 基础操作
mimo git status
mimo git diff
mimo git log --oneline -20
mimo git blame src/main.py

# 智能提交
mimo git commit          # 自动生成 commit message
mimo git commit --fix    # 基于最近修改自动修复式提交

# 分支管理
mimo git branch create feature/user-auth
mimo git branch switch main
mimo git branch delete old-feature

# PR/MR 生成
mimo git pr create       # 自动生成 PR 标题和描述
mimo git pr review       # 代码审查

# 冲突解决
mimo git merge resolve   # 智能合并冲突解决

# 历史分析
mimo git blame explain   # 解释某行代码的变更历史
mimo git history "这个函数是怎么演变的"
```

#### 10.2 Commit Message 自动生成

基于 diff 内容自动生成符合 Conventional Commits 规范的提交信息：

```
feat(auth): add JWT token refresh endpoint

- Implement token refresh logic with rotation
- Add rate limiting for refresh requests
- Include refresh token in response payload

Closes #142
```

#### 10.3 PR 描述自动生成

- 变更摘要
- 影响范围分析
- 测试覆盖情况
- Breaking Changes 提示
- 截图/UI 变更预览（如有）

---

### 模块 11：Testing Integration — 测试集成

#### 11.1 测试生成

```
mimo test generate app/services/user_service.py
```

- 自动生成单元测试
- 自动生成集成测试
- 边界条件覆盖
- Mock 对象生成
- 测试数据工厂生成

#### 11.2 测试执行与分析

```
mimo test run                    # 运行所有测试
mimo test run --coverage         # 带覆盖率
mimo test run --failed-only      # 只运行失败的测试
mimo test fix                    # 自动修复失败的测试
mimo test analyze                # 分析测试质量
```

#### 11.3 TDD 模式

```
mimo tdd "用户注册功能"
```

自动进入测试驱动开发循环：
1. 先写失败的测试
2. 实现最小代码让测试通过
3. 重构
4. 重复

---

### 模块 12：Debug Assistant — 调试助手

#### 12.1 错误分析

```
mimo debug "程序报错 KeyError: 'user_id'"
cat traceback.log | mimo debug
```

- 堆栈跟踪智能解析
- 根因分析
- 修复方案建议（多个候选）
- 相关文档/StackOverflow 链接

#### 12.2 日志分析

```
mimo log analyze /var/log/app.log --last 1h
mimo log analyze /var/log/app.log --level error
```

- 错误模式聚类
- 异常时间线重建
- 性能瓶颈识别

#### 12.3 交互式调试

```
mimo debug interactive
```

- 设置断点
- 单步执行模拟
- 变量值追踪
- 调用栈可视化

---

### 模块 13：Code Review — 代码审查

```
mimo review                    # 审查未提交的变更
mimo review --pr 142          # 审查指定 PR
mimo review --since HEAD~5    # 审查最近 5 个提交
mimo review --file src/main.py # 审查指定文件
```

#### 13.1 审查维度

| 维度     | 检查内容                            |
| -------- | ----------------------------------- |
| 正确性   | 逻辑错误、边界条件、空值处理        |
| 安全性   | SQL 注入、XSS、硬编码密钥、权限检查 |
| 性能     | N+1 查询、内存泄漏、不必要的循环    |
| 可维护性 | 函数长度、命名规范、注释质量        |
| 测试覆盖 | 新代码是否有对应测试                |
| 代码风格 | 是否符合项目规范                    |
| 依赖安全 | 新依赖是否有已知漏洞                |

#### 13.2 审查输出

```
┌─ Code Review Report ─────────────────────────────────────────┐
│                                                              │
│  Summary: 3 issues found (1 critical, 2 suggestions)        │
│                                                              │
│  🔴 CRITICAL │ src/services/auth.py:45                       │
│  │ SQL injection vulnerability: string concatenation in     │
│  │ query. Use parameterized queries instead.                │
│  │                                                          │
│  │  - query = f"SELECT * FROM users WHERE email='{email}'" │
│  │  + query = "SELECT * FROM users WHERE email=%s"          │
│  │  + cursor.execute(query, (email,))                       │
│                                                              │
│  🟡 SUGGESTION │ src/utils/cache.py:12                      │
│  │ Cache key collision risk: consider adding namespace.     │
│                                                              │
│  🟡 SUGGESTION │ src/models/order.py:78                     │
│  │ Consider using enum instead of magic strings for status. │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

---

### 模块 14：Docker / Sandbox — 沙箱执行

#### 14.1 隔离执行环境

```yaml
sandbox:
  enabled: true
  provider: "docker"        # docker | firecracker | wasm
  
  image: "mimo/sandbox:latest"
  
  resource_limits:
    cpu: "2"
    memory: "2Gi"
    disk: "10Gi"
    timeout: "300s"
    
  network:
    mode: "restricted"      # unrestricted | restricted | none
    allowed_hosts:
      - "registry.npmjs.org"
      - "pypi.org"
      - "github.com"
      
  volumes:
    - source: "${PROJECT_DIR}"
      target: "/workspace"
      readonly: false
      
  environment:
    - "NODE_ENV=development"
```

#### 14.2 沙箱功能

- 在容器内执行不确定安全性的代码
- 依赖安装隔离（不影响宿主环境）
- 测试执行隔离
- 自动清理（会话结束后销毁）
- 资源超限自动终止
- 执行快照（可回放）

---

### 模块 15：MCP Protocol — 模型上下文协议支持

#### 15.1 MCP 客户端

支持连接 MCP 服务器，扩展代理能力：

```yaml
mcp:
  servers:
    - name: "filesystem"
      transport: "stdio"
      command: "npx"
      args: ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/dir"]
      
    - name: "github"
      transport: "stdio"
      command: "npx"
      args: ["-y", "@modelcontextprotocol/server-github"]
      env:
        GITHUB_TOKEN: "${GITHUB_TOKEN}"
        
    - name: "postgres"
      transport: "stdio"
      command: "npx"
      args: ["-y", "@modelcontextprotocol/server-postgres"]
      env:
        DATABASE_URL: "${DATABASE_URL}"
        
    - name: "custom-api"
      transport: "sse"
      url: "http://localhost:3000/mcp"
```

#### 15.2 MCP 服务器

同时作为 MCP 服务器暴露 MiMo CLI 的能力，供其他工具调用：

```yaml
mcp_server:
  enabled: true
  transport: "stdio"
  capabilities:
    - "file_read"
    - "file_write"
    - "shell"
    - "search"
    - "git"
```

---

### 模块 16：Configuration — 配置系统

#### 16.1 配置文件层级

```
优先级从高到低:

1. 命令行参数         mimo --model deepseek "..."
2. 项目级配置         ./mimo.yaml
3. 用户级配置         ~/.mimo/config.yaml
4. 系统级配置         /etc/mimo/config.yaml
5. 内置默认值
```

#### 16.2 完整配置示例

```yaml
# ~/.mimo/config.yaml

# 基本配置
default_model: "mimo"
language: "zh-CN"              # zh-CN | en-US | ja-JP
theme: "dark"                  # dark | light | auto
editor: "code"                 # 默认编辑器
shell: "/bin/zsh"              # 默认 Shell

# 模型配置
models:
  mimo:
    api_base: "https://api.mimo.xiaomi.com/v1"
    api_key: "${MIMO_API_KEY}"
    model: "mimo-v2"
    max_tokens: 32768
    temperature: 0.3
    top_p: 0.95
    
# 安全配置
safety:
  level: "confirm"             # lockdown | confirm | auto
  backup_before_write: true
  backup_dir: ".mimo/backups"
  audit_log: "~/.mimo/audit.log"
  blocked_commands:
    - "sudo"
    - "chmod 777"
  protected_files:
    - ".env"
    - "*.pem"
    - "*.key"
    - "id_rsa*"
  protected_branches:
    - "main"
    - "master"
    - "release/*"
  max_file_write_size: "1MB"
  max_session_cost: "¥10"
  
# 上下文配置
context:
  max_tokens: 128000
  directory_tree_depth: 3
  ignore_patterns:
    - "node_modules"
    - ".git"
    - "__pycache__"
    - "dist"
    - "build"
    - ".venv"
    - "vendor"
    - "*.min.js"
    - "*.min.css"
    
# 代理行为
agent:
  max_iterations: 25
  max_parallel_tools: 5
  planning_mode: "plan-execute" # react | plan-execute | auto
  auto_confirm_low_risk: true
  show_token_usage: true
  show_cost: true
  verbose: false
  
# 会话配置
sessions:
  auto_save: true
  max_history: 50
  storage_dir: "~/.mimo/sessions"
  
# 记忆配置
memory:
  enabled: true
  vector_store: "sqlite"       # sqlite | chroma | pinecone
  vector_store_path: "~/.mimo/memory.db"
  embedding_model: "text-embedding-3-small"
  
# 技能配置
skills:
  directories:
    - "~/.mimo/skills/builtin"
    - "~/.mimo/skills/community"
    - "~/.mimo/skills/custom"
    - "./.mimo/skills"         # 项目级技能
  auto_update: true
  marketplace_url: "https://skills.mimo.xiaomi.com"
  
# MCP 配置
mcp:
  servers: []
  
# 沙箱配置
sandbox:
  enabled: false
  provider: "docker"
  
# 忽略模式
ignore:
  - ".git/"
  - "node_modules/"
  - "__pycache__/"
  - "*.pyc"
  - ".DS_Store"

# 自定义别名
aliases:
  "c": "commit"
  "r": "review"
  "t": "test run"
  "d": "debug"
```

---

### 模块 17：Plugin System — 插件系统

#### 17.1 插件类型

| 类型         | 说明             | 示例                    |
| ------------ | ---------------- | ----------------------- |
| `tool`       | 新增工具能力     | 数据库查询、K8s 管理    |
| `skill`      | 新增技能         | 代码迁移、国际化        |
| `theme`      | TUI 主题         | 自定义颜色和布局        |
| `model`      | 自定义模型提供商 | 私有部署模型            |
| `hook`       | 生命周期钩子     | 提交前 lint、部署前测试 |
| `middleware` | 请求/响应中间件  | 日志记录、Token 计费    |

#### 17.2 插件开发

```python
# my_plugin/__init__.py
from mimo_cli.plugins import Plugin, plugin_metadata

@plugin_metadata(
    name="k8s-assistant",
    version="1.0.0",
    description="Kubernetes 集群管理助手",
    author="Community",
    hooks=["before_execute", "after_execute"]
)
class K8sAssistant(Plugin):
    
    def register_tools(self, registry):
        registry.register("k8s_apply", self.k8s_apply)
        registry.register("k8s_get", self.k8s_get)
        registry.register("k8s_logs", self.k8s_logs)
    
    def register_skills(self, registry):
        registry.register("k8s-deploy", "skills/k8s-deploy")
    
    def before_execute(self, context):
        # 执行前检查 kubectl 连接
        pass
    
    def after_execute(self, context, result):
        # 执行后记录操作
        pass
```

---

### 模块 18：CI/CD Integration — 持续集成

#### 18.1 命令行模式（适合 CI）

```yaml
# GitHub Actions 示例
- name: AI Code Review
  run: |
    mimo review \
      --pr ${{ github.event.pull_request.number }} \
      --output github-pr-comment \
      --safety lockdown \
      --non-interactive

# 生成文档
- name: Generate Docs
  run: |
    mimo run "为所有公共 API 生成 OpenAPI 文档" \
      --safety auto \
      --output file:docs/api.yaml
```

#### 18.2 Git Hooks 集成

```bash
#!/bin/bash
# .git/hooks/pre-commit

# 自动运行代码审查
mimo review --staged --safety lockdown --non-interactive
if [ $? -ne 0 ]; then
    echo "Code review failed. Please fix issues before committing."
    exit 1
fi
```

#### 18.3 CI 支持功能

- 非交互模式（`--non-interactive`）
- 输出格式化（JSON、Markdown、GitHub PR Comment）
- 退出码语义化（0=成功，1=失败，2=需要人工审核）
- CI Token 认证
- PR Bot（自动审查 PR 并评论）

---

### 模块 19：Multi-Agent — 多代理协作（高级）

#### 19.1 代理角色

```yaml
agents:
  architect:
    model: "mimo"
    role: "系统架构师"
    skills: ["system-design", "api-design"]
    
  developer:
    model: "mimo"
    role: "开发工程师"
    skills: ["code-gen", "refactor", "debug"]
    
  reviewer:
    model: "mimo"
    role: "代码审查员"
    skills: ["code-review", "security-audit"]
    
  tester:
    model: "mimo"
    role: "测试工程师"
    skills: ["test-gen", "test-run"]
    
  devops:
    model: "mimo"
    role: "运维工程师"
    skills: ["deploy", "monitor", "docker"]
```

#### 19.2 协作模式

```
# 自动多代理模式
mimo team "设计并实现一个微服务电商平台"

# 流程:
# 1. architect 设计系统架构
# 2. developer 按模块并行开发
# 3. reviewer 逐模块审查
# 4. tester 生成并运行测试
# 5. devops 编写部署配置
# 6. architect 最终验收
```

---

### 模块 20：Telemetry & Update — 遥测与更新

#### 20.1 使用统计（Opt-in）

```yaml
telemetry:
  enabled: false               # 默认关闭
  
  # 开启后收集（匿名）:
  # - 会话数 / 任务类型分布
  # - 平均执行时间
  # - 错误率
  # - 模型使用分布
  # 不收集:
  # - 代码内容
  # - 对话内容
  # - 文件名/路径
  # - 用户身份信息
```

#### 20.2 自动更新

```bash
mimo update                    # 检查并更新
mimo update --check            # 仅检查
mimo update --rollback         # 回退版本
mimo version                   # 查看当前版本
```

---

## 四、命令参考手册

```bash
# ═══════════════════════════════════════
# MiMo CLI 命令参考
# ═══════════════════════════════════════

# 基础
mimo                            # 进入交互模式
mimo run "<task>"               # 单次任务执行
mimo resume [session-id]        # 恢复会话
mimo version                    # 查看版本
mimo update                     # 更新

# 任务
mimo code "<description>"       # 生成代码
mimo refactor "<target>"        # 重构代码
mimo fix "<error>"              # 修复错误
mimo explain "<code-or-file>"   # 解释代码
mimo optimize "<target>"        # 性能优化

# 文件操作
mimo read <file>                # 读取文件
mimo write <file> "<content>"   # 写入文件
mimo edit <file> "<changes>"    # 编辑文件
mimo find "<pattern>"           # 搜索文件
mimo diff <file1> <file2>       # 文件对比

# Git
mimo git status                 # 状态查看
mimo git commit                 # 智能提交
mimo git pr create              # 创建 PR
mimo git pr review [number]     # 审查 PR

# 代码质量
mimo review [target]            # 代码审查
mimo lint [target]              # 代码检查
mimo format [target]            # 代码格式化

# 测试
mimo test run [target]          # 运行测试
mimo test generate <file>       # 生成测试
mimo test fix                   # 修复失败测试
mimo test coverage              # 覆盖率报告

# 调试
mimo debug "<error>"            # 错误分析
mimo debug interactive          # 交互调试
mimo log analyze <file>         # 日志分析

# 项目
mimo init                       # 初始化 AGENT.md
mimo scaffold "<type>" "<name>" # 项目脚手架
mimo info                       # 项目信息

# 技能
mimo skill list                 # 查看技能列表
mimo skill install <name>       # 安装技能
mimo skill publish <path>       # 发布技能
mimo skill search "<query>"     # 搜索技能

# 会话
mimo sessions list              # 会话列表
mimo sessions resume <id>       # 恢复会话
mimo sessions delete <id>       # 删除会话
mimo sessions export <id>       # 导出会话

# 配置
mimo config show                # 显示配置
mimo config set <key> <value>   # 设置配置
mimo config edit                # 编辑配置文件
mimo config reset               # 重置配置

# 模型
mimo model list                 # 查看可用模型
mimo model switch <name>        # 切换模型
mimo model test <name>          # 测试模型连接

# 记忆
mimo memory list                # 查看记忆
mimo memory clear               # 清除记忆
mimo memory export              # 导出记忆
mimo memory import <file>       # 导入记忆

# MCP
mimo mcp list                   # 查看 MCP 服务器
mimo mcp add <config>           # 添加 MCP 服务器
mimo mcp remove <name>          # 移除 MCP 服务器

# 管道
echo "content" | mimo "<task>"  # 管道输入
cat file | mimo explain         # 文件输入

# 通用选项
--model <name>                  # 指定模型
--safety <level>                # 安全级别
--skill <name>                  # 使用技能
--output <format>               # 输出格式 (text|json|markdown)
--non-interactive               # 非交互模式
--verbose                       # 详细输出
--quiet                         # 静默模式
--no-color                      # 禁用颜色
--help                          # 帮助信息
```

---

## 五、技术选型

### 语言选择：Go

| 考量   | 选择理由                       |
| ------ | ------------------------------ |
| 性能   | 启动速度快，内存占用低         |
| 分发   | 单二进制，无运行时依赖         |
| 并发   | goroutine 原生支持并行工具调用 |
| 生态   | Cobra/Rich 标准库丰富          |
| 跨平台 | Linux/macOS/Windows 统一体验   |

> 备选：Rust（性能极致但开发速度慢）、Python + Typer（开发最快但分发困难）、TypeScript + Ink（生态好但启动慢）

### 依赖清单

```
核心依赖:
├── cobra              — CLI 框架
├── bubbletea          — TUI 框架 (charmbracelet)
├── lipgloss           — 终端样式
├── glamour            — Markdown 渲染
├── chroma             — 代码高亮
├── go-resty           — HTTP 客户端
├── yaml.v3            — YAML 解析
├── sqlite             — 本地存储
├── go-git             — Git 操作
├── websocket          — WebSocket
└── cobra-cobra        — 自动补全

可选依赖:
├── docker-sdk         — Docker API
├── go-mcp             — MCP 协议
└── chromem-go         — 向量存储
```

---

## 六、开发路线图

### v0.1.0 — MVP（第 1-4 周）

- [ ] 项目骨架搭建（Go + Cobra + BubbleTea）
- [ ] 基础交互模式（对话式输入输出）
- [ ] LLM API 对接（MiMo + OpenAI 兼容接口）
- [ ] 流式输出渲染
- [ ] 基础工具：shell、file_read、file_write、file_edit
- [ ] 简单的安全确认机制
- [ ] 配置文件系统
- [ ] `mimo run` 单次执行模式

### v0.2.0 — 核心能力（第 5-8 周）

- [ ] ReAct + Plan-Execute 规划引擎
- [ ] 完整工具集（20+ 工具）
- [ ] AGENT.md 解析与上下文收集
- [ ] Git 集成（status/diff/commit）
- [ ] 会话持久化与恢复
- [ ] 多模型切换
- [ ] 安全护栏三级策略

### v0.3.0 — 技能与扩展（第 9-12 周）

- [ ] 技能系统（加载/执行/市场）
- [ ] 代码审查功能
- [ ] 测试生成与执行
- [ ] 调试助手
- [ ] 插件系统基础
- [ ] MCP 客户端

### v0.4.0 — 进阶功能（第 13-16 周）

- [ ] 记忆系统（向量化存储）
- [ ] 沙箱执行（Docker）
- [ ] CI/CD 集成
- [ ] IDE 扩展（VS Code）
- [ ] MCP 服务器模式
- [ ] 多代理协作原型

### v1.0.0 — 正式发布（第 17-20 周）

- [ ] 性能优化
- [ ] 完善文档
- [ ] 安全审计
- [ ] 跨平台测试
- [ ] 技能市场上线
- [ ] 社区建设

---

## 七、项目文件结构

```
mimo-cli/
├── .github/
│   └── workflows/
│       ├── ci.yml
│       ├── release.yml
│       └── pr-review.yml
├── cmd/
│   ├── root.go                 # 根命令
│   ├── run.go                  # mimo run
│   ├── interactive.go          # 交互模式
│   ├── resume.go               # 会话恢复
│   ├── skill.go                # 技能管理
│   ├── config.go               # 配置管理
│   ├── git.go                  # Git 子命令
│   ├── test.go                 # 测试子命令
│   ├── review.go               # 代码审查
│   ├── debug.go                # 调试助手
│   ├── memory.go               # 记忆管理
│   └── version.go              # 版本信息
├── internal/
│   ├── agent/
│   │   ├── agent.go            # 代理核心
│   │   ├── planner.go          # 规划引擎
│   │   ├── executor.go         # 执行引擎
│   │   ├── observer.go         # 观察引擎
│   │   └── intent.go           # 意图解析
│   ├── tools/
│   │   ├── registry.go         # 工具注册表
│   │   ├── shell.go
│   │   ├── file.go
│   │   ├── git.go
│   │   ├── search.go
│   │   ├── web.go
│   │   ├── http.go
│   │   ├── docker.go
│   │   └── database.go
│   ├── skills/
│   │   ├── loader.go
│   │   ├── executor.go
│   │   ├── marketplace.go
│   │   └── builtin/
│   │       ├── scaffold/
│   │       ├── git-workflow/
│   │       ├── test-gen/
│   │       └── ...
│   ├── context/
│   │   ├── manager.go
│   │   ├── project.go
│   │   ├── compressor.go
│   │   └── agent_md.go
│   ├── llm/
│   │   ├── gateway.go          # 模型网关
│   │   ├── provider.go         # 提供商接口
│   │   ├── openai.go           # OpenAI 兼容
│   │   ├── anthropic.go        # Anthropic 兼容
│   │   ├── ollama.go           # Ollama 本地
│   │   └── router.go           # 模型路由
│   ├── safety/
│   │   ├── guardrail.go
│   │   ├── classifier.go
│   │   ├── backup.go
│   │   └── audit.go
│   ├── memory/
│   │   ├── store.go
│   │   ├── session.go
│   │   ├── vector.go
│   │   └── rules.go
│   ├── sandbox/
│   │   ├── docker.go
│   │   └── interface.go
│   ├── mcp/
│   │   ├── client.go
│   │   └── server.go
│   ├── config/
│   │   ├── config.go
│   │   └── schema.go
│   ├── tui/
│   │   ├── app.go
│   │   ├── chat.go
│   │   ├── plan.go
│   │   ├── diff.go
│   │   ├── progress.go
│   │   ├── theme.go
│   │   └── keybind.go
│   └── plugin/
│       ├── manager.go
│       └── interface.go
├── pkg/
│   ├── logger/
│   ├── util/
│   └── version/
├── skills/                     # 内置技能资源
├── themes/                     # TUI 主题
├── docs/
│   ├── ARCHITECTURE.md
│   ├── CONTRIBUTING.md
│   ├── SKILL_DEV_GUIDE.md
│   └── PLUGIN_DEV_GUIDE.md
├── tests/
│   ├── e2e/
│   ├── integration/
│   └── unit/
├── scripts/
│   ├── install.sh
│   ├── uninstall.sh
│   └── release.sh
├── AGENT.md
├── CHANGELOG.md
├── LICENSE
├── Makefile
├── README.md
└── go.mod
```

---

## 八、非功能需求

| 维度           | 要求                                                      |
| -------------- | --------------------------------------------------------- |
| **启动时间**   | 冷启动 < 200ms                                            |
| **响应延迟**   | 首 Token 延迟 < 500ms（取决于模型）                       |
| **内存占用**   | 空闲 < 30MB，活跃 < 200MB                                 |
| **二进制大小** | < 50MB（不含内置技能资源）                                |
| **跨平台**     | Linux (amd64/arm64), macOS (amd64/arm64), Windows (amd64) |
| **可测试性**   | 核心模块单元测试覆盖率 > 80%                              |
| **可观测性**   | 结构化日志 + 操作审计                                     |
| **国际化**     | 所有用户界面字符串支持 i18n                               |
| **无障碍**     | 终端无障碍支持（screen reader 兼容）                      |
| **许可证**     | Apache 2.0                                                |

---

## 九、命名与品牌

```
命令名:     mimo
全称:       MiMo CLI
Slogan:     "Your AI pair programmer in the terminal"
配色:       小米橙 (#FF6900) + 深灰 (#1A1A1A)
图标:       终端 + 闪电（象征效率）
```

---

