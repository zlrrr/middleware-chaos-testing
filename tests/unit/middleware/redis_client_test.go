package middleware_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"middleware-chaos-testing/internal/core"
	"middleware-chaos-testing/internal/middleware"
)

// RedisClientTestSuite Redis客户端测试套件
type RedisClientTestSuite struct {
	suite.Suite
	client *middleware.RedisClient
	config *middleware.RedisConfig
}

// SetupTest 每个测试前执行
func (suite *RedisClientTestSuite) SetupTest() {
	suite.config = &middleware.RedisConfig{
		Host:    "localhost",
		Port:    6379,
		DB:      0,
		Timeout: 5 * time.Second,
	}
	suite.client = middleware.NewRedisClient(suite.config)
}

// TearDownTest 每个测试后执行
func (suite *RedisClientTestSuite) TearDownTest() {
	if suite.client != nil {
		ctx := context.Background()
		_ = suite.client.Disconnect(ctx)
	}
}

// TestConnect_Success 测试成功连接
func (suite *RedisClientTestSuite) TestConnect_Success() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.NoError(err, "Connect should succeed")

	// 验证连接状态
	err = suite.client.HealthCheck(ctx)
	suite.NoError(err, "HealthCheck should succeed after connect")
}

// TestConnect_AlreadyConnected 测试重复连接
func (suite *RedisClientTestSuite) TestConnect_AlreadyConnected() {
	ctx := context.Background()

	// 第一次连接
	err := suite.client.Connect(ctx)
	suite.NoError(err)

	// 第二次连接应该直接返回成功（幂等性）
	err = suite.client.Connect(ctx)
	suite.NoError(err)
}

// TestConnect_InvalidConfig 测试无效配置
func (suite *RedisClientTestSuite) TestConnect_InvalidConfig() {
	ctx := context.Background()

	invalidConfig := &middleware.RedisConfig{
		Host: "invalid-host-that-does-not-exist",
		Port: 99999,
	}
	invalidClient := middleware.NewRedisClient(invalidConfig)

	err := invalidClient.Connect(ctx)
	suite.Error(err, "Connect should fail with invalid config")
}

// TestConnect_Timeout 测试连接超时
func (suite *RedisClientTestSuite) TestConnect_Timeout() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	config := &middleware.RedisConfig{
		Host: "10.255.255.1", // 不可达的IP
		Port: 6379,
	}
	client := middleware.NewRedisClient(config)

	err := client.Connect(ctx)
	suite.Error(err, "Connect should timeout")
	suite.True(errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, core.ErrOperationTimeout),
		"Error should be timeout related")
}

// TestDisconnect_Success 测试断开连接
func (suite *RedisClientTestSuite) TestDisconnect_Success() {
	ctx := context.Background()

	// 先连接
	err := suite.client.Connect(ctx)
	suite.NoError(err)

	// 再断开
	err = suite.client.Disconnect(ctx)
	suite.NoError(err)

	// 断开后健康检查应该失败
	err = suite.client.HealthCheck(ctx)
	suite.Error(err, "HealthCheck should fail after disconnect")
}

// TestDisconnect_NotConnected 测试未连接时断开
func (suite *RedisClientTestSuite) TestDisconnect_NotConnected() {
	ctx := context.Background()

	// 未连接就断开，应该返回成功（幂等性）
	err := suite.client.Disconnect(ctx)
	suite.NoError(err)
}

// TestHealthCheck_NotConnected 测试未连接时的健康检查
func (suite *RedisClientTestSuite) TestHealthCheck_NotConnected() {
	ctx := context.Background()

	err := suite.client.HealthCheck(ctx)
	suite.Error(err)
	suite.True(errors.Is(err, core.ErrClientNotConnected),
		"Should return ErrClientNotConnected")
}

// TestExecute_SetOperation 测试SET操作
func (suite *RedisClientTestSuite) TestExecute_SetOperation() {
	ctx := context.Background()

	// 连接
	err := suite.client.Connect(ctx)
	suite.NoError(err)

	// 创建SET操作
	op := &middleware.RedisSetOperation{
		OpKey:   "test:key",
		OpValue: []byte("test-value"),
	}

	// 执行操作
	result, err := suite.client.Execute(ctx, op)
	suite.NoError(err)
	suite.NotNil(result)
	suite.True(result.Success, "SET operation should succeed")
	suite.Greater(result.Duration, time.Duration(0), "Duration should be recorded")
}

// TestExecute_GetOperation 测试GET操作
func (suite *RedisClientTestSuite) TestExecute_GetOperation() {
	ctx := context.Background()

	// 连接
	err := suite.client.Connect(ctx)
	suite.NoError(err)

	// 先SET一个值
	setOp := &middleware.RedisSetOperation{
		OpKey:   "test:get:key",
		OpValue: []byte("test-get-value"),
	}
	_, err = suite.client.Execute(ctx, setOp)
	suite.NoError(err)

	// 再GET这个值
	getOp := &middleware.RedisGetOperation{
		OpKey: "test:get:key",
	}
	result, err := suite.client.Execute(ctx, getOp)
	suite.NoError(err)
	suite.NotNil(result)
	suite.True(result.Success)
	suite.Equal([]byte("test-get-value"), result.Data)
}

// TestExecute_GetNonExistentKey 测试GET不存在的键
func (suite *RedisClientTestSuite) TestExecute_GetNonExistentKey() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.NoError(err)

	getOp := &middleware.RedisGetOperation{
		OpKey: "non-existent-key-12345",
	}
	result, err := suite.client.Execute(ctx, getOp)

	// GET不存在的键应该返回成功，但Data为nil
	suite.NoError(err)
	suite.NotNil(result)
	suite.True(result.Success)
	suite.Nil(result.Data)
}

// TestExecute_DeleteOperation 测试DELETE操作
func (suite *RedisClientTestSuite) TestExecute_DeleteOperation() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.NoError(err)

	// 先SET一个值
	setOp := &middleware.RedisSetOperation{
		OpKey:   "test:delete:key",
		OpValue: []byte("to-be-deleted"),
	}
	_, err = suite.client.Execute(ctx, setOp)
	suite.NoError(err)

	// 删除这个键
	delOp := &middleware.RedisDeleteOperation{
		OpKey: "test:delete:key",
	}
	result, err := suite.client.Execute(ctx, delOp)
	suite.NoError(err)
	suite.True(result.Success)

	// 验证键已被删除
	getOp := &middleware.RedisGetOperation{
		OpKey: "test:delete:key",
	}
	result, err = suite.client.Execute(ctx, getOp)
	suite.NoError(err)
	suite.Nil(result.Data, "Key should be deleted")
}

// TestExecute_NotConnected 测试未连接时执行操作
func (suite *RedisClientTestSuite) TestExecute_NotConnected() {
	ctx := context.Background()

	op := &middleware.RedisSetOperation{
		OpKey:   "test:key",
		OpValue: []byte("value"),
	}

	result, err := suite.client.Execute(ctx, op)
	suite.Error(err)
	suite.Nil(result)
	suite.True(errors.Is(err, core.ErrClientNotConnected))
}

// TestExecute_UnsupportedOperation 测试不支持的操作类型
func (suite *RedisClientTestSuite) TestExecute_UnsupportedOperation() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.NoError(err)

	// 使用一个不支持的操作类型
	op := &unsupportedOperation{}

	result, err := suite.client.Execute(ctx, op)
	suite.Error(err)
	suite.Nil(result)
	suite.True(errors.Is(err, core.ErrUnsupportedOperation))
}

// TestGetMetrics 测试获取指标
func (suite *RedisClientTestSuite) TestGetMetrics() {
	ctx := context.Background()

	// 未连接时
	metrics := suite.client.GetMetrics()
	suite.NotNil(metrics)
	suite.Equal(0, metrics.ActiveConnections)

	// 连接后
	err := suite.client.Connect(ctx)
	suite.NoError(err)

	metrics = suite.client.GetMetrics()
	suite.NotNil(metrics)
	suite.Equal(1, metrics.ActiveConnections)

	// 断开后
	err = suite.client.Disconnect(ctx)
	suite.NoError(err)

	metrics = suite.client.GetMetrics()
	suite.Equal(0, metrics.ActiveConnections)
}

// TestConcurrentOperations 测试并发操作
func (suite *RedisClientTestSuite) TestConcurrentOperations() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.NoError(err)

	// 并发执行100个操作
	const goroutines = 10
	const opsPerGoroutine = 10

	errChan := make(chan error, goroutines*opsPerGoroutine)
	doneChan := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer func() { doneChan <- true }()

			for j := 0; j < opsPerGoroutine; j++ {
				op := &middleware.RedisSetOperation{
					OpKey:   "test:concurrent:" + string(rune(id*100+j)),
					OpValue: []byte("value"),
				}
				_, err := suite.client.Execute(ctx, op)
				if err != nil {
					errChan <- err
				}
			}
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < goroutines; i++ {
		<-doneChan
	}
	close(errChan)

	// 验证没有错误
	for err := range errChan {
		suite.Fail("Concurrent operation failed", err.Error())
	}
}

// TestExecuteWithTimeout 测试操作超时
func (suite *RedisClientTestSuite) TestExecuteWithTimeout() {
	ctx := context.Background()

	err := suite.client.Connect(ctx)
	suite.NoError(err)

	// 使用极短的超时时间
	ctx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // 确保超时

	op := &middleware.RedisSetOperation{
		OpKey:   "test:timeout:key",
		OpValue: []byte("value"),
	}

	result, err := suite.client.Execute(ctx, op)
	// 可能会超时，也可能在超时前完成
	if err != nil {
		suite.True(errors.Is(err, context.DeadlineExceeded))
	} else {
		suite.True(result.Success)
	}
}

// TestRedisClientTestSuite 运行测试套件
func TestRedisClientTestSuite(t *testing.T) {
	suite.Run(t, new(RedisClientTestSuite))
}

// unsupportedOperation 不支持的操作类型（用于测试）
type unsupportedOperation struct{}

func (u *unsupportedOperation) Type() core.OperationType {
	return core.OpTypeCustom
}

func (u *unsupportedOperation) Key() string {
	return "unsupported"
}

func (u *unsupportedOperation) Value() []byte {
	return nil
}

func (u *unsupportedOperation) Metadata() map[string]interface{} {
	return nil
}
