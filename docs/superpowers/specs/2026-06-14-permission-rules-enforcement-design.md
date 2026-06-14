# Permission Rules Enforcement Design

> Date: 2026-06-14
> Scope: MiMo Desktop tool execution permission enforcement

## Goal

Persisted `permission.rules` must affect live tool execution. When the user saves advanced permission settings, subsequent agent tool calls should honor the configured `allow`, `ask`, and `deny` decisions without restarting the app.

## Scope

This change covers the already persisted rule groups used by the advanced settings UI:

- `read`
- `write`
- `edit`
- `bash`
- `external_directory`

Rules apply before a tool executes. They do not replace the existing safety classifier. They only decide whether a tool category is allowed, denied, or requires confirmation. The existing `safety.LevelLockdown`, `safety.LevelConfirm`, `safety.LevelAuto`, blocked commands, protected files, and read-only/write/exec permission mode remain active.

## Approach

The recommended approach is to attach `permission.Ruleset` to `safety.Guardrail` and evaluate it inside `Guardrail.CheckWithConfirmAll`. This keeps one enforcement path for ReAct, streamed ReAct, Plan-Execute, and future MCP tools because those paths already call the guardrail before execution.

The alternatives considered were:

- Enforce in `tools.Registry.Execute`: catches all direct registry calls, but has no access to the existing confirmation bridge and would duplicate safety behavior.
- Enforce in each tool implementation: explicit but repetitive, easy to miss new tools, and harder to test consistently.
- Enforce in `agent.executeToolCall` only: covers normal chat, but risks divergence with plan execution and non-agent callers.

## Permission Mapping

Tool names map to permission categories before rule evaluation:

- `file_read`, `dir_list`, `search`, `glob`, `git_status`, `git_diff`, `git_log`, `web_fetch`, `web_search`, and read-like MCP tools map to `read`.
- `file_write`, `file_delete`, `dir_create`, `clipboard`, and create/write-like MCP tools map to `write`.
- `file_edit`, `file_diff`, and patch/edit-like MCP tools map to `edit`.
- `shell`, process execution tools, and command-like MCP tools map to `bash`.
- Any file operation targeting a path outside the current working directory maps to `external_directory` first. If that rule allows the path, the regular action category still applies.

Unknown tools keep the existing safety classifier behavior and use their raw tool name for permission lookup. If no permission rule matches, the default decision is `ask`.

## Decision Flow

For each tool call:

1. Classify the tool action with the existing safety classifier.
2. Evaluate `external_directory` when path-like parameters point outside the workspace.
3. Evaluate the mapped permission category.
4. If a matching rule returns `deny`, block execution and return an error.
5. If a matching rule returns `ask`, use the existing confirmation callback unless `confirmAll` is active.
6. If rules allow the action, continue through the existing safety level checks.
7. Audit logging still records the classified action.

The rule layer is intentionally conservative. Invalid rule actions are ignored when converting config to a runtime ruleset, so malformed saved config does not accidentally allow more access. If every rule is invalid or missing, defaults from `permission.DefaultRuleset()` are used.

## Runtime Updates

`desktop.NewApp` converts `cfg.Permission.Rules` into `permission.Ruleset` and installs it on the guardrail.

`desktop.UpdateAdvancedSettings` saves the config and immediately refreshes the active guardrail rules. This makes the current app session honor new settings after the save promise resolves.

`desktop.SetSafetyLevel` recreates the guardrail, so it also reinstalls the active runtime rules.

## Error Handling

Denied permission errors should include the mapped permission category and tool name. Confirmation denials should continue to use the existing safety confirmation flow so the frontend sees one confirmation dialog style.

Invalid or unknown config actions are skipped during conversion. Unknown permission names are preserved only if they are non-empty, which allows future tool-specific rules without widening current defaults.

## Tests

Add or update Go tests for:

- `permission` category mapping for core tools.
- Guardrail denial when `bash` is configured as `deny`.
- Guardrail confirmation when `write` is configured as `ask`.
- Agent execution refusing a denied `shell` call before the tool runs.
- `UpdateAdvancedSettings` refreshing runtime rules without app restart.

Existing tests for config persistence and agent tool allowlists should continue to pass.

