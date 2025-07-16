# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Relayooor is an IBC packet clearing platform that helps users clear stuck IBC transfers. The platform consists of:
1. **Packet Clearing Service** - Token-based authorization system with on-chain payment verification
2. **Web Application** - Vue.js frontend with Keplr wallet integration for viewing and clearing stuck packets
3. **API Backend** - Go-based API handling clearing operations, payment verification, and WebSocket updates
4. **Monitoring Infrastructure** - Chainpulse for IBC metrics, Hermes relayer for packet clearing execution

**Critical User Experience Goal**: The web app must be simple enough that non-technical users can connect their wallet, see stuck transfers, pay a fee, and have their packets cleared automatically.

## Key Commands

### 🚨 IMPORTANT: Development on macOS
**You cannot run development servers directly on macOS. All services must run through Docker.**

### Building and Deploying (Required Workflow)
```bash
# 1. Build the frontend (creates production bundle)
cd webapp && yarn build

# 2. Build and start all services via Docker
cd .. && docker-compose up -d

# Alternative deployment options:
docker-compose -f docker-compose.webapp.yml up -d  # Minimal webapp setup
docker-compose -f docker-compose.full.yml up -d    # Complete stack with Grafana
```

### Frontend Development
```bash
cd webapp && yarn install     # Install dependencies
cd webapp && yarn build       # Build production bundle (REQUIRED before Docker)
cd webapp && yarn type-check  # Run TypeScript type checking
```

### Backend Development
```bash
cd api && go mod download     # Download dependencies
cd api && go test ./...       # Run tests
cd api && go build -o api-server ./cmd/server  # Build binary

# Also test relayer-middleware API (legacy)
cd relayer-middleware/api && go test ./...
```

### Docker Service Management
```bash
# Main commands (from root directory)
./scripts/start.sh            # Start full stack with checks
docker-compose up -d          # Start all services
docker-compose down           # Stop all services
docker-compose logs -f        # View all logs
docker-compose ps             # Check service status

# Individual service restarts (using Makefile.docker)
make -f Makefile.docker webapp-restart      # Rebuild and restart webapp
make -f Makefile.docker api-restart         # Rebuild and restart API
make -f Makefile.docker chainpulse-restart  # Restart chainpulse

# Quick iteration for development
make -f Makefile.docker quick-fix-webapp    # Build and restart webapp
make -f Makefile.docker quick-fix-api       # Build and restart API

```

### Testing and Monitoring
```bash
# Test API endpoints
make -f Makefile.docker test-api

# Check service status
make -f Makefile.docker status

# View individual service logs
make -f Makefile.docker logs-webapp
make -f Makefile.docker logs-api
make -f Makefile.docker logs-chainpulse

# Run test scripts
./scripts/run-tests.sh                     # Run all tests
./scripts/test-chainpulse-integration.sh   # Test chainpulse endpoints
./scripts/test-packet-clearing-scenarios.sh # Test clearing scenarios
./scripts/check-chain-compatibility.sh     # Verify chain compatibility
./scripts/debug-neutron.sh                 # Debug Neutron issues
```

## Architecture & Structure

### Technology Stack
- **Frontend**: Vue 3, TypeScript, Vite, TailwindCSS, Pinia, Vue Query, WebSocket
- **Backend**: Go with Gin framework, HMAC token signing, WebSocket support, Redis, PostgreSQL
- **Monitoring**: Rust-based Chainpulse (modified for user data), Prometheus, Grafana
- **Infrastructure**: Docker Compose, Nginx, Hermes REST API

### Service Ports
- Web App: 80 (production) / 5173 (development)
- API Backend: 8080
- Hermes REST API: 5185
- Chainpulse Metrics: 3001
- Grafana: 3003
- Prometheus: 9090
- PostgreSQL: 5432
- Redis: 6379

### Key Directories
- `/webapp` - Vue.js application with packet clearing wizard
- `/api` - Go backend API with clearing service (note: NOT in relayer-middleware)
- `/monitoring/chainpulse` - Rust-based IBC monitoring (modified for user data collection)
- `/relayer-middleware` - Legacy directory with Hermes relayer configs and old API
- `/relayer-middleware/api/migrations` - Database migration files
- `/docs` - Comprehensive documentation for all components
- `/config` - Configuration files for chainpulse and other services
- `/scripts` - Utility scripts including chain compatibility checker

### Service Dependencies (Why Docker is Required)
The services form a tightly coupled system:
```
webapp (port 80)
  └─> api-backend (port 8080)
       ├─> chainpulse (port 3001) - Provides IBC metrics
       ├─> postgres (port 5432) - Stores clearing data
       └─> redis (port 6379) - Caches API responses

monitoring stack:
  └─> prometheus (port 9090) -> chainpulse metrics
       └─> grafana (port 3003) - Visualization
```

### Authentication & Authorization Flow
1. Users connect Keplr wallet to view their stuck packets
2. API generates cryptographically signed clearing tokens (5-minute expiry)
3. Users make on-chain payment with token in memo
4. API verifies payment and triggers packet clearing via Hermes
5. WebSocket provides real-time clearing status updates

## Development Priorities

### Current Focus (as of 2025-07-13)
- Production-ready packet clearing service
- User-friendly clearing wizard with payment flow
- Comprehensive monitoring and statistics
- Support for Osmosis, Cosmos Hub, Neutron chains
- IBC v2/Eureka compatibility planning

### When Making Changes
1. **Web App**: 
   - Maintain simple 5-step wizard flow for non-technical users
   - ALWAYS run `yarn build` before deploying changes
   - Use `make -f Makefile.docker webapp-restart` for quick iterations
2. **API**: 
   - Preserve token-based security and payment verification logic
   - Build with `go build` before Docker deployment
   - Use `make -f Makefile.docker api-restart` for quick iterations
3. **Monitoring**: Keep Chainpulse modifications minimal (only data collection)
4. **Documentation**: Update all relevant docs when adding features
5. **Testing**: Run type checking and tests before building

### Testing Strategy
- Frontend: `yarn type-check` before building (NOT yarn dev on macOS)
- Backend: `go test ./...` for unit tests
- Integration: Full stack testing with Docker Compose
- API Testing: `make -f Makefile.docker test-api` after deployment
- Payment Flow: Test with small amounts on testnet first

## 🚨 CRITICAL: Development Workflow on macOS

### Why Docker is Required
On macOS, you CANNOT run the development servers directly (no `yarn dev` or `go run`). All services are interconnected and must run together in Docker containers. The webapp needs the API, the API needs chainpulse and the database, etc.

### Correct Development Workflow

#### For Frontend Changes:
1. Make your code changes in `/webapp`
2. Build the production bundle: `cd webapp && yarn build`
3. Rebuild and restart the webapp container:
   ```bash
   make -f Makefile.docker webapp-restart
   ```
   OR use the full rebuild one-liner:
   ```bash
   cd webapp && yarn build && cd .. && docker-compose -f docker-compose.webapp.yml build --no-cache && docker-compose -f docker-compose.webapp.yml up -d
   ```

#### For Backend Changes:
1. Make your code changes in `/api`
2. Rebuild and restart the API container:
   ```bash
   make -f Makefile.docker api-restart
   ```

#### For Full Stack Changes:
1. Stop everything: `docker-compose down`
2. Build frontend: `cd webapp && yarn build`
3. Start everything: `cd .. && docker-compose up -d`

⚠️ **IMPORTANT**: 
- Frontend changes REQUIRE running `yarn build` first
- You'll access the app at http://localhost (port 80), NOT port 5173
- The webapp Dockerfile runs `vite build` during container build
- If you skip the build step, you'll see old/cached content

## Important Context

### From Project Documentation
- This is a monorepo structure with independent deployable components
- Supports multiple relayer types (Hermes and Go relayer) with legacy versions
- Production deployment configured for Fly.io with GitHub Actions CI/CD
- Uses mock-chainpulse for local development when needed
- RPC endpoints require authentication (username/password in .env)
- Redis used for API response caching with TTL
- Database includes optimizations: connection pooling, materialized views, smart indexing
- Circuit breaker pattern implemented for Hermes integration
- Graceful shutdown with 5-minute grace period for active clearings

### Known Issues & Important Notes
1. **Neutron Chain Compatibility**: Neutron shows as "degraded" due to protobuf decoding errors. This is because Neutron uses ABCI++ vote extensions for their Slinky oracle, which the current chainpulse version cannot decode. This is a known limitation.
2. **Chain Integration**: Before adding new chains, run `./scripts/check-chain-compatibility.sh` to identify potential compatibility issues
3. **Docker Management**: Use `make -f Makefile.docker` commands for easier service management
4. **API Dynamic Chain Detection**: The API now dynamically detects chains from chainpulse metrics - no need to hardcode chain lists
5. **Troubleshooting Guide**: See `/docs/chain-integration-troubleshooting.md` for detailed chain integration guidance

### Key Features to Maintain
1. Secure token-based clearing authorization
2. On-chain payment verification with memo parsing
3. Automated packet clearing via Hermes REST API
4. Real-time status updates via WebSocket
5. User statistics with wallet signature authentication
6. Platform-wide analytics and monitoring

## Common Development Pitfalls to Avoid

1. **DON'T run `yarn dev`** - It won't work on macOS due to service dependencies
2. **DON'T forget to run `yarn build`** before restarting webapp container
3. **DON'T try to run services individually** - They depend on each other
4. **DON'T edit files in running containers** - Changes will be lost
5. **DON'T use localhost for service-to-service communication** - Use Docker service names (e.g., `api-backend`, `chainpulse`)
6. **DON'T skip the nginx proxy** - The webapp expects `/api/` routes to be proxied

## Quick Command Reference

```bash
# Check what's running
docker ps

# Full rebuild and restart (frontend changes)
cd webapp && yarn build && cd .. && make -f Makefile.docker webapp-restart

# Check if API is healthy
curl http://localhost:8080/health

# View real-time logs
docker-compose logs -f webapp api-backend

# Clean slate restart
docker-compose down && docker-compose up -d
```

## Environment Variables

Key environment variables to configure (create `.env` file from `.env.example`):
```bash
# Authentication & Security
JWT_SECRET=your-secret-key
ADMIN_API_KEY=your-admin-key

# Relayer Selection
ACTIVE_RELAYER=hermes  # or go-relayer

# Database Configuration
POSTGRES_USER=relayooor_user
POSTGRES_PASSWORD=secure-password
POSTGRES_DB=relayooor
REDIS_PASSWORD=redis-password

# Chain RPC Endpoints (with authentication)
COSMOS_RPC_ENDPOINT=https://username:password@cosmos-rpc.example.com:443
OSMOSIS_RPC_ENDPOINT=https://username:password@osmosis-rpc.example.com:443
NEUTRON_RPC_ENDPOINT=https://username:password@neutron-rpc.example.com:443

# Backup RPC Endpoints (for failover)
COSMOS_BACKUP_RPC=https://backup-cosmos.example.com:443
OSMOSIS_BACKUP_RPC=https://backup-osmosis.example.com:443
```

## Database Migrations

Database migrations are located in `/relayer-middleware/api/migrations/`:
```bash
# Run migrations manually (if needed)
cd relayer-middleware/api
go run migrations/*.go up

# Create new migration
go run migrations/*.go create <migration_name>
```

## CI/CD and Deployment

### GitHub Actions Workflow
The project uses GitHub Actions for CI/CD (`.github/workflows/`):
- Automated testing on push to main branch
- Docker image building and pushing to GitHub Container Registry
- Fly.io deployment for production
- Manual workflow dispatch support

### Production Deployment
```bash
# Deploy to Fly.io (requires fly CLI)
fly deploy

# Check deployment status
fly status

# View production logs
fly logs
```

## Docker Compose Configurations

Multiple compose files are available for different scenarios:
- `docker-compose.yml` - Standard development stack
- `docker-compose.webapp.yml` - Minimal webapp-only setup
- `docker-compose.full.yml` - Complete stack including Grafana monitoring
- `docker-compose.minimal.yml` - Lightweight development environment
- `docker-compose.simple.yml` - Basic services only

Use with `-f` flag:
```bash
docker-compose -f docker-compose.full.yml up -d
```

## Internal Documentation

Comprehensive internal documentation for Claude Code is available in the `.claude/` directory:
- `PROJECT_BLUEPRINT.md` - Complete project overview, architecture, and current state
- `DEPLOYMENT_GUIDE.md` - Detailed deployment and operations guide
- Other technical documentation and analysis reports

These documents contain detailed information about the codebase, known issues, and development workflows.

## Running Single Tests

### Frontend Tests
```bash
# Run a specific test file
cd webapp && yarn test src/components/Card.test.ts

# Run tests in watch mode
cd webapp && yarn test:watch

# Run tests with coverage
cd webapp && yarn test:coverage
```

### Backend Tests
```bash
# Run tests for a specific package
cd api && go test ./pkg/handlers -v

# Run a single test function
cd api && go test -run TestHealthHandler ./pkg/handlers

# Run tests with coverage
cd api && go test -cover ./...

# For relayer-middleware API
cd relayer-middleware/api && go test ./pkg/clearing -v
```

## Code Architecture Patterns

### Frontend (Vue.js)
- **Components**: Use composition API with TypeScript
- **State Management**: Pinia stores in `/webapp/src/stores/`
- **API Calls**: Centralized in `/webapp/src/services/api.ts`
- **Types**: Shared types in `/webapp/src/types/`
- **Composables**: Reusable logic in `/webapp/src/composables/`

### Backend (Go)
- **Handler Pattern**: HTTP handlers in `/api/pkg/handlers/`
- **Middleware**: Auth, CORS, logging in `/api/pkg/middleware/`
- **Services**: Business logic separated from handlers
- **Database**: Repository pattern for data access
- **Error Handling**: Consistent error types and responses

### Cross-Service Communication
- Frontend → API: REST endpoints with `/api/` prefix (proxied by nginx)
- API → Chainpulse: HTTP calls to port 3001
- API → Database: Connection pooling with pgx
- WebSocket: Real-time updates on `/api/ws`