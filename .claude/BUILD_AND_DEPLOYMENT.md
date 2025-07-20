# Build and Deployment Guide

## Overview

This guide provides comprehensive instructions for building and deploying the Relayooor platform across different environments.

## Prerequisites

### Development Environment
- Docker 20.10+ and Docker Compose 2.0+
- Node.js 18+ and Yarn 1.22+
- Go 1.21+
- PostgreSQL client tools
- Git

### Production Environment
- Docker and Docker Compose
- Domain with SSL certificates
- PostgreSQL database
- Redis instance
- RPC endpoints for supported chains

## Build Process

### 1. Local Development Build

#### Complete Stack Build
```bash
# Clone repository
git clone https://github.com/your-org/relayooor.git
cd relayooor

# Copy and configure environment
cp .env.example .env
# Edit .env with your configuration

# Build frontend first (CRITICAL!)
cd webapp
yarn install
yarn build
cd ..

# Start all services
./scripts/setup-and-launch.sh
```

#### Manual Build Steps
```bash
# 1. Build frontend
cd webapp
yarn install
yarn build
cd ..

# 2. Build Docker images
docker-compose build --no-cache

# 3. Start services
docker-compose up -d

# 4. Run database migrations
docker-compose exec api-backend go run . migrate up
```

### 2. Frontend-Only Build
```bash
cd webapp
yarn install
yarn build

# For development with hot reload (requires Docker backend)
yarn dev
```

### 3. API-Only Build
```bash
cd relayer-middleware/api
go mod download
go build -o api ./cmd/server
```

### 4. Production Build
```bash
# Use production Docker Compose
docker-compose -f docker-compose.yml -f docker-compose.prod.yml build

# Or use multi-stage builds
docker build -t relayooor/webapp:latest -f webapp/Dockerfile ./webapp
docker build -t relayooor/api:latest -f relayer-middleware/api/Dockerfile ./relayer-middleware/api
```

## Deployment Configurations

### 1. Docker Compose Variants

#### Standard Development (docker-compose.yml)
```yaml
version: '3.8'
services:
  webapp:
    build: ./webapp
    ports:
      - "80:80"
    depends_on:
      - api-backend
  
  api-backend:
    build: ./relayer-middleware/api
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgresql://relayooor:relayooor@postgres:5432/relayooor
    depends_on:
      - postgres
      - redis
```

#### Full Stack (docker-compose.full.yml)
Includes:
- All core services
- Chainpulse monitoring
- Hermes relayer
- Grafana dashboards
- Prometheus metrics

#### Webapp Only (docker-compose.webapp.yml)
For frontend development with external API

#### Local Development (docker-compose.local.yml)
Optimized for local development with:
- Volume mounts for hot reload
- Debug ports exposed
- Local database data persistence

### 2. Environment Configuration

#### Development (.env.development)
```env
# API Configuration
API_URL=http://localhost:8080
JWT_SECRET=dev-secret-change-in-production

# Database
POSTGRES_USER=relayooor_user
POSTGRES_PASSWORD=relayooor_dev
POSTGRES_DB=relayooor

# Redis
REDIS_URL=redis://redis:6379

# Chain RPC (no auth for dev)
COSMOS_RPC_ENDPOINT=https://rpc.cosmos.network:443
OSMOSIS_RPC_ENDPOINT=https://rpc.osmosis.zone:443
```

#### Production (.env.production)
```env
# API Configuration
API_URL=https://api.relayooor.com
JWT_SECRET=<secure-random-string>

# Database
POSTGRES_USER=relayooor_prod
POSTGRES_PASSWORD=<secure-password>
POSTGRES_DB=relayooor_production
DATABASE_SSL_MODE=require

# Redis
REDIS_URL=redis://:<password>@redis:6379
REDIS_TLS=true

# Chain RPC (with auth)
COSMOS_RPC_ENDPOINT=https://user:pass@private-rpc.cosmos.network:443
OSMOSIS_RPC_ENDPOINT=https://user:pass@private-rpc.osmosis.zone:443

# Security
CORS_ORIGINS=https://app.relayooor.com
SECURE_COOKIES=true
```

## Deployment Procedures

### 1. Local Development Deployment

```bash
# Quick start
make dev-all

# Or step by step
make dev-backend  # Start backend services
make dev-frontend # Start frontend dev server

# View logs
docker-compose logs -f api-backend chainpulse

# Reset everything
make clean
make dev-all
```

### 2. Staging Deployment

```bash
# Build and push images
docker build -t relayooor/webapp:staging .
docker push relayooor/webapp:staging

# Deploy to staging
ssh staging-server
cd /opt/relayooor
docker-compose pull
docker-compose up -d

# Run migrations
docker-compose exec api-backend go run . migrate up
```

### 3. Production Deployment

#### Using Docker Swarm
```bash
# Initialize swarm (if not already)
docker swarm init

# Deploy stack
docker stack deploy -c docker-compose.prod.yml relayooor

# Scale services
docker service scale relayooor_api=3

# Update service
docker service update --image relayooor/api:v1.2.0 relayooor_api
```

#### Using Kubernetes
```yaml
# webapp-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp
spec:
  replicas: 3
  selector:
    matchLabels:
      app: webapp
  template:
    metadata:
      labels:
        app: webapp
    spec:
      containers:
      - name: webapp
        image: relayooor/webapp:latest
        ports:
        - containerPort: 80
        env:
        - name: API_URL
          value: "http://api-service:8080"
---
apiVersion: v1
kind: Service
metadata:
  name: webapp-service
spec:
  selector:
    app: webapp
  ports:
  - port: 80
    targetPort: 80
  type: LoadBalancer
```

### 4. Cloud Platform Deployments

#### AWS ECS
```json
{
  "family": "relayooor-webapp",
  "taskRoleArn": "arn:aws:iam::123456789:role/ecsTaskRole",
  "executionRoleArn": "arn:aws:iam::123456789:role/ecsExecutionRole",
  "networkMode": "awsvpc",
  "containerDefinitions": [
    {
      "name": "webapp",
      "image": "relayooor/webapp:latest",
      "memory": 512,
      "cpu": 256,
      "essential": true,
      "portMappings": [
        {
          "containerPort": 80,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "API_URL",
          "value": "https://api.relayooor.com"
        }
      ]
    }
  ]
}
```

#### Google Cloud Run
```bash
# Build and push to GCR
gcloud builds submit --tag gcr.io/PROJECT_ID/relayooor-webapp

# Deploy
gcloud run deploy relayooor-webapp \
  --image gcr.io/PROJECT_ID/relayooor-webapp \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars API_URL=https://api.relayooor.com
```

#### Fly.io
```toml
# fly.toml
app = "relayooor"
primary_region = "dfw"

[build]
  dockerfile = "webapp/Dockerfile"

[env]
  API_URL = "https://api.relayooor.com"

[http_service]
  internal_port = 80
  force_https = true
  auto_stop_machines = true
  auto_start_machines = true

[[services]]
  protocol = "tcp"
  internal_port = 80
  
  [[services.ports]]
    port = 80
    handlers = ["http"]
    
  [[services.ports]]
    port = 443
    handlers = ["tls", "http"]
```

## Database Management

### 1. Initial Setup
```bash
# Create database
psql -U postgres -c "CREATE DATABASE relayooor;"
psql -U postgres -c "CREATE USER relayooor_user WITH PASSWORD 'password';"
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE relayooor TO relayooor_user;"

# Run migrations
docker-compose exec api-backend go run . migrate up
```

### 2. Migration Management
```bash
# Create new migration
docker-compose exec api-backend go run . migrate create add_user_preferences

# Run pending migrations
docker-compose exec api-backend go run . migrate up

# Rollback last migration
docker-compose exec api-backend go run . migrate down 1

# Check migration status
docker-compose exec api-backend go run . migrate status
```

### 3. Backup and Restore
```bash
# Backup
docker-compose exec postgres pg_dump -U relayooor_user relayooor > backup.sql

# Restore
docker-compose exec -T postgres psql -U relayooor_user relayooor < backup.sql

# Automated backup script
#!/bin/bash
BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)
docker-compose exec postgres pg_dump -U relayooor_user relayooor | \
  gzip > "${BACKUP_DIR}/relayooor_${DATE}.sql.gz"
```

## Monitoring Setup

### 1. Prometheus Configuration
```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'api'
    static_configs:
      - targets: ['api-backend:9090']
      
  - job_name: 'chainpulse'
    static_configs:
      - targets: ['chainpulse:3001']
      
  - job_name: 'hermes'
    static_configs:
      - targets: ['hermes:3001']
```

### 2. Grafana Dashboards
Import dashboards from `/monitoring/grafana/dashboards/`:
- `api-metrics.json` - API performance metrics
- `chainpulse-ibc.json` - IBC packet monitoring
- `hermes-relayer.json` - Relayer performance

### 3. Alerting Rules
```yaml
# alerts.yml
groups:
  - name: relayooor
    rules:
      - alert: HighErrorRate
        expr: rate(api_errors_total[5m]) > 0.05
        for: 5m
        annotations:
          summary: "High error rate detected"
          
      - alert: StuckPacketsIncreasing
        expr: increase(stuck_packets_total[1h]) > 100
        for: 15m
        annotations:
          summary: "Stuck packets increasing rapidly"
```

## SSL/TLS Configuration

### 1. Let's Encrypt with Nginx
```nginx
server {
    listen 80;
    server_name app.relayooor.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name app.relayooor.com;
    
    ssl_certificate /etc/letsencrypt/live/app.relayooor.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/app.relayooor.com/privkey.pem;
    
    location / {
        proxy_pass http://webapp:80;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 2. Automated Certificate Renewal
```bash
# Certbot renewal cron
0 0 * * * certbot renew --quiet --post-hook "docker-compose restart nginx"
```

## Health Checks and Monitoring

### 1. Service Health Endpoints
- API: `GET http://localhost:8080/health`
- Chainpulse: `GET http://localhost:3000/health`
- Hermes: `GET http://localhost:5185/health`

### 2. Docker Health Checks
```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 40s
```

### 3. Monitoring Script
```bash
#!/bin/bash
# check-services.sh

check_service() {
    local name=$1
    local url=$2
    
    if curl -f -s "$url" > /dev/null; then
        echo "✓ $name is healthy"
    else
        echo "✗ $name is down!"
        exit 1
    fi
}

check_service "API" "http://localhost:8080/health"
check_service "Chainpulse" "http://localhost:3000/health"
check_service "Hermes" "http://localhost:5185/health"
```

## Rollback Procedures

### 1. Docker Image Rollback
```bash
# Tag current version as backup
docker tag relayooor/webapp:latest relayooor/webapp:backup

# Deploy new version
docker-compose pull
docker-compose up -d

# If issues, rollback
docker tag relayooor/webapp:backup relayooor/webapp:latest
docker-compose up -d
```

### 2. Database Rollback
```bash
# Before deployment, create backup
pg_dump -U relayooor_user relayooor > pre-deploy-backup.sql

# If needed, restore
psql -U relayooor_user relayooor < pre-deploy-backup.sql
```

## Performance Optimization

### 1. Frontend Optimization
```bash
# Build with production optimizations
cd webapp
yarn build --mode production

# Analyze bundle size
yarn build --report
```

### 2. API Optimization
```go
// Enable pprof profiling
import _ "net/http/pprof"

// In main()
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

### 3. Database Optimization
```sql
-- Add indexes for common queries
CREATE INDEX idx_packets_status_created ON packets(status, created_at);
CREATE INDEX idx_packets_sender ON packets(sender);
CREATE INDEX idx_packets_chains ON packets(src_chain_id, dst_chain_id);

-- Analyze query performance
EXPLAIN ANALYZE SELECT * FROM packets WHERE status = 'pending';
```

## Troubleshooting Deployment

### Common Issues

1. **Frontend shows blank page**
   - Check if API URL is correctly configured
   - Verify CORS settings
   - Check browser console for errors

2. **Services can't connect**
   - Verify Docker network configuration
   - Check service names in connection strings
   - Ensure ports are properly exposed

3. **Database connection fails**
   - Check credentials in environment variables
   - Verify PostgreSQL is running
   - Check if migrations have been run

4. **Build fails**
   - Clear Docker cache: `docker system prune -a`
   - Check disk space
   - Verify all dependencies are available

### Debug Commands
```bash
# Check running containers
docker-compose ps

# View service logs
docker-compose logs -f [service-name]

# Execute commands in container
docker-compose exec api-backend sh

# Network debugging
docker network ls
docker network inspect relayooor_default

# Resource usage
docker stats
```