# Chainpulse Packet Data Analysis

Date: 2025-01-19

## Current State

### What Chainpulse IS Tracking

1. **Packet Routing Information**:
   - Source channel & port
   - Destination channel & port
   - Sequence number

2. **Relayer Information**:
   - Signer address (the relayer who submitted the packet)
   - Memo field (relayer identification)
   - Effectiveness (whether packet was successfully relayed or frontrun)

3. **Transaction Metadata**:
   - Transaction hash
   - Block height
   - Chain ID

### What Chainpulse is NOT Tracking

1. **Packet Data Content**:
   - Sender address (original user who initiated the transfer)
   - Receiver address (destination user)
   - Transfer amount
   - Token denomination
   - Any other packet-specific data

## The Missing Piece: Packet Data

According to the REST examples in REST-Examples.md, IBC packet data is available in transaction events:

```json
{
  "type": "send_packet",
  "attributes": [
    {
      "key": "packet_data",
      "value": "{\"amount\":\"2230211\",\"denom\":\"uusdc\",\"receiver\":\"osmo1wev8ptzj27aueu04wgvvl4gvurax6rj5p5lw4u\",\"sender\":\"noble1zw4x6cptan8au9ezpmknms3uf9eeydl3v2lvje\"}",
      "index": true
    }
  ]
}
```

This data is also available in the `fungible_token_packet` event type:
```json
{
  "type": "fungible_token_packet",
  "attributes": [
    {
      "key": "sender",
      "value": "sei1wev8ptzj27aueu04wgvvl4gvurax6rj5yrag90"
    },
    {
      "key": "receiver",
      "value": "noble1wev8ptzj27aueu04wgvvl4gvurax6rj5pvekmq"
    },
    {
      "key": "amount",
      "value": "2230211"
    },
    {
      "key": "denom",
      "value": "transfer/channel-45/uusdc"
    }
  ]
}
```

## Current Code Analysis

### 1. Message Parsing (msg.rs)
- Decodes IBC messages including `MsgTransfer`
- Extracts packet structure but not packet data content
- Only gets the `signer` (relayer), not sender/receiver

### 2. Database Schema (db.rs)
- No columns for sender, receiver, amount, or denom
- Only tracks routing and relayer information

### 3. Metrics (metrics.rs)
- No metrics include sender/receiver labels
- The `ibc_stuck_packets_detailed` metric mentioned in output doesn't exist in code

## Why the API Uses Mock Data

The simple API (/api) generates mock data for stuck packets because:
1. Chainpulse doesn't store sender/receiver information
2. The Prometheus metrics don't include this data
3. There's no way to retrieve this information without querying the blockchain

## Required Changes to Remove Mocks

### Option 1: Enhance Chainpulse (Recommended)

1. **Parse packet data in collect.rs**:
   ```rust
   // In process_packet function
   if let Some(packet_data) = parse_packet_data_from_events(events) {
       // Decode JSON packet data
       let data: FungibleTokenPacketData = serde_json::from_str(&packet_data)?;
       // Store sender, receiver, amount, denom
   }
   ```

2. **Update database schema**:
   ```sql
   ALTER TABLE packets ADD COLUMN sender TEXT;
   ALTER TABLE packets ADD COLUMN receiver TEXT;
   ALTER TABLE packets ADD COLUMN amount TEXT;
   ALTER TABLE packets ADD COLUMN denom TEXT;
   ```

3. **Add new metrics**:
   ```rust
   // New detailed metric with sender/receiver
   ibc_stuck_packets_detailed{
       src_chain,
       dst_chain,
       src_channel,
       sender,
       receiver,
       denom
   }
   ```

### Option 2: Query Chain REST APIs

Use the REST endpoints documented in REST-Examples.md to fetch packet details on-demand:

1. **For stuck packets**: Query `/ibc/core/channel/v1/channels/{channel_id}/ports/{port_id}/packet_commitments/{sequence}`
2. **For packet data**: Parse transaction events from chain REST API
3. **Cache results** in Redis to avoid repeated queries

### Option 3: Use Existing Chain Indexers

Instead of building our own indexer:
1. Use existing services like Mintscan API
2. Query indexed packet data from block explorers
3. Focus on packet clearing logic rather than indexing

## Recommendation

The best approach depends on requirements:

1. **If real-time monitoring is critical**: Enhance Chainpulse to parse and store packet data
2. **If on-demand queries are sufficient**: Use chain REST APIs with caching
3. **If minimizing development is priority**: Use existing indexer services

The current mock data serves as a placeholder until one of these solutions is implemented.