# Configuration Guide

This document describes how configuration is managed in the Relayooor platform to avoid hardcoded values and enable easy deployment across different environments.

## Overview

The configuration system follows these principles:
1. **No hardcoded chains** - Chain data loaded from API
2. **Environment-based URLs** - Service endpoints configured via environment variables
3. **Runtime discovery** - Services discover each other dynamically
4. **Centralized registry** - Single source of truth for chain/channel data

## Configuration Layers

### 1. Chain Registry (API)

The API provides a central registry of chain and channel configurations at `/api/config`:

```json
{
  "chains": {
    "osmosis-1": {
      "chain_id": "osmosis-1",
      "chain_name": "Osmosis",
      "address_prefix": "osmo",
      "explorer": "https://www.mintscan.io/osmosis/txs",
      "logo": "/images/chains/osmosis.svg"
    }
  },
  "channels": [
    {
      "source_chain": "osmosis-1",
      "source_channel": "channel-0",
      "dest_chain": "cosmoshub-4",
      "dest_channel": "channel-141",
      "source_port": "transfer",
      "dest_port": "transfer",
      "status": "active"
    }
  ]
}
```

### 2. Environment Variables

#### Backend (.env)
```bash
# Service URLs
CHAINPULSE_METRICS_URL=http://relayooor-chainpulse-1:3001
HERMES_API_URL=http://hermes:5185

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=relayooor
DB_PASSWORD=secure_password
DB_NAME=relayooor

# Redis
REDIS_ADDR=redis:6379

# RPC Credentials
RPC_USERNAME=skip
RPC_PASSWORD=secure_password

# WebSocket URLs for chains
COSMOS_WS_URL=wss://cosmoshub-4-skip-rpc.polkachu.com/websocket
OSMOSIS_WS_URL=wss://osmosis-1-skip-rpc.polkachu.com/websocket
NEUTRON_WS_URL=wss://neutron-1-skip-rpc.polkachu.com/websocket
NOBLE_WS_URL=wss://noble-1-skip-rpc.polkachu.com/websocket
```

#### Frontend (.env)
```bash
# API Configuration
VITE_API_URL=  # Empty = use relative URLs
VITE_API_VERSION=v1

# Service URLs
VITE_CHAINPULSE_URL=http://localhost:3001
VITE_HERMES_URL=http://localhost:5185

# WebSocket
VITE_WS_HOST=  # Empty = use current host

# Features
VITE_MOCK_DATA=false
VITE_DEBUG=false
VITE_ANALYTICS_ENABLED=true

# Defaults
VITE_DEFAULT_CHAIN=osmosis-1
```

### 3. Docker Configuration

#### docker-compose.yml
```yaml
services:
  api:
    environment:
      - CHAINPULSE_METRICS_URL=http://chainpulse:3001
      - DB_HOST=postgres
      - REDIS_ADDR=redis:6379
    
  webapp:
    environment:
      - VITE_API_URL=/api
```

### 4. Runtime Configuration

#### Frontend Config Service
```typescript
// webapp/src/services/config.ts
class ConfigService {
  async loadRegistry(): Promise<ChainRegistry>
  async getChain(chainId: string): Promise<ChainConfig>
  async getChainByPrefix(prefix: string): Promise<ChainConfig>
  async getChannels(chainId?: string): Promise<ChannelConfig[]>
  getExplorerUrl(chainId: string, txHash: string): string
}
```

#### Backend Chain Registry
```go
// api/internal/config/chains.go
type ChainRegistry struct {
  Chains   map[string]ChainConfig
  Channels []ChannelConfig
}
```

## Adding New Chains

### 1. Add to Chainpulse Config
```toml
# config/chainpulse-selected.toml
[chains.newchain-1]
url = "${NEWCHAIN_WS_URL}"
comet_version = "0.37"
username = "${RPC_USERNAME}"
password = "${RPC_PASSWORD}"
```

### 2. Add Environment Variable
```bash
# .env
NEWCHAIN_WS_URL=wss://newchain-1-rpc.example.com/websocket
```

### 3. Update API Registry
The API will automatically detect the new chain from chainpulse metrics. To add metadata:

```go
// api/cmd/server/main.go - getChainRegistry()
{
  "chain_id": "newchain-1",
  "chain_name": "New Chain",
  "address_prefix": "new",
  "explorer": "https://explorer.newchain.com/tx",
  "logo": "/images/chains/newchain.svg"
}
```

### 4. Frontend Auto-Discovery
The frontend will automatically:
- Load chain from API config endpoint
- Display in chain selectors
- Use proper explorer URLs
- Handle address prefix detection

## Port Configuration

All ports are configurable via environment:

| Service | Default Port | Environment Variable |
|---------|-------------|---------------------|
| Web App | 80 | WEBAPP_PORT |
| API | 8080 | API_PORT |
| Chainpulse | 3001 | CHAINPULSE_PORT |
| Hermes | 5185 | HERMES_PORT |
| Grafana | 3003 | GRAFANA_PORT |
| Prometheus | 9090 | PROMETHEUS_PORT |

## Service Discovery

### Local Development
Services use localhost with default ports:
- API: http://localhost:8080
- Chainpulse: http://localhost:3001
- Webapp: http://localhost:80

### Docker Environment
Services use container names:
- API: http://relayooor-api:8080
- Chainpulse: http://relayooor-chainpulse-1:3001
- Webapp: http://relayooor-webapp:80

### Production
Services should use:
- Internal DNS names
- Service mesh discovery
- Load balancer endpoints

## Configuration Best Practices

1. **Never hardcode chain IDs** - Use API registry
2. **Use relative URLs** - Let proxies handle routing
3. **Environment-specific defaults** - Different for dev/prod
4. **Validate at startup** - Check required config exists
5. **Provide fallbacks** - Graceful degradation
6. **Document all variables** - In .env.example
7. **Use structured config** - Not scattered constants

## Migration Guide

To migrate from hardcoded values:

1. **Identify hardcoded values**
   ```bash
   grep -r "cosmoshub-4\|osmosis-1" webapp/src
   grep -r "localhost:3001\|localhost:8080" .
   ```

2. **Replace with config calls**
   ```typescript
   // Before
   const chain = 'osmosis-1'
   
   // After
   const chain = await configService.getChain('osmosis-1')
   ```

3. **Use environment variables**
   ```typescript
   // Before
   const url = 'http://localhost:8080'
   
   // After
   const url = config.getApiUrl()
   ```

4. **Test all environments**
   - Local development
   - Docker compose
   - Production deployment