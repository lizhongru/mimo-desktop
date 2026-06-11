package tools

import (
	"context"
	"encoding/json"

	"github.com/mimo-cli/mimo-cli/internal/actor"
	"github.com/mimo-cli/mimo-cli/internal/memory"
	"github.com/mimo-cli/mimo-cli/internal/task"
)

type actorRegistryProvider func() *actor.Registry
type memoryServiceProvider func() *memory.Service
type taskRegistryProvider func() *task.Registry
type sessionIDProvider func() string

// RegisterMemoryTools exposes the persistent memory system to the agent.
func (r *Registry) RegisterMemoryTools(provider memoryServiceProvider) {
	r.Register(&MemorySearchTool{provider: provider})
	r.Register(&MemoryReconcileTool{provider: provider})
}

// RegisterTaskTools exposes the task tracker to the agent.
func (r *Registry) RegisterTaskTools(provider taskRegistryProvider, currentSession sessionIDProvider) {
	r.Register(&TaskCreateTool{provider: provider, currentSession: currentSession})
	r.Register(&TaskListTool{provider: provider, currentSession: currentSession})
	r.Register(&TaskStartTool{provider: provider})
	r.Register(&TaskDoneTool{provider: provider})
	r.Register(&TaskBlockTool{provider: provider})
}

// RegisterActorTools exposes sub-agent lifecycle controls to the agent.
func (r *Registry) RegisterActorTools(provider actorRegistryProvider, currentSession sessionIDProvider) {
	r.Register(&ActorSpawnTool{provider: provider, currentSession: currentSession})
	r.Register(&ActorListTool{provider: provider, currentSession: currentSession})
	r.Register(&ActorStatusTool{provider: provider})
	r.Register(&ActorCancelTool{provider: provider})
	r.Register(&ActorSendTool{provider: provider})
}

type MemorySearchTool struct {
	provider memoryServiceProvider
}

func (t *MemorySearchTool) Name() string { return "memory_search" }

func (t *MemorySearchTool) Description() string {
	return "Search persistent project, global, and session memories"
}

func (t *MemorySearchTool) GetSafetyLevel() SafetyLevel { return SafetyLow }

func (t *MemorySearchTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query for memory content",
			},
			"scope": map[string]interface{}{
				"type":        "string",
				"description": "Optional scope filter: global, project, or session",
			},
			"scope_id": map[string]interface{}{
				"type":        "string",
				"description": "Optional scope identifier filter",
			},
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Optional memory type filter, such as project, checkpoint, notes, progress, or free",
			},
			"limit": map[string]interface{}{
				"type":        "integer",
				"description": "Maximum number of results to return, default 10",
			},
		},
		"required": []string{"query"},
	}
}

func (t *MemorySearchTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "query")
	return err
}

func (t *MemorySearchTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *MemorySearchTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	svc := t.provider()
	if svc == nil {
		return ToolError("memory service is not available"), nil
	}

	query, _ := StringParam(params, "query")
	limit := 10
	if v, err := IntParam(params, "limit"); err == nil && v > 0 {
		limit = v
	}

	results, err := svc.Search(query, memory.SearchOpts{
		Scope:   OptionalStringParam(params, "scope", ""),
		ScopeID: OptionalStringParam(params, "scope_id", ""),
		Type:    OptionalStringParam(params, "type", ""),
		Limit:   limit,
	})
	if err != nil {
		return ToolError("memory search failed: %v", err), nil
	}

	return jsonToolResult(results)
}

type MemoryReconcileTool struct {
	provider memoryServiceProvider
}

func (t *MemoryReconcileTool) Name() string { return "memory_reconcile" }

func (t *MemoryReconcileTool) Description() string {
	return "Scan memory markdown files and update the persistent memory index"
}

func (t *MemoryReconcileTool) GetSafetyLevel() SafetyLevel { return SafetyLow }

func (t *MemoryReconcileTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []string{},
	}
}

func (t *MemoryReconcileTool) Validate(params map[string]interface{}) error { return nil }

func (t *MemoryReconcileTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *MemoryReconcileTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	svc := t.provider()
	if svc == nil {
		return ToolError("memory service is not available"), nil
	}

	indexed, pruned, err := svc.Reconcile()
	if err != nil {
		return ToolError("memory reconcile failed: %v", err), nil
	}

	return jsonToolResult(map[string]int{"indexed": indexed, "pruned": pruned})
}

type TaskCreateTool struct {
	provider       taskRegistryProvider
	currentSession sessionIDProvider
}

func (t *TaskCreateTool) Name() string { return "task_create" }

func (t *TaskCreateTool) Description() string {
	return "Create a task in the current session task tracker"
}

func (t *TaskCreateTool) GetSafetyLevel() SafetyLevel { return SafetyLow }

func (t *TaskCreateTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"summary": map[string]interface{}{
				"type":        "string",
				"description": "Short task summary",
			},
			"parent_id": map[string]interface{}{
				"type":        "string",
				"description": "Optional parent task ID",
			},
			"session_id": map[string]interface{}{
				"type":        "string",
				"description": "Optional session ID; defaults to the active session",
			},
		},
		"required": []string{"summary"},
	}
}

func (t *TaskCreateTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "summary")
	return err
}

func (t *TaskCreateTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *TaskCreateTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	registry := t.provider()
	if registry == nil {
		return ToolError("task registry is not available"), nil
	}

	sessionID := OptionalStringParam(params, "session_id", "")
	if sessionID == "" && t.currentSession != nil {
		sessionID = t.currentSession()
	}
	if sessionID == "" {
		return ToolError("no active session for task_create"), nil
	}

	summary, _ := StringParam(params, "summary")
	parentID := OptionalStringParam(params, "parent_id", "")
	var parent *string
	if parentID != "" {
		parent = &parentID
	}

	created, err := registry.Create(sessionID, summary, parent)
	if err != nil {
		return ToolError("task create failed: %v", err), nil
	}
	return jsonToolResult(created)
}

type TaskListTool struct {
	provider       taskRegistryProvider
	currentSession sessionIDProvider
}

func (t *TaskListTool) Name() string { return "task_list" }

func (t *TaskListTool) Description() string {
	return "List tasks for the current session task tracker"
}

func (t *TaskListTool) GetSafetyLevel() SafetyLevel { return SafetyLow }

func (t *TaskListTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Optional status filter: open, in_progress, blocked, done, or abandoned",
			},
			"include_terminal": map[string]interface{}{
				"type":        "boolean",
				"description": "Include done and abandoned tasks",
			},
			"session_id": map[string]interface{}{
				"type":        "string",
				"description": "Optional session ID; defaults to the active session",
			},
		},
		"required": []string{},
	}
}

func (t *TaskListTool) Validate(params map[string]interface{}) error { return nil }

func (t *TaskListTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *TaskListTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	registry := t.provider()
	if registry == nil {
		return ToolError("task registry is not available"), nil
	}

	sessionID := OptionalStringParam(params, "session_id", "")
	if sessionID == "" && t.currentSession != nil {
		sessionID = t.currentSession()
	}
	if sessionID == "" {
		return ToolError("no active session for task_list"), nil
	}

	var status *task.TaskStatus
	if raw := OptionalStringParam(params, "status", ""); raw != "" {
		s := task.TaskStatus(raw)
		status = &s
	}

	tasks, err := registry.List(sessionID, status, boolParam(params, "include_terminal", false))
	if err != nil {
		return ToolError("task list failed: %v", err), nil
	}
	return jsonToolResult(tasks)
}

type TaskStartTool struct {
	provider taskRegistryProvider
}

func (t *TaskStartTool) Name() string { return "task_start" }

func (t *TaskStartTool) Description() string {
	return "Mark a task as in progress"
}

func (t *TaskStartTool) GetSafetyLevel() SafetyLevel { return SafetyLow }

func (t *TaskStartTool) Parameters() map[string]interface{} {
	return taskTransitionParams([]string{"id"})
}

func (t *TaskStartTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "id")
	return err
}

func (t *TaskStartTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *TaskStartTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	registry := t.provider()
	if registry == nil {
		return ToolError("task registry is not available"), nil
	}
	id, _ := StringParam(params, "id")
	owner := OptionalStringParam(params, "owner", "agent")
	eventSummary := OptionalStringParam(params, "event_summary", "Task started")
	updated, err := registry.Start(id, owner, eventSummary)
	if err != nil {
		return ToolError("task start failed: %v", err), nil
	}
	return jsonToolResult(updated)
}

type TaskDoneTool struct {
	provider taskRegistryProvider
}

func (t *TaskDoneTool) Name() string { return "task_done" }

func (t *TaskDoneTool) Description() string {
	return "Mark a task as done"
}

func (t *TaskDoneTool) GetSafetyLevel() SafetyLevel { return SafetyLow }

func (t *TaskDoneTool) Parameters() map[string]interface{} {
	return taskTransitionParams([]string{"id"})
}

func (t *TaskDoneTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "id")
	return err
}

func (t *TaskDoneTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *TaskDoneTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	registry := t.provider()
	if registry == nil {
		return ToolError("task registry is not available"), nil
	}
	id, _ := StringParam(params, "id")
	eventSummary := OptionalStringParam(params, "event_summary", "Task completed")
	updated, err := registry.Done(id, eventSummary)
	if err != nil {
		return ToolError("task done failed: %v", err), nil
	}
	return jsonToolResult(updated)
}

type TaskBlockTool struct {
	provider taskRegistryProvider
}

func (t *TaskBlockTool) Name() string { return "task_block" }

func (t *TaskBlockTool) Description() string {
	return "Mark a task as blocked"
}

func (t *TaskBlockTool) GetSafetyLevel() SafetyLevel { return SafetyLow }

func (t *TaskBlockTool) Parameters() map[string]interface{} {
	return taskTransitionParams([]string{"id"})
}

func (t *TaskBlockTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "id")
	return err
}

func (t *TaskBlockTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *TaskBlockTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	registry := t.provider()
	if registry == nil {
		return ToolError("task registry is not available"), nil
	}
	id, _ := StringParam(params, "id")
	eventSummary := OptionalStringParam(params, "event_summary", "Task blocked")
	updated, err := registry.Block(id, eventSummary)
	if err != nil {
		return ToolError("task block failed: %v", err), nil
	}
	return jsonToolResult(updated)
}

func taskTransitionParams(required []string) map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Task ID",
			},
			"owner": map[string]interface{}{
				"type":        "string",
				"description": "Optional owner for task_start",
			},
			"event_summary": map[string]interface{}{
				"type":        "string",
				"description": "Optional status change summary",
			},
		},
		"required": required,
	}
}

func jsonToolResult(v interface{}) (*ToolResult, error) {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ToolError("failed to encode tool output: %v", err), nil
	}
	return &ToolResult{Output: string(out)}, nil
}

func boolParam(params map[string]interface{}, key string, defaultVal bool) bool {
	val, ok := params[key]
	if !ok {
		return defaultVal
	}
	b, ok := val.(bool)
	if !ok {
		return defaultVal
	}
	return b
}

type ActorSpawnTool struct {
	provider       actorRegistryProvider
	currentSession sessionIDProvider
}

func (t *ActorSpawnTool) Name() string { return "actor_spawn" }

func (t *ActorSpawnTool) Description() string {
	return "Spawn a sub-agent actor for exploration, general work, title generation, summary, compaction, dream, distill, or checkpoint writing"
}

func (t *ActorSpawnTool) GetSafetyLevel() SafetyLevel { return SafetyLow }

func (t *ActorSpawnTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Actor type: explore, general, title, summary, compaction, dream, distill, or checkpoint-writer",
			},
			"prompt": map[string]interface{}{
				"type":        "string",
				"description": "Prompt or task for the actor",
			},
			"parent_id": map[string]interface{}{
				"type":        "string",
				"description": "Optional parent task or actor ID",
			},
			"session_id": map[string]interface{}{
				"type":        "string",
				"description": "Optional session ID; defaults to the active session",
			},
		},
		"required": []string{"type", "prompt"},
	}
}

func (t *ActorSpawnTool) Validate(params map[string]interface{}) error {
	if _, err := StringParam(params, "type"); err != nil {
		return err
	}
	_, err := StringParam(params, "prompt")
	return err
}

func (t *ActorSpawnTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *ActorSpawnTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	registry := t.provider()
	if registry == nil {
		return ToolError("actor registry is not available"), nil
	}

	sessionID := OptionalStringParam(params, "session_id", "")
	if sessionID == "" && t.currentSession != nil {
		sessionID = t.currentSession()
	}
	if sessionID == "" {
		return ToolError("no active session for actor_spawn"), nil
	}

	actorType, _ := StringParam(params, "type")
	prompt, _ := StringParam(params, "prompt")
	parentID := OptionalStringParam(params, "parent_id", "")
	var parent *string
	if parentID != "" {
		parent = &parentID
	}

	spawned, err := registry.Spawn(ctx, actor.SpawnOpts{
		Type:      actor.ActorType(actorType),
		SessionID: sessionID,
		ParentID:  parent,
		Prompt:    prompt,
	})
	if err != nil {
		return ToolError("actor spawn failed: %v", err), nil
	}
	return jsonToolResult(spawned)
}

type ActorListTool struct {
	provider       actorRegistryProvider
	currentSession sessionIDProvider
}

func (t *ActorListTool) Name() string { return "actor_list" }

func (t *ActorListTool) Description() string {
	return "List sub-agent actors for the active session"
}

func (t *ActorListTool) GetSafetyLevel() SafetyLevel { return SafetyLow }

func (t *ActorListTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Optional status filter: pending, running, completed, failed, or cancelled",
			},
			"session_id": map[string]interface{}{
				"type":        "string",
				"description": "Optional session ID; defaults to the active session",
			},
		},
		"required": []string{},
	}
}

func (t *ActorListTool) Validate(params map[string]interface{}) error { return nil }

func (t *ActorListTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *ActorListTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	registry := t.provider()
	if registry == nil {
		return ToolError("actor registry is not available"), nil
	}

	status := OptionalStringParam(params, "status", "")
	if status != "" {
		return jsonToolResult(registry.ListByStatus(actor.ActorStatus(status)))
	}

	sessionID := OptionalStringParam(params, "session_id", "")
	if sessionID == "" && t.currentSession != nil {
		sessionID = t.currentSession()
	}
	return jsonToolResult(registry.List(sessionID))
}

type ActorStatusTool struct {
	provider actorRegistryProvider
}

func (t *ActorStatusTool) Name() string { return "actor_status" }

func (t *ActorStatusTool) Description() string {
	return "Get the status and result of a sub-agent actor"
}

func (t *ActorStatusTool) GetSafetyLevel() SafetyLevel { return SafetyLow }

func (t *ActorStatusTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Actor ID",
			},
		},
		"required": []string{"id"},
	}
}

func (t *ActorStatusTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "id")
	return err
}

func (t *ActorStatusTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *ActorStatusTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	registry := t.provider()
	if registry == nil {
		return ToolError("actor registry is not available"), nil
	}
	id, _ := StringParam(params, "id")
	act, ok := registry.Get(id)
	if !ok {
		return ToolError("actor %s not found", id), nil
	}
	return jsonToolResult(act)
}

type ActorCancelTool struct {
	provider actorRegistryProvider
}

func (t *ActorCancelTool) Name() string { return "actor_cancel" }

func (t *ActorCancelTool) Description() string {
	return "Cancel a pending or running sub-agent actor"
}

func (t *ActorCancelTool) GetSafetyLevel() SafetyLevel { return SafetyLow }

func (t *ActorCancelTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Actor ID",
			},
		},
		"required": []string{"id"},
	}
}

func (t *ActorCancelTool) Validate(params map[string]interface{}) error {
	_, err := StringParam(params, "id")
	return err
}

func (t *ActorCancelTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *ActorCancelTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	registry := t.provider()
	if registry == nil {
		return ToolError("actor registry is not available"), nil
	}
	id, _ := StringParam(params, "id")
	if err := registry.Cancel(id); err != nil {
		return ToolError("actor cancel failed: %v", err), nil
	}
	return jsonToolResult(map[string]string{"id": id, "status": string(actor.ActorCancelled)})
}

type ActorSendTool struct {
	provider actorRegistryProvider
}

func (t *ActorSendTool) Name() string { return "actor_send" }

func (t *ActorSendTool) Description() string {
	return "Send a message to a running sub-agent actor"
}

func (t *ActorSendTool) GetSafetyLevel() SafetyLevel { return SafetyLow }

func (t *ActorSendTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Actor ID",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Message content to send",
			},
		},
		"required": []string{"id", "content"},
	}
}

func (t *ActorSendTool) Validate(params map[string]interface{}) error {
	if _, err := StringParam(params, "id"); err != nil {
		return err
	}
	_, err := StringParam(params, "content")
	return err
}

func (t *ActorSendTool) RequiresConfirmation(params map[string]interface{}) bool { return false }

func (t *ActorSendTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	registry := t.provider()
	if registry == nil {
		return ToolError("actor registry is not available"), nil
	}
	id, _ := StringParam(params, "id")
	content, _ := StringParam(params, "content")
	if err := registry.Send(id, content); err != nil {
		return ToolError("actor send failed: %v", err), nil
	}
	return jsonToolResult(map[string]string{"id": id, "message": "sent"})
}
