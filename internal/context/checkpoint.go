package context

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CheckpointConfig holds checkpoint configuration
type CheckpointConfig struct {
	AutoCheckpoint       bool    `yaml:"auto_checkpoint"`
	TokenThreshold       float64 `yaml:"token_threshold"`       // auto-checkpoint when context reaches this ratio
	MaxCheckpoints       int     `yaml:"max_checkpoints"`       // max checkpoints per session
	ReconstructOnResume  bool    `yaml:"reconstruct_on_resume"` // auto-reconstruct context on session resume
	ContextBudget        int     `yaml:"context_budget"`        // token budget for context reconstruction
}

// DefaultCheckpointConfig returns default checkpoint configuration
func DefaultCheckpointConfig() CheckpointConfig {
	return CheckpointConfig{
		AutoCheckpoint:      true,
		TokenThreshold:      0.75, // 75% of max tokens
		MaxCheckpoints:      10,
		ReconstructOnResume: true,
		ContextBudget:       128000,
	}
}

// CheckpointManager manages checkpoint creation and context reconstruction
type CheckpointManager struct {
	config    CheckpointConfig
	projectDir string
}

// NewCheckpointManager creates a new checkpoint manager
func NewCheckpointManager(config CheckpointConfig, projectDir string) *CheckpointManager {
	return &CheckpointManager{
		config:     config,
		projectDir: projectDir,
	}
}

// CheckpointState represents the state to be checkpointed
type CheckpointState struct {
	SessionID     string            `json:"session_id"`
	Messages      []MessageSnapshot `json:"messages"`
	Summary       string            `json:"summary"`
	ProjectInfo   *ProjectInfo      `json:"project_info,omitempty"`
	TokenCount    int               `json:"token_count"`
	CreatedAt     time.Time         `json:"created_at"`
}

// MessageSnapshot is a lightweight message representation for checkpointing
type MessageSnapshot struct {
	Role      string   `json:"role"`
	Content   string   `json:"content"`
	ToolCalls []string `json:"tool_calls,omitempty"`
}

// ShouldCheckpoint determines if a checkpoint should be created
func (m *CheckpointManager) ShouldCheckpoint(currentTokens, maxTokens int) bool {
	if !m.config.AutoCheckpoint {
		return false
	}
	ratio := float64(currentTokens) / float64(maxTokens)
	return ratio >= m.config.TokenThreshold
}

// CreateCheckpointFile creates a checkpoint markdown file in the session memory directory
func (m *CheckpointManager) CreateCheckpointFile(state *CheckpointState, sessionDir string) (string, error) {
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return "", fmt.Errorf("cannot create session dir: %w", err)
	}

	checkpointPath := filepath.Join(sessionDir, "checkpoint.md")

	// Build checkpoint content
	content := m.buildCheckpointContent(state)

	if err := os.WriteFile(checkpointPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("cannot write checkpoint: %w", err)
	}

	return checkpointPath, nil
}

// buildCheckpointContent builds the checkpoint markdown content
func (m *CheckpointManager) buildCheckpointContent(state *CheckpointState) string {
	var lines []string

	lines = append(lines, "# Session Checkpoint")
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("**Session ID**: %s", state.SessionID))
	lines = append(lines, fmt.Sprintf("**Created At**: %s", state.CreatedAt.Format(time.RFC3339)))
	lines = append(lines, fmt.Sprintf("**Token Count**: %d", state.TokenCount))
	lines = append(lines, "")

	if state.Summary != "" {
		lines = append(lines, "## Summary")
		lines = append(lines, "")
		lines = append(lines, state.Summary)
		lines = append(lines, "")
	}

	if state.ProjectInfo != nil {
		lines = append(lines, "## Project Context")
		lines = append(lines, "")
		if state.ProjectInfo.ProjectType != "" {
			lines = append(lines, fmt.Sprintf("- **Project Type**: %s", state.ProjectInfo.ProjectType))
		}
		if state.ProjectInfo.GitBranch != "" {
			lines = append(lines, fmt.Sprintf("- **Git Branch**: %s", state.ProjectInfo.GitBranch))
		}
		lines = append(lines, "")
	}

	if len(state.Messages) > 0 {
		lines = append(lines, "## Recent Messages")
		lines = append(lines, "")
		// Only include last 10 messages in checkpoint
		start := 0
		if len(state.Messages) > 10 {
			start = len(state.Messages) - 10
		}
		for _, msg := range state.Messages[start:] {
			lines = append(lines, fmt.Sprintf("**%s**: %s", msg.Role, truncateString(msg.Content, 200)))
			lines = append(lines, "")
		}
	}

	return joinLines(lines)
}

// RebuildContextFromCheckpoint rebuilds context from a checkpoint file
func (m *CheckpointManager) RebuildContextFromCheckpoint(checkpointPath string) (*CheckpointState, error) {
	data, err := os.ReadFile(checkpointPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read checkpoint: %w", err)
	}

	state := &CheckpointState{
		Messages: []MessageSnapshot{},
	}

	// Parse the markdown checkpoint
	content := string(data)
	lines := splitLines(content)

	var currentSection string
	var summaryLines []string

	for _, line := range lines {
		if line == "## Summary" {
			currentSection = "summary"
			continue
		}
		if line == "## Project Context" {
			currentSection = "project"
			continue
		}
		if line == "## Recent Messages" {
			currentSection = "messages"
			continue
		}

		switch currentSection {
		case "summary":
			if line != "" {
				summaryLines = append(summaryLines, line)
			}
		}
	}

	if len(summaryLines) > 0 {
		state.Summary = joinLines(summaryLines)
	}

	return state, nil
}

// SerializeState serializes checkpoint state to JSON
func (m *CheckpointManager) SerializeState(state *CheckpointState) (string, error) {
	data, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DeserializeState deserializes checkpoint state from JSON
func (m *CheckpointManager) DeserializeState(data string) (*CheckpointState, error) {
	state := &CheckpointState{}
	if err := json.Unmarshal([]byte(data), state); err != nil {
		return nil, err
	}
	return state, nil
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// splitLines splits a string into lines
func splitLines(s string) []string {
	result := []string{}
	current := []byte{}
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			result = append(result, string(current))
			current = []byte{}
		} else {
			current = append(current, s[i])
		}
	}
	if len(current) > 0 {
		result = append(result, string(current))
	}
	return result
}

// joinLines joins lines with newlines
func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}
