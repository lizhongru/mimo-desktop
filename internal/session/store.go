package session

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Store manages session persistence with SQLite
type Store struct {
	db *sql.DB
}

// DB returns the underlying database connection
func (s *Store) DB() *sql.DB {
	return s.db
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

	// Repair rows affected by an older CreateSession placeholder bug that wrote
	// user_name into created_at, making LoadSession fail when scanning time.Time.
	db.Exec(`UPDATE sessions SET updated_at = CURRENT_TIMESTAMP WHERE updated_at IS NULL OR updated_at = '' OR updated_at NOT GLOB '????-??-??*'`)
	db.Exec(`UPDATE sessions SET created_at = updated_at WHERE created_at IS NULL OR created_at = '' OR created_at NOT GLOB '????-??-??*'`)

	// --- Migration v4: Memory system tables ---
	db.Exec(`
		CREATE TABLE IF NOT EXISTS memory (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT NOT NULL,
			scope TEXT NOT NULL DEFAULT 'project',
			scope_id TEXT DEFAULT '',
			type TEXT DEFAULT '',
			body TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			UNIQUE(path, scope)
		)
	`)
	db.Exec("CREATE INDEX IF NOT EXISTS idx_memory_scope ON memory(scope)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_memory_path ON memory(path)")

	// --- Migration v5: Task system tables ---
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
			cleanup_after INTEGER,
			FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
		)
	`)
	db.Exec(`
		CREATE TABLE IF NOT EXISTS task_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id TEXT NOT NULL,
			at INTEGER NOT NULL,
			kind TEXT NOT NULL,
			summary TEXT,
			FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
		)
	`)
	db.Exec("CREATE INDEX IF NOT EXISTS idx_tasks_session ON tasks(session_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_task_events_task ON task_events(task_id)")

	// --- Migration v6: Checkpoint system tables ---
	db.Exec(`
		CREATE TABLE IF NOT EXISTS checkpoints (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			summary TEXT NOT NULL DEFAULT '',
			message_offset INTEGER NOT NULL DEFAULT 0,
			token_count INTEGER NOT NULL DEFAULT 0,
			metadata TEXT NOT NULL DEFAULT '{}',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
		)
	`)
	db.Exec("CREATE INDEX IF NOT EXISTS idx_checkpoints_session ON checkpoints(session_id)")

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

func normalizeWorkspaceID(workspaceID string) string {
	if workspaceID == "" {
		return DefaultWorkspaceID
	}
	return workspaceID
}

// WorkspacePathFromID returns the filesystem path encoded in a folder workspace ID.
func WorkspacePathFromID(workspaceID string) string {
	if strings.HasPrefix(workspaceID, "ws:") {
		return strings.TrimPrefix(workspaceID, "ws:")
	}
	return ""
}

func workspaceNameFromPath(path string) string {
	name := filepath.Base(path)
	if name == "" || name == "." || name == string(filepath.Separator) {
		return path
	}
	return name
}

func (s *Store) ensureWorkspaceForID(workspaceID string) error {
	workspaceID = normalizeWorkspaceID(workspaceID)
	if workspaceID == DefaultWorkspaceID {
		_, err := s.db.Exec(
			`INSERT OR IGNORE INTO workspaces (id, name, type, path) VALUES (?, 'default', 'chat', '')`,
			DefaultWorkspaceID,
		)
		return err
	}

	if strings.HasPrefix(workspaceID, "ws:") {
		path := WorkspacePathFromID(workspaceID)
		_, err := s.db.Exec(
			`INSERT OR IGNORE INTO workspaces (id, name, type, path) VALUES (?, ?, 'folder', ?)`,
			workspaceID,
			workspaceNameFromPath(path),
			path,
		)
		return err
	}

	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO workspaces (id, name, type, path) VALUES (?, ?, 'chat', '')`,
		workspaceID,
		workspaceID,
	)
	return err
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
	workspaceID = normalizeWorkspaceID(workspaceID)
	if err := s.ensureWorkspaceForID(workspaceID); err != nil {
		return err
	}
	_, err := s.db.Exec(`
		INSERT INTO sessions (id, workspace_id, model_name, user_name, last_message, created_at, updated_at)
		VALUES (?, ?, ?, ?, '', ?, ?)
	`, id, workspaceID, modelName, userName, time.Now(), time.Now())
	return err
}

// SaveSession creates or updates a session with its messages
func (s *Store) SaveSession(id, workspaceID, modelName, userName string, messages []Message) error {
	workspaceID = normalizeWorkspaceID(workspaceID)
	if err := s.ensureWorkspaceForID(workspaceID); err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

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
	workspaceID = normalizeWorkspaceID(workspaceID)
	if err := s.ensureWorkspaceForID(workspaceID); err != nil {
		return err
	}
	_, err := s.db.Exec("UPDATE sessions SET workspace_id = ?, updated_at = ? WHERE id = ?", workspaceID, time.Now(), sessionID)
	return err
}

// CountMessages returns the number of messages in a session
func (s *Store) CountMessages(sessionID string) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM messages WHERE session_id = ?", sessionID).Scan(&count)
	return count, err
}

// Checkpoint represents a session checkpoint for context reconstruction
type Checkpoint struct {
	ID            string    `json:"id"`
	SessionID     string    `json:"session_id"`
	Summary       string    `json:"summary"`
	MessageOffset int       `json:"message_offset"`
	TokenCount    int       `json:"token_count"`
	Metadata      string    `json:"metadata"`
	CreatedAt     time.Time `json:"created_at"`
}

// SaveCheckpoint creates or updates a checkpoint
func (s *Store) SaveCheckpoint(cp *Checkpoint) error {
	if cp.ID == "" {
		cp.ID = fmt.Sprintf("cp_%d", time.Now().UnixNano())
	}
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO checkpoints (id, session_id, summary, message_offset, token_count, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, cp.ID, cp.SessionID, cp.Summary, cp.MessageOffset, cp.TokenCount, cp.Metadata, cp.CreatedAt)
	return err
}

// LoadCheckpoint loads a checkpoint by ID
func (s *Store) LoadCheckpoint(id string) (*Checkpoint, error) {
	cp := &Checkpoint{}
	err := s.db.QueryRow(`
		SELECT id, session_id, summary, message_offset, token_count, metadata, created_at
		FROM checkpoints WHERE id = ?
	`, id).Scan(&cp.ID, &cp.SessionID, &cp.Summary, &cp.MessageOffset, &cp.TokenCount, &cp.Metadata, &cp.CreatedAt)
	if err != nil {
		return nil, err
	}
	return cp, nil
}

// ListCheckpoints returns all checkpoints for a session, ordered by created_at DESC
func (s *Store) ListCheckpoints(sessionID string) ([]Checkpoint, error) {
	rows, err := s.db.Query(`
		SELECT id, session_id, summary, message_offset, token_count, metadata, created_at
		FROM checkpoints WHERE session_id = ? ORDER BY created_at DESC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checkpoints []Checkpoint
	for rows.Next() {
		var cp Checkpoint
		if err := rows.Scan(&cp.ID, &cp.SessionID, &cp.Summary, &cp.MessageOffset, &cp.TokenCount, &cp.Metadata, &cp.CreatedAt); err != nil {
			return nil, err
		}
		checkpoints = append(checkpoints, cp)
	}
	return checkpoints, rows.Err()
}

// GetLatestCheckpoint returns the most recent checkpoint for a session
func (s *Store) GetLatestCheckpoint(sessionID string) (*Checkpoint, error) {
	cp := &Checkpoint{}
	err := s.db.QueryRow(`
		SELECT id, session_id, summary, message_offset, token_count, metadata, created_at
		FROM checkpoints WHERE session_id = ? ORDER BY created_at DESC LIMIT 1
	`, sessionID).Scan(&cp.ID, &cp.SessionID, &cp.Summary, &cp.MessageOffset, &cp.TokenCount, &cp.Metadata, &cp.CreatedAt)
	if err != nil {
		return nil, err
	}
	return cp, nil
}

// DeleteCheckpoint deletes a checkpoint by ID
func (s *Store) DeleteCheckpoint(id string) error {
	_, err := s.db.Exec("DELETE FROM checkpoints WHERE id = ?", id)
	return err
}

// LoadMessagesFromOffset loads messages after a message-count offset.
func (s *Store) LoadMessagesFromOffset(sessionID string, offset int) ([]Message, error) {
	if offset < 0 {
		offset = 0
	}

	rows, err := s.db.Query(`
		SELECT id, session_id, role, content, tokens, tool_calls, duration_ms, thinking, tool_lines, created_at
		FROM messages WHERE session_id = ? ORDER BY id ASC LIMIT -1 OFFSET ?
	`, sessionID, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var toolLinesJSON string
		if err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.Tokens, &msg.ToolCalls, &msg.DurationMs, &msg.Thinking, &toolLinesJSON, &msg.CreatedAt); err != nil {
			return nil, err
		}
		if toolLinesJSON != "" {
			json.Unmarshal([]byte(toolLinesJSON), &msg.ToolLines)
		}
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}
