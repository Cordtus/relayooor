# IBC Monitoring

Chainpulse-based monitoring system for IBC networks with Prometheus and Grafana integration.

## Features

- **Real-time IBC Monitoring**: Track packet flow, stuck packets, and channel health
- **Multi-chain Support**: Monitor multiple IBC-enabled chains simultaneously
- **Authenticated RPC**: Support for username/password protected RPC endpoints
- **Prometheus Integration**: Export metrics in Prometheus format
- **Grafana Dashboards**: Pre-configured dashboards for visualization
- **Extensible**: Easy to add new chains and custom metrics

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ Chainpulse  │────>│ Prometheus  │────>│   Grafana   │
│  (Metrics)  │     │ (Storage)   │     │ (Visualize) │
└─────────────┘     └─────────────┘     └─────────────┘
       │
       v
┌─────────────┐
│IBC Networks │
└─────────────┘
```

## Quick Start

1. Configure your chains in `config/chainpulse.toml`

2. Set environment variables:
```bash
# Optional RPC authentication
export RPC_USERNAME=your_username
export RPC_PASSWORD=your_password

# Optional chain-specific WebSocket URLs
export CHAIN_OSMOSIS_WS=wss://your-osmosis-rpc/websocket
export CHAIN_COSMOSHUB_WS=wss://your-cosmos-rpc/websocket
```

3. Start the monitoring stack:
```bash
docker-compose up -d
```

4. Access the services:
   - Grafana: http://localhost:3000 (admin/admin)
   - Prometheus: http://localhost:9090
   - Chainpulse metrics: http://localhost:3001/metrics

## Configuration

### Adding New Chains

Edit `config/chainpulse.toml`:
```toml
[chains.your-chain-id]
url = "${CHAIN_YOUR_WS}"
comet_version = "0.37"
```

Then set the environment variable:
```bash
export CHAIN_YOUR_WS=wss://your-chain-rpc/websocket
```

### Custom Dashboards

Add Grafana dashboard JSON files to `config/grafana/provisioning/dashboards/`

### Prometheus Scrape Targets

Edit `config/prometheus.yml` to add new metrics sources.

## Available Metrics

Chainpulse provides various IBC metrics including:
- `ibc_stuck_packets`: Number of stuck packets per channel
- `ibc_effected_packets`: Successfully relayed packets
- `ibc_uneffected_packets`: Failed relay attempts
- `ibc_frontrun_counter`: Frontrunning events by relayer
- `ibc_handshake_states`: Channel handshake status

## Environment Variables

- `RPC_USERNAME`: Username for RPC authentication
- `RPC_PASSWORD`: Password for RPC authentication
- `CHAIN_*_WS`: WebSocket URLs for specific chains
- `GF_ADMIN_USER`: Grafana admin username (default: admin)
- `GF_ADMIN_PASSWORD`: Grafana admin password (default: admin)

## Building from Source

```bash
docker build -t ibc-monitoring .
```

## Troubleshooting

### Connection Issues
- Verify WebSocket URLs are accessible
- Check if RPC authentication is required
- Ensure firewall rules allow WebSocket connections

### Missing Metrics
- Check chainpulse logs: `docker-compose logs chainpulse`
- Verify chains are producing blocks
- Ensure IBC activity exists on monitored channels

### Performance
- Adjust Prometheus retention if disk usage is high
- Reduce scrape frequency for less critical metrics
- Consider running on dedicated monitoring infrastructure