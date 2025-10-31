package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"middleware-chaos-testing/internal/core"
)

// KafkaClient Kafka客户端实现
type KafkaClient struct {
	config   *KafkaConfig
	writer   *kafka.Writer
	reader   *kafka.Reader
	topic    string
	groupID  string
	brokers  []string
}

// NewKafkaClient 创建新的Kafka客户端
func NewKafkaClient(config *KafkaConfig) *KafkaClient {
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}

	return &KafkaClient{
		config:  config,
		topic:   config.Topic,
		groupID: config.GroupID,
		brokers: config.Brokers,
	}
}

// Connect 连接到Kafka
func (k *KafkaClient) Connect(ctx context.Context) error {
	// 创建Writer（生产者）
	k.writer = &kafka.Writer{
		Addr:         kafka.TCP(k.brokers...),
		Topic:        k.topic,
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: k.config.Timeout,
		ReadTimeout:  k.config.Timeout,
	}

	// 创建Reader（消费者）
	k.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        k.brokers,
		Topic:          k.topic,
		GroupID:        k.groupID,
		MinBytes:       1,
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})

	// 测试连接 - 尝试获取topic元数据
	conn, err := kafka.DialLeader(ctx, "tcp", k.brokers[0], k.topic, 0)
	if err != nil {
		return fmt.Errorf("failed to dial leader: %w", err)
	}
	conn.Close()

	return nil
}

// Disconnect 断开连接
func (k *KafkaClient) Disconnect(ctx context.Context) error {
	var errs []error

	if k.writer != nil {
		if err := k.writer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close writer: %w", err))
		}
	}

	if k.reader != nil {
		if err := k.reader.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close reader: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("disconnect errors: %v", errs)
	}

	return nil
}

// Execute 执行操作
func (k *KafkaClient) Execute(ctx context.Context, op core.Operation) (*core.Result, error) {
	startTime := time.Now()

	switch op.Type() {
	case core.OpTypeWrite:
		return k.executeProduce(ctx, op, startTime)
	case core.OpTypeRead:
		return k.executeConsume(ctx, op, startTime)
	default:
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("unsupported operation type: %s", op.Type())), nil
	}
}

// executeProduce 执行生产消息操作
func (k *KafkaClient) executeProduce(ctx context.Context, op core.Operation, startTime time.Time) (*core.Result, error) {
	// 构造Kafka消息
	msg := kafka.Message{
		Key:   []byte(op.Key()),
		Value: op.Value(),
	}

	// 如果操作元数据中指定了topic，使用该topic
	if meta := op.Metadata(); meta != nil {
		if topic, ok := meta["topic"].(string); ok && topic != "" {
			msg.Topic = topic
		}
	}

	// 发送消息
	err := k.writer.WriteMessages(ctx, msg)
	duration := time.Since(startTime)

	if err != nil {
		return core.NewResult(false, duration, fmt.Errorf("failed to produce message: %w", err)), nil
	}

	return core.NewResult(true, duration, nil), nil
}

// executeConsume 执行消费消息操作
func (k *KafkaClient) executeConsume(ctx context.Context, op core.Operation, startTime time.Time) (*core.Result, error) {
	// 设置读取超时
	readCtx := ctx
	if meta := op.Metadata(); meta != nil {
		if maxWait, ok := meta["max_wait"].(time.Duration); ok && maxWait > 0 {
			var cancel context.CancelFunc
			readCtx, cancel = context.WithTimeout(ctx, maxWait)
			defer cancel()
		}
	}

	// 读取消息
	msg, err := k.reader.ReadMessage(readCtx)
	duration := time.Since(startTime)

	if err != nil {
		// 超时不算作错误，只是没有消息
		if err == context.DeadlineExceeded {
			result := core.NewResult(true, duration, nil)
			result.Metadata["no_message"] = true
			return result, nil
		}
		return core.NewResult(false, duration, fmt.Errorf("failed to consume message: %w", err)), nil
	}

	// 成功读取消息
	result := core.NewResult(true, duration, nil)
	result.Data = msg.Value
	result.Metadata["offset"] = msg.Offset
	result.Metadata["partition"] = msg.Partition
	result.Metadata["key"] = string(msg.Key)

	return result, nil
}

// Ping 检查连接是否正常
func (k *KafkaClient) Ping(ctx context.Context) error {
	if k.writer == nil {
		return fmt.Errorf("writer not initialized")
	}

	// 尝试连接到broker
	conn, err := kafka.DialLeader(ctx, "tcp", k.brokers[0], k.topic, 0)
	if err != nil {
		return fmt.Errorf("failed to ping kafka: %w", err)
	}
	defer conn.Close()

	return nil
}

// GetStats 获取统计信息
func (k *KafkaClient) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	if k.writer != nil {
		writerStats := k.writer.Stats()
		stats["writer_messages"] = writerStats.Messages
		stats["writer_bytes"] = writerStats.Bytes
		stats["writer_errors"] = writerStats.Errors
	}

	if k.reader != nil {
		readerStats := k.reader.Stats()
		stats["reader_messages"] = readerStats.Messages
		stats["reader_bytes"] = readerStats.Bytes
		stats["reader_errors"] = readerStats.Errors
		stats["reader_lag"] = readerStats.Lag
	}

	return stats
}
