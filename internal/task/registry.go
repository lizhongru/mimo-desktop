package task

import (
	"database/sql"
	"fmt"
	"time"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskOpen       TaskStatus = "open"
	TaskInProgress TaskStatus = "in_progress"
	TaskBlocked    TaskStatus = "blocked"
	TaskDone       TaskStatus = "done"
	TaskAbandoned  TaskStatus = "abandoned"
)

// Task represents a tracking task
type Task struct {
	ID           string     `json:"id"`
	SessionID    string     `json:"session_id"`
	ParentTaskID *string    `json:"parent_task_id,omitempty"`
	Status       TaskStatus `json:"status"`
	Summary      string     `json:"summary"`
	Owner        *string    `json:"owner,omitempty"`
	CreatedAt    int64      `json:"created_at"`
	LastEventAt  int64      `json:"last_event_at"`
	EndedAt      *int64     `json:"ended_at,omitempty"`
	CleanupAfter *int64     `json:"cleanup_after,omitempty"`
}

// TaskEvent represents an event in a task's lifecycle
type TaskEvent struct {
	ID       int64  `json:"id"`
	TaskID   string `json:"task_id"`
	At       int64  `json:"at"`
	Kind     string `json:"kind"`
	Summary  string `json:"summary,omitempty"`
}

// Registry manages task persistence and state
type Registry struct {
	db *sql.DB
}

// NewRegistry creates a new task registry
func NewRegistry(db *sql.DB) *Registry {
	return &Registry{db: db}
}

// Create creates a new task
func (r *Registry) Create(sessionID, summary string, parentID *string) (*Task, error) {
	id := fmt.Sprintf("T%d", time.Now().UnixNano())
	now := time.Now().Unix()

	_, err := r.db.Exec(`
		INSERT INTO tasks (id, session_id, parent_task_id, status, summary, created_at, last_event_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, sessionID, parentID, TaskOpen, summary, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Create initial event
	r.addEvent(id, "created", "Task created")

	return r.Get(id)
}

// Get returns a task by ID
func (r *Registry) Get(id string) (*Task, error) {
	task := &Task{}
	err := r.db.QueryRow(`
		SELECT id, session_id, parent_task_id, status, summary, owner, created_at, last_event_at, ended_at, cleanup_after
		FROM tasks WHERE id = ?
	`, id).Scan(
		&task.ID, &task.SessionID, &task.ParentTaskID, &task.Status, &task.Summary,
		&task.Owner, &task.CreatedAt, &task.LastEventAt, &task.EndedAt, &task.CleanupAfter,
	)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// List returns tasks for a session
func (r *Registry) List(sessionID string, status *TaskStatus, includeTerminal bool) ([]Task, error) {
	query := `SELECT id, session_id, parent_task_id, status, summary, owner, created_at, last_event_at, ended_at, cleanup_after FROM tasks WHERE session_id = ?`
	args := []interface{}{sessionID}

	if status != nil {
		query += ` AND status = ?`
		args = append(args, *status)
	} else if !includeTerminal {
		query += ` AND status NOT IN (?, ?)`
		args = append(args, TaskDone, TaskAbandoned)
	}

	query += ` ORDER BY created_at ASC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(
			&task.ID, &task.SessionID, &task.ParentTaskID, &task.Status, &task.Summary,
			&task.Owner, &task.CreatedAt, &task.LastEventAt, &task.EndedAt, &task.CleanupAfter,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

// Start marks a task as in progress
func (r *Registry) Start(id, owner, eventSummary string) (*Task, error) {
	now := time.Now().Unix()
	_, err := r.db.Exec(`
		UPDATE tasks SET status = ?, owner = ?, last_event_at = ? WHERE id = ?
	`, TaskInProgress, owner, now, id)
	if err != nil {
		return nil, err
	}
	r.addEvent(id, "started", eventSummary)
	return r.Get(id)
}

// Done marks a task as completed
func (r *Registry) Done(id, eventSummary string) (*Task, error) {
	now := time.Now().Unix()
	_, err := r.db.Exec(`
		UPDATE tasks SET status = ?, ended_at = ?, last_event_at = ? WHERE id = ?
	`, TaskDone, now, now, id)
	if err != nil {
		return nil, err
	}
	r.addEvent(id, "completed", eventSummary)
	return r.Get(id)
}

// Block marks a task as blocked
func (r *Registry) Block(id, eventSummary string) (*Task, error) {
	now := time.Now().Unix()
	_, err := r.db.Exec(`
		UPDATE tasks SET status = ?, last_event_at = ? WHERE id = ?
	`, TaskBlocked, now, id)
	if err != nil {
		return nil, err
	}
	r.addEvent(id, "blocked", eventSummary)
	return r.Get(id)
}

// Unblock unblocks a task
func (r *Registry) Unblock(id, eventSummary string) (*Task, error) {
	now := time.Now().Unix()
	_, err := r.db.Exec(`
		UPDATE tasks SET status = ?, last_event_at = ? WHERE id = ?
	`, TaskInProgress, now, id)
	if err != nil {
		return nil, err
	}
	r.addEvent(id, "unblocked", eventSummary)
	return r.Get(id)
}

// Abandon marks a task as abandoned
func (r *Registry) Abandon(id, eventSummary string) (*Task, error) {
	now := time.Now().Unix()
	_, err := r.db.Exec(`
		UPDATE tasks SET status = ?, ended_at = ?, last_event_at = ? WHERE id = ?
	`, TaskAbandoned, now, now, id)
	if err != nil {
		return nil, err
	}
	r.addEvent(id, "abandoned", eventSummary)
	return r.Get(id)
}

// addEvent adds an event to a task
func (r *Registry) addEvent(taskID, kind, summary string) {
	now := time.Now().Unix()
	r.db.Exec(`
		INSERT INTO task_events (task_id, at, kind, summary)
		VALUES (?, ?, ?, ?)
	`, taskID, now, kind, summary)
}

// GetEvents returns events for a task
func (r *Registry) GetEvents(taskID string) ([]TaskEvent, error) {
	rows, err := r.db.Query(`
		SELECT id, task_id, at, kind, summary
		FROM task_events WHERE task_id = ? ORDER BY at ASC
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []TaskEvent
	for rows.Next() {
		var event TaskEvent
		if err := rows.Scan(&event.ID, &event.TaskID, &event.At, &event.Kind, &event.Summary); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

// GetChildren returns child tasks of a parent task
func (r *Registry) GetChildren(parentID string) ([]Task, error) {
	rows, err := r.db.Query(`
		SELECT id, session_id, parent_task_id, status, summary, owner, created_at, last_event_at, ended_at, cleanup_after
		FROM tasks WHERE parent_task_id = ? ORDER BY created_at ASC
	`, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(
			&task.ID, &task.SessionID, &task.ParentTaskID, &task.Status, &task.Summary,
			&task.Owner, &task.CreatedAt, &task.LastEventAt, &task.EndedAt, &task.CleanupAfter,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

// Delete removes a task and its events
func (r *Registry) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM task_events WHERE task_id = ?", id)
	if err != nil {
		return err
	}
	_, err = r.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}
