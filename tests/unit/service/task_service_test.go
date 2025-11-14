package service_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"middleware-chaos-testing/internal/service"
	"middleware-chaos-testing/internal/storage"
	"middleware-chaos-testing/internal/types"
)

// MockTaskStore is a mock implementation of TaskStore
type MockTaskStore struct {
	mock.Mock
}

func (m *MockTaskStore) CreateTask(ctx context.Context, task *types.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskStore) GetTask(ctx context.Context, taskID string) (*types.Task, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Task), args.Error(1)
}

func (m *MockTaskStore) ListTasks(ctx context.Context, filter types.TaskFilter) (*types.TaskListResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.TaskListResponse), args.Error(1)
}

func (m *MockTaskStore) UpdateTask(ctx context.Context, task *types.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskStore) DeleteTask(ctx context.Context, taskID string) error {
	args := m.Called(ctx, taskID)
	return args.Error(0)
}

func (m *MockTaskStore) UpdateTaskStatus(ctx context.Context, taskID string, status types.TaskStatus) error {
	args := m.Called(ctx, taskID, status)
	return args.Error(0)
}

// MockExecutor is a mock implementation of MCTExecutor
type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) Execute(ctx context.Context, task *types.Task) (*types.TestResult, error) {
	args := m.Called(ctx, task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.TestResult), args.Error(1)
}

// TestTaskService_CreateTask tests task creation
func TestTaskService_CreateTask(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockTaskStore)
	mockExecutor := new(MockExecutor)

	svc := service.NewTaskService(mockStore, nil, mockExecutor)

	req := &types.TaskRequest{
		Middleware: "redis",
		Config: map[string]string{
			"host": "localhost",
			"port": "6379",
		},
		Duration:    "30s",
		Operations:  5000,
		Concurrency: 10,
	}

	// Setup mock expectations
	mockStore.On("CreateTask", ctx, mock.AnythingOfType("*types.Task")).Return(nil)

	// Execute
	task, err := svc.CreateTask(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.NotEmpty(t, task.ID)
	assert.Equal(t, "redis", task.Middleware)
	assert.Equal(t, types.TaskStatusPending, task.Status)
	mockStore.AssertExpectations(t)
}

// TestTaskService_CreateTask_InvalidMiddleware tests validation
func TestTaskService_CreateTask_InvalidMiddleware(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockTaskStore)
	mockExecutor := new(MockExecutor)

	svc := service.NewTaskService(mockStore, nil, mockExecutor)

	req := &types.TaskRequest{
		Middleware: "invalid",
		Config: map[string]string{
			"host": "localhost",
		},
		Duration: "30s",
	}

	// Execute
	task, err := svc.CreateTask(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "unsupported middleware")
}

// TestTaskService_GetTask tests retrieving a task
func TestTaskService_GetTask(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockTaskStore)

	svc := service.NewTaskService(mockStore, nil, nil)

	taskID := "test-task-id"
	expectedTask := &types.Task{
		ID:         taskID,
		Middleware: "redis",
		Status:     types.TaskStatusPending,
		Config:     `{"host":"localhost","port":"6379"}`,
		CreatedAt:  time.Now(),
	}

	// Setup mock
	mockStore.On("GetTask", ctx, taskID).Return(expectedTask, nil)

	// Execute
	task, err := svc.GetTask(ctx, taskID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, taskID, task.ID)
	assert.Equal(t, "redis", task.Middleware)
	mockStore.AssertExpectations(t)
}

// TestTaskService_GetTask_NotFound tests getting non-existent task
func TestTaskService_GetTask_NotFound(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockTaskStore)

	svc := service.NewTaskService(mockStore, nil, nil)

	taskID := "non-existent"

	// Setup mock
	mockStore.On("GetTask", ctx, taskID).Return(nil, storage.ErrTaskNotFound)

	// Execute
	task, err := svc.GetTask(ctx, taskID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Equal(t, storage.ErrTaskNotFound, err)
	mockStore.AssertExpectations(t)
}

// TestTaskService_ListTasks tests listing tasks
func TestTaskService_ListTasks(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockTaskStore)

	svc := service.NewTaskService(mockStore, nil, nil)

	filter := types.TaskFilter{
		Page:     1,
		PageSize: 10,
	}

	expectedTasks := []*types.Task{
		{
			ID:         "task-1",
			Middleware: "redis",
			Status:     types.TaskStatusCompleted,
			CreatedAt:  time.Now(),
		},
		{
			ID:         "task-2",
			Middleware: "kafka",
			Status:     types.TaskStatusRunning,
			CreatedAt:  time.Now(),
		},
	}

	expectedResp := &types.TaskListResponse{
		Tasks:      expectedTasks,
		Total:      2,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
	}

	// Setup mock
	mockStore.On("ListTasks", ctx, filter).Return(expectedResp, nil)

	// Execute
	resp, err := svc.ListTasks(ctx, filter)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, len(resp.Tasks))
	assert.Equal(t, int64(2), resp.Total)
	mockStore.AssertExpectations(t)
}

// TestTaskService_DeleteTask tests task deletion
func TestTaskService_DeleteTask(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockTaskStore)

	svc := service.NewTaskService(mockStore, nil, nil)

	taskID := "task-to-delete"

	// Setup mock
	mockStore.On("DeleteTask", ctx, taskID).Return(nil)

	// Execute
	err := svc.DeleteTask(ctx, taskID)

	// Assert
	assert.NoError(t, err)
	mockStore.AssertExpectations(t)
}

// TestTaskService_RunTask tests executing a task
func TestTaskService_RunTask(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockTaskStore)
	mockExecutor := new(MockExecutor)

	svc := service.NewTaskService(mockStore, nil, mockExecutor)

	task := &types.Task{
		ID:         "test-task",
		Middleware: "redis",
		Status:     types.TaskStatusPending,
		Config:     `{"host":"localhost","port":"6379"}`,
		CreatedAt:  time.Now(),
	}

	expectedResult := &types.TestResult{
		TaskID:     "test-task",
		Middleware: "redis",
		Score:      85.5,
		Grade:      "GOOD",
		Status:     "PASS",
	}

	// Setup mocks
	mockStore.On("UpdateTaskStatus", ctx, task.ID, types.TaskStatusRunning).Return(nil)
	mockExecutor.On("Execute", ctx, task).Return(expectedResult, nil)
	mockStore.On("UpdateTaskStatus", ctx, task.ID, types.TaskStatusCompleted).Return(nil)
	mockStore.On("UpdateTask", ctx, mock.AnythingOfType("*types.Task")).Return(nil)

	// Execute
	result, err := svc.RunTask(ctx, task)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 85.5, result.Score)
	assert.Equal(t, "GOOD", string(result.Grade))
	mockStore.AssertExpectations(t)
	mockExecutor.AssertExpectations(t)
}

// TestTaskService_RunTask_ExecutorError tests handling executor errors
func TestTaskService_RunTask_ExecutorError(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockTaskStore)
	mockExecutor := new(MockExecutor)

	svc := service.NewTaskService(mockStore, nil, mockExecutor)

	task := &types.Task{
		ID:         "test-task",
		Middleware: "redis",
		Status:     types.TaskStatusPending,
		Config:     `{"host":"localhost","port":"6379"}`,
		CreatedAt:  time.Now(),
	}

	// Setup mocks
	mockStore.On("UpdateTaskStatus", ctx, task.ID, types.TaskStatusRunning).Return(nil)
	mockExecutor.On("Execute", ctx, task).Return(nil, assert.AnError)
	mockStore.On("UpdateTaskStatus", ctx, task.ID, types.TaskStatusFailed).Return(nil)
	mockStore.On("UpdateTask", ctx, mock.AnythingOfType("*types.Task")).Return(nil)

	// Execute
	result, err := svc.RunTask(ctx, task)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	mockStore.AssertExpectations(t)
	mockExecutor.AssertExpectations(t)
}

// TestTaskRequest_Validate tests request validation
func TestTaskRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     *types.TaskRequest
		wantErr bool
	}{
		{
			name: "valid_redis_request",
			req: &types.TaskRequest{
				Middleware: "redis",
				Config: map[string]string{
					"host": "localhost",
					"port": "6379",
				},
				Duration:    "30s",
				Operations:  5000,
				Concurrency: 10,
			},
			wantErr: false,
		},
		{
			name: "missing_middleware",
			req: &types.TaskRequest{
				Config:   map[string]string{},
				Duration: "30s",
			},
			wantErr: true,
		},
		{
			name: "missing_duration",
			req: &types.TaskRequest{
				Middleware: "redis",
				Config:     map[string]string{},
			},
			wantErr: true,
		},
		{
			name: "invalid_duration_format",
			req: &types.TaskRequest{
				Middleware: "redis",
				Config:     map[string]string{},
				Duration:   "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestTask_ToJSON tests task JSON serialization
func TestTask_ToJSON(t *testing.T) {
	task := &types.Task{
		ID:         "test-task",
		Middleware: "redis",
		Status:     types.TaskStatusPending,
		Config:     `{"host":"localhost","port":"6379"}`,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	data, err := json.Marshal(task)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	var decoded types.Task
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, task.ID, decoded.ID)
	assert.Equal(t, task.Middleware, decoded.Middleware)
}
