package middleware

import (
	"time"

	"middleware-chaos-testing/internal/core"
)

// EMQXConfig EMQX配置（EMQX 5.8社区版）
type EMQXConfig struct {
	Broker   string // MQTT Broker地址: tcp://host:port 或 ssl://host:port
	ClientID string // 客户端ID
	Username string // 用户名
	Password string // 密码

	// MQTT协议配置
	ProtocolVersion byte // MQTT协议版本: 3 (MQTT 3.1), 4 (MQTT 3.1.1), 5 (MQTT 5.0)
	CleanSession    bool // 清理会话，默认: true
	KeepAlive       int  // 心跳间隔（秒），默认: 60

	// 遗嘱消息（Last Will）配置
	WillEnabled  bool   // 是否启用遗嘱消息
	WillTopic    string // 遗嘱消息主题
	WillPayload  []byte // 遗嘱消息内容
	WillQoS      byte   // 遗嘱消息QoS级别
	WillRetained bool   // 遗嘱消息是否保留

	// TLS/SSL配置
	UseTLS     bool   // 是否使用TLS
	CACert     string // CA证书路径
	ClientCert string // 客户端证书路径
	ClientKey  string // 客户端私钥路径

	// 性能配置
	MaxReconnectInterval time.Duration // 最大重连间隔，默认: 10分钟
	ConnectTimeout       time.Duration // 连接超时，默认: 30s
	WriteTimeout         time.Duration // 写入超时，默认: 30s
	PingTimeout          time.Duration // Ping超时，默认: 10s
	Order                bool          // 是否保证消息顺序，默认: true
	ResumeSubs           bool          // 是否恢复订阅，默认: false

	// 重试配置
	AutoReconnect    bool          // 是否自动重连，默认: true
	MaxReconnectWait time.Duration // 最大重连等待时间
}

// ApplyDefaults 应用默认配置
func (c *EMQXConfig) ApplyDefaults() {
	if c.ProtocolVersion == 0 {
		c.ProtocolVersion = 4 // 默认使用MQTT 3.1.1
	}
	if c.KeepAlive == 0 {
		c.KeepAlive = 60 // 60秒
	}
	if c.MaxReconnectInterval == 0 {
		c.MaxReconnectInterval = 10 * time.Minute
	}
	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = 30 * time.Second
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = 30 * time.Second
	}
	if c.PingTimeout == 0 {
		c.PingTimeout = 10 * time.Second
	}
	// CleanSession默认为false（Go zero value），需要显式设置
	// Order默认为false，AutoReconnect默认为false
}

// EMQX消息结构

// EMQXMessage EMQX MQTT消息
type EMQXMessage struct {
	Topic    string // 主题
	Payload  []byte // 消息负载
	QoS      byte   // 服务质量: 0, 1, 2
	Retained bool   // 是否保留消息
	MessageID uint16 // 消息ID（由客户端库自动分配）
}

// EMQX操作类型

// EMQXPublishOperation 发布消息操作
type EMQXPublishOperation struct {
	core.BaseOperation
	Message *EMQXMessage // 要发布的消息
}

// EMQXSubscribeOperation 订阅主题操作
type EMQXSubscribeOperation struct {
	core.BaseOperation
	Topic   string        // 订阅主题（支持通配符: + 和 #）
	QoS     byte          // 订阅QoS级别
	Timeout time.Duration // 订阅超时时间
}

// EMQXUnsubscribeOperation 取消订阅操作
type EMQXUnsubscribeOperation struct {
	core.BaseOperation
	Topic string // 取消订阅的主题
}

// MQTT QoS级别常量
const (
	QoS0 byte = 0 // At most once（至多一次，fire-and-forget）
	QoS1 byte = 1 // At least once（至少一次，需要确认）
	QoS2 byte = 2 // Exactly once（精确一次，四次握手）
)

// MQTT协议版本常量
const (
	ProtocolMQTT31  byte = 3 // MQTT 3.1
	ProtocolMQTT311 byte = 4 // MQTT 3.1.1
	ProtocolMQTT5   byte = 5 // MQTT 5.0
)

// MQTT主题通配符说明
// + : 单级通配符，例如 "sensor/+/temperature" 可匹配 "sensor/1/temperature"
// # : 多级通配符，例如 "sensor/#" 可匹配 "sensor/1/temperature" 和 "sensor/2/humidity"
// $share/{group}/{topic} : 共享订阅（EMQX特性），用于负载均衡
