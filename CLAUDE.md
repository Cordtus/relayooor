# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Relayooor is an IBC (Inter-Blockchain Communication) packet clearing platform for the Cosmos ecosystem. It helps users clear stuck IBC transfers through a secure, token-based authorization system.

## Key Architecture Components

### Frontend (webapp/)

- Vue 3 with TypeScript, Composition API, and `<script setup>` syntax
- Pinia for state management
- Tailwind CSS for styling
- Vue Router for navigation
- Vite as build tool

### Backend APIs

- **Simple API (/api)**: Main API implementation using Chainpulse APIs directly
- **Full API (/relayer-middleware/api)**: Advanced implementation with packet clearing logic (for future features)

### Services

- **Chainpulse**: Fork of IBC monitoring service with CometBFT 0.38 support
- **PostgreSQL**: Main database with optimized schemas
- **Redis**: Caching layer
- **Hermes**: IBC relayer for packet clearing execution

## Essential Development Commands

### Frontend Development

```bash
cd webapp
yarn install
yarn dev           # Start dev server on http://localhost:5173
yarn test          # Run unit tests with Vitest
yarn test:ui       # Run tests with UI
yarn build         # Build for production
yarn preview       # Preview production build
yarn type-check    # Run TypeScript type checking
```

### Backend Development

```bash
# From project root
make dev-backend   # Start backend services
make dev-frontend  # Start frontend development
make test          # Run all tests (backend + frontend)
make start         # Start all services

# Go API development
cd relayer-middleware/api
go test ./...      # Run tests
go run .           # Start API server
```

### Database Management

```bash
# Apply migrations
cd relayer-middleware/api
go run . migrate up

# Connect to database
psql -h localhost -U relayooor_user -d relayooor
```

## Critical Implementation Notes

### API Integration

The frontend uses the simple API implementation at `/api/cmd/server/main.go` which provides all necessary endpoints:

1. The main API is in `/api/cmd/server/main.go` (not `/api/handlers/` which is unused)
2. Frontend API client is at `webapp/src/services/api.ts`
3. All packet data comes from Chainpulse APIs, not Prometheus metrics parsing
4. The api/handlers package exists but is not currently used in the application

### State Management Pattern

```typescript
// Use Pinia stores in webapp/src/stores/
const store = usePacketStore()
await store.fetchPackets()
```

### Component Structure

- Components use `<script setup lang="ts">` syntax
- Props defined with `defineProps<{...}>()`
- Emit events with `defineEmits<{...}>()`
- Follow existing patterns in `webapp/src/components/`

### Testing Approach

```typescript
// Frontend component tests
import { render, screen } from '@testing-library/vue'
import { describe, it, expect } from 'vitest'

// Backend tests
func TestHandler(t *testing.T) {
    // Use testify/assert for assertions
    assert.Equal(t, expected, actual)
}
```

## Recent Features

### Packet Search (Added 2025-07-20)
- Comprehensive search functionality for IBC packets
- Search by wallet address (as sender or receiver)
- Filter by chain, token/denom, and age
- Export results to CSV
- Component: `webapp/src/components/search/PacketSearch.vue`
- API endpoint: `GET /api/packets/search`

## Known Issues & Workarounds

1. **Neutron Chain Support**: Chainpulse cannot decode Neutron's ABCI++ blocks. See `docs/neutron-slinky-issue.md`
2. **Local Development**: Use `docker-compose.local.yml` for full local setup
3. **Hermes Authentication**: RPC and WebSocket endpoints require authentication, gRPC does not

## Project-Specific Conventions

### Git Workflow

- Feature branches: `feature/description`
- Bug fixes: `fix/description`
- Commit format: `type: description` (feat, fix, docs, refactor, test)

### API Response Format

```json
{
  "success": true,
  "data": {...},
  "error": null
}
```

### Error Handling

- Frontend: Use toast notifications for user-facing errors
- Backend: Return structured error responses with appropriate HTTP status codes

### Security Considerations

- All packet clearing requests require JWT authentication
- Token validation happens in middleware
- Never log sensitive information (keys, tokens)

## Debugging Tips

### Frontend Debugging

- Vue DevTools for component inspection
- Network tab for API calls
- Check Pinia stores for state issues

### Backend Debugging

- Enable debug logging: `LOG_LEVEL=debug`
- Check PostgreSQL logs for query issues
- Monitor Redis for caching problems

### Common Development Tasks

1. **Adding a new API endpoint**:
   - Add handler function in `/api/cmd/server/main.go`
   - Register route with the Gorilla Mux router
   - Update frontend API client in `webapp/src/services/api.ts`

2. **Creating a new Vue component**:
   - Follow existing patterns in `webapp/src/components/`
   - Add TypeScript interfaces in `webapp/src/types/`
   - Write tests in same directory with `.spec.ts` extension

3. **Database schema changes**:
   - Create migration in `/relayer-middleware/api/migrations/`
   - Test migration up and down
   - Update models in `/relayer-middleware/api/models/`

## Internal Documentation

Comprehensive internal documentation is maintained in the `.claude` directory. This documentation provides detailed guidance for all aspects of development, deployment, and operations.

### Documentation Index

#### Core Documentation
- **[.claude/PROJECT_BLUEPRINT.md](.claude/PROJECT_BLUEPRINT.md)** - System architecture and design overview
- **[.claude/BUILD_AND_DEPLOYMENT.md](.claude/BUILD_AND_DEPLOYMENT.md)** - Complete build and deployment procedures
- **[.claude/FILE_MAPPING.md](.claude/FILE_MAPPING.md)** - Comprehensive file reference and structure
- **[.claude/API_INTERFACES.md](.claude/API_INTERFACES.md)** - External API documentation (Hermes, Chainpulse)
- **[.claude/TROUBLESHOOTING.md](.claude/TROUBLESHOOTING.md)** - Common issues and solutions

#### Module Blueprints
- **[.claude/implementations/FRONTEND_MODULE.md](.claude/implementations/FRONTEND_MODULE.md)** - Frontend architecture details
- **[.claude/implementations/API_MODULE.md](.claude/implementations/API_MODULE.md)** - Backend API implementation
- **[.claude/implementations/CHAINPULSE_MODULE.md](.claude/implementations/CHAINPULSE_MODULE.md)** - Monitoring service details
- **[.claude/implementations/HERMES_MODULE.md](.claude/implementations/HERMES_MODULE.md)** - IBC relayer integration

#### Working Documentation
- **[.claude/sessions/DEVELOPMENT_CACHE.md](.claude/sessions/DEVELOPMENT_CACHE.md)** - Current development status and notes
- **[.claude/INDEX.md](.claude/INDEX.md)** - Complete documentation index with usage guide

### When to Consult Internal Docs

- **Starting development**: Read PROJECT_BLUEPRINT and BUILD_AND_DEPLOYMENT
- **Adding features**: Consult relevant module blueprint
- **Debugging issues**: Check TROUBLESHOOTING first
- **Finding files**: Use FILE_MAPPING reference
- **API integration**: Refer to API_INTERFACES
- **Current work**: Review DEVELOPMENT_CACHE

### Quick Access Commands
```bash
# View all documentation
ls -la .claude/

# Search documentation
grep -r "search-term" .claude/

# Open documentation index
cat .claude/INDEX.md
```
