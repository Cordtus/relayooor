# API Module Deployment

## Overview

The Go API serves as the backend, proxying metrics and handling user requests. It can scale horizontally based on load.

## Requirements

- 256MB RAM per instance
- Can scale to zero
- Stateless (no persistent storage needed)

## Configuration

### Required Environment Variables

```bash
PORT              # HTTP port (default: 8080)
CHAINPULSE_URL    # Internal URL to Chainpulse service
DATABASE_URL      # PostgreSQL connection string (optional)
JWT_SECRET        # Secret for JWT tokens
```

### Optional Environment Variables

```bash
CORS_ORIGINS      # Allowed origins (default: *)
LOG_LEVEL         # debug, info, warn, error (default: info)
```

## Fly.io Deployment

### fly.toml Configuration

```toml
app = "relayooor-api"
primary_region = "iad"

[build]
  builder = "paketobuildpacks/builder:base"
  buildpacks = ["gcr.io/paketo-buildpacks/go"]

[env]
  PORT = "8080"
  CHAINPULSE_URL = "http://relayooor-chainpulse.internal:3001"

[[services]]
  internal_port = 8080
  protocol = "tcp"
  auto_stop_machines = true
  auto_start_machines = true
  min_machines_running = 1

  [services.concurrency]
    type = "requests"
    hard_limit = 250
    soft_limit = 200

  [[services.ports]]
    port = 80
    handlers = ["http"]
    
  [[services.ports]]
    port = 443
    handlers = ["tls", "http"]

  [[services.http_checks]]
    interval = "15s"
    grace_period = "5s"
    method = "get"
    path = "/health"
    protocol = "http"
    timeout = "2s"
```

### Deployment Commands

```bash
# Set secrets
fly secrets set JWT_SECRET=$(openssl rand -base64 32)

# Deploy
fly deploy

# Scale based on load
fly autoscale set min=1 max=10

# Monitor
fly status
```

## API Endpoints

### Health Check
```
GET /health
Returns: {"status": "ok"}
```

### Metrics Proxy
```
GET /api/metrics/chainpulse
Returns: Prometheus metrics (text format)
```

### Monitoring Data
```
GET /api/monitoring/data
Returns: Structured JSON data
```

## Performance Tuning

### Connection Pooling
The API maintains a connection pool to Chainpulse. Default settings:
- Max idle connections: 10
- Max open connections: 100
- Connection timeout: 5s

### Caching
Metrics responses are cached for 5 seconds to reduce load on Chainpulse.

## Troubleshooting

### Connection refused to Chainpulse
- Verify internal DNS: `relayooor-chainpulse.internal`
- Check both services are in same Fly organization
- Ensure Chainpulse is running

### High response times
- Check Chainpulse health
- Increase instance count
- Review slow query logs

### Memory issues
- Normal usage: 50-100MB
- If exceeding 200MB, check for memory leaks
- Restart instances if needed