package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"middleware-chaos-testing/internal/collector"
	"middleware-chaos-testing/internal/core"
	"middleware-chaos-testing/internal/evaluator"
	"middleware-chaos-testing/internal/middleware"
	"middleware-chaos-testing/internal/reporter"
)

var (
	version = "0.1.0-dev"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "mct",
	Short: "Middleware Chaos Testing Tool",
	Long: `A tool for testing middleware stability under chaos scenarios.
Supports Redis, Kafka, and other middleware systems.`,
	Version: version,
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run stability test",
	Long:  "Run a stability test against the specified middleware",
	RunE:  runTest,
}

var (
	middlewareType string
	host           string
	port           int
	duration       time.Duration
	operations     int
	outputFormat   string
	reportPath     string
	configFile     string
)

func init() {
	testCmd.Flags().StringVar(&middlewareType, "middleware", "", "Middleware type (redis|kafka) [required]")
	testCmd.Flags().StringVar(&host, "host", "localhost", "Middleware host")
	testCmd.Flags().IntVar(&port, "port", 0, "Middleware port (default: 6379 for redis, 9092 for kafka)")
	testCmd.Flags().DurationVar(&duration, "duration", 60*time.Second, "Test duration")
	testCmd.Flags().IntVar(&operations, "operations", 10000, "Number of operations to perform")
	testCmd.Flags().StringVar(&outputFormat, "output", "console", "Output format (console|json|markdown)")
	testCmd.Flags().StringVar(&reportPath, "report-path", "", "Report output path (default: stdout)")
	testCmd.Flags().StringVar(&configFile, "config", "", "Config file path")

	testCmd.MarkFlagRequired("middleware")

	rootCmd.AddCommand(testCmd)
}

func runTest(cmd *cobra.Command, args []string) error {
	// 设置默认端口
	if port == 0 {
		switch middlewareType {
		case "redis":
			port = 6379
		case "kafka":
			port = 9092
		default:
			return fmt.Errorf("unsupported middleware type: %s", middlewareType)
		}
	}

	fmt.Printf("Starting %s stability test...\n", middlewareType)
	fmt.Printf("Target: %s:%d\n", host, port)
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Operations: %d\n\n", operations)

	// 执行测试
	ctx, cancel := context.WithTimeout(context.Background(), duration+30*time.Second)
	defer cancel()

	metrics, err := executeTest(ctx, middlewareType, host, port, duration, operations)
	if err != nil {
		return fmt.Errorf("test execution failed: %w", err)
	}

	// 评分 - 根据中间件类型使用不同的阈值
	var eval *evaluator.StabilityEvaluator
	if middlewareType == "kafka" {
		// Kafka使用专用阈值（符合业界最佳实践）
		eval = evaluator.NewStabilityEvaluator(evaluator.KafkaThresholds())
	} else {
		// 其他中间件使用默认阈值
		eval = evaluator.NewStabilityEvaluator(nil)
	}

	var result *core.EvaluationResult

	switch middlewareType {
	case "redis":
		result = eval.EvaluateRedis(metrics)
	case "kafka":
		result = eval.EvaluateKafka(metrics)
	default:
		result = eval.Evaluate(metrics)
	}

	// 生成报告
	output := os.Stdout
	if reportPath != "" {
		f, err := os.Create(reportPath)
		if err != nil {
			return fmt.Errorf("failed to create report file: %w", err)
		}
		defer f.Close()
		output = f
	}

	if err := generateReport(metrics, result, outputFormat, output); err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	// 根据测试状态设置退出码
	switch result.Status {
	case core.StatusFail:
		os.Exit(1)
	case core.StatusWarning:
		os.Exit(2)
	default:
		os.Exit(0)
	}

	return nil
}

func executeTest(ctx context.Context, middlewareType, host string, port int, duration time.Duration, operations int) (*core.StabilityMetrics, error) {
	coll := collector.NewMetricsCollector()

	switch middlewareType {
	case "redis":
		return executeRedisTest(ctx, host, port, duration, operations, coll)
	case "kafka":
		return executeKafkaTest(ctx, host, port, duration, operations, coll)
	default:
		return nil, fmt.Errorf("unsupported middleware type: %s", middlewareType)
	}
}

func executeRedisTest(ctx context.Context, host string, port int, duration time.Duration, operations int, coll *collector.MetricsCollector) (*core.StabilityMetrics, error) {
	// 创建Redis客户端
	cfg := &middleware.RedisConfig{
		Host:     host,
		Port:     port,
		Password: "",
		DB:       0,
		Timeout:  5 * time.Second,
	}

	client := middleware.NewRedisClient(cfg)

	// 连接
	startConnect := time.Now()
	if err := client.Connect(ctx); err != nil {
		coll.RecordConnectionAttempt(false, time.Since(startConnect))
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	coll.RecordConnectionAttempt(true, time.Since(startConnect))
	defer client.Disconnect(ctx)

	// 运行测试
	testCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	opsPerformed := 0
	ticker := time.NewTicker(duration / time.Duration(operations))
	defer ticker.Stop()

	for {
		select {
		case <-testCtx.Done():
			goto DONE
		case <-ticker.C:
			if opsPerformed >= operations {
				goto DONE
			}

			// 执行操作（SET + GET）
			key := fmt.Sprintf("test:key:%d", opsPerformed)
			value := fmt.Sprintf("value-%d", opsPerformed)

			// SET操作
			setResult, _ := client.Execute(testCtx, &middleware.RedisSetOperation{
				OpKey:   key,
				OpValue: []byte(value),
			})
			if setResult != nil {
				coll.RecordOperation(setResult)
			}

			// GET操作
			getResult, _ := client.Execute(testCtx, &middleware.RedisGetOperation{
				OpKey: key,
			})
			if getResult != nil {
				coll.RecordOperation(getResult)
			}

			opsPerformed += 2
		}
	}

DONE:
	return coll.GetMetrics(), nil
}

func executeKafkaTest(ctx context.Context, host string, port int, duration time.Duration, operations int, coll *collector.MetricsCollector) (*core.StabilityMetrics, error) {
	// 创建Kafka客户端
	brokers := []string{fmt.Sprintf("%s:%d", host, port)}
	cfg := &middleware.KafkaConfig{
		Brokers: brokers,
		Topic:   "chaos-test-topic",
		GroupID: "chaos-test-group",
		Timeout: 5 * time.Second,
	}

	client := middleware.NewKafkaClient(cfg)

	// 连接
	startConnect := time.Now()
	if err := client.Connect(ctx); err != nil {
		coll.RecordConnectionAttempt(false, time.Since(startConnect))
		return nil, fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	coll.RecordConnectionAttempt(true, time.Since(startConnect))
	defer client.Disconnect(ctx)

	// 运行测试
	testCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	opsPerformed := 0
	ticker := time.NewTicker(duration / time.Duration(operations))
	defer ticker.Stop()

	for {
		select {
		case <-testCtx.Done():
			goto DONE
		case <-ticker.C:
			if opsPerformed >= operations {
				goto DONE
			}

			// 执行操作（Produce + Consume）
			key := fmt.Sprintf("test-key-%d", opsPerformed)
			value := fmt.Sprintf("test-value-%d-%d", opsPerformed, time.Now().Unix())

			// Produce操作
			produceOp := &middleware.KafkaProduceOperation{
				OpKey:   key,
				OpValue: []byte(value),
			}
			produceResult, _ := client.Execute(testCtx, produceOp)
			if produceResult != nil {
				coll.RecordOperation(produceResult)
			}

			// Consume操作（尝试读取）
			consumeOp := &middleware.KafkaConsumeOperation{
				MaxWait: 100 * time.Millisecond,
			}
			consumeResult, _ := client.Execute(testCtx, consumeOp)
			if consumeResult != nil {
				coll.RecordOperation(consumeResult)
			}

			opsPerformed += 2
		}
	}

DONE:
	// 获取Kafka统计信息
	stats := client.GetStats()
	metrics := coll.GetMetrics()

	// 添加Kafka特定指标
	if lag, ok := stats["reader_lag"].(int64); ok {
		metrics.MessageLag = lag
	}

	return metrics, nil
}

func generateReport(metrics *core.StabilityMetrics, evaluation *core.EvaluationResult, format string, output *os.File) error {
	switch format {
	case "json":
		rep := reporter.NewJSONReporter()
		return rep.GenerateReport(metrics, evaluation, output)
	case "markdown", "md":
		rep := reporter.NewMarkdownReporter()
		return rep.GenerateReport(metrics, evaluation, output)
	case "console", "":
		rep := reporter.NewConsoleReporter()
		return rep.GenerateReport(metrics, evaluation, output)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}
