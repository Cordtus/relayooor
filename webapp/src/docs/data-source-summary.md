# Data Source Summary

## Available Data Sources

### 1. Configuration API (`/api/config`)
- **Provides**: Chain registry, channel mappings, explorer URLs
- **Usage**: Load once on app startup, cache in config service
- **Components using**: Dashboard, Monitoring, Settings

### 2. Monitoring Data (`/api/monitoring/data`)
- **Provides**: System overview, chains, channels, relayers
- **Usage**: Real-time monitoring data
- **Refresh**: 10 seconds (configurable)

### 3. Metrics Endpoints
- `/api/v1/chainpulse/metrics` - Raw Prometheus metrics
- `/api/monitoring/metrics` - Structured metrics
- **Usage**: Fallback when structured data unavailable

### 4. Channel Congestion (`/api/channels/congestion`)
- **Provides**: Stuck packets by channel
- **Refresh**: 30 seconds
- **Duplicate of**: Data available in monitoring endpoint

### 5. Platform Statistics (`/api/v1/platform/statistics`)
- **Provides**: Global platform metrics
- **Usage**: Analytics page
- **Refresh**: 30 seconds

## Data Issues Found

### Duplicate Data Sources
1. **Stuck Packets** - Available from:
   - `/api/v1/packets/stuck`
   - `/api/v1/packets/stuck/stream`
   - `/api/v1/packets/stuck/changes`
   - `/api/channels/congestion`
   - `/api/monitoring/data`

2. **Channel Information** - Available from:
   - `/api/monitoring/data`
   - `/api/v1/chainpulse/metrics`
   - `/api/channels/congestion`

### Hardcoded Values Still Present
1. **Chain inferrence from channel IDs** (Dashboard.vue:191)
   - `channel-141` → cosmoshub-4
   - Need channel-to-chain mapping from config

2. **Default denom mappings** (Settings.vue:267-269)
   - Still hardcoded in Settings
   - Should come from chain config

3. **Time ranges** in various components
   - Now using constants but still static
   - Could be configurable

## Recommendations

### 1. Consolidate Data Sources
- Use `/api/monitoring/data` as primary source
- Remove duplicate endpoints
- Cache appropriately

### 2. Complete Config Migration
- Add denom info to chain registry
- Add channel-to-chain mappings
- Move all chain-specific logic to config

### 3. Create Data Hooks
```typescript
// Suggested patterns
useChainData() - Load and cache chain config
useMonitoringData() - Real-time monitoring
useChannelData() - Channel-specific data
useAnalytics() - Analytics data with time range
```

### 4. Standardize Refresh Patterns
- Use constants everywhere
- Allow user configuration
- Implement smart refresh (pause when hidden)

## Next Steps

1. Add missing data to chain registry:
   - Default denoms
   - Clearing fees
   - Channel ownership

2. Create composables for common data patterns

3. Remove duplicate API calls

4. Implement proper caching strategy