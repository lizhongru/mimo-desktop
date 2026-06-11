package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mimo-cli/mimo-cli/internal/ignore"
)

// SearchTool searches for text patterns in files
type SearchTool struct {
	ign *ignore.Matcher
}

func NewSearchTool(ign *ignore.Matcher) *SearchTool { return &SearchTool{ign: ign} }

func (t *SearchTool) Name() string                { return "search" }
func (t *SearchTool) Description() string         { return "Search for a text pattern in files" }
func (t *SearchTool) GetSafetyLevel() SafetyLevel { return SafetyLow }

func (t *SearchTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "Regex pattern to search for",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Directory to search in (default: current directory)",
			},
			"include": map[string]interface{}{
				"type":        "string",
				"description": "File glob pattern to filter (e.g. '*.go')",
			},
			"max_results": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of results (default: 50)",
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *SearchTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "pattern")
	return err
}

func (t *SearchTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *SearchTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	pattern, _ := StringParam(params, "pattern")
	re, err := regexp.Compile(pattern)
	if err != nil {
		return ToolError("invalid regex: %v", err), nil
	}

	dir := "."
	if v, err := StringParam(params, "path"); err == nil && v != "" {
		dir = v
	}
	dir = ResolvePath(ctx, dir)

	include := ""
	if v, err := StringParam(params, "include"); err == nil {
		include = v
	}

	maxResults := 50
	if v, err := IntParam(params, "max_results"); err == nil && v > 0 {
		maxResults = v
	}

	var results []string
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if t.ign != nil && t.ign.ShouldIgnore(info.Name(), true, dir, path) {
				return filepath.SkipDir
			}
			return nil
		}
		if t.ign != nil && t.ign.ShouldIgnore(info.Name(), false, dir, path) {
			return nil
		}
		if info.Size() > 1024*1024 {
			return nil
		}
		if include != "" {
			if matched, _ := filepath.Match(include, info.Name()); !matched {
				return nil
			}
		}

		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		rel, _ := filepath.Rel(dir, path)
		scanner := bufio.NewScanner(f)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			if re.MatchString(line) {
				results = append(results, fmt.Sprintf("%s:%d: %s", rel, lineNum, strings.TrimSpace(line)))
				if len(results) >= maxResults {
					return fmt.Errorf("limit reached")
				}
			}
		}
		return nil
	})

	if len(results) == 0 {
		return &ToolResult{Output: "No matches found."}, nil
	}
	return &ToolResult{Output: strings.Join(results, "\n")}, nil
}
