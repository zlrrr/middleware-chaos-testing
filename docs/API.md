# MCT Platform API Documentation

## Overview

The MCT Platform provides a RESTful HTTP API for managing chaos testing tasks and retrieving test results. This document describes all available endpoints, request/response formats, and error handling.

**Base URL**: `http://localhost:8080`

**API Version**: `v1`

**Content-Type**: `application/json`

## Table of Contents

- [Authentication](#authentication)
- [Response Format](#response-format)
- [Error Codes](#error-codes)
- [Endpoints](#endpoints)
  - [Health Check](#health-check)
  - [Middleware Management](#middleware-management)
  - [Task Management](#task-management)
  - [Results](#results)
- [Data Models](#data-models)
- [Examples](#examples)

## Authentication

**Current Status**: No authentication required

**Future**: Bearer token authentication will be implemented in a future release.

## Response Format

All API responses follow a unified JSON format:

### Success Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    // Response data here
  }
}
```

### Error Response

```json
{
  "code": 1001,
  "message": "Task not found",
  "data": null
}
```

### Pagination Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [...],
    "total": 100,
    "page": 1,
    "page_size": 10,
    "total_pages": 10
  }
}
```

## Error Codes

| Code | Message | Description |
|------|---------|-------------|
| 0 | success | Request succeeded |
| 1000 | Invalid request | Request validation failed |
| 1001 | Task not found | Specified task does not exist |
| 1002 | Task already running | Cannot run a task that is already running |
| 1003 | Invalid middleware | Unsupported middleware type |
| 1004 | Invalid configuration | Middleware configuration is invalid |
| 1005 | Result not available | Test result not yet available |
| 5000 | Internal server error | Unexpected server error |
| 5001 | Storage error | Database operation failed |
| 5002 | Executor error | Task execution failed |

## Endpoints

### Health Check

#### GET /health

Check if the API server is healthy and responsive.

**Request**:
```bash
curl http://localhost:8080/health
```

**Response** (200 OK):
```json
{
  "status": "healthy",
  "timestamp": "2025-11-14T10:00:00Z",
  "version": "1.0.0"
}
```

**Response Codes**:
- `200 OK`: Server is healthy
- `503 Service Unavailable`: Server is unhealthy

---

### Middleware Management

#### GET /api/v1/middlewares

Get information about all supported middleware types.

**Request**:
```bash
curl http://localhost:8080/api/v1/middlewares
```

**Response** (200 OK):
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
      "supported_chaos": [
        "network_delay",
        "network_loss",
        "service_kill",
        "container_kill"
      ],
      "config_schema": {
        "host": {
          "type": "string",
          "required": true,
          "description": "Redis server address",
          "example": "localhost:6379"
        },
        "password": {
          "type": "string",
          "required": false,
          "description": "Redis password"
        },
        "db": {
          "type": "integer",
          "required": false,
          "default": 0,
          "description": "Redis database number"
        }
      }
    },
    {
      "name": "kafka",
      "display_name": "Apache Kafka",
      "description": "Distributed event streaming platform",
      "version": "latest",
      "supported_chaos": [
        "network_delay",
        "network_loss",
        "service_kill",
        "container_kill",
        "disk_io"
      ],
      "config_schema": {
        "brokers": {
          "type": "string",
          "required": true,
          "description": "Comma-separated broker addresses",
          "example": "localhost:9092"
        },
        "topic": {
          "type": "string",
          "required": true,
          "description": "Kafka topic name"
        }
      }
    },
    {
      "name": "mongodb",
      "display_name": "MongoDB",
      "description": "Document database",
      "version": "4.4",
      "supported_chaos": [
        "network_delay",
        "network_loss",
        "service_kill",
        "container_kill"
      ],
      "config_schema": {
        "uri": {
          "type": "string",
          "required": true,
          "description": "MongoDB connection URI",
          "example": "mongodb://localhost:27017"
        },
        "database": {
          "type": "string",
          "required": true,
          "description": "Database name"
        },
        "collection": {
          "type": "string",
          "required": true,
          "description": "Collection name"
        }
      }
    },
    {
      "name": "rabbitmq",
      "display_name": "RabbitMQ",
      "description": "Message broker",
      "version": "3.12",
      "supported_chaos": [
        "network_delay",
        "network_loss",
        "service_kill",
        "container_kill"
      ],
      "config_schema": {
        "url": {
          "type": "string",
          "required": true,
          "description": "RabbitMQ connection URL",
          "example": "amqp://guest:guest@localhost:5672/"
        },
        "queue": {
          "type": "string",
          "required": true,
          "description": "Queue name"
        }
      }
    },
    {
      "name": "emqx",
      "display_name": "EMQX",
      "description": "MQTT broker",
      "version": "5.8",
      "supported_chaos": [
        "network_delay",
        "network_loss",
        "service_kill",
        "container_kill"
      ],
      "config_schema": {
        "broker": {
          "type": "string",
          "required": true,
          "description": "MQTT broker address",
          "example": "tcp://localhost:1883"
        },
        "topic": {
          "type": "string",
          "required": true,
          "description": "MQTT topic"
        },
        "client_id": {
          "type": "string",
          "required": true,
          "description": "MQTT client ID"
        }
      }
    },
    {
      "name": "nacos",
      "display_name": "Nacos",
      "description": "Service discovery and configuration",
      "version": "2.4",
      "supported_chaos": [
        "network_delay",
        "network_loss",
        "service_kill",
        "container_kill"
      ],
      "config_schema": {
        "server_addr": {
          "type": "string",
          "required": true,
          "description": "Nacos server address",
          "example": "localhost:8848"
        },
        "namespace_id": {
          "type": "string",
          "required": false,
          "default": "public",
          "description": "Namespace ID"
        },
        "group": {
          "type": "string",
          "required": false,
          "default": "DEFAULT_GROUP",
          "description": "Configuration group"
        },
        "data_id": {
          "type": "string",
          "required": true,
          "description": "Configuration data ID"
        }
      }
    },
    {
      "name": "elasticsearch",
      "display_name": "Elasticsearch",
      "description": "Search and analytics engine",
      "version": "7.x",
      "supported_chaos": [
        "network_delay",
        "network_loss",
        "service_kill",
        "container_kill",
        "disk_io"
      ],
      "config_schema": {
        "addresses": {
          "type": "string",
          "required": true,
          "description": "Comma-separated ES addresses",
          "example": "http://localhost:9200"
        },
        "index": {
          "type": "string",
          "required": true,
          "description": "Index name"
        },
        "username": {
          "type": "string",
          "required": false,
          "description": "Username for authentication"
        },
        "password": {
          "type": "string",
          "required": false,
          "description": "Password for authentication"
        }
      }
    }
  ]
}
```

**Response Codes**:
- `200 OK`: Successfully retrieved middleware list

---

### Task Management

#### POST /api/v1/tasks

Create a new chaos testing task.

**Request**:
```bash
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

**Request Body**:
```json
{
  "middleware": "redis",
  "config": {
    "host": "redis:6379",
    "password": "",
    "db": 0
  }
}
```

**Request Fields**:
- `middleware` (string, required): Middleware type (redis, kafka, mongodb, rabbitmq, emqx, nacos, elasticsearch)
- `config` (object, required): Middleware-specific configuration

**Response** (200 OK):
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "task-1731600000",
    "middleware": "redis",
    "status": "pending",
    "start_time": "2025-11-14T10:00:00Z",
    "end_time": null,
    "config": "{\"host\":\"redis:6379\",\"password\":\"\",\"db\":0}",
    "error": ""
  }
}
```

**Response Codes**:
- `200 OK`: Task created successfully
- `400 Bad Request`: Invalid request body or configuration
- `500 Internal Server Error`: Failed to create task

---

#### GET /api/v1/tasks

Get a list of all tasks with optional filtering and pagination.

**Request**:
```bash
# Get all tasks
curl http://localhost:8080/api/v1/tasks

# Filter by middleware
curl http://localhost:8080/api/v1/tasks?middleware=redis

# Filter by status
curl http://localhost:8080/api/v1/tasks?status=completed

# Pagination
curl http://localhost:8080/api/v1/tasks?page=1&page_size=10

# Combined filters
curl "http://localhost:8080/api/v1/tasks?middleware=redis&status=completed&page=1&page_size=20"
```

**Query Parameters**:
- `middleware` (string, optional): Filter by middleware type
- `status` (string, optional): Filter by status (pending, running, completed, failed)
- `page` (integer, optional): Page number (default: 1)
- `page_size` (integer, optional): Items per page (default: 20, max: 100)

**Response** (200 OK):
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": "task-1731600000",
        "middleware": "redis",
        "status": "completed",
        "start_time": "2025-11-14T10:00:00Z",
        "end_time": "2025-11-14T10:01:30Z",
        "config": "{\"host\":\"redis:6379\"}",
        "error": ""
      },
      {
        "id": "task-1731600100",
        "middleware": "kafka",
        "status": "running",
        "start_time": "2025-11-14T10:05:00Z",
        "end_time": null,
        "config": "{\"brokers\":\"kafka:9092\"}",
        "error": ""
      }
    ],
    "total": 25,
    "page": 1,
    "page_size": 10,
    "total_pages": 3
  }
}
```

**Response Codes**:
- `200 OK`: Successfully retrieved task list
- `400 Bad Request`: Invalid query parameters
- `500 Internal Server Error`: Failed to retrieve tasks

---

#### GET /api/v1/tasks/:id

Get details of a specific task.

**Request**:
```bash
curl http://localhost:8080/api/v1/tasks/task-1731600000
```

**URL Parameters**:
- `id` (string, required): Task ID

**Response** (200 OK):
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "task-1731600000",
    "middleware": "redis",
    "status": "completed",
    "start_time": "2025-11-14T10:00:00Z",
    "end_time": "2025-11-14T10:01:30Z",
    "config": "{\"host\":\"redis:6379\",\"password\":\"\",\"db\":0}",
    "error": ""
  }
}
```

**Response Codes**:
- `200 OK`: Successfully retrieved task
- `404 Not Found`: Task not found
- `500 Internal Server Error`: Failed to retrieve task

---

#### POST /api/v1/tasks/:id/run

Execute a chaos testing task.

**Request**:
```bash
curl -X POST http://localhost:8080/api/v1/tasks/task-1731600000/run
```

**URL Parameters**:
- `id` (string, required): Task ID

**Response** (200 OK):
```json
{
  "code": 0,
  "message": "Task started successfully",
  "data": null
}
```

**Response Codes**:
- `200 OK`: Task started successfully
- `400 Bad Request`: Task is already running or completed
- `404 Not Found`: Task not found
- `500 Internal Server Error`: Failed to start task

**Note**: This is an asynchronous operation. The task runs in the background. Use GET /api/v1/tasks/:id to check status.

---

#### DELETE /api/v1/tasks/:id

Delete a task and its associated results.

**Request**:
```bash
curl -X DELETE http://localhost:8080/api/v1/tasks/task-1731600000
```

**URL Parameters**:
- `id` (string, required): Task ID

**Response** (200 OK):
```json
{
  "code": 0,
  "message": "Task deleted successfully",
  "data": null
}
```

**Response Codes**:
- `200 OK`: Task deleted successfully
- `400 Bad Request`: Cannot delete running task
- `404 Not Found`: Task not found
- `500 Internal Server Error`: Failed to delete task

---

### Results

#### GET /api/v1/tasks/:id/result

Get the test result for a completed task.

**Request**:
```bash
curl http://localhost:8080/api/v1/tasks/task-1731600000/result
```

**URL Parameters**:
- `id` (string, required): Task ID

**Response** (200 OK):
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "task_id": "task-1731600000",
    "score": 87.5,
    "grade": "B",
    "dimensions": {
      "availability": 92.0,
      "performance": 85.0,
      "resilience": 88.0,
      "data_integrity": 95.0,
      "recovery": 77.5
    },
    "metrics": {
      "total_operations": 10000,
      "successful_operations": 9850,
      "failed_operations": 150,
      "average_latency_ms": 12.5,
      "p95_latency_ms": 25.0,
      "p99_latency_ms": 45.0,
      "error_rate": 1.5,
      "recovery_time_ms": 3500
    },
    "issues": [
      {
        "severity": "Medium",
        "type": "Performance",
        "description": "Elevated latency during network delay scenario",
        "timestamp": "2025-11-14T10:00:45Z",
        "details": {
          "scenario": "network_delay",
          "observed_latency": 45.0,
          "threshold": 30.0
        }
      },
      {
        "severity": "Low",
        "type": "Availability",
        "description": "Brief connection timeout during service kill",
        "timestamp": "2025-11-14T10:01:10Z",
        "details": {
          "scenario": "service_kill",
          "duration_ms": 500
        }
      }
    ],
    "recommendations": [
      {
        "priority": "High",
        "category": "Performance",
        "suggestion": "Enable connection pooling to reduce latency",
        "impact": "Expected 20-30% latency reduction",
        "implementation": "Configure minIdle and maxTotal in connection pool settings"
      },
      {
        "priority": "Medium",
        "category": "Resilience",
        "suggestion": "Implement retry mechanism with exponential backoff",
        "impact": "Improved fault tolerance during transient failures",
        "implementation": "Add retry logic with 3 attempts and exponential backoff"
      }
    ],
    "test_duration_ms": 90000,
    "timestamp": "2025-11-14T10:01:30Z"
  }
}
```

**Response Codes**:
- `200 OK`: Successfully retrieved result
- `404 Not Found`: Task or result not found
- `400 Bad Request`: Result not yet available (task not completed)
- `500 Internal Server Error`: Failed to retrieve result

---

## Data Models

### Task

```typescript
interface Task {
  id: string;              // Unique task identifier
  middleware: string;      // Middleware type
  status: TaskStatus;      // Current status
  start_time: string;      // ISO 8601 timestamp
  end_time: string | null; // ISO 8601 timestamp or null if not finished
  config: string;          // JSON-encoded configuration
  error: string;           // Error message if failed
}

enum TaskStatus {
  pending = "pending",
  running = "running",
  completed = "completed",
  failed = "failed"
}
```

### TestResult

```typescript
interface TestResult {
  task_id: string;
  score: number;                    // Overall score (0-100)
  grade: Grade;                     // Letter grade
  dimensions: DimensionScores;
  metrics: StabilityMetrics;
  issues: Issue[];
  recommendations: Recommendation[];
  test_duration_ms: number;
  timestamp: string;                // ISO 8601 timestamp
}

enum Grade {
  A = "A",  // 90-100
  B = "B",  // 80-89
  C = "C",  // 70-79
  D = "D",  // 60-69
  F = "F"   // 0-59
}

interface DimensionScores {
  availability: number;     // 0-100
  performance: number;      // 0-100
  resilience: number;       // 0-100
  data_integrity: number;   // 0-100
  recovery: number;         // 0-100
}

interface StabilityMetrics {
  total_operations: number;
  successful_operations: number;
  failed_operations: number;
  average_latency_ms: number;
  p95_latency_ms: number;
  p99_latency_ms: number;
  error_rate: number;           // Percentage
  recovery_time_ms: number;
}

interface Issue {
  severity: Severity;
  type: string;
  description: string;
  timestamp: string;            // ISO 8601 timestamp
  details: Record<string, any>;
}

enum Severity {
  Critical = "Critical",
  High = "High",
  Medium = "Medium",
  Low = "Low"
}

interface Recommendation {
  priority: Priority;
  category: string;
  suggestion: string;
  impact: string;
  implementation: string;
}

enum Priority {
  High = "High",
  Medium = "Medium",
  Low = "Low"
}
```

### MiddlewareInfo

```typescript
interface MiddlewareInfo {
  name: string;
  display_name: string;
  description: string;
  version: string;
  supported_chaos: string[];
  config_schema: Record<string, ConfigField>;
}

interface ConfigField {
  type: string;
  required: boolean;
  description: string;
  default?: any;
  example?: string;
}
```

### Pagination

```typescript
interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}
```

---

## Examples

### Complete Workflow Example

```bash
#!/bin/bash

# 1. Check API health
curl http://localhost:8080/health

# 2. Get available middlewares
curl http://localhost:8080/api/v1/middlewares

# 3. Create a Redis test task
TASK_RESPONSE=$(curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "middleware": "redis",
    "config": {
      "host": "redis:6379",
      "password": "",
      "db": 0
    }
  }')

echo "Create task response: $TASK_RESPONSE"

# Extract task ID using jq
TASK_ID=$(echo $TASK_RESPONSE | jq -r '.data.id')
echo "Task ID: $TASK_ID"

# 4. Run the task
curl -X POST http://localhost:8080/api/v1/tasks/$TASK_ID/run

# 5. Poll for completion
while true; do
  TASK=$(curl -s http://localhost:8080/api/v1/tasks/$TASK_ID)
  STATUS=$(echo $TASK | jq -r '.data.status')

  echo "Current status: $STATUS"

  if [ "$STATUS" = "completed" ] || [ "$STATUS" = "failed" ]; then
    break
  fi

  sleep 5
done

# 6. Get the result
if [ "$STATUS" = "completed" ]; then
  RESULT=$(curl -s http://localhost:8080/api/v1/tasks/$TASK_ID/result)
  echo "Test result:"
  echo $RESULT | jq '.'

  SCORE=$(echo $RESULT | jq -r '.data.score')
  GRADE=$(echo $RESULT | jq -r '.data.grade')
  echo "Score: $SCORE, Grade: $GRADE"
else
  echo "Task failed"
  ERROR=$(echo $TASK | jq -r '.data.error')
  echo "Error: $ERROR"
fi

# 7. Clean up (optional)
# curl -X DELETE http://localhost:8080/api/v1/tasks/$TASK_ID
```

### Multiple Middleware Tests

```bash
#!/bin/bash

MIDDLEWARES=("redis" "kafka" "mongodb" "rabbitmq" "emqx" "nacos")

for MW in "${MIDDLEWARES[@]}"; do
  echo "Testing $MW..."

  # Create task
  TASK_ID=$(curl -s -X POST http://localhost:8080/api/v1/tasks \
    -H "Content-Type: application/json" \
    -d "{\"middleware\":\"$MW\",\"config\":{}}" \
    | jq -r '.data.id')

  # Run task
  curl -s -X POST http://localhost:8080/api/v1/tasks/$TASK_ID/run > /dev/null

  # Wait and get result
  sleep 60

  RESULT=$(curl -s http://localhost:8080/api/v1/tasks/$TASK_ID/result)
  SCORE=$(echo $RESULT | jq -r '.data.score')

  echo "$MW Score: $SCORE"
done
```

### Filtering and Pagination

```bash
# Get first page of completed Redis tasks
curl "http://localhost:8080/api/v1/tasks?middleware=redis&status=completed&page=1&page_size=10"

# Get second page
curl "http://localhost:8080/api/v1/tasks?middleware=redis&status=completed&page=2&page_size=10"

# Get all running tasks
curl "http://localhost:8080/api/v1/tasks?status=running"

# Get all tasks for Kafka
curl "http://localhost:8080/api/v1/tasks?middleware=kafka"
```

### JavaScript/TypeScript Client

```typescript
class MCTClient {
  private baseURL: string;

  constructor(baseURL: string = 'http://localhost:8080') {
    this.baseURL = baseURL;
  }

  async health(): Promise<any> {
    const response = await fetch(`${this.baseURL}/health`);
    return response.json();
  }

  async getMiddlewares(): Promise<any> {
    const response = await fetch(`${this.baseURL}/api/v1/middlewares`);
    const data = await response.json();
    return data.data;
  }

  async createTask(middleware: string, config: any): Promise<string> {
    const response = await fetch(`${this.baseURL}/api/v1/tasks`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ middleware, config })
    });
    const data = await response.json();
    return data.data.id;
  }

  async runTask(taskId: string): Promise<void> {
    await fetch(`${this.baseURL}/api/v1/tasks/${taskId}/run`, {
      method: 'POST'
    });
  }

  async getTask(taskId: string): Promise<any> {
    const response = await fetch(`${this.baseURL}/api/v1/tasks/${taskId}`);
    const data = await response.json();
    return data.data;
  }

  async getTaskResult(taskId: string): Promise<any> {
    const response = await fetch(`${this.baseURL}/api/v1/tasks/${taskId}/result`);
    const data = await response.json();
    return data.data;
  }

  async waitForCompletion(taskId: string, pollInterval: number = 5000): Promise<any> {
    while (true) {
      const task = await this.getTask(taskId);
      if (task.status === 'completed' || task.status === 'failed') {
        return task;
      }
      await new Promise(resolve => setTimeout(resolve, pollInterval));
    }
  }
}

// Usage
const client = new MCTClient();

async function runTest() {
  // Create task
  const taskId = await client.createTask('redis', {
    host: 'redis:6379',
    db: 0
  });

  // Run task
  await client.runTask(taskId);

  // Wait for completion
  const task = await client.waitForCompletion(taskId);

  if (task.status === 'completed') {
    // Get result
    const result = await client.getTaskResult(taskId);
    console.log(`Score: ${result.score}, Grade: ${result.grade}`);
  } else {
    console.error(`Task failed: ${task.error}`);
  }
}

runTest();
```

### Python Client

```python
import requests
import time
from typing import Dict, Any, Optional

class MCTClient:
    def __init__(self, base_url: str = 'http://localhost:8080'):
        self.base_url = base_url

    def health(self) -> Dict[str, Any]:
        response = requests.get(f'{self.base_url}/health')
        return response.json()

    def get_middlewares(self) -> list:
        response = requests.get(f'{self.base_url}/api/v1/middlewares')
        return response.json()['data']

    def create_task(self, middleware: str, config: Dict[str, Any]) -> str:
        response = requests.post(
            f'{self.base_url}/api/v1/tasks',
            json={'middleware': middleware, 'config': config}
        )
        return response.json()['data']['id']

    def run_task(self, task_id: str) -> None:
        requests.post(f'{self.base_url}/api/v1/tasks/{task_id}/run')

    def get_task(self, task_id: str) -> Dict[str, Any]:
        response = requests.get(f'{self.base_url}/api/v1/tasks/{task_id}')
        return response.json()['data']

    def get_task_result(self, task_id: str) -> Dict[str, Any]:
        response = requests.get(f'{self.base_url}/api/v1/tasks/{task_id}/result')
        return response.json()['data']

    def wait_for_completion(self, task_id: str, poll_interval: int = 5) -> Dict[str, Any]:
        while True:
            task = self.get_task(task_id)
            if task['status'] in ['completed', 'failed']:
                return task
            time.sleep(poll_interval)

# Usage
client = MCTClient()

# Create and run test
task_id = client.create_task('redis', {'host': 'redis:6379', 'db': 0})
client.run_task(task_id)

# Wait for completion
task = client.wait_for_completion(task_id)

if task['status'] == 'completed':
    result = client.get_task_result(task_id)
    print(f"Score: {result['score']}, Grade: {result['grade']}")
else:
    print(f"Task failed: {task['error']}")
```

---

## Rate Limiting

**Current Status**: No rate limiting implemented

**Future**: Rate limiting will be added based on API key or IP address.

---

## Versioning

The API uses URI versioning. The current version is `v1` as indicated in the base path `/api/v1`.

Future versions will be released as `/api/v2`, `/api/v3`, etc., with appropriate deprecation notices.

---

## Support

For issues, questions, or feature requests:
- GitHub Issues: https://github.com/zlrrr/middleware-chaos-testing/issues
- Documentation: See `docs/platform-guide.md`
- Main README: See `README.md`
