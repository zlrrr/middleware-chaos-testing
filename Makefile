.PHONY: help build test test-unit test-integration test-e2e coverage lint fmt clean run docker-build docker-run dev

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
BINARY_NAME=mct
BINARY_PATH=bin/$(BINARY_NAME)
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# 颜色定义
BLUE=\033[0;34m
GREEN=\033[0;32m
RED=\033[0;31m
NC=\033[0m # No Color

help: ## 显示帮助信息
	@echo "$(BLUE)可用命令:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-18s$(NC) %s\n", $$1, $$2}'

build: ## 构建二进制文件
	@echo "$(BLUE)构建中...$(NC)"
	@go build -o $(BINARY_PATH) cmd/mct/main.go
	@echo "$(GREEN)构建完成: $(BINARY_PATH)$(NC)"

install: build ## 安装到 GOPATH/bin
	@echo "$(BLUE)安装中...$(NC)"
	@go install cmd/mct/main.go
	@echo "$(GREEN)安装完成$(NC)"

test: test-unit ## 运行所有测试

test-unit: ## 运行单元测试
	@echo "$(BLUE)运行单元测试...$(NC)"
	@go test -short -v -race ./...

test-unit-quiet: ## 安静模式运行单元测试
	@go test -short ./...

test-integration: ## 运行集成测试
	@echo "$(BLUE)启动测试环境...$(NC)"
	@docker-compose -f docker-compose.test.yml up -d
	@echo "$(BLUE)等待服务就绪...$(NC)"
	@sleep 5
	@echo "$(BLUE)运行集成测试...$(NC)"
	@go test -v ./tests/integration/... || (docker-compose -f docker-compose.test.yml down && exit 1)
	@echo "$(BLUE)停止测试环境...$(NC)"
	@docker-compose -f docker-compose.test.yml down

test-e2e: build ## 运行端到端测试
	@echo "$(BLUE)启动测试环境...$(NC)"
	@docker-compose -f docker-compose.test.yml up -d
	@echo "$(BLUE)等待服务就绪...$(NC)"
	@sleep 5
	@echo "$(BLUE)运行E2E测试...$(NC)"
	@go test -v ./tests/e2e/... || (docker-compose -f docker-compose.test.yml down && exit 1)
	@echo "$(BLUE)停止测试环境...$(NC)"
	@docker-compose -f docker-compose.test.yml down

coverage: ## 生成覆盖率报告
	@echo "$(BLUE)生成覆盖率报告...$(NC)"
	@go test -coverprofile=$(COVERAGE_FILE) ./...
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "$(GREEN)覆盖率报告已生成: $(COVERAGE_HTML)$(NC)"
	@go tool cover -func=$(COVERAGE_FILE) | grep total

coverage-check: ## 检查覆盖率是否达标 (>= 85%)
	@echo "$(BLUE)检查覆盖率...$(NC)"
	@go test -coverprofile=$(COVERAGE_FILE) ./... > /dev/null 2>&1
	@coverage=$$(go tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ $$(echo "$$coverage < 85" | bc -l) -eq 1 ]; then \
		echo "$(RED)覆盖率 $$coverage% 低于阈值 85%$(NC)"; \
		exit 1; \
	else \
		echo "$(GREEN)覆盖率 $$coverage% 达标 (>= 85%)$(NC)"; \
	fi

lint: ## 运行代码检查
	@echo "$(BLUE)运行代码检查...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "$(RED)golangci-lint 未安装，跳过检查$(NC)"; \
	fi

fmt: ## 格式化代码
	@echo "$(BLUE)格式化代码...$(NC)"
	@gofmt -w .
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi
	@echo "$(GREEN)代码格式化完成$(NC)"

vet: ## 运行 go vet
	@echo "$(BLUE)运行 go vet...$(NC)"
	@go vet ./...

clean: ## 清理构建产物
	@echo "$(BLUE)清理中...$(NC)"
	@rm -rf $(BINARY_PATH)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@go clean
	@echo "$(GREEN)清理完成$(NC)"

run: build ## 运行应用程序
	@echo "$(BLUE)启动测试环境...$(NC)"
	@docker-compose up -d
	@sleep 3
	@echo "$(BLUE)运行应用...$(NC)"
	@./$(BINARY_PATH) test --middleware redis --duration 10s

run-kafka: build ## 运行Kafka测试
	@echo "$(BLUE)启动测试环境...$(NC)"
	@docker-compose up -d
	@sleep 10
	@echo "$(BLUE)运行Kafka测试...$(NC)"
	@./$(BINARY_PATH) test --middleware kafka --duration 10s

docker-build: ## 构建Docker镜像
	@echo "$(BLUE)构建Docker镜像...$(NC)"
	@docker build -t mct:latest .
	@echo "$(GREEN)Docker镜像构建完成$(NC)"

docker-run: ## 在Docker中运行
	@echo "$(BLUE)启动Docker环境...$(NC)"
	@docker-compose up -d
	@echo "$(GREEN)服务已启动$(NC)"

docker-stop: ## 停止Docker环境
	@echo "$(BLUE)停止Docker环境...$(NC)"
	@docker-compose down
	@echo "$(GREEN)服务已停止$(NC)"

docker-logs: ## 查看Docker日志
	@docker-compose logs -f

mod-download: ## 下载依赖
	@echo "$(BLUE)下载依赖...$(NC)"
	@go mod download

mod-tidy: ## 整理依赖
	@echo "$(BLUE)整理依赖...$(NC)"
	@go mod tidy

mod-vendor: ## 创建vendor目录
	@echo "$(BLUE)创建vendor目录...$(NC)"
	@go mod vendor

dev: ## 开发模式 (启动测试环境)
	@echo "$(BLUE)启动开发环境...$(NC)"
	@docker-compose up -d
	@echo "$(GREEN)开发环境已启动$(NC)"
	@echo ""
	@echo "Redis: localhost:6379"
	@echo "Kafka: localhost:9092"
	@echo ""
	@echo "运行 'make run' 启动测试"

bench: ## 运行基准测试
	@echo "$(BLUE)运行基准测试...$(NC)"
	@go test -bench=. -benchmem ./...

check: fmt vet lint test-unit-quiet ## 运行所有检查 (格式化、vet、lint、测试)

pre-commit: check coverage-check ## Pre-commit检查 (格式、lint、测试、覆盖率)
	@echo "$(GREEN)所有检查通过!$(NC)"

setup: ## 设置开发环境
	@echo "$(BLUE)设置开发环境...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golang/mock/mockgen@latest
	@go mod download
	@echo "$(GREEN)开发环境设置完成$(NC)"
