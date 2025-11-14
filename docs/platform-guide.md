# MCT Platform User Guide

## Overview

The Middleware Chaos Testing (MCT) Platform provides a comprehensive solution for testing middleware stability under various chaos scenarios. This guide will help you get started with the platform and make the most of its features.

## Table of Contents

- [Quick Start](#quick-start)
- [Using the Web UI](#using-the-web-ui)
- [Creating Test Tasks](#creating-test-tasks)
- [Viewing Test Results](#viewing-test-results)
- [Using the API](#using-the-api)
- [WebSocket Subscriptions](#websocket-subscriptions)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Quick Start

### Prerequisites

- Docker 20.10+ and Docker Compose 2.0+
- 8GB RAM minimum (16GB recommended)
- 20GB free disk space

### Starting the Platform

```bash
# Clone the repository
git clone https://github.com/zlrrr/middleware-chaos-testing.git
cd middleware-chaos-testing

# Start all services
./scripts/start-platform.sh
```

Wait for all services to be healthy. The script will display access URLs when ready.

### Accessing the Platform

- **Web UI**: http://localhost:3000
- **API Server**: http://localhost:8080
- **API Health Check**: http://localhost:8080/health

## Using the Web UI

### Dashboard

The dashboard provides an overview of your testing activities:

- **Task Statistics**: Total tasks, running tasks, completed tasks, failed tasks
- **Middleware Coverage**: Which middleware types have been tested
- **Recent Activity**: Latest test results and their scores

### Navigation

- **Dashboard** (📊): Overview and statistics
- **Tasks** (📋): Task management and execution
- **Results** (📈): Detailed test result analysis

## Creating Test Tasks

### Via Web UI

1. Navigate to the **Tasks** page
2. Click the **"Create Task"** button
3. Fill in the task details:

   **Middleware Selection**: Choose from 7 supported middleware types:
   - Redis
   - Kafka
   - MongoDB
   - RabbitMQ
   - EMQX
   - Nacos
   - Elasticsearch

   **Configuration**: Each middleware has specific configuration fields:

   **Redis Example**:
   ```
   Host: redis:6379
   Password: (leave empty for no auth)
   DB: 0
   ```

   **Kafka Example**:
   ```
   Brokers: kafka:9092
   Topic: test-topic
   ```

   **MongoDB Example**:
   ```
   URI: mongodb://admin:password@mongodb:27017
   Database: testdb
   Collection: testcol
   ```

   **RabbitMQ Example**:
   ```
   URL: amqp://admin:password@rabbitmq:5672/
   Queue: test-queue
   ```

   **EMQX Example**:
   ```
   Broker: tcp://emqx:1883
   Topic: test/topic
   Client ID: test-client
   ```

   **Nacos Example**:
   ```
   Server Addr: nacos:8848
   Namespace ID: public
   Group: DEFAULT_GROUP
   Data ID: test-config
   ```

4. Click **"Create"** to create the task
5. The task will appear in the task list with status "pending"

### Via API

```bash
# Create a Redis test task
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "middleware": "redis",
    "config": {
      "host": "redis:6379",
      "password": "",
      "db": 0
    }
  }'
```

Response:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "task-1731600000",
    "middleware": "redis",
    "status": "pending",
    "start_time": "2025-11-14T10:00:00Z",
    "config": "{\"host\":\"redis:6379\",\"password\":\"\",\"db\":0}"
  }
}
```

## Running Test Tasks

### Via Web UI

1. Navigate to the **Tasks** page
2. Find the task you want to run
3. Click the **"Run"** button (▶️) in the Actions column
4. The task status will change to "running"
5. Wait for the test to complete (typically 30-60 seconds)
6. Status will change to "completed" or "failed"
7. Click **"View Results"** (👁️) to see detailed results

### Via API

```bash
# Run a task
curl -X POST http://localhost:8080/api/v1/tasks/task-1731600000/run

# Check task status
curl http://localhost:8080/api/v1/tasks/task-1731600000

# Get task result
curl http://localhost:8080/api/v1/tasks/task-1731600000/result
```

## Viewing Test Results

### Result Overview

When viewing a test result, you'll see:

1. **Overall Score Card**:
   - Total score (0-100)
   - Grade (A, B, C, D, F)
   - Five dimensional scores:
     - Availability
     - Performance
     - Resilience
     - Data Integrity
     - Recovery

2. **Key Metrics**:
   - Total operations
   - Successful operations
   - Failed operations
   - Average latency
   - P95/P99 latency
   - Error rate
   - Recovery time

3. **Issues Found**:
   - Severity level (Critical, High, Medium, Low)
   - Issue type
   - Description
   - Timestamp

4. **Recommendations**:
   - Priority level
   - Category
   - Actionable suggestions
   - Expected impact

### Understanding Scores

**Grade Scale**:
- **A (90-100)**: Excellent stability, production-ready
- **B (80-89)**: Good stability, minor improvements needed
- **C (70-79)**: Acceptable stability, several issues to address
- **D (60-69)**: Poor stability, significant improvements required
- **F (0-59)**: Critical issues, not production-ready

**Dimensional Scores**:

1. **Availability** (0-100):
   - Measures middleware uptime and accessibility
   - Key metric: Success rate during chaos scenarios

2. **Performance** (0-100):
   - Measures response time and throughput
   - Key metric: Latency percentiles (P50, P95, P99)

3. **Resilience** (0-100):
   - Measures ability to handle failures
   - Key metric: Recovery from network/service disruptions

4. **Data Integrity** (0-100):
   - Measures data consistency and correctness
   - Key metric: Data loss and corruption detection

5. **Recovery** (0-100):
   - Measures time to recover from failures
   - Key metric: Recovery time objective (RTO)

## Using the API

### List All Middlewares

```bash
curl http://localhost:8080/api/v1/middlewares
```

Response:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "name": "redis",
      "display_name": "Redis",
      "description": "In-memory data structure store",
      "version": "7.0",
      "supported_chaos": ["network_delay", "network_loss", "service_kill", "container_kill"]
    },
    ...
  ]
}
```

### List Tasks

```bash
# Get all tasks
curl http://localhost:8080/api/v1/tasks

# Filter by middleware
curl http://localhost:8080/api/v1/tasks?middleware=redis

# Filter by status
curl http://localhost:8080/api/v1/tasks?status=completed

# Pagination
curl http://localhost:8080/api/v1/tasks?page=1&page_size=10
```

### Get Task Details

```bash
curl http://localhost:8080/api/v1/tasks/task-1731600000
```

### Delete Task

```bash
curl -X DELETE http://localhost:8080/api/v1/tasks/task-1731600000
```

### Health Check

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2025-11-14T10:00:00Z",
  "version": "1.0.0"
}
```

## WebSocket Subscriptions

### Subscribe to Task Updates

```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/ws/tasks');

ws.onopen = () => {
  console.log('Connected to task updates');
};

ws.onmessage = (event) => {
  const update = JSON.parse(event.data);
  console.log('Task update:', update);
  // update = { task_id, status, progress, ... }
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('Disconnected from task updates');
};
```

### Subscribe to Specific Task

```javascript
const taskId = 'task-1731600000';
const ws = new WebSocket(`ws://localhost:8080/api/v1/ws/tasks/${taskId}`);

ws.onmessage = (event) => {
  const result = JSON.parse(event.data);
  if (result.status === 'completed') {
    console.log('Task completed!', result);
  }
};
```

## Best Practices

### Task Configuration

1. **Use Realistic Configurations**:
   - Mirror your production middleware settings
   - Use appropriate connection pools and timeouts
   - Configure proper resource limits

2. **Start with Basic Tests**:
   - Run baseline tests without chaos first
   - Gradually introduce chaos scenarios
   - Compare results to identify issues

3. **Test Incrementally**:
   - Test one middleware at a time
   - Verify each dimension separately
   - Combine scenarios after individual validation

### Interpreting Results

1. **Focus on Trends**:
   - Run multiple tests over time
   - Track score improvements
   - Monitor regression

2. **Prioritize Critical Issues**:
   - Address "Critical" and "High" severity issues first
   - Implement recommendations with high expected impact
   - Verify fixes with follow-up tests

3. **Understand Context**:
   - Consider your use case requirements
   - Some scenarios may not apply to your deployment
   - Adjust testing parameters based on your SLAs

### Performance Optimization

1. **Task Management**:
   - Delete old completed tasks regularly
   - Don't run too many tasks concurrently
   - Monitor resource usage

2. **Configuration Tuning**:
   - Adjust chaos parameters based on your needs
   - Use appropriate test durations
   - Configure timeouts properly

## Troubleshooting

### Task Stuck in "Running" Status

**Symptoms**: Task remains in "running" state for extended period

**Solutions**:
1. Check mct-server logs:
   ```bash
   docker-compose logs -f mct-server
   ```

2. Verify middleware service is healthy:
   ```bash
   docker-compose ps
   ```

3. Check task result for partial data:
   ```bash
   curl http://localhost:8080/api/v1/tasks/<task-id>/result
   ```

4. Restart the task if necessary:
   ```bash
   # Delete stuck task
   curl -X DELETE http://localhost:8080/api/v1/tasks/<task-id>

   # Create new task
   curl -X POST http://localhost:8080/api/v1/tasks -d '{...}'
   ```

### Task Failed Immediately

**Symptoms**: Task status changes to "failed" within seconds

**Common Causes**:
1. **Incorrect configuration**: Verify middleware connection details
2. **Middleware not accessible**: Ensure middleware service is running
3. **Invalid credentials**: Check username/password for authenticated services
4. **Network issues**: Verify Docker network connectivity

**Solutions**:
1. Check task error message:
   ```bash
   curl http://localhost:8080/api/v1/tasks/<task-id>
   ```

2. Verify middleware accessibility:
   ```bash
   # Redis
   docker exec -it mct-redis redis-cli ping

   # MongoDB
   docker exec -it mct-mongodb mongo --eval "db.adminCommand('ping')"

   # RabbitMQ
   docker exec -it mct-rabbitmq rabbitmq-diagnostics ping
   ```

3. Check mct-server can reach middleware:
   ```bash
   docker exec -it mct-server ping redis
   docker exec -it mct-server ping mongodb
   ```

### Low Test Scores

**Symptoms**: Consistently getting low scores (D or F grades)

**Analysis**:
1. Check which dimensions have low scores
2. Review issues list for specific problems
3. Examine recommendations for guidance

**Common Issues**:
1. **High latency**: Middleware may be under-resourced
   - Solution: Increase Docker memory/CPU limits

2. **Connection timeouts**: Network or service overload
   - Solution: Adjust timeout configurations, reduce load

3. **Data loss**: Persistence not configured properly
   - Solution: Verify volume mounts and persistence settings

4. **Slow recovery**: Insufficient replicas or resources
   - Solution: Enable HA mode, add more resources

### API Errors

**401 Unauthorized**: Authentication not yet implemented (future feature)

**404 Not Found**: Check endpoint URL and task ID

**500 Internal Server Error**:
- Check mct-server logs
- Verify middleware services are running
- Check for resource exhaustion

### UI Not Loading

**Symptoms**: Web UI doesn't load or shows errors

**Solutions**:
1. Verify mct-web container is running:
   ```bash
   docker-compose ps mct-web
   ```

2. Check mct-web logs:
   ```bash
   docker-compose logs -f mct-web
   ```

3. Verify API server is accessible:
   ```bash
   curl http://localhost:8080/health
   ```

4. Clear browser cache and reload

5. Check browser console for JavaScript errors

## Advanced Usage

### Custom Chaos Scenarios

Currently, the MCT tool uses predefined chaos scenarios. Future versions will support custom scenarios via the API.

### Automated Testing

Integrate MCT into your CI/CD pipeline:

```bash
#!/bin/bash
# ci-test.sh

# Start platform
./scripts/start-platform.sh

# Create test task
TASK_ID=$(curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"middleware":"redis","config":{"host":"redis:6379"}}' \
  | jq -r '.data.id')

# Run task
curl -X POST http://localhost:8080/api/v1/tasks/$TASK_ID/run

# Wait for completion
while true; do
  STATUS=$(curl -s http://localhost:8080/api/v1/tasks/$TASK_ID | jq -r '.data.status')
  if [ "$STATUS" = "completed" ]; then
    break
  fi
  sleep 5
done

# Check score
SCORE=$(curl -s http://localhost:8080/api/v1/tasks/$TASK_ID/result | jq -r '.data.score')

if [ $(echo "$SCORE >= 80" | bc) -eq 1 ]; then
  echo "✓ Test passed with score: $SCORE"
  exit 0
else
  echo "✗ Test failed with score: $SCORE"
  exit 1
fi
```

### Monitoring and Alerting

Export test results to your monitoring system:

```bash
# Get latest result
curl http://localhost:8080/api/v1/tasks/<task-id>/result \
  | jq '{
      task_id: .data.task_id,
      score: .data.score,
      grade: .data.grade,
      availability: .data.dimensions.availability,
      performance: .data.dimensions.performance,
      critical_issues: [.data.issues[] | select(.severity == "Critical")]
    }'
```

## Support and Resources

- **GitHub Repository**: https://github.com/zlrrr/middleware-chaos-testing
- **Issue Tracker**: https://github.com/zlrrr/middleware-chaos-testing/issues
- **API Documentation**: See `docs/API.md`
- **Deployment Guide**: See `DOCKER_DEPLOYMENT.md`
- **Main Documentation**: See `README.md`

## Appendix

### Supported Middleware Matrix

| Middleware | Version | Connection Modes | HA Support | Persistence |
|------------|---------|------------------|------------|-------------|
| Redis | 7.0 | Standalone, Cluster | ✓ | RDB, AOF |
| Kafka | Latest | Single broker, Multi-broker | ✓ | Log segments |
| MongoDB | 4.4.13 | Standalone, Replica Set | ✓ | WiredTiger |
| RabbitMQ | 3.12.7 | Standalone, Cluster | ✓ | Durable queues |
| EMQX | 5.8 | Standalone, Cluster | ✓ | Message persistence |
| Nacos | 2.4.3 | Standalone, Cluster | ✓ | Derby, MySQL |
| Elasticsearch | 7.x | Standalone, Cluster | ✓ | Index segments |

### Chaos Scenario Types

1. **Network Delay**: Introduces latency to network traffic
2. **Network Loss**: Simulates packet loss
3. **Service Kill**: Terminates middleware process
4. **Container Kill**: Stops middleware container
5. **Resource Limit**: Constrains CPU/memory
6. **Disk IO**: Throttles disk operations
7. **Clock Skew**: Adjusts system time

### Evaluation Criteria

Each dimension is scored based on specific metrics:

**Availability** = (Successful Operations / Total Operations) × 100

**Performance** = 100 - (Normalized Latency Degradation)

**Resilience** = (Operations During Chaos / Baseline Operations) × 100

**Data Integrity** = 100 - (Data Loss Percentage)

**Recovery** = 100 - (Normalized Recovery Time)

**Overall Score** = Weighted average of all dimensions
