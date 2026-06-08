package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// JSONQueryTool reads and queries JSON/YAML files
type JSONQueryTool struct{}

func NewJSONQueryTool() *JSONQueryTool { return &JSONQueryTool{} }

func (t *JSONQueryTool) Name() string        { return "json_query" }
func (t *JSONQueryTool) Description() string  { return "Read and query JSON/YAML files using dot-notation paths" }
func (t *JSONQueryTool) GetSafetyLevel() SafetyLevel { return SafetyLow }
func (t *JSONQueryTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "File path (JSON or YAML)",
			},
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Dot-notation query path (e.g. 'name' or 'dependencies.react.version'). Empty to show full content.",
			},
		},
		"required": []string{"path"},
	}
}
func (t *JSONQueryTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "path")
	return err
}
func (t *JSONQueryTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *JSONQueryTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	path, _ := StringParam(params, "path")
	query, _ := StringParam(params, "query")

	data, err := os.ReadFile(path)
	if err != nil {
		return ToolError("cannot read file: %v", err), nil
	}

	// Parse as JSON or YAML based on file extension
	var parsed interface{}
	ext := strings.ToLower(path)
	if strings.HasSuffix(ext, ".yaml") || strings.HasSuffix(ext, ".yml") {
		if err := yaml.Unmarshal(data, &parsed); err != nil {
			return ToolError("YAML parse error: %v", err), nil
		}
	} else {
		if err := json.Unmarshal(data, &parsed); err != nil {
			return ToolError("JSON parse error: %v", err), nil
		}
	}

	if query == "" {
		// Return full content formatted
		out, _ := json.MarshalIndent(parsed, "", "  ")
		result := string(out)
		if len(result) > 6000 {
			result = result[:6000] + "\n... (truncated)"
		}
		return &ToolResult{Output: result}, nil
	}

	// Navigate dot-notation path
	result := navigatePath(parsed, query)
	if result == nil {
		return &ToolResult{Output: fmt.Sprintf("Path not found: %s", query)}, nil
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return &ToolResult{Output: string(out)}, nil
}

func navigatePath(data interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := data
	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[part]
			if !ok {
				return nil
			}
			current = val
		case map[interface{}]interface{}:
			// YAML sometimes produces this
			val, ok := v[part]
			if !ok {
				return nil
			}
			current = val
		case []interface{}:
			// Try to use part as array index
			idx := 0
			for _, c := range part {
				if c >= '0' && c <= '9' {
					idx = idx*10 + int(c-'0')
				} else {
					return nil
				}
			}
			if idx < 0 || idx >= len(v) {
				return nil
			}
			current = v[idx]
		default:
			return nil
		}
	}
	return current
}
