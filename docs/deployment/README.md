# Relayooor Production Deployment Guide

Deploy the complete Relayooor stack on Fly.io with this streamlined guide.

## Architecture Overview

The application consists of three modules:
1. **Chainpulse** - IBC metrics collector (Rust)
2. **API** - Backend service (Go) 
3. **Web App** - User interface (Vue.js)

## Prerequisites

- Fly.io CLI installed (`brew install flyctl`)
- Fly.io account with payment method
- RPC node credentials for target chains

## Quick Start

### 1. Initial Setup

```bash
# Authenticate with Fly.io
fly auth login

# Clone repository
git clone https://github.com/your-org/relayooor
cd relayooor
```

### 2. Deploy Chainpulse

```bash
cd chainpulse

# Create persistent volume
fly volumes create chainpulse_data --size 10 --region iad

# Set RPC credentials
fly secrets set RPC_USERNAME=your-username RPC_PASSWORD=your-password

# Deploy
fly deploy

# Verify it's running
fly status
```

See [chainpulse-module.md](./chainpulse-module.md) for configuration details.

### 3. Deploy API Service

```bash
cd ../api

# Set JWT secret
fly secrets set JWT_SECRET=$(openssl rand -base64 32)

# Deploy with autoscaling
fly deploy
fly autoscale set min=1 max=10

# Check health
curl https://relayooor-api.fly.dev/health
```

See [api-module.md](./api-module.md) for API endpoints and scaling options.

### 4. Deploy Web Application

```bash
cd ../webapp

# Build and deploy
fly deploy

# Enable autoscaling
fly autoscale set min=1 max=5

# Access at
open https://relayooor-webapp.fly.dev
```

See [webapp-module.md](./webapp-module.md) for nginx configuration.

## Network Configuration

All services communicate via Fly.io's private network:

```
Internet → relayooor-webapp.fly.dev
               ↓
       relayooor-api.internal:8080
               ↓
    relayooor-chainpulse.internal:3001
```

## Essential Commands

### Monitor All Services
```bash
# View logs across all apps
fly logs -a relayooor-chainpulse
fly logs -a relayooor-api  
fly logs -a relayooor-webapp

# Check metrics endpoint
curl https://relayooor-api.fly.dev/api/metrics/chainpulse
```

### Scale Services
```bash
# Scale API based on load
fly scale count 3 -a relayooor-api

# Add regions for global distribution
fly regions add sin -a relayooor-webapp
```

### Update Configuration
```bash
# Update secrets
fly secrets set NEW_SECRET=value -a relayooor-api

# Redeploy after changes
fly deploy --strategy rolling -a relayooor-api
```

## Health Monitoring

Each service exposes health endpoints:
- Chainpulse: `http://relayooor-chainpulse.internal:3001/metrics`
- API: `https://relayooor-api.fly.dev/health`
- Web App: `https://relayooor-webapp.fly.dev/`

## Troubleshooting

### Service Communication Issues
```bash
# Test internal DNS
fly ssh console -a relayooor-api
> nslookup relayooor-chainpulse.internal
```

### Database Issues
```bash
# Check Chainpulse volume
fly volumes list -a relayooor-chainpulse

# Extend volume if needed
fly volumes extend <volume-id> --size 20
```

### Performance Problems
```bash
# View resource usage
fly scale show -a relayooor-api

# Check response times
fly metrics -a relayooor-webapp
```

## Maintenance

### Backups
```bash
# Backup Chainpulse database
fly ssh console -a relayooor-chainpulse
> sqlite3 /data/chainpulse.db ".backup /data/backup.db"
```

### Updates
```bash
# Deploy new version with zero downtime
fly deploy --strategy canary -a relayooor-api
```

### Logs
```bash
# Export logs for analysis
fly logs -a relayooor-chainpulse --json > chainpulse.log
```

## Cost Optimization

- Chainpulse: Runs 24/7 (required for WebSocket connections)
- API: Auto-scales based on traffic (can scale to zero)
- Web App: Serves static files (minimal resources)

Estimated monthly cost: $15-50 depending on traffic

## Next Steps

1. Set up monitoring alerts
2. Configure custom domains
3. Enable CDN for web app
4. Implement backup automation

For detailed module configuration, refer to:
- [Chainpulse Module](./chainpulse-module.md)
- [API Module](./api-module.md)  
- [Web App Module](./webapp-module.md)