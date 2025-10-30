package core

import "io"

// Reporter 报告生成器接口
type Reporter interface {
	// GenerateReport 生成报告
	// metrics: 稳定性指标
	// evaluation: 评估结果
	// output: 输出目标（io.Writer）
	GenerateReport(
		metrics *StabilityMetrics,
		evaluation *EvaluationResult,
		output io.Writer,
	) error
}

// ConsoleReporter 控制台报告生成器
type ConsoleReporter interface {
	Reporter
	// SetColorEnabled 设置是否启用颜色
	SetColorEnabled(enabled bool)
}

// JSONReporter JSON报告生成器
type JSONReporter interface {
	Reporter
	// SetIndent 设置JSON缩进
	SetIndent(indent string)
}

// MarkdownReporter Markdown报告生成器
type MarkdownReporter interface {
	Reporter
	// SetTemplate 设置自定义模板
	SetTemplate(template string)
}
