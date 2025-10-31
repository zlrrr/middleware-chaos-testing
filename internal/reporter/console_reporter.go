package reporter

import (
	"fmt"
	"io"
	"strings"
	"time"

	"middleware-chaos-testing/internal/core"
)

// ConsoleReporter 控制台报告生成器
type ConsoleReporterImpl struct {
	colorEnabled bool
}

// NewConsoleReporter 创建新的控制台报告生成器
func NewConsoleReporter() *ConsoleReporterImpl {
	return &ConsoleReporterImpl{
		colorEnabled: true,
	}
}

// SetColorEnabled 设置是否启用颜色
func (r *ConsoleReporterImpl) SetColorEnabled(enabled bool) {
	r.colorEnabled = enabled
}

// GenerateReport 生成控制台报告
func (r *ConsoleReporterImpl) GenerateReport(
	metrics *core.StabilityMetrics,
	evaluation *core.EvaluationResult,
	output io.Writer,
) error {
	var sb strings.Builder

	// 标题
	sb.WriteString("==========================================\n")
	sb.WriteString("   中间件稳定性测试报告\n")
	sb.WriteString("==========================================\n\n")

	// 测试信息
	sb.WriteString(fmt.Sprintf("测试时长: %v\n", metrics.Duration.Round(time.Second)))
	sb.WriteString(fmt.Sprintf("测试完成: %s\n\n", evaluation.EvaluatedAt.Format("2006-01-02 15:04:05")))

	// 总体评分
	sb.WriteString("------------------------------------------\n")
	statusSymbol := r.getStatusSymbol(evaluation.Status)
	sb.WriteString(fmt.Sprintf("  总体评分: %.1f/100 (%s) %s\n",
		evaluation.Score, evaluation.Grade, statusSymbol))
	sb.WriteString("------------------------------------------\n\n")

	// 各维度得分
	sb.WriteString("各维度得分:\n")
	sb.WriteString(fmt.Sprintf("  %s 可用性   %.1f/30  (%.1f%%)  - 权重30%%\n",
		r.getCheckmark(evaluation.Scores.Availability >= 20),
		evaluation.Scores.Availability,
		evaluation.Scores.Availability/30*100))
	sb.WriteString(fmt.Sprintf("  %s 性能     %.1f/25  (%.1f%%)  - 权重25%%\n",
		r.getCheckmark(evaluation.Scores.Performance >= 17),
		evaluation.Scores.Performance,
		evaluation.Scores.Performance/25*100))
	sb.WriteString(fmt.Sprintf("  %s 可靠性   %.1f/25  (%.1f%%)  - 权重25%%\n",
		r.getCheckmark(evaluation.Scores.Reliability >= 17),
		evaluation.Scores.Reliability,
		evaluation.Scores.Reliability/25*100))
	sb.WriteString(fmt.Sprintf("  %s 恢复力   %.1f/20  (%.1f%%)  - 权重20%%\n\n",
		r.getCheckmark(evaluation.Scores.Resilience >= 14),
		evaluation.Scores.Resilience,
		evaluation.Scores.Resilience/20*100))

	// 核心指标
	sb.WriteString("------------------------------------------\n")
	sb.WriteString("  核心指标\n")
	sb.WriteString("------------------------------------------\n")

	// 可用性
	sb.WriteString(fmt.Sprintf("可用性: %.2f%% %s\n",
		metrics.Availability*100,
		r.getCheckmark(metrics.Availability >= 0.95)))
	sb.WriteString(fmt.Sprintf("  - 总操作数: %d\n", metrics.TotalOperations))
	sb.WriteString(fmt.Sprintf("  - 成功操作: %d\n", metrics.SuccessfulOperations))
	sb.WriteString(fmt.Sprintf("  - 失败操作: %d\n", metrics.FailedOperations))
	sb.WriteString(fmt.Sprintf("  - 错误率: %.2f%%\n\n", metrics.ErrorRate*100))

	// 性能指标
	sb.WriteString("性能指标:\n")
	sb.WriteString(fmt.Sprintf("  - P50 延迟: %v %s\n",
		metrics.P50Latency.Round(time.Millisecond),
		r.getCheckmark(metrics.P50Latency <= 50*time.Millisecond)))
	sb.WriteString(fmt.Sprintf("  - P95 延迟: %v %s\n",
		metrics.P95Latency.Round(time.Millisecond),
		r.getCheckmark(metrics.P95Latency <= 200*time.Millisecond)))
	sb.WriteString(fmt.Sprintf("  - P99 延迟: %v %s\n",
		metrics.P99Latency.Round(time.Millisecond),
		r.getCheckmark(metrics.P99Latency <= 500*time.Millisecond)))
	sb.WriteString(fmt.Sprintf("  - 平均吞吐: %.0f ops/s\n\n", metrics.Throughput))

	// 可靠性
	sb.WriteString("可靠性:\n")
	sb.WriteString(fmt.Sprintf("  - 数据丢失率: %.4f%% %s\n",
		metrics.DataLossRate*100,
		r.getCheckmark(metrics.DataLossRate == 0)))

	// 恢复性
	if metrics.MTTR > 0 {
		sb.WriteString("\n恢复性:\n")
		sb.WriteString(fmt.Sprintf("  - MTTR: %v %s\n",
			metrics.MTTR.Round(time.Second),
			r.getCheckmark(metrics.MTTR <= 300*time.Second)))
	}
	if metrics.ReconnectSuccessRate > 0 {
		sb.WriteString(fmt.Sprintf("  - 重连成功率: %.0f%% %s\n",
			metrics.ReconnectSuccessRate*100,
			r.getCheckmark(metrics.ReconnectSuccessRate >= 0.95)))
	}

	// 发现的问题
	if len(evaluation.Issues) > 0 {
		sb.WriteString("\n------------------------------------------\n")
		sb.WriteString(fmt.Sprintf("  发现的问题 (%d个)\n", len(evaluation.Issues)))
		sb.WriteString("------------------------------------------\n")
		for _, issue := range evaluation.Issues {
			sb.WriteString(fmt.Sprintf("[%s] %s\n", issue.Severity, issue.Type))
			sb.WriteString(fmt.Sprintf("  指标: %s\n", issue.Metric))
			sb.WriteString(fmt.Sprintf("  当前值: %.2f\n", issue.Current))
			sb.WriteString(fmt.Sprintf("  期望值: %.2f\n", issue.Expected))
			sb.WriteString(fmt.Sprintf("  说明: %s\n\n", issue.Message))
		}
	}

	// 改进建议
	if len(evaluation.Recommendations) > 0 {
		sb.WriteString("------------------------------------------\n")
		sb.WriteString("  改进建议 (按优先级排序)\n")
		sb.WriteString("------------------------------------------\n\n")
		for i, rec := range evaluation.Recommendations {
			sb.WriteString(fmt.Sprintf("[%s] %s\n", rec.Priority, rec.Title))
			sb.WriteString(fmt.Sprintf("分类: %s\n", rec.Category))
			if rec.Message != "" {
				sb.WriteString(fmt.Sprintf("说明: %s\n", rec.Message))
			}
			if len(rec.Actions) > 0 {
				sb.WriteString("具体行动:\n")
				for j, action := range rec.Actions {
					sb.WriteString(fmt.Sprintf("  %d. %s\n", j+1, action))
				}
			}
			if len(rec.References) > 0 {
				sb.WriteString("\n参考文档:\n")
				for _, ref := range rec.References {
					sb.WriteString(fmt.Sprintf("  - %s\n", ref))
				}
			}
			if i < len(evaluation.Recommendations)-1 {
				sb.WriteString("\n")
			}
		}
		sb.WriteString("\n")
	}

	// 结论
	sb.WriteString("------------------------------------------\n")
	sb.WriteString("  结论\n")
	sb.WriteString("------------------------------------------\n")
	sb.WriteString(evaluation.Rationale)
	sb.WriteString("\n==========================================\n")

	_, err := output.Write([]byte(sb.String()))
	return err
}

func (r *ConsoleReporterImpl) getStatusSymbol(status core.TestStatus) string {
	switch status {
	case core.StatusPass:
		return "✅ PASS"
	case core.StatusWarning:
		return "⚠️  WARNING"
	case core.StatusFail:
		return "❌ FAIL"
	default:
		return ""
	}
}

func (r *ConsoleReporterImpl) getCheckmark(pass bool) string {
	if pass {
		return "✓"
	}
	return "⚠️"
}
