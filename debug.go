package main

import (
	"fmt"
	"time"

	"middleware-chaos-testing/internal/core"
	"middleware-chaos-testing/internal/evaluator"
)

func main() {
	eval := evaluator.NewStabilityEvaluator(nil)

	// TestEvaluate_ScoreDimensions的输入
	metrics := &core.StabilityMetrics{
		Availability: 0.999,
		P95Latency:   50 * time.Millisecond,
		P99Latency:   100 * time.Millisecond,
		ErrorRate:    0.001,
		DataLossRate: 0.0,
		MTTR:         30 * time.Second,
		ReconnectSuccessRate: 0.95,
	}

	result := eval.Evaluate(metrics)

	fmt.Printf("Total Score: %.2f\n", result.Score)
	fmt.Printf("Availability Score: %.2f\n", result.Scores.Availability)
	fmt.Printf("Performance Score: %.2f\n", result.Scores.Performance)
	fmt.Printf("Reliability Score: %.2f\n", result.Scores.Reliability)
	fmt.Printf("Resilience Score: %.2f\n", result.Scores.Resilience)

	// 测试期望的加权计算
	expected := result.Scores.Availability*0.30 +
		result.Scores.Performance*0.25 +
		result.Scores.Reliability*0.25 +
		result.Scores.Resilience*0.20

	fmt.Printf("\nExpected (weighted): %.2f\n", expected)
	fmt.Printf("Actual: %.2f\n", result.Score)
	fmt.Printf("Match: %v\n", expected == result.Score)
}
