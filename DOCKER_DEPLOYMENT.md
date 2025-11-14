# MCT Platform Docker Deployment Guide

## Quick Start

### Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- 8GB RAM minimum (16GB recommended)
- 20GB free disk space

### Start the Platform

```bash
# Option 1: Use the startup script
./scripts/start-platform.sh

# Option 2: Use docker-compose directly
docker-compose up -d
```

### Stop the Platform

```bash
# Option 1: Use the stop script
./scripts/stop-platform.sh

# Option 2: Use docker-compose directly
docker-compose down
```

## Service Architecture

```
┌─────────────────────────────────────────────┐
│          MCT Service Platform               │
│                                             │
│  ┌──────────┐         ┌──────────┐         │
│  │ mct-web  │◄───────►│mct-server│         │
│  │ (React)  │  HTTP   │  (Go)    │         │
│  │  :3000   │         │  :8080   │         │
│  └──────────┘         └─────┬────┘         │
│                             │               │
│                    ┌────────┴────────┐      │
│                    │                 │      │
│         ┌──────────▼─────┐  ┌───────▼────┐ │
│         │  Middlewares   │  │ mct binary │ │
│         │                │  └────────────┘ │
│         │ - Redis        │                 │
│         │ - Kafka        │                 │
│         │ - MongoDB      │                 │
│         │ - RabbitMQ     │                 │
│         │ - EMQX         │                 │
│         │ - Nacos        │                 │
│         └────────────────┘                 │
└─────────────────────────────────────────────┘
```

## Services

### MCT Platform Services

| Service | Port | Description |
|---------|------|-------------|
| mct-web | 3000 | React frontend UI |
| mct-server | 8080 | Go backend API server |

### Middleware Services

| Service | Ports | Credentials | Description |
|---------|-------|-------------|-------------|
| Redis | 6379 | - | In-memory data store |
| Kafka | 9092, 29092 | - | Message streaming platform |
| Zookeeper | 2181 | - | Kafka dependency |
| MongoDB | 27017 | admin/password | Document database |
| RabbitMQ | 5672, 15672 | admin/password | Message broker |
| EMQX | 1883, 18083 | - | MQTT broker |
| Nacos | 8848, 9848 | - | Service discovery |

## Access URLs

### Platform

- **Web UI**: http://localhost:3000
- **API Server**: http://localhost:8080
- **API Health**: http://localhost:8080/health
- **API Docs**: http://localhost:8080/api/v1/middlewares

### Management Interfaces

- **RabbitMQ Management**: http://localhost:15672 (admin/password)
- **EMQX Dashboard**: http://localhost:18083
- **Nacos Console**: http://localhost:8848/nacos

## Docker Compose Commands

### View logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f mct-server
docker-compose logs -f mct-web
```

### Restart services

```bash
# Restart all
docker-compose restart

# Restart specific service
docker-compose restart mct-server
```

### Check service status

```bash
docker-compose ps
```

### Rebuild images

```bash
docker-compose build

# Rebuild specific service
docker-compose build mct-server
docker-compose build mct-web
```

### Clean up

```bash
# Stop and remove containers
docker-compose down

# Stop, remove containers and volumes
docker-compose down -v

# Stop, remove containers, volumes, and images
docker-compose down -v --rmi all
```

## Troubleshooting

### Service not starting

1. Check Docker is running:
   ```bash
   docker info
   ```

2. Check logs:
   ```bash
   docker-compose logs [service-name]
   ```

3. Verify port availability:
   ```bash
   netstat -tulpn | grep [port]
   ```

### Out of memory

Increase Docker memory limit in Docker Desktop settings (minimum 8GB).

### Port conflicts

If ports are already in use, modify `docker-compose.yml` to use different ports:

```yaml
services:
  mct-web:
    ports:
      - "3001:80"  # Change 3000 to 3001
```

### Rebuild from scratch

```bash
# Stop and remove everything
docker-compose down -v --rmi all

# Rebuild and start
docker-compose build --no-cache
docker-compose up -d
```

## Development Mode

### Run with live reload

For frontend development:

```bash
cd web
npm run dev
```

For backend development:

```bash
# Stop the mct-server container
docker-compose stop mct-server

# Run locally
go run cmd/mct-server/main.go
```

## Production Deployment

### Recommendations

1. **Use external databases** for production
2. **Enable TLS/SSL** for API and Web
3. **Configure resource limits**:
   ```yaml
   services:
     mct-server:
       deploy:
         resources:
           limits:
             cpus: '2'
             memory: 2G
   ```
4. **Set up monitoring** (Prometheus, Grafana)
5. **Configure logging** to external systems
6. **Use secrets management** for credentials
7. **Enable authentication** on API

## Environment Variables

### MCT Server

| Variable | Default | Description |
|----------|---------|-------------|
| MCT_BINARY_PATH | /usr/local/bin/mct | Path to mct binary |
| MCT_WORK_DIR | /tmp/mct-tasks | Working directory |
| MCT_SERVER_PORT | 8080 | Server port |
| GIN_MODE | release | Gin mode (debug/release) |

### Override in docker-compose.yml

```yaml
services:
  mct-server:
    environment:
      - MCT_SERVER_PORT=9090
      - GIN_MODE=debug
```

## Volumes

Persistent data is stored in Docker volumes:

- `mct-data`: MCT application data
- `mct-logs`: Application logs
- `redis-data`: Redis persistence
- `mongodb-data`: MongoDB data
- `rabbitmq-data`: RabbitMQ data

### Backup volumes

```bash
docker run --rm -v mct-data:/data -v $(pwd):/backup \
  alpine tar czf /backup/mct-data-backup.tar.gz -C /data .
```

### Restore volumes

```bash
docker run --rm -v mct-data:/data -v $(pwd):/backup \
  alpine tar xzf /backup/mct-data-backup.tar.gz -C /data
```

## Network

All services run on the `mct-network` bridge network. Services can communicate using their service names as hostnames.

Example:
- Frontend → Backend: `http://mct-server:8080`
- Backend → Redis: `redis:6379`
- Backend → MongoDB: `mongodb:27017`

## Health Checks

All services have health checks configured. View health status:

```bash
docker-compose ps
```

Healthy services show `(healthy)` status.

## Support

For issues and questions:
- GitHub Issues: https://github.com/zlrrr/middleware-chaos-testing/issues
- Documentation: README.md
