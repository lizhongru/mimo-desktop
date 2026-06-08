package tools

import (
	"context"
	"fmt"
)

// SafetyLevel defines the safety level of a tool
type SafetyLevel string

const (
	SafetyLow    SafetyLevel = "LOW"
	SafetyMedium SafetyLevel = "MEDIUM"
	SafetyHigh   SafetyLevel = "HIGH"
	SafetyCritical SafetyLevel = "CRITICAL"
)

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Output   string `json:"output"`
	Error    string `json:"error,omitempty"`
	ExitCode int    `json:"exit_code,omitempty"`
	Duration float64 `json:"duration,omitempty"`
}

// BaseTool defines the interface that all tools must implement
type BaseTool interface {
	// Name returns the tool name
	Name() string

	// Description returns a human-readable description
	Description() string

	// SafetyLevel returns the safety level of this tool
	GetSafetyLevel() SafetyLevel

	// Parameters returns the JSON Schema for tool parameters
	Parameters() map[string]interface{}

	// Execute runs the tool with the given parameters
	Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error)

	// Validate checks if the parameters are valid
	Validate(params map[string]interface{}) error

	// RequiresConfirmation returns true if the tool requires user confirmation
	RequiresConfirmation(params map[string]interface{}) bool
}

// ToolDefinition converts a tool to an LLM tool definition
func ToolDefinition(tool BaseTool) map[string]interface{} {
	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        tool.Name(),
			"description": tool.Description(),
			"parameters":  tool.Parameters(),
		},
	}
}

// ToolFunc is a convenience function to create a simple tool result
func ToolFunc(output string, err error) *ToolResult {
	if err != nil {
		return &ToolResult{Error: err.Error(), ExitCode: 1}
	}
	return &ToolResult{Output: output}
}

// ToolError creates a tool result with an error
func ToolError(format string, args ...interface{}) *ToolResult {
	return &ToolResult{
		Error:    fmt.Sprintf(format, args...),
		ExitCode: 1,
	}
}
