# Neutron Chain Compatibility Issue

## Problem Description

Neutron chain shows as "degraded" or causes errors in chainpulse because it uses Slinky oracle vote extensions that the current version of chainpulse cannot decode.

## Technical Details

Starting from block `12255559` on mainnet (`16308954` on testnet), Neutron includes Slinky oracle data in the first "transaction" of each block. This is not actually a transaction but compressed protobuf data containing oracle price information.

### Structure of Neutron Blocks with Slinky Data

1. **First transaction**: Contains vote extensions (VE) - protobuf encoded `ExtendedCommitInfo` compressed with zstd
2. **Each vote extension**: Contains oracle prices reported by validators
3. **Price data**: GOB-encoded big.Int deltas (differences from previous block)

### Decoding Process Required

To properly handle Neutron blocks, chainpulse would need to:

1. Detect if the first transaction is actually vote extension data
2. Decompress using zstd algorithm
3. Decode the protobuf structure (`ExtendedCommitInfo`)
4. For each vote extension:
   - Decompress with zstd
   - Decode `OracleVoteExtension` protobuf
   - Handle GOB-encoded price deltas

## Current Workaround

We've created an alternative configuration file `chainpulse-selected-no-neutron.toml` that excludes Neutron from monitoring. This prevents chainpulse from crashing while still monitoring other major chains.

To use this configuration:

```bash
# Update docker-compose.yml to use the no-neutron config
volumes:
  - ./config/chainpulse-selected-no-neutron.toml:/config/chainpulse.toml:ro
```

## Long-term Solutions

### Option 1: Update Chainpulse Fork

Modify the chainpulse code to handle Slinky vote extensions:

1. Add detection for Neutron chain ID
2. Implement special handling for first transaction when chain is Neutron
3. Add zstd decompression support
4. Parse ExtendedCommitInfo without treating it as a regular transaction

### Option 2: Proxy/Filter Approach

Create a proxy service that:
1. Intercepts Neutron WebSocket connections
2. Filters out or modifies the problematic first transaction
3. Passes cleaned data to chainpulse

### Option 3: Alternative Monitoring

Use a different monitoring solution specifically for Neutron that understands Slinky data.

## References

- [Slinky Oracle Documentation](https://github.com/skip-mev/connect)
- [ExtendedCommitInfo Proto](https://github.com/cometbft/cometbft/blob/e1b4453baf0af6487ad187c7f17dc50517126673/proto/tendermint/abci/types.proto#L379)
- [OracleVoteExtension Proto](https://github.com/skip-mev/connect/blob/fc1f860af86ee873253bb9471a01831be5f319bc/proto/connect/abci/v2/vote_extensions.proto#L7)

## Impact on Relayooor

- Neutron chain will show as unavailable in monitoring dashboards
- IBC packets to/from Neutron won't be tracked by chainpulse
- Manual monitoring or alternative solutions needed for Neutron routes

## Temporary UI Adjustments

The webapp has been updated to show Neutron as "degraded" with an explanation about the compatibility issue rather than showing errors.