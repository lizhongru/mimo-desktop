package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/mimo-cli/mimo-cli/internal/tools"
)

// MCPToolAdapter wraps an MCP tool to implement the tools.BaseTool interface
type MCPToolAdapter struct {
	client     *Client
	tool       Tool
	serverName string
	fullName   string
}

// NewMCPToolAdapter creates a new MCP tool adapter
func NewMCPToolAdapter(client *Client, tool Tool, serverName string) *MCPToolAdapter {
	// Prefix tool name with server name to avoid conflicts: server__toolname
	fullName := serverName + "__" + tool.Name
	return &MCPToolAdapter{
		client:     client,
		tool:       tool,
		serverName: serverName,
		fullName:   fullName,
	}
}

// Name returns the tool name (prefixed with server name)
func (a *MCPToolAdapter) Name() string {
	return a.fullName
}

// Description returns the tool description
func (a *MCPToolAdapter) Description() string {
	desc := a.tool.Description
	if desc == "" {
		desc = fmt.Sprintf("MCP tool %s from server %s", a.tool.Name, a.serverName)
	}
	return fmt.Sprintf("[MCP:%s] %s", a.serverName, desc)
}

// GetSafetyLevel returns the safety level (MCP tools default to MEDIUM)
func (a *MCPToolAdapter) GetSafetyLevel() tools.SafetyLevel {
	return tools.SafetyMedium
}

// Parameters returns the JSON Schema for the tool parameters
func (a *MCPToolAdapter) Parameters() map[string]interface{} {
	if a.tool.InputSchema != nil {
		return a.tool.InputSchema
	}
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

// Execute calls the MCP tool
func (a *MCPToolAdapter) Execute(ctx context.Context, params map[string]interface{}) (*tools.ToolResult, error) {
	result, err := a.client.CallTool(a.tool.Name, params)
	if err != nil {
		return &tools.ToolResult{
			Error:    err.Error(),
			ExitCode: 1,
		}, nil
	}

	// Extract text content from result
	var output strings.Builder
	for _, content := range result.Content {
		switch content.Type {
		case "text":
			output.WriteString(content.Text)
			output.WriteString("\n")
		default:
			output.WriteString(fmt.Sprintf("[%s content]\n", content.Type))
		}
	}

	outputStr := strings.TrimSpace(output.String())

	if result.IsError {
		return &tools.ToolResult{
			Error:    outputStr,
			ExitCode: 1,
		}, nil
	}

	return &tools.ToolResult{
		Output: outputStr,
	}, nil
}

// Validate checks if the parameters are valid
func (a *MCPToolAdapter) Validate(params map[string]interface{}) error {
	// Basic validation - MCP tools handle their own validation
	return nil
}

// RequiresConfirmation returns false (safety is handled by guardrail)
func (a *MCPToolAdapter) RequiresConfirmation(params map[string]interface{}) bool {
	return false
}
