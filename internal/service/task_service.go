package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"middleware-chaos-testing/internal/executor"
	"middleware-chaos-testing/internal/storage"
	"middleware-chaos-testing/internal/types"
)

var (
	// Supported middleware types
	supportedMiddlewares = map[string]bool{
		"redis":     true,
		"kafka":     true,
		"mongodb":   true,
		"rocketmq":  true,
		"rabbitmq":  true,
		"emqx":      true,
		"nacos":     true,
	}
)

// TaskService handles task management operations
type TaskService struct {
	taskStore   storage.TaskStore
	resultStore storage.ResultStore
	executor    executor.MCTExecutor
	logger      *zap.Logger
}

// NewTaskService creates a new task service
func NewTaskService(taskStore storage.TaskStore, resultStore storage.ResultStore, executor executor.MCTExecutor) *TaskService {
	logger, _ := zap.NewDevelopment()
	return &TaskService{
		taskStore:   taskStore,
		resultStore: resultStore,
		executor:    executor,
		logger:      logger,
	}
}

// CreateTask creates a new test task
func (s *TaskService) CreateTask(ctx context.Context, req *types.TaskRequest) (*types.Task, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Validate middleware
	if !supportedMiddlewares[req.Middleware] {
		return nil, fmt.Errorf("unsupported middleware: %s", req.Middleware)
	}

	// Convert config to JSON string
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	// Create task
	now := time.Now()
	task := &types.Task{
		ID:         uuid.New().String(),
		Middleware: req.Middleware,
		Status:     types.TaskStatusPending,
		Config:     string(configJSON),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Store task
	if err := s.taskStore.CreateTask(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	s.logger.Info("Task created",
		zap.String("task_id", task.ID),
		zap.String("middleware", task.Middleware),
	)

	return task, nil
}

// GetTask retrieves a task by ID
func (s *TaskService) GetTask(ctx context.Context, taskID string) (*types.Task, error) {
	task, err := s.taskStore.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// ListTasks retrieves a paginated list of tasks with optional filters
func (s *TaskService) ListTasks(ctx context.Context, filter types.TaskFilter) (*types.TaskListResponse, error) {
	resp, err := s.taskStore.ListTasks(ctx, filter)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// DeleteTask deletes a task and its associated result
func (s *TaskService) DeleteTask(ctx context.Context, taskID string) error {
	// Delete result if exists
	if s.resultStore != nil {
		_ = s.resultStore.DeleteResult(ctx, taskID) // Ignore error if not found
	}

	// Delete task
	if err := s.taskStore.DeleteTask(ctx, taskID); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	s.logger.Info("Task deleted", zap.String("task_id", taskID))
	return nil
}

// RunTask executes a test task
func (s *TaskService) RunTask(ctx context.Context, task *types.Task) (*types.TestResult, error) {
	s.logger.Info("Starting task execution",
		zap.String("task_id", task.ID),
		zap.String("middleware", task.Middleware),
	)

	// Update task status to running
	task.Status = types.TaskStatusRunning
	task.StartTime = time.Now()
	if err := s.taskStore.UpdateTaskStatus(ctx, task.ID, types.TaskStatusRunning); err != nil {
		return nil, fmt.Errorf("failed to update task status: %w", err)
	}

	// Execute task
	result, err := s.executor.Execute(ctx, task)
	if err != nil {
		// Update task status to failed
		task.Status = types.TaskStatusFailed
		task.EndTime = time.Now()
		task.ErrorMsg = err.Error()
		_ = s.taskStore.UpdateTask(ctx, task)
		_ = s.taskStore.UpdateTaskStatus(ctx, task.ID, types.TaskStatusFailed)

		s.logger.Error("Task execution failed",
			zap.String("task_id", task.ID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("task execution failed: %w", err)
	}

	// Update task status to completed
	task.Status = types.TaskStatusCompleted
	task.EndTime = time.Now()
	if err := s.taskStore.UpdateTask(ctx, task); err != nil {
		s.logger.Warn("Failed to update task",
			zap.String("task_id", task.ID),
			zap.Error(err),
		)
	}
	if err := s.taskStore.UpdateTaskStatus(ctx, task.ID, types.TaskStatusCompleted); err != nil {
		s.logger.Warn("Failed to update task status",
			zap.String("task_id", task.ID),
			zap.Error(err),
		)
	}

	// Save result if result store is available
	if s.resultStore != nil {
		if err := s.resultStore.SaveResult(ctx, result); err != nil {
			s.logger.Warn("Failed to save result",
				zap.String("task_id", task.ID),
				zap.Error(err),
			)
		}
	}

	s.logger.Info("Task completed successfully",
		zap.String("task_id", task.ID),
		zap.Float64("score", result.Score),
		zap.String("grade", string(result.Grade)),
	)

	return result, nil
}

// GetTaskResult retrieves the test result for a task
func (s *TaskService) GetTaskResult(ctx context.Context, taskID string) (*types.TestResult, error) {
	if s.resultStore == nil {
		return nil, errors.New("result store not configured")
	}

	result, err := s.resultStore.GetResult(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return result, nil
}
