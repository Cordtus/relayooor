# PROJECT BLUEPRINT - Relayooor IBC Packet Clearing Platform

## Executive Summary

Relayooor is an IBC packet clearing platform that helps users clear stuck IBC transfers. The platform uses a token-based payment system where users pay on-chain and the service automatically clears their stuck packets.

## Current State (as of 2025-07-16)

### What Works
- ✅ Full stack deployment via Docker Compose
- ✅ API integration with real Chainpulse data
- ✅ Real-time IBC metrics collection (15,442 packets/day across 4 chains)
- ✅ Stuck packet detection (60+ stuck packets detected with 12-13h stuck duration)
- ✅ Monitoring data endpoints with chain statistics
- ✅ Token-based clearing authorization with HMAC signing
- ✅ WebSocket real-time updates
- ✅ Chain compatibility checking
- ✅ Basic UI with wallet integration

### Known Issues
- ❌ Neutron chain shows as "degraded" (ABCI++ vote extensions not supported)
- ❌ Local development mode (`yarn dev`) doesn't work - must use Docker
- ❌ Settings page temporarily disabled due to build issue
- ❌ Some chains may have RPC authentication issues
- ⚠️ Payment verification not fully implemented
- ⚠️ Actual packet clearing via Hermes not tested end-to-end

## Architecture Overview

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Web App       │────▶│   API Backend   │────▶│   Chainpulse    │
│  (Vue 3/TS)     │     │     (Go)        │     │     (Rust)      │
│   Port 80       │     │   Port 8080     │     │   Port 3001     │
└─────────────────┘     └─────────────────┘     └─────────────────┘
         │                       │                        │
         │                       ▼                        │
         │              ┌─────────────────┐              │
         │              │   PostgreSQL    │              │
         │              │   Port 5432     │              │
         │              └─────────────────┘              │
         │                       │                        │
         └───────────────────────┴────────────────────────┘
                                 │
                         ┌───────▼────────┐
                         │     Redis      │
                         │   Port 6379    │
                         └────────────────┘
```

## Critical Build & Run Instructions

### ⚠️ IMPORTANT: macOS Development Workflow

**You CANNOT run development servers directly on macOS.** All services must run in Docker due to tight service dependencies.

### Correct Build Process

```bash
# 1. Clone and setup
git clone <repo>
cd relayooor

# 2. Create .env file
cp .env.example .env
# Edit .env with your RPC endpoints and credentials

# 3. Build frontend FIRST (critical!)
cd webapp
yarn install
yarn build
cd ..

# 4. Start all services
./scripts/setup-and-launch.sh

# OR manually:
docker-compose down  # Clean slate
docker-compose build --no-cache  # Fresh build
docker-compose up -d
```

### Access Points
- Web App: http://localhost
- API: http://localhost:8080
- Chainpulse Metrics: http://localhost:3001/metrics
- Grafana: http://localhost:3003 (if using full stack)

### Quick Iteration Commands
```bash
# Frontend changes
cd webapp && yarn build && cd .. && make -f Makefile.docker webapp-restart

# Backend changes  
make -f Makefile.docker api-restart

# View logs
docker-compose logs -f webapp api-backend chainpulse
```

## Service Configuration

### Required Environment Variables
```bash
# Authentication
JWT_SECRET=<your-secret>
ADMIN_API_KEY=<admin-key>

# Database
POSTGRES_USER=relayooor_user
POSTGRES_PASSWORD=<secure-password>
POSTGRES_DB=relayooor
REDIS_PASSWORD=<redis-password>

# Chain RPC (with auth)
COSMOS_RPC_ENDPOINT=https://username:password@cosmos-rpc.example.com:443
OSMOSIS_RPC_ENDPOINT=https://username:password@osmosis-rpc.example.com:443
NEUTRON_RPC_ENDPOINT=https://username:password@neutron-rpc.example.com:443
```

### Docker Services
- `webapp` - Vue.js frontend (nginx)
- `api-backend` - Go API server  
- `chainpulse` - Rust IBC monitoring
- `postgres` - Database
- `redis` - Caching
- `hermes` - IBC relayer (pulled from ghcr.io)

## Key API Endpoints

### Chainpulse Data (Port 3001)
- `/metrics` - Prometheus metrics with IBC data (working ✅)

### API Backend (Port 3000 external, 8080 internal)
- `/health` - Health check (working ✅)
- `/api/metrics` - Basic metrics (working ✅)
- `/api/channels` - Channel list (working ✅)
- `/api/packets/stuck` - Stuck packets list (working ✅)
- `/api/monitoring/data` - Monitoring statistics (working ✅)
- `/api/monitoring/metrics` - Metrics data (working ✅)
- `/api/statistics/platform` - Platform stats (working ✅)
- `/api/generate-clearing-token` - Create payment token (not tested)
- `/api/verify-payment` - Verify on-chain payment (not tested)
- `/api/clear-packet` - Trigger packet clearing (not tested)
- `/ws` - WebSocket for real-time updates (not tested)

## Recent Improvements (from git history)

1. **Real Data Integration** (16c944c, 4c8f436)
   - Connected API to real Chainpulse metrics
   - Removed hardcoded mock data

2. **API Routing Fix** (0e261b0)
   - Fixed doubled /api paths in routing

3. **Build Process** (1f519fd)
   - Added setup script and documentation
   - Improved Docker workflow

4. **Configuration System** (1e928f0, b434349)
   - Centralized config service
   - Removed hardcoded values

5. **Chain Compatibility** (9920b57)
   - Added chain compatibility checking scripts
   - Documented Neutron issues

## Outstanding Tasks & Issues

### High Priority
1. Complete payment verification implementation
2. Test end-to-end packet clearing with Hermes
3. Fix Neutron chain compatibility (ABCI++ support)
4. Implement proper error handling for failed clearings
5. Add comprehensive logging

### Medium Priority  
1. Implement refund mechanism for failed clearings
2. Add more chain support (validate with compatibility script)
3. Improve error messages for users
4. Add admin dashboard
5. Implement rate limiting

### Low Priority
1. Performance optimizations
2. Add more detailed metrics
3. Implement batch clearing
4. Mobile responsive improvements

## Testing

### Available Test Scripts
```bash
# Run all tests
./scripts/run-tests.sh

# Test chainpulse integration
./scripts/test-chainpulse-integration.sh  

# Test packet clearing scenarios
./scripts/test-packet-clearing-scenarios.sh

# Check chain compatibility
./scripts/check-chain-compatibility.sh <chain-id>

# Debug Neutron issues
./scripts/debug-neutron.sh
```

### Manual Testing Checklist
1. [ ] Connect Keplr wallet
2. [ ] View stuck packets
3. [ ] Generate clearing token
4. [ ] Make payment with token in memo
5. [ ] Verify payment processed
6. [ ] Check packet clearing status
7. [ ] Verify WebSocket updates

## Deployment

### Production (Fly.io)
```bash
fly deploy
fly status
fly logs
```

### Docker Compose Variants
- `docker-compose.yml` - Standard development
- `docker-compose.webapp.yml` - Minimal webapp only
- `docker-compose.full.yml` - With Grafana monitoring

## Common Issues & Solutions

### Issue: Frontend shows old data after changes
**Solution**: Always run `yarn build` before restarting container

### Issue: Services can't connect to each other
**Solution**: Use Docker service names (not localhost) for inter-service communication

### Issue: Neutron chain shows as degraded
**Solution**: This is expected - Neutron uses ABCI++ vote extensions not supported by current chainpulse

### Issue: Build fails with cache issues
**Solution**: Use `docker-compose build --no-cache`

### Issue: Can't connect to services
**Solution**: Check if all services are running with `docker-compose ps`

## Code Organization

### Frontend (`/webapp`)
- `/src/components` - Vue components
- `/src/views` - Page components  
- `/src/services` - API clients
- `/src/stores` - Pinia state management
- `/src/config` - Configuration

### Backend (`/api`)
- `/cmd/server` - Main entry point
- `/handlers` - HTTP handlers
- `/middleware` - Auth, CORS, etc
- `/services` - Business logic
- `/models` - Data models

### Monitoring (`/monitoring/chainpulse`)
- Modified fork of chainpulse
- Collects IBC metrics
- Exposes Prometheus metrics

## Important Notes

1. **Never run `yarn dev`** on macOS - it won't work due to service dependencies
2. **Always build frontend first** before Docker deployment
3. **Use correct ports** - API is on 8080, not 3000
4. **Check logs frequently** during development
5. **Test with small amounts** when testing payments

---

Last Updated: 2025-07-16
Git Revision: 16c944c