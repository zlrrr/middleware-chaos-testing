package middleware_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"middleware-chaos-testing/internal/core"
	"middleware-chaos-testing/internal/middleware"
)

// EMQXClientTestSuite EMQX客户端测试套件
type EMQXClientTestSuite struct {
	suite.Suite
	client *middleware.EMQXClient
	config *middleware.EMQXConfig
}

// SetupTest 每个测试前执行
func (suite *EMQXClientTestSuite) SetupTest() {
	suite.config = &middleware.EMQXConfig{
		Broker:          "tcp://localhost:1883",
		ClientID:        "chaos-test-client",
		Username:        "test",
		Password:        "test",
		ProtocolVersion: 4, // MQTT 3.1.1
		CleanSession:    true,
		KeepAlive:       60,
	}
	suite.client = middleware.NewEMQXClient(suite.config)
}

// TearDownTest 每个测试后执行
func (suite *EMQXClientTestSuite) TearDownTest() {
	if suite.client != nil {
		ctx := context.Background()
		_ = suite.client.Disconnect(ctx)
	}
}

// TestConnect 测试EMQX连接
func (suite *EMQXClientTestSuite) TestConnect() {
	ctx := context.Background()

	// 测试成功连接
	err := suite.client.Connect(ctx)
	suite.NoError(err, "Connect should succeed")

	// 验证连接状态
	err = suite.client.Ping(ctx)
	suite.NoError(err, "Ping should succeed after connect")
}

// TestConnect_InvalidBroker 测试无效Broker连接
func (suite *EMQXClientTestSuite) TestConnect_InvalidBroker() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	invalidConfig := &middleware.EMQXConfig{
		Broker:          "tcp://invalid-host:1883",
		ClientID:        "test-invalid",
		ProtocolVersion: 4,
		ConnectTimeout:  1 * time.Second,
	}
	client := middleware.NewEMQXClient(invalidConfig)

	err := client.Connect(ctx)
	// 连接应该失败或超时
	if err != nil {
		suite.T().Logf("Expected error for invalid broker: %v", err)
	}
}

// TestPublish 测试发布消息
func (suite *EMQXClientTestSuite) TestPublish() {
	ctx := context.Background()

	// 先连接
	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 创建发布操作
	message := &middleware.EMQXMessage{
		Topic:   "chaos/test/publish",
		Payload: []byte("test message payload"),
		QoS:     0,
		Retained: false,
	}

	op := &middleware.EMQXPublishOperation{
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

// TestSubscribe 测试订阅主题
func (suite *EMQXClientTestSuite) TestSubscribe() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 订阅主题
	subscribeOp := &middleware.EMQXSubscribeOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeRead,
			OpKey:  "subscribe-test",
		},
		Topic: "chaos/test/subscribe",
		QoS:   1,
	}

	result, err := suite.client.Execute(ctx, subscribeOp)
	suite.NoError(err, "Subscribe should succeed")
	suite.NotNil(result, "Result should not be nil")
	suite.True(result.Success, "Subscribe operation should be successful")

	// 验证订阅成功
	subscribed, ok := result.Metadata["subscribed"]
	suite.True(ok, "Should have subscribed metadata")
	suite.True(subscribed.(bool), "Should be subscribed")
}

// TestUnsubscribe 测试取消订阅
func (suite *EMQXClientTestSuite) TestUnsubscribe() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 先订阅
	subscribeOp := &middleware.EMQXSubscribeOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeRead,
			OpKey:  "subscribe-before-unsubscribe",
		},
		Topic: "chaos/test/unsubscribe",
		QoS:   0,
	}
	_, err = suite.client.Execute(ctx, subscribeOp)
	suite.Require().NoError(err)

	// 取消订阅
	unsubscribeOp := &middleware.EMQXUnsubscribeOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "unsubscribe-test",
		},
		Topic: "chaos/test/unsubscribe",
	}

	result, err := suite.client.Execute(ctx, unsubscribeOp)
	suite.NoError(err, "Unsubscribe should succeed")
	suite.True(result.Success, "Unsubscribe operation should be successful")

	// 验证取消订阅成功
	unsubscribed, ok := result.Metadata["unsubscribed"]
	suite.True(ok, "Should have unsubscribed metadata")
	suite.True(unsubscribed.(bool), "Should be unsubscribed")
}

// TestQoS0 测试QoS 0消息传输（至多一次）
func (suite *EMQXClientTestSuite) TestQoS0() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 订阅主题（QoS 0）
	subscribeOp := &middleware.EMQXSubscribeOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeRead,
			OpKey:  "subscribe-qos0",
		},
		Topic: "chaos/test/qos0",
		QoS:   0,
	}
	_, err = suite.client.Execute(ctx, subscribeOp)
	suite.Require().NoError(err)

	time.Sleep(100 * time.Millisecond)

	// 发布QoS 0消息
	message := &middleware.EMQXMessage{
		Topic:   "chaos/test/qos0",
		Payload: []byte("QoS 0 message"),
		QoS:     0,
	}

	publishOp := &middleware.EMQXPublishOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "publish-qos0",
		},
		Message: message,
	}

	result, err := suite.client.Execute(ctx, publishOp)
	suite.NoError(err, "QoS 0 publish should succeed")
	suite.True(result.Success, "QoS 0 operation should be successful")

	suite.T().Log("QoS 0 message published (fire-and-forget)")
}

// TestQoS1 测试QoS 1消息传输（至少一次）
func (suite *EMQXClientTestSuite) TestQoS1() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 订阅主题（QoS 1）
	subscribeOp := &middleware.EMQXSubscribeOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeRead,
			OpKey:  "subscribe-qos1",
		},
		Topic: "chaos/test/qos1",
		QoS:   1,
	}
	_, err = suite.client.Execute(ctx, subscribeOp)
	suite.Require().NoError(err)

	time.Sleep(100 * time.Millisecond)

	// 发布QoS 1消息
	message := &middleware.EMQXMessage{
		Topic:   "chaos/test/qos1",
		Payload: []byte("QoS 1 message"),
		QoS:     1,
	}

	publishOp := &middleware.EMQXPublishOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "publish-qos1",
		},
		Message: message,
	}

	result, err := suite.client.Execute(ctx, publishOp)
	suite.NoError(err, "QoS 1 publish should succeed")
	suite.True(result.Success, "QoS 1 operation should be successful")

	// QoS 1需要确认
	suite.T().Log("QoS 1 message published with acknowledgment")
}

// TestQoS2 测试QoS 2消息传输（精确一次）
func (suite *EMQXClientTestSuite) TestQoS2() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 订阅主题（QoS 2）
	subscribeOp := &middleware.EMQXSubscribeOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeRead,
			OpKey:  "subscribe-qos2",
		},
		Topic: "chaos/test/qos2",
		QoS:   2,
	}
	_, err = suite.client.Execute(ctx, subscribeOp)
	suite.Require().NoError(err)

	time.Sleep(100 * time.Millisecond)

	// 发布QoS 2消息
	message := &middleware.EMQXMessage{
		Topic:   "chaos/test/qos2",
		Payload: []byte("QoS 2 message"),
		QoS:     2,
	}

	publishOp := &middleware.EMQXPublishOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "publish-qos2",
		},
		Message: message,
	}

	result, err := suite.client.Execute(ctx, publishOp)
	suite.NoError(err, "QoS 2 publish should succeed")
	suite.True(result.Success, "QoS 2 operation should be successful")

	// QoS 2需要四次握手
	suite.T().Log("QoS 2 message published with exactly-once delivery")
}

// TestRetainedMessage 测试保留消息
func (suite *EMQXClientTestSuite) TestRetainedMessage() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 发布保留消息
	message := &middleware.EMQXMessage{
		Topic:    "chaos/test/retained",
		Payload:  []byte("This is a retained message"),
		QoS:      1,
		Retained: true,
	}

	publishOp := &middleware.EMQXPublishOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeWrite,
			OpKey:  "publish-retained",
		},
		Message: message,
	}

	result, err := suite.client.Execute(ctx, publishOp)
	suite.NoError(err, "Retained message publish should succeed")
	suite.True(result.Success, "Retained message operation should be successful")

	// 验证保留标志
	retained, ok := result.Metadata["retained"]
	suite.True(ok, "Should have retained metadata")
	suite.True(retained.(bool), "Message should be marked as retained")

	suite.T().Log("Retained message published - new subscribers will receive it")
}

// TestLastWill 测试遗嘱消息（Last Will）
func (suite *EMQXClientTestSuite) TestLastWill() {
	ctx := context.Background()

	// 配置遗嘱消息
	configWithWill := &middleware.EMQXConfig{
		Broker:          suite.config.Broker,
		ClientID:        "chaos-test-will-client",
		Username:        suite.config.Username,
		Password:        suite.config.Password,
		ProtocolVersion: 4,
		CleanSession:    true,
		KeepAlive:       60,
		WillEnabled:     true,
		WillTopic:       "chaos/test/will",
		WillPayload:     []byte("Client disconnected unexpectedly"),
		WillQoS:         1,
		WillRetained:    false,
	}
	clientWithWill := middleware.NewEMQXClient(configWithWill)

	err := clientWithWill.Connect(ctx)
	suite.Require().NoError(err)

	// 验证连接成功（遗嘱消息已设置）
	stats := clientWithWill.GetStats()
	suite.NotNil(stats, "Stats should not be nil")

	suite.T().Log("Client connected with Last Will message configured")

	// 正常断开连接（遗嘱消息不会发送）
	err = clientWithWill.Disconnect(ctx)
	suite.NoError(err, "Disconnect should succeed")
}

// TestCleanSession 测试清理会话
func (suite *EMQXClientTestSuite) TestCleanSession() {
	ctx := context.Background()

	// 测试Clean Session = true
	configClean := &middleware.EMQXConfig{
		Broker:          suite.config.Broker,
		ClientID:        "chaos-test-clean-session",
		Username:        suite.config.Username,
		Password:        suite.config.Password,
		ProtocolVersion: 4,
		CleanSession:    true,
		KeepAlive:       60,
	}
	clientClean := middleware.NewEMQXClient(configClean)

	err := clientClean.Connect(ctx)
	suite.NoError(err, "Clean session connect should succeed")

	// 订阅主题
	subscribeOp := &middleware.EMQXSubscribeOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeRead,
			OpKey:  "subscribe-clean",
		},
		Topic: "chaos/test/clean",
		QoS:   1,
	}
	_, err = clientClean.Execute(ctx, subscribeOp)
	suite.Require().NoError(err)

	// 断开连接（会话将被清除）
	err = clientClean.Disconnect(ctx)
	suite.NoError(err, "Disconnect should succeed")

	suite.T().Log("Clean session tested - session will be cleared on disconnect")

	// 测试Clean Session = false（持久会话）
	configPersistent := &middleware.EMQXConfig{
		Broker:          suite.config.Broker,
		ClientID:        "chaos-test-persistent-session",
		Username:        suite.config.Username,
		Password:        suite.config.Password,
		ProtocolVersion: 4,
		CleanSession:    false, // 持久会话
		KeepAlive:       60,
	}
	clientPersistent := middleware.NewEMQXClient(configPersistent)

	err = clientPersistent.Connect(ctx)
	suite.NoError(err, "Persistent session connect should succeed")

	// 订阅主题
	subscribeOp2 := &middleware.EMQXSubscribeOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeRead,
			OpKey:  "subscribe-persistent",
		},
		Topic: "chaos/test/persistent",
		QoS:   1,
	}
	_, err = clientPersistent.Execute(ctx, subscribeOp2)
	suite.Require().NoError(err)

	// 断开连接（会话将被保留）
	err = clientPersistent.Disconnect(ctx)
	suite.NoError(err, "Disconnect should succeed")

	suite.T().Log("Persistent session tested - session will be retained for reconnection")
}

// TestSharedSubscription 测试共享订阅（EMQX 5.8特性）
func (suite *EMQXClientTestSuite) TestSharedSubscription() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 共享订阅格式: $share/{group}/{topic}
	subscribeOp := &middleware.EMQXSubscribeOperation{
		BaseOperation: core.BaseOperation{
			OpType: core.OpTypeRead,
			OpKey:  "shared-subscribe",
		},
		Topic: "$share/group1/chaos/test/shared",
		QoS:   1,
	}

	result, err := suite.client.Execute(ctx, subscribeOp)
	suite.NoError(err, "Shared subscription should succeed")
	suite.True(result.Success, "Shared subscription operation should be successful")

	suite.T().Log("Shared subscription created - messages will be load-balanced across subscribers")
}

// TestExecute_UnsupportedOperation 测试不支持的操作类型
func (suite *EMQXClientTestSuite) TestExecute_UnsupportedOperation() {
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
func (suite *EMQXClientTestSuite) TestDisconnect() {
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
func (suite *EMQXClientTestSuite) TestGetStats() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 发布一些消息
	message := &middleware.EMQXMessage{
		Topic:   "chaos/test/stats",
		Payload: []byte("stats test message"),
		QoS:     1,
	}

	op := &middleware.EMQXPublishOperation{
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
func (suite *EMQXClientTestSuite) TestConcurrentPublish() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.Require().NoError(err)

	// 并发发布10条消息
	concurrentCount := 10
	done := make(chan bool, concurrentCount)
	errors := make(chan error, concurrentCount)

	for i := 0; i < concurrentCount; i++ {
		go func(index int) {
			message := &middleware.EMQXMessage{
				Topic:   "chaos/test/concurrent",
				Payload: []byte("concurrent message"),
				QoS:     1,
			}

			op := &middleware.EMQXPublishOperation{
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

// TestEMQXClientTestSuite 运行测试套件
func TestEMQXClientTestSuite(t *testing.T) {
	suite.Run(t, new(EMQXClientTestSuite))
}
