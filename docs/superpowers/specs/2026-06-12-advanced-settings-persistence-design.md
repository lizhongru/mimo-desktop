# Advanced Settings Persistence Design

## Goal

Persist the Advanced Settings panel to `~/.mimo/config.yaml` and make automatic checkpoint behavior read those saved values.

## Scope

This change covers the existing Advanced Settings groups:

- Checkpoint settings: auto checkpoint, token threshold, maximum checkpoints.
- Memory settings: Claude Code memory indexing flag, search score floor.
- Permission settings: read, write, edit, bash actions.

Permission settings are persisted now but are not wired into tool execution in this change.

## Approach

Add `MemoryConfig`, `CheckpointConfig`, and `PermissionConfig` to `internal/config.Config`, including default values and merge behavior for YAML loading. Expose the three config sections through `desktop/app_config.go` via `GetConfig`, and add one save endpoint that updates all advanced settings at once.

The checkpoint runtime should use the saved checkpoint config when creating manual and automatic checkpoints. The UI should load saved values when the settings modal opens, keep local form state while editing, and call the new backend save endpoint from the Advanced Settings save button.

## Data Model

`MemoryConfig` stores:

- `cc_index bool`
- `search_score_floor float64`

`CheckpointConfig` stores:

- `auto_checkpoint bool`
- `token_threshold float64`
- `max_checkpoints int`
- `reconstruct_on_resume bool`
- `context_budget int`

`PermissionConfig` stores:

- `rules []PermissionRuleConfig`

Each `PermissionRuleConfig` stores:

- `permission string`
- `action string`
- optional `pattern string`

The frontend keeps its current simple four-control permission view and maps it to rules for `read`, `write`, `edit`, and `bash`.

## Runtime Behavior

Manual checkpoint creation uses the saved checkpoint config for file checkpoint generation.

Automatic checkpoint creation uses:

- saved `auto_checkpoint`
- saved `token_threshold`
- saved `max_checkpoints`
- saved `context_budget`, falling back to `cfg.Context.MaxTokens` when no explicit checkpoint budget is set

Saving Advanced Settings updates `a.cfg` in memory and writes the full config through `SaveUserConfig`.

## Testing

Add Go tests for config defaults, YAML persistence, and backend advanced-setting save behavior.

Add a checkpoint test proving that disabling `auto_checkpoint` prevents automatic checkpoint creation.

Run frontend type checking to verify the new Wails method typings used by the settings modal.
