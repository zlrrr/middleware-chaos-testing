package middleware

import (
	"time"

	"middleware-chaos-testing/internal/core"
)

// KafkaConfig Kafka配置
type KafkaConfig struct {
	Brokers []string      // Broker地址列表
	Topic   string        // Topic名称
	GroupID string        // 消费者组ID
	Timeout time.Duration // 超时时间
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
