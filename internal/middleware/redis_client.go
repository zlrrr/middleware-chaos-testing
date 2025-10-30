package middleware

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"middleware-chaos-testing/internal/core"
)

// RedisClient Redis客户端实现
type RedisClient struct {
	config  *RedisConfig
	client  *redis.Client
	mu      sync.RWMutex
	metrics *redisClientMetrics
}

// redisClientMetrics Redis客户端内部指标
type redisClientMetrics struct {
	mu                       sync.RWMutex
	activeConnections        int
	totalConnectionAttempts  int64
	failedConnectionAttempts int64
}

// NewRedisClient 创建新的Redis客户端
func NewRedisClient(config *RedisConfig) *RedisClient {
	return &RedisClient{
		config: config,
		metrics: &redisClientMetrics{
			activeConnections:        0,
			totalConnectionAttempts:  0,
			failedConnectionAttempts: 0,
		},
	}
}

// Connect 建立连接
func (r *RedisClient) Connect(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 增加连接尝试计数
	r.metrics.mu.Lock()
	r.metrics.totalConnectionAttempts++
	r.metrics.mu.Unlock()

	// 如果已经连接，直接返回（幂等性）
	if r.client != nil {
		// 验证连接是否有效
		if err := r.client.Ping(ctx).Err(); err == nil {
			return nil
		}
		// 连接无效，关闭旧连接
		_ = r.client.Close()
		r.client = nil
	}

	// 创建Redis客户端配置
	options := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", r.config.Host, r.config.Port),
		Password: r.config.Password,
		DB:       r.config.DB,
	}

	// 设置超时
	if r.config.Timeout > 0 {
		options.DialTimeout = r.config.Timeout
		options.ReadTimeout = r.config.Timeout
		options.WriteTimeout = r.config.Timeout
	}

	// 创建客户端
	client := redis.NewClient(options)

	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		r.metrics.mu.Lock()
		r.metrics.failedConnectionAttempts++
		r.metrics.mu.Unlock()
		_ = client.Close()
		return fmt.Errorf("%w: %v", core.ErrConnectionFailed, err)
	}

	r.client = client
	r.metrics.mu.Lock()
	r.metrics.activeConnections = 1
	r.metrics.mu.Unlock()

	return nil
}

// Disconnect 断开连接
func (r *RedisClient) Disconnect(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.client == nil {
		return nil // 幂等性：未连接时断开也返回成功
	}

	err := r.client.Close()
	r.client = nil

	r.metrics.mu.Lock()
	r.metrics.activeConnections = 0
	r.metrics.mu.Unlock()

	return err
}

// Execute 执行操作
func (r *RedisClient) Execute(ctx context.Context, op core.Operation) (*core.Result, error) {
	r.mu.RLock()
	client := r.client
	r.mu.RUnlock()

	if client == nil {
		return nil, core.ErrClientNotConnected
	}

	startTime := time.Now()

	// 根据操作类型执行不同的命令
	switch v := op.(type) {
	case *RedisSetOperation:
		return r.executeSet(ctx, client, v, startTime)
	case *RedisGetOperation:
		return r.executeGet(ctx, client, v, startTime)
	case *RedisDeleteOperation:
		return r.executeDelete(ctx, client, v, startTime)
	default:
		return nil, fmt.Errorf("%w: %T", core.ErrUnsupportedOperation, op)
	}
}

// executeSet 执行SET操作
func (r *RedisClient) executeSet(
	ctx context.Context,
	client *redis.Client,
	op *RedisSetOperation,
	startTime time.Time,
) (*core.Result, error) {
	err := client.Set(ctx, op.Key(), op.Value(), 0).Err()
	duration := time.Since(startTime)

	if err != nil {
		return &core.Result{
			Success:   false,
			Duration:  duration,
			Error:     err,
			Timestamp: time.Now(),
			Metadata:  make(map[string]interface{}),
		}, err
	}

	return &core.Result{
		Success:   true,
		Duration:  duration,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}, nil
}

// executeGet 执行GET操作
func (r *RedisClient) executeGet(
	ctx context.Context,
	client *redis.Client,
	op *RedisGetOperation,
	startTime time.Time,
) (*core.Result, error) {
	val, err := client.Get(ctx, op.Key()).Bytes()
	duration := time.Since(startTime)

	// Redis的GET命令，键不存在时返回redis.Nil错误
	// 这不算操作失败，而是正常的空值返回
	if err == redis.Nil {
		return &core.Result{
			Success:   true,
			Duration:  duration,
			Data:      nil,
			Timestamp: time.Now(),
			Metadata:  make(map[string]interface{}),
		}, nil
	}

	if err != nil {
		return &core.Result{
			Success:   false,
			Duration:  duration,
			Error:     err,
			Timestamp: time.Now(),
			Metadata:  make(map[string]interface{}),
		}, err
	}

	return &core.Result{
		Success:   true,
		Duration:  duration,
		Data:      val,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}, nil
}

// executeDelete 执行DELETE操作
func (r *RedisClient) executeDelete(
	ctx context.Context,
	client *redis.Client,
	op *RedisDeleteOperation,
	startTime time.Time,
) (*core.Result, error) {
	err := client.Del(ctx, op.Key()).Err()
	duration := time.Since(startTime)

	if err != nil {
		return &core.Result{
			Success:   false,
			Duration:  duration,
			Error:     err,
			Timestamp: time.Now(),
			Metadata:  make(map[string]interface{}),
		}, err
	}

	return &core.Result{
		Success:   true,
		Duration:  duration,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}, nil
}

// HealthCheck 健康检查
func (r *RedisClient) HealthCheck(ctx context.Context) error {
	r.mu.RLock()
	client := r.client
	r.mu.RUnlock()

	if client == nil {
		return core.ErrClientNotConnected
	}

	return client.Ping(ctx).Err()
}

// GetMetrics 获取客户端指标
func (r *RedisClient) GetMetrics() *core.ClientMetrics {
	r.metrics.mu.RLock()
	defer r.metrics.mu.RUnlock()

	return &core.ClientMetrics{
		ActiveConnections:        r.metrics.activeConnections,
		TotalConnectionAttempts:  r.metrics.totalConnectionAttempts,
		FailedConnectionAttempts: r.metrics.failedConnectionAttempts,
	}
}
