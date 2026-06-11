package memory

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SearchResult represents a memory search result
type SearchResult struct {
	Path    string  `json:"path"`
	Snippet string  `json:"snippet"`
	Score   float64 `json:"score"`
	Scope   string  `json:"scope"`
	ScopeID string  `json:"scope_id"`
	Type    string  `json:"type"`
}

// SearchOpts contains options for memory search
type SearchOpts struct {
	Scope   string `json:"scope"`
	ScopeID string `json:"scope_id"`
	Type    string `json:"type"`
	Limit   int    `json:"limit"`
}

// Service manages the memory system
type Service struct {
	db    *sql.DB
	paths *Paths
}

// NewService creates a new memory service
func NewService(db *sql.DB, projectDir, sessionID string) *Service {
	paths := NewPaths(projectDir, sessionID)
	return &Service{
		db:    db,
		paths: paths,
	}
}

// Init initializes the memory service and creates necessary tables
func (s *Service) Init() error {
	return s.createTables()
}

// createTables creates the memory_fts table and FTS5 virtual table
func (s *Service) createTables() error {
	// Create the main memory_fts table
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS memory_fts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT NOT NULL,
			scope TEXT NOT NULL DEFAULT 'project',
			scope_id TEXT DEFAULT '',
			type TEXT DEFAULT '',
			body TEXT NOT NULL,
			created_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("create memory_fts table: %w", err)
	}

	// Create FTS5 virtual table for full-text search
	// Note: FTS5 table creation is done separately to handle existing tables
	s.db.Exec(`DROP TABLE IF EXISTS memory_fts_idx`)
	_, err = s.db.Exec(`
		CREATE VIRTUAL TABLE IF NOT EXISTS memory_fts_idx USING fts5(
			body, path, scope, scope_id, type,
			content=memory_fts,
			content_rowid=id
		)
	`)
	if err != nil {
		// FTS5 might not be available, log warning but don't fail
		fmt.Fprintf(os.Stderr, "Warning: FTS5 not available: %v\n", err)
	}

	return nil
}

// Search performs a full-text search on memory
func (s *Service) Search(query string, opts SearchOpts) ([]SearchResult, error) {
	if opts.Limit <= 0 {
		opts.Limit = 10
	}

	ftsQuery := BuildFtsQuery(query)
	if ftsQuery == "" {
		return nil, nil
	}

	// Build WHERE clauses for filtering
	var conditions []string
	var params []interface{}

	if opts.Scope != "" {
		conditions = append(conditions, "scope = ?")
		params = append(params, opts.Scope)
	}
	if opts.ScopeID != "" {
		conditions = append(conditions, "scope_id = ?")
		params = append(params, opts.ScopeID)
	}
	if opts.Type != "" {
		conditions = append(conditions, "type = ?")
		params = append(params, opts.Type)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "AND " + strings.Join(conditions, " AND ")
	}

	// Try FTS5 search first
	results, err := s.ftsSearch(ftsQuery, whereClause, params, opts.Limit)
	if err == nil && len(results) > 0 {
		return results, nil
	}

	// Fallback to LIKE search if FTS5 is not available or returns no results
	return s.likeSearch(query, opts)
}

// ftsSearch performs FTS5 search
func (s *Service) ftsSearch(ftsQuery, whereClause string, params []interface{}, limit int) ([]SearchResult, error) {
	fetchLimit := limit * 3
	if fetchLimit > 50 {
		fetchLimit = 50
	}

	sql := fmt.Sprintf(`
		SELECT path, scope, scope_id, type,
			   snippet(memory_fts_idx, 0, '<<', '>>', '...', 32) AS snippet,
			   bm25(memory_fts_idx) AS score
		FROM memory_fts_idx
		JOIN memory_fts ON memory_fts.id = memory_fts_idx.rowid
		WHERE memory_fts_idx MATCH ?
		%s
		ORDER BY score
		LIMIT ?
	`, whereClause)

	allParams := append([]interface{}{ftsQuery}, params...)
	allParams = append(allParams, fetchLimit)

	rows, err := s.db.Query(sql, allParams...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.Path, &r.Scope, &r.ScopeID, &r.Type, &r.Snippet, &r.Score); err != nil {
			continue
		}
		// FTS5 bm25() returns lower = better, convert to higher = better
		r.Score = -r.Score
		results = append(results, r)
	}

	// Apply score floor (keep top result and those within 15% of top score)
	if len(results) > 1 {
		topScore := results[0].Score
		cutoff := topScore * 0.15
		var filtered []SearchResult
		for i, r := range results {
			if i == 0 || r.Score >= cutoff {
				filtered = append(filtered, r)
			}
		}
		results = filtered
	}

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// likeSearch performs a LIKE-based search as fallback
func (s *Service) likeSearch(query string, opts SearchOpts) ([]SearchResult, error) {
	var conditions []string
	var params []interface{}

	// Add text search condition
	conditions = append(conditions, "body LIKE ?")
	params = append(params, "%"+query+"%")

	if opts.Scope != "" {
		conditions = append(conditions, "scope = ?")
		params = append(params, opts.Scope)
	}
	if opts.ScopeID != "" {
		conditions = append(conditions, "scope_id = ?")
		params = append(params, opts.ScopeID)
	}
	if opts.Type != "" {
		conditions = append(conditions, "type = ?")
		params = append(params, opts.Type)
	}

	whereClause := strings.Join(conditions, " AND ")
	sql := fmt.Sprintf(`
		SELECT path, scope, scope_id, type, body
		FROM memory_fts
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ?
	`, whereClause)

	params = append(params, opts.Limit)
	rows, err := s.db.Query(sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var body string
		if err := rows.Scan(&r.Path, &r.Scope, &r.ScopeID, &r.Type, &body); err != nil {
			continue
		}
		r.Snippet = ExtractSnippet(body, query, 64)
		r.Score = 1.0 // Default score for LIKE search
		results = append(results, r)
	}

	return results, nil
}

// IndexFile indexes a markdown file into the memory system
func (s *Service) IndexFile(path, scope, scopeID, memType string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	body := string(data)

	// Check if file already exists
	var existingID int
	err = s.db.QueryRow("SELECT id FROM memory_fts WHERE path = ? AND scope = ?", path, scope).
		Scan(&existingID)

	if err == nil {
		// Update existing entry
		_, err = s.db.Exec(
			"UPDATE memory_fts SET body = ?, type = ?, created_at = ? WHERE id = ?",
			body, memType, time.Now().Unix(), existingID,
		)
	} else {
		// Insert new entry
		_, err = s.db.Exec(
			"INSERT INTO memory_fts (path, scope, scope_id, type, body, created_at) VALUES (?, ?, ?, ?, ?, ?)",
			path, scope, scopeID, memType, body, time.Now().Unix(),
		)
	}

	if err != nil {
		return err
	}

	// Update FTS index
	s.db.Exec("DELETE FROM memory_fts_idx WHERE rowid = ?", existingID)
	s.db.Exec("INSERT INTO memory_fts_idx(rowid, body, path, scope, scope_id, type) SELECT id, body, path, scope, scope_id, type FROM memory_fts WHERE path = ? AND scope = ?", path, scope)

	return nil
}

// RemoveFile removes a file from the memory index
func (s *Service) RemoveFile(path, scope string) error {
	var id int
	err := s.db.QueryRow("SELECT id FROM memory_fts WHERE path = ? AND scope = ?", path, scope).Scan(&id)
	if err != nil {
		return nil // Not found, nothing to remove
	}

	_, err = s.db.Exec("DELETE FROM memory_fts WHERE id = ?", id)
	if err != nil {
		return err
	}

	s.db.Exec("DELETE FROM memory_fts_idx WHERE rowid = ?", id)
	return nil
}

// Reconcile scans memory directories and updates the index
func (s *Service) Reconcile() (indexed int, pruned int, err error) {
	// Ensure directories exist
	if err := s.paths.EnsureDirectories(); err != nil {
		return 0, 0, err
	}

	// Index global memory
	if n, err := s.reconcileDir(s.paths.GlobalDir, "global", ""); err == nil {
		indexed += n
	}

	// Index project memory
	if n, err := s.reconcileDir(s.paths.ProjectDir, "project", ""); err == nil {
		indexed += n
	}

	// Index session memory
	if s.paths.SessionDir != "" {
		if n, err := s.reconcileDir(s.paths.SessionDir, "session", filepath.Base(s.paths.SessionDir)); err == nil {
			indexed += n
		}
	}

	// Prune deleted files
	pruned, err = s.pruneDeleted()

	return
}

// reconcileDir indexes all markdown files in a directory
func (s *Service) reconcileDir(dir, scope, scopeID string) (int, error) {
	var count int

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Determine file type from name
		memType := determineType(path)

		if err := s.IndexFile(path, scope, scopeID, memType); err == nil {
			count++
		}

		return nil
	})

	return count, err
}

// pruneDeleted removes index entries for files that no longer exist
func (s *Service) pruneDeleted() (int, error) {
	rows, err := s.db.Query("SELECT id, path FROM memory_fts")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var toDelete []int
	for rows.Next() {
		var id int
		var path string
		if err := rows.Scan(&id, &path); err != nil {
			continue
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			toDelete = append(toDelete, id)
		}
	}

	for _, id := range toDelete {
		s.db.Exec("DELETE FROM memory_fts WHERE id = ?", id)
		s.db.Exec("DELETE FROM memory_fts_idx WHERE rowid = ?", id)
	}

	return len(toDelete), nil
}

// determineType determines the memory type from the file path
func determineType(path string) string {
	base := filepath.Base(path)
	switch {
	case base == "MEMORY.md":
		return "project"
	case base == "checkpoint.md":
		return "checkpoint"
	case base == "notes.md":
		return "notes"
	case strings.HasSuffix(base, "progress.md"):
		return "progress"
	default:
		return "free"
	}
}

// Count returns the total number of indexed memory entries
func (s *Service) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM memory_fts").Scan(&count)
	return count, err
}
