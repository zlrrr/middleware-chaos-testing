# 指标详细定义

## 1. 指标体系概述

中间件混沌测试框架采用四维度指标体系，全面评估中间件的稳定性：

```
稳定性指标体系
├── 可用性指标 (Availability) - 权重 30%
├── 性能指标 (Performance) - 权重 25%
├── 可靠性指标 (Reliability) - 权重 25%
└── 恢复性指标 (Resilience) - 权重 20%
```

## 2. 可用性指标 (Availability Metrics)

### 2.1 总操作数 (Total Operations)

**定义**: 测试期间执行的操作总数

**计算公式**:
```
TotalOperations = SuccessfulOperations + FailedOperations
```

**数据类型**: int64

**示例值**: 10000

---

### 2.2 成功操作数 (Successful Operations)

**定义**: 成功完成的操作数量

**计算方法**: 操作返回成功状态时计数加1

**数据类型**: int64

**示例值**: 9992

---

### 2.3 失败操作数 (Failed Operations)

**定义**: 失败的操作数量

**计算方法**: 操作返回失败状态时计数加1

**数据类型**: int64

**示例值**: 8

---

### 2.4 可用性 (Availability)

**定义**: 成功操作占总操作的比例

**计算公式**:
```
Availability = SuccessfulOperations / TotalOperations
```

**数据类型**: float64 (0.0 - 1.0)

**单位**: 百分比

**示例值**: 0.9992 (99.92%)

**阈值标准**:
- 优秀 (Excellent): >= 99.99%
- 良好 (Good): >= 99.9%
- 一般 (Fair): >= 99.0%
- 及格 (Pass): >= 95.0%
- 失败 (Fail): < 95.0%

---

### 2.5 连接成功率 (Connection Success Rate)

**定义**: 成功建立连接的比例

**计算公式**:
```
ConnectionSuccessRate = SuccessfulConnectionAttempts / TotalConnectionAttempts
```

**数据类型**: float64 (0.0 - 1.0)

**单位**: 百分比

**示例值**: 0.98 (98%)

**阈值标准**:
- 优秀: >= 99%
- 良好: >= 95%
- 及格: >= 90%

---

### 2.6 服务可用时间百分比 (Service Uptime Percentage)

**定义**: 服务可用时间占总测试时间的比例

**计算公式**:
```
UptimePercentage = (TotalDuration - DowntimeDuration) / TotalDuration
```

**数据类型**: float64 (0.0 - 1.0)

**单位**: 百分比

**示例值**: 0.9995 (99.95%)

---

## 3. 性能指标 (Performance Metrics)

### 3.1 P50 延迟 (P50 Latency)

**定义**: 第50百分位延迟，即50%的操作延迟低于此值

**计算方法**: 对所有操作延迟排序，取中位数

**数据类型**: time.Duration

**单位**: 毫秒 (ms)

**示例值**: 8ms

**阈值标准**:
- 优秀: <= 5ms
- 良好: <= 20ms
- 一般: <= 50ms
- 及格: <= 100ms

---

### 3.2 P95 延迟 (P95 Latency)

**定义**: 第95百分位延迟，即95%的操作延迟低于此值

**计算方法**: 对所有操作延迟排序，取第95百分位值

**数据类型**: time.Duration

**单位**: 毫秒 (ms)

**示例值**: 45ms

**阈值标准**:
- 优秀: <= 10ms
- 良好: <= 50ms
- 一般: <= 100ms
- 及格: <= 200ms

**权重**: 在性能评分中占 60% (15/25分)

---

### 3.3 P99 延迟 (P99 Latency)

**定义**: 第99百分位延迟，即99%的操作延迟低于此值

**计算方法**: 对所有操作延迟排序，取第99百分位值

**数据类型**: time.Duration

**单位**: 毫秒 (ms)

**示例值**: 120ms

**阈值标准**:
- 优秀: <= 20ms
- 良好: <= 100ms
- 一般: <= 200ms
- 及格: <= 500ms

**权重**: 在性能评分中占 40% (10/25分)

**说明**: P99延迟反映了尾部延迟，对用户体验有重要影响

---

### 3.4 平均延迟 (Average Latency)

**定义**: 所有操作的平均延迟时间

**计算公式**:
```
AvgLatency = Sum(AllOperationDurations) / TotalOperations
```

**数据类型**: time.Duration

**单位**: 毫秒 (ms)

**示例值**: 12ms

---

### 3.5 最大延迟 (Max Latency)

**定义**: 所有操作中的最大延迟时间

**计算方法**: 记录过程中遇到的最大延迟值

**数据类型**: time.Duration

**单位**: 毫秒 (ms)

**示例值**: 850ms

---

### 3.6 最小延迟 (Min Latency)

**定义**: 所有操作中的最小延迟时间

**计算方法**: 记录过程中遇到的最小延迟值

**数据类型**: time.Duration

**单位**: 毫秒 (ms)

**示例值**: 2ms

---

### 3.7 吞吐量 (Throughput)

**定义**: 每秒完成的操作数

**计算公式**:
```
Throughput = TotalOperations / Duration.Seconds()
```

**数据类型**: float64

**单位**: ops/s (operations per second)

**示例值**: 167 ops/s

**阈值标准**:
- 优秀: >= 1000 ops/s
- 良好: >= 500 ops/s
- 一般: >= 100 ops/s

---

### 3.8 并发连接数 (Concurrent Connections)

**定义**: 同时活跃的连接数

**计算方法**: 实时统计当前活跃连接数，记录峰值

**数据类型**: int

**示例值**: 50

---

## 4. 可靠性指标 (Reliability Metrics)

### 4.1 错误率 (Error Rate)

**定义**: 失败操作占总操作的比例

**计算公式**:
```
ErrorRate = FailedOperations / TotalOperations
```

**数据类型**: float64 (0.0 - 1.0)

**单位**: 百分比

**示例值**: 0.0008 (0.08%)

**阈值标准**:
- 优秀: <= 0.01%
- 良好: <= 0.1%
- 一般: <= 0.5%
- 及格: <= 1.0%

**权重**: 在可靠性评分中占 60% (15/25分)

---

### 4.2 数据一致性 (Data Consistency)

**定义**: 写入后能正确读取的数据比例

**计算方法**:
```
1. 写入测试数据时记录哈希值
2. 读取时验证哈希值是否匹配
3. DataConsistency = MatchedReads / TotalValidationReads
```

**数据类型**: float64 (0.0 - 1.0)

**单位**: 百分比

**示例值**: 1.0 (100%)

**阈值标准**:
- 优秀: == 100%
- 良好: >= 99.99%
- 及格: >= 99.9%

---

### 4.3 数据丢失率 (Data Loss Rate)

**定义**: 丢失的数据占总写入数据的比例

**计算方法**:
```
1. 记录所有写入的键
2. 测试结束后验证这些键是否存在
3. DataLossRate = LostKeys / TotalWrittenKeys
```

**数据类型**: float64 (0.0 - 1.0)

**单位**: 百分比

**示例值**: 0.0 (0%)

**阈值标准**:
- 优秀: == 0%
- 可接受: < 0.01%
- 不可接受: >= 0.1%

**权重**: 在可靠性评分中占 40% (10/25分)

**严重性**: CRITICAL（如果 > 0.1%）

---

### 4.4 重复率 (Duplicate Rate)

**定义**: 重复接收/处理的消息占总消息的比例（主要用于Kafka）

**计算方法**:
```
DuplicateRate = DuplicateMessages / TotalMessages
```

**数据类型**: float64 (0.0 - 1.0)

**单位**: 百分比

**示例值**: 0.002 (0.2%)

**阈值标准**:
- 优秀: < 0.1%
- 良好: < 1%
- 可接受: < 5%

---

### 4.5 乱序率 (Out-of-Order Rate)

**定义**: 乱序消息占总消息的比例（主要用于Kafka）

**计算方法**:
```
1. 发送消息时按顺序标记序号
2. 接收时检测序号是否递增
3. OutOfOrderRate = OutOfOrderMessages / TotalMessages
```

**数据类型**: float64 (0.0 - 1.0)

**单位**: 百分比

**示例值**: 0.0 (0%)

**阈值标准**:
- 优秀: == 0%
- 良好: < 0.01%
- 可接受: < 0.1%

---

## 5. 恢复性指标 (Resilience Metrics)

### 5.1 MTBF (Mean Time Between Failures)

**定义**: 平均故障间隔时间

**计算公式**:
```
MTBF = TotalUptime / NumberOfFailures
```

**数据类型**: time.Duration

**单位**: 秒 (s)

**示例值**: 600s (10分钟)

**阈值标准**:
- 优秀: >= 1小时
- 良好: >= 30分钟
- 及格: >= 10分钟

---

### 5.2 MTTR (Mean Time To Recovery)

**定义**: 平均恢复时间（从检测到故障到恢复服务的平均时间）

**计算方法**:
```
1. 记录每次故障发生时间
2. 记录每次恢复成功时间
3. MTTR = Sum(RecoveryTimes) / NumberOfFailures
```

**数据类型**: time.Duration

**单位**: 秒 (s)

**示例值**: 25s

**阈值标准**:
- 优秀: <= 5s
- 良好: <= 30s
- 一般: <= 60s
- 及格: <= 300s

**权重**: 在恢复力评分中占 60% (12/20分)

---

### 5.3 故障检测时间 (Fault Detection Time)

**定义**: 从故障发生到检测到故障的平均时间

**计算方法**:
```
1. 模拟故障注入
2. 记录检测到故障的时间
3. FaultDetectionTime = DetectionTime - FaultInjectionTime
```

**数据类型**: time.Duration

**单位**: 毫秒 (ms)

**示例值**: 500ms

**阈值标准**:
- 优秀: <= 100ms
- 良好: <= 500ms
- 及格: <= 1000ms

---

### 5.4 重连成功率 (Reconnect Success Rate)

**定义**: 连接断开后重连成功的比例

**计算公式**:
```
ReconnectSuccessRate = SuccessfulReconnects / TotalReconnectAttempts
```

**数据类型**: float64 (0.0 - 1.0)

**单位**: 百分比

**示例值**: 0.96 (96%)

**阈值标准**:
- 优秀: >= 99%
- 良好: >= 95%
- 一般: >= 90%
- 及格: >= 85%

**权重**: 在恢复力评分中占 40% (8/20分)

---

## 6. 中间件特定指标

### 6.1 Redis 特定指标

#### 6.1.1 缓存命中率 (Cache Hit Rate)

**定义**: GET命令命中缓存的比例

**计算公式**:
```
CacheHitRate = CacheHits / (CacheHits + CacheMisses)
```

**数据类型**: float64 (0.0 - 1.0)

**单位**: 百分比

**示例值**: 0.85 (85%)

**阈值标准**:
- 优秀: >= 95%
- 良好: >= 90%
- 一般: >= 80%

**建议**: 如果命中率低于80%，建议优化缓存策略

---

#### 6.1.2 内存使用率 (Memory Usage)

**定义**: 已使用内存占最大内存的比例

**计算方法**: 通过 INFO memory 命令获取

**数据类型**: float64 (0.0 - 1.0)

**单位**: 百分比

**示例值**: 0.65 (65%)

**阈值标准**:
- 正常: < 80%
- 警告: 80% - 90%
- 危险: > 90%

---

#### 6.1.3 键空间利用率 (Keyspace Utilization)

**定义**: 已用键数占最大键数的比例

**计算方法**: 通过 DBSIZE 命令获取当前键数

**数据类型**: float64 (0.0 - 1.0)

**单位**: 百分比

**示例值**: 0.45 (45%)

---

#### 6.1.4 持久化延迟 (Persistence Latency)

**定义**: RDB/AOF 写入的延迟

**计算方法**: 监控 BGSAVE 和 AOF 重写的耗时

**数据类型**: time.Duration

**单位**: 毫秒 (ms)

**阈值标准**:
- 良好: < 100ms
- 可接受: < 1000ms

---

### 6.2 Kafka 特定指标

#### 6.2.1 消息积压 (Message Lag)

**定义**: 未消费的消息数量

**计算方法**: 通过消费者组监控获取

**数据类型**: int64

**单位**: 消息数

**示例值**: 1500

**阈值标准**:
- 良好: < 1000
- 警告: 1000 - 10000
- 危险: > 10000

---

#### 6.2.2 消费延迟 (Consumer Lag)

**定义**: 消息生产时间到消费时间的延迟

**计算公式**:
```
ConsumerLag = ConsumeTimestamp - ProduceTimestamp
```

**数据类型**: time.Duration

**单位**: 毫秒 (ms)

**示例值**: 350ms

**阈值标准**:
- 优秀: < 100ms
- 良好: < 500ms
- 可接受: < 1000ms

---

#### 6.2.3 分区重平衡次数 (Rebalance Count)

**定义**: 消费者组发生重平衡的次数

**计算方法**: 监听 rebalance 事件

**数据类型**: int64

**示例值**: 2

**阈值标准**:
- 良好: <= 2次
- 警告: 3-5次
- 危险: > 5次

**说明**: 频繁重平衡会影响消费性能

---

#### 6.2.4 ISR 收缩次数 (ISR Shrink Count)

**定义**: In-Sync Replicas 集合收缩的次数

**计算方法**: 监控 ISR 变化

**数据类型**: int64

**示例值**: 0

**阈值标准**:
- 良好: == 0
- 警告: 1-3次
- 危险: > 3次

**说明**: ISR收缩表明副本同步出现问题

---

## 7. 错误分类指标

### 7.1 按错误类型统计 (Errors By Type)

**定义**: 按错误类型分类的错误计数

**数据结构**:
```go
ErrorsByType map[ErrorType]int64
```

**错误类型**:
- `network`: 网络错误
- `timeout`: 超时错误
- `authentication`: 认证错误
- `data_loss`: 数据丢失
- `other`: 其他错误

**示例值**:
```json
{
  "network": 5,
  "timeout": 3,
  "authentication": 0,
  "data_loss": 0,
  "other": 2
}
```

---

## 8. 时间相关指标

### 8.1 测试开始时间 (Start Time)

**定义**: 测试开始的时间戳

**数据类型**: time.Time

**示例值**: 2025-10-30T14:00:00Z

---

### 8.2 测试结束时间 (End Time)

**定义**: 测试结束的时间戳

**数据类型**: time.Time

**示例值**: 2025-10-30T14:01:00Z

---

### 8.3 测试持续时间 (Duration)

**定义**: 测试实际运行的时长

**计算公式**:
```
Duration = EndTime - StartTime
```

**数据类型**: time.Duration

**单位**: 秒 (s)

**示例值**: 60s

---

## 9. 指标收集方法

### 9.1 实时收集

在每次操作完成时立即记录：
- 操作延迟
- 操作结果（成功/失败）
- 错误类型

### 9.2 周期性采样

每隔固定时间（如1秒）采样：
- 并发连接数
- 内存使用率
- CPU使用率

### 9.3 测试结束计算

在测试结束后计算：
- 百分位延迟（P50/P95/P99）
- 平均值
- 总体评分

### 9.4 数据结构

使用高效的数据结构存储指标：
- **计数器**: atomic.Int64
- **延迟**: 时间序列数组 + 排序
- **百分位**: 使用 HdrHistogram 或类似算法

---

## 10. 指标验证

### 10.1 数据完整性检查

```
TotalOperations == SuccessfulOperations + FailedOperations
Availability == SuccessfulOperations / TotalOperations
ErrorRate == 1 - Availability
```

### 10.2 合理性检查

```
0 <= Availability <= 1
0 <= ErrorRate <= 1
P50 <= P95 <= P99
MinLatency <= AvgLatency <= MaxLatency
MTBF >= 0
MTTR >= 0
```

---

## 11. 指标输出格式

### 11.1 JSON格式

```json
{
  "availability": 0.9992,
  "total_operations": 10000,
  "successful_operations": 9992,
  "failed_operations": 8,
  "error_rate": 0.0008,
  "latency": {
    "p50": "8ms",
    "p95": "45ms",
    "p99": "120ms",
    "avg": "12ms",
    "min": "2ms",
    "max": "850ms"
  },
  "throughput": 167,
  "mttr": "25s",
  "reconnect_success_rate": 0.96
}
```

### 11.2 控制台格式

```
可用性: 99.92% ✓
  - 总操作数: 10,000
  - 成功操作: 9,992
  - 失败操作: 8
  - 错误率: 0.08%

性能指标:
  - P50 延迟: 8ms ✓
  - P95 延迟: 45ms ✓
  - P99 延迟: 120ms ⚠
  - 平均吞吐: 167 ops/s
```

---

## 12. 指标使用示例

### 12.1 评分计算

```go
// 可用性得分 (30分)
availabilityScore := calculateAvailabilityScore(metrics.Availability)

// 性能得分 (25分)
performanceScore := calculatePerformanceScore(metrics.P95Latency, metrics.P99Latency)

// 可靠性得分 (25分)
reliabilityScore := calculateReliabilityScore(metrics.ErrorRate, metrics.DataLossRate)

// 恢复力得分 (20分)
resilienceScore := calculateResilienceScore(metrics.MTTR, metrics.ReconnectSuccessRate)

// 总分
totalScore := availabilityScore*0.3 + performanceScore*0.25 +
              reliabilityScore*0.25 + resilienceScore*0.2
```

### 12.2 问题识别

```go
if metrics.Availability < 0.95 {
    issues.Add(Issue{
        Type: "low_availability",
        Severity: "CRITICAL",
        Current: metrics.Availability,
        Expected: 0.95,
    })
}

if metrics.P99Latency > 500*time.Millisecond {
    issues.Add(Issue{
        Type: "high_p99_latency",
        Severity: "MEDIUM",
        Current: metrics.P99Latency,
        Expected: 500*time.Millisecond,
    })
}
```

---

## 13. 指标扩展

### 13.1 自定义指标

支持添加自定义指标到 `Metadata` 字段：

```go
type StabilityMetrics struct {
    // ... 标准指标

    // 自定义指标
    CustomMetrics map[string]interface{}
}
```

### 13.2 中间件特定指标扩展

每种中间件可以扩展特定的指标字段，通过接口实现多态。

---

**文档版本**: v1.0
**最后更新**: 2025-10-30
