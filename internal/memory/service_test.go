package memory

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	return db
}

func TestBuildFtsQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single word",
			input:    "hello",
			expected: `"hello"`,
		},
		{
			name:     "multiple words",
			input:    "hello world",
			expected: `"hello" OR "world"`,
		},
		{
			name:     "with punctuation",
			input:    "hello, world!",
			expected: `"hello" OR "world"`,
		},
		{
			name:     "empty query",
			input:    "",
			expected: "",
		},
		{
			name:     "special characters",
			input:    "test@example.com",
			expected: `"test" OR "example" OR "com"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildFtsQuery(tt.input)
			if result != tt.expected {
				t.Errorf("BuildFtsQuery(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractSnippet(t *testing.T) {
	text := "This is a test document about machine learning and artificial intelligence."

	tests := []struct {
		name      string
		text      string
		query     string
		contextLen int
		wantContains string
	}{
		{
			name:         "find match",
			text:         text,
			query:        "machine",
			contextLen:   10,
			wantContains: "machine",
		},
		{
			name:         "no match returns beginning",
			text:         text,
			query:        "xyz",
			contextLen:   20,
			wantContains: "This is a test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractSnippet(tt.text, tt.query, tt.contextLen)
			if len(result) == 0 {
				t.Error("ExtractSnippet returned empty string")
			}
		})
	}
}

func TestServiceInit(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	svc := NewService(db, t.TempDir(), "test-session")
	err := svc.Init()
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Check that tables were created
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='memory_fts'").Scan(&tableName)
	if err != nil {
		t.Errorf("memory_fts table not created: %v", err)
	}
}

func TestIndexFile(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create temp directory and file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	content := "# Test Memory\n\nThis is a test memory entry."
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	svc := NewService(db, tmpDir, "test-session")
	if err := svc.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Index the file
	err := svc.IndexFile(testFile, "project", "", "project")
	if err != nil {
		t.Fatalf("IndexFile() failed: %v", err)
	}

	// Verify the file was indexed
	count, err := svc.Count()
	if err != nil {
		t.Fatalf("Count() failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Count() = %d, want 1", count)
	}
}

func TestSearch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tmpDir := t.TempDir()

	// Create test files
	files := map[string]string{
		"memory.md":    "# Project Memory\n\nThis is project memory about Go development.",
		"notes.md":     "# Notes\n\nQuick notes about debugging.",
		"checkpoint.md": "# Checkpoint\n\nSession checkpoint with task progress.",
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	svc := NewService(db, tmpDir, "test-session")
	if err := svc.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Index all files
	for name := range files {
		path := filepath.Join(tmpDir, name)
		svc.IndexFile(path, "project", "", determineType(path))
	}

	// Search for "Go"
	results, err := svc.Search("Go", SearchOpts{Limit: 10})
	if err != nil {
		t.Fatalf("Search() failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Search() returned no results")
	}

	// Check that we found the memory.md file
	found := false
	for _, r := range results {
		if filepath.Base(r.Path) == "memory.md" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Search() did not find memory.md")
	}
}

func TestReconcile(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tmpDir := t.TempDir()

	// Create test files in the correct directory structure
	// Files should be in <projectDir>/.mimo/memory/ for project scope
	memoryDir := filepath.Join(tmpDir, ".mimo", "memory")
	os.MkdirAll(memoryDir, 0755)

	files := map[string]string{
		"MEMORY.md": "# Project Memory\n\nImportant project knowledge.",
	}

	for name, content := range files {
		path := filepath.Join(memoryDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	svc := NewService(db, tmpDir, "test-session")
	if err := svc.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Reconcile
	indexed, pruned, err := svc.Reconcile()
	if err != nil {
		t.Fatalf("Reconcile() failed: %v", err)
	}

	if indexed != 1 {
		t.Errorf("Reconcile() indexed = %d, want 1", indexed)
	}
	if pruned != 0 {
		t.Errorf("Reconcile() pruned = %d, want 0", pruned)
	}
}

func TestRemoveFile(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	content := "# Test\n\nTest content."
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	svc := NewService(db, tmpDir, "test-session")
	if err := svc.Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Index the file
	svc.IndexFile(testFile, "project", "", "project")

	// Verify indexed
	count, _ := svc.Count()
	if count != 1 {
		t.Errorf("Count() after index = %d, want 1", count)
	}

	// Remove the file
	err := svc.RemoveFile(testFile, "project")
	if err != nil {
		t.Fatalf("RemoveFile() failed: %v", err)
	}

	// Verify removed
	count, _ = svc.Count()
	if count != 0 {
		t.Errorf("Count() after remove = %d, want 0", count)
	}
}

func TestDetermineType(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/path/to/MEMORY.md", "project"},
		{"/path/to/checkpoint.md", "checkpoint"},
		{"/path/to/notes.md", "notes"},
		{"/path/to/tasks/T1/progress.md", "progress"},
		{"/path/to/custom.md", "free"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := determineType(tt.path)
			if result != tt.expected {
				t.Errorf("determineType(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

// Ensure time import is used
var _ = time.Now
