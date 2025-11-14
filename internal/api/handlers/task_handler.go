package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"middleware-chaos-testing/internal/service"
	"middleware-chaos-testing/internal/storage"
	"middleware-chaos-testing/internal/types"
)

// TaskHandler handles task-related HTTP requests
type TaskHandler struct {
	taskService *service.TaskService
	logger      *zap.Logger
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(taskService *service.TaskService, logger *zap.Logger) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
		logger:      logger,
	}
}

// CreateTask handles POST /api/v1/tasks
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req types.TaskRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body", zap.Error(err))
		BadRequestError(c, err)
		return
	}

	// Create task
	task, err := h.taskService.CreateTask(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create task", zap.Error(err))
		InternalServerError(c, err)
		return
	}

	h.logger.Info("Task created", zap.String("task_id", task.ID))
	CreatedResponse(c, task)
}

// GetTask handles GET /api/v1/tasks/:id
func (h *TaskHandler) GetTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		BadRequestError(c, errors.New("task ID is required"))
		return
	}

	task, err := h.taskService.GetTask(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, storage.ErrTaskNotFound) {
			NotFoundError(c, "Task not found")
			return
		}
		h.logger.Error("Failed to get task", zap.Error(err), zap.String("task_id", taskID))
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, task)
}

// ListTasks handles GET /api/v1/tasks
func (h *TaskHandler) ListTasks(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	middleware := c.Query("middleware")
	status := c.Query("status")

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	filter := types.TaskFilter{
		Middleware: middleware,
		Status:     types.TaskStatus(status),
		Page:       page,
		PageSize:   pageSize,
	}

	resp, err := h.taskService.ListTasks(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list tasks", zap.Error(err))
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, resp)
}

// DeleteTask handles DELETE /api/v1/tasks/:id
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		BadRequestError(c, errors.New("task ID is required"))
		return
	}

	err := h.taskService.DeleteTask(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, storage.ErrTaskNotFound) {
			NotFoundError(c, "Task not found")
			return
		}
		h.logger.Error("Failed to delete task", zap.Error(err), zap.String("task_id", taskID))
		InternalServerError(c, err)
		return
	}

	SuccessResponseWithMessage(c, "Task deleted successfully", nil)
}

// RunTask handles POST /api/v1/tasks/:id/run
func (h *TaskHandler) RunTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		BadRequestError(c, errors.New("task ID is required"))
		return
	}

	// Get task
	task, err := h.taskService.GetTask(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, storage.ErrTaskNotFound) {
			NotFoundError(c, "Task not found")
			return
		}
		h.logger.Error("Failed to get task", zap.Error(err), zap.String("task_id", taskID))
		InternalServerError(c, err)
		return
	}

	// Check if task is already running
	if task.Status == types.TaskStatusRunning {
		ErrorResponseWithMessage(c, http.StatusConflict, "Task is already running")
		return
	}

	// Run task asynchronously
	go func() {
		ctx := context.Background()
		_, err := h.taskService.RunTask(ctx, task)
		if err != nil {
			h.logger.Error("Task execution failed",
				zap.String("task_id", taskID),
				zap.Error(err),
			)
		}
	}()

	SuccessResponseWithMessage(c, "Task execution started", gin.H{
		"task_id": taskID,
		"status":  types.TaskStatusRunning,
	})
}

// GetTaskResult handles GET /api/v1/tasks/:id/result
func (h *TaskHandler) GetTaskResult(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		BadRequestError(c, errors.New("task ID is required"))
		return
	}

	result, err := h.taskService.GetTaskResult(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, storage.ErrResultNotFound) {
			NotFoundError(c, "Result not found")
			return
		}
		h.logger.Error("Failed to get task result", zap.Error(err), zap.String("task_id", taskID))
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, result)
}

// RegisterRoutes registers task-related routes
func (h *TaskHandler) RegisterRoutes(r *gin.RouterGroup) {
	tasks := r.Group("/tasks")
	{
		tasks.POST("", h.CreateTask)
		tasks.GET("", h.ListTasks)
		tasks.GET("/:id", h.GetTask)
		tasks.DELETE("/:id", h.DeleteTask)
		tasks.POST("/:id/run", h.RunTask)
		tasks.GET("/:id/result", h.GetTaskResult)
	}
}
