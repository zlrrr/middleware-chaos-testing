# 开发环境搭建指南

## 1. 环境要求

### 1.1 操作系统

支持以下操作系统：
- **Linux**: Ubuntu 20.04+, CentOS 7+
- **macOS**: 11.0+
- **Windows**: 10+ (使用WSL2)

### 1.2 必需软件

| 软件 | 版本要求 | 用途 |
|------|---------|------|
| Go | >= 1.21 | 主要开发语言 |
| Docker | >= 20.10 | 容器化部署和测试 |
| Docker Compose | >= 2.0 | 多容器编排 |
| Git | >= 2.30 | 版本控制 |
| Make | >= 4.0 | 构建工具 |

### 1.3 可选软件

| 软件 | 用途 |
|------|------|
| golangci-lint | 代码静态检查 |
| mockgen | Mock生成 |
| Redis CLI | Redis调试 |
| Kafka Tools | Kafka调试 |

---

## 2. 安装步骤

### 2.1 安装Go

#### Linux (Ubuntu/Debian)
```bash
# 下载Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz

# 解压到/usr/local
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# 配置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# 验证安装
go version
```

#### macOS
```bash
# 使用Homebrew
brew install go@1.21

# 或者下载安装包
# https://go.dev/dl/go1.21.5.darwin-amd64.pkg

# 验证安装
go version
```

#### Windows (WSL2)
```bash
# 在WSL2中按照Linux方法安装
# 或使用Windows安装包: https://go.dev/dl/go1.21.5.windows-amd64.msi
```

### 2.2 安装Docker

#### Linux (Ubuntu)
```bash
# 卸载旧版本
sudo apt-get remove docker docker-engine docker.io containerd runc

# 安装依赖
sudo apt-get update
sudo apt-get install \
    ca-certificates \
    curl \
    gnupg \
    lsb-release

# 添加Docker GPG密钥
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

# 设置仓库
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# 安装Docker
sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io docker-compose-plugin

# 添加当前用户到docker组
sudo usermod -aG docker $USER
newgrp docker

# 验证安装
docker --version
docker compose version
```

#### macOS
```bash
# 下载并安装Docker Desktop
# https://www.docker.com/products/docker-desktop

# 或使用Homebrew
brew install --cask docker

# 验证安装
docker --version
docker compose version
```

### 2.3 安装开发工具

```bash
# 安装golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 安装mockgen
go install github.com/golang/mock/mockgen@latest

# 安装其他工具
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/go-delve/delve/cmd/dlv@latest
```

---

## 3. 项目设置

### 3.1 克隆项目

```bash
# 克隆仓库
git clone https://github.com/username/middleware-chaos-testing.git
cd middleware-chaos-testing

# 切换到开发分支
git checkout develop
```

### 3.2 初始化Go模块

```bash
# 初始化go.mod
go mod init middleware-chaos-testing

# 下载依赖
go mod download

# 整理依赖
go mod tidy
```

### 3.3 配置Git

```bash
# 配置用户信息
git config user.name "Your Name"
git config user.email "your.email@example.com"

# 安装pre-commit钩子
cp scripts/pre-commit.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

---

## 4. 开发环境配置

### 4.1 VS Code配置

创建 `.vscode/settings.json`:

```json
{
  "go.useLanguageServer": true,
  "go.toolsManagement.autoUpdate": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "go.formatTool": "goimports",
  "editor.formatOnSave": true,
  "go.testFlags": ["-v", "-race"],
  "go.coverOnSave": true,
  "go.coverageOptions": "showCoveredCodeOnly",
  "[go]": {
    "editor.codeActionsOnSave": {
      "source.organizeImports": true
    }
  }
}
```

创建 `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch MCT",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/mct",
      "args": [
        "test",
        "--middleware", "redis",
        "--host", "localhost",
        "--port", "6379",
        "--duration", "10s"
      ]
    },
    {
      "name": "Test Current File",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${file}"
    },
    {
      "name": "Test Current Package",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${fileDirname}"
    }
  ]
}
```

创建 `.vscode/extensions.json`:

```json
{
  "recommendations": [
    "golang.go",
    "eamodio.gitlens",
    "streetsidesoftware.code-spell-checker",
    "ms-azuretools.vscode-docker"
  ]
}
```

### 4.2 GoLand/IntelliJ IDEA配置

1. 打开项目
2. File → Settings → Go → GOROOT: 选择Go安装路径
3. File → Settings → Go → GOPATH: 设置为项目路径
4. File → Settings → Tools → File Watchers: 添加gofmt和goimports
5. Run → Edit Configurations: 添加调试配置

---

## 5. 启动测试环境

### 5.1 使用Docker Compose

创建 `docker-compose.yml`:

```yaml
version: '3.8'

services:
  redis:
    image: redis:7-alpine
    container_name: mct-redis
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes
    volumes:
      - redis-data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  kafka:
    image: confluentinc/cp-kafka:latest
    container_name: mct-kafka
    ports:
      - "9092:9092"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_AUTO_CREATE_TOPICS_ENABLE: "true"
    depends_on:
      - zookeeper

  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    container_name: mct-zookeeper
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    ports:
      - "2181:2181"

volumes:
  redis-data:
```

### 5.2 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 停止并删除数据
docker-compose down -v
```

### 5.3 验证服务

```bash
# 验证Redis
redis-cli ping
# 应该返回: PONG

# 验证Kafka
docker-compose exec kafka kafka-topics --list --bootstrap-server localhost:9092
# 如果能列出topic则表示正常
```

---

## 6. 构建和运行

### 6.1 Makefile

创建项目根目录的 `Makefile`:

```makefile
.PHONY: help build test clean run docker

# 默认目标
help:
	@echo "Available commands:"
	@echo "  make build              - Build the binary"
	@echo "  make test               - Run all tests"
	@echo "  make test-unit          - Run unit tests"
	@echo "  make test-integration   - Run integration tests"
	@echo "  make coverage           - Generate coverage report"
	@echo "  make lint               - Run linter"
	@echo "  make fmt                - Format code"
	@echo "  make clean              - Clean build artifacts"
	@echo "  make run                - Run the application"
	@echo "  make docker-build       - Build Docker image"
	@echo "  make docker-run         - Run in Docker"

# 构建二进制文件
build:
	@echo "Building..."
	@go build -o bin/mct cmd/mct/main.go

# 运行所有测试
test: test-unit

# 运行单元测试
test-unit:
	@echo "Running unit tests..."
	@go test -short -v -race -coverprofile=coverage.out ./...

# 运行集成测试
test-integration:
	@echo "Starting test environment..."
	@docker-compose -f docker-compose.test.yml up -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Running integration tests..."
	@go test -v ./tests/integration/... || (docker-compose -f docker-compose.test.yml down && exit 1)
	@echo "Stopping test environment..."
	@docker-compose -f docker-compose.test.yml down

# 生成覆盖率报告
coverage:
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | grep total
	@echo "Coverage report generated: coverage.html"

# 运行代码检查
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

# 格式化代码
fmt:
	@echo "Formatting code..."
	@gofmt -w .
	@goimports -w .

# 清理构建产物
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean

# 运行应用
run: build
	@./bin/mct test --middleware redis --duration 10s

# 构建Docker镜像
docker-build:
	@echo "Building Docker image..."
	@docker build -t mct:latest .

# 在Docker中运行
docker-run:
	@docker-compose up -d

# 开发模式（监听文件变化）
dev:
	@echo "Starting development mode..."
	@docker-compose up -d
	@go run cmd/mct/main.go test --middleware redis --duration 10s
```

### 6.2 构建项目

```bash
# 构建二进制文件
make build

# 查看帮助
make help

# 运行测试
make test

# 运行应用
make run
```

---

## 7. 调试指南

### 7.1 使用VS Code调试

1. 在代码中设置断点
2. 按F5或点击"Run and Debug"
3. 选择"Launch MCT"配置
4. 开始调试

### 7.2 使用Delve调试

```bash
# 安装delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试main程序
dlv debug cmd/mct/main.go -- test --middleware redis --duration 10s

# 调试测试
dlv test ./internal/evaluator -- -test.run TestEvaluate
```

### 7.3 日志调试

```go
import "log/slog"

// 设置日志级别
slog.SetLogLoggerLevel(slog.LevelDebug)

// 在代码中添加日志
slog.Debug("debug message", "key", value)
slog.Info("info message", "key", value)
slog.Warn("warning message", "key", value)
slog.Error("error message", "error", err)
```

---

## 8. 代码质量工具

### 8.1 golangci-lint配置

创建 `.golangci.yml`:

```yaml
linters-settings:
  govet:
    check-shadowing: true
  gocyclo:
    min-complexity: 15
  dupl:
    threshold: 100
  goconst:
    min-len: 3
    min-occurrences: 3

linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - structcheck
    - varcheck
    - ineffassign
    - deadcode
    - typecheck
    - gocyclo
    - dupl
    - goconst
    - misspell
    - unparam

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0

run:
  timeout: 5m
  tests: true
```

### 8.2 运行代码检查

```bash
# 运行所有检查
make lint

# 自动修复一些问题
golangci-lint run --fix

# 只运行特定linter
golangci-lint run --disable-all -E errcheck
```

---

## 9. 依赖管理

### 9.1 添加依赖

```bash
# 添加新依赖
go get github.com/some/package

# 添加特定版本
go get github.com/some/package@v1.2.3

# 更新依赖
go get -u github.com/some/package

# 整理依赖
go mod tidy
```

### 9.2 vendor目录

```bash
# 创建vendor目录
go mod vendor

# 使用vendor构建
go build -mod=vendor
```

---

## 10. 常见问题

### 10.1 Go环境问题

**问题**: `go: command not found`

**解决**:
```bash
# 确保Go已添加到PATH
export PATH=$PATH:/usr/local/go/bin
source ~/.bashrc
```

**问题**: `package xxx is not in GOROOT`

**解决**:
```bash
# 下载依赖
go mod download
go mod tidy
```

### 10.2 Docker问题

**问题**: `Cannot connect to the Docker daemon`

**解决**:
```bash
# 启动Docker服务
sudo systemctl start docker

# 或使用Docker Desktop
```

**问题**: `permission denied while trying to connect to the Docker daemon`

**解决**:
```bash
# 添加用户到docker组
sudo usermod -aG docker $USER
newgrp docker
```

### 10.3 测试问题

**问题**: 集成测试失败 - 连接被拒绝

**解决**:
```bash
# 确保测试环境已启动
docker-compose up -d

# 检查服务状态
docker-compose ps

# 查看日志
docker-compose logs redis
```

---

## 11. 开发工作流

### 11.1 功能开发流程

```bash
# 1. 创建功能分支
git checkout -b feature/your-feature

# 2. 编写测试（TDD）
vim internal/xxx/xxx_test.go

# 3. 运行测试（应该失败 - 红灯）
make test-unit

# 4. 实现功能
vim internal/xxx/xxx.go

# 5. 运行测试（应该通过 - 绿灯）
make test-unit

# 6. 重构代码
# 确保测试仍然通过

# 7. 提交代码
git add .
git commit -m "feat: add your feature"

# 8. 推送到远程
git push origin feature/your-feature

# 9. 创建Pull Request
```

### 11.2 代码审查清单

- [ ] 代码符合Go编码规范
- [ ] 所有测试通过
- [ ] 覆盖率达标（>= 85%）
- [ ] 代码通过lint检查
- [ ] 有适当的错误处理
- [ ] 有必要的注释和文档
- [ ] 没有遗留的调试代码

---

## 12. 性能分析

### 12.1 CPU分析

```bash
# 生成CPU profile
go test -cpuprofile=cpu.prof -bench=.

# 查看profile
go tool pprof cpu.prof
```

### 12.2 内存分析

```bash
# 生成内存profile
go test -memprofile=mem.prof -bench=.

# 查看profile
go tool pprof mem.prof
```

### 12.3 竞态检测

```bash
# 运行竞态检测
go test -race ./...

# 构建时启用竞态检测
go build -race
```

---

## 13. 文档生成

### 13.1 生成API文档

```bash
# 启动文档服务器
godoc -http=:6060

# 在浏览器中访问
# http://localhost:6060/pkg/middleware-chaos-testing/
```

### 13.2 生成README

使用工具自动生成项目README：

```bash
# 安装readme-gen
go install github.com/ktr0731/readme-gen@latest

# 生成README
readme-gen
```

---

## 14. 快速开始示例

### 14.1 完整的开发环境设置

```bash
#!/bin/bash
# setup-dev.sh - 一键设置开发环境

set -e

echo "Setting up development environment..."

# 1. 检查Go安装
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Please install Go 1.21+"
    exit 1
fi

# 2. 检查Docker安装
if ! command -v docker &> /dev/null; then
    echo "Docker is not installed. Please install Docker"
    exit 1
fi

# 3. 安装开发工具
echo "Installing development tools..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/golang/mock/mockgen@latest
go install golang.org/x/tools/cmd/goimports@latest

# 4. 下载依赖
echo "Downloading dependencies..."
go mod download

# 5. 启动测试环境
echo "Starting test environment..."
docker-compose up -d

# 6. 等待服务就绪
echo "Waiting for services..."
sleep 10

# 7. 验证环境
echo "Verifying environment..."
make test-unit

echo "Development environment is ready!"
echo "Run 'make help' to see available commands"
```

使用方法：
```bash
chmod +x setup-dev.sh
./setup-dev.sh
```

---

**文档版本**: v1.0
**最后更新**: 2025-10-30
