package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mimo-cli/mimo-cli/internal/actor"
	"github.com/mimo-cli/mimo-cli/internal/memory"
	"github.com/mimo-cli/mimo-cli/internal/task"
	_ "modernc.org/sqlite"
)

func setupMimoToolTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE tasks (
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
		);
		CREATE TABLE task_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id TEXT NOT NULL,
			at INTEGER NOT NULL,
			kind TEXT NOT NULL,
			summary TEXT
		);
	`)
	if err != nil {
		db.Close()
		t.Fatalf("create tables: %v", err)
	}

	return db
}

func TestRegisterMimoMemoryToolsMakesMemorySearchExecutable(t *testing.T) {
	db := setupMimoToolTestDB(t)
	defer db.Close()

	projectDir := t.TempDir()
	svc := memory.NewService(db, projectDir, "session-1")
	if err := svc.Init(); err != nil {
		t.Fatalf("init memory service: %v", err)
	}

	memFile := filepath.Join(projectDir, "MEMORY.md")
	t.Cleanup(func() {})
	if err := writeTestFile(memFile, "Important architecture: use the Go tool registry for agent tools."); err != nil {
		t.Fatalf("write memory file: %v", err)
	}
	if err := svc.IndexFile(memFile, "project", "", "project"); err != nil {
		t.Fatalf("index memory file: %v", err)
	}

	registry := NewRegistry()
	registry.RegisterMemoryTools(func() *memory.Service { return svc })

	if _, ok := registry.Get("memory_search"); !ok {
		t.Fatal("memory_search should be registered")
	}
	if _, ok := registry.Get("memory_reconcile"); !ok {
		t.Fatal("memory_reconcile should be registered")
	}

	result, err := registry.Execute(context.Background(), "memory_search", map[string]interface{}{
		"query": "architecture",
		"limit": float64(3),
	})
	if err != nil {
		t.Fatalf("execute memory_search: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("memory_search returned error: %s", result.Error)
	}
	if !strings.Contains(result.Output, "MEMORY.md") {
		t.Fatalf("memory_search output should mention indexed file, got: %s", result.Output)
	}
}

func TestRegisterMimoTaskToolsUsesCurrentSession(t *testing.T) {
	db := setupMimoToolTestDB(t)
	defer db.Close()

	taskRegistry := task.NewRegistry(db)
	registry := NewRegistry()
	registry.RegisterTaskTools(func() *task.Registry { return taskRegistry }, func() string { return "session-1" })

	for _, name := range []string{"task_create", "task_list", "task_start", "task_done", "task_block"} {
		if _, ok := registry.Get(name); !ok {
			t.Fatalf("%s should be registered", name)
		}
	}

	createResult, err := registry.Execute(context.Background(), "task_create", map[string]interface{}{
		"summary": "Wire memory and task tools into the agent registry",
	})
	if err != nil {
		t.Fatalf("execute task_create: %v", err)
	}
	if createResult.Error != "" {
		t.Fatalf("task_create returned error: %s", createResult.Error)
	}

	var created task.Task
	if err := json.Unmarshal([]byte(createResult.Output), &created); err != nil {
		t.Fatalf("decode task_create output: %v\n%s", err, createResult.Output)
	}
	if created.SessionID != "session-1" {
		t.Fatalf("created task session = %q, want session-1", created.SessionID)
	}

	listResult, err := registry.Execute(context.Background(), "task_list", map[string]interface{}{})
	if err != nil {
		t.Fatalf("execute task_list: %v", err)
	}
	if !strings.Contains(listResult.Output, created.ID) {
		t.Fatalf("task_list output should include created task %s, got: %s", created.ID, listResult.Output)
	}
}

func TestRegisterMimoActorToolsUsesCurrentSession(t *testing.T) {
	actorRegistry := actor.NewRegistry()
	registry := NewRegistry()
	registry.RegisterActorTools(func() *actor.Registry { return actorRegistry }, func() string { return "session-1" })

	for _, name := range []string{"actor_spawn", "actor_list", "actor_status", "actor_cancel", "actor_send"} {
		if _, ok := registry.Get(name); !ok {
			t.Fatalf("%s should be registered", name)
		}
	}

	spawnResult, err := registry.Execute(context.Background(), "actor_spawn", map[string]interface{}{
		"type":   "title",
		"prompt": "Generate a session title",
	})
	if err != nil {
		t.Fatalf("execute actor_spawn: %v", err)
	}
	if spawnResult.Error != "" {
		t.Fatalf("actor_spawn returned error: %s", spawnResult.Error)
	}

	var spawned actor.Actor
	if err := json.Unmarshal([]byte(spawnResult.Output), &spawned); err != nil {
		t.Fatalf("decode actor_spawn output: %v\n%s", err, spawnResult.Output)
	}
	if spawned.SessionID != "session-1" {
		t.Fatalf("spawned actor session = %q, want session-1", spawned.SessionID)
	}

	statusResult, err := registry.Execute(context.Background(), "actor_status", map[string]interface{}{
		"id": spawned.ID,
	})
	if err != nil {
		t.Fatalf("execute actor_status: %v", err)
	}
	if !strings.Contains(statusResult.Output, spawned.ID) {
		t.Fatalf("actor_status output should include actor id %s, got: %s", spawned.ID, statusResult.Output)
	}

	listResult, err := registry.Execute(context.Background(), "actor_list", map[string]interface{}{})
	if err != nil {
		t.Fatalf("execute actor_list: %v", err)
	}
	if !strings.Contains(listResult.Output, spawned.ID) {
		t.Fatalf("actor_list output should include actor id %s, got: %s", spawned.ID, listResult.Output)
	}
}

func writeTestFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
