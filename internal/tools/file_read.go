package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileReadTool reads file contents
type FileReadTool struct{}

func NewFileReadTool() *FileReadTool {
	return &FileReadTool{}
}

func (t *FileReadTool) Name() string { return "file_read" }

func (t *FileReadTool) Description() string {
	return "Read the contents of a file at the given path"
}

func (t *FileReadTool) GetSafetyLevel() SafetyLevel { return SafetyLow }

func (t *FileReadTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to read",
			},
			"offset": map[string]interface{}{
				"type":        "integer",
				"description": "Line number to start reading from (0-based)",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of lines to read",
			},
		},
		"required": []string{"path"},
	}
}

func (t *FileReadTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "path")
	return err
}

func (t *FileReadTool) RequiresConfirmation(params map[string]interface{}) bool {
	return false
}

func (t *FileReadTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	path, _ := StringParam(params, "path")

	// Resolve relative path
	if !filepath.IsAbs(path) {
		wd, err := os.Getwd()
		if err != nil {
			return ToolError("failed to get working directory: %v", err), nil
		}
		path = filepath.Join(wd, path)
	}

	// Check if file exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ToolError("file not found: %s", path), nil
		}
		return ToolError("failed to access file: %v", err), nil
	}

	if info.IsDir() {
		return ToolError("path is a directory, not a file: %s", path), nil
	}

	// Check file size (max 1MB)
	if info.Size() > 1024*1024 {
		return ToolError("file too large (%d bytes). Use offset/limit to read portions.", info.Size()), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return ToolError("failed to read file: %v", err), nil
	}

	content := string(data)

	// Track the original offset for correct line numbering
	lineOffset := 0

	// Apply offset/limit if specified
	if offset, err := IntParam(params, "offset"); err == nil {
		lines := strings.Split(content, "\n")
		if offset >= len(lines) {
			return &ToolResult{Output: "(empty: offset beyond file end)"}, nil
		}
		lineOffset = offset
		lines = lines[offset:]
		content = strings.Join(lines, "\n")
	}

	if limit, err := IntParam(params, "limit"); err == nil {
		lines := strings.Split(content, "\n")
		if limit < len(lines) {
			lines = lines[:limit]
			content = strings.Join(lines, "\n") + "\n... (truncated)"
		}
	}

	// Add line numbers
	lines := strings.Split(content, "\n")
	var numbered []string
	for i, line := range lines {
		numbered = append(numbered, fmt.Sprintf("%4d\t%s", lineOffset+i+1, line))
	}

	return &ToolResult{Output: strings.Join(numbered, "\n")}, nil
}
