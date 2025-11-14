#!/bin/bash

# MCT Service Platform Startup Script
# This script starts all services using Docker Compose

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

echo "========================================="
echo "MCT Service Platform Startup"
echo "========================================="
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "Error: Docker is not installed"
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "Error: Docker Compose is not installed"
    exit 1
fi

# Detect docker-compose command
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
else
    DOCKER_COMPOSE="docker compose"
fi

# Stop and remove existing containers
echo "Stopping existing containers..."
$DOCKER_COMPOSE down

# Build images
echo ""
echo "Building Docker images..."
$DOCKER_COMPOSE build

# Start services
echo ""
echo "Starting services..."
$DOCKER_COMPOSE up -d

# Wait for services to be ready
echo ""
echo "Waiting for services to be ready..."
sleep 5

# Check health
echo ""
echo "Checking service health..."

# Check mct-server health
for i in {1..30}; do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo "✓ MCT Server is healthy"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "✗ MCT Server failed to start"
    fi
    sleep 2
done

# Check mct-web health
for i in {1..30}; do
    if curl -s http://localhost:3000 > /dev/null 2>&1; then
        echo "✓ MCT Web UI is healthy"
        break
    fi
    if [ $i -eq 30 ]; then
        echo "✗ MCT Web UI failed to start"
    fi
    sleep 2
done

echo ""
echo "========================================="
echo "MCT Service Platform is Ready!"
echo "========================================="
echo ""
echo "Web UI:     http://localhost:3000"
echo "API Server: http://localhost:8080"
echo "API Docs:   http://localhost:8080/api/v1/middlewares"
echo ""
echo "Middleware Services:"
echo "  Redis:     localhost:6379"
echo "  Kafka:     localhost:9092"
echo "  MongoDB:   localhost:27017 (admin/password)"
echo "  RabbitMQ:  localhost:5672 (admin/password)"
echo "             Management: http://localhost:15672"
echo "  EMQX:      localhost:1883"
echo "             Dashboard: http://localhost:18083"
echo "  Nacos:     http://localhost:8848/nacos"
echo ""
echo "Useful commands:"
echo "  View logs:     $DOCKER_COMPOSE logs -f [service]"
echo "  Stop platform: $DOCKER_COMPOSE down"
echo "  Restart:       $DOCKER_COMPOSE restart [service]"
echo ""
