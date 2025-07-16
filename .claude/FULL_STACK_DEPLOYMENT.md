# Full Stack Deployment Guide

This guide covers the deployment of the complete Relayooor application stack using Docker Compose.

## Architecture Overview

The full stack consists of:
- **Chainpulse**: IBC monitoring service (forked from https://github.com/Cordtus/chainpulse.git)
- **API Backend**: Go service providing REST API and WebSocket endpoints
- **Web Application**: Vue.js 3 frontend with packet clearing interface
- **PostgreSQL**: Database for storing clearing transactions and user data
- **Prometheus**: Metrics collection and storage
- **Grafana**: Metrics visualization dashboard

## Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+
- At least 8GB RAM available
- 10GB free disk space

## Quick Start

1. Clone the repository:
```bash
git clone https://github.com/cordtus/relayooor.git
cd relayooor
```

2. Run the full stack:
```bash
./start-full-stack.sh
```

## Manual Deployment

### Build all services:
```bash
docker-compose -f docker-compose.full.yml build
```

### Start all services:
```bash
docker-compose -f docker-compose.full.yml up -d
```

### Check service status:
```bash
docker-compose -f docker-compose.full.yml ps
```

### View logs:
```bash
# All services
docker-compose -f docker-compose.full.yml logs -f

# Specific service
docker-compose -f docker-compose.full.yml logs -f chainpulse
```

## Service Endpoints

| Service | URL | Description |
|---------|-----|-------------|
| Web Application | http://localhost | Main user interface |
| API Backend | http://localhost:8080 | REST API endpoints |
| Chainpulse API | http://localhost:3000 | IBC monitoring API |
| Chainpulse Metrics | http://localhost:3001/metrics | Prometheus metrics |
| Prometheus | http://localhost:9090 | Metrics query interface |
| Grafana | http://localhost:3003 | Dashboards (admin/admin) |
| PostgreSQL | localhost:5432 | Database (postgres/postgres) |

## Configuration

### Chainpulse Configuration
Edit `config/chainpulse.toml` to configure monitored chains and endpoints.

### API Backend Environment
Configure in `docker-compose.full.yml`:
```yaml
environment:
  - CHAINPULSE_URL=http://chainpulse:3000
  - DATABASE_URL=postgres://postgres:postgres@postgres:5432/relayooor
```

### Web Application Environment
Create `webapp/.env.production`:
```
VITE_API_URL=http://localhost:8080
VITE_CHAINPULSE_URL=http://localhost:3000
```

## Monitoring

### Prometheus Queries
Access Prometheus at http://localhost:9090 and try these queries:
- `ibc_channels_total` - Total number of IBC channels
- `ibc_stuck_packets_total` - Number of stuck packets
- `ibc_relayer_balance` - Relayer wallet balances

### Grafana Dashboards
1. Access Grafana at http://localhost:3003
2. Login with admin/admin
3. Import dashboard from `config/grafana-poc/dashboards/`

## Troubleshooting

### Service won't start
```bash
# Check logs
docker-compose -f docker-compose.full.yml logs [service-name]

# Restart service
docker-compose -f docker-compose.full.yml restart [service-name]
```

### Port conflicts
If ports are already in use, modify the port mappings in `docker-compose.full.yml`.

### Database issues
```bash
# Connect to database
docker-compose -f docker-compose.full.yml exec postgres psql -U postgres -d relayooor

# Reset database
docker-compose -f docker-compose.full.yml down -v
docker-compose -f docker-compose.full.yml up -d postgres
```

### Chainpulse not connecting
1. Check RPC endpoints in `config/chainpulse.toml`
2. Ensure chains are accessible from Docker network
3. Check Chainpulse logs: `docker-compose -f docker-compose.full.yml logs chainpulse`

## Development

### Rebuilding after changes
```bash
# Rebuild specific service
docker-compose -f docker-compose.full.yml build [service-name]

# Rebuild and restart
docker-compose -f docker-compose.full.yml up -d --build [service-name]
```

### Running locally (outside Docker)
See individual README files:
- API Backend: `api/README.md`
- Web Application: `webapp/README.md`
- Chainpulse: `monitoring/chainpulse/README.md`

## Production Deployment

For production deployments:
1. Use environment-specific configurations
2. Enable SSL/TLS termination
3. Configure proper database credentials
4. Set up monitoring alerts
5. Enable log aggregation
6. Configure backup strategies

## Cleanup

To stop and remove all containers, networks, and volumes:
```bash
docker-compose -f docker-compose.full.yml down -v
```