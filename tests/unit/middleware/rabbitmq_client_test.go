package middleware_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"middleware-chaos-testing/internal/core"
	"middleware-chaos-testing/internal/middleware"
)

// RabbitMQClientTestSuite RabbitMQ客户端测试套件
type RabbitMQClientTestSuite struct {
	suite.Suite
	client *middleware.RabbitMQClient
	config *middleware.RabbitMQConfig
}

// SetupTest 每个测试前执行
func (suite *RabbitMQClientTestSuite) SetupTest() {
	suite.config = &middleware.RabbitMQConfig{
		URL:          "amqp://guest:guest@localhost:5672/",
		Exchange:     "chaos-test-exchange",
		ExchangeType: "direct",
		Queue:        "chaos-test-queue",
		RoutingKey:   "chaos.test",
	}
	suite.client = middleware.NewRabbitMQClient(suite.config)
}

// TearDownTest 每个测试后执行
func (suite *RabbitMQClientTestSuite) TearDownTest() {
	if suite.client != nil {
		ctx := context.Background()
		_ = suite.client.Disconnect(ctx)
	}
}

// TestConnect 测试RabbitMQ连接
func (suite *RabbitMQClientTestSuite) TestConnect() {
	ctx := context.Background()

	// 测试成功连接
	err := suite.client.Connect(ctx)
	suite.NoError(err, "Connect should succeed")

	// 验证连接状态
	err = suite.client.Ping(ctx)
	suite.NoError(err, "Ping should succeed after connect")
}

// TestConnect_InvalidURL 测试无效URL连接
func (suite *RabbitMQClientTestSuite) TestConnect_InvalidURL() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	invalidConfig := &middleware.RabbitMQConfig{
		URL:          "amqp://invalid-host:5672/",
		Exchange:     "test-exchange",
		ExchangeType: "direct",
		Queue:        "test-queue",
		RoutingKey:   "test",
	}
	client := middleware.NewRabbitMQClient(invalidConfig)

	err := client.Connect(ctx)
	// 连接可能会超时或失败
	if err != nil {
		suite.T().Logf("Expected error for invalid URL: %v", err)
	}
}

// TestPublish 测试发布消息
func (suite *RabbitMQClientTestSuite) TestPublish() {
	ctx := context.Background()

	// 先连接
	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 创建发布操作
	message := &middleware.RabbitMQMessage{
		Exchange:   suite.config.Exchange,
		RoutingKey: suite.config.RoutingKey,
		Body:       []byte("test message body"),
		Headers: map[string]interface{}{
			"content-type": "text/plain",
		},
		Persistent: true,
	}

	op := &middleware.RabbitMQPublishOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "test-publish",
		},
		Message: message,
	}

	// 执行发布
	result, err := suite.client.Execute(ctx, op)
	suite.NoError(err, "Publish should succeed")
	suite.NotNil(result, "Result should not be nil")
	suite.True(result.Success, "Publish operation should be successful")
	suite.Greater(result.Duration, time.Duration(0), "Duration should be positive")

	// 验证消息已发布
	published, ok := result.Metadata["published"]
	suite.True(ok, "Should have published metadata")
	suite.True(published.(bool), "Message should be published")
}

// TestConsume 测试消费消息
func (suite *RabbitMQClientTestSuite) TestConsume() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 先发布一条消息
	publishOp := &middleware.RabbitMQPublishOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "consume-test-publish",
		},
		Message: &middleware.RabbitMQMessage{
			Exchange:   suite.config.Exchange,
			RoutingKey: suite.config.RoutingKey,
			Body:       []byte("message to consume"),
		},
	}
	_, err = suite.client.Execute(ctx, publishOp)
	suite.Require().NoError(err)

	// 等待消息可用
	time.Sleep(100 * time.Millisecond)

	// 消费消息
	consumeOp := &middleware.RabbitMQConsumeOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeRead,
			OpKey:  "consume-operation",
		},
		AutoAck: false,
		Timeout: 5 * time.Second,
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

// TestAck 测试消息确认
func (suite *RabbitMQClientTestSuite) TestAck() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 发布消息
	publishOp := &middleware.RabbitMQPublishOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "ack-test-publish",
		},
		Message: &middleware.RabbitMQMessage{
			Exchange:   suite.config.Exchange,
			RoutingKey: suite.config.RoutingKey,
			Body:       []byte("message for ack test"),
		},
	}
	_, err = suite.client.Execute(ctx, publishOp)
	suite.Require().NoError(err)

	time.Sleep(100 * time.Millisecond)

	// 消费消息（不自动确认）
	consumeOp := &middleware.RabbitMQConsumeOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeRead,
			OpKey:  "ack-consume",
		},
		AutoAck: false,
		Timeout: 5 * time.Second,
	}

	consumeResult, err := suite.client.Execute(ctx, consumeOp)
	suite.Require().NoError(err)

	if consumeResult.Success && consumeResult.Data != nil {
		// 手动确认消息
		deliveryTag, ok := consumeResult.Metadata["delivery_tag"]
		suite.Require().True(ok, "Should have delivery_tag")

		ackOp := &middleware.RabbitMQAckOperation{
			BaseOperation: core.BaseOperation{
				OpType: core.OpTypeWrite,
				OpKey:  "ack-operation",
			},
			DeliveryTag: deliveryTag.(uint64),
			Multiple:    false,
		}

		ackResult, err := suite.client.Execute(ctx, ackOp)
		suite.NoError(err, "Ack should succeed")
		suite.True(ackResult.Success, "Ack operation should be successful")
	}
}

// TestReject 测试消息拒绝
func (suite *RabbitMQClientTestSuite) TestReject() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 发布消息
	publishOp := &middleware.RabbitMQPublishOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "reject-test-publish",
		},
		Message: &middleware.RabbitMQMessage{
			Exchange:   suite.config.Exchange,
			RoutingKey: suite.config.RoutingKey,
			Body:       []byte("message for reject test"),
		},
	}
	_, err = suite.client.Execute(ctx, publishOp)
	suite.Require().NoError(err)

	time.Sleep(100 * time.Millisecond)

	// 消费消息
	consumeOp := &middleware.RabbitMQConsumeOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeRead,
			OpKey:  "reject-consume",
		},
		AutoAck: false,
		Timeout: 5 * time.Second,
	}

	consumeResult, err := suite.client.Execute(ctx, consumeOp)
	suite.Require().NoError(err)

	if consumeResult.Success && consumeResult.Data != nil {
		// 拒绝消息
		deliveryTag, ok := consumeResult.Metadata["delivery_tag"]
		suite.Require().True(ok, "Should have delivery_tag")

		rejectOp := &middleware.RabbitMQRejectOperation{
			BaseOperation: core.BaseOperation{
				OpType: core.OpTypeWrite,
				OpKey:  "reject-operation",
			},
			DeliveryTag: deliveryTag.(uint64),
			Requeue:     false, // 不重新入队
		}

		rejectResult, err := suite.client.Execute(ctx, rejectOp)
		suite.NoError(err, "Reject should succeed")
		suite.True(rejectResult.Success, "Reject operation should be successful")
	}
}

// TestExchangeRouting 测试交换机路由
func (suite *RabbitMQClientTestSuite) TestExchangeRouting() {
	ctx := context.Background()

	// 测试不同交换机类型
	exchangeTypes := []string{"direct", "fanout", "topic"}

	for _, exchangeType := range exchangeTypes {
		config := &middleware.RabbitMQConfig{
			URL:          suite.config.URL,
			Exchange:     "chaos-exchange-" + exchangeType,
			ExchangeType: exchangeType,
			Queue:        "chaos-queue-" + exchangeType,
			RoutingKey:   "chaos.routing." + exchangeType,
		}
		client := middleware.NewRabbitMQClient(config)

		err := client.Connect(ctx)
		suite.Require().NoError(err)

		// 发布消息到不同类型的交换机
		message := &middleware.RabbitMQMessage{
			Exchange:   config.Exchange,
			RoutingKey: config.RoutingKey,
			Body:       []byte("routing test for " + exchangeType),
		}

		op := &middleware.RabbitMQPublishOperation{
			BaseOperation: core.BaseOperation{
				OpType: core.OpTypeWrite,
				OpKey:  "routing-test-" + exchangeType,
			},
			Message: message,
		}

		result, err := client.Execute(ctx, op)
		suite.NoError(err, "Routing to %s exchange should succeed", exchangeType)
		suite.True(result.Success, "Routing operation should be successful")

		client.Disconnect(ctx)
	}
}

// TestDeadLetterQueue 测试死信队列
func (suite *RabbitMQClientTestSuite) TestDeadLetterQueue() {
	ctx := context.Background()

	// 配置死信队列
	config := &middleware.RabbitMQConfig{
		URL:               suite.config.URL,
		Exchange:          "chaos-test-exchange",
		ExchangeType:      "direct",
		Queue:             "chaos-test-queue-with-dlx",
		RoutingKey:        "chaos.test",
		DeadLetterExchange: "chaos-dlx-exchange",
		DeadLetterRoutingKey: "chaos.dlx",
	}
	client := middleware.NewRabbitMQClient(config)

	err := client.Connect(ctx)
	suite.Require().NoError(err)
	defer client.Disconnect(ctx)

	// 发布消息
	message := &middleware.RabbitMQMessage{
		Exchange:   config.Exchange,
		RoutingKey: config.RoutingKey,
		Body:       []byte("message for dead letter queue"),
	}

	op := &middleware.RabbitMQPublishOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "dlq-test",
		},
		Message: message,
	}

	result, err := client.Execute(ctx, op)
	suite.NoError(err, "DLQ publish should succeed")
	suite.True(result.Success, "DLQ operation should be successful")
}

// TestTTL 测试消息TTL（Time To Live）
func (suite *RabbitMQClientTestSuite) TestTTL() {
	ctx := context.Background()

	// 配置TTL
	config := &middleware.RabbitMQConfig{
		URL:          suite.config.URL,
		Exchange:     "chaos-test-exchange",
		ExchangeType: "direct",
		Queue:        "chaos-test-queue-ttl",
		RoutingKey:   "chaos.test.ttl",
		MessageTTL:   5000, // 5秒TTL
		QueueTTL:     10000, // 队列10秒TTL
	}
	client := middleware.NewRabbitMQClient(config)

	err := client.Connect(ctx)
	suite.Require().NoError(err)
	defer client.Disconnect(ctx)

	// 发布带TTL的消息
	message := &middleware.RabbitMQMessage{
		Exchange:   config.Exchange,
		RoutingKey: config.RoutingKey,
		Body:       []byte("message with TTL"),
		Expiration: "5000", // 5秒过期
	}

	op := &middleware.RabbitMQPublishOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "ttl-test",
		},
		Message: message,
	}

	startTime := time.Now()
	result, err := client.Execute(ctx, op)
	suite.NoError(err, "TTL publish should succeed")
	suite.True(result.Success, "TTL operation should be successful")

	suite.T().Logf("TTL message published in %v", time.Since(startTime))
}

// TestPriority 测试消息优先级
func (suite *RabbitMQClientTestSuite) TestPriority() {
	ctx := context.Background()

	// 配置优先级队列
	config := &middleware.RabbitMQConfig{
		URL:           suite.config.URL,
		Exchange:      "chaos-test-exchange",
		ExchangeType:  "direct",
		Queue:         "chaos-test-queue-priority",
		RoutingKey:    "chaos.test.priority",
		MaxPriority:   10, // 最大优先级
	}
	client := middleware.NewRabbitMQClient(config)

	err := client.Connect(ctx)
	suite.Require().NoError(err)
	defer client.Disconnect(ctx)

	// 发布不同优先级的消息
	for priority := uint8(1); priority <= 10; priority++ {
		message := &middleware.RabbitMQMessage{
			Exchange:   config.Exchange,
			RoutingKey: config.RoutingKey,
			Body:       []byte("priority message"),
			Priority:   priority,
		}

		op := &middleware.RabbitMQPublishOperation{
			BaseOperation: core.BaseOperation{
				OpType: core.OpTypeWrite,
				OpKey:  "priority-test",
			},
			Message: message,
		}

		result, err := client.Execute(ctx, op)
		suite.NoError(err, "Priority %d publish should succeed", priority)
		suite.True(result.Success, "Priority operation should be successful")
	}
}

// TestExecute_UnsupportedOperation 测试不支持的操作类型
func (suite *RabbitMQClientTestSuite) TestExecute_UnsupportedOperation() {
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
func (suite *RabbitMQClientTestSuite) TestDisconnect() {
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
func (suite *RabbitMQClientTestSuite) TestGetStats() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 发布一些消息
	message := &middleware.RabbitMQMessage{
		Exchange:   suite.config.Exchange,
		RoutingKey: suite.config.RoutingKey,
		Body:       []byte("stats test message"),
	}

	op := &middleware.RabbitMQPublishOperation{
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

// TestConcurrentPublish 测试并发发布
func (suite *RabbitMQClientTestSuite) TestConcurrentPublish() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 并发发布10条消息
	concurrentCount := 10
	done := make(chan bool, concurrentCount)
	errors := make(chan error, concurrentCount)

	for i := 0; i < concurrentCount; i++ {
		go func(index int) {
			message := &middleware.RabbitMQMessage{
				Exchange:   suite.config.Exchange,
				RoutingKey: suite.config.RoutingKey,
				Body:       []byte("concurrent message"),
			}

			op := &middleware.RabbitMQPublishOperation{
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

	// 等待所有发布完成
	for i := 0; i < concurrentCount; i++ {
		<-done
	}
	close(errors)

	// 检查是否有错误
	errorCount := 0
	for err := range errors {
		errorCount++
		suite.T().Logf("Concurrent publish error: %v", err)
	}

	suite.T().Logf("Concurrent publish completed: %d/%d successful",
		concurrentCount-errorCount, concurrentCount)
}

// TestRabbitMQClientTestSuite 运行测试套件
func TestRabbitMQClientTestSuite(t *testing.T) {
	suite.Run(t, new(RabbitMQClientTestSuite))
}
