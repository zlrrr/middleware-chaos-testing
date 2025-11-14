package handlers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"middleware-chaos-testing/internal/executor"
)

// MiddlewareHandler handles middleware metadata requests
type MiddlewareHandler struct {
	logger *zap.Logger
}

// NewMiddlewareHandler creates a new middleware metadata handler
func NewMiddlewareHandler(logger *zap.Logger) *MiddlewareHandler {
	return &MiddlewareHandler{
		logger: logger,
	}
}

// MiddlewareInfo represents metadata about a supported middleware
type MiddlewareInfo struct {
	Name        string                 `json:"name"`
	DisplayName string                 `json:"display_name"`
	Description string                 `json:"description"`
	ConfigSpec  map[string]ConfigField `json:"config_spec"`
}

// ConfigField represents a configuration field specification
type ConfigField struct {
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description"`
	Example     string      `json:"example,omitempty"`
}

// ListMiddlewares handles GET /api/v1/middlewares
func (h *MiddlewareHandler) ListMiddlewares(c *gin.Context) {
	middlewares := []MiddlewareInfo{
		{
			Name:        "redis",
			DisplayName: "Redis",
			Description: "Redis in-memory data store",
			ConfigSpec: map[string]ConfigField{
				"host": {
					Type:        "string",
					Required:    true,
					Default:     "localhost",
					Description: "Redis server host",
					Example:     "localhost",
				},
				"port": {
					Type:        "string",
					Required:    true,
					Default:     "6379",
					Description: "Redis server port",
					Example:     "6379",
				},
				"password": {
					Type:        "string",
					Required:    false,
					Description: "Redis password (optional)",
				},
				"db": {
					Type:        "number",
					Required:    false,
					Default:     0,
					Description: "Redis database number",
				},
			},
		},
		{
			Name:        "kafka",
			DisplayName: "Apache Kafka",
			Description: "Distributed event streaming platform",
			ConfigSpec: map[string]ConfigField{
				"brokers": {
					Type:        "string",
					Required:    true,
					Description: "Kafka broker addresses (comma-separated)",
					Example:     "localhost:9092",
				},
				"topic": {
					Type:        "string",
					Required:    true,
					Description: "Kafka topic name",
					Example:     "test-topic",
				},
			},
		},
		{
			Name:        "mongodb",
			DisplayName: "MongoDB",
			Description: "Document database",
			ConfigSpec: map[string]ConfigField{
				"uri": {
					Type:        "string",
					Required:    true,
					Description: "MongoDB connection URI",
					Example:     "mongodb://localhost:27017",
				},
				"database": {
					Type:        "string",
					Required:    true,
					Description: "Database name",
					Example:     "test",
				},
				"collection": {
					Type:        "string",
					Required:    true,
					Description: "Collection name",
					Example:     "test_collection",
				},
			},
		},
		{
			Name:        "rocketmq",
			DisplayName: "Apache RocketMQ",
			Description: "Distributed messaging and streaming platform",
			ConfigSpec: map[string]ConfigField{
				"nameserver": {
					Type:        "string",
					Required:    true,
					Description: "RocketMQ nameserver address",
					Example:     "localhost:9876",
				},
				"topic": {
					Type:        "string",
					Required:    true,
					Description: "Topic name",
					Example:     "test-topic",
				},
				"group": {
					Type:        "string",
					Required:    true,
					Description: "Consumer group name",
					Example:     "test-group",
				},
			},
		},
		{
			Name:        "rabbitmq",
			DisplayName: "RabbitMQ",
			Description: "Message broker",
			ConfigSpec: map[string]ConfigField{
				"url": {
					Type:        "string",
					Required:    true,
					Description: "RabbitMQ connection URL",
					Example:     "amqp://guest:guest@localhost:5672/",
				},
				"queue": {
					Type:        "string",
					Required:    true,
					Description: "Queue name",
					Example:     "test-queue",
				},
			},
		},
		{
			Name:        "emqx",
			DisplayName: "EMQX",
			Description: "MQTT broker",
			ConfigSpec: map[string]ConfigField{
				"broker": {
					Type:        "string",
					Required:    true,
					Description: "MQTT broker address",
					Example:     "tcp://localhost:1883",
				},
				"topic": {
					Type:        "string",
					Required:    true,
					Description: "MQTT topic",
					Example:     "test/topic",
				},
				"client_id": {
					Type:        "string",
					Required:    false,
					Description: "MQTT client ID (auto-generated if not provided)",
				},
			},
		},
		{
			Name:        "nacos",
			DisplayName: "Nacos",
			Description: "Dynamic naming and configuration service",
			ConfigSpec: map[string]ConfigField{
				"server_addr": {
					Type:        "string",
					Required:    true,
					Description: "Nacos server address",
					Example:     "localhost:8848",
				},
				"namespace": {
					Type:        "string",
					Required:    false,
					Default:     "public",
					Description: "Namespace ID",
				},
				"group": {
					Type:        "string",
					Required:    false,
					Default:     "DEFAULT_GROUP",
					Description: "Group name",
				},
			},
		},
	}

	SuccessResponse(c, middlewares)
}

// GetMiddlewareConfig handles GET /api/v1/middlewares/:name/config
func (h *MiddlewareHandler) GetMiddlewareConfig(c *gin.Context) {
	name := c.Param("name")

	// Validate middleware exists
	if !executor.IsValidMiddleware(name) {
		NotFoundError(c, "Middleware not found")
		return
	}

	// For simplicity, return the same info from the list
	// In a real application, you might want to store this in a database
	h.ListMiddlewares(c)
}

// RegisterRoutes registers middleware metadata routes
func (h *MiddlewareHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/middlewares", h.ListMiddlewares)
	r.GET("/middlewares/:name/config", h.GetMiddlewareConfig)
}
