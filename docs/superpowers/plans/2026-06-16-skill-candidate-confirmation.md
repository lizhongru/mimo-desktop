# Skill Candidate Confirmation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 Distill 生成的 skill 候选增加桌面端确认启用与删除闭环。

**Architecture:** 后端 `internal/skill` 负责候选文件状态变更，`desktop/app_dream.go` 暴露 Wails DTO/API，前端在现有 Distill 面板上提供操作按钮并刷新列表。启用采用最小状态文件 `.mimo/skills/enabled.json`，不引入完整插件系统。

**Tech Stack:** Go 标准库、Wails 绑定、React/TypeScript、现有 frontend build。

---

### Task 1: 后端 Skill 候选状态

**Files:**
- Modify: `internal/skill/distill.go`
- Test: `internal/skill/distill_test.go`

- [ ] Step 1: 写失败测试，覆盖启用状态文件、删除候选目录、非法名称拒绝。
- [ ] Step 2: 运行 `go test ./internal/skill -count=1` 确认失败。
- [ ] Step 3: 实现 `EnableCandidate`、`DeleteCandidate`、`ListEnabledCandidates` 与安全名称解析。
- [ ] Step 4: 运行 `go test ./internal/skill -count=1` 确认通过。

### Task 2: Wails API 映射

**Files:**
- Modify: `desktop/app_dream.go`
- Modify: `desktop/frontend/src/wails/wailsjs/go/desktop/App.js`
- Modify: `desktop/frontend/src/wails/wailsjs/go/desktop/App.d.ts`
- Modify: `desktop/frontend/src/wails/wailsjs/go/models.ts`

- [ ] Step 1: 为 `SkillCandidateInfo` 增加 `enabled` 字段。
- [ ] Step 2: 暴露 `DistillEnableCandidate(name string)` 和 `DistillDeleteCandidate(name string)`。
- [ ] Step 3: 手动同步 Wails TS 绑定。
- [ ] Step 4: 运行 Go 桌面包测试。

### Task 3: 前端确认入口

**Files:**
- Modify: current Distill UI file discovered by search

- [ ] Step 1: 找到调用 `DistillListCandidates` 的组件。
- [ ] Step 2: 添加启用/删除按钮、loading 状态和结果提示。
- [ ] Step 3: 操作成功后刷新候选列表。
- [ ] Step 4: 运行 `cd desktop/frontend; npm run build`。
