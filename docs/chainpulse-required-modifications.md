# Chainpulse Required Modifications for Packet Clearing

## Overview

This document outlines the minimal modifications required for Chainpulse to support the packet clearing feature. These changes focus on enhancing data collection to identify packet senders/receivers and track stuck packets, while maintaining compatibility with future IBC v2/Eureka support.

## Core Requirements

### 1. Enhanced Packet Data Collection

**Current State:** Chainpulse tracks packet relay attempts but doesn't parse packet data to identify senders/receivers.

**Required Change:** Parse IBC packet data field to extract user addresses for fungible token transfers.

#### Implementation in `src/msg.rs`

Add packet data parsing:

```rust
use serde::{Deserialize, Serialize};

/// IBC Fungible Token Transfer packet data structure
/// This structure is standard across IBC v1 and will have a compatibility layer in IBC v2
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FungibleTokenPacketData {
    pub denom: String,
    pub amount: String,
    pub sender: String,
    pub receiver: String,
    #[serde(default)]
    pub memo: String,
}

/// Enhanced packet info that works for both IBC v1 and future v2
#[derive(Debug, Clone)]
pub struct UniversalPacketInfo {
    pub sequence: u64,
    pub source_channel: String,
    pub destination_channel: String,
    pub source_port: String,
    pub destination_port: String,
    pub timeout_timestamp: Option<u64>,
    pub timeout_height: Option<Height>,
    
    // User data (when available)
    pub sender: Option<String>,
    pub receiver: Option<String>,
    pub amount: Option<String>,
    pub denom: Option<String>,
    
    // Version info for future compatibility
    pub ibc_version: String, // "v1" or "v2"
}
```

#### Modify packet processing in `src/collect.rs`

Update the `process_packet` function:

```rust
// In the existing packet processing logic, add:
fn extract_packet_user_data(packet: &Packet) -> (Option<String>, Option<String>, Option<String>, Option<String>) {
    // Only parse transfer port packets
    if packet.source_port != "transfer" {
        return (None, None, None, None);
    }
    
    // Try to parse as fungible token data
    match serde_json::from_slice::<FungibleTokenPacketData>(&packet.data) {
        Ok(ft_data) => (
            Some(ft_data.sender),
            Some(ft_data.receiver),
            Some(ft_data.denom),
            Some(ft_data.amount),
        ),
        Err(_) => (None, None, None, None),
    }
}
```

### 2. Database Schema Additions

Add columns to store user data without breaking existing functionality:

```sql
-- Add to existing packets table (backward compatible)
ALTER TABLE packets ADD COLUMN sender TEXT;
ALTER TABLE packets ADD COLUMN receiver TEXT;
ALTER TABLE packets ADD COLUMN denom TEXT;
ALTER TABLE packets ADD COLUMN amount TEXT;
ALTER TABLE packets ADD COLUMN ibc_version TEXT DEFAULT 'v1';

-- Create indexes for efficient user queries
CREATE INDEX idx_packets_sender ON packets(sender) WHERE sender IS NOT NULL;
CREATE INDEX idx_packets_receiver ON packets(receiver) WHERE receiver IS NOT NULL;
CREATE INDEX idx_packets_pending_by_sender ON packets(sender, effected) WHERE effected = false;
```

### 3. Enhanced Metrics for Stuck Packet Detection

Update `src/metrics.rs` to properly track stuck packets:

```rust
lazy_static! {
    /// Actual stuck packets gauge (currently not implemented)
    pub static ref IBC_STUCK_PACKETS_DETAILED: IntGaugeVec = register_int_gauge_vec!(
        "ibc_stuck_packets_detailed",
        "Detailed stuck packet tracking with user info",
        &["src_chain", "dst_chain", "src_channel", "dst_channel", "has_user_data"]
    ).unwrap();
    
    /// Time since packet creation for unrelayed packets
    pub static ref IBC_PACKET_AGE_UNRELAYED: GaugeVec = register_gauge_vec!(
        "ibc_packet_age_seconds",
        "Age of unrelayed packets in seconds",
        &["src_chain", "dst_chain", "channel"]
    ).unwrap();
}
```

### 4. Status Endpoint Enhancement

Modify `src/status.rs` to expose packet details via HTTP API:

```rust
/// Enhanced status response that includes user-queryable packet data
#[derive(Serialize)]
pub struct EnhancedStatus {
    pub chains: Vec<ChainStatus>,
    pub packets: PacketStats,
    pub stuck_packets: StuckPacketsSummary,
    pub ibc_version_support: Vec<String>, // ["v1", "v2-ready"]
}

#[derive(Serialize)]
pub struct StuckPacketsSummary {
    pub total_stuck: u64,
    pub by_channel: HashMap<String, u64>,
    pub oldest_stuck_age_seconds: Option<u64>,
    pub queryable_packets: u64, // Packets with sender/receiver data
}

// Add new endpoint for packet queries
pub async fn handle_packet_query(query: PacketQuery) -> Result<Vec<PacketInfo>> {
    // Query packets by sender/receiver address
    // This endpoint will be used by the Relayooor API
}
```

### 5. Configuration for IBC v2 Readiness

Update `chainpulse.toml` format to support version configuration:

```toml
[global]
# IBC version support
ibc_versions = ["v1"]  # Will become ["v1", "v2"] when ready

[chains.osmosis-1]
url = "wss://rpc.osmosis.zone/websocket"
comet_version = "0.37"
ibc_version = "v1"  # Per-chain version override

# Future IBC v2 chain example (not implemented yet)
# [chains.eureka-testnet-1]
# url = "wss://eureka-rpc.example.com/websocket"
# comet_version = "0.38"
# ibc_version = "v2"
```

## API Endpoints to Add

Add these REST endpoints to Chainpulse for the Relayooor API to query:

```
GET /api/v1/packets/by-user?address={address}&role={sender|receiver|both}
GET /api/v1/packets/stuck?min_age_seconds=900
GET /api/v1/packets/{chain}/{channel}/{sequence}
GET /api/v1/channels/congestion
```

Example response format:
```json
{
  "packets": [
    {
      "chain_id": "osmosis-1",
      "sequence": 12345,
      "src_channel": "channel-0",
      "dst_channel": "channel-141",
      "sender": "osmo1abc...",
      "receiver": "cosmos1xyz...",
      "amount": "1000000",
      "denom": "uosmo",
      "age_seconds": 1820,
      "relay_attempts": 3,
      "last_attempt_by": "osmo1relayer...",
      "ibc_version": "v1"
    }
  ],
  "total": 42,
  "api_version": "1.0"
}
```

## Migration Path for IBC v2

When IBC v2/Eureka support is added:

1. **Packet Structure Abstraction:**
   ```rust
   pub trait PacketData {
       fn get_sender(&self) -> Option<String>;
       fn get_receiver(&self) -> Option<String>;
       fn get_amount(&self) -> Option<String>;
       fn get_denom(&self) -> Option<String>;
   }
   
   impl PacketData for FungibleTokenPacketDataV1 { ... }
   impl PacketData for FungibleTokenPacketDataV2 { ... }
   ```

2. **Version Detection:**
   ```rust
   fn detect_packet_version(data: &[u8]) -> IbcVersion {
       // Check packet structure to determine version
       // V2 packets will have different structure
   }
   ```

3. **Unified Processing:**
   ```rust
   fn process_any_packet(packet: &Packet) -> UniversalPacketInfo {
       match detect_packet_version(&packet.data) {
           IbcVersion::V1 => process_v1_packet(packet),
           IbcVersion::V2 => process_v2_packet(packet),
       }
   }
   ```

## Minimal Implementation Checklist

1. Parse fungible token packet data to extract sender/receiver
2. Add database columns for user addresses (nullable for non-transfer packets)
3. Create indexes for efficient user queries
4. Implement stuck packet detection (packets pending > 15 minutes)
5. Add HTTP endpoints for packet queries
6. Update metrics to include stuck packet details

## Testing the Modifications

Test queries to verify the implementation:

```sql
-- Find packets for a specific user
SELECT * FROM packets 
WHERE sender = 'osmo1abc...' OR receiver = 'osmo1abc...'
ORDER BY created_at DESC;

-- Find stuck packets
SELECT chain_id, sequence, src_channel, sender, receiver, amount, denom,
       (strftime('%s', 'now') - strftime('%s', created_at)) as age_seconds
FROM packets
WHERE effected = 0 
  AND created_at < datetime('now', '-15 minutes')
  AND sender IS NOT NULL;

-- Channel congestion analysis
SELECT src_channel, dst_channel, 
       COUNT(*) as stuck_count,
       MIN(created_at) as oldest_stuck
FROM packets
WHERE effected = 0 
  AND created_at < datetime('now', '-15 minutes')
GROUP BY src_channel, dst_channel;
```

## Performance Considerations

1. **Parsing Overhead:** Only parse packet data for transfer port packets
2. **Index Strategy:** Indexes on sender/receiver are partial (WHERE NOT NULL)
3. **Query Limits:** Enforce reasonable limits on API queries (max 1000 packets)
4. **Caching:** Consider caching user packet queries for 30 seconds

## Security Notes

1. **Address Validation:** Validate queried addresses are valid bech32
2. **Rate Limiting:** Implement per-IP rate limiting on query endpoints
3. **Data Privacy:** Consider whether all packet data should be publicly queryable

## No Hermes Modifications Required

The packet clearing feature will use Hermes's existing REST API:
- `GET /clear-packets` - Clear specific packets
- `POST /clear-packets` - Clear packets matching criteria

No modifications to Hermes are required as the existing API is sufficient.

## Summary

These minimal modifications to Chainpulse will:
1. Enable user-centric packet queries for the clearing feature
2. Properly track stuck packets with user context
3. Maintain backward compatibility
4. Prepare for seamless IBC v2 integration
5. Keep the user experience consistent regardless of IBC version

The key design principle is to abstract IBC version differences in Chainpulse and the API layer, so the frontend remains simple and consistent for users.