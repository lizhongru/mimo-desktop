package desktop

import (
	"os"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// SelectDirectory opens a directory picker dialog and returns the selected path.
func (a *App) SelectDirectory() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Workspace Directory",
	})
	if err != nil {
		return "", err
	}
	return dir, nil
}

// GetWorkingDir returns the current working directory.
func (a *App) GetWorkingDir() string {
	wd, _ := os.Getwd()
	return wd
}

// ToolInfo describes a registered tool.
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	SafetyLevel string `json:"safetyLevel"`
	IsMCP       bool   `json:"isMcp"`
	ServerName  string `json:"serverName,omitempty"`
}

// GetTools returns all registered tools (built-in + MCP).
func (a *App) GetTools() []ToolInfo {
	names := a.registry.List()
	result := make([]ToolInfo, 0, len(names))
	for _, name := range names {
		tool, ok := a.registry.Get(name)
		if !ok {
			continue
		}
		isMcp := false
		serverName := ""
		// MCP tools use "servername__toolname" format — check with SplitN
		// to avoid false positives from regular tools containing "__".
		if idx := strings.Index(name, "__"); idx > 0 && idx < len(name)-2 {
			isMcp = true
			serverName = name[:idx]
		}
		result = append(result, ToolInfo{
			Name:        tool.Name(),
			Description: tool.Description(),
			SafetyLevel: string(tool.GetSafetyLevel()),
			IsMCP:       isMcp,
			ServerName:  serverName,
		})
	}
	return result
}

// MCPServerInfo describes an MCP server.
type MCPServerInfo struct {
	Name      string   `json:"name"`
	Connected bool     `json:"connected"`
	ToolCount int      `json:"toolCount"`
	Tools     []string `json:"tools"`
}

// GetMCPServers returns all configured MCP servers.
func (a *App) GetMCPServers() []MCPServerInfo {
	if a.mcpManager == nil {
		return nil
	}
	names := a.mcpManager.ServerNames()
	result := make([]MCPServerInfo, len(names))
	for i, name := range names {
		client, ok := a.mcpManager.GetClient(name)
		connected := ok && client.IsConnected()
		// Count tools for this server
		toolNames := []string{}
		allTools := a.registry.List()
		for _, tn := range allTools {
			if len(tn) > len(name)+2 && tn[:len(name)+2] == name+"__" {
				toolNames = append(toolNames, tn[len(name)+2:])
			}
		}
		result[i] = MCPServerInfo{
			Name:      name,
			Connected: connected,
			ToolCount: len(toolNames),
			Tools:     toolNames,
		}
	}
	return result
}
