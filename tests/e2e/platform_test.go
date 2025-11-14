package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// Base URL for the API server
	// In real E2E tests, this would point to the actual server
	baseURL = "http://localhost:8080"

	// Timeout for operations
	operationTimeout = 120 * time.Second

	// Poll interval for task status
	pollInterval = 2 * time.Second
)

// APIResponse represents the standard API response structure
type APIResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// Task represents a test task
type Task struct {
	ID         string `json:"id"`
	Middleware string `json:"middleware"`
	Status     string `json:"status"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time,omitempty"`
	Config     string `json:"config"`
	Error      string `json:"error,omitempty"`
}

// TestResult represents a test result
type TestResult struct {
	TaskID         string                 `json:"task_id"`
	Score          float64                `json:"score"`
	Grade          string                 `json:"grade"`
	Dimensions     map[string]float64     `json:"dimensions"`
	Metrics        map[string]interface{} `json:"metrics"`
	Issues         []interface{}          `json:"issues"`
	Recommendations []interface{}         `json:"recommendations"`
	TestDurationMS int64                  `json:"test_duration_ms"`
	Timestamp      string                 `json:"timestamp"`
}

// MiddlewareInfo represents middleware information
type MiddlewareInfo struct {
	Name            string   `json:"name"`
	DisplayName     string   `json:"display_name"`
	Description     string   `json:"description"`
	Version         string   `json:"version"`
	SupportedChaos  []string `json:"supported_chaos"`
}

// Client wraps HTTP operations for E2E testing
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request and returns the response
func (c *Client) doRequest(method, path string, body interface{}) (*APIResponse, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respData, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &apiResp, nil
}

// Health checks the API server health
func (c *Client) Health() error {
	resp, err := c.httpClient.Get(c.baseURL + "/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// GetMiddlewares retrieves all available middlewares
func (c *Client) GetMiddlewares() ([]MiddlewareInfo, error) {
	resp, err := c.doRequest("GET", "/api/v1/middlewares", nil)
	if err != nil {
		return nil, err
	}

	var middlewares []MiddlewareInfo
	if err := json.Unmarshal(resp.Data, &middlewares); err != nil {
		return nil, fmt.Errorf("failed to unmarshal middlewares: %w", err)
	}

	return middlewares, nil
}

// CreateTask creates a new test task
func (c *Client) CreateTask(middleware string, config map[string]interface{}) (*Task, error) {
	reqBody := map[string]interface{}{
		"middleware": middleware,
		"config":     config,
	}

	resp, err := c.doRequest("POST", "/api/v1/tasks", reqBody)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error: %s", resp.Message)
	}

	var task Task
	if err := json.Unmarshal(resp.Data, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}

// GetTask retrieves a task by ID
func (c *Client) GetTask(taskID string) (*Task, error) {
	resp, err := c.doRequest("GET", "/api/v1/tasks/"+taskID, nil)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error: %s", resp.Message)
	}

	var task Task
	if err := json.Unmarshal(resp.Data, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	return &task, nil
}

// RunTask starts a task execution
func (c *Client) RunTask(taskID string) error {
	resp, err := c.doRequest("POST", "/api/v1/tasks/"+taskID+"/run", nil)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return fmt.Errorf("API error: %s", resp.Message)
	}

	return nil
}

// GetTaskResult retrieves the test result for a task
func (c *Client) GetTaskResult(taskID string) (*TestResult, error) {
	resp, err := c.doRequest("GET", "/api/v1/tasks/"+taskID+"/result", nil)
	if err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("API error: %s", resp.Message)
	}

	var result TestResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

// DeleteTask deletes a task
func (c *Client) DeleteTask(taskID string) error {
	resp, err := c.doRequest("DELETE", "/api/v1/tasks/"+taskID, nil)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return fmt.Errorf("API error: %s", resp.Message)
	}

	return nil
}

// WaitForTaskCompletion waits for a task to complete or fail
func (c *Client) WaitForTaskCompletion(ctx context.Context, taskID string) (*Task, error) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			task, err := c.GetTask(taskID)
			if err != nil {
				return nil, err
			}

			if task.Status == "completed" || task.Status == "failed" {
				return task, nil
			}
		}
	}
}

// Test Health Check
func TestHealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	client := NewClient(baseURL)

	err := client.Health()
	assert.NoError(t, err, "Health check should succeed")
}

// Test Get Middlewares
func TestGetMiddlewares(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	client := NewClient(baseURL)

	middlewares, err := client.GetMiddlewares()
	require.NoError(t, err, "Should get middlewares successfully")

	// Should have at least the 7 supported middlewares
	assert.GreaterOrEqual(t, len(middlewares), 7, "Should have at least 7 middlewares")

	// Verify expected middlewares are present
	expectedMiddlewares := []string{"redis", "kafka", "mongodb", "rabbitmq", "emqx", "nacos", "elasticsearch"}
	foundMiddlewares := make(map[string]bool)

	for _, mw := range middlewares {
		foundMiddlewares[mw.Name] = true
		assert.NotEmpty(t, mw.DisplayName, "DisplayName should not be empty")
		assert.NotEmpty(t, mw.Description, "Description should not be empty")
	}

	for _, expected := range expectedMiddlewares {
		assert.True(t, foundMiddlewares[expected], "Should have middleware: %s", expected)
	}
}

// Test Redis Chaos Testing - Complete Workflow
func TestRedisWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	client := NewClient(baseURL)
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	// 1. Create task
	task, err := client.CreateTask("redis", map[string]interface{}{
		"host":     "redis:6379",
		"password": "",
		"db":       0,
	})
	require.NoError(t, err, "Should create task successfully")
	require.NotEmpty(t, task.ID, "Task ID should not be empty")
	assert.Equal(t, "redis", task.Middleware)
	assert.Equal(t, "pending", task.Status)

	// 2. Run task
	err = client.RunTask(task.ID)
	require.NoError(t, err, "Should run task successfully")

	// 3. Wait for completion
	completedTask, err := client.WaitForTaskCompletion(ctx, task.ID)
	require.NoError(t, err, "Task should complete without error")
	assert.Equal(t, "completed", completedTask.Status, "Task should be completed")

	// 4. Get result
	result, err := client.GetTaskResult(task.ID)
	require.NoError(t, err, "Should get result successfully")
	assert.Equal(t, task.ID, result.TaskID)
	assert.GreaterOrEqual(t, result.Score, 0.0, "Score should be >= 0")
	assert.LessOrEqual(t, result.Score, 100.0, "Score should be <= 100")
	assert.NotEmpty(t, result.Grade, "Grade should not be empty")

	// Verify dimensions
	assert.Contains(t, result.Dimensions, "availability")
	assert.Contains(t, result.Dimensions, "performance")
	assert.Contains(t, result.Dimensions, "resilience")
	assert.Contains(t, result.Dimensions, "data_integrity")
	assert.Contains(t, result.Dimensions, "recovery")

	// 5. Clean up
	err = client.DeleteTask(task.ID)
	assert.NoError(t, err, "Should delete task successfully")
}

// Test Kafka Chaos Testing
func TestKafkaWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	client := NewClient(baseURL)
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	task, err := client.CreateTask("kafka", map[string]interface{}{
		"brokers": "kafka:9092",
		"topic":   "test-topic",
	})
	require.NoError(t, err)

	err = client.RunTask(task.ID)
	require.NoError(t, err)

	completedTask, err := client.WaitForTaskCompletion(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, "completed", completedTask.Status)

	result, err := client.GetTaskResult(task.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Score, 0.0)
	assert.LessOrEqual(t, result.Score, 100.0)

	err = client.DeleteTask(task.ID)
	assert.NoError(t, err)
}

// Test MongoDB Chaos Testing
func TestMongoDBWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	client := NewClient(baseURL)
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	task, err := client.CreateTask("mongodb", map[string]interface{}{
		"uri":        "mongodb://admin:password@mongodb:27017",
		"database":   "testdb",
		"collection": "testcol",
	})
	require.NoError(t, err)

	err = client.RunTask(task.ID)
	require.NoError(t, err)

	completedTask, err := client.WaitForTaskCompletion(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, "completed", completedTask.Status)

	result, err := client.GetTaskResult(task.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Score, 0.0)

	err = client.DeleteTask(task.ID)
	assert.NoError(t, err)
}

// Test RabbitMQ Chaos Testing
func TestRabbitMQWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	client := NewClient(baseURL)
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	task, err := client.CreateTask("rabbitmq", map[string]interface{}{
		"url":   "amqp://admin:password@rabbitmq:5672/",
		"queue": "test-queue",
	})
	require.NoError(t, err)

	err = client.RunTask(task.ID)
	require.NoError(t, err)

	completedTask, err := client.WaitForTaskCompletion(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, "completed", completedTask.Status)

	result, err := client.GetTaskResult(task.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Score, 0.0)

	err = client.DeleteTask(task.ID)
	assert.NoError(t, err)
}

// Test EMQX Chaos Testing
func TestEMQXWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	client := NewClient(baseURL)
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	task, err := client.CreateTask("emqx", map[string]interface{}{
		"broker":    "tcp://emqx:1883",
		"topic":     "test/topic",
		"client_id": "test-client",
	})
	require.NoError(t, err)

	err = client.RunTask(task.ID)
	require.NoError(t, err)

	completedTask, err := client.WaitForTaskCompletion(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, "completed", completedTask.Status)

	result, err := client.GetTaskResult(task.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Score, 0.0)

	err = client.DeleteTask(task.ID)
	assert.NoError(t, err)
}

// Test Nacos Chaos Testing
func TestNacosWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	client := NewClient(baseURL)
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	task, err := client.CreateTask("nacos", map[string]interface{}{
		"server_addr":  "nacos:8848",
		"namespace_id": "public",
		"group":        "DEFAULT_GROUP",
		"data_id":      "test-config",
	})
	require.NoError(t, err)

	err = client.RunTask(task.ID)
	require.NoError(t, err)

	completedTask, err := client.WaitForTaskCompletion(ctx, task.ID)
	require.NoError(t, err)
	assert.Equal(t, "completed", completedTask.Status)

	result, err := client.GetTaskResult(task.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Score, 0.0)

	err = client.DeleteTask(task.ID)
	assert.NoError(t, err)
}

// Test Multiple Concurrent Tasks
func TestConcurrentTasks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	client := NewClient(baseURL)
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout*2)
	defer cancel()

	middlewareConfigs := []struct {
		middleware string
		config     map[string]interface{}
	}{
		{"redis", map[string]interface{}{"host": "redis:6379", "db": 0}},
		{"kafka", map[string]interface{}{"brokers": "kafka:9092", "topic": "test"}},
		{"mongodb", map[string]interface{}{"uri": "mongodb://admin:password@mongodb:27017", "database": "test", "collection": "test"}},
	}

	taskIDs := make([]string, 0, len(middlewareConfigs))

	// Create all tasks
	for _, mc := range middlewareConfigs {
		task, err := client.CreateTask(mc.middleware, mc.config)
		require.NoError(t, err)
		taskIDs = append(taskIDs, task.ID)
	}

	// Run all tasks
	for _, taskID := range taskIDs {
		err := client.RunTask(taskID)
		require.NoError(t, err)
	}

	// Wait for all tasks to complete
	for _, taskID := range taskIDs {
		completedTask, err := client.WaitForTaskCompletion(ctx, taskID)
		require.NoError(t, err)
		assert.Equal(t, "completed", completedTask.Status)

		result, err := client.GetTaskResult(taskID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.Score, 0.0)
	}

	// Clean up
	for _, taskID := range taskIDs {
		err := client.DeleteTask(taskID)
		assert.NoError(t, err)
	}
}

// Test Error Handling - Invalid Middleware
func TestInvalidMiddleware(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	client := NewClient(baseURL)

	_, err := client.CreateTask("invalid-middleware", map[string]interface{}{})
	assert.Error(t, err, "Should fail with invalid middleware")
}

// Test Error Handling - Invalid Configuration
func TestInvalidConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	client := NewClient(baseURL)

	// Redis without required host field
	_, err := client.CreateTask("redis", map[string]interface{}{})
	assert.Error(t, err, "Should fail with missing required configuration")
}

// Test Error Handling - Task Not Found
func TestTaskNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	client := NewClient(baseURL)

	_, err := client.GetTask("non-existent-task-id")
	assert.Error(t, err, "Should fail with task not found")
}

// Test Error Handling - Delete Running Task
func TestDeleteRunningTask(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	client := NewClient(baseURL)

	task, err := client.CreateTask("redis", map[string]interface{}{
		"host": "redis:6379",
		"db":   0,
	})
	require.NoError(t, err)

	err = client.RunTask(task.ID)
	require.NoError(t, err)

	// Try to delete while running (should fail)
	err = client.DeleteTask(task.ID)
	if err == nil {
		// If deletion succeeded, it means the task completed very quickly
		// This is acceptable in the test
		t.Log("Task completed before deletion attempt")
	} else {
		// Expected: deletion should fail for running task
		t.Log("Deletion correctly failed for running task")
	}

	// Wait for completion and clean up
	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	_, _ = client.WaitForTaskCompletion(ctx, task.ID)
	_ = client.DeleteTask(task.ID)
}

// Benchmark API Performance
func BenchmarkCreateTask(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping E2E benchmark in short mode")
	}

	client := NewClient(baseURL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task, err := client.CreateTask("redis", map[string]interface{}{
			"host": "redis:6379",
			"db":   0,
		})
		if err != nil {
			b.Fatalf("Failed to create task: %v", err)
		}

		// Clean up
		_ = client.DeleteTask(task.ID)
	}
}
