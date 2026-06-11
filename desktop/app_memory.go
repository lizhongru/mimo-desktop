package desktop

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mimo-cli/mimo-cli/internal/memory"
)

// MemorySearchResult is the frontend-friendly memory search result
type MemorySearchResult struct {
	Path    string  `json:"path"`
	Snippet string  `json:"snippet"`
	Score   float64 `json:"score"`
	Scope   string  `json:"scope"`
	ScopeID string  `json:"scope_id"`
	Type    string  `json:"type"`
}

// MemoryService returns the memory service (lazy init)
func (a *App) memoryService() *memory.Service {
	if a.sessionStore == nil {
		return nil
	}
	if a.memorySvc == nil {
		wd, _ := os.Getwd()
		a.memorySvc = memory.NewService(a.sessionStore.DB(), wd, a.currentSessionID)
		a.memorySvc.Init()
	}
	return a.memorySvc
}

// MemorySearch searches the memory system
func (a *App) MemorySearch(query string, scope string, limit int) ([]MemorySearchResult, error) {
	svc := a.memoryService()
	if svc == nil {
		return nil, fmt.Errorf("memory service not initialized")
	}

	// Reconcile before search to ensure index is up to date
	svc.Reconcile()

	results, err := svc.Search(query, memory.SearchOpts{
		Scope: scope,
		Limit: limit,
	})
	if err != nil {
		return nil, fmt.Errorf("memory search failed: %w", err)
	}

	// Convert to frontend-friendly format
	searchResults := make([]MemorySearchResult, len(results))
	for i, r := range results {
		searchResults[i] = MemorySearchResult{
			Path:    r.Path,
			Snippet: r.Snippet,
			Score:   r.Score,
			Scope:   r.Scope,
			ScopeID: r.ScopeID,
			Type:    r.Type,
		}
	}

	return searchResults, nil
}

// MemoryReconcile re-indexes all memory files
func (a *App) MemoryReconcile() (int, int) {
	svc := a.memoryService()
	if svc == nil {
		return 0, 0
	}
	indexed, pruned, _ := svc.Reconcile()
	return indexed, pruned
}

// MemoryCount returns the total number of indexed memory entries
func (a *App) MemoryCount() int {
	svc := a.memoryService()
	if svc == nil {
		return 0
	}
	count, _ := svc.Count()
	return count
}

// MemoryIndexFile indexes a specific file into memory
func (a *App) MemoryIndexFile(path, scope, scopeID, memType string) error {
	svc := a.memoryService()
	if svc == nil {
		return fmt.Errorf("memory service not initialized")
	}
	return svc.IndexFile(path, scope, scopeID, memType)
}

// WriteMemory writes content to a memory file
func (a *App) WriteMemory(scope, content string) error {
	wd, _ := os.Getwd()
	paths := memory.NewPaths(wd, a.currentSessionID)

	// Ensure directory exists
	if err := paths.EnsureDirectories(); err != nil {
		return err
	}

	// Get the file path for the scope
	var filePath string
	switch scope {
	case "global":
		filePath = paths.MemoryFilePath("global")
	case "project", "projects":
		filePath = paths.MemoryFilePath("project")
	case "session":
		filePath = paths.CheckpointFilePath()
	case "notes":
		filePath = paths.NotesFilePath()
	default:
		return fmt.Errorf("unknown scope: %s", scope)
	}

	if filePath == "" {
		return fmt.Errorf("no file path for scope: %s", scope)
	}

	// Ensure parent directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write the file
	return os.WriteFile(filePath, []byte(content), 0644)
}

// ReadMemory reads content from a memory file
func (a *App) ReadMemory(scope string) (string, error) {
	wd, _ := os.Getwd()
	paths := memory.NewPaths(wd, a.currentSessionID)

	var filePath string
	switch scope {
	case "global":
		filePath = paths.MemoryFilePath("global")
	case "project", "projects":
		filePath = paths.MemoryFilePath("project")
	case "session":
		filePath = paths.CheckpointFilePath()
	case "notes":
		filePath = paths.NotesFilePath()
	default:
		return "", fmt.Errorf("unknown scope: %s", scope)
	}

	if filePath == "" {
		return "", fmt.Errorf("no file path for scope: %s", scope)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // Empty file
		}
		return "", err
	}

	return string(data), nil
}

// ListMemoryFiles lists all memory files in the project
func (a *App) ListMemoryFiles() []MemoryFileInfo {
	wd, _ := os.Getwd()
	paths := memory.NewPaths(wd, a.currentSessionID)

	var files []MemoryFileInfo

	// List project memory files
	if _, err := os.Stat(paths.ProjectDir); err == nil {
		filepath.Walk(paths.ProjectDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			if filepath.Ext(path) == ".md" {
				stat, _ := os.Stat(path)
				files = append(files, MemoryFileInfo{
					Path:      path,
					Name:      info.Name(),
					Size:      info.Size(),
					UpdatedAt: stat.ModTime(),
					Scope:     "project",
				})
			}
			return nil
		})
	}

	// List session memory files
	if paths.SessionDir != "" {
		if _, err := os.Stat(paths.SessionDir); err == nil {
			filepath.Walk(paths.SessionDir, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				if filepath.Ext(path) == ".md" {
					stat, _ := os.Stat(path)
					files = append(files, MemoryFileInfo{
						Path:      path,
						Name:      info.Name(),
						Size:      info.Size(),
						UpdatedAt: stat.ModTime(),
						Scope:     "session",
					})
				}
				return nil
			})
		}
	}

	return files
}

// MemoryFileInfo describes a memory file
type MemoryFileInfo struct {
	Path      string    `json:"path"`
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	UpdatedAt time.Time `json:"updatedAt"`
	Scope     string    `json:"scope"`
}
