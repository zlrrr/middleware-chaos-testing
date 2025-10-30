package evaluator

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"middleware-chaos-testing/internal/core"
)

// StabilityEvaluator 稳定性评估器
type StabilityEvaluator struct {
	thresholds *core.Thresholds
}

// NewStabilityEvaluator 创建新的稳定性评估器
func NewStabilityEvaluator(thresholds *core.Thresholds) *StabilityEvaluator {
	// 从默认阈值开始
	finalThresholds := DefaultThresholds()

	// 如果提供了自定义阈值，合并非零值
	if thresholds != nil {
		if thresholds.AvailabilityExcellent > 0 {
			finalThresholds.AvailabilityExcellent = thresholds.AvailabilityExcellent
		}
		if thresholds.AvailabilityGood > 0 {
			finalThresholds.AvailabilityGood = thresholds.AvailabilityGood
		}
		if thresholds.AvailabilityFair > 0 {
			finalThresholds.AvailabilityFair = thresholds.AvailabilityFair
		}
		if thresholds.AvailabilityPass > 0 {
			finalThresholds.AvailabilityPass = thresholds.AvailabilityPass
		}

		if thresholds.P95LatencyExcellent > 0 {
			finalThresholds.P95LatencyExcellent = thresholds.P95LatencyExcellent
		}
		if thresholds.P95LatencyGood > 0 {
			finalThresholds.P95LatencyGood = thresholds.P95LatencyGood
		}
		if thresholds.P95LatencyFair > 0 {
			finalThresholds.P95LatencyFair = thresholds.P95LatencyFair
		}
		if thresholds.P95LatencyPass > 0 {
			finalThresholds.P95LatencyPass = thresholds.P95LatencyPass
		}

		if thresholds.P99LatencyExcellent > 0 {
			finalThresholds.P99LatencyExcellent = thresholds.P99LatencyExcellent
		}
		if thresholds.P99LatencyGood > 0 {
			finalThresholds.P99LatencyGood = thresholds.P99LatencyGood
		}
		if thresholds.P99LatencyFair > 0 {
			finalThresholds.P99LatencyFair = thresholds.P99LatencyFair
		}
		if thresholds.P99LatencyPass > 0 {
			finalThresholds.P99LatencyPass = thresholds.P99LatencyPass
		}

		if thresholds.ErrorRateExcellent > 0 {
			finalThresholds.ErrorRateExcellent = thresholds.ErrorRateExcellent
		}
		if thresholds.ErrorRateGood > 0 {
			finalThresholds.ErrorRateGood = thresholds.ErrorRateGood
		}
		if thresholds.ErrorRateFair > 0 {
			finalThresholds.ErrorRateFair = thresholds.ErrorRateFair
		}
		if thresholds.ErrorRatePass > 0 {
			finalThresholds.ErrorRatePass = thresholds.ErrorRatePass
		}

		if thresholds.MTTRExcellent > 0 {
			finalThresholds.MTTRExcellent = thresholds.MTTRExcellent
		}
		if thresholds.MTTRGood > 0 {
			finalThresholds.MTTRGood = thresholds.MTTRGood
		}
		if thresholds.MTTRFair > 0 {
			finalThresholds.MTTRFair = thresholds.MTTRFair
		}
		if thresholds.MTTRPass > 0 {
			finalThresholds.MTTRPass = thresholds.MTTRPass
		}
	}

	return &StabilityEvaluator{
		thresholds: finalThresholds,
	}
}

// DefaultThresholds 返回默认阈值
func DefaultThresholds() *core.Thresholds {
	return &core.Thresholds{
		AvailabilityExcellent: 0.9999,
		AvailabilityGood:      0.999,
		AvailabilityFair:      0.99,
		AvailabilityPass:      0.95,

		P95LatencyExcellent: 10 * time.Millisecond,
		P95LatencyGood:      50 * time.Millisecond,
		P95LatencyFair:      100 * time.Millisecond,
		P95LatencyPass:      200 * time.Millisecond,

		P99LatencyExcellent: 20 * time.Millisecond,
		P99LatencyGood:      100 * time.Millisecond,
		P99LatencyFair:      200 * time.Millisecond,
		P99LatencyPass:      500 * time.Millisecond,

		ErrorRateExcellent: 0.0001,
		ErrorRateGood:      0.001,
		ErrorRateFair:      0.005,
		ErrorRatePass:      0.01,

		MTTRExcellent: 5 * time.Second,
		MTTRGood:      30 * time.Second,
		MTTRFair:      60 * time.Second,
		MTTRPass:      300 * time.Second,
	}
}

// Evaluate 评估稳定性指标
func (se *StabilityEvaluator) Evaluate(metrics *core.StabilityMetrics) *core.EvaluationResult {
	result := &core.EvaluationResult{
		EvaluatedAt: time.Now(),
		Issues:      make([]core.Issue, 0),
		Recommendations: make([]core.Recommendation, 0),
	}

	// 计算各维度得分
	result.Scores.Availability = se.calculateAvailabilityScore(metrics, result)
	result.Scores.Performance = se.calculatePerformanceScore(metrics, result)
	result.Scores.Reliability = se.calculateReliabilityScore(metrics, result)
	result.Scores.Resilience = se.calculateResilienceScore(metrics, result)

	// 计算总分（直接相加，各维度满分分别是30/25/25/20）
	result.Score = result.Scores.Availability +
		result.Scores.Performance +
		result.Scores.Reliability +
		result.Scores.Resilience

	// 确定等级和状态
	result.Grade = se.determineGrade(result.Score)
	result.Status = se.determineStatus(result)

	// 生成建议和判断依据
	result.Recommendations = se.generateRecommendations(result)
	result.Rationale = se.generateRationale(result)

	return result
}

// calculateAvailabilityScore 计算可用性得分 (满分30分)
func (se *StabilityEvaluator) calculateAvailabilityScore(
	metrics *core.StabilityMetrics,
	result *core.EvaluationResult,
) float64 {
	availability := metrics.Availability

	var score float64
	switch {
	case availability >= se.thresholds.AvailabilityExcellent:
		score = 30.0
	case availability >= se.thresholds.AvailabilityGood:
		score = 27.0
	case availability >= se.thresholds.AvailabilityFair:
		score = 24.0
	case availability >= se.thresholds.AvailabilityPass:
		score = 20.0
	default:
		score = availability * 100 * 0.2
		result.Issues = append(result.Issues, core.Issue{
			Type:     "low_availability",
			Severity: "CRITICAL",
			Metric:   "availability",
			Current:  availability * 100,
			Expected: se.thresholds.AvailabilityPass * 100,
			Message: fmt.Sprintf("可用性%.2f%%低于最低要求%.2f%%",
				availability*100,
				se.thresholds.AvailabilityPass*100),
		})
	}

	return score
}

// calculatePerformanceScore 计算性能得分 (满分25分)
func (se *StabilityEvaluator) calculatePerformanceScore(
	metrics *core.StabilityMetrics,
	result *core.EvaluationResult,
) float64 {
	p95 := metrics.P95Latency
	p99 := metrics.P99Latency

	// P95得分 (15分)
	var p95Score float64
	switch {
	case p95 <= se.thresholds.P95LatencyExcellent:
		p95Score = 15.0
	case p95 <= se.thresholds.P95LatencyGood:
		p95Score = 13.5
	case p95 <= se.thresholds.P95LatencyFair:
		p95Score = 12.0
	case p95 <= se.thresholds.P95LatencyPass:
		p95Score = 10.0
	default:
		p95Score = 8.0
		result.Issues = append(result.Issues, core.Issue{
			Type:     "high_p95_latency",
			Severity: "HIGH",
			Metric:   "p95_latency",
			Current:  float64(p95.Milliseconds()),
			Expected: float64(se.thresholds.P95LatencyPass.Milliseconds()),
			Message:  fmt.Sprintf("P95延迟%v超过阈值%v", p95, se.thresholds.P95LatencyPass),
		})
	}

	// P99得分 (10分)
	var p99Score float64
	switch {
	case p99 <= se.thresholds.P99LatencyExcellent:
		p99Score = 10.0
	case p99 <= se.thresholds.P99LatencyGood:
		p99Score = 9.0
	case p99 <= se.thresholds.P99LatencyFair:
		p99Score = 8.0
	case p99 <= se.thresholds.P99LatencyPass:
		p99Score = 6.5
	default:
		p99Score = 5.0
		result.Issues = append(result.Issues, core.Issue{
			Type:     "high_p99_latency",
			Severity: "MEDIUM",
			Metric:   "p99_latency",
			Current:  float64(p99.Milliseconds()),
			Expected: float64(se.thresholds.P99LatencyPass.Milliseconds()),
			Message:  fmt.Sprintf("P99延迟%v超过阈值%v", p99, se.thresholds.P99LatencyPass),
		})
	}

	return p95Score + p99Score
}

// calculateReliabilityScore 计算可靠性得分 (满分25分)
func (se *StabilityEvaluator) calculateReliabilityScore(
	metrics *core.StabilityMetrics,
	result *core.EvaluationResult,
) float64 {
	errorRate := metrics.ErrorRate
	dataLossRate := metrics.DataLossRate

	// 错误率得分 (15分)
	var errorScore float64
	switch {
	case errorRate <= se.thresholds.ErrorRateExcellent:
		errorScore = 15.0
	case errorRate <= se.thresholds.ErrorRateGood:
		errorScore = 13.5
	case errorRate <= se.thresholds.ErrorRateFair:
		errorScore = 12.0
	case errorRate <= se.thresholds.ErrorRatePass:
		errorScore = 10.0
	default:
		errorScore = 7.0
		result.Issues = append(result.Issues, core.Issue{
			Type:     "high_error_rate",
			Severity: "HIGH",
			Metric:   "error_rate",
			Current:  errorRate * 100,
			Expected: se.thresholds.ErrorRatePass * 100,
			Message: fmt.Sprintf("错误率%.4f%%超过阈值%.2f%%",
				errorRate*100,
				se.thresholds.ErrorRatePass*100),
		})
	}

	// 数据丢失率得分 (10分)
	var lossScore float64
	switch {
	case dataLossRate == 0:
		lossScore = 10.0
	case dataLossRate < 0.0001:
		lossScore = 8.0
	case dataLossRate < 0.001:
		lossScore = 6.0
	default:
		lossScore = 3.0
		result.Issues = append(result.Issues, core.Issue{
			Type:     "data_loss_detected",
			Severity: "CRITICAL",
			Metric:   "data_loss_rate",
			Current:  dataLossRate * 100,
			Expected: 0,
			Message:  fmt.Sprintf("检测到数据丢失，丢失率%.4f%%", dataLossRate*100),
		})
	}

	return errorScore + lossScore
}

// calculateResilienceScore 计算恢复力得分 (满分20分)
func (se *StabilityEvaluator) calculateResilienceScore(
	metrics *core.StabilityMetrics,
	result *core.EvaluationResult,
) float64 {
	mttr := metrics.MTTR
	reconnectRate := metrics.ReconnectSuccessRate

	// 恢复时间得分 (12分)
	var mttrScore float64
	switch {
	case mttr <= se.thresholds.MTTRExcellent:
		mttrScore = 12.0
	case mttr <= se.thresholds.MTTRGood:
		mttrScore = 10.5
	case mttr <= se.thresholds.MTTRFair:
		mttrScore = 9.0
	case mttr <= se.thresholds.MTTRPass:
		mttrScore = 7.0
	default:
		mttrScore = 5.0
		result.Issues = append(result.Issues, core.Issue{
			Type:     "slow_recovery",
			Severity: "MEDIUM",
			Metric:   "mttr",
			Current:  float64(mttr.Seconds()),
			Expected: float64(se.thresholds.MTTRPass.Seconds()),
			Message:  fmt.Sprintf("平均恢复时间%v超过阈值%v", mttr, se.thresholds.MTTRPass),
		})
	}

	// 重连成功率得分 (8分)
	var reconnectScore float64
	switch {
	case reconnectRate >= 0.99:
		reconnectScore = 8.0
	case reconnectRate >= 0.95:
		reconnectScore = 7.0
	case reconnectRate >= 0.90:
		reconnectScore = 6.0
	default:
		reconnectScore = 4.0
		result.Issues = append(result.Issues, core.Issue{
			Type:     "low_reconnect_rate",
			Severity: "MEDIUM",
			Metric:   "reconnect_success_rate",
			Current:  reconnectRate * 100,
			Expected: 95.0,
			Message:  fmt.Sprintf("重连成功率%.2f%%低于预期", reconnectRate*100),
		})
	}

	return mttrScore + reconnectScore
}

// determineGrade 确定等级
func (se *StabilityEvaluator) determineGrade(score float64) core.StabilityGrade {
	switch {
	case score >= 90:
		return core.GradeExcellent
	case score >= 80:
		return core.GradeGood
	case score >= 70:
		return core.GradeFair
	case score >= 60:
		return core.GradePoor
	default:
		return core.GradeFailed
	}
}

// determineStatus 确定状态
func (se *StabilityEvaluator) determineStatus(result *core.EvaluationResult) core.TestStatus {
	// CRITICAL问题直接失败
	for _, issue := range result.Issues {
		if issue.Severity == "CRITICAL" {
			return core.StatusFail
		}
	}

	// 分数低于70失败
	if result.Score < 70 {
		return core.StatusFail
	}

	// HIGH问题为警告
	for _, issue := range result.Issues {
		if issue.Severity == "HIGH" {
			return core.StatusWarning
		}
	}

	// 分数低于85为警告
	if result.Score < 85 {
		return core.StatusWarning
	}

	return core.StatusPass
}

// generateRecommendations 生成建议
func (se *StabilityEvaluator) generateRecommendations(result *core.EvaluationResult) []core.Recommendation {
	recommendations := make([]core.Recommendation, 0)

	for _, issue := range result.Issues {
		switch issue.Type {
		case "low_availability":
			recommendations = append(recommendations, core.Recommendation{
				Priority: "HIGH",
				Category: "SCALING",
				Title:    "提高系统可用性",
				Message:  "当前可用性不满足生产环境要求",
				Actions: []string{
					"检查服务健康状态，排查频繁失败原因",
					"增加实例数量，实现高可用部署",
					"配置健康检查和自动重启机制",
					"实施熔断和降级策略",
				},
				References: []string{
					"https://redis.io/topics/sentinel",
					"https://kafka.apache.org/documentation/#replication",
				},
			})

		case "high_p95_latency", "high_p99_latency":
			recommendations = append(recommendations, core.Recommendation{
				Priority: "MEDIUM",
				Category: "OPTIMIZATION",
				Title:    "优化响应延迟",
				Message:  "延迟指标超出可接受范围",
				Actions: []string{
					"分析慢查询日志，优化热点操作",
					"检查网络延迟和带宽瓶颈",
					"优化数据结构和查询模式",
					"考虑增加缓存层或读写分离",
					"评估硬件资源是否充足",
				},
			})

		case "high_error_rate":
			recommendations = append(recommendations, core.Recommendation{
				Priority: "HIGH",
				Category: "CONFIGURATION",
				Title:    "降低错误率",
				Message:  "错误率过高可能导致业务中断",
				Actions: []string{
					"查看错误日志，分析错误类型",
					"检查客户端配置（超时、重试）",
					"验证服务端配置",
					"实施错误处理和重试逻辑",
				},
			})

		case "data_loss_detected":
			recommendations = append(recommendations, core.Recommendation{
				Priority: "HIGH",
				Category: "CONFIGURATION",
				Title:    "防止数据丢失",
				Message:  "检测到数据丢失，需立即处理",
				Actions: []string{
					"检查持久化配置",
					"确保有足够的副本数",
					"配置fsync策略",
					"实施数据校验机制",
				},
			})

		case "slow_recovery":
			recommendations = append(recommendations, core.Recommendation{
				Priority: "MEDIUM",
				Category: "OPTIMIZATION",
				Title:    "加快故障恢复",
				Message:  "平均恢复时间过长，影响系统可用性",
				Actions: []string{
					"优化健康检查频率和超时设置",
					"实施更激进的重试策略",
					"增加备用连接池",
					"优化故障检测算法",
				},
			})

		case "low_reconnect_rate":
			recommendations = append(recommendations, core.Recommendation{
				Priority: "MEDIUM",
				Category: "CONFIGURATION",
				Title:    "提高重连成功率",
				Message:  "重连成功率低，影响系统稳定性",
				Actions: []string{
					"检查网络稳定性",
					"调整重连间隔和最大重试次数",
					"实施指数退避算法",
					"检查服务端连接限制",
				},
			})
		}
	}

	// 按优先级排序
	sort.Slice(recommendations, func(i, j int) bool {
		priority := map[string]int{"HIGH": 3, "MEDIUM": 2, "LOW": 1}
		return priority[recommendations[i].Priority] > priority[recommendations[j].Priority]
	})

	// 去重（相同类型的建议只保留一个）
	seen := make(map[string]bool)
	unique := make([]core.Recommendation, 0)
	for _, rec := range recommendations {
		if !seen[rec.Title] {
			seen[rec.Title] = true
			unique = append(unique, rec)
		}
	}

	return unique
}

// generateRationale 生成判断依据
func (se *StabilityEvaluator) generateRationale(result *core.EvaluationResult) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("综合评分: %.2f/100 (%s)\n\n", result.Score, result.Grade))
	b.WriteString("各维度得分:\n")
	b.WriteString(fmt.Sprintf("- 可用性: %.2f/30 (权重30%%)\n", result.Scores.Availability))
	b.WriteString(fmt.Sprintf("- 性能: %.2f/25 (权重25%%)\n", result.Scores.Performance))
	b.WriteString(fmt.Sprintf("- 可靠性: %.2f/25 (权重25%%)\n", result.Scores.Reliability))
	b.WriteString(fmt.Sprintf("- 恢复力: %.2f/20 (权重20%%)\n\n", result.Scores.Resilience))

	switch result.Status {
	case core.StatusPass:
		b.WriteString("✅ 测试通过: 系统稳定性符合预期，可以用于生产环境。\n")
	case core.StatusWarning:
		b.WriteString("⚠️  警告: 系统存在需要关注的问题，建议优化后再部署。\n")
	case core.StatusFail:
		b.WriteString("❌ 测试失败: 系统稳定性不满足最低要求，不建议用于生产环境。\n")
	}

	if len(result.Issues) > 0 {
		b.WriteString(fmt.Sprintf("\n发现 %d 个问题需要处理。\n", len(result.Issues)))
	}

	return b.String()
}

// EvaluateRedis Redis特定评估
func (se *StabilityEvaluator) EvaluateRedis(metrics *core.StabilityMetrics) *core.EvaluationResult {
	result := se.Evaluate(metrics)

	// 添加Redis特定检查
	if metrics.CacheHitRate > 0 && metrics.CacheHitRate < 0.90 {
		result.Recommendations = append(result.Recommendations, core.Recommendation{
			Priority: "MEDIUM",
			Category: "OPTIMIZATION",
			Title:    "提高缓存命中率",
			Message:  fmt.Sprintf("当前命中率%.2f%%偏低", metrics.CacheHitRate*100),
			Actions: []string{
				"分析缓存键的访问模式",
				"调整缓存过期策略",
				"考虑增加缓存容量",
			},
		})
	}

	return result
}

// EvaluateKafka Kafka特定评估
func (se *StabilityEvaluator) EvaluateKafka(metrics *core.StabilityMetrics) *core.EvaluationResult {
	result := se.Evaluate(metrics)

	// 添加Kafka特定检查
	if metrics.MessageLag > 1000 {
		result.Issues = append(result.Issues, core.Issue{
			Type:     "high_message_lag",
			Severity: "MEDIUM",
			Metric:   "message_lag",
			Current:  float64(metrics.MessageLag),
			Expected: 1000,
			Message:  "消息积压过多",
		})
	}

	return result
}

// SetThresholds 设置自定义阈值
func (se *StabilityEvaluator) SetThresholds(thresholds *core.Thresholds) {
	se.thresholds = thresholds
}

// GetDefaultThresholds 获取默认阈值
func (se *StabilityEvaluator) GetDefaultThresholds() *core.Thresholds {
	return DefaultThresholds()
}
