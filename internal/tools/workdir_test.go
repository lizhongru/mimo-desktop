package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDirListUsesWorkingDirFromContext(t *testing.T) {
	processDir := t.TempDir()
	selectedDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(processDir, "process-only.txt"), []byte("wrong"), 0644); err != nil {
		t.Fatalf("write process file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(selectedDir, "selected-only.txt"), []byte("right"), 0644); err != nil {
		t.Fatalf("write selected file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalDir)
	})
	if err := os.Chdir(processDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	result, err := NewDirListTool(nil).Execute(WithWorkingDir(context.Background(), selectedDir), nil)
	if err != nil {
		t.Fatalf("execute dir_list: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("dir_list returned error: %s", result.Error)
	}
	if !strings.Contains(result.Output, "selected-only.txt") {
		t.Fatalf("dir_list output = %q, want selected workspace file", result.Output)
	}
	if strings.Contains(result.Output, "process-only.txt") {
		t.Fatalf("dir_list output = %q, should not include process cwd file", result.Output)
	}
}

func TestFileReadUsesWorkingDirFromContext(t *testing.T) {
	processDir := t.TempDir()
	selectedDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(processDir, "note.txt"), []byte("process cwd"), 0644); err != nil {
		t.Fatalf("write process file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(selectedDir, "note.txt"), []byte("selected workspace"), 0644); err != nil {
		t.Fatalf("write selected file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalDir)
	})
	if err := os.Chdir(processDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	result, err := NewFileReadTool().Execute(WithWorkingDir(context.Background(), selectedDir), map[string]interface{}{
		"path": "note.txt",
	})
	if err != nil {
		t.Fatalf("execute file_read: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("file_read returned error: %s", result.Error)
	}
	if !strings.Contains(result.Output, "selected workspace") {
		t.Fatalf("file_read output = %q, want selected workspace content", result.Output)
	}
	if strings.Contains(result.Output, "process cwd") {
		t.Fatalf("file_read output = %q, should not include process cwd content", result.Output)
	}
}
