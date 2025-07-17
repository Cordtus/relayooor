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
- **Simple API (/api)**: Basic mock implementation - DO NOT USE for production features
- **Full API (/relayer-middleware/api)**: Complete implementation with packet clearing logic - THIS IS THE MAIN API

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
The frontend expects the full-featured API at `/api/*`. Currently, a simple mock API is deployed. When implementing features:
1. Check if endpoint exists in `/relayer-middleware/api/handlers/`
2. If not, implement in the full API, not the simple one
3. Frontend API client is at `webapp/src/api/client.ts`

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
   - Implement handler in `/relayer-middleware/api/handlers/`
   - Add route in `/relayer-middleware/api/routes/routes.go`
   - Update frontend API client if needed

2. **Creating a new Vue component**:
   - Follow existing patterns in `webapp/src/components/`
   - Add TypeScript interfaces in `webapp/src/types/`
   - Write tests in same directory with `.spec.ts` extension

3. **Database schema changes**:
   - Create migration in `/relayer-middleware/api/migrations/`
   - Test migration up and down
   - Update models in `/relayer-middleware/api/models/`