# Skill Selection Message Badges Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Show which manually selected Skills were used for each user message and preserve that metadata in saved sessions.

**Architecture:** Add `selectedSkills` to frontend and backend message DTOs, persist it as JSON in the session database, and render compact badges under user messages. Keep agent injection behavior unchanged; this feature is display and history metadata only.

**Tech Stack:** Go session store and DTOs, Wails generated TypeScript bindings, React/Zustand frontend, existing i18n helper.

---

### Task 1: Persist selected Skills in messages

**Files:**
- Modify: `internal/session/store.go`
- Modify: `desktop/app_session.go`
- Modify: `desktop/frontend/src/wails/wailsjs/go/models.ts`

- [ ] Add `SelectedSkills []string` to `session.Message` and `ChatMessageDTO`.
- [ ] Add a nullable/defaulted `selected_skills` messages column during migration.
- [ ] Save and load `selected_skills` as JSON alongside `tool_lines`.
- [ ] Update Wails TS model with `selectedSkills?: string[]`.

### Task 2: Carry selected Skills through frontend state

**Files:**
- Modify: `desktop/frontend/src/lib/types.ts`
- Modify: `desktop/frontend/src/stores/chatStore.ts`
- Modify: `desktop/frontend/src/App.tsx`

- [ ] Add `selectedSkills?: string[]` to `ChatMessage`.
- [ ] Update `addUserMessage(content, selectedSkills)`.
- [ ] Map selected Skills when loading and saving sessions.

### Task 3: Render Skill badges under user messages

**Files:**
- Modify: `desktop/frontend/src/components/chat/MessageList.tsx`
- Modify: `desktop/frontend/src/lib/i18n.ts`

- [ ] Add i18n key `message_selected_skills`.
- [ ] Render compact badges under user messages when `selectedSkills.length > 0`.
- [ ] Keep assistant messages unchanged.

### Task 4: Validate

**Files:**
- Test: existing Go and frontend build commands.

- [ ] Run `go test ./desktop/... ./internal/skill -count=1`.
- [ ] Run `npm run build` from `desktop/frontend`.
- [ ] Commit the final implementation.
