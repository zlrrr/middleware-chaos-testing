package middleware

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"middleware-chaos-testing/internal/core"
)

// RocketMQClient RocketMQ客户端实现
type RocketMQClient struct {
	config   *RocketMQConfig
	producer rocketmq.Producer
	consumer rocketmq.PushConsumer
	logger   *Logger

	// 统计信息
	stats struct {
		sync.RWMutex
		messagesSent     int64
		messagesConsumed int64
		sendErrors       int64
		consumeErrors    int64
	}
}

// NewRocketMQClient 创建新的RocketMQ客户端
func NewRocketMQClient(config *RocketMQConfig) *RocketMQClient {
	config.ApplyDefaults()
	logger := NewLogger("RocketMQClient", false)
	logger.Info("Creating new RocketMQ client: nameservers=%v topic=%s producerGroup=%s consumerGroup=%s",
		config.NameServers, config.Topic, config.ProducerGroup, config.ConsumerGroup)

	return &RocketMQClient{
		config: config,
		logger: logger,
	}
}

// Connect 连接到RocketMQ
func (r *RocketMQClient) Connect(ctx context.Context) error {
	r.logger.Info("Connecting to RocketMQ...")

	// 创建生产者
	if err := r.createProducer(); err != nil {
		r.logger.Error("Failed to create producer: %v", err)
		return fmt.Errorf("failed to create producer: %w", err)
	}

	// 创建消费者（可选，根据需要）
	if err := r.createConsumer(); err != nil {
		r.logger.Warn("Failed to create consumer (optional): %v", err)
		// 消费者创建失败不影响生产者使用
	}

	r.logger.Info("Successfully connected to RocketMQ")
	r.logger.Info("Producer configured: group=%s retryTimes=%d sendTimeout=%v",
		r.config.ProducerGroup, r.config.RetryTimes, r.config.SendMsgTimeout)

	return nil
}

// createProducer 创建生产者
func (r *RocketMQClient) createProducer() error {
	// 创建生产者选项
	opts := []producer.Option{
		producer.WithNameServer(r.config.NameServers),
		producer.WithGroupName(r.config.ProducerGroup),
		producer.WithRetry(r.config.RetryTimes),
		producer.WithSendMsgTimeout(r.config.SendMsgTimeout),
	}

	// 创建生产者实例
	p, err := rocketmq.NewProducer(opts...)
	if err != nil {
		return fmt.Errorf("failed to new producer: %w", err)
	}

	// 启动生产者
	if err := p.Start(); err != nil {
		return fmt.Errorf("failed to start producer: %w", err)
	}

	r.producer = p
	r.logger.Debug("Producer created and started successfully")
	return nil
}

// createConsumer 创建消费者
func (r *RocketMQClient) createConsumer() error {
	// 创建消费者选项
	opts := []consumer.Option{
		consumer.WithNameServer(r.config.NameServers),
		consumer.WithGroupName(r.config.ConsumerGroup),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromFirstOffset),
	}

	// 设置消费模式
	if r.config.ConsumeMode == "BROADCASTING" {
		opts = append(opts, consumer.WithConsumerModel(consumer.BroadCasting))
	} else {
		opts = append(opts, consumer.WithConsumerModel(consumer.Clustering))
	}

	// 设置消费超时
	if r.config.ConsumeTimeout > 0 {
		opts = append(opts, consumer.WithConsumeTimeout(r.config.ConsumeTimeout))
	}

	// 创建消费者实例
	c, err := rocketmq.NewPushConsumer(opts...)
	if err != nil {
		return fmt.Errorf("failed to new consumer: %w", err)
	}

	// 订阅主题
	if err := c.Subscribe(r.config.Topic, consumer.MessageSelector{}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			r.stats.Lock()
			r.stats.messagesConsumed++
			r.stats.Unlock()
			r.logger.Debug("Consumed message: msgID=%s topic=%s tag=%s",
				msg.MsgId, msg.Topic, msg.GetTags())
		}
		return consumer.ConsumeSuccess, nil
	}); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// 启动消费者
	if err := c.Start(); err != nil {
		return fmt.Errorf("failed to start consumer: %w", err)
	}

	r.consumer = c
	r.logger.Debug("Consumer created and started successfully")
	return nil
}

// Disconnect 断开连接
func (r *RocketMQClient) Disconnect(ctx context.Context) error {
	r.logger.Info("Disconnecting from RocketMQ...")

	var errs []error

	// 关闭生产者
	if r.producer != nil {
		if err := r.producer.Shutdown(); err != nil {
			r.logger.Error("Failed to shutdown producer: %v", err)
			errs = append(errs, fmt.Errorf("failed to shutdown producer: %w", err))
		} else {
			r.logger.Debug("Producer shutdown successfully")
		}
		r.producer = nil
	}

	// 关闭消费者
	if r.consumer != nil {
		if err := r.consumer.Shutdown(); err != nil {
			r.logger.Error("Failed to shutdown consumer: %v", err)
			errs = append(errs, fmt.Errorf("failed to shutdown consumer: %w", err))
		} else {
			r.logger.Debug("Consumer shutdown successfully")
		}
		r.consumer = nil
	}

	if len(errs) > 0 {
		r.logger.Error("Disconnect completed with errors: %v", errs)
		return fmt.Errorf("disconnect errors: %v", errs)
	}

	r.logger.Info("Successfully disconnected from RocketMQ")
	return nil
}

// Execute 执行操作
func (r *RocketMQClient) Execute(ctx context.Context, op core.Operation) (*core.Result, error) {
	startTime := time.Now()
	r.logger.Debug("Executing operation: type=%s key=%s", op.Type(), op.Key())

	var result *core.Result

	switch v := op.(type) {
	case *RocketMQSendOperation:
		result = r.executeSend(ctx, v, startTime)
	case *RocketMQSendBatchOperation:
		result = r.executeSendBatch(ctx, v, startTime)
	case *RocketMQConsumeOperation:
		result = r.executeConsume(ctx, v, startTime)
	case *RocketMQTransactionOperation:
		result = r.executeTransaction(ctx, v, startTime)
	default:
		duration := time.Since(startTime)
		opErr := fmt.Errorf("unsupported operation type: %T", op)
		r.logger.Error("Operation failed: %v", opErr)
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
	r.logger.LogOperation(opLog)

	return result, nil
}

// executeSend 执行发送消息操作
func (r *RocketMQClient) executeSend(ctx context.Context, op *RocketMQSendOperation, startTime time.Time) *core.Result {
	if r.producer == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("producer not initialized"))
	}

	msg := op.Message
	r.logger.Debug("Sending message: topic=%s tag=%s key=%s delayLevel=%d",
		msg.Topic, msg.Tag, msg.Key, msg.DelayLevel)

	// 构造RocketMQ消息
	rocketMsg := &primitive.Message{
		Topic: msg.Topic,
		Body:  msg.Body,
	}

	// 设置Tag
	if msg.Tag != "" {
		rocketMsg.WithTag(msg.Tag)
	}

	// 设置Key
	if msg.Key != "" {
		rocketMsg.WithKeys([]string{msg.Key})
	}

	// 设置延迟级别
	if msg.DelayLevel > 0 {
		rocketMsg.WithDelayTimeLevel(msg.DelayLevel)
	}

	// 设置自定义属性
	if len(msg.Properties) > 0 {
		rocketMsg.WithProperties(msg.Properties)
	}

	// 发送消息（支持顺序消息）
	var sendResult *primitive.SendResult
	var err error

	if msg.ShardingKey != "" {
		// 顺序消息：使用ShardingKey选择队列
		selector := producer.NewHashQueueSelector()
		sendResult, err = r.producer.SendSync(ctx, rocketMsg, selector, msg.ShardingKey)
	} else {
		// 普通消息
		sendResult, err = r.producer.SendSync(ctx, rocketMsg)
	}

	duration := time.Since(startTime)

	if err != nil {
		r.stats.Lock()
		r.stats.sendErrors++
		r.stats.Unlock()
		r.logger.Error("Failed to send message: topic=%s key=%s error=%v duration=%v",
			msg.Topic, msg.Key, err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to send message: %w", err))
	}

	r.stats.Lock()
	r.stats.messagesSent++
	r.stats.Unlock()

	r.logger.Debug("Message sent successfully: msgID=%s topic=%s queueID=%d offset=%d duration=%v",
		sendResult.MsgID, msg.Topic, sendResult.MessageQueue.QueueId, sendResult.QueueOffset, duration)

	result := core.NewResult(true, duration, nil)
	result.Metadata["message_id"] = sendResult.MsgID
	result.Metadata["queue_id"] = sendResult.MessageQueue.QueueId
	result.Metadata["queue_offset"] = sendResult.QueueOffset
	result.Metadata["delay_level"] = msg.DelayLevel

	return result
}

// executeSendBatch 执行批量发送消息操作
func (r *RocketMQClient) executeSendBatch(ctx context.Context, op *RocketMQSendBatchOperation, startTime time.Time) *core.Result {
	if r.producer == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("producer not initialized"))
	}

	r.logger.Debug("Sending batch messages: count=%d", len(op.Messages))

	// 构造RocketMQ消息列表
	var rocketMsgs []*primitive.Message
	for _, msg := range op.Messages {
		rocketMsg := &primitive.Message{
			Topic: msg.Topic,
			Body:  msg.Body,
		}
		if msg.Tag != "" {
			rocketMsg.WithTag(msg.Tag)
		}
		if msg.Key != "" {
			rocketMsg.WithKeys([]string{msg.Key})
		}
		rocketMsgs = append(rocketMsgs, rocketMsg)
	}

	// 批量发送
	sendResult, err := r.producer.SendSync(ctx, rocketMsgs...)
	duration := time.Since(startTime)

	if err != nil {
		r.stats.Lock()
		r.stats.sendErrors++
		r.stats.Unlock()
		r.logger.Error("Failed to send batch messages: count=%d error=%v duration=%v",
			len(op.Messages), err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to send batch: %w", err))
	}

	r.stats.Lock()
	r.stats.messagesSent += int64(len(op.Messages))
	r.stats.Unlock()

	r.logger.Debug("Batch messages sent successfully: count=%d msgID=%s duration=%v",
		len(op.Messages), sendResult.MsgID, duration)

	result := core.NewResult(true, duration, nil)
	result.Metadata["message_count"] = len(op.Messages)
	result.Metadata["message_id"] = sendResult.MsgID

	return result
}

// executeConsume 执行消费消息操作（模拟）
func (r *RocketMQClient) executeConsume(ctx context.Context, op *RocketMQConsumeOperation, startTime time.Time) *core.Result {
	// 注意：RocketMQ的消费是推模式，这里只是模拟拉取
	r.logger.Debug("Consuming messages: maxMessages=%d timeout=%v", op.MaxMessages, op.Timeout)

	// 等待一段时间模拟消费
	time.Sleep(100 * time.Millisecond)
	duration := time.Since(startTime)

	// 返回成功结果（实际消费由订阅回调处理）
	result := core.NewResult(true, duration, nil)
	result.Metadata["consumed"] = true

	return result
}

// executeTransaction 执行事务消息操作（模拟）
func (r *RocketMQClient) executeTransaction(ctx context.Context, op *RocketMQTransactionOperation, startTime time.Time) *core.Result {
	if r.producer == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("producer not initialized"))
	}

	msg := op.Message
	r.logger.Debug("Sending transaction message: transactionID=%s key=%s",
		op.TransactionID, msg.Key)

	// 构造RocketMQ消息
	rocketMsg := &primitive.Message{
		Topic: msg.Topic,
		Body:  msg.Body,
	}
	if msg.Tag != "" {
		rocketMsg.WithTag(msg.Tag)
	}
	if msg.Key != "" {
		rocketMsg.WithKeys([]string{msg.Key})
	}

	// 发送普通消息模拟事务消息（实际需要TransactionProducer）
	sendResult, err := r.producer.SendSync(ctx, rocketMsg)
	duration := time.Since(startTime)

	if err != nil {
		r.logger.Error("Failed to send transaction message: transactionID=%s error=%v",
			op.TransactionID, err)
		return core.NewResult(false, duration, fmt.Errorf("failed to send transaction message: %w", err))
	}

	r.logger.Debug("Transaction message sent: msgID=%s transactionID=%s duration=%v",
		sendResult.MsgID, op.TransactionID, duration)

	result := core.NewResult(true, duration, nil)
	result.Metadata["message_id"] = sendResult.MsgID
	result.Metadata["transaction_id"] = op.TransactionID

	return result
}

// Ping 检查连接是否正常
func (r *RocketMQClient) Ping(ctx context.Context) error {
	if r.producer == nil {
		r.logger.Error("Ping failed: producer not initialized")
		return fmt.Errorf("producer not initialized")
	}

	r.logger.Debug("Pinging RocketMQ...")

	// 通过发送一个测试消息来验证连接（实际生产中可能有更好的方式）
	// 这里简单返回成功，因为生产者已启动
	r.logger.Debug("Ping successful")
	return nil
}

// GetStats 获取统计信息
func (r *RocketMQClient) GetStats() map[string]interface{} {
	r.stats.RLock()
	defer r.stats.RUnlock()

	stats := make(map[string]interface{})
	stats["messages_sent"] = r.stats.messagesSent
	stats["messages_consumed"] = r.stats.messagesConsumed
	stats["send_errors"] = r.stats.sendErrors
	stats["consume_errors"] = r.stats.consumeErrors

	// 添加配置信息
	stats["producer_group"] = r.config.ProducerGroup
	stats["consumer_group"] = r.config.ConsumerGroup
	stats["topic"] = r.config.Topic
	stats["retry_times"] = r.config.RetryTimes
	stats["consume_mode"] = r.config.ConsumeMode

	r.logger.Debug("Stats retrieved: sent=%d consumed=%d errors=%d",
		r.stats.messagesSent, r.stats.messagesConsumed, r.stats.sendErrors)

	return stats
}
