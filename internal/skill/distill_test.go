package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDistillAnalyze(t *testing.T) {
	config := DefaultDistillConfig()
	config.MinConfidence = 0.1
	distill := NewDistill(config, t.TempDir())

	sessionData := "$ go test ./...\n$ go build ./...\n$ go test ./...\n$ go build ./...\n$ go test ./...\n$ go build ./...\n$ go test ./...\n$ go build ./...\n"

	candidates, err := distill.Analyze(sessionData)
	if err != nil {
		t.Fatalf("failed to analyze: %v", err)
	}

	if len(candidates) == 0 {
		t.Error("expected some candidates")
	}
}

func TestDistillSaveCandidates(t *testing.T) {
	tmpDir := t.TempDir()
	distill := NewDistill(DefaultDistillConfig(), tmpDir)

	candidates := []SkillCandidate{
		{Name: "test_skill", Description: "Test skill", Confidence: 0.8},
	}

	err := distill.SaveCandidates(candidates)
	if err != nil {
		t.Fatalf("failed to save candidates: %v", err)
	}

	// Check if file was created
	skillFile := filepath.Join(tmpDir, ".mimo", "skills", "candidates.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		t.Error("candidates.md should exist")
	}
}

func TestDistillDisabled(t *testing.T) {
	config := DefaultDistillConfig()
	config.Enabled = false
	distill := NewDistill(config, t.TempDir())

	candidates, err := distill.Analyze("test data")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if candidates != nil {
		t.Error("expected nil candidates when disabled")
	}
}
