package collector

import (
	"sort"
	"sync"
	"time"

	"middleware-chaos-testing/internal/core"
)

// MetricsCollector 指标收集器实现
type MetricsCollector struct {
	mu sync.RWMutex

	// 操作统计
	totalOps      int64
	successOps    int64
	failedOps     int64
	latencies     []time.Duration
	errors        map[core.ErrorType]int64

	// 连接统计
	totalConnAttempts      int64
	successfulConnAttempts int64

	// 重连统计
	totalReconnectAttempts int64
	successfulReconnects   int64

	// 故障恢复统计
	failureStartTimes []time.Time
	recoveryTimes     []time.Duration

	// 时间统计
	startTime time.Time
	endTime   time.Time
}

// NewMetricsCollector 创建新的指标收集器
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		latencies: make([]time.Duration, 0, 10000),
		errors:    make(map[core.ErrorType]int64),
		startTime: time.Now(),
	}
}

// RecordOperation 记录一次操作
func (mc *MetricsCollector) RecordOperation(result *core.Result) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.totalOps++
	mc.latencies = append(mc.latencies, result.Duration)

	if result.Success {
		mc.successOps++
	} else {
		mc.failedOps++
	}
}

// RecordConnectionAttempt 记录连接尝试
func (mc *MetricsCollector) RecordConnectionAttempt(success bool, duration time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.totalConnAttempts++
	if success {
		mc.successfulConnAttempts++
	}
}

// RecordError 记录错误
func (mc *MetricsCollector) RecordError(err error, errorType core.ErrorType) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.errors[errorType]++
}

// GetMetrics 获取当前聚合的指标
func (mc *MetricsCollector) GetMetrics() *core.StabilityMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	mc.endTime = time.Now()

	metrics := &core.StabilityMetrics{
		TotalOperations:      mc.totalOps,
		SuccessfulOperations: mc.successOps,
		FailedOperations:     mc.failedOps,

		TotalConnectionAttempts:      mc.totalConnAttempts,
		SuccessfulConnectionAttempts: mc.successfulConnAttempts,

		TotalReconnectAttempts: mc.totalReconnectAttempts,
		SuccessfulReconnects:   mc.successfulReconnects,

		ErrorsByType: make(map[core.ErrorType]int64),

		StartTime: mc.startTime,
		EndTime:   mc.endTime,
		Duration:  mc.endTime.Sub(mc.startTime),
	}

	// 复制错误统计
	for k, v := range mc.errors {
		metrics.ErrorsByType[k] = v
	}

	// 计算可用性
	if mc.totalOps > 0 {
		metrics.Availability = float64(mc.successOps) / float64(mc.totalOps)
		metrics.ErrorRate = float64(mc.failedOps) / float64(mc.totalOps)
	}

	// 计算连接成功率
	if mc.totalConnAttempts > 0 {
		metrics.ConnectionSuccessRate = float64(mc.successfulConnAttempts) / float64(mc.totalConnAttempts)
	}

	// 计算重连成功率
	if mc.totalReconnectAttempts > 0 {
		metrics.ReconnectSuccessRate = float64(mc.successfulReconnects) / float64(mc.totalReconnectAttempts)
	} else {
		// 如果没有重连尝试，假设重连成功率为100%
		metrics.ReconnectSuccessRate = 1.0
	}

	// 计算延迟指标
	if len(mc.latencies) > 0 {
		sortedLatencies := make([]time.Duration, len(mc.latencies))
		copy(sortedLatencies, mc.latencies)
		sort.Slice(sortedLatencies, func(i, j int) bool {
			return sortedLatencies[i] < sortedLatencies[j]
		})

		metrics.P50Latency = sortedLatencies[len(sortedLatencies)*50/100]
		metrics.P95Latency = sortedLatencies[len(sortedLatencies)*95/100]
		metrics.P99Latency = sortedLatencies[len(sortedLatencies)*99/100]
		metrics.MinLatency = sortedLatencies[0]
		metrics.MaxLatency = sortedLatencies[len(sortedLatencies)-1]

		// 计算平均延迟
		var totalLatency time.Duration
		for _, lat := range mc.latencies {
			totalLatency += lat
		}
		metrics.AvgLatency = totalLatency / time.Duration(len(mc.latencies))
	}

	// 计算吞吐量
	if metrics.Duration > 0 {
		metrics.Throughput = float64(mc.totalOps) / metrics.Duration.Seconds()
	}

	// 计算MTTR（如果有恢复数据）
	if len(mc.recoveryTimes) > 0 {
		var totalRecovery time.Duration
		for _, rt := range mc.recoveryTimes {
			totalRecovery += rt
		}
		metrics.MTTR = totalRecovery / time.Duration(len(mc.recoveryTimes))
	}

	return metrics
}

// Reset 重置指标
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.totalOps = 0
	mc.successOps = 0
	mc.failedOps = 0
	mc.latencies = make([]time.Duration, 0, 10000)
	mc.errors = make(map[core.ErrorType]int64)
	mc.totalConnAttempts = 0
	mc.successfulConnAttempts = 0
	mc.totalReconnectAttempts = 0
	mc.successfulReconnects = 0
	mc.failureStartTimes = nil
	mc.recoveryTimes = nil
	mc.startTime = time.Now()
	mc.endTime = time.Time{}
}
