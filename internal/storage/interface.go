package storage

import (
	"context"

	"middleware-chaos-testing/internal/types"
)

// TaskStore defines the interface for task persistence
type TaskStore interface {
	// CreateTask creates a new task
	CreateTask(ctx context.Context, task *types.Task) error

	// GetTask retrieves a task by ID
	GetTask(ctx context.Context, taskID string) (*types.Task, error)

	// ListTasks retrieves a paginated list of tasks with optional filters
	ListTasks(ctx context.Context, filter types.TaskFilter) (*types.TaskListResponse, error)

	// UpdateTask updates an existing task
	UpdateTask(ctx context.Context, task *types.Task) error

	// DeleteTask deletes a task by ID
	DeleteTask(ctx context.Context, taskID string) error

	// UpdateTaskStatus updates the status of a task
	UpdateTaskStatus(ctx context.Context, taskID string, status types.TaskStatus) error
}

// ResultStore defines the interface for test result persistence
type ResultStore interface {
	// SaveResult saves a test result
	SaveResult(ctx context.Context, result *types.TestResult) error

	// GetResult retrieves a test result by task ID
	GetResult(ctx context.Context, taskID string) (*types.TestResult, error)

	// DeleteResult deletes a test result by task ID
	DeleteResult(ctx context.Context, taskID string) error
}
