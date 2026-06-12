package desktop

import (
	"os"
	"testing"
	"time"

	"github.com/mimo-cli/mimo-cli/internal/agent"
	iconfig "github.com/mimo-cli/mimo-cli/internal/config"
	"github.com/mimo-cli/mimo-cli/internal/session"
)

func newCheckpointTestApp(t *testing.T, maxTokens int) (*App, *session.Store, string) {
	t.Helper()

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")
	t.Chdir(t.TempDir())

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

	app := &App{
		sessionStore:     store,
		currentSessionID: sessionID,
		cfg: &iconfig.Config{
			DefaultModel: "mimo",
			UserName:     "tester",
			Context: iconfig.ContextConfig{
				MaxTokens: maxTokens,
			},
		},
	}
	return app, store, sessionID
}

func TestRestoreCheckpointLoadsCheckpointContextIntoAgent(t *testing.T) {
	app, store, sessionID := newCheckpointTestApp(t, 128000)
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

	app.agent = &agent.Agent{}

	result := app.RestoreCheckpoint(cp.ID)
	if !result.Success {
		t.Fatalf("restore checkpoint failed: %s", result.Message)
	}

	if got, want := app.agent.GetMessageCount(), 3; got != want {
		t.Fatalf("agent message count after restore = %d, want %d", got, want)
	}
}

func TestSaveSessionFromFrontendCreatesAutoCheckpointWhenThresholdReached(t *testing.T) {
	app, store, sessionID := newCheckpointTestApp(t, 20)

	err := app.SaveSessionFromFrontend(sessionID, []ChatMessageDTO{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "world"},
	})
	if err != nil {
		t.Fatalf("save session: %v", err)
	}

	checkpoints, err := store.ListCheckpoints(sessionID)
	if err != nil {
		t.Fatalf("list checkpoints: %v", err)
	}
	if len(checkpoints) != 1 {
		t.Fatalf("checkpoint count = %d, want 1", len(checkpoints))
	}
	if checkpoints[0].MessageOffset != 2 {
		t.Fatalf("message offset = %d, want 2", checkpoints[0].MessageOffset)
	}
	if checkpoints[0].Summary == "" {
		t.Fatal("summary is empty")
	}
}

func TestSaveSessionFromFrontendSkipsAutoCheckpointBelowThreshold(t *testing.T) {
	app, store, sessionID := newCheckpointTestApp(t, 1000)

	err := app.SaveSessionFromFrontend(sessionID, []ChatMessageDTO{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "world"},
	})
	if err != nil {
		t.Fatalf("save session: %v", err)
	}

	checkpoints, err := store.ListCheckpoints(sessionID)
	if err != nil {
		t.Fatalf("list checkpoints: %v", err)
	}
	if len(checkpoints) != 0 {
		t.Fatalf("checkpoint count = %d, want 0", len(checkpoints))
	}
}

func TestSaveSessionFromFrontendRespectsDisabledAutoCheckpointConfig(t *testing.T) {
	app, store, sessionID := newCheckpointTestApp(t, 20)
	app.cfg.Checkpoint.AutoCheckpoint = false
	app.cfg.Checkpoint.TokenThreshold = 0.01
	app.cfg.Checkpoint.MaxCheckpoints = 10
	app.cfg.Checkpoint.ContextBudget = 20

	err := app.SaveSessionFromFrontend(sessionID, []ChatMessageDTO{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "world"},
	})
	if err != nil {
		t.Fatalf("save session: %v", err)
	}

	checkpoints, err := store.ListCheckpoints(sessionID)
	if err != nil {
		t.Fatalf("list checkpoints: %v", err)
	}
	if len(checkpoints) != 0 {
		t.Fatalf("checkpoint count = %d, want 0", len(checkpoints))
	}
}

func TestSaveSessionFromFrontendDoesNotDuplicateAutoCheckpointAtSameOffset(t *testing.T) {
	app, store, sessionID := newCheckpointTestApp(t, 20)
	messages := []ChatMessageDTO{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "world"},
	}

	if err := app.SaveSessionFromFrontend(sessionID, messages); err != nil {
		t.Fatalf("first save session: %v", err)
	}
	if err := app.SaveSessionFromFrontend(sessionID, messages); err != nil {
		t.Fatalf("second save session: %v", err)
	}

	checkpoints, err := store.ListCheckpoints(sessionID)
	if err != nil {
		t.Fatalf("list checkpoints: %v", err)
	}
	if len(checkpoints) != 1 {
		t.Fatalf("checkpoint count = %d, want 1", len(checkpoints))
	}
}

func TestSaveSessionFromFrontendPrunesAutoCheckpointsToDefaultLimit(t *testing.T) {
	app, store, sessionID := newCheckpointTestApp(t, 20)
	var messages []ChatMessageDTO

	for i := 0; i < 12; i++ {
		messages = append(messages,
			ChatMessageDTO{Role: "user", Content: "hello"},
			ChatMessageDTO{Role: "assistant", Content: "world"},
		)
		if err := app.SaveSessionFromFrontend(sessionID, messages); err != nil {
			t.Fatalf("save session %d: %v", i, err)
		}
	}

	checkpoints, err := store.ListCheckpoints(sessionID)
	if err != nil {
		t.Fatalf("list checkpoints: %v", err)
	}
	if len(checkpoints) != 10 {
		t.Fatalf("checkpoint count = %d, want 10", len(checkpoints))
	}
}
