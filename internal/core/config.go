package core

import "time"

// Config 配置接口
type Config interface {
	// GetMiddlewareType 获取中间件类型
	GetMiddlewareType() string

	// GetConnectionConfig 获取连接配置
	GetConnectionConfig() *ConnectionConfig

	// GetTestConfig 获取测试配置
	GetTestConfig() *TestConfig

	// GetThresholds 获取阈值配置
	GetThresholds() *Thresholds

	// GetOutputConfig 获取输出配置
	GetOutputConfig() *OutputConfig

	// Validate 验证配置
	Validate() error
}

// ConnectionConfig 连接配置
type ConnectionConfig struct {
	Host     string        // 主机
	Port     int           // 端口
	Username string        // 用户名
	Password string        // 密码
	Database int           // Redis DB
	Timeout  time.Duration // 超时时间

	// Kafka特定
	Brokers []string // Broker列表
	Topic   string   // Topic
	GroupID string   // 消费者组ID
}

// TestConfig 测试配置
type TestConfig struct {
	Duration    time.Duration    // 测试持续时间
	Operations  int              // 操作次数
	Concurrency int              // 并发数
	Workload    []WorkloadConfig // 工作负载配置
}

// WorkloadConfig 工作负载配置
type WorkloadConfig struct {
	Operation  string // 操作类型
	Weight     int    // 权重
	KeyPattern string // 键模式
	ValueSize  int    // 值大小
}

// OutputConfig 输出配置
type OutputConfig struct {
	Format                 string // console, json, markdown
	Path                   string // 报告保存路径
	IncludeRecommendations bool   // 是否包含建议
}
