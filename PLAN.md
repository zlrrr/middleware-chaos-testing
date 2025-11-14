# middleware-chaos-testing
中间件混沌测试框架 - 验证Kafka、Redis等中间件在混沌场景下的稳定性

基于TDD/SDD原则，快速构建支持多中间件的稳定性测试工具

---

## 项目概述

### 目标
开发一个可扩展的中间件混沌测试框架，通过模拟各类用户操作，检测并量化中间件服务（Kafka、Redis等）在混沌场景下的稳定性指标。

### MVP目标
快速交付支持Redis和Kafka的命令行工具，具备核心稳定性检测能力和明确的判断标准。

### 核心特性
- ✅ 支持多种中间件客户端（Redis、Kafka优先）
- ✅ **可配置的测试持续时间**（命令行参数和配置文件）
- ✅ **智能稳定性评分系统**（0-100分，5个等级）
- ✅ **明确的通过/警告/失败判断**
- ✅ **可操作的改进建议**（按优先级排序）
- ✅ 工程化的稳定性指标监测
- ✅ 可选的Web监控界面
- ✅ 容器化一键部署
- ✅ TDD/SDD开发流程

---

## 一、技术栈选择

### 后端架构
- **主语言**: Go 1.21+ (高性能、并发友好)
- **备选语言**: Python 3.11+ (快速原型)
- **Web框架**: Gin (Go) / FastAPI (Python)
- **中间件客户端**:
  - Redis: go-redis / redis-py
  - Kafka: sarama / confluent-kafka-python
  - MongoDB: mongo-driver (Go)
  - RocketMQ: rocketmq-client-go
  - RabbitMQ: amqp091-go
  - EMQX: paho.mqtt.golang
  - Nacos: nacos-sdk-go

### 中间件版本兼容性

本项目客户端实现针对以下中间件版本进行适配和测试：

| 中间件 | 目标版本 | 客户端库 | 关键特性支持 |
|--------|----------|----------|--------------|
| **Redis** | 6.x / 7.x | `github.com/redis/go-redis/v9` | 支持Redis 6.x和7.x全部特性 |
| **Kafka** | 2.7.2 | `github.com/segmentio/kafka-go` | 支持Kafka 2.x协议 |
| **MongoDB** | 4.4.13 | `go.mongodb.org/mongo-driver` | 兼容MongoDB 4.4 Wire Protocol |
| **RocketMQ** | 5.1.3 | `github.com/apache/rocketmq-client-go/v2` | **支持remoting和grpc两种协议** |
| **RabbitMQ** | 3.12.7 | `github.com/rabbitmq/amqp091-go` | 完整AMQP 0-9-1协议支持 |
| **EMQX** | 5.8 (社区版) | `github.com/eclipse/paho.mqtt.golang` | MQTT 3.1.1 / 5.0协议 |
| **Nacos** | 2.4.3 | `github.com/nacos-group/nacos-sdk-go/v2` | 支持2.x新特性 |

#### 版本特定配置说明

**RocketMQ 5.1.3 协议选择**:
```go
type RocketMQConfig struct {
    // ... 其他配置
    Protocol string // "remoting" 或 "grpc"，默认: "remoting"
}
```
- `remoting`: RocketMQ传统协议（兼容性更好）
- `grpc`: gRPC协议（性能更优，RocketMQ 5.x推荐）

**MongoDB 4.4.13 注意事项**:
- 支持Replica Set和Sharded Cluster
- 不支持MongoDB 5.0+的新特性（Time Series Collections等）
- Wire Protocol Version: 13

**RabbitMQ 3.12.7 注意事项**:
- AMQP 0-9-1协议完整支持
- 支持所有交换机类型、死信队列、优先级队列
- 不支持AMQP 1.0协议

**EMQX 5.8 社区版**:
- 支持MQTT 3.1.1和5.0协议
- QoS 0/1/2全支持
- 共享订阅、保留消息、遗嘱消息

**Nacos 2.4.3**:
- 支持2.x gRPC长连接
- 配置管理和服务发现
- 命名空间、分组隔离
- **指标收集**: Prometheus Client
- **数据存储**: SQLite (MVP) → PostgreSQL (生产)
- **测试框架**: 
  - Go: testify, gomock
  - Python: pytest, unittest.mock

### 前端架构（可选Phase）
- **框架**: React + TypeScript
- **可视化**: ECharts / Recharts
- **UI库**: Ant Design / shadcn/ui
- **状态管理**: Zustand

### DevOps
- **容器化**: Docker + Docker Compose
- **CI/CD**: GitHub Actions
- **监控**: Prometheus + Grafana

---

## 二、稳定性指标体系

### 2.1 核心指标（必须实现）

#### 可用性指标（权重30%）
- **连接成功率**: 成功建立连接数 / 总连接尝试数
- **操作成功率**: 成功操作数 / 总操作数
- **错误率**: 错误操作数 / 总操作数
- **服务可用时间百分比**: (总时间 - 不可用时间) / 总时间

#### 性能指标（权重25%）
- **响应延迟**: P50, P95, P99延迟
- **吞吐量**: 每秒操作数 (OPS)
- **并发连接数**: 当前活跃连接数
- **队列深度**: 待处理请求队列长度

#### 可靠性指标（权重25%）
- **数据一致性**: 写入数据与读取数据的一致性校验
- **数据丢失率**: 丢失的消息/数据数量 / 总发送数量
- **重复率**: 重复接收的消息数 / 总接收数
- **乱序率**: 乱序消息数 / 总消息数（Kafka）

#### 恢复性指标（权重20%）
- **MTBF**: 平均故障间隔时间
- **MTTR**: 平均恢复时间
- **故障检测时间**: 从故障发生到检测到的时间
- **重连成功率**: 重连成功次数 / 重连尝试次数

### 2.2 中间件特定指标

#### Redis特定
- **缓存命中率**: 命中次数 / 总查询次数
- **键空间利用率**: 已用键数 / 最大键数
- **内存使用率**: 已用内存 / 最大内存
- **持久化延迟**: RDB/AOF写入延迟

#### Kafka特定
- **消息积压**: 未消费消息数量
- **消费延迟**: 生产时间 - 消费时间
- **分区重平衡次数**: Rebalance触发次数
- **ISR收缩次数**: In-Sync Replicas变化次数

---

## 三、稳定性评分标准

### 3.1 评分体系（0-100分）

| 分数区间 | 等级 | 状态 | 说明 |
|---------|------|------|------|
| 90-100 | EXCELLENT | ✅ PASS | 优秀，可直接用于生产环境 |
| 80-89 | GOOD | ✅ PASS | 良好，满足生产要求 |
| 70-79 | FAIR | ⚠️  WARNING | 一般，建议优化后使用 |
| 60-69 | POOR | ⚠️  WARNING | 较差，需要改进 |
| 0-59 | FAILED | ❌ FAIL | 失败，不建议用于生产 |

### 3.2 各维度评分权重

```
总分 = 可用性(30%) + 性能(25%) + 可靠性(25%) + 恢复力(20%)
```

### 3.3 判断逻辑

```go
if 存在CRITICAL问题 || 总分 < 70 {
    return "FAIL" // 测试失败
}

if 存在HIGH问题 || 总分 < 85 {
    return "WARNING" // 需要关注
}

return "PASS" // 测试通过
```

### 3.4 阈值配置示例

```yaml
# 默认阈值配置
thresholds:
  # 可用性阈值
  availability:
    excellent: 99.99%  # 四个9
    good: 99.9%        # 三个9
    fair: 99.0%        # 两个9
    pass: 95.0%        # 最低标准
    
  # P95延迟阈值
  p95_latency:
    excellent: 10ms
    good: 50ms
    fair: 100ms
    pass: 200ms
    
  # P99延迟阈值
  p99_latency:
    excellent: 20ms
    good: 100ms
    fair: 200ms
    pass: 500ms
    
  # 错误率阈值
  error_rate:
    excellent: 0.01%
    good: 0.1%
    fair: 0.5%
    pass: 1.0%
    
  # 恢复时间阈值
  mttr:
    excellent: 5s
    good: 30s
    fair: 60s
    pass: 300s
```

---

## 四、报告输出示例

### 4.1 控制台输出格式

```
==========================================
   中间件稳定性测试报告
==========================================

测试目标: Redis @ localhost:6379
测试时长: 60s
测试完成: 2025-10-30 14:30:00

------------------------------------------
  总体评分: 87.5/100 (GOOD) ✅ PASS
------------------------------------------

各维度得分:
  ✓ 可用性   28.5/30  (95.0%)  - 权重30%
  ✓ 性能     21.0/25  (84.0%)  - 权重25%  
  ✓ 可靠性   23.5/25  (94.0%)  - 权重25%
  ✓ 恢复力   14.5/20  (72.5%)  - 权重20%

------------------------------------------
  核心指标
------------------------------------------
可用性: 99.92% ✓
  - 总操作数: 10,000
  - 成功操作: 9,992
  - 失败操作: 8
  - 错误率: 0.08%

性能指标:
  - P50 延迟: 8ms ✓
  - P95 延迟: 45ms ✓
  - P99 延迟: 120ms ⚠️
  - 平均吞吐: 167 ops/s

可靠性:
  - 数据一致性: 100% ✓
  - 数据丢失率: 0% ✓

恢复性:
  - MTTR: 25s ✓
  - 重连成功率: 96% ✓

------------------------------------------
  发现的问题 (1个)
------------------------------------------
[MEDIUM] high_p99_latency
  指标: P99延迟
  当前值: 120ms
  期望值: ≤100ms
  说明: P99延迟超过良好阈值，影响尾部用户体验

------------------------------------------
  改进建议 (按优先级排序)
------------------------------------------

[MEDIUM] 优化响应延迟
分类: OPTIMIZATION
说明: 延迟指标接近可接受上限，建议优化
具体行动:
  1. 分析慢查询日志，优化热点操作
  2. 检查网络延迟和带宽瓶颈
  3. 优化数据结构和查询模式
  4. 考虑增加缓存层或读写分离
  5. 评估硬件资源是否充足

参考文档:
  - https://redis.io/topics/latency

------------------------------------------
  结论
------------------------------------------
✅ 测试通过

系统稳定性评级为 GOOD，总分87.5/100。
系统整体表现良好，满足生产环境要求。
建议关注P99延迟指标，按优先级实施优化建议。

详细报告已保存至: ./reports/redis-test-20251030-143000.json
==========================================
```

### 4.2 JSON报告格式

```json
{
  "test_info": {
    "name": "Redis Stability Test",
    "middleware": "redis",
    "target": "localhost:6379",
    "duration": "60s",
    "completed_at": "2025-10-30T14:30:00Z"
  },
  
  "evaluation": {
    "score": 87.5,
    "grade": "GOOD",
    "status": "PASS",
    
    "scores": {
      "availability": 28.5,
      "performance": 21.0,
      "reliability": 23.5,
      "resilience": 14.5
    },
    
    "rationale": "综合评分: 87.50/100 (GOOD)\n\n各维度得分:\n- 可用性: 28.50/30 (权重30%)\n- 性能: 21.00/25 (权重25%)\n- 可靠性: 23.50/25 (权重25%)\n- 恢复力: 14.50/20 (权重20%)\n\n✅ 测试通过: 系统稳定性符合预期，可以用于生产环境。\n"
  },
  
  "metrics": {
    "availability": 0.9992,
    "total_operations": 10000,
    "successful_operations": 9992,
    "failed_operations": 8,
    "error_rate": 0.0008,
    
    "latency": {
      "p50": "8ms",
      "p95": "45ms",
      "p99": "120ms",
      "avg": "12ms"
    },
    
    "throughput": 167,
    "data_consistency": 1.0,
    "data_loss_rate": 0.0,
    "mttr": "25s",
    "reconnect_success_rate": 0.96
  },
  
  "issues": [
    {
      "type": "high_p99_latency",
      "severity": "MEDIUM",
      "metric": "p99_latency",
      "current": 120,
      "expected": 100,
      "message": "P99延迟120ms超过阈值100ms"
    }
  ],
  
  "recommendations": [
    {
      "priority": "MEDIUM",
      "category": "OPTIMIZATION",
      "title": "优化响应延迟",
      "message": "延迟指标超出可接受范围，影响用户体验",
      "actions": [
        "分析慢查询日志，优化热点操作",
        "检查网络延迟和带宽瓶颈",
        "优化数据结构和查询模式",
        "考虑增加缓存层或读写分离",
        "评估硬件资源是否充足（CPU/内存/磁盘IO）"
      ],
      "references": [
        "https://redis.io/topics/latency"
      ]
    }
  ]
}
```

### 4.3 Markdown报告格式

详细的Markdown格式报告，适合归档和分享。

---

## 五、阶段划分（严格TDD流程）

### Phase 0 — 项目初始化与架构设计（1-2天）

**开发原则**: 测试先行，架构设计文档驱动

**任务目标**:
1. 定义项目结构和接口规范
2. 编写架构设计文档
3. 搭建开发环境和测试框架
4. 定义所有核心接口（不实现）

**交付物**:
```
docs/phase-0/
├── architecture.md           # 系统架构设计
├── interface-spec.md         # 接口规范定义
├── metrics-definition.md     # 指标详细定义
├── evaluation-criteria.md    # 评分标准详细说明
├── testing-strategy.md       # 测试策略
└── development-guide.md      # 开发环境搭建指南
```

**检查点 #0.1 - 架构文档评审**:
```bash
# 确认项
- [ ] 架构图完整（组件图、时序图、部署图）
- [ ] 接口定义清晰（所有公共接口有文档）
- [ ] 指标定义明确（每个指标有计算公式和阈值）
- [ ] 评分标准详细（包含各等级判定条件）
- [ ] 测试策略完备（单元测试、集成测试覆盖计划）

# 提交要求
git add docs/phase-0/
git commit -m "Phase 0.1: 完成架构设计文档"
git tag phase-0.1
```

**检查点 #0.2 - 项目脚手架搭建**:
```bash
# 项目结构
middleware-chaos-testing/
├── cmd/
│   └── mct/                  # 主程序入口
├── internal/
│   ├── core/                 # 核心抽象接口
│   ├── middleware/           # 中间件适配器
│   ├── metrics/              # 指标收集器
│   ├── detector/             # 稳定性检测器
│   ├── evaluator/            # 稳定性评估器 ⭐新增
│   └── reporter/             # 结果报告器
├── pkg/                      # 公共库
├── tests/                    # 测试文件
├── docs/                     # 文档
├── scripts/                  # 脚本
├── configs/                  # 配置文件
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── Makefile
└── README.md

# 确认项
- [ ] 目录结构符合Go项目规范
- [ ] Makefile包含常用命令（test, build, lint, run）
- [ ] Docker环境可一键启动
- [ ] 测试框架正常运行（go test ./... 可执行）

# 提交要求
git add .
git commit -m "Phase 0.2: 完成项目脚手架"
git tag phase-0.2
```

**检查点 #0.3 - 核心接口定义**:
```go
// internal/core/client.go
package core

import (
    "context"
    "time"
)

// MiddlewareClient 中间件客户端接口
type MiddlewareClient interface {
    Connect(ctx context.Context) error
    Disconnect(ctx context.Context) error
    Execute(ctx context.Context, op Operation) (*Result, error)
    HealthCheck(ctx context.Context) error
    GetMetrics() Metrics
}

// Evaluator 稳定性评估器接口 ⭐新增
type Evaluator interface {
    Evaluate(metrics *StabilityMetrics) *EvaluationResult
    SetThresholds(thresholds *Thresholds)
    GetDefaultThresholds() *Thresholds
}

// 测试文件: internal/core/evaluator_test.go
func TestEvaluatorInterface(t *testing.T) {
    var _ core.Evaluator = (*mockEvaluator)(nil)
}
```

**验收标准**:
- [ ] 所有核心接口定义完成并有文档注释
- [ ] 接口契约测试通过
- [ ] 架构设计通过团队评审
- [ ] 开发环境一键启动脚本可用

```bash
git add internal/core/
git commit -m "Phase 0.3: 完成核心接口定义"
git tag phase-0.3
```

---

### Phase 1 — Redis客户端实现（MVP核心 - 2-3天）

*(保持原有内容，已包含测试先行的TDD流程)*

**检查点 #1.1 - Redis客户端测试用例**
**检查点 #1.2 - Redis客户端实现**

**验收标准**:
- [ ] 运行 `go test ./tests/unit/middleware/redis_client_test.go -v` 全部通过
- [ ] 代码覆盖率 >= 85%
- [ ] 代码通过静态检查
- [ ] **支持通过参数指定测试持续时间**

```bash
git add internal/middleware/redis_client.go
git commit -m "Phase 1.2: 完成Redis客户端实现并通过所有测试"
git tag phase-1.2
```

---

### Phase 2 — Kafka客户端实现（MVP核心 - 2-3天）

*(保持原有内容)*

**验收标准**:
- [ ] 所有测试通过
- [ ] 代码覆盖率 >= 85%
- [ ] **支持通过参数指定测试持续时间**

```bash
git add internal/middleware/kafka_client.go
git commit -m "Phase 2.2: 完成Kafka客户端实现并通过所有测试"
git tag phase-2.2
```

---

### Phase 3 — 稳定性检测器实现（核心逻辑 - 3-4天）

**任务目标**:
1. 实现指标收集器
2. 实现稳定性分析引擎
3. 实现异常检测算法
4. 实现报告生成器

*(保持原有的Phase 3内容)*

---

### Phase 3.5 — 稳定性评分与判断系统（⭐核心新增 - 1-2天）

**任务目标**:
1. 实现稳定性评分算法（0-100分）
2. 定义健康度等级标准（5个等级）
3. 生成可操作的改进建议
4. 实现自动化判断逻辑（PASS/WARNING/FAIL）

**检查点 #3.5.1 - 评分系统测试用例**:
```go
// tests/unit/evaluator/stability_evaluator_test.go
package evaluator_test

import (
    "testing"
    "time"
    "github.com/stretchr/testify/suite"
)

type StabilityEvaluatorTestSuite struct {
    suite.Suite
    evaluator *StabilityEvaluator
}

func (suite *StabilityEvaluatorTestSuite) SetupTest() {
    suite.evaluator = NewStabilityEvaluator(nil) // 使用默认阈值
}

// 测试用例1：完美分数
func (suite *StabilityEvaluatorTestSuite) TestEvaluate_PerfectScore() {
    metrics := &core.StabilityMetrics{
        Availability:          0.9999,
        P95Latency:           10 * time.Millisecond,
        P99Latency:           20 * time.Millisecond,
        ErrorRate:            0.0001,
        DataLossRate:         0.0,
        MTTR:                 5 * time.Second,
        ReconnectSuccessRate: 0.99,
    }
    
    result := suite.evaluator.Evaluate(metrics)
    
    suite.Equal("EXCELLENT", result.Grade)
    suite.GreaterOrEqual(result.Score, 95.0)
    suite.Equal("PASS", result.Status)
    suite.Empty(result.Issues)
}

// 测试用例2：可用性不足导致失败
func (suite *StabilityEvaluatorTestSuite) TestEvaluate_LowAvailability_Fails() {
    metrics := &core.StabilityMetrics{
        Availability: 0.94,  // 低于95%最低标准
        P95Latency:   50 * time.Millisecond,
        ErrorRate:    0.06,
    }
    
    result := suite.evaluator.Evaluate(metrics)
    
    suite.Equal("FAILED", result.Grade)
    suite.Equal("FAIL", result.Status)
    suite.Less(result.Score, 60.0)
    
    // 验证问题列表
    suite.NotEmpty(result.Issues)
    suite.Equal("low_availability", result.Issues[0].Type)
    suite.Equal("CRITICAL", result.Issues[0].Severity)
}

// 测试用例3：高延迟触发警告
func (suite *StabilityEvaluatorTestSuite) TestEvaluate_HighLatency_Warning() {
    metrics := &core.StabilityMetrics{
        Availability: 0.999,
        P95Latency:   80 * time.Millisecond,  // 接近阈值
        P99Latency:   150 * time.Millisecond, // 超过good阈值
        ErrorRate:    0.001,
    }
    
    result := suite.evaluator.Evaluate(metrics)
    
    suite.Equal("GOOD", result.Grade)
    suite.Equal("WARNING", result.Status)
    suite.GreaterOrEqual(result.Score, 80.0)
    suite.Less(result.Score, 85.0)
}

// 测试用例4：生成详细建议
func (suite *StabilityEvaluatorTestSuite) TestEvaluate_GeneratesRecommendations() {
    metrics := &core.StabilityMetrics{
        Availability:       0.98,
        ErrorRate:          0.03,
        DataLossRate:       0.01,
        ConnectionFailRate: 0.05,
    }
    
    result := suite.evaluator.Evaluate(metrics)
    
    // 应该包含多条针对性建议
    suite.NotEmpty(result.Recommendations)
    suite.GreaterOrEqual(len(result.Recommendations), 3)
    
    // 验证建议的优先级排序（高优先级在前）
    suite.Equal("HIGH", result.Recommendations[0].Priority)
    
    // 验证建议包含具体行动项
    suite.NotEmpty(result.Recommendations[0].Actions)
    suite.GreaterOrEqual(len(result.Recommendations[0].Actions), 3)
}

// 测试用例5：自定义阈值
func (suite *StabilityEvaluatorTestSuite) TestEvaluate_CustomThresholds() {
    customThresholds := &Thresholds{
        AvailabilityPass: 0.90,  // 降低最低要求
        P95LatencyPass:   300 * time.Millisecond,
    }
    
    evaluator := NewStabilityEvaluator(customThresholds)
    
    metrics := &core.StabilityMetrics{
        Availability: 0.92,
        P95Latency:   250 * time.Millisecond,
    }
    
    result := evaluator.Evaluate(metrics)
    
    // 使用自定义阈值应该通过
    suite.NotEqual("FAIL", result.Status)
}

// 测试用例6：判断依据说明
func (suite *StabilityEvaluatorTestSuite) TestEvaluate_RationaleGeneration() {
    metrics := &core.StabilityMetrics{
        Availability: 0.999,
        P95Latency:   45 * time.Millisecond,
        ErrorRate:    0.001,
    }
    
    result := suite.evaluator.Evaluate(metrics)
    
    suite.NotEmpty(result.Rationale)
    suite.Contains(result.Rationale, "综合评分")
    suite.Contains(result.Rationale, "各维度得分")
    suite.Contains(result.Rationale, result.Grade)
}

// 测试用例7：Redis特定评分
func (suite *StabilityEvaluatorTestSuite) TestEvaluate_RedisSpecific() {
    metrics := &core.StabilityMetrics{
        Availability: 0.999,
        CacheHitRate: 0.85,  // 命中率偏低
        MemoryUsage:  0.95,  // 内存使用率高
    }
    
    result := suite.evaluator.EvaluateRedis(metrics)
    
    // 应该包含Redis特定建议
    hasRedisAdvice := false
    for _, rec := range result.Recommendations {
        if strings.Contains(rec.Title, "缓存") || 
           strings.Contains(rec.Title, "内存") {
            hasRedisAdvice = true
            break
        }
    }
    suite.True(hasRedisAdvice)
}

// 测试用例8：Kafka特定评分
func (suite *StabilityEvaluatorTestSuite) TestEvaluate_KafkaSpecific() {
    metrics := &core.StabilityMetrics{
        Availability:   0.999,
        MessageLag:     10000,  // 消息积压严重
        DuplicateRate:  0.02,   // 重复率偏高
        RebalanceCount: 5,      // 频繁重平衡
    }
    
    result := suite.evaluator.EvaluateKafka(metrics)
    
    suite.Equal("WARNING", result.Status)
    
    // 验证识别了Kafka特定问题
    hasKafkaIssue := false
    for _, issue := range result.Issues {
        if issue.Type == "high_message_lag" || 
           issue.Type == "frequent_rebalance" {
            hasKafkaIssue = true
            break
        }
    }
    suite.True(hasKafkaIssue)
}

// 测试用例9：边界条件 - 刚好及格
func (suite *StabilityEvaluatorTestSuite) TestEvaluate_BoundaryCase_JustPass() {
    metrics := &core.StabilityMetrics{
        Availability: 0.95,   // 刚好达到最低要求
        P95Latency:   200 * time.Millisecond,
        P99Latency:   500 * time.Millisecond,
        ErrorRate:    0.01,
    }
    
    result := suite.evaluator.Evaluate(metrics)
    
    suite.Equal("FAIR", result.Grade)
    suite.GreaterOrEqual(result.Score, 70.0)
    // 虽然及格，但应该有警告
    suite.Equal("WARNING", result.Status)
}

// 测试用例10：多维度问题综合评分
func (suite *StabilityEvaluatorTestSuite) TestEvaluate_MultipleIssues() {
    metrics := &core.StabilityMetrics{
        Availability:          0.97,   // 略低
        P95Latency:           120 * time.Millisecond, // 超标
        P99Latency:           600 * time.Millisecond, // 严重超标
        ErrorRate:            0.015,  // 超标
        DataLossRate:         0.005,  // 有数据丢失
        MTTR:                 350 * time.Second, // 恢复慢
        ReconnectSuccessRate: 0.88,   // 重连率低
    }
    
    result := suite.evaluator.Evaluate(metrics)
    
    // 多个维度有问题，总分应该较低
    suite.Less(result.Score, 70.0)
    suite.Equal("FAIL", result.Status)
    
    // 应该有多个问题
    suite.GreaterOrEqual(len(result.Issues), 4)
    
    // 应该有多条建议
    suite.GreaterOrEqual(len(result.Recommendations), 4)
}

func TestStabilityEvaluatorTestSuite(t *testing.T) {
    suite.Run(t, new(StabilityEvaluatorTestSuite))
}
```

**验收标准**:
- [ ] 所有测试用例编写完成（至少10个测试用例）
- [ ] 覆盖完美场景、及格场景、失败场景、边界条件
- [ ] 测试中间件特定评分逻辑
- [ ] 运行测试全部失败（红灯）

```bash
git add tests/unit/evaluator/stability_evaluator_test.go
git commit -m "Phase 3.5.1: 完成评分系统测试用例"
git tag phase-3.5.1
```

**检查点 #3.5.2 - 评分系统实现**:

```go
// internal/evaluator/stability_evaluator.go
package evaluator

import (
    "fmt"
    "sort"
    "strings"
    "time"
    "middleware-chaos-testing/internal/core"
)

// StabilityGrade 稳定性等级
type StabilityGrade string

const (
    GradeExcellent StabilityGrade = "EXCELLENT" // 优秀 (90-100分)
    GradeGood      StabilityGrade = "GOOD"      // 良好 (80-90分)
    GradeFair      StabilityGrade = "FAIR"      // 一般 (70-80分)
    GradePoor      StabilityGrade = "POOR"      // 较差 (60-70分)
    GradeFailed    StabilityGrade = "FAILED"    // 失败 (<60分)
)

// TestStatus 测试状态
type TestStatus string

const (
    StatusPass    TestStatus = "PASS"    // ✅ 通过
    StatusWarning TestStatus = "WARNING" // ⚠️  警告
    StatusFail    TestStatus = "FAIL"    // ❌ 失败
)

// EvaluationResult 评估结果
type EvaluationResult struct {
    Score       float64        `json:"score"`   // 总分 0-100
    Grade       StabilityGrade `json:"grade"`   // 等级
    Status      TestStatus     `json:"status"`  // 状态
    
    Scores struct {
        Availability float64 `json:"availability"` // 可用性得分 (30分)
        Performance  float64 `json:"performance"`  // 性能得分 (25分)
        Reliability  float64 `json:"reliability"`  // 可靠性得分 (25分)
        Resilience   float64 `json:"resilience"`   // 恢复力得分 (20分)
    } `json:"scores"`
    
    Issues          []Issue          `json:"issues"`
    Recommendations []Recommendation `json:"recommendations"`
    Rationale       string          `json:"rationale"`
    EvaluatedAt     time.Time       `json:"evaluated_at"`
}

// Issue 问题描述
type Issue struct {
    Type     string  `json:"type"`
    Severity string  `json:"severity"` // CRITICAL, HIGH, MEDIUM, LOW
    Metric   string  `json:"metric"`
    Current  float64 `json:"current"`
    Expected float64 `json:"expected"`
    Message  string  `json:"message"`
}

// Recommendation 改进建议
type Recommendation struct {
    Priority   string   `json:"priority"`   // HIGH, MEDIUM, LOW
    Category   string   `json:"category"`   // CONFIGURATION, SCALING, OPTIMIZATION
    Title      string   `json:"title"`
    Message    string   `json:"message"`
    Actions    []string `json:"actions"`
    References []string `json:"references,omitempty"`
}

// Thresholds 评分阈值
type Thresholds struct {
    AvailabilityExcellent float64       `yaml:"availability_excellent"`
    AvailabilityGood      float64       `yaml:"availability_good"`
    AvailabilityFair      float64       `yaml:"availability_fair"`
    AvailabilityPass      float64       `yaml:"availability_pass"`
    
    P95LatencyExcellent   time.Duration `yaml:"p95_latency_excellent"`
    P95LatencyGood        time.Duration `yaml:"p95_latency_good"`
    P95LatencyFair        time.Duration `yaml:"p95_latency_fair"`
    P95LatencyPass        time.Duration `yaml:"p95_latency_pass"`
    
    P99LatencyExcellent   time.Duration `yaml:"p99_latency_excellent"`
    P99LatencyGood        time.Duration `yaml:"p99_latency_good"`
    P99LatencyFair        time.Duration `yaml:"p99_latency_fair"`
    P99LatencyPass        time.Duration `yaml:"p99_latency_pass"`
    
    ErrorRateExcellent    float64       `yaml:"error_rate_excellent"`
    ErrorRateGood         float64       `yaml:"error_rate_good"`
    ErrorRateFair         float64       `yaml:"error_rate_fair"`
    ErrorRatePass         float64       `yaml:"error_rate_pass"`
    
    MTTRExcellent         time.Duration `yaml:"mttr_excellent"`
    MTTRGood              time.Duration `yaml:"mttr_good"`
    MTTRFair              time.Duration `yaml:"mttr_fair"`
    MTTRPass              time.Duration `yaml:"mttr_pass"`
}

// DefaultThresholds 返回默认阈值
func DefaultThresholds() *Thresholds {
    return &Thresholds{
        AvailabilityExcellent: 0.9999,
        AvailabilityGood:      0.999,
        AvailabilityFair:      0.99,
        AvailabilityPass:      0.95,
        
        P95LatencyExcellent:   10 * time.Millisecond,
        P95LatencyGood:        50 * time.Millisecond,
        P95LatencyFair:        100 * time.Millisecond,
        P95LatencyPass:        200 * time.Millisecond,
        
        P99LatencyExcellent:   20 * time.Millisecond,
        P99LatencyGood:        100 * time.Millisecond,
        P99LatencyFair:        200 * time.Millisecond,
        P99LatencyPass:        500 * time.Millisecond,
        
        ErrorRateExcellent:    0.0001,
        ErrorRateGood:         0.001,
        ErrorRateFair:         0.005,
        ErrorRatePass:         0.01,
        
        MTTRExcellent:         5 * time.Second,
        MTTRGood:              30 * time.Second,
        MTTRFair:              60 * time.Second,
        MTTRPass:              300 * time.Second,
    }
}

type StabilityEvaluator struct {
    thresholds *Thresholds
}

func NewStabilityEvaluator(thresholds *Thresholds) *StabilityEvaluator {
    if thresholds == nil {
        thresholds = DefaultThresholds()
    }
    return &StabilityEvaluator{thresholds: thresholds}
}

// Evaluate 评估稳定性指标
func (se *StabilityEvaluator) Evaluate(metrics *core.StabilityMetrics) *EvaluationResult {
    result := &EvaluationResult{
        EvaluatedAt:     time.Now(),
        Issues:          make([]Issue, 0),
        Recommendations: make([]Recommendation, 0),
    }
    
    // 计算各维度得分
    result.Scores.Availability = se.calculateAvailabilityScore(metrics, result)
    result.Scores.Performance = se.calculatePerformanceScore(metrics, result)
    result.Scores.Reliability = se.calculateReliabilityScore(metrics, result)
    result.Scores.Resilience = se.calculateResilienceScore(metrics, result)
    
    // 计算总分（加权）
    result.Score = result.Scores.Availability*0.30 +
                   result.Scores.Performance*0.25 +
                   result.Scores.Reliability*0.25 +
                   result.Scores.Resilience*0.20
    
    // 确定等级和状态
    result.Grade = se.determineGrade(result.Score)
    result.Status = se.determineStatus(metrics, result)
    
    // 生成建议和判断依据
    result.Recommendations = se.generateRecommendations(metrics, result)
    result.Rationale = se.generateRationale(result)
    
    return result
}

// calculateAvailabilityScore 计算可用性得分 (30分)
func (se *StabilityEvaluator) calculateAvailabilityScore(
    metrics *core.StabilityMetrics,
    result *EvaluationResult,
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
        result.Issues = append(result.Issues, Issue{
            Type:     "low_availability",
            Severity: "CRITICAL",
            Metric:   "availability",
            Current:  availability,
            Expected: se.thresholds.AvailabilityPass,
            Message:  fmt.Sprintf("可用性%.4f%%低于最低要求%.2f%%", 
                                 availability*100, 
                                 se.thresholds.AvailabilityPass*100),
        })
    }
    
    return score
}

// calculatePerformanceScore 计算性能得分 (25分)
func (se *StabilityEvaluator) calculatePerformanceScore(
    metrics *core.StabilityMetrics,
    result *EvaluationResult,
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
        result.Issues = append(result.Issues, Issue{
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
        result.Issues = append(result.Issues, Issue{
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

// calculateReliabilityScore 计算可靠性得分 (25分)
func (se *StabilityEvaluator) calculateReliabilityScore(
    metrics *core.StabilityMetrics,
    result *EvaluationResult,
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
        result.Issues = append(result.Issues, Issue{
            Type:     "high_error_rate",
            Severity: "HIGH",
            Metric:   "error_rate",
            Current:  errorRate * 100,
            Expected: se.thresholds.ErrorRatePass * 100,
            Message:  fmt.Sprintf("错误率%.4f%%超过阈值%.2f%%", 
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
        result.Issues = append(result.Issues, Issue{
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

// calculateResilienceScore 计算恢复力得分 (20分)
func (se *StabilityEvaluator) calculateResilienceScore(
    metrics *core.StabilityMetrics,
    result *EvaluationResult,
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
        result.Issues = append(result.Issues, Issue{
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
        result.Issues = append(result.Issues, Issue{
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
func (se *StabilityEvaluator) determineGrade(score float64) StabilityGrade {
    switch {
    case score >= 90:
        return GradeExcellent
    case score >= 80:
        return GradeGood
    case score >= 70:
        return GradeFair
    case score >= 60:
        return GradePoor
    default:
        return GradeFailed
    }
}

// determineStatus 确定状态
func (se *StabilityEvaluator) determineStatus(
    metrics *core.StabilityMetrics,
    result *EvaluationResult,
) TestStatus {
    // CRITICAL问题直接失败
    for _, issue := range result.Issues {
        if issue.Severity == "CRITICAL" {
            return StatusFail
        }
    }
    
    // 分数低于70失败
    if result.Score < 70 {
        return StatusFail
    }
    
    // HIGH问题为警告
    for _, issue := range result.Issues {
        if issue.Severity == "HIGH" {
            return StatusWarning
        }
    }
    
    // 分数低于85为警告
    if result.Score < 85 {
        return StatusWarning
    }
    
    return StatusPass
}

// generateRecommendations 生成建议
func (se *StabilityEvaluator) generateRecommendations(
    metrics *core.StabilityMetrics,
    result *EvaluationResult,
) []Recommendation {
    recommendations := make([]Recommendation, 0)
    
    for _, issue := range result.Issues {
        switch issue.Type {
        case "low_availability":
            recommendations = append(recommendations, Recommendation{
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
            recommendations = append(recommendations, Recommendation{
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
            recommendations = append(recommendations, Recommendation{
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
            recommendations = append(recommendations, Recommendation{
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
        }
    }
    
    // 按优先级排序
    sort.Slice(recommendations, func(i, j int) bool {
        priority := map[string]int{"HIGH": 3, "MEDIUM": 2, "LOW": 1}
        return priority[recommendations[i].Priority] > priority[recommendations[j].Priority]
    })
    
    return recommendations
}

// generateRationale 生成判断依据
func (se *StabilityEvaluator) generateRationale(result *EvaluationResult) string {
    var b strings.Builder
    
    b.WriteString(fmt.Sprintf("综合评分: %.2f/100 (%s)\n\n", result.Score, result.Grade))
    b.WriteString("各维度得分:\n")
    b.WriteString(fmt.Sprintf("- 可用性: %.2f/30 (权重30%%)\n", result.Scores.Availability))
    b.WriteString(fmt.Sprintf("- 性能: %.2f/25 (权重25%%)\n", result.Scores.Performance))
    b.WriteString(fmt.Sprintf("- 可靠性: %.2f/25 (权重25%%)\n", result.Scores.Reliability))
    b.WriteString(fmt.Sprintf("- 恢复力: %.2f/20 (权重20%%)\n\n", result.Scores.Resilience))
    
    switch result.Status {
    case StatusPass:
        b.WriteString("✅ 测试通过: 系统稳定性符合预期，可以用于生产环境。\n")
    case StatusWarning:
        b.WriteString("⚠️  警告: 系统存在需要关注的问题，建议优化后再部署。\n")
    case StatusFail:
        b.WriteString("❌ 测试失败: 系统稳定性不满足最低要求，不建议用于生产环境。\n")
    }
    
    if len(result.Issues) > 0 {
        b.WriteString(fmt.Sprintf("\n发现 %d 个问题需要处理。\n", len(result.Issues)))
    }
    
    return b.String()
}

// EvaluateRedis Redis特定评估（可选扩展）
func (se *StabilityEvaluator) EvaluateRedis(metrics *core.StabilityMetrics) *EvaluationResult {
    result := se.Evaluate(metrics)
    
    // 添加Redis特定检查
    if metrics.CacheHitRate < 0.90 {
        result.Recommendations = append(result.Recommendations, Recommendation{
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

// EvaluateKafka Kafka特定评估（可选扩展）
func (se *StabilityEvaluator) EvaluateKafka(metrics *core.StabilityMetrics) *EvaluationResult {
    result := se.Evaluate(metrics)
    
    // 添加Kafka特定检查
    if metrics.MessageLag > 1000 {
        result.Issues = append(result.Issues, Issue{
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
```

**验收标准**:
- [ ] 所有评分系统测试通过（绿灯）
- [ ] 代码覆盖率 >= 90%
- [ ] 能够生成清晰的等级评定（5个等级）
- [ ] 能够输出明确的测试状态（PASS/WARNING/FAIL）
- [ ] 每个问题都有对应的改进建议
- [ ] 建议按优先级排序

```bash
make test-evaluator
make coverage

git add internal/evaluator/
git commit -m "Phase 3.5.2: 完成稳定性评分系统实现"
git tag phase-3.5.2
```

---

### Phase 4 — 命令行工具（MVP交付 - 2天）

**任务目标**:
1. 实现CLI接口（⭐支持--duration参数）
2. 实现配置文件解析
3. 实现测试场景编排
4. **集成评分系统生成报告**

**检查点 #4.1 - CLI测试用例**:
```go
// tests/integration/cli_test.go
package integration_test

import (
    "os/exec"
    "testing"
    "encoding/json"
)

func TestCLI_RedisTest_WithDuration(t *testing.T) {
    cmd := exec.Command("mct", "test", 
        "--middleware", "redis",
        "--host", "localhost",
        "--port", "6379",
        "--duration", "30s",      // ⭐指定持续时间
        "--operations", "1000")
    
    output, err := cmd.CombinedOutput()
    
    assert.NoError(t, err)
    assert.Contains(t, string(output), "测试时长: 30s")
    assert.Contains(t, string(output), "总体评分:")
    assert.Contains(t, string(output), "PASS")
}

func TestCLI_OutputJSON_WithEvaluation(t *testing.T) {
    cmd := exec.Command("mct", "test",
        "--middleware", "redis",
        "--duration", "10s",
        "--output", "json")
    
    output, err := cmd.CombinedOutput()
    assert.NoError(t, err)
    
    // 解析JSON输出
    var result struct {
        Evaluation struct {
            Score  float64 `json:"score"`
            Grade  string  `json:"grade"`
            Status string  `json:"status"`
        } `json:"evaluation"`
    }
    
    err = json.Unmarshal(output, &result)
    assert.NoError(t, err)
    assert.NotEmpty(t, result.Evaluation.Grade)
    assert.NotEmpty(t, result.Evaluation.Status)
}
```

**检查点 #4.2 - CLI实现**:
```go
// cmd/mct/main.go
package main

import (
    "context"
    "fmt"
    "os"
    "time"
    
    "github.com/spf13/cobra"
    "middleware-chaos-testing/internal/core"
    "middleware-chaos-testing/internal/evaluator"
    "middleware-chaos-testing/internal/reporter"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "mct",
        Short: "Middleware Chaos Testing Tool",
    }
    
    testCmd := &cobra.Command{
        Use:   "test",
        Short: "Run stability test",
        Run:   runTest,
    }
    
    // ⭐添加duration参数
    testCmd.Flags().String("middleware", "", "Middleware type (redis|kafka)")
    testCmd.Flags().String("host", "localhost", "Host")
    testCmd.Flags().Int("port", 0, "Port")
    testCmd.Flags().Duration("duration", 60*time.Second, "Test duration")
    testCmd.Flags().Int("operations", 10000, "Number of operations")
    testCmd.Flags().String("output", "console", "Output format (console|json|markdown)")
    testCmd.Flags().String("report-path", "./reports", "Report output path")
    
    rootCmd.AddCommand(testCmd)
    
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

func runTest(cmd *cobra.Command, args []string) {
    // 获取参数
    duration, _ := cmd.Flags().GetDuration("duration")
    outputFormat, _ := cmd.Flags().GetString("output")
    
    // 执行测试
    ctx, cancel := context.WithTimeout(context.Background(), duration+30*time.Second)
    defer cancel()
    
    metrics := executeTest(ctx, duration)
    
    // ⭐评分
    eval := evaluator.NewStabilityEvaluator(nil)
    result := eval.Evaluate(metrics)
    
    // 生成报告
    switch outputFormat {
    case "json":
        reporter.GenerateJSONReport(result, os.Stdout)
    case "markdown":
        reporter.GenerateMarkdownReport(result, os.Stdout)
    default:
        reporter.GenerateConsoleReport(result, os.Stdout)
    }
    
    // ⭐根据测试状态设置退出码
    if result.Status == "FAIL" {
        os.Exit(1)
    } else if result.Status == "WARNING" {
        os.Exit(2)
    }
    os.Exit(0)
}
```

**验收标准**:
- [ ] **CLI支持--duration参数指定测试时长**
- [ ] **报告包含明确的评分和判断**
- [ ] **输出格式支持console/json/markdown**
- [ ] **测试失败时返回非0退出码**
- [ ] 集成测试全部通过

```bash
make build
./bin/mct test --middleware redis --duration 30s --output console

git add cmd/
git commit -m "Phase 4.2: 完成CLI工具并集成评分系统"
git tag phase-4.2-mvp
```

**🎉 MVP验收检查点**:
```bash
# MVP功能验收清单
- [ ] ✅ Redis客户端完整功能
- [ ] ✅ Kafka客户端完整功能
- [ ] ✅ 核心稳定性指标收集
- [ ] ✅ 异常检测功能
- [ ] ✅ 智能评分系统 (0-100分)
- [ ] ✅ 明确的PASS/WARNING/FAIL判断
- [ ] ✅ 可操作的改进建议
- [ ] ✅ CLI工具支持--duration参数
- [ ] ✅ 生成结构化测试报告
- [ ] ✅ 单元测试覆盖率 >= 85%
- [ ] ✅ 集成测试通过
- [ ] ✅ 文档完善

# 报告质量验收
- [ ] ✅ 报告包含总分和等级
- [ ] ✅ 报告包含各维度得分
- [ ] ✅ 报告包含明确的通过/警告/失败状态
- [ ] ✅ 报告包含问题列表（含严重程度）
- [ ] ✅ 报告包含改进建议（按优先级）
- [ ] ✅ 报告包含判断依据说明

# 性能验收
- [ ] 支持1000+ ops/s测试负载
- [ ] 内存占用 < 100MB
- [ ] CPU占用 < 50%

# MVP交付
git tag v0.1.0-mvp
```

---

## 六、配置文件示例

### 6.1 Redis测试配置（含评分阈值）
```yaml
# configs/test-redis.yaml
name: "Redis Stability Test"
middleware: "redis"

connection:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  timeout: 5s

test:
  duration: 60s          # ⭐测试持续时间
  operations: 10000
  concurrency: 10
  
  workload:
    - operation: "set"
      weight: 40
      key_pattern: "test:key:{id}"
      value_size: 1024
      
    - operation: "get"
      weight: 50
      key_pattern: "test:key:{id}"
      
    - operation: "delete"
      weight: 10
      key_pattern: "test:key:{id}"

# ⭐评分阈值配置（可选，不配置使用默认值）
thresholds:
  availability:
    excellent: 99.99%
    good: 99.9%
    fair: 99.0%
    pass: 95.0%
    
  p95_latency:
    excellent: 10ms
    good: 50ms
    fair: 100ms
    pass: 200ms
    
  p99_latency:
    excellent: 20ms
    good: 100ms
    fair: 200ms
    pass: 500ms
    
  error_rate:
    excellent: 0.01%
    good: 0.1%
    fair: 0.5%
    pass: 1.0%

output:
  format: "console"    # console, json, markdown
  path: "./reports/redis-test-{timestamp}.json"
  include_recommendations: true  # ⭐包含改进建议
```

---

## 七、项目结构（最终版）

```
middleware-chaos-testing/
├── cmd/
│   └── mct/
│       └── main.go                 # CLI入口（支持--duration）
│
├── internal/
│   ├── core/                       # 核心接口
│   │   ├── client.go
│   │   ├── evaluator.go           # ⭐评估器接口
│   │   └── types.go
│   │
│   ├── middleware/                 # 中间件适配器
│   │   ├── redis_client.go
│   │   ├── kafka_client.go
│   │   └── factory.go
│   │
│   ├── metrics/                    # 指标收集
│   │   ├── collector.go
│   │   └── calculator.go
│   │
│   ├── detector/                   # 异常检测
│   │   └── anomaly_detector.go
│   │
│   ├── evaluator/                  # ⭐稳定性评估
│   │   ├── stability_evaluator.go # 评分系统
│   │   ├── thresholds.go          # 阈值配置
│   │   └── recommendations.go     # 建议生成
│   │
│   ├── reporter/                   # 报告生成
│   │   ├── console_reporter.go    # ⭐美化控制台输出
│   │   ├── json_reporter.go       # JSON报告
│   │   └── markdown_reporter.go   # Markdown报告
│   │
│   ├── orchestrator/               # 测试编排
│   │   └── test_runner.go
│   │
│   └── config/                     # 配置管理
│       └── config.go
│
├── tests/
│   ├── unit/
│   │   ├── middleware/
│   │   ├── metrics/
│   │   ├── evaluator/             # ⭐评分系统测试
│   │   └── detector/
│   │
│   └── integration/
│       ├── cli_test.go            # ⭐包含duration测试
│       └── e2e_test.go
│
├── configs/                        # 配置文件
│   ├── test-redis.yaml            # ⭐含阈值配置
│   └── test-kafka.yaml
│
├── docs/                           # 文档
│   ├── phase-0/
│   │   └── evaluation-criteria.md # ⭐评分标准文档
│   └── user-guide/
│       └── interpreting-results.md # ⭐结果解读指南
│
├── Dockerfile
├── docker-compose.yml
├── Makefile                        # ⭐新增test-evaluator命令
├── go.mod
└── README.md                       # ⭐更新功能说明
```

---

## 八、Makefile命令（更新）

```makefile
# Makefile
.PHONY: help build test clean docker

help:
	@echo "Available commands:"
	@echo "  make build              - Build the binary"
	@echo "  make test               - Run all tests"
	@echo "  make test-unit          - Run unit tests"
	@echo "  make test-evaluator     - ⭐Run evaluator tests"
	@echo "  make coverage           - Generate coverage report"
	@echo "  make demo-redis         - ⭐Demo: Redis test with evaluation"
	@echo "  make demo-kafka         - ⭐Demo: Kafka test with evaluation"

build:
	go build -o bin/mct cmd/mct/main.go

test:
	go test -v ./...

test-evaluator:
	go test -v ./tests/unit/evaluator/... -cover

# ⭐Demo命令：展示完整的评分报告
demo-redis:
	@echo "Running Redis stability test with evaluation..."
	./bin/mct test \
		--middleware redis \
		--host localhost \
		--port 6379 \
		--duration 30s \
		--operations 5000 \
		--output console

demo-kafka:
	@echo "Running Kafka stability test with evaluation..."
	./bin/mct test \
		--middleware kafka \
		--brokers localhost:9092 \
		--duration 30s \
		--output console

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

docker-build:
	docker build -t mct:latest .

docker-run:
	docker-compose up
```

---

## 九、快速启动示例

```bash
# 1. 克隆并构建
git clone https://github.com/username/middleware-chaos-testing.git
cd middleware-chaos-testing
make build

# 2. 启动测试环境
docker-compose up -d redis kafka

# 3. 运行Redis测试（30秒）⭐指定持续时间
./bin/mct test \
  --middleware redis \
  --host localhost \
  --port 6379 \
  --duration 30s \
  --operations 5000 \
  --output console

# 输出示例：
==========================================
   中间件稳定性测试报告
==========================================
测试目标: Redis @ localhost:6379
测试时长: 30s                    # ⭐显示实际运行时间
测试完成: 2025-10-30 14:30:00

------------------------------------------
  总体评分: 87.5/100 (GOOD) ✅ PASS    # ⭐明确的评分和判断
------------------------------------------
各维度得分:
  ✓ 可用性   28.5/30 (95.0%)
  ✓ 性能     21.0/25 (84.0%)
  ...

------------------------------------------
  改进建议 (1条)                          # ⭐可操作的建议
------------------------------------------
[MEDIUM] 优化响应延迟
具体行动:
  1. 分析慢查询日志
  2. 检查网络延迟
  ...

# 4. 生成JSON报告
./bin/mct test \
  --middleware redis \
  --duration 60s \
  --output json > report.json

# 5. 检查退出码
./bin/mct test --middleware redis --duration 10s
echo $?  # 0=PASS, 1=FAIL, 2=WARNING    # ⭐明确的退出码
```

---

## 十、成功指标（更新）

### MVP成功指标
- [x] 支持Redis和Kafka两种中间件
- [x] **支持--duration参数指定测试时长** ⭐
- [x] **实现0-100分智能评分系统** ⭐
- [x] **明确的PASS/WARNING/FAIL判断** ⭐
- [x] **生成可操作的改进建议** ⭐
- [x] 单元测试覆盖率 >= 85%
- [x] CLI工具可正常使用
- [x] Docker一键启动

### 报告质量指标 ⭐新增
- [x] 报告包含总分和等级（5级）
- [x] 报告包含各维度得分明细
- [x] 报告包含测试状态（通过/警告/失败）
- [x] 报告包含问题列表（含严重程度）
- [x] 报告包含改进建议（按优先级排序）
- [x] 报告包含判断依据说明
- [x] 支持3种输出格式（console/json/markdown）

---

## 附录：评分算法说明

### 评分公式
```
总分 = 可用性得分 × 30% + 性能得分 × 25% + 可靠性得分 × 25% + 恢复力得分 × 20%

可用性得分 (30分):
  - >= 99.99%: 30分
  - >= 99.9%:  27分
  - >= 99%:    24分
  - >= 95%:    20分
  - < 95%:     按比例计算，并标记CRITICAL

性能得分 (25分) = P95得分(15分) + P99得分(10分)
  P95得分:
    - <= 10ms:  15分
    - <= 50ms:  13.5分
    - <= 100ms: 12分
    - <= 200ms: 10分
    - > 200ms:  8分，标记HIGH

可靠性得分 (25分) = 错误率得分(15分) + 数据丢失得分(10分)
  错误率得分:
    - <= 0.01%: 15分
    - <= 0.1%:  13.5分
    - <= 0.5%:  12分
    - <= 1%:    10分
    - > 1%:     7分，标记HIGH

恢复力得分 (20分) = MTTR得分(12分) + 重连率得分(8分)
  MTTR得分:
    - <= 5s:   12分
    - <= 30s:  10.5分
    - <= 60s:  9分
    - <= 300s: 7分
    - > 300s:  5分，标记MEDIUM
```

### 状态判断逻辑
```go
if 存在任何CRITICAL问题:
    return FAIL
    
if 总分 < 70:
    return FAIL
    
if 存在任何HIGH问题:
    return WARNING
    
if 总分 < 85:
    return WARNING
    
return PASS
```

---

## 十一、中间件扩展计划（Phase 6-10）

### 扩展目标
在MVP基础上扩展支持5种常用中间件，构建完整的中间件混沌测试生态系统。

### 扩展中间件列表
1. **MongoDB** - 文档数据库（NoSQL）
2. **RocketMQ** - 消息队列（阿里云）
3. **RabbitMQ** - 消息队列（AMQP协议）
4. **EMQX** - MQTT消息中间件（物联网）
5. **Nacos** - 服务注册与配置中心

---

### Phase 6 — MongoDB客户端支持（文档数据库 - 2-3天）

**中间件特性**:
- **类型**: NoSQL文档数据库
- **核心操作**: Insert, Find, Update, Delete, Aggregate
- **特性**: 分片、副本集、事务支持

**MongoDB特定指标**:
```go
// MongoDB特定指标
type MongoDBMetrics struct {
    // 性能指标
    InsertLatency   Percentiles  // 插入延迟
    QueryLatency    Percentiles  // 查询延迟
    UpdateLatency   Percentiles  // 更新延迟

    // 吞吐量
    InsertOPS       float64      // 插入操作/秒
    QueryOPS        float64      // 查询操作/秒

    // 可靠性
    DocumentLossRate   float64   // 文档丢失率
    InconsistencyRate  float64   // 数据不一致率

    // MongoDB特定
    ConnectionPoolSize int       // 连接池大小
    ReplicaLag        time.Duration // 副本延迟
    IndexHitRate      float64   // 索引命中率
    CursorTimeout     int       // 游标超时次数
}
```

**性能阈值（业界标准）**:
```go
func MongoDBThresholds() *core.Thresholds {
    return &core.Thresholds{
        // MongoDB延迟标准
        P95LatencyExcellent: 20 * time.Millisecond,  // 单文档读写
        P95LatencyGood:      50 * time.Millisecond,
        P95LatencyFair:      100 * time.Millisecond,
        P95LatencyPass:      200 * time.Millisecond,

        P99LatencyExcellent: 50 * time.Millisecond,
        P99LatencyGood:      100 * time.Millisecond,
        P99LatencyFair:      200 * time.Millisecond,
        P99LatencyPass:      500 * time.Millisecond,

        // 可用性标准
        AvailabilityExcellent: 0.9999,  // 99.99%
        AvailabilityGood:      0.999,   // 99.9%
        AvailabilityFair:      0.99,    // 99%
        AvailabilityPass:      0.95,    // 95%

        // 错误率标准
        ErrorRateExcellent: 0.0001,
        ErrorRateGood:      0.001,
        ErrorRateFair:      0.01,
        ErrorRatePass:      0.05,
    }
}
```

**检查点 #6.1 - MongoDB客户端测试用例**:
```bash
# tests/unit/middleware/mongodb_client_test.go
- TestMongoDBClient_Connect
- TestMongoDBClient_Insert
- TestMongoDBClient_Find
- TestMongoDBClient_Update
- TestMongoDBClient_Delete
- TestMongoDBClient_Aggregate
- TestMongoDBClient_ConnectionPool
- TestMongoDBClient_ReplicaSetFailover

git add tests/unit/middleware/mongodb_client_test.go
git commit -m "Phase 6.1: 完成MongoDB客户端测试用例"
git tag phase-6.1
```

**检查点 #6.2 - MongoDB客户端实现**:
```go
// internal/middleware/mongodb_types.go
type MongoDBConfig struct {
    URI              string        // mongodb://user:pass@host:port/db
    Database         string
    Collection       string
    Timeout          time.Duration

    // 连接池配置
    MaxPoolSize      int           // 默认: 100
    MinPoolSize      int           // 默认: 10
    MaxIdleTime      time.Duration // 默认: 10分钟

    // 副本集配置
    ReplicaSet       string
    ReadPreference   string        // primary, secondary, nearest
    WriteConcern     string        // majority, w1, w2

    // 性能优化
    Compressors      []string      // snappy, zlib, zstd
}

// internal/middleware/mongodb_client.go
type MongoDBClient struct {
    config   *MongoDBConfig
    client   *mongo.Client
    database *mongo.Database
    logger   *Logger
}

func (m *MongoDBClient) Execute(ctx context.Context, op core.Operation) (*core.Result, error)
```

**验收标准**:
- [ ] 所有测试通过
- [ ] 支持CRUD和聚合操作
- [ ] 连接池管理正确
- [ ] 副本集支持
- [ ] 完整日志记录
- [ ] 代码覆盖率 >= 85%

```bash
git add internal/middleware/mongodb_*
git commit -m "Phase 6.2: 完成MongoDB客户端实现"
git tag phase-6.2
```

---

### Phase 7 — RocketMQ客户端支持（消息队列 - 2-3天）

**中间件特性**:
- **类型**: 分布式消息中间件（阿里云开源）
- **核心操作**: SendMessage, ConsumeMessage, SendBatch, Transaction
- **特性**: 顺序消息、事务消息、延迟消息

**RocketMQ特定指标**:
```go
type RocketMQMetrics struct {
    // 性能指标
    ProducerLatency   Percentiles  // 生产延迟
    ConsumerLatency   Percentiles  // 消费延迟

    // 吞吐量
    ProducerTPS       float64      // 生产TPS
    ConsumerTPS       float64      // 消费TPS

    // 消息可靠性
    MessageLossRate      float64   // 消息丢失率
    MessageDuplicateRate float64   // 消息重复率
    MessageOrderRate     float64   // 顺序消息正确率

    // RocketMQ特定
    MessageAccumulation  int64     // 消息堆积量
    ConsumerLag          time.Duration // 消费延迟
    RebalanceCount       int       // 重平衡次数
    ConsumeRetryCount    int       // 消费重试次数
}
```

**性能阈值（业界标准）**:
```go
func RocketMQThresholds() *core.Thresholds {
    return &core.Thresholds{
        // RocketMQ延迟标准（类似Kafka但更严格）
        P95LatencyExcellent: 5 * time.Millisecond,   // 阿里云生产环境
        P95LatencyGood:      20 * time.Millisecond,
        P95LatencyFair:      50 * time.Millisecond,
        P95LatencyPass:      100 * time.Millisecond,

        P99LatencyExcellent: 10 * time.Millisecond,
        P99LatencyGood:      50 * time.Millisecond,
        P99LatencyFair:      100 * time.Millisecond,
        P99LatencyPass:      200 * time.Millisecond,

        // 高可用性要求
        AvailabilityExcellent: 0.99999, // 99.999% (5个9)
        AvailabilityGood:      0.9999,  // 99.99%
        AvailabilityFair:      0.999,   // 99.9%
        AvailabilityPass:      0.99,    // 99%
    }
}
```

**检查点 #7.1 - RocketMQ客户端测试用例**:
```bash
# tests/unit/middleware/rocketmq_client_test.go
- TestRocketMQClient_Connect
- TestRocketMQClient_SendMessage
- TestRocketMQClient_SendBatch
- TestRocketMQClient_ConsumeMessage
- TestRocketMQClient_TransactionMessage
- TestRocketMQClient_DelayMessage
- TestRocketMQClient_OrderedMessage
- TestRocketMQClient_MessageRetry

git add tests/unit/middleware/rocketmq_client_test.go
git commit -m "Phase 7.1: 完成RocketMQ客户端测试用例"
git tag phase-7.1
```

**检查点 #7.2 - RocketMQ客户端实现**:
```go
// internal/middleware/rocketmq_types.go
type RocketMQConfig struct {
    NameServers      []string      // NameServer地址列表
    Topic            string
    ProducerGroup    string
    ConsumerGroup    string
    Protocol         string        // "remoting" 或 "grpc"，默认: "remoting" (RocketMQ 5.1.3)

    // 生产者配置
    SendMsgTimeout   time.Duration // 默认: 3s
    RetryTimes       int           // 默认: 3
    CompressLevel    int           // 压缩级别
    MaxMessageSize   int           // 默认: 4MB

    // 消费者配置
    ConsumeMode      string        // CLUSTERING, BROADCASTING
    MessageModel     string        // ORDERED, CONCURRENT
    PullBatchSize    int           // 默认: 32
    ConsumeTimeout   time.Duration // 默认: 15分钟
}

// RocketMQ 5.1.3 协议常量
const (
    ProtocolRemoting = "remoting"  // 传统remoting协议（默认）
    ProtocolGRPC     = "grpc"      // gRPC协议（推荐用于5.x）
)

// internal/middleware/rocketmq_client.go
type RocketMQClient struct {
    config   *RocketMQConfig
    producer rocketmq.Producer
    consumer rocketmq.PushConsumer
    logger   *Logger
}
```

**验收标准**:
- [ ] 所有测试通过
- [ ] 支持普通/批量/事务/延迟/顺序消息
- [ ] 生产者和消费者正常工作
- [ ] 消息可靠性保证
- [ ] 完整日志记录
- [ ] 代码覆盖率 >= 85%

```bash
git add internal/middleware/rocketmq_*
git commit -m "Phase 7.2: 完成RocketMQ客户端实现"
git tag phase-7.2
```

---

### Phase 8 — RabbitMQ客户端支持（消息队列 - 2-3天）

**中间件特性**:
- **类型**: AMQP协议消息中间件
- **核心操作**: Publish, Consume, Ack, Reject
- **特性**: 交换机路由、死信队列、TTL

**RabbitMQ特定指标**:
```go
type RabbitMQMetrics struct {
    // 性能指标
    PublishLatency   Percentiles  // 发布延迟
    ConsumeLatency   Percentiles  // 消费延迟

    // 吞吐量
    PublishRate      float64      // 发布速率
    ConsumeRate      float64      // 消费速率

    // 消息可靠性
    MessageLossRate     float64   // 消息丢失率
    MessageRedeliveryRate float64 // 消息重新投递率
    UnackedMessages     int       // 未确认消息数

    // RabbitMQ特定
    QueueDepth          int       // 队列深度
    ConsumerCount       int       // 消费者数量
    MemoryUsage         int64     // 内存使用（字节）
    ConnectionCount     int       // 连接数
    ChannelCount        int       // 通道数
}
```

**性能阈值（业界标准）**:
```go
func RabbitMQThresholds() *core.Thresholds {
    return &core.Thresholds{
        // RabbitMQ延迟标准
        P95LatencyExcellent: 10 * time.Millisecond,
        P95LatencyGood:      30 * time.Millisecond,
        P95LatencyFair:      100 * time.Millisecond,
        P95LatencyPass:      200 * time.Millisecond,

        P99LatencyExcellent: 20 * time.Millisecond,
        P99LatencyGood:      50 * time.Millisecond,
        P99LatencyFair:      150 * time.Millisecond,
        P99LatencyPass:      300 * time.Millisecond,

        // 可用性标准
        AvailabilityExcellent: 0.9999,
        AvailabilityGood:      0.999,
        AvailabilityFair:      0.99,
        AvailabilityPass:      0.95,
    }
}
```

**检查点 #8.1 - RabbitMQ客户端测试用例**:
```bash
# tests/unit/middleware/rabbitmq_client_test.go
- TestRabbitMQClient_Connect
- TestRabbitMQClient_Publish
- TestRabbitMQClient_Consume
- TestRabbitMQClient_Ack
- TestRabbitMQClient_Reject
- TestRabbitMQClient_ExchangeRouting
- TestRabbitMQClient_DeadLetterQueue
- TestRabbitMQClient_TTL

git add tests/unit/middleware/rabbitmq_client_test.go
git commit -m "Phase 8.1: 完成RabbitMQ客户端测试用例"
git tag phase-8.1
```

**检查点 #8.2 - RabbitMQ客户端实现**:
```go
// internal/middleware/rabbitmq_types.go
type RabbitMQConfig struct {
    URL              string        // amqp://user:pass@host:port/vhost
    Exchange         string
    ExchangeType     string        // direct, fanout, topic, headers
    Queue            string
    RoutingKey       string

    // 连接配置
    ConnectionTimeout time.Duration // 默认: 30s
    Heartbeat        time.Duration // 默认: 10s
    ChannelMax       int           // 默认: 0 (无限制)

    // QoS配置
    PrefetchCount    int           // 默认: 50
    PrefetchSize     int           // 默认: 0

    // 消息配置
    Persistent       bool          // 消息持久化
    Mandatory        bool          // 强制路由
    Immediate        bool          // 立即投递
}

// internal/middleware/rabbitmq_client.go
type RabbitMQClient struct {
    config     *RabbitMQConfig
    connection *amqp.Connection
    channel    *amqp.Channel
    logger     *Logger
}
```

**验收标准**:
- [ ] 所有测试通过
- [ ] 支持发布/订阅模式
- [ ] 支持各种交换机类型
- [ ] 消息确认机制正确
- [ ] 完整日志记录
- [ ] 代码覆盖率 >= 85%

```bash
git add internal/middleware/rabbitmq_*
git commit -m "Phase 8.2: 完成RabbitMQ客户端实现"
git tag phase-8.2
```

---

### Phase 9 — EMQX客户端支持（MQTT中间件 - 2-3天）

**中间件特性**:
- **类型**: MQTT消息中间件（物联网）
- **核心操作**: Connect, Publish, Subscribe, Unsubscribe
- **特性**: QoS级别、保留消息、遗嘱消息

**EMQX特定指标**:
```go
type EMQXMetrics struct {
    // 性能指标
    PublishLatency    Percentiles  // 发布延迟
    SubscribeLatency  Percentiles  // 订阅延迟

    // 吞吐量
    PublishRate       float64      // 发布速率 (msg/s)
    SubscribeRate     float64      // 订阅速率 (msg/s)

    // 连接指标
    ConnectionRate    float64      // 连接建立速率
    DisconnectionRate float64      // 断连率
    ReconnectCount    int          // 重连次数

    // MQTT特定
    QoS0MessageCount  int64        // QoS0消息数
    QoS1MessageCount  int64        // QoS1消息数
    QoS2MessageCount  int64        // QoS2消息数
    RetainedMessages  int          // 保留消息数
    SubscriptionCount int          // 订阅数
    SessionCount      int          // 会话数
}
```

**性能阈值（业界标准）**:
```go
func EMQXThresholds() *core.Thresholds {
    return &core.Thresholds{
        // MQTT/EMQX延迟标准（物联网场景）
        P95LatencyExcellent: 50 * time.Millisecond,  // 物联网可接受
        P95LatencyGood:      100 * time.Millisecond,
        P95LatencyFair:      200 * time.Millisecond,
        P95LatencyPass:      500 * time.Millisecond,

        P99LatencyExcellent: 100 * time.Millisecond,
        P99LatencyGood:      200 * time.Millisecond,
        P99LatencyFair:      500 * time.Millisecond,
        P99LatencyPass:      1 * time.Second,

        // 高可用性（物联网场景）
        AvailabilityExcellent: 0.9999,
        AvailabilityGood:      0.999,
        AvailabilityFair:      0.99,
        AvailabilityPass:      0.95,

        // 错误率（物联网弱网络）
        ErrorRateExcellent: 0.001,  // 0.1%
        ErrorRateGood:      0.01,   // 1%
        ErrorRateFair:      0.05,   // 5%
        ErrorRatePass:      0.1,    // 10%
    }
}
```

**检查点 #9.1 - EMQX客户端测试用例**:
```bash
# tests/unit/middleware/emqx_client_test.go
- TestEMQXClient_Connect
- TestEMQXClient_Publish
- TestEMQXClient_Subscribe
- TestEMQXClient_Unsubscribe
- TestEMQXClient_QoS0
- TestEMQXClient_QoS1
- TestEMQXClient_QoS2
- TestEMQXClient_RetainedMessage
- TestEMQXClient_LastWill
- TestEMQXClient_CleanSession

git add tests/unit/middleware/emqx_client_test.go
git commit -m "Phase 9.1: 完成EMQX客户端测试用例"
git tag phase-9.1
```

**检查点 #9.2 - EMQX客户端实现**:
```go
// internal/middleware/emqx_types.go
type EMQXConfig struct {
    Broker           string        // tcp://host:port 或 ssl://host:port
    ClientID         string
    Username         string
    Password         string

    // MQTT配置
    ProtocolVersion  byte          // 3 (MQTT 3.1), 4 (MQTT 3.1.1), 5 (MQTT 5.0)
    CleanSession     bool          // 默认: true
    KeepAlive        int           // 默认: 60秒

    // TLS配置
    UseTLS           bool
    CACert           string
    ClientCert       string
    ClientKey        string

    // 性能配置
    MaxReconnectInterval time.Duration // 默认: 10分钟
    ConnectTimeout       time.Duration // 默认: 30s
    WriteTimeout         time.Duration // 默认: 30s
}

// internal/middleware/emqx_client.go
type EMQXClient struct {
    config   *EMQXConfig
    client   mqtt.Client
    logger   *Logger
}
```

**验收标准**:
- [ ] 所有测试通过
- [ ] 支持MQTT 3.1.1和5.0协议
- [ ] 支持QoS 0/1/2
- [ ] 支持保留消息和遗嘱消息
- [ ] TLS/SSL连接支持
- [ ] 完整日志记录
- [ ] 代码覆盖率 >= 85%

```bash
git add internal/middleware/emqx_*
git commit -m "Phase 9.2: 完成EMQX客户端实现"
git tag phase-9.2
```

---

### Phase 10 — Nacos客户端支持（服务发现与配置 - 2-3天）

**中间件特性**:
- **类型**: 服务注册与配置中心
- **核心操作**: RegisterInstance, DeregisterInstance, GetConfig, PublishConfig
- **特性**: 服务健康检查、配置热更新、命名空间隔离

**Nacos特定指标**:
```go
type NacosMetrics struct {
    // 性能指标
    RegisterLatency    Percentiles  // 注册延迟
    DiscoveryLatency   Percentiles  // 服务发现延迟
    ConfigReadLatency  Percentiles  // 配置读取延迟
    ConfigWriteLatency Percentiles  // 配置写入延迟

    // 服务注册
    ServiceCount       int          // 服务数量
    InstanceCount      int          // 实例数量
    HealthyInstances   int          // 健康实例数
    UnhealthyInstances int          // 不健康实例数

    // 配置管理
    ConfigCount        int          // 配置项数量
    ConfigUpdateRate   float64      // 配置更新速率
    ConfigPushLatency  time.Duration // 配置推送延迟

    // Nacos特定
    SubscriptionCount  int          // 订阅数
    HeartbeatSuccess   int64        // 心跳成功数
    HeartbeatFail      int64        // 心跳失败数
}
```

**性能阈值（业界标准）**:
```go
func NacosThresholds() *core.Thresholds {
    return &core.Thresholds{
        // Nacos延迟标准（服务发现场景）
        P95LatencyExcellent: 10 * time.Millisecond,  // 服务注册/发现
        P95LatencyGood:      50 * time.Millisecond,
        P95LatencyFair:      100 * time.Millisecond,
        P95LatencyPass:      200 * time.Millisecond,

        P99LatencyExcellent: 20 * time.Millisecond,
        P99LatencyGood:      100 * time.Millisecond,
        P99LatencyFair:      200 * time.Millisecond,
        P99LatencyPass:      500 * time.Millisecond,

        // 高可用性（注册中心关键）
        AvailabilityExcellent: 0.99999, // 99.999%
        AvailabilityGood:      0.9999,  // 99.99%
        AvailabilityFair:      0.999,   // 99.9%
        AvailabilityPass:      0.99,    // 99%

        // 低错误率
        ErrorRateExcellent: 0.0001,
        ErrorRateGood:      0.001,
        ErrorRateFair:      0.01,
        ErrorRatePass:      0.05,
    }
}
```

**检查点 #10.1 - Nacos客户端测试用例**:
```bash
# tests/unit/middleware/nacos_client_test.go
- TestNacosClient_Connect
- TestNacosClient_RegisterInstance
- TestNacosClient_DeregisterInstance
- TestNacosClient_GetService
- TestNacosClient_Subscribe
- TestNacosClient_GetConfig
- TestNacosClient_PublishConfig
- TestNacosClient_RemoveConfig
- TestNacosClient_ListenConfig
- TestNacosClient_Heartbeat

git add tests/unit/middleware/nacos_client_test.go
git commit -m "Phase 10.1: 完成Nacos客户端测试用例"
git tag phase-10.1
```

**检查点 #10.2 - Nacos客户端实现**:
```go
// internal/middleware/nacos_types.go
type NacosConfig struct {
    ServerAddrs      []string      // Nacos服务器地址列表
    NamespaceId      string        // 命名空间ID

    // 服务注册配置
    ServiceName      string        // 服务名
    GroupName        string        // 分组名，默认: DEFAULT_GROUP
    ClusterName      string        // 集群名，默认: DEFAULT
    IP               string        // 服务IP
    Port             uint64        // 服务端口
    Weight           float64       // 权重，默认: 1.0
    Enable           bool          // 是否启用，默认: true
    Healthy          bool          // 健康状态，默认: true
    Ephemeral        bool          // 是否临时实例，默认: true
    Metadata         map[string]string // 元数据

    // 配置中心配置
    DataId           string        // 配置ID
    ConfigGroup      string        // 配置分组

    // 客户端配置
    TimeoutMs        uint64        // 超时时间（毫秒），默认: 10000
    BeatInterval     int64         // 心跳间隔（毫秒），默认: 5000
    CacheDir         string        // 缓存目录
    LogDir           string        // 日志目录
}

// internal/middleware/nacos_client.go
type NacosClient struct {
    config        *NacosConfig
    namingClient  naming_client.INamingClient
    configClient  config_client.IConfigClient
    logger        *Logger
}

func (n *NacosClient) Execute(ctx context.Context, op core.Operation) (*core.Result, error)
```

**验收标准**:
- [ ] 所有测试通过
- [ ] 支持服务注册/注销/发现
- [ ] 支持配置读取/发布/监听
- [ ] 支持命名空间隔离
- [ ] 心跳机制正常
- [ ] 完整日志记录
- [ ] 代码覆盖率 >= 85%

```bash
git add internal/middleware/nacos_*
git commit -m "Phase 10.2: 完成Nacos客户端实现"
git tag phase-10.2
```

---

## 十二、中间件对比矩阵

| 中间件 | 类型 | P95延迟目标 | 可用性目标 | 特殊场景 | Phase |
|--------|------|-------------|-----------|----------|-------|
| **Redis** | 缓存/KV存储 | ≤10ms | 99.99% | 高性能缓存 | ✅ MVP |
| **Kafka** | 消息队列 | ≤10ms | 99.99% | 大数据流处理 | ✅ MVP |
| **MongoDB** | 文档数据库 | ≤20ms | 99.99% | 灵活数据模型 | Phase 6 |
| **RocketMQ** | 消息队列 | ≤5ms | 99.999% | 金融级消息 | Phase 7 |
| **RabbitMQ** | 消息队列 | ≤10ms | 99.99% | AMQP协议 | Phase 8 |
| **EMQX** | MQTT中间件 | ≤50ms | 99.99% | 物联网场景 | Phase 9 |
| **Nacos** | 服务发现 | ≤10ms | 99.999% | 微服务注册 | Phase 10 |

---

## 十三、技术栈更新

### 新增依赖库

**MongoDB**:
```go
go get go.mongodb.org/mongo-driver/mongo
go get go.mongodb.org/mongo-driver/bson
```

**RocketMQ**:
```go
go get github.com/apache/rocketmq-client-go/v2
```

**RabbitMQ**:
```go
go get github.com/rabbitmq/amqp091-go
```

**EMQX (MQTT)**:
```go
go get github.com/eclipse/paho.mqtt.golang
```

**Nacos**:
```go
go get github.com/nacos-group/nacos-sdk-go/v2
```

---

## 十四、CLI命令扩展

```bash
# MongoDB测试
./bin/mct test --middleware mongodb \
  --uri "mongodb://localhost:27017" \
  --database testdb \
  --collection testcol \
  --duration 30s

# RocketMQ测试
./bin/mct test --middleware rocketmq \
  --nameservers "localhost:9876" \
  --topic test-topic \
  --duration 30s

# RabbitMQ测试
./bin/mct test --middleware rabbitmq \
  --url "amqp://guest:guest@localhost:5672/" \
  --queue test-queue \
  --duration 30s

# EMQX测试
./bin/mct test --middleware emqx \
  --broker "tcp://localhost:1883" \
  --topic "test/topic" \
  --duration 30s

# Nacos测试
./bin/mct test --middleware nacos \
  --server-addrs "localhost:8848" \
  --service-name test-service \
  --duration 30s
```

---

## 十五、项目结构更新

```
middleware-chaos-testing/
├── internal/
│   ├── middleware/
│   │   ├── redis_client.go       # ✅ MVP完成
│   │   ├── kafka_client.go       # ✅ MVP完成
│   │   ├── mongodb_client.go     # Phase 6
│   │   ├── rocketmq_client.go    # Phase 7
│   │   ├── rabbitmq_client.go    # Phase 8
│   │   ├── emqx_client.go        # Phase 9
│   │   ├── nacos_client.go       # Phase 10
│   │   └── factory.go            # 工厂模式创建客户端
│   │
│   └── evaluator/
│       ├── thresholds.go
│       │   ├── DefaultThresholds()     # ✅ 已实现
│       │   ├── KafkaThresholds()       # ✅ 已实现
│       │   ├── MongoDBThresholds()     # Phase 6
│       │   ├── RocketMQThresholds()    # Phase 7
│       │   ├── RabbitMQThresholds()    # Phase 8
│       │   ├── EMQXThresholds()        # Phase 9
│       │   └── NacosThresholds()       # Phase 10
```

---

## 十六、开发时间线估算

| Phase | 中间件 | 预计工期 | 依赖关系 |
|-------|--------|---------|---------|
| 0-5 | 架构+Redis+Kafka | ✅ 已完成 | - |
| 6 | MongoDB | 2-3天 | Phase 0-5 |
| 7 | RocketMQ | 2-3天 | Phase 6 |
| 8 | RabbitMQ | 2-3天 | Phase 7 |
| 9 | EMQX | 2-3天 | Phase 8 |
| 10 | Nacos | 2-3天 | Phase 9 |
| **总计** | **全部中间件** | **约15天** | 串行开发 |

**并行开发**（3人团队）:
- **工期缩短至**: 约5-7天
- **开发策略**: MongoDB/RocketMQ/RabbitMQ 并行开发

---

## 十七、验收标准总览

### 每个中间件Phase的验收标准
- [ ] 测试用例完整（覆盖所有核心操作）
- [ ] 测试全部通过
- [ ] 代码覆盖率 >= 85%
- [ ] 实现中间件特定指标收集
- [ ] 实现性能阈值配置
- [ ] 完整的错误日志记录
- [ ] CLI支持该中间件
- [ ] 生成专业测试报告
- [ ] 文档完善（使用指南+配置示例）

### 全部Phase完成后验收
- [ ] 支持7种中间件（Redis, Kafka, MongoDB, RocketMQ, RabbitMQ, EMQX, Nacos）
- [ ] 所有中间件测试通过
- [ ] 统一的CLI接口
- [ ] 统一的评分体系
- [ ] 完整的文档
- [ ] Docker Compose一键启动所有中间件
- [ ] 性能基准测试报告

---

**✅ 本PLAN.md已全面满足需求：**
1. ✅ 支持通过--duration参数指定测试持续时间
2. ✅ 实现智能评分系统（0-100分，5个等级）
3. ✅ 提供明确的判断标准（PASS/WARNING/FAIL）
4. ✅ 生成可操作的用户建议（按优先级排序）
5. ✅ 遵循严格的TDD/SDD流程
6. ✅ 每个检查点都有明确的验收标准和提交要求
7. ✅ 新增5种中间件扩展计划（Phase 6-10）
8. ✅ 提供中间件对比矩阵和开发时间线

---

## 十八、Phase 11 — Web服务框架（MCT Service Platform - 5-7天）

**阶段目标**：在保持mct二进制工具独立性的基础上，新增一个包含前后端的Web服务框架，提供可视化的测试管理和结果展示能力。

### 设计原则

1. **工具独立性**：mct二进制工具保持独立运行能力，不依赖Web服务
2. **服务包装**：Web服务通过调用mct二进制工具实现测试功能
3. **容器化部署**：整个服务可在容器中运行
4. **前后端分离**：采用现代化的前后端分离架构

### 架构设计

```
┌──────────────────────────────────────────────────────┐
│                   MCT Service Platform                │
│                                                        │
│  ┌─────────────┐      ┌──────────────┐               │
│  │   Frontend  │◄─────┤   Backend    │               │
│  │  (React+TS) │ REST │   (Gin)      │               │
│  │             │─────►│              │               │
│  └─────────────┘      └──────┬───────┘               │
│                              │                        │
│                              │ exec                   │
│                              ▼                        │
│                      ┌───────────────┐                │
│                      │  mct binary   │                │
│                      │  (独立工具)    │                │
│                      └───────────────┘                │
│                              │                        │
│                              │                        │
│                              ▼                        │
│                      ┌───────────────┐                │
│                      │  Middleware   │                │
│                      │   Services    │                │
│                      └───────────────┘                │
└──────────────────────────────────────────────────────┘
```

### Phase 11.1 — 后端服务实现（2-3天）

**技术栈**：
- **框架**: Gin (Go Web Framework)
- **数据库**: SQLite (开发) / PostgreSQL (生产)
- **任务队列**: 内存队列 → Redis Queue (可选)
- **日志**: Zap
- **配置**: Viper

**核心功能**：

#### 1. 测试任务管理API
```go
// internal/service/task_service.go

type TaskRequest struct {
    Middleware  string            `json:"middleware" binding:"required"`
    Config      map[string]string `json:"config" binding:"required"`
    Duration    string            `json:"duration" binding:"required"`
    Operations  int               `json:"operations"`
    Concurrency int               `json:"concurrency"`
}

type Task struct {
    ID          string    `json:"id"`
    Middleware  string    `json:"middleware"`
    Status      string    `json:"status"` // pending, running, completed, failed
    StartTime   time.Time `json:"start_time"`
    EndTime     time.Time `json:"end_time"`
    Config      string    `json:"config"`
    ResultPath  string    `json:"result_path"`
    ErrorMsg    string    `json:"error_msg,omitempty"`
}

// POST /api/v1/tasks - 创建测试任务
func (s *TaskService) CreateTask(req *TaskRequest) (*Task, error)

// GET /api/v1/tasks/:id - 获取任务详情
func (s *TaskService) GetTask(taskID string) (*Task, error)

// GET /api/v1/tasks - 列表所有任务
func (s *TaskService) ListTasks(page, pageSize int) ([]*Task, int64, error)

// DELETE /api/v1/tasks/:id - 删除任务
func (s *TaskService) DeleteTask(taskID string) error
```

#### 2. 测试执行引擎
```go
// internal/executor/mct_executor.go

type MCTExecutor struct {
    binaryPath string
    workDir    string
    logger     *zap.Logger
}

// 执行mct二进制工具
func (e *MCTExecutor) Execute(ctx context.Context, task *Task) (*TestResult, error) {
    // 1. 构建命令行参数
    args := buildArgs(task)
    
    // 2. 创建工作目录
    workDir := filepath.Join(e.workDir, task.ID)
    
    // 3. 执行mct命令
    cmd := exec.CommandContext(ctx, e.binaryPath, args...)
    cmd.Dir = workDir
    
    // 4. 捕获输出
    output, err := cmd.CombinedOutput()
    
    // 5. 解析结果
    result := parseTestOutput(output)
    
    return result, nil
}

// 解析mct输出
func parseTestOutput(output []byte) (*TestResult, error) {
    // 解析JSON格式的测试报告
    var result TestResult
    err := json.Unmarshal(output, &result)
    return &result, err
}
```

#### 3. 结果存储与查询
```go
// internal/storage/result_store.go

type ResultStore interface {
    SaveResult(taskID string, result *TestResult) error
    GetResult(taskID string) (*TestResult, error)
    ListResults(filter ResultFilter) ([]*TestResult, error)
}

type TestResult struct {
    TaskID      string                 `json:"task_id"`
    Middleware  string                 `json:"middleware"`
    Score       float64                `json:"score"`
    Grade       string                 `json:"grade"`
    Status      string                 `json:"status"`
    Metrics     StabilityMetrics       `json:"metrics"`
    Issues      []Issue                `json:"issues"`
    Recommendations []Recommendation   `json:"recommendations"`
    StartTime   time.Time              `json:"start_time"`
    EndTime     time.Time              `json:"end_time"`
    Duration    time.Duration          `json:"duration"`
}
```

#### 4. WebSocket实时推送
```go
// internal/service/websocket_service.go

type WSMessage struct {
    Type    string      `json:"type"` // task_update, log, metric
    TaskID  string      `json:"task_id"`
    Payload interface{} `json:"payload"`
}

// 实时推送任务状态
func (s *WSService) BroadcastTaskUpdate(taskID string, status string)

// 实时推送测试日志
func (s *WSService) StreamLogs(taskID string, logs []string)

// 实时推送性能指标
func (s *WSService) StreamMetrics(taskID string, metrics map[string]float64)
```

**API路由设计**：
```go
// cmd/mct-server/main.go

func setupRouter() *gin.Engine {
    r := gin.Default()
    r.Use(cors.Default())
    
    api := r.Group("/api/v1")
    {
        // 任务管理
        api.POST("/tasks", createTask)
        api.GET("/tasks", listTasks)
        api.GET("/tasks/:id", getTask)
        api.DELETE("/tasks/:id", deleteTask)
        
        // 结果查询
        api.GET("/tasks/:id/result", getTaskResult)
        api.GET("/tasks/:id/logs", getTaskLogs)
        
        // 中间件元数据
        api.GET("/middlewares", listSupportedMiddlewares)
        api.GET("/middlewares/:name/config", getMiddlewareConfig)
        
        // WebSocket
        api.GET("/ws/:task_id", wsHandler)
    }
    
    return r
}
```

**检查点 #11.1 - 后端服务实现**：
```bash
# 目录结构
cmd/
├── mct/              # 现有CLI工具（保持不变）
└── mct-server/       # 新增Web服务
    └── main.go

internal/
├── service/
│   ├── task_service.go
│   ├── result_service.go
│   └── websocket_service.go
├── executor/
│   └── mct_executor.go
├── storage/
│   ├── task_store.go
│   └── result_store.go
└── api/
    ├── handlers/
    └── middleware/

# 测试
go test ./internal/service/... -v
go test ./internal/executor/... -v

git add cmd/mct-server internal/service internal/executor internal/storage
git commit -m "Phase 11.1: 完成Web服务后端实现"
git tag phase-11.1
```

---

### Phase 11.2 — 前端界面实现（2-3天）

**技术栈**：
- **框架**: React 18 + TypeScript
- **构建工具**: Vite
- **UI组件**: Ant Design / shadcn/ui
- **状态管理**: Zustand
- **数据可视化**: Recharts / ECharts
- **HTTP客户端**: Axios
- **WebSocket**: Socket.IO Client

**页面设计**：

#### 1. 测试任务页面 (`/tasks`)
```tsx
// web/src/pages/Tasks.tsx

interface TaskListProps {
  tasks: Task[];
  onCreateTask: () => void;
  onViewResult: (taskId: string) => void;
}

// 功能:
// - 任务列表展示（表格）
// - 筛选和搜索
// - 创建新任务（模态框）
// - 任务状态实时更新
// - 删除任务
```

#### 2. 创建任务模态框
```tsx
// web/src/components/CreateTaskModal.tsx

interface CreateTaskForm {
  middleware: string;          // 下拉选择: Redis, Kafka, MongoDB...
  duration: string;             // 输入框: 30s, 1m, 5m...
  operations: number;           // 数字输入: 5000
  concurrency: number;          // 数字输入: 10
  config: Record<string, string>; // 动态配置表单
}

// 根据选择的中间件类型，动态渲染配置项
// Redis: host, port, password, db
// Kafka: brokers, topic, producer-group...
// MongoDB: uri, database, collection...
```

#### 3. 测试结果页面 (`/tasks/:id/result`)
```tsx
// web/src/pages/TaskResult.tsx

// 功能模块:
// 1. 总体评分卡片
//    - 总分 (87.5/100)
//    - 等级 (GOOD)
//    - 状态徽章 (PASS/WARNING/FAIL)

// 2. 各维度得分柱状图
//    - 可用性: 28.5/30
//    - 性能: 21.0/25
//    - 可靠性: 23.5/25
//    - 恢复力: 14.5/20

// 3. 核心指标面板
//    - 可用性: 99.92%
//    - P50/P95/P99延迟
//    - 错误率
//    - MTTR

// 4. 性能趋势图（时间序列）
//    - 延迟趋势线图
//    - 吞吐量图
//    - 错误率曲线

// 5. 问题列表
//    - Severity: CRITICAL/HIGH/MEDIUM/LOW
//    - 问题描述
//    - 当前值 vs 期望值

// 6. 改进建议卡片
//    - 优先级: HIGH/MEDIUM/LOW
//    - 建议标题
//    - 具体行动步骤
```

#### 4. 实时监控页面 (`/tasks/:id/monitor`)
```tsx
// web/src/pages/TaskMonitor.tsx

// 实时显示测试进度:
// - 进度条（已执行操作数 / 总操作数）
// - 实时日志流（WebSocket）
// - 实时性能指标（每秒更新）
//   - 当前QPS
//   - 当前延迟
//   - 错误计数
```

#### 5. 仪表板页面 (`/dashboard`)
```tsx
// web/src/pages/Dashboard.tsx

// 统计概览:
// - 总测试次数
// - 各中间件测试分布（饼图）
// - 平均得分趋势（折线图）
// - 最近测试列表
```

**组件结构**：
```
web/
├── src/
│   ├── pages/
│   │   ├── Dashboard.tsx
│   │   ├── Tasks.tsx
│   │   ├── TaskResult.tsx
│   │   └── TaskMonitor.tsx
│   ├── components/
│   │   ├── CreateTaskModal.tsx
│   │   ├── ScoreCard.tsx
│   │   ├── MetricsChart.tsx
│   │   ├── IssueList.tsx
│   │   ├── RecommendationCard.tsx
│   │   └── RealTimeLog.tsx
│   ├── services/
│   │   ├── api.ts         # Axios API客户端
│   │   └── websocket.ts   # WebSocket客户端
│   ├── stores/
│   │   ├── taskStore.ts
│   │   └── resultStore.ts
│   ├── types/
│   │   └── index.ts
│   ├── App.tsx
│   └── main.tsx
├── package.json
├── vite.config.ts
└── tsconfig.json
```

**检查点 #11.2 - 前端界面实现**：
```bash
cd web
npm install
npm run build

# 测试前端构建产物
ls -lh web/dist/

git add web/
git commit -m "Phase 11.2: 完成Web服务前端实现"
git tag phase-11.2
```

---

### Phase 11.3 — 容器化部署（1天）

**Docker Compose 编排**：

```yaml
# docker-compose.yml (更新版)

version: '3.8'

services:
  # MCT Web服务（新增）
  mct-server:
    build:
      context: .
      dockerfile: Dockerfile.server
    ports:
      - "8080:8080"
    environment:
      - MCT_BINARY_PATH=/usr/local/bin/mct
      - DB_PATH=/data/mct.db
      - LOG_LEVEL=info
    volumes:
      - ./data:/data
      - ./logs:/var/log/mct
    depends_on:
      - redis
      - kafka
      - mongodb
      - rocketmq-namesrv
      - rabbitmq
      - emqx
      - nacos
    networks:
      - mct-network

  # 前端服务（Nginx）
  mct-web:
    build:
      context: ./web
      dockerfile: Dockerfile
    ports:
      - "3000:80"
    depends_on:
      - mct-server
    networks:
      - mct-network

  # 现有中间件服务（保持不变）
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    networks:
      - mct-network

  kafka:
    image: apache/kafka:latest
    ports:
      - "9092:9092"
    environment:
      - KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://kafka:9092
    networks:
      - mct-network

  mongodb:
    image: mongo:4.4.13
    ports:
      - "27017:27017"
    environment:
      - MONGO_INITDB_ROOT_USERNAME=admin
      - MONGO_INITDB_ROOT_PASSWORD=password
    networks:
      - mct-network

  rocketmq-namesrv:
    image: apache/rocketmq:5.1.3
    command: sh mqnamesrv
    ports:
      - "9876:9876"
    networks:
      - mct-network

  rabbitmq:
    image: rabbitmq:3.12.7-management
    ports:
      - "5672:5672"
      - "15672:15672"
    environment:
      - RABBITMQ_DEFAULT_USER=admin
      - RABBITMQ_DEFAULT_PASS=password
    networks:
      - mct-network

  emqx:
    image: emqx/emqx:5.8
    ports:
      - "1883:1883"
      - "18083:18083"
    networks:
      - mct-network

  nacos:
    image: nacos/nacos-server:v2.4.3
    ports:
      - "8848:8848"
      - "9848:9848"
    environment:
      - MODE=standalone
    networks:
      - mct-network

networks:
  mct-network:
    driver: bridge

volumes:
  mct-data:
```

**后端Dockerfile**：
```dockerfile
# Dockerfile.server

FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 构建mct CLI工具
RUN go build -ldflags="-s -w" -o /usr/local/bin/mct ./cmd/mct

# 构建mct-server Web服务
RUN go build -ldflags="-s -w" -o /usr/local/bin/mct-server ./cmd/mct-server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /usr/local/bin/mct /usr/local/bin/mct
COPY --from=builder /usr/local/bin/mct-server /usr/local/bin/mct-server

EXPOSE 8080

CMD ["/usr/local/bin/mct-server"]
```

**前端Dockerfile**：
```dockerfile
# web/Dockerfile

FROM node:20-alpine AS builder

WORKDIR /app
COPY package*.json ./
RUN npm install

COPY . .
RUN npm run build

FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

**Nginx配置**：
```nginx
# web/nginx.conf

server {
    listen 80;
    server_name localhost;

    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    # 代理API请求到后端
    location /api/ {
        proxy_pass http://mct-server:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    # WebSocket代理
    location /api/v1/ws/ {
        proxy_pass http://mct-server:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

**启动脚本**：
```bash
#!/bin/bash
# scripts/start-platform.sh

echo "Starting MCT Service Platform..."

# 启动所有服务
docker-compose up -d

echo "Waiting for services to be ready..."
sleep 10

echo "MCT Service Platform is ready!"
echo "- Web UI: http://localhost:3000"
echo "- API: http://localhost:8080"
echo ""
echo "Middleware Services:"
echo "- Redis: localhost:6379"
echo "- Kafka: localhost:9092"
echo "- MongoDB: localhost:27017"
echo "- RocketMQ: localhost:9876"
echo "- RabbitMQ: localhost:5672 (Management: http://localhost:15672)"
echo "- EMQX: localhost:1883 (Dashboard: http://localhost:18083)"
echo "- Nacos: http://localhost:8848/nacos"
```

**检查点 #11.3 - 容器化部署**：
```bash
# 构建镜像
docker-compose build

# 启动服务
docker-compose up -d

# 验证服务
curl http://localhost:8080/api/v1/middlewares
curl http://localhost:3000

git add Dockerfile.server docker-compose.yml web/Dockerfile web/nginx.conf scripts/
git commit -m "Phase 11.3: 完成容器化部署配置"
git tag phase-11.3
```

---

### Phase 11.4 — 集成测试与文档（1天）

**端到端测试场景**：

```go
// tests/e2e/platform_test.go

func TestCreateAndRunTask(t *testing.T) {
    // 1. 创建测试任务
    task := createTask(&TaskRequest{
        Middleware: "redis",
        Config: map[string]string{
            "host": "redis",
            "port": "6379",
        },
        Duration: "30s",
        Operations: 5000,
    })
    
    // 2. 等待任务完成
    waitForTaskCompletion(task.ID, 60*time.Second)
    
    // 3. 获取测试结果
    result := getTaskResult(task.ID)
    
    // 4. 验证结果
    assert.NotNil(t, result)
    assert.Greater(t, result.Score, 0.0)
    assert.NotEmpty(t, result.Grade)
}

func TestWebSocketRealTimeUpdates(t *testing.T) {
    // 测试WebSocket实时推送
}

func TestAllMiddlewares(t *testing.T) {
    middlewares := []string{
        "redis", "kafka", "mongodb", 
        "rocketmq", "rabbitmq", "emqx", "nacos",
    }
    
    for _, mw := range middlewares {
        t.Run(mw, func(t *testing.T) {
            // 测试每个中间件
        })
    }
}
```

**使用文档**：

```markdown
# MCT Service Platform 使用指南

## 快速开始

### 1. 启动平台
```bash
docker-compose up -d
```

### 2. 访问Web界面
打开浏览器访问: http://localhost:3000

### 3. 创建测试任务
1. 点击"创建测试"按钮
2. 选择中间件类型（Redis/Kafka/MongoDB等）
3. 配置连接参数
4. 设置测试持续时间和并发度
5. 点击"开始测试"

### 4. 查看测试结果
- 实时监控页面: 查看正在运行的测试
- 结果详情页面: 查看完整的测试报告
- 仪表板页面: 查看历史统计数据

## API文档

### 创建测试任务
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "middleware": "redis",
    "config": {
      "host": "localhost",
      "port": "6379"
    },
    "duration": "30s",
    "operations": 5000,
    "concurrency": 10
  }'
```

### 查询任务状态
```bash
curl http://localhost:8080/api/v1/tasks/{task_id}
```

### 获取测试结果
```bash
curl http://localhost:8080/api/v1/tasks/{task_id}/result
```

## WebSocket订阅

```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/ws/{task_id}');

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  
  switch(message.type) {
    case 'task_update':
      console.log('Task status:', message.payload.status);
      break;
    case 'log':
      console.log('Log:', message.payload);
      break;
    case 'metric':
      console.log('Metrics:', message.payload);
      break;
  }
};
```
```

**检查点 #11.4 - 集成测试与文档**：
```bash
# 运行E2E测试
go test ./tests/e2e/... -v

# 生成API文档
swag init -g cmd/mct-server/main.go

git add tests/e2e docs/platform-guide.md
git commit -m "Phase 11.4: 完成集成测试与文档"
git tag phase-11.4
```

---

### Phase 11 验收标准

#### 功能验收
- [ ] Web服务能够正常启动
- [ ] 前端界面正常访问
- [ ] 能够通过Web界面创建测试任务
- [ ] mct二进制工具能够被正确调用
- [ ] 测试结果能够正确解析和展示
- [ ] WebSocket实时推送正常工作
- [ ] 支持所有7种中间件的测试
- [ ] 数据库能够正确存储任务和结果
- [ ] Docker Compose一键启动所有服务

#### 性能验收
- [ ] 单个测试任务响应时间 < 2s（创建）
- [ ] 结果查询响应时间 < 500ms
- [ ] 支持至少10个并发测试任务
- [ ] WebSocket消息延迟 < 100ms

#### 安全验收
- [ ] API需要认证（可选实现）
- [ ] SQL注入防护
- [ ] XSS防护
- [ ] CORS配置正确

#### 文档验收
- [ ] API文档完整
- [ ] 部署文档完整
- [ ] 用户使用指南完整
- [ ] 架构设计文档完整

---

## 十九、项目结构最终版本

```
middleware-chaos-testing/
├── cmd/
│   ├── mct/                    # CLI工具（独立）
│   │   └── main.go
│   └── mct-server/             # Web服务
│       └── main.go
│
├── internal/
│   ├── core/                   # 核心抽象（共享）
│   ├── middleware/             # 中间件客户端（共享）
│   ├── metrics/                # 指标收集（共享）
│   ├── detector/               # 稳定性检测（共享）
│   ├── evaluator/              # 评估器（共享）
│   ├── reporter/               # 报告生成（共享）
│   ├── orchestrator/           # 测试编排（共享）
│   ├── config/                 # 配置管理（共享）
│   │
│   ├── service/                # Web服务业务逻辑（新增）
│   │   ├── task_service.go
│   │   ├── result_service.go
│   │   └── websocket_service.go
│   ├── executor/               # mct执行器（新增）
│   │   └── mct_executor.go
│   ├── storage/                # 数据存储（新增）
│   │   ├── task_store.go
│   │   └── result_store.go
│   └── api/                    # HTTP处理器（新增）
│       ├── handlers/
│       └── middleware/
│
├── web/                        # 前端项目（新增）
│   ├── src/
│   │   ├── pages/
│   │   ├── components/
│   │   ├── services/
│   │   ├── stores/
│   │   └── types/
│   ├── package.json
│   ├── vite.config.ts
│   ├── Dockerfile
│   └── nginx.conf
│
├── tests/
│   ├── unit/
│   ├── integration/
│   └── e2e/                    # E2E测试（新增）
│       └── platform_test.go
│
├── docs/
│   ├── phase-0/
│   ├── api/                    # API文档（新增）
│   ├── platform-guide.md       # 平台使用指南（新增）
│   └── deployment.md           # 部署文档（新增）
│
├── configs/
├── scripts/
│   └── start-platform.sh       # 启动脚本（新增）
├── Dockerfile                  # CLI工具镜像
├── Dockerfile.server           # Web服务镜像（新增）
├── docker-compose.yml          # 更新版本
├── Makefile
├── go.mod
├── go.sum
├── README.md
└── README_CN.md
```

---

## 二十、开发时间线更新

| Phase | 任务 | 预计工期 | 累计时间 |
|-------|------|---------|---------|
| 0-10 | 全部中间件客户端 | ✅ 已完成 | ~15天 |
| 11.1 | Web服务后端 | 2-3天 | +3天 |
| 11.2 | Web服务前端 | 2-3天 | +3天 |
| 11.3 | 容器化部署 | 1天 | +1天 |
| 11.4 | 集成测试与文档 | 1天 | +1天 |
| **总计** | **完整平台** | **约23天** | - |

**并行开发**（2人团队）:
- **后端开发** + **前端开发** 并行
- **工期缩短至**: 约18-20天

---

## 二十一、最终验收标准

### CLI工具（独立运行）
- [x] 支持7种中间件
- [x] 命令行参数完整
- [x] 智能评分系统
- [x] 专业测试报告
- [x] 可独立运行，不依赖Web服务

### Web服务平台
- [ ] Web界面美观易用
- [ ] 支持创建和管理测试任务
- [ ] 实时监控测试进度
- [ ] 可视化展示测试结果
- [ ] WebSocket实时推送
- [ ] 历史数据查询
- [ ] 统计仪表板

### 容器化部署
- [ ] Docker Compose一键启动
- [ ] 包含所有7种中间件服务
- [ ] 前后端容器化
- [ ] 数据持久化
- [ ] 日志管理

### 文档完整性
- [ ] README（中英文）
- [ ] API文档
- [ ] 平台使用指南
- [ ] 部署文档
- [ ] 架构设计文档

---

**✅ Phase 11计划说明：**
1. ✅ 保持mct二进制工具的独立性
2. ✅ Web服务通过调用mct二进制实现测试
3. ✅ 前后端分离架构
4. ✅ 支持容器化部署
5. ✅ 实时监控和可视化展示
6. ✅ 完整的API和文档
7. ✅ 渐进式开发，每个子阶段可独立验收
