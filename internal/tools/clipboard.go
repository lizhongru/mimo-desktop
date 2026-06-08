package tools

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// ClipboardTool reads from or writes to the system clipboard
type ClipboardTool struct{}

func NewClipboardTool() *ClipboardTool { return &ClipboardTool{} }

func (t *ClipboardTool) Name() string        { return "clipboard" }
func (t *ClipboardTool) Description() string  { return "Read from or write to the system clipboard" }
func (t *ClipboardTool) GetSafetyLevel() SafetyLevel { return SafetyLow }
func (t *ClipboardTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Action: read or write",
				"enum":        []string{"read", "write"},
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Content to write (required for write action)",
			},
		},
		"required": []string{"action"},
	}
}
func (t *ClipboardTool) Validate(params map[string]interface{}) error {
	action, err := StringParam(params, "action")
	if err != nil {
		return err
	}
	if action == "write" {
		_, err := StringParam(params, "content")
		return err
	}
	if action != "read" {
		return fmt.Errorf("action must be 'read' or 'write'")
	}
	return nil
}
func (t *ClipboardTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *ClipboardTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	action, _ := StringParam(params, "action")

	if action == "read" {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = exec.CommandContext(ctx, "powershell", "-Command", "Get-Clipboard")
		case "darwin":
			cmd = exec.CommandContext(ctx, "pbpaste")
		default:
			cmd = exec.CommandContext(ctx, "xclip", "-selection", "clipboard", "-o")
		}
		out, err := cmd.Output()
		if err != nil {
			return ToolError("clipboard read failed: %v", err), nil
		}
		content := strings.TrimSpace(string(out))
		if content == "" {
			return &ToolResult{Output: "Clipboard is empty"}, nil
		}
		return &ToolResult{Output: content}, nil
	}

	// write
	content, _ := StringParam(params, "content")
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.CommandContext(ctx, "powershell", "-Command", "Set-Clipboard -Value "+content)
		cmd.Stdin = strings.NewReader(content)
		// Use stdin approach for Windows
		cmd = exec.CommandContext(ctx, "powershell", "-Command", "$input | Set-Clipboard")
		cmd.Stdin = strings.NewReader(content)
	case "darwin":
		cmd = exec.CommandContext(ctx, "pbcopy")
		cmd.Stdin = strings.NewReader(content)
	default:
		cmd = exec.CommandContext(ctx, "xclip", "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(content)
	}

	if err := cmd.Run(); err != nil {
		return ToolError("clipboard write failed: %v", err), nil
	}
	return &ToolResult{Output: fmt.Sprintf("Copied %d bytes to clipboard", len(content))}, nil
}
