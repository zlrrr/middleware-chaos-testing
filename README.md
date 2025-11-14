# Middleware Chaos Testing (MCT)

**English** | [中文](#中文文档)

A comprehensive chaos testing framework for middleware stability validation across Redis, Kafka, MongoDB, RocketMQ, RabbitMQ, EMQX, and Nacos.

[![Go Version](https://img.shields.io/badge/go-1.23+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## Overview

MCT is an extensible middleware chaos testing framework that simulates various user operations to detect and quantify middleware service stability metrics under chaotic scenarios.

### Core Features

- ✅ **Multi-Middleware Support**: Redis, Kafka, MongoDB, RocketMQ, RabbitMQ, EMQX, Nacos
- ✅ **Configurable Test Duration**: Command-line and configuration file support
- ✅ **Intelligent Scoring System**: 0-100 points with 5 grade levels
- ✅ **Clear Test Results**: Pass/Warning/Fail status with actionable insights
- ✅ **Prioritized Recommendations**: Improvement suggestions ranked by priority
- ✅ **Production-Grade Metrics**: Industry-standard stability indicators
- ✅ **Containerized Deployment**: One-command Docker setup
- ✅ **TDD Methodology**: Test-Driven Development workflow

## Table of Contents

- [Quick Start](#quick-start)
- [Compilation & Build](#compilation--build)
- [Running Tests](#running-tests)
  - [Redis Testing](#redis-testing)
  - [Kafka Testing](#kafka-testing)
  - [MongoDB Testing](#mongodb-testing)
  - [RocketMQ Testing](#rocketmq-testing)
  - [RabbitMQ Testing](#rabbitmq-testing)
  - [EMQX Testing](#emqx-testing)
  - [Nacos Testing](#nacos-testing)
- [Running Test Cases](#running-test-cases)
- [Configuration](#configuration)
- [Scoring System](#scoring-system)
- [Development](#development)

---

## Quick Start

### Prerequisites

- **Go**: 1.23 or higher
- **Docker**: 20.10+ (for running middleware containers)
- **Docker Compose**: 2.0+ (optional, for multi-service setup)

### Installation

```bash
# Clone the repository
git clone https://github.com/username/middleware-chaos-testing.git
cd middleware-chaos-testing

# Download dependencies
go mod download

# Build the project
go build -o bin/mct ./cmd/mct
```

---

## Compilation & Build

### Build from Source

```bash
# Build for current platform
go build -o bin/mct ./cmd/mct

# Build with optimizations (smaller binary)
go build -ldflags="-s -w" -o bin/mct ./cmd/mct

# Cross-compile for Linux (from macOS/Windows)
GOOS=linux GOARCH=amd64 go build -o bin/mct-linux ./cmd/mct

# Cross-compile for Windows
GOOS=windows GOARCH=amd64 go build -o bin/mct.exe ./cmd/mct
```

### Build with Makefile (if available)

```bash
# Build binary
make build

# Build and install to $GOPATH/bin
make install

# Clean build artifacts
make clean
```

### Verify Build

```bash
# Check binary
./bin/mct version

# Show help
./bin/mct --help
```

---

## Running Tests

### Redis Testing

**Supported Versions**: Redis 6.x, 7.x

#### Start Redis Server

```bash
# Using Docker
docker run -d --name redis-test \
  -p 6379:6379 \
  redis:7-alpine

# Or using Docker Compose
docker-compose up -d redis
```

#### Run Stability Test

```bash
# Basic test (30 seconds, 5000 operations)
./bin/mct test \
  --middleware redis \
  --host localhost \
  --port 6379 \
  --duration 30s \
  --operations 5000

# Advanced test with custom configuration
./bin/mct test \
  --middleware redis \
  --host localhost \
  --port 6379 \
  --duration 60s \
  --operations 10000 \
  --concurrency 10 \
  --password "your-redis-password"
```

#### Redis Test Operations

- `SET`: Write key-value pairs
- `GET`: Read values by key
- `DEL`: Delete keys
- `INCR`: Increment counters
- `EXPIRE`: Set key expiration

---

### Kafka Testing

**Supported Versions**: Kafka 2.7.2

#### Start Kafka Cluster

```bash
# Using Docker Compose (with ZooKeeper)
docker-compose up -d zookeeper kafka

# Or standalone Kafka (KRaft mode)
docker run -d --name kafka-test \
  -p 9092:9092 \
  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  apache/kafka:latest
```

#### Run Stability Test

```bash
# Basic test
./bin/mct test \
  --middleware kafka \
  --brokers localhost:9092 \
  --topic test-topic \
  --duration 30s \
  --operations 5000

# Advanced test with consumer groups
./bin/mct test \
  --middleware kafka \
  --brokers localhost:9092,localhost:9093 \
  --topic test-topic \
  --producer-group test-producer \
  --consumer-group test-consumer \
  --duration 60s \
  --operations 10000 \
  --concurrency 5
```

#### Kafka Test Operations

- `PRODUCE`: Publish messages to topics
- `CONSUME`: Subscribe and consume messages
- `COMMIT`: Commit consumer offsets

---

### MongoDB Testing

**Supported Versions**: MongoDB 4.4.13

#### Start MongoDB Server

```bash
# Using Docker
docker run -d --name mongodb-test \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=password \
  mongo:4.4.13

# Or using Docker Compose
docker-compose up -d mongodb
```

#### Run Stability Test

```bash
# Basic test
./bin/mct test \
  --middleware mongodb \
  --host localhost \
  --port 27017 \
  --database testdb \
  --collection testcol \
  --duration 30s \
  --operations 5000

# With authentication
./bin/mct test \
  --middleware mongodb \
  --uri "mongodb://admin:password@localhost:27017/testdb?authSource=admin" \
  --collection testcol \
  --duration 60s \
  --operations 10000 \
  --concurrency 10
```

#### MongoDB Test Operations

- `INSERT`: Insert documents
- `FIND`: Query documents
- `UPDATE`: Update documents
- `DELETE`: Delete documents
- `AGGREGATE`: Run aggregation pipelines

---

### RocketMQ Testing

**Supported Versions**: RocketMQ 5.1.3
**Protocols**: Remoting (default), gRPC

#### Start RocketMQ Server

```bash
# Using Docker
docker run -d --name rocketmq-namesrv \
  -p 9876:9876 \
  apache/rocketmq:5.1.3 sh mqnamesrv

docker run -d --name rocketmq-broker \
  -p 10909:10909 -p 10911:10911 \
  -e NAMESRV_ADDR=localhost:9876 \
  apache/rocketmq:5.1.3 sh mqbroker
```

#### Run Stability Test

```bash
# Basic test (Remoting protocol - default)
./bin/mct test \
  --middleware rocketmq \
  --nameservers localhost:9876 \
  --topic test-topic \
  --duration 30s \
  --operations 5000

# Using gRPC protocol (RocketMQ 5.x recommended)
./bin/mct test \
  --middleware rocketmq \
  --nameservers localhost:9876 \
  --topic test-topic \
  --protocol grpc \
  --producer-group test-producer \
  --consumer-group test-consumer \
  --duration 60s \
  --operations 10000
```

#### RocketMQ Test Operations

- `SEND`: Send messages to topics
- `RECEIVE`: Consume messages from queues
- Protocol selection: `remoting` (compatible) or `grpc` (high performance)

---

### RabbitMQ Testing

**Supported Versions**: RabbitMQ 3.12.7

#### Start RabbitMQ Server

```bash
# Using Docker with management plugin
docker run -d --name rabbitmq-test \
  -p 5672:5672 \
  -p 15672:15672 \
  -e RABBITMQ_DEFAULT_USER=admin \
  -e RABBITMQ_DEFAULT_PASS=password \
  rabbitmq:3.12.7-management

# Access management UI: http://localhost:15672
```

#### Run Stability Test

```bash
# Basic test
./bin/mct test \
  --middleware rabbitmq \
  --host localhost \
  --port 5672 \
  --user admin \
  --password password \
  --duration 30s \
  --operations 5000

# Advanced test with routing
./bin/mct test \
  --middleware rabbitmq \
  --uri "amqp://admin:password@localhost:5672/" \
  --exchange test-exchange \
  --queue test-queue \
  --routing-key test-key \
  --duration 60s \
  --operations 10000 \
  --concurrency 10
```

#### RabbitMQ Test Operations

- `PUBLISH`: Publish messages to exchanges
- `CONSUME`: Consume messages from queues
- `ACK`: Acknowledge message delivery
- Supports: Direct, Topic, Fanout exchanges

---

### EMQX Testing

**Supported Versions**: EMQX 5.8 Community Edition
**Protocol**: MQTT 3.1.1, MQTT 5.0

#### Start EMQX Server

```bash
# Using Docker
docker run -d --name emqx-test \
  -p 1883:1883 \
  -p 18083:18083 \
  emqx/emqx:5.8

# Access dashboard: http://localhost:18083
# Default credentials: admin / public
```

#### Run Stability Test

```bash
# Basic MQTT test
./bin/mct test \
  --middleware emqx \
  --broker tcp://localhost:1883 \
  --client-id test-client \
  --topic test/topic \
  --duration 30s \
  --operations 5000

# MQTT with QoS levels
./bin/mct test \
  --middleware emqx \
  --broker tcp://localhost:1883 \
  --topic test/topic \
  --qos 1 \
  --clean-session true \
  --duration 60s \
  --operations 10000

# MQTT over TLS
./bin/mct test \
  --middleware emqx \
  --broker ssl://localhost:8883 \
  --topic test/topic \
  --ca-cert /path/to/ca.crt \
  --client-cert /path/to/client.crt \
  --client-key /path/to/client.key \
  --duration 60s
```

#### EMQX Test Operations

- `PUBLISH`: Publish MQTT messages
- `SUBSCRIBE`: Subscribe to topics
- QoS levels: 0 (At most once), 1 (At least once), 2 (Exactly once)
- Features: Retained messages, Last Will, Clean session
- Shared subscriptions (EMQX 5.8 feature)

---

### Nacos Testing

**Supported Versions**: Nacos 2.4.3
**Features**: Service Discovery, Configuration Management

#### Start Nacos Server

```bash
# Using Docker (standalone mode)
docker run -d --name nacos-test \
  -p 8848:8848 \
  -p 9848:9848 \
  -e MODE=standalone \
  nacos/nacos-server:v2.4.3

# Access console: http://localhost:8848/nacos
# Default credentials: nacos / nacos
```

#### Run Stability Test

```bash
# Service Registry test
./bin/mct test \
  --middleware nacos \
  --servers localhost:8848 \
  --service-name test-service \
  --namespace public \
  --duration 30s \
  --operations 5000

# Configuration Management test
./bin/mct test \
  --middleware nacos \
  --servers localhost:8848 \
  --data-id test-config \
  --group DEFAULT_GROUP \
  --namespace public \
  --duration 60s \
  --operations 10000

# With authentication (Nacos 2.x)
./bin/mct test \
  --middleware nacos \
  --servers localhost:8848 \
  --service-name test-service \
  --username nacos \
  --password nacos \
  --namespace dev \
  --duration 60s
```

#### Nacos Test Operations

- **Service Registry**:
  - `REGISTER`: Register service instances
  - `DEREGISTER`: Deregister instances
  - `DISCOVER`: Query service instances
  - `SUBSCRIBE`: Watch service changes
  - `HEARTBEAT`: Send heartbeat (ephemeral instances)
- **Configuration Center**:
  - `PUBLISH_CONFIG`: Publish configuration
  - `GET_CONFIG`: Retrieve configuration
  - `REMOVE_CONFIG`: Delete configuration
  - `LISTEN_CONFIG`: Listen for config changes

---

## Running Test Cases

### Unit Tests

Run all unit tests for middleware clients:

```bash
# Run all unit tests
go test ./tests/unit/... -v

# Run tests for specific middleware
go test ./tests/unit/middleware/redis_client_test.go -v
go test ./tests/unit/middleware/kafka_client_test.go -v
go test ./tests/unit/middleware/mongodb_client_test.go -v
go test ./tests/unit/middleware/rocketmq_client_test.go -v
go test ./tests/unit/middleware/rabbitmq_client_test.go -v
go test ./tests/unit/middleware/emqx_client_test.go -v
go test ./tests/unit/middleware/nacos_client_test.go -v

# Run with coverage
go test ./tests/unit/... -cover -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out
```

### Integration Tests

```bash
# Run integration tests (requires running middleware services)
go test ./tests/integration/... -v

# Run specific integration test
go test ./tests/integration/redis_integration_test.go -v
```

### End-to-End Tests

```bash
# Run E2E tests
go test ./tests/e2e/... -v

# Run with timeout
go test ./tests/e2e/... -v -timeout 10m
```

### Test Makefile Targets (if available)

```bash
# Run unit tests
make test-unit

# Run integration tests
make test-integration

# Run all tests
make test

# Generate coverage report
make coverage

# Run tests with race detector
make test-race
```

---

## Configuration

### YAML Configuration File

Create a configuration file for repeatable tests:

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
  duration: 60s
  operations: 10000
  concurrency: 10

operations:
  - type: "SET"
    weight: 40
  - type: "GET"
    weight: 40
  - type: "DEL"
    weight: 10
  - type: "INCR"
    weight: 10

thresholds:
  availability:
    excellent: 99.99
    good: 99.9
    fair: 99.0
    pass: 95.0
  p95_latency:
    excellent: 10ms
    good: 50ms
    fair: 100ms
    pass: 200ms
  error_rate:
    excellent: 0.01
    good: 0.1
    fair: 0.5
    pass: 1.0
```

### Run with Configuration File

```bash
./bin/mct test --config configs/test-redis.yaml
```

---

## Scoring System

MCT uses a 0-100 point intelligent scoring system:

| Score | Grade | Status | Description |
|-------|-------|--------|-------------|
| 90-100 | EXCELLENT | ✅ PASS | Excellent, production-ready |
| 80-89 | GOOD | ✅ PASS | Good, meets production standards |
| 70-79 | FAIR | ⚠️ WARNING | Fair, optimization recommended |
| 60-69 | POOR | ⚠️ WARNING | Poor, improvements needed |
| 0-59 | FAILED | ❌ FAIL | Failed, not recommended for production |

### Scoring Dimensions

```
Total Score = Availability (30%) + Performance (25%) + Reliability (25%) + Resilience (20%)
```

**Availability (30 points)**:
- Uptime percentage
- Success rate of operations

**Performance (25 points)**:
- P95 latency (15 points)
- P99 latency (10 points)

**Reliability (25 points)**:
- Error rate (15 points)
- Data loss rate (10 points)

**Resilience (20 points)**:
- Mean Time To Recovery (MTTR) (12 points)
- Reconnect success rate (8 points)

See [Evaluation Criteria](docs/phase-0/evaluation-criteria.md) for details.

---

## Development

### Development Environment Setup

```bash
# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Download dependencies
go mod download

# Start development environment (all middleware)
docker-compose up -d
```

### Code Quality

```bash
# Format code
go fmt ./...
goimports -w .

# Lint code
golangci-lint run

# Run all checks
go fmt ./... && golangci-lint run && go test ./tests/unit/... -cover
```

### Adding a New Middleware Client

1. Create client interface implementation in `internal/middleware/`
2. Create type definitions in `internal/middleware/{name}_types.go`
3. Create client implementation in `internal/middleware/{name}_client.go`
4. Add threshold function in `internal/evaluator/stability_evaluator.go`
5. Write 15+ unit tests in `tests/unit/middleware/{name}_client_test.go`
6. Update documentation

### Project Structure

```
middleware-chaos-testing/
├── cmd/
│   └── mct/                  # Main program entry
├── internal/
│   ├── core/                 # Core abstractions
│   ├── middleware/           # Middleware adapters (7 clients)
│   │   ├── redis_client.go
│   │   ├── kafka_client.go
│   │   ├── mongodb_client.go
│   │   ├── rocketmq_client.go
│   │   ├── rabbitmq_client.go
│   │   ├── emqx_client.go
│   │   └── nacos_client.go
│   ├── metrics/              # Metrics collector
│   ├── detector/             # Stability detector
│   ├── evaluator/            # Stability evaluator
│   ├── reporter/             # Report generator
│   ├── orchestrator/         # Test orchestrator
│   └── config/               # Configuration manager
├── tests/
│   ├── unit/                 # Unit tests (15+ per middleware)
│   ├── integration/          # Integration tests
│   └── e2e/                  # End-to-end tests
├── docs/                     # Documentation
├── configs/                  # Configuration examples
├── scripts/                  # Build & deployment scripts
├── Makefile                  # Build automation
├── Dockerfile                # Container image
└── docker-compose.yml        # Multi-service orchestration
```

---

## Supported Middleware Versions

| Middleware | Version | Client Library | Key Features |
|------------|---------|----------------|--------------|
| **Redis** | 6.x, 7.x | `github.com/redis/go-redis/v9` | KV operations, pipelining |
| **Kafka** | 2.7.2 | `github.com/segmentio/kafka-go` | Producer/Consumer, consumer groups |
| **MongoDB** | 4.4.13 | `go.mongodb.org/mongo-driver` | CRUD, aggregations |
| **RocketMQ** | 5.1.3 | `github.com/apache/rocketmq-client-go/v2` | **Remoting/gRPC protocols** |
| **RabbitMQ** | 3.12.7 | `github.com/rabbitmq/amqp091-go` | AMQP 0.9.1, exchanges, queues |
| **EMQX** | 5.8 CE | `github.com/eclipse/paho.mqtt.golang` | MQTT 3.1.1/5.0, QoS 0/1/2 |
| **Nacos** | 2.4.3 | `github.com/nacos-group/nacos-sdk-go/v2` | Service discovery, config management |

---

## Documentation

- [Architecture Design](docs/phase-0/architecture.md)
- [Interface Specification](docs/phase-0/interface-spec.md)
- [Metrics Definition](docs/phase-0/metrics-definition.md)
- [Evaluation Criteria](docs/phase-0/evaluation-criteria.md)
- [Testing Strategy](docs/phase-0/testing-strategy.md)
- [Development Guide](docs/phase-0/development-guide.md)
- [Project Plan](PLAN.md)

---

## Development Roadmap

- [x] Phase 0: Project initialization and architecture design
- [x] Phase 1: Redis client implementation
- [x] Phase 2: Kafka client implementation
- [x] Phase 3: Stability detector implementation
- [x] Phase 3.5: Stability scoring system
- [ ] Phase 4: CLI tool implementation
- [x] Phase 6: MongoDB client support
- [x] Phase 7: RocketMQ client support
- [x] Phase 8: RabbitMQ client support
- [x] Phase 9: EMQX client support
- [x] Phase 10: Nacos client support

See [PLAN.md](PLAN.md) for detailed roadmap.

---

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork this repository
2. Create a feature branch (`git checkout -b feature/your-feature`)
3. Commit your changes (`git commit -m 'Add some feature'`)
4. Push to the branch (`git push origin feature/your-feature`)
5. Create a Pull Request

### Code Standards

- Follow Go coding conventions
- All code must pass `golangci-lint run`
- Unit test coverage >= 85%
- Add necessary documentation and comments
- Follow TDD (Test-Driven Development) methodology

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- [go-redis](https://github.com/redis/go-redis) - Redis client
- [kafka-go](https://github.com/segmentio/kafka-go) - Kafka client
- [mongo-driver](https://github.com/mongodb/mongo-go-driver) - MongoDB client
- [rocketmq-client-go](https://github.com/apache/rocketmq-client-go) - RocketMQ client
- [amqp091-go](https://github.com/rabbitmq/amqp091-go) - RabbitMQ client
- [paho.mqtt.golang](https://github.com/eclipse/paho.mqtt.golang) - MQTT client
- [nacos-sdk-go](https://github.com/nacos-group/nacos-sdk-go) - Nacos client
- [cobra](https://github.com/spf13/cobra) - CLI framework

---

## Contact

- **Issues**: [GitHub Issues](https://github.com/username/middleware-chaos-testing/issues)
- **Email**: your.email@example.com

---

# 中文文档

[English](#middleware-chaos-testing-mct) | **中文**

一个全面的中间件混沌测试框架，用于验证 Redis、Kafka、MongoDB、RocketMQ、RabbitMQ、EMQX 和 Nacos 的稳定性。

## 概述

MCT 是一个可扩展的中间件混沌测试框架，通过模拟各类用户操作，检测并量化中间件服务在混沌场景下的稳定性指标。

### 核心特性

- ✅ **多中间件支持**：Redis、Kafka、MongoDB、RocketMQ、RabbitMQ、EMQX、Nacos
- ✅ **可配置测试时长**：支持命令行参数和配置文件
- ✅ **智能评分系统**：0-100分，5个等级
- ✅ **清晰的测试结果**：通过/警告/失败状态及可操作的建议
- ✅ **优先级建议**：按优先级排序的改进建议
- ✅ **生产级指标**：行业标准的稳定性指标
- ✅ **容器化部署**：一键 Docker 启动
- ✅ **TDD 方法论**：测试驱动开发工作流

## 快速开始

### 环境要求

- **Go**: 1.23 或更高版本
- **Docker**: 20.10+（用于运行中间件容器）
- **Docker Compose**: 2.0+（可选，用于多服务部署）

### 安装

```bash
# 克隆仓库
git clone https://github.com/username/middleware-chaos-testing.git
cd middleware-chaos-testing

# 下载依赖
go mod download

# 构建项目
go build -o bin/mct ./cmd/mct
```

## 编译构建

### 从源码构建

```bash
# 构建当前平台
go build -o bin/mct ./cmd/mct

# 优化构建（更小的二进制文件）
go build -ldflags="-s -w" -o bin/mct ./cmd/mct

# 交叉编译到 Linux
GOOS=linux GOARCH=amd64 go build -o bin/mct-linux ./cmd/mct

# 交叉编译到 Windows
GOOS=windows GOARCH=amd64 go build -o bin/mct.exe ./cmd/mct
```

## 运行测试

详细的测试指南请参考 [英文文档](#running-tests)，包括：

- **Redis 测试** - 支持 6.x/7.x 版本
- **Kafka 测试** - 支持 2.7.2 版本
- **MongoDB 测试** - 支持 4.4.13 版本
- **RocketMQ 测试** - 支持 5.1.3 版本（Remoting/gRPC 协议）
- **RabbitMQ 测试** - 支持 3.12.7 版本
- **EMQX 测试** - 支持 5.8 社区版（MQTT）
- **Nacos 测试** - 支持 2.4.3 版本（服务发现与配置管理）

## 运行测试用例

```bash
# 运行所有单元测试
go test ./tests/unit/... -v

# 运行特定中间件测试
go test ./tests/unit/middleware/redis_client_test.go -v
go test ./tests/unit/middleware/kafka_client_test.go -v
go test ./tests/unit/middleware/mongodb_client_test.go -v
go test ./tests/unit/middleware/rocketmq_client_test.go -v
go test ./tests/unit/middleware/rabbitmq_client_test.go -v
go test ./tests/unit/middleware/emqx_client_test.go -v
go test ./tests/unit/middleware/nacos_client_test.go -v

# 生成覆盖率报告
go test ./tests/unit/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 评分体系

MCT 采用 0-100 分的智能评分系统：

| 分数区间 | 等级 | 状态 | 说明 |
|---------|------|------|------|
| 90-100 | EXCELLENT | ✅ PASS | 优秀，可直接用于生产环境 |
| 80-89 | GOOD | ✅ PASS | 良好，满足生产要求 |
| 70-79 | FAIR | ⚠️ WARNING | 一般，建议优化后使用 |
| 60-69 | POOR | ⚠️ WARNING | 较差，需要改进 |
| 0-59 | FAILED | ❌ FAIL | 失败，不建议用于生产 |

### 评分维度

```
总分 = 可用性(30%) + 性能(25%) + 可靠性(25%) + 恢复力(20%)
```

详见 [评分标准文档](docs/phase-0/evaluation-criteria.md)

## 支持的中间件版本

| 中间件 | 版本 | 客户端库 | 关键特性 |
|--------|------|----------|----------|
| **Redis** | 6.x, 7.x | `github.com/redis/go-redis/v9` | KV操作、管道 |
| **Kafka** | 2.7.2 | `github.com/segmentio/kafka-go` | 生产者/消费者、消费组 |
| **MongoDB** | 4.4.13 | `go.mongodb.org/mongo-driver` | CRUD、聚合 |
| **RocketMQ** | 5.1.3 | `github.com/apache/rocketmq-client-go/v2` | **Remoting/gRPC 协议** |
| **RabbitMQ** | 3.12.7 | `github.com/rabbitmq/amqp091-go` | AMQP 0.9.1、交换机、队列 |
| **EMQX** | 5.8 社区版 | `github.com/eclipse/paho.mqtt.golang` | MQTT 3.1.1/5.0、QoS 0/1/2 |
| **Nacos** | 2.4.3 | `github.com/nacos-group/nacos-sdk-go/v2` | 服务发现、配置管理 |

## 开发路线图

- [x] Phase 0: 项目初始化与架构设计
- [x] Phase 1: Redis 客户端实现
- [x] Phase 2: Kafka 客户端实现
- [x] Phase 3: 稳定性检测器实现
- [x] Phase 3.5: 稳定性评分系统
- [ ] Phase 4: CLI 工具实现
- [x] Phase 6: MongoDB 客户端支持
- [x] Phase 7: RocketMQ 客户端支持
- [x] Phase 8: RabbitMQ 客户端支持
- [x] Phase 9: EMQX 客户端支持
- [x] Phase 10: Nacos 客户端支持

详见 [PLAN.md](PLAN.md)

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

---

**Made with ❤️ for middleware reliability testing**
