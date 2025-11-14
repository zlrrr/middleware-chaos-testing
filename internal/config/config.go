package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// TestConfig represents the complete test configuration
type TestConfig struct {
	Name       string            `yaml:"name"`
	Middleware string            `yaml:"middleware"`
	Connection ConnectionConfig  `yaml:"connection"`
	Test       TestParameters    `yaml:"test"`
	Thresholds *ThresholdsConfig `yaml:"thresholds,omitempty"`
	Output     OutputConfig      `yaml:"output"`
}

// ConnectionConfig represents middleware connection settings
type ConnectionConfig struct {
	Host     string        `yaml:"host"`
	Port     int           `yaml:"port"`
	Password string        `yaml:"password,omitempty"`
	DB       int           `yaml:"db,omitempty"`
	Timeout  time.Duration `yaml:"timeout"`

	// Kafka specific
	Brokers []string `yaml:"brokers,omitempty"`
	Topic   string   `yaml:"topic,omitempty"`
	GroupID string   `yaml:"group_id,omitempty"`

	// MongoDB specific
	URI        string `yaml:"uri,omitempty"`
	Database   string `yaml:"database,omitempty"`
	Collection string `yaml:"collection,omitempty"`

	// RocketMQ specific
	NameServers   []string `yaml:"nameservers,omitempty"`
	ProducerGroup string   `yaml:"producer_group,omitempty"`
	ConsumerGroup string   `yaml:"consumer_group,omitempty"`

	// RabbitMQ specific
	URL      string `yaml:"url,omitempty"`
	Queue    string `yaml:"queue,omitempty"`
	Exchange string `yaml:"exchange,omitempty"`

	// EMQX specific
	Broker   string `yaml:"broker,omitempty"`
	ClientID string `yaml:"client_id,omitempty"`
	QoS      int    `yaml:"qos,omitempty"`

	// Nacos specific
	ServerAddr  string `yaml:"server_addr,omitempty"`
	NamespaceID string `yaml:"namespace_id,omitempty"`
	Group       string `yaml:"group,omitempty"`
	DataID      string `yaml:"data_id,omitempty"`
}

// TestParameters represents test execution parameters
type TestParameters struct {
	Duration    time.Duration       `yaml:"duration"`
	Operations  int                 `yaml:"operations"`
	Concurrency int                 `yaml:"concurrency"`
	Workload    []WorkloadOperation `yaml:"workload,omitempty"`
}

// WorkloadOperation represents a single operation in the workload
type WorkloadOperation struct {
	Operation   string `yaml:"operation"`
	Weight      int    `yaml:"weight"`
	KeyPattern  string `yaml:"key_pattern,omitempty"`
	ValueSize   int    `yaml:"value_size,omitempty"`
}

// ThresholdsConfig represents evaluation thresholds
type ThresholdsConfig struct {
	Availability ThresholdLevels `yaml:"availability"`
	P95Latency   ThresholdLevels `yaml:"p95_latency"`
	P99Latency   ThresholdLevels `yaml:"p99_latency"`
	ErrorRate    ThresholdLevels `yaml:"error_rate"`
}

// ThresholdLevels represents the four quality levels
type ThresholdLevels struct {
	Excellent interface{} `yaml:"excellent"`
	Good      interface{} `yaml:"good"`
	Fair      interface{} `yaml:"fair"`
	Pass      interface{} `yaml:"pass"`
}

// OutputConfig represents output settings
type OutputConfig struct {
	Format                  string `yaml:"format"` // console, json, markdown
	Path                    string `yaml:"path,omitempty"`
	IncludeRecommendations bool   `yaml:"include_recommendations"`
}

// LoadFromFile loads configuration from a YAML file
func LoadFromFile(path string) (*TestConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config TestConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Validate checks if the configuration is valid
func (c *TestConfig) Validate() error {
	if c.Middleware == "" {
		return fmt.Errorf("middleware type is required")
	}

	validMiddlewares := map[string]bool{
		"redis": true, "kafka": true, "mongodb": true,
		"rocketmq": true, "rabbitmq": true, "emqx": true, "nacos": true,
	}

	if !validMiddlewares[c.Middleware] {
		return fmt.Errorf("unsupported middleware: %s", c.Middleware)
	}

	if c.Test.Duration <= 0 {
		return fmt.Errorf("test duration must be positive")
	}

	if c.Test.Operations <= 0 {
		return fmt.Errorf("operations must be positive")
	}

	// Validate middleware-specific connection settings
	switch c.Middleware {
	case "redis":
		if c.Connection.Host == "" || c.Connection.Port == 0 {
			return fmt.Errorf("redis requires host and port")
		}
	case "kafka":
		if len(c.Connection.Brokers) == 0 && c.Connection.Host == "" {
			return fmt.Errorf("kafka requires brokers or host")
		}
	case "mongodb":
		if c.Connection.URI == "" && c.Connection.Host == "" {
			return fmt.Errorf("mongodb requires uri or host")
		}
	}

	return nil
}

// ToConnectionString converts connection config to a connection string
func (c *ConnectionConfig) ToConnectionString() string {
	switch {
	case c.URI != "":
		return c.URI
	case c.URL != "":
		return c.URL
	case c.Broker != "":
		return c.Broker
	case c.ServerAddr != "":
		return c.ServerAddr
	default:
		return fmt.Sprintf("%s:%d", c.Host, c.Port)
	}
}
