package core

import "errors"

// 标准错误定义
var (
	// ErrConnectionFailed 连接失败
	ErrConnectionFailed = errors.New("connection failed")

	// ErrOperationTimeout 操作超时
	ErrOperationTimeout = errors.New("operation timeout")

	// ErrInvalidConfig 无效配置
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrClientNotConnected 客户端未连接
	ErrClientNotConnected = errors.New("client not connected")

	// ErrUnsupportedOperation 不支持的操作
	ErrUnsupportedOperation = errors.New("unsupported operation")

	// ErrInvalidThresholds 无效的阈值配置
	ErrInvalidThresholds = errors.New("invalid thresholds")

	// ErrInvalidMetrics 无效的指标数据
	ErrInvalidMetrics = errors.New("invalid metrics")
)
