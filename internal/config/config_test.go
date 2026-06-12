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

func TestLoadFromFilePreservesAdvancedDefaultsWhenSectionsAreAbsent(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.yaml")
	data := []byte(`default_model: mimo`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if !cfg.Memory.CCIndex {
		t.Fatal("memory cc_index should keep its default when section is absent")
	}
	if !cfg.Checkpoint.AutoCheckpoint {
		t.Fatal("auto checkpoint should keep its default when section is absent")
	}
	if !cfg.Checkpoint.ReconstructOnResume {
		t.Fatal("reconstruct_on_resume should keep its default when section is absent")
	}
}
