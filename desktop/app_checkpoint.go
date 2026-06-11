package desktop

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mimo-cli/mimo-cli/internal/context"
	"github.com/mimo-cli/mimo-cli/internal/llm"
	"github.com/mimo-cli/mimo-cli/internal/session"
)

// CheckpointResult represents the result of a checkpoint operation
type CheckpointResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	ID      string `json:"id,omitempty"`
}

// CheckpointInfo represents checkpoint information for the frontend
type CheckpointInfo struct {
	ID            string    `json:"id"`
	Summary       string    `json:"summary"`
	TokenCount    int       `json:"token_count"`
	MessageOffset int       `json:"message_offset"`
	CreatedAt     time.Time `json:"created_at"`
}

// CreateCheckpoint creates a checkpoint for the current session
func (a *App) CreateCheckpoint(summary string) CheckpointResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.currentSessionID == "" {
		return CheckpointResult{Success: false, Message: "No active session"}
	}

	if a.sessionStore == nil {
		return CheckpointResult{Success: false, Message: "Session store not initialized"}
	}

	// Get current message count
	msgCount, err := a.sessionStore.CountMessages(a.currentSessionID)
	if err != nil {
		return CheckpointResult{Success: false, Message: fmt.Sprintf("Failed to count messages: %v", err)}
	}

	// Estimate token count from messages
	tokenCount := msgCount * 50 // rough estimate

	// Create checkpoint
	cp := &session.Checkpoint{
		ID:            fmt.Sprintf("cp_%d", time.Now().UnixNano()),
		SessionID:     a.currentSessionID,
		Summary:       summary,
		MessageOffset: msgCount,
		TokenCount:    tokenCount,
		Metadata:      "{}",
		CreatedAt:     time.Now(),
	}

	if err := a.sessionStore.SaveCheckpoint(cp); err != nil {
		return CheckpointResult{Success: false, Message: fmt.Sprintf("Failed to save checkpoint: %v", err)}
	}

	// Also create a checkpoint file in the session memory directory
	wd, _ := os.Getwd()
	checkpointMgr := context.NewCheckpointManager(
		context.DefaultCheckpointConfig(),
		wd,
	)
	sessionDir := filepath.Join(wd, ".mimo", "memory", "sessions", a.currentSessionID)

	state := &context.CheckpointState{
		SessionID:  a.currentSessionID,
		Summary:    summary,
		TokenCount: tokenCount,
		CreatedAt:  time.Now(),
	}

	if _, err := checkpointMgr.CreateCheckpointFile(state, sessionDir); err != nil {
		// Log warning but don't fail - SQLite checkpoint is primary
		fmt.Fprintf(os.Stderr, "Warning: failed to create checkpoint file: %v\n", err)
	}

	return CheckpointResult{
		Success: true,
		Message: "Checkpoint created successfully",
		ID:      cp.ID,
	}
}

// ListCheckpoints returns all checkpoints for the current session
func (a *App) ListCheckpoints() []CheckpointInfo {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.currentSessionID == "" || a.sessionStore == nil {
		return []CheckpointInfo{}
	}

	checkpoints, err := a.sessionStore.ListCheckpoints(a.currentSessionID)
	if err != nil {
		return []CheckpointInfo{}
	}

	var result []CheckpointInfo
	for _, cp := range checkpoints {
		result = append(result, CheckpointInfo{
			ID:            cp.ID,
			Summary:       cp.Summary,
			TokenCount:    cp.TokenCount,
			MessageOffset: cp.MessageOffset,
			CreatedAt:     cp.CreatedAt,
		})
	}
	return result
}

// RestoreCheckpoint restores context from a checkpoint
func (a *App) RestoreCheckpoint(checkpointID string) CheckpointResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.currentSessionID == "" {
		return CheckpointResult{Success: false, Message: "No active session"}
	}

	if a.sessionStore == nil {
		return CheckpointResult{Success: false, Message: "Session store not initialized"}
	}

	// Load the checkpoint
	cp, err := a.sessionStore.LoadCheckpoint(checkpointID)
	if err != nil {
		return CheckpointResult{Success: false, Message: fmt.Sprintf("Failed to load checkpoint: %v", err)}
	}

	// Load messages from the checkpoint offset
	messages, err := a.sessionStore.LoadMessagesFromOffset(cp.SessionID, cp.MessageOffset)
	if err != nil {
		return CheckpointResult{Success: false, Message: fmt.Sprintf("Failed to load messages: %v", err)}
	}

	// Convert to agent messages and load them into the active agent context.
	agentMessages := []llm.Message{{
		Role: llm.RoleSystem,
		Content: fmt.Sprintf(
			"Restored checkpoint summary for session %s:\n%s",
			cp.SessionID,
			cp.Summary,
		),
	}}
	for _, msg := range messages {
		agentMessages = append(agentMessages, llm.Message{
			Role:    llm.Role(msg.Role),
			Content: msg.Content,
		})
	}
	if a.agent != nil {
		a.agent.LoadMessages(agentMessages)
	}

	return CheckpointResult{
		Success: true,
		Message: fmt.Sprintf("Checkpoint restored. Context rebuilt from checkpoint with %d messages.", len(messages)),
		ID:      cp.ID,
	}
}

// DeleteCheckpoint deletes a checkpoint
func (a *App) DeleteCheckpoint(checkpointID string) CheckpointResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.sessionStore == nil {
		return CheckpointResult{Success: false, Message: "Session store not initialized"}
	}

	if err := a.sessionStore.DeleteCheckpoint(checkpointID); err != nil {
		return CheckpointResult{Success: false, Message: fmt.Sprintf("Failed to delete checkpoint: %v", err)}
	}

	return CheckpointResult{
		Success: true,
		Message: "Checkpoint deleted successfully",
	}
}

// GetCheckpointSummary returns a summary of the checkpoint for display
func (a *App) GetCheckpointSummary(checkpointID string) string {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.sessionStore == nil {
		return ""
	}

	cp, err := a.sessionStore.LoadCheckpoint(checkpointID)
	if err != nil {
		return ""
	}

	return cp.Summary
}

// ExportCheckpoints exports all checkpoints for the current session as JSON
func (a *App) ExportCheckpoints() string {
	checkpoints := a.ListCheckpoints()
	data, err := json.MarshalIndent(checkpoints, "", "  ")
	if err != nil {
		return "[]"
	}
	return string(data)
}
