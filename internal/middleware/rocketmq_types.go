package middleware

import (
	"time"

	"middleware-chaos-testing/internal/core"
)

// RocketMQConfig RocketMQ配置
type RocketMQConfig struct {
	NameServers []string // NameServer地址列表
	Topic       string   // 主题
	ProducerGroup string // 生产者组
	ConsumerGroup string // 消费者组
	Protocol      string // 连接协议: "remoting" 或 "grpc"，默认: "remoting" (RocketMQ 5.1.3)

	// 生产者配置
	SendMsgTimeout time.Duration // 发送消息超时，默认: 3s
	RetryTimes     int           // 重试次数，默认: 3
	CompressLevel  int           // 压缩级别，0-9
	MaxMessageSize int           // 最大消息大小，默认: 4MB

	// 消费者配置
	ConsumeMode    string        // 消费模式: CLUSTERING, BROADCASTING
	MessageModel   string        // 消息模型: ORDERED, CONCURRENT
	PullBatchSize  int           // 拉取批量大小，默认: 32
	ConsumeTimeout time.Duration // 消费超时，默认: 15分钟

	// 性能配置
	MaxReconsumeTimes int // 最大重新消费次数，默认: -1（无限）
}

// RocketMQ 5.1.3 协议常量
const (
	ProtocolRemoting = "remoting" // 传统remoting协议（默认，兼容性好）
	ProtocolGRPC     = "grpc"     // gRPC协议（RocketMQ 5.x推荐，性能更优）
)

// ApplyDefaults 应用默认配置
func (c *RocketMQConfig) ApplyDefaults() {
	if c.Protocol == "" {
		c.Protocol = ProtocolRemoting // 默认使用remoting协议（兼容性好）
	}
	if c.SendMsgTimeout == 0 {
		c.SendMsgTimeout = 3 * time.Second
	}
	if c.RetryTimes == 0 {
		c.RetryTimes = 3
	}
	if c.MaxMessageSize == 0 {
		c.MaxMessageSize = 4 * 1024 * 1024 // 4MB
	}
	if c.ConsumeMode == "" {
		c.ConsumeMode = "CLUSTERING"
	}
	if c.MessageModel == "" {
		c.MessageModel = "CONCURRENT"
	}
	if c.PullBatchSize == 0 {
		c.PullBatchSize = 32
	}
	if c.ConsumeTimeout == 0 {
		c.ConsumeTimeout = 15 * time.Minute
	}
	if c.MaxReconsumeTimes == 0 {
		c.MaxReconsumeTimes = -1 // 无限重试
	}
}

// RocketMQ消息结构

// RocketMQMessage RocketMQ消息
type RocketMQMessage struct {
	Topic       string            // 主题
	Tag         string            // 标签（用于消息过滤）
	Key         string            // 消息键（用于消息查询）
	Body        []byte            // 消息体
	Properties  map[string]string // 自定义属性
	DelayLevel  int               // 延迟级别（1-18，分别对应不同延迟时间）
	ShardingKey string            // 分片键（用于顺序消息）
}

// RocketMQ操作类型

// RocketMQSendOperation 发送消息操作
type RocketMQSendOperation struct {
	core.BaseOperation
	Message *RocketMQMessage // 要发送的消息
}

// RocketMQSendBatchOperation 批量发送消息操作
type RocketMQSendBatchOperation struct {
	core.BaseOperation
	Messages []*RocketMQMessage // 要批量发送的消息列表
}

// RocketMQConsumeOperation 消费消息操作
type RocketMQConsumeOperation struct {
	core.BaseOperation
	MaxMessages int           // 最大消费消息数
	Timeout     time.Duration // 消费超时时间
}

// RocketMQTransactionOperation 事务消息操作
type RocketMQTransactionOperation struct {
	core.BaseOperation
	Message       *RocketMQMessage // 事务消息
	TransactionID string           // 事务ID
	CheckTimes    int              // 事务回查次数
}

// RocketMQ延迟级别定义
// Level 1: 1s, Level 2: 5s, Level 3: 10s, Level 4: 30s
// Level 5: 1m, Level 6: 2m, Level 7: 3m, Level 8: 4m, Level 9: 5m
// Level 10: 6m, Level 11: 7m, Level 12: 8m, Level 13: 9m, Level 14: 10m
// Level 15: 20m, Level 16: 30m, Level 17: 1h, Level 18: 2h
const (
	DelayLevel1s   = 1  // 1秒
	DelayLevel5s   = 2  // 5秒
	DelayLevel10s  = 3  // 10秒
	DelayLevel30s  = 4  // 30秒
	DelayLevel1m   = 5  // 1分钟
	DelayLevel2m   = 6  // 2分钟
	DelayLevel5m   = 9  // 5分钟
	DelayLevel10m  = 14 // 10分钟
	DelayLevel30m  = 16 // 30分钟
	DelayLevel1h   = 17 // 1小时
	DelayLevel2h   = 18 // 2小时
)
