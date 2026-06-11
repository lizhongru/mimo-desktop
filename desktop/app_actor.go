package desktop

import (
	"context"
	"fmt"
	"time"

	"github.com/mimo-cli/mimo-cli/internal/actor"
)

// ActorInfo represents actor information for the frontend
type ActorInfo struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	SessionID   string  `json:"session_id"`
	ParentID    *string `json:"parent_id,omitempty"`
	Status      string  `json:"status"`
	Prompt      string  `json:"prompt"`
	Result      string  `json:"result,omitempty"`
	Error       string  `json:"error,omitempty"`
	CreatedAt   int64   `json:"created_at"`
	StartedAt   *int64  `json:"started_at,omitempty"`
	CompletedAt *int64  `json:"completed_at,omitempty"`
}

// ActorResult represents the result of an actor operation
type ActorResult struct {
	Success bool       `json:"success"`
	Message string     `json:"message"`
	Actor   *ActorInfo `json:"actor,omitempty"`
}

// actorRegistry holds the actor registry (singleton)
var actorRegistry *actor.Registry

func getActorRegistry() *actor.Registry {
	if actorRegistry == nil {
		actorRegistry = actor.NewRegistry()
	}
	return actorRegistry
}

// ActorSpawn creates and starts a new actor
func (a *App) ActorSpawn(actorType string, prompt string, taskID string) ActorResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.currentSessionID == "" {
		return ActorResult{Success: false, Message: "No active session"}
	}

	registry := getActorRegistry()

	var parentID *string
	if taskID != "" {
		parentID = &taskID
	}

	act, err := registry.Spawn(context.Background(), actor.SpawnOpts{
		Type:      actor.ActorType(actorType),
		SessionID: a.currentSessionID,
		ParentID:  parentID,
		Prompt:    prompt,
	})

	if err != nil {
		return ActorResult{Success: false, Message: fmt.Sprintf("Failed to spawn actor: %v", err)}
	}

	return ActorResult{
		Success: true,
		Message: "Actor spawned successfully",
		Actor:   actorInfoFromActor(act),
	}
}

// ActorList returns actors for the current session
func (a *App) ActorList(status string) []ActorInfo {
	a.mu.Lock()
	defer a.mu.Unlock()

	registry := getActorRegistry()

	var actors []*actor.Actor
	if status != "" {
		actors = registry.ListByStatus(actor.ActorStatus(status))
	} else {
		actors = registry.List(a.currentSessionID)
	}

	var result []ActorInfo
	for _, act := range actors {
		result = append(result, *actorInfoFromActor(act))
	}
	return result
}

// ActorGet returns an actor by ID
func (a *App) ActorGet(id string) *ActorInfo {
	registry := getActorRegistry()
	act, ok := registry.Get(id)
	if !ok {
		return nil
	}
	return actorInfoFromActor(act)
}

// ActorCancel cancels an actor
func (a *App) ActorCancel(id string) ActorResult {
	registry := getActorRegistry()
	if err := registry.Cancel(id); err != nil {
		return ActorResult{Success: false, Message: err.Error()}
	}
	return ActorResult{Success: true, Message: "Actor cancelled"}
}

// ActorCleanup removes old completed actors
func (a *App) ActorCleanup(maxAge int) int {
	registry := getActorRegistry()
	return registry.Cleanup(time.Duration(maxAge) * time.Second)
}

func actorInfoFromActor(act *actor.Actor) *ActorInfo {
	if act == nil {
		return nil
	}
	return &ActorInfo{
		ID:          act.ID,
		Type:        string(act.Type),
		SessionID:   act.SessionID,
		ParentID:    act.ParentID,
		Status:      string(act.Status),
		Prompt:      act.Prompt,
		Result:      act.Result,
		Error:       act.Error,
		CreatedAt:   act.CreatedAt,
		StartedAt:   act.StartedAt,
		CompletedAt: act.CompletedAt,
	}
}
