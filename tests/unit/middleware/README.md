# Redis客户端测试

## 测试类型

### 单元测试
不需要外部依赖，测试客户端的基本逻辑：
- 配置创建
- 未连接状态下的行为
- 操作类型实现

运行单元测试：
```bash
go test -short -v ./tests/unit/middleware/...
```

### 集成测试
需要真实的Redis服务器，测试完整功能：
- 连接管理
- SET/GET/DELETE操作
- 并发操作
- 错误处理
- 超时处理

## 运行集成测试

### 方式1：使用Docker Compose
```bash
# 启动Redis
docker-compose up -d redis

# 或使用测试环境
docker-compose -f docker-compose.test.yml up -d redis-test

# 运行集成测试（不加-short）
go test -v ./tests/unit/middleware/...

# 清理
docker-compose down
```

### 方式2：使用本地Redis
```bash
# 确保Redis运行在localhost:6379
redis-server

# 运行集成测试
go test -v ./tests/unit/middleware/...
```

## 测试结果

### 单元测试（不需要Redis）
```
=== RUN   TestRedisClient_UnitTests
--- PASS: TestRedisClient_UnitTests
=== RUN   TestRedisOperations
--- PASS: TestRedisOperations
PASS
```

### 集成测试（需要Redis）
```
=== RUN   TestRedisClientTestSuite
=== RUN   TestRedisClientTestSuite/TestConnect_Success
--- PASS: TestRedisClientTestSuite/TestConnect_Success
=== RUN   TestRedisClientTestSuite/TestExecute_SetOperation
--- PASS: TestRedisClientTestSuite/TestExecute_SetOperation
... (17个测试全部通过)
PASS
```

## 代码覆盖率

仅运行单元测试时覆盖率较低，因为大部分代码路径需要Redis连接。

运行完整集成测试时覆盖率预期 >= 85%。

## CI/CD建议

在CI环境中：
1. 使用Docker Compose启动Redis服务
2. 运行完整测试套件
3. 检查覆盖率阈值

示例 GitHub Actions 配置：
```yaml
services:
  redis:
    image: redis:7-alpine
    ports:
      - 6379:6379

steps:
  - name: Run tests
    run: go test -v -coverprofile=coverage.out ./...

  - name: Check coverage
    run: |
      coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
      if (( $(echo "$coverage < 85" | bc -l) )); then
        echo "Coverage $coverage% is below threshold 85%"
        exit 1
      fi
```
