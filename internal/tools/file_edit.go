package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mimo-cli/mimo-cli/internal/backup"
)

// FileEditTool performs line-level edits on files
type FileEditTool struct {
	bm *backup.Manager
}

func NewFileEditTool(bm *backup.Manager) *FileEditTool {
	return &FileEditTool{bm: bm}
}

func (t *FileEditTool) Name() string { return "file_edit" }

func (t *FileEditTool) Description() string {
	return "Edit a file by replacing specific text with new text (exact string match)"
}

func (t *FileEditTool) GetSafetyLevel() SafetyLevel { return SafetyMedium }

func (t *FileEditTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to edit",
			},
			"old_text": map[string]interface{}{
				"type":        "string",
				"description": "The exact text to find and replace (must be unique in the file)",
			},
			"new_text": map[string]interface{}{
				"type":        "string",
				"description": "The replacement text",
			},
			"replaceAll": map[string]interface{}{
				"type":        "boolean",
				"description": "Replace all occurrences (default: false, requires unique match)",
			},
		},
		"required": []string{"path", "old_text", "new_text"},
	}
}

func (t *FileEditTool) Validate(params map[string]interface{}) error {
	if _, err := StringParam(params, "path"); err != nil {
		return err
	}
	if _, err := StringParam(params, "old_text"); err != nil {
		return err
	}
	if _, err := StringParam(params, "new_text"); err != nil {
		return err
	}
	return nil
}

func (t *FileEditTool) RequiresConfirmation(params map[string]interface{}) bool {
	path, _ := StringParam(params, "path")
	return isProtectedFile(path)
}

func (t *FileEditTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	path, _ := StringParam(params, "path")
	oldText, _ := StringParam(params, "old_text")
	newText, _ := StringParam(params, "new_text")
	replaceAll := false
	if v, ok := params["replaceAll"].(bool); ok {
		replaceAll = v
	}

	// Resolve relative path
	if !filepath.IsAbs(path) {
		wd, err := os.Getwd()
		if err != nil {
			return ToolError("failed to get working directory: %v", err), nil
		}
		path = filepath.Join(wd, path)
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ToolError("file not found: %s", path), nil
		}
		return ToolError("failed to read file: %v", err), nil
	}

	content := string(data)

	// Check if oldText exists
	if !strings.Contains(content, oldText) {
		return ToolError("text not found in file: %s", oldText), nil
	}

	// Check uniqueness if not replaceAll
	if !replaceAll {
		count := strings.Count(content, oldText)
		if count > 1 {
			return ToolError("text appears %d times in file. Use replaceAll=true or provide more unique text", count), nil
		}
	}

	// Backup before edit
	if t.bm != nil {
		t.bm.Backup(path, "file_edit")
	}

	// Perform replacement
	var result string
	var replaced int
	if replaceAll {
		result = strings.ReplaceAll(content, oldText, newText)
		replaced = strings.Count(content, oldText)
	} else {
		result = strings.Replace(content, oldText, newText, 1)
		replaced = 1
	}

	// Write back
	if err := os.WriteFile(path, []byte(result), 0644); err != nil {
		return ToolError("failed to write file: %v", err), nil
	}

	return &ToolResult{
		Output: fmt.Sprintf("Replaced %d occurrence(s) in %s", replaced, path),
	}, nil
}
