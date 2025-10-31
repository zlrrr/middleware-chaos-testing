package reporter

import (
	"fmt"
	"io"
	"strings"
	"time"

	"middleware-chaos-testing/internal/core"
)

// MarkdownReporterImpl Markdown报告生成器
type MarkdownReporterImpl struct {
	template string
}

// NewMarkdownReporter 创建新的Markdown报告生成器
func NewMarkdownReporter() *MarkdownReporterImpl {
	return &MarkdownReporterImpl{}
}

// SetTemplate 设置自定义模板
func (r *MarkdownReporterImpl) SetTemplate(template string) {
	r.template = template
}

// GenerateReport 生成Markdown报告
func (r *MarkdownReporterImpl) GenerateReport(
	metrics *core.StabilityMetrics,
	evaluation *core.EvaluationResult,
	output io.Writer,
) error {
	var sb strings.Builder

	// 标题
	sb.WriteString("# 中间件稳定性测试报告\n\n")

	// 测试信息
	sb.WriteString("## 测试信息\n\n")
	sb.WriteString(fmt.Sprintf("- **测试时长**: %v\n", metrics.Duration.Round(time.Second)))
	sb.WriteString(fmt.Sprintf("- **测试完成**: %s\n\n",
		evaluation.EvaluatedAt.Format("2006-01-02 15:04:05")))

	// 总体评分
	sb.WriteString("## 总体评分\n\n")
	statusSymbol := r.getStatusSymbol(evaluation.Status)
	sb.WriteString(fmt.Sprintf("**%.1f/100** (%s) %s\n\n", evaluation.Score, evaluation.Grade, statusSymbol))

	// 各维度得分
	sb.WriteString("### 各维度得分\n\n")
	sb.WriteString("| 维度 | 得分 | 百分比 | 权重 |\n")
	sb.WriteString("|------|------|--------|------|\n")
	sb.WriteString(fmt.Sprintf("| 可用性 | %.1f/30 | %.1f%% | 30%% |\n",
		evaluation.Scores.Availability,
		evaluation.Scores.Availability/30*100))
	sb.WriteString(fmt.Sprintf("| 性能 | %.1f/25 | %.1f%% | 25%% |\n",
		evaluation.Scores.Performance,
		evaluation.Scores.Performance/25*100))
	sb.WriteString(fmt.Sprintf("| 可靠性 | %.1f/25 | %.1f%% | 25%% |\n",
		evaluation.Scores.Reliability,
		evaluation.Scores.Reliability/25*100))
	sb.WriteString(fmt.Sprintf("| 恢复力 | %.1f/20 | %.1f%% | 20%% |\n\n",
		evaluation.Scores.Resilience,
		evaluation.Scores.Resilience/20*100))

	// 核心指标
	sb.WriteString("## 核心指标\n\n")

	// 可用性
	sb.WriteString("### 可用性\n\n")
	sb.WriteString(fmt.Sprintf("- **可用性率**: %.2f%%\n", metrics.Availability*100))
	sb.WriteString(fmt.Sprintf("- **总操作数**: %d\n", metrics.TotalOperations))
	sb.WriteString(fmt.Sprintf("- **成功操作**: %d\n", metrics.SuccessfulOperations))
	sb.WriteString(fmt.Sprintf("- **失败操作**: %d\n", metrics.FailedOperations))
	sb.WriteString(fmt.Sprintf("- **错误率**: %.2f%%\n\n", metrics.ErrorRate*100))

	// 性能指标
	sb.WriteString("### 性能指标\n\n")
	sb.WriteString(fmt.Sprintf("- **P50 延迟**: %v\n", metrics.P50Latency.Round(time.Millisecond)))
	sb.WriteString(fmt.Sprintf("- **P95 延迟**: %v\n", metrics.P95Latency.Round(time.Millisecond)))
	sb.WriteString(fmt.Sprintf("- **P99 延迟**: %v\n", metrics.P99Latency.Round(time.Millisecond)))
	sb.WriteString(fmt.Sprintf("- **平均吞吐**: %.0f ops/s\n\n", metrics.Throughput))

	// 可靠性
	sb.WriteString("### 可靠性\n\n")
	sb.WriteString(fmt.Sprintf("- **数据丢失率**: %.4f%%\n\n", metrics.DataLossRate*100))

	// 恢复性
	if metrics.MTTR > 0 || metrics.ReconnectSuccessRate > 0 {
		sb.WriteString("### 恢复性\n\n")
		if metrics.MTTR > 0 {
			sb.WriteString(fmt.Sprintf("- **MTTR**: %v\n", metrics.MTTR.Round(time.Second)))
		}
		if metrics.ReconnectSuccessRate > 0 {
			sb.WriteString(fmt.Sprintf("- **重连成功率**: %.0f%%\n", metrics.ReconnectSuccessRate*100))
		}
		sb.WriteString("\n")
	}

	// 发现的问题
	if len(evaluation.Issues) > 0 {
		sb.WriteString(fmt.Sprintf("## 发现的问题 (%d个)\n\n", len(evaluation.Issues)))
		for _, issue := range evaluation.Issues {
			sb.WriteString(fmt.Sprintf("### [%s] %s\n\n", issue.Severity, issue.Type))
			sb.WriteString(fmt.Sprintf("- **指标**: %s\n", issue.Metric))
			sb.WriteString(fmt.Sprintf("- **当前值**: %.2f\n", issue.Current))
			sb.WriteString(fmt.Sprintf("- **期望值**: %.2f\n", issue.Expected))
			sb.WriteString(fmt.Sprintf("- **说明**: %s\n\n", issue.Message))
		}
	}

	// 改进建议
	if len(evaluation.Recommendations) > 0 {
		sb.WriteString("## 改进建议\n\n")
		for _, rec := range evaluation.Recommendations {
			sb.WriteString(fmt.Sprintf("### [%s] %s\n\n", rec.Priority, rec.Title))
			sb.WriteString(fmt.Sprintf("**分类**: %s\n\n", rec.Category))
			if rec.Message != "" {
				sb.WriteString(fmt.Sprintf("%s\n\n", rec.Message))
			}
			if len(rec.Actions) > 0 {
				sb.WriteString("**具体行动**:\n\n")
				for _, action := range rec.Actions {
					sb.WriteString(fmt.Sprintf("- %s\n", action))
				}
				sb.WriteString("\n")
			}
			if len(rec.References) > 0 {
				sb.WriteString("**参考文档**:\n\n")
				for _, ref := range rec.References {
					sb.WriteString(fmt.Sprintf("- %s\n", ref))
				}
				sb.WriteString("\n")
			}
		}
	}

	// 结论
	sb.WriteString("## 结论\n\n")
	sb.WriteString(evaluation.Rationale)
	sb.WriteString("\n")

	_, err := output.Write([]byte(sb.String()))
	return err
}

func (r *MarkdownReporterImpl) getStatusSymbol(status core.TestStatus) string {
	switch status {
	case core.StatusPass:
		return "✅ PASS"
	case core.StatusWarning:
		return "⚠️ WARNING"
	case core.StatusFail:
		return "❌ FAIL"
	default:
		return ""
	}
}
