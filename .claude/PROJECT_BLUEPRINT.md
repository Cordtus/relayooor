# Relayooor Project Blueprint

## Project Overview

Relayooor is a comprehensive IBC (Inter-Blockchain Communication) packet clearing platform for the Cosmos ecosystem. It provides users with tools to identify and clear stuck IBC transfers through a secure, token-based authorization system.

### Core Value Proposition

- **Problem**: IBC transfers can get stuck due to relayer failures, leaving user funds in limbo
- **Solution**: A user-friendly platform that allows users to clear their own stuck packets without technical expertise
- **Key Differentiator**: Secure token-based authorization ensures only packet owners can clear their transfers

## System Architecture

### High-Level Architecture

```
┌─────────────────┐     ┌─────────────────┐
│   Vue Frontend  │     │ Packet Manager  │
│   (Primary UI)  │     │  (Simple UI)    │
└────────┬────────┘     └────────┬────────┘
         │                       │
         ├───────────┬───────────┘
         │           │
    ┌────▼───────────▼────┐
    │      Nginx Proxy    │
    └────────┬────────────┘
             │
    ┌────────▼────────────┐
    │    Go API Server    │
    │  (Main Business     │
    │      Logic)         │
    └────┬───┬───┬───┬───┘
         │   │   │   │
    ┌────▼┐  │   │  ┌▼────────┐
    │Redis│  │   │  │Chainpulse│
    └─────┘  │   │  │(Monitor) │
             │   │  └──────────┘
    ┌────────▼┐  │
    │PostgreSQL│ │
    └──────────┘ │
                 │
          ┌──────▼──────┐
          │   Hermes    │
          │ (IBC Relay) │
          └─────────────┘
```

### Component Interactions

1. **User Journey**: User → Frontend → API → Authorization → Payment → Clearing → Success
2. **Data Flow**: Chainpulse monitors chains → Stores in PostgreSQL → API queries → Frontend displays
3. **Clearing Flow**: API validates request → Generates token → Calls Hermes → Clears packet

## Key Components

### 1. Frontend Applications

#### Main Web Application (`/webapp`)

- **Technology**: Vue 3, TypeScript, Pinia, Tailwind CSS, Vite
- **Purpose**: Primary user interface for packet monitoring and clearing
- **Key Features**:
  - Real-time packet monitoring dashboard
  - Wallet integration (Keplr, Leap, etc.)
  - Multi-step clearing wizard
  - Analytics and statistics
  - Channel management
  - WebSocket real-time updates

#### Packet Manager (`/packet-manager`)

- **Technology**: Vue 3, minimal dependencies
- **Purpose**: Simplified interface for basic packet clearing
- **Key Features**:
  - Streamlined clearing process
  - Mobile-friendly design
  - Minimal configuration required
  - Docker-based deployment

### 2. Backend Services

#### Main API (`/relayer-middleware/api`)

- **Technology**: Go, Gin framework, PostgreSQL, Redis
- **Purpose**: Core business logic and API endpoints
- **Key Features**:
  - JWT-based authentication
  - Secure token generation for packet clearing
  - Payment verification
  - Rate limiting and caching
  - WebSocket support for real-time updates
  - Prometheus metrics export

#### Simple API (`/api`)

- **Technology**: Go, basic HTTP server
- **Purpose**: Mock API for development
- **Note**: NOT for production use

### 3. External Services

#### Chainpulse

- **Technology**: Rust, async runtime
- **Purpose**: IBC monitoring and data collection
- **Modifications**: Custom fork with user-based packet queries and CometBFT 0.38 support
- **Integration**: REST API on port 3000, metrics on port 3001
- **Known Issues**: Cannot decode Neutron's ABCI++ vote extensions

#### Hermes IBC Relayer

- **Technology**: Rust, IBC protocol implementation
- **Purpose**: Execute packet clearing operations
- **Integration**: REST API on port 5185, telemetry on port 3001
- **Configuration**: Custom entrypoint scripts for Docker deployment
- **Authentication**: RPC/WebSocket require auth, gRPC does not

### 4. Data Storage

#### PostgreSQL

- **Purpose**: Primary data store
- **Schema**: Optimized for packet queries with indexes and materialized views
- **Key Tables**:
  - `packets` - IBC packet data
  - `channels` - IBC channel information
  - `chains` - Supported chain configurations
  - `users` - User accounts and preferences
  - `clearing_requests` - Clearing request history
  - `transactions` - Payment transactions

#### Redis

- **Purpose**: Caching and session management
- **Usage**:
  - Token storage
  - Rate limiting
  - Session management
  - Temporary data cache
  - WebSocket connection state

## Security Architecture

### Authentication Flow

1. User connects wallet
2. Signs message with private key
3. API verifies signature
4. Issues JWT token
5. Token used for authenticated requests

### Authorization Model

- **Packet Ownership**: Only packet sender/receiver can clear
- **Token-Based**: Unique token per clearing request
- **Time-Limited**: Tokens expire after set period
- **Payment Verification**: Must verify payment before clearing

### Security Measures

- HTTPS everywhere
- CORS configuration
- Rate limiting
- Input validation
- SQL injection prevention
- XSS protection
- HMAC signing for tokens

## Deployment Architecture

### Docker-Based Deployment

- All services containerized
- Docker Compose orchestration
- Health checks and restart policies
- Volume mounts for persistence

### Service Dependencies

```
postgres → api, chainpulse
redis → api
chainpulse → api
hermes → api
api → webapp, packet-manager
```

### Port Mapping

- 80: Nginx (HTTP)
- 443: Nginx (HTTPS)
- 3000: Chainpulse API
- 3001: Metrics/Telemetry
- 5185: Hermes API
- 5432: PostgreSQL
- 6379: Redis
- 8080: API (internal)

## Development Workflow

### Local Development

**IMPORTANT**: On macOS, you CANNOT run `yarn dev` directly. All services must run in Docker.

```bash
# Initial setup
cd webapp
yarn install
yarn build
cd ..

# Start services
./scripts/setup-and-launch.sh

# OR manually
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

### Quick Iteration

```bash
# Frontend changes
cd webapp && yarn build && cd .. && make -f Makefile.docker webapp-restart

# Backend changes
make -f Makefile.docker api-restart

# View logs
docker-compose logs -f webapp api-backend chainpulse
```

### Testing Strategy

- Unit tests for business logic
- Integration tests for API endpoints
- Component tests for frontend (Vitest)
- E2E tests for critical flows

### Available Test Scripts

```bash
./scripts/run-tests.sh                    # All tests
./scripts/test-chainpulse-integration.sh  # Chainpulse integration
./scripts/test-packet-clearing-scenarios.sh # Clearing scenarios
./scripts/check-chain-compatibility.sh    # Chain compatibility
```

## Configuration Management

### Environment Variables

Required variables in `.env`:

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
```

### Service Configuration

- TOML files for Chainpulse/Hermes
- JSON for frontend config
- Environment injection for Docker

## API Endpoints Reference

### Public Endpoints

- `GET /health` - Health check
- `GET /api/chains/{id}/packets/stuck` - Get stuck packets
- `GET /api/help/glossary` - Get help tooltips

### Clearing Service

- `POST /api/clearing/request` - Request clearing authorization
- `POST /api/clearing/verify-payment` - Verify payment transaction
- `GET /api/clearing/status/{token}` - Check clearing status
- `POST /api/clearing/execute` - Execute packet clearing

### Authentication

- `POST /api/auth/wallet-sign` - Authenticate with wallet signature
- `POST /api/auth/refresh` - Refresh authentication token

### Monitoring & Analytics

- `GET /api/monitoring/data` - Monitoring statistics
- `GET /api/statistics/platform` - Platform-wide metrics
- `GET /api/channels/congestion` - Channel congestion data

### WebSocket

- `GET /ws` - WebSocket for real-time updates
- `GET /api/packets/stuck/stream` - Server-sent events

## Known Issues & Workarounds

1. **Neutron Chain Support**: Chainpulse cannot decode Neutron's ABCI++ blocks
   - Status: Shows as "degraded"
   - Workaround: None currently, requires Chainpulse update

2. **Local Development on macOS**: Cannot use `yarn dev` directly
   - Workaround: Use Docker-based development

3. **RPC Authentication**: Some chains require auth tokens
   - Solution: Configure in environment variables

## Current Status (as of 2025-07-19)

### Working Features ✅

- Full stack deployment via Docker Compose
- API integration with real Chainpulse data
- Real-time IBC metrics collection
- Stuck packet detection
- Token-based clearing authorization
- WebSocket real-time updates
- Chain compatibility checking
- Wallet integration

### In Progress ⚠️

- Payment verification implementation
- End-to-end packet clearing testing
- Admin dashboard
- Batch clearing

### Known Issues ❌

- Neutron chain shows as "degraded"
- Settings page temporarily disabled
- Some chains may have RPC authentication issues

## Future Enhancements

### Roadmap

1. **Phase 1**: Core clearing functionality (current)
2. **Phase 2**: Multi-relayer support
3. **Phase 3**: Advanced analytics and reporting
4. **Phase 4**: Mobile applications
5. **Phase 5**: SDK for third-party integrations

### Scalability Considerations

- Horizontal scaling for API
- Read replicas for database
- CDN for static assets
- Queue system for clearing requests
- Microservices architecture

## Success Metrics

### Technical Metrics

- Packet clearing success rate > 95%
- API response time < 200ms
- System uptime > 99.9%
- Zero security incidents

### Business Metrics

- User adoption rate
- Total packets cleared
- Revenue from fees
- User satisfaction score
- Time to clear packets

## Troubleshooting Guide

### Common Issues

1. **Frontend shows old data**: Always run `yarn build` before restarting
2. **Services can't connect**: Use Docker service names, not localhost
3. **Build fails with cache**: Use `docker-compose build --no-cache`
4. **Can't connect to services**: Check with `docker-compose ps`

### Debug Commands

```bash
# Check service status
docker-compose ps

# View logs
docker-compose logs -f [service-name]

# Connect to database
psql -h localhost -U relayooor_user -d relayooor

# Test API endpoint
curl http://localhost:8080/health
```

---
Last Updated: 2025-07-19
