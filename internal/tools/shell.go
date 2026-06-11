package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// ShellTool executes shell commands
type ShellTool struct{}

func NewShellTool() *ShellTool {
	return &ShellTool{}
}

func (t *ShellTool) Name() string { return "shell" }

func (t *ShellTool) Description() string {
	return "Execute shell commands in the system terminal"
}

func (t *ShellTool) GetSafetyLevel() SafetyLevel { return SafetyHigh }

func (t *ShellTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "The shell command to execute",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in seconds (default: 120)",
			},
			"working_dir": map[string]interface{}{
				"type":        "string",
				"description": "Working directory for the command",
			},
		},
		"required": []string{"command"},
	}
}

func (t *ShellTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "command")
	return err
}

func (t *ShellTool) RequiresConfirmation(params map[string]interface{}) bool {
	cmd, _ := StringParam(params, "command")
	return isDangerousCommand(cmd)
}

func (t *ShellTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	command, _ := StringParam(params, "command")
	timeout := 120
	if t, err := IntParam(params, "timeout"); err == nil {
		timeout = t
	}
	workingDir := OptionalStringParam(params, "working_dir", "")
	workingDir = ResolvePath(ctx, workingDir)

	// Create command with timeout
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	cmd.Dir = workingDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start).Seconds()

	result := &ToolResult{
		Output:   stdout.String(),
		Duration: duration,
	}

	if stderr.Len() > 0 {
		if result.Output != "" {
			result.Output += "\n--- stderr ---\n" + stderr.String()
		} else {
			result.Output = stderr.String()
		}
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Error = fmt.Sprintf("command exited with code %d", exitErr.ExitCode())
		} else if ctx.Err() == context.DeadlineExceeded {
			result.Error = fmt.Sprintf("command timed out after %ds", timeout)
			result.ExitCode = -1
		} else {
			result.Error = err.Error()
			result.ExitCode = 1
		}
	}

	return result, nil
}

// isDangerousCommand checks if a command is potentially dangerous
func isDangerousCommand(cmd string) bool {
	dangerous := []string{
		"rm -rf", "rm -r /", "mkfs", "dd if=", "format",
		"sudo rm", "chmod 777", "chown", "> /dev/",
		":(){ :|:& };:", "shutdown", "reboot", "init 0",
	}

	lower := strings.ToLower(strings.TrimSpace(cmd))
	for _, pattern := range dangerous {
		if strings.Contains(lower, strings.ToLower(pattern)) {
			return true
		}
	}

	// Check for pipe to shell (potential code injection)
	if strings.Contains(cmd, "| sh") || strings.Contains(cmd, "| bash") ||
		strings.Contains(cmd, "|exec") {
		return true
	}

	return false
}
