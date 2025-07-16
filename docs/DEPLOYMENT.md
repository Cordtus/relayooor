# Deployment Guide

This guide will help you deploy the Relayooor IBC Packet Clearing Platform.

## Prerequisites

- Docker and Docker Compose
- Node.js 18+ and Yarn
- Git
- Minimum 8GB RAM
- Available ports: 80, 8080, 3001, 5432, 6379

## Quick Start

```bash
# 1. Clone the repository
git clone <repository-url>
cd relayooor

# 2. Setup environment
cp .env.example .env
# Edit .env with your RPC endpoints

# 3. Build and start
cd webapp && yarn install && yarn build && cd ..
./scripts/setup-and-launch.sh
```

The application will be available at http://localhost

## Configuration

### Environment Variables

Create a `.env` file with:

```env
# Required
JWT_SECRET=your-secret-key
POSTGRES_PASSWORD=secure-password
REDIS_PASSWORD=redis-password

# Chain RPC Endpoints
COSMOS_RPC_ENDPOINT=https://your-cosmos-rpc:443
OSMOSIS_RPC_ENDPOINT=https://your-osmosis-rpc:443
NEUTRON_RPC_ENDPOINT=https://your-neutron-rpc:443
```

## Service Management

### Start Services
```bash
docker-compose up -d
```

### Stop Services
```bash
docker-compose down
```

### View Logs
```bash
docker-compose logs -f
```

### Restart a Service
```bash
docker-compose restart api-backend
```

## Updating the Application

### Frontend Updates
```bash
cd webapp
yarn build
cd ..
docker-compose restart webapp
```

### Backend Updates
```bash
docker-compose restart api-backend
```

## Troubleshooting

### Services Won't Start
- Check if ports are already in use
- Ensure Docker is running
- Check logs: `docker-compose logs`

### Frontend Shows Old Data
- Rebuild frontend: `cd webapp && yarn build`
- Clear browser cache

### Database Connection Issues
- Verify PostgreSQL is running: `docker-compose ps postgres`
- Check credentials in `.env`

## Production Deployment

For production deployments, consider:
- Using HTTPS with SSL certificates
- Setting strong passwords
- Configuring firewall rules
- Setting up monitoring
- Regular backups

## Support

For issues or questions, please check the documentation or open an issue on GitHub.