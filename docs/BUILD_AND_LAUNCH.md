# Build and Launch Guide

This guide documents the standard process for building and launching the Relayooor application.

## Prerequisites

1. **Docker Desktop** must be installed and running
2. **RPC Credentials** for Skip endpoints (username: `skip`, password: contact Skip for access)
3. **Git** for version control
4. **Minimum 4GB RAM** available for Docker

## Quick Start

The easiest way to build and launch Relayooor is using the setup script:

```bash
./setup-and-launch.sh
```

This will:
1. Check Docker is running
2. Create/verify `.env` configuration
3. Clean up any existing containers
4. Build all services
5. Start all services
6. Wait for services to be healthy
7. Display service URLs

## Manual Build Process

If you prefer to build manually:

### 1. Environment Setup

Create a `.env` file with your RPC credentials:

```bash
# RPC Authentication
RPC_USERNAME=skip
RPC_PASSWORD=your_password_here

# Optional WebSocket URLs (defaults shown)
COSMOS_WS_URL=wss://cosmos-rpc.polkachu.com/websocket
OSMOSIS_WS_URL=wss://osmosis-rpc.polkachu.com/websocket
NEUTRON_WS_URL=wss://neutron-rpc.polkachu.com/websocket
NOBLE_WS_URL=wss://noble-rpc.polkachu.com/websocket
```

### 2. Build Services

```bash
# Build all services
docker-compose build

# Or build specific services
docker-compose build api-backend
docker-compose build chainpulse
docker-compose build webapp
```

### 3. Start Services

```bash
# Start all services in background
docker-compose up -d

# Or start with logs visible
docker-compose up
```

### 4. Verify Services

Check each service is running:

```bash
# API health check
curl http://localhost:3000/health

# Chainpulse metrics
curl http://localhost:3001/metrics | head

# Web application
curl -I http://localhost:80
```

## Service Architecture

The application consists of the following services:

### Core Services

1. **api-backend** (port 3000)
   - REST API for the web application
   - Connects to Chainpulse for metrics
   - Handles clearing operations (when implemented)

2. **chainpulse** (port 3001)
   - Monitors IBC activity on configured chains
   - Provides Prometheus metrics
   - WebSocket connections to chain RPCs

3. **webapp** (port 80)
   - Vue.js single-page application
   - Nginx reverse proxy for API calls
   - Real-time monitoring dashboard

### Supporting Services

4. **postgres** (port 5432)
   - Database for clearing operations
   - User: relayooor, Password: relayooor

5. **redis** (port 6379)
   - Cache and session storage
   - Pub/sub for real-time updates

6. **prometheus** (port 9090)
   - Metrics aggregation
   - Scrapes Chainpulse metrics

7. **grafana** (port 3003)
   - Metrics visualization
   - Default login: admin/admin

## Common Operations

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f api-backend

# Last 100 lines
docker-compose logs --tail=100
```

### Restart Services

```bash
# All services
docker-compose restart

# Specific service
docker-compose restart webapp
```

### Stop Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (clean slate)
docker-compose down -v
```

### Update Code and Rebuild

```bash
# Pull latest changes
git pull

# Rebuild changed services
docker-compose up -d --build
```

## Troubleshooting

### Service Won't Start

1. Check logs: `docker-compose logs [service-name]`
2. Verify port availability: `lsof -i :3000` (replace with service port)
3. Check Docker resources: Ensure Docker has enough memory

### Chainpulse Connection Issues

1. Verify RPC credentials in `.env`
2. Check WebSocket URLs are accessible
3. Look for auth errors in logs: `docker-compose logs chainpulse`

### Webapp Shows No Data

1. Verify API is running: `curl http://localhost:3000/health`
2. Check Chainpulse metrics: `curl http://localhost:3001/metrics`
3. Clear browser cache and reload

### Database Connection Issues

1. Ensure postgres container is running
2. Check credentials match in docker-compose.yml
3. Verify no other service is using port 5432

## Development Workflow

### Making Changes

1. **Frontend Changes** (webapp/):
   ```bash
   # Rebuild just the webapp
   docker-compose up -d --build webapp
   ```

2. **API Changes** (api/):
   ```bash
   # Rebuild API
   docker-compose up -d --build api-backend
   ```

3. **Configuration Changes**:
   - Update relevant config files
   - Restart affected services

### Local Development

For faster development iteration:

```bash
# Frontend development with hot reload
cd webapp
yarn install
yarn dev

# API development
cd api
go run cmd/server/main.go
```

## Production Considerations

1. **Security**:
   - Change default passwords
   - Use proper RPC credentials
   - Enable HTTPS

2. **Performance**:
   - Increase Docker memory limits
   - Configure proper database indexes
   - Enable Redis persistence

3. **Monitoring**:
   - Set up Grafana alerts
   - Configure log aggregation
   - Monitor disk usage

## Script Commands

The `setup-and-launch.sh` script supports:

- `./setup-and-launch.sh` - Full setup and launch
- `./setup-and-launch.sh stop` - Stop all services
- `./setup-and-launch.sh restart` - Restart all services
- `./setup-and-launch.sh logs` - Follow logs
- `./setup-and-launch.sh status` - Show service status
- `./setup-and-launch.sh clean` - Remove all data