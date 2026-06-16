package desktop

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mimo-cli/mimo-cli/internal/agent"
	iconfig "github.com/mimo-cli/mimo-cli/internal/config"
	"github.com/mimo-cli/mimo-cli/internal/ignore"
)

func TestBuildSystemPromptIncludesCurrentAgentPrompt(t *testing.T) {
	t.Cleanup(func() {
		multiAgentManager = nil
	})

	manager := getMultiAgentManager()
	manager.SetConfig("plan", &agent.AgentConfig{
		Name:        "Plan",
		Mode:        agent.ModePlan,
		Color:       "#3b82f6",
		Description: "Plan only",
		Prompt:      "CUSTOM PLAN PROMPT SENTINEL",
	})
	if err := manager.SetCurrent("plan"); err != nil {
		t.Fatalf("set current agent: %v", err)
	}

	app := &App{
		cfg:           iconfig.DefaultConfig(),
		ignoreMatcher: ignore.New(),
	}

	prompt := app.buildSystemPrompt(t.TempDir())
	if !strings.Contains(prompt, "CUSTOM PLAN PROMPT SENTINEL") {
		t.Fatalf("system prompt does not include current agent prompt")
	}
}

func TestBuildSystemPromptIncludesEnabledProjectSkills(t *testing.T) {
	projectDir := t.TempDir()
	skillDir := filepath.Join(projectDir, ".mimo", "skills", "test_skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("create skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: test_skill\ndescription: Test skill\n---\n\n# test_skill\n\nCUSTOM ENABLED SKILL SENTINEL"), 0644); err != nil {
		t.Fatalf("write skill file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".mimo", "skills", "enabled.json"), []byte(`{"skills":["test_skill"]}`), 0644); err != nil {
		t.Fatalf("write enabled skills: %v", err)
	}

	app := &App{
		cfg:           iconfig.DefaultConfig(),
		ignoreMatcher: ignore.New(),
	}

	prompt := app.buildSystemPrompt(projectDir)
	if strings.Contains(prompt, "CUSTOM ENABLED SKILL SENTINEL") {
		t.Fatalf("system prompt should not include enabled skill content until selected")
	}

	prompt = app.buildSystemPrompt(projectDir, "test_skill")
	if !strings.Contains(prompt, "## Enabled Project Skills") {
		t.Fatalf("system prompt does not include enabled skills section")
	}
	if !strings.Contains(prompt, "CUSTOM ENABLED SKILL SENTINEL") {
		t.Fatalf("system prompt does not include enabled skill content")
	}
}
