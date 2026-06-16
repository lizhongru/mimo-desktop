# Workspace 架构 — 涉及文件完整代码

---

## 1. internal/session/store.go

数据库层。定义 Workspace / Session / Message 结构体、建表迁移、所有 CRUD 方法。

```go
package session

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Store manages session persistence with SQLite
type Store struct {
	db *sql.DB
}

// DefaultWorkspaceID is the permanent "conversations" workspace
const DefaultWorkspaceID = "default"

// Workspace represents a project/workspace that groups sessions
type Workspace struct {
	ID        string
	Name      string
	Type      string // "chat" or "folder"
	Path      string // filesystem path (empty for chat type)
	CreatedAt time.Time
}

// Session represents a conversation session
type Session struct {
	ID          string
	WorkspaceID string
	ModelName   string
	UserName    string
	LastMessage string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Message represents a single message in a session
type Message struct {
	ID         int64
	SessionID  string
	Role       string
	Content    string
	Tokens     int
	ToolCalls  int
	DurationMs int64
	Thinking   string
	ToolLines  []string
	CreatedAt  time.Time
}

// Open opens or creates the SQLite database
func Open() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot get home dir: %w", err)
	}

	dbDir := filepath.Join(home, ".mimo")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create .mimo dir: %w", err)
	}

	dbPath := filepath.Join(dbDir, "sessions.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open database: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("cannot set WAL mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("cannot enable foreign keys: %w", err)
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func migrate(db *sql.DB) error {
	// Create tables
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS workspaces (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL DEFAULT '',
			type TEXT NOT NULL DEFAULT 'chat',
			path TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			workspace_id TEXT NOT NULL DEFAULT 'default',
			model_name TEXT NOT NULL DEFAULT '',
			user_name TEXT NOT NULL DEFAULT '',
			last_message TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL DEFAULT '',
			tokens INTEGER NOT NULL DEFAULT 0,
			tool_calls INTEGER NOT NULL DEFAULT 0,
			duration_ms INTEGER NOT NULL DEFAULT 0,
			thinking TEXT NOT NULL DEFAULT '',
			tool_lines TEXT NOT NULL DEFAULT '[]',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_messages_session ON messages(session_id);
		CREATE INDEX IF NOT EXISTS idx_sessions_workspace ON sessions(workspace_id);
	`)
	if err != nil {
		return err
	}

	// --- Migration v2: thinking + tool_lines ---
	db.Exec("ALTER TABLE messages ADD COLUMN thinking TEXT NOT NULL DEFAULT ''")
	db.Exec("ALTER TABLE messages ADD COLUMN tool_lines TEXT NOT NULL DEFAULT '[]'")

	// --- Migration v3: workspaces + workspace_id ---
	// Create default workspace if it doesn't exist
	db.Exec(`INSERT OR IGNORE INTO workspaces (id, name, type, path) VALUES (?, 'default', 'chat', '')`, DefaultWorkspaceID)

	// Add workspace_id column to sessions (old DBs had working_dir)
	db.Exec("ALTER TABLE sessions ADD COLUMN workspace_id TEXT NOT NULL DEFAULT 'default'")

	// Migrate old working_dir data: create folder workspaces for each unique working_dir
	rows, err := db.Query(`SELECT DISTINCT working_dir FROM sessions WHERE working_dir != '' AND working_dir IS NOT NULL`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var wd string
			if rows.Scan(&wd) == nil {
				wsID := "ws:" + wd
				db.Exec(`INSERT OR IGNORE INTO workspaces (id, name, type, path) VALUES (?, ?, 'folder', ?)`, wsID, filepath.Base(wd), wd)
				db.Exec(`UPDATE sessions SET workspace_id = ? WHERE working_dir = ?`, wsID, wd)
			}
		}
	}

	// Drop old working_dir column (SQLite doesn't support DROP COLUMN before 3.35, so we leave it)

	// Ensure default workspace exists
	db.Exec(`INSERT OR IGNORE INTO workspaces (id, name, type, path) VALUES (?, 'default', 'chat', '')`, DefaultWorkspaceID)

	return nil
}

// ============================================================
// Workspace CRUD
// ============================================================

// EnsureWorkspace creates a workspace if it doesn't exist, returns the workspace
func (s *Store) EnsureWorkspace(id, name, wsType, path string) (*Workspace, error) {
	_, err := s.db.Exec(`
		INSERT INTO workspaces (id, name, type, path, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name, path = excluded.path
	`, id, name, wsType, path, time.Now())
	if err != nil {
		return nil, err
	}
	return &Workspace{ID: id, Name: name, Type: wsType, Path: path, CreatedAt: time.Now()}, nil
}

// GetDefaultWorkspace returns the permanent "conversations" workspace
func (s *Store) GetDefaultWorkspace() (*Workspace, error) {
	return s.EnsureWorkspace(DefaultWorkspaceID, "default", "chat", "")
}

// GetWorkspace returns a workspace by ID
func (s *Store) GetWorkspace(id string) (*Workspace, error) {
	ws := &Workspace{}
	err := s.db.QueryRow(`SELECT id, name, type, path, created_at FROM workspaces WHERE id = ?`, id).
		Scan(&ws.ID, &ws.Name, &ws.Type, &ws.Path, &ws.CreatedAt)
	if err != nil {
		return nil, err
	}
	return ws, nil
}

// ListWorkspaces returns all workspaces
func (s *Store) ListWorkspaces() ([]Workspace, error) {
	rows, err := s.db.Query(`SELECT id, name, type, path, created_at FROM workspaces ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var workspaces []Workspace
	for rows.Next() {
		var ws Workspace
		if err := rows.Scan(&ws.ID, &ws.Name, &ws.Type, &ws.Path, &ws.CreatedAt); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, ws)
	}
	return workspaces, nil
}

// DeleteWorkspace removes a workspace and all its sessions (cascade)
func (s *Store) DeleteWorkspace(id string) error {
	if id == DefaultWorkspaceID {
		return fmt.Errorf("cannot delete the default workspace")
	}
	_, err := s.db.Exec("DELETE FROM workspaces WHERE id = ?", id)
	return err
}

// ============================================================
// Session CRUD
// ============================================================

// CreateSession creates an empty session bound to a workspace
func (s *Store) CreateSession(id, workspaceID, modelName, userName string) error {
	if workspaceID == "" {
		workspaceID = DefaultWorkspaceID
	}
	_, err := s.db.Exec(`
		INSERT INTO sessions (id, workspace_id, model_name, user_name, last_message, created_at, updated_at)
		VALUES (?, ?, ?, '', '', ?, ?)
	`, id, workspaceID, modelName, userName, time.Now(), time.Now())
	return err
}

// SaveSession creates or updates a session with its messages
func (s *Store) SaveSession(id, workspaceID, modelName, userName string, messages []Message) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if workspaceID == "" {
		workspaceID = DefaultWorkspaceID
	}

	// Find last user message for display title
	lastMsg := ""
	if messages != nil {
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == "user" {
				lastMsg = messages[i].Content
				if len(lastMsg) > 200 {
					lastMsg = lastMsg[:200]
				}
				break
			}
		}
	}

	// Upsert session
	_, err = tx.Exec(`
		INSERT INTO sessions (id, workspace_id, model_name, user_name, last_message, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			model_name = excluded.model_name,
			user_name = excluded.user_name,
			last_message = excluded.last_message,
			updated_at = excluded.updated_at
	`, id, workspaceID, modelName, userName, lastMsg, time.Now(), time.Now())
	if err != nil {
		return err
	}

	if messages != nil {
		_, err = tx.Exec("DELETE FROM messages WHERE session_id = ?", id)
		if err != nil {
			return err
		}
		stmt, err := tx.Prepare(`
			INSERT INTO messages (session_id, role, content, tokens, tool_calls, duration_ms, thinking, tool_lines, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`)
		if err != nil {
			return err
		}
		defer stmt.Close()
		for _, msg := range messages {
			toolLinesJSON, _ := json.Marshal(msg.ToolLines)
			_, err = stmt.Exec(id, msg.Role, msg.Content, msg.Tokens, msg.ToolCalls, msg.DurationMs, msg.Thinking, string(toolLinesJSON), msg.CreatedAt)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// LoadSession loads a session and its messages by ID
func (s *Store) LoadSession(id string) (*Session, []Message, error) {
	sess := &Session{}
	err := s.db.QueryRow(`
		SELECT id, workspace_id, model_name, user_name, last_message, created_at, updated_at
		FROM sessions WHERE id = ?
	`, id).Scan(&sess.ID, &sess.WorkspaceID, &sess.ModelName, &sess.UserName, &sess.LastMessage, &sess.CreatedAt, &sess.UpdatedAt)
	if err != nil {
		return nil, nil, err
	}

	rows, err := s.db.Query(`
		SELECT id, session_id, role, content, tokens, tool_calls, duration_ms, thinking, tool_lines, created_at
		FROM messages WHERE session_id = ?
		ORDER BY id ASC
	`, id)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var toolLinesJSON string
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.Tokens, &msg.ToolCalls, &msg.DurationMs, &msg.Thinking, &toolLinesJSON, &msg.CreatedAt); err != nil {
			return nil, nil, err
		}
		json.Unmarshal([]byte(toolLinesJSON), &msg.ToolLines)
		messages = append(messages, msg)
	}
	return sess, messages, nil
}

// ListSessions returns all sessions that have at least one message, most recent first
func (s *Store) ListSessions(limit int) ([]Session, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(`
		SELECT DISTINCT s.id, s.workspace_id, s.model_name, s.user_name, s.last_message, s.created_at, s.updated_at
		FROM sessions s
		INNER JOIN messages m ON m.session_id = s.id
		ORDER BY s.updated_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var sess Session
		if err := rows.Scan(&sess.ID, &sess.WorkspaceID, &sess.ModelName, &sess.UserName, &sess.LastMessage, &sess.CreatedAt, &sess.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, sess)
	}
	return sessions, nil
}

// ListSessionsByWorkspace returns sessions for a specific workspace
func (s *Store) ListSessionsByWorkspace(workspaceID string, limit int) ([]Session, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(`
		SELECT DISTINCT s.id, s.workspace_id, s.model_name, s.user_name, s.last_message, s.created_at, s.updated_at
		FROM sessions s
		INNER JOIN messages m ON m.session_id = s.id
		WHERE s.workspace_id = ?
		ORDER BY s.updated_at DESC
		LIMIT ?
	`, workspaceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []Session
	for rows.Next() {
		var sess Session
		if err := rows.Scan(&sess.ID, &sess.WorkspaceID, &sess.ModelName, &sess.UserName, &sess.LastMessage, &sess.CreatedAt, &sess.UpdatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, sess)
	}
	return sessions, nil
}

// DeleteSession removes a session and its messages
func (s *Store) DeleteSession(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec("DELETE FROM messages WHERE session_id = ?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM sessions WHERE id = ?", id); err != nil {
		return err
	}
	return tx.Commit()
}

// RenameSession updates the last_message (display title) of a session
func (s *Store) RenameSession(id, title string) error {
	_, err := s.db.Exec("UPDATE sessions SET last_message = ?, updated_at = ? WHERE id = ?", title, time.Now(), id)
	return err
}

// MoveSession moves a session to a different workspace
func (s *Store) MoveSession(sessionID, workspaceID string) error {
	_, err := s.db.Exec("UPDATE sessions SET workspace_id = ?, updated_at = ? WHERE id = ?", workspaceID, time.Now(), sessionID)
	return err
}

// CountMessages returns the number of messages in a session
func (s *Store) CountMessages(sessionID string) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM messages WHERE session_id = ?", sessionID).Scan(&count)
	return count, err
}
```

---

## 2. desktop/app_session.go

Wails 桥接层。暴露给前端的 Go 方法，包括 Workspace 和 Session 的增删改查。

```go
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
	ID           string `json:"id"`
	WorkspaceID  string `json:"workspaceId"`
	ModelName    string `json:"modelName"`
	UserName     string `json:"userName"`
	FirstMessage string `json:"firstMessage"`
	LastMessage  string `json:"lastMessage"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
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

// CreateNewSession creates a new session bound to a workspace and returns its ID.
func (a *App) CreateNewSession(workspaceID string) string {
	a.mu.Lock()
	a.currentSessionID = uuid.New().String()
	id := a.currentSessionID
	a.mu.Unlock()
	a.agent.LoadMessages(nil)
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
```

---

## 3. stores/sessionStore.ts

前端 Zustand 状态管理。定义 SessionItem、WorkspaceItem 接口，管理 sessions / workspaces / selectedWorkspace 状态。

```typescript
import { create } from "zustand";

export interface SessionItem {
  id: string;
  workspaceId: string;
  modelName: string;
  userName: string;
  lastMessage: string;
  createdAt: string;
  updatedAt: string;
}

export interface WorkspaceItem {
  id: string;
  name: string;
  type: string; // "chat" | "folder"
  path: string;
}

const WS_STORAGE_KEY = "mimo_selectedWorkspace";

interface SessionState {
  sessions: SessionItem[];
  workspaces: WorkspaceItem[];
  currentSessionId: string | null;
  streamingSessionId: string | null;
  exportingSessionId: string | null;
  leftSidebarOpen: boolean;
  selectedWorkspace: string; // workspace ID

  setSessions: (sessions: SessionItem[]) => void;
  setWorkspaces: (workspaces: WorkspaceItem[]) => void;
  setCurrentSessionId: (id: string | null) => void;
  setStreamingSessionId: (id: string | null) => void;
  setExportingSessionId: (id: string | null) => void;
  addSession: (session: SessionItem) => void;
  removeSession: (id: string) => void;
  updateSession: (id: string, lastMessage: string) => void;
  toggleLeftSidebar: () => void;
  setLeftSidebarOpen: (open: boolean) => void;
  setSelectedWorkspace: (id: string) => void;
}

function loadSelectedWorkspace(): string {
  try {
    return localStorage.getItem(WS_STORAGE_KEY) || "";
  } catch {
    return "";
  }
}

export const useSessionStore = create<SessionState>((set) => ({
  sessions: [],
  workspaces: [],
  currentSessionId: null,
  streamingSessionId: null,
  exportingSessionId: null,
  leftSidebarOpen: true,
  selectedWorkspace: loadSelectedWorkspace(),

  setSessions: (sessions) => set({ sessions }),
  setWorkspaces: (workspaces) => set({ workspaces }),
  setCurrentSessionId: (id) => set({ currentSessionId: id }),
  setStreamingSessionId: (id) => set({ streamingSessionId: id }),
  setExportingSessionId: (id) => set({ exportingSessionId: id }),
  addSession: (session) =>
    set((s) => ({ sessions: [session, ...s.sessions] })),
  removeSession: (id) =>
    set((s) => ({
      sessions: s.sessions.filter((sess) => sess.id !== id),
      currentSessionId: s.currentSessionId === id ? null : s.currentSessionId,
    })),
  updateSession: (id, lastMessage) =>
    set((s) => ({
      sessions: s.sessions.map((sess) =>
        sess.id === id ? { ...sess, lastMessage } : sess
      ),
    })),
  toggleLeftSidebar: () =>
    set((s) => ({ leftSidebarOpen: !s.leftSidebarOpen })),
  setLeftSidebarOpen: (open) => set({ leftSidebarOpen: open }),
  setSelectedWorkspace: (id) => {
    try {
      if (id) {
        localStorage.setItem(WS_STORAGE_KEY, id);
      } else {
        localStorage.removeItem(WS_STORAGE_KEY);
      }
    } catch { /* ignore */ }
    set({ selectedWorkspace: id });
  },
}));
```

---

## 4. App.tsx

主组件。核心业务逻辑：handleSelectWorkspace、handleNewChat、handleSend、handleLoadSession。

```typescript
import { useState, useEffect, useCallback } from "react";
import { AppLayout } from "./components/layout/AppLayout";
import { useAgent } from "./hooks/useAgent";
import { useChatStore } from "./stores/chatStore";
import { useSessionStore } from "./stores/sessionStore";
import { useSettingsStore } from "./stores/settingsStore";
import { useActivityStore } from "./stores/activityStore";
import { t } from "./lib/i18n";

const DEFAULT_WS = "default";

declare global {
  interface Window {
    go: {
      desktop: {
        App: {
          SendMessage: (message: string, attachmentsJSON?: string) => Promise<void>;
          CancelOperation: () => Promise<void>;
          IsBusy: () => Promise<boolean>;
          RespondToConfirm: (approved: boolean) => Promise<void>;
          RespondToConfirmAll: (approved: boolean) => Promise<void>;
          GetModelName: () => Promise<string>;
          GetVersion: () => Promise<Record<string, string>>;
          CompressContext: () => Promise<{ before: number; after: number }>;
          ExportChat: (messages: Array<{ role: string; content: string }>) => Promise<void>;
          GetWorkingDir: () => Promise<string>;
          GetTools: () => Promise<Array<{ name: string; description: string; safetyLevel: string; isMcp: boolean; serverName: string }>>;
          GetMCPServers: () => Promise<Array<{ name: string; connected: boolean; toolCount: number; tools: string[] }>>;
          // Workspace methods
          ListWorkspaces: () => Promise<Array<{ id: string; name: string; type: string; path: string }>>;
          CreateWorkspace: (path: string) => Promise<{ id: string; name: string; type: string; path: string }>;
          // Session methods
          ListSessions: (limit: number) => Promise<Array<{ id: string; workspaceId: string; modelName: string; userName: string; lastMessage: string; createdAt: string; updatedAt: string }>>;
          CreateNewSession: (workspaceId: string) => Promise<string>;
          LoadSession: (id: string) => Promise<{ id: string; workspaceId: string; modelName: string; messages: Array<{ role: string; content: string; thinking?: string; toolLines?: string[]; tokens: number; toolCalls: number; durationMs: number }> }>;
          DeleteSession: (id: string) => Promise<void>;
          RenameSession: (id: string, title: string) => Promise<void>;
          MoveSession: (sessionId: string, workspaceId: string) => Promise<void>;
          SaveSessionFromFrontend: (sessionId: string, messages: unknown[]) => Promise<void>;
          // Config methods
          GetConfig: () => Promise<{ defaultModel: string; language: string; theme: string; userName: string; models: Record<string, { provider: string; website: string; apiBase: string; apiKey: string; model: string; models: string[]; fallback: string; maxTokens: number; temperature: number; topP: number; streaming: boolean; vision: boolean; tools: boolean }>; safety: { level: string; permission: string }; agent: { maxIterations: number; planningMode: string; permission: string; reasoningLevel: string; showTokenUsage: boolean } }>;
          SetTheme: (theme: string) => Promise<void>;
          SetLanguage: (lang: string) => Promise<void>;
          SetDefaultModel: (name: string) => Promise<void>;
          AddModel: (name: string, provider: string, website: string, apiBase: string, apiKey: string, model: string, models: string[], fallback: string, maxTokens: number, temperature: number, topP: number, streaming: boolean, vision: boolean, tools: boolean) => Promise<void>;
          UpdateModel: (name: string, provider: string, website: string, apiBase: string, apiKey: string, model: string, models: string[], fallback: string, maxTokens: number, temperature: number, topP: number, streaming: boolean, vision: boolean, tools: boolean) => Promise<void>;
          RemoveModel: (name: string) => Promise<void>;
          SetSafetyLevel: (level: string) => Promise<void>;
          SetPlanningMode: (mode: string) => Promise<void>;
          SetPermission: (perm: string) => Promise<void>;
          SetReasoningLevel: (level: string) => Promise<void>;
          WindowMinimise: () => Promise<void>;
          WindowMaximise: () => Promise<void>;
          WindowClose: () => Promise<void>;
          WindowIsMaximised: () => Promise<boolean>;
          OpenInExplorer: (path: string) => Promise<void>;
          SelectDirectory: () => Promise<string>;
          ListRemoteModels: (modelName: string) => Promise<Array<{ id: string; owned_by: string; description?: string; context_window?: number; max_output?: number; capabilities?: string[] }>>;
          ListRemoteModelsWithConfig: (apiBase: string, apiKey: string) => Promise<Array<{ id: string; owned_by: string; description?: string; context_window?: number; max_output?: number; capabilities?: string[] }>>;
        };
      };
    };
  }
}

export default function App() {
  const currentModel = useSettingsStore((s) => s.currentModel);
  useAgent();

  // Load sessions, workspaces and config on mount
  useEffect(() => {
    Promise.all([
      window.go?.desktop?.App?.ListSessions?.(30),
      window.go?.desktop?.App?.ListWorkspaces?.(),
    ]).then(([sessions, workspaces]) => {
      useSessionStore.getState().setSessions(sessions || []);
      useSessionStore.getState().setWorkspaces(workspaces || []);
    }).catch(console.error);

    window.go?.desktop?.App?.GetConfig?.().then((cfg) => {
      useSettingsStore.getState().initFromConfig({
        theme: cfg.theme, language: cfg.language, defaultModel: cfg.defaultModel,
        models: cfg.models, planningMode: cfg.agent?.planningMode,
        safetyLevel: cfg.safety?.level, permission: cfg.agent?.permission,
        reasoningLevel: cfg.agent?.reasoningLevel,
      });
    }).catch(console.error);
  }, []);

  // Send a message — lazy-creates session if needed
  const handleSend = useCallback(async (message: string, attachments?: { name: string; type: string; dataUrl: string }[]) => {
    const sessStore = useSessionStore.getState();
    let sid = sessStore.currentSessionId;

    if (!sid) {
      const ws = sessStore.selectedWorkspace || DEFAULT_WS;
      try {
        sid = await window.go?.desktop?.App?.CreateNewSession?.(ws);
        sessStore.setCurrentSessionId(sid);
      } catch (e) {
        console.error("CreateNewSession failed:", e);
        return;
      }
    }

    if (!sessStore.sessions.find((s) => s.id === sid)) {
      sessStore.addSession({
        id: sid!,
        workspaceId: sessStore.selectedWorkspace || DEFAULT_WS,
        modelName: "",
        userName: "",
        lastMessage: message,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      });
    }

    sessStore.setStreamingSessionId(sid);
    useChatStore.getState().addUserMessage(message);
    const attachmentsJSON = attachments && attachments.length > 0 ? JSON.stringify(attachments) : "";
    window.go?.desktop?.App?.SendMessage?.(message, attachmentsJSON).catch((err) => {
      console.error("SendMessage failed:", err);
      useSessionStore.getState().setStreamingSessionId(null);
      useChatStore.getState().finalizeResponse(`${t("error_prefix")}: ${err}`, 0);
    });
  }, []);

  const handleCancel = useCallback(() => {
    window.go?.desktop?.App?.CancelOperation?.().catch(console.error);
  }, []);

  // New chat — create session bound to current workspace
  const handleNewChat = useCallback(() => {
    const ws = useSessionStore.getState().selectedWorkspace || DEFAULT_WS;
    window.go?.desktop?.App?.CreateNewSession?.(ws).then((id) => {
      useSessionStore.getState().setCurrentSessionId(id);
      useChatStore.getState().clearMessages();
      useActivityStore.getState().clear();
    }).catch(console.error);
  }, []);

  // Load existing session — restore workspace from backend
  const handleLoadSession = useCallback((id: string) => {
    window.go?.desktop?.App?.LoadSession?.(id).then((data) => {
      useSessionStore.getState().setCurrentSessionId(id);
      useChatStore.getState().clearMessages();
      useActivityStore.getState().clear();
      if (data?.messages) {
        for (const msg of data.messages) {
          useChatStore.getState().addRestoredMessage(msg as { role: "user" | "assistant"; content: string; thinking?: string; toolLines?: string[]; tokens?: number; toolCalls?: number; durationMs?: number });
        }
      }
      // Restore workspace selection from backend
      if (data?.workspaceId) {
        useSessionStore.getState().setSelectedWorkspace(data.workspaceId);
      }
    }).catch(console.error);
  }, []);

  const handleDeleteSession = useCallback(async (id: string) => {
    await window.go?.desktop?.App?.DeleteSession?.(id);
    useSessionStore.getState().removeSession(id);
    if (useSessionStore.getState().currentSessionId === id) {
      useChatStore.getState().clearMessages();
      useSessionStore.getState().setCurrentSessionId(null);
    }
  }, []);

  const [toast, setToast] = useState<string | null>(null);

  // Select workspace — update store; session binding happens at creation time
  const handleSelectWorkspace = useCallback(async (dir: string) => {
    // dir is a filesystem path; convert to workspace ID
    if (dir) {
      try {
        const ws = await window.go?.desktop?.App?.CreateWorkspace?.(dir);
        useSessionStore.getState().setSelectedWorkspace(ws.id);
        // Refresh workspaces list
        const list = await window.go?.desktop?.App?.ListWorkspaces?.();
        useSessionStore.getState().setWorkspaces(list || []);
      } catch (e) {
        console.error("CreateWorkspace failed:", e);
      }
    } else {
      useSessionStore.getState().setSelectedWorkspace(DEFAULT_WS);
    }
  }, []);

  const handleExportSession = useCallback(async (id: string) => {
    useSessionStore.getState().setExportingSessionId(id);
    try {
      const data = await window.go?.desktop?.App?.LoadSession?.(id);
      if (!data?.messages?.length) {
        setToast(t("export_empty"));
        return;
      }
      const exportMsgs = data.messages.map((m) => ({ role: m.role, content: m.content }));
      await window.go?.desktop?.App?.ExportChat?.(exportMsgs);
      setToast(t("export_success"));
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e);
      if (!msg.includes("cancelled")) {
        setToast(t("export_failed"));
      }
    } finally {
      useSessionStore.getState().setExportingSessionId(null);
      setTimeout(() => setToast(null), 2500);
    }
  }, []);

  const handleConfirmApprove = useCallback(() => {
    window.go?.desktop?.App?.RespondToConfirm?.(true).catch(console.error);
  }, []);
  const handleConfirmDeny = useCallback(() => {
    window.go?.desktop?.App?.RespondToConfirm?.(false).catch(console.error);
  }, []);
  const handleConfirmApproveAll = useCallback(() => {
    window.go?.desktop?.App?.RespondToConfirmAll?.(true).catch(console.error);
  }, []);

  // Keyboard shortcuts
  useEffect(() => {
    const handleRegenerate = () => {
      const msgs = useChatStore.getState().messages;
      const lastUser = [...msgs].reverse().find((m) => m.role === "user");
      if (lastUser) {
        const lastUserIdx = msgs.lastIndexOf(lastUser);
        useChatStore.setState({ messages: msgs.slice(0, lastUserIdx) });
        handleSend(lastUser.content);
      }
    };
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        const state = useChatStore.getState();
        if (state.confirmAction) {
          state.setConfirmAction(null);
          window.go?.desktop?.App?.RespondToConfirm?.(false).catch(console.error);
        } else if (state.isStreaming) {
          handleCancel();
        }
        return;
      }
      const mod = e.ctrlKey || e.metaKey;
      if (mod) {
        switch (e.key.toLowerCase()) {
          case "b":
            e.preventDefault();
            useSessionStore.getState().toggleLeftSidebar();
            break;
          case "i":
            e.preventDefault();
            useActivityStore.getState().toggleRightSidebar();
            break;
          case "n":
            e.preventDefault();
            handleNewChat();
            break;
          case "k":
            e.preventDefault();
            if (!useChatStore.getState().isCompressing && !useChatStore.getState().isStreaming) {
              window.go?.desktop?.App?.CompressContext?.().catch(console.error);
            }
            break;
        }
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    window.addEventListener("mimo:regenerate", handleRegenerate);
    return () => {
      window.removeEventListener("keydown", handleKeyDown);
      window.removeEventListener("mimo:regenerate", handleRegenerate);
    };
  }, [handleCancel, handleNewChat, handleSend]);

  return (
    <>
      <AppLayout
        modelName={currentModel}
        onSend={handleSend}
        onCancel={handleCancel}
        onNewChat={handleNewChat}
        onLoadSession={handleLoadSession}
        onDeleteSession={handleDeleteSession}
        onExportSession={handleExportSession}
        onConfirmApprove={handleConfirmApprove}
        onConfirmDeny={handleConfirmDeny}
        onConfirmApproveAll={handleConfirmApproveAll}
        onSelectWorkspace={handleSelectWorkspace}
      />
      {toast && (
        <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-[200] bg-elevated border border-bdr text-txt text-sm px-4 py-2 rounded-lg shadow-lg animate-fade-in">
          {toast}
        </div>
      )}
    </>
  );
}
```

---

## 5. components/layout/AppLayout.tsx

布局组件。传递 onSend / onSelectWorkspace 等 props 给 WelcomeView 和 ChatInput。

```typescript
import { useState, useEffect, useCallback, useRef } from "react";
import { useSessionStore } from "../../stores/sessionStore";
import { useSettingsStore } from "../../stores/settingsStore";
import { useActivityStore } from "../../stores/activityStore";
import { LeftSidebar } from "./LeftSidebar";
import { RightSidebar } from "./RightSidebar";
import { MessageList } from "../chat/MessageList";
import { ChatInput } from "../chat/ChatInput";
import { StatusBar } from "../chat/StatusBar";
import { ConfirmDialog } from "../confirm/ConfirmDialog";
import { ToolsViewer } from "../common/ToolsViewer";
import { SettingsPage } from "../settings/SettingsPage";
import { WelcomeView } from "../welcome/WelcomeView";
import { useChatStore } from "../../stores/chatStore";
import {
  PanelLeft,
  PanelRight,
  Minus,
  Square,
  X,
  Copy,
  Wrench,
} from "lucide-react";
import { t } from "../../lib/i18n";

function formatTokens(n: number): string {
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`;
  return String(n);
}

interface Props {
  modelName: string;
  onSend: (message: string, attachments?: { name: string; type: string; dataUrl: string }[]) => void;
  onCancel: () => void;
  onNewChat: () => void;
  onLoadSession: (id: string) => void;
  onDeleteSession: (id: string) => Promise<void>;
  onExportSession: (id: string) => Promise<void>;
  onConfirmApprove: () => void;
  onConfirmDeny: () => void;
  onConfirmApproveAll: () => void;
  onSelectWorkspace: (dir: string) => Promise<void>;
}

export function AppLayout({
  modelName,
  onSend,
  onCancel,
  onNewChat,
  onLoadSession,
  onDeleteSession,
  onExportSession,
  onConfirmApprove,
  onConfirmDeny,
  onConfirmApproveAll,
  onSelectWorkspace,
}: Props) {
  useSettingsStore((s) => s.language);
  const messages = useChatStore((s) => s.messages);
  const leftOpen = useSessionStore((s) => s.leftSidebarOpen);
  const rightOpen = useActivityStore((s) => s.rightSidebarOpen);
  const confirmAction = useChatStore((s) => s.confirmAction);
  const usage = useChatStore((s) => s.usage);
  const [toolsOpen, setToolsOpen] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);

  const [isMaximised, setIsMaximised] = useState(false);
  const [dragActive, setDragActive] = useState(false);
  const dragCounter = useRef(0);

  useEffect(() => {
    window.go?.desktop?.App?.WindowIsMaximised?.()
      .then(setIsMaximised)
      .catch(() => {});
  }, []);

  const handleMinimise = useCallback(() => {
    window.go?.desktop?.App?.WindowMinimise?.().catch(console.error);
  }, []);

  const handleMaximise = useCallback(() => {
    window.go?.desktop?.App?.WindowMaximise?.().catch(console.error);
    setTimeout(() => {
      window.go?.desktop?.App?.WindowIsMaximised?.()
        .then(setIsMaximised)
        .catch(() => {});
    }, 100);
  }, []);

  const handleClose = useCallback(() => {
    window.go?.desktop?.App?.WindowClose?.().catch(console.error);
  }, []);

  return (
    <div
      className={`h-screen flex flex-col bg-root text-txt select-none relative ${dragActive ? "drag-active" : ""}`} onDragEnter={(e) => { e.preventDefault(); dragCounter.current += 1; if (dragCounter.current === 1) setDragActive(true); }} onDragLeave={(e) => { e.preventDefault(); dragCounter.current -= 1; if (dragCounter.current <= 0) { dragCounter.current = 0; setDragActive(false); } }} onDragOver={(e) => e.preventDefault()} onDrop={(e) => { e.preventDefault(); dragCounter.current = 0; setDragActive(false); }}>
      {/* Modern Title Bar */}
      <div className="relative h-10 flex items-center border-b border-bdr-div bg-root flex-shrink-0 drag-region">
        {/* Left: sidebar toggle */}
        <div className="flex items-center gap-1 pl-3 no-drag z-10">
          <button
            onClick={() => useSessionStore.getState().toggleLeftSidebar()}
            className={`p-1.5 rounded-md hover:bg-elevated/80 transition-colors cursor-pointer no-drag ${
              leftOpen ? "text-txt" : "text-txt-g"
            }`}
            title={t("toggle_left_sidebar")}
          >
            <PanelLeft className="w-[15px] h-[15px]" />
          </button>
        </div>

        {/* Center: App title — absolute centered */}
        <div className="absolute inset-0 flex items-center justify-center pointer-events-none drag-region">
          <span className="text-[13px] font-semibold tracking-wide text-txt">
            {t("app_name")}
          </span>
        </div>

        {/* Right: sidebar toggle + window controls */}
        <div className="flex items-center no-drag z-10 ml-auto">
          <button
            onClick={() => setToolsOpen(true)}
            className={`p-1.5 rounded-md hover:bg-elevated/80 transition-colors cursor-pointer text-txt-g hover:text-txt`}
            title={t("tools")}
          >
            <Wrench className="w-[15px] h-[15px]" />
          </button>

          <button
            onClick={() => useActivityStore.getState().toggleRightSidebar()}
            className={`p-1.5 rounded-md hover:bg-elevated/80 transition-colors cursor-pointer ${
              rightOpen ? "text-txt" : "text-txt-g"
            }`}
            title={t("toggle_right_sidebar")}
          >
            <PanelRight className="w-[15px] h-[15px]" />
          </button>

          {/* Divider */}
          <div className="w-px h-4 bg-bdr-div mx-1" />

          {/* Window controls */}
          <button
            onClick={handleMinimise}
            className="w-[46px] h-10 flex items-center justify-center hover:bg-elevated/80 transition-colors cursor-pointer text-txt-g hover:text-txt"
            title={t("minimize")}
          >
            <Minus className="w-[14px] h-[14px]" strokeWidth={1.5} />
          </button>
          <button
            onClick={handleMaximise}
            className="w-[46px] h-10 flex items-center justify-center hover:bg-elevated/80 transition-colors cursor-pointer text-txt-g hover:text-txt"
            title={t("maximize")}
          >
            {isMaximised ? (
              <Copy className="w-[12px] h-[12px]" strokeWidth={1.5} />
            ) : (
              <Square className="w-[12px] h-[12px]" strokeWidth={1.5} />
            )}
          </button>
          <button
            onClick={handleClose}
            className="w-[46px] h-10 flex items-center justify-center hover:bg-red-500/80 transition-colors cursor-pointer text-txt-g hover:text-white rounded-tr-md"
            title={t("close_tooltip")}
          >
            <X className="w-[15px] h-[15px]" strokeWidth={1.5} />
          </button>
        </div>
      </div>

      {/* Main area */}
      <div className="flex-1 flex overflow-hidden">
        {/* Left Sidebar */}
        <div
          className={`border-r border-bdr bg-sidebar transition-all duration-200 flex-shrink-0 overflow-hidden ${
            leftOpen ? "w-[260px]" : "w-0"
          }`}
        >
          {leftOpen && (
            <LeftSidebar
              onNewChat={onNewChat}
              onLoadSession={onLoadSession}
              onDeleteSession={onDeleteSession}
              onExportSession={onExportSession}
              onOpenSettings={() => setSettingsOpen(true)}
            />
          )}
        </div>

        {/* Center: Chat */}
        <div className="flex-1 flex flex-col min-w-0">
          {messages.length === 0 ? (
            <WelcomeView onSend={onSend} onSelectWorkspace={onSelectWorkspace} />
          ) : (
            <>
              <MessageList />
              <ChatInput onSend={onSend} onCancel={onCancel} />
              <StatusBar modelName={modelName} />
            </>
          )}
        </div>

        {/* Right Sidebar */}
        <div
          className={`border-l border-bdr bg-surface transition-all duration-200 flex-shrink-0 overflow-hidden ${
            rightOpen ? "w-[320px]" : "w-0"
          }`}
        >
          {rightOpen && <RightSidebar />}
        </div>
      </div>

      {/* Settings Page Modal */}
      <SettingsPage
        open={settingsOpen}
        onClose={() => setSettingsOpen(false)}
        defaultModel={modelName}
      />

      {/* Tools Viewer Modal */}
      <ToolsViewer open={toolsOpen} onClose={() => setToolsOpen(false)} />

      {/* Confirm Dialog (global overlay) */}
      <ConfirmDialog
        action={confirmAction}
        onApprove={() => {
          onConfirmApprove();
          useChatStore.getState().setConfirmAction(null);
        }}
        onDeny={() => {
          onConfirmDeny();
          useChatStore.getState().setConfirmAction(null);
        }}
        onApproveAll={() => {
          onConfirmApproveAll();
          useChatStore.getState().setConfirmAction(null);
        }}
      />
    </div>
  );
}
```

---

## 6. components/welcome/WelcomeView.tsx

欢迎页。包含 WorkspacePicker 组件（选择文件夹工作区）和消息输入框。

```typescript
import { useState, useRef, useCallback, useEffect, KeyboardEvent } from "react";
import {
  Send,
  FolderOpen,
  ChevronDown,
  Route,
  Shield,
  Brain,
  Cpu,
  Check,
  ChevronRight,
  Paperclip,
  ImageIcon,
  X,
  FileText,
} from "lucide-react";
import { useChatStore } from "../../stores/chatStore";
import { useSettingsStore } from "../../stores/settingsStore";
import { useSessionStore } from "../../stores/sessionStore";
import { t } from "../../lib/i18n";
import { useAnimatedOpen } from "../../lib/useAnimatedOpen";

interface Props {
  onSend: (message: string, attachments?: { name: string; type: string; dataUrl: string }[]) => void;
  onSelectWorkspace: (dir: string) => Promise<void>;
}

function WorkspacePicker({
  value,
  onChange,
}: {
  value: string;
  onChange: (dir: string) => void;
}) {
  const [rawOpen, setRawOpen] = useState(false);
  const { shouldRender, closing } = useAnimatedOpen(rawOpen, 120);
  const [recentDirs, setRecentDirs] = useState<string[]>([]);
  const sessions = useSessionStore((s) => s.sessions);
  const workspaces = useSessionStore((s) => s.workspaces);

  // Collect folder workspaces for recent list
  useEffect(() => {
    const dirs = workspaces
      .filter((ws) => ws.type === "folder" && ws.path)
      .map((ws) => ws.path);
    setRecentDirs(dirs.slice(0, 5));
  }, [workspaces]);

  const handleSelect = (dir: string) => {
    onChange(dir);
    setRawOpen(false);
  };

  const handleBrowse = async () => {
    try {
      const dir = await window.go?.desktop?.App?.SelectDirectory?.();
      if (dir) {
        handleSelect(dir);
      } else {
        setRawOpen(false);
      }
    } catch (err) {
      console.error("Failed to select directory:", err);
      setRawOpen(false);
    }
  };

  const displayName = value
    ? value.split(/[/\\]/).filter(Boolean).pop() || value
    : t("select_workspace");

  return (
    <div className="relative inline-block">
      <button
        onClick={() => setRawOpen(!rawOpen)}
        className="flex items-center gap-2 px-3 py-2 rounded-lg bg-elevated border border-bdr hover:border-accent/30 transition-colors cursor-pointer text-sm text-txt-2 hover:text-txt"
      >
        <FolderOpen className="w-4 h-4 text-txt-g" />
        <span className="max-w-[200px] truncate">{displayName}</span>
        <ChevronDown className="w-3.5 h-3.5 text-txt-2" />
      </button>

      {shouldRender && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setRawOpen(false)} />
          <div
            className={`absolute bottom-full mb-2 left-0 z-50 bg-surface border border-bdr rounded-lg shadow-xl px-2 py-1 min-w-[220px] ${
              closing ? "animate-pop-out" : "animate-pop-up"
            }`}
          >
            {/* No project option — conversation without workspace */}
            <button
              onClick={() => handleSelect("")}
              className={`w-full flex items-center gap-2 px-2.5 py-2 text-xs rounded-md transition-colors cursor-pointer ${
                !value
                  ? "text-accent bg-accent/10"
                  : "text-txt-2 hover:bg-elevated"
              }`}
            >
              <span className="w-3.5 h-3.5 flex items-center justify-center text-[11px] font-bold text-txt-g">&#x2205;</span>
              <span>{t("no_project")}</span>
              {!value && <Check className="w-3 h-3 ml-auto flex-shrink-0" />}
            </button>
            <div className="h-px bg-bdr-div my-1" />
            <div className="px-2.5 py-1 text-[10px] text-txt-m uppercase tracking-wider">
              {t("recent_workspaces")}
            </div>

            {recentDirs.length > 0 ? (
              recentDirs.map((dir) => {
                const name = dir.split(/[/\\]/).filter(Boolean).pop() || dir;
                return (
                  <button
                    key={dir}
                    onClick={() => handleSelect(dir)}
                    className={`w-full flex items-center gap-2 px-2.5 py-2 text-xs rounded-md transition-colors cursor-pointer ${
                      dir === value
                        ? "text-accent bg-accent/10"
                        : "text-txt-2 hover:bg-elevated"
                    }`}
                  >
                    <FolderOpen className="w-3.5 h-3.5 flex-shrink-0" />
                    <span className="truncate">{name}</span>
                    {dir === value && <Check className="w-3 h-3 ml-auto flex-shrink-0" />}
                  </button>
                );
              })
            ) : (
              <div className="px-2.5 py-2 text-xs text-txt-m">
                {t("no_recent_workspaces")}
              </div>
            )}

            <div className="h-px bg-bdr-div my-1" />

            <button
              onClick={handleBrowse}
              className="w-full flex items-center gap-2 px-2.5 py-2 text-xs text-txt-2 hover:bg-elevated rounded-md transition-colors cursor-pointer"
            >
              <FolderOpen className="w-3.5 h-3.5" />
              {t("browse_folder")}
            </button>
          </div>
        </>
      )}
    </div>
  );
}

function MiniDropdown({
  icon: Icon,
  label,
  value,
  options,
  onChange,
}: {
  icon: typeof ChevronDown;
  label: string;
  value: string;
  options: { key: string; label: string }[];
  onChange: (v: string) => void;
}) {
  const [rawOpen, setRawOpen] = useState(false);
  const { shouldRender, closing } = useAnimatedOpen(rawOpen, 120);

  return (
    <div className="relative">
      <button
        onClick={() => setRawOpen(!rawOpen)}
        className="flex items-center gap-1 px-1.5 py-0.5 rounded text-[11px] text-txt-2 hover:text-txt hover:bg-bdr/40 transition-colors cursor-pointer"
        title={label}
      >
        <Icon className="w-3 h-3" />
        <span className="max-w-[72px] truncate">
          {options.find((o) => o.key === value)?.label || value}
        </span>
        <ChevronDown className="w-2.5 h-2.5 text-txt-2" />
      </button>
      {shouldRender && (
        <>
          <div className="fixed inset-0 z-40" onClick={() => setRawOpen(false)} />
          <div
            className={`absolute bottom-full mb-1.5 left-0 z-50 bg-surface border border-bdr rounded-lg shadow-xl px-2 py-1 min-w-[120px] ${
              closing ? "animate-pop-out" : "animate-pop-up"
            }`}
          >
            {options.map((opt) => (
              <button
                key={opt.key}
                onClick={() => {
                  onChange(opt.key);
                  setRawOpen(false);
                }}
                className={`w-full text-left px-3 py-2 text-xs rounded-md transition-colors cursor-pointer ${
                  opt.key === value
                    ? "text-accent bg-accent/10"
                    : "text-txt-2 hover:bg-elevated"
                }`}
              >
                {opt.label}
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  );
}

function ModelReasoningPicker() {
  const [rawOpen, setRawOpen] = useState(false);
  const { shouldRender: open, closing } = useAnimatedOpen(rawOpen, 150);
  const currentModel = useSettingsStore((s) => s.currentModel);
  const currentModelKey = useSettingsStore((s) => s.currentModelKey);
  const models = useSettingsStore((s) => s.models);
  const reasoningLevel = useSettingsStore((s) => s.reasoningLevel);
  useSettingsStore((s) => s.language);
  const triggerRef = useRef<HTMLButtonElement>(null);

  const reasoningOptions = [
    { key: "low", label: t("reasoning_low"), icon: "\u26a1" },
    { key: "medium", label: t("reasoning_medium"), icon: "\u2696\ufe0f" },
    { key: "high", label: t("reasoning_high"), icon: "\ud83e\udde0" },
  ];
  const reasoningIdx = reasoningOptions.findIndex((o) => o.key === reasoningLevel);
  const [showModels, setShowModels] = useState(false);
  const [panelLeft, setPanelLeft] = useState(false);

  const close = () => { setRawOpen(false); setShowModels(false); };
  const handleOpen = () => {
    if (rawOpen) { close(); return; }
    if (triggerRef.current) {
      const rect = triggerRef.current.getBoundingClientRect();
      setPanelLeft(window.innerWidth - rect.right < 460);
    }
    setRawOpen(true);
  };

  return (
    <div className="relative">
      <button ref={triggerRef} onClick={handleOpen}
        className={`flex items-center gap-1 px-2 py-1 rounded-lg text-[11px] transition-all cursor-pointer ${rawOpen ? "bg-accent/15 text-accent border border-accent/30" : "text-txt-2 hover:text-txt hover:bg-bdr/40 border border-transparent"}`}
        title={t("current_model")}>
        <Cpu className="w-3 h-3" />
        <span className="max-w-[120px] truncate font-medium">{currentModel || "..."}</span>
        <ChevronDown className={`w-2.5 h-2.5 transition-transform ${rawOpen ? "rotate-180" : ""}`} />
      </button>
      {open && <div className="fixed inset-0 z-40" onClick={close} />}
      {open && (
        <div className={`absolute bottom-full mb-2 mt-1 z-50 ${panelLeft ? "right-0" : "left-0"} ${closing ? "animate-pop-out" : "animate-pop-up"}`}>
          <div className="flex items-end gap-0.5">
            {showModels && (
              <div className={`w-[180px] bg-surface border border-bdr rounded-xl shadow-2xl px-2 py-1.5 ${panelLeft ? "order-2" : "order-1"} animate-slide-right`}>
                <div className="px-2.5 py-1 text-[10px] text-txt-m uppercase tracking-wider mb-1">{t("current_model")}</div>
                <div className="space-y-0.5 max-h-[160px] overflow-y-auto">
                  {models.map((m) => (
                    <button key={m}
                      onClick={() => { useSettingsStore.getState().setCurrentModel(m); close(); }}
                      className={`w-full flex items-center justify-between px-2.5 py-2 text-xs rounded-lg transition-all cursor-pointer ${m === currentModelKey ? "bg-accent/10 text-accent" : "text-txt-2 hover:bg-elevated"}`}>
                      <span className="truncate">{m}</span>
                      {m === currentModelKey && <Check className="w-3.5 h-3.5 text-accent flex-shrink-0" />}
                    </button>
                  ))}
                </div>
              </div>
            )}
            <div className={`${showModels ? (panelLeft ? "order-1" : "order-2") : ""} w-[260px] bg-surface border border-bdr rounded-xl shadow-2xl overflow-hidden`}>
              <div className="px-3.5 pt-3 pb-2.5">
                <div className="flex items-center gap-1.5 mb-2.5">
                  <Brain className="w-3.5 h-3.5 text-accent" />
                  <span className="text-[11px] font-medium text-txt">{t("reasoning_label")}</span>
                </div>
                <div className="relative flex bg-elevated rounded-lg p-0.5">
                  <div className="absolute top-0.5 bottom-0.5 rounded-md bg-accent shadow-sm transition-all duration-200 ease-out"
                    style={{
                      width: `calc(${100 / reasoningOptions.length}% - 2px)`,
                      left: `calc(${reasoningIdx * 100 / reasoningOptions.length}% + 2px)`,
                    }} />
                  {reasoningOptions.map((opt) => (
                    <button key={opt.key}
                      onClick={() => useSettingsStore.getState().setReasoningLevel(opt.key)}
                      className={`relative z-10 flex-1 flex items-center justify-center gap-1 px-2 py-1.5 text-[11px] rounded-md transition-colors duration-200 cursor-pointer ${opt.key === reasoningLevel ? "text-white font-medium" : "text-txt-2 hover:text-txt"}`}>
                      <span className="text-[10px]">{opt.icon}</span>
                      <span>{opt.label}</span>
                    </button>
                  ))}
                </div>
              </div>
              <div className="h-px bg-bdr-div mx-3" />
              <button onClick={() => setShowModels(!showModels)}
                className="w-full flex items-center justify-between px-3.5 py-2.5 text-xs text-txt cursor-pointer hover:bg-elevated transition-colors">
                <span className="flex items-center gap-2">
                  <Cpu className="w-3.5 h-3.5 text-accent" />
                  <span className="font-medium truncate">{currentModel || "..."}</span>
                </span>
                <ChevronRight className={`w-3 h-3 text-txt-2 transition-transform duration-200 ${showModels ? "rotate-180" : ""}`} />
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}


export function WelcomeView({ onSend, onSelectWorkspace }: Props) {
  const [text, setText] = useState("");
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const isStreaming = useChatStore((s) => s.isStreaming);
  const planningMode = useSettingsStore((s) => s.planningMode);
  const permission = useSettingsStore((s) => s.permission);
  const [attachments, setAttachments] = useState<{ name: string; type: string; dataUrl: string }[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [workspace, setWorkspace] = useState("");

  const handleSend = useCallback(() => {
    const trimmed = text.trim();
    if (!trimmed || isStreaming) return;
    onSend(trimmed, attachments.length > 0 ? attachments : undefined);
    setText("");
    setAttachments([]);
    if (textareaRef.current) textareaRef.current.style.height = "auto";
  }, [text, isStreaming, onSend]);

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); handleSend(); }
  };

  const handleFiles = useCallback((files: FileList) => {
    Array.from(files).forEach((file) => {
      const reader = new FileReader();
      reader.onload = () => {
        setAttachments((prev) => [...prev, { name: file.name, type: file.type, dataUrl: reader.result as string }]);
      };
      reader.readAsDataURL(file);
    });
  }, []);

  const removeAttachment = (i: number) => setAttachments((prev) => prev.filter((_, idx) => idx !== i));
  const handleDrop = useCallback((e: React.DragEvent) => { e.preventDefault(); e.stopPropagation(); if (e.dataTransfer.files.length > 0) handleFiles(e.dataTransfer.files); }, [handleFiles]);
  const handleDragOver = (e: React.DragEvent) => { e.preventDefault(); e.stopPropagation(); };
  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => { setText(e.target.value); const el = e.target; el.style.height = "auto"; el.style.height = Math.min(el.scrollHeight, 200) + "px"; };

  const handleWorkspaceChange = async (dir: string) => {
    setWorkspace(dir);
    onSelectWorkspace(dir);
  };

  return (
    <div className="flex-1 flex flex-col items-center justify-center px-4 pb-8">
      <div className="mb-8 text-center animate-fade-in">
        <div className="w-20 h-20 mx-auto mb-4 rounded-2xl bg-accent/15 flex items-center justify-center">
          <span className="text-4xl font-bold text-accent">M</span>
        </div>
        <h1 className="text-2xl font-semibold text-txt mb-2">MiMo Desktop</h1>
        <p className="text-sm text-txt-g">{t("welcome_subtitle")}</p>
      </div>

      <div className="w-full max-w-2xl">
        <div className="bg-elevated border border-bdr rounded-xl focus-within:border-accent/50 focus-within:ring-1 focus-within:ring-accent/20 transition-colors" onDrop={handleDrop} onDragOver={handleDragOver}>
          {attachments.length > 0 && (
            <div className="flex flex-wrap gap-2 px-4 pt-2">
              {attachments.map((att, i) => (
                <div key={i} className="flex items-center gap-1.5 bg-surface border border-bdr rounded-md px-2 py-1 text-xs">
                  <Paperclip className="w-3 h-3 text-txt-g" />
                  <span className="text-txt-2 max-w-[120px] truncate">{att.name}</span>
                  <button onClick={() => removeAttachment(i)} className="text-txt-g hover:text-red-400 cursor-pointer"><X className="w-3 h-3" /></button>
                </div>
              ))}
            </div>
          )}
          <textarea ref={textareaRef} value={text} onChange={handleChange} onKeyDown={handleKeyDown} placeholder={t("welcome_input_placeholder")} rows={2} className="w-full resize-none bg-transparent px-4 pt-3 pb-1 text-sm text-txt placeholder:text-txt-2 focus:outline-none min-h-[56px]" readOnly={isStreaming} autoFocus />
          <div className="flex items-center justify-between px-2 pb-2 pt-0.5">
            <div className="flex items-center gap-0.5">
              <input ref={fileInputRef} type="file" multiple accept="image/*,.pdf,.txt,.md,.json,.csv" className="hidden" onChange={(e) => { if (e.target.files) handleFiles(e.target.files); e.target.value = ""; }} />
              <button onClick={() => fileInputRef.current?.click()} className="flex items-center gap-1 px-1.5 py-0.5 rounded text-[11px] text-txt-2 hover:text-txt hover:bg-bdr/40 transition-colors cursor-pointer" title={t("attach_file")}><Paperclip className="w-3 h-3" /></button>
              <div className="w-px h-3 bg-txt-2/30 mx-0.5" />
              <MiniDropdown icon={Route} label={t("planning_mode")} value={planningMode} options={[{ key: "auto", label: t("plan_auto") }, { key: "react", label: t("plan_react") }, { key: "plan-execute", label: t("plan_execute") }]} onChange={(v) => useSettingsStore.getState().setPlanningMode(v)} />
              <div className="w-px h-3 bg-txt-2/30 mx-0.5" />
              <MiniDropdown icon={Shield} label={t("safety_level")} value={permission} options={[{ key: "readonly", label: t("perm_readonly") }, { key: "write", label: t("perm_write") }, { key: "exec", label: t("perm_exec") }]} onChange={(v) => useSettingsStore.getState().setPermission(v)} />
            </div>
            <div className="flex items-center gap-0.5">
              <ModelReasoningPicker />
              <div className="w-px h-3 bg-txt-2/30 mx-0.5" />
              <button onClick={handleSend} disabled={!text.trim()} className="w-7 h-7 flex items-center justify-center rounded-lg bg-accent/20 text-accent hover:bg-accent/30 disabled:opacity-30 disabled:cursor-not-allowed transition-colors cursor-pointer" title={t("send_enter")}><Send className="w-3.5 h-3.5" /></button>
            </div>
          </div>
        </div>
        <div className="flex justify-center mt-4">
          <WorkspacePicker value={workspace} onChange={handleWorkspaceChange} />
        </div>
        <p className="text-center text-[11px] text-txt-m mt-4">{t("welcome_hint")}</p>
      </div>
    </div>
  );
}

```

---

## 7. components/chat/ChatInput.tsx

聊天输入框。不涉及 workspace，只调 onSend。

```typescript
import { useState, useRef, useCallback, useEffect, KeyboardEvent } from "react";
import {
  Send, Square, ChevronDown, ChevronRight, Brain, Shield, Route,
  Cpu, Check, Paperclip, ImageIcon, X,
} from "lucide-react";
import { useChatStore } from "../../stores/chatStore";
import { useSettingsStore } from "../../stores/settingsStore";
import { t } from "../../lib/i18n";
import { useAnimatedOpen } from "../../lib/useAnimatedOpen";

interface Props {
  onSend: (message: string, attachments?: { name: string; type: string; dataUrl: string }[]) => void;
  onCancel: () => void;
}

function Dropdown({ icon: Icon, label, value, options, onChange }: {
  icon: typeof ChevronDown; label: string; value: string;
  options: { key: string; label: string }[]; onChange: (v: string) => void;
}) {
  const [rawOpen, setRawOpen] = useState(false);
  const { shouldRender, closing } = useAnimatedOpen(rawOpen, 120);
  const close = () => setRawOpen(false);
  return (
    <div className="relative">
      <button onClick={() => setRawOpen(!rawOpen)}
        className="flex items-center gap-1 px-1.5 py-0.5 rounded text-[11px] text-txt-2 hover:text-txt hover:bg-bdr/40 transition-colors cursor-pointer"
        title={label}>
        <Icon className="w-3 h-3" />
        <span className="max-w-[72px] truncate">{options.find((o) => o.key === value)?.label || value}</span>
        <ChevronDown className="w-2.5 h-2.5 text-txt-2" />
      </button>
      {shouldRender && (<>
        <div className="fixed inset-0 z-40" onClick={close} />
        <div className={`absolute bottom-full mb-1.5 left-0 z-50 bg-surface border border-bdr rounded-lg shadow-xl px-2 py-1 min-w-[120px] ${closing ? "animate-pop-out" : "animate-pop-up"}`}>
          {options.map((opt) => (
            <button key={opt.key} onClick={() => { onChange(opt.key); close(); }}
              className={`w-full text-left px-3 py-2 text-xs rounded-md transition-colors cursor-pointer ${opt.key === value ? "text-accent bg-accent/10" : "text-txt-2 hover:bg-elevated"}`}>
              {opt.label}
            </button>
          ))}
        </div>
      </>)}
    </div>
  );
}

function ModelReasoningPicker() {
  const [rawOpen, setRawOpen] = useState(false);
  const { shouldRender: open, closing } = useAnimatedOpen(rawOpen, 150);
  const currentModel = useSettingsStore((s) => s.currentModel);
  const currentModelKey = useSettingsStore((s) => s.currentModelKey);
  const models = useSettingsStore((s) => s.models);
  const reasoningLevel = useSettingsStore((s) => s.reasoningLevel);
  useSettingsStore((s) => s.language);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const [showModels, setShowModels] = useState(false);

  useEffect(() => { useSettingsStore.getState().refreshModels(); }, []);

  const reasoningOptions = [
    { key: "low", label: t("reasoning_low"), icon: "\u26a1" },
    { key: "medium", label: t("reasoning_medium"), icon: "\u2696\ufe0f" },
    { key: "high", label: t("reasoning_high"), icon: "\ud83e\udde0" },
  ];
  const reasoningIdx = reasoningOptions.findIndex((o) => o.key === reasoningLevel);

  const close = () => { setRawOpen(false); setShowModels(false); };
  const handleOpen = () => { if (rawOpen) { close(); return; } setRawOpen(true); };

  const modelsMap = (useSettingsStore.getState() as any)._modelsMap as Record<string, { model?: string }> | undefined;

  return (
    <div className="relative">
      <button ref={triggerRef} onClick={handleOpen}
        className={`flex items-center gap-1 px-2 py-1 rounded-lg text-[11px] transition-all cursor-pointer ${rawOpen ? "bg-accent/15 text-accent border border-accent/30" : "text-txt-2 hover:text-txt hover:bg-bdr/40 border border-transparent"}`}
        title={t("current_model")}>
        <Cpu className="w-3 h-3" />
        <span className="max-w-[120px] truncate font-medium">{currentModel || "..."}</span>
        <ChevronDown className={`w-2.5 h-2.5 transition-transform ${rawOpen ? "rotate-180" : ""}`} />
      </button>
      {open && <div className="fixed inset-0 z-40" onClick={close} />}
      {open && (
        <div className={`absolute bottom-full mb-2 right-0 z-50 ${closing ? "animate-pop-out" : "animate-pop-up"}`}>
          <div className="flex items-end gap-0.5">
            {showModels && (
              <div className="w-[220px] bg-surface border border-bdr rounded-xl shadow-2xl overflow-hidden order-1 animate-slide-right">
                <div className="px-3 pt-2.5 pb-1">
                  <div className="text-[10px] text-txt-2 uppercase tracking-wider mb-1.5">{t("current_model")}</div>
                </div>
                <div className="max-h-[200px] overflow-y-auto px-1.5 pb-1.5 space-y-0.5">
                  {models.map((m) => {
                    const display = modelsMap?.[m]?.model || m;
                    const isActive = m === currentModelKey;
                    return (
                      <button key={m}
                        onClick={() => { useSettingsStore.getState().setCurrentModel(m); setShowModels(false); setRawOpen(false); }}
                        className={`w-full flex items-center justify-between px-2.5 py-2 rounded-lg transition-all cursor-pointer ${isActive ? "bg-accent/10" : "hover:bg-elevated"}`}>
                        <span className={`text-xs font-medium truncate ${isActive ? "text-accent" : "text-txt"}`}>{display}</span>
                        {isActive && <Check className="w-3.5 h-3.5 text-accent flex-shrink-0" />}
                      </button>
                    );
                  })}
                </div>
              </div>
            )}
            <div className={`${showModels ? "order-2" : ""} w-[260px] bg-surface border border-bdr rounded-xl shadow-2xl overflow-hidden`}>
              <div className="px-3.5 pt-3 pb-2.5">
                <div className="flex items-center gap-1.5 mb-2.5">
                  <Brain className="w-3.5 h-3.5 text-accent" />
                  <span className="text-[11px] font-medium text-txt">{t("reasoning_label")}</span>
                </div>
                <div className="relative flex bg-elevated rounded-lg p-0.5">
                  <div className="absolute top-0.5 bottom-0.5 rounded-md bg-accent shadow-sm transition-all duration-200 ease-out"
                    style={{ width: `calc(${100 / reasoningOptions.length}% - 2px)`, left: `calc(${reasoningIdx * 100 / reasoningOptions.length}% + 2px)` }} />
                  {reasoningOptions.map((opt) => (
                    <button key={opt.key}
                      onClick={() => useSettingsStore.getState().setReasoningLevel(opt.key)}
                      className={`relative z-10 flex-1 flex items-center justify-center gap-1 px-2 py-1.5 text-[11px] rounded-md transition-colors duration-200 cursor-pointer ${opt.key === reasoningLevel ? "text-white font-medium" : "text-txt-2 hover:text-txt"}`}>
                      <span className="text-[10px]">{opt.icon}</span>
                      <span>{opt.label}</span>
                    </button>
                  ))}
                </div>
              </div>
              <div className="h-px bg-bdr-div mx-3" />
              <button onClick={() => setShowModels(!showModels)}
                className="w-full flex items-center justify-between px-3.5 py-2.5 text-xs text-txt cursor-pointer hover:bg-elevated transition-colors">
                <span className="flex items-center gap-2">
                  <Cpu className="w-3.5 h-3.5 text-accent" />
                  <span className="font-medium truncate max-w-[160px]">{currentModel || "..."}</span>
                </span>
                <ChevronRight className={`w-3 h-3 text-txt-2 transition-transform duration-200 ${showModels ? "rotate-180" : ""}`} />
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
export function ChatInput({ onSend, onCancel }: Props) {
  const [text, setText] = useState("");
  const isStreaming = useChatStore((s) => s.isStreaming);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const planningMode = useSettingsStore((s) => s.planningMode);
  const permission = useSettingsStore((s) => s.permission);
  const [attachments, setAttachments] = useState<{ name: string; type: string; dataUrl: string }[]>([]);
  const [dragHighlight, setDragHighlight] = useState(false);
  const dragCounter = useRef(0);

  const handleSend = useCallback(() => {
    const trimmed = text.trim();
    if (!trimmed || isStreaming) return;
    onSend(trimmed, attachments.length > 0 ? attachments : undefined);
    setText("");
    setAttachments([]);
    if (textareaRef.current) textareaRef.current.style.height = "auto";
  }, [text, isStreaming, onSend, attachments]);

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); handleSend(); }
  };

  const handleFiles = useCallback((files: FileList) => {
    Array.from(files).forEach((file) => {
      const reader = new FileReader();
      reader.onload = () => {
        setAttachments((prev) => [...prev, { name: file.name, type: file.type, dataUrl: reader.result as string }]);
      };
      reader.readAsDataURL(file);
    });
  }, []);

  const removeAttachment = (i: number) => setAttachments((prev) => prev.filter((_, idx) => idx !== i));
  const handleDrop = useCallback((e: React.DragEvent) => { e.preventDefault(); dragCounter.current = 0; setDragHighlight(false); if (e.dataTransfer.files.length > 0) handleFiles(e.dataTransfer.files); }, [handleFiles]);
  const handleDragOver = (e: React.DragEvent) => { e.preventDefault(); e.dataTransfer.dropEffect = "copy"; };
  const handleDragEnter = useCallback((e: React.DragEvent) => { e.preventDefault(); dragCounter.current += 1; if (dragCounter.current === 1) setDragHighlight(true); }, []);
  const handleDragLeave = useCallback((e: React.DragEvent) => { e.preventDefault(); dragCounter.current -= 1; if (dragCounter.current <= 0) { dragCounter.current = 0; setDragHighlight(false); } }, []);
  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => { setText(e.target.value); const el = e.target; el.style.height = "auto"; el.style.height = Math.min(el.scrollHeight, 200) + "px"; };

  return (
    <div className="px-4 py-3">
      <div className="max-w-4xl mx-auto">
        <div className={`bg-elevated border rounded-xl focus-within:border-accent/50 focus-within:ring-1 focus-within:ring-accent/20 transition-colors ${dragHighlight ? "border-accent ring-2 ring-accent/30 bg-accent/5 drag-drop-zone" : "border-bdr"}`} onDrop={handleDrop} onDragOver={handleDragOver} onDragEnter={handleDragEnter} onDragLeave={handleDragLeave}>
          {dragHighlight && attachments.length === 0 && (
            <div className="flex items-center justify-center gap-2 px-4 py-3 text-accent/80">
              <Paperclip className="w-4 h-4" />
              <span className="text-xs font-medium">{t("drop_to_add")}</span>
            </div>
          )}
          {attachments.length > 0 && (
            <div className="flex flex-wrap gap-2 px-4 pt-2">
              {attachments.map((att, i) => (
                <div key={i} className="flex items-center gap-1.5 bg-surface border border-bdr rounded-md px-2 py-1 text-xs">
                  {att.type.startsWith("image/") ? <ImageIcon className="w-3 h-3 text-accent" /> : <Paperclip className="w-3 h-3 text-txt-g" />}
                  <span className="text-txt-2 max-w-[120px] truncate">{att.name}</span>
                  <button onClick={() => removeAttachment(i)} className="text-txt-g hover:text-red-400 cursor-pointer"><X className="w-3 h-3" /></button>
                </div>
              ))}
            </div>
          )}
          <textarea ref={textareaRef} value={text} onChange={handleChange} onKeyDown={handleKeyDown} placeholder={t("input_placeholder")} rows={2} className="w-full resize-none bg-transparent px-4 pt-3 pb-1 text-sm text-txt placeholder:text-txt-2 focus:outline-none min-h-[56px]" readOnly={isStreaming} />
          <div className="flex items-center justify-between px-2 pb-2 pt-0.5">
            <div className="flex items-center gap-0.5">
              <Dropdown icon={Route} label={t("planning_mode")} value={planningMode} options={[{ key: "auto", label: t("plan_auto") }, { key: "react", label: t("plan_react") }, { key: "plan-execute", label: t("plan_execute") }]} onChange={(v) => useSettingsStore.getState().setPlanningMode(v)} />
              <div className="w-px h-3 bg-txt-2/30 mx-0.5" />
              <Dropdown icon={Shield} label={t("safety_level")} value={permission} options={[{ key: "readonly", label: t("perm_readonly") }, { key: "write", label: t("perm_write") }, { key: "exec", label: t("perm_exec") }]} onChange={(v) => useSettingsStore.getState().setPermission(v)} />
            </div>
            <div className="flex items-center gap-0.5">
              <ModelReasoningPicker />
              <div className="w-px h-3 bg-txt-2/30 mx-0.5" />
              {isStreaming
                ? <button onClick={onCancel} className="w-7 h-7 flex items-center justify-center rounded-lg bg-red-500/15 text-red-400 hover:bg-red-500/25 transition-colors cursor-pointer" title={t("cancel_esc")}><Square className="w-3.5 h-3.5" /></button>
                : <button onClick={handleSend} disabled={!text.trim()} className="w-7 h-7 flex items-center justify-center rounded-lg bg-accent/20 text-accent hover:bg-accent/30 disabled:opacity-30 disabled:cursor-not-allowed transition-colors cursor-pointer" title={t("send_enter")}><Send className="w-3.5 h-3.5" /></button>
              }
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
```

---

## 8. hooks/useAgent.ts

Agent 事件监听。CHAT_DONE 时自动保存 session 并刷新列表。

```typescript
import { useEffect } from "react";
import { EventsOn, EVENTS } from "../lib/events";
import { useChatStore } from "../stores/chatStore";
import { useActivityStore } from "../stores/activityStore";
import { useSessionStore } from "../stores/sessionStore";
import { t } from "../lib/i18n";
import type { AgentUsage, ConfirmAction } from "../lib/types";

export function useAgent() {
  const store = useChatStore;
  const activity = useActivityStore;

  useEffect(() => {
    const unsubs: (() => void)[] = [];

    unsubs.push(
      EventsOn(EVENTS.DELTA, (...args: unknown[]) => {
        store.getState().appendDelta(args[0] as string);
      })
    );

    unsubs.push(
      EventsOn(EVENTS.THINKING, (...args: unknown[]) => {
        store.getState().appendThinking(args[0] as string);
      })
    );

    unsubs.push(
      EventsOn(EVENTS.TOOL_CALL, (...args: unknown[]) => {
        const name = args[0] as string;
        const toolArgs = args[1] as string;
        store.getState().addToolCall(name, toolArgs);
        // Add to activity log
        activity.getState().addEntry({
          type: "tool_call",
          name,
          detail: toolArgs,
          status: "running",
        });
      })
    );

    unsubs.push(
      EventsOn(EVENTS.TOOL_RESULT, (...args: unknown[]) => {
        const name = args[0] as string;
        const result = args[1] as string;
        store.getState().updateToolResult(name, result);
        // Update activity log
        activity.getState().updateEntry(name, {
          status: result.startsWith("Error") ? "error" : "done",
          detail: result,
        });
      })
    );

    unsubs.push(
      EventsOn(EVENTS.USAGE, (...args: unknown[]) => {
        store.getState().setUsage(args[0] as AgentUsage);
      })
    );

    unsubs.push(
      EventsOn(EVENTS.ERROR, (...args: unknown[]) => {
        const err = args[0] as string;
        console.error("Agent error:", err);
      })
    );

    unsubs.push(
      EventsOn(EVENTS.COMPRESSING, () => {
        store.getState().setCompressing(true);
      })
    );

    unsubs.push(
      EventsOn(EVENTS.COMPRESS_DONE, (...args: unknown[]) => {
        const data = args[0] as { before: number; after: number };
        store.getState().setCompressing(false);
        const saved = data.before - data.after;
        const pct = data.before > 0 ? Math.round((saved / data.before) * 100) : 0;
        activity.getState().addEntry({
          type: "tool_call",
          name: "compress_done",
          detail: `${t("compress_done")} ${data.before} → ${data.after} (-${saved}, -${pct}%)`,
          status: "done",
        });
      })
    );

    unsubs.push(
      EventsOn(EVENTS.PLANNING, (...args: unknown[]) => {
        const msg = args[0] as string;
        activity.getState().addEntry({
          type: "plan_step",
          name: "planning",
          detail: msg,
          status: "running",
        });
      })
    );

    unsubs.push(
      EventsOn(EVENTS.PLAN_GENERATED, (...args: unknown[]) => {
        const plan = args[0] as {
          goal: string;
          steps: Array<{
            id: number;
            description: string;
            status: string;
          }>;
          totalSteps: number;
        };
        activity.getState().setPlan({
          goal: plan.goal,
          steps: plan.steps.map((s) => ({
            id: s.id,
            description: s.description,
            status: s.status as "pending" | "in_progress" | "completed" | "failed" | "skipped",
          })),
          currentStep: 0,
          totalSteps: plan.totalSteps,
        });
      })
    );

    unsubs.push(
      EventsOn(EVENTS.PLAN_STEP_START, (...args: unknown[]) => {
        const step = args[0] as { id: number; description: string };
        activity.getState().updatePlanStep(step.id, "in_progress");
      })
    );

    unsubs.push(
      EventsOn(EVENTS.PLAN_STEP_DONE, (...args: unknown[]) => {
        const step = args[0] as { id: number; status: string };
        activity.getState().updatePlanStep(
          step.id,
          step.status === "failed" ? "failed" : "completed"
        );
      })
    );

    unsubs.push(
      EventsOn(EVENTS.CHAT_DONE, (...args: unknown[]) => {
        const data = args[0] as { response: string; duration: number };
        store.getState().finalizeResponse(data.response, data.duration);
        useSessionStore.getState().setStreamingSessionId(null);
        // Auto-save session
        try {
          const msgs = store.getState().messages;
          const sid = useSessionStore.getState().currentSessionId;
          if (sid && msgs.length > 0) {
            // workingDir is no longer passed - backend reads it from session record
            window.go?.desktop?.App?.SaveSessionFromFrontend?.(sid, msgs).then(() => {
              // Refresh session list so the new session appears in sidebar
              window.go?.desktop?.App?.ListSessions?.(30).then((list) => {
                useSessionStore.getState().setSessions(list || []);
              }).catch(console.error);
            }).catch((err) => {
              console.warn("Auto-save session failed:", err);
            });
          }
        } catch (e) {
          console.warn("Auto-save error:", e);
        }
      })
    );

    unsubs.push(
      EventsOn(EVENTS.CHAT_ERROR, (...args: unknown[]) => {
        console.error("Chat error:", args[0]);
        useSessionStore.getState().setStreamingSessionId(null);
        store.getState().finalizeResponse(`${t("error_prefix")}: ${args[0]}`, 0);
      })
    );

    unsubs.push(
      EventsOn(EVENTS.CHAT_CANCELLED, () => {
        useSessionStore.getState().setStreamingSessionId(null);
        store.getState().resetStreamState();
      })
    );

    unsubs.push(
      EventsOn(EVENTS.SAFETY_CONFIRM, (...args: unknown[]) => {
        store.getState().setConfirmAction(args[0] as ConfirmAction);
      })
    );

    return () => {
      unsubs.forEach((fn) => fn());
    };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps
}

```

---

## 9. components/layout/LeftSidebar.tsx

左侧栏。按 workspace 分组显示 session 列表。

```typescript
import { useEffect, useState, useRef } from "react";
import {
  Plus,
  MessageSquare,
  Trash2,
  FolderOpen,
  Settings,
  ChevronDown,
  ChevronRight,
  SquareCheck,
  Square,
  Pencil,
  X,
  Pin,
  PinOff,
  FolderInput,
  Edit3,
  Keyboard,
  HelpCircle,
  Info,
  Languages,
  Moon,
  Sun,
} from "lucide-react";
import { useSessionStore, type SessionItem, type WorkspaceItem } from "../../stores/sessionStore";
import { useSettingsStore } from "../../stores/settingsStore";
import { t } from "../../lib/i18n";
import { ShortcutsPanel } from "../common/ShortcutsPanel";
import { HelpLogPanel } from "../common/HelpLogPanel";
import { AboutPanel } from "../common/AboutPanel";
import { animateThemeSwitch } from "../../lib/theme-transition";

interface Props {
  onNewChat: () => void;
  onLoadSession: (id: string) => void;
  onDeleteSession: (id: string) => void;
  onExportSession?: (id: string) => Promise<void>;
  onOpenSettings: () => void;
}

interface ContextMenuState {
  x: number;
  y: number;
  sessionId?: string;
  workspaceDir?: string;
}

function formatDate(dateStr: string): string {
  try {
    const d = new Date(dateStr);
    const now = new Date();
    const isToday = d.toDateString() === now.toDateString();
    if (isToday) {
      return d.toLocaleTimeString(undefined, {
        hour: "2-digit",
        minute: "2-digit",
      });
    }
    return d.toLocaleDateString(undefined, { month: "short", day: "numeric" });
  } catch {
    return "";
  }
}

function truncate(str: string, max: number): string {
  if (!str) return "";
  return str.length > max ? str.slice(0, max) + "..." : str;
}

function getDirName(path: string): string {
  if (!path) return "其他";
  const parts = path.replace(/\\/g, "/").split("/");
  return parts[parts.length - 1] || path;
}

function ContextMenu({
  menu,
  pinned,
  onClose,
  onPin,
  onOpenExplorer,
  onRename,
  onRemove,
}: {
  menu: ContextMenuState;
  pinned: boolean;
  onClose: () => void;
  onPin: () => void;
  onOpenExplorer: () => void;
  onRename: () => void;
  onRemove: () => void;
}) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) onClose();
    };
    const handleScroll = () => onClose();
    document.addEventListener("mousedown", handleClick);
    document.addEventListener("scroll", handleScroll, true);
    return () => {
      document.removeEventListener("mousedown", handleClick);
      document.removeEventListener("scroll", handleScroll, true);
    };
  }, [onClose]);

  const items = [
    {
      icon: pinned ? PinOff : Pin,
      label: pinned ? t("unpin") : t("pin"),
      onClick: onPin,
    },
    { icon: FolderInput, label: t("open_in_explorer"), onClick: onOpenExplorer },
    ...(menu.sessionId
      ? [{ icon: Edit3, label: t("rename"), onClick: onRename }]
      : []),
    { icon: Trash2, label: t("remove_project"), onClick: onRemove, danger: true },
  ];

  return (
    <div
      ref={ref}
      className="fixed z-[100] bg-surface border border-bdr rounded-lg shadow-xl py-1 min-w-[180px]"
      style={{ left: menu.x, top: menu.y }}
    >
      {items.map((item, i) => (
        <button
          key={i}
          onClick={() => {
            item.onClick();
            onClose();
          }}
          className={`w-full flex items-center gap-2.5 px-3 py-1.5 text-sm transition-colors cursor-pointer ${
            item.danger
              ? "text-red-400 hover:bg-red-500/10"
              : "text-txt-2 hover:bg-elevated"
          }`}
        >
          <item.icon className="w-3.5 h-3.5" />
          {item.label}
        </button>
      ))}
    </div>
  );
}

function SessionItemRow({
  session,
  isActive,
  isManageMode,
  isSelected,
  onLoad,
  onDelete,
  onToggleSelect,
  onContextMenu,
}: {
  session: SessionItem;
  isActive: boolean;
  isManageMode: boolean;
  isSelected: boolean;
  onLoad: (id: string) => void;
  onDelete: (id: string) => void;
  onToggleSelect: (id: string) => void;
  onContextMenu: (e: React.MouseEvent, sessionId: string) => void;
}) {
  return (
    <div
      className={`group flex items-start gap-2 px-3 py-2 mx-2 rounded-md cursor-pointer transition-colors ${
        isActive && !isManageMode
          ? "bg-accent/10 border-l-2 border-accent"
          : "border-l-2 border-transparent hover:bg-elevated/40"
      } ${isSelected ? "bg-accent/10" : ""}`}
      onClick={() => (isManageMode ? onToggleSelect(session.id) : onLoad(session.id))}
      onContextMenu={(e) => onContextMenu(e, session.id)}
    >
      {isManageMode ? (
        <div className="flex-shrink-0 mt-0.5">
          {isSelected ? (
            <SquareCheck className="w-4 h-4 text-accent" />
          ) : (
            <Square className="w-4 h-4 text-txt-g" />
          )}
        </div>
      ) : (
        <MessageSquare className={`w-3.5 h-3.5 mt-0.5 flex-shrink-0 ${isActive ? "text-accent" : "text-txt-g"}`} />
      )}
      <div className="flex-1 min-w-0">
        <div className={`text-sm truncate ${isActive ? "text-accent" : "text-txt"}`}>
          {truncate(session.lastMessage, 35)}
        </div>
        <div className="text-[10px] text-txt-m mt-0.5">
          {formatDate(session.updatedAt)}
        </div>
      </div>
      {!isManageMode && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onDelete(session.id);
          }}
          className="opacity-0 group-hover:opacity-100 p-1 hover:text-red-400 text-txt-g transition-all cursor-pointer"
        >
          <Trash2 className="w-3 h-3" />
        </button>
      )}
    </div>
  );
}

export function LeftSidebar({
  onNewChat,
  onLoadSession,
  onDeleteSession,
  onOpenSettings,
}: Props) {
  const sessions = useSessionStore((s) => s.sessions);
  const currentSessionId = useSessionStore((s) => s.currentSessionId);
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const workspaces = useSessionStore((s) => s.workspaces);
  const [currentExpanded, setCurrentExpanded] = useState(true);
  const [otherExpanded, setOtherExpanded] = useState(false);
  const [noWsExpanded, setNoWsExpanded] = useState(true);
  const [manageMode, setManageMode] = useState(false);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [pinnedDirs, setPinnedDirs] = useState<Set<string>>(new Set());
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null);
  const [menuOpen, setMenuOpen] = useState(false);
  const [shortcutsOpen, setShortcutsOpen] = useState(false);
  const [helpLogOpen, setHelpLogOpen] = useState(false);
  const [aboutOpen, setAboutOpen] = useState(false);
  const [langPanelOpen, setLangPanelOpen] = useState(false);
  const theme = useSettingsStore((s) => s.theme);
  const language = useSettingsStore((s) => s.language);
  const setTheme = useSettingsStore((s) => s.setTheme);
  const setLanguage = useSettingsStore((s) => s.setLanguage);
  const [renameTarget, setRenameTarget] = useState<string | null>(null);
  const [renameValue, setRenameValue] = useState("");

  useEffect(() => {
  // Workspaces are loaded in App.tsx on mount



  }, []);

  // Close context menu on escape
  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setContextMenu(null);
    };
    window.addEventListener("keydown", handleKey);
    return () => window.removeEventListener("keydown", handleKey);
  }, []);

  const handleDelete = (id: string) => setDeleteTarget(id);

  const confirmDelete = () => {
    if (deleteTarget) {
      onDeleteSession(deleteTarget);
      setDeleteTarget(null);
    }
  };

  const toggleSelect = (id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const handleContextMenu = (e: React.MouseEvent, sessionId: string) => {
    e.preventDefault();
    e.stopPropagation();
    const session = sessions.find((s) => s.id === sessionId);
    setContextMenu({
      x: e.clientX,
      y: e.clientY,
      sessionId,
      workspaceDir: session?.workspaceId || "",
    });
  };

  const handleWorkspaceContextMenu = (e: React.MouseEvent, dir: string) => {
    e.preventDefault();
    e.stopPropagation();
    setContextMenu({
      x: e.clientX,
      y: e.clientY,
      workspaceDir: dir,
    });
  };

  const togglePin = (dir: string) => {
    setPinnedDirs((prev) => {
      const next = new Set(prev);
      if (next.has(dir)) next.delete(dir);
      else next.add(dir);
      return next;
    });
  };

  const deleteSelected = () => {
    for (const id of selectedIds) onDeleteSession(id);
    setSelectedIds(new Set());
    setManageMode(false);
  };

  const deleteWorkspace = (dir: string) => {
    const dirSessions = sessions.filter((s) => s.workspaceId === dir);
    for (const s of dirSessions) onDeleteSession(s.id);
  };

  const confirmRename = () => {
    if (renameTarget && renameValue.trim()) {
      window.go?.desktop?.App?.RenameSession?.(renameTarget, renameValue.trim())
        .then(() => {
          useSessionStore.getState().updateSession(renameTarget, renameValue.trim());
        })
        .catch(console.error);
    }
    setRenameTarget(null);
    setRenameValue("");
  };

  // Group sessions by workspace
  const workspaceGroups = new Map<string, SessionItem[]>();

  for (const session of sessions) {
    const wsId = session.workspaceId || "default";
    if (!workspaceGroups.has(wsId)) workspaceGroups.set(wsId, []);
    workspaceGroups.get(wsId)!.push(session);
  }

  // Build display list: default workspace first, then folder workspaces
  const defaultSessions = workspaceGroups.get("default") || [];
  workspaceGroups.delete("default");
  const otherEntries = Array.from(workspaceGroups.entries()).sort(([a], [b]) => {
    const aPinned = pinnedDirs.has(a) ? 0 : 1;
    const bPinned = pinnedDirs.has(b) ? 0 : 1;
    return aPinned - bPinned;
  });

  const getWorkspaceName = (wsId: string): string => {
    const ws = workspaces.find((w) => w.id === wsId);
    if (ws) return ws.name;
    // Fallback: extract from wsId (e.g. "ws:D:\path" -> "path")
    if (wsId.startsWith("ws:")) {
      const parts = wsId.slice(3).replace(/\\/g, "/").split("/").filter(Boolean);
      return parts[parts.length - 1] || wsId;
    }
    return wsId;
  };


  const renderSessionRow = (session: SessionItem) => (
    <SessionItemRow
      key={session.id}
      session={session}
      isActive={currentSessionId === session.id}
      isManageMode={manageMode}
      isSelected={selectedIds.has(session.id)}
      onLoad={onLoadSession}
      onDelete={handleDelete}
      onToggleSelect={toggleSelect}
      onContextMenu={handleContextMenu}
    />
  );

  const renderWorkspaceGroup = (wsId: string, dirSessions: SessionItem[], expanded: boolean, onToggle: () => void) => {
    const isPinned = pinnedDirs.has(wsId);
    const wsManage = manageMode && selectedIds.size > 0;

    return (
      <div key={wsId} className="mb-1">
        <button
          onClick={onToggle}
          onContextMenu={(e) => handleWorkspaceContextMenu(e, wsId)}
          className={`flex items-center gap-1.5 px-4 py-1.5 text-[10px] uppercase tracking-wider w-full transition-colors cursor-pointer ${
            isPinned ? "text-accent" : "text-txt-m hover:text-txt-g"
          }`}
        >
          <ChevronDown className={`w-3 h-3 transition-transform ${expanded ? "" : "-rotate-90"}`} />
          <FolderOpen className="w-3 h-3" />
          <span className="truncate">{getWorkspaceName(wsId)}</span>
          {isPinned && <Pin className="w-2.5 h-2.5 ml-0.5" />}
          {manageMode && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                const ids = dirSessions.map((s) => s.id);
                const allSelected = ids.every((id) => selectedIds.has(id));
                setSelectedIds((prev) => {
                  const next = new Set(prev);
                  if (allSelected) ids.forEach((id) => next.delete(id));
                  else ids.forEach((id) => next.add(id));
                  return next;
                });
              }}
              className="ml-auto text-txt-g hover:text-accent cursor-pointer"
            >
              {dirSessions.every((s) => selectedIds.has(s.id)) ? (
                <SquareCheck className="w-3 h-3 text-accent" />
              ) : (
                <Square className="w-3 h-3" />
              )}
            </button>
          )}
          {!manageMode && (
            <span className="ml-auto text-txt-g">{dirSessions.length}</span>
          )}
        </button>
        {expanded && dirSessions.map(renderSessionRow)}
      </div>
    );
  };

  return (
    <div className="flex flex-col h-full w-[260px]">
      {/* Header */}
      <div className="p-3 border-b border-bdr-sub space-y-2">
        <div className="flex items-center gap-2">
          <button
            onClick={onNewChat}
            className="flex-1 flex items-center gap-2 px-3 py-2 rounded-lg bg-accent/15 text-accent hover:bg-accent/25 transition-colors text-sm font-medium cursor-pointer"
          >
            <Plus className="w-4 h-4" />
            {t("new_chat")}
          </button>
          <button
            onClick={() => {
              if (manageMode) {
                setManageMode(false);
                setSelectedIds(new Set());
              } else {
                setManageMode(true);
              }
            }}
            className={`p-2 rounded-lg transition-colors cursor-pointer ${
              manageMode ? "bg-accent/20 text-accent" : "bg-elevated text-txt-g hover:text-txt"
            }`}
            title={t("manage")}
          >
            {manageMode ? <X className="w-4 h-4" /> : <Pencil className="w-4 h-4" />}
          </button>
        </div>

        {manageMode && selectedIds.size > 0 && (
          <div className="flex items-center gap-2 text-xs">
            <span className="text-txt-m">
              {t("selected_count")} {selectedIds.size}
            </span>
            <button
              onClick={deleteSelected}
              className="ml-auto flex items-center gap-1 px-2 py-1 rounded bg-red-500/20 text-red-400 hover:bg-red-500/30 transition-colors cursor-pointer"
            >
              <Trash2 className="w-3 h-3" />
              {t("delete_selected")}
            </button>
          </div>
        )}
      </div>

      {/* Session List */}
      <div className="flex-1 overflow-y-auto py-1">
        {sessions.length === 0 && (
          <div className="px-4 py-8 text-center text-txt-m text-xs">
            {t("no_sessions")}
          </div>
        )}

        {defaultSessions.length > 0 &&
          renderWorkspaceGroup("default", defaultSessions, currentExpanded, () => setCurrentExpanded(!currentExpanded))}

        

        {otherEntries.map(([wsId, dirSessions]) =>
          renderWorkspaceGroup(wsId, dirSessions, otherExpanded, () => setOtherExpanded(!otherExpanded))
        )}
      </div>

      {/* Footer ? User profile menu */}
      <UserProfileFooter onOpenSettings={onOpenSettings} />

      {/* Context Menu */}
      {contextMenu && (
        <ContextMenu
          menu={contextMenu}
          pinned={pinnedDirs.has(contextMenu.workspaceDir || "")}
          onClose={() => setContextMenu(null)}
          onPin={() => {
            if (contextMenu.workspaceDir) togglePin(contextMenu.workspaceDir);
          }}
          onOpenExplorer={() => {
            const dir = contextMenu.workspaceDir || "";
            window.go?.desktop?.App?.OpenInExplorer?.(dir).catch(console.error);
          }}
          onRename={() => {
            if (contextMenu.sessionId) {
              setRenameTarget(contextMenu.sessionId);
              const s = sessions.find((s) => s.id === contextMenu.sessionId);
              setRenameValue(s?.lastMessage || "");
            }
          }}
          onRemove={() => {
            if (contextMenu.sessionId) {
              onDeleteSession(contextMenu.sessionId);
            } else if (contextMenu.workspaceDir) {
              deleteWorkspace(contextMenu.workspaceDir);
            }
          }}
        />
      )}

      {/* Rename Dialog */}
      {renameTarget && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 backdrop-blur-sm">
          <div className="bg-surface border border-bdr rounded-xl w-[320px] mx-4 shadow-2xl">
            <div className="px-5 py-4">
              <h3 className="text-sm font-medium text-txt mb-3">{t("rename")}</h3>
              <input
                value={renameValue}
                onChange={(e) => setRenameValue(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && confirmRename()}
                className="w-full bg-elevated border border-bdr rounded-md px-3 py-1.5 text-sm text-txt focus:outline-none focus:border-accent/50"
                autoFocus
              />
            </div>
            <div className="flex items-center gap-2 px-5 py-3 border-t border-bdr-sub justify-end">
              <button
                onClick={() => { setRenameTarget(null); setRenameValue(""); }}
                className="px-3 py-1.5 rounded-md text-sm bg-elevated text-txt-2 hover:bg-elevated/80 transition-colors cursor-pointer"
              >
                {t("cancel")}
              </button>
              <button
                onClick={confirmRename}
                className="px-3 py-1.5 rounded-md text-sm bg-accent/20 text-accent hover:bg-accent/30 transition-colors cursor-pointer"
              >
                {t("save")}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Dialog */}
      {deleteTarget && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 backdrop-blur-sm">
          <div className="bg-surface border border-bdr rounded-xl w-[320px] mx-4 shadow-2xl">
            <div className="px-5 py-4">
              <h3 className="text-sm font-medium text-txt mb-2">
                {t("delete_session")}
              </h3>
              <p className="text-xs text-txt-g">{t("delete_confirm")}</p>
            </div>
            <div className="flex items-center gap-2 px-5 py-3 border-t border-bdr-sub justify-end">
              <button
                onClick={() => setDeleteTarget(null)}
                className="px-3 py-1.5 rounded-md text-sm bg-elevated text-txt-2 hover:bg-elevated/80 transition-colors cursor-pointer"
              >
                {t("cancel")}
              </button>
              <button
                onClick={confirmDelete}
                className="px-3 py-1.5 rounded-md text-sm bg-red-500/20 text-red-400 hover:bg-red-500/30 transition-colors cursor-pointer"
              >
                {t("delete")}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );


function UserProfileFooter({
  onOpenSettings,
}: {
  onOpenSettings: () => void;
}) {
  const language = useSettingsStore((s) => s.language);
  const theme = useSettingsStore((s) => s.theme);
  const setLanguage = useSettingsStore((s) => s.setLanguage);
  const setTheme = useSettingsStore((s) => s.setTheme);

  const footerRef = useRef<HTMLDivElement>(null);
  const [rawMenuOpen, setRawMenuOpen] = useState(false);
  const [shortcutsOpen, setShortcutsOpen] = useState(false);
  const [helpLogOpen, setHelpLogOpen] = useState(false);
  const [aboutOpen, setAboutOpen] = useState(false);
  const [langPanelOpen, setLangPanelOpen] = useState(false);
  const [menuPos, setMenuPos] = useState({ left: 0, bottom: 0 });

  const close = () => {
    setRawMenuOpen(false);
    setLangPanelOpen(false);
  };

  const openMenu = () => {
    if (footerRef.current) {
      const rect = footerRef.current.getBoundingClientRect();
      setMenuPos({ left: rect.left + 8, bottom: window.innerHeight - rect.top + 8 });
    }
    setRawMenuOpen(true);
  };

  // Click outside to close
  useEffect(() => {
    if (!rawMenuOpen) return;
    const handler = (e: MouseEvent) => {
      if (footerRef.current && !footerRef.current.contains(e.target as Node)) {
        close();
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [rawMenuOpen]);

  return (
    <div ref={footerRef} className="relative border-t border-bdr-sub">
      <button
        onClick={() => (rawMenuOpen ? close() : openMenu())}
        className="flex items-center gap-2.5 w-full px-3 py-2.5 rounded-md hover:bg-elevated transition-colors cursor-pointer"
      >
        <div className="w-7 h-7 rounded-full bg-accent/20 flex items-center justify-center flex-shrink-0">
          <span className="text-xs font-bold text-accent">M</span>
        </div>
        <div className="flex-1 min-w-0 text-left">
          <div className="text-xs font-medium text-txt truncate">MiMo User</div>
          <div className="text-[10px] text-txt-g truncate">{t("click_to_settings")}</div>
        </div>
        <ChevronDown
          className={`w-3.5 h-3.5 text-txt-g transition-transform ${rawMenuOpen ? "rotate-180" : ""}`}
        />
      </button>

      {rawMenuOpen && (
        <div
          className="fixed z-[100]"
          style={{ left: menuPos.left, bottom: menuPos.bottom }}
        >
          {/* Main menu 170px */}
          <div className="w-[170px] bg-surface border border-bdr rounded-lg shadow-xl px-2 py-1.5 animate-pop-up">
            {/* Settings */}
            <button
              onClick={() => { onOpenSettings(); close(); }}
              className="w-full flex items-center gap-2.5 px-2.5 py-1.5 text-xs text-txt-2 hover:bg-elevated transition-colors cursor-pointer rounded-md"
            >
              <Settings className="w-3.5 h-3.5 text-txt-g" />
              {t("settings")}
            </button>

            {/* Shortcuts */}
            <button
              onClick={() => { close(); setShortcutsOpen(true); }}
              className="w-full flex items-center gap-2.5 px-2.5 py-1.5 text-xs text-txt-2 hover:bg-elevated transition-colors cursor-pointer rounded-md"
            >
              <Keyboard className="w-3.5 h-3.5 text-txt-g" />
              {t("shortcuts")}
            </button>

            {/* Help Log */}
            <button
              onClick={() => { close(); setHelpLogOpen(true); }}
              className="w-full flex items-center gap-2.5 px-2.5 py-1.5 text-xs text-txt-2 hover:bg-elevated transition-colors cursor-pointer rounded-md"
            >
              <HelpCircle className="w-3.5 h-3.5 text-txt-g" />
              {t("help_log")}
            </button>

            {/* About */}
            <button
              onClick={() => { close(); setAboutOpen(true); }}
              className="w-full flex items-center gap-2.5 px-2.5 py-1.5 text-xs text-txt-2 hover:bg-elevated transition-colors cursor-pointer rounded-md"
            >
              <Info className="w-3.5 h-3.5 text-txt-g" />
              {t("about")}
            </button>

            {/* Divider */}
            <div className="border-t border-bdr-sub my-1 mx-1" />

            {/* Language */}
            <div
              className="relative"
              onMouseEnter={() => setLangPanelOpen(true)}
              onMouseLeave={() => setLangPanelOpen(false)}
            >
              <div className="w-full flex items-center justify-between px-2.5 py-1.5 text-xs text-txt-2 hover:bg-elevated transition-colors cursor-pointer rounded-md">
                <span className="flex items-center gap-2.5">
                  <Languages className="w-3.5 h-3.5 text-txt-g" />
                  {t("language")}
                </span>
                <span className="flex items-center gap-1 text-txt-m">
                  {language === "zh" ? t("chinese") : t("english")}
                  <ChevronRight className="w-3 h-3" />
                </span>
              </div>

              {langPanelOpen && (
                <div className="absolute left-full bottom-0 w-[120px] bg-surface border border-bdr rounded-lg shadow-xl px-2 py-1.5 animate-pop-up">
                  <div className="px-2.5 py-1 text-[10px] text-txt-g uppercase tracking-wider">
                    {t("language")}
                  </div>
                  <button
                    onClick={() => { setLanguage("zh"); close(); }}
                    className={`w-full flex items-center justify-between px-2.5 py-1.5 text-xs rounded-md transition-colors cursor-pointer ${
                      language === "zh" ? "text-accent bg-accent/10" : "text-txt-2 hover:bg-elevated"
                    }`}
                  >
                    {t("chinese")}
                    {language === "zh" && <span className="text-accent">&#10003;</span>}
                  </button>
                  <button
                    onClick={() => { setLanguage("en"); close(); }}
                    className={`w-full flex items-center justify-between px-2.5 py-1.5 text-xs rounded-md transition-colors cursor-pointer ${
                      language === "en" ? "text-accent bg-accent/10" : "text-txt-2 hover:bg-elevated"
                    }`}
                  >
                    {t("english")}
                    {language === "en" && <span className="text-accent">&#10003;</span>}
                  </button>
                </div>
              )}
            </div>

            {/* Theme toggle */}
            <div className="w-full flex items-center justify-between px-2.5 py-1.5 text-xs text-txt-2 rounded-md">
              <span className="flex items-center gap-2.5">
                {theme === "dark" ? (
                  <Moon className="w-3.5 h-3.5 text-txt-g" />
                ) : (
                  <Sun className="w-3.5 h-3.5 text-txt-g" />
                )}
                {t("theme")}
              </span>
              <button
                onClick={(e) => setTheme(theme === "dark" ? "light" : "dark", e.clientX, e.clientY)}
                className={`relative w-9 h-5 rounded-full transition-colors duration-200 cursor-pointer flex-shrink-0 ${
                  theme === "dark" ? "bg-accent" : "bg-elevated border border-bdr"
                }`}
              >
                <div
                  className={`absolute top-0.5 w-4 h-4 rounded-full bg-white shadow-md transition-transform duration-200 ${
                    theme === "dark" ? "translate-x-4" : "translate-x-0.5"
                  }`}
                />
              </button>
            </div>
          </div>
        </div>
      )}

      <ShortcutsPanel open={shortcutsOpen} onClose={() => setShortcutsOpen(false)} />
      <HelpLogPanel open={helpLogOpen} onClose={() => setHelpLogOpen(false)} />
      <AboutPanel open={aboutOpen} onClose={() => setAboutOpen(false)} />
    </div>
  );
}
}

```
