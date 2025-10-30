# 接口规范定义

## 1. 核心接口

### 1.1 MiddlewareClient 接口

中间件客户端的统一接口，所有中间件适配器必须实现此接口。

```go
package core

import (
    "context"
    "time"
)

// MiddlewareClient 中间件客户端接口
type MiddlewareClient interface {
    // Connect 建立连接
    // 返回错误表示连接失败
    Connect(ctx context.Context) error

    // Disconnect 断开连接
    // 返回错误表示断开失败
    Disconnect(ctx context.Context) error

    // Execute 执行操作
    // op: 要执行的操作
    // 返回操作结果和错误
    Execute(ctx context.Context, op Operation) (*Result, error)

    // HealthCheck 健康检查
    // 返回错误表示服务不健康
    HealthCheck(ctx context.Context) error

    // GetMetrics 获取客户端级别的指标
    // 返回当前的指标快照
    GetMetrics() *ClientMetrics
}
```

**接口设计原则:**
- 所有方法都接受 `context.Context` 以支持超时和取消
- 方法名清晰表达意图
- 返回值统一为 (结果, error) 的形式
- 线程安全，支持并发调用

### 1.2 Operation 接口

表示一个可执行的操作。

```go
// Operation 操作接口
type Operation interface {
    // Type 返回操作类型
    Type() OperationType

    // Key 返回操作的键（如果适用）
    Key() string

    // Value 返回操作的值（如果适用）
    Value() []byte

    // Metadata 返回操作的元数据
    Metadata() map[string]interface{}
}

// OperationType 操作类型
type OperationType string

const (
    OpTypeRead   OperationType = "read"
    OpTypeWrite  OperationType = "write"
    OpTypeDelete OperationType = "delete"
    OpTypeCustom OperationType = "custom"
)
```

### 1.3 Result 接口

表示操作执行的结果。

```go
// Result 操作结果
type Result struct {
    // Success 操作是否成功
    Success bool

    // Duration 操作耗时
    Duration time.Duration

    // Error 错误信息（如果失败）
    Error error

    // Data 返回的数据（如果有）
    Data []byte

    // Metadata 结果元数据
    Metadata map[string]interface{}

    // Timestamp 操作完成时间
    Timestamp time.Time
}
```

### 1.4 MetricsCollector 接口

指标收集器接口，负责收集和聚合测试指标。

```go
// MetricsCollector 指标收集器接口
type MetricsCollector interface {
    // RecordOperation 记录一次操作
    RecordOperation(result *Result)

    // RecordConnectionAttempt 记录连接尝试
    RecordConnectionAttempt(success bool, duration time.Duration)

    // RecordError 记录错误
    RecordError(err error, errorType ErrorType)

    // GetMetrics 获取当前聚合的指标
    GetMetrics() *StabilityMetrics

    // Reset 重置指标
    Reset()
}

// ErrorType 错误类型
type ErrorType string

const (
    ErrorTypeNetwork       ErrorType = "network"
    ErrorTypeTimeout       ErrorType = "timeout"
    ErrorTypeAuthentication ErrorType = "authentication"
    ErrorTypeDataLoss      ErrorType = "data_loss"
    ErrorTypeOther         ErrorType = "other"
)
```

### 1.5 StabilityMetrics 结构

包含所有稳定性相关的指标。

```go
// StabilityMetrics 稳定性指标
type StabilityMetrics struct {
    // 可用性指标
    TotalOperations      int64   // 总操作数
    SuccessfulOperations int64   // 成功操作数
    FailedOperations     int64   // 失败操作数
    Availability         float64 // 可用性 (成功率)

    // 连接指标
    TotalConnectionAttempts      int64   // 总连接尝试数
    SuccessfulConnectionAttempts int64   // 成功连接数
    ConnectionSuccessRate        float64 // 连接成功率

    // 性能指标
    P50Latency    time.Duration // P50延迟
    P95Latency    time.Duration // P95延迟
    P99Latency    time.Duration // P99延迟
    AvgLatency    time.Duration // 平均延迟
    MaxLatency    time.Duration // 最大延迟
    MinLatency    time.Duration // 最小延迟
    Throughput    float64       // 吞吐量 (ops/s)

    // 可靠性指标
    ErrorRate         float64 // 错误率
    DataLossRate      float64 // 数据丢失率
    DataConsistency   float64 // 数据一致性
    DuplicateRate     float64 // 重复率

    // 恢复性指标
    MTBF                  time.Duration // 平均故障间隔时间
    MTTR                  time.Duration // 平均恢复时间
    TotalReconnectAttempts int64         // 重连尝试次数
    SuccessfulReconnects   int64         // 成功重连次数
    ReconnectSuccessRate   float64       // 重连成功率

    // 错误统计
    ErrorsByType map[ErrorType]int64 // 按类型分类的错误数

    // 时间相关
    StartTime time.Time     // 测试开始时间
    EndTime   time.Time     // 测试结束时间
    Duration  time.Duration // 测试持续时间

    // 中间件特定指标（可选）
    // Redis
    CacheHitRate   float64 // 缓存命中率
    MemoryUsage    float64 // 内存使用率
    KeyspaceUtilization float64 // 键空间利用率

    // Kafka
    MessageLag      int64   // 消息积压
    ConsumerLag     time.Duration // 消费延迟
    DuplicateMessages int64 // 重复消息数
    RebalanceCount  int64   // 重平衡次数
}
```

### 1.6 Evaluator 接口

稳定性评估器接口，负责评分和生成建议。

```go
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
```

### 1.7 EvaluationResult 结构

评估结果，包含评分、等级、问题和建议。

```go
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

// StabilityGrade 稳定性等级
type StabilityGrade string

const (
    GradeExcellent StabilityGrade = "EXCELLENT" // 优秀 (90-100)
    GradeGood      StabilityGrade = "GOOD"      // 良好 (80-89)
    GradeFair      StabilityGrade = "FAIR"      // 一般 (70-79)
    GradePoor      StabilityGrade = "POOR"      // 较差 (60-69)
    GradeFailed    StabilityGrade = "FAILED"    // 失败 (<60)
)

// TestStatus 测试状态
type TestStatus string

const (
    StatusPass    TestStatus = "PASS"    // 通过
    StatusWarning TestStatus = "WARNING" // 警告
    StatusFail    TestStatus = "FAIL"    // 失败
)

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
```

### 1.8 Reporter 接口

报告生成器接口。

```go
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
```

### 1.9 Config 接口

配置管理接口。

```go
// Config 配置接口
type Config interface {
    // GetMiddlewareType 获取中间件类型
    GetMiddlewareType() string

    // GetConnectionConfig 获取连接配置
    GetConnectionConfig() *ConnectionConfig

    // GetTestConfig 获取测试配置
    GetTestConfig() *TestConfig

    // GetThresholds 获取阈值配置
    GetThresholds() *Thresholds

    // GetOutputConfig 获取输出配置
    GetOutputConfig() *OutputConfig

    // Validate 验证配置
    Validate() error
}

// ConnectionConfig 连接配置
type ConnectionConfig struct {
    Host     string
    Port     int
    Username string
    Password string
    Database int           // Redis DB
    Timeout  time.Duration
    // Kafka特定
    Brokers []string
    Topic   string
    GroupID string
}

// TestConfig 测试配置
type TestConfig struct {
    Duration    time.Duration // 测试持续时间
    Operations  int           // 操作次数
    Concurrency int           // 并发数
    Workload    []WorkloadConfig
}

// WorkloadConfig 工作负载配置
type WorkloadConfig struct {
    Operation   string  // 操作类型
    Weight      int     // 权重
    KeyPattern  string  // 键模式
    ValueSize   int     // 值大小
}

// OutputConfig 输出配置
type OutputConfig struct {
    Format                  string // console, json, markdown
    Path                    string // 报告保存路径
    IncludeRecommendations bool   // 是否包含建议
}
```

### 1.10 Orchestrator 接口

测试编排器接口。

```go
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
    GetStatus() *TestStatus
}

// TestStatus 测试状态
type TestStatus struct {
    State       string        // running, paused, stopped, completed
    Progress    float64       // 进度 0-1
    ElapsedTime time.Duration // 已运行时间
    Operations  int64         // 已完成操作数
}
```

## 2. 接口实现规范

### 2.1 错误处理

所有接口实现必须遵循以下错误处理规范：

```go
// 定义标准错误类型
var (
    ErrConnectionFailed    = errors.New("connection failed")
    ErrOperationTimeout    = errors.New("operation timeout")
    ErrInvalidConfig       = errors.New("invalid configuration")
    ErrClientNotConnected  = errors.New("client not connected")
    ErrUnsupportedOperation = errors.New("unsupported operation")
)

// 使用 errors.Is 和 errors.As 进行错误判断
if errors.Is(err, ErrConnectionFailed) {
    // 处理连接失败
}

// 包装错误以保留上下文
return fmt.Errorf("failed to execute operation: %w", err)
```

### 2.2 上下文传递

所有长时间运行的操作必须接受并正确处理 `context.Context`：

```go
func (c *RedisClient) Execute(ctx context.Context, op Operation) (*Result, error) {
    // 检查上下文是否已取消
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // 在执行操作时使用上下文
    result, err := c.client.Get(ctx, op.Key()).Result()

    return &Result{...}, nil
}
```

### 2.3 并发安全

所有接口实现必须是线程安全的：

```go
type MetricsCollector struct {
    mu      sync.RWMutex
    metrics *StabilityMetrics
}

func (mc *MetricsCollector) RecordOperation(result *Result) {
    mc.mu.Lock()
    defer mc.mu.Unlock()

    mc.metrics.TotalOperations++
    if result.Success {
        mc.metrics.SuccessfulOperations++
    } else {
        mc.metrics.FailedOperations++
    }
}

func (mc *MetricsCollector) GetMetrics() *StabilityMetrics {
    mc.mu.RLock()
    defer mc.mu.RUnlock()

    // 返回副本而不是原始数据
    return mc.metrics.Clone()
}
```

### 2.4 资源管理

实现必须正确管理资源，避免泄漏：

```go
func (c *RedisClient) Connect(ctx context.Context) error {
    // 创建连接
    client := redis.NewClient(&redis.Options{...})

    // 测试连接
    if err := client.Ping(ctx).Err(); err != nil {
        client.Close() // 确保失败时关闭连接
        return fmt.Errorf("ping failed: %w", err)
    }

    c.client = client
    return nil
}

func (c *RedisClient) Disconnect(ctx context.Context) error {
    if c.client != nil {
        err := c.client.Close()
        c.client = nil // 避免重复关闭
        return err
    }
    return nil
}
```

### 2.5 日志和追踪

所有接口实现应支持结构化日志：

```go
import "log/slog"

func (c *RedisClient) Execute(ctx context.Context, op Operation) (*Result, error) {
    start := time.Now()

    slog.DebugContext(ctx, "executing operation",
        "type", op.Type(),
        "key", op.Key(),
    )

    result, err := c.executeInternal(ctx, op)

    slog.InfoContext(ctx, "operation completed",
        "type", op.Type(),
        "success", result.Success,
        "duration", time.Since(start),
    )

    return result, err
}
```

## 3. 测试接口

### 3.1 Mock接口

所有核心接口都应提供Mock实现用于测试：

```go
// MockMiddlewareClient Mock中间件客户端
type MockMiddlewareClient struct {
    ConnectFunc      func(ctx context.Context) error
    DisconnectFunc   func(ctx context.Context) error
    ExecuteFunc      func(ctx context.Context, op Operation) (*Result, error)
    HealthCheckFunc  func(ctx context.Context) error
    GetMetricsFunc   func() *ClientMetrics
}

func (m *MockMiddlewareClient) Connect(ctx context.Context) error {
    if m.ConnectFunc != nil {
        return m.ConnectFunc(ctx)
    }
    return nil
}

// ... 其他方法类似
```

### 3.2 测试工具接口

```go
// TestHelper 测试辅助工具接口
type TestHelper interface {
    // SetupTestEnvironment 设置测试环境
    SetupTestEnvironment() error

    // TeardownTestEnvironment 清理测试环境
    TeardownTestEnvironment() error

    // CreateTestData 创建测试数据
    CreateTestData(count int) error

    // VerifyTestData 验证测试数据
    VerifyTestData() (bool, error)

    // CleanTestData 清理测试数据
    CleanTestData() error
}
```

## 4. 扩展接口

### 4.1 插件接口

预留插件扩展能力：

```go
// Plugin 插件接口
type Plugin interface {
    // Name 返回插件名称
    Name() string

    // Version 返回插件版本
    Version() string

    // Initialize 初始化插件
    Initialize(config map[string]interface{}) error

    // OnBeforeTest 测试开始前回调
    OnBeforeTest(ctx context.Context) error

    // OnAfterTest 测试结束后回调
    OnAfterTest(ctx context.Context, metrics *StabilityMetrics) error

    // OnOperationComplete 操作完成回调
    OnOperationComplete(result *Result) error

    // Shutdown 关闭插件
    Shutdown() error
}

// PluginManager 插件管理器接口
type PluginManager interface {
    // RegisterPlugin 注册插件
    RegisterPlugin(plugin Plugin) error

    // GetPlugin 获取插件
    GetPlugin(name string) (Plugin, error)

    // ListPlugins 列出所有插件
    ListPlugins() []Plugin
}
```

## 5. 版本兼容性

### 5.1 接口版本控制

接口遵循语义化版本控制：

- **MAJOR**: 不兼容的API变更
- **MINOR**: 向后兼容的功能新增
- **PATCH**: 向后兼容的问题修正

### 5.2 废弃策略

接口废弃遵循以下原则：

1. 在废弃前至少保留一个大版本
2. 使用 `@deprecated` 标记废弃的接口
3. 提供迁移指南

```go
// Deprecated: Use NewMiddlewareClientV2 instead.
// This function will be removed in v2.0.0
func NewMiddlewareClient() MiddlewareClient {
    // ...
}
```

## 6. 性能要求

### 6.1 接口性能指标

- **Connect/Disconnect**: < 5s
- **Execute**: < 100ms (P99)
- **GetMetrics**: < 1ms
- **HealthCheck**: < 1s
- **Evaluate**: < 100ms

### 6.2 并发性能

- 支持至少 **1000 并发连接**
- 支持至少 **10000 ops/s** 的操作执行

## 7. 文档要求

所有公共接口和方法必须包含：

1. 功能描述
2. 参数说明
3. 返回值说明
4. 使用示例
5. 错误情况说明
6. 并发安全性说明

示例：

```go
// Execute 执行中间件操作
//
// 此方法执行指定的操作并返回结果。该方法是线程安全的，
// 可以从多个goroutine并发调用。
//
// Parameters:
//   - ctx: 上下文，用于超时和取消控制
//   - op: 要执行的操作
//
// Returns:
//   - *Result: 操作结果，包含成功状态、耗时和数据
//   - error: 错误信息，如果操作失败
//
// Errors:
//   - ErrClientNotConnected: 客户端未连接
//   - ErrOperationTimeout: 操作超时
//   - ErrUnsupportedOperation: 不支持的操作类型
//
// Example:
//
//     client := NewRedisClient(config)
//     ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//     defer cancel()
//
//     op := &SetOperation{key: "test", value: []byte("data")}
//     result, err := client.Execute(ctx, op)
//     if err != nil {
//         log.Fatal(err)
//     }
//     fmt.Printf("Success: %v, Duration: %v\n", result.Success, result.Duration)
//
func (c *RedisClient) Execute(ctx context.Context, op Operation) (*Result, error)
```
