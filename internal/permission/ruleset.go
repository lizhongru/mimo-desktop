package permission

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PermissionAction represents the action for a permission rule
type PermissionAction string

const (
	Allow PermissionAction = "allow"
	Deny  PermissionAction = "deny"
	Ask   PermissionAction = "ask"
)

// PermissionRule defines a permission rule
type PermissionRule struct {
	Permission string           `json:"permission"` // "read", "write", "edit", "bash", "external_directory"
	Action     PermissionAction `json:"action"`
	Pattern    string           `json:"pattern,omitempty"` // glob pattern
}

// Ruleset is a collection of permission rules
type Ruleset []PermissionRule

// DefaultRuleset returns the default permission rules
func DefaultRuleset() Ruleset {
	return Ruleset{
		{Permission: "read", Action: Allow},
		{Permission: "write", Action: Ask},
		{Permission: "edit", Action: Ask},
		{Permission: "bash", Action: Ask},
		{Permission: "external_directory", Action: Deny},
	}
}

// Evaluate evaluates whether a tool action is allowed
func (r Ruleset) Evaluate(tool string, params map[string]interface{}) PermissionAction {
	// Find matching rule
	for _, rule := range r {
		if rule.Permission == tool || rule.Permission == "*" {
			if rule.Pattern != "" {
				if matchPattern(rule.Pattern, params) {
					return rule.Action
				}
			} else {
				return rule.Action
			}
		}
	}
	// Default: ask
	return Ask
}

// Merge merges two rulesets, with the second taking precedence
func Merge(a, b Ruleset) Ruleset {
	result := make(Ruleset, 0, len(a)+len(b))
	result = append(result, a...)
	result = append(result, b...)
	return result
}

// matchPattern checks if params match a pattern
func matchPattern(pattern string, params map[string]interface{}) bool {
	// Simple pattern matching for file paths
	if path, ok := params["path"].(string); ok {
		return matchPath(pattern, path)
	}
	if filePath, ok := params["file_path"].(string); ok {
		return matchPath(pattern, filePath)
	}
	return false
}

// matchPath matches a file path against a pattern
func matchPath(pattern, path string) bool {
	// Convert glob pattern to simple prefix/suffix matching
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}
	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(path, suffix)
	}
	return pattern == path
}

// LoadRuleset loads a ruleset from a JSON file
func LoadRuleset(path string) (Ruleset, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read ruleset: %w", err)
	}

	var ruleset Ruleset
	if err := json.Unmarshal(data, &ruleset); err != nil {
		return nil, fmt.Errorf("failed to parse ruleset: %w", err)
	}

	return ruleset, nil
}

// SaveRuleset saves a ruleset to a JSON file
func SaveRuleset(path string, ruleset Ruleset) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(ruleset, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal ruleset: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write ruleset: %w", err)
	}

	return nil
}
