package memory

import (
	"os"
	"path/filepath"
	"runtime"
)

// Paths manages memory file paths
type Paths struct {
	// Global memory directory (~/.mimo/memory/)
	GlobalDir string
	// Project memory directory (<project>/.mimo/memory/)
	ProjectDir string
	// Session memory directory (<project>/.mimo/sessions/<sid>/)
	SessionDir string
}

// NewPaths creates a Paths instance
func NewPaths(projectDir, sessionID string) *Paths {
	home, _ := os.UserHomeDir()

	globalDir := filepath.Join(home, ".mimo", "memory")
	projectDir = filepath.Join(projectDir, ".mimo", "memory")
	sessionDir := ""
	if sessionID != "" {
		sessionDir = filepath.Join(projectDir, "sessions", sessionID)
	}

	return &Paths{
		GlobalDir:  globalDir,
		ProjectDir: projectDir,
		SessionDir: sessionDir,
	}
}

// EnsureDirectories creates all memory directories
func (p *Paths) EnsureDirectories() error {
	dirs := []string{p.GlobalDir, p.ProjectDir}
	if p.SessionDir != "" {
		dirs = append(dirs, p.SessionDir)
		dirs = append(dirs, filepath.Join(p.SessionDir, "tasks"))
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// MemoryFilePath returns the path to MEMORY.md for a given scope
func (p *Paths) MemoryFilePath(scope string) string {
	switch scope {
	case "global":
		return filepath.Join(p.GlobalDir, "MEMORY.md")
	case "project", "projects":
		return filepath.Join(p.ProjectDir, "MEMORY.md")
	default:
		return filepath.Join(p.ProjectDir, "MEMORY.md")
	}
}

// CheckpointFilePath returns the path to checkpoint.md for a session
func (p *Paths) CheckpointFilePath() string {
	if p.SessionDir == "" {
		return ""
	}
	return filepath.Join(p.SessionDir, "checkpoint.md")
}

// NotesFilePath returns the path to notes.md for a session
func (p *Paths) NotesFilePath() string {
	if p.SessionDir == "" {
		return ""
	}
	return filepath.Join(p.SessionDir, "notes.md")
}

// TaskProgressFilePath returns the path to a task's progress.md
func (p *Paths) TaskProgressFilePath(taskID string) string {
	if p.SessionDir == "" {
		return ""
	}
	return filepath.Join(p.SessionDir, "tasks", taskID, "progress.md")
}

// MemoryDir returns the directory for memory files
func (p *Paths) MemoryDir() string {
	if runtime.GOOS == "windows" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".mimo", "memory")
	}
	return filepath.Join(p.ProjectDir)
}
