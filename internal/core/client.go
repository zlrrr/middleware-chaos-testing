package core

import "context"

// MiddlewareClient 中间件客户端接口
// 所有中间件适配器必须实现此接口
type MiddlewareClient interface {
	// Connect 建立连接
	// 返回错误表示连接失败
	Connect(ctx context.Context) error

	// Disconnect 断开连接
	// 返回错误表示断开失败
	Disconnect(ctx context.Context) error

	// Execute 执行操作
	// op: 要执行的操作
	// 返回操作结果和错误
	Execute(ctx context.Context, op Operation) (*Result, error)

	// HealthCheck 健康检查
	// 返回错误表示服务不健康
	HealthCheck(ctx context.Context) error

	// GetMetrics 获取客户端级别的指标
	// 返回当前的指标快照
	GetMetrics() *ClientMetrics
}

// ClientMetrics 客户端级别的指标
type ClientMetrics struct {
	// ActiveConnections 当前活跃连接数
	ActiveConnections int

	// TotalConnectionAttempts 总连接尝试数
	TotalConnectionAttempts int64

	// FailedConnectionAttempts 失败的连接尝试数
	FailedConnectionAttempts int64
}
