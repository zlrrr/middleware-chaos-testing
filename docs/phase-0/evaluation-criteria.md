# 评分标准详细说明

## 1. 评分体系概述

### 1.1 总体评分结构

```
总分 (0-100分) = 可用性得分 × 30% + 性能得分 × 25% + 可靠性得分 × 25% + 恢复力得分 × 20%
```

### 1.2 评分等级

| 分数区间 | 等级 | 英文 | 状态 | 说明 |
|---------|------|------|------|------|
| 90-100 | 优秀 | EXCELLENT | PASS | 可直接用于生产环境 |
| 80-89 | 良好 | GOOD | PASS | 满足生产要求 |
| 70-79 | 一般 | FAIR | WARNING | 建议优化后使用 |
| 60-69 | 较差 | POOR | WARNING | 需要改进 |
| 0-59 | 失败 | FAILED | FAIL | 不建议用于生产 |

### 1.3 测试状态判定

```
IF 存在 CRITICAL 问题:
    返回 FAIL

IF 总分 < 70:
    返回 FAIL

IF 存在 HIGH 问题:
    返回 WARNING

IF 总分 < 85:
    返回 WARNING

OTHERWISE:
    返回 PASS
```

---

## 2. 可用性评分 (30分)

### 2.1 评分规则

可用性评分完全基于可用性百分比：

| 可用性 | 得分 | 等级 | 说明 |
|--------|------|------|------|
| >= 99.99% | 30.0 | EXCELLENT | 四个9，生产级别 |
| >= 99.9% | 27.0 | GOOD | 三个9，良好水平 |
| >= 99.0% | 24.0 | FAIR | 两个9，基本可用 |
| >= 95.0% | 20.0 | POOR | 最低要求 |
| < 95.0% | 比例计算 | FAILED | 不可接受 |

### 2.2 计算公式

```go
func calculateAvailabilityScore(availability float64) float64 {
    switch {
    case availability >= 0.9999:
        return 30.0
    case availability >= 0.999:
        return 27.0
    case availability >= 0.99:
        return 24.0
    case availability >= 0.95:
        return 20.0
    default:
        // 低于95%按比例计算
        return availability * 100 * 0.2
    }
}
```

### 2.3 问题识别

**CRITICAL 级别**:
- 可用性 < 95%

**示例**:
```
[CRITICAL] low_availability
  当前值: 94.5%
  期望值: >= 95.0%
  说明: 可用性低于最低要求，系统不稳定
```

### 2.4 改进建议

可用性低时的建议：

**优先级: HIGH**
- 检查服务健康状态，排查频繁失败原因
- 增加实例数量，实现高可用部署
- 配置健康检查和自动重启机制
- 实施熔断和降级策略
- 检查网络连接稳定性

---

## 3. 性能评分 (25分)

### 3.1 评分组成

性能评分由两部分组成：
- **P95延迟得分**: 15分 (60%)
- **P99延迟得分**: 10分 (40%)

### 3.2 P95延迟评分 (15分)

| P95延迟 | 得分 | 等级 | 说明 |
|---------|------|------|------|
| <= 10ms | 15.0 | EXCELLENT | 极优延迟 |
| <= 50ms | 13.5 | GOOD | 良好延迟 |
| <= 100ms | 12.0 | FAIR | 可接受延迟 |
| <= 200ms | 10.0 | POOR | 偏高延迟 |
| > 200ms | 8.0 | FAILED | 超出阈值 |

**计算公式**:
```go
func calculateP95Score(p95 time.Duration) float64 {
    switch {
    case p95 <= 10*time.Millisecond:
        return 15.0
    case p95 <= 50*time.Millisecond:
        return 13.5
    case p95 <= 100*time.Millisecond:
        return 12.0
    case p95 <= 200*time.Millisecond:
        return 10.0
    default:
        return 8.0
    }
}
```

### 3.3 P99延迟评分 (10分)

| P99延迟 | 得分 | 等级 | 说明 |
|---------|------|------|------|
| <= 20ms | 10.0 | EXCELLENT | 极优尾延迟 |
| <= 100ms | 9.0 | GOOD | 良好尾延迟 |
| <= 200ms | 8.0 | FAIR | 可接受尾延迟 |
| <= 500ms | 6.5 | POOR | 偏高尾延迟 |
| > 500ms | 5.0 | FAILED | 超出阈值 |

**计算公式**:
```go
func calculateP99Score(p99 time.Duration) float64 {
    switch {
    case p99 <= 20*time.Millisecond:
        return 10.0
    case p99 <= 100*time.Millisecond:
        return 9.0
    case p99 <= 200*time.Millisecond:
        return 8.0
    case p99 <= 500*time.Millisecond:
        return 6.5
    default:
        return 5.0
    }
}
```

### 3.4 问题识别

**HIGH 级别**:
- P95延迟 > 200ms

**MEDIUM 级别**:
- P99延迟 > 500ms
- P95延迟 > 100ms 且 <= 200ms

**示例**:
```
[HIGH] high_p95_latency
  当前值: 250ms
  期望值: <= 200ms
  说明: P95延迟超过阈值，影响用户体验
```

### 3.5 改进建议

延迟过高时的建议：

**优先级: MEDIUM-HIGH**
- 分析慢查询日志，优化热点操作
- 检查网络延迟和带宽瓶颈
- 优化数据结构和查询模式
- 考虑增加缓存层或读写分离
- 评估硬件资源是否充足（CPU/内存/磁盘IO）
- 检查是否存在锁竞争

---

## 4. 可靠性评分 (25分)

### 4.1 评分组成

可靠性评分由两部分组成：
- **错误率得分**: 15分 (60%)
- **数据丢失率得分**: 10分 (40%)

### 4.2 错误率评分 (15分)

| 错误率 | 得分 | 等级 | 说明 |
|--------|------|------|------|
| <= 0.01% | 15.0 | EXCELLENT | 极低错误率 |
| <= 0.1% | 13.5 | GOOD | 良好错误率 |
| <= 0.5% | 12.0 | FAIR | 可接受错误率 |
| <= 1.0% | 10.0 | POOR | 偏高错误率 |
| > 1.0% | 7.0 | FAILED | 超出阈值 |

**计算公式**:
```go
func calculateErrorRateScore(errorRate float64) float64 {
    switch {
    case errorRate <= 0.0001:
        return 15.0
    case errorRate <= 0.001:
        return 13.5
    case errorRate <= 0.005:
        return 12.0
    case errorRate <= 0.01:
        return 10.0
    default:
        return 7.0
    }
}
```

### 4.3 数据丢失率评分 (10分)

| 数据丢失率 | 得分 | 等级 | 说明 |
|-----------|------|------|------|
| == 0% | 10.0 | EXCELLENT | 无数据丢失 |
| < 0.01% | 8.0 | GOOD | 极少量丢失 |
| < 0.1% | 6.0 | FAIR | 少量丢失 |
| >= 0.1% | 3.0 | FAILED | 数据丢失严重 |

**计算公式**:
```go
func calculateDataLossScore(dataLossRate float64) float64 {
    switch {
    case dataLossRate == 0:
        return 10.0
    case dataLossRate < 0.0001:
        return 8.0
    case dataLossRate < 0.001:
        return 6.0
    default:
        return 3.0
    }
}
```

### 4.4 问题识别

**CRITICAL 级别**:
- 数据丢失率 >= 0.1%

**HIGH 级别**:
- 错误率 > 1.0%
- 数据丢失率 > 0 且 < 0.1%

**示例**:
```
[CRITICAL] data_loss_detected
  当前值: 0.15%
  期望值: 0%
  说明: 检测到数据丢失，需立即处理
```

### 4.5 改进建议

**数据丢失** (优先级: HIGH):
- 检查持久化配置（RDB/AOF设置）
- 确保有足够的副本数
- 配置合适的fsync策略
- 实施数据校验机制
- 检查磁盘空间和IO性能

**高错误率** (优先级: HIGH):
- 查看错误日志，分析错误类型
- 检查客户端配置（超时、重试设置）
- 验证服务端配置
- 实施更好的错误处理和重试逻辑
- 检查网络稳定性

---

## 5. 恢复力评分 (20分)

### 5.1 评分组成

恢复力评分由两部分组成：
- **MTTR得分**: 12分 (60%)
- **重连成功率得分**: 8分 (40%)

### 5.2 MTTR评分 (12分)

| MTTR | 得分 | 等级 | 说明 |
|------|------|------|------|
| <= 5s | 12.0 | EXCELLENT | 极快恢复 |
| <= 30s | 10.5 | GOOD | 快速恢复 |
| <= 60s | 9.0 | FAIR | 正常恢复 |
| <= 300s | 7.0 | POOR | 恢复较慢 |
| > 300s | 5.0 | FAILED | 恢复过慢 |

**计算公式**:
```go
func calculateMTTRScore(mttr time.Duration) float64 {
    switch {
    case mttr <= 5*time.Second:
        return 12.0
    case mttr <= 30*time.Second:
        return 10.5
    case mttr <= 60*time.Second:
        return 9.0
    case mttr <= 300*time.Second:
        return 7.0
    default:
        return 5.0
    }
}
```

### 5.3 重连成功率评分 (8分)

| 重连成功率 | 得分 | 等级 | 说明 |
|-----------|------|------|------|
| >= 99% | 8.0 | EXCELLENT | 极高重连率 |
| >= 95% | 7.0 | GOOD | 良好重连率 |
| >= 90% | 6.0 | FAIR | 可接受重连率 |
| < 90% | 4.0 | POOR | 重连率偏低 |

**计算公式**:
```go
func calculateReconnectScore(reconnectRate float64) float64 {
    switch {
    case reconnectRate >= 0.99:
        return 8.0
    case reconnectRate >= 0.95:
        return 7.0
    case reconnectRate >= 0.90:
        return 6.0
    default:
        return 4.0
    }
}
```

### 5.4 问题识别

**MEDIUM 级别**:
- MTTR > 300s
- 重连成功率 < 90%

**示例**:
```
[MEDIUM] slow_recovery
  当前值: 350s
  期望值: <= 300s
  说明: 平均恢复时间过长，影响系统可用性
```

### 5.5 改进建议

**恢复慢** (优先级: MEDIUM):
- 优化健康检查频率和超时设置
- 实施更激进的重试策略
- 增加备用连接池
- 优化故障检测算法
- 考虑实施预热连接

**重连率低** (优先级: MEDIUM):
- 检查网络稳定性
- 调整重连间隔和最大重试次数
- 实施指数退避算法
- 检查服务端连接限制
- 优化连接池配置

---

## 6. 综合评分示例

### 6.1 示例1: 优秀系统

**指标**:
- 可用性: 99.99%
- P95延迟: 8ms
- P99延迟: 18ms
- 错误率: 0.005%
- 数据丢失率: 0%
- MTTR: 4s
- 重连成功率: 99.5%

**评分**:
- 可用性得分: 30.0 / 30
- 性能得分: 15.0 + 10.0 = 25.0 / 25
- 可靠性得分: 15.0 + 10.0 = 25.0 / 25
- 恢复力得分: 12.0 + 8.0 = 20.0 / 20

**总分**: 100 / 100

**等级**: EXCELLENT

**状态**: PASS

---

### 6.2 示例2: 良好系统

**指标**:
- 可用性: 99.92%
- P95延迟: 45ms
- P99延迟: 120ms
- 错误率: 0.08%
- 数据丢失率: 0%
- MTTR: 25s
- 重连成功率: 96%

**评分**:
- 可用性得分: 27.0 / 30
- 性能得分: 13.5 + 9.0 = 22.5 / 25
- 可靠性得分: 13.5 + 10.0 = 23.5 / 25
- 恢复力得分: 10.5 + 7.0 = 17.5 / 20

**总分**: 90.5 / 100 → 实际 27×0.3 + 22.5×0.25 + 23.5×0.25 + 17.5×0.2 = 87.5 / 100

**等级**: GOOD

**状态**: PASS

**识别问题**: P99延迟略高 (MEDIUM)

---

### 6.3 示例3: 需要警告的系统

**指标**:
- 可用性: 98.5%
- P95延迟: 80ms
- P99延迟: 250ms
- 错误率: 1.5%
- 数据丢失率: 0.01%
- MTTR: 120s
- 重连成功率: 88%

**评分**:
- 可用性得分: 24.0 / 30
- 性能得分: 13.5 + 8.0 = 21.5 / 25
- 可靠性得分: 7.0 + 8.0 = 15.0 / 25
- 恢复力得分: 9.0 + 4.0 = 13.0 / 20

**总分**: 73.5 / 100 → 实际 24×0.3 + 21.5×0.25 + 15×0.25 + 13×0.2 = 73.725 / 100

**等级**: FAIR

**状态**: WARNING

**识别问题**:
- 高错误率 (HIGH)
- 数据丢失 (HIGH)
- 重连率低 (MEDIUM)

---

### 6.4 示例4: 失败的系统

**指标**:
- 可用性: 92.0%
- P95延迟: 350ms
- P99延迟: 800ms
- 错误率: 8.0%
- 数据丢失率: 0.5%
- MTTR: 600s
- 重连成功率: 75%

**评分**:
- 可用性得分: 18.4 / 30 (92×100×0.2)
- 性能得分: 8.0 + 5.0 = 13.0 / 25
- 可靠性得分: 7.0 + 3.0 = 10.0 / 25
- 恢复力得分: 5.0 + 4.0 = 9.0 / 20

**总分**: 50.4 / 100 → 实际 18.4×0.3 + 13×0.25 + 10×0.25 + 9×0.2 = 50.02 / 100

**等级**: FAILED

**状态**: FAIL

**识别问题**:
- 低可用性 (CRITICAL)
- 数据丢失严重 (CRITICAL)
- 高延迟 (HIGH)
- 高错误率 (HIGH)
- 恢复慢 (MEDIUM)

---

## 7. 边界情况处理

### 7.1 刚好及格 (70分)

```
IF 总分 == 70.0:
    等级 = FAIR
    状态 = WARNING (因为 < 85)
```

### 7.2 临界PASS/WARNING (85分)

```
IF 总分 == 85.0:
    等级 = GOOD
    状态 = WARNING (因为 < 85)
```

**说明**: 状态判断使用严格小于 (<)，因此85分刚好为WARNING

### 7.3 临界WARNING/FAIL (70分)

```
IF 总分 == 70.0:
    等级 = FAIR
    状态 = WARNING
```

---

## 8. 问题严重程度定义

### 8.1 CRITICAL

**定义**: 严重影响系统可用性或数据安全的问题

**触发条件**:
- 可用性 < 95%
- 数据丢失率 >= 0.1%
- 系统完全不可用

**影响**: 自动触发 FAIL 状态

---

### 8.2 HIGH

**定义**: 严重影响系统性能或可靠性的问题

**触发条件**:
- 错误率 > 1.0%
- P95延迟 > 200ms
- 数据丢失率 > 0 且 < 0.1%

**影响**: 自动触发 WARNING 状态

---

### 8.3 MEDIUM

**定义**: 对系统有一定影响但不严重的问题

**触发条件**:
- P99延迟 > 500ms
- MTTR > 300s
- 重连成功率 < 90%
- P95延迟 > 100ms 且 <= 200ms

**影响**: 不影响状态判定，但会生成建议

---

### 8.4 LOW

**定义**: 轻微问题，可能影响优化

**触发条件**:
- 缓存命中率 < 80% (Redis)
- 消息积压 > 1000 (Kafka)

**影响**: 仅生成建议

---

## 9. 改进建议生成规则

### 9.1 建议优先级

建议按优先级排序：
1. **HIGH**: 关键问题，必须解决
2. **MEDIUM**: 重要问题，建议尽快解决
3. **LOW**: 优化建议，可选

### 9.2 建议内容结构

```go
type Recommendation struct {
    Priority   string   // HIGH, MEDIUM, LOW
    Category   string   // CONFIGURATION, SCALING, OPTIMIZATION
    Title      string   // 简短标题
    Message    string   // 问题描述
    Actions    []string // 具体行动项（3-5条）
    References []string // 参考文档链接
}
```

### 9.3 建议生成映射

| 问题类型 | 优先级 | 类别 | 标题 |
|---------|--------|------|------|
| low_availability | HIGH | SCALING | 提高系统可用性 |
| high_error_rate | HIGH | CONFIGURATION | 降低错误率 |
| data_loss_detected | HIGH | CONFIGURATION | 防止数据丢失 |
| high_p95_latency | MEDIUM | OPTIMIZATION | 优化响应延迟 |
| high_p99_latency | MEDIUM | OPTIMIZATION | 优化尾部延迟 |
| slow_recovery | MEDIUM | OPTIMIZATION | 加快故障恢复 |
| low_reconnect_rate | MEDIUM | CONFIGURATION | 提高重连成功率 |

---

## 10. 判断依据生成

### 10.1 依据格式

```
综合评分: {score}/100 ({grade})

各维度得分:
- 可用性: {availability_score}/30 (权重30%)
- 性能: {performance_score}/25 (权重25%)
- 可靠性: {reliability_score}/25 (权重25%)
- 恢复力: {resilience_score}/20 (权重20%)

{状态说明}

{问题总结}
```

### 10.2 状态说明

**PASS**:
```
✅ 测试通过: 系统稳定性符合预期，可以用于生产环境。
```

**WARNING**:
```
⚠️  警告: 系统存在需要关注的问题，建议优化后再部署。
```

**FAIL**:
```
❌ 测试失败: 系统稳定性不满足最低要求，不建议用于生产环境。
```

---

## 11. 中间件特定评分

### 11.1 Redis特定检查

**缓存命中率低** (< 90%):
- 优先级: MEDIUM
- 类别: OPTIMIZATION
- 建议: 提高缓存命中率

**内存使用率高** (> 80%):
- 优先级: MEDIUM
- 类别: SCALING
- 建议: 增加内存或优化数据结构

### 11.2 Kafka特定检查

**消息积压高** (> 1000):
- 优先级: MEDIUM
- 类别: SCALING
- 建议: 增加消费者或优化消费逻辑

**频繁重平衡** (> 3次):
- 优先级: MEDIUM
- 类别: CONFIGURATION
- 建议: 优化消费者配置

---

## 12. 配置化阈值

支持通过配置文件自定义所有阈值：

```yaml
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

  mttr:
    excellent: 5s
    good: 30s
    fair: 60s
    pass: 300s

  reconnect_rate:
    excellent: 99%
    good: 95%
    fair: 90%
```

---

## 13. 评分算法验证

### 13.1 单元测试用例

所有评分边界都需要有测试用例覆盖：

```go
// 测试可用性边界
func TestAvailabilityBoundaries(t *testing.T) {
    cases := []struct{
        availability float64
        expectedScore float64
        expectedGrade string
    }{
        {0.9999, 30.0, "EXCELLENT"},
        {0.999, 27.0, "GOOD"},
        {0.99, 24.0, "FAIR"},
        {0.95, 20.0, "POOR"},
        {0.94, 18.8, "FAILED"}, // 94 * 100 * 0.2
    }

    for _, tc := range cases {
        score := calculateAvailabilityScore(tc.availability)
        assert.Equal(t, tc.expectedScore, score)
    }
}
```

### 13.2 集成测试

验证完整的评分流程：
1. 模拟各种指标组合
2. 验证总分计算正确
3. 验证等级和状态判定正确
4. 验证问题识别正确
5. 验证建议生成正确

---

**文档版本**: v1.0
**最后更新**: 2025-10-30
