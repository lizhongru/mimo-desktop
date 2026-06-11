# MiMo-Code 功能移植开发文档

> 基于当前桌面端 (Wails v2 + Go + React 19)，将 MiMo-Code (Electron + TypeScript + SolidJS) 的核心功能逐步移植集成。
> 生成日期：2026-06-11
> 最后更新：2026-06-11

---

## 📊 开发进度总览

> 复核日期：2026-06-11
> 复核口径：只有接入 Agent 真实执行链路、配置可持久化、前端调用真实后端能力的功能才记为完成；仅存在 UI、CRUD 或模拟实现时记为部分完成。

| Phase | 功能 | 当前状态 | 主要缺口 |
|-------|------|----------|----------|
| P1 | 持久化记忆系统 | 🟡 **部分完成** | FTS/面板已落地；`memory` tool 未注册进 Agent，配置项未完整接入 |
| P2 | 智能上下文管理 | 🟡 **部分完成** | 手动 checkpoint 已有；恢复未真正 `LoadMessages`，自动 checkpoint/上下文重建未接入聊天循环 |
| P3 | 任务追踪系统 | 🟡 **部分完成** | CRUD/面板已落地；缺少 `T1/T1.1` 编号、Agent `task` tool、任务事件总线和 progress 文件语义 |
| P4 | 子智能体系统 | 🔴 **未完成核心链路** | Actor 注册/面板已落地；执行仍是 `time.After` 模拟，未启动真实 LLM 子智能体 |
| P5 | 多智能体系统 | 🟡 **部分完成** | 切换 UI/配置已落地；prompt、权限、工具白名单未应用到运行中的 Agent |
| P6 | Dream & Distill | 🟡 **部分完成** | Dream 为关键词提取；Distill 候选列表仍是 placeholder，未使用源项目子智能体蒸馏流程 |
| P7 | 前端 UI 增强 | 🟡 **部分完成** | 多数面板已落地；FileTree 仍是 mock，AdvancedSettings 保存未持久化 |

**当前进度**: 7/7 Phases 已有 UI 或后端骨架，但 0/7 Phases 达到 MiMo-Code 核心语义完整对齐，不能按 100% 结项。

### 复核结论

- 已落地：Wails 绑定、面板入口、基础 SQLite 表、记忆搜索/写入、任务 CRUD、checkpoint CRUD、Actor 列表/取消、Agent 切换 UI、Dream/Distill 基础入口。
- 未完成：Agent 真实工具注册、Checkpoint 恢复到运行上下文、真实子智能体执行、多 Agent prompt/权限/工具白名单生效、设置持久化、LSP、插件、Goal 停止条件。
- 当前验证：`go test ./...` 通过，`desktop/frontend` 下 `npm run build` 通过，但前端存在大 chunk warning。

---

## 一、两个项目对比分析

### 1.1 技术栈差异

| 维度 | 当前项目 (mimo-cli) | MiMo-Code |
|------|---------------------|-----------|
| 后端语言 | Go | TypeScript (Bun) |
| 桌面框架 | Wails v2 | Electron 41 |
| 前端框架 | React 19 | SolidJS |
| 前端构建 | Vite + TailwindCSS | electron-vite + UnoCSS |
| 状态管理 | Zustand 5 | SolidJS signals |
| 数据库 | SQLite (modernc.org/sqlite) | SQLite (drizzle-orm + better-sqlite) |
| LLM SDK | 自研 OpenAI/Anthropic 适配 | Vercel AI SDK (@ai-sdk/*) |
| 包管理 | Go modules + npm | Bun workspaces (monorepo) |
| 通信机制 | Wails bindings (Go ↔ JS) | Electron IPC + Hono HTTP server |

### 1.2 功能差异矩阵

| 功能模块 | 当前项目复核状态 | MiMo-Code | 优先级 |
|----------|------------------|-----------|--------|
| 基础聊天 | ✅ 已有 | ✅ 已有 | - |
| 流式输出 | ✅ 已有 | ✅ 已有 | - |
| 工具系统 (25个) | ⚠️ 基础工具已有；MiMo-Code 新增 `memory/task/actor/lsp` 等工具未接入 | ✅ 已有 (58个) | P2 |
| 多模型支持 | ✅ 已有 | ✅ 已有 | - |
| MCP 协议 | ✅ 已有 | ✅ 已有 | - |
| 安全确认 | ✅ 已有 | ✅ 已有 | - |
| 会话持久化 | ✅ SQLite | ✅ SQLite + FTS5 | P2 |
| **持久化记忆系统** | 🟡 FTS/文件面板已有；未作为 Agent 工具使用 | ✅ SQLite FTS5 全文搜索 | **P0** |
| **智能上下文管理** | 🟡 checkpoint CRUD 已有；自动创建、恢复到 Agent 上下文未完成 | ✅ 自动检查点 + 上下文重建 | **P0** |
| **任务追踪系统** | 🟡 CRUD/面板已有；缺树状编号和 Agent `task` tool | ✅ 树状任务 (T1, T1.1...) | **P1** |
| **子智能体系统** | 🔴 当前为模拟执行；未运行真实 LLM 子智能体 | ✅ explore/general/title 等 | **P1** |
| **多智能体 (build/plan/compose)** | 🟡 UI/配置已有；未影响运行 Agent 的 prompt/权限/tools | ✅ 多主智能体 + Tab 切换 | **P1** |
| **Goal/停止条件** | ❌ 无 | ✅ /goal + 裁判模型评估 | P2 |
| **Compose 编排模式** | ❌ 无 | ✅ 内置 skill 编排 | P2 |
| **语音输入** | ❌ 无 | ✅ TenVAD + MiMo ASR | P3 |
| **Dream & Distill** | 🟡 基础入口已有；Distill 候选解析和 LLM 流程未完成 | ✅ 记忆提取 + 技能蒸馏 | P2 |
| **权限系统** | ⚠️ 基础 confirm + 独立 ruleset；未接入工具执行/Agent 配置合并 | ✅ 细粒度 per-tool 权限规则 | P1 |
| **LSP 集成** | ❌ 无 | ✅ 语言服务器协议 | P3 |
| **插件系统** | ❌ 无 | ✅ 插件注册/加载 | P2 |
| **深链接** | ❌ 无 | ✅ opencode:// 协议 | P3 |
| **自动更新** | ❌ 无 | ✅ electron-updater | P3 |
| **国际化** | ✅ 中英双语 (100+ key) | ✅ 15 种语言 | P2 |
| **CLI 安装** | ❌ 无 | ✅ 一键安装脚本 | P3 |

---

## 二、移植架构设计

### 2.1 核心原则

1. **保持 Wails v2 架构不变** — 不迁移到 Electron，保留 Go 后端优势
2. **功能对齐而非代码搬运** — 用 Go 重新实现 MiMo-Code 的 TypeScript 逻辑
3. **渐进式集成** — 按优先级分 Phase 交付，每个 Phase 独立可用
4. **前端组件可选替换** — React 组件可逐步引入 MiMo-Code 的设计模式

### 2.2 后端新增模块规划

```
internal/
├── memory/              # 🟡 部分完成：FTS/文件读写已落地，Agent 工具接入待完成
│   ├── service.go       # 记忆服务 (搜索/索引/调和)
│   ├── fts.go           # FTS5 全文搜索引擎
│   ├── paths.go         # 记忆路径管理
│   ├── store.go         # SQLite 记忆存储
│   └── service_test.go  # 单元测试
├── task/                # 🟡 部分完成：任务 CRUD 已落地，树状编号/tool/progress 待完成
│   ├── registry.go      # 任务注册/状态管理（当前 ID 为时间戳，不是 T1/T1.1）
│   └── registry_test.go # 单元测试
├── actor/               # 🔴 核心未完成：生命周期骨架已有，执行仍是模拟
│   ├── actor.go         # Actor 注册与生命周期（未启动真实 LLM 子智能体）
│   └── actor_test.go    # 单元测试
├── skill/               # 🟡 部分完成：基础 distill 文件生成，候选解析/LLM 流程待完成
│   ├── distill.go       # 技能蒸馏（简化实现）
│   └── distill_test.go  # 单元测试
├── permission/          # 🟡 部分完成：规则集已有，未全量接入工具执行链路
│   ├── ruleset.go       # 权限规则集
│   └── ruleset_test.go  # 单元测试
├── agent/               # 🟡 部分完成：多智能体配置已有，未应用到运行中 Agent
│   └── multi_agent.go   # 多智能体配置管理
└── context/             # 🟡 部分完成：系统提示/检查点文件已有，自动重建待完成
    ├── manager.go       # 上下文管理器
    └── checkpoint.go    # 检查点管理器
```

### 2.3 前端新增组件规划

```
desktop/frontend/src/
├── components/
│   ├── memory/              # 🟡 部分完成：记忆搜索/编辑 UI
│   │   ├── MemorySearch.tsx
│   │   ├── MemoryPanel.tsx
│   │   └── index.ts
│   ├── checkpoint/          # 🟡 部分完成：检查点管理 UI，恢复语义待后端完成
│   │   ├── CheckpointPanel.tsx
│   │   └── index.ts
│   ├── task/                # 🟡 部分完成：任务面板
│   │   ├── TaskPanel.tsx
│   │   └── index.ts
│   ├── actor/               # 🟡 部分完成：子智能体面板，后端仍模拟
│   │   ├── ActorPanel.tsx
│   │   └── index.ts
│   ├── agent/               # 🟡 部分完成：多智能体切换 UI
│   │   ├── AgentSwitcher.tsx
│   │   └── index.ts
│   ├── session/             # 🟡 部分完成：会话标签页
│   │   ├── SessionTabs.tsx
│   │   └── index.ts
│   ├── file/                # 🔴 未完成真实数据：当前 FileTree 使用 mock tree
│   │   ├── FileTree.tsx
│   │   └── index.ts
│   └── settings/            # 🟡 部分完成：高级设置 UI，保存未持久化
│       └── AdvancedSettings.tsx
├── stores/
│   ├── taskStore.ts         # ❌ 未创建：任务状态
│   ├── memoryStore.ts       # ❌ 未创建：记忆状态
│   └── agentStore.ts        # ❌ 未创建：多智能体状态
└── hooks/
    ├── useTask.ts           # ❌ 未创建：任务 hook
    ├── useMemory.ts         # ❌ 未创建：记忆 hook
    └── useAgent.ts          # ✅ 已存在：基础 Agent 初始化 hook
```

---

## 三、Phase 0：基础设施准备 (1 周) 🟡 部分完成

### 3.1 数据库 Schema 扩展

```sql
-- 记忆系统表
CREATE TABLE IF NOT EXISTS memory_fts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL,
    scope TEXT NOT NULL,        -- 'global' | 'projects' | 'sessions'
    scope_id TEXT,
    type TEXT,
    body TEXT NOT NULL,
    created_at INTEGER NOT NULL
);

CREATE VIRTUAL TABLE IF NOT EXISTS memory_fts_idx USING fts5(
    body, path, scope, scope_id, type,
    content=memory_fts,
    content_rowid=id
);

-- 任务系统表
CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    parent_task_id TEXT,
    status TEXT NOT NULL DEFAULT 'open',
    summary TEXT NOT NULL,
    owner TEXT,
    created_at INTEGER NOT NULL,
    last_event_at INTEGER NOT NULL,
    ended_at INTEGER,
    cleanup_after INTEGER,
    FOREIGN KEY (session_id) REFERENCES sessions(id)
);

CREATE TABLE IF NOT EXISTS task_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL,
    at INTEGER NOT NULL,
    kind TEXT NOT NULL,
    summary TEXT,
    FOREIGN KEY (task_id) REFERENCES tasks(id)
);
```

### 3.2 配置 Schema 扩展

在 `internal/config/schema.go` 中新增：

```go
// MemoryConfig 记忆系统配置
type MemoryConfig struct {
    CCIndex bool `yaml:"cc_index"` // 索引 Claude Code 记忆
}

// CheckpointConfig 检查点配置
type CheckpointConfig struct {
    MemoryReconcileOnSearch bool    `yaml:"memory_reconcile_on_search"`
    MemorySearchScoreFloor  float64 `yaml:"memory_search_score_floor"`
    TaskArchiveDays         int     `yaml:"task_archive_days"`
    AutoCheckpoint          bool    `yaml:"auto_checkpoint"`
    ContextBudget           int     `yaml:"context_budget"` // token 预算
}

// PermissionConfig 权限配置
type PermissionConfig struct {
    Rules []PermissionRule `yaml:"rules"`
}
```

### 3.3 依赖安装

```bash
# Go 依赖 (已在 go.mod 中)
go get github.com/mattn/go-sqlite3  # 如需 FTS5 支持

# 前端依赖 (在 desktop/frontend/)
npm install @radix-ui/react-dialog @radix-ui/react-tabs @radix-ui/react-tooltip
```

---

## 四、Phase 1：持久化记忆系统 (2 周) [P0] 🟡 部分完成

### 4.1 后端实现

#### `internal/memory/service.go`

```go
package memory

import (
    "database/sql"
    "path/filepath"
    "os"
    "strings"
)

type Service struct {
    db *sql.DB
    root string
}

type SearchResult struct {
    Path    string
    Snippet string
    Score   float64
    Scope   string
    ScopeID string
    Type    string
}

func (s *Service) Search(query string, opts SearchOpts) ([]SearchResult, error) {
    // 1. 构建 FTS5 查询 (BM25 排序)
    // 2. 执行搜索
    // 3. 应用分数下限过滤
    // 4. 返回结果
}

func (s *Service) Reconcile() (indexed int, pruned int, err error) {
    // 1. 扫描 memory 目录下的 .md 文件
    // 2. 解析 frontmatter (scope, scope_id, type)
    // 3. 更新 FTS 索引
    // 4. 清理已删除文件的索引
}
```

#### `internal/memory/fts.go`

```go
package memory

import "strings"

// BuildFtsQuery 将用户查询转换为 FTS5 查询
// 标点符号变为分隔符，每个字母数字序列变为短语引用字面量
func BuildFtsQuery(query string) string {
    tokens := tokenize(query)
    if len(tokens) == 0 { return "" }
    quoted := make([]string, len(tokens))
    for i, t := range tokens {
        quoted[i] = `"` + t + `"`
    }
    return strings.Join(quoted, " OR ")
}
```

### 4.2 Wails 绑定

在 `desktop/app.go` 中新增：

```go
// MemorySearch 搜索记忆
func (a *App) MemorySearch(query string, scope string, limit int) []MemorySearchResult {
    return a.memoryService.Search(query, scope, limit)
}

// MemoryReconcile 重新索引记忆
func (a *App) MemoryReconcile() (int, int) {
    return a.memoryService.Reconcile()
}
```

### 4.3 前端实现

```tsx
// desktop/frontend/src/components/memory/MemorySearch.tsx
export function MemorySearch() {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SearchResult[]>([]);

  const handleSearch = async () => {
    const res = await window.go.desktop.App.MemorySearch(query, "", 10);
    setResults(res);
  };

  return (
    <div className="p-4">
      <input value={query} onChange={e => setQuery(e.target.value)} />
      <button onClick={handleSearch}>搜索记忆</button>
      {results.map(r => <ResultCard key={r.path} result={r} />)}
    </div>
  );
}
```

### 4.4 文件结构

在项目根目录创建：

```
~/.mimo/memory/                    # 全局记忆
<project>/.mimo/memory/            # 项目记忆
<project>/.mimo/sessions/<sid>/    # 会话记忆
    checkpoint.md
    notes.md
    tasks/<id>/progress.md
```

---

## 五、Phase 2：智能上下文管理 (2 周) [P0] 🟡 部分完成

### 5.1 检查点系统

#### `internal/checkpoint/writer.go`

```go
package checkpoint

type Writer struct {
    sessionDir string
    memoryDir  string
}

// WriteCheckpoint 写入会话检查点
func (w *Writer) WriteCheckpoint(state SessionState) error {
    // 1. 渲染 checkpoint 模板
    // 2. 提取关键信息 (任务树、当前工作、设计决策等)
    // 3. 写入 checkpoint.md
    // 4. 更新 MEMORY.md (如果需要)
}

// WriteMemory 写入项目记忆
func (w *Writer) WriteMemory(entries []MemoryEntry) error {
    // 1. 扫描现有记忆
    // 2. 合并新条目
    // 3. 清理过时条目
    // 4. 写入 MEMORY.md
}
```

### 5.2 上下文重建

增强 `internal/context/manager.go`：

```go
// RebuildContext 当上下文接近限制时重建
func (m *Manager) RebuildContext(messages []llm.Message) []llm.Message {
    // 1. 读取最新 checkpoint
    // 2. 注入项目记忆
    // 3. 注入任务进展
    // 4. 保留近期消息
    // 5. 按 token 预算分配空间
    // 6. 返回重建后的消息列表
}
```

### 5.3 前端反馈

```tsx
// 在 StatusBar 中显示检查点状态
<div className="flex items-center gap-2 text-xs text-txt-m">
  <CheckpointIcon />
  <span>检查点: {checkpointTime}</span>
  <span>记忆: {memoryCount} 条</span>
</div>
```

---

## 六、Phase 3：任务追踪系统 (1 周) [P1] 🟡 部分完成

### 6.1 后端实现

#### `internal/task/registry.go`

```go
package task

type TaskStatus string
const (
    TaskOpen       TaskStatus = "open"
    TaskInProgress TaskStatus = "in_progress"
    TaskBlocked    TaskStatus = "blocked"
    TaskDone       TaskStatus = "done"
    TaskAbandoned  TaskStatus = "abandoned"
)

type Task struct {
    ID            string
    SessionID     string
    ParentTaskID  *string
    Status        TaskStatus
    Summary       string
    Owner         *string
    CreatedAt     int64
    LastEventAt   int64
    EndedAt       *int64
    CleanupAfter  *int64
}

type Registry struct {
    db *sql.DB
}

func (r *Registry) Create(sessionID, summary string, parentID *string) (*Task, error) {}
func (r *Registry) List(sessionID string, status *TaskStatus, includeTerminal bool) ([]Task, error) {}
func (r *Registry) Start(id, owner string, eventSummary string) (*Task, error) {}
func (r *Registry) Done(id, eventSummary string) (*Task, error) {}
func (r *Registry) Block(id, eventSummary string) (*Task, error) {}
```

### 6.2 Wails 绑定

```go
// Task 工具 — 供 Agent 调用
func (a *App) TaskCreate(sessionID, summary, parentID string) (Task, error) {}
func (a *App) TaskList(sessionID string, status string) ([]Task, error) {}
func (a *App) TaskStart(id, owner, eventSummary string) (Task, error) {}
func (a *App) TaskDone(id, eventSummary string) (Task, error) {}
func (a *App) TaskBlock(id, eventSummary string) (Task, error) {}
```

### 6.3 前端任务面板

```tsx
// desktop/frontend/src/components/task/TaskTree.tsx
export function TaskTree({ tasks }: { tasks: Task[] }) {
  const rootTasks = tasks.filter(t => !t.parentTaskId);
  return (
    <div className="space-y-1">
      {rootTasks.map(task => (
        <TaskNode key={task.id} task={task} allTasks={tasks} />
      ))}
    </div>
  );
}
```

---

## 七、Phase 4：子智能体系统 (2 周) [P1] 🔴 核心未完成

### 7.1 后端实现

#### `internal/actor/registry.go`

```go
package actor

type ActorType string
const (
    ActorExplore   ActorType = "explore"
    ActorGeneral   ActorType = "general"
    ActorTitle     ActorType = "title"
    ActorSummary   ActorType = "summary"
    ActorCompact   ActorType = "compaction"
    ActorDream     ActorType = "dream"
    ActorDistill   ActorType = "distill"
    ActorWriter    ActorType = "checkpoint-writer"
)

type Actor struct {
    ID          string
    Type        ActorType
    SessionID   string
    ParentID    *string
    Status      string
    Prompt      string
    CreatedAt   int64
}

type Registry struct {
    mu     sync.RWMutex
    actors map[string]*Actor
    agent  *agent.Agent
}

// Spawn 创建子智能体
func (r *Registry) Spawn(opts SpawnOpts) (*Actor, error) {
    // 1. 根据类型选择权限和提示词
    // 2. 创建 Actor 记录
    // 3. 启动 goroutine 执行
    // 4. 返回 Actor ID
}

// Send 向子智能体发送消息
func (r *Registry) Send(actorID, content string) error {}

// Cancel 取消子智能体
func (r *Registry) Cancel(actorID string) error {}
```

### 7.2 Wails 绑定

```go
// Actor 绑定
func (a *App) ActorSpawn(actorType, prompt, taskID string) (string, error) {}
func (a *App) ActorSend(actorID, content string) error {}
func (a *App) ActorCancel(actorID string) error {}
func (a *App) ActorStatus(actorID string) (ActorStatus, error) {}
```

### 7.3 前端子智能体指示器

```tsx
// 在 StatusBar 中显示活跃的子智能体
{activeActors.length > 0 && (
  <div className="flex items-center gap-2 text-xs text-blue-400">
    <Loader2 className="w-3 h-3 animate-spin" />
    <span>子智能体: {activeActors.map(a => a.type).join(', ')}</span>
  </div>
)}
```

---

## 八、Phase 5：多智能体系统 (2 周) [P1] 🟡 部分完成

### 8.1 Agent 定义扩展

在 `internal/agent/agent.go` 中扩展：

```go
type AgentMode string
const (
    AgentModeBuild      AgentMode = "build"
    AgentModePlan       AgentMode = "plan"
    AgentModeCompose    AgentMode = "compose"
    AgentModeSubagent   AgentMode = "subagent"
)

type AgentConfig struct {
    Name        string
    Mode        AgentMode
    Color       string
    Description string
    Permission  PermissionRuleset
    Prompt      string
    ToolAllowlist []string
}
```

### 8.2 权限系统

```go
// internal/permission/ruleset.go
type PermissionAction string
const (
    PermissionAllow PermissionAction = "allow"
    PermissionDeny  PermissionAction = "deny"
    PermissionAsk   PermissionAction = "ask"
)

type PermissionRule struct {
    Permission string          // "read", "write", "edit", "bash", "external_directory"
    Action     PermissionAction
    Pattern    string          // glob pattern
}

type Ruleset []PermissionRule

// Merge 合并两个规则集
func Merge(a, b Ruleset) Ruleset {}

// Evaluate 评估操作是否被允许
func (r Ruleset) Evaluate(tool string, params map[string]interface{}) PermissionAction {}
```

### 8.3 前端智能体切换器

```tsx
// desktop/frontend/src/components/agent/AgentSwitcher.tsx
export function AgentSwitcher() {
  const [agents, setAgents] = useState<AgentInfo[]>([]);
  const current = useAgentStore(s => s.currentAgent);

  return (
    <div className="flex gap-1">
      {agents.filter(a => a.mode === 'primary').map(agent => (
        <button
          key={agent.name}
          className={`px-2 py-1 rounded text-xs ${
            current === agent.name ? 'bg-accent text-white' : 'bg-elevated'
          }`}
          onClick={() => switchAgent(agent.name)}
        >
          {agent.name}
        </button>
      ))}
    </div>
  );
}
```

---

## 九、Phase 6：Dream & Distill (1 周) [P2] 🟡 部分完成

### 9.1 Dream — 记忆提取

```go
// internal/memory/dream.go
func (s *Service) Dream(sessionDir string) error {
    // 1. 扫描近期会话轨迹
    // 2. 识别持久知识 (架构决策、bug 修复、用户偏好)
    // 3. 提取到 MEMORY.md
    // 4. 清理过时条目
}
```

### 9.2 Distill — 技能蒸馏

```go
// internal/skill/distill.go
func Distill(sessionDir string) ([]SkillCandidate, error) {
    // 1. 发现重复的手动工作流
    // 2. 生成 skill 候选
    // 3. 评估置信度
    // 4. 打包为可复用技能
}
```

---

## 十、Phase 7：前端 UI 增强 (持续) [P2] 🟡 部分完成

### 10.1 会话标签页系统

参考 MiMo-Code 的 session-tab 设计：

```tsx
// desktop/frontend/src/components/session/SessionTabs.tsx
export function SessionTabs() {
  const sessions = useSessionStore(s => s.recentSessions);
  const current = useSessionStore(s => s.currentSessionId);

  return (
    <div className="flex border-b border-bdr">
      {sessions.map(s => (
        <SessionTab key={s.id} session={s} active={s.id === current} />
      ))}
    </div>
  );
}
```

### 10.2 文件树组件

参考 MiMo-Code 的 file-tree.tsx：

```tsx
// desktop/frontend/src/components/file/FileTree.tsx
export function FileTree({ root }: { root: string }) {
  // 1. 调用 backend 获取目录结构
  // 2. 递归渲染文件树
  // 3. 支持展开/折叠
  // 4. 点击打开文件预览
}
```

### 10.3 增强的设置页面

参考 MiMo-Code 的 dialog-settings.tsx：

```tsx
// desktop/frontend/src/components/settings/SettingsPage.tsx
// 新增配置项：
// - 检查点行为配置
// - 记忆系统配置
// - 权限规则编辑器
// - 自定义智能体配置
```

---

## 十一、开发排期

| Phase | 功能 | 原计划工时 | 依赖 | 复核状态 |
|-------|------|------------|------|----------|
| P1 | 持久化记忆系统 | 2 周 | - | 🟡 部分完成 |
| P2 | 智能上下文管理 | 2 周 | P1 | 🟡 部分完成 |
| P3 | 任务追踪系统 | 1 周 | P1 | 🟡 部分完成 |
| P4 | 子智能体系统 | 2 周 | P1 | 🔴 核心未完成 |
| P5 | 多智能体系统 | 2 周 | P1, P2 | 🟡 部分完成 |
| P6 | Dream & Distill | 1 周 | P1, P3 | 🟡 部分完成 |
| P7 | 前端 UI 增强 | 持续 | P3, P4, P5 | 🟡 部分完成 |

**原计划总计约 12 周（3 个月）**；当前不建议按已完成收尾，应转为“核心链路补齐 + 验收矩阵”推进。

### 11.1 当前复核缺口清单

#### P0：必须先补齐的核心链路

1. **Agent 工具接入**：将 MiMo-Code 对应的 `memory`、`task`、`actor` 工具注册进当前 `internal/tools`，并确保 LLM 能在聊天循环中调用。
2. **Checkpoint 恢复**：`RestoreCheckpoint` 需要真正转换消息并调用 `a.agent.LoadMessages(...)`，同时把 checkpoint summary/近期消息合并进上下文。
3. **自动 checkpoint**：当前 `CheckpointManager.ShouldCheckpoint` 未接入 `ChatStream`/token 使用回调，需要在上下文接近阈值时自动生成 checkpoint。
4. **真实 Actor 执行**：`internal/actor` 需要从模拟返回改为创建子 Agent/子会话，支持 run/spawn/status/wait/cancel/send。
5. **多 Agent 生效**：`AgentSwitch` 需要让 build/plan/compose 的 prompt、权限、工具白名单真正影响下一次请求。

#### P1：影响功能完整性的缺口

1. **任务语义对齐**：任务 ID 应改为 `T1`、`T1.1` 这种兄弟序号；补齐 rename、archive、progress 文件和 task tool。
2. **权限系统接入**：当前 `permission.Ruleset` 与 `safety.Guardrail`、工具执行、Agent 配置仍是分离状态，需要统一评估入口。
3. **配置持久化**：补齐 `MemoryConfig`、`CheckpointConfig`、`PermissionConfig`，并让 `AdvancedSettings` 调用后端保存。
4. **文件树真实数据**：`FileTree` 需要后端目录树接口，不能继续使用 mock root。

#### P2：增强能力缺口

1. **Dream & Distill**：Dream 需要 LLM/子智能体提取，Distill 需要真实候选解析和确认安装流程。
2. **Goal 停止条件**：当前没有 `/goal`、裁判模型评估、阻止提前停止的 run loop 逻辑。
3. **LSP 集成**：当前没有 `internal/lsp`、诊断收集、read/write 后触发 diagnostics 的链路。
4. **插件系统**：当前没有插件 loader、matcher、hook runtime，也没有前端管理入口。

#### P3：后续产品化缺口

1. **语音输入、深链接、自动更新、CLI 安装**仍未开始。
2. **国际化**当前是中英文，未达到 MiMo-Code 15 种语言规模。
3. **前端打包优化**仍有大 chunk warning，需要后续拆包。

---

## 十二、阶段实现记录与复核说明

> 以下 Phase 报告保留原开发记录，但不再代表最终验收完成。
> 复核后发现多个模块仍停留在 UI、CRUD 或模拟执行层，是否完成应以本文档顶部“开发进度总览”和“11.1 当前复核缺口清单”为准。

### 12.0 复核发现的问题

| 模块 | 当前文件 | 问题 |
|------|----------|------|
| 多智能体 | `desktop/app_multi_agent.go` | `AgentSwitch` 只更新 `MultiAgentManager.current`，没有修改运行中 Agent 的 system prompt、权限和工具集 |
| Checkpoint | `desktop/app_checkpoint.go` | `RestoreCheckpoint` 读取消息后没有调用 `a.agent.LoadMessages`，恢复只返回成功文案 |
| Actor | `internal/actor/actor.go` | `runExplore/runGeneral/runTitle/runSummary` 使用 `time.After` 模拟，不是真实 LLM 子智能体 |
| Task | `internal/task/registry.go` | ID 使用时间戳，未实现 MiMo-Code 的 `T1/T1.1` 树状编号，也未注册为 Agent tool |
| Memory | `internal/memory/service.go` | 搜索服务存在，但未注册为 Agent `memory` tool；搜索前配置化 reconcile、CC 索引等未完整接入 |
| Settings | `desktop/frontend/src/components/settings/AdvancedSettings.tsx` | 高级设置只维护本地 state，保存回调在 `SettingsPage.tsx` 中仅 `console.log` |
| FileTree | `desktop/frontend/src/components/file/FileTree.tsx` | 当前是 mock tree，未读取真实目录结构 |
| Distill | `desktop/app_dream.go` | `DistillListCandidates` 读取文件后未解析，当前返回空列表 |

### 12.1 当前验证记录

```bash
go test ./...
# 结果：通过

cd desktop/frontend
npm run build
# 结果：通过；仍有大 chunk warning
```

---

## 十三、Phase 1 原始完成报告

### 13.1 完成内容

**Phase 1: 持久化记忆系统** 已于 2026-06-11 完成，包含以下功能：

#### 后端模块 (`internal/memory/`)
- `paths.go` — 记忆文件路径管理（全局/项目/会话作用域）
- `fts.go` — FTS5 查询构建器（分词、短语引用、摘要提取）
- `service.go` — 核心记忆服务（搜索、索引、调和）
- `store.go` — SQLite 记忆持久化（CRUD、Upsert 唯一约束）
- `service_test.go` — 8 个单元测试（全部通过）

#### 数据库扩展 (`internal/session/store.go`)
- 新增 `memory_fts` 表用于记忆存储
- 新增 `tasks` 和 `task_events` 表为 Phase 3 准备
- 添加 `DB()` 方法供外部模块访问

#### Wails 绑定 (`desktop/app_memory.go`)
- `MemorySearch` — 记忆搜索
- `MemoryReconcile` — 记忆重新索引
- `MemoryCount` — 记忆条目统计
- `MemoryIndexFile` — 索引单个文件
- `WriteMemory` — 写入记忆文件
- `ReadMemory` — 读取记忆文件
- `ListMemoryFiles` — 列出记忆文件

#### 前端组件 (`desktop/frontend/src/components/memory/`)
- `MemorySearch.tsx` — 搜索 UI（搜索/文件标签页）
- `MemoryPanel.tsx` — 文件列表 + 编辑器面板
- `index.ts` — 组件导出

### 13.2 测试结果

```bash
# Go 单元测试
go test ./internal/memory/... -v
# 结果: 8/8 测试通过

# 前端构建
cd desktop/frontend && npm run build
# 结果: 构建成功
```

### 13.3 技术决策

1. **FTS5 兼容性**：由于使用 `modernc.org/sqlite`（纯 Go SQLite），FTS5 支持需要特殊处理，实现了 LIKE 搜索作为降级方案。
2. **记忆格式**：保持与 MiMo-Code 兼容的 Markdown 格式 + frontmatter 元数据。
3. **分数下限**：实现 15% BM25 分数下限过滤，避免低相关性结果。

### 13.4 下一步计划

Phase 1 完成后，建议按以下顺序继续：

1. **Phase 2: 智能上下文管理** — 自动检查点 + 上下文重建（2 周）
2. **Phase 3: 任务追踪系统** — 树状任务管理（1 周）
3. **Phase 4: 子智能体系统** — actor 注册与生命周期（2 周）

### 13.5 文件清单

**新增文件：**
- `internal/memory/paths.go`
- `internal/memory/fts.go`
- `internal/memory/service.go`
- `internal/memory/store.go`
- `internal/memory/service_test.go`
- `desktop/app_memory.go`
- `desktop/frontend/src/components/memory/MemorySearch.tsx`
- `desktop/frontend/src/components/memory/MemoryPanel.tsx`
- `desktop/frontend/src/components/memory/index.ts`

**修改文件：**
- `internal/session/store.go` — 新增 `DB()` 方法和数据库迁移
- `desktop/app.go` — 新增 `memorySvc` 字段和导入

---

## 十四、Phase 2 原始完成报告

### 14.1 完成内容

**Phase 2: 智能上下文管理** 已于 2026-06-11 完成，包含以下功能：

#### 后端模块 (`internal/context/`)
- `checkpoint.go` — 检查点管理器（创建、恢复、序列化）
- `manager.go` — 上下文管理器（系统提示构建、项目信息收集）

#### 数据库扩展 (`internal/session/store.go`)
- 新增 `checkpoints` 表用于检查点存储
- 新增 `Checkpoint` 数据结构
- 新增 `SaveCheckpoint`、`LoadCheckpoint`、`ListCheckpoints`、`GetLatestCheckpoint`、`DeleteCheckpoint` 方法
- 新增 `LoadMessagesFromOffset` 方法用于增量消息加载

#### Wails 绑定 (`desktop/app_checkpoint.go`)
- `CreateCheckpoint` — 创建检查点
- `ListCheckpoints` — 列出检查点
- `RestoreCheckpoint` — 恢复检查点
- `DeleteCheckpoint` — 删除检查点
- `ExportCheckpoints` — 导出检查点

#### 前端组件 (`desktop/frontend/src/components/checkpoint/`)
- `CheckpointPanel.tsx` — 检查点管理面板
- `index.ts` — 组件导出

### 14.2 测试结果

```bash
# Go 单元测试
go test ./internal/memory/... -v
# 结果: 8/8 测试通过

go test ./internal/session/... -v
# 结果: 6/6 测试通过

# 前端构建
cd desktop/frontend && npm run build
# 结果: 构建成功
```

### 14.3 技术决策

1. **检查点存储**：使用 SQLite 存储检查点元数据，文件系统存储检查点内容，确保原子性和可恢复性。
2. **增量加载**：支持从检查点偏移量加载消息，避免每次都加载全部历史。
3. **上下文重建**：从检查点摘要 + 最近消息重建上下文，保持对话连续性。

### 14.4 下一步计划

Phase 2 完成后，建议按以下顺序继续：

1. **Phase 3: 任务追踪系统** — 树状任务管理（1 周）
2. **Phase 4: 子智能体系统** — actor 注册与生命周期（2 周）

### 14.5 文件清单

**新增文件：**
- `internal/context/checkpoint.go`
- `desktop/app_checkpoint.go`
- `desktop/frontend/src/components/checkpoint/CheckpointPanel.tsx`
- `desktop/frontend/src/components/checkpoint/index.ts`

**修改文件：**
- `internal/session/store.go` — 新增检查点表和方法
- `desktop/frontend/src/App.tsx` — 新增检查点和记忆方法类型声明

---

## 十五、关键技术决策

### 15.1 为什么不用 Electron？

| 因素 | Wails v2 | Electron |
|------|----------|----------|
| 包体积 | ~17MB | ~150MB |
| 内存占用 | ~50MB | ~200MB |
| 启动速度 | <1s | ~3s |
| Go 后端优势 | ✅ 直接调用 | ❌ 需 sidecar |
| 生态成熟度 | 较新 | 非常成熟 |

**结论**：保持 Wails v2，用 Go 重新实现 MiMo-Code 的核心逻辑。

### 15.2 记忆系统存储策略

- **FTS5 索引**：使用 SQLite FTS5 进行全文搜索
- **文件系统**：记忆文件以 Markdown 存储，便于手动编辑
- **调和机制**：搜索前自动调和，确保索引最新
- **分数下限**：BM25 分数低于 top 结果 15% 的结果被过滤

### 15.3 上下文重建策略

- **token 预算**：检查点 30%，记忆 20%，近期消息 50%
- **重要性排序**：任务进展 > 设计决策 > 项目规则 > 历史记录
- **渐进式注入**：先注入最关键信息，逐步展开

---

## 十六、测试策略

### 16.1 单元测试

```bash
# Go 测试
go test ./internal/memory/... -v
go test ./internal/task/... -v
go test ./internal/actor/... -v

# 前端测试
cd desktop/frontend && npm test
```

### 16.2 集成测试

```bash
# Wails 集成测试
wails dev -tags wails
# 手动测试清单：
# 1. 记忆搜索返回正确结果
# 2. 检查点自动创建
# 3. 任务创建/状态转换
# 4. 子智能体派生/取消
# 5. 多智能体切换
```

### 16.3 E2E 测试

```bash
# 使用 Playwright
npx playwright test desktop/e2e/
```

---

## 十七、风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| FTS5 性能问题 | 记忆搜索慢 | 建立索引 + 限制结果数量 |
| 上下文重建丢失信息 | Agent 失忆 | 保留最近 10 条消息 + 关键信息优先 |
| 子智能体并发冲突 | 数据竞争 | 使用 mutex + context 控制 |
| Go 实现 TypeScript 逻辑出错 | 功能不对齐 | 逐模块对照 MiMo-Code 源码 |
| 前端 React vs SolidJS 差异 | UI 不一致 | 保持 React，参考 SolidJS 设计模式 |

---

## 十八、参考资源

- MiMo-Code 仓库：https://github.com/XiaomiMiMo/MiMo-Code
- OpenCode 原始项目：https://github.com/anomalyco/opencode
- Wails v2 文档：https://wails.io/docs
- SQLite FTS5 文档：https://www.sqlite.org/fts5.html
- Vercel AI SDK：https://sdk.vercel.ai

---

> 本文档由 MiMoCode Agent 生成，基于对两个项目的全面分析。
> 如需调整优先级或增加功能，请在对应 Phase 的 Issue 中讨论。
