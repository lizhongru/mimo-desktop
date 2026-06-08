package tools

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// EnvTool reads or lists environment variables
type EnvTool struct{}

func NewEnvTool() *EnvTool { return &EnvTool{} }

func (t *EnvTool) Name() string        { return "env" }
func (t *EnvTool) Description() string  { return "Read or list environment variables" }
func (t *EnvTool) GetSafetyLevel() SafetyLevel { return SafetyLow }
func (t *EnvTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Action: get or list",
				"enum":        []string{"get", "list"},
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Variable name (required for get action)",
			},
			"filter": map[string]interface{}{
				"type":        "string",
				"description": "Filter variables by name prefix (for list action)",
			},
		},
		"required": []string{"action"},
	}
}
func (t *EnvTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "action")
	return err
}
func (t *EnvTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *EnvTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	action, _ := StringParam(params, "action")

	switch action {
	case "get":
		name, err := StringParam(params, "name")
		if err != nil {
			return ToolError("name is required for get action"), nil
		}
		val := os.Getenv(name)
		if val == "" {
			return &ToolResult{Output: fmt.Sprintf("%s is not set", name)}, nil
		}
		return &ToolResult{Output: fmt.Sprintf("%s=%s", name, val)}, nil

	case "list":
		filter, _ := StringParam(params, "filter")
		var sb strings.Builder
		count := 0
		for _, env := range os.Environ() {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) != 2 {
				continue
			}
			if filter != "" && !strings.HasPrefix(strings.ToUpper(parts[0]), strings.ToUpper(filter)) {
				continue
			}
			sb.WriteString(env + "\n")
			count++
			if count >= 50 {
				sb.WriteString("... (truncated, showing first 50)\n")
				break
			}
		}
		if count == 0 {
			return &ToolResult{Output: "No matching environment variables"}, nil
		}
		return &ToolResult{Output: sb.String()}, nil

	default:
		return ToolError("unknown action: %s (use get/list)", action), nil
	}
}
