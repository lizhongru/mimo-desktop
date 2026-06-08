package desktop

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/mimo-cli/mimo-cli/internal/session"
)

// SessionDTO is a frontend-friendly session representation.
type SessionDTO struct {
	ID          string `json:"id"`
	ModelName   string `json:"modelName"`
	UserName    string `json:"userName"`
	LastMessage string `json:"lastMessage"`
	WorkingDir  string `json:"workingDir"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// SessionData contains a full session with messages for the frontend.
type SessionData struct {
	ID        string        `json:"id"`
	ModelName string        `json:"modelName"`
	Messages  []ChatMessageDTO `json:"messages"`
}

// ChatMessageDTO is a frontend-friendly message representation.
type ChatMessageDTO struct {
	Role       string   `json:"role"`
	Content    string   `json:"content"`
	Thinking   string   `json:"thinking,omitempty"`
	ToolLines  []string `json:"toolLines,omitempty"`
	Tokens     int      `json:"tokens"`
	ToolCalls  int      `json:"toolCalls"`
	DurationMs int64    `json:"durationMs"`
}

// ListSessions returns all saved sessions (most recent first).
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
			ModelName:   s.ModelName,
			UserName:    s.UserName,
			LastMessage: s.LastMessage,
			WorkingDir:  s.WorkingDir,
			CreatedAt:   s.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   s.UpdatedAt.Format(time.RFC3339),
		}
	}
	return result, nil
}

// LoadSession loads a session and returns its messages.
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
			Role:       m.Role,
			Content:    m.Content,
			Thinking:   m.Thinking,
			ToolLines:  m.ToolLines,
			Tokens:     m.Tokens,
			ToolCalls:  m.ToolCalls,
			DurationMs: m.DurationMs,
		}
	}
	a.mu.Lock()
	a.currentSessionID = sessionID
	a.mu.Unlock()
	return &SessionData{
		ID:        sess.ID,
		ModelName: sess.ModelName,
		Messages:  msgs,
	}, nil
}

// CreateNewSession creates a new session and returns its ID.
func (a *App) CreateNewSession() string {
	a.mu.Lock()
	a.currentSessionID = uuid.New().String()
	a.mu.Unlock()
	a.agent.LoadMessages(nil)
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.currentSessionID
}

// SaveCurrentSession saves the current conversation to the session store.
func (a *App) SaveCurrentSession() error {
	a.mu.Lock()
	sid := a.currentSessionID
	a.mu.Unlock()
	if a.sessionStore == nil || sid == "" {
		return nil
	}
	// Build session messages from agent's internal state
	// For now, we save what we have — the frontend sends messages via SendMessage
	// and we track them on the Go side
	return nil
}

// DeleteSession removes a session from the store.
func (a *App) DeleteSession(sessionID string) error {
	if a.sessionStore == nil {
		return fmt.Errorf("session store not available")
	}
	return a.sessionStore.DeleteSession(sessionID)
}

// RenameSession renames a session's display title (stored as lastMessage override).
func (a *App) RenameSession(sessionID string, title string) error {
	if a.sessionStore == nil {
		return fmt.Errorf("session store not available")
	}
	return a.sessionStore.RenameSession(sessionID, title)
}

// SaveSessionFromFrontend saves a session with messages from the frontend.
func (a *App) SaveSessionFromFrontend(sessionID string, messages []ChatMessageDTO) error {
	if a.sessionStore == nil {
		return fmt.Errorf("session store not available")
	}
	msgs := make([]session.Message, len(messages))
	for i, m := range messages {
		msgs[i] = session.Message{
			Role:       m.Role,
			Content:    m.Content,
			Tokens:     m.Tokens,
			ToolCalls:  m.ToolCalls,
			DurationMs: m.DurationMs,
			Thinking:   m.Thinking,
			ToolLines:  m.ToolLines,
			CreatedAt:  time.Now(),
		}
	}
	workingDir, _ := os.Getwd()
	return a.sessionStore.SaveSession(sessionID, a.cfg.DefaultModel, a.cfg.UserName, workingDir, msgs)
}
