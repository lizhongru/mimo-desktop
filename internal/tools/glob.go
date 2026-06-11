package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/mimo-cli/mimo-cli/internal/ignore"
)

// GlobTool finds files matching a glob pattern
type GlobTool struct {
	ign *ignore.Matcher
}

func NewGlobTool(ign *ignore.Matcher) *GlobTool { return &GlobTool{ign: ign} }

func (t *GlobTool) Name() string                { return "glob" }
func (t *GlobTool) Description() string         { return "Find files matching a glob pattern" }
func (t *GlobTool) GetSafetyLevel() SafetyLevel { return SafetyLow }

func (t *GlobTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Glob pattern (e.g. '**/*.go', '*.txt')",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Base directory (default: current directory)",
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *GlobTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "pattern")
	return err
}

func (t *GlobTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *GlobTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	pattern, _ := StringParam(params, "pattern")

	dir := "."
	if v, err := StringParam(params, "path"); err == nil && v != "" {
		dir = v
	}
	dir = ResolvePath(ctx, dir)

	// Handle ** patterns by walking the directory
	if strings.Contains(pattern, "**") {
		return t.globWalk(dir, pattern)
	}

	// Simple glob
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return ToolError("invalid pattern: %v", err), nil
	}

	var results []string
	for _, m := range matches {
		rel, _ := filepath.Rel(dir, m)
		results = append(results, rel)
	}

	if len(results) == 0 {
		return &ToolResult{Output: "No files matched."}, nil
	}
	return &ToolResult{Output: strings.Join(results, "\n")}, nil
}

func (t *GlobTool) globWalk(baseDir, pattern string) (*ToolResult, error) {
	// Split pattern at **
	parts := strings.SplitN(pattern, "**", 2)
	afterStar := ""
	if len(parts) > 1 {
		afterStar = strings.TrimPrefix(parts[1], "/")
	}

	var results []string
	_ = filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if t.ign != nil && t.ign.ShouldIgnore(info.Name(), true, baseDir, path) {
				return filepath.SkipDir
			}
			return nil
		}
		if t.ign != nil && t.ign.ShouldIgnore(info.Name(), false, baseDir, path) {
			return nil
		}

		rel, _ := filepath.Rel(baseDir, path)

		if afterStar != "" {
			if matched, _ := filepath.Match(afterStar, info.Name()); !matched {
				// Also try matching the full relative path
				if matched, _ := filepath.Match(afterStar, rel); !matched {
					return nil
				}
			}
		}

		results = append(results, rel)
		if len(results) > 200 {
			return filepath.SkipDir
		}
		return nil
	})

	if len(results) == 0 {
		return &ToolResult{Output: "No files matched."}, nil
	}
	return &ToolResult{Output: strings.Join(results, "\n")}, nil
}
