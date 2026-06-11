package task

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	// Create tables
	db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			parent_task_id TEXT,
			status TEXT NOT NULL DEFAULT 'open',
			summary TEXT NOT NULL,
			owner TEXT,
			created_at INTEGER NOT NULL,
			last_event_at INTEGER NOT NULL,
			ended_at INTEGER,
			cleanup_after INTEGER
		)
	`)
	db.Exec(`
		CREATE TABLE IF NOT EXISTS task_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id TEXT NOT NULL,
			at INTEGER NOT NULL,
			kind TEXT NOT NULL,
			summary TEXT
		)
	`)

	return db
}

func TestCreateTask(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	registry := NewRegistry(db)

	task, err := registry.Create("session1", "Test task", nil)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	if task.ID == "" {
		t.Error("task ID should not be empty")
	}
	if task.Summary != "Test task" {
		t.Errorf("expected summary 'Test task', got '%s'", task.Summary)
	}
	if task.Status != TaskOpen {
		t.Errorf("expected status 'open', got '%s'", task.Status)
	}
}

func TestCreateChildTask(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	registry := NewRegistry(db)

	parent, err := registry.Create("session1", "Parent task", nil)
	if err != nil {
		t.Fatalf("failed to create parent task: %v", err)
	}

	child, err := registry.Create("session1", "Child task", &parent.ID)
	if err != nil {
		t.Fatalf("failed to create child task: %v", err)
	}

	if child.ParentTaskID == nil || *child.ParentTaskID != parent.ID {
		t.Error("child task should have parent ID")
	}
}

func TestTaskStatusTransitions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	registry := NewRegistry(db)

	task, _ := registry.Create("session1", "Test task", nil)

	// Start
	task, err := registry.Start(task.ID, "user", "Starting")
	if err != nil {
		t.Fatalf("failed to start task: %v", err)
	}
	if task.Status != TaskInProgress {
		t.Errorf("expected status 'in_progress', got '%s'", task.Status)
	}
	if task.Owner == nil || *task.Owner != "user" {
		t.Error("owner should be set")
	}

	// Block
	task, err = registry.Block(task.ID, "Blocked")
	if err != nil {
		t.Fatalf("failed to block task: %v", err)
	}
	if task.Status != TaskBlocked {
		t.Errorf("expected status 'blocked', got '%s'", task.Status)
	}

	// Done
	task, err = registry.Done(task.ID, "Completed")
	if err != nil {
		t.Fatalf("failed to complete task: %v", err)
	}
	if task.Status != TaskDone {
		t.Errorf("expected status 'done', got '%s'", task.Status)
	}
	if task.EndedAt == nil {
		t.Error("ended_at should be set")
	}
}

func TestListTasks(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	registry := NewRegistry(db)

	registry.Create("session1", "Task 1", nil)
	registry.Create("session1", "Task 2", nil)
	registry.Create("session2", "Task 3", nil)

	tasks, err := registry.List("session1", nil, false)
	if err != nil {
		t.Fatalf("failed to list tasks: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestDeleteTask(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	registry := NewRegistry(db)

	task, _ := registry.Create("session1", "Test task", nil)
	err := registry.Delete(task.ID)
	if err != nil {
		t.Fatalf("failed to delete task: %v", err)
	}

	_, err = registry.Get(task.ID)
	if err == nil {
		t.Error("expected error when getting deleted task")
	}
}
