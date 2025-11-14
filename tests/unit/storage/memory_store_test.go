package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"middleware-chaos-testing/internal/storage"
	"middleware-chaos-testing/internal/types"
)

// TestMemoryTaskStore_CreateAndGet tests creating and retrieving tasks
func TestMemoryTaskStore_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryTaskStore()

	task := &types.Task{
		ID:         "test-task-1",
		Middleware: "redis",
		Status:     types.TaskStatusPending,
		Config:     `{"host":"localhost"}`,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Create task
	err := store.CreateTask(ctx, task)
	assert.NoError(t, err)

	// Get task
	retrieved, err := store.GetTask(ctx, task.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, task.ID, retrieved.ID)
	assert.Equal(t, task.Middleware, retrieved.Middleware)
	assert.Equal(t, task.Status, retrieved.Status)
}

// TestMemoryTaskStore_GetNotFound tests getting non-existent task
func TestMemoryTaskStore_GetNotFound(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryTaskStore()

	task, err := store.GetTask(ctx, "non-existent")
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Equal(t, storage.ErrTaskNotFound, err)
}

// TestMemoryTaskStore_ListTasks tests listing tasks
func TestMemoryTaskStore_ListTasks(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryTaskStore()

	// Create multiple tasks
	tasks := []*types.Task{
		{
			ID:         "task-1",
			Middleware: "redis",
			Status:     types.TaskStatusCompleted,
			CreatedAt:  time.Now().Add(-2 * time.Hour),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         "task-2",
			Middleware: "kafka",
			Status:     types.TaskStatusRunning,
			CreatedAt:  time.Now().Add(-1 * time.Hour),
			UpdatedAt:  time.Now(),
		},
		{
			ID:         "task-3",
			Middleware: "redis",
			Status:     types.TaskStatusPending,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	for _, task := range tasks {
		err := store.CreateTask(ctx, task)
		assert.NoError(t, err)
	}

	// List all tasks
	filter := types.TaskFilter{
		Page:     1,
		PageSize: 10,
	}

	resp, err := store.ListTasks(ctx, filter)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 3, len(resp.Tasks))
	assert.Equal(t, int64(3), resp.Total)
	assert.Equal(t, 1, resp.TotalPages)
}

// TestMemoryTaskStore_ListTasks_WithFilter tests filtered listing
func TestMemoryTaskStore_ListTasks_WithFilter(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryTaskStore()

	// Create multiple tasks
	tasks := []*types.Task{
		{
			ID:         "task-1",
			Middleware: "redis",
			Status:     types.TaskStatusCompleted,
			CreatedAt:  time.Now(),
		},
		{
			ID:         "task-2",
			Middleware: "kafka",
			Status:     types.TaskStatusCompleted,
			CreatedAt:  time.Now(),
		},
		{
			ID:         "task-3",
			Middleware: "redis",
			Status:     types.TaskStatusRunning,
			CreatedAt:  time.Now(),
		},
	}

	for _, task := range tasks {
		err := store.CreateTask(ctx, task)
		assert.NoError(t, err)
	}

	// Filter by middleware
	filter := types.TaskFilter{
		Middleware: "redis",
		Page:       1,
		PageSize:   10,
	}

	resp, err := store.ListTasks(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(resp.Tasks))
	assert.Equal(t, int64(2), resp.Total)

	// Filter by status
	filter = types.TaskFilter{
		Status:   types.TaskStatusCompleted,
		Page:     1,
		PageSize: 10,
	}

	resp, err = store.ListTasks(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(resp.Tasks))
	assert.Equal(t, int64(2), resp.Total)
}

// TestMemoryTaskStore_ListTasks_Pagination tests pagination
func TestMemoryTaskStore_ListTasks_Pagination(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryTaskStore()

	// Create 25 tasks
	for i := 0; i < 25; i++ {
		task := &types.Task{
			ID:         string(rune('A' + i)),
			Middleware: "redis",
			Status:     types.TaskStatusCompleted,
			CreatedAt:  time.Now(),
		}
		err := store.CreateTask(ctx, task)
		assert.NoError(t, err)
	}

	// Get first page
	filter := types.TaskFilter{
		Page:     1,
		PageSize: 10,
	}

	resp, err := store.ListTasks(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(resp.Tasks))
	assert.Equal(t, int64(25), resp.Total)
	assert.Equal(t, 3, resp.TotalPages)

	// Get second page
	filter.Page = 2
	resp, err = store.ListTasks(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(resp.Tasks))

	// Get third page
	filter.Page = 3
	resp, err = store.ListTasks(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(resp.Tasks))
}

// TestMemoryTaskStore_UpdateTask tests updating a task
func TestMemoryTaskStore_UpdateTask(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryTaskStore()

	task := &types.Task{
		ID:         "test-task",
		Middleware: "redis",
		Status:     types.TaskStatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Create task
	err := store.CreateTask(ctx, task)
	assert.NoError(t, err)

	// Update task
	task.Status = types.TaskStatusRunning
	task.StartTime = time.Now()
	err = store.UpdateTask(ctx, task)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := store.GetTask(ctx, task.ID)
	assert.NoError(t, err)
	assert.Equal(t, types.TaskStatusRunning, retrieved.Status)
	assert.False(t, retrieved.StartTime.IsZero())
}

// TestMemoryTaskStore_UpdateTaskStatus tests updating task status
func TestMemoryTaskStore_UpdateTaskStatus(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryTaskStore()

	task := &types.Task{
		ID:         "test-task",
		Middleware: "redis",
		Status:     types.TaskStatusPending,
		CreatedAt:  time.Now(),
	}

	// Create task
	err := store.CreateTask(ctx, task)
	assert.NoError(t, err)

	// Update status
	err = store.UpdateTaskStatus(ctx, task.ID, types.TaskStatusRunning)
	assert.NoError(t, err)

	// Verify
	retrieved, err := store.GetTask(ctx, task.ID)
	assert.NoError(t, err)
	assert.Equal(t, types.TaskStatusRunning, retrieved.Status)
}

// TestMemoryTaskStore_DeleteTask tests deleting a task
func TestMemoryTaskStore_DeleteTask(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryTaskStore()

	task := &types.Task{
		ID:         "test-task",
		Middleware: "redis",
		Status:     types.TaskStatusCompleted,
		CreatedAt:  time.Now(),
	}

	// Create task
	err := store.CreateTask(ctx, task)
	assert.NoError(t, err)

	// Delete task
	err = store.DeleteTask(ctx, task.ID)
	assert.NoError(t, err)

	// Verify deletion
	retrieved, err := store.GetTask(ctx, task.ID)
	assert.Error(t, err)
	assert.Nil(t, retrieved)
	assert.Equal(t, storage.ErrTaskNotFound, err)
}

// TestMemoryResultStore_SaveAndGet tests result storage
func TestMemoryResultStore_SaveAndGet(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryResultStore()

	result := &types.TestResult{
		TaskID:     "test-task",
		Middleware: "redis",
		Score:      87.5,
		Grade:      "GOOD",
		Status:     "PASS",
		StartTime:  time.Now(),
		EndTime:    time.Now().Add(30 * time.Second),
		Duration:   30 * time.Second,
	}

	// Save result
	err := store.SaveResult(ctx, result)
	assert.NoError(t, err)

	// Get result
	retrieved, err := store.GetResult(ctx, result.TaskID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, result.TaskID, retrieved.TaskID)
	assert.Equal(t, result.Score, retrieved.Score)
}

// TestMemoryResultStore_GetNotFound tests getting non-existent result
func TestMemoryResultStore_GetNotFound(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryResultStore()

	result, err := store.GetResult(ctx, "non-existent")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, storage.ErrResultNotFound, err)
}

// TestMemoryResultStore_DeleteResult tests deleting a result
func TestMemoryResultStore_DeleteResult(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemoryResultStore()

	result := &types.TestResult{
		TaskID:     "test-task",
		Middleware: "redis",
		Score:      87.5,
	}

	// Save result
	err := store.SaveResult(ctx, result)
	assert.NoError(t, err)

	// Delete result
	err = store.DeleteResult(ctx, result.TaskID)
	assert.NoError(t, err)

	// Verify deletion
	retrieved, err := store.GetResult(ctx, result.TaskID)
	assert.Error(t, err)
	assert.Nil(t, retrieved)
}
