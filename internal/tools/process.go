package tools

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// ProcessTool lists or kills system processes
type ProcessTool struct{}

func NewProcessTool() *ProcessTool { return &ProcessTool{} }

func (t *ProcessTool) Name() string        { return "process" }
func (t *ProcessTool) Description() string  { return "List running processes or kill a process by PID" }
func (t *ProcessTool) GetSafetyLevel() SafetyLevel { return SafetyHigh }
func (t *ProcessTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Action: list or kill",
				"enum":        []string{"list", "kill"},
			},
			"pid": map[string]interface{}{
				"type":        "integer",
				"description": "Process ID to kill (required for kill action)",
			},
			"name_filter": map[string]interface{}{
				"type":        "string",
				"description": "Filter processes by name (for list action)",
			},
		},
		"required": []string{"action"},
	}
}
func (t *ProcessTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "action")
	return err
}
func (t *ProcessTool) RequiresConfirmation(params map[string]interface{}) bool {
	action, _ := StringParam(params, "action")
	return action == "kill"
}

func (t *ProcessTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	action, _ := StringParam(params, "action")

	switch action {
	case "list":
		var cmd *exec.Cmd
		nameFilter, _ := StringParam(params, "name_filter")
		if runtime.GOOS == "windows" {
			if nameFilter != "" {
				cmd = exec.CommandContext(ctx, "powershell", "-Command",
					fmt.Sprintf("Get-Process | Where-Object { $_.ProcessName -like '*%s*' } | Format-Table Id, ProcessName, CPU, WorkingSet -AutoSize", nameFilter))
			} else {
				cmd = exec.CommandContext(ctx, "powershell", "-Command",
					"Get-Process | Sort-Object CPU -Descending | Select-Object -First 30 | Format-Table Id, ProcessName, CPU, WorkingSet -AutoSize")
			}
		} else {
			if nameFilter != "" {
				cmd = exec.CommandContext(ctx, "sh", "-c", "ps aux | grep -i "+nameFilter)
			} else {
				cmd = exec.CommandContext(ctx, "sh", "-c", "ps aux --sort=-%cpu | head -30")
			}
		}
		out, err := cmd.CombinedOutput()
		if err != nil {
			return ToolError("process list failed: %v\n%s", err, string(out)), nil
		}
		result := strings.TrimSpace(string(out))
		if result == "" {
			return &ToolResult{Output: "No matching processes found"}, nil
		}
		return &ToolResult{Output: result}, nil

	case "kill":
		pid, err := IntParam(params, "pid")
		if err != nil {
			return ToolError("pid is required for kill action"), nil
		}
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.CommandContext(ctx, "taskkill", "/F", "/PID", fmt.Sprintf("%d", pid))
		} else {
			cmd = exec.CommandContext(ctx, "kill", fmt.Sprintf("%d", pid))
		}
		out, err := cmd.CombinedOutput()
		if err != nil {
			return ToolError("kill failed: %v\n%s", err, string(out)), nil
		}
		return &ToolResult{Output: fmt.Sprintf("Process %d killed", pid)}, nil

	default:
		return ToolError("unknown action: %s (use list/kill)", action), nil
	}
}
