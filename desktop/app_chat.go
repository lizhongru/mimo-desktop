package desktop

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/mimo-cli/mimo-cli/internal/llm"
	"github.com/mimo-cli/mimo-cli/internal/tools"
)

// appVersion and appCommit are set at build time via -ldflags.
var (
	appVersion = "dev"
	appCommit  = "unknown"
)

// cancelFunc holds the cancel function for the current SendMessage goroutine.
// Stored on App so CancelOperation can cancel the context.

// SendMessage sends a user message and starts streaming the agent response.
// attachmentsJSON is an optional JSON array of {name, type, dataUrl} objects.
func (a *App) SendMessage(message string, attachmentsJSON string) error {
	a.mu.Lock()
	if a.isBusy {
		a.mu.Unlock()
		return fmt.Errorf("agent is busy")
	}
	a.isBusy = true
	// Reset confirmAll for new message
	a.confirmAll = false
	a.mu.Unlock()

	runtime.EventsEmit(a.ctx, EventChatStart, message)

	ctx, cancel := context.WithCancel(context.Background())
	workingDir := a.currentSessionWorkingDir()
	if config := getMultiAgentManager().GetCurrent(); config != nil {
		a.agent.SetToolAllowlist(config.ToolAllowlist)
	}
	a.agent.SystemPrompt(a.buildSystemPrompt(workingDir))
	if workingDir != "" {
		ctx = tools.WithWorkingDir(ctx, workingDir)
	}
	a.mu.Lock()
	a.cancelChat = cancel
	a.mu.Unlock()

	go func() {
		defer func() {
			cancel()
			a.mu.Lock()
			a.isBusy = false
			a.cancelChat = nil
			a.mu.Unlock()
		}()

		// Parse attachments if provided
		var attachments []llm.Attachment
		if attachmentsJSON != "" {
			if err := json.Unmarshal([]byte(attachmentsJSON), &attachments); err != nil {
				runtime.EventsEmit(a.ctx, EventChatError, fmt.Sprintf("failed to parse attachments: %v", err))
				return
			}
		}

		start := time.Now()
		response, err := a.agent.ChatStream(ctx, message, attachments)
		duration := time.Since(start)

		if err != nil {
			runtime.EventsEmit(a.ctx, EventChatError, err.Error())
			return
		}

		runtime.EventsEmit(a.ctx, EventChatDone, map[string]interface{}{
			"response": response,
			"duration": duration.Milliseconds(),
		})
	}()

	return nil
}

// CancelOperation cancels the current agent operation.
func (a *App) CancelOperation() {
	a.agent.Cancel()
	a.mu.Lock()
	if a.cancelChat != nil {
		a.cancelChat()
	}
	a.isBusy = false
	a.mu.Unlock()
	runtime.EventsEmit(a.ctx, EventChatCancelled)
}

// IsBusy returns whether the agent is currently processing.
func (a *App) IsBusy() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.isBusy
}

// RespondToConfirm responds to a safety confirmation dialog.
func (a *App) RespondToConfirm(approved bool) {
	select {
	case a.confirmChan <- approved:
	case <-time.After(30 * time.Second):
		// Timed out waiting for receiver; avoid permanent block.
	}
}

// RespondToConfirmAll responds to a safety confirmation and sets confirm-all mode.
func (a *App) RespondToConfirmAll(approved bool) {
	a.mu.Lock()
	a.confirmAll = true
	a.mu.Unlock()
	a.agent.SetConfirmAll(true)
	select {
	case a.confirmChan <- approved:
	case <-time.After(30 * time.Second):
		// Timed out waiting for receiver; avoid permanent block.
	}
}

// GetModelName returns the current model display name.
func (a *App) GetModelName() string {
	return a.gateway.GetCurrentModel()
}

// GetVersion returns the app version info.
func (a *App) GetVersion() map[string]string {
	return map[string]string{
		"version": appVersion,
		"commit":  appCommit,
	}
}

// CompressContext manually triggers context compression.
func (a *App) CompressContext() (map[string]int, error) {
	before, after, err := a.agent.CompressContext(context.Background())
	if err != nil {
		return nil, err
	}
	runtime.EventsEmit(a.ctx, EventCompressDone, map[string]interface{}{
		"before": before,
		"after":  after,
	})
	return map[string]int{"before": before, "after": after}, nil
}

// ExportMessage is a frontend-friendly message for export.
type ExportMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ExportChat exports chat messages to a Markdown file via save dialog.
func (a *App) ExportChat(messages []ExportMessage) error {
	if len(messages) == 0 {
		return fmt.Errorf("no messages to export")
	}

	var sb strings.Builder
	sb.WriteString("# MiMo Chat Export\n\n")
	sb.WriteString(fmt.Sprintf("_Exported: %s_\n\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString("---\n\n")

	for _, msg := range messages {
		role := "\U0001F916 Assistant"
		if msg.Role == "user" {
			role = "\U0001F464 User"
		}
		sb.WriteString(fmt.Sprintf("## %s\n\n%s\n\n", role, msg.Content))
	}

	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export Chat",
		DefaultFilename: fmt.Sprintf("chat_%s.md", time.Now().Format("20060102_150405")),
		Filters: []runtime.FileFilter{
			{DisplayName: "Markdown", Pattern: "*.md"},
		},
	})
	if err != nil {
		return err
	}
	if filePath == "" {
		return fmt.Errorf("export cancelled")
	}
	if !strings.HasSuffix(filePath, ".md") {
		filePath += ".md"
	}
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	if err := os.WriteFile(filePath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}
