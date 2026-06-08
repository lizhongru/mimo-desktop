package safety

import (
	"strings"

	"github.com/mimo-cli/mimo-cli/internal/tools"
)

// ActionLevel defines the danger level of an action
type ActionLevel string

const (
	ActionCritical ActionLevel = "CRITICAL"
	ActionHigh     ActionLevel = "HIGH"
	ActionMedium   ActionLevel = "MEDIUM"
	ActionLow      ActionLevel = "LOW"
)

// Action represents a classified action
type Action struct {
	Level       ActionLevel
	Description string
	Tool        string
	Params      map[string]interface{}
}

// Classifier classifies tool actions into safety levels
type Classifier struct {
	blockedCommands   []string
	protectedFiles    []string
	protectedBranches []string
}

// NewClassifier creates a new action classifier
func NewClassifier(blockedCmds, protectedFiles, protectedBranches []string) *Classifier {
	return &Classifier{
		blockedCommands:   blockedCmds,
		protectedFiles:    protectedFiles,
		protectedBranches: protectedBranches,
	}
}

// Classify determines the safety level of a tool action
func (c *Classifier) Classify(toolName string, params map[string]interface{}) Action {
	// Get base safety level from tool
	tool, exists := getToolSafetyLevel(toolName)
	if !exists {
		return Action{
			Level:       ActionMedium,
			Description: "Unknown tool: " + toolName,
			Tool:        toolName,
			Params:      params,
		}
	}

	action := Action{
		Tool:   toolName,
		Params: params,
	}

	switch tool {
	case tools.SafetyCritical:
		action.Level = ActionCritical
		action.Description = "Critical operation blocked"
	case tools.SafetyHigh:
		action.Level = ActionHigh
		action.Description = "High-risk operation requires confirmation"
	case tools.SafetyMedium:
		action.Level = ActionMedium
		action.Description = "Medium-risk operation"
	case tools.SafetyLow:
		action.Level = ActionLow
		action.Description = "Low-risk operation"
	}

	// Override based on specific patterns
	if toolName == "shell" {
		if cmd, ok := params["command"].(string); ok {
			if c.isBlockedCommand(cmd) {
				action.Level = ActionCritical
				action.Description = "Blocked command: " + cmd
			} else if !isDangerousShellCommand(cmd) {
				// 普通命令降级为 MEDIUM，无需确认
				action.Level = ActionMedium
				action.Description = "Shell command: " + cmd
			} else {
				// 危险命令保持 HIGH，需要确认
				action.Description = "Dangerous command: " + cmd
			}
		}
	}

	if toolName == "file_write" || toolName == "file_edit" {
		if path, ok := params["path"].(string); ok {
			if c.isProtectedFile(path) {
				action.Level = ActionHigh
				action.Description = "Protected file: " + path
			}
		}
	}

	return action
}

// isBlockedCommand checks if a command is in the blocked list
func (c *Classifier) isBlockedCommand(cmd string) bool {
	for _, blocked := range c.blockedCommands {
		if containsIgnoreCase(cmd, blocked) {
			return true
		}
	}
	return false
}

// isProtectedFile checks if a file is in the protected list
func (c *Classifier) isProtectedFile(path string) bool {
	for _, protected := range c.protectedFiles {
		if matchPattern(path, protected) {
			return true
		}
	}
	return false
}

// getToolSafetyLevel returns the safety level for a known tool
func getToolSafetyLevel(name string) (tools.SafetyLevel, bool) {
	levels := map[string]tools.SafetyLevel{
		"shell":      tools.SafetyHigh,
		"file_read":  tools.SafetyLow,
		"file_write": tools.SafetyMedium,
		"file_edit":  tools.SafetyMedium,
		"dir_list":   tools.SafetyLow,
		"dir_create": tools.SafetyLow,
		"search":     tools.SafetyLow,
		"glob":       tools.SafetyLow,
		"git":        tools.SafetyMedium,
		"git_status": tools.SafetyLow,
		"git_diff":   tools.SafetyLow,
		"git_log":    tools.SafetyLow,
		"git_commit": tools.SafetyMedium,
		"git_branch":   tools.SafetyLow,
		"git_checkout": tools.SafetyMedium,
		"git_merge":    tools.SafetyMedium,
		"web_fetch":    tools.SafetyLow,
		"docker":     tools.SafetyHigh,
		"process":    tools.SafetyHigh,
		"env":        tools.SafetyMedium,
	}
	level, ok := levels[name]
	return level, ok
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		stringsContains(stringsToLower(s), stringsToLower(substr))
}

func stringsToLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func stringsContains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func matchPattern(path, pattern string) bool {
	// Simple pattern matching: * at end means prefix match
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(path) >= len(prefix) && path[:len(prefix)] == prefix
	}
	return path == pattern
}

// isDangerousShellCommand checks if a shell command is potentially dangerous
func isDangerousShellCommand(cmd string) bool {
	dangerous := []string{
		"rm -rf", "rm -r /", "mkfs", "dd if=", "format",
		"sudo rm", "chmod 777", "chown", "> /dev/",
		":(){ :|:& };:", "shutdown", "reboot", "init 0",
	}
	lower := strings.ToLower(strings.TrimSpace(cmd))
	for _, pattern := range dangerous {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	if strings.Contains(cmd, "| sh") || strings.Contains(cmd, "| bash") ||
		strings.Contains(cmd, "|exec") {
		return true
	}
	return false
}
