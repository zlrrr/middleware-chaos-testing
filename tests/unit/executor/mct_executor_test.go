package executor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"middleware-chaos-testing/internal/executor"
	"middleware-chaos-testing/internal/types"
)

// TestExecutorConfig_Validate tests configuration validation
func TestExecutorConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *executor.ExecutorConfig
		wantErr bool
	}{
		{
			name: "valid_config",
			config: &executor.ExecutorConfig{
				BinaryPath: "/usr/local/bin/mct",
				WorkDir:    "/tmp/mct",
			},
			wantErr: false,
		},
		{
			name: "missing_binary_path",
			config: &executor.ExecutorConfig{
				WorkDir: "/tmp/mct",
			},
			wantErr: true,
		},
		{
			name: "missing_work_dir",
			config: &executor.ExecutorConfig{
				BinaryPath: "/usr/local/bin/mct",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestMCTExecutor_NewMCTExecutor tests executor creation
func TestMCTExecutor_NewMCTExecutor(t *testing.T) {
	config := &executor.ExecutorConfig{
		BinaryPath: "/usr/local/bin/mct",
		WorkDir:    "/tmp/mct",
	}

	exec := executor.NewMCTExecutor(config)
	assert.NotNil(t, exec)
}

// TestMCTExecutor_Execute_InvalidTask tests error handling
func TestMCTExecutor_Execute_InvalidTask(t *testing.T) {
	t.Skip("Skipping - requires actual mct binary for integration testing")

	config := &executor.ExecutorConfig{
		BinaryPath: "/usr/local/bin/mct",
		WorkDir:    t.TempDir(),
	}

	exec := executor.NewMCTExecutor(config)

	// Test with invalid config
	task := &types.Task{
		ID:         "test-task",
		Middleware: "redis",
		Config:     `invalid json`,
	}

	result, err := exec.Execute(context.Background(), task)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestIsValidMiddleware tests middleware validation
func TestIsValidMiddleware(t *testing.T) {
	supported := []string{"redis", "kafka", "mongodb", "rocketmq", "rabbitmq", "emqx", "nacos"}

	for _, mw := range supported {
		t.Run(mw, func(t *testing.T) {
			valid := executor.IsValidMiddleware(mw)
			assert.True(t, valid, "Middleware %s should be supported", mw)
		})
	}

	// Test invalid middleware
	t.Run("invalid", func(t *testing.T) {
		invalid := executor.IsValidMiddleware("invalid")
		assert.False(t, invalid)
	})
}

// TestMCTExecutor_Execute_Timeout tests timeout handling
func TestMCTExecutor_Execute_Timeout(t *testing.T) {
	t.Skip("Skipping - requires actual mct binary for integration testing")

	// This test would require the actual binary to be present
	// and would test timeout behavior in integration tests
}

// TestMCTExecutor_Execute_Success tests successful execution
func TestMCTExecutor_Execute_Success(t *testing.T) {
	t.Skip("Skipping - requires actual mct binary for integration testing")

	// This test would require the actual binary to be present
	// and would test end-to-end execution in integration tests
}
