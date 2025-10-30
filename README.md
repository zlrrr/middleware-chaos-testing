# Middleware Chaos Testing (MCT)

中间件混沌测试框架 - 验证Kafka、Redis等中间件在混沌场景下的稳定性

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## 概述

MCT是一个可扩展的中间件混沌测试框架,通过模拟各类用户操作,检测并量化中间件服务(Kafka、Redis等)在混沌场景下的稳定性指标。

### 核心特性

- ✅ 支持多种中间件客户端(Redis、Kafka优先)
- ✅ 可配置的测试持续时间(命令行参数和配置文件)
- ✅ 智能稳定性评分系统(0-100分,5个等级)
- ✅ 明确的通过/警告/失败判断
- ✅ 可操作的改进建议(按优先级排序)
- ✅ 工程化的稳定性指标监测
- ✅ 容器化一键部署
- ✅ TDD/SDD开发流程

## 快速开始

### 环境要求

- Go 1.21+
- Docker 20.10+
- Docker Compose 2.0+

### 安装

```bash
# 克隆仓库
git clone https://github.com/username/middleware-chaos-testing.git
cd middleware-chaos-testing

# 下载依赖
make mod-download

# 构建
make build
```

### 启动测试环境

```bash
# 启动Redis和Kafka
make dev

# 或使用docker-compose
docker-compose up -d
```

### 运行测试

```bash
# Redis测试
./bin/mct test \
  --middleware redis \
  --host localhost \
  --port 6379 \
  --duration 30s \
  --operations 5000

# Kafka测试
./bin/mct test \
  --middleware kafka \
  --brokers localhost:9092 \
  --duration 30s \
  --operations 5000
```

## 项目结构

```
middleware-chaos-testing/
├── cmd/
│   └── mct/                  # 主程序入口
├── internal/
│   ├── core/                 # 核心抽象接口
│   ├── middleware/           # 中间件适配器
│   ├── metrics/              # 指标收集器
│   ├── detector/             # 稳定性检测器
│   ├── evaluator/            # 稳定性评估器
│   ├── reporter/             # 结果报告器
│   ├── orchestrator/         # 测试编排器
│   └── config/               # 配置管理
├── tests/
│   ├── unit/                 # 单元测试
│   ├── integration/          # 集成测试
│   └── e2e/                  # 端到端测试
├── docs/                     # 文档
├── configs/                  # 配置文件
├── scripts/                  # 脚本
├── Makefile                  # 构建脚本
├── Dockerfile                # Docker镜像
└── docker-compose.yml        # Docker编排
```

## 开发

### 设置开发环境

```bash
# 安装开发工具
make setup

# 启动开发环境
make dev
```

### 运行测试

```bash
# 单元测试
make test-unit

# 集成测试
make test-integration

# 所有测试
make test

# 覆盖率报告
make coverage
```

### 代码质量

```bash
# 代码格式化
make fmt

# 代码检查
make lint

# 运行所有检查
make check
```

## 评分体系

MCT采用0-100分的智能评分系统:

| 分数区间 | 等级 | 状态 | 说明 |
|---------|------|------|------|
| 90-100 | EXCELLENT | PASS | 优秀,可直接用于生产环境 |
| 80-89 | GOOD | PASS | 良好,满足生产要求 |
| 70-79 | FAIR | WARNING | 一般,建议优化后使用 |
| 60-69 | POOR | WARNING | 较差,需要改进 |
| 0-59 | FAILED | FAIL | 失败,不建议用于生产 |

### 评分维度

```
总分 = 可用性(30%) + 性能(25%) + 可靠性(25%) + 恢复力(20%)
```

详见[评分标准文档](docs/phase-0/evaluation-criteria.md)

## 测试报告示例

```
==========================================
   中间件稳定性测试报告
==========================================

测试目标: Redis @ localhost:6379
测试时长: 30s
测试完成: 2025-10-30 14:30:00

------------------------------------------
  总体评分: 87.5/100 (GOOD) ✅ PASS
------------------------------------------

各维度得分:
  ✓ 可用性   28.5/30  (95.0%)
  ✓ 性能     21.0/25  (84.0%)
  ✓ 可靠性   23.5/25  (94.0%)
  ✓ 恢复力   14.5/20  (72.5%)

------------------------------------------
  核心指标
------------------------------------------
可用性: 99.92% ✓
  - 总操作数: 10,000
  - 成功操作: 9,992
  - 失败操作: 8

性能指标:
  - P50 延迟: 8ms ✓
  - P95 延迟: 45ms ✓
  - P99 延迟: 120ms ⚠️

------------------------------------------
  改进建议 (按优先级排序)
------------------------------------------

[MEDIUM] 优化响应延迟
  1. 分析慢查询日志
  2. 检查网络延迟
  3. 优化数据结构
```

## 配置

支持YAML配置文件:

```yaml
# configs/test-redis.yaml
name: "Redis Stability Test"
middleware: "redis"

connection:
  host: "localhost"
  port: 6379
  timeout: 5s

test:
  duration: 60s
  operations: 10000
  concurrency: 10

thresholds:
  availability:
    excellent: 99.99%
    good: 99.9%
    pass: 95.0%

  p95_latency:
    excellent: 10ms
    good: 50ms
    pass: 200ms
```

## 文档

- [架构设计](docs/phase-0/architecture.md)
- [接口规范](docs/phase-0/interface-spec.md)
- [指标定义](docs/phase-0/metrics-definition.md)
- [评分标准](docs/phase-0/evaluation-criteria.md)
- [测试策略](docs/phase-0/testing-strategy.md)
- [开发指南](docs/phase-0/development-guide.md)

## 开发路线图

- [x] Phase 0: 项目初始化与架构设计
- [ ] Phase 1: Redis客户端实现
- [ ] Phase 2: Kafka客户端实现
- [ ] Phase 3: 稳定性检测器实现
- [ ] Phase 3.5: 稳定性评分系统
- [ ] Phase 4: CLI工具实现

详见[PLAN.md](PLAN.md)

## 贡献

欢迎贡献代码!请遵循以下步骤:

1. Fork本仓库
2. 创建特性分支 (`git checkout -b feature/your-feature`)
3. 提交更改 (`git commit -m 'Add some feature'`)
4. 推送到分支 (`git push origin feature/your-feature`)
5. 创建Pull Request

### 代码规范

- 遵循Go编码规范
- 所有代码必须通过`make check`
- 单元测试覆盖率 >= 85%
- 添加必要的文档和注释

## 许可证

本项目采用MIT许可证 - 详见[LICENSE](LICENSE)文件

## 致谢

- [go-redis](https://github.com/redis/go-redis) - Redis客户端
- [sarama](https://github.com/IBM/sarama) - Kafka客户端
- [cobra](https://github.com/spf13/cobra) - CLI框架

## 联系方式

- Issue: [GitHub Issues](https://github.com/username/middleware-chaos-testing/issues)
- Email: your.email@example.com
