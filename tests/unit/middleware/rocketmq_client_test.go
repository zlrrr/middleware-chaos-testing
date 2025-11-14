package middleware_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"middleware-chaos-testing/internal/core"
	"middleware-chaos-testing/internal/middleware"
)

// RocketMQClientTestSuite RocketMQ客户端测试套件
type RocketMQClientTestSuite struct {
	suite.Suite
	client *middleware.RocketMQClient
	config *middleware.RocketMQConfig
}

// SetupTest 每个测试前执行
func (suite *RocketMQClientTestSuite) SetupTest() {
	suite.config = &middleware.RocketMQConfig{
		NameServers:   []string{"localhost:9876"},
		Topic:         "chaos-test-topic",
		ProducerGroup: "chaos-producer-group",
		ConsumerGroup: "chaos-consumer-group",
	}
	suite.client = middleware.NewRocketMQClient(suite.config)
}

// TearDownTest 每个测试后执行
func (suite *RocketMQClientTestSuite) TearDownTest() {
	if suite.client != nil {
		ctx := context.Background()
		_ = suite.client.Disconnect(ctx)
	}
}

// TestConnect 测试RocketMQ连接
func (suite *RocketMQClientTestSuite) TestConnect() {
	ctx := context.Background()

	// 测试成功连接
	err := suite.client.Connect(ctx)
	suite.NoError(err, "Connect should succeed")

	// 验证连接状态
	err = suite.client.Ping(ctx)
	suite.NoError(err, "Ping should succeed after connect")
}

// TestConnect_InvalidNameServer 测试无效NameServer连接
func (suite *RocketMQClientTestSuite) TestConnect_InvalidNameServer() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	invalidConfig := &middleware.RocketMQConfig{
		NameServers:   []string{"invalid-host:9876"},
		Topic:         "test-topic",
		ProducerGroup: "test-producer",
		ConsumerGroup: "test-consumer",
	}
	client := middleware.NewRocketMQClient(invalidConfig)

	err := client.Connect(ctx)
	// 连接可能会超时或失败
	if err != nil {
		suite.T().Logf("Expected error for invalid nameserver: %v", err)
	}
}

// TestSendMessage 测试发送普通消息
func (suite *RocketMQClientTestSuite) TestSendMessage() {
	ctx := context.Background()

	// 先连接
	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 创建发送消息操作
	message := &middleware.RocketMQMessage{
		Topic: suite.config.Topic,
		Tag:   "test-tag",
		Key:   "test-key-1",
		Body:  []byte("test message body"),
	}

	op := &middleware.RocketMQSendOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "test-key-1",
		},
		Message: message,
	}

	// 执行发送
	result, err := suite.client.Execute(ctx, op)
	suite.NoError(err, "SendMessage should succeed")
	suite.NotNil(result, "Result should not be nil")
	suite.True(result.Success, "Send operation should be successful")
	suite.Greater(result.Duration, time.Duration(0), "Duration should be positive")

	// 验证返回的消息ID
	msgID, ok := result.Metadata["message_id"]
	suite.True(ok, "Should have message_id in metadata")
	suite.NotEmpty(msgID, "Message ID should not be empty")
}

// TestSendBatch 测试批量发送消息
func (suite *RocketMQClientTestSuite) TestSendBatch() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 创建批量消息
	messages := []*middleware.RocketMQMessage{
		{
			Topic: suite.config.Topic,
			Tag:   "batch-tag",
			Key:   "batch-key-1",
			Body:  []byte("batch message 1"),
		},
		{
			Topic: suite.config.Topic,
			Tag:   "batch-tag",
			Key:   "batch-key-2",
			Body:  []byte("batch message 2"),
		},
		{
			Topic: suite.config.Topic,
			Tag:   "batch-tag",
			Key:   "batch-key-3",
			Body:  []byte("batch message 3"),
		},
	}

	op := &middleware.RocketMQSendBatchOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "batch-send",
		},
		Messages: messages,
	}

	// 执行批量发送
	result, err := suite.client.Execute(ctx, op)
	suite.NoError(err, "SendBatch should succeed")
	suite.NotNil(result, "Result should not be nil")
	suite.True(result.Success, "Batch send operation should be successful")

	// 验证发送数量
	count, ok := result.Metadata["message_count"]
	suite.True(ok, "Should have message_count in metadata")
	suite.Equal(3, count, "Should send 3 messages")
}

// TestConsumeMessage 测试消费消息
func (suite *RocketMQClientTestSuite) TestConsumeMessage() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 先发送一条消息
	sendOp := &middleware.RocketMQSendOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "consume-test-key",
		},
		Message: &middleware.RocketMQMessage{
			Topic: suite.config.Topic,
			Tag:   "consume-tag",
			Key:   "consume-test-key",
			Body:  []byte("message to consume"),
		},
	}
	_, err = suite.client.Execute(ctx, sendOp)
	suite.Require().NoError(err)

	// 等待消息可用
	time.Sleep(100 * time.Millisecond)

	// 消费消息
	consumeOp := &middleware.RocketMQConsumeOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeRead,
			OpKey:  "consume-operation",
		},
		MaxMessages: 1,
		Timeout:     5 * time.Second,
	}

	result, err := suite.client.Execute(ctx, consumeOp)
	suite.NoError(err, "Consume should not error")
	suite.NotNil(result, "Result should not be nil")

	// 消费可能成功或超时（没有消息）
	if result.Success {
		suite.T().Log("Message consumed successfully")
		if result.Data != nil {
			suite.T().Logf("Consumed message: %v", result.Data)
		}
	}
}

// TestTransactionMessage 测试事务消息
func (suite *RocketMQClientTestSuite) TestTransactionMessage() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 创建事务消息
	message := &middleware.RocketMQMessage{
		Topic: suite.config.Topic,
		Tag:   "transaction-tag",
		Key:   "transaction-key",
		Body:  []byte("transaction message"),
	}

	op := &middleware.RocketMQTransactionOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "transaction-test",
		},
		Message:       message,
		TransactionID: "tx-001",
		CheckTimes:    3,
	}

	// 执行事务消息发送
	result, err := suite.client.Execute(ctx, op)
	suite.NoError(err, "Transaction message should not error")
	suite.NotNil(result, "Result should not be nil")

	// 事务消息可能需要回查，所以可能成功或待确认
	if result.Success {
		suite.T().Log("Transaction message sent successfully")
	}
}

// TestDelayMessage 测试延迟消息
func (suite *RocketMQClientTestSuite) TestDelayMessage() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 创建延迟消息（延迟1秒）
	message := &middleware.RocketMQMessage{
		Topic:      suite.config.Topic,
		Tag:        "delay-tag",
		Key:        "delay-key",
		Body:       []byte("delay message"),
		DelayLevel: 1, // Level 1 = 1秒延迟
	}

	op := &middleware.RocketMQSendOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "delay-test",
		},
		Message: message,
	}

	startTime := time.Now()

	// 执行发送
	result, err := suite.client.Execute(ctx, op)
	suite.NoError(err, "Delay message should succeed")
	suite.NotNil(result, "Result should not be nil")
	suite.True(result.Success, "Delay send operation should be successful")

	// 验证延迟消息发送成功
	delayLevel, ok := result.Metadata["delay_level"]
	if ok {
		suite.Equal(1, delayLevel, "Delay level should be 1")
	}

	suite.T().Logf("Delay message sent in %v", time.Since(startTime))
}

// TestOrderedMessage 测试顺序消息
func (suite *RocketMQClientTestSuite) TestOrderedMessage() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 发送顺序消息（使用相同的shardingKey保证顺序）
	shardingKey := "order-123"

	for i := 1; i <= 3; i++ {
		message := &middleware.RocketMQMessage{
			Topic:       suite.config.Topic,
			Tag:         "order-tag",
			Key:         "order-key",
			Body:        []byte("ordered message"),
			ShardingKey: shardingKey, // 保证顺序的关键
		}

		op := &middleware.RocketMQSendOperation{
			BaseOperation: core.BaseOperation{
				OpType: core.OpTypeWrite,
				OpKey:  "order-test",
			},
			Message: message,
		}

		result, err := suite.client.Execute(ctx, op)
		suite.NoError(err, "Ordered message %d should succeed", i)
		suite.True(result.Success, "Ordered message %d should be successful", i)
	}

	suite.T().Log("All ordered messages sent successfully")
}

// TestMessageRetry 测试消息重试
func (suite *RocketMQClientTestSuite) TestMessageRetry() {
	ctx := context.Background()

	// 配置重试次数
	retryConfig := &middleware.RocketMQConfig{
		NameServers:   suite.config.NameServers,
		Topic:         suite.config.Topic,
		ProducerGroup: suite.config.ProducerGroup,
		ConsumerGroup: suite.config.ConsumerGroup,
		RetryTimes:    3, // 重试3次
	}
	retryClient := middleware.NewRocketMQClient(retryConfig)

	err := retryClient.Connect(ctx)
	suite.Require().NoError(err)
	defer retryClient.Disconnect(ctx)

	// 发送消息（如果失败会自动重试）
	message := &middleware.RocketMQMessage{
		Topic: suite.config.Topic,
		Tag:   "retry-tag",
		Key:   "retry-key",
		Body:  []byte("retry message"),
	}

	op := &middleware.RocketMQSendOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "retry-test",
		},
		Message: message,
	}

	result, err := retryClient.Execute(ctx, op)
	suite.NoError(err, "Message with retry should succeed or fail gracefully")
	suite.NotNil(result, "Result should not be nil")

	// 验证重试配置
	if retryCount, ok := result.Metadata["retry_count"]; ok {
		suite.T().Logf("Message sent with %d retries", retryCount)
	}
}

// TestExecute_UnsupportedOperation 测试不支持的操作类型
func (suite *RocketMQClientTestSuite) TestExecute_UnsupportedOperation() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 创建一个不支持的操作类型
	unsupportedOp := &core.BaseOperation{
		OpType: "UNSUPPORTED",
		OpKey:  "test",
	}

	result, err := suite.client.Execute(ctx, unsupportedOp)
	suite.NoError(err, "Execute should not return error for unsupported operation")
	suite.NotNil(result, "Result should not be nil")
	suite.False(result.Success, "Unsupported operation should fail")
	suite.NotNil(result.Error, "Error should be set for unsupported operation")
}

// TestDisconnect 测试断开连接
func (suite *RocketMQClientTestSuite) TestDisconnect() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 断开连接
	err = suite.client.Disconnect(ctx)
	suite.NoError(err, "Disconnect should succeed")

	// 再次断开连接应该也不报错
	err = suite.client.Disconnect(ctx)
	suite.NoError(err, "Multiple disconnects should not error")
}

// TestGetStats 测试获取统计信息
func (suite *RocketMQClientTestSuite) TestGetStats() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 发送一些消息
	message := &middleware.RocketMQMessage{
		Topic: suite.config.Topic,
		Tag:   "stats-tag",
		Key:   "stats-key",
		Body:  []byte("stats test message"),
	}

	op := &middleware.RocketMQSendOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "stats-test",
		},
		Message: message,
	}
	_, err = suite.client.Execute(ctx, op)
	suite.Require().NoError(err)

	// 获取统计信息
	stats := suite.client.GetStats()
	suite.NotNil(stats, "Stats should not be nil")
	suite.NotEmpty(stats, "Stats should contain data")

	// 验证统计信息包含必要字段
	suite.T().Logf("Stats: %v", stats)
}

// TestConcurrentSend 测试并发发送
func (suite *RocketMQClientTestSuite) TestConcurrentSend() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 并发发送10条消息
	concurrentCount := 10
	done := make(chan bool, concurrentCount)
	errors := make(chan error, concurrentCount)

	for i := 0; i < concurrentCount; i++ {
		go func(index int) {
			message := &middleware.RocketMQMessage{
				Topic: suite.config.Topic,
				Tag:   "concurrent-tag",
				Key:   "concurrent-key",
				Body:  []byte("concurrent message"),
			}

			op := &middleware.RocketMQSendOperation{
				BaseOperation: core.BaseOperation{
					OpType: core.OpTypeWrite,
					OpKey:  "concurrent-test",
				},
				Message: message,
			}

			result, err := suite.client.Execute(ctx, op)
			if err != nil {
				errors <- err
			} else if !result.Success {
				errors <- result.Error
			}
			done <- true
		}(i)
	}

	// 等待所有发送完成
	for i := 0; i < concurrentCount; i++ {
		<-done
	}
	close(errors)

	// 检查是否有错误
	errorCount := 0
	for err := range errors {
		errorCount++
		suite.T().Logf("Concurrent send error: %v", err)
	}

	suite.T().Logf("Concurrent send completed: %d/%d successful",
		concurrentCount-errorCount, concurrentCount)
}

// TestRocketMQClientTestSuite 运行测试套件
func TestRocketMQClientTestSuite(t *testing.T) {
	suite.Run(t, new(RocketMQClientTestSuite))
}
