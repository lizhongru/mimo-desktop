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

// Session represents a conversation session
type Session struct {
	ID          string
	ModelName   string
	UserName    string
	LastMessage string
	WorkingDir  string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Message represents a single message in a session
type Message struct {
	ID          int64
	SessionID   string
	Role        string
	Content     string
	Tokens      int
	ToolCalls   int
	DurationMs  int64
	Thinking    string
	ToolLines   []string
	CreatedAt   time.Time
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

	// Enable WAL mode for better concurrent access
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("cannot set WAL mode: %w", err)
	}

	// Enable foreign key constraints for CASCADE deletes
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
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			model_name TEXT NOT NULL DEFAULT '',
			user_name TEXT NOT NULL DEFAULT '',
			last_message TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
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
	`)
	if err != nil {
		return err
	}
	// Add columns for existing databases (migration v2)
	db.Exec("ALTER TABLE messages ADD COLUMN thinking TEXT NOT NULL DEFAULT ''")
	db.Exec("ALTER TABLE messages ADD COLUMN tool_lines TEXT NOT NULL DEFAULT '[]'")
	// Migration v3: add working_dir to sessions
	db.Exec("ALTER TABLE sessions ADD COLUMN working_dir TEXT NOT NULL DEFAULT ''")
	return nil
}

// SaveSession creates or updates a session with its messages
func (s *Store) SaveSession(id, modelName, userName, workingDir string, messages []Message) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Find last user message
	lastMsg := ""
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			lastMsg = messages[i].Content
			if len(lastMsg) > 200 {
				lastMsg = lastMsg[:200]
			}
			break
		}
	}

	// Upsert session
	_, err = tx.Exec(`
		INSERT INTO sessions (id, model_name, user_name, last_message, working_dir, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			model_name = excluded.model_name,
			user_name = excluded.user_name,
			last_message = excluded.last_message,
			working_dir = excluded.working_dir,
			updated_at = excluded.updated_at
	`, id, modelName, userName, lastMsg, workingDir, time.Now(), time.Now())
	if err != nil {
		return err
	}

	// Delete old messages and re-insert
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

	return tx.Commit()
}

// LoadSession loads a session and its messages by ID
func (s *Store) LoadSession(id string) (*Session, []Message, error) {
	// Load session
	sess := &Session{}
	err := s.db.QueryRow(`
		SELECT id, model_name, user_name, last_message, working_dir, created_at, updated_at
		FROM sessions WHERE id = ?
	`, id).Scan(&sess.ID, &sess.ModelName, &sess.UserName, &sess.LastMessage, &sess.WorkingDir, &sess.CreatedAt, &sess.UpdatedAt)
	if err != nil {
		return nil, nil, err
	}

	// Load messages
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

// ListSessions returns all sessions, most recent first
func (s *Store) ListSessions(limit int) ([]Session, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(`
		SELECT id, model_name, user_name, last_message, working_dir, created_at, updated_at
		FROM sessions
		ORDER BY updated_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var sess Session
		if err := rows.Scan(&sess.ID, &sess.ModelName, &sess.UserName, &sess.LastMessage, &sess.WorkingDir, &sess.CreatedAt, &sess.UpdatedAt); err != nil {
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

	// Explicitly delete messages first (in case foreign_keys is not enabled)
	if _, err := tx.Exec("DELETE FROM messages WHERE session_id = ?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM sessions WHERE id = ?", id); err != nil {
		return err
	}
	return tx.Commit()
}

// RenameSession updates the last_message (display title) of a session.
func (s *Store) RenameSession(id, title string) error {
	_, err := s.db.Exec("UPDATE sessions SET last_message = ?, updated_at = ? WHERE id = ?", title, time.Now(), id)
	return err
}

// CountMessages returns the number of messages in a session
func (s *Store) CountMessages(sessionID string) (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM messages WHERE session_id = ?", sessionID).Scan(&count)
	return count, err
}
