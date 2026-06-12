# Advanced Settings Persistence Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Persist the Advanced Settings panel to `~/.mimo/config.yaml` and make automatic checkpoint behavior read those saved values.

**Architecture:** Extend `internal/config.Config` with advanced settings sections, expose them through `desktop/app_config.go`, and make `desktop/app_checkpoint.go` build its runtime checkpoint manager from saved config. Update the React settings modal to load saved values and save them through a new Wails method.

**Tech Stack:** Go, YAML config via `gopkg.in/yaml.v3`, Wails method bindings, React + TypeScript.

---

## File Map

- Modify `internal/config/schema.go`: add `MemoryConfig`, `CheckpointConfig`, `PermissionConfig`, `PermissionRuleConfig`, and defaults.
- Modify `internal/config/config.go`: merge the new config sections.
- Add `internal/config/config_test.go`: verify defaults, save/load, and merge behavior.
- Modify `desktop/app_config.go`: add DTOs and `UpdateAdvancedSettings`.
- Add or modify `desktop/app_config_test.go`: verify backend save updates config and writes YAML.
- Modify `desktop/app_checkpoint.go`: build checkpoint runtime config from saved config.
- Modify `desktop/app_checkpoint_test.go`: verify disabled auto checkpoint prevents creation.
- Modify `desktop/frontend/src/App.tsx`: extend Wails type declarations.
- Modify `desktop/frontend/src/components/settings/AdvancedSettings.tsx`: accept initial config, expose save status, and call parent save.
- Modify `desktop/frontend/src/components/settings/SettingsPage.tsx`: load advanced config from `GetConfig`, pass it into `AdvancedSettings`, and call `UpdateAdvancedSettings`.
- Run `go test ./... -count=1`, `npx tsc --noEmit`, `npm run build`, and `git diff --check`.

---

### Task 1: Config Schema And Persistence

**Files:**
- Modify: `internal/config/schema.go`
- Modify: `internal/config/config.go`
- Add: `internal/config/config_test.go`

- [ ] **Step 1: Write failing config tests**

Add `internal/config/config_test.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfigIncludesAdvancedSettings(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.Memory.CCIndex {
		t.Fatal("default memory cc_index should be enabled")
	}
	if cfg.Memory.SearchScoreFloor != 0.15 {
		t.Fatalf("memory search score floor = %v, want 0.15", cfg.Memory.SearchScoreFloor)
	}
	if !cfg.Checkpoint.AutoCheckpoint {
		t.Fatal("default auto checkpoint should be enabled")
	}
	if cfg.Checkpoint.TokenThreshold != 0.75 {
		t.Fatalf("checkpoint threshold = %v, want 0.75", cfg.Checkpoint.TokenThreshold)
	}
	if cfg.Checkpoint.MaxCheckpoints != 10 {
		t.Fatalf("max checkpoints = %d, want 10", cfg.Checkpoint.MaxCheckpoints)
	}
	if len(cfg.Permission.Rules) == 0 {
		t.Fatal("default permission rules should not be empty")
	}
}

func TestLoadFromFileMergesAdvancedSettings(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	data := []byte(`
memory:
  cc_index: false
  search_score_floor: 0.3
checkpoint:
  auto_checkpoint: false
  token_threshold: 0.8
  max_checkpoints: 5
permission:
  rules:
    - permission: read
      action: allow
    - permission: bash
      action: deny
`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Memory.CCIndex {
		t.Fatal("memory cc_index = true, want false")
	}
	if cfg.Memory.SearchScoreFloor != 0.3 {
		t.Fatalf("memory search score floor = %v, want 0.3", cfg.Memory.SearchScoreFloor)
	}
	if cfg.Checkpoint.AutoCheckpoint {
		t.Fatal("auto checkpoint = true, want false")
	}
	if cfg.Checkpoint.TokenThreshold != 0.8 {
		t.Fatalf("checkpoint threshold = %v, want 0.8", cfg.Checkpoint.TokenThreshold)
	}
	if cfg.Checkpoint.MaxCheckpoints != 5 {
		t.Fatalf("max checkpoints = %d, want 5", cfg.Checkpoint.MaxCheckpoints)
	}
	if got := cfg.Permission.Rules[1].Action; got != "deny" {
		t.Fatalf("second permission action = %q, want deny", got)
	}
}
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```powershell
go test ./internal/config -count=1
```

Expected: fail because `Config` has no `Memory`, `Checkpoint`, or `Permission` fields.

- [ ] **Step 3: Implement config structs and defaults**

In `internal/config/schema.go`, add after `ContextConfig`:

```go
// MemoryConfig represents memory-system configuration.
type MemoryConfig struct {
	CCIndex          bool    `yaml:"cc_index" mapstructure:"cc_index"`
	SearchScoreFloor float64 `yaml:"search_score_floor" mapstructure:"search_score_floor"`
}

// CheckpointConfig represents checkpoint behavior configuration.
type CheckpointConfig struct {
	AutoCheckpoint     bool    `yaml:"auto_checkpoint" mapstructure:"auto_checkpoint"`
	TokenThreshold     float64 `yaml:"token_threshold" mapstructure:"token_threshold"`
	MaxCheckpoints     int     `yaml:"max_checkpoints" mapstructure:"max_checkpoints"`
	ReconstructOnResume bool   `yaml:"reconstruct_on_resume" mapstructure:"reconstruct_on_resume"`
	ContextBudget      int     `yaml:"context_budget" mapstructure:"context_budget"`
}

// PermissionConfig represents persisted permission rules.
type PermissionConfig struct {
	Rules []PermissionRuleConfig `yaml:"rules" mapstructure:"rules"`
}

// PermissionRuleConfig represents one persisted permission rule.
type PermissionRuleConfig struct {
	Permission string `yaml:"permission" mapstructure:"permission" json:"permission"`
	Action     string `yaml:"action" mapstructure:"action" json:"action"`
	Pattern    string `yaml:"pattern,omitempty" mapstructure:"pattern" json:"pattern,omitempty"`
}
```

Add these fields to `Config`:

```go
	// Memory
	Memory MemoryConfig `yaml:"memory" mapstructure:"memory"`

	// Checkpoint
	Checkpoint CheckpointConfig `yaml:"checkpoint" mapstructure:"checkpoint"`

	// Permission rules
	Permission PermissionConfig `yaml:"permission" mapstructure:"permission"`
```

Add default values in `DefaultConfig`:

```go
		Memory: MemoryConfig{
			CCIndex:          true,
			SearchScoreFloor: 0.15,
		},
		Checkpoint: CheckpointConfig{
			AutoCheckpoint:      true,
			TokenThreshold:      0.75,
			MaxCheckpoints:      10,
			ReconstructOnResume: true,
			ContextBudget:       128000,
		},
		Permission: PermissionConfig{
			Rules: []PermissionRuleConfig{
				{Permission: "read", Action: "allow"},
				{Permission: "write", Action: "ask"},
				{Permission: "edit", Action: "ask"},
				{Permission: "bash", Action: "ask"},
				{Permission: "external_directory", Action: "deny"},
			},
		},
```

- [ ] **Step 4: Implement merge behavior**

In `internal/config/config.go`, update `mergeConfig` to merge the new sections:

```go
	// Memory - merge fields
	dst.Memory.CCIndex = src.Memory.CCIndex
	if src.Memory.SearchScoreFloor != 0 {
		dst.Memory.SearchScoreFloor = src.Memory.SearchScoreFloor
	}

	// Checkpoint - merge fields
	dst.Checkpoint.AutoCheckpoint = src.Checkpoint.AutoCheckpoint
	if src.Checkpoint.TokenThreshold != 0 {
		dst.Checkpoint.TokenThreshold = src.Checkpoint.TokenThreshold
	}
	if src.Checkpoint.MaxCheckpoints != 0 {
		dst.Checkpoint.MaxCheckpoints = src.Checkpoint.MaxCheckpoints
	}
	dst.Checkpoint.ReconstructOnResume = src.Checkpoint.ReconstructOnResume
	if src.Checkpoint.ContextBudget != 0 {
		dst.Checkpoint.ContextBudget = src.Checkpoint.ContextBudget
	}

	// Permission - replace rules when provided
	if src.Permission.Rules != nil {
		dst.Permission.Rules = src.Permission.Rules
	}
```

- [ ] **Step 5: Run config tests**

Run:

```powershell
go test ./internal/config -count=1
```

Expected: pass.

---

### Task 2: Backend DTOs And Save Endpoint

**Files:**
- Modify: `desktop/app_config.go`
- Add: `desktop/app_config_test.go`

- [ ] **Step 1: Write failing backend test**

Add `desktop/app_config_test.go`:

```go
package desktop

import (
	"os"
	"path/filepath"
	"testing"

	iconfig "github.com/mimo-cli/mimo-cli/internal/config"
)

func TestUpdateAdvancedSettingsSavesConfig(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")

	app := &App{cfg: iconfig.DefaultConfig()}
	input := AdvancedSettingsDTO{
		Memory: MemorySettingsDTO{
			CCIndex:          false,
			SearchScoreFloor: 0.33,
		},
		Checkpoint: CheckpointSettingsDTO{
			AutoCheckpoint:     false,
			TokenThreshold:     0.82,
			MaxCheckpoints:     7,
			ReconstructOnResume: true,
			ContextBudget:      64000,
		},
		Permission: PermissionSettingsDTO{
			Rules: []PermissionRuleDTO{
				{Permission: "read", Action: "allow"},
				{Permission: "bash", Action: "deny"},
			},
		},
	}

	if err := app.UpdateAdvancedSettings(input); err != nil {
		t.Fatalf("update advanced settings: %v", err)
	}

	if app.cfg.Checkpoint.AutoCheckpoint {
		t.Fatal("in-memory auto checkpoint = true, want false")
	}
	configPath := filepath.Join(tmpHome, ".mimo", "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected config file at %s: %v", configPath, err)
	}

	loaded, err := iconfig.LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("load saved config: %v", err)
	}
	if loaded.Memory.SearchScoreFloor != 0.33 {
		t.Fatalf("saved memory floor = %v, want 0.33", loaded.Memory.SearchScoreFloor)
	}
	if loaded.Checkpoint.ContextBudget != 64000 {
		t.Fatalf("saved checkpoint budget = %d, want 64000", loaded.Checkpoint.ContextBudget)
	}
	if loaded.Permission.Rules[1].Action != "deny" {
		t.Fatalf("saved bash action = %q, want deny", loaded.Permission.Rules[1].Action)
	}
}
```

- [ ] **Step 2: Run test to verify failure**

Run:

```powershell
go test ./desktop -run TestUpdateAdvancedSettingsSavesConfig -count=1
```

Expected: fail because DTOs and `UpdateAdvancedSettings` do not exist.

- [ ] **Step 3: Add DTOs to app config**

In `desktop/app_config.go`, extend `AppConfigDTO`:

```go
	Memory     MemorySettingsDTO     `json:"memory"`
	Checkpoint CheckpointSettingsDTO `json:"checkpoint"`
	Permission PermissionSettingsDTO `json:"permission"`
```

Add DTO types near `AgentDTO`:

```go
// MemorySettingsDTO is frontend-friendly memory config.
type MemorySettingsDTO struct {
	CCIndex          bool    `json:"ccIndex"`
	SearchScoreFloor float64 `json:"searchScoreFloor"`
}

// CheckpointSettingsDTO is frontend-friendly checkpoint config.
type CheckpointSettingsDTO struct {
	AutoCheckpoint     bool    `json:"autoCheckpoint"`
	TokenThreshold     float64 `json:"tokenThreshold"`
	MaxCheckpoints     int     `json:"maxCheckpoints"`
	ReconstructOnResume bool   `json:"reconstructOnResume"`
	ContextBudget      int     `json:"contextBudget"`
}

// PermissionSettingsDTO is frontend-friendly permission config.
type PermissionSettingsDTO struct {
	Rules []PermissionRuleDTO `json:"rules"`
}

// PermissionRuleDTO is frontend-friendly permission rule config.
type PermissionRuleDTO struct {
	Permission string `json:"permission"`
	Action     string `json:"action"`
	Pattern    string `json:"pattern,omitempty"`
}
```

- [ ] **Step 4: Expose advanced config in GetConfig**

In `GetConfig`, populate:

```go
		Memory: MemorySettingsDTO{
			CCIndex:          a.cfg.Memory.CCIndex,
			SearchScoreFloor: a.cfg.Memory.SearchScoreFloor,
		},
		Checkpoint: CheckpointSettingsDTO{
			AutoCheckpoint:      a.cfg.Checkpoint.AutoCheckpoint,
			TokenThreshold:      a.cfg.Checkpoint.TokenThreshold,
			MaxCheckpoints:      a.cfg.Checkpoint.MaxCheckpoints,
			ReconstructOnResume: a.cfg.Checkpoint.ReconstructOnResume,
			ContextBudget:       a.cfg.Checkpoint.ContextBudget,
		},
		Permission: PermissionSettingsDTO{
			Rules: permissionRulesToDTO(a.cfg.Permission.Rules),
		},
```

Add helpers:

```go
func permissionRulesToDTO(rules []iconfig.PermissionRuleConfig) []PermissionRuleDTO {
	result := make([]PermissionRuleDTO, 0, len(rules))
	for _, rule := range rules {
		result = append(result, PermissionRuleDTO{
			Permission: rule.Permission,
			Action:     rule.Action,
			Pattern:    rule.Pattern,
		})
	}
	return result
}

func permissionRulesFromDTO(rules []PermissionRuleDTO) []iconfig.PermissionRuleConfig {
	result := make([]iconfig.PermissionRuleConfig, 0, len(rules))
	for _, rule := range rules {
		result = append(result, iconfig.PermissionRuleConfig{
			Permission: rule.Permission,
			Action:     rule.Action,
			Pattern:    rule.Pattern,
		})
	}
	return result
}
```

- [ ] **Step 5: Implement save endpoint**

Add to `desktop/app_config.go`:

```go
// UpdateAdvancedSettings updates memory, checkpoint, and permission settings.
func (a *App) UpdateAdvancedSettings(settings AdvancedSettingsDTO) error {
	a.cfg.Memory = iconfig.MemoryConfig{
		CCIndex:          settings.Memory.CCIndex,
		SearchScoreFloor: settings.Memory.SearchScoreFloor,
	}
	a.cfg.Checkpoint = iconfig.CheckpointConfig{
		AutoCheckpoint:      settings.Checkpoint.AutoCheckpoint,
		TokenThreshold:      settings.Checkpoint.TokenThreshold,
		MaxCheckpoints:      settings.Checkpoint.MaxCheckpoints,
		ReconstructOnResume: settings.Checkpoint.ReconstructOnResume,
		ContextBudget:       settings.Checkpoint.ContextBudget,
	}
	a.cfg.Permission = iconfig.PermissionConfig{
		Rules: permissionRulesFromDTO(settings.Permission.Rules),
	}
	return iconfig.SaveUserConfig(a.cfg)
}
```

Add the wrapper DTO:

```go
// AdvancedSettingsDTO groups advanced settings edited in one form.
type AdvancedSettingsDTO struct {
	Memory     MemorySettingsDTO     `json:"memory"`
	Checkpoint CheckpointSettingsDTO `json:"checkpoint"`
	Permission PermissionSettingsDTO `json:"permission"`
}
```

- [ ] **Step 6: Run backend config tests**

Run:

```powershell
go test ./desktop -run TestUpdateAdvancedSettingsSavesConfig -count=1
```

Expected: pass.

---

### Task 3: Checkpoint Runtime Uses Saved Config

**Files:**
- Modify: `desktop/app_checkpoint.go`
- Modify: `desktop/app_checkpoint_test.go`

- [ ] **Step 1: Write failing checkpoint test**

Add to `desktop/app_checkpoint_test.go`:

```go
func TestSaveSessionFromFrontendRespectsDisabledAutoCheckpointConfig(t *testing.T) {
	app, store, sessionID := newCheckpointTestApp(t, 20)
	app.cfg.Checkpoint.AutoCheckpoint = false
	app.cfg.Checkpoint.TokenThreshold = 0.01
	app.cfg.Checkpoint.MaxCheckpoints = 10
	app.cfg.Checkpoint.ContextBudget = 20

	err := app.SaveSessionFromFrontend(sessionID, []ChatMessageDTO{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "world"},
	})
	if err != nil {
		t.Fatalf("save session: %v", err)
	}

	checkpoints, err := store.ListCheckpoints(sessionID)
	if err != nil {
		t.Fatalf("list checkpoints: %v", err)
	}
	if len(checkpoints) != 0 {
		t.Fatalf("checkpoint count = %d, want 0", len(checkpoints))
	}
}
```

- [ ] **Step 2: Run test to verify failure**

Run:

```powershell
go test ./desktop -run TestSaveSessionFromFrontendRespectsDisabledAutoCheckpointConfig -count=1
```

Expected: fail because `maybeCreateAutoCheckpoint` currently ignores `cfg.Checkpoint.AutoCheckpoint`.

- [ ] **Step 3: Add runtime config helper**

In `desktop/app_checkpoint.go`, add:

```go
func (a *App) checkpointRuntimeConfig() context.CheckpointConfig {
	cfg := context.DefaultCheckpointConfig()
	if a.cfg == nil {
		return cfg
	}

	cfg.AutoCheckpoint = a.cfg.Checkpoint.AutoCheckpoint
	if a.cfg.Checkpoint.TokenThreshold > 0 {
		cfg.TokenThreshold = a.cfg.Checkpoint.TokenThreshold
	}
	if a.cfg.Checkpoint.MaxCheckpoints > 0 {
		cfg.MaxCheckpoints = a.cfg.Checkpoint.MaxCheckpoints
	}
	cfg.ReconstructOnResume = a.cfg.Checkpoint.ReconstructOnResume
	if a.cfg.Checkpoint.ContextBudget > 0 {
		cfg.ContextBudget = a.cfg.Checkpoint.ContextBudget
	} else if a.cfg.Context.MaxTokens > 0 {
		cfg.ContextBudget = a.cfg.Context.MaxTokens
	}
	return cfg
}
```

- [ ] **Step 4: Use helper in checkpoint creation**

Replace `context.DefaultCheckpointConfig()` usage in `createCheckpointForSession` and `maybeCreateAutoCheckpoint` with `a.checkpointRuntimeConfig()`.

In `maybeCreateAutoCheckpoint`, set:

```go
	checkpointCfg := a.checkpointRuntimeConfig()
	maxTokens := checkpointCfg.ContextBudget
```

Keep the rest of the function behavior unchanged.

- [ ] **Step 5: Run checkpoint tests**

Run:

```powershell
go test ./desktop -run Checkpoint -count=1
```

Expected: pass.

---

### Task 4: Frontend Advanced Settings Wiring

**Files:**
- Modify: `desktop/frontend/src/App.tsx`
- Modify: `desktop/frontend/src/components/settings/AdvancedSettings.tsx`
- Modify: `desktop/frontend/src/components/settings/SettingsPage.tsx`

- [ ] **Step 1: Update Wails TypeScript declaration**

In `desktop/frontend/src/App.tsx`, add types:

```ts
type AdvancedSettingsConfig = {
  memory: { ccIndex: boolean; searchScoreFloor: number };
  checkpoint: {
    autoCheckpoint: boolean;
    tokenThreshold: number;
    maxCheckpoints: number;
    reconstructOnResume: boolean;
    contextBudget: number;
  };
  permission: {
    rules: Array<{ permission: string; action: string; pattern?: string }>;
  };
};
```

Extend `GetConfig` return type with:

```ts
memory: AdvancedSettingsConfig["memory"];
checkpoint: AdvancedSettingsConfig["checkpoint"];
permission: AdvancedSettingsConfig["permission"];
```

Add method:

```ts
UpdateAdvancedSettings: (settings: AdvancedSettingsConfig) => Promise<void>;
```

- [ ] **Step 2: Make AdvancedSettings accept controlled initial config**

In `AdvancedSettings.tsx`, export config type and update `Props`:

```ts
export interface AdvancedSettingsConfig {
  checkpoint: CheckpointConfig;
  memory: MemoryConfig;
  permission: PermissionConfig;
}

interface Props {
  value: AdvancedSettingsConfig;
  saving?: boolean;
  onSave?: (config: AdvancedSettingsConfig) => void;
}
```

Add `useEffect` import and sync state:

```ts
useEffect(() => {
  setCheckpoint(value.checkpoint);
  setMemory(value.memory);
  setPermission(value.permission);
}, [value]);
```

Change save button disabled state and label:

```tsx
<button
  onClick={handleSave}
  disabled={saving}
  className="flex items-center gap-2 px-4 py-2 bg-accent text-white rounded-md text-sm hover:bg-accent/90 disabled:opacity-60 disabled:cursor-not-allowed"
>
  <Save className="w-4 h-4" />
  {saving ? "保存中..." : "保存设置"}
</button>
```

- [ ] **Step 3: Map permission rules in SettingsPage**

In `SettingsPage.tsx`, import `type AdvancedSettingsConfig`.

Add default config:

```ts
const defaultAdvancedSettings: AdvancedSettingsConfig = {
  checkpoint: {
    autoCheckpoint: true,
    tokenThreshold: 0.75,
    maxCheckpoints: 10,
    reconstructOnResume: true,
    contextBudget: 128000,
  },
  memory: {
    ccIndex: true,
    searchScoreFloor: 0.15,
  },
  permission: {
    read: "allow",
    write: "ask",
    edit: "ask",
    bash: "ask",
  },
};
```

Add local state:

```ts
const [advancedSettings, setAdvancedSettings] = useState<AdvancedSettingsConfig>(defaultAdvancedSettings);
const [advancedSaving, setAdvancedSaving] = useState(false);
```

Add helpers:

```ts
function permissionRulesToForm(rules?: Array<{ permission: string; action: string }>) {
  const result = { ...defaultAdvancedSettings.permission };
  for (const rule of rules || []) {
    if (rule.permission === "read" || rule.permission === "write" || rule.permission === "edit" || rule.permission === "bash") {
      result[rule.permission] = rule.action;
    }
  }
  return result;
}

function permissionFormToRules(permission: AdvancedSettingsConfig["permission"]) {
  return Object.entries(permission).map(([key, action]) => ({
    permission: key,
    action,
  }));
}
```

When `GetConfig` resolves, set:

```ts
setAdvancedSettings({
  checkpoint: cfg.checkpoint || defaultAdvancedSettings.checkpoint,
  memory: cfg.memory || defaultAdvancedSettings.memory,
  permission: permissionRulesToForm(cfg.permission?.rules),
});
```

Use the component:

```tsx
<AdvancedSettings
  value={advancedSettings}
  saving={advancedSaving}
  onSave={(config) => {
    setAdvancedSaving(true);
    window.go?.desktop?.App?.UpdateAdvancedSettings?.({
      checkpoint: config.checkpoint,
      memory: config.memory,
      permission: { rules: permissionFormToRules(config.permission) },
    })
      .then(() => setAdvancedSettings(config))
      .catch(console.error)
      .finally(() => setAdvancedSaving(false));
  }}
/>
```

- [ ] **Step 4: Run frontend type check**

Run:

```powershell
cd desktop/frontend
npx tsc --noEmit
```

Expected: pass.

---

### Task 5: Full Verification And Commit

**Files:**
- All files from previous tasks.

- [ ] **Step 1: Run Go tests**

Run:

```powershell
go test ./... -count=1
```

Expected: pass.

- [ ] **Step 2: Run frontend checks**

Run:

```powershell
cd desktop/frontend
npx tsc --noEmit
npm run build
```

Expected: both pass. Existing bundle-size and Node ESM warnings are acceptable if the commands exit 0.

- [ ] **Step 3: Run diff whitespace check**

Run:

```powershell
cd ../..
git diff --check
```

Expected: exit 0.

- [ ] **Step 4: Review diff**

Run:

```powershell
git diff --stat
git diff -- internal/config/schema.go internal/config/config.go desktop/app_config.go desktop/app_checkpoint.go desktop/frontend/src/components/settings/AdvancedSettings.tsx desktop/frontend/src/components/settings/SettingsPage.tsx desktop/frontend/src/App.tsx
```

Expected: diff only covers advanced settings persistence.

- [ ] **Step 5: Commit**

Run:

```powershell
git add internal/config/schema.go internal/config/config.go internal/config/config_test.go desktop/app_config.go desktop/app_config_test.go desktop/app_checkpoint.go desktop/app_checkpoint_test.go desktop/frontend/src/App.tsx desktop/frontend/src/components/settings/AdvancedSettings.tsx desktop/frontend/src/components/settings/SettingsPage.tsx docs/superpowers/plans/2026-06-12-advanced-settings-persistence.md
git commit -m "feat: persist advanced settings"
```

Expected: commit succeeds.
