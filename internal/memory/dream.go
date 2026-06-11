package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DreamConfig holds dream configuration
type DreamConfig struct {
	Enabled           bool `yaml:"enabled"`
	ExtractOnComplete bool `yaml:"extract_on_complete"`
	MaxEntries        int  `yaml:"max_entries"`
}

// DefaultDreamConfig returns default dream configuration
func DefaultDreamConfig() DreamConfig {
	return DreamConfig{
		Enabled:           true,
		ExtractOnComplete: true,
		MaxEntries:        100,
	}
}

// DreamEntry represents a memory entry extracted from a session
type DreamEntry struct {
	Type      string    `json:"type"`      // "decision", "bugfix", "preference", "knowledge"
	Summary   string    `json:"summary"`
	Context   string    `json:"context,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Dream extracts persistent knowledge from session history
type Dream struct {
	config    DreamConfig
	projectDir string
}

// NewDream creates a new dream instance
func NewDream(config DreamConfig, projectDir string) *Dream {
	return &Dream{
		config:     config,
		projectDir: projectDir,
	}
}

// Extract extracts memory entries from session data
func (d *Dream) Extract(sessionData string) ([]DreamEntry, error) {
	if !d.config.Enabled {
		return nil, nil
	}

	var entries []DreamEntry

	// Extract decisions
	decisions := d.extractDecisions(sessionData)
	entries = append(entries, decisions...)

	// Extract bug fixes
	bugfixes := d.extractBugFixes(sessionData)
	entries = append(entries, bugfixes...)

	// Extract preferences
	preferences := d.extractPreferences(sessionData)
	entries = append(entries, preferences...)

	// Extract knowledge
	knowledge := d.extractKnowledge(sessionData)
	entries = append(entries, knowledge...)

	// Limit entries
	if len(entries) > d.config.MaxEntries {
		entries = entries[:d.config.MaxEntries]
	}

	return entries, nil
}

// extractDecisions extracts architectural decisions from session data
func (d *Dream) extractDecisions(data string) []DreamEntry {
	var entries []DreamEntry
	lines := strings.Split(data, "\n")

	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "decision") ||
			strings.Contains(lower, "decided to") ||
			strings.Contains(lower, "chose to") ||
			strings.Contains(lower, "will use") {
			entries = append(entries, DreamEntry{
				Type:      "decision",
				Summary:   strings.TrimSpace(line),
				CreatedAt: time.Now(),
			})
		}
	}
	return entries
}

// extractBugFixes extracts bug fix information from session data
func (d *Dream) extractBugFixes(data string) []DreamEntry {
	var entries []DreamEntry
	lines := strings.Split(data, "\n")

	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "fixed") ||
			strings.Contains(lower, "bug fix") ||
			strings.Contains(lower, "resolved") ||
			strings.Contains(lower, "patch") {
			entries = append(entries, DreamEntry{
				Type:      "bugfix",
				Summary:   strings.TrimSpace(line),
				CreatedAt: time.Now(),
			})
		}
	}
	return entries
}

// extractPreferences extracts user preferences from session data
func (d *Dream) extractPreferences(data string) []DreamEntry {
	var entries []DreamEntry
	lines := strings.Split(data, "\n")

	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "prefer") ||
			strings.Contains(lower, "like to") ||
			strings.Contains(lower, "always use") ||
			strings.Contains(lower, "never use") {
			entries = append(entries, DreamEntry{
				Type:      "preference",
				Summary:   strings.TrimSpace(line),
				CreatedAt: time.Now(),
			})
		}
	}
	return entries
}

// extractKnowledge extracts general knowledge from session data
func (d *Dream) extractKnowledge(data string) []DreamEntry {
	var entries []DreamEntry
	lines := strings.Split(data, "\n")

	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "note:") ||
			strings.Contains(lower, "important:") ||
			strings.Contains(lower, "remember:") ||
			strings.Contains(lower, "learned that") {
			entries = append(entries, DreamEntry{
				Type:      "knowledge",
				Summary:   strings.TrimSpace(line),
				CreatedAt: time.Now(),
			})
		}
	}
	return entries
}

// SaveToMemory saves extracted entries to MEMORY.md
func (d *Dream) SaveToMemory(entries []DreamEntry) error {
	if len(entries) == 0 {
		return nil
	}

	memoryDir := filepath.Join(d.projectDir, ".mimo", "memory")
	if err := os.MkdirAll(memoryDir, 0755); err != nil {
		return fmt.Errorf("failed to create memory dir: %w", err)
	}

	memoryFile := filepath.Join(memoryDir, "MEMORY.md")

	// Read existing content
	var existingContent string
	if data, err := os.ReadFile(memoryFile); err == nil {
		existingContent = string(data)
	}

	// Build new content
	var content strings.Builder
	content.WriteString("# Project Memory\n\n")
	content.WriteString(fmt.Sprintf("Last updated: %s\n\n", time.Now().Format(time.RFC3339)))

	// Group entries by type
	byType := make(map[string][]DreamEntry)
	for _, entry := range entries {
		byType[entry.Type] = append(byType[entry.Type], entry)
	}

	// Write decisions
	if decisions, ok := byType["decision"]; ok {
		content.WriteString("## Decisions\n\n")
		for _, entry := range decisions {
			content.WriteString(fmt.Sprintf("- %s\n", entry.Summary))
		}
		content.WriteString("\n")
	}

	// Write bug fixes
	if bugfixes, ok := byType["bugfix"]; ok {
		content.WriteString("## Bug Fixes\n\n")
		for _, entry := range bugfixes {
			content.WriteString(fmt.Sprintf("- %s\n", entry.Summary))
		}
		content.WriteString("\n")
	}

	// Write preferences
	if preferences, ok := byType["preference"]; ok {
		content.WriteString("## Preferences\n\n")
		for _, entry := range preferences {
			content.WriteString(fmt.Sprintf("- %s\n", entry.Summary))
		}
		content.WriteString("\n")
	}

	// Write knowledge
	if knowledge, ok := byType["knowledge"]; ok {
		content.WriteString("## Knowledge\n\n")
		for _, entry := range knowledge {
			content.WriteString(fmt.Sprintf("- %s\n", entry.Summary))
		}
		content.WriteString("\n")
	}

	// Append to existing content if any
	if existingContent != "" {
		content.WriteString("\n---\n\n")
		content.WriteString(existingContent)
	}

	return os.WriteFile(memoryFile, []byte(content.String()), 0644)
}

// Run executes the dream process on session history
func (d *Dream) Run(sessionDir string) (int, error) {
	// Read session files
	sessionFile := filepath.Join(sessionDir, "checkpoint.md")
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to read session: %w", err)
	}

	// Extract entries
	entries, err := d.Extract(string(data))
	if err != nil {
		return 0, err
	}

	// Save to memory
	if err := d.SaveToMemory(entries); err != nil {
		return 0, err
	}

	return len(entries), nil
}
