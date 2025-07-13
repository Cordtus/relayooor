# Chainpulse Module Deployment

## Overview

Chainpulse monitors blockchain nodes and collects IBC packet metrics. It requires constant WebSocket connections to chain nodes.

## Requirements

- Persistent storage for SQLite database (1GB minimum)
- Stable network connection
- 512MB RAM minimum
- Always-on process (cannot scale to zero)

## Configuration

### Required Environment Variables

```bash
RPC_USERNAME    # Authentication for RPC nodes
RPC_PASSWORD    # Authentication password
DATABASE_PATH   # Path to SQLite file (default: /data/chainpulse.db)
```

### Configuration File

Create `chainpulse.toml`:

```toml
[database]
path = "/data/chainpulse.db"

[metrics]
enabled = true
port = 3001
stuck_packets = true

[chains.cosmoshub-4]
url = "wss://rpc-cosmoshub.example.com/websocket"
comet_version = "0.37"
username = "${RPC_USERNAME}"
password = "${RPC_PASSWORD}"

[chains.osmosis-1]
url = "wss://rpc-osmosis.example.com/websocket"
comet_version = "0.37"
username = "${RPC_USERNAME}"
password = "${RPC_PASSWORD}"
```

## Fly.io Deployment

### fly.toml Configuration

```toml
app = "relayooor-chainpulse"
primary_region = "iad"

[build]
  image = "relayooor/chainpulse:latest"

[env]
  PORT = "3001"

[mounts]
  source = "chainpulse_data"
  destination = "/data"

[[services]]
  internal_port = 3001
  protocol = "tcp"

  [[services.ports]]
    port = 3001

[checks]
  [checks.metrics]
    grace_period = "30s"
    interval = "15s"
    method = "get"
    path = "/metrics"
    port = 3001
    timeout = "10s"
```

### Deployment Commands

```bash
# Create volume for persistent storage
fly volumes create chainpulse_data --size 10 --region iad

# Set secrets
fly secrets set RPC_USERNAME=your-username RPC_PASSWORD=your-password

# Deploy
fly deploy

# Check logs
fly logs
```

## Health Monitoring

Monitor these key metrics:

1. **Connection Status**: Check `chainpulse_reconnects` metric
2. **Error Rate**: Monitor `chainpulse_errors` metric
3. **Data Flow**: Verify `chainpulse_packets` increases

## Troubleshooting

### No data appearing
- Check WebSocket URLs are accessible
- Verify RPC credentials
- Check logs for connection errors

### High memory usage
- Normal for long-running connections
- Restart if memory exceeds 1GB

### Database growth
- SQLite file grows with historical data
- Implement cleanup job if exceeds 5GB