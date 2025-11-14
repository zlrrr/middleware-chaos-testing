package storage

import (
	"context"
	"sort"
	"sync"
	"time"

	"middleware-chaos-testing/internal/types"
)

// MemoryTaskStore is an in-memory implementation of TaskStore
type MemoryTaskStore struct {
	mu    sync.RWMutex
	tasks map[string]*types.Task
}

// NewMemoryTaskStore creates a new in-memory task store
func NewMemoryTaskStore() *MemoryTaskStore {
	return &MemoryTaskStore{
		tasks: make(map[string]*types.Task),
	}
}

// CreateTask creates a new task
func (m *MemoryTaskStore) CreateTask(ctx context.Context, task *types.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[task.ID]; exists {
		return ErrTaskAlreadyExists
	}

	// Create a copy to avoid external modifications
	taskCopy := *task
	m.tasks[task.ID] = &taskCopy

	return nil
}

// GetTask retrieves a task by ID
func (m *MemoryTaskStore) GetTask(ctx context.Context, taskID string) (*types.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return nil, ErrTaskNotFound
	}

	// Return a copy
	taskCopy := *task
	return &taskCopy, nil
}

// ListTasks retrieves a paginated list of tasks with optional filters
func (m *MemoryTaskStore) ListTasks(ctx context.Context, filter types.TaskFilter) (*types.TaskListResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Collect all tasks into a slice
	allTasks := make([]*types.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		// Apply filters
		if filter.Middleware != "" && task.Middleware != filter.Middleware {
			continue
		}
		if filter.Status != "" && task.Status != filter.Status {
			continue
		}

		// Create a copy
		taskCopy := *task
		allTasks = append(allTasks, &taskCopy)
	}

	// Sort by created time (newest first)
	sort.Slice(allTasks, func(i, j int) bool {
		return allTasks[i].CreatedAt.After(allTasks[j].CreatedAt)
	})

	total := int64(len(allTasks))

	// Apply pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(allTasks) {
		allTasks = []*types.Task{}
	} else {
		if end > len(allTasks) {
			end = len(allTasks)
		}
		allTasks = allTasks[start:end]
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &types.TaskListResponse{
		Tasks:      allTasks,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateTask updates an existing task
func (m *MemoryTaskStore) UpdateTask(ctx context.Context, task *types.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[task.ID]; !exists {
		return ErrTaskNotFound
	}

	task.UpdatedAt = time.Now()
	taskCopy := *task
	m.tasks[task.ID] = &taskCopy

	return nil
}

// DeleteTask deletes a task by ID
func (m *MemoryTaskStore) DeleteTask(ctx context.Context, taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[taskID]; !exists {
		return ErrTaskNotFound
	}

	delete(m.tasks, taskID)
	return nil
}

// UpdateTaskStatus updates the status of a task
func (m *MemoryTaskStore) UpdateTaskStatus(ctx context.Context, taskID string, status types.TaskStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return ErrTaskNotFound
	}

	task.Status = status
	task.UpdatedAt = time.Now()

	return nil
}

// MemoryResultStore is an in-memory implementation of ResultStore
type MemoryResultStore struct {
	mu      sync.RWMutex
	results map[string]*types.TestResult
}

// NewMemoryResultStore creates a new in-memory result store
func NewMemoryResultStore() *MemoryResultStore {
	return &MemoryResultStore{
		results: make(map[string]*types.TestResult),
	}
}

// SaveResult saves a test result
func (m *MemoryResultStore) SaveResult(ctx context.Context, result *types.TestResult) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a copy
	resultCopy := *result
	m.results[result.TaskID] = &resultCopy

	return nil
}

// GetResult retrieves a test result by task ID
func (m *MemoryResultStore) GetResult(ctx context.Context, taskID string) (*types.TestResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result, exists := m.results[taskID]
	if !exists {
		return nil, ErrResultNotFound
	}

	// Return a copy
	resultCopy := *result
	return &resultCopy, nil
}

// DeleteResult deletes a test result by task ID
func (m *MemoryResultStore) DeleteResult(ctx context.Context, taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.results[taskID]; !exists {
		return ErrResultNotFound
	}

	delete(m.results, taskID)
	return nil
}
