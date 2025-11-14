# Middleware Chaos Testing (MCT)

A comprehensive chaos testing framework for middleware stability validation across Redis, Kafka, MongoDB, RocketMQ, RabbitMQ, EMQX, and Nacos.

[![Go Version](https://img.shields.io/badge/go-1.23+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Build Binaries](https://github.com/zlrrr/middleware-chaos-testing/actions/workflows/release-binaries.yml/badge.svg)](https://github.com/zlrrr/middleware-chaos-testing/actions/workflows/release-binaries.yml)
[![Build Docker Images](https://github.com/zlrrr/middleware-chaos-testing/actions/workflows/docker-images.yml/badge.svg)](https://github.com/zlrrr/middleware-chaos-testing/actions/workflows/docker-images.yml)

**[中文文档](README_CN.md)** | **[Installation Guide](INSTALL.md)**

## Overview

MCT is an extensible middleware chaos testing framework that simulates various user operations to detect and quantify middleware service stability metrics under chaotic scenarios.

### Core Features

- ✅ **Multi-Middleware Support**: Redis, Kafka, MongoDB, RocketMQ, RabbitMQ, EMQX, Nacos
- ✅ **Web Platform**: Modern React UI + RESTful API server for easy management
- ✅ **Configurable Test Duration**: Command-line and configuration file support
- ✅ **Intelligent Scoring System**: 0-100 points with 5 grade levels
- ✅ **Clear Test Results**: Pass/Warning/Fail status with actionable insights
- ✅ **Prioritized Recommendations**: Improvement suggestions ranked by priority
- ✅ **Production-Grade Metrics**: Industry-standard stability indicators
- ✅ **Containerized Deployment**: One-command Docker setup with all services
- ✅ **TDD Methodology**: Test-Driven Development workflow

## Table of Contents

- [Quick Start](#quick-start)
- [Web Platform](#web-platform)
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

### Installation Options

MCT can be installed in multiple ways depending on your needs:

#### Option 1: Pre-built Binary (Recommended)

Download the latest release for your platform:

```bash
# One-line install (Linux/macOS)
curl -fsSL https://raw.githubusercontent.com/zlrrr/middleware-chaos-testing/main/install.sh | bash

# Verify installation
mct version
```

**Manual download:**
- Visit the [Releases page](https://github.com/zlrrr/middleware-chaos-testing/releases)
- Download the appropriate binary for your OS and architecture
- Extract and move to your PATH

#### Option 2: Docker Images (For Web Platform)

Use pre-built Docker images for the complete platform:

```bash
# Pull images
docker pull ghcr.io/zlrrr/middleware-chaos-testing/mct-server:latest
docker pull ghcr.io/zlrrr/middleware-chaos-testing/mct-web:latest

# Download and run with docker-compose
curl -O https://raw.githubusercontent.com/zlrrr/middleware-chaos-testing/main/docker-compose.release.yml
docker-compose -f docker-compose.release.yml up -d
```

Access the platform at:
- **Web UI**: http://localhost:3000
- **API Server**: http://localhost:8080

#### Option 3: Build from Source

Requirements:
- **Go**: 1.23 or higher
- **Node.js**: 18+ (for web UI)
- **Docker**: 20.10+ (for middleware containers)
- **Docker Compose**: 2.0+

```bash
# Clone the repository
git clone https://github.com/zlrrr/middleware-chaos-testing.git
cd middleware-chaos-testing

# Build CLI tool
go build -o bin/mct ./cmd/mct

# Build API server
go build -o bin/mct-server ./cmd/mct-server

# Build web UI
cd web
npm install
npm run build
cd ..

# Or use the startup script to build and run everything
./scripts/start-platform.sh
```

### Prerequisites

**For Docker deployment:**
- Docker 20.10+
- Docker Compose 2.0+
- 8GB RAM minimum (16GB recommended)
- 20GB disk space

**For source build:**
- Go 1.23+
- Node.js 18+
- Git

---

## Web Platform

MCT Platform provides a complete web-based solution for managing chaos testing tasks through a modern UI and RESTful API.

### Start the Platform

```bash
# Start all services (frontend, backend, and 7 middlewares)
./scripts/start-platform.sh

# Or use docker-compose directly
docker-compose up -d
```

### Access the Platform

- **Web UI**: http://localhost:3000
- **API Server**: http://localhost:8080
- **API Documentation**: http://localhost:8080/api/v1/middlewares

### Features

**Web UI (React)**:
- Dashboard with task statistics and middleware coverage
- Task management (create, run, view, delete)
- Interactive test result visualization
- Real-time task status updates
- Five-dimensional score cards
- Detailed metrics and recommendations

**API Server (Go)**:
- RESTful API for task management
- Support for all 7 middleware types
- Task execution and result retrieval
- Health checks and monitoring
- CORS-enabled for frontend integration

**Middleware Services**:
- Redis (6379) - In-memory data store
- Kafka (9092) - Message streaming
- MongoDB (27017) - Document database
- RabbitMQ (5672, 15672) - Message broker
- EMQX (1883, 18083) - MQTT broker
- Nacos (8848, 9848) - Service discovery
- Elasticsearch (9200) - Search engine

### Quick Start Guide

#### Step 1: Start the Platform

```bash
# Clone repository (if not already done)
git clone https://github.com/zlrrr/middleware-chaos-testing.git
cd middleware-chaos-testing

# Start all services
./scripts/start-platform.sh
```

Wait for all services to become healthy. The script will display:
```
✅ All services are healthy!

Access URLs:
- Web UI: http://localhost:3000
- API Server: http://localhost:8080
- API Health: http://localhost:8080/health

Middleware Services:
- Redis: localhost:6379
- Kafka: localhost:9092
- MongoDB: localhost:27017
- RabbitMQ: localhost:5672 (Management: http://localhost:15672)
- EMQX: localhost:1883 (Dashboard: http://localhost:18083)
- Nacos: localhost:8848 (Console: http://localhost:8848/nacos)
```

#### Step 2: Access the Web UI

Open your browser and navigate to **http://localhost:3000**

You'll see the MCT Dashboard with:
- **Task Statistics**: Total, running, completed, failed tasks
- **Middleware Coverage**: Which middleware types have been tested
- **Recent Activity**: Latest test results

#### Step 3: Create a Test Task

**Via Web UI:**

1. Click the **"Create Task"** button in the top right
2. Fill in the task details:

   **For Redis:**
   ```
   Middleware: Redis
   Host: redis
   Port: 6379
   Password: (leave empty)
   DB: 0
   ```

   **For Kafka:**
   ```
   Middleware: Kafka
   Brokers: kafka:9092
   Topic: test-topic
   ```

   **For MongoDB:**
   ```
   Middleware: MongoDB
   URI: mongodb://admin:password@mongodb:27017
   Database: testdb
   Collection: testcol
   ```

3. Click **"Create"** to save the task

**Via API:**

```bash
# Create a Redis test task
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "middleware": "redis",
    "config": {
      "host": "redis",
      "port": 6379,
      "db": 0
    }
  }'
```

#### Step 4: Run the Test

**Via Web UI:**

1. Navigate to the **Tasks** page
2. Find your task in the list
3. Click the **"Run" (▶️)** button
4. Watch the status change from "pending" → "running" → "completed"

**Via API:**

```bash
# Run the task (replace with your task ID)
curl -X POST http://localhost:8080/api/v1/tasks/{task-id}/run

# Check status
curl http://localhost:8080/api/v1/tasks/{task-id}
```

#### Step 5: View Test Results

**Via Web UI:**

1. Click **"View Results" (👁️)** button next to the completed task
2. You'll see:
   - **Overall Score** (0-100) and Grade (A/B/C/D/F)
   - **Five Dimensions**: Availability, Performance, Resilience, Data Integrity, Recovery
   - **Key Metrics**: Operations, latency, error rate
   - **Issues Found**: Categorized by severity (Critical/High/Medium/Low)
   - **Recommendations**: Prioritized improvement suggestions

**Via API:**

```bash
# Get test result
curl http://localhost:8080/api/v1/tasks/{task-id}/result | jq

# Output example:
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "task-123",
    "score": 87.5,
    "grade": "B",
    "dimensions": {
      "availability": 92.0,
      "performance": 85.0,
      "resilience": 88.0,
      "data_integrity": 95.0,
      "recovery": 77.5
    },
    ...
  }
}
```

### Understanding Test Results

**Score Interpretation:**

| Score | Grade | Status | Meaning |
|-------|-------|--------|---------|
| 90-100 | A | ✅ PASS | Excellent - Production ready |
| 80-89 | B | ✅ PASS | Good - Meets production standards |
| 70-79 | C | ⚠️ WARNING | Fair - Optimization recommended |
| 60-69 | D | ⚠️ WARNING | Poor - Improvements needed |
| 0-59 | F | ❌ FAIL | Failed - Not production ready |

**Five Dimensions:**

1. **Availability** (0-100): Uptime and success rate during chaos
2. **Performance** (0-100): Response time and throughput
3. **Resilience** (0-100): Ability to handle failures
4. **Data Integrity** (0-100): Data consistency and correctness
5. **Recovery** (0-100): Time to recover from failures

### Common Use Cases

#### Use Case 1: Pre-Production Validation

Test middleware before deploying to production:

```bash
# Test Redis cluster
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "middleware": "redis",
    "config": {"host": "redis", "port": 6379}
  }'

# If score >= 80: Safe to deploy ✅
# If score < 80: Review recommendations ⚠️
```

#### Use Case 2: Continuous Monitoring

Run periodic tests to track stability trends:

```bash
# Automated daily test
#!/bin/bash
for middleware in redis kafka mongodb; do
  TASK_ID=$(curl -s -X POST http://localhost:8080/api/v1/tasks \
    -H "Content-Type: application/json" \
    -d "{\"middleware\":\"$middleware\",\"config\":{}}" \
    | jq -r '.data.id')

  curl -X POST http://localhost:8080/api/v1/tasks/$TASK_ID/run
  echo "Started test for $middleware: $TASK_ID"
done
```

#### Use Case 3: Comparing Configurations

Test different configurations to find the optimal setup:

1. Create tasks with different configurations
2. Run all tasks
3. Compare scores to determine best configuration

### Troubleshooting

**Problem: Services not starting**

```bash
# Check service status
docker-compose ps

# Check logs
docker-compose logs mct-server
docker-compose logs mct-web

# Restart specific service
docker-compose restart mct-server
```

**Problem: Cannot connect to middleware**

Make sure to use service names (not localhost) when creating tasks:
- ✅ Correct: `redis`, `kafka`, `mongodb`
- ❌ Incorrect: `localhost`, `127.0.0.1`

**Problem: Task stuck in "running" status**

```bash
# Check mct-server logs for errors
docker-compose logs -f mct-server

# Check if middleware service is healthy
docker-compose ps | grep redis
```

**Problem: Web UI not loading**

```bash
# Check if mct-web is running
docker-compose ps mct-web

# Restart web service
docker-compose restart mct-web

# Clear browser cache and reload
```

### Advanced Configuration

#### Custom Network Settings

Edit `docker-compose.yml` to customize ports:

```yaml
services:
  mct-web:
    ports:
      - "8080:80"  # Change web UI port to 8080
  mct-server:
    ports:
      - "9090:8080"  # Change API port to 9090
```

#### Persistent Storage

Test results are stored in memory by default. For persistence:

```yaml
services:
  mct-server:
    volumes:
      - ./data:/data
    environment:
      - STORAGE_TYPE=file
      - STORAGE_PATH=/data
```

#### Resource Limits

Adjust resource limits for middleware services:

```yaml
services:
  redis:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
```

### Documentation

- **Platform Guide**: [docs/platform-guide.md](docs/platform-guide.md)
- **API Reference**: [docs/API.md](docs/API.md)
- **Deployment Guide**: [DOCKER_DEPLOYMENT.md](DOCKER_DEPLOYMENT.md)

### Stop the Platform

```bash
./scripts/stop-platform.sh

# Or use docker-compose directly
docker-compose down
```

---

## CLI Tool Usage

The `mct` command-line tool provides direct access to middleware chaos testing without the web platform.

### Basic Usage

```bash
# Test Redis
./bin/mct test --middleware redis --host localhost --duration 30s --operations 5000

# Test Kafka
./bin/mct test --middleware kafka --host localhost --duration 60s --operations 10000

# Test MongoDB
./bin/mct test --middleware mongodb --host localhost --duration 30s

# Test all 7 middleware types
./bin/mct test --middleware redis|kafka|mongodb|rocketmq|rabbitmq|emqx|nacos
```

### CLI Options

```
Flags:
  --middleware string    Middleware type (redis|kafka|mongodb|rocketmq|rabbitmq|emqx|nacos) [required]
  --host string          Middleware host (default "localhost")
  --port int             Middleware port (auto-detected by middleware type)
  --duration duration    Test duration (default 60s)
  --operations int       Number of operations (default 10000)
  --output string        Output format: console|json|markdown (default "console")
  --report-path string   Report output path (default: stdout)
  --config string        Config file path (YAML)
```

### Output Formats

```bash
# Console output (human-readable)
./bin/mct test --middleware redis --output console

# JSON output (machine-readable)
./bin/mct test --middleware kafka --output json

# Markdown report
./bin/mct test --middleware mongodb --output markdown --report-path ./report.md
```

### Using Configuration Files

```bash
# Use a YAML configuration file
./bin/mct test --config configs/test-redis.yaml
```

Example configuration file (`configs/test-redis.yaml`):
```yaml
name: "Redis Stability Test"
middleware: "redis"
connection:
  host: "localhost"
  port: 6379
  timeout: 5s
test:
  duration: 60s
  operations: 10000
output:
  format: "console"
  include_recommendations: true
```

### Exit Codes

- `0`: Test passed (score ≥ 80)
- `1`: Test failed (score < 60)
- `2`: Test warning (60 ≤ score < 80)

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
│   ├── mct/                  # Main CLI program
│   └── mct-server/           # Web API server
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
│   ├── config/               # Configuration manager
│   ├── types/                # Shared type definitions
│   ├── storage/              # Task and result storage
│   ├── executor/             # Task execution engine
│   ├── service/              # Business logic layer
│   └── api/                  # HTTP API handlers and middleware
│       ├── handlers/         # API request handlers
│       └── middleware/       # HTTP middleware (CORS, logging, recovery)
├── web/                      # React frontend
│   ├── src/
│   │   ├── components/       # React components
│   │   ├── pages/            # Page components
│   │   ├── services/         # API client
│   │   ├── stores/           # State management (Zustand)
│   │   └── types/            # TypeScript definitions
│   ├── Dockerfile            # Frontend container image
│   └── nginx.conf            # Nginx configuration
├── tests/
│   ├── unit/                 # Unit tests (15+ per middleware)
│   ├── integration/          # Integration tests (API)
│   └── e2e/                  # End-to-end platform tests
├── docs/                     # Documentation
│   ├── platform-guide.md     # Platform user guide
│   └── API.md                # API reference
├── configs/                  # Configuration examples
├── scripts/                  # Build & deployment scripts
│   ├── start-platform.sh     # Platform startup script
│   └── stop-platform.sh      # Platform shutdown script
├── Makefile                  # Build automation
├── Dockerfile.server         # Backend container image
├── docker-compose.yml        # Multi-service orchestration
└── DOCKER_DEPLOYMENT.md      # Docker deployment guide
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

**Platform Documentation**:
- [Platform User Guide](docs/platform-guide.md)
- [API Reference](docs/API.md)
- [Docker Deployment Guide](DOCKER_DEPLOYMENT.md)

**Architecture Documentation**:
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
- [x] Phase 4: CLI tool implementation (支持全部7种中间件)
- [x] Phase 6: MongoDB client support
- [x] Phase 7: RocketMQ client support
- [x] Phase 8: RabbitMQ client support
- [x] Phase 9: EMQX client support
- [x] Phase 10: Nacos client support
- [x] Phase 11: Web Service Platform
  - [x] Phase 11.1: Backend API server
  - [x] Phase 11.2: React frontend UI
  - [x] Phase 11.3: Docker containerization
  - [x] Phase 11.4: Integration testing and documentation

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

**Made with ❤️ for middleware reliability testing**
