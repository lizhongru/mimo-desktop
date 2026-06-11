package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mimo-cli/mimo-cli/internal/ignore"
)

// DirListTool lists directory contents
type DirListTool struct {
	ign *ignore.Matcher
}

func NewDirListTool(ign *ignore.Matcher) *DirListTool { return &DirListTool{ign: ign} }

func (t *DirListTool) Name() string                { return "dir_list" }
func (t *DirListTool) Description() string         { return "List files and directories at the given path" }
func (t *DirListTool) GetSafetyLevel() SafetyLevel { return SafetyLow }

func (t *DirListTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Directory path to list (default: current directory)",
			},
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Glob pattern to filter files (e.g. '*.go')",
			},
			"recursive": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to list recursively (default: false)",
			},
			"max_depth": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum recursion depth (default: 3)",
			},
		},
		"required": []string{},
	}
}

func (t *DirListTool) Validate(params map[string]interface{}) error { return nil }

func (t *DirListTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *DirListTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	dir := "."
	if v, err := StringParam(params, "path"); err == nil && v != "" {
		dir = v
	}

	dir = ResolvePath(ctx, dir)

	info, err := os.Stat(dir)
	if err != nil {
		return ToolError("path not found: %s", dir), nil
	}
	if !info.IsDir() {
		return ToolError("path is not a directory: %s", dir), nil
	}

	recursive := false
	if v, ok := params["recursive"]; ok {
		if b, ok := v.(bool); ok {
			recursive = b
		}
	}

	maxDepth := 3
	if v, err := IntParam(params, "max_depth"); err == nil && v > 0 {
		maxDepth = v
	}

	pattern := ""
	if v, err := StringParam(params, "pattern"); err == nil {
		pattern = v
	}

	var lines []string
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		rel, _ := filepath.Rel(dir, path)
		if rel == "." {
			return nil
		}

		depth := strings.Count(rel, string(os.PathSeparator))
		if !recursive && depth > 0 && info.IsDir() {
			return filepath.SkipDir
		}
		if depth >= maxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() && t.ign != nil && t.ign.ShouldIgnore(info.Name(), true, dir, path) {
			return filepath.SkipDir
		}

		// Skip ignored files
		if !info.IsDir() && t.ign != nil && t.ign.ShouldIgnore(info.Name(), false, dir, path) {
			return nil
		}

		if pattern != "" {
			if matched, _ := filepath.Match(pattern, info.Name()); !matched && !info.IsDir() {
				return nil
			}
		}

		indent := strings.Repeat("  ", depth)
		if info.IsDir() {
			lines = append(lines, fmt.Sprintf("%s%s/", indent, info.Name()))
		} else {
			lines = append(lines, fmt.Sprintf("%s%s", indent, info.Name()))
		}

		if len(lines) > 500 {
			return fmt.Errorf("limit reached")
		}
		return nil
	})

	if len(lines) == 0 {
		return &ToolResult{Output: "(empty directory)"}, nil
	}
	return &ToolResult{Output: strings.Join(lines, "\n")}, nil
}
