package core

import "time"

// ErrorType 错误类型
type ErrorType string

const (
	// ErrorTypeNetwork 网络错误
	ErrorTypeNetwork ErrorType = "network"
	// ErrorTypeTimeout 超时错误
	ErrorTypeTimeout ErrorType = "timeout"
	// ErrorTypeAuthentication 认证错误
	ErrorTypeAuthentication ErrorType = "authentication"
	// ErrorTypeDataLoss 数据丢失
	ErrorTypeDataLoss ErrorType = "data_loss"
	// ErrorTypeOther 其他错误
	ErrorTypeOther ErrorType = "other"
)

// MetricsCollector 指标收集器接口
type MetricsCollector interface {
	// RecordOperation 记录一次操作
	RecordOperation(result *Result)

	// RecordConnectionAttempt 记录连接尝试
	RecordConnectionAttempt(success bool, duration time.Duration)

	// RecordError 记录错误
	RecordError(err error, errorType ErrorType)

	// GetMetrics 获取当前聚合的指标
	GetMetrics() *StabilityMetrics

	// Reset 重置指标
	Reset()
}

// StabilityMetrics 稳定性指标
type StabilityMetrics struct {
	// 可用性指标
	TotalOperations      int64   // 总操作数
	SuccessfulOperations int64   // 成功操作数
	FailedOperations     int64   // 失败操作数
	Availability         float64 // 可用性 (成功率)

	// 连接指标
	TotalConnectionAttempts      int64   // 总连接尝试数
	SuccessfulConnectionAttempts int64   // 成功连接数
	ConnectionSuccessRate        float64 // 连接成功率

	// 性能指标
	P50Latency time.Duration // P50延迟
	P95Latency time.Duration // P95延迟
	P99Latency time.Duration // P99延迟
	AvgLatency time.Duration // 平均延迟
	MaxLatency time.Duration // 最大延迟
	MinLatency time.Duration // 最小延迟
	Throughput float64       // 吞吐量 (ops/s)

	// 可靠性指标
	ErrorRate       float64 // 错误率
	DataLossRate    float64 // 数据丢失率
	DataConsistency float64 // 数据一致性
	DuplicateRate   float64 // 重复率

	// 恢复性指标
	MTBF                   time.Duration // 平均故障间隔时间
	MTTR                   time.Duration // 平均恢复时间
	TotalReconnectAttempts int64         // 重连尝试次数
	SuccessfulReconnects   int64         // 成功重连次数
	ReconnectSuccessRate   float64       // 重连成功率

	// 错误统计
	ErrorsByType map[ErrorType]int64 // 按类型分类的错误数

	// 时间相关
	StartTime time.Time     // 测试开始时间
	EndTime   time.Time     // 测试结束时间
	Duration  time.Duration // 测试持续时间

	// 中间件特定指标（可选）
	// Redis
	CacheHitRate        float64 // 缓存命中率
	MemoryUsage         float64 // 内存使用率
	KeyspaceUtilization float64 // 键空间利用率

	// Kafka
	MessageLag        int64         // 消息积压
	ConsumerLag       time.Duration // 消费延迟
	DuplicateMessages int64         // 重复消息数
	RebalanceCount    int64         // 重平衡次数
}

// Clone 克隆指标（用于并发安全读取）
func (sm *StabilityMetrics) Clone() *StabilityMetrics {
	clone := *sm
	if sm.ErrorsByType != nil {
		clone.ErrorsByType = make(map[ErrorType]int64)
		for k, v := range sm.ErrorsByType {
			clone.ErrorsByType[k] = v
		}
	}
	return &clone
}
