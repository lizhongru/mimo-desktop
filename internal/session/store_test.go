package session

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()

	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "sessions.db"))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}
	if err := migrate(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	return &Store{db: db}
}

func TestCreateSessionEnsuresFolderWorkspace(t *testing.T) {
	store := newTestStore(t)
	workspacePath := filepath.Join(t.TempDir(), "ProjectA")
	workspaceID := "ws:" + workspacePath

	if err := store.CreateSession("session-1", workspaceID, "mimo", "tester"); err != nil {
		t.Fatalf("create session: %v", err)
	}

	assertSessionWorkspace(t, store, "session-1", workspaceID)

	ws, err := store.GetWorkspace(workspaceID)
	if err != nil {
		t.Fatalf("get workspace: %v", err)
	}
	if ws.Path != workspacePath {
		t.Fatalf("workspace path = %q, want %q", ws.Path, workspacePath)
	}
}

func TestMoveSessionEnsuresFolderWorkspace(t *testing.T) {
	store := newTestStore(t)
	workspacePath := filepath.Join(t.TempDir(), "ProjectB")
	workspaceID := "ws:" + workspacePath

	if err := store.CreateSession("session-1", DefaultWorkspaceID, "mimo", "tester"); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := store.MoveSession("session-1", workspaceID); err != nil {
		t.Fatalf("move session: %v", err)
	}

	assertSessionWorkspace(t, store, "session-1", workspaceID)
}

func TestLoadSessionAfterCreateThenSaveReturnsSavedMessages(t *testing.T) {
	store := newTestStore(t)

	if err := store.CreateSession("session-1", DefaultWorkspaceID, "mimo", "tester"); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := store.SaveSession("session-1", DefaultWorkspaceID, "mimo", "tester", []Message{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "hi there"},
	}); err != nil {
		t.Fatalf("save session: %v", err)
	}

	sess, messages, err := store.LoadSession("session-1")
	if err != nil {
		t.Fatalf("load session: %v", err)
	}
	if sess.ID != "session-1" {
		t.Fatalf("session id = %q, want session-1", sess.ID)
	}
	if len(messages) != 2 {
		t.Fatalf("message count = %d, want 2", len(messages))
	}
	if messages[0].Content != "hello" || messages[1].Content != "hi there" {
		t.Fatalf("messages = %#v", messages)
	}
}

func TestMigrateRepairsInvalidSessionCreatedAt(t *testing.T) {
	store := newTestStore(t)

	if err := store.CreateSession("session-1", DefaultWorkspaceID, "mimo", "tester"); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := store.SaveSession("session-1", DefaultWorkspaceID, "mimo", "tester", []Message{
		{Role: "user", Content: "hello"},
	}); err != nil {
		t.Fatalf("save session: %v", err)
	}
	if _, err := store.db.Exec("UPDATE sessions SET created_at = ? WHERE id = ?", "tester", "session-1"); err != nil {
		t.Fatalf("corrupt created_at: %v", err)
	}

	if err := migrate(store.db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	if _, _, err := store.LoadSession("session-1"); err != nil {
		t.Fatalf("load session after repair: %v", err)
	}
}

func assertSessionWorkspace(t *testing.T, store *Store, sessionID string, want string) {
	t.Helper()

	var got string
	err := store.db.QueryRow("SELECT workspace_id FROM sessions WHERE id = ?", sessionID).Scan(&got)
	if err != nil {
		t.Fatalf("query session workspace: %v", err)
	}
	if got != want {
		t.Fatalf("workspace id = %q, want %q", got, want)
	}
}
