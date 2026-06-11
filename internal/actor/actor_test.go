package actor

import (
	"context"
	"testing"
	"time"
)

func TestSpawnActor(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	actor, err := registry.Spawn(ctx, SpawnOpts{
		Type:      ActorExplore,
		SessionID: "session1",
		Prompt:    "test prompt",
	})

	if err != nil {
		t.Fatalf("failed to spawn actor: %v", err)
	}

	if actor.ID == "" {
		t.Error("actor ID should not be empty")
	}

	if actor.Type != ActorExplore {
		t.Errorf("expected type 'explore', got '%s'", actor.Type)
	}

	if actor.Status != ActorPending {
		t.Errorf("expected status 'pending', got '%s'", actor.Status)
	}
}

func TestActorExecution(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	actor, _ := registry.Spawn(ctx, SpawnOpts{
		Type:      ActorTitle,
		SessionID: "session1",
		Prompt:    "generate title",
	})

	// Wait for actor to complete
	time.Sleep(200 * time.Millisecond)

	result, ok := registry.Get(actor.ID)
	if !ok {
		t.Fatal("actor not found")
	}

	if result.Status != ActorCompleted {
		t.Errorf("expected status 'completed', got '%s'", result.Status)
	}

	if result.Result == "" {
		t.Error("result should not be empty")
	}
}

func TestCancelActor(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	actor, _ := registry.Spawn(ctx, SpawnOpts{
		Type:      ActorGeneral,
		SessionID: "session1",
		Prompt:    "long running task",
	})

	err := registry.Cancel(actor.ID)
	if err != nil {
		t.Fatalf("failed to cancel actor: %v", err)
	}

	result, _ := registry.Get(actor.ID)
	if result.Status != ActorCancelled {
		t.Errorf("expected status 'cancelled', got '%s'", result.Status)
	}
}

func TestListActors(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	registry.Spawn(ctx, SpawnOpts{Type: ActorExplore, SessionID: "session1", Prompt: "test1"})
	registry.Spawn(ctx, SpawnOpts{Type: ActorGeneral, SessionID: "session1", Prompt: "test2"})
	registry.Spawn(ctx, SpawnOpts{Type: ActorTitle, SessionID: "session2", Prompt: "test3"})

	actors := registry.List("session1")
	if len(actors) != 2 {
		t.Errorf("expected 2 actors, got %d", len(actors))
	}

	allActors := registry.List("")
	if len(allActors) != 3 {
		t.Errorf("expected 3 actors, got %d", len(allActors))
	}
}

func TestCleanup(t *testing.T) {
	registry := NewRegistry()
	ctx := context.Background()

	// Spawn and complete an actor
	actor, _ := registry.Spawn(ctx, SpawnOpts{
		Type:      ActorTitle,
		SessionID: "session1",
		Prompt:    "test",
	})

	// Wait for actor to complete
	time.Sleep(300 * time.Millisecond)

	// Verify actor is completed
	a, ok := registry.Get(actor.ID)
	if !ok || a.Status != ActorCompleted {
		t.Fatalf("actor should be completed, got status: %v", a.Status)
	}

	// Cleanup with 0 duration should remove completed actors
	removed := registry.Cleanup(0)
	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}

	_, ok = registry.Get(actor.ID)
	if ok {
		t.Error("actor should have been removed")
	}
}
