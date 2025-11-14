# 中间件混沌测试框架 (MCT)

一个全面的中间件混沌测试框架，用于验证 Redis、Kafka、MongoDB、RocketMQ、RabbitMQ、EMQX 和 Nacos 的稳定性。

[![Go 版本](https://img.shields.io/badge/go-1.23+-blue.svg)](https://golang.org/dl/)
[![许可证](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

**[English Documentation](README.md)**

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

## 目录

- [快速开始](#快速开始)
- [编译构建](#编译构建)
- [运行测试](#运行测试)
  - [Redis 测试](#redis-测试)
  - [Kafka 测试](#kafka-测试)
  - [MongoDB 测试](#mongodb-测试)
  - [RocketMQ 测试](#rocketmq-测试)
  - [RabbitMQ 测试](#rabbitmq-测试)
  - [EMQX 测试](#emqx-测试)
  - [Nacos 测试](#nacos-测试)
- [运行测试用例](#运行测试用例)
- [配置说明](#配置说明)
- [评分体系](#评分体系)
- [开发指南](#开发指南)

---

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

---

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

### 使用 Makefile 构建（如果可用）

```bash
# 构建二进制文件
make build

# 构建并安装到 $GOPATH/bin
make install

# 清理构建产物
make clean
```

### 验证构建

```bash
# 检查二进制文件
./bin/mct version

# 显示帮助信息
./bin/mct --help
```

---

## 运行测试

### Redis 测试

**支持版本**: Redis 6.x, 7.x

#### 启动 Redis 服务器

```bash
# 使用 Docker
docker run -d --name redis-test \
  -p 6379:6379 \
  redis:7-alpine

# 或使用 Docker Compose
docker-compose up -d redis
```

#### 运行稳定性测试

```bash
# 基础测试（30秒，5000次操作）
./bin/mct test \
  --middleware redis \
  --host localhost \
  --port 6379 \
  --duration 30s \
  --operations 5000

# 高级测试，自定义配置
./bin/mct test \
  --middleware redis \
  --host localhost \
  --port 6379 \
  --duration 60s \
  --operations 10000 \
  --concurrency 10 \
  --password "your-redis-password"
```

#### Redis 测试操作

- `SET`: 写入键值对
- `GET`: 根据键读取值
- `DEL`: 删除键
- `INCR`: 递增计数器
- `EXPIRE`: 设置键过期时间

---

### Kafka 测试

**支持版本**: Kafka 2.7.2

#### 启动 Kafka 集群

```bash
# 使用 Docker Compose（带 ZooKeeper）
docker-compose up -d zookeeper kafka

# 或独立 Kafka（KRaft 模式）
docker run -d --name kafka-test \
  -p 9092:9092 \
  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  apache/kafka:latest
```

#### 运行稳定性测试

```bash
# 基础测试
./bin/mct test \
  --middleware kafka \
  --brokers localhost:9092 \
  --topic test-topic \
  --duration 30s \
  --operations 5000

# 高级测试，带消费者组
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

#### Kafka 测试操作

- `PRODUCE`: 发布消息到主题
- `CONSUME`: 订阅并消费消息
- `COMMIT`: 提交消费者偏移量

---

### MongoDB 测试

**支持版本**: MongoDB 4.4.13

#### 启动 MongoDB 服务器

```bash
# 使用 Docker
docker run -d --name mongodb-test \
  -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=password \
  mongo:4.4.13

# 或使用 Docker Compose
docker-compose up -d mongodb
```

#### 运行稳定性测试

```bash
# 基础测试
./bin/mct test \
  --middleware mongodb \
  --host localhost \
  --port 27017 \
  --database testdb \
  --collection testcol \
  --duration 30s \
  --operations 5000

# 带认证
./bin/mct test \
  --middleware mongodb \
  --uri "mongodb://admin:password@localhost:27017/testdb?authSource=admin" \
  --collection testcol \
  --duration 60s \
  --operations 10000 \
  --concurrency 10
```

#### MongoDB 测试操作

- `INSERT`: 插入文档
- `FIND`: 查询文档
- `UPDATE`: 更新文档
- `DELETE`: 删除文档
- `AGGREGATE`: 运行聚合管道

---

### RocketMQ 测试

**支持版本**: RocketMQ 5.1.3
**协议**: Remoting（默认）、gRPC

#### 启动 RocketMQ 服务器

```bash
# 使用 Docker
docker run -d --name rocketmq-namesrv \
  -p 9876:9876 \
  apache/rocketmq:5.1.3 sh mqnamesrv

docker run -d --name rocketmq-broker \
  -p 10909:10909 -p 10911:10911 \
  -e NAMESRV_ADDR=localhost:9876 \
  apache/rocketmq:5.1.3 sh mqbroker
```

#### 运行稳定性测试

```bash
# 基础测试（Remoting 协议 - 默认）
./bin/mct test \
  --middleware rocketmq \
  --nameservers localhost:9876 \
  --topic test-topic \
  --duration 30s \
  --operations 5000

# 使用 gRPC 协议（RocketMQ 5.x 推荐）
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

#### RocketMQ 测试操作

- `SEND`: 发送消息到主题
- `RECEIVE`: 从队列消费消息
- 协议选择: `remoting`（兼容性好）或 `grpc`（高性能）

---

### RabbitMQ 测试

**支持版本**: RabbitMQ 3.12.7

#### 启动 RabbitMQ 服务器

```bash
# 使用 Docker，带管理插件
docker run -d --name rabbitmq-test \
  -p 5672:5672 \
  -p 15672:15672 \
  -e RABBITMQ_DEFAULT_USER=admin \
  -e RABBITMQ_DEFAULT_PASS=password \
  rabbitmq:3.12.7-management

# 访问管理界面: http://localhost:15672
```

#### 运行稳定性测试

```bash
# 基础测试
./bin/mct test \
  --middleware rabbitmq \
  --host localhost \
  --port 5672 \
  --user admin \
  --password password \
  --duration 30s \
  --operations 5000

# 高级测试，带路由
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

#### RabbitMQ 测试操作

- `PUBLISH`: 发布消息到交换机
- `CONSUME`: 从队列消费消息
- `ACK`: 确认消息投递
- 支持: Direct、Topic、Fanout 交换机

---

### EMQX 测试

**支持版本**: EMQX 5.8 社区版
**协议**: MQTT 3.1.1, MQTT 5.0

#### 启动 EMQX 服务器

```bash
# 使用 Docker
docker run -d --name emqx-test \
  -p 1883:1883 \
  -p 18083:18083 \
  emqx/emqx:5.8

# 访问仪表板: http://localhost:18083
# 默认凭据: admin / public
```

#### 运行稳定性测试

```bash
# 基础 MQTT 测试
./bin/mct test \
  --middleware emqx \
  --broker tcp://localhost:1883 \
  --client-id test-client \
  --topic test/topic \
  --duration 30s \
  --operations 5000

# MQTT 带 QoS 级别
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

#### EMQX 测试操作

- `PUBLISH`: 发布 MQTT 消息
- `SUBSCRIBE`: 订阅主题
- QoS 级别: 0（最多一次）、1（至少一次）、2（恰好一次）
- 特性: 保留消息、遗嘱消息、清理会话
- 共享订阅（EMQX 5.8 特性）

---

### Nacos 测试

**支持版本**: Nacos 2.4.3
**功能**: 服务发现、配置管理

#### 启动 Nacos 服务器

```bash
# 使用 Docker（独立模式）
docker run -d --name nacos-test \
  -p 8848:8848 \
  -p 9848:9848 \
  -e MODE=standalone \
  nacos/nacos-server:v2.4.3

# 访问控制台: http://localhost:8848/nacos
# 默认凭据: nacos / nacos
```

#### 运行稳定性测试

```bash
# 服务注册测试
./bin/mct test \
  --middleware nacos \
  --servers localhost:8848 \
  --service-name test-service \
  --namespace public \
  --duration 30s \
  --operations 5000

# 配置管理测试
./bin/mct test \
  --middleware nacos \
  --servers localhost:8848 \
  --data-id test-config \
  --group DEFAULT_GROUP \
  --namespace public \
  --duration 60s \
  --operations 10000

# 带认证（Nacos 2.x）
./bin/mct test \
  --middleware nacos \
  --servers localhost:8848 \
  --service-name test-service \
  --username nacos \
  --password nacos \
  --namespace dev \
  --duration 60s
```

#### Nacos 测试操作

- **服务注册中心**:
  - `REGISTER`: 注册服务实例
  - `DEREGISTER`: 注销实例
  - `DISCOVER`: 查询服务实例
  - `SUBSCRIBE`: 监听服务变化
  - `HEARTBEAT`: 发送心跳（临时实例）
- **配置中心**:
  - `PUBLISH_CONFIG`: 发布配置
  - `GET_CONFIG`: 获取配置
  - `REMOVE_CONFIG`: 删除配置
  - `LISTEN_CONFIG`: 监听配置变化

---

## 运行测试用例

### 单元测试

运行所有中间件客户端的单元测试：

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

# 运行并生成覆盖率
go test ./tests/unit/... -cover -coverprofile=coverage.out

# 查看覆盖率报告
go tool cover -html=coverage.out
```

### 集成测试

```bash
# 运行集成测试（需要运行中间件服务）
go test ./tests/integration/... -v

# 运行特定集成测试
go test ./tests/integration/redis_integration_test.go -v
```

### 端到端测试

```bash
# 运行 E2E 测试
go test ./tests/e2e/... -v

# 带超时运行
go test ./tests/e2e/... -v -timeout 10m
```

### Makefile 测试目标（如果可用）

```bash
# 运行单元测试
make test-unit

# 运行集成测试
make test-integration

# 运行所有测试
make test

# 生成覆盖率报告
make coverage

# 运行竞态检测测试
make test-race
```

---

## 配置说明

### YAML 配置文件

创建配置文件以实现可重复测试：

```yaml
# configs/test-redis.yaml
name: "Redis 稳定性测试"
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

### 使用配置文件运行

```bash
./bin/mct test --config configs/test-redis.yaml
```

---

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

**可用性（30分）**：
- 正常运行时间百分比
- 操作成功率

**性能（25分）**：
- P95 延迟（15分）
- P99 延迟（10分）

**可靠性（25分）**：
- 错误率（15分）
- 数据丢失率（10分）

**恢复力（20分）**：
- 平均恢复时间 MTTR（12分）
- 重连成功率（8分）

详见 [评分标准文档](docs/phase-0/evaluation-criteria.md)

---

## 开发指南

### 开发环境设置

```bash
# 安装开发工具
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 下载依赖
go mod download

# 启动开发环境（所有中间件）
docker-compose up -d
```

### 代码质量

```bash
# 格式化代码
go fmt ./...
goimports -w .

# 代码检查
golangci-lint run

# 运行所有检查
go fmt ./... && golangci-lint run && go test ./tests/unit/... -cover
```

### 添加新的中间件客户端

1. 在 `internal/middleware/` 中创建客户端接口实现
2. 在 `internal/middleware/{name}_types.go` 中创建类型定义
3. 在 `internal/middleware/{name}_client.go` 中创建客户端实现
4. 在 `internal/evaluator/stability_evaluator.go` 中添加阈值函数
5. 在 `tests/unit/middleware/{name}_client_test.go` 中编写 15+ 个单元测试
6. 更新文档

### 项目结构

```
middleware-chaos-testing/
├── cmd/
│   └── mct/                  # 主程序入口
├── internal/
│   ├── core/                 # 核心抽象
│   ├── middleware/           # 中间件适配器（7个客户端）
│   │   ├── redis_client.go
│   │   ├── kafka_client.go
│   │   ├── mongodb_client.go
│   │   ├── rocketmq_client.go
│   │   ├── rabbitmq_client.go
│   │   ├── emqx_client.go
│   │   └── nacos_client.go
│   ├── metrics/              # 指标收集器
│   ├── detector/             # 稳定性检测器
│   ├── evaluator/            # 稳定性评估器
│   ├── reporter/             # 报告生成器
│   ├── orchestrator/         # 测试编排器
│   └── config/               # 配置管理器
├── tests/
│   ├── unit/                 # 单元测试（每个中间件15+个）
│   ├── integration/          # 集成测试
│   └── e2e/                  # 端到端测试
├── docs/                     # 文档
├── configs/                  # 配置示例
├── scripts/                  # 构建和部署脚本
├── Makefile                  # 构建自动化
├── Dockerfile                # 容器镜像
└── docker-compose.yml        # 多服务编排
```

---

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

---

## 文档

- [架构设计](docs/phase-0/architecture.md)
- [接口规范](docs/phase-0/interface-spec.md)
- [指标定义](docs/phase-0/metrics-definition.md)
- [评分标准](docs/phase-0/evaluation-criteria.md)
- [测试策略](docs/phase-0/testing-strategy.md)
- [开发指南](docs/phase-0/development-guide.md)
- [项目计划](PLAN.md)

---

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

---

## 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/your-feature`)
3. 提交更改 (`git commit -m 'Add some feature'`)
4. 推送到分支 (`git push origin feature/your-feature`)
5. 创建 Pull Request

### 代码规范

- 遵循 Go 编码规范
- 所有代码必须通过 `golangci-lint run`
- 单元测试覆盖率 >= 85%
- 添加必要的文档和注释
- 遵循 TDD（测试驱动开发）方法论

---

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

---

## 致谢

- [go-redis](https://github.com/redis/go-redis) - Redis 客户端
- [kafka-go](https://github.com/segmentio/kafka-go) - Kafka 客户端
- [mongo-driver](https://github.com/mongodb/mongo-go-driver) - MongoDB 客户端
- [rocketmq-client-go](https://github.com/apache/rocketmq-client-go) - RocketMQ 客户端
- [amqp091-go](https://github.com/rabbitmq/amqp091-go) - RabbitMQ 客户端
- [paho.mqtt.golang](https://github.com/eclipse/paho.mqtt.golang) - MQTT 客户端
- [nacos-sdk-go](https://github.com/nacos-group/nacos-sdk-go) - Nacos 客户端
- [cobra](https://github.com/spf13/cobra) - CLI 框架

---

## 联系方式

- **问题反馈**: [GitHub Issues](https://github.com/username/middleware-chaos-testing/issues)
- **邮箱**: your.email@example.com

---

**用 ❤️ 为中间件可靠性测试而构建**
