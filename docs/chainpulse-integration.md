# Chainpulse Integration Documentation

## Overview
This document describes the integration of Chainpulse monitoring service with the Relayooor packet clearing system.

## Architecture

### Components
1. **Chainpulse Service** (`../chainpulse`)
   - Runs on port 3001
   - Provides Prometheus metrics at `/metrics`
   - Exposes API endpoints for packet queries

2. **API Backend** (`/api`)
   - Simple Go HTTP server on port 8080
   - Integrates with Chainpulse for user packet data
   - Falls back to mock data if Chainpulse unavailable

3. **Relayer Middleware** (`/relayer-middleware/api`)
   - Advanced Go API with database support
   - Chainpulse client service at `/pkg/chainpulse/client.go`
   - Handler at `/pkg/handlers/chainpulse.go`

4. **Frontend** (`/webapp`)
   - Vue.js application
   - Packets service at `/src/services/packets.ts`
   - Real-time integration with wallet connections

## API Endpoints

### Chainpulse Native Endpoints
- `GET /metrics` - Prometheus metrics
- `GET /api/v1/packets/by-user?address={address}` - User's packets
- `GET /api/v1/packets/stuck?min_stuck_minutes={mins}` - Stuck packets
- `GET /api/v1/packets/{chain}/{channel}/{sequence}` - Packet details
- `GET /api/v1/channels/congestion` - Channel congestion status

### API Backend Integration
- `GET /api/packets/stuck` - Global stuck packets (uses Chainpulse)
- `GET /api/user/{wallet}/transfers` - User transfers (uses Chainpulse)
- `GET /api/user/{wallet}/stuck` - User's stuck packets

### Relayer Middleware Integration
- `GET /api/v1/chainpulse/packets/by-user` - Proxied from Chainpulse
- `GET /api/v1/chainpulse/packets/stuck` - Proxied from Chainpulse
- `GET /api/v1/chainpulse/channels/congestion` - Proxied from Chainpulse
- `GET /api/v1/chainpulse/metrics` - Raw Prometheus metrics
- `GET /api/v1/chainpulse/health` - Health check

## Configuration

### Docker Compose
```yaml
chainpulse:
  build:
    context: ./monitoring
    dockerfile: Dockerfile
  environment:
    - RPC_USERNAME=${RPC_USERNAME}
    - RPC_PASSWORD=${RPC_PASSWORD}
  ports:
    - "3002:3001"  # Mapped to avoid conflicts
```

### Environment Variables
- `CHAINPULSE_URL` - URL of Chainpulse service (default: `http://localhost:3001`)
- `RPC_USERNAME` - RPC authentication username
- `RPC_PASSWORD` - RPC authentication password

## Data Flow

1. **User connects wallet** → Frontend detects wallet address
2. **Frontend requests user packets** → API backend queries Chainpulse
3. **Chainpulse returns packet data** → Includes stuck status, age, attempts
4. **Frontend displays packets** → User can select packets to clear
5. **Clearing request** → Goes through clearing service with payment

## Testing

Run the integration test script:
```bash
./test-chainpulse-integration.sh
```

This tests:
1. Chainpulse health
2. Metrics endpoint
3. API backend integration
4. User transfer queries
5. Middleware proxy endpoints

## Monitoring

Prometheus scrapes Chainpulse metrics every 10 seconds:
- `ibc_effected_packets` - Successfully relayed packets
- `ibc_uneffected_packets` - Failed relay attempts
- `ibc_stuck_packets` - Currently stuck packets
- `chainpulse_chains` - Number of monitored chains
- `chainpulse_txs` - Total transactions processed
- `chainpulse_packets` - Total packets processed

## Error Handling

- API backend falls back to mock data if Chainpulse is unavailable
- Frontend shows loading states and error messages
- All errors are logged with context

## Security

- Wallet signatures required for clearing operations
- API validates wallet addresses before queries
- Chainpulse runs in isolated container
- No direct database access from frontend

## Future Enhancements

1. **Caching Layer** - Redis cache for frequently accessed packet data
2. **WebSocket Support** - Real-time packet status updates
3. **Batch Operations** - Clear multiple packets in single transaction
4. **Advanced Filtering** - Filter by date, amount, channel
5. **Historical Data** - Track clearing success rates over time