package skill

import (
	"os"
	"path/filepath"
	"strings"
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

	skillFile := filepath.Join(tmpDir, ".mimo", "skills", "candidates.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		t.Error("candidates.md should exist")
	}
}

func TestDistillSaveCandidatesWritesSkillFiles(t *testing.T) {
	tmpDir := t.TempDir()
	distill := NewDistill(DefaultDistillConfig(), tmpDir)

	candidates := []SkillCandidate{
		{
			Name:        "test skill",
			Description: "Test skill",
			Confidence:  0.8,
			Pattern:     "go test ./...",
			Commands:    []string{"go test ./..."},
		},
	}

	if err := distill.SaveCandidates(candidates); err != nil {
		t.Fatalf("failed to save candidates: %v", err)
	}

	skillFile := filepath.Join(tmpDir, ".mimo", "skills", "test_skill", "SKILL.md")
	data, err := os.ReadFile(skillFile)
	if err != nil {
		t.Fatalf("expected SKILL.md to exist: %v", err)
	}

	content := string(data)
	for _, expected := range []string{
		"name: test_skill",
		"description: Test skill",
		"confidence: 0.80",
		"go test ./...",
	} {
		if !strings.Contains(content, expected) {
			t.Fatalf("expected SKILL.md to contain %q, got:\n%s", expected, content)
		}
	}
}

func TestParseCandidatesMarkdown(t *testing.T) {
	input := "# Skill Candidates\n\n" +
		"Generated: 2026-06-16T00:00:00+08:00\n\n" +
		"## skill_go_test\n\n" +
		"- **Description**: Automated skill for: go test ./...\n" +
		"- **Confidence**: 0.80\n" +
		"- **Pattern**: go test ./...\n" +
		"- **Commands**:\n" +
		"  - `go test ./...`\n\n" +
		"## build_workflow\n\n" +
		"- **Description**: Build workflow\n" +
		"- **Confidence**: 0.70\n"

	candidates := ParseCandidatesMarkdown([]byte(input))
	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(candidates))
	}

	first := candidates[0]
	if first.Name != "skill_go_test" {
		t.Fatalf("expected first name skill_go_test, got %q", first.Name)
	}
	if first.Description != "Automated skill for: go test ./..." {
		t.Fatalf("unexpected description: %q", first.Description)
	}
	if first.Confidence != 0.8 {
		t.Fatalf("expected confidence 0.8, got %.2f", first.Confidence)
	}
	if first.Pattern != "go test ./..." {
		t.Fatalf("unexpected pattern: %q", first.Pattern)
	}
	if len(first.Commands) != 1 || first.Commands[0] != "go test ./..." {
		t.Fatalf("unexpected commands: %#v", first.Commands)
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
