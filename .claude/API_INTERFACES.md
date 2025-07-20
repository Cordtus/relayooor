# API Interfaces Documentation

## Overview

This document provides comprehensive documentation for the external API interfaces used by Relayooor: Hermes IBC Relayer and Chainpulse monitoring service.

**Last Verified**: 2025-01-19

## Deployment Architecture

All services run as Docker containers:
- **chainpulse**: Port 3001 (metrics)
- **api-backend**: Port 3000 (maps to internal 8080)
- **hermes**: Port 5185 (REST API), Port 3010 (telemetry/metrics)
- **postgres**: Port 5432
- **redis**: Port 6379
- **webapp**: Port 80
- **prometheus**: Port 9090
- **grafana**: Port 3003

## Hermes API Interface

### Base Configuration

- **Base URL**: `http://localhost:5185`
- **Telemetry URL**: `http://localhost:3001`
- **Authentication**: None
- **Protocol**: REST/HTTP

### Endpoints

#### 1. Version Information

```http
GET /version

Response:
{
  "name": "hermes",
  "version": "1.7.4"
}
```

#### 2. Relayer State

```http
GET /state

Response:
{
  "chains": [
    {
      "id": "cosmoshub-4",
      "connected": true,
      "height": 18234567
    },
    {
      "id": "osmosis-1",
      "connected": true,
      "height": 12345678
    }
  ]
}
```

#### 3. List Configured Chains

```http
GET /chains

Response:
[
  {
    "id": "cosmoshub-4",
    "type": "CosmosSdk",
    "account_prefix": "cosmos",
    "gas_price": {
      "price": 0.025,
      "denom": "uatom"
    }
  }
]
```

#### 4. Chain Details

```http
GET /chain/{chain_id}

Response:
{
  "id": "cosmoshub-4",
  "rpc_addr": "https://cosmos-rpc.example.com:443",
  "grpc_addr": "https://cosmos-grpc.example.com:9090",
  "key_name": "relayer",
  "account": "cosmos1...",
  "balance": {
    "amount": "1000000",
    "denom": "uatom"
  }
}
```

#### 5. Clear Packets (Primary Endpoint)

```http
POST /clear_packets
Content-Type: application/json

Request:
{
  "chain_id": "cosmoshub-4",
  "channel_id": "channel-141",
  "port_id": "transfer",
  "sequences": [123, 124, 125]
}

Response (Success):
{
  "status": "success",
  "tx_hash": "A1B2C3D4E5F6...",
  "cleared_sequences": [123, 124, 125]
}

Response (Error):
{
  "status": "error",
  "error": "insufficient gas",
  "details": {
    "required_gas": 200000,
    "provided_gas": 100000
  }
}
```

#### 6. Query Unreceived Packets

```http
GET /query/packet/unreceived?chain={chain}&channel={channel}&port={port}

Response:
{
  "sequences": [123, 124, 125, 126],
  "total": 4
}
```

#### 7. Query Unacknowledged Packets

```http
GET /query/packet/pending?chain={chain}&channel={channel}&port={port}

Response:
{
  "sequences": [127, 128],
  "total": 2
}
```

#### 8. Health Check

```http
GET /health

Response:
{
  "status": "healthy",
  "chains": {
    "cosmoshub-4": "connected",
    "osmosis-1": "connected",
    "neutron-1": "disconnected"
  },
  "version": "1.7.4"
}
```

### Telemetry Metrics (Prometheus Format)

```http
GET http://localhost:3001/metrics

Response (text/plain):
# HELP hermes_acknowledged_packets_total Total number of acknowledged packets
# TYPE hermes_acknowledged_packets_total counter
hermes_acknowledged_packets_total{chain="cosmoshub-4",channel="channel-141",port="transfer"} 1234

# HELP hermes_chain_height Current height of the chain
# TYPE hermes_chain_height gauge
hermes_chain_height{chain="cosmoshub-4"} 18234567

# HELP hermes_wallet_balance Wallet balance in uatom
# TYPE hermes_wallet_balance gauge
hermes_wallet_balance{chain="cosmoshub-4",denom="uatom"} 1000000

# HELP hermes_tx_latency_confirmed Time taken for transaction confirmation
# TYPE hermes_tx_latency_confirmed histogram
hermes_tx_latency_confirmed_bucket{chain="cosmoshub-4",le="1"} 45
hermes_tx_latency_confirmed_bucket{chain="cosmoshub-4",le="5"} 89
hermes_tx_latency_confirmed_bucket{chain="cosmoshub-4",le="+Inf"} 92
```

## Chainpulse API Interface (Enhanced Fork v0.4.0-pre)

### Base Configuration

- **Base URL**: `http://localhost:3001`
- **Metrics URL**: `http://localhost:3001/metrics`
- **API URL**: `http://localhost:3001/api/v1`
- **Authentication**: None required
- **Protocol**: Prometheus metrics + REST API (NEW)

### Endpoints Overview

- **Base URL**: `http://localhost:3001`
- **API Base**: `http://localhost:3001/api/v1`
- **Metrics**: `http://localhost:3001/metrics`

### Verified Endpoints

#### 1. Prometheus Metrics

```http
GET /metrics

Response (text/plain):
# HELP chainpulse_chains The number of chains being monitored
# TYPE chainpulse_chains gauge
chainpulse_chains 6

# HELP chainpulse_errors The number of times an error was encountered
# TYPE chainpulse_errors counter
chainpulse_errors{chain_id="jackal-1"} 32

# HELP chainpulse_packets The number of packets processed
# TYPE chainpulse_packets counter
chainpulse_packets{chain_id="osmosis-1"} 32783

# HELP chainpulse_reconnects The number of times we had to reconnect to the WebSocket
# TYPE chainpulse_reconnects counter
chainpulse_reconnects{chain_id="stride-1"} 30

# HELP chainpulse_txs The number of txs processed
# TYPE chainpulse_txs counter
chainpulse_txs{chain_id="cosmoshub-4"} 9064

# HELP ibc_effected_packets The number of IBC packets that have been relayed and were effected
# TYPE ibc_effected_packets counter
ibc_effected_packets{chain_id="axelar-dojo-1",dst_channel="channel-0",dst_port="transfer",memo="",signer="axelar1...",src_channel="channel-19",src_port="transfer"} 5

# HELP ibc_uneffected_packets The number of IBC packets that were relayed but not effected
# TYPE ibc_uneffected_packets counter

# HELP ibc_frontrun_counter The number of times a signer gets frontrun by the original signer
# TYPE ibc_frontrun_counter counter

# HELP ibc_stuck_packets The number of packets stuck on an IBC channel
# TYPE ibc_stuck_packets gauge
ibc_stuck_packets{dst_chain="channel-0",src_chain="osmosis-1",src_channel="channel-165"} 6220

# HELP ibc_stuck_packets_detailed Detailed stuck packet tracking with user info
# TYPE ibc_stuck_packets_detailed gauge

# HELP ibc_packet_age_seconds Age of unrelayed packets in seconds
# TYPE ibc_packet_age_seconds gauge

# HELP ibc_packets_near_timeout Number of packets nearing their timeout deadline
# TYPE ibc_packets_near_timeout gauge
ibc_packets_near_timeout{src_chain="osmosis-1",dst_chain="cosmoshub-4",src_channel="channel-0",dst_channel="channel-141",timeout_type="timestamp"} 2

# HELP ibc_packet_timeout_seconds Time until packet timeout in seconds (negative if expired)
# TYPE ibc_packet_timeout_seconds gauge
ibc_packet_timeout_seconds{src_chain="osmosis-1",dst_chain="cosmoshub-4",src_channel="channel-0",dst_channel="channel-141"} -1800

# HELP ibc_stuck_packets_by_user Number of stuck packets by user address
# TYPE ibc_stuck_packets_by_user gauge
ibc_stuck_packets_by_user{address="osmo1...",role="sender"} 3

# HELP ibc_stuck_value Total value stuck in packets by denomination
# TYPE ibc_stuck_value gauge
ibc_stuck_value{denom="uosmo",src_chain="osmosis-1",dst_chain="cosmoshub-4"} 1500000000
```

#### 2. REST API Endpoints (NEW in v0.4.0-pre)

##### Get Stuck Packets
```http
GET /api/v1/packets/stuck?min_age_seconds=900&limit=100

Response:
{
  "packets": [
    {
      "chain_id": "osmosis-1",
      "sequence": 895396,
      "src_channel": "channel-750",
      "dst_channel": "channel-1",
      "sender": "osmo1m8wg4vxkefhs374qxmmqpyusgz289wmulex5qdwpfx7jnrxzer5s9cv83q",
      "receiver": "noble1ejmfwh88dxrehv345kj4743uznwpzkaz5tpv8z",
      "amount": "10832264",
      "denom": "transfer/channel-750/uusdc",
      "age_seconds": 330516,
      "relay_attempts": 3,
      "last_attempt_by": "osmo1j6swju2q7zywxmpcttcw4k98j7fphx5nu4scjy",
      "timeout_timestamp": 1234567890000000000,
      "ibc_version": "v1"
    }
  ],
  "total": 150,
  "api_version": "1.0"
}
```

##### Find Packets by User
```http
GET /api/v1/packets/by-user?address={address}&role={sender|receiver}&status={all|pending|delivered}

Query Parameters:
- address (required): User address
- role (optional): Filter by sender or receiver, default: both
- status (optional): Filter by delivery status, default: all
- limit (optional): Max results, default: 100
```

##### Get Packet Details
```http
GET /api/v1/packets/{chain_id}/{channel}/{sequence}

Response includes full packet data with timeout information.
```

##### Get Expiring Packets
```http
GET /api/v1/packets/expiring?minutes=60

Returns packets that will timeout within the specified window.
```

##### Get Channel Congestion
```http
GET /api/v1/channels/congestion

Response:
{
  "channels": [
    {
      "src_channel": "channel-0",
      "dst_channel": "channel-141",
      "chain_id": "osmosis-1",
      "stuck_count": 150,
      "oldest_stuck_age_seconds": 7200,
      "total_value": {
        "uosmo": "1500000000",
        "uusdc": "500000000"
      }
    }
  ]
}
```

##### Get User Transfer History
```http
GET /api/v1/users/{address}/transfers?direction={sent|received|both}

Returns complete transfer history for a user address.
```

##### Get Packet Analytics
```http
GET /api/v1/packets/analytics?chain_id={chain}&period={24h|7d|30d}

Returns analytics including delivery rates, timeout statistics, and total value transferred.
```

##### Get Expired Packets (NEW)
```http
GET /api/v1/packets/expired

Response:
{
  "packets": [
    {
      "chain_id": "osmosis-1",
      "sequence": 123456,
      "src_channel": "channel-0",
      "dst_channel": "channel-141",
      "sender": "osmo1abc...",
      "receiver": "cosmos1xyz...",
      "amount": "1000000",
      "denom": "uosmo",
      "seconds_since_timeout": 1800,
      "timeout_type": "height"
    }
  ],
  "api_version": "1.0"
}
```

##### Find Duplicate Packets (NEW)
```http
GET /api/v1/packets/duplicates

Response:
{
  "duplicates": [
    {
      "data_hash": "2ec9535148dfadb1c99e22ec8a9a7fb3b07b3d40fbd63225aef93ef112ec34eb",
      "count": 41,
      "packets": [
        {
          "chain_id": "osmosis-1",
          "sequence": 361754,
          "src_channel": "channel-19774",
          "sender": "osmo1w70qj355ra7yu43d0pma5ds6zft043zhjgr5d6",
          "created_at": "2025-07-19 23:35:05"
        }
      ]
    }
  ],
  "api_version": "1.0"
}
```

##### Enhanced User Queries (NEW)
```http
GET /api/v1/packets/by-user?address={address}&role={sender|receiver}&limit={limit}&offset={offset}

Query Parameters:
- address (required): User address
- role (optional): Filter by sender or receiver, default: both
- limit (optional): Max results, default: 100
- offset (optional): Pagination offset

Response includes comprehensive packet data with relay attempts and timeout info.
```

### Relayooor API Facade (Verified Endpoints)

The Relayooor API (`api-backend` container on port 3000) parses Chainpulse Prometheus metrics and provides structured JSON responses:

#### 1. Monitoring Data (VERIFIED)

```http
GET /api/monitoring/data

Response:
{
  "chains": [
    {
      "chain_id": "osmosis-1",
      "errors": 0,
      "name": "Osmosis",
      "packets_24h": 32855,
      "status": "connected",
      "txs_total": 24965
    }
  ],
  "channels": [
    {
      "dst_channel": "channel-37",
      "dst_port": "transfer",
      "effected": 4,
      "packets_pending": 0,
      "src": "osmosis-1",
      "src_channel": "channel-0",
      "src_port": "transfer",
      "status": "active",
      "success_rate": 100,
      "total_packets": 4,
      "uneffected": 0
    }
  ],
  "top_relayers": [
    {
      "address": "axelar16vmp7sz28pnvgz6f3zm6q93y39jsd33aszn9np",
      "earnings_24h": "$2.50",
      "name": "Inter Blockchain Services Relayer",
      "packets_relayed": 123,
      "success_rate": 95.2
    }
  ],
  "alerts": [],
  "timestamp": "2025-01-19T18:44:17.309296-06:00"
}
```

#### 2. Stuck Packets (VERIFIED - Partially Mock Data)

```http
GET /api/packets/stuck

Response:
[
  {
    "id": "osmosis-1-895396",
    "channelId": "channel-750",
    "sequence": 895396,
    "sourceChain": "osmosis-1",
    "destinationChain": "channel-1",  // Note: Shows channel ID instead of chain ID
    "stuckDuration": "3d",
    "amount": "10832264",
    "denom": "transfer/channel-750/uusdc",
    "sender": "osmo1m8wg4vxkefhs374qxmmqpyusgz289wmulex5qdwpfx7jnrxzer5s9cv83q",
    "receiver": "noble1ejmfwh88dxrehv345kj4743uznwpzkaz5tpv8z",
    "timestamp": "2025-07-15T21:57:53.006771-06:00",
    "txHash": "<nil>",
    "relayAttempts": 3
  }
]
```

**Note**: This endpoint parses `ibc_stuck_packets` metrics but adds mock data for sender, receiver, amount, and timestamp fields.

#### 3. Packet Details

```http
GET /api/v1/chainpulse/packets/{chain}/{channel}/{sequence}

Response:
{
  "packet": {
    "id": "cosmoshub-4-channel-141-123",
    // ... full packet details
  },
  "clearing_eligible": true,
  "clearing_requirements": {
    "needs_payment": true,
    "minimum_fee_usd": "0.50",
    "estimated_gas": 200000
  }
}
```

#### 4. Channel Congestion

```http
GET /api/v1/chainpulse/channels/congestion

Response:
{
  "channels": [
    {
      "src_chain": "cosmoshub-4",
      "dst_chain": "osmosis-1",
      "channel_id": "channel-141",
      "pending_packets": 4,
      "average_clear_time_hours": 2.5,
      "congestion_level": "medium"
    }
  ]
}
```

## WebSocket Interface

### Connection

```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/ws');

// Authentication after connection
ws.onopen = () => {
  ws.send(JSON.stringify({
    type: 'auth',
    token: 'Bearer eyJhbGciOiJ...'
  }));
};
```

### Message Types

#### 1. Subscribe to Updates

```json
{
  "type": "subscribe",
  "topics": ["packet_updates", "clearing_status"],
  "filters": {
    "chain_id": "cosmoshub-4",
    "user_address": "cosmos1..."
  }
}
```

#### 2. Packet Update Notification

```json
{
  "type": "packet_update",
  "data": {
    "packet_id": "cosmoshub-4-channel-141-123",
    "status": "cleared",
    "tx_hash": "A1B2C3D4E5F6...",
    "timestamp": "2024-01-15T10:35:00Z"
  }
}
```

#### 3. Clearing Status Update

```json
{
  "type": "clearing_status",
  "data": {
    "token_id": "tok_abc123",
    "status": "completed",
    "packet_id": "cosmoshub-4-channel-141-123",
    "tx_hash": "A1B2C3D4E5F6..."
  }
}
```

#### 4. Error Notification

```json
{
  "type": "error",
  "error": {
    "code": "INSUFFICIENT_GAS",
    "message": "Transaction failed due to insufficient gas",
    "details": {
      "required": 200000,
      "provided": 100000
    }
  }
}
```

## Authentication

### JWT Token Structure

```json
{
  "user_id": "usr_123456",
  "wallet_address": "cosmos1...",
  "chain_id": "cosmoshub-4",
  "exp": 1704456000,
  "iat": 1704369600
}
```

### Required Headers

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

### Public Endpoints (No Auth Required)

- All Chainpulse metrics endpoints
- Hermes health and metrics endpoints
- Stuck packets queries (read-only)
- Channel congestion data

### Protected Endpoints (Auth Required)

- Packet clearing operations
- User-specific packet queries
- Payment verification
- WebSocket connections

## Error Response Format

### Standard Error Response

```json
{
  "error": {
    "code": "PACKET_NOT_FOUND",
    "title": "Packet Not Found",
    "message": "The requested packet could not be found",
    "details": {
      "packet_id": "cosmoshub-4-channel-141-999",
      "searched_chains": ["cosmoshub-4", "osmosis-1"]
    },
    "suggestion": "Verify the packet ID and try again"
  },
  "request_id": "req_abc123",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Common Error Codes

- `UNAUTHORIZED` - Missing or invalid authentication
- `PACKET_NOT_FOUND` - Requested packet doesn't exist
- `PACKET_NOT_STUCK` - Packet is not eligible for clearing
- `INSUFFICIENT_BALANCE` - Relayer wallet has insufficient funds
- `CHAIN_UNAVAILABLE` - Chain RPC is not responding
- `RATE_LIMITED` - Too many requests
- `INTERNAL_ERROR` - Server error

## Rate Limiting

### Current Limits (Not Yet Enforced)

- **API Requests**: 100 requests per minute per IP
- **WebSocket Connections**: 5 concurrent connections per user
- **Clearing Operations**: 10 per hour per user

### Rate Limit Headers

```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1704370200
```

## Integration Best Practices

### 1. Connection Management

```go
// Implement exponential backoff for retries
func connectWithRetry(url string, maxRetries int) (*Client, error) {
    for i := 0; i < maxRetries; i++ {
        client, err := NewClient(url)
        if err == nil {
            return client, nil
        }

        waitTime := time.Duration(math.Pow(2, float64(i))) * time.Second
        time.Sleep(waitTime)
    }
    return nil, fmt.Errorf("failed after %d retries", maxRetries)
}
```

### 2. Error Handling

```go
// Parse and handle specific error codes
switch errResp.Code {
case "INSUFFICIENT_GAS":
    // Increase gas and retry
    req.GasLimit = errResp.Details["required_gas"].(int) * 1.2
    return retry(req)

case "PACKET_NOT_STUCK":
    // Packet was already cleared
    return nil

default:
    return fmt.Errorf("clearing failed: %s", errResp.Message)
}
```

### 3. Monitoring Integration

```go
// Collect metrics from both services
func collectMetrics() {
    hermesMetrics := fetchMetrics("http://hermes:3001/metrics")
    chainpulseMetrics := fetchMetrics("http://chainpulse:3001/metrics")

    // Parse and forward to monitoring system
    prometheus.Write(hermesMetrics)
    prometheus.Write(chainpulseMetrics)
}
```

## Testing Endpoints

### Mock Responses for Development

```bash
# Simulate stuck packet
curl -X POST http://localhost:8080/api/v1/test/create-stuck-packet \
  -H "Content-Type: application/json" \
  -d '{"chain": "cosmoshub-4", "channel": "channel-141"}'

# Simulate clearing success
curl -X POST http://localhost:8080/api/v1/test/simulate-clearing \
  -H "Content-Type: application/json" \
  -d '{"packet_id": "cosmoshub-4-channel-141-123", "success": true}'
```

## Current Implementation Status

### Data Sources

1. **Chainpulse**: Provides ONLY Prometheus metrics at `/metrics` endpoint
   - No JSON API endpoints
   - No health check endpoint
   - All packet data must be extracted from metrics

2. **Relayooor API**: Parses Chainpulse metrics and provides JSON API
   - Real data: Chain status, channel activity, relayer performance
   - Mock data: User addresses, amounts, timestamps, transaction hashes
   - Mixed: Stuck packets (real counts, mock details)

3. **Hermes**: Currently unable to verify due to authentication issues
   - Expected to provide packet clearing functionality
   - REST API on port 5185
   - Telemetry metrics on port 3010

### Key Findings

1. **No Native JSON APIs**: Chainpulse only exposes Prometheus metrics, not the JSON APIs documented
2. **Mock Data Still Present**: The API adds mock data for user-specific fields since these aren't available in metrics
3. **Authentication Issues**: Hermes requires proper URL encoding of credentials in configuration
4. **Container Architecture**: All services run as Docker containers with specific port mappings

## Migration Notes

### From Direct Chainpulse Access

Since Chainpulse only provides Prometheus metrics:

1. Use the Relayooor API facade for JSON responses
2. Accept that user-specific data will be incomplete without additional data sources
3. Consider implementing a proper database to store packet details

### From Manual Hermes Commands

If migrating from CLI-based clearing to API:

1. Fix Hermes authentication configuration first
2. Use the REST API on port 5185 once operational
3. Monitor status via telemetry metrics on port 3010
