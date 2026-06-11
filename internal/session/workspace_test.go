package session

import (
	"path/filepath"
	"testing"
)

func TestWorkspacePathFromIDReturnsFolderPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ProjectA")
	got := WorkspacePathFromID("ws:" + path)
	if got != path {
		t.Fatalf("workspace path = %q, want %q", got, path)
	}
}

func TestWorkspacePathFromIDIgnoresDefaultWorkspace(t *testing.T) {
	if got := WorkspacePathFromID(DefaultWorkspaceID); got != "" {
		t.Fatalf("default workspace path = %q, want empty", got)
	}
}
