package types

import (
	"errors"
	"fmt"
	"time"

	"middleware-chaos-testing/internal/core"
)

// Validate validates a task request
func (r *TaskRequest) Validate() error {
	if r.Middleware == "" {
		return errors.New("middleware is required")
	}
	if r.Duration == "" {
		return errors.New("duration is required")
	}
	if len(r.Config) == 0 {
		return errors.New("config is required")
	}

	// Validate duration format
	_, err := time.ParseDuration(r.Duration)
	if err != nil {
		return fmt.Errorf("invalid duration format: %w", err)
	}

	return nil
}

// TaskStatus represents the status of a test task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

// TaskRequest represents a request to create a test task
type TaskRequest struct {
	Middleware  string            `json:"middleware" binding:"required"`
	Config      map[string]string `json:"config" binding:"required"`
	Duration    string            `json:"duration" binding:"required"`
	Operations  int               `json:"operations"`
	Concurrency int               `json:"concurrency"`
}

// Task represents a test task
type Task struct {
	ID          string     `json:"id"`
	Middleware  string     `json:"middleware"`
	Status      TaskStatus `json:"status"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     time.Time  `json:"end_time,omitempty"`
	Config      string     `json:"config"` // JSON string
	ResultPath  string     `json:"result_path,omitempty"`
	ErrorMsg    string     `json:"error_msg,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// DimensionScores represents scores across different dimensions
type DimensionScores struct {
	Availability float64 `json:"availability"` // 可用性得分 (30分)
	Performance  float64 `json:"performance"`  // 性能得分 (25分)
	Reliability  float64 `json:"reliability"`  // 可靠性得分 (25分)
	Resilience   float64 `json:"resilience"`   // 恢复力得分 (20分)
}

// TestResult represents the result of a test execution
type TestResult struct {
	TaskID          string                   `json:"task_id"`
	Middleware      string                   `json:"middleware"`
	Score           float64                  `json:"score"`
	Grade           core.StabilityGrade      `json:"grade"`
	Status          core.TestStatus          `json:"status"`
	Metrics         *core.StabilityMetrics   `json:"metrics"`
	Issues          []core.Issue             `json:"issues"`
	Recommendations []core.Recommendation    `json:"recommendations"`
	Rationale       string                   `json:"rationale"`
	StartTime       time.Time                `json:"start_time"`
	EndTime         time.Time                `json:"end_time"`
	Duration        time.Duration            `json:"duration"`
	Scores          DimensionScores          `json:"scores"`
}

// TaskFilter represents filter criteria for listing tasks
type TaskFilter struct {
	Middleware string
	Status     TaskStatus
	Page       int
	PageSize   int
}

// TaskListResponse represents a paginated list of tasks
type TaskListResponse struct {
	Tasks      []*Task `json:"tasks"`
	Total      int64   `json:"total"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	TotalPages int     `json:"total_pages"`
}
