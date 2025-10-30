package evaluator_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"middleware-chaos-testing/internal/core"
	"middleware-chaos-testing/internal/evaluator"
)

// StabilityEvaluatorTestSuite 稳定性评估器测试套件
type StabilityEvaluatorTestSuite struct {
	suite.Suite
	evaluator *evaluator.StabilityEvaluator
}

// SetupTest 每个测试前执行
func (suite *StabilityEvaluatorTestSuite) SetupTest() {
	suite.evaluator = evaluator.NewStabilityEvaluator(nil) // 使用默认阈值
}

// TestEvaluate_PerfectScore 测试完美分数 (100分)
func (suite *StabilityEvaluatorTestSuite) TestEvaluate_PerfectScore() {
	metrics := &core.StabilityMetrics{
		Availability:         0.9999, // 99.99%
		P95Latency:           10 * time.Millisecond,
		P99Latency:           20 * time.Millisecond,
		ErrorRate:            0.0001, // 0.01%
		DataLossRate:         0.0,
		MTTR:                 5 * time.Second,
		ReconnectSuccessRate: 0.99,
	}

	result := suite.evaluator.Evaluate(metrics)

	suite.Equal(core.GradeExcellent, result.Grade, "Grade should be EXCELLENT")
	suite.GreaterOrEqual(result.Score, 90.0, "Score should be >= 90")
	suite.Equal(core.StatusPass, result.Status, "Status should be PASS")
	suite.Empty(result.Issues, "Should have no issues")
	suite.NotEmpty(result.Rationale, "Should have rationale")
}

// TestEvaluate_LowAvailability_Fails 测试可用性不足导致失败
func (suite *StabilityEvaluatorTestSuite) TestEvaluate_LowAvailability_Fails() {
	metrics := &core.StabilityMetrics{
		Availability: 0.94,  // 94% - 低于95%最低标准
		P95Latency:   50 * time.Millisecond,
		P99Latency:   100 * time.Millisecond,
		ErrorRate:    0.06,  // 6%
		MTTR:         30 * time.Second,
	}

	result := suite.evaluator.Evaluate(metrics)

	suite.Equal(core.GradeFailed, result.Grade, "Grade should be FAILED")
	suite.Equal(core.StatusFail, result.Status, "Status should be FAIL")
	suite.Less(result.Score, 60.0, "Score should be < 60")

	// 验证问题列表
	suite.NotEmpty(result.Issues, "Should have issues")

	// 应该包含低可用性的CRITICAL问题
	hasCriticalIssue := false
	for _, issue := range result.Issues {
		if issue.Severity == "CRITICAL" && strings.Contains(issue.Type, "availability") {
			hasCriticalIssue = true
			break
		}
	}
	suite.True(hasCriticalIssue, "Should have CRITICAL availability issue")
}

// TestEvaluate_HighLatency_Warning 测试高延迟触发警告
func (suite *StabilityEvaluatorTestSuite) TestEvaluate_HighLatency_Warning() {
	metrics := &core.StabilityMetrics{
		Availability: 0.999,  // 99.9%
		P95Latency:   80 * time.Millisecond,  // 接近阈值
		P99Latency:   150 * time.Millisecond, // 超过good阈值
		ErrorRate:    0.001,  // 0.1%
		MTTR:         30 * time.Second,
		ReconnectSuccessRate: 0.96,
	}

	result := suite.evaluator.Evaluate(metrics)

	suite.Equal(core.GradeGood, result.Grade, "Grade should be GOOD")
	suite.GreaterOrEqual(result.Score, 80.0, "Score should be >= 80")
	suite.Less(result.Score, 90.0, "Score should be < 90")

	// 应该有警告状态或通过状态（取决于是否有HIGH问题）
	suite.NotEqual(core.StatusFail, result.Status, "Should not FAIL")
}

// TestEvaluate_GeneratesRecommendations 测试生成详细建议
func (suite *StabilityEvaluatorTestSuite) TestEvaluate_GeneratesRecommendations() {
	metrics := &core.StabilityMetrics{
		Availability:       0.98,  // 98% - 低于优秀标准
		ErrorRate:          0.02,  // 2% - 偏高
		DataLossRate:       0.001, // 有数据丢失
		P95Latency:         120 * time.Millisecond,
		P99Latency:         250 * time.Millisecond,
		MTTR:               40 * time.Second,
		ReconnectSuccessRate: 0.92,
	}

	result := suite.evaluator.Evaluate(metrics)

	// 应该包含多条针对性建议
	suite.NotEmpty(result.Recommendations, "Should have recommendations")
	suite.GreaterOrEqual(len(result.Recommendations), 1, "Should have at least 1 recommendation")

	// 验证建议包含具体行动项
	for _, rec := range result.Recommendations {
		suite.NotEmpty(rec.Title, "Recommendation should have title")
		suite.NotEmpty(rec.Actions, "Recommendation should have actions")
		suite.GreaterOrEqual(len(rec.Actions), 1, "Should have at least 1 action")
	}
}

// TestEvaluate_CustomThresholds 测试自定义阈值
func (suite *StabilityEvaluatorTestSuite) TestEvaluate_CustomThresholds() {
	customThresholds := &core.Thresholds{
		AvailabilityPass: 0.90,  // 降低最低要求到90%
		P95LatencyPass:   300 * time.Millisecond,
		ErrorRatePass:    0.05, // 5%
	}

	customEvaluator := evaluator.NewStabilityEvaluator(customThresholds)

	metrics := &core.StabilityMetrics{
		Availability: 0.92,  // 92%
		P95Latency:   250 * time.Millisecond,
		P99Latency:   450 * time.Millisecond,
		ErrorRate:    0.04,  // 4%
		MTTR:         60 * time.Second,
	}

	result := customEvaluator.Evaluate(metrics)

	// 使用自定义阈值应该不会失败（92% >= 90%）
	suite.NotEqual(core.StatusFail, result.Status, "Should not fail with custom thresholds")
	suite.GreaterOrEqual(result.Score, 60.0, "Score should be >= 60 with custom thresholds")
}

// TestEvaluate_RationaleGeneration 测试判断依据说明
func (suite *StabilityEvaluatorTestSuite) TestEvaluate_RationaleGeneration() {
	metrics := &core.StabilityMetrics{
		Availability: 0.999,
		P95Latency:   45 * time.Millisecond,
		P99Latency:   95 * time.Millisecond,
		ErrorRate:    0.001,
		MTTR:         25 * time.Second,
		ReconnectSuccessRate: 0.97,
	}

	result := suite.evaluator.Evaluate(metrics)

	suite.NotEmpty(result.Rationale, "Rationale should not be empty")
	suite.Contains(result.Rationale, "综合评分", "Should contain overall score")
	suite.Contains(result.Rationale, "各维度得分", "Should contain dimension scores")
	suite.Contains(result.Rationale, string(result.Grade), "Should contain grade")
}

// TestEvaluate_BoundaryCase_JustPass 测试边界条件 - 刚好及格
func (suite *StabilityEvaluatorTestSuite) TestEvaluate_BoundaryCase_JustPass() {
	metrics := &core.StabilityMetrics{
		Availability: 0.95,   // 刚好达到最低要求95%
		P95Latency:   200 * time.Millisecond, // 刚好达到pass阈值
		P99Latency:   500 * time.Millisecond, // 刚好达到pass阈值
		ErrorRate:    0.01,   // 刚好1%
		MTTR:         300 * time.Second, // 刚好300s
		ReconnectSuccessRate: 0.90,
	}

	result := suite.evaluator.Evaluate(metrics)

	suite.Equal(core.GradeFair, result.Grade, "Grade should be FAIR")
	suite.GreaterOrEqual(result.Score, 70.0, "Score should be >= 70")
	suite.Less(result.Score, 80.0, "Score should be < 80")

	// 虽然及格，但应该有警告
	suite.Equal(core.StatusWarning, result.Status, "Should be WARNING at boundary")
}

// TestEvaluate_MultipleIssues 测试多维度问题综合评分
func (suite *StabilityEvaluatorTestSuite) TestEvaluate_MultipleIssues() {
	metrics := &core.StabilityMetrics{
		Availability:          0.96,   // 略低
		P95Latency:           150 * time.Millisecond, // 超标
		P99Latency:           600 * time.Millisecond, // 严重超标
		ErrorRate:            0.015,  // 1.5% - 超标
		DataLossRate:         0.005,  // 有数据丢失
		MTTR:                 350 * time.Second, // 恢复慢
		ReconnectSuccessRate: 0.88,   // 重连率低
	}

	result := suite.evaluator.Evaluate(metrics)

	// 多个维度有问题，总分应该较低
	suite.Less(result.Score, 75.0, "Score should be low with multiple issues")

	// 应该有多个问题
	suite.GreaterOrEqual(len(result.Issues), 2, "Should have multiple issues")

	// 应该有建议
	suite.NotEmpty(result.Recommendations, "Should have recommendations")
}

// TestEvaluate_ScoreDimensions 测试各维度得分正确性
func (suite *StabilityEvaluatorTestSuite) TestEvaluate_ScoreDimensions() {
	metrics := &core.StabilityMetrics{
		Availability: 0.999,
		P95Latency:   50 * time.Millisecond,
		P99Latency:   100 * time.Millisecond,
		ErrorRate:    0.001,
		DataLossRate: 0.0,
		MTTR:         30 * time.Second,
		ReconnectSuccessRate: 0.95,
	}

	result := suite.evaluator.Evaluate(metrics)

	// 验证各维度得分
	suite.GreaterOrEqual(result.Scores.Availability, 0.0, "Availability score should be >= 0")
	suite.LessOrEqual(result.Scores.Availability, 30.0, "Availability score should be <= 30")

	suite.GreaterOrEqual(result.Scores.Performance, 0.0, "Performance score should be >= 0")
	suite.LessOrEqual(result.Scores.Performance, 25.0, "Performance score should be <= 25")

	suite.GreaterOrEqual(result.Scores.Reliability, 0.0, "Reliability score should be >= 0")
	suite.LessOrEqual(result.Scores.Reliability, 25.0, "Reliability score should be <= 25")

	suite.GreaterOrEqual(result.Scores.Resilience, 0.0, "Resilience score should be >= 0")
	suite.LessOrEqual(result.Scores.Resilience, 20.0, "Resilience score should be <= 20")

	// 验证总分是各维度的加权和
	expectedScore := result.Scores.Availability*0.30 +
		result.Scores.Performance*0.25 +
		result.Scores.Reliability*0.25 +
		result.Scores.Resilience*0.20

	suite.InDelta(expectedScore, result.Score, 0.01, "Total score should be weighted sum of dimensions")
}

// TestGetDefaultThresholds 测试获取默认阈值
func (suite *StabilityEvaluatorTestSuite) TestGetDefaultThresholds() {
	thresholds := suite.evaluator.GetDefaultThresholds()

	suite.NotNil(thresholds, "Default thresholds should not be nil")

	// 验证关键阈值
	suite.Equal(0.9999, thresholds.AvailabilityExcellent, "Availability excellent should be 99.99%")
	suite.Equal(0.95, thresholds.AvailabilityPass, "Availability pass should be 95%")

	suite.Equal(10*time.Millisecond, thresholds.P95LatencyExcellent, "P95 excellent should be 10ms")
	suite.Equal(200*time.Millisecond, thresholds.P95LatencyPass, "P95 pass should be 200ms")
}

// TestSetThresholds 测试设置自定义阈值
func (suite *StabilityEvaluatorTestSuite) TestSetThresholds() {
	customThresholds := &core.Thresholds{
		AvailabilityPass: 0.80,
		P95LatencyPass:   500 * time.Millisecond,
	}

	suite.evaluator.SetThresholds(customThresholds)

	// 验证设置后的阈值生效
	metrics := &core.StabilityMetrics{
		Availability: 0.85,  // 85%，高于新阈值80%
		P95Latency:   400 * time.Millisecond, // 低于新阈值500ms
		ErrorRate:    0.01,
		MTTR:         50 * time.Second,
	}

	result := suite.evaluator.Evaluate(metrics)

	// 使用自定义阈值，应该不会因为可用性失败
	suite.NotEqual(core.StatusFail, result.Status, "Should not fail with custom thresholds")
}

// TestEvaluate_GradeMapping 测试分数到等级的映射
func (suite *StabilityEvaluatorTestSuite) TestEvaluate_GradeMapping() {
	testCases := []struct {
		name          string
		availability  float64
		expectedGrade core.StabilityGrade
		minScore      float64
		maxScore      float64
	}{
		{
			name:          "Excellent - 99.99%",
			availability:  0.9999,
			expectedGrade: core.GradeExcellent,
			minScore:      90.0,
			maxScore:      100.0,
		},
		{
			name:          "Good - 99.9%",
			availability:  0.999,
			expectedGrade: core.GradeGood,
			minScore:      80.0,
			maxScore:      89.9,
		},
		{
			name:          "Fair - 99%",
			availability:  0.99,
			expectedGrade: core.GradeFair,
			minScore:      70.0,
			maxScore:      79.9,
		},
		{
			name:          "Poor - 96%",
			availability:  0.96,
			expectedGrade: core.GradePoor,
			minScore:      60.0,
			maxScore:      69.9,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			metrics := &core.StabilityMetrics{
				Availability:         tc.availability,
				P95Latency:           50 * time.Millisecond,
				P99Latency:           100 * time.Millisecond,
				ErrorRate:            0.001,
				MTTR:                 30 * time.Second,
				ReconnectSuccessRate: 0.95,
			}

			result := suite.evaluator.Evaluate(metrics)

			suite.Equal(tc.expectedGrade, result.Grade, "Grade should match expected")
			suite.GreaterOrEqual(result.Score, tc.minScore, "Score should be >= min")
			suite.LessOrEqual(result.Score, tc.maxScore, "Score should be <= max")
		})
	}
}

// TestStabilityEvaluatorTestSuite 运行测试套件
func TestStabilityEvaluatorTestSuite(t *testing.T) {
	suite.Run(t, new(StabilityEvaluatorTestSuite))
}
