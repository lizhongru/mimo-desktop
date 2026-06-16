# Skill Distill Files Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 打通 Distill 最小闭环，把候选 skill 落盘为 `.mimo/skills/<name>/SKILL.md`，并让桌面端能读取候选列表。

**Architecture:** 复用 `internal/skill` 作为候选分析、保存、解析的唯一来源；`desktop/app_dream.go` 只负责调用并映射 DTO。生成的 `SKILL.md` 使用稳定 Markdown 模板，`candidates.md` 保持现有可读格式并提供解析器。

**Tech Stack:** Go 标准库、现有 `internal/skill` 包、Wails 桌面绑定。

---

### Task 1: 候选解析器

**Files:**
- Modify: `internal/skill/distill.go`
- Test: `internal/skill/distill_test.go`

- [ ] **Step 1: Write failing test**

`TestParseCandidatesMarkdown` 验证 `candidates.md` 可解析出名称、描述、置信度、模式和命令。

- [ ] **Step 2: Implement parser**

`ParseCandidatesMarkdown(data []byte) []SkillCandidate` 读取现有 Markdown section。

### Task 2: 生成 skill 文件

**Files:**
- Modify: `internal/skill/distill.go`
- Test: `internal/skill/distill_test.go`

- [ ] **Step 1: Write failing test**

`TestDistillSaveCandidatesWritesSkillFiles` 验证 `.mimo/skills/<name>/SKILL.md` 被生成。

- [ ] **Step 2: Implement writer**

`SaveCandidates` 同时写 `candidates.md` 和每个候选目录下的 `SKILL.md`。

### Task 3: Desktop candidate list

**Files:**
- Modify: `desktop/app_dream.go`

- [ ] **Step 1: Replace placeholder parser**

`DistillListCandidates` 调用 `skill.ParseCandidatesMarkdown` 并映射 DTO。

### Task 4: Verify and handoff

**Files:**
- Modify: `HANDOFF_CURRENT.md`

- [ ] **Step 1: Run verification**

`go test ./desktop/... ./internal/session/... ./internal/skill -count=1` and `npm run build`.

- [ ] **Step 2: Update handoff and commit**

记录 Skill Distill 最小闭环和验证状态后提交。
