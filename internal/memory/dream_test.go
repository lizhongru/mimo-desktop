package memory

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDreamExtract(t *testing.T) {
	dream := NewDream(DefaultDreamConfig(), t.TempDir())

	sessionData := `
		# Session History

		We decided to use PostgreSQL for the database.

		I fixed the bug in the login flow.

		The user prefers dark mode.

		Note: Always use error handling.
	`

	entries, err := dream.Extract(sessionData)
	if err != nil {
		t.Fatalf("failed to extract entries: %v", err)
	}

	if len(entries) == 0 {
		t.Error("expected some entries to be extracted")
	}

	// Check for different types
	types := make(map[string]bool)
	for _, entry := range entries {
		types[entry.Type] = true
	}

	if !types["decision"] {
		t.Error("expected decision entries")
	}
}

func TestDreamSaveToMemory(t *testing.T) {
	tmpDir := t.TempDir()
	dream := NewDream(DefaultDreamConfig(), tmpDir)

	entries := []DreamEntry{
		{Type: "decision", Summary: "Use Go for backend"},
		{Type: "preference", Summary: "Prefer tabs over spaces"},
	}

	err := dream.SaveToMemory(entries)
	if err != nil {
		t.Fatalf("failed to save to memory: %v", err)
	}

	// Check if file was created
	memoryFile := filepath.Join(tmpDir, ".mimo", "memory", "MEMORY.md")
	if _, err := os.Stat(memoryFile); os.IsNotExist(err) {
		t.Error("MEMORY.md should exist")
	}
}

func TestDreamDisabled(t *testing.T) {
	config := DefaultDreamConfig()
	config.Enabled = false
	dream := NewDream(config, t.TempDir())

	entries, err := dream.Extract("test data")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entries != nil {
		t.Error("expected nil entries when disabled")
	}
}
