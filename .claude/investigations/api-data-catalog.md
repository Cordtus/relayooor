# API Data Catalog

Date: 2025-01-19
Status: Verified against running instances

## Simple API (/api)

Base URL: http://localhost:8080

### Available Endpoints (Verified)

#### 1. GET /api/packets/stuck
Returns stuck IBC packets parsed from Chainpulse metrics.

**Response Structure**:
```json
[
  {
    "id": "osmosis-1-895396",
    "channelId": "channel-750",
    "sequence": 895396,
    "sourceChain": "osmosis-1",
    "destinationChain": "channel-1",
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

**Data Source**: Parsed from `ibc_stuck_packets_detailed` metrics
**Current Issues**: 
- destinationChain shows channel ID instead of chain ID
- txHash is always nil
- Timestamps appear to be mock/generated

#### 2. GET /api/monitoring/data
Returns structured monitoring data with chains, channels, and relayers.

**Response Structure**:
```json
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

**Data Sources**:
- chains: Parsed from `chainpulse_packets`, `chainpulse_txs`, `chainpulse_errors`
- channels: Parsed from `ibc_effected_packets`, `ibc_uneffected_packets`
- top_relayers: Parsed from `ibc_effected_packets` grouped by signer
- alerts: Currently always empty array

#### 3. GET /api/metrics/chainpulse
Proxies raw Prometheus metrics from Chainpulse.

**Response**: Raw Prometheus text format
**Source**: Direct proxy to http://chainpulse:3001/metrics

#### 4. GET /api/channels/congestion
Returns channel congestion data (currently falls back to mock data when Chainpulse doesn't provide this endpoint).

#### 5. GET /api/packets/{chain}/{channel}/{sequence}
Returns detailed packet information (currently falls back to mock data).

#### 6. GET /api/user/{wallet}/transfers
Returns transfers for a specific wallet address.

#### 7. GET /api/user/{wallet}/stuck
Returns only stuck transfers for a specific wallet.

#### 8. POST /api/packets/clear
Initiates packet clearing (currently returns mock response).

#### 9. GET /api/chains/registry
Returns comprehensive chain information including RPC endpoints.

#### 10. GET /api/monitoring/metrics
Returns platform-wide metrics aggregated from Chainpulse.

#### 11. GET /api/statistics/platform
Returns platform statistics including total value locked and transaction counts.

### Data Processing Flow

1. **Chainpulse Metrics** → **API Parser** → **Structured JSON**
2. The API parses Prometheus metrics using regex patterns
3. Aggregates data by chain, channel, and relayer
4. Adds calculated fields (success_rate, packets_24h, etc.)
5. Returns structured JSON suitable for frontend consumption

### Current Implementation Status

**Working with Real Data**:
- Basic monitoring metrics (chains, packets, errors)
- Channel activity tracking
- Relayer performance metrics
- Stuck packet counts

**Still Using Mock Data**:
- Detailed stuck packet information (sender, receiver, amounts)
- Packet clearing operations
- User-specific transfer queries
- Channel congestion details
- Individual packet details

### Integration Notes

1. **Chainpulse Dependency**: The API heavily depends on Chainpulse metrics being available
2. **Fallback Behavior**: When Chainpulse is unavailable, most endpoints fall back to mock data
3. **Data Freshness**: Metrics are fetched in real-time from Chainpulse
4. **Parsing Logic**: Complex regex patterns extract data from Prometheus format

## Chainpulse Service

Base URL: http://localhost:3001

### Available Endpoints (Verified)

#### 1. GET /metrics
The only endpoint exposed by Chainpulse.

**Format**: Prometheus text format
**Content**: See chainpulse-metrics-catalog.md for detailed metric descriptions

### Missing Expected Endpoints

Based on API_INTERFACES.md, these endpoints were expected but not found:
- GET /health
- GET /api/v1/packets/stuck
- GET /api/v1/packets/by-user
- GET /api/v1/channels/congestion
- GET /api/v1/packets/{chain}/{channel}/{sequence}

**Conclusion**: Chainpulse only provides Prometheus metrics. All JSON API functionality is implemented in the Relayooor API layer by parsing these metrics.

## Hermes Service

Status: Currently failing to start due to authentication issues
Next Step: Fix URL encoding in entrypoint script or use pre-authenticated config

### Expected Endpoints (from API_INTERFACES.md)

- GET /version
- GET /state
- GET /chains
- GET /chain/{chain_id}
- POST /clear_packets
- GET /query/packet/unreceived
- GET /query/packet/pending
- GET /health
- GET http://localhost:3010/metrics (Telemetry)

## Next Steps

1. Fix Hermes authentication to verify its endpoints
2. Implement real data fetching for currently mocked endpoints
3. Consider implementing missing Chainpulse API endpoints in the middleware
4. Map all data flows to remove mock implementations systematically