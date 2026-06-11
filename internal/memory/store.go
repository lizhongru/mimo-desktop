package memory

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Store manages memory persistence with SQLite
type Store struct {
	db *sql.DB
}

// MemoryEntry represents a memory entry
type MemoryEntry struct {
	ID        int64
	Path      string
	Scope     string
	ScopeID   string
	Type      string
	Body      string
	CreatedAt time.Time
}

// NewStore creates a new memory store
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// Init creates the memory tables
func (s *Store) Init() error {
	_, err := s.db.Exec(`
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
	if err != nil {
		return fmt.Errorf("create memory table: %w", err)
	}

	// Create index for faster lookups
	s.db.Exec("CREATE INDEX IF NOT EXISTS idx_memory_scope ON memory(scope)")
	s.db.Exec("CREATE INDEX IF NOT EXISTS idx_memory_path ON memory(path)")

	return nil
}

// Upsert inserts or updates a memory entry
func (s *Store) Upsert(entry *MemoryEntry) error {
	_, err := s.db.Exec(`
		INSERT INTO memory (path, scope, scope_id, type, body, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(path, scope) DO UPDATE SET
			scope_id = excluded.scope_id,
			type = excluded.type,
			body = excluded.body,
			created_at = excluded.created_at
	`, entry.Path, entry.Scope, entry.ScopeID, entry.Type, entry.Body, entry.CreatedAt.Unix())
	return err
}

// Get retrieves a memory entry by path and scope
func (s *Store) Get(path, scope string) (*MemoryEntry, error) {
	entry := &MemoryEntry{}
	var createdAt int64
	err := s.db.QueryRow(
		"SELECT id, path, scope, scope_id, type, body, created_at FROM memory WHERE path = ? AND scope = ?",
		path, scope,
	).Scan(&entry.ID, &entry.Path, &entry.Scope, &entry.ScopeID, &entry.Type, &entry.Body, &createdAt)
	if err != nil {
		return nil, err
	}
	entry.CreatedAt = time.Unix(createdAt, 0)
	return entry, nil
}

// List returns all memory entries for a scope
func (s *Store) List(scope string) ([]MemoryEntry, error) {
	rows, err := s.db.Query(
		"SELECT id, path, scope, scope_id, type, body, created_at FROM memory WHERE scope = ? ORDER BY created_at DESC",
		scope,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []MemoryEntry
	for rows.Next() {
		var entry MemoryEntry
		var createdAt int64
		if err := rows.Scan(&entry.ID, &entry.Path, &entry.Scope, &entry.ScopeID, &entry.Type, &entry.Body, &createdAt); err != nil {
			return nil, err
		}
		entry.CreatedAt = time.Unix(createdAt, 0)
		entries = append(entries, entry)
	}
	return entries, nil
}

// Delete removes a memory entry
func (s *Store) Delete(path, scope string) error {
	_, err := s.db.Exec("DELETE FROM memory WHERE path = ? AND scope = ?", path, scope)
	return err
}

// Search performs a text search on memory entries
func (s *Store) Search(query string, scope string, limit int) ([]MemoryEntry, error) {
	if limit <= 0 {
		limit = 10
	}

	// Use LIKE for basic text search
	sqlQuery := "SELECT id, path, scope, scope_id, type, body, created_at FROM memory WHERE body LIKE ?"
	params := []interface{}{"%" + query + "%"}

	if scope != "" {
		sqlQuery += " AND scope = ?"
		params = append(params, scope)
	}

	sqlQuery += " ORDER BY created_at DESC LIMIT ?"
	params = append(params, limit)

	rows, err := s.db.Query(sqlQuery, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []MemoryEntry
	for rows.Next() {
		var entry MemoryEntry
		var createdAt int64
		if err := rows.Scan(&entry.ID, &entry.Path, &entry.Scope, &entry.ScopeID, &entry.Type, &entry.Body, &createdAt); err != nil {
			return nil, err
		}
		entry.CreatedAt = time.Unix(createdAt, 0)
		entries = append(entries, entry)
	}
	return entries, nil
}

// Count returns the total number of memory entries
func (s *Store) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM memory").Scan(&count)
	return count, err
}

// IndexFile indexes a markdown file into the store
func (s *Store) IndexFile(path, scope, scopeID, memType string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	body := string(data)

	entry := &MemoryEntry{
		Path:      path,
		Scope:     scope,
		ScopeID:   scopeID,
		Type:      memType,
		Body:      body,
		CreatedAt: time.Now(),
	}

	return s.Upsert(entry)
}

// Reconcile scans a directory and indexes all markdown files
func (s *Store) Reconcile(dir, scope, scopeID string) (int, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return 0, nil
	}

	var count int
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		memType := determineType(path)
		if err := s.IndexFile(path, scope, scopeID, memType); err == nil {
			count++
		}

		return nil
	})

	return count, err
}
