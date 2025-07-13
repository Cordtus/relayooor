# Monitoring Dashboard Setup

This document explains how to set up and use the IBC monitoring dashboard with Chainpulse.

## Overview

The monitoring dashboard provides real-time visibility into IBC packet flows, relay performance, and stuck packets. It integrates with Chainpulse to collect and display metrics.

## Architecture

```
┌─────────────────┐     ┌──────────────┐     ┌─────────────┐
│  Web Dashboard  │────▶│  API Backend │────▶│ Chainpulse  │
│   (React/TS)    │     │  (Go/Gin)    │     │   (Rust)    │
└─────────────────┘     └──────────────┘     └─────────────┘
                                                     │
                                                     ▼
                                              ┌─────────────┐
                                              │ Blockchain  │
                                              │    Nodes    │
                                              └─────────────┘
```

## Running Chainpulse

### Option 1: External Chainpulse Instance

1. Navigate to your Chainpulse directory:
   ```bash
   cd ../chainpulse
   ```

2. Create a configuration file:
   ```toml
   # chainpulse.toml
   [database]
   path = "./chainpulse.db"

   [metrics]
   enabled = true
   port = 3001
   stuck_packets = true
   populate_on_start = false

   [chains.cosmoshub-4]
   url = "wss://rpc.cosmos.network/websocket"
   comet_version = "0.37"

   [chains.osmosis-1]
   url = "wss://rpc.osmosis.zone/websocket"
   comet_version = "0.37"
   ```

3. Run Chainpulse:
   ```bash
   cargo run -- --config chainpulse.toml
   ```

4. Verify metrics are available:
   ```bash
   curl http://localhost:3001/metrics
   ```

### Option 2: Docker Compose Stack

The project includes Chainpulse in the Docker Compose setup:

```bash
make start  # Starts all services including Chainpulse
```

## Dashboard Features

### 1. System Overview
- Active chains and connection status
- Total packet volume (24h, 7d, 30d)
- Overall success rates

### 2. Packet Flow Visualization
- Real-time packet flow chart
- Effected vs uneffected packets
- Customizable time intervals

### 3. Channel Performance
- Channel-by-channel metrics
- Success rates per channel
- Packet volume and trends

### 4. Relayer Leaderboard
- Top performing relayers
- Success rates and packet counts
- Frontrunning statistics

### 5. Stuck Packet Alerts
- Real-time stuck packet detection
- Duration and value estimates
- Quick access to clearing functions

## Configuration

### Frontend Configuration

Edit `webapp/src/config/monitoring.ts`:

```typescript
export const MONITORING_CONFIG = {
  // Point to your Chainpulse instance
  chainpulseUrl: 'http://localhost:3001/metrics',
  
  // Refresh intervals
  refreshIntervals: {
    metrics: 5000,      // 5 seconds
    channels: 10000,    // 10 seconds
    stuckPackets: 30000 // 30 seconds
  }
};
```

### API Configuration

The API backend proxies Chainpulse metrics to handle CORS and authentication:

```go
// api/cmd/server/main.go
api.HandleFunc("/metrics/chainpulse", func(w http.ResponseWriter, r *http.Request) {
    // Proxy to actual Chainpulse instance
    resp, err := http.Get("http://chainpulse:3001/metrics")
    // ... handle response
})
```

## Metrics Reference

### System Metrics
- `chainpulse_chains` - Number of monitored chains
- `chainpulse_txs` - Total transactions processed
- `chainpulse_packets` - Total packets processed
- `chainpulse_reconnects` - WebSocket reconnections
- `chainpulse_errors` - Error count per chain

### IBC Metrics
- `ibc_effected_packets` - Successfully relayed packets
- `ibc_uneffected_packets` - Frontrun packets
- `ibc_frontrun_counter` - Frontrun events
- `ibc_stuck_packets` - Stuck packet gauge

## Development

### Running Locally

1. Start the API backend:
   ```bash
   cd api && go run cmd/server/main.go
   ```

2. Start the frontend:
   ```bash
   cd webapp && yarn dev
   ```

3. Access the dashboard at http://localhost:5173/monitoring

### Adding New Metrics

1. Update types in `webapp/src/types/monitoring.ts`
2. Update parser in `webapp/src/utils/metricsParser.ts`
3. Create new visualization components
4. Update the monitoring page

## Troubleshooting

### No metrics showing
- Check Chainpulse is running: `curl http://localhost:3001/metrics`
- Verify API backend is proxying correctly
- Check browser console for CORS errors

### Stale data
- Verify WebSocket connections to blockchain nodes
- Check Chainpulse logs for errors
- Ensure refresh intervals are configured correctly

### Performance issues
- Reduce refresh intervals for large datasets
- Enable metric aggregation in Chainpulse
- Use pagination for historical data queries