# MiMo CLI

> **你的终端中的 AI 结对编程伙伴** 🤖💻

MiMo CLI 是一款基于大语言模型的代理式命令行开发工具。通过自然语言描述，MiMo CLI 能够自主完成代码编写、文件操作、项目构建、测试调试等开发全流程任务。

## ✨ 核心特性

| 特性 | 描述 |
|------|------|
| 🧠 **智能代理** | 基于 MiMo 大模型，理解自然语言指令并自主执行任务 |
| 🔧 **丰富工具集** | 20+ 内置工具，覆盖文件操作、Git、网络、数据库等开发场景 |
| 🎯 **多模型支持** | 支持 MiMo、OpenAI、Anthropic、DeepSeek 等多模型热切换 |
| 🛡️ **安全可靠** | 分级安全策略 + 沙箱执行 + 操作审计 |
| 🚀 **高效执行** | 并行任务执行 + 流式输出 + 智能上下文管理 |
| 🌍 **中英双语** | 原生中英双语支持，无语言障碍 |
| 🔌 **可扩展** | 开放技能系统 + 插件架构 + MCP 协议支持 |

## 🚀 快速开始

### 安装

```bash
# 使用 Go 安装
go install github.com/mimo-cli/mimo-cli@latest

# 或者下载预编译二进制
curl -fsSL https://install.mimo.xiaomi.com | sh
```

### 配置

```bash
# 设置 API 密钥
export MIMO_API_KEY="your-api-key-here"

# 或使用配置文件
mimo config set model.api_key "your-api-key-here"
```

### 基本用法

```bash
# 进入交互模式
mimo

# 单次任务执行
mimo run "创建一个 FastAPI 用户管理 API"

# 代码审查
mimo review --file src/main.py

# 测试生成
mimo test generate app/services/user_service.py

# 智能提交
mimo git commit
```

## 📖 使用示例

### 1. 项目脚手架

```bash
# 创建 Python FastAPI 项目
mimo scaffold python my-api

# 创建 React 前端项目
mimo scaffold react my-frontend
```

### 2. 代码生成与重构

```bash
# 生成 REST API
mimo code "用户注册登录 API，包含 JWT 认证"

# 重构现有代码
mimo refactor "将这个函数重构为使用策略模式"

# 修复错误
mimo fix "程序报错 KeyError: 'user_id'"
```

### 3. 测试与调试

```bash
# 生成测试用例
mimo test generate src/services/payment.py

# 运行测试
mimo test run --coverage

# 调试错误
mimo debug "程序报错 ConnectionRefusedError"
```

### 4. Git 工作流

```bash
# 智能提交（自动生成 commit message）
mimo git commit

# 创建 PR
mimo git pr create

# 代码审查
mimo git pr review 142
```

## 🏗️ 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                     MiMo CLI Architecture                    │
├─────────────────────────────────────────────────────────────┤
│  Interface Layer (Terminal/IDE/Web)                         │
│          ↓                                                  │
│  Core Agent Engine                                          │
│    ├── Intent Parser (意图解析)                              │
│    ├── Planner Engine (规划引擎)                            │
│    ├── Executor Engine (执行引擎)                           │
│    └── Observer Engine (观察引擎)                           │
│          ↓                                                  │
│  Infrastructure Layer                                       │
│    ├── Tool System (工具系统)                               │
│    ├── Context Manager (上下文管理)                         │
│    ├── Safety Guardrail (安全护栏)                          │
│    └── Memory System (记忆系统)                             │
└─────────────────────────────────────────────────────────────┘
```

## 🛠️ 内置工具集

| 工具 | 功能 | 安全等级 |
|------|------|----------|
| `shell` | 执行 Shell 命令 | HIGH |
| `file_read` | 读取文件内容 | LOW |
| `file_write` | 写入/创建文件 | MEDIUM |
| `file_edit` | 精确编辑文件 | MEDIUM |
| `file_delete` | 删除文件 | HIGH |
| `dir_list` | 列出目录内容 | LOW |
| `search` | 项目内文本搜索 | LOW |
| `git` | Git 操作全套 | MEDIUM |
| `web_fetch` | 抓取网页内容 | LOW |
| `web_search` | 网络搜索 | LOW |
| `http_request` | 发送 HTTP 请求 | MEDIUM |
| `json_query` | JSON/YAML 解析查询 | LOW |
| `diff` | 文件差异对比 | LOW |
| `docker` | Docker 容器操作 | HIGH |
| `clipboard` | 剪贴板读写 | LOW |
| `process` | 进程管理 | HIGH |
| `env` | 环境变量管理 | MEDIUM |
| `dependency` | 包管理器操作 | MEDIUM |

## ⚙️ 配置说明

MiMo CLI 支持多层级配置：

1. **命令行参数** - `mimo --model deepseek "..."`
2. **项目级配置** - `./mimo.yaml`
3. **用户级配置** - `~/.mimo/config.yaml`
4. **系统级配置** - `/etc/mimo/config.yaml`
5. **内置默认值**

### 配置示例

```yaml
# ~/.mimo/config.yaml
default_model: "mimo"
language: "zh-CN"
theme: "dark"

models:
  mimo:
    api_base: "https://api.mimo.xiaomi.com/v1"
    api_key: "${MIMO_API_KEY}"
    model: "mimo-v2"

safety:
  level: "confirm"
  backup_before_write: true
  audit_log: "~/.mimo/audit.log"

agent:
  max_iterations: 25
  planning_mode: "plan-execute"
```

## 🧩 技能系统

MiMo CLI 支持通过技能扩展功能：

```bash
# 查看可用技能
mimo skill list

# 安装技能
mimo skill install @community/k8s-deploy

# 搜索技能
mimo skill search "docker deploy"

# 发布自己的技能
mimo skill publish ./my-skill
```

## 🔒 安全特性

- **三级安全策略**：锁定模式、确认模式、自动模式
- **危险操作拦截**：自动识别并拦截危险命令
- **文件备份**：写入前自动备份到 `.mimo/backups/`
- **操作审计**：所有操作记录到审计日志
- **沙箱执行**：支持 Docker 容器隔离执行

## 📊 性能指标

| 指标 | 目标值 |
|------|--------|
| 启动时间 | < 200ms |
| 首 Token 延迟 | < 500ms |
| 内存占用 | 空闲 < 30MB |
| 二进制大小 | < 50MB |

## 🤝 参与贡献

我们欢迎社区贡献！请查看 [CONTRIBUTING.md](docs/CONTRIBUTING.md) 了解详细信息。

### 开发环境

```bash
# 克隆仓库
git clone https://github.com/mimo-cli/mimo-cli.git
cd mimo-cli

# 安装依赖
go mod tidy

# 构建
go build -o mimo .

# 运行测试
go test ./...
```

## 📄 许可证

本项目采用 [Apache 2.0 许可证](LICENSE)。

## 📞 联系我们

- 项目主页：https://github.com/mimo-cli/mimo-cli
- 问题反馈：https://github.com/mimo-cli/mimo-cli/issues
- 技能市场：https://skills.mimo.xiaomi.com

---

**MiMo CLI** - 让开发更高效，让编程更智能 🚀