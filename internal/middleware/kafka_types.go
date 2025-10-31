package middleware

import (
	"time"

	"middleware-chaos-testing/internal/core"
)

// KafkaConfig Kafka配置（业界最佳实践）
type KafkaConfig struct {
	// 基础配置
	Brokers []string      // Broker地址列表
	Topic   string        // Topic名称
	GroupID string        // 消费者组ID
	Timeout time.Duration // 超时时间

	// 生产者性能配置（最佳实践）
	BatchSize    int           // 批处理大小（默认：100条）
	BatchTimeout time.Duration // 批处理超时（默认：10ms）
	Compression  int           // 压缩类型：0=none, 1=gzip, 2=snappy, 3=lz4, 4=zstd（推荐：snappy）
	MaxAttempts  int           // 最大重试次数（默认：3）
	RequiredAcks int           // ACK要求：-1=all, 0=none, 1=leader（推荐：1）
	Async        bool          // 异步发送（推荐：false以保证可靠性）

	// 消费者性能配置（最佳实践）
	MinBytes        int           // 最小读取字节（默认：1KB）
	MaxBytes        int           // 最大读取字节（默认：1MB）
	MaxWait         time.Duration // 最大等待时间（默认：100ms）
	CommitInterval  time.Duration // 提交间隔（默认：1s）
	StartOffset     int64         // 起始offset：-1=newest, -2=oldest
	HeartbeatInterval time.Duration // 心跳间隔（默认：3s）
	SessionTimeout  time.Duration // 会话超时（默认：10s）
	RebalanceTimeout time.Duration // 重平衡超时（默认：60s）

	// 连接池配置
	MaxIdleConns int           // 最大空闲连接数（默认：10）
	IdleTimeout  time.Duration // 空闲连接超时（默认：30s）
}

// ApplyDefaults 应用默认配置（业界最佳实践）
func (c *KafkaConfig) ApplyDefaults() {
	if c.Timeout == 0 {
		c.Timeout = 5 * time.Second
	}

	// 生产者最佳实践配置
	if c.BatchSize == 0 {
		c.BatchSize = 100 // 平衡延迟和吞吐量
	}
	if c.BatchTimeout == 0 {
		c.BatchTimeout = 10 * time.Millisecond // 低延迟
	}
	if c.Compression == 0 {
		c.Compression = 2 // snappy - 平衡压缩率和CPU
	}
	if c.MaxAttempts == 0 {
		c.MaxAttempts = 3
	}
	if c.RequiredAcks == 0 {
		c.RequiredAcks = 1 // leader确认，平衡性能和可靠性
	}

	// 消费者最佳实践配置
	if c.MinBytes == 0 {
		c.MinBytes = 1024 // 1KB
	}
	if c.MaxBytes == 0 {
		c.MaxBytes = 1024 * 1024 // 1MB
	}
	if c.MaxWait == 0 {
		c.MaxWait = 100 * time.Millisecond // 低延迟
	}
	if c.CommitInterval == 0 {
		c.CommitInterval = 1 * time.Second
	}
	if c.StartOffset == 0 {
		c.StartOffset = -1 // 从最新消息开始
	}
	if c.HeartbeatInterval == 0 {
		c.HeartbeatInterval = 3 * time.Second
	}
	if c.SessionTimeout == 0 {
		c.SessionTimeout = 10 * time.Second
	}
	if c.RebalanceTimeout == 0 {
		c.RebalanceTimeout = 60 * time.Second
	}

	// 连接池配置
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 10
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = 30 * time.Second
	}
}

// KafkaProduceOperation 生产消息操作
type KafkaProduceOperation struct {
	OpKey   string // 消息Key
	OpValue []byte // 消息内容
	OpTopic string // Topic（可选，如果为空使用默认Topic）
}

func (k *KafkaProduceOperation) Type() core.OperationType {
	return core.OpTypeWrite
}

func (k *KafkaProduceOperation) Key() string {
	return k.OpKey
}

func (k *KafkaProduceOperation) Value() []byte {
	return k.OpValue
}

func (k *KafkaProduceOperation) Metadata() map[string]interface{} {
	meta := make(map[string]interface{})
	if k.OpTopic != "" {
		meta["topic"] = k.OpTopic
	}
	return meta
}

// KafkaConsumeOperation 消费消息操作
type KafkaConsumeOperation struct {
	OpTopic string // Topic（可选，如果为空使用默认Topic）
	MaxWait time.Duration // 最大等待时间
}

func (k *KafkaConsumeOperation) Type() core.OperationType {
	return core.OpTypeRead
}

func (k *KafkaConsumeOperation) Key() string {
	return ""
}

func (k *KafkaConsumeOperation) Value() []byte {
	return nil
}

func (k *KafkaConsumeOperation) Metadata() map[string]interface{} {
	meta := make(map[string]interface{})
	if k.OpTopic != "" {
		meta["topic"] = k.OpTopic
	}
	if k.MaxWait > 0 {
		meta["max_wait"] = k.MaxWait
	}
	return meta
}
