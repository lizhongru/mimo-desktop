package desktop

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/mimo-cli/mimo-cli/internal/session"
)

// ============================================================
// DTOs
// ============================================================

type WorkspaceDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
}

type SessionDTO struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspaceId"`
	ModelName   string `json:"modelName"`
	UserName    string `json:"userName"`
	LastMessage string `json:"lastMessage"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type SessionData struct {
	ID          string           `json:"id"`
	WorkspaceID string           `json:"workspaceId"`
	ModelName   string           `json:"modelName"`
	Messages    []ChatMessageDTO `json:"messages"`
}

type ChatMessageDTO struct {
	Role       string   `json:"role"`
	Content    string   `json:"content"`
	Thinking   string   `json:"thinking,omitempty"`
	ToolLines  []string `json:"toolLines,omitempty"`
	Tokens     int      `json:"tokens"`
	ToolCalls  int      `json:"toolCalls"`
	DurationMs int64    `json:"durationMs"`
}

// ============================================================
// Workspace methods
// ============================================================

// ListWorkspaces returns all workspaces
func (a *App) ListWorkspaces() ([]WorkspaceDTO, error) {
	if a.sessionStore == nil {
		return nil, fmt.Errorf("session store not available")
	}
	workspaces, err := a.sessionStore.ListWorkspaces()
	if err != nil {
		return nil, err
	}
	result := make([]WorkspaceDTO, len(workspaces))
	for i, ws := range workspaces {
		result[i] = WorkspaceDTO{ID: ws.ID, Name: ws.Name, Type: ws.Type, Path: ws.Path}
	}
	return result, nil
}

// CreateWorkspace creates a folder-type workspace from a filesystem path
func (a *App) CreateWorkspace(path string) (WorkspaceDTO, error) {
	if a.sessionStore == nil {
		return WorkspaceDTO{}, fmt.Errorf("session store not available")
	}
	id := "ws:" + path
	name := filepath.Base(path)
	ws, err := a.sessionStore.EnsureWorkspace(id, name, "folder", path)
	if err != nil {
		return WorkspaceDTO{}, err
	}
	return WorkspaceDTO{ID: ws.ID, Name: ws.Name, Type: ws.Type, Path: ws.Path}, nil
}

// ============================================================
// Session methods
// ============================================================

// ListSessions returns all saved sessions (most recent first)
func (a *App) ListSessions(limit int) ([]SessionDTO, error) {
	if a.sessionStore == nil {
		return nil, fmt.Errorf("session store not available")
	}
	sessions, err := a.sessionStore.ListSessions(limit)
	if err != nil {
		return nil, err
	}
	result := make([]SessionDTO, len(sessions))
	for i, s := range sessions {
		result[i] = SessionDTO{
			ID:          s.ID,
			WorkspaceID: s.WorkspaceID,
			ModelName:   s.ModelName,
			UserName:    s.UserName,
			LastMessage: s.LastMessage,
			CreatedAt:   s.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   s.UpdatedAt.Format(time.RFC3339),
		}
	}
	return result, nil
}

// LoadSession loads a session and returns its messages
func (a *App) LoadSession(sessionID string) (*SessionData, error) {
	if a.sessionStore == nil {
		return nil, fmt.Errorf("session store not available")
	}
	sess, messages, err := a.sessionStore.LoadSession(sessionID)
	if err != nil {
		return nil, err
	}
	msgs := make([]ChatMessageDTO, len(messages))
	for i, m := range messages {
		msgs[i] = ChatMessageDTO{
			Role: m.Role, Content: m.Content, Thinking: m.Thinking,
			ToolLines: m.ToolLines, Tokens: m.Tokens, ToolCalls: m.ToolCalls, DurationMs: m.DurationMs,
		}
	}
	a.mu.Lock()
	a.currentSessionID = sessionID
	a.mu.Unlock()
	return &SessionData{
		ID:          sess.ID,
		WorkspaceID: sess.WorkspaceID,
		ModelName:   sess.ModelName,
		Messages:    msgs,
	}, nil
}

func (a *App) currentSessionWorkingDir() string {
	a.mu.Lock()
	sessionID := a.currentSessionID
	a.mu.Unlock()

	if sessionID == "" || a.sessionStore == nil {
		return ""
	}
	sess, _, err := a.sessionStore.LoadSession(sessionID)
	if err != nil || sess == nil {
		return ""
	}
	return session.WorkspacePathFromID(sess.WorkspaceID)
}

// CreateNewSession creates a new session bound to a workspace and returns its ID.
func (a *App) CreateNewSession(workspaceID string) string {
	a.mu.Lock()
	a.currentSessionID = uuid.New().String()
	id := a.currentSessionID
	a.mu.Unlock()
	a.agent.LoadMessages(nil)
	fmt.Println("[CreateNewSession] workspaceID:", workspaceID)
	if workspaceID == "" {
		workspaceID = session.DefaultWorkspaceID
	}
	if a.sessionStore != nil {
		a.sessionStore.CreateSession(id, workspaceID, a.cfg.DefaultModel, a.cfg.UserName)
	}
	return id
}

// DeleteSession removes a session from the store
func (a *App) DeleteSession(sessionID string) error {
	if a.sessionStore == nil {
		return fmt.Errorf("session store not available")
	}
	return a.sessionStore.DeleteSession(sessionID)
}

// RenameSession updates the display title of a session
func (a *App) RenameSession(sessionID string, title string) error {
	if a.sessionStore == nil {
		return fmt.Errorf("session store not available")
	}
	return a.sessionStore.RenameSession(sessionID, title)
}

// MoveSession moves a session to a different workspace
func (a *App) MoveSession(sessionID string, workspaceID string) error {
	if a.sessionStore == nil {
		return fmt.Errorf("session store not available")
	}
	return a.sessionStore.MoveSession(sessionID, workspaceID)
}

// SaveSessionFromFrontend saves a session with messages from the frontend.
// workspaceID is read from the existing session record.
func (a *App) SaveSessionFromFrontend(sessionID string, messages []ChatMessageDTO) error {
	if a.sessionStore == nil {
		return fmt.Errorf("session store not available")
	}
	msgs := make([]session.Message, len(messages))
	for i, m := range messages {
		msgs[i] = session.Message{
			Role: m.Role, Content: m.Content, Tokens: m.Tokens,
			ToolCalls: m.ToolCalls, DurationMs: m.DurationMs,
			Thinking: m.Thinking, ToolLines: m.ToolLines, CreatedAt: time.Now(),
		}
	}
	// Read workspaceID from the existing session record
	existingSess, _, err := a.sessionStore.LoadSession(sessionID)
	workspaceID := session.DefaultWorkspaceID
	if err == nil && existingSess != nil {
		workspaceID = existingSess.WorkspaceID
	}
	return a.sessionStore.SaveSession(sessionID, workspaceID, a.cfg.DefaultModel, a.cfg.UserName, msgs)
}
