# Legacy Version Support

This guide explains how to use legacy versions of Hermes and the Go relayer for networks that require older IBC implementations.

## Available Versions

### Current (Default)
- **Hermes**: Latest from the repository
- **Go Relayer**: Latest from the repository

### Legacy
- **Hermes**: v0.15.0 (IBC v2 compatible)
- **Go Relayer**: v2.1.2 (Cosmos SDK v0.45 compatible)

## Using Legacy Versions

### 1. Environment Variable Method

Set the `RELAYER_VERSION` environment variable:

```bash
# Use legacy versions
export RELAYER_VERSION=legacy
docker-compose -f docker-compose.yml -f docker-compose.legacy.yml up

# Use current versions (default)
export RELAYER_VERSION=current
docker-compose up
```

### 2. Per-Chain Configuration

You can configure specific chains to use legacy versions by updating the API configuration:

```json
{
  "chains": {
    "cosmoshub-3": {
      "relayer": "hermes",
      "version": "legacy"
    },
    "osmosis-1": {
      "relayer": "rly",
      "version": "current"
    }
  }
}
```

### 3. Runtime Switching

The dashboard API supports switching versions at runtime:

```bash
# Switch to legacy Hermes
curl -X POST http://localhost:3000/relayer/version \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"relayer": "hermes", "version": "legacy"}'
```

## Configuration Differences

### Hermes v0.15.0 vs Current

Key differences in configuration:

1. **Gas Configuration**:
   - Legacy: Uses fixed gas prices
   - Current: Supports dynamic gas pricing

2. **Event Source**:
   - Legacy: WebSocket only
   - Current: WebSocket + RPC polling

Example legacy config snippet:
```toml
[chains.gas_price]
price = 0.025
denom = 'uatom'
```

### Go Relayer v2.1.2 vs Current

Key differences:

1. **Path Configuration**:
   - Legacy: YAML format
   - Current: JSON format with enhanced options

2. **Fee Middleware**:
   - Legacy: Not supported
   - Current: Full ICS-29 fee support

## Network Compatibility

### Networks Requiring Legacy Versions

Some networks may require legacy relayer versions:

| Network | Required Version | Reason |
|---------|-----------------|--------|
| Terra Classic | Hermes v0.15.0 | IBC v2 compatibility |
| Cosmos Hub (pre-v7) | Go Relayer v2.1.2 | SDK compatibility |
| Older Osmosis | Hermes v0.15.0 | Event format |

### Testing Compatibility

Before switching versions, test the connection:

```bash
# Test with legacy Hermes
docker exec relayer-dashboard /usr/local/bin/relayer-selector hermes health-check

# Test with current Go relayer
docker exec relayer-dashboard /usr/local/bin/relayer-selector rly chains list
```

## Troubleshooting

### Common Issues

1. **Event Format Mismatch**:
   - Symptom: Relayer not detecting events
   - Solution: Switch to legacy version for older chains

2. **Gas Estimation Failures**:
   - Symptom: Transactions failing with out-of-gas
   - Solution: Use legacy version with fixed gas prices

3. **Client Update Errors**:
   - Symptom: "client state not found" errors
   - Solution: Ensure version matches chain's IBC version

### Debug Mode

Enable debug logging for version issues:

```bash
# In docker-compose.yml
environment:
  - RUST_LOG=debug  # For Hermes
  - RLY_DEBUG=true  # For Go relayer
```

## Migration Guide

### Upgrading from Legacy to Current

1. **Backup Configurations**:
   ```bash
   cp -r config/hermes config/hermes.backup
   cp -r config/relayer config/relayer.backup
   ```

2. **Update Chain Configs**:
   - Add dynamic gas configuration
   - Update event source settings
   - Verify chain IDs and endpoints

3. **Test on Testnet**:
   - Always test version changes on testnet first
   - Monitor for packet relay issues
   - Verify client updates work correctly

4. **Gradual Migration**:
   - Migrate one chain at a time
   - Monitor metrics after each change
   - Keep legacy configs as fallback

## API Endpoints for Version Management

### Get Current Versions
```bash
GET /api/relayer/versions
```

### Set Relayer Version
```bash
POST /api/relayer/version
{
  "relayer": "hermes",
  "version": "legacy"
}
```

### Get Version Compatibility
```bash
GET /api/chains/:chain_id/compatibility
```

## Best Practices

1. **Document Version Requirements**: Keep a record of which chains require which versions
2. **Monitor After Switching**: Watch logs and metrics after version changes
3. **Keep Configs Separate**: Maintain separate config directories for different versions
4. **Test Thoroughly**: Always test version switches in a non-production environment first