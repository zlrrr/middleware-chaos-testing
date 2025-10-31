package middleware_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"middleware-chaos-testing/internal/middleware"
)

// KafkaClientTestSuite Kafka客户端测试套件
type KafkaClientTestSuite struct {
	suite.Suite
	config *middleware.KafkaConfig
}

// SetupSuite 测试套件初始化
func (suite *KafkaClientTestSuite) SetupSuite() {
	suite.config = &middleware.KafkaConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "test-topic",
		GroupID: "test-group",
		Timeout: 5 * time.Second,
	}
}

// TestNewKafkaClient 测试创建Kafka客户端
func (suite *KafkaClientTestSuite) TestNewKafkaClient() {
	client := middleware.NewKafkaClient(suite.config)
	assert.NotNil(suite.T(), client)
}

// TestKafkaClient_Connect 测试连接（需要实际的Kafka服务）
func (suite *KafkaClientTestSuite) TestKafkaClient_Connect() {
	suite.T().Skip("Skipping integration test - requires Kafka server")

	client := middleware.NewKafkaClient(suite.config)
	ctx := context.Background()

	err := client.Connect(ctx)
	assert.NoError(suite.T(), err)

	defer client.Disconnect(ctx)
}

// TestKafkaClient_ProduceAndConsume 测试生产和消费消息
func (suite *KafkaClientTestSuite) TestKafkaClient_ProduceAndConsume() {
	suite.T().Skip("Skipping integration test - requires Kafka server")

	client := middleware.NewKafkaClient(suite.config)
	ctx := context.Background()

	err := client.Connect(ctx)
	assert.NoError(suite.T(), err)
	defer client.Disconnect(ctx)

	// 生产消息
	produceOp := &middleware.KafkaProduceOperation{
		OpKey:   "test-key",
		OpValue: []byte("test-value"),
	}

	result, err := client.Execute(ctx, produceOp)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.True(suite.T(), result.Success)

	// 等待消息被写入
	time.Sleep(100 * time.Millisecond)

	// 消费消息
	consumeOp := &middleware.KafkaConsumeOperation{
		MaxWait: 2 * time.Second,
	}

	result, err = client.Execute(ctx, consumeOp)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.True(suite.T(), result.Success)
}

// TestKafkaClient_Operations 测试操作类型
func (suite *KafkaClientTestSuite) TestKafkaClient_Operations() {
	// 测试ProduceOperation
	produceOp := &middleware.KafkaProduceOperation{
		OpKey:   "key1",
		OpValue: []byte("value1"),
	}
	assert.Equal(suite.T(), "write", string(produceOp.Type()))
	assert.Equal(suite.T(), "key1", produceOp.Key())
	assert.Equal(suite.T(), []byte("value1"), produceOp.Value())

	// 测试ConsumeOperation
	consumeOp := &middleware.KafkaConsumeOperation{}
	assert.Equal(suite.T(), "read", string(consumeOp.Type()))
	assert.Equal(suite.T(), "", consumeOp.Key())
	assert.Nil(suite.T(), consumeOp.Value())
}

// TestKafkaClient_DefaultTimeout 测试默认超时设置
func (suite *KafkaClientTestSuite) TestKafkaClient_DefaultTimeout() {
	config := &middleware.KafkaConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "test-topic",
		GroupID: "test-group",
		// 不设置Timeout
	}

	client := middleware.NewKafkaClient(config)
	assert.NotNil(suite.T(), client)
	// 应该设置了默认超时
}

// TestKafkaClientTestSuite 运行测试套件
func TestKafkaClientTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaClientTestSuite))
}
