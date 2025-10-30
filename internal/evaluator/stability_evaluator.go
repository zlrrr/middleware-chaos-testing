package evaluator

import (
	"middleware-chaos-testing/internal/core"
)

// StabilityEvaluator 稳定性评估器（占位符实现）
type StabilityEvaluator struct {
	thresholds *core.Thresholds
}

// NewStabilityEvaluator 创建新的稳定性评估器
func NewStabilityEvaluator(thresholds *core.Thresholds) *StabilityEvaluator {
	// 占位符实现
	return &StabilityEvaluator{
		thresholds: thresholds,
	}
}

// Evaluate 评估稳定性指标（占位符）
func (se *StabilityEvaluator) Evaluate(metrics *core.StabilityMetrics) *core.EvaluationResult {
	// 占位符：返回失败结果
	return &core.EvaluationResult{
		Score:  0,
		Grade:  core.GradeFailed,
		Status: core.StatusFail,
	}
}

// EvaluateRedis Redis特定评估（占位符）
func (se *StabilityEvaluator) EvaluateRedis(metrics *core.StabilityMetrics) *core.EvaluationResult {
	return se.Evaluate(metrics)
}

// EvaluateKafka Kafka特定评估（占位符）
func (se *StabilityEvaluator) EvaluateKafka(metrics *core.StabilityMetrics) *core.EvaluationResult {
	return se.Evaluate(metrics)
}

// SetThresholds 设置自定义阈值（占位符）
func (se *StabilityEvaluator) SetThresholds(thresholds *core.Thresholds) {
	se.thresholds = thresholds
}

// GetDefaultThresholds 获取默认阈值（占位符）
func (se *StabilityEvaluator) GetDefaultThresholds() *core.Thresholds {
	// 返回空阈值
	return &core.Thresholds{}
}
