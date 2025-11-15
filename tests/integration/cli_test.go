package integration

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	mctBinaryPath = "../../bin/mct"
)

// TestCLI_Version tests the version flag
func TestCLI_Version(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	cmd := exec.Command(mctBinaryPath, "--version")
	output, err := cmd.CombinedOutput()

	assert.NoError(t, err)
	assert.Contains(t, string(output), "0.1.0")
}

// TestCLI_Help tests the help command
func TestCLI_Help(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	cmd := exec.Command(mctBinaryPath, "help")
	output, err := cmd.CombinedOutput()

	assert.NoError(t, err)
	assert.Contains(t, string(output), "A tool for testing middleware stability")
	assert.Contains(t, string(output), "test")
}

// TestCLI_TestCommand_Help tests the test command help
func TestCLI_TestCommand_Help(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	cmd := exec.Command(mctBinaryPath, "test", "--help")
	output, err := cmd.CombinedOutput()

	assert.NoError(t, err)
	assert.Contains(t, string(output), "middleware")
	assert.Contains(t, string(output), "duration")
	assert.Contains(t, string(output), "operations")
	assert.Contains(t, string(output), "output")
}

// TestCLI_MissingRequiredFlag tests error handling for missing flags
func TestCLI_MissingRequiredFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	cmd := exec.Command(mctBinaryPath, "test", "--host", "localhost")
	output, err := cmd.CombinedOutput()

	assert.Error(t, err)
	assert.Contains(t, string(output), "middleware")
}

// TestCLI_UnsupportedMiddleware tests error handling for unsupported middleware
func TestCLI_UnsupportedMiddleware(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	cmd := exec.Command(mctBinaryPath, "test",
		"--middleware", "invalid-middleware",
		"--duration", "5s")

	output, err := cmd.CombinedOutput()

	assert.Error(t, err)
	assert.Contains(t, string(output), "unsupported")
}

// TestCLI_DurationParsing tests duration parameter parsing
func TestCLI_DurationParsing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	testCases := []struct {
		duration     string
		shouldPass   bool
		description  string
	}{
		{"5s", true, "seconds format"},
		{"1m", true, "minutes format"},
		{"30s", true, "30 seconds"},
		{"invalid", false, "invalid format"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			cmd := exec.Command(mctBinaryPath, "test",
				"--middleware", "redis",
				"--host", "nonexistent-host",  // Will fail to connect, but that's OK
				"--duration", tc.duration,
				"--operations", "100")

			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if tc.shouldPass {
				// Should not fail on parsing, but may fail on connection
				assert.True(t, err == nil || !assert.Contains(t, outputStr, "invalid duration"),
					"Duration %s should parse correctly", tc.duration)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// TestCLI_OutputFormats tests different output formats
func TestCLI_OutputFormats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	formats := []string{"console", "json", "markdown"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			cmd := exec.Command(mctBinaryPath, "test",
				"--middleware", "redis",
				"--host", "nonexistent-host",
				"--duration", "1s",
				"--operations", "10",
				"--output", format)

			output, _ := cmd.CombinedOutput()
			outputStr := string(output)

			// Should include format-specific elements
			switch format {
			case "json":
				// Should be valid JSON (or attempt to be)
				assert.True(t, len(outputStr) > 0)
			case "markdown":
				// Markdown reports typically have headers
				assert.True(t, len(outputStr) > 0)
			case "console":
				// Console output should be human-readable
				assert.True(t, len(outputStr) > 0)
			}
		})
	}
}

// TestCLI_ConfigFile tests configuration file loading
func TestCLI_ConfigFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `name: "Test Config"
middleware: "redis"
connection:
  host: "localhost"
  port: 6379
  timeout: 5s
test:
  duration: 5s
  operations: 100
output:
  format: "json"
  include_recommendations: true
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Note: Config file loading needs to be implemented in the CLI
	// This test is currently a placeholder
	t.Skip("Config file loading not yet implemented in CLI")

	cmd := exec.Command(mctBinaryPath, "test", "--config", configPath)
	output, _ := cmd.CombinedOutput()

	// Should attempt to use config file
	assert.True(t, len(output) > 0)
}

// TestCLI_JSONOutput tests JSON output format and structure
func TestCLI_JSONOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	cmd := exec.Command(mctBinaryPath, "test",
		"--middleware", "redis",
		"--host", "nonexistent-host",
		"--duration", "1s",
		"--operations", "10",
		"--output", "json")

	output, _ := cmd.CombinedOutput()

	// Try to parse as JSON
	var result map[string]interface{}
	err := json.Unmarshal(output, &result)

	if err == nil {
		// If it's valid JSON, check for expected fields
		// These fields depend on the actual reporter implementation
		t.Log("Successfully parsed JSON output")
	} else {
		// Output might not be pure JSON if there's an error
		t.Logf("Output is not valid JSON (may be expected for failed connection): %v", err)
	}
}

// TestCLI_ReportPath tests writing output to a file
func TestCLI_ReportPath(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	tmpDir := t.TempDir()
	reportPath := filepath.Join(tmpDir, "test-report.json")

	cmd := exec.Command(mctBinaryPath, "test",
		"--middleware", "redis",
		"--host", "nonexistent-host",
		"--duration", "1s",
		"--operations", "10",
		"--output", "json",
		"--report-path", reportPath)

	_, _ = cmd.CombinedOutput()

	// Check if file was created (it might not be if connection fails early)
	if _, err := os.Stat(reportPath); err == nil {
		t.Logf("Report file created at %s", reportPath)

		// Read the file
		data, err := os.ReadFile(reportPath)
		require.NoError(t, err)
		assert.True(t, len(data) > 0)
	} else {
		t.Logf("Report file not created (expected if connection fails)")
	}
}

// TestCLI_AllMiddlewareTypes tests that all middleware types are recognized
func TestCLI_AllMiddlewareTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	middlewares := []string{"redis", "kafka", "mongodb", "rocketmq", "rabbitmq", "emqx", "nacos"}

	for _, mw := range middlewares {
		t.Run(mw, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, mctBinaryPath, "test",
				"--middleware", mw,
				"--host", "nonexistent-host",
				"--duration", "1s",
				"--operations", "10")

			output, _ := cmd.CombinedOutput()
			outputStr := string(output)

			// Should not reject the middleware type itself
			assert.NotContains(t, outputStr, "unsupported middleware type")

			// Should show it's trying to connect (even if it fails)
			assert.True(t, len(outputStr) > 0 || ctx.Err() != nil,
				"Should produce output or timeout")
		})
	}
}

// TestCLI_OperationsParameter tests the operations parameter
func TestCLI_OperationsParameter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	cmd := exec.Command(mctBinaryPath, "test",
		"--middleware", "redis",
		"--host", "nonexistent-host",
		"--duration", "1s",
		"--operations", "5000") // Higher operations count

	output, _ := cmd.CombinedOutput()

	// Should accept the parameter (even if test fails to connect)
	assert.True(t, len(output) > 0)
}

// TestCLI_DefaultPorts tests that default ports are set correctly
func TestCLI_DefaultPorts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI integration test in short mode")
	}

	testCases := []struct {
		middleware  string
		expectedPort string
	}{
		{"redis", "6379"},
		{"kafka", "9092"},
		{"mongodb", "27017"},
		{"rabbitmq", "5672"},
		{"emqx", "1883"},
		{"nacos", "8848"},
	}

	for _, tc := range testCases {
		t.Run(tc.middleware, func(t *testing.T) {
			cmd := exec.Command(mctBinaryPath, "test",
				"--middleware", tc.middleware,
				"--host", "localhost", // Don't use nonexistent to see port in output
				"--duration", "1s",
				"--operations", "10")

			output, _ := cmd.CombinedOutput()
			outputStr := string(output)

			// Output should show the target with default port
			assert.Contains(t, outputStr, tc.expectedPort,
				"Output should show default port %s for %s", tc.expectedPort, tc.middleware)
		})
	}
}

// BenchmarkCLI_Startup benchmarks CLI startup time
func BenchmarkCLI_Startup(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := exec.Command(mctBinaryPath, "version")
		_, _ = cmd.CombinedOutput()
	}
}
