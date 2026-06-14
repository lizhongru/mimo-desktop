# MiMo Desktop — 当前上下文交接

> 日期：2026-06-14
> 工作区：`D:\works\study\mimo cli`
> 当前分支：`master`
> 远端：`origin/master`
> 当前状态：本轮侧栏 UI 与工作区展开修复已完成，准备提交并推送到 `origin/master`。

---

## 1. 最高优先级对话规则

- 用户要求：所有回复必须用中文。
- 用户要求：每句话都必须叫用户“哥哥”。
- 不要使用 `git reset --hard`、`git checkout --` 等会丢改动的命令。
- 若遇到未提交改动，默认视为用户或当前 agent 的工作成果，不要回退。

---

## 2. 当前 Git 状态

当前本地分支：

```text
master
```

提交历史顶部在本轮提交前为：

```text
aef153e feat: enforce persisted permission rules
7b75b8f feat: persist advanced settings
0e55a2e docs: define advanced settings persistence
0bdaebc feat: auto checkpoint saved chat sessions
587f7e5 feat: apply active agent mode to chat runtime
```

说明：

- 本地只保留 `master` 分支。
- `aef153e feat: enforce persisted permission rules` 已在 `master` 上。
- 本轮会新增一个侧栏 UI / 工作区展开相关提交并推送。
- 若 GitHub HTTPS 推送再次遇到 connection reset，可用：

```powershell
git -c http.sslBackend=schannel -c http.version=HTTP/1.1 push origin master
```

---

## 3. 本轮完成内容

### 3.1 权限系统接入

提交：

```text
aef153e feat: enforce persisted permission rules
```

主要内容：

- 将已持久化的 permission rules 接入工具权限判断。
- 覆盖 read/write/edit/bash 等权限路径。
- 合并后本地只剩 `master` 分支。

### 3.2 左侧对话列表 UI 优化

主要文件：

- `desktop/frontend/src/components/layout/LeftSidebar.tsx`
- `desktop/frontend/src/components/layout/AppLayout.tsx`
- `desktop/frontend/src/styles/globals.css`

主要内容：

- 左侧栏宽度调整为 `284px`。
- 顶部改为轻量标题、icon-only 新建按钮、管理按钮。
- 新增搜索输入框和搜索无结果空态。
- 对话行改为更现代的紧凑单行样式。
- 空态改为更轻的居中样式。
- 底部用户区视觉统一为侧栏账号行。
- 新增侧栏专用主题变量，适配深色/浅色切换。

### 3.3 新建对话与工作区展开修复

主要文件：

- `desktop/frontend/src/App.tsx`
- `desktop/frontend/src/components/layout/LeftSidebar.tsx`
- `desktop/frontend/src/stores/sessionStore.ts`

主要内容：

- 新建空对话后立即加入前端 session 列表。
- 左侧栏会根据 `selectedWorkspace` 和 `currentSessionId` 自动展开对应工作区分组。
- 选择文件夹后，如果当前是空对话，会同步更新本地 session 的 `workspaceId`。
- 发送首条消息前移动 session 时也同步本地工作区，避免 UI 等后端刷新。

---

## 4. 已验证

本轮按用户要求未跑测试。

已做的轻量检查：

- 使用本地页面截图查看深色侧栏效果。
- 使用本地 Chrome CDP 临时切到浅色类名查看浅色侧栏效果。
- 未执行 `go test`、`npm run build`、`npx tsc --noEmit`。

---

## 5. 当前潜在问题

1. Wails 自动生成文件仍有改动：`desktop/frontend/src/wails/wailsjs/go/desktop/App.d.ts`、`App.js`、`models.ts`。
2. 本轮没有按测试流程验证类型检查或构建，因为用户明确要求不需要测试。
3. `npm run build` 既有大 chunk warning 和 Node ESM warning 仍未处理。

---

## 6. 建议下一步

建议优先级：

1. 文件树真实数据：把前端 `FileTree` 从 mock root 改为后端目录树接口。
2. 子智能体真实执行：把 `internal/actor` 的模拟执行替换为真实 LLM 子 Agent 生命周期。
3. Memory 配置生效：让 `ccIndex` 和 `searchScoreFloor` 真正影响 memory reconcile/search 行为。
4. 任务 ID 语义：改成 `T1`、`T1.1` 这种树状编号，并补齐 task rename/archive/progress。

---

## 7. 下个 session 建议接手步骤

先确认状态：

```powershell
git status --short --branch --untracked-files=all
git log --oneline -6
git branch -r --format='%(refname:short)'
```

如果继续前端体验优化，建议先看：

- `desktop/frontend/src/components/layout/LeftSidebar.tsx`
- `desktop/frontend/src/App.tsx`
- `desktop/frontend/src/stores/sessionStore.ts`
- `desktop/frontend/src/styles/globals.css`

如果继续后端功能开发，建议先看：

- `internal/session/store.go`
- `desktop/app_session.go`
- `internal/tools/*`
- `internal/agent/agent.go`

---

## 8. 关键参考文件

后端：

- `desktop/app_session.go`
- `internal/session/store.go`
- `internal/permission/ruleset.go`
- `internal/safety/guardrail.go`
- `internal/agent/agent.go`
- `internal/tools/*`

前端：

- `desktop/frontend/src/App.tsx`
- `desktop/frontend/src/components/layout/LeftSidebar.tsx`
- `desktop/frontend/src/components/layout/AppLayout.tsx`
- `desktop/frontend/src/stores/sessionStore.ts`
- `desktop/frontend/src/styles/globals.css`
- `desktop/frontend/src/wails/wailsjs/go/desktop/App.d.ts`
- `desktop/frontend/src/wails/wailsjs/go/desktop/App.js`
- `desktop/frontend/src/wails/wailsjs/go/models.ts`

规划文档：

- `docs/MiMo-Code-Integration-Plan.md`
- `docs/superpowers/specs/2026-06-12-advanced-settings-persistence-design.md`
- `docs/superpowers/plans/2026-06-12-advanced-settings-persistence.md`
