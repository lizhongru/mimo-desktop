# MiMo Desktop — 当前上下文交接

> 日期：2026-06-12
> 工作区：`D:\works\study\mimo cli`
> 分支：`master`
> 远端：`origin/master`
> 当前状态：`master` 与 `origin/master` 一致，代码改动已提交并推送；本地仅 `HANDOFF_CURRENT.md` 为本次交接刷新产生的文档改动。

---

## 1. 最高优先级对话规则

- 用户要求：所有回复必须用中文。
- 用户要求：每句话都必须叫用户“哥哥”。
- 不要使用 `git reset --hard`、`git checkout --` 等会丢改动的命令。
- 若遇到未提交改动，默认视为用户或当前 agent 的工作成果，不要回退。

---

## 2. 当前提交状态

当前 `HEAD` 与 `origin/master` 均为：

```text
0bdaebc feat: auto checkpoint saved chat sessions
```

最近提交：

```text
0bdaebc feat: auto checkpoint saved chat sessions
587f7e5 feat: apply active agent mode to chat runtime
9c7c450 feat: add mimo integration panels and settings
c002108 feat: wire mimo memory task actor tools
72cacdd 更新多个文件
```

之前 GitHub 普通 DNS/SSL 有问题，如果后续需要推送，已知可用命令：

```powershell
git -c http.sslBackend=schannel -c http.curloptResolve=github.com:443:140.82.116.4 push origin master
```

---

## 3. 已完成并提交的工作

### 3.1 前端样式统一与多 Agent 切换生效

提交：

```text
587f7e5 feat: apply active agent mode to chat runtime
```

主要内容：

- 统一检查点输入框、任务描述输入框、子智能体选择/提示词输入框边框样式。
- `buildSystemPrompt` 会读取当前 Agent 配置，并把名称、模式、描述、prompt 写入 system prompt。
- 聊天发送和 Agent 切换都会同步当前 Agent 的 `ToolAllowlist`。
- `internal/agent` 支持工具白名单过滤，禁止未授权工具定义和执行。
- `AgentSwitch` 返回完整 Agent 配置，不再丢 `prompt` 和 `tool_allowlist`。

相关测试：

- `desktop/app_multi_agent_test.go`
- `internal/agent/agent_tool_allowlist_test.go`

### 3.2 自动 checkpoint 接入保存链路

提交：

```text
0bdaebc feat: auto checkpoint saved chat sessions
```

主要内容：

- `SaveSessionFromFrontend` 保存消息后会调用 `maybeCreateAutoCheckpoint`。
- 自动 checkpoint 使用当前上下文 token 预算阈值，默认遵循 `context.DefaultCheckpointConfig()`。
- 如果 `a.cfg.Context.MaxTokens` 有值，会用该值作为 checkpoint 阈值预算。
- 同一消息 offset 不重复创建 checkpoint。
- checkpoint 超过默认最大数量后会修剪旧记录。
- `RestoreCheckpoint` 会把 checkpoint summary 和 offset 之后的消息重新加载进运行中的 Agent 上下文。
- SQLite checkpoint 仍是主存储，同时保留 `.mimo/memory/sessions/<sessionID>` 文件 checkpoint 的兼容写入。

相关测试：

- `desktop/app_checkpoint_test.go`
- `internal/session/store_test.go`

---

## 4. 当前工作区状态

当前真实状态：

```text
## master...origin/master
 M HANDOFF_CURRENT.md
```

说明：

- 代码文件当前没有未提交 diff。
- 新增测试文件已经在提交中，不再是未跟踪文件。
- `HANDOFF_CURRENT.md` 是本次为了修正过期交接信息而修改。

---

## 5. 最近已验证

交接前一轮已跑过：

```powershell
go test ./... -count=1
cd desktop/frontend
npx tsc --noEmit
npm run build
cd ../..
git diff --check
```

结果：

- Go 测试通过。
- TypeScript 类型检查通过。
- Vite 生产构建通过。
- `git diff --check` 通过。
- `npm run build` 仍有既有 warning：bundle chunk 大于 500 kB、Node ESM warning；这些不是失败项。

当前本轮只刷新交接文档，若要继续改代码，建议重新跑相关验证。

---

## 6. 下个 session 建议接手步骤

1. 先确认状态：

```powershell
git status --short --branch --untracked-files=all
git log --oneline -5
git diff --stat
```

2. 若只想保存本交接刷新，可以提交：

```powershell
git add HANDOFF_CURRENT.md
git commit -m "docs: refresh current handoff state"
```

3. 若要继续功能开发，建议先选一个明确目标，再按范围补测试并实现。

---

## 7. 推荐下一步功能

建议优先级：

1. 配置持久化：补齐 `MemoryConfig`、`CheckpointConfig`、`PermissionConfig`，并让高级设置面板调用后端保存。
2. 文件树真实数据：把前端文件树从 mock root 改为后端目录树接口。
3. 子智能体真实执行：把 `internal/actor` 的模拟执行替换为真实 LLM 子 Agent 生命周期。
4. 任务 ID 语义：改成 `T1`、`T1.1` 这种树状编号，并补齐任务 rename/archive/progress 能力。
5. Dream & Distill：把关键词/placeholder 流程升级为真实候选解析和确认安装流程。

---

## 8. 关键参考文件

后端：

- `desktop/app_checkpoint.go`
- `desktop/app_session.go`
- `desktop/app_multi_agent.go`
- `desktop/app_chat.go`
- `internal/agent/agent.go`
- `internal/session/store.go`
- `internal/actor/actor.go`

前端：

- `desktop/frontend/src/App.tsx`
- `desktop/frontend/src/components/checkpoint/CheckpointPanel.tsx`
- `desktop/frontend/src/components/task/TaskPanel.tsx`
- `desktop/frontend/src/components/actor/ActorPanel.tsx`
- `desktop/frontend/src/components/layout/LeftSidebar.tsx`

规划文档：

- `docs/MiMo-Code-Integration-Plan.md`
