#!/bin/bash

# MCT Service Platform Stop Script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

echo "========================================="
echo "Stopping MCT Service Platform"
echo "========================================="
echo ""

# Detect docker-compose command
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
else
    DOCKER_COMPOSE="docker compose"
fi

# Stop all services
$DOCKER_COMPOSE down

echo ""
echo "MCT Service Platform stopped successfully"
echo ""
