package middleware

import (
	"time"

	"middleware-chaos-testing/internal/core"
)

// RabbitMQConfig RabbitMQ配置
type RabbitMQConfig struct {
	URL              string // amqp://user:pass@host:port/vhost
	Exchange         string // 交换机名称
	ExchangeType     string // 交换机类型: direct, fanout, topic, headers
	Queue            string // 队列名称
	RoutingKey       string // 路由键

	// 连接配置
	ConnectionTimeout time.Duration // 连接超时，默认: 30s
	Heartbeat         time.Duration // 心跳间隔，默认: 10s
	ChannelMax        int           // 最大通道数，默认: 0 (无限制)
	FrameSize         int           // 帧大小，默认: 131072 (128KB)

	// QoS配置
	PrefetchCount int // 预取消息数，默认: 50
	PrefetchSize  int // 预取消息大小（字节），默认: 0

	// 消息配置
	Persistent bool // 消息持久化，默认: true
	Mandatory  bool // 强制路由（消息必须路由到队列），默认: false
	Immediate  bool // 立即投递（消息必须立即投递给消费者），默认: false

	// 死信队列配置
	DeadLetterExchange   string // 死信交换机
	DeadLetterRoutingKey string // 死信路由键

	// TTL配置
	MessageTTL int64 // 消息TTL（毫秒），默认: 0 (永不过期)
	QueueTTL   int64 // 队列TTL（毫秒），默认: 0 (永不删除)

	// 优先级配置
	MaxPriority uint8 // 队列最大优先级，默认: 0 (不启用优先级)

	// 重试配置
	MaxRetries  int           // 最大重试次数，默认: 3
	RetryDelay  time.Duration // 重试延迟，默认: 1s
	Timeout     time.Duration // 操作超时，默认: 30s
}

// ApplyDefaults 应用默认配置
func (c *RabbitMQConfig) ApplyDefaults() {
	if c.ConnectionTimeout == 0 {
		c.ConnectionTimeout = 30 * time.Second
	}
	if c.Heartbeat == 0 {
		c.Heartbeat = 10 * time.Second
	}
	if c.FrameSize == 0 {
		c.FrameSize = 131072 // 128KB
	}
	if c.PrefetchCount == 0 {
		c.PrefetchCount = 50
	}
	if c.ExchangeType == "" {
		c.ExchangeType = "direct"
	}
	if c.MaxRetries == 0 {
		c.MaxRetries = 3
	}
	if c.RetryDelay == 0 {
		c.RetryDelay = 1 * time.Second
	}
	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second
	}
}

// RabbitMQ消息结构

// RabbitMQMessage RabbitMQ消息
type RabbitMQMessage struct {
	Exchange   string                 // 交换机
	RoutingKey string                 // 路由键
	Body       []byte                 // 消息体
	Headers    map[string]interface{} // 消息头

	// 消息属性
	ContentType     string // 内容类型，如 "text/plain", "application/json"
	ContentEncoding string // 内容编码
	CorrelationID   string // 关联ID（用于RPC）
	ReplyTo         string // 回复队列
	MessageID       string // 消息ID
	Timestamp       int64  // 时间戳
	Type            string // 消息类型
	UserID          string // 用户ID
	AppID           string // 应用ID

	// 投递选项
	Persistent bool   // 是否持久化
	Priority   uint8  // 优先级（0-9）
	Expiration string // 过期时间（毫秒字符串）
}

// RabbitMQ操作类型

// RabbitMQPublishOperation 发布消息操作
type RabbitMQPublishOperation struct {
	core.BaseOperation
	Message   *RabbitMQMessage // 要发布的消息
	Mandatory bool             // 强制路由
	Immediate bool             // 立即投递
}

// RabbitMQConsumeOperation 消费消息操作
type RabbitMQConsumeOperation struct {
	core.BaseOperation
	Queue     string        // 队列名称（可选，使用配置中的队列）
	AutoAck   bool          // 自动确认
	Exclusive bool          // 独占消费
	NoLocal   bool          // 不消费自己发布的消息
	NoWait    bool          // 不等待服务器确认
	Timeout   time.Duration // 消费超时时间
}

// RabbitMQAckOperation 消息确认操作
type RabbitMQAckOperation struct {
	core.BaseOperation
	DeliveryTag uint64 // 投递标签
	Multiple    bool   // 是否批量确认
}

// RabbitMQRejectOperation 消息拒绝操作
type RabbitMQRejectOperation struct {
	core.BaseOperation
	DeliveryTag uint64 // 投递标签
	Requeue     bool   // 是否重新入队
}

// RabbitMQNackOperation 消息否定确认操作
type RabbitMQNackOperation struct {
	core.BaseOperation
	DeliveryTag uint64 // 投递标签
	Multiple    bool   // 是否批量否定确认
	Requeue     bool   // 是否重新入队
}

// 交换机类型常量
const (
	ExchangeTypeDirect  = "direct"  // 直连交换机（精确匹配路由键）
	ExchangeTypeFanout  = "fanout"  // 扇出交换机（广播到所有绑定的队列）
	ExchangeTypeTopic   = "topic"   // 主题交换机（通配符匹配路由键）
	ExchangeTypeHeaders = "headers" // 头交换机（基于消息头匹配）
)

// 投递模式常量
const (
	DeliveryModeNonPersistent = 1 // 非持久化
	DeliveryModePersistent    = 2 // 持久化
)
