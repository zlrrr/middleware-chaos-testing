package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"middleware-chaos-testing/internal/core"
)

// RabbitMQClient RabbitMQ客户端实现
type RabbitMQClient struct {
	config     *RabbitMQConfig
	connection *amqp091.Connection
	channel    *amqp091.Channel
	logger     *Logger

	// 统计信息
	stats struct {
		sync.RWMutex
		messagesPublished int64
		messagesConsumed  int64
		messagesAcked     int64
		messagesRejected  int64
		publishErrors     int64
		consumeErrors     int64
	}
}

// NewRabbitMQClient 创建新的RabbitMQ客户端
func NewRabbitMQClient(config *RabbitMQConfig) *RabbitMQClient {
	config.ApplyDefaults()
	logger := NewLogger("RabbitMQClient", false)
	logger.Info("Creating new RabbitMQ client: url=%s exchange=%s queue=%s",
		config.URL, config.Exchange, config.Queue)

	return &RabbitMQClient{
		config: config,
		logger: logger,
	}
}

// Connect 连接到RabbitMQ
func (r *RabbitMQClient) Connect(ctx context.Context) error {
	r.logger.Info("Connecting to RabbitMQ...")

	// 创建连接
	config := amqp091.Config{
		Heartbeat: r.config.Heartbeat,
		Locale:    "en_US",
	}

	conn, err := amqp091.DialConfig(r.config.URL, config)
	if err != nil {
		r.logger.Error("Failed to connect to RabbitMQ: %v", err)
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	r.connection = conn

	// 创建通道
	channel, err := conn.Channel()
	if err != nil {
		r.logger.Error("Failed to open channel: %v", err)
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	r.channel = channel

	// 设置QoS
	if err := channel.Qos(
		r.config.PrefetchCount,
		r.config.PrefetchSize,
		false,
	); err != nil {
		r.logger.Error("Failed to set QoS: %v", err)
		channel.Close()
		conn.Close()
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// 声明交换机
	if r.config.Exchange != "" {
		if err := r.declareExchange(); err != nil {
			r.logger.Error("Failed to declare exchange: %v", err)
			channel.Close()
			conn.Close()
			return fmt.Errorf("failed to declare exchange: %w", err)
		}
	}

	// 声明队列
	if r.config.Queue != "" {
		if err := r.declareQueue(); err != nil {
			r.logger.Error("Failed to declare queue: %v", err)
			channel.Close()
			conn.Close()
			return fmt.Errorf("failed to declare queue: %w", err)
		}

		// 绑定队列到交换机
		if r.config.Exchange != "" && r.config.RoutingKey != "" {
			if err := r.bindQueue(); err != nil {
				r.logger.Error("Failed to bind queue: %v", err)
				channel.Close()
				conn.Close()
				return fmt.Errorf("failed to bind queue: %w", err)
			}
		}
	}

	r.logger.Info("Successfully connected to RabbitMQ")
	r.logger.Info("Exchange: %s (type=%s), Queue: %s, RoutingKey: %s",
		r.config.Exchange, r.config.ExchangeType, r.config.Queue, r.config.RoutingKey)

	return nil
}

// declareExchange 声明交换机
func (r *RabbitMQClient) declareExchange() error {
	return r.channel.ExchangeDeclare(
		r.config.Exchange,     // name
		r.config.ExchangeType, // type
		true,                  // durable
		false,                 // auto-deleted
		false,                 // internal
		false,                 // no-wait
		nil,                   // arguments
	)
}

// declareQueue 声明队列
func (r *RabbitMQClient) declareQueue() error {
	args := amqp091.Table{}

	// 配置死信队列
	if r.config.DeadLetterExchange != "" {
		args["x-dead-letter-exchange"] = r.config.DeadLetterExchange
		if r.config.DeadLetterRoutingKey != "" {
			args["x-dead-letter-routing-key"] = r.config.DeadLetterRoutingKey
		}
	}

	// 配置TTL
	if r.config.MessageTTL > 0 {
		args["x-message-ttl"] = r.config.MessageTTL
	}
	if r.config.QueueTTL > 0 {
		args["x-expires"] = r.config.QueueTTL
	}

	// 配置优先级
	if r.config.MaxPriority > 0 {
		args["x-max-priority"] = r.config.MaxPriority
	}

	_, err := r.channel.QueueDeclare(
		r.config.Queue, // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		args,           // arguments
	)

	return err
}

// bindQueue 绑定队列到交换机
func (r *RabbitMQClient) bindQueue() error {
	return r.channel.QueueBind(
		r.config.Queue,      // queue name
		r.config.RoutingKey, // routing key
		r.config.Exchange,   // exchange
		false,               // no-wait
		nil,                 // arguments
	)
}

// Disconnect 断开连接
func (r *RabbitMQClient) Disconnect(ctx context.Context) error {
	r.logger.Info("Disconnecting from RabbitMQ...")

	var errs []error

	// 关闭通道
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			r.logger.Error("Failed to close channel: %v", err)
			errs = append(errs, fmt.Errorf("failed to close channel: %w", err))
		} else {
			r.logger.Debug("Channel closed successfully")
		}
		r.channel = nil
	}

	// 关闭连接
	if r.connection != nil {
		if err := r.connection.Close(); err != nil {
			r.logger.Error("Failed to close connection: %v", err)
			errs = append(errs, fmt.Errorf("failed to close connection: %w", err))
		} else {
			r.logger.Debug("Connection closed successfully")
		}
		r.connection = nil
	}

	if len(errs) > 0 {
		r.logger.Error("Disconnect completed with errors: %v", errs)
		return fmt.Errorf("disconnect errors: %v", errs)
	}

	r.logger.Info("Successfully disconnected from RabbitMQ")
	return nil
}

// Execute 执行操作
func (r *RabbitMQClient) Execute(ctx context.Context, op core.Operation) (*core.Result, error) {
	startTime := time.Now()
	r.logger.Debug("Executing operation: type=%s key=%s", op.Type(), op.Key())

	var result *core.Result

	switch v := op.(type) {
	case *RabbitMQPublishOperation:
		result = r.executePublish(ctx, v, startTime)
	case *RabbitMQConsumeOperation:
		result = r.executeConsume(ctx, v, startTime)
	case *RabbitMQAckOperation:
		result = r.executeAck(ctx, v, startTime)
	case *RabbitMQRejectOperation:
		result = r.executeReject(ctx, v, startTime)
	case *RabbitMQNackOperation:
		result = r.executeNack(ctx, v, startTime)
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

// executePublish 执行发布消息操作
func (r *RabbitMQClient) executePublish(ctx context.Context, op *RabbitMQPublishOperation, startTime time.Time) *core.Result {
	if r.channel == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("channel not initialized"))
	}

	msg := op.Message
	r.logger.Debug("Publishing message: exchange=%s routingKey=%s persistent=%v",
		msg.Exchange, msg.RoutingKey, msg.Persistent)

	// 构造AMQP消息
	publishing := amqp091.Publishing{
		ContentType: msg.ContentType,
		Body:        msg.Body,
		Headers:     amqp091.Table(msg.Headers),
	}

	// 设置持久化
	if msg.Persistent || r.config.Persistent {
		publishing.DeliveryMode = DeliveryModePersistent
	} else {
		publishing.DeliveryMode = DeliveryModeNonPersistent
	}

	// 设置其他属性
	if msg.ContentEncoding != "" {
		publishing.ContentEncoding = msg.ContentEncoding
	}
	if msg.CorrelationID != "" {
		publishing.CorrelationId = msg.CorrelationID
	}
	if msg.ReplyTo != "" {
		publishing.ReplyTo = msg.ReplyTo
	}
	if msg.MessageID != "" {
		publishing.MessageId = msg.MessageID
	}
	if msg.Type != "" {
		publishing.Type = msg.Type
	}
	if msg.UserID != "" {
		publishing.UserId = msg.UserID
	}
	if msg.AppID != "" {
		publishing.AppId = msg.AppID
	}
	if msg.Priority > 0 {
		publishing.Priority = msg.Priority
	}
	if msg.Expiration != "" {
		publishing.Expiration = msg.Expiration
	}
	if msg.Timestamp > 0 {
		publishing.Timestamp = time.Unix(msg.Timestamp, 0)
	}

	// 发布消息
	exchange := msg.Exchange
	if exchange == "" {
		exchange = r.config.Exchange
	}

	routingKey := msg.RoutingKey
	if routingKey == "" {
		routingKey = r.config.RoutingKey
	}

	mandatory := op.Mandatory || r.config.Mandatory
	immediate := op.Immediate || r.config.Immediate

	err := r.channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		mandatory,
		immediate,
		publishing,
	)

	duration := time.Since(startTime)

	if err != nil {
		r.stats.Lock()
		r.stats.publishErrors++
		r.stats.Unlock()
		r.logger.Error("Failed to publish message: exchange=%s routingKey=%s error=%v duration=%v",
			exchange, routingKey, err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to publish message: %w", err))
	}

	r.stats.Lock()
	r.stats.messagesPublished++
	r.stats.Unlock()

	r.logger.Debug("Message published successfully: exchange=%s routingKey=%s duration=%v",
		exchange, routingKey, duration)

	result := core.NewResult(true, duration, nil)
	result.Metadata["published"] = true
	result.Metadata["exchange"] = exchange
	result.Metadata["routing_key"] = routingKey

	return result
}

// executeConsume 执行消费消息操作
func (r *RabbitMQClient) executeConsume(ctx context.Context, op *RabbitMQConsumeOperation, startTime time.Time) *core.Result {
	if r.channel == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("channel not initialized"))
	}

	queue := op.Queue
	if queue == "" {
		queue = r.config.Queue
	}

	r.logger.Debug("Consuming message: queue=%s autoAck=%v timeout=%v",
		queue, op.AutoAck, op.Timeout)

	// 尝试获取一条消息
	delivery, ok, err := r.channel.Get(queue, op.AutoAck)
	duration := time.Since(startTime)

	if err != nil {
		r.stats.Lock()
		r.stats.consumeErrors++
		r.stats.Unlock()
		r.logger.Error("Failed to consume message: queue=%s error=%v duration=%v",
			queue, err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to consume message: %w", err))
	}

	if !ok {
		// 队列为空
		r.logger.Debug("No message available: queue=%s duration=%v", queue, duration)
		result := core.NewResult(true, duration, nil)
		result.Metadata["consumed"] = false
		result.Metadata["queue"] = queue
		return result
	}

	r.stats.Lock()
	r.stats.messagesConsumed++
	r.stats.Unlock()

	r.logger.Debug("Message consumed successfully: queue=%s deliveryTag=%d duration=%v",
		queue, delivery.DeliveryTag, duration)

	// 构造消费结果
	message := &RabbitMQMessage{
		Exchange:   delivery.Exchange,
		RoutingKey: delivery.RoutingKey,
		Body:       delivery.Body,
		Headers:    map[string]interface{}(delivery.Headers),
	}

	result := core.NewResult(true, duration, nil)
	// 将message序列化为JSON
	if data, err := json.Marshal(message); err == nil {
		result.Data = data
	}
	result.Metadata["consumed"] = true
	result.Metadata["queue"] = queue
	result.Metadata["delivery_tag"] = delivery.DeliveryTag
	result.Metadata["redelivered"] = delivery.Redelivered
	result.Metadata["exchange"] = delivery.Exchange
	result.Metadata["routing_key"] = delivery.RoutingKey

	return result
}

// executeAck 执行消息确认操作
func (r *RabbitMQClient) executeAck(ctx context.Context, op *RabbitMQAckOperation, startTime time.Time) *core.Result {
	if r.channel == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("channel not initialized"))
	}

	r.logger.Debug("Acknowledging message: deliveryTag=%d multiple=%v",
		op.DeliveryTag, op.Multiple)

	err := r.channel.Ack(op.DeliveryTag, op.Multiple)
	duration := time.Since(startTime)

	if err != nil {
		r.logger.Error("Failed to acknowledge message: deliveryTag=%d error=%v duration=%v",
			op.DeliveryTag, err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to acknowledge message: %w", err))
	}

	r.stats.Lock()
	r.stats.messagesAcked++
	r.stats.Unlock()

	r.logger.Debug("Message acknowledged successfully: deliveryTag=%d duration=%v",
		op.DeliveryTag, duration)

	result := core.NewResult(true, duration, nil)
	result.Metadata["acknowledged"] = true
	result.Metadata["delivery_tag"] = op.DeliveryTag

	return result
}

// executeReject 执行消息拒绝操作
func (r *RabbitMQClient) executeReject(ctx context.Context, op *RabbitMQRejectOperation, startTime time.Time) *core.Result {
	if r.channel == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("channel not initialized"))
	}

	r.logger.Debug("Rejecting message: deliveryTag=%d requeue=%v",
		op.DeliveryTag, op.Requeue)

	err := r.channel.Reject(op.DeliveryTag, op.Requeue)
	duration := time.Since(startTime)

	if err != nil {
		r.logger.Error("Failed to reject message: deliveryTag=%d error=%v duration=%v",
			op.DeliveryTag, err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to reject message: %w", err))
	}

	r.stats.Lock()
	r.stats.messagesRejected++
	r.stats.Unlock()

	r.logger.Debug("Message rejected successfully: deliveryTag=%d requeue=%v duration=%v",
		op.DeliveryTag, op.Requeue, duration)

	result := core.NewResult(true, duration, nil)
	result.Metadata["rejected"] = true
	result.Metadata["delivery_tag"] = op.DeliveryTag
	result.Metadata["requeued"] = op.Requeue

	return result
}

// executeNack 执行消息否定确认操作
func (r *RabbitMQClient) executeNack(ctx context.Context, op *RabbitMQNackOperation, startTime time.Time) *core.Result {
	if r.channel == nil {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("channel not initialized"))
	}

	r.logger.Debug("Nacking message: deliveryTag=%d multiple=%v requeue=%v",
		op.DeliveryTag, op.Multiple, op.Requeue)

	err := r.channel.Nack(op.DeliveryTag, op.Multiple, op.Requeue)
	duration := time.Since(startTime)

	if err != nil {
		r.logger.Error("Failed to nack message: deliveryTag=%d error=%v duration=%v",
			op.DeliveryTag, err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to nack message: %w", err))
	}

	r.logger.Debug("Message nacked successfully: deliveryTag=%d duration=%v",
		op.DeliveryTag, duration)

	result := core.NewResult(true, duration, nil)
	result.Metadata["nacked"] = true
	result.Metadata["delivery_tag"] = op.DeliveryTag
	result.Metadata["requeued"] = op.Requeue

	return result
}

// Ping 检查连接是否正常
func (r *RabbitMQClient) Ping(ctx context.Context) error {
	if r.connection == nil || r.connection.IsClosed() {
		r.logger.Error("Ping failed: connection not established or closed")
		return fmt.Errorf("connection not established or closed")
	}

	if r.channel == nil {
		r.logger.Error("Ping failed: channel not initialized")
		return fmt.Errorf("channel not initialized")
	}

	r.logger.Debug("Ping successful")
	return nil
}

// GetStats 获取统计信息
func (r *RabbitMQClient) GetStats() map[string]interface{} {
	r.stats.RLock()
	defer r.stats.RUnlock()

	stats := make(map[string]interface{})
	stats["messages_published"] = r.stats.messagesPublished
	stats["messages_consumed"] = r.stats.messagesConsumed
	stats["messages_acked"] = r.stats.messagesAcked
	stats["messages_rejected"] = r.stats.messagesRejected
	stats["publish_errors"] = r.stats.publishErrors
	stats["consume_errors"] = r.stats.consumeErrors

	// 添加配置信息
	stats["exchange"] = r.config.Exchange
	stats["exchange_type"] = r.config.ExchangeType
	stats["queue"] = r.config.Queue
	stats["routing_key"] = r.config.RoutingKey
	stats["prefetch_count"] = r.config.PrefetchCount

	// 连接状态
	if r.connection != nil {
		stats["connected"] = !r.connection.IsClosed()
	} else {
		stats["connected"] = false
	}

	r.logger.Debug("Stats retrieved: published=%d consumed=%d acked=%d rejected=%d errors=%d",
		r.stats.messagesPublished, r.stats.messagesConsumed, r.stats.messagesAcked,
		r.stats.messagesRejected, r.stats.publishErrors)

	return stats
}
