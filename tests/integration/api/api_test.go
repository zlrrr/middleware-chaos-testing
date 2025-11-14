package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"middleware-chaos-testing/internal/api/handlers"
	apimiddleware "middleware-chaos-testing/internal/api/middleware"
	"middleware-chaos-testing/internal/executor"
	"middleware-chaos-testing/internal/service"
	"middleware-chaos-testing/internal/storage"
	"middleware-chaos-testing/internal/types"
)

// setupTestRouter creates a test router with all dependencies
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	logger, _ := zap.NewDevelopment()

	// Initialize storage
	taskStore := storage.NewMemoryTaskStore()
	resultStore := storage.NewMemoryResultStore()

	// Initialize executor (with test binary path)
	execConfig := &executor.ExecutorConfig{
		BinaryPath: "/usr/local/bin/mct",
		WorkDir:    "/tmp/mct-test",
	}
	exec := executor.NewMCTExecutor(execConfig)

	// Initialize service
	taskService := service.NewTaskService(taskStore, resultStore, exec)

	// Setup router
	router := gin.New()
	router.Use(apimiddleware.Recovery(logger))
	router.Use(apimiddleware.Logger(logger))
	router.Use(apimiddleware.CORS())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		taskHandler := handlers.NewTaskHandler(taskService, logger)
		middlewareHandler := handlers.NewMiddlewareHandler(logger)

		taskHandler.RegisterRoutes(v1)
		middlewareHandler.RegisterRoutes(v1)
	}

	return router
}

// TestHealthCheck tests the health check endpoint
func TestHealthCheck(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}

// TestCreateTask tests POST /api/v1/tasks
func TestCreateTask(t *testing.T) {
	router := setupTestRouter()

	taskReq := types.TaskRequest{
		Middleware: "redis",
		Config: map[string]string{
			"host": "localhost",
			"port": "6379",
		},
		Duration:    "30s",
		Operations:  5000,
		Concurrency: 10,
	}

	body, _ := json.Marshal(taskReq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response handlers.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
}

// TestCreateTask_InvalidMiddleware tests validation
func TestCreateTask_InvalidMiddleware(t *testing.T) {
	router := setupTestRouter()

	taskReq := types.TaskRequest{
		Middleware: "invalid",
		Config: map[string]string{
			"host": "localhost",
		},
		Duration: "30s",
	}

	body, _ := json.Marshal(taskReq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response handlers.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "unsupported middleware")
}

// TestGetTask tests GET /api/v1/tasks/:id
func TestGetTask(t *testing.T) {
	router := setupTestRouter()

	// First create a task
	taskReq := types.TaskRequest{
		Middleware: "redis",
		Config: map[string]string{
			"host": "localhost",
			"port": "6379",
		},
		Duration: "30s",
	}

	body, _ := json.Marshal(taskReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var createResponse handlers.Response
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	taskData := createResponse.Data.(map[string]interface{})
	taskID := taskData["id"].(string)

	// Now get the task
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/tasks/"+taskID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
}

// TestGetTask_NotFound tests 404 error
func TestGetTask_NotFound(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tasks/non-existent-id", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response handlers.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "not found")
}

// TestListTasks tests GET /api/v1/tasks
func TestListTasks(t *testing.T) {
	router := setupTestRouter()

	// Create a few tasks first
	for i := 0; i < 3; i++ {
		taskReq := types.TaskRequest{
			Middleware: "redis",
			Config: map[string]string{
				"host": "localhost",
				"port": "6379",
			},
			Duration: "30s",
		}

		body, _ := json.Marshal(taskReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
	}

	// List tasks
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tasks?page=1&page_size=10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)
}

// TestDeleteTask tests DELETE /api/v1/tasks/:id
func TestDeleteTask(t *testing.T) {
	router := setupTestRouter()

	// Create a task first
	taskReq := types.TaskRequest{
		Middleware: "redis",
		Config: map[string]string{
			"host": "localhost",
			"port": "6379",
		},
		Duration: "30s",
	}

	body, _ := json.Marshal(taskReq)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var createResponse handlers.Response
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	taskData := createResponse.Data.(map[string]interface{})
	taskID := taskData["id"].(string)

	// Delete the task
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/v1/tasks/"+taskID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Contains(t, response.Message, "deleted successfully")
}

// TestListMiddlewares tests GET /api/v1/middlewares
func TestListMiddlewares(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/middlewares", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Data)

	// Verify we have all 7 middlewares
	data := response.Data.([]interface{})
	assert.Equal(t, 7, len(data))
}
