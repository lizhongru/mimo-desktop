package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Entry represents a single backup
type Entry struct {
	ID          int       `json:"id"`
	OrigPath    string    `json:"orig_path"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	BackupFile  string    `json:"backup_file"` // relative to backupDir
}

// manifest is the on-disk structure for backups.json
type manifest struct {
	NextID  int     `json:"next_id"`
	Entries []Entry `json:"entries"`
}

// Manager handles file backups for rollback support
type Manager struct {
	mu        sync.Mutex
	backupDir string
}

// NewManager creates a backup manager
func NewManager(backupDir string) *Manager {
	if backupDir == "" {
		backupDir = ".mimo/backups"
	}
	return &Manager{backupDir: backupDir}
}

// Backup creates a backup of the given file before modification.
// Returns the backup ID. If the file doesn't exist (new file), returns 0.
func (m *Manager) Backup(origPath, description string) (int, error) {
	if _, err := os.Stat(origPath); os.IsNotExist(err) {
		return 0, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err := os.MkdirAll(m.backupDir, 0755); err != nil {
		return 0, fmt.Errorf("failed to create backup dir: %w", err)
	}

	mf, err := m.loadManifest()
	if err != nil {
		return 0, err
	}

	// Resolve to absolute path for reliable matching later
	absPath, err := filepath.Abs(origPath)
	if err != nil {
		absPath = origPath
	}

	mf.NextID++
	id := mf.NextID
	backupFile := fmt.Sprintf("%d.bak", id)
	backupPath := filepath.Join(m.backupDir, backupFile)

	// Read and write backup
	data, err := os.ReadFile(origPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read original: %w", err)
	}
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return 0, fmt.Errorf("failed to write backup: %w", err)
	}

	entry := Entry{
		ID:          id,
		OrigPath:    absPath,
		Description: description,
		Timestamp:   time.Now(),
		BackupFile:  backupFile,
	}
	mf.Entries = append(mf.Entries, entry)

	if err := m.saveManifest(mf); err != nil {
		return 0, err
	}
	return id, nil
}

// List returns all backups, sorted by time (newest first)
func (m *Manager) List() ([]Entry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	mf, err := m.loadManifest()
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, len(mf.Entries))
	copy(entries, mf.Entries)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})
	return entries, nil
}

// Restore restores a file from a backup by ID
func (m *Manager) Restore(id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	mf, err := m.loadManifest()
	if err != nil {
		return err
	}

	for _, e := range mf.Entries {
		if e.ID == id {
			backupPath := filepath.Join(m.backupDir, e.BackupFile)
			data, err := os.ReadFile(backupPath)
			if err != nil {
				return fmt.Errorf("failed to read backup: %w", err)
			}
			parent := filepath.Dir(e.OrigPath)
			if err := os.MkdirAll(parent, 0755); err != nil {
				return fmt.Errorf("failed to create dir: %w", err)
			}
			return os.WriteFile(e.OrigPath, data, 0644)
		}
	}
	return fmt.Errorf("backup #%d not found", id)
}

// loadManifest reads the manifest file (creates empty if missing)
func (m *Manager) loadManifest() (*manifest, error) {
	path := filepath.Join(m.backupDir, "manifest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &manifest{NextID: 0}, nil
		}
		return nil, err
	}
	var mf manifest
	if err := json.Unmarshal(data, &mf); err != nil {
		return &manifest{NextID: 0}, nil
	}
	return &mf, nil
}

// saveManifest writes the manifest file
func (m *Manager) saveManifest(mf *manifest) error {
	path := filepath.Join(m.backupDir, "manifest.json")
	data, err := json.MarshalIndent(mf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
