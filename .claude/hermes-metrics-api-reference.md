# Hermes Metrics and API Reference

This document provides a comprehensive reference for Hermes IBC Relayer metrics and API endpoints, specifically for integration with Relayooor.

## Overview

Hermes exposes two main interfaces:
1. **REST API** (Port 5185) - For controlling and querying the relayer
2. **Metrics Endpoint** (Port 3001) - Prometheus-compatible metrics for monitoring

## REST API Endpoints

### Base URL
```
http://hermes:5185
```

### Available Endpoints

#### 1. Version Information
```http
GET /version
```
Returns Hermes version details.

**Response:**
```json
{
  "name": "hermes",
  "version": "1.10.5"
}
```

#### 2. State Information
```http
GET /state
```
Returns the current state of the relayer.

**Response:**
```json
{
  "chains": [
    {
      "id": "cosmoshub-4",
      "account": "cosmos1...",
      "gas_price": {
        "price": 0.025,
        "denom": "uatom"
      }
    }
  ]
}
```

#### 3. List Chains
```http
GET /chains
```
Returns all configured chains.

**Response:**
```json
[
  {
    "id": "cosmoshub-4",
    "type": "CosmosSdk",
    "pretty_name": "Cosmos Hub",
    "state": "Open",
    "counterparty_chains": ["osmosis-1", "noble-1", "stride-1"]
  }
]
```

#### 4. Chain Details
```http
GET /chain/{chain_id}
```
Returns detailed information about a specific chain.

**Response:**
```json
{
  "id": "cosmoshub-4",
  "type": "CosmosSdk",
  "account": {
    "address": "cosmos1...",
    "balance": {
      "amount": "1000000",
      "denom": "uatom"
    }
  },
  "clients": ["07-tendermint-1", "07-tendermint-2"],
  "connections": ["connection-0", "connection-1"],
  "channels": [
    {
      "id": "channel-0",
      "port": "transfer",
      "state": "Open",
      "counterparty": {
        "chain_id": "osmosis-1",
        "channel_id": "channel-1",
        "port_id": "transfer"
      }
    }
  ]
}
```

#### 5. Clear Packets (Not standard - custom endpoint)
```http
POST /clear_packets
```
Triggers packet clearing for a specific channel.

**Request Body:**
```json
{
  "chain_id": "cosmoshub-4",
  "port": "transfer",
  "channel": "channel-0",
  "sequences": [1, 2, 3]  // Optional, if not provided clears all
}
```

**Response:**
```json
{
  "success": true,
  "cleared_packets": 3,
  "tx_hash": "0x..."
}
```

## Prometheus Metrics Endpoint

### Base URL
```
http://hermes:3001/metrics
```

### Available Metrics

#### System Metrics

**`hermes_uptime_seconds`**
- Type: Counter
- Description: Time since Hermes started
- Usage: Monitor relayer uptime and detect restarts

**`hermes_build_info`**
- Type: Gauge
- Labels: `version`, `commit`
- Description: Build information about Hermes
- Usage: Track deployed versions

#### Chain Metrics

**`hermes_chains_count`**
- Type: Gauge
- Description: Number of chains configured
- Usage: Ensure all expected chains are configured

**`hermes_chain_status`**
- Type: Gauge
- Labels: `chain_id`, `status`
- Values: 1 (healthy), 0 (unhealthy)
- Description: Health status of each chain
- Usage: Alert on unhealthy chains

#### Packet Metrics

**`hermes_acknowledgment_packets_confirmed_total`**
- Type: Counter
- Labels: `src_chain`, `dst_chain`, `src_channel`, `dst_channel`
- Description: Total acknowledgment packets confirmed
- Usage: Track IBC activity between chains

**`hermes_receive_packets_confirmed_total`**
- Type: Counter
- Labels: `src_chain`, `dst_chain`, `src_channel`, `dst_channel`
- Description: Total receive packets confirmed
- Usage: Monitor packet flow

**`hermes_timeout_packets_confirmed_total`**
- Type: Counter
- Labels: `src_chain`, `dst_chain`, `src_channel`, `dst_channel`
- Description: Total timeout packets confirmed
- Usage: Identify problematic channels

#### Pending Packet Metrics

**`hermes_pending_acknowledgments`**
- Type: Gauge
- Labels: `chain_id`, `channel_id`, `port_id`
- Description: Number of pending acknowledgments
- Usage: Identify congested channels

**`hermes_pending_packets`**
- Type: Gauge
- Labels: `chain_id`, `channel_id`, `port_id`
- Description: Number of pending receive packets
- Usage: Monitor packet backlog

#### Transaction Metrics

**`hermes_tx_latency_submitted`**
- Type: Histogram
- Labels: `chain_id`, `channel_id`, `port_id`
- Description: Latency from packet event to transaction submission
- Usage: Monitor relayer responsiveness

**`hermes_tx_latency_confirmed`**
- Type: Histogram
- Labels: `chain_id`, `channel_id`, `port_id`
- Description: Latency from packet event to transaction confirmation
- Usage: Track end-to-end latency

#### Error Metrics

**`hermes_send_packet_errors_total`**
- Type: Counter
- Labels: `chain_id`, `channel_id`, `port_id`
- Description: Total errors sending packets
- Usage: Alert on persistent errors

**`hermes_acknowledgment_errors_total`**
- Type: Counter
- Labels: `chain_id`, `channel_id`, `port_id`
- Description: Total errors processing acknowledgments
- Usage: Identify problematic channels

#### Query Metrics

**`hermes_queries_total`**
- Type: Counter
- Labels: `chain_id`, `query_type`
- Description: Total number of queries to chain nodes
- Usage: Monitor RPC usage

**`hermes_queries_cache_hits_total`**
- Type: Counter
- Labels: `chain_id`, `query_type`
- Description: Total cache hits for queries
- Usage: Evaluate cache effectiveness

#### Wallet Metrics

**`hermes_wallet_balance`**
- Type: Gauge
- Labels: `chain_id`, `account`, `denom`
- Description: Current wallet balance
- Usage: Alert on low balances

## Integration with Relayooor

### 1. Monitoring Dashboard Integration

The frontend can consume these metrics to display:
- Real-time packet flow between chains
- Channel health and congestion
- Relayer wallet balances
- Performance metrics (latency, success rates)

### 2. Alerting Rules

Example Prometheus alerting rules:

```yaml
groups:
  - name: hermes_alerts
    rules:
      - alert: HermesDown
        expr: up{job="hermes"} == 0
        for: 5m
        annotations:
          summary: "Hermes is down"
          
      - alert: ChannelCongested
        expr: hermes_pending_packets > 100
        for: 10m
        annotations:
          summary: "Channel {{ $labels.channel_id }} is congested"
          
      - alert: LowWalletBalance
        expr: hermes_wallet_balance < 1000000
        for: 5m
        annotations:
          summary: "Low balance on {{ $labels.chain_id }}"
```

### 3. API Client Usage

The Relayooor API uses these endpoints through the `hermesClient`:

```go
// Check Hermes health
version, err := hermesClient.GetVersion(ctx)

// Clear packets programmatically
response, err := hermesClient.ClearPackets(ctx, &ClearPacketsRequest{
    Chain:     "cosmoshub-4",
    Channel:   "channel-0",
    Port:      "transfer",
    Sequences: []uint64{1, 2, 3},
})
```

### 4. Metric Collection Strategy

For efficient metric collection:
1. Scrape interval: 15-30 seconds
2. Retention: 15 days for raw metrics
3. Aggregation: 5-minute averages for dashboards
4. Key metrics to track:
   - `hermes_pending_packets` (immediate action needed)
   - `hermes_wallet_balance` (operational requirement)
   - `hermes_tx_latency_confirmed` (performance indicator)
   - Error metrics (system health)

## Utilizing Metrics in Relayooor

### 1. Channel Selection Algorithm
Use pending packet metrics to intelligently route clearing requests:
```javascript
// Select least congested channel
const channels = await getChannelMetrics();
const bestChannel = channels.reduce((best, current) => 
  current.pending_packets < best.pending_packets ? current : best
);
```

### 2. Dynamic Pricing
Adjust clearing fees based on congestion:
```javascript
const baseFee = 1.0;
const congestionMultiplier = Math.min(pendingPackets / 50, 3);
const clearingFee = baseFee * congestionMultiplier;
```

### 3. Predictive Maintenance
Alert operators before issues occur:
- Low wallet balance warnings
- Rising error rates
- Increasing latency trends

### 4. Performance Optimization
Use metrics to optimize operations:
- Batch clearing during low-congestion periods
- Pre-fund wallets based on historical usage
- Route packets through fastest channels

## Configuration Notes

### Disable Active Relaying
To ensure Hermes only clears packets on demand:

```toml
[mode.clients]
enabled = false
refresh = false
misbehaviour = false

[mode.connections]
enabled = false

[mode.channels]
enabled = false

[mode.packets]
enabled = true
clear_interval = 0  # Disable automatic clearing
clear_on_start = false
tx_confirmation = true
```

### Security Considerations
1. Metrics endpoint should not be publicly exposed
2. Use nginx proxy with authentication for external access
3. Sanitize chain/channel IDs in public dashboards
4. Rate limit API endpoints

## Troubleshooting

### Common Issues

1. **No metrics data**
   - Check telemetry is enabled in config
   - Verify port 3001 is accessible
   - Check Hermes logs for errors

2. **API not responding**
   - Verify REST is enabled in config
   - Check port 5185 is accessible
   - Ensure Hermes process is running

3. **High pending packets**
   - Check wallet balances
   - Verify chain RPC endpoints are responsive
   - Check for chain halts or upgrades

### Useful Queries

```bash
# Check Hermes version
curl http://localhost:5185/version

# Get current state
curl http://localhost:5185/state

# Check metrics
curl http://localhost:3001/metrics | grep hermes_pending

# Monitor specific channel
curl http://localhost:3001/metrics | grep 'channel="channel-0"'
```