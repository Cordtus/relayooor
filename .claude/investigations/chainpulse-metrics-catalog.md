# Chainpulse Metrics Catalog

Date: 2025-01-19
Status: Verified against running instance

## Base Information

- **Metrics URL**: http://localhost:3001/metrics
- **Format**: Prometheus text format
- **Current Status**: Running and collecting data from 6 chains

## Available Metrics

### 1. System Metrics

#### chainpulse_chains
- **Type**: gauge
- **Description**: The number of chains being monitored
- **Current Value**: 6
- **Example**: `chainpulse_chains 6`

#### chainpulse_errors
- **Type**: counter
- **Description**: The number of times an error was encountered
- **Labels**: chain_id
- **Example**: `chainpulse_errors{chain_id="jackal-1"} 32`

#### chainpulse_packets
- **Type**: counter
- **Description**: The number of packets processed
- **Labels**: chain_id
- **Example**: `chainpulse_packets{chain_id="osmosis-1"} 32783`

#### chainpulse_reconnects
- **Type**: counter
- **Description**: The number of times we had to reconnect to the WebSocket
- **Labels**: chain_id
- **Example**: `chainpulse_reconnects{chain_id="stride-1"} 30`

#### chainpulse_txs
- **Type**: counter
- **Description**: The number of txs processed
- **Labels**: chain_id
- **Example**: `chainpulse_txs{chain_id="cosmoshub-4"} 123456`

### 2. IBC Packet Metrics

#### ibc_effected_packets
- **Type**: counter
- **Description**: The number of IBC packets that have been relayed and were effected
- **Labels**: 
  - chain_id
  - src_channel
  - src_port
  - dst_channel
  - dst_port
  - signer (relayer address)
  - memo (relayer identifier)
- **Example**: `ibc_effected_packets{chain_id="axelar-dojo-1",dst_channel="channel-0",dst_port="transfer",memo="",signer="axelar1zzssvjchht00qag3x9987esrwux09tv5j2lr90",src_channel="channel-19",src_port="transfer"} 5`

#### ibc_uneffected_packets
- **Type**: counter
- **Description**: The number of IBC packets that were relayed but not effected (frontrun)
- **Labels**: Same as ibc_effected_packets
- **Example**: Similar structure to effected packets

#### ibc_frontrun_counter
- **Type**: counter
- **Description**: The number of times a signer gets frontrun by the original signer
- **Labels**:
  - chain_id
  - src_channel
  - src_port
  - dst_channel
  - dst_port
  - signer
  - frontrunned_by
  - memo
  - effected_memo
- **Example**: `ibc_frontrun_counter{...} 10`

#### ibc_stuck_packets
- **Type**: gauge
- **Description**: The number of packets stuck on an IBC channel
- **Labels**:
  - src_chain
  - dst_chain (actually dst_channel in current data)
  - src_channel
- **Example**: `ibc_stuck_packets{dst_chain="channel-0",src_chain="osmosis-1",src_channel="channel-165"} 6220`

#### ibc_stuck_packets_detailed
- **Type**: gauge
- **Description**: Detailed stuck packet tracking with user info
- **Labels**: Additional user/sender information
- **Example**: Not seen in current sample

#### ibc_packet_age_seconds
- **Type**: gauge/histogram
- **Description**: Age of unrelayed packets in seconds
- **Labels**: Various packet identifiers
- **Example**: Not seen in current sample

## Observed Data Patterns

### Active Chains
Based on the metrics, Chainpulse is monitoring:
1. cosmoshub-4 (Cosmos Hub)
2. osmosis-1 (Osmosis)
3. axelar-dojo-1 (Axelar)
4. noble-1 (Noble)
5. stride-1 (Stride)
6. jackal-1 (Jackal)

### Connection Issues
Some chains showing errors and reconnects:
- jackal-1: 32 errors, 32 reconnects
- stride-1: 30 errors, 30 reconnects
- noble-1: 1 reconnect
- osmosis-1: 1 reconnect

### Stuck Packets
Major congestion observed:
- osmosis-1 → channel-165: 6,220 stuck packets
- osmosis-1 → channel-141: 362 stuck packets
- Multiple other channels with lower stuck packet counts

### Active Relayers
Multiple professional relayers identified in memo fields:
- Lavender.Five Nodes
- SynergyNodes.com
- cosmosrescue.dev
- Polkachu
- CroutonDigital
- Inter Blockchain Services
- Qubelabs.io
- Rakoff
- staketown-relayer
- Apeiron Nodes
- Jackal Labs

## API Endpoints Beyond Metrics

Based on the API_INTERFACES.md documentation, Chainpulse should also provide:
- GET /health - Health check endpoint
- Additional JSON API endpoints through the Relayooor API facade

## Integration Notes

1. **Metric Format**: All metrics follow Prometheus format conventions
2. **Labels**: Extensive labeling allows for detailed filtering and aggregation
3. **Real-time Data**: Metrics are updated in real-time as packets are processed
4. **Relayer Identification**: Memo field contains relayer service names and versions
5. **Channel Identification**: Stuck packets tracked by source chain and channel

## Next Steps

1. Verify JSON API endpoints (if available)
2. Test WebSocket connections for real-time updates
3. Document data refresh rates and latency
4. Map metric relationships to frontend requirements