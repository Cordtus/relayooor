# Hardcoded Values Refactoring Plan

## Summary of Findings

Based on the comprehensive audit, here are the main areas with hardcoded values that need refactoring:

### 1. Frontend Chain Data (HIGH PRIORITY)
- **Location**: `/webapp/src/config/chains.ts`
- **Issue**: 139 lines of hardcoded chain configurations
- **Solution**: Load from API `/api/config` endpoint
- **Status**: ✅ Config service created, needs integration

### 2. Duplicate Chain Mappings (HIGH PRIORITY)
- **Locations**: 
  - `/webapp/src/services/api.ts` (lines 88-96, 356-359)
  - `/webapp/src/utils/metricsParser.ts` (lines 398-431)
  - `/webapp/src/config/chains.ts` (lines 154-190)
- **Issue**: Same data duplicated 3+ times
- **Solution**: Use central config service
- **Status**: ✅ Config service created, needs cleanup

### 3. Service URLs (MEDIUM PRIORITY)
- **Backend**: Multiple `http://localhost:3000` references
- **Frontend**: Hardcoded API and service URLs
- **Solution**: Environment variables + config module
- **Status**: ✅ Environment config created

### 4. Explorer URLs (COMPLETED)
- **Locations**: ClearingWizard.vue, TransferCard.vue
- **Issue**: Hardcoded mintscan URLs
- **Solution**: Config service with getExplorerUrl()
- **Status**: ✅ Completed

### 5. Docker Service Names (LOW PRIORITY)
- **Location**: docker-compose.yml, nginx.conf
- **Issue**: Hardcoded container names
- **Solution**: Use COMPOSE_PROJECT_NAME
- **Status**: ⏳ Pending

## Implementation Steps

### Phase 1: Core Infrastructure ✅
1. Create chain registry API endpoint
2. Create frontend config service
3. Create environment config module

### Phase 2: Frontend Cleanup (IN PROGRESS)
1. Replace chains.ts with dynamic loading
2. Remove duplicate chain mappings
3. Update all components to use config service
4. Test chain auto-discovery

### Phase 3: Backend Cleanup
1. Replace all localhost URLs with env vars
2. Create service discovery module
3. Update Docker configs
4. Test in all environments

### Phase 4: Documentation
1. Update CLAUDE.md with new patterns
2. Create configuration guide
3. Update deployment docs
4. Create migration checklist

## Components Needing Updates

### High Priority Components
- [ ] `/webapp/src/config/chains.ts` - Replace with API calls
- [ ] `/webapp/src/views/Dashboard.vue` - Use config service
- [ ] `/webapp/src/views/Monitoring.vue` - Use config service
- [ ] `/webapp/src/components/monitoring/*` - Remove hardcoded chains

### Medium Priority Components
- [ ] `/webapp/src/services/api.ts` - Remove duplicate mappings
- [ ] `/webapp/src/utils/metricsParser.ts` - Remove duplicate mappings
- [ ] `/api/cmd/server/main.go` - Use env vars for all URLs
- [ ] `/scripts/*.sh` - Accept base URLs as parameters

### Low Priority Components
- [ ] Docker compose files - Use variables
- [ ] Prometheus config - Template generation
- [ ] Nginx config - Environment substitution

## Testing Checklist

### Unit Tests
- [ ] Config service loads chain data
- [ ] Fallback works when API unavailable
- [ ] Explorer URL generation
- [ ] Address prefix detection

### Integration Tests
- [ ] New chain appears in UI automatically
- [ ] Channel data updates dynamically
- [ ] Service discovery works in Docker
- [ ] Environment overrides work

### End-to-End Tests
- [ ] Add new chain via config only
- [ ] Deploy with different service URLs
- [ ] Verify no hardcoded values remain

## Benefits After Refactoring

1. **Dynamic Chain Support** - Add chains without code changes
2. **Environment Flexibility** - Deploy anywhere easily
3. **Reduced Maintenance** - Single source of truth
4. **Better Testing** - Mock configurations easily
5. **Improved Documentation** - Clear configuration guide

## Next Steps

1. Complete Phase 2 frontend cleanup
2. Create automated tests for config service
3. Update all documentation
4. Create migration script for existing deployments
5. Plan Phase 3 backend cleanup