# Chain Integration Troubleshooting Guide

## Overview
This document provides guidance for troubleshooting chain integration issues with Chainpulse monitoring, particularly focusing on protobuf decoding errors and establishing a reliable process for adding new chains.

## Known Issues

### Neutron-1 Protobuf Decoding Error
**Symptom:** 
```
ERROR collect{chain=neutron-1}: failed to decode Protobuf message: invalid tag value: 0
```

**Status:** Under investigation
- Connection to WebSocket succeeds
- Authentication works properly
- Blocks are received but fail during protobuf decoding
- Tried both comet_version 0.37 and 0.38 with same error

**Impact:**
- Chain shows as "degraded" in monitoring
- No packet/transaction data collected for Neutron
- Other chains (Cosmos Hub, Osmosis, Noble) work correctly

**Investigation Results:**
- Neutron runs CometBFT version 0.38.17 (confirmed via RPC status)
- Application version: neutrond 6.0.3
- Protocol version: block=11, app=0
- WebSocket connection and authentication succeed
- Block data is received but fails protobuf decoding

**Root Cause:** 
Neutron uses ABCI++ vote extensions through the Slinky oracle implementation. This adds extra data to blocks that standard Cosmos chains don't have:
- Vote extensions are part of ABCI 2.0 (Cosmos SDK 0.50+ and CometBFT 0.38+)
- Slinky is an enshrined oracle that uses vote extensions for validator consensus on price data
- The extended commit info in blocks contains additional protobuf fields that chainpulse doesn't know how to decode
- The "invalid tag value: 0" error occurs when chainpulse tries to parse these vote extension fields

**Technical Details:**
- Neutron version: 6.0.3 with CometBFT 0.38.17
- Uses vote extensions from block n-1 to submit oracle price data for block n
- Recent Neutron releases (v6.0.1-rpc, v6.0.3) addressed various RPC/query issues

**Workaround:**
Until chainpulse is updated to handle ABCI++ vote extensions, Neutron will show as "degraded" but other monitoring continues to function.

**Potential Solutions:**
1. Update chainpulse to handle ABCI++ vote extensions (requires understanding Slinky protobuf schemas)
2. Use a different WebSocket event subscription that doesn't include vote extension data
3. Fork chainpulse and add custom decoder for Neutron blocks
4. Wait for Neutron team to provide RPC compatibility layer

**References:**
- Neutron GitHub: https://github.com/neutron-org/neutron
- Slinky Oracle (vote extensions): Used for price feed consensus
- ABCI++ Specification: https://docs.cometbft.com/main/spec/abci/

## Process for Adding New Chains

### 1. Prerequisites
- Chain must have a WebSocket RPC endpoint
- Authentication credentials (if required)
- Knowledge of the chain's Tendermint/CometBFT version

### 2. Configuration Steps

#### Step 1: Add WebSocket URL to .env
```bash
# Add to /Users/cordt/repos/relayooor/.env
CHAINNAME_WS_URL=wss://chainname-skip-rpc.polkachu.com/websocket
```

#### Step 2: Update chainpulse configuration
```toml
# Add to /Users/cordt/repos/relayooor/config/chainpulse-selected.toml
[chains.chainname-1]
url = "${CHAINNAME_WS_URL}"
comet_version = "0.37"  # Try 0.37 first, then 0.38 if issues
username = "${RPC_USERNAME}"
password = "${RPC_PASSWORD}"
```

#### Step 3: Add chain display name to API
```go
// Update chainNames map in api/cmd/server/main.go
chainNames := map[string]string{
    "cosmoshub-4": "Cosmos Hub",
    "osmosis-1":   "Osmosis",
    "neutron-1":   "Neutron",
    "noble-1":     "Noble",
    "chainname-1": "Chain Display Name",  // Add here
}
```

### 3. Testing Process

#### Quick Add Using Makefile
```bash
# Use the add-chain command
make -f Makefile.docker add-chain CHAIN=akash-1 WS=wss://akash-1-skip-rpc.polkachu.com/websocket
```

#### Manual Testing Steps
1. Restart chainpulse: `make -f Makefile.docker chainpulse-restart`
2. Check logs: `make -f Makefile.docker logs-chainpulse`
3. Verify metrics: `make -f Makefile.docker check-chains`
4. Monitor for errors: `docker logs -f relayooor-chainpulse-1 | grep "chain_id"`

### 4. Common Troubleshooting

#### Determining Correct comet_version
1. Query the chain's status endpoint:
```bash
curl -s https://chainname-rpc.example.com/status | jq '.result.node_info.version'
```

2. Version mapping:
- Tendermint v0.34.x → comet_version = "0.34"
- CometBFT v0.37.x → comet_version = "0.37"  
- CometBFT v0.38.x → comet_version = "0.38"

#### Common Error Patterns

**WebSocket connection issues:**
- Check authentication credentials
- Verify WebSocket URL format
- Test basic connectivity: `curl -u 'username:password' https://rpc-url/status`

**Protobuf decoding errors:**
- Try different comet_version values
- Check if chain uses custom ABCI extensions
- Look for chain-specific documentation on RPC compatibility

**No data appearing:**
- Verify chain is producing IBC traffic
- Check if metrics are being exposed: `curl http://localhost:3001/metrics | grep chain_id`
- Ensure chain ID matches exactly (case-sensitive)

### 5. Monitoring Chain Health

After adding a chain, monitor its health:
```bash
# Check overall status
make -f Makefile.docker status

# Check specific chain metrics
curl -s http://localhost:3001/metrics | grep -E "chainpulse_(packets|txs|errors){chain_id=\"chainname-1\"}"

# View in web interface
open http://localhost/monitoring
```

## Dynamic Chain Detection

The system is designed to automatically detect and display chains in the frontend:
- Chainpulse exposes metrics for all configured chains
- API dynamically parses these metrics and converts to structured data
- Frontend queries `/api/monitoring/data` to get all available chains
- No frontend changes needed when adding new chains

## Known Chain Compatibility Issues

### Chains with Vote Extensions (ABCI++)
These chains use vote extensions and may not work with standard chainpulse:
- **Neutron** - Uses Slinky oracle for price feeds
- **dYdX** - Custom implementation with oracle functionality
- **Any chain using Slinky oracle** - Will have similar protobuf decoding issues

### Chains with Custom Implementations
These chains may require special handling:
- **Injective** - Custom Tendermint fork
- **Sei** - Performance optimizations may affect compatibility
- **Evmos** - EVM integration may introduce unique challenges

### Compatibility Check Script
Use the provided script to check chain compatibility before adding:
```bash
./scripts/check-chain-compatibility.sh
```

This script will:
- Query chain version and protocol information
- Identify potential compatibility issues
- Warn about known problematic implementations

## Future Improvements

1. **ABCI++ Support**: Add vote extension handling to chainpulse
2. **Auto-detection of comet_version**: Query chain status endpoint to determine version
3. **Chain registry integration**: Pull chain metadata from cosmos chain registry
4. **Error recovery**: Implement automatic reconnection with version fallback
5. **Metrics dashboard**: Add Grafana dashboard for chain health monitoring
6. **Compatibility matrix**: Maintain a list of tested chains with their compatibility status