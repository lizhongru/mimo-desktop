# Permission Rules Enforcement Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enforce saved advanced permission rules during live tool execution.

**Architecture:** Convert persisted config rules into `permission.Ruleset`, install them on `safety.Guardrail`, and evaluate them before existing safety level decisions. Keep tool execution paths unchanged so ReAct, streaming, and Plan-Execute continue to use the same guardrail.

**Tech Stack:** Go, existing MiMo Desktop config DTOs, `internal/permission`, `internal/safety`, `internal/agent`, Wails backend tests.

---

## File Structure

- Modify `internal/permission/ruleset.go` to add tool-to-permission category mapping and safer runtime rule helpers.
- Modify `internal/permission/ruleset_test.go` to cover category mapping and default-safe rule behavior.
- Modify `internal/safety/guardrail.go` to store and evaluate a runtime permission ruleset.
- Add or modify `internal/safety/guardrail_test.go` to verify deny and ask behavior.
- Modify `internal/agent/agent_tool_allowlist_test.go` or add a focused agent test to verify denied tools do not execute.
- Modify `desktop/app_config.go` and `desktop/app.go` to convert config rules and refresh guardrail rules after settings changes.
- Modify `desktop/app_config_test.go` to verify settings save updates the live guardrail.

### Task 1: Permission Runtime Helpers

**Files:**
- Modify: `internal/permission/ruleset.go`
- Modify: `internal/permission/ruleset_test.go`

- [x] **Step 1: Write failing tests**

Add tests that expect core tools to map to `read`, `write`, `edit`, and `bash`, and expect invalid config actions to fall back to defaults.

- [x] **Step 2: Run tests to verify failure**

Run: `go test ./internal/permission -run "TestPermissionForTool|TestRulesetFromConfig" -count=1`

Expected: FAIL because helper functions do not exist yet.

- [x] **Step 3: Implement minimal helpers**

Add exported helpers for category mapping and config conversion.

- [x] **Step 4: Run tests to verify pass**

Run: `go test ./internal/permission -count=1`

Expected: PASS.

### Task 2: Guardrail Permission Enforcement

**Files:**
- Modify: `internal/safety/guardrail.go`
- Add: `internal/safety/guardrail_test.go`

- [x] **Step 1: Write failing tests**

Add tests for:

- `bash=deny` blocks `shell`.
- `write=ask` invokes confirmation for `file_write`.
- `confirmAll=true` bypasses the rule confirmation but not deny.

- [x] **Step 2: Run tests to verify failure**

Run: `go test ./internal/safety -run "TestGuardrailPermission" -count=1`

Expected: FAIL because guardrail has no runtime ruleset.

- [x] **Step 3: Implement minimal enforcement**

Store rules on `Guardrail`, add `SetRuleset`, evaluate mapped permission action before safety-level decisions, and reuse the existing confirmation callback for `ask`.

- [x] **Step 4: Run tests to verify pass**

Run: `go test ./internal/safety -count=1`

Expected: PASS.

### Task 3: Agent Execution Coverage

**Files:**
- Modify: `internal/agent/agent_tool_allowlist_test.go`

- [x] **Step 1: Write failing test**

Add a test that registers a `shell` test tool, configures guardrail rules with `bash=deny`, calls `executeToolCall`, and asserts the test tool was not executed.

- [x] **Step 2: Run tests to verify failure**

Run: `go test ./internal/agent -run TestExecuteToolCallRejectsDeniedPermission -count=1`

Expected: FAIL before guardrail enforcement is wired into safety.

- [x] **Step 3: Keep implementation in guardrail**

No agent production code should be needed unless the test reveals a missing execution path.

- [x] **Step 4: Run tests to verify pass**

Run: `go test ./internal/agent -count=1`

Expected: PASS.

### Task 4: Desktop Config Runtime Refresh

**Files:**
- Modify: `desktop/app.go`
- Modify: `desktop/app_config.go`
- Modify: `desktop/app_config_test.go`

- [x] **Step 1: Write failing test**

Add a test that creates an `App` with an existing guardrail, calls `UpdateAdvancedSettings` with `bash=deny`, then checks that the active guardrail blocks `shell`.

- [x] **Step 2: Run tests to verify failure**

Run: `go test ./desktop -run TestUpdateAdvancedSettingsRefreshesPermissionRules -count=1`

Expected: FAIL because `UpdateAdvancedSettings` currently only saves config.

- [x] **Step 3: Implement runtime refresh**

Add conversion from DTO/config permission rules to `permission.Ruleset`, install rules in `NewApp`, reinstall rules in `SetSafetyLevel`, and refresh rules in `UpdateAdvancedSettings`.

- [x] **Step 4: Run tests to verify pass**

Run: `go test ./desktop -run TestUpdateAdvancedSettingsRefreshesPermissionRules -count=1`

Expected: PASS.

### Task 5: Full Verification

**Files:**
- Verify all modified Go packages and whitespace.

- [x] **Step 1: Run focused package tests**

Run: `go test ./internal/permission ./internal/safety ./internal/agent ./desktop -count=1`

Expected: PASS.

- [x] **Step 2: Run full Go suite**

Run: `go test ./... -count=1`

Expected: PASS.

- [x] **Step 3: Run whitespace check**

Run: `git diff --check`

Expected: PASS, ignoring no files.
