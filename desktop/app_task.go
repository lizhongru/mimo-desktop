package desktop

import (
	"fmt"

	"github.com/mimo-cli/mimo-cli/internal/task"
)

// TaskInfo represents task information for the frontend
type TaskInfo struct {
	ID           string  `json:"id"`
	SessionID    string  `json:"session_id"`
	ParentTaskID *string `json:"parent_task_id,omitempty"`
	Status       string  `json:"status"`
	Summary      string  `json:"summary"`
	Owner        *string `json:"owner,omitempty"`
	CreatedAt    int64   `json:"created_at"`
	LastEventAt  int64   `json:"last_event_at"`
	EndedAt      *int64  `json:"ended_at,omitempty"`
}

// TaskEventInfo represents task event information for the frontend
type TaskEventInfo struct {
	ID      int64  `json:"id"`
	TaskID  string `json:"task_id"`
	At      int64  `json:"at"`
	Kind    string `json:"kind"`
	Summary string `json:"summary,omitempty"`
}

// TaskResult represents the result of a task operation
type TaskResult struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Task    *TaskInfo `json:"task,omitempty"`
}

// TaskCreate creates a new task
func (a *App) TaskCreate(summary string, parentID string) TaskResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.currentSessionID == "" {
		return TaskResult{Success: false, Message: "No active session"}
	}

	if a.sessionStore == nil {
		return TaskResult{Success: false, Message: "Session store not initialized"}
	}

	registry := task.NewRegistry(a.sessionStore.DB())
	if registry == nil {
		return TaskResult{Success: false, Message: "Task registry not available"}
	}

	var parent *string
	if parentID != "" {
		parent = &parentID
	}

	t, err := registry.Create(a.currentSessionID, summary, parent)
	if err != nil {
		return TaskResult{Success: false, Message: fmt.Sprintf("Failed to create task: %v", err)}
	}

	return TaskResult{
		Success: true,
		Message: "Task created successfully",
		Task:    taskInfoFromTask(t),
	}
}

// TaskList returns tasks for the current session
func (a *App) TaskList(status string, includeTerminal bool) []TaskInfo {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.currentSessionID == "" || a.sessionStore == nil {
		return []TaskInfo{}
	}

	registry := task.NewRegistry(a.sessionStore.DB())
	if registry == nil {
		return []TaskInfo{}
	}

	var statusPtr *task.TaskStatus
	if status != "" {
		s := task.TaskStatus(status)
		statusPtr = &s
	}

	tasks, err := registry.List(a.currentSessionID, statusPtr, includeTerminal)
	if err != nil {
		return []TaskInfo{}
	}

	var result []TaskInfo
	for _, t := range tasks {
		result = append(result, *taskInfoFromTask(&t))
	}
	return result
}

// TaskStart marks a task as in progress
func (a *App) TaskStart(id, owner, eventSummary string) TaskResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	registry := task.NewRegistry(a.sessionStore.DB())
	if registry == nil {
		return TaskResult{Success: false, Message: "Task registry not available"}
	}

	t, err := registry.Start(id, owner, eventSummary)
	if err != nil {
		return TaskResult{Success: false, Message: fmt.Sprintf("Failed to start task: %v", err)}
	}

	return TaskResult{
		Success: true,
		Message: "Task started",
		Task:    taskInfoFromTask(t),
	}
}

// TaskDone marks a task as completed
func (a *App) TaskDone(id, eventSummary string) TaskResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	registry := task.NewRegistry(a.sessionStore.DB())
	if registry == nil {
		return TaskResult{Success: false, Message: "Task registry not available"}
	}

	t, err := registry.Done(id, eventSummary)
	if err != nil {
		return TaskResult{Success: false, Message: fmt.Sprintf("Failed to complete task: %v", err)}
	}

	return TaskResult{
		Success: true,
		Message: "Task completed",
		Task:    taskInfoFromTask(t),
	}
}

// TaskBlock marks a task as blocked
func (a *App) TaskBlock(id, eventSummary string) TaskResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	registry := task.NewRegistry(a.sessionStore.DB())
	if registry == nil {
		return TaskResult{Success: false, Message: "Task registry not available"}
	}

	t, err := registry.Block(id, eventSummary)
	if err != nil {
		return TaskResult{Success: false, Message: fmt.Sprintf("Failed to block task: %v", err)}
	}

	return TaskResult{
		Success: true,
		Message: "Task blocked",
		Task:    taskInfoFromTask(t),
	}
}

// TaskDelete deletes a task
func (a *App) TaskDelete(id string) TaskResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	registry := task.NewRegistry(a.sessionStore.DB())
	if registry == nil {
		return TaskResult{Success: false, Message: "Task registry not available"}
	}

	if err := registry.Delete(id); err != nil {
		return TaskResult{Success: false, Message: fmt.Sprintf("Failed to delete task: %v", err)}
	}

	return TaskResult{
		Success: true,
		Message: "Task deleted",
	}
}

// TaskGetEvents returns events for a task
func (a *App) TaskGetEvents(taskID string) []TaskEventInfo {
	a.mu.Lock()
	defer a.mu.Unlock()

	registry := task.NewRegistry(a.sessionStore.DB())
	if registry == nil {
		return []TaskEventInfo{}
	}

	events, err := registry.GetEvents(taskID)
	if err != nil {
		return []TaskEventInfo{}
	}

	var result []TaskEventInfo
	for _, event := range events {
		result = append(result, TaskEventInfo{
			ID:      event.ID,
			TaskID:  event.TaskID,
			At:      event.At,
			Kind:    event.Kind,
			Summary: event.Summary,
		})
	}
	return result
}

// TaskRename updates a task summary.
func (a *App) TaskRename(id string, newSummary string) TaskResult {
	a.mu.Lock()
	defer a.mu.Unlock()
	registry := task.NewRegistry(a.sessionStore.DB())
	t, err := registry.Rename(id, newSummary)
	if err != nil {
		return TaskResult{Success: false, Message: err.Error()}
	}
	return TaskResult{Success: true, Message: "Task renamed", Task: taskInfoFromTask(t)}
}

// TaskArchive archives a completed/abandoned/blocked task.
func (a *App) TaskArchive(id string) TaskResult {
	a.mu.Lock()
	defer a.mu.Unlock()
	registry := task.NewRegistry(a.sessionStore.DB())
	t, err := registry.Archive(id)
	if err != nil {
		return TaskResult{Success: false, Message: err.Error()}
	}
	return TaskResult{Success: true, Message: "Task archived", Task: taskInfoFromTask(t)}
}

// TaskProgress adds a progress event to a task.
func (a *App) TaskProgress(id string, summary string) TaskResult {
	a.mu.Lock()
	defer a.mu.Unlock()
	registry := task.NewRegistry(a.sessionStore.DB())
	if err := registry.Progress(id, summary); err != nil {
		return TaskResult{Success: false, Message: err.Error()}
	}
	return TaskResult{Success: true, Message: "Progress recorded"}
}

func taskInfoFromTask(t *task.Task) *TaskInfo {
	if t == nil {
		return nil
	}
	return &TaskInfo{
		ID:           t.ID,
		SessionID:    t.SessionID,
		ParentTaskID: t.ParentTaskID,
		Status:       string(t.Status),
		Summary:      t.Summary,
		Owner:        t.Owner,
		CreatedAt:    t.CreatedAt,
		LastEventAt:  t.LastEventAt,
		EndedAt:      t.EndedAt,
	}
}
