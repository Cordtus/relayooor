# Claude Internal Guide - Relayooor Project

## Project Overview

Relayooor is an IBC packet clearing platform with three main components:
1. **Vue.js webapp** - User interface for packet clearing
2. **Go API backend** - Handles clearing operations and payments
3. **Chainpulse monitoring** - Real-time IBC metrics (forked from Cordtus/chainpulse)

## Quick Commands

### Full Stack Launch
```bash
# Fastest way to launch everything
cd /Users/cordt/repos/relayooor
./start-full-stack.sh

# Or manually with Docker Compose
docker-compose -f docker-compose.full.yml up -d
```

### Individual Component Launch
```bash
# Backend only
cd relayer-middleware/api
go run cmd/server/main.go

# Frontend (M4 Mac - use Docker)
docker-compose -f docker-compose.full.yml up webapp

# Database migrations
cd relayer-middleware/api
migrate -path migrations -database $DATABASE_URL up
```

## Key Technical Details

### TypeScript Configuration
- **Relaxed mode**: `strict: false` for faster development
- **Path aliases**: `@/*` maps to `src/*`
- **Import extensions**: `.ts` files can be imported directly

### M4 Mac Considerations
- Vite dev server has binding issues on Apple Silicon
- Always use Docker for frontend development
- Alternative: `yarn build && yarn preview`

### Database Optimizations
- **Connection pooling**: pgx/v5 with 25 max connections
- **TimescaleDB**: For time-series packet data
- **Key indexes**: 
  - `idx_clearing_operations_wallet_created`
  - `idx_packet_flow_status_created`
  - `idx_network_stats_chain_date`

### API Authentication
- **Token signing**: HMAC-SHA256 with `CLEARING_SECRET_KEY`
- **Token expiry**: 5 minutes
- **Wallet auth**: EIP-191 signature verification

## Common Tasks

### Adding New Chain Support
1. Update `config/chainpulse.toml`:
   ```toml
   [chains.new-chain-id]
   url = "wss://rpc.example.com/websocket"
   comet_version = "0.38"
   ```
2. Add to `relayer-middleware/api/pkg/config/config.go`
3. Update frontend chain list in `webapp/src/types/chains.ts`

### Fixing TypeScript Errors
1. Check imports use `@/` prefix
2. Verify all interfaces are exported
3. Run `yarn type-check` to see all errors
4. Common fix: Add optional fields for type compatibility

### Database Schema Changes
1. Create migration file: `migrations/XXX_description.sql`
2. Always use `IF NOT EXISTS` for indexes
3. Create indexes `CONCURRENTLY` in production
4. Test rollback before applying

### Monitoring Performance
```sql
-- Slow queries
SELECT * FROM pg_stat_statements 
WHERE mean_exec_time > 1000 
ORDER BY mean_exec_time DESC;

-- Connection pool stats
SELECT * FROM pg_stat_activity;

-- Table sizes
SELECT pg_size_pretty(pg_total_relation_size(tablename::regclass))
FROM pg_tables WHERE schemaname = 'public';
```

## Environment Variables

### Required
- `CLEARING_SECRET_KEY`: Generate with `openssl rand -hex 32`
- `SERVICE_WALLET_ADDRESS`: Your service fee collection address
- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis for caching and rate limiting

### Chain RPCs
- `OSMOSIS_RPC`: wss://rpc.osmosis.zone/websocket
- `COSMOSHUB_RPC`: wss://rpc.cosmos.network/websocket
- `NEUTRON_RPC`: wss://rpc-kralum.neutron-1.neutron.org/websocket

## Debugging Tips

### Frontend Issues
- **Vite not accessible**: Use Docker or preview mode
- **Component not found**: Check file exists at exact path
- **Type errors**: Look for StuckPacket interface mismatches

### Backend Issues
- **Token validation fails**: Check CLEARING_SECRET_KEY matches
- **Payment not verified**: Ensure correct memo format
- **Database timeout**: Check connection pool stats

### Chainpulse Issues
- **No metrics**: Verify WebSocket URLs are correct
- **Missing user data**: Check packet parsing modifications
- **High memory**: Reduce block buffer size

## Testing Workflows

### Manual Payment Test
1. Get stuck packets: `GET /api/v1/users/{address}/stuck-packets`
2. Request token: `POST /api/v1/clearing/request-token`
3. Make IBC transfer with memo to service address
4. Verify payment: `POST /api/v1/clearing/verify-payment`
5. Check status: `GET /api/v1/clearing/status/{token}`

### Load Testing
```bash
# Generate load
hey -n 1000 -c 10 http://localhost:8080/api/v1/health

# Monitor performance
watch -n 1 'psql -c "SELECT * FROM pg_stat_activity"'
```

## Common Fixes

### "Cannot find module" Error
```typescript
// Add to tsconfig.app.json
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"]
    }
  }
}
```

### Database Migration Failed
```bash
# Check current version
migrate -path migrations -database $DATABASE_URL version

# Force version if needed
migrate -path migrations -database $DATABASE_URL force VERSION
```

### Redis Connection Refused
```bash
# Check Redis is running
docker ps | grep redis

# Test connection
redis-cli ping
```

## Code Patterns

### API Error Handling
```go
if err != nil {
    logger.Error("Operation failed", zap.Error(err))
    c.JSON(500, gin.H{"error": "Internal error"})
    return
}
```

### Vue Composition API
```vue
<script setup lang="ts">
import { ref, computed } from 'vue'

const data = ref<DataType>()
const processed = computed(() => data.value?.field)
</script>
```

### Database Queries
```go
query := NewQueryBuilder("table").
    Where("status = $1", "active").
    OrderBy("created_at", true).
    Limit(100)
```

## Deployment Checklist

1. [ ] Update version in package.json
2. [ ] Run database migrations
3. [ ] Build frontend: `yarn build`
4. [ ] Test with Docker Compose
5. [ ] Check all health endpoints
6. [ ] Verify Chainpulse metrics
7. [ ] Test payment flow end-to-end
8. [ ] Monitor logs for errors

## Quick Diagnostics

```bash
# Check all services
curl -s http://localhost:8080/api/v1/health | jq
curl -s http://localhost:3000/metrics
curl -s http://localhost:5173/ | head -n 10

# Database health
psql $DATABASE_URL -c "SELECT version();"

# Redis health
redis-cli ping

# Docker status
docker-compose -f docker-compose.full.yml ps
```

## Project Status Notes

- **Completed**: Docker setup, TypeScript fixes, database optimizations
- **Current focus**: Building and launching full stack locally
- **Known issues**: Vite dev server on M4 Macs (use Docker)
- **Next steps**: Production deployment preparation

## File Locations

- Main config: `/Users/cordt/repos/relayooor/.env`
- Chainpulse config: `/Users/cordt/repos/relayooor/config/chainpulse.toml`
- Database schema: `/Users/cordt/repos/relayooor/relayer-middleware/api/migrations/`
- API handlers: `/Users/cordt/repos/relayooor/relayer-middleware/api/pkg/handlers/`
- Vue components: `/Users/cordt/repos/relayooor/webapp/src/components/`

## Performance Targets

- API response time: <100ms for queries
- Database query time: <50ms for indexed queries
- Frontend build size: <500KB gzipped
- Chainpulse memory: <1GB per chain
- Clearing success rate: >95%

## Security Reminders

- Never log CLEARING_SECRET_KEY
- Validate all user inputs
- Use prepared statements for SQL
- Rate limit all endpoints
- Sanitize error messages
- Keep dependencies updated