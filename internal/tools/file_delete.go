package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// FileDeleteTool deletes files (moves to recycle bin on Windows)
type FileDeleteTool struct{}

func NewFileDeleteTool() *FileDeleteTool { return &FileDeleteTool{} }

func (t *FileDeleteTool) Name() string                { return "file_delete" }
func (t *FileDeleteTool) Description() string         { return "Delete a file or empty directory" }
func (t *FileDeleteTool) GetSafetyLevel() SafetyLevel { return SafetyHigh }
func (t *FileDeleteTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "File or directory path to delete",
			},
		},
		"required": []string{"path"},
	}
}
func (t *FileDeleteTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "path")
	return err
}
func (t *FileDeleteTool) RequiresConfirmation(params map[string]interface{}) bool { return true }

func (t *FileDeleteTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	path, _ := StringParam(params, "path")
	path = ResolvePath(ctx, path)

	// Check existence
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return ToolError("path does not exist: %s", path), nil
	}
	if err != nil {
		return ToolError("cannot access path: %v", err), nil
	}

	// Safety: don't delete critical directories
	absPath, _ := filepath.Abs(path)
	base := filepath.Base(absPath)
	if base == ".git" || base == ".mimo" || base == "node_modules" {
		return ToolError("refusing to delete protected directory: %s", base), nil
	}

	if runtime.GOOS == "windows" {
		// On Windows, try to move to recycle bin using PowerShell
		// Fall back to os.Remove if not available
		if info.IsDir() {
			err = os.Remove(path)
		} else {
			err = os.Remove(path)
		}
	} else {
		if info.IsDir() {
			err = os.Remove(path)
		} else {
			err = os.Remove(path)
		}
	}

	if err != nil {
		return ToolError("delete failed: %v", err), nil
	}

	kind := "file"
	if info.IsDir() {
		kind = "directory"
	}
	return &ToolResult{Output: fmt.Sprintf("Deleted %s: %s", kind, path)}, nil
}

// DirCreateTool creates directories
type DirCreateTool struct{}

func NewDirCreateTool() *DirCreateTool { return &DirCreateTool{} }

func (t *DirCreateTool) Name() string { return "dir_create" }
func (t *DirCreateTool) Description() string {
	return "Create a directory (including parent directories)"
}
func (t *DirCreateTool) GetSafetyLevel() SafetyLevel { return SafetyLow }
func (t *DirCreateTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Directory path to create",
			},
		},
		"required": []string{"path"},
	}
}
func (t *DirCreateTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "path")
	return err
}
func (t *DirCreateTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *DirCreateTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	path, _ := StringParam(params, "path")
	path = ResolvePath(ctx, path)

	if _, err := os.Stat(path); err == nil {
		return &ToolResult{Output: fmt.Sprintf("Directory already exists: %s", path)}, nil
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		return ToolError("failed to create directory: %v", err), nil
	}

	return &ToolResult{Output: fmt.Sprintf("Created directory: %s", path)}, nil
}
