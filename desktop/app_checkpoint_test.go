package desktop

import (
	"os"
	"testing"
	"time"

	"github.com/mimo-cli/mimo-cli/internal/agent"
	"github.com/mimo-cli/mimo-cli/internal/session"
)

func TestRestoreCheckpointLoadsCheckpointContextIntoAgent(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")

	store, err := session.Open()
	if err != nil {
		t.Fatalf("open session store: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
		_ = os.RemoveAll(tmpHome)
	})

	const sessionID = "session-1"
	if err := store.CreateSession(sessionID, session.DefaultWorkspaceID, "mimo", "tester"); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := store.SaveSession(sessionID, session.DefaultWorkspaceID, "mimo", "tester", []session.Message{
		{Role: "user", Content: "before checkpoint", CreatedAt: time.Now()},
		{Role: "assistant", Content: "after checkpoint assistant context", CreatedAt: time.Now()},
		{Role: "user", Content: "after checkpoint user context", CreatedAt: time.Now()},
	}); err != nil {
		t.Fatalf("save session: %v", err)
	}

	_, savedMessages, err := store.LoadSession(sessionID)
	if err != nil {
		t.Fatalf("load session: %v", err)
	}
	if len(savedMessages) != 3 {
		t.Fatalf("saved message count = %d, want 3", len(savedMessages))
	}

	cp := &session.Checkpoint{
		ID:            "cp-1",
		SessionID:     sessionID,
		Summary:       "Checkpoint says the user is wiring MiMo memory and task tools.",
		MessageOffset: int(savedMessages[0].ID),
		TokenCount:    120,
		Metadata:      "{}",
		CreatedAt:     time.Now(),
	}
	if err := store.SaveCheckpoint(cp); err != nil {
		t.Fatalf("save checkpoint: %v", err)
	}

	app := &App{
		agent:            &agent.Agent{},
		sessionStore:     store,
		currentSessionID: sessionID,
	}

	result := app.RestoreCheckpoint(cp.ID)
	if !result.Success {
		t.Fatalf("restore checkpoint failed: %s", result.Message)
	}

	if got, want := app.agent.GetMessageCount(), 3; got != want {
		t.Fatalf("agent message count after restore = %d, want %d", got, want)
	}
}
