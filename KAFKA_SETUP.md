# Kafka支持设置指南

## 概述

项目已实现完整的Kafka客户端支持，但由于网络限制，kafka-go依赖尚未完全下载。

## 已完成的功能

### ✅ Kafka客户端实现
- **文件**: `internal/middleware/kafka_client.go`
- **功能**:
  - 连接到Kafka集群
  - 生产消息（Produce）
  - 消费消息（Consume）
  - 获取统计信息（消息延迟、积压等）

### ✅ Kafka操作类型
- **文件**: `internal/middleware/kafka_types.go`
- **类型**:
  - `KafkaConfig`: Kafka配置
  - `KafkaProduceOperation`: 生产消息操作
  - `KafkaConsumeOperation`: 消费消息操作

### ✅ CLI工具集成
- **文件**: `cmd/mct/main.go`
- **功能**: `executeKafkaTest()` 函数实现完整的Kafka稳定性测试流程

### ✅ 单元测试
- **文件**: `tests/unit/middleware/kafka_client_test.go`
- **测试覆盖**: 基本功能测试（需要实际Kafka服务）

## 完成设置步骤

### 1. 下载依赖

在有网络连接的环境中运行：

```bash
go mod tidy
go mod download
```

这将下载以下依赖：
- `github.com/segmentio/kafka-go` v0.4.49
- 相关的压缩库（klauspost/compress, pierrec/lz4等）

### 2. 构建项目

```bash
go build -o bin/mct ./cmd/mct
```

### 3. 启动Kafka服务（测试用）

使用Docker快速启动Kafka：

```bash
# 启动Zookeeper
docker run -d --name zookeeper -p 2181:2181 zookeeper:3.8

# 启动Kafka
docker run -d --name kafka \
  -p 9092:9092 \
  -e KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181 \
  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
  --link zookeeper \
  confluentinc/cp-kafka:latest
```

或使用docker-compose（推荐）：

```yaml
version: '3'
services:
  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
    ports:
      - "2181:2181"

  kafka:
    image: confluentinc/cp-kafka:latest
    depends_on:
      - zookeeper
    ports:
      - "9092:9092"
    environment:
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
```

```bash
docker-compose up -d
```

### 4. 运行Kafka测试

```bash
# 运行Kafka稳定性测试
./bin/mct test --middleware kafka \
  --host localhost \
  --port 9092 \
  --duration 30s \
  --operations 1000

# 输出JSON格式
./bin/mct test --middleware kafka \
  --duration 10s \
  --output json

# 输出Markdown格式
./bin/mct test --middleware kafka \
  --duration 10s \
  --output markdown \
  --report-path kafka-report.md
```

## 测试验证

### 运行单元测试

```bash
# 运行所有Kafka测试（需要本地Kafka服务）
go test ./tests/unit/middleware/kafka_client_test.go -v

# 跳过集成测试（不需要Kafka服务）
go test ./tests/unit/middleware/kafka_client_test.go -v -short
```

### 验证构建

```bash
# 验证构建成功
go build ./cmd/mct

# 查看帮助
./mct --help
./mct test --help
```

## Kafka特定指标

测试会收集以下Kafka特定指标：

- **消息积压 (MessageLag)**: 未消费的消息数量
- **消费延迟 (ConsumerLag)**: 生产和消费的时间差
- **重复消息数 (DuplicateMessages)**: 检测到的重复消息
- **重平衡次数 (RebalanceCount)**: 消费者组重平衡次数

这些指标会在评分系统中被使用，通过`EvaluateKafka()`方法进行Kafka特定的评估。

## 示例输出

```
Starting kafka stability test...
Target: localhost:9092
Duration: 30s
Operations: 1000

==========================================
   中间件稳定性测试报告
==========================================

测试目标: Kafka @ localhost:9092
测试时长: 30s
测试完成: 2025-10-31 15:30:00

------------------------------------------
  总体评分: 88.5/100 (GOOD) ✅ PASS
------------------------------------------

各维度得分:
  ✓ 可用性   27.0/30  (90.0%)  - 权重30%
  ✓ 性能     22.5/25  (90.0%)  - 权重25%
  ✓ 可靠性   23.0/25  (92.0%)  - 权重25%
  ✓ 恢复力   16.0/20  (80.0%)  - 权重20%
...
```

## 故障排查

### 问题: 无法连接到Kafka

**解决方案**:
1. 确认Kafka服务正在运行: `docker ps | grep kafka`
2. 检查端口是否可访问: `nc -zv localhost 9092`
3. 查看Kafka日志: `docker logs kafka`

### 问题: 消息积压严重

**原因**: 消费速度慢于生产速度

**解决方案**:
1. 增加消费者实例数
2. 优化消费逻辑
3. 调整batch size和fetch配置

### 问题: 依赖下载失败

**解决方案**:
1. 配置Go代理: `export GOPROXY=https://goproxy.cn,direct`
2. 使用国内镜像: `export GOPROXY=https://goproxy.io,direct`
3. 离线下载依赖包并手动安装

## 下一步

1. ✅ Kafka客户端实现完成
2. ⏳ 需要在有网络环境下完成依赖下载
3. ⏳ 在实际Kafka环境中测试和验证
4. ⏳ 根据测试结果调整评分阈值
5. ⏳ 添加更多Kafka特定的混沌测试场景

## 参考文档

- [kafka-go官方文档](https://github.com/segmentio/kafka-go)
- [Kafka官方文档](https://kafka.apache.org/documentation/)
- [项目PLAN.md](./PLAN.md) - 完整的开发计划
