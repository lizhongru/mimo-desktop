package desktop

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// appVersion and appCommit are set at build time via -ldflags.
var (
	appVersion = "dev"
	appCommit  = "unknown"
)

// cancelFunc holds the cancel function for the current SendMessage goroutine.
// Stored on App so CancelOperation can cancel the context.

// SendMessage sends a user message and starts streaming the agent response.
func (a *App) SendMessage(message string) error {
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

		start := time.Now()
		response, err := a.agent.ChatStream(ctx, message)
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
	return map[string]int{"before": before, "after": after}, nil
}

// toJSON is a helper to marshal any value to a JSON string for event payloads.
func toJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}
