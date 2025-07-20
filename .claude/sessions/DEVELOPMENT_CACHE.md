# Development Cache

## Purpose
This file serves as a temporary cache for ongoing development work, tracking current issues, solutions in progress, and notes that haven't been formalized into other documentation yet.

## Current Development Status (as of 2025-07-19)

### Active Work Items

#### 1. Payment Verification Implementation
**Status**: In Progress
**Priority**: High
**Details**:
- Token generation is working
- Need to implement blockchain transaction query
- Need to verify memo field matches payment token
- Need to handle multiple payment tokens (ATOM, OSMO, etc.)

**Next Steps**:
```go
// TODO: Implement in clearing service
func (s *ClearingService) queryTransaction(txHash string) (*Transaction, error) {
    // Query chain for transaction details
    // Verify memo matches token
    // Check amount is sufficient
}
```

#### 2. End-to-End Packet Clearing Testing
**Status**: Not Started
**Priority**: High
**Blockers**: Payment verification must be completed first

**Test Plan**:
1. Create stuck packet in test environment
2. Generate clearing token
3. Make payment with token in memo
4. Verify payment
5. Execute clearing via Hermes
6. Confirm packet cleared

#### 3. Neutron Chain Support
**Status**: Blocked
**Issue**: ABCI++ vote extensions not supported by Chainpulse
**Impact**: Shows as "degraded", limited functionality

**Research Notes**:
- Neutron uses Slinky oracle with vote extensions
- Would require significant Chainpulse changes
- Consider alternative data sources

### Recent Discoveries

#### Docker Networking on macOS
- Cannot use `localhost` between containers
- Must use service names (e.g., `postgres`, not `localhost:5432`)
- Affects all inter-service communication

#### Frontend Build Requirements
- MUST build frontend before Docker deployment
- Hot reload doesn't work properly in Docker on macOS
- Use quick iteration commands for development

#### WebSocket Configuration
- Requires special nginx configuration for proxying
- Connection timeout needs to be increased
- Implement reconnection logic on frontend

### Temporary Workarounds

#### 1. Settings Page Disabled
```javascript
// In router/index.ts
// Temporarily disabled due to build issue
// { path: '/settings', component: Settings }
```
**TODO**: Re-enable once component is fixed

#### 2. Mock Payment Verification
```go
// In clearing_handler.go
// TODO: Remove this mock implementation
func verifyPayment(token string, txHash string) error {
    // MOCK: Always return success for testing
    log.Println("MOCK: Payment verification bypassed")
    return nil
}
```

### Performance Observations

#### Database Query Optimization Needed
- Stuck packets query is slow with large dataset
- Need to add compound index on (status, created_at)
- Consider materialized view for frequently accessed data

#### Memory Usage Patterns
- Chainpulse memory grows over time
- Implement packet archival after 30 days
- Consider pagination for large result sets

### Configuration Notes

#### RPC Endpoints
- Most require authentication in URL
- Format: `https://username:password@host:port`
- WebSocket URLs also need auth
- gRPC doesn't support URL auth

#### Environment Variables
- JWT_SECRET must be strong in production
- ADMIN_API_KEY for administrative operations
- Chain RPC endpoints need auth credentials

### Integration Challenges

#### Hermes REST API
- Limited documentation
- Some endpoints return non-standard responses
- Error messages not always helpful
- Need timeout handling for long operations

#### Chainpulse Metrics
- Prometheus format requires parsing
- No structured JSON API in original
- Created facade in main API for easier consumption

### Security Considerations

#### Token Security
- Clearing tokens have 30-minute expiry
- Stored in Redis with TTL
- HMAC signed to prevent tampering
- Need to implement token revocation

#### Rate Limiting
- Defined error codes but not implemented
- Need to add middleware
- Consider different limits per endpoint
- WebSocket connections need limiting

### Testing Gaps

#### Missing Test Coverage
1. WebSocket message handling
2. Payment verification flow
3. Hermes error scenarios
4. Chain-specific edge cases
5. Concurrent clearing requests

#### Test Data Needed
- Realistic stuck packet scenarios
- Various chain configurations
- Different token types
- Error transaction hashes

### Deployment Considerations

#### Docker Image Sizes
- Frontend: ~50MB (nginx alpine)
- API: ~20MB (scratch with binary)
- Chainpulse: ~100MB (debian slim)
- Consider multi-stage builds

#### Health Check Improvements
- Add deep health checks
- Check dependent service health
- Include version information
- Add readiness vs liveness

### Code Cleanup Tasks

#### Technical Debt
1. Remove mock API (`/api` directory)
2. Consolidate error handling
3. Standardize logging format
4. Remove commented code
5. Update deprecated dependencies

#### Refactoring Opportunities
- Extract clearing logic to separate service
- Implement proper event sourcing
- Create shared types package
- Improve test structure

### Documentation Gaps

#### Need to Document
1. Production deployment checklist
2. Disaster recovery procedures
3. Performance tuning guide
4. API client SDK examples
5. Monitoring setup guide

### Next Session Priorities

1. Complete payment verification implementation
2. Test full clearing flow end-to-end
3. Re-enable settings page
4. Implement rate limiting
5. Add missing test coverage

### Questions for Team

1. Preferred payment tokens to support?
2. Rate limiting thresholds?
3. Retention policy for cleared packets?
4. Monitoring/alerting preferences?
5. Deployment target (K8s, ECS, etc.)?

### Useful Commands Discovered

```bash
# Quick frontend rebuild and restart
cd webapp && yarn build && cd .. && docker-compose restart webapp

# Check stuck packets in DB
docker-compose exec postgres psql -U relayooor_user -d relayooor -c "SELECT COUNT(*) FROM packets WHERE status = 'pending'"

# Monitor Hermes logs for errors
docker-compose logs -f hermes | grep -i error

# Test Chainpulse metrics
curl -s http://localhost:3001/metrics | grep stuck_packets
```

### Environment-Specific Notes

#### Local Development
- Use docker-compose.local.yml for persistence
- Disable SSL for local RPC endpoints
- Use test mnemonics for Hermes

#### Staging
- Enable debug logging
- Use testnet chains
- Shorter token expiry for testing

#### Production
- Enable all security features
- Use hardware security module
- Implement audit logging

---

## Notes for Next Session

This cache should be reviewed at the start of the next development session. Items should be:
1. Converted to formal documentation if stable
2. Moved to issue tracker if they're bugs
3. Deleted if no longer relevant
4. Updated with current status

**Last Updated**: 2025-07-20
**Session Duration**: Search feature implementation and API consolidation
**Next Focus**: Hermes authentication fix and production testing

## Recent Session Updates (2025-07-20)

### Completed Features
1. **Comprehensive Packet Search**
   - Added PacketSearch.vue component with multi-mode search
   - Implemented /api/packets/search endpoint
   - Supports search by wallet, chain, token, age
   - CSV export functionality
   - Integrated into main dashboard

### API Architecture Clarification
- The simple API (/api/cmd/server/main.go) is the main implementation
- Uses Chainpulse APIs directly for data
- The api/handlers package is unused (compilation fixed but not integrated)
- All packet data comes from Chainpulse, not Prometheus parsing

### Known Issues
- Hermes authentication with encoded passwords in URLs not working
- Decided to postpone Hermes fix to focus on other features

### Next Steps
1. Fix Hermes authentication issue
2. Complete end-to-end testing with live data
3. Performance optimization for large datasets