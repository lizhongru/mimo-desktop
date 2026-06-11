package desktop

import (
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
