package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

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

// mctExecutor implements MCTExecutor interface
type mctExecutor struct {
	config *ExecutorConfig
	logger *zap.Logger
}

// NewMCTExecutor creates a new MCT executor
func NewMCTExecutor(config *ExecutorConfig) MCTExecutor {
	logger, _ := zap.NewDevelopment()
	return &mctExecutor{
		config: config,
		logger: logger,
	}
}

// Execute runs the mct binary with the given task configuration
func (e *mctExecutor) Execute(ctx context.Context, task *types.Task) (*types.TestResult, error) {
	// Create work directory
	workDir, err := e.CreateWorkDir(task)
	if err != nil {
		return nil, fmt.Errorf("failed to create work directory: %w", err)
	}
	// Note: In production, you might want to keep workDir for logs
	// defer e.CleanupWorkDir(workDir)

	// Build command arguments
	args, err := e.BuildArgs(task)
	if err != nil {
		return nil, fmt.Errorf("failed to build arguments: %w", err)
	}

	// Execute mct binary
	cmd := exec.CommandContext(ctx, e.config.BinaryPath, args...)
	cmd.Dir = workDir

	e.logger.Info("Executing MCT binary",
		zap.String("binary", e.config.BinaryPath),
		zap.Strings("args", args),
		zap.String("workDir", workDir),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it's a context cancellation
		if ctx.Err() != nil {
			return nil, fmt.Errorf("execution cancelled: %w", ctx.Err())
		}
		return nil, fmt.Errorf("execution failed: %w, output: %s", err, string(output))
	}

	// Parse output
	result, err := e.ParseOutput(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse output: %w", err)
	}

	// Set task ID
	result.TaskID = task.ID

	return result, nil
}

// BuildArgs builds command-line arguments from task configuration
func (e *mctExecutor) BuildArgs(task *types.Task) ([]string, error) {
	args := []string{"test"}

	// Add middleware
	args = append(args, "--middleware", task.Middleware)

	// Parse config
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(task.Config), &config); err != nil {
		return nil, fmt.Errorf("invalid config JSON: %w", err)
	}

	// Add config parameters
	for key, value := range config {
		// Convert value to string
		var valueStr string
		switch v := value.(type) {
		case string:
			valueStr = v
		case float64:
			valueStr = fmt.Sprintf("%.0f", v)
		case bool:
			valueStr = fmt.Sprintf("%t", v)
		default:
			valueStr = fmt.Sprintf("%v", v)
		}

		args = append(args, fmt.Sprintf("--%s", key), valueStr)
	}

	// Always request JSON output
	args = append(args, "--output", "json")

	return args, nil
}

// ParseOutput parses the JSON output from mct binary
func (e *mctExecutor) ParseOutput(output []byte) (*types.TestResult, error) {
	var result types.TestResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return &result, nil
}

// CreateWorkDir creates a work directory for the task
func (e *mctExecutor) CreateWorkDir(task *types.Task) (string, error) {
	workDir := filepath.Join(e.config.WorkDir, task.ID)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}
	return workDir, nil
}

// CleanupWorkDir removes the work directory
func (e *mctExecutor) CleanupWorkDir(workDir string) error {
	return os.RemoveAll(workDir)
}

// BuildConfigFile creates a temporary config file for the task
func (e *mctExecutor) BuildConfigFile(task *types.Task, workDir string) (string, error) {
	// Parse config
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(task.Config), &config); err != nil {
		return "", fmt.Errorf("invalid config JSON: %w", err)
	}

	// Build YAML config
	var yaml strings.Builder
	yaml.WriteString(fmt.Sprintf("middleware: %s\n", task.Middleware))
	yaml.WriteString("\nconnection:\n")
	for key, value := range config {
		yaml.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
	}

	// Write to file
	configPath := filepath.Join(workDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yaml.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to write config file: %w", err)
	}

	return configPath, nil
}

// Validate validates the executor configuration
func (c *ExecutorConfig) Validate() error {
	if c.BinaryPath == "" {
		return errors.New("binary path is required")
	}
	if c.WorkDir == "" {
		return errors.New("work directory is required")
	}
	return nil
}

// IsValidMiddleware checks if a middleware type is supported
func IsValidMiddleware(middleware string) bool {
	return supportedMiddlewares[middleware]
}
