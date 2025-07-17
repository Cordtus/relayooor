# Hermes Integration Guide

This document describes how Hermes IBC Relayer is integrated into the Relayooor platform.

## Overview

Hermes is configured as a packet clearing service that only operates on-demand through authenticated API requests. It does not actively relay packets unless explicitly instructed.

## Configuration

### Operational Mode

Hermes is configured with the following modes:

```toml
[mode.clients]
enabled = false  # Disabled - no automatic client updates

[mode.connections]
enabled = false  # Disabled - no connection handshakes

[mode.channels]
enabled = false  # Disabled - no channel handshakes

[mode.packets]
enabled = true
clear_interval = 0      # No automatic clearing
clear_on_start = false  # No clearing on startup
tx_confirmation = true  # Wait for transaction confirmation
```

### Supported Chains

- **Cosmos Hub** (cosmoshub-4)
- **Osmosis** (osmosis-1)
- **Noble** (noble-1)
- **Stride** (stride-1)

Note: Neutron is temporarily excluded due to Slinky oracle compatibility issues.

### Authentication

- **RPC Endpoints**: Require authentication (username/password in URL)
- **WebSocket Endpoints**: Require authentication
- **gRPC Endpoints**: Do not require authentication

Authentication is automatically injected by the entrypoint script using environment variables:
- `RPC_USERNAME`
- `RPC_PASSWORD`

## API Integration

### REST API Endpoints

Hermes exposes a REST API on port 5185:

- `GET /version` - Get Hermes version
- `GET /state` - Get current state
- `GET /chains` - List configured chains
- `GET /chain/{chain_id}` - Get chain details
- `POST /clear_packets` - Clear specific packets

### Telemetry Metrics

Prometheus metrics are exposed on port 3001:

- `hermes_pending_packets` - Number of pending packets per channel
- `hermes_wallet_balance` - Relayer wallet balances
- `hermes_tx_latency_confirmed` - Transaction confirmation latency
- Various error and performance metrics

## Usage in Relayooor

### 1. On-Demand Packet Clearing

When a user pays for packet clearing:

```javascript
// API call to clear packets
POST /api/v1/relayer/hermes/clear
{
  "chain": "cosmoshub-4",
  "channel": "channel-0",
  "port": "transfer",
  "sequences": [1, 2, 3]
}
```

### 2. Health Monitoring

The system continuously monitors Hermes health:

```javascript
// Health check endpoint
GET /api/v1/relayer/hermes/health

// Response
{
  "healthy": true,
  "service": "hermes",
  "url": "http://hermes:5185"
}
```

### 3. Metrics Integration

Frontend dashboards consume Hermes metrics for:
- Channel congestion visualization
- Wallet balance monitoring
- Performance analytics
- Error tracking

## Docker Integration

Hermes runs as a Docker service with:
- Custom entrypoint for authentication
- Persistent volume for data
- Health checks
- Automatic restart on failure

```yaml
hermes:
  image: ghcr.io/informalsystems/hermes:1.10.5
  environment:
    - RUST_LOG=info
    - RPC_USERNAME=${RPC_USERNAME}
    - RPC_PASSWORD=${RPC_PASSWORD}
  volumes:
    - ./config/hermes:/config:ro
    - hermes-data:/data
  ports:
    - "3010:3001"  # Telemetry
    - "5185:5185"  # REST API
```

## Security Considerations

1. **Access Control**: Hermes API is not publicly exposed
2. **Authentication**: All RPC calls use authenticated endpoints
3. **Rate Limiting**: Implemented at API gateway level
4. **Audit Logging**: All clearing operations are logged

## Troubleshooting

### Common Issues

1. **Hermes not starting**
   - Check logs: `docker logs relayooor-hermes-1`
   - Verify config syntax: `hermes config validate`
   - Ensure RPC endpoints are accessible

2. **Authentication failures**
   - Verify RPC_USERNAME and RPC_PASSWORD are set
   - Check URL encoding of special characters
   - Test RPC endpoints manually

3. **Packet clearing failures**
   - Check wallet balances
   - Verify chain is not halted
   - Review gas price configuration

### Useful Commands

```bash
# Check Hermes version
curl http://localhost:5185/version

# View pending packets
curl http://localhost:3001/metrics | grep pending

# Monitor logs
docker logs -f relayooor-hermes-1

# Restart Hermes
docker-compose restart hermes
```

## Future Enhancements

1. **Multi-relayer Support**: Add support for Go relayer as fallback
2. **Dynamic Configuration**: Update chains without restart
3. **Advanced Routing**: Intelligent path selection for packets
4. **Fee Optimization**: Dynamic gas price adjustment