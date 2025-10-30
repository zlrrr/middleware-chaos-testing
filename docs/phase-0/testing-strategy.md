# 测试策略

## 1. 测试策略概述

本项目采用 **测试驱动开发（TDD）** 和 **规范驱动开发（SDD）** 相结合的方式，确保代码质量和设计正确性。

### 1.1 核心原则

1. **测试先行**: 先写测试，后写实现
2. **红-绿-重构**: 测试失败 → 实现通过 → 重构优化
3. **高覆盖率**: 单元测试覆盖率目标 >= 85%
4. **接口驱动**: 基于接口定义进行测试和实现
5. **持续验证**: 每个检查点都要通过测试才能继续

### 1.2 测试金字塔

```
        /\
       /  \
      / E2E\          ← 端到端测试 (10%)
     /------\
    /Integra\         ← 集成测试 (20%)
   /----------\
  /   Unit     \      ← 单元测试 (70%)
 /--------------\
```

---

## 2. 单元测试策略

### 2.1 测试范围

**必须编写单元测试的组件**:
- 所有核心接口实现
- 所有计算和评分逻辑
- 所有数据处理函数
- 所有配置解析逻辑

**可选单元测试**:
- 简单的getter/setter
- 纯数据结构

### 2.2 单元测试框架

使用 `testify/suite` 进行测试组织：

```go
import (
    "testing"
    "github.com/stretchr/testify/suite"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

type RedisClientTestSuite struct {
    suite.Suite
    client *RedisClient
    mock   *MockRedisServer
}

func (suite *RedisClientTestSuite) SetupTest() {
    // 每个测试前执行
    suite.mock = NewMockRedisServer()
    suite.client = NewRedisClient(suite.mock.Config())
}

func (suite *RedisClientTestSuite) TearDownTest() {
    // 每个测试后执行
    suite.client.Disconnect(context.Background())
    suite.mock.Close()
}

func (suite *RedisClientTestSuite) TestConnect_Success() {
    err := suite.client.Connect(context.Background())
    suite.NoError(err)
    suite.True(suite.client.IsConnected())
}

func TestRedisClientTestSuite(t *testing.T) {
    suite.Run(t, new(RedisClientTestSuite))
}
```

### 2.3 Mock策略

**使用场景**:
- 外部依赖（Redis、Kafka等）
- 时间相关逻辑
- 随机数生成
- 网络IO

**Mock工具**:
- **testify/mock**: 基础Mock
- **gomock**: 高级Mock需求
- **miniredis**: Redis内存实现（用于快速测试）

**示例**:
```go
type MockMiddlewareClient struct {
    mock.Mock
}

func (m *MockMiddlewareClient) Execute(ctx context.Context, op Operation) (*Result, error) {
    args := m.Called(ctx, op)
    return args.Get(0).(*Result), args.Error(1)
}

// 使用
func TestOrchestrator_WithMock(t *testing.T) {
    mockClient := new(MockMiddlewareClient)
    mockClient.On("Execute", mock.Anything, mock.Anything).
        Return(&Result{Success: true}, nil)

    orchestrator := NewOrchestrator(mockClient)
    // ... 测试逻辑
}
```

### 2.4 测试用例设计

每个功能至少包含以下测试用例：

1. **正常场景** (Happy Path)
2. **边界条件** (Boundary Cases)
3. **错误处理** (Error Cases)
4. **并发安全** (Concurrency)

**示例**:
```go
// 1. 正常场景
func (suite *MetricsCollectorTestSuite) TestRecordOperation_Success() {
    result := &Result{Success: true, Duration: 10*time.Millisecond}
    suite.collector.RecordOperation(result)

    metrics := suite.collector.GetMetrics()
    suite.Equal(int64(1), metrics.TotalOperations)
    suite.Equal(int64(1), metrics.SuccessfulOperations)
}

// 2. 边界条件
func (suite *MetricsCollectorTestSuite) TestRecordOperation_ZeroDuration() {
    result := &Result{Success: true, Duration: 0}
    suite.collector.RecordOperation(result)

    metrics := suite.collector.GetMetrics()
    suite.Equal(time.Duration(0), metrics.MinLatency)
}

// 3. 错误处理
func (suite *MetricsCollectorTestSuite) TestRecordOperation_Failed() {
    result := &Result{Success: false, Error: errors.New("test error")}
    suite.collector.RecordOperation(result)

    metrics := suite.collector.GetMetrics()
    suite.Equal(int64(1), metrics.FailedOperations)
}

// 4. 并发安全
func (suite *MetricsCollectorTestSuite) TestRecordOperation_Concurrent() {
    const goroutines = 100
    const opsPerGoroutine = 100

    var wg sync.WaitGroup
    wg.Add(goroutines)

    for i := 0; i < goroutines; i++ {
        go func() {
            defer wg.Done()
            for j := 0; j < opsPerGoroutine; j++ {
                suite.collector.RecordOperation(&Result{Success: true})
            }
        }()
    }

    wg.Wait()

    metrics := suite.collector.GetMetrics()
    suite.Equal(int64(goroutines*opsPerGoroutine), metrics.TotalOperations)
}
```

### 2.5 表驱动测试

对于相似的测试场景，使用表驱动测试：

```go
func TestCalculateAvailabilityScore(t *testing.T) {
    tests := []struct {
        name         string
        availability float64
        wantScore    float64
        wantGrade    StabilityGrade
    }{
        {
            name:         "Excellent - 99.99%",
            availability: 0.9999,
            wantScore:    30.0,
            wantGrade:    GradeExcellent,
        },
        {
            name:         "Good - 99.9%",
            availability: 0.999,
            wantScore:    27.0,
            wantGrade:    GradeGood,
        },
        {
            name:         "Fair - 99%",
            availability: 0.99,
            wantScore:    24.0,
            wantGrade:    GradeFair,
        },
        {
            name:         "Failed - Below threshold",
            availability: 0.94,
            wantScore:    18.8,
            wantGrade:    GradeFailed,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            evaluator := NewStabilityEvaluator(nil)
            score := evaluator.calculateAvailabilityScore(tt.availability)
            assert.Equal(t, tt.wantScore, score)
        })
    }
}
```

### 2.6 覆盖率目标

- **总体覆盖率**: >= 85%
- **核心逻辑**: >= 90%
- **评分系统**: >= 95%
- **边界条件**: 100%

**检查命令**:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
go tool cover -func=coverage.out | grep total
```

---

## 3. 集成测试策略

### 3.1 测试范围

**集成测试场景**:
- 各组件协作流程
- 实际中间件连接测试
- 配置文件解析和验证
- 报告生成端到端流程

### 3.2 测试环境

使用Docker Compose搭建测试环境：

```yaml
# docker-compose.test.yml
version: '3.8'

services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  kafka:
    image: confluentinc/cp-kafka:latest
    ports:
      - "9092:9092"
    environment:
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092

  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
```

**启动测试环境**:
```bash
docker-compose -f docker-compose.test.yml up -d
```

### 3.3 集成测试示例

```go
// tests/integration/redis_integration_test.go
package integration_test

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/suite"
)

type RedisIntegrationTestSuite struct {
    suite.Suite
    client *RedisClient
}

func (suite *RedisIntegrationTestSuite) SetupSuite() {
    // 整个测试套件开始前执行一次
    // 等待Redis服务就绪
    suite.waitForRedis()
}

func (suite *RedisIntegrationTestSuite) SetupTest() {
    // 每个测试前执行
    config := &RedisConfig{
        Host: "localhost",
        Port: 6379,
    }
    suite.client = NewRedisClient(config)
    err := suite.client.Connect(context.Background())
    suite.Require().NoError(err)
}

func (suite *RedisIntegrationTestSuite) TearDownTest() {
    // 清理测试数据
    suite.client.FlushTestData()
    suite.client.Disconnect(context.Background())
}

func (suite *RedisIntegrationTestSuite) TestFullWorkflow() {
    ctx := context.Background()

    // 1. 写入数据
    op := &SetOperation{Key: "test:key", Value: []byte("test-value")}
    result, err := suite.client.Execute(ctx, op)
    suite.NoError(err)
    suite.True(result.Success)

    // 2. 读取数据
    getOp := &GetOperation{Key: "test:key"}
    result, err = suite.client.Execute(ctx, getOp)
    suite.NoError(err)
    suite.True(result.Success)
    suite.Equal([]byte("test-value"), result.Data)

    // 3. 验证指标
    metrics := suite.client.GetMetrics()
    suite.Equal(int64(2), metrics.TotalOperations)
    suite.Equal(int64(2), metrics.SuccessfulOperations)
}

func (suite *RedisIntegrationTestSuite) waitForRedis() {
    // 等待Redis就绪
    maxRetries := 30
    for i := 0; i < maxRetries; i++ {
        client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
        if err := client.Ping(context.Background()).Err(); err == nil {
            client.Close()
            return
        }
        time.Sleep(1 * time.Second)
    }
    suite.T().Fatal("Redis not ready after 30 seconds")
}

func TestRedisIntegrationTestSuite(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    suite.Run(t, new(RedisIntegrationTestSuite))
}
```

### 3.4 运行集成测试

```bash
# 启动测试环境
docker-compose -f docker-compose.test.yml up -d

# 运行集成测试
go test -v ./tests/integration/...

# 仅运行单元测试（跳过集成测试）
go test -short ./...

# 清理测试环境
docker-compose -f docker-compose.test.yml down -v
```

---

## 4. 端到端测试策略

### 4.1 CLI端到端测试

测试完整的CLI命令执行流程：

```go
// tests/e2e/cli_test.go
package e2e_test

import (
    "encoding/json"
    "os/exec"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestCLI_RedisTest_FullWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    // 执行CLI命令
    cmd := exec.Command("./bin/mct", "test",
        "--middleware", "redis",
        "--host", "localhost",
        "--port", "6379",
        "--duration", "10s",
        "--operations", "1000",
        "--output", "json",
    )

    output, err := cmd.CombinedOutput()
    assert.NoError(t, err, "CLI command should succeed")

    // 解析JSON输出
    var report struct {
        Evaluation struct {
            Score  float64 `json:"score"`
            Grade  string  `json:"grade"`
            Status string  `json:"status"`
        } `json:"evaluation"`
        Metrics struct {
            TotalOperations int64 `json:"total_operations"`
        } `json:"metrics"`
    }

    err = json.Unmarshal(output, &report)
    assert.NoError(t, err, "Should parse JSON output")

    // 验证结果
    assert.GreaterOrEqual(t, report.Evaluation.Score, 70.0, "Score should be passing")
    assert.NotEmpty(t, report.Evaluation.Grade)
    assert.Equal(t, int64(1000), report.Metrics.TotalOperations)
}

func TestCLI_ExitCodes(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        wantExit int // 0=PASS, 1=FAIL, 2=WARNING
    }{
        {
            name: "Successful test returns 0",
            args: []string{"test", "--middleware", "redis", "--duration", "5s"},
            wantExit: 0,
        },
        {
            name: "Failed connection returns 1",
            args: []string{"test", "--middleware", "redis", "--host", "invalid", "--duration", "5s"},
            wantExit: 1,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := exec.Command("./bin/mct", tt.args...)
            err := cmd.Run()

            if tt.wantExit == 0 {
                assert.NoError(t, err)
            } else {
                exitError, ok := err.(*exec.ExitError)
                assert.True(t, ok, "Should be exit error")
                assert.Equal(t, tt.wantExit, exitError.ExitCode())
            }
        })
    }
}
```

### 4.2 配置文件测试

```go
func TestCLI_WithConfigFile(t *testing.T) {
    // 创建临时配置文件
    configContent := `
name: "Redis Test"
middleware: "redis"
connection:
  host: "localhost"
  port: 6379
test:
  duration: 10s
  operations: 500
`
    configFile := createTempFile(t, configContent)
    defer os.Remove(configFile)

    // 使用配置文件运行测试
    cmd := exec.Command("./bin/mct", "test", "--config", configFile)
    output, err := cmd.CombinedOutput()
    assert.NoError(t, err)
    assert.Contains(t, string(output), "Redis Test")
}
```

---

## 5. 性能测试策略

### 5.1 基准测试

使用Go的benchmark功能进行性能测试：

```go
// internal/metrics/collector_bench_test.go
package metrics_test

import (
    "testing"
    "time"
)

func BenchmarkMetricsCollector_RecordOperation(b *testing.B) {
    collector := NewMetricsCollector()
    result := &Result{
        Success:  true,
        Duration: 10 * time.Millisecond,
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        collector.RecordOperation(result)
    }
}

func BenchmarkMetricsCollector_GetMetrics(b *testing.B) {
    collector := NewMetricsCollector()
    // 预先记录一些数据
    for i := 0; i < 10000; i++ {
        collector.RecordOperation(&Result{Success: true})
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = collector.GetMetrics()
    }
}

func BenchmarkMetricsCollector_Concurrent(b *testing.B) {
    collector := NewMetricsCollector()
    result := &Result{Success: true}

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            collector.RecordOperation(result)
        }
    })
}
```

**运行基准测试**:
```bash
go test -bench=. -benchmem ./internal/metrics/
```

### 5.2 负载测试

测试系统在高负载下的表现：

```go
func TestHighLoad_1000OpsPerSecond(t *testing.T) {
    orchestrator := NewOrchestrator(client)
    config := &TestConfig{
        Duration:    60 * time.Second,
        Concurrency: 50,
        TargetOPS:   1000,
    }

    ctx := context.Background()
    metrics, err := orchestrator.Run(ctx, config)

    assert.NoError(t, err)
    assert.GreaterOrEqual(t, metrics.Throughput, 950.0, "Should achieve at least 95% of target OPS")
    assert.LessOrEqual(t, metrics.P99Latency, 100*time.Millisecond, "P99 should be under 100ms")
}
```

---

## 6. 测试数据管理

### 6.1 测试数据生成

```go
// tests/testdata/generator.go
package testdata

type TestDataGenerator struct {
    prefix string
}

func NewTestDataGenerator() *TestDataGenerator {
    return &TestDataGenerator{
        prefix: fmt.Sprintf("test:%d:", time.Now().Unix()),
    }
}

func (g *TestDataGenerator) GenerateKey(id int) string {
    return fmt.Sprintf("%skey:%d", g.prefix, id)
}

func (g *TestDataGenerator) GenerateValue(size int) []byte {
    return make([]byte, size)
}

func (g *TestDataGenerator) Cleanup(client MiddlewareClient) error {
    // 清理所有带测试前缀的数据
    return client.DeleteByPrefix(context.Background(), g.prefix)
}
```

### 6.2 测试数据隔离

- 使用唯一前缀（包含时间戳）
- 测试结束后自动清理
- 避免污染生产数据

---

## 7. 持续集成测试

### 7.1 GitHub Actions配置

```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

      kafka:
        image: confluentinc/cp-kafka:latest
        ports:
          - 9092:9092

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Download dependencies
        run: go mod download

      - name: Run unit tests
        run: go test -short -coverprofile=coverage.out ./...

      - name: Run integration tests
        run: go test -v ./tests/integration/...

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out

      - name: Check coverage threshold
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$coverage < 85" | bc -l) )); then
            echo "Coverage $coverage% is below threshold 85%"
            exit 1
          fi
```

### 7.2 Pre-commit钩子

```bash
# .git/hooks/pre-commit
#!/bin/bash

echo "Running tests before commit..."

# 运行单元测试
go test -short ./...
if [ $? -ne 0 ]; then
    echo "Unit tests failed. Commit aborted."
    exit 1
fi

# 检查代码格式
gofmt -l . | grep -v vendor
if [ $? -eq 0 ]; then
    echo "Code is not formatted. Run 'gofmt -w .' first."
    exit 1
fi

# 运行静态检查
golangci-lint run
if [ $? -ne 0 ]; then
    echo "Linting failed. Commit aborted."
    exit 1
fi

echo "All checks passed!"
exit 0
```

---

## 8. 测试检查清单

### 8.1 Phase 0 测试检查清单

- [ ] 所有接口都有测试用例
- [ ] Mock实现可用
- [ ] 测试框架正常运行

### 8.2 Phase 1 Redis客户端测试检查清单

- [ ] 连接/断开测试
- [ ] 基本操作测试（SET/GET/DELETE）
- [ ] 错误处理测试
- [ ] 超时测试
- [ ] 重连测试
- [ ] 并发测试
- [ ] 指标收集测试
- [ ] 覆盖率 >= 85%

### 8.3 Phase 2 Kafka客户端测试检查清单

- [ ] 生产者测试
- [ ] 消费者测试
- [ ] 消息顺序测试
- [ ] 重复消息测试
- [ ] 错误处理测试
- [ ] 重平衡测试
- [ ] 覆盖率 >= 85%

### 8.4 Phase 3.5 评分系统测试检查清单

- [ ] 完美分数测试 (100分)
- [ ] 边界条件测试 (70分, 85分, 90分)
- [ ] 失败场景测试 (<60分)
- [ ] 各维度评分测试
- [ ] 问题识别测试
- [ ] 建议生成测试
- [ ] 自定义阈值测试
- [ ] Redis特定评分测试
- [ ] Kafka特定评分测试
- [ ] 覆盖率 >= 95%

### 8.5 Phase 4 CLI测试检查清单

- [ ] 参数解析测试
- [ ] 配置文件加载测试
- [ ] 完整工作流测试
- [ ] 输出格式测试（console/json/markdown）
- [ ] 退出码测试
- [ ] 错误处理测试

---

## 9. 测试最佳实践

### 9.1 命名规范

```
测试文件: <source_file>_test.go
测试函数: Test<FunctionName>_<Scenario>
基准测试: Benchmark<FunctionName>_<Scenario>
测试套件: <Component>TestSuite
```

### 9.2 测试结构

遵循 AAA 模式：
```go
func TestExample(t *testing.T) {
    // Arrange - 准备
    client := NewClient()
    expected := "expected value"

    // Act - 执行
    actual := client.DoSomething()

    // Assert - 断言
    assert.Equal(t, expected, actual)
}
```

### 9.3 错误信息

```go
// 好的错误信息
assert.Equal(t, 100, score, "Score should be 100 for perfect metrics")

// 差的错误信息
assert.Equal(t, 100, score)
```

### 9.4 测试隔离

- 每个测试独立运行
- 不依赖其他测试的执行顺序
- 使用 SetupTest/TearDownTest 进行清理

### 9.5 测试性能

- 避免睡眠（使用Mock时间）
- 使用合适的超时
- 并行运行独立测试

```go
func TestParallel(t *testing.T) {
    t.Parallel() // 并行运行
    // ... 测试逻辑
}
```

---

## 10. 测试工具链

### 10.1 必需工具

```bash
# 安装testify
go get github.com/stretchr/testify

# 安装gomock
go install github.com/golang/mock/mockgen@latest

# 安装golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 10.2 Makefile测试命令

```makefile
.PHONY: test test-unit test-integration test-e2e coverage

test: test-unit test-integration

test-unit:
	go test -short -v ./...

test-integration:
	docker-compose -f docker-compose.test.yml up -d
	go test -v ./tests/integration/...
	docker-compose -f docker-compose.test.yml down

test-e2e:
	make build
	docker-compose -f docker-compose.test.yml up -d
	go test -v ./tests/e2e/...
	docker-compose -f docker-compose.test.yml down

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out | grep total

bench:
	go test -bench=. -benchmem ./...

lint:
	golangci-lint run ./...
```

---

**文档版本**: v1.0
**最后更新**: 2025-10-30
