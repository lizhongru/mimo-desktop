package desktop

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mimo-cli/mimo-cli/internal/session"
)

func TestDistillRunUsesSavedSessionMessages(t *testing.T) {
	app, store, sessionID := newCheckpointTestApp(t, 128000)

	messages := []session.Message{
		{Role: "user", Content: "请依次执行：\ngo test ./internal/skill -count=1\ngo test ./internal/skill -count=1\ngo test ./internal/skill -count=1", CreatedAt: time.Now()},
		{Role: "assistant", Content: "已执行：\n$ go test ./internal/skill -count=1\n$ go test ./internal/skill -count=1\n$ go test ./internal/skill -count=1", CreatedAt: time.Now()},
	}
	if err := store.SaveSession(sessionID, session.DefaultWorkspaceID, "mimo", "tester", messages); err != nil {
		t.Fatalf("save session: %v", err)
	}

	legacyCheckpoint := filepath.Join(".mimo", "memory", "sessions", sessionID, "checkpoint.md")
	if _, err := os.Stat(legacyCheckpoint); !os.IsNotExist(err) {
		t.Fatalf("test should not rely on legacy checkpoint, stat err: %v", err)
	}

	result := app.DistillRun()
	if !result.Success {
		t.Fatalf("distill failed: %s", result.Message)
	}
	if result.Count == 0 {
		t.Fatalf("expected candidates from saved session messages, got message %q", result.Message)
	}

	candidates := app.DistillListCandidates()
	if len(candidates) == 0 {
		t.Fatal("expected saved candidates to be listed")
	}
}

func TestDistillListCandidatesIncludesSkillMetadata(t *testing.T) {
	app, store, sessionID := newCheckpointTestApp(t, 128000)
	messages := []session.Message{
		{Role: "assistant", Content: "$ go test ./internal/skill -count=1\n$ go test ./internal/skill -count=1", CreatedAt: time.Now()},
	}
	if err := store.SaveSession(sessionID, session.DefaultWorkspaceID, "mimo", "tester", messages); err != nil {
		t.Fatalf("save session: %v", err)
	}

	result := app.DistillRun()
	if !result.Success || result.Count == 0 {
		t.Fatalf("distill result = %#v", result)
	}

	candidates := app.DistillListCandidates()
	if len(candidates) == 0 {
		t.Fatal("expected candidates")
	}
	if len(candidates[0].Commands) == 0 && candidates[0].Pattern == "" {
		t.Fatalf("expected metadata for frontend explanation: %#v", candidates[0])
	}
}
