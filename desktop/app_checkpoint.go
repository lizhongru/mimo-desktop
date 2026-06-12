package desktop

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	iconfig "github.com/mimo-cli/mimo-cli/internal/config"
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
	sessionID := a.currentSessionID
	a.mu.Unlock()

	if sessionID == "" {
		return CheckpointResult{Success: false, Message: "No active session"}
	}
	return a.createCheckpointForSession(sessionID, summary, "manual")
}

func (a *App) createCheckpointForSession(sessionID string, summary string, source string) CheckpointResult {
	if a.sessionStore == nil {
		return CheckpointResult{Success: false, Message: "Session store not initialized"}
	}

	_, messages, err := a.sessionStore.LoadSession(sessionID)
	if err != nil {
		return CheckpointResult{Success: false, Message: fmt.Sprintf("Failed to load session messages: %v", err)}
	}

	messageOffset := len(messages)
	tokenCount := estimateCheckpointTokens(messages)
	if strings.TrimSpace(summary) == "" {
		summary = buildAutoCheckpointSummary(messages)
	}

	cp := &session.Checkpoint{
		ID:            fmt.Sprintf("cp_%d", time.Now().UnixNano()),
		SessionID:     sessionID,
		Summary:       summary,
		MessageOffset: messageOffset,
		TokenCount:    tokenCount,
		Metadata:      fmt.Sprintf(`{"source":%q}`, source),
		CreatedAt:     time.Now(),
	}

	if err := a.sessionStore.SaveCheckpoint(cp); err != nil {
		return CheckpointResult{Success: false, Message: fmt.Sprintf("Failed to save checkpoint: %v", err)}
	}

	// Also create a checkpoint file in the session memory directory
	wd, _ := os.Getwd()
	checkpointMgr := context.NewCheckpointManager(a.checkpointRuntimeConfig(), wd)
	sessionDir := filepath.Join(wd, ".mimo", "memory", "sessions", sessionID)

	state := &context.CheckpointState{
		SessionID:  sessionID,
		Messages:   checkpointMessageSnapshots(messages),
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

func (a *App) maybeCreateAutoCheckpoint(sessionID string) error {
	if sessionID == "" || a.sessionStore == nil {
		return nil
	}

	checkpointCfg := a.checkpointRuntimeConfig()
	maxTokens := checkpointCfg.ContextBudget

	_, messages, err := a.sessionStore.LoadSession(sessionID)
	if err != nil {
		return err
	}
	if len(messages) == 0 {
		return nil
	}

	currentTokens := estimateCheckpointTokens(messages)
	checkpointMgr := context.NewCheckpointManager(checkpointCfg, "")
	if !checkpointMgr.ShouldCheckpoint(currentTokens, maxTokens) {
		return nil
	}

	latest, err := a.sessionStore.GetLatestCheckpoint(sessionID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if latest != nil && latest.MessageOffset >= len(messages) {
		return nil
	}

	result := a.createCheckpointForSession(sessionID, buildAutoCheckpointSummary(messages), "auto")
	if !result.Success {
		return fmt.Errorf("%s", result.Message)
	}

	return a.pruneCheckpoints(sessionID, checkpointCfg.MaxCheckpoints)
}

func (a *App) checkpointRuntimeConfig() context.CheckpointConfig {
	cfg := context.DefaultCheckpointConfig()
	if a.cfg == nil {
		return cfg
	}

	if checkpointConfigIsSet(a.cfg.Checkpoint) {
		checkpointCfg := a.cfg.Checkpoint
		cfg.AutoCheckpoint = checkpointCfg.AutoCheckpoint
		cfg.ReconstructOnResume = checkpointCfg.ReconstructOnResume
		if checkpointCfg.TokenThreshold > 0 {
			cfg.TokenThreshold = checkpointCfg.TokenThreshold
		}
		if checkpointCfg.MaxCheckpoints > 0 {
			cfg.MaxCheckpoints = checkpointCfg.MaxCheckpoints
		}
		if checkpointCfg.ContextBudget > 0 {
			cfg.ContextBudget = checkpointCfg.ContextBudget
		}
	} else if a.cfg.Context.MaxTokens > 0 {
		cfg.ContextBudget = a.cfg.Context.MaxTokens
	}
	return cfg
}

func checkpointConfigIsSet(cfg iconfig.CheckpointConfig) bool {
	return cfg.AutoCheckpoint ||
		cfg.TokenThreshold > 0 ||
		cfg.MaxCheckpoints > 0 ||
		cfg.ReconstructOnResume ||
		cfg.ContextBudget > 0
}

func (a *App) pruneCheckpoints(sessionID string, maxCheckpoints int) error {
	if maxCheckpoints <= 0 || a.sessionStore == nil {
		return nil
	}
	checkpoints, err := a.sessionStore.ListCheckpoints(sessionID)
	if err != nil {
		return err
	}
	for i := maxCheckpoints; i < len(checkpoints); i++ {
		if err := a.sessionStore.DeleteCheckpoint(checkpoints[i].ID); err != nil {
			return err
		}
	}
	return nil
}

func estimateCheckpointTokens(messages []session.Message) int {
	total := 0
	for _, msg := range messages {
		if msg.Tokens > 0 {
			total += msg.Tokens
		} else {
			total += len(msg.Content) / 3
		}
		total += msg.ToolCalls * 10
		for _, line := range msg.ToolLines {
			total += len(line) / 3
		}
		total += 10
	}
	return total
}

func buildAutoCheckpointSummary(messages []session.Message) string {
	latestUserMessage := ""
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			latestUserMessage = truncateCheckpointText(messages[i].Content, 160)
			break
		}
	}
	if latestUserMessage == "" {
		return fmt.Sprintf("自动检查点：已保存 %d 条消息。", len(messages))
	}
	return fmt.Sprintf("自动检查点：已保存 %d 条消息。最近用户消息：%s", len(messages), latestUserMessage)
}

func checkpointMessageSnapshots(messages []session.Message) []context.MessageSnapshot {
	snapshots := make([]context.MessageSnapshot, 0, len(messages))
	for _, msg := range messages {
		snapshots = append(snapshots, context.MessageSnapshot{
			Role:      msg.Role,
			Content:   msg.Content,
			ToolCalls: msg.ToolLines,
		})
	}
	return snapshots
}

func truncateCheckpointText(text string, maxLen int) string {
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen]) + "..."
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
