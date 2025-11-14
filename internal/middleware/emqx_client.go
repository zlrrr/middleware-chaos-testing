package middleware

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"middleware-chaos-testing/internal/core"
)

// EMQXClient EMQX客户端实现（支持EMQX 5.8社区版）
type EMQXClient struct {
	config *EMQXConfig
	client mqtt.Client
	logger *Logger

	// 消息接收处理
	messageHandlers map[string]mqtt.MessageHandler
	mu              sync.RWMutex

	// 统计信息
	stats struct {
		sync.RWMutex
		messagesPublished   int64
		messagesReceived    int64
		subscriptions       int64
		unsubscriptions     int64
		publishErrors       int64
		subscribeErrors     int64
		reconnectCount      int64
		connectionLost      int64
	}
}

// NewEMQXClient 创建新的EMQX客户端
func NewEMQXClient(config *EMQXConfig) *EMQXClient {
	config.ApplyDefaults()
	logger := NewLogger("EMQXClient", false)
	logger.Info("Creating new EMQX client: broker=%s clientID=%s protocol=%d",
		config.Broker, config.ClientID, config.ProtocolVersion)

	return &EMQXClient{
		config:          config,
		logger:          logger,
		messageHandlers: make(map[string]mqtt.MessageHandler),
	}
}

// Connect 连接到EMQX
func (e *EMQXClient) Connect(ctx context.Context) error {
	e.logger.Info("Connecting to EMQX broker: %s", e.config.Broker)

	// 创建MQTT客户端选项
	opts := mqtt.NewClientOptions()
	opts.AddBroker(e.config.Broker)
	opts.SetClientID(e.config.ClientID)
	opts.SetUsername(e.config.Username)
	opts.SetPassword(e.config.Password)

	// 设置MQTT协议版本
	if e.config.ProtocolVersion == ProtocolMQTT311 {
		opts.SetProtocolVersion(4) // MQTT 3.1.1
	} else if e.config.ProtocolVersion == ProtocolMQTT5 {
		opts.SetProtocolVersion(5) // MQTT 5.0
	} else {
		opts.SetProtocolVersion(3) // MQTT 3.1
	}

	// 设置CleanSession
	opts.SetCleanSession(e.config.CleanSession)

	// 设置心跳间隔
	opts.SetKeepAlive(time.Duration(e.config.KeepAlive) * time.Second)

	// 设置连接超时
	opts.SetConnectTimeout(e.config.ConnectTimeout)
	opts.SetWriteTimeout(e.config.WriteTimeout)
	opts.SetPingTimeout(e.config.PingTimeout)

	// 设置自动重连
	if e.config.AutoReconnect {
		opts.SetAutoReconnect(true)
		opts.SetMaxReconnectInterval(e.config.MaxReconnectInterval)
	}

	// 设置消息顺序
	opts.SetOrderMatters(e.config.Order)

	// 设置遗嘱消息（Last Will）
	if e.config.WillEnabled {
		opts.SetWill(
			e.config.WillTopic,
			string(e.config.WillPayload),
			e.config.WillQoS,
			e.config.WillRetained,
		)
		e.logger.Info("Last Will configured: topic=%s qos=%d",
			e.config.WillTopic, e.config.WillQoS)
	}

	// 设置TLS
	if e.config.UseTLS {
		tlsConfig, err := e.createTLSConfig()
		if err != nil {
			e.logger.Error("Failed to create TLS config: %v", err)
			return fmt.Errorf("failed to create TLS config: %w", err)
		}
		opts.SetTLSConfig(tlsConfig)
		e.logger.Info("TLS enabled")
	}

	// 设置连接回调
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		e.logger.Info("Connected to EMQX broker successfully")
	})

	// 设置连接丢失回调
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		e.stats.Lock()
		e.stats.connectionLost++
		e.stats.Unlock()
		e.logger.Error("Connection lost: %v", err)
	})

	// 设置重连回调
	opts.SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
		e.stats.Lock()
		e.stats.reconnectCount++
		e.stats.Unlock()
		e.logger.Info("Reconnecting to EMQX broker...")
	})

	// 创建客户端
	e.client = mqtt.NewClient(opts)

	// 连接到broker
	token := e.client.Connect()
	if !token.WaitTimeout(e.config.ConnectTimeout) {
		return fmt.Errorf("connection timeout after %v", e.config.ConnectTimeout)
	}

	if err := token.Error(); err != nil {
		e.logger.Error("Failed to connect: %v", err)
		return fmt.Errorf("failed to connect to EMQX: %w", err)
	}

	e.logger.Info("Successfully connected to EMQX")
	e.logger.Info("ClientID: %s, CleanSession: %v, KeepAlive: %ds",
		e.config.ClientID, e.config.CleanSession, e.config.KeepAlive)

	return nil
}

// createTLSConfig 创建TLS配置
func (e *EMQXClient) createTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
	}

	// 加载CA证书
	if e.config.CACert != "" {
		caCert, err := os.ReadFile(e.config.CACert)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA cert")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// 加载客户端证书
	if e.config.ClientCert != "" && e.config.ClientKey != "" {
		cert, err := tls.LoadX509KeyPair(e.config.ClientCert, e.config.ClientKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}

// Disconnect 断开连接
func (e *EMQXClient) Disconnect(ctx context.Context) error {
	e.logger.Info("Disconnecting from EMQX...")

	if e.client == nil {
		e.logger.Debug("Client not initialized, nothing to disconnect")
		return nil
	}

	if !e.client.IsConnected() {
		e.logger.Debug("Client not connected, nothing to disconnect")
		return nil
	}

	// 优雅断开连接（250ms等待）
	e.client.Disconnect(250)
	e.logger.Info("Successfully disconnected from EMQX")

	return nil
}

// Execute 执行操作
func (e *EMQXClient) Execute(ctx context.Context, op core.Operation) (*core.Result, error) {
	startTime := time.Now()
	e.logger.Debug("Executing operation: type=%s key=%s", op.Type(), op.Key())

	var result *core.Result

	switch v := op.(type) {
	case *EMQXPublishOperation:
		result = e.executePublish(ctx, v, startTime)
	case *EMQXSubscribeOperation:
		result = e.executeSubscribe(ctx, v, startTime)
	case *EMQXUnsubscribeOperation:
		result = e.executeUnsubscribe(ctx, v, startTime)
	default:
		duration := time.Since(startTime)
		opErr := fmt.Errorf("unsupported operation type: %T", op)
		e.logger.Error("Operation failed: %v", opErr)
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
	e.logger.LogOperation(opLog)

	return result, nil
}

// executePublish 执行发布消息操作
func (e *EMQXClient) executePublish(ctx context.Context, op *EMQXPublishOperation, startTime time.Time) *core.Result {
	if e.client == nil || !e.client.IsConnected() {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("client not connected"))
	}

	msg := op.Message
	e.logger.Debug("Publishing message: topic=%s qos=%d retained=%v payload_size=%d",
		msg.Topic, msg.QoS, msg.Retained, len(msg.Payload))

	// 发布消息
	token := e.client.Publish(msg.Topic, msg.QoS, msg.Retained, msg.Payload)

	// 等待发布完成
	timeout := e.config.WriteTimeout
	if !token.WaitTimeout(timeout) {
		duration := time.Since(startTime)
		e.stats.Lock()
		e.stats.publishErrors++
		e.stats.Unlock()
		e.logger.Error("Publish timeout after %v: topic=%s", timeout, msg.Topic)
		return core.NewResult(false, duration, fmt.Errorf("publish timeout"))
	}

	duration := time.Since(startTime)

	if err := token.Error(); err != nil {
		e.stats.Lock()
		e.stats.publishErrors++
		e.stats.Unlock()
		e.logger.Error("Failed to publish message: topic=%s error=%v duration=%v",
			msg.Topic, err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to publish: %w", err))
	}

	e.stats.Lock()
	e.stats.messagesPublished++
	e.stats.Unlock()

	e.logger.Debug("Message published successfully: topic=%s qos=%d duration=%v",
		msg.Topic, msg.QoS, duration)

	result := core.NewResult(true, duration, nil)
	result.Metadata["published"] = true
	result.Metadata["topic"] = msg.Topic
	result.Metadata["qos"] = msg.QoS
	result.Metadata["retained"] = msg.Retained
	result.Metadata["payload_size"] = len(msg.Payload)

	return result
}

// executeSubscribe 执行订阅主题操作
func (e *EMQXClient) executeSubscribe(ctx context.Context, op *EMQXSubscribeOperation, startTime time.Time) *core.Result {
	if e.client == nil || !e.client.IsConnected() {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("client not connected"))
	}

	topic := op.Topic
	qos := op.QoS

	e.logger.Debug("Subscribing to topic: %s qos=%d", topic, qos)

	// 创建消息处理器
	handler := func(client mqtt.Client, msg mqtt.Message) {
		e.stats.Lock()
		e.stats.messagesReceived++
		e.stats.Unlock()

		e.logger.Debug("Message received: topic=%s qos=%d retained=%v payload_size=%d duplicate=%v",
			msg.Topic(), msg.Qos(), msg.Retained(), len(msg.Payload()), msg.Duplicate())
	}

	// 保存处理器
	e.mu.Lock()
	e.messageHandlers[topic] = handler
	e.mu.Unlock()

	// 订阅主题
	token := e.client.Subscribe(topic, qos, handler)

	// 等待订阅完成
	timeout := e.config.ConnectTimeout
	if op.Timeout > 0 {
		timeout = op.Timeout
	}

	if !token.WaitTimeout(timeout) {
		duration := time.Since(startTime)
		e.stats.Lock()
		e.stats.subscribeErrors++
		e.stats.Unlock()
		e.logger.Error("Subscribe timeout after %v: topic=%s", timeout, topic)
		return core.NewResult(false, duration, fmt.Errorf("subscribe timeout"))
	}

	duration := time.Since(startTime)

	if err := token.Error(); err != nil {
		e.stats.Lock()
		e.stats.subscribeErrors++
		e.stats.Unlock()
		e.logger.Error("Failed to subscribe: topic=%s error=%v duration=%v",
			topic, err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to subscribe: %w", err))
	}

	e.stats.Lock()
	e.stats.subscriptions++
	e.stats.Unlock()

	e.logger.Debug("Subscribed successfully: topic=%s qos=%d duration=%v",
		topic, qos, duration)

	result := core.NewResult(true, duration, nil)
	result.Metadata["subscribed"] = true
	result.Metadata["topic"] = topic
	result.Metadata["qos"] = qos

	return result
}

// executeUnsubscribe 执行取消订阅操作
func (e *EMQXClient) executeUnsubscribe(ctx context.Context, op *EMQXUnsubscribeOperation, startTime time.Time) *core.Result {
	if e.client == nil || !e.client.IsConnected() {
		duration := time.Since(startTime)
		return core.NewResult(false, duration, fmt.Errorf("client not connected"))
	}

	topic := op.Topic

	e.logger.Debug("Unsubscribing from topic: %s", topic)

	// 取消订阅
	token := e.client.Unsubscribe(topic)

	// 等待完成
	timeout := e.config.ConnectTimeout
	if !token.WaitTimeout(timeout) {
		duration := time.Since(startTime)
		e.logger.Error("Unsubscribe timeout after %v: topic=%s", timeout, topic)
		return core.NewResult(false, duration, fmt.Errorf("unsubscribe timeout"))
	}

	duration := time.Since(startTime)

	if err := token.Error(); err != nil {
		e.logger.Error("Failed to unsubscribe: topic=%s error=%v duration=%v",
			topic, err, duration)
		return core.NewResult(false, duration, fmt.Errorf("failed to unsubscribe: %w", err))
	}

	// 移除处理器
	e.mu.Lock()
	delete(e.messageHandlers, topic)
	e.mu.Unlock()

	e.stats.Lock()
	e.stats.unsubscriptions++
	e.stats.Unlock()

	e.logger.Debug("Unsubscribed successfully: topic=%s duration=%v", topic, duration)

	result := core.NewResult(true, duration, nil)
	result.Metadata["unsubscribed"] = true
	result.Metadata["topic"] = topic

	return result
}

// Ping 检查连接是否正常
func (e *EMQXClient) Ping(ctx context.Context) error {
	if e.client == nil {
		e.logger.Error("Ping failed: client not initialized")
		return fmt.Errorf("client not initialized")
	}

	if !e.client.IsConnected() {
		e.logger.Error("Ping failed: client not connected")
		return fmt.Errorf("client not connected")
	}

	e.logger.Debug("Ping successful")
	return nil
}

// GetStats 获取统计信息
func (e *EMQXClient) GetStats() map[string]interface{} {
	e.stats.RLock()
	defer e.stats.RUnlock()

	stats := make(map[string]interface{})
	stats["messages_published"] = e.stats.messagesPublished
	stats["messages_received"] = e.stats.messagesReceived
	stats["subscriptions"] = e.stats.subscriptions
	stats["unsubscriptions"] = e.stats.unsubscriptions
	stats["publish_errors"] = e.stats.publishErrors
	stats["subscribe_errors"] = e.stats.subscribeErrors
	stats["reconnect_count"] = e.stats.reconnectCount
	stats["connection_lost"] = e.stats.connectionLost

	// 添加配置信息
	stats["broker"] = e.config.Broker
	stats["client_id"] = e.config.ClientID
	stats["protocol_version"] = e.config.ProtocolVersion
	stats["clean_session"] = e.config.CleanSession
	stats["keep_alive"] = e.config.KeepAlive

	// 连接状态
	if e.client != nil {
		stats["connected"] = e.client.IsConnected()
	} else {
		stats["connected"] = false
	}

	// 活跃订阅数
	e.mu.RLock()
	stats["active_subscriptions"] = len(e.messageHandlers)
	e.mu.RUnlock()

	e.logger.Debug("Stats retrieved: published=%d received=%d subscriptions=%d errors=%d",
		e.stats.messagesPublished, e.stats.messagesReceived,
		e.stats.subscriptions, e.stats.publishErrors)

	return stats
}
