package tools

import (
	"context"
	"os"
	"path/filepath"

	"github.com/mimo-cli/mimo-cli/internal/backup"
)

// FileWriteTool writes content to a file
type FileWriteTool struct {
	bm *backup.Manager
}

func NewFileWriteTool(bm *backup.Manager) *FileWriteTool {
	return &FileWriteTool{bm: bm}
}

func (t *FileWriteTool) Name() string { return "file_write" }

func (t *FileWriteTool) Description() string {
	return "Create or overwrite a file with the given content"
}

func (t *FileWriteTool) GetSafetyLevel() SafetyLevel { return SafetyMedium }

func (t *FileWriteTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Path to the file to write",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Content to write to the file",
			},
			"create_dirs": map[string]interface{}{
				"type":        "boolean",
				"description": "Create parent directories if they don't exist",
			},
		},
		"required": []string{"path", "content"},
	}
}

func (t *FileWriteTool) Validate(params map[string]interface{}) error {
	if _, err := StringParam(params, "path"); err != nil {
		return err
	}
	if _, err := StringParam(params, "content"); err != nil {
		return err
	}
	return nil
}

func (t *FileWriteTool) RequiresConfirmation(params map[string]interface{}) bool {
	path, _ := StringParam(params, "path")
	return isProtectedFile(path)
}

func (t *FileWriteTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	path, _ := StringParam(params, "path")
	content, _ := StringParam(params, "content")

	// Resolve relative path
	if !filepath.IsAbs(path) {
		wd, err := os.Getwd()
		if err != nil {
			return ToolError("failed to get working directory: %v", err), nil
		}
		path = filepath.Join(wd, path)
	}

	// Create directories if requested
	if createDirs, ok := params["create_dirs"].(bool); ok && createDirs {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return ToolError("failed to create directories: %v", err), nil
		}
	}

	// Backup before write
	if t.bm != nil {
		t.bm.Backup(path, "file_write")
	}

	// Write file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return ToolError("failed to write file: %v", err), nil
	}

	return &ToolResult{Output: "File written successfully: " + path}, nil
}

// isProtectedFile checks if a file is protected
func isProtectedFile(path string) bool {
	base := filepath.Base(path)
	protected := []string{
		".env", ".env.local", ".env.production",
		"id_rsa", "id_ed25519",
	}
	for _, p := range protected {
		if base == p {
			return true
		}
	}
	// Check extensions
	ext := filepath.Ext(path)
	switch ext {
	case ".pem", ".key", ".cert", ".crt":
		return true
	}
	return false
}
