package core

import "time"

// StabilityGrade 稳定性等级
type StabilityGrade string

const (
	// GradeExcellent 优秀 (90-100分)
	GradeExcellent StabilityGrade = "EXCELLENT"
	// GradeGood 良好 (80-89分)
	GradeGood StabilityGrade = "GOOD"
	// GradeFair 一般 (70-79分)
	GradeFair StabilityGrade = "FAIR"
	// GradePoor 较差 (60-69分)
	GradePoor StabilityGrade = "POOR"
	// GradeFailed 失败 (<60分)
	GradeFailed StabilityGrade = "FAILED"
)

// TestStatus 测试状态
type TestStatus string

const (
	// StatusPass 通过
	StatusPass TestStatus = "PASS"
	// StatusWarning 警告
	StatusWarning TestStatus = "WARNING"
	// StatusFail 失败
	StatusFail TestStatus = "FAIL"
)

// Evaluator 稳定性评估器接口
type Evaluator interface {
	// Evaluate 评估稳定性指标
	// 返回评估结果
	Evaluate(metrics *StabilityMetrics) *EvaluationResult

	// EvaluateRedis Redis特定评估
	EvaluateRedis(metrics *StabilityMetrics) *EvaluationResult

	// EvaluateKafka Kafka特定评估
	EvaluateKafka(metrics *StabilityMetrics) *EvaluationResult

	// SetThresholds 设置自定义阈值
	SetThresholds(thresholds *Thresholds)

	// GetDefaultThresholds 获取默认阈值
	GetDefaultThresholds() *Thresholds
}

// EvaluationResult 评估结果
type EvaluationResult struct {
	// 总体评分
	Score  float64        // 0-100分
	Grade  StabilityGrade // 等级
	Status TestStatus     // 状态

	// 各维度得分
	Scores struct {
		Availability float64 // 可用性得分 (30分)
		Performance  float64 // 性能得分 (25分)
		Reliability  float64 // 可靠性得分 (25分)
		Resilience   float64 // 恢复力得分 (20分)
	}

	// 识别的问题
	Issues []Issue

	// 改进建议
	Recommendations []Recommendation

	// 判断依据
	Rationale string

	// 评估时间
	EvaluatedAt time.Time
}

// Issue 问题描述
type Issue struct {
	Type     string  // 问题类型
	Severity string  // 严重程度: CRITICAL, HIGH, MEDIUM, LOW
	Metric   string  // 相关指标
	Current  float64 // 当前值
	Expected float64 // 期望值
	Message  string  // 问题描述
}

// Recommendation 改进建议
type Recommendation struct {
	Priority   string   // 优先级: HIGH, MEDIUM, LOW
	Category   string   // 类别: CONFIGURATION, SCALING, OPTIMIZATION
	Title      string   // 标题
	Message    string   // 描述
	Actions    []string // 具体行动项
	References []string // 参考文档链接
}

// Thresholds 评分阈值
type Thresholds struct {
	// 可用性阈值
	AvailabilityExcellent float64 // >= 99.99%
	AvailabilityGood      float64 // >= 99.9%
	AvailabilityFair      float64 // >= 99.0%
	AvailabilityPass      float64 // >= 95.0%

	// P95延迟阈值
	P95LatencyExcellent time.Duration // <= 10ms
	P95LatencyGood      time.Duration // <= 50ms
	P95LatencyFair      time.Duration // <= 100ms
	P95LatencyPass      time.Duration // <= 200ms

	// P99延迟阈值
	P99LatencyExcellent time.Duration // <= 20ms
	P99LatencyGood      time.Duration // <= 100ms
	P99LatencyFair      time.Duration // <= 200ms
	P99LatencyPass      time.Duration // <= 500ms

	// 错误率阈值
	ErrorRateExcellent float64 // <= 0.01%
	ErrorRateGood      float64 // <= 0.1%
	ErrorRateFair      float64 // <= 0.5%
	ErrorRatePass      float64 // <= 1.0%

	// MTTR阈值
	MTTRExcellent time.Duration // <= 5s
	MTTRGood      time.Duration // <= 30s
	MTTRFair      time.Duration // <= 60s
	MTTRPass      time.Duration // <= 300s
}
