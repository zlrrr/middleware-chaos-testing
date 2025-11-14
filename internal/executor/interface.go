package executor

import (
	"context"

	"middleware-chaos-testing/internal/types"
)

// MCTExecutor defines the interface for executing mct binary
type MCTExecutor interface {
	// Execute runs the mct binary with the given task configuration
	Execute(ctx context.Context, task *types.Task) (*types.TestResult, error)
}

// ExecutorConfig holds configuration for the MCT executor
type ExecutorConfig struct {
	BinaryPath string // Path to mct binary
	WorkDir    string // Working directory for test execution
}
