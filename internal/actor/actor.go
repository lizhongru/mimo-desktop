package actor

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ActorType represents the type of actor
type ActorType string

const (
	ActorExplore    ActorType = "explore"
	ActorGeneral    ActorType = "general"
	ActorTitle      ActorType = "title"
	ActorSummary    ActorType = "summary"
	ActorCompaction ActorType = "compaction"
	ActorDream      ActorType = "dream"
	ActorDistill    ActorType = "distill"
	ActorCheckpoint ActorType = "checkpoint-writer"
)

// ActorStatus represents the status of an actor
type ActorStatus string

const (
	ActorPending   ActorStatus = "pending"
	ActorRunning   ActorStatus = "running"
	ActorCompleted ActorStatus = "completed"
	ActorFailed    ActorStatus = "failed"
	ActorCancelled ActorStatus = "cancelled"
)

// Actor represents a sub-agent
type Actor struct {
	ID          string      `json:"id"`
	Type        ActorType   `json:"type"`
	SessionID   string      `json:"session_id"`
	ParentID    *string     `json:"parent_id,omitempty"`
	Status      ActorStatus `json:"status"`
	Prompt      string      `json:"prompt"`
	Result      string      `json:"result,omitempty"`
	Error       string      `json:"error,omitempty"`
	CreatedAt   int64       `json:"created_at"`
	StartedAt   *int64      `json:"started_at,omitempty"`
	CompletedAt *int64      `json:"completed_at,omitempty"`
}

// SpawnOpts contains options for spawning an actor
type SpawnOpts struct {
	Type      ActorType
	SessionID string
	ParentID  *string
	Prompt    string
	TaskID    *string
}

// Message represents a message sent to an actor
type Message struct {
	Content string
}

// Executor runs the actual task for an actor.
// Implementations can use LLM, mock logic, or any other backend.
type Executor interface {
	ExecuteActor(ctx context.Context, actor *Actor) (string, error)
}

// Registry manages actor lifecycle
type Registry struct {
	mu       sync.RWMutex
	actors   map[string]*Actor
	nextID   int
	executor Executor // nil = use mock fallback
}

// NewRegistry creates a new actor registry with mock execution
func NewRegistry() *Registry {
	return &Registry{
		actors:  make(map[string]*Actor),
		nextID: 1,
	}
}

// NewRegistryWithExecutor creates a new actor registry with a real executor
func NewRegistryWithExecutor(exec Executor) *Registry {
	return &Registry{
		actors:  make(map[string]*Actor),
		nextID: 1,
		executor: exec,
	}
}

// Spawn creates and starts a new actor
func (r *Registry) Spawn(ctx context.Context, opts SpawnOpts) (*Actor, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := fmt.Sprintf("actor-%d", r.nextID)
	r.nextID++

	now := time.Now().Unix()
	actor := &Actor{
		ID:        id,
		Type:      opts.Type,
		SessionID: opts.SessionID,
		ParentID:  opts.ParentID,
		Status:    ActorPending,
		Prompt:    opts.Prompt,
		CreatedAt: now,
	}

	r.actors[id] = actor

	// Start actor in background
	go r.runActor(ctx, actor)

	return actor, nil
}

// runActor executes the actor's task
func (r *Registry) runActor(ctx context.Context, act *Actor) {
	r.mu.Lock()
	act.Status = ActorRunning
	now := time.Now().Unix()
	act.StartedAt = &now
	r.mu.Unlock()

	var result string
	var err error

	if r.executor != nil {
		result, err = r.executor.ExecuteActor(ctx, act)
	} else {
		result, err = r.runMock(ctx, act)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	completedAt := time.Now().Unix()
	act.CompletedAt = &completedAt

	if err != nil {
		act.Status = ActorFailed
		act.Error = err.Error()
	} else {
		act.Status = ActorCompleted
		act.Result = result
	}
}

// runMock is the fallback when no executor is configured
func (r *Registry) runMock(ctx context.Context, act *Actor) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(100 * time.Millisecond):
		return fmt.Sprintf("[%s] Processed: %s", act.Type, act.Prompt), nil
	}
}

// Get returns an actor by ID
func (r *Registry) Get(id string) (*Actor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	actor, ok := r.actors[id]
	return actor, ok
}

// List returns all actors for a session
func (r *Registry) List(sessionID string) []*Actor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var actors []*Actor
	for _, actor := range r.actors {
		if sessionID == "" || actor.SessionID == sessionID {
			actors = append(actors, actor)
		}
	}
	return actors
}

// ListByStatus returns actors with a specific status
func (r *Registry) ListByStatus(status ActorStatus) []*Actor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var actors []*Actor
	for _, actor := range r.actors {
		if actor.Status == status {
			actors = append(actors, actor)
		}
	}
	return actors
}

// Send sends a message to an actor (placeholder)
func (r *Registry) Send(actorID string, content string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	actor, ok := r.actors[actorID]
	if !ok {
		return fmt.Errorf("actor %s not found", actorID)
	}

	if actor.Status != ActorRunning {
		return fmt.Errorf("actor %s is not running", actorID)
	}

	// Placeholder for message handling
	return nil
}

// Cancel cancels an actor
func (r *Registry) Cancel(actorID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	actor, ok := r.actors[actorID]
	if !ok {
		return fmt.Errorf("actor %s not found", actorID)
	}

	if actor.Status != ActorRunning && actor.Status != ActorPending {
		return fmt.Errorf("actor %s cannot be cancelled", actorID)
	}

	now := time.Now().Unix()
	actor.Status = ActorCancelled
	actor.CompletedAt = &now

	return nil
}

// Cleanup removes completed/failed/cancelled actors older than the given duration
func (r *Registry) Cleanup(olderThan time.Duration) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	cutoff := time.Now().Add(-olderThan).Unix()
	removed := 0

	for id, actor := range r.actors {
		if actor.Status == ActorCompleted ||
			actor.Status == ActorFailed ||
			actor.Status == ActorCancelled {
			if actor.CompletedAt != nil && *actor.CompletedAt <= cutoff {
				delete(r.actors, id)
				removed++
			}
		}
	}
	return removed
}
