# DEPLOYMENT GUIDE - Relayooor

## Quick Start (TL;DR)

```bash
# Setup and run everything
cd webapp && yarn install && yarn build && cd ..
./scripts/setup-and-launch.sh
```

Access at: http://localhost

## Prerequisites

- Docker & Docker Compose installed
- Node.js 18+ and Yarn
- Git
- 8GB+ RAM recommended
- Ports 80, 3000-3003, 5432, 6379, 8080 available

## Complete Setup Process

### 1. Clone Repository
```bash
git clone <repository-url>
cd relayooor
```

### 2. Environment Configuration
```bash
cp .env.example .env
```

Edit `.env` with your configuration:
```env
# Required
JWT_SECRET=your-secret-key-here
POSTGRES_PASSWORD=secure-database-password
REDIS_PASSWORD=secure-redis-password

# Chain RPC Endpoints (with authentication)
COSMOS_RPC_ENDPOINT=https://username:password@cosmos-rpc.example.com:443
OSMOSIS_RPC_ENDPOINT=https://username:password@osmosis-rpc.example.com:443
NEUTRON_RPC_ENDPOINT=https://username:password@neutron-rpc.example.com:443

# Optional
ADMIN_API_KEY=admin-api-key
ACTIVE_RELAYER=hermes
```

### 3. Build Frontend (CRITICAL - Must be done first!)
```bash
cd webapp
yarn install
yarn build  # This creates the production bundle
cd ..
```

### 4. Start Services

#### Option A: Using Setup Script (Recommended)
```bash
./scripts/setup-and-launch.sh
```

This script:
- Checks Docker installation
- Verifies environment setup
- Builds all containers
- Starts services with health checks
- Shows status

#### Option B: Manual Docker Compose
```bash
# Clean start
docker-compose down -v  # Remove volumes for fresh start
docker-compose build --no-cache
docker-compose up -d

# Check status
docker-compose ps
docker-compose logs -f
```

### 5. Verify Deployment
```bash
# Check all services are running
docker-compose ps

# Test endpoints
curl http://localhost/health           # Frontend
curl http://localhost:8080/health      # API
curl http://localhost:3001/metrics     # Chainpulse

# View logs
docker-compose logs -f webapp api-backend chainpulse
```

## Service Ports

**Important**: The API is exposed on port 3000 externally but runs on 8080 internally.

- Web Application: `http://localhost` (port 80)
- API Backend: `http://localhost:3000` (external) / port 8080 (internal)
- Chainpulse Metrics: `http://localhost:3001/metrics`
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3003`

## Service Management

### Start/Stop Services
```bash
# Stop all
docker-compose down

# Stop and remove volumes (full reset)
docker-compose down -v

# Start all
docker-compose up -d

# Restart specific service
docker-compose restart api-backend
```

### Quick Development Iteration

#### Frontend Changes
```bash
# After making changes to webapp
cd webapp && yarn build && cd ..
make -f Makefile.docker webapp-restart

# OR manually
docker-compose stop webapp
docker-compose build webapp
docker-compose up -d webapp
```

#### Backend Changes
```bash
# After making changes to api
make -f Makefile.docker api-restart

# OR manually
docker-compose stop api-backend
docker-compose build api-backend
docker-compose up -d api-backend
```

### View Logs
```bash
# All services
docker-compose logs -f

# Specific services
docker-compose logs -f api-backend
docker-compose logs -f chainpulse
docker-compose logs -f webapp

# With timestamps
docker-compose logs -f -t api-backend
```

## Docker Compose Variants

### Standard Development Stack
```bash
docker-compose up -d
```
Includes: webapp, api, chainpulse, postgres, redis, hermes

### Minimal Webapp Only
```bash
docker-compose -f docker-compose.webapp.yml up -d
```
Includes: webapp, api, minimal dependencies

### Full Stack with Monitoring
```bash
docker-compose -f docker-compose.full.yml up -d
```
Includes: All services plus Prometheus and Grafana

## Production Deployment (Fly.io)

### Prerequisites
```bash
# Install flyctl
curl -L https://fly.io/install.sh | sh

# Login
fly auth login
```

### Deploy
```bash
# Deploy application
fly deploy

# Check status
fly status

# View logs
fly logs

# Scale
fly scale count webapp=2
```

### Environment Variables on Fly
```bash
fly secrets set JWT_SECRET=your-secret
fly secrets set POSTGRES_PASSWORD=your-password
# ... set all required secrets
```

## Troubleshooting

### Problem: Services won't start
```bash
# Check for port conflicts
lsof -i :80
lsof -i :8080
lsof -i :3001

# Check Docker daemon
docker ps
```

### Problem: Frontend shows stale data
```bash
# Always rebuild frontend first
cd webapp && yarn build && cd ..
docker-compose restart webapp
```

### Problem: Can't connect to database
```bash
# Check postgres is running
docker-compose ps postgres
docker-compose logs postgres

# Test connection
docker-compose exec postgres psql -U relayooor_user -d relayooor
```

### Problem: Build failures
```bash
# Clean build
docker-compose down -v
docker system prune -a  # WARNING: Removes all unused images
docker-compose build --no-cache
docker-compose up -d
```

### Problem: Services can't communicate
- Use service names from docker-compose.yml (not localhost)
- Example: `api-backend:8080` not `localhost:8080`

## Health Checks

### Manual Health Checks
```bash
# Frontend
curl http://localhost/

# API
curl http://localhost:8080/health

# Chainpulse
curl http://localhost:3001/metrics | grep "ibc_"

# Database
docker-compose exec postgres pg_isready

# Redis
docker-compose exec redis redis-cli ping
```

### Automated Health Check Script
```bash
./scripts/setup-and-launch.sh status
```

## Backup and Restore

### Backup Database
```bash
# Backup
docker-compose exec postgres pg_dump -U relayooor_user relayooor > backup.sql

# Backup with timestamp
docker-compose exec postgres pg_dump -U relayooor_user relayooor > backup_$(date +%Y%m%d_%H%M%S).sql
```

### Restore Database
```bash
# Restore
docker-compose exec -T postgres psql -U relayooor_user relayooor < backup.sql
```

## Security Considerations

1. **Change default passwords** in production
2. **Use HTTPS** with proper certificates
3. **Restrict database access** to local network
4. **Rotate JWT secrets** regularly
5. **Monitor logs** for suspicious activity
6. **Keep Docker images updated**

## Performance Tuning

### Database
```yaml
# In docker-compose.yml under postgres
environment:
  POSTGRES_MAX_CONNECTIONS: 200
  POSTGRES_SHARED_BUFFERS: 256MB
```

### Redis
```yaml
# In docker-compose.yml under redis
command: redis-server --maxmemory 512mb --maxmemory-policy allkeys-lru
```

### API
- Adjust connection pool sizes
- Configure rate limiting
- Enable response caching

## Monitoring

### With Grafana (Full Stack)
```bash
docker-compose -f docker-compose.full.yml up -d
```
Access Grafana at: http://localhost:3003

### Prometheus Metrics
- Chainpulse: http://localhost:3001/metrics
- Configure Prometheus to scrape these endpoints

### Application Logs
- Use centralized logging in production
- Consider ELK stack or similar

---

Last Updated: 2025-07-16