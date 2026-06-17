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

func TestSelectedCommandOnlySkillRunsDirectly(t *testing.T) {
	projectDir := t.TempDir()
	skillDir := filepath.Join(projectDir, ".mimo", "skills", "skill_npm_run_build")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("create skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: skill_npm_run_build\ndescription: Run frontend build\n---\n\n# skill_npm_run_build\n\nRun frontend build\n\n## Pattern\n\n```text\nnpm run build\n```\n\n## Commands\n\n- `npm run build`\n"), 0644); err != nil {
		t.Fatalf("write skill file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".mimo", "skills", "enabled.json"), []byte(`{"skills":["skill_npm_run_build"]}`), 0644); err != nil {
		t.Fatalf("write enabled skills: %v", err)
	}
	frontendDir := filepath.Join(projectDir, "desktop", "frontend")
	if err := os.MkdirAll(frontendDir, 0755); err != nil {
		t.Fatalf("create frontend dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(frontendDir, "package.json"), []byte(`{"scripts":{"build":"vite build"}}`), 0644); err != nil {
		t.Fatalf("write package json: %v", err)
	}

	runs := selectedCommandOnlySkillRuns(projectDir, []string{"skill_npm_run_build"})
	if len(runs) != 1 {
		t.Fatalf("expected one direct skill command, got %#v", runs)
	}
	if runs[0].Skill != "skill_npm_run_build" {
		t.Fatalf("skill = %q", runs[0].Skill)
	}
	if runs[0].Command != "npm run build" {
		t.Fatalf("command = %q", runs[0].Command)
	}
	if runs[0].WorkingDir != frontendDir {
		t.Fatalf("working dir = %q, want %q", runs[0].WorkingDir, frontendDir)
	}
}

func TestSelectedCommandOnlySkillSkipsComplexWorkflow(t *testing.T) {
	projectDir := t.TempDir()
	skillDir := filepath.Join(projectDir, ".mimo", "skills", "complex_skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("create skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: complex_skill\ndescription: Complex workflow\n---\n\n# complex_skill\n\n## Workflow\n\nInspect files first.\n\n## Commands\n\n- `npm run build`\n"), 0644); err != nil {
		t.Fatalf("write skill file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".mimo", "skills", "enabled.json"), []byte(`{"skills":["complex_skill"]}`), 0644); err != nil {
		t.Fatalf("write enabled skills: %v", err)
	}

	runs := selectedCommandOnlySkillRuns(projectDir, []string{"complex_skill"})
	if len(runs) != 0 {
		t.Fatalf("expected complex workflow to fall back to agent, got %#v", runs)
	}
}

func TestSelectedCommandOnlySkillSkipsUnsafeCommand(t *testing.T) {
	projectDir := t.TempDir()
	skillDir := filepath.Join(projectDir, ".mimo", "skills", "unsafe_skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("create skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: unsafe_skill\ndescription: Unsafe command\n---\n\n# unsafe_skill\n\n## Commands\n\n- `rm -rf .`\n"), 0644); err != nil {
		t.Fatalf("write skill file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".mimo", "skills", "enabled.json"), []byte(`{"skills":["unsafe_skill"]}`), 0644); err != nil {
		t.Fatalf("write enabled skills: %v", err)
	}

	runs := selectedCommandOnlySkillRuns(projectDir, []string{"unsafe_skill"})
	if len(runs) != 0 {
		t.Fatalf("expected unsafe command to fall back to agent, got %#v", runs)
	}
}

func TestSummarizeShellOutputStripsANSIAndKeepsImportantLines(t *testing.T) {
	output := "\x1b[2mdist/\x1b[22m\x1b[36massets/ToolsViewer.js\x1b[39m 5.22 kB\n" +
		"\x1b[32m✓ built in 4.74s\x1b[39m\n" +
		"--- stderr ---\n" +
		"(node:123) Warning: To load an ES module, set \"type\": \"module\" in the package.json\n" +
		"\x1b[33m(!) Some chunks are larger than 500 kB after minification. Consider:\x1b[39m\n" +
		"- Using dynamic import() to code-split the application\n"

	summary := summarizeShellOutput(output)
	if strings.Contains(summary, "\x1b") || strings.Contains(summary, "[2m") || strings.Contains(summary, "[39m") {
		t.Fatalf("summary still contains ANSI fragments: %q", summary)
	}
	if strings.Contains(summary, "ToolsViewer.js") {
		t.Fatalf("summary should omit verbose asset list: %q", summary)
	}
	for _, expected := range []string{"built in 4.74s", "Warning: To load an ES module", "Some chunks are larger"} {
		if !strings.Contains(summary, expected) {
			t.Fatalf("summary missing %q: %q", expected, summary)
		}
	}
}

func TestBuildSystemPromptIncludesEnabledProjectSkills(t *testing.T) {
	projectDir := t.TempDir()
	skillDir := filepath.Join(projectDir, ".mimo", "skills", "test_skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("create skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: test_skill\ndescription: Test skill\n---\n\n# test_skill\n\nCUSTOM ENABLED SKILL SENTINEL\n\n## Commands\n\n- `npm run build`"), 0644); err != nil {
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
	if !strings.Contains(prompt, "## Selected Project Skills") {
		t.Fatalf("system prompt does not include selected skills section")
	}
	if !strings.Contains(prompt, "run those exact commands") {
		t.Fatalf("system prompt should instruct agent to apply selected skill commands")
	}
	if !strings.Contains(prompt, "override prior conversation interpretations") {
		t.Fatalf("system prompt should isolate selected skills from prior context interpretations")
	}
	if !strings.Contains(prompt, "## Current Turn Selected Skill Commands") || !strings.Contains(prompt, "- npm run build") {
		t.Fatalf("system prompt should list current turn selected skill commands")
	}
	if !strings.Contains(prompt, "CUSTOM ENABLED SKILL SENTINEL") {
		t.Fatalf("system prompt does not include enabled skill content")
	}
}
