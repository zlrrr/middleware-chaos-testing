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
	logger   *Logger
}

// NewKafkaClient 创建新的Kafka客户端
func NewKafkaClient(config *KafkaConfig) *KafkaClient {
	// 应用默认配置（业界最佳实践）
	config.ApplyDefaults()

	logger := NewLogger("KafkaClient", false)
	logger.Info("Creating new Kafka client: brokers=%v topic=%s groupID=%s",
		config.Brokers, config.Topic, config.GroupID)

	return &KafkaClient{
		config:  config,
		topic:   config.Topic,
		groupID: config.GroupID,
		brokers: config.Brokers,
		logger:  logger,
	}
}

// Connect 连接到Kafka
func (k *KafkaClient) Connect(ctx context.Context) error {
	k.logger.Info("Connecting to Kafka: brokers=%v topic=%s", k.brokers, k.topic)

	// 创建Writer（生产者）- 使用业界最佳实践配置
	compressionCodec := k.getCompressionCodec()
	k.writer = &kafka.Writer{
		Addr:         kafka.TCP(k.brokers...),
		Topic:        k.topic,
		Balancer:     &kafka.LeastBytes{}, // 负载均衡
		// 性能配置（最佳实践）
		BatchSize:    k.config.BatchSize,              // 批处理大小
		BatchTimeout: k.config.BatchTimeout,           // 批处理超时（低延迟）
		Compression:  compressionCodec,                // 压缩算法
		MaxAttempts:  k.config.MaxAttempts,            // 最大重试次数
		RequiredAcks: kafka.RequiredAcks(k.config.RequiredAcks), // ACK要求
		Async:        k.config.Async,                  // 同步/异步发送
		// 超时配置
		WriteTimeout: k.config.Timeout,
		ReadTimeout:  k.config.Timeout,
	}

	k.logger.Info("Writer configured: batchSize=%d batchTimeout=%v compression=%d acks=%d async=%v",
		k.config.BatchSize, k.config.BatchTimeout, k.config.Compression,
		k.config.RequiredAcks, k.config.Async)

	// 创建Reader（消费者）- 使用业界最佳实践配置
	k.reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  k.brokers,
		Topic:    k.topic,
		GroupID:  k.groupID,
		// 性能配置（最佳实践）
		MinBytes:          k.config.MinBytes,          // 最小读取字节
		MaxBytes:          k.config.MaxBytes,          // 最大读取字节
		MaxWait:           k.config.MaxWait,           // 最大等待时间（低延迟）
		CommitInterval:    k.config.CommitInterval,    // 提交间隔
		StartOffset:       k.config.StartOffset,       // 起始offset
		HeartbeatInterval: k.config.HeartbeatInterval, // 心跳间隔
		SessionTimeout:    k.config.SessionTimeout,    // 会话超时
		RebalanceTimeout:  k.config.RebalanceTimeout,  // 重平衡超时
	})

	k.logger.Info("Reader configured: minBytes=%d maxBytes=%d maxWait=%v commitInterval=%v",
		k.config.MinBytes, k.config.MaxBytes, k.config.MaxWait, k.config.CommitInterval)

	// 测试连接 - 尝试获取topic元数据
	k.logger.Debug("Testing connection to Kafka broker...")
	conn, err := kafka.DialLeader(ctx, "tcp", k.brokers[0], k.topic, 0)
	if err != nil {
		k.logger.Error("Failed to connect to Kafka: %v", err)
		return fmt.Errorf("failed to dial leader: %w", err)
	}
	conn.Close()

	k.logger.Info("Successfully connected to Kafka")
	return nil
}

// getCompressionCodec 获取压缩编码器
func (k *KafkaClient) getCompressionCodec() kafka.Compression {
	switch k.config.Compression {
	case 0:
		return kafka.Compression(0) // None
	case 1:
		return kafka.Gzip
	case 2:
		return kafka.Snappy
	case 3:
		return kafka.Lz4
	case 4:
		return kafka.Zstd
	default:
		return kafka.Snappy // 默认使用snappy
	}
}

// Disconnect 断开连接
func (k *KafkaClient) Disconnect(ctx context.Context) error {
	k.logger.Info("Disconnecting from Kafka...")
	var errs []error

	if k.writer != nil {
		if err := k.writer.Close(); err != nil {
			k.logger.Error("Failed to close writer: %v", err)
			errs = append(errs, fmt.Errorf("failed to close writer: %w", err))
		} else {
			k.logger.Debug("Writer closed successfully")
		}
	}

	if k.reader != nil {
		if err := k.reader.Close(); err != nil {
			k.logger.Error("Failed to close reader: %v", err)
			errs = append(errs, fmt.Errorf("failed to close reader: %w", err))
		} else {
			k.logger.Debug("Reader closed successfully")
		}
	}

	if len(errs) > 0 {
		k.logger.Error("Disconnect completed with errors: %v", errs)
		return fmt.Errorf("disconnect errors: %v", errs)
	}

	k.logger.Info("Successfully disconnected from Kafka")
	return nil
}

// Execute 执行操作
func (k *KafkaClient) Execute(ctx context.Context, op core.Operation) (*core.Result, error) {
	startTime := time.Now()
	k.logger.Debug("Executing operation: type=%s key=%s", op.Type(), op.Key())

	var result *core.Result
	var err error

	switch op.Type() {
	case core.OpTypeWrite:
		result, err = k.executeProduce(ctx, op, startTime)
	case core.OpTypeRead:
		result, err = k.executeConsume(ctx, op, startTime)
	default:
		duration := time.Since(startTime)
		opErr := fmt.Errorf("unsupported operation type: %s", op.Type())
		k.logger.Error("Operation failed: %v", opErr)
		return core.NewResult(false, duration, opErr), nil
	}

	// 记录操作日志
	opLog := &OperationLog{
		Timestamp: startTime,
		Operation: string(op.Type()),
		Key:       op.Key(),
		Success:   result.Success,
		Duration:  result.Duration,
		Metadata:  result.Metadata,
	}
	if result.Error != nil {
		opLog.Error = result.Error.Error()
	}
	k.logger.LogOperation(opLog)

	return result, err
}

// executeProduce 执行生产消息操作
func (k *KafkaClient) executeProduce(ctx context.Context, op core.Operation, startTime time.Time) (*core.Result, error) {
	// 构造Kafka消息
	msg := kafka.Message{
		Key:   []byte(op.Key()),
		Value: op.Value(),
	}

	// 如果操作元数据中指定了topic，使用该topic
	topic := k.topic
	if meta := op.Metadata(); meta != nil {
		if t, ok := meta["topic"].(string); ok && t != "" {
			msg.Topic = t
			topic = t
		}
	}

	k.logger.Debug("Producing message: key=%s topic=%s valueSize=%d",
		op.Key(), topic, len(op.Value()))

	// 发送消息
	err := k.writer.WriteMessages(ctx, msg)
	duration := time.Since(startTime)

	if err != nil {
		k.logger.Error("Failed to produce message: key=%s topic=%s error=%v duration=%v",
			op.Key(), topic, err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to produce message: %w", err)), nil
	}

	k.logger.Debug("Message produced successfully: key=%s topic=%s duration=%v",
		op.Key(), topic, duration)
	return core.NewResult(true, duration, nil), nil
}

// executeConsume 执行消费消息操作
func (k *KafkaClient) executeConsume(ctx context.Context, op core.Operation, startTime time.Time) (*core.Result, error) {
	// 设置读取超时
	readCtx := ctx
	maxWait := k.config.MaxWait
	if meta := op.Metadata(); meta != nil {
		if mw, ok := meta["max_wait"].(time.Duration); ok && mw > 0 {
			maxWait = mw
			var cancel context.CancelFunc
			readCtx, cancel = context.WithTimeout(ctx, maxWait)
			defer cancel()
		}
	}

	k.logger.Debug("Consuming message: topic=%s groupID=%s maxWait=%v",
		k.topic, k.groupID, maxWait)

	// 读取消息
	msg, err := k.reader.ReadMessage(readCtx)
	duration := time.Since(startTime)

	if err != nil {
		// 超时不算作错误，只是没有消息
		if err == context.DeadlineExceeded {
			k.logger.Debug("No message available (timeout): duration=%v", duration)
			result := core.NewResult(true, duration, nil)
			result.Metadata["no_message"] = true
			return result, nil
		}
		k.logger.Error("Failed to consume message: error=%v duration=%v", err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to consume message: %w", err)), nil
	}

	// 成功读取消息
	k.logger.Debug("Message consumed successfully: key=%s offset=%d partition=%d size=%d duration=%v",
		string(msg.Key), msg.Offset, msg.Partition, len(msg.Value), duration)

	result := core.NewResult(true, duration, nil)
	result.Data = msg.Value
	result.Metadata["offset"] = msg.Offset
	result.Metadata["partition"] = msg.Partition
	result.Metadata["key"] = string(msg.Key)
	result.Metadata["topic"] = msg.Topic

	return result, nil
}

// Ping 检查连接是否正常
func (k *KafkaClient) Ping(ctx context.Context) error {
	if k.writer == nil {
		k.logger.Error("Ping failed: writer not initialized")
		return fmt.Errorf("writer not initialized")
	}

	k.logger.Debug("Pinging Kafka broker: %s", k.brokers[0])

	// 尝试连接到broker
	conn, err := kafka.DialLeader(ctx, "tcp", k.brokers[0], k.topic, 0)
	if err != nil {
		k.logger.Error("Ping failed: %v", err)
		return fmt.Errorf("failed to ping kafka: %w", err)
	}
	defer conn.Close()

	k.logger.Debug("Ping successful")
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
		k.logger.Debug("Writer stats: messages=%d bytes=%d errors=%d",
			writerStats.Messages, writerStats.Bytes, writerStats.Errors)
	}

	if k.reader != nil {
		readerStats := k.reader.Stats()
		stats["reader_messages"] = readerStats.Messages
		stats["reader_bytes"] = readerStats.Bytes
		stats["reader_errors"] = readerStats.Errors
		stats["reader_lag"] = readerStats.Lag
		k.logger.Debug("Reader stats: messages=%d bytes=%d errors=%d lag=%d",
			readerStats.Messages, readerStats.Bytes, readerStats.Errors, readerStats.Lag)

		// 如果有消息积压，发出警告
		if readerStats.Lag > 1000 {
			k.logger.Warn("High message lag detected: %d messages", readerStats.Lag)
		}
	}

	return stats
}
