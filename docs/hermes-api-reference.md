# Hermes REST API and Telemetry Reference

## Overview

This document provides a comprehensive reference for Hermes REST API endpoints and telemetry metrics used in the Relayooor project.

## Hermes REST API Endpoints

The Hermes REST API runs on port 5185 by default and provides the following endpoints:

### Core Endpoints

#### GET /version
Returns the Hermes version information.

**Response:**
```json
{
  "name": "hermes",
  "version": "1.7.0"
}
```

#### GET /state
Returns the current state of the Hermes relayer.

**Response:**
```json
{
  "status": "running",
  "chains": ["cosmoshub-4", "osmosis-1"],
  "active_connections": 5
}
```

#### GET /chains
Lists all configured chains in Hermes.

**Response:**
```json
[
  {
    "id": "cosmoshub-4",
    "type": "CosmosSdk",
    "account_prefix": "cosmos",
    "key_name": "relayer"
  },
  {
    "id": "osmosis-1",
    "type": "CosmosSdk",
    "account_prefix": "osmo",
    "key_name": "relayer"
  }
]
```

#### GET /chain/{chain_id}
Returns detailed information about a specific chain.

**Parameters:**
- `chain_id`: The chain identifier (e.g., "cosmoshub-4")

**Response:**
```json
{
  "id": "cosmoshub-4",
  "type": "CosmosSdk",
  "rpc_addr": "https://cosmoshub-4-skip-rpc.polkachu.com",
  "grpc_addr": "https://cosmoshub-4-skip-grpc.polkachu.com:14990",
  "account_prefix": "cosmos",
  "key_name": "relayer",
  "gas_price": {
    "price": 0.025,
    "denom": "uatom"
  }
}
```

#### POST /clear_packets
Clears pending packets on a specific channel.

**Request Body:**
```json
{
  "chain_id": "cosmoshub-4",
  "port": "transfer",
  "channel": "channel-141",
  "sequences": [1, 2, 3]  // Optional: specific sequences to clear
}
```

**Response:**
```json
{
  "success": true,
  "tx_hash": "0x123...",
  "cleared_sequences": [1, 2, 3]
}
```

## Hermes Telemetry Metrics

Hermes exposes Prometheus-compatible metrics on port 3001 by default. These metrics provide detailed information about the relayer's performance and health.

### System Metrics

#### hermes_uptime_seconds
Time since Hermes started (in seconds).

#### hermes_build_info
Build information about Hermes.
- Labels: `version`, `commit`

### Chain Metrics

#### hermes_chains_count
Total number of chains configured.

#### hermes_chain_status
Status of each configured chain (1 = healthy, 0 = unhealthy).
- Labels: `chain_id`

### Packet Metrics

#### hermes_acknowledgment_packets_confirmed_total
Total number of acknowledgment packets confirmed.
- Labels: `src_chain`, `dst_chain`, `src_channel`, `dst_channel`, `src_port`, `dst_port`

#### hermes_receive_packets_confirmed_total
Total number of receive packets confirmed.
- Labels: `src_chain`, `dst_chain`, `src_channel`, `dst_channel`, `src_port`, `dst_port`

#### hermes_timeout_packets_confirmed_total
Total number of timeout packets confirmed.
- Labels: `src_chain`, `dst_chain`, `src_channel`, `dst_channel`, `src_port`, `dst_port`

### Transaction Metrics

#### hermes_tx_latency_submitted
Latency between transaction submission (in milliseconds).
- Labels: `chain_id`, `channel_id`, `port_id`
- Type: Histogram with buckets: [500, 1000, 2000, 3000, 5000, 7500, 10000, 15000, 20000]

#### hermes_tx_latency_confirmed
Latency between transaction submission and confirmation (in milliseconds).
- Labels: `chain_id`, `channel_id`, `port_id`
- Type: Histogram with buckets: [1000, 2000, 4000, 6000, 8000, 10000, 15000, 20000, 25000, 30000]

### Connection Metrics

#### hermes_ws_reconnects_total
Total number of WebSocket reconnections.
- Labels: `chain_id`

#### hermes_ws_events_total
Total number of WebSocket events received.
- Labels: `chain_id`

### Error Metrics

#### hermes_send_packet_errors_total
Total number of errors when sending packets.
- Labels: `chain_id`, `channel_id`, `port_id`

#### hermes_acknowledgment_errors_total
Total number of errors when processing acknowledgments.
- Labels: `chain_id`, `channel_id`, `port_id`

#### hermes_timeout_errors_total
Total number of errors when processing timeouts.
- Labels: `chain_id`, `channel_id`, `port_id`

### Query Metrics

#### hermes_queries_total
Total number of queries made to chains.
- Labels: `chain_id`, `query_type`

#### hermes_queries_cache_hits_total
Total number of query cache hits.
- Labels: `chain_id`, `query_type`

### Wallet Metrics

#### hermes_wallet_balance
Current wallet balance for each configured chain.
- Labels: `chain_id`, `account`, `denom`

## Integration with Relayooor

### API Integration

The Relayooor API integrates with Hermes through the following handlers:

1. **HermesClient** (`/relayer-middleware/api/pkg/clearing/hermes_client.go`)
   - Provides a Go client for interacting with Hermes REST API
   - Handles packet clearing requests
   - Manages version and health checks

2. **Hermes Handlers** (`/relayer-middleware/api/pkg/handlers/hermes.go`)
   - `GetHermesVersion`: Proxies version requests
   - `GetHermesHealth`: Checks Hermes connectivity
   - `ClearPacketsWithHermes`: Initiates packet clearing

### Telemetry Integration

The telemetry metrics are consumed by:

1. **Prometheus** - Scrapes metrics from port 3001
2. **Chainpulse** - Aggregates and enhances metrics
3. **Frontend Dashboard** - Displays metrics via the monitoring API

### Configuration

#### Environment Variables

```bash
# Hermes REST API
HERMES_REST_URL=http://localhost:5185

# Hermes Telemetry
HERMES_TELEMETRY_PORT=3001
HERMES_LEGACY_TELEMETRY_PORT=3002
```

#### Docker Compose Configuration

```yaml
hermes:
  image: informalsystems/hermes:latest
  ports:
    - "5185:5185"  # REST API
    - "3001:3001"  # Telemetry
  environment:
    - HERMES_REST_ENABLED=true
    - HERMES_TELEMETRY_ENABLED=true
```

### Example Usage

#### Checking Hermes Health

```bash
# Via direct API
curl http://localhost:5185/version

# Via Relayooor API
curl http://localhost:8080/api/v1/hermes/health
```

#### Viewing Metrics

```bash
# Raw Prometheus metrics
curl http://localhost:3001/metrics

# Specific metric
curl -s http://localhost:3001/metrics | grep hermes_acknowledgment_packets_confirmed_total
```

#### Clearing Packets

```bash
# Via Relayooor API
curl -X POST http://localhost:8080/api/v1/hermes/clear \
  -H "Content-Type: application/json" \
  -d '{
    "chain": "cosmoshub-4",
    "channel": "channel-141",
    "port": "transfer",
    "sequences": [1, 2, 3]
  }'
```

## Metric Alerts and Monitoring

### Recommended Alerts

1. **High Error Rate**
   ```promql
   rate(hermes_send_packet_errors_total[5m]) > 0.1
   ```

2. **Low Wallet Balance**
   ```promql
   hermes_wallet_balance < 1000000
   ```

3. **High Transaction Latency**
   ```promql
   histogram_quantile(0.95, hermes_tx_latency_confirmed) > 20000
   ```

4. **Frequent Reconnections**
   ```promql
   rate(hermes_ws_reconnects_total[1h]) > 5
   ```

### Grafana Dashboard Queries

1. **Packet Success Rate**
   ```promql
   sum(rate(hermes_receive_packets_confirmed_total[5m])) / 
   sum(rate(hermes_send_packet_errors_total[5m]) + rate(hermes_receive_packets_confirmed_total[5m]))
   ```

2. **Average Transaction Latency**
   ```promql
   histogram_quantile(0.5, sum(rate(hermes_tx_latency_confirmed_bucket[5m])) by (le))
   ```

3. **Packets per Channel**
   ```promql
   sum by (src_channel, dst_channel) (
     rate(hermes_receive_packets_confirmed_total[5m])
   )
   ```

## Troubleshooting

### Common Issues

1. **Hermes REST API Not Responding**
   - Check if Hermes is running: `supervisorctl status hermes hermes-rest`
   - Verify REST is enabled in config: `rest.enabled = true`
   - Check port binding: `netstat -an | grep 5185`

2. **No Metrics Available**
   - Verify telemetry is enabled: `telemetry.enabled = true`
   - Check port binding: `netstat -an | grep 3001`
   - Ensure Prometheus can reach the endpoint

3. **Packet Clearing Failures**
   - Check wallet balance for gas fees
   - Verify chain connectivity
   - Review Hermes logs for errors

### Debug Commands

```bash
# Check Hermes status
curl -s http://localhost:5185/state | jq .

# View recent packet activity
curl -s http://localhost:3001/metrics | grep -E "hermes_.*packets.*_total"

# Check wallet balances
curl -s http://localhost:3001/metrics | grep hermes_wallet_balance

# Monitor error rates
watch -n 5 'curl -s http://localhost:3001/metrics | grep -E "hermes_.*errors_total"'
```

## References

- [Hermes Documentation](https://hermes.informal.systems/)
- [Prometheus Metric Types](https://prometheus.io/docs/concepts/metric_types/)
- [Relayooor API Documentation](./API_IMPROVEMENTS.md)