package core

import "time"

// Result 操作结果
type Result struct {
	// Success 操作是否成功
	Success bool

	// Duration 操作耗时
	Duration time.Duration

	// Error 错误信息（如果失败）
	Error error

	// Data 返回的数据（如果有）
	Data []byte

	// Metadata 结果元数据
	Metadata map[string]interface{}

	// Timestamp 操作完成时间
	Timestamp time.Time
}

// NewResult 创建新的结果
func NewResult(success bool, duration time.Duration, err error) *Result {
	return &Result{
		Success:   success,
		Duration:  duration,
		Error:     err,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}
