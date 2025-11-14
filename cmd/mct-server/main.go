package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"middleware-chaos-testing/internal/api/handlers"
	apimiddleware "middleware-chaos-testing/internal/api/middleware"
	"middleware-chaos-testing/internal/executor"
	"middleware-chaos-testing/internal/service"
	"middleware-chaos-testing/internal/storage"
)

const (
	defaultPort       = "8080"
	defaultMCTBinary  = "./mct"
	defaultWorkDir    = "/tmp/mct-tasks"
	shutdownTimeout   = 10 * time.Second
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Get configuration from environment
	port := getEnv("MCT_SERVER_PORT", defaultPort)
	mctBinary := getEnv("MCT_BINARY_PATH", defaultMCTBinary)
	workDir := getEnv("MCT_WORK_DIR", defaultWorkDir)

	logger.Info("Starting MCT Server",
		zap.String("port", port),
		zap.String("mct_binary", mctBinary),
		zap.String("work_dir", workDir),
	)

	// Initialize storage
	taskStore := storage.NewMemoryTaskStore()
	resultStore := storage.NewMemoryResultStore()

	// Initialize executor
	execConfig := &executor.ExecutorConfig{
		BinaryPath: mctBinary,
		WorkDir:    workDir,
	}
	if err := execConfig.Validate(); err != nil {
		logger.Fatal("Invalid executor configuration", zap.Error(err))
	}
	exec := executor.NewMCTExecutor(execConfig)

	// Initialize service
	taskService := service.NewTaskService(taskStore, resultStore, exec)

	// Setup router
	router := setupRouter(taskService, logger)

	// Create HTTP server
	server := &http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server listening", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

// setupRouter configures the HTTP router with all routes and middleware
func setupRouter(taskService *service.TaskService, logger *zap.Logger) *gin.Engine {
	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Apply global middleware
	router.Use(apimiddleware.Recovery(logger))
	router.Use(apimiddleware.Logger(logger))
	router.Use(apimiddleware.CORS())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().UTC(),
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Initialize handlers
		taskHandler := handlers.NewTaskHandler(taskService, logger)
		middlewareHandler := handlers.NewMiddlewareHandler(logger)

		// Register routes
		taskHandler.RegisterRoutes(v1)
		middlewareHandler.RegisterRoutes(v1)
	}

	return router
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
