package tools

import (
	"context"
	"fmt"
	"sync"

	"github.com/mimo-cli/mimo-cli/internal/backup"
	"github.com/mimo-cli/mimo-cli/internal/ignore"
)

// Registry manages all available tools
type Registry struct {
	mu    sync.RWMutex
	tools map[string]BaseTool
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]BaseTool),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool BaseTool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name()] = tool
}

// Get returns a tool by name
func (r *Registry) Get(name string) (BaseTool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, ok := r.tools[name]
	return tool, ok
}

// List returns all registered tool names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// Definitions returns LLM tool definitions for all registered tools
func (r *Registry) Definitions() []map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	defs := make([]map[string]interface{}, 0, len(r.tools))
	for _, tool := range r.tools {
		defs = append(defs, ToolDefinition(tool))
	}
	return defs
}

// Execute runs a tool by name with the given parameters
func (r *Registry) Execute(ctx context.Context, name string, params map[string]interface{}) (*ToolResult, error) {
	tool, ok := r.Get(name)
	if !ok {
		return ToolError("tool %q not found", name), nil
	}

	// Validate parameters
	if err := tool.Validate(params); err != nil {
		return ToolError("invalid parameters: %v", err), nil
	}

	return tool.Execute(ctx, params)
}

// DefaultRegistry creates a registry with all built-in tools
func DefaultRegistry(ign *ignore.Matcher, bm *backup.Manager) *Registry {
	r := NewRegistry()
	r.Register(NewShellTool())
	r.Register(NewFileReadTool())
	r.Register(NewFileWriteTool(bm))
	r.Register(NewFileEditTool(bm))
	r.Register(NewDirListTool(ign))
	r.Register(NewSearchTool(ign))
	r.Register(NewGlobTool(ign))
	r.Register(NewGitStatusTool())
	r.Register(NewGitDiffTool())
	r.Register(NewGitLogTool())
	r.Register(NewGitCommitTool())
	r.Register(NewGitBranchTool())
	r.Register(NewGitCheckoutTool())
	r.Register(NewGitMergeTool())
	r.Register(NewWebFetchTool())
	// Phase 3 tools
	r.Register(NewFileDiffTool())
	r.Register(NewClipboardTool())
	r.Register(NewProcessTool())
	r.Register(NewEnvTool())
	r.Register(NewDependencyTool())
	r.Register(NewHTTPRequestTool())
	r.Register(NewWebSearchTool())
	r.Register(NewFileDeleteTool())
	r.Register(NewDirCreateTool())
	r.Register(NewJSONQueryTool())
	return r
}

// StringParam extracts a string parameter from params
func StringParam(params map[string]interface{}, key string) (string, error) {
	val, ok := params[key]
	if !ok {
		return "", fmt.Errorf("missing required parameter: %s", key)
	}
	s, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("parameter %s must be a string", key)
	}
	return s, nil
}

// OptionalStringParam extracts an optional string parameter
func OptionalStringParam(params map[string]interface{}, key string, defaultVal string) string {
	val, ok := params[key]
	if !ok {
		return defaultVal
	}
	s, ok := val.(string)
	if !ok {
		return defaultVal
	}
	return s
}

// IntParam extracts an integer parameter
func IntParam(params map[string]interface{}, key string) (int, error) {
	val, ok := params[key]
	if !ok {
		return 0, fmt.Errorf("missing required parameter: %s", key)
	}
	switch v := val.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("parameter %s must be an integer", key)
	}
}
