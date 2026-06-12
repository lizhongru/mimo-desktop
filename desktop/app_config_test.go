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
			AutoCheckpoint:      false,
			TokenThreshold:      0.82,
			MaxCheckpoints:      7,
			ReconstructOnResume: true,
			ContextBudget:       64000,
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
