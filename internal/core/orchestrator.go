package core

import (
	"context"
	"time"
)

// Orchestrator 测试编排器接口
type Orchestrator interface {
	// Run 运行测试
	// ctx: 上下文，用于超时和取消
	// config: 测试配置
	// 返回稳定性指标和错误
	Run(ctx context.Context, config Config) (*StabilityMetrics, error)

	// Pause 暂停测试
	Pause() error

	// Resume 恢复测试
	Resume() error

	// Stop 停止测试
	Stop() error

	// GetStatus 获取测试状态
	GetStatus() *OrchestratorStatus
}

// OrchestratorStatus 测试编排器状态
type OrchestratorStatus struct {
	State       string        // running, paused, stopped, completed
	Progress    float64       // 进度 0-1
	ElapsedTime time.Duration // 已运行时间
	Operations  int64         // 已完成操作数
}
