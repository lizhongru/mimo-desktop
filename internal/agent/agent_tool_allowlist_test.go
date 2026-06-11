package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/mimo-cli/mimo-cli/internal/llm"
	"github.com/mimo-cli/mimo-cli/internal/tools"
)

type allowlistTestTool struct {
	name string
}

func (t allowlistTestTool) Name() string { return t.name }

func (t allowlistTestTool) Description() string { return "test tool " + t.name }

func (t allowlistTestTool) GetSafetyLevel() tools.SafetyLevel { return tools.SafetyLow }

func (t allowlistTestTool) Parameters() map[string]interface{} {
	return map[string]interface{}{"type": "object"}
}

func (t allowlistTestTool) Execute(context.Context, map[string]interface{}) (*tools.ToolResult, error) {
	return &tools.ToolResult{Output: "ran " + t.name}, nil
}

func (t allowlistTestTool) Validate(map[string]interface{}) error { return nil }

func (t allowlistTestTool) RequiresConfirmation(map[string]interface{}) bool { return false }

func TestToolDefinitionsHonorToolAllowlist(t *testing.T) {
	registry := tools.NewRegistry()
	registry.Register(allowlistTestTool{name: "file_read"})
	registry.Register(allowlistTestTool{name: "shell"})

	a := &Agent{registry: registry}
	a.SetToolAllowlist([]string{"file_read"})

	defs := a.toolDefinitions()
	if got, want := len(defs), 1; got != want {
		t.Fatalf("tool definition count = %d, want %d", got, want)
	}
	if got, want := defs[0].Function.Name, "file_read"; got != want {
		t.Fatalf("tool definition name = %q, want %q", got, want)
	}
}

func TestExecuteToolCallRejectsDisallowedTool(t *testing.T) {
	registry := tools.NewRegistry()
	registry.Register(allowlistTestTool{name: "file_read"})
	registry.Register(allowlistTestTool{name: "shell"})

	a := &Agent{registry: registry}
	a.SetToolAllowlist([]string{"file_read"})

	_, err := a.executeToolCall(context.Background(), llm.ToolCall{
		Function: llm.FunctionCall{
			Name:      "shell",
			Arguments: "{}",
		},
	})
	if err == nil {
		t.Fatal("executeToolCall returned nil error for disallowed tool")
	}
	if !strings.Contains(err.Error(), "not allowed") {
		t.Fatalf("executeToolCall error = %q, want not allowed", err.Error())
	}
}
