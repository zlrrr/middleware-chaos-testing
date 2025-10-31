package reporter

import (
	"encoding/json"
	"io"

	"middleware-chaos-testing/internal/core"
)

// JSONReporterImpl JSON报告生成器
type JSONReporterImpl struct {
	indent string
}

// NewJSONReporter 创建新的JSON报告生成器
func NewJSONReporter() *JSONReporterImpl {
	return &JSONReporterImpl{
		indent: "  ",
	}
}

// SetIndent 设置JSON缩进
func (r *JSONReporterImpl) SetIndent(indent string) {
	r.indent = indent
}

// GenerateReport 生成JSON报告
func (r *JSONReporterImpl) GenerateReport(
	metrics *core.StabilityMetrics,
	evaluation *core.EvaluationResult,
	output io.Writer,
) error {
	report := map[string]interface{}{
		"test_info": map[string]interface{}{
			"duration":     metrics.Duration.String(),
			"completed_at": evaluation.EvaluatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
		"evaluation": map[string]interface{}{
			"score":  evaluation.Score,
			"grade":  evaluation.Grade,
			"status": evaluation.Status,
			"scores": map[string]interface{}{
				"availability": evaluation.Scores.Availability,
				"performance":  evaluation.Scores.Performance,
				"reliability":  evaluation.Scores.Reliability,
				"resilience":   evaluation.Scores.Resilience,
			},
			"rationale": evaluation.Rationale,
		},
		"metrics": map[string]interface{}{
			"availability": map[string]interface{}{
				"rate":                metrics.Availability,
				"total_operations":    metrics.TotalOperations,
				"successful_operations": metrics.SuccessfulOperations,
				"failed_operations":   metrics.FailedOperations,
				"error_rate":          metrics.ErrorRate,
			},
			"performance": map[string]interface{}{
				"p50_latency_ms": metrics.P50Latency.Milliseconds(),
				"p95_latency_ms": metrics.P95Latency.Milliseconds(),
				"p99_latency_ms": metrics.P99Latency.Milliseconds(),
				"avg_latency_ms": metrics.AvgLatency.Milliseconds(),
				"throughput":     metrics.Throughput,
			},
			"reliability": map[string]interface{}{
				"data_loss_rate": metrics.DataLossRate,
			},
		},
		"issues":          evaluation.Issues,
		"recommendations": evaluation.Recommendations,
	}

	// 添加恢复性指标（如果有）
	if metrics.MTTR > 0 || metrics.ReconnectSuccessRate > 0 {
		report["metrics"].(map[string]interface{})["resilience"] = map[string]interface{}{
			"mttr_seconds":          metrics.MTTR.Seconds(),
			"reconnect_success_rate": metrics.ReconnectSuccessRate,
		}
	}

	encoder := json.NewEncoder(output)
	encoder.SetIndent("", r.indent)
	return encoder.Encode(report)
}
