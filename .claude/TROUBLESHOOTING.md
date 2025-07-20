# Troubleshooting Guide

## Overview

This guide covers common issues encountered during development and production operations of the Relayooor platform, along with their solutions.

## Development Issues

### 1. macOS Local Development Problems

#### Issue: `yarn dev` fails with connection errors
```
Error: connect ECONNREFUSED 127.0.0.1:8080
```

**Root Cause**: Frontend dev server cannot connect to backend services on macOS due to Docker networking limitations.

**Solution**:
```bash
# Don't use yarn dev directly. Use Docker-based development:
cd webapp
yarn build
cd ..
docker-compose up -d
```

**Quick Iteration**:
```bash
# After making frontend changes:
cd webapp && yarn build && cd .. && make -f Makefile.docker webapp-restart
```

### 2. Frontend Build Issues

#### Issue: Frontend shows old content after changes
**Symptoms**: Changes not reflected even after container restart

**Solution**:
```bash
# Always rebuild frontend before restarting:
cd webapp
yarn build
cd ..
docker-compose restart webapp

# Or use the make command:
make -f Makefile.docker webapp-restart
```

#### Issue: Build fails with memory error
```
FATAL ERROR: Reached heap limit Allocation failed - JavaScript heap out of memory
```

**Solution**:
```bash
# Increase Node.js memory limit:
export NODE_OPTIONS="--max-old-space-size=4096"
yarn build
```

### 3. Docker Issues

#### Issue: Containers fail to start
```
ERROR: for api-backend Cannot start service api-backend: driver failed programming external connectivity
```

**Solution**:
```bash
# Clean up and restart:
docker-compose down
docker system prune -f
docker-compose up -d

# If port conflicts:
lsof -i :8080  # Check what's using the port
kill -9 <PID>  # Kill the process
```

#### Issue: "No space left on device"
**Solution**:
```bash
# Clean up Docker resources:
docker system prune -a --volumes
docker image prune -a
docker volume prune
```

### 4. Database Connection Issues

#### Issue: API cannot connect to PostgreSQL
```
error: dial tcp 127.0.0.1:5432: connect: connection refused
```

**Solution**:
```bash
# Check if postgres is running:
docker-compose ps postgres

# Use correct connection string (service name, not localhost):
DATABASE_URL=postgresql://relayooor_user:password@postgres:5432/relayooor

# Not:
DATABASE_URL=postgresql://relayooor_user:password@localhost:5432/relayooor
```

#### Issue: Migration fails
```
error: pq: password authentication failed for user "relayooor_user"
```

**Solution**:
```bash
# Verify credentials match in .env and docker-compose.yml
# Reset database if needed:
docker-compose down -v
docker-compose up -d postgres
docker-compose exec api-backend go run . migrate up
```

## Service-Specific Issues

### 1. Chainpulse Issues

#### Issue: Neutron chain shows as "degraded"
**Root Cause**: Chainpulse cannot decode Neutron's ABCI++ vote extensions

**Current Status**: Known limitation, no workaround available

**Impact**: Limited packet tracking for Neutron chain

**Monitoring**:
```bash
# Check Chainpulse logs for details:
docker-compose logs chainpulse | grep -i neutron
```

#### Issue: High memory usage
**Symptoms**: Chainpulse consuming excessive memory

**Solution**:
```yaml
# Add memory limits in docker-compose.yml:
chainpulse:
  mem_limit: 2g
  memswap_limit: 2g
```

**Long-term fix**: Implement packet archival after 30 days

### 2. Hermes Issues

#### Issue: RPC connection failed
```
Error: RPC error to endpoint https://cosmos-rpc.example.com:443
```

**Solutions**:
1. **Check authentication**:
   ```toml
   # In config/hermes/config.toml
   rpc_addr = 'https://username:password@cosmos-rpc.example.com:443'
   ```

2. **Verify RPC endpoint**:
   ```bash
   curl https://cosmos-rpc.example.com:443/status
   ```

3. **Check proxy settings**:
   ```bash
   # Use proxy-enabled entrypoint:
   docker-compose exec hermes env | grep -i proxy
   ```

#### Issue: Insufficient gas
```
Error: out of gas in location: WritePerByte; gasWanted: 100000, gasUsed: 150000
```

**Solution**:
```toml
# Increase gas settings in config.toml:
gas_multiplier = 1.5  # Increase from 1.2
default_gas = 200000  # Increase from 100000
max_gas = 2000000     # Increase from 1000000
```

#### Issue: Account sequence mismatch
```
Error: account sequence mismatch, expected 123, got 125
```

**Solution**:
```bash
# Usually self-corrects, but if persistent:
docker-compose restart hermes

# Check account status:
docker exec hermes hermes keys balance --chain cosmoshub-4
```

### 3. API Issues

#### Issue: CORS errors in browser
```
Access to fetch at 'http://localhost:8080/api' from origin 'http://localhost:5173' has been blocked by CORS policy
```

**Solution**:
```go
// Ensure CORS middleware is properly configured:
router.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"http://localhost:5173", "http://localhost"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"*"},
    AllowCredentials: true,
}))
```

#### Issue: WebSocket connection drops
**Symptoms**: Real-time updates stop working

**Solutions**:
1. **Check nginx configuration**:
   ```nginx
   location /ws {
       proxy_pass http://api-backend:8080;
       proxy_http_version 1.1;
       proxy_set_header Upgrade $http_upgrade;
       proxy_set_header Connection "upgrade";
       proxy_read_timeout 3600s;  # Increase timeout
   }
   ```

2. **Implement reconnection logic**:
   ```javascript
   // Frontend WebSocket service
   class WebSocketService {
     reconnect() {
       setTimeout(() => {
         this.connect();
       }, this.reconnectDelay);
       this.reconnectDelay = Math.min(this.reconnectDelay * 2, 30000);
     }
   }
   ```

## Production Issues

### 1. Performance Issues

#### Issue: Slow API response times
**Diagnosis**:
```bash
# Check API metrics:
curl http://localhost:9090/metrics | grep api_request_duration

# Check database slow queries:
docker-compose exec postgres psql -U relayooor_user -d relayooor -c "SELECT * FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"
```

**Solutions**:
1. **Add database indexes**:
   ```sql
   CREATE INDEX CONCURRENTLY idx_packets_status_created ON packets(status, created_at);
   CREATE INDEX CONCURRENTLY idx_packets_sender ON packets(sender);
   ```

2. **Enable connection pooling**:
   ```go
   config.MaxConns = 25
   config.MinConns = 5
   ```

3. **Implement caching**:
   ```go
   // Cache frequently accessed data
   cached, err := cache.Get(key)
   if err == nil {
       return cached
   }
   ```

### 2. Security Issues

#### Issue: JWT token exposed in logs
**Solution**:
```go
// Implement log sanitization:
func sanitizeLogs(message string) string {
    // Remove JWT tokens
    re := regexp.MustCompile(`Bearer [A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+`)
    return re.ReplaceAllString(message, "Bearer [REDACTED]")
}
```

#### Issue: Rate limiting not working
**Solution**:
```go
// Implement proper rate limiting:
limiter := rate.NewLimiter(rate.Every(time.Minute/100), 100)

func rateLimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(429, gin.H{"error": "Rate limit exceeded"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

## Debugging Tools

### 1. Service Logs
```bash
# View all logs:
docker-compose logs -f

# Service-specific logs:
docker-compose logs -f api-backend
docker-compose logs -f chainpulse
docker-compose logs -f hermes

# Filter logs:
docker-compose logs api-backend | grep -i error
docker-compose logs chainpulse | grep -i "stuck packet"
```

### 2. Database Queries
```bash
# Connect to database:
docker-compose exec postgres psql -U relayooor_user -d relayooor

# Useful queries:
-- Check stuck packets
SELECT COUNT(*) FROM packets WHERE status = 'pending' AND created_at < NOW() - INTERVAL '1 hour';

-- Check recent clearing requests
SELECT * FROM clearing_requests ORDER BY created_at DESC LIMIT 10;

-- Check user activity
SELECT wallet_address, COUNT(*) as request_count 
FROM clearing_requests 
GROUP BY wallet_address 
ORDER BY request_count DESC;
```

### 3. API Testing
```bash
# Test endpoints:
# Health check
curl http://localhost:8080/health

# Get stuck packets
curl http://localhost:8080/api/chains/cosmoshub-4/packets/stuck

# Test with auth:
TOKEN="your-jwt-token"
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/users/profile
```

### 4. Network Debugging
```bash
# Check Docker networks:
docker network ls
docker network inspect relayooor_default

# Test service connectivity:
docker-compose exec api-backend ping postgres
docker-compose exec api-backend curl http://chainpulse:3000/health

# Check port bindings:
docker-compose ps
netstat -an | grep -E ":(80|8080|3000|5432|6379)"
```

## Recovery Procedures

### 1. Database Recovery
```bash
# Backup before any recovery:
docker-compose exec postgres pg_dump -U relayooor_user relayooor > backup-$(date +%Y%m%d-%H%M%S).sql

# Restore from backup:
docker-compose exec -T postgres psql -U relayooor_user relayooor < backup.sql

# Reset database completely:
docker-compose down -v
docker-compose up -d postgres
docker-compose exec api-backend go run . migrate up
```

### 2. Service Recovery
```bash
# Restart single service:
docker-compose restart api-backend

# Restart with clean state:
docker-compose down
docker-compose up -d

# Force recreate containers:
docker-compose up -d --force-recreate

# Complete reset:
docker-compose down -v
docker system prune -f
docker-compose build --no-cache
docker-compose up -d
```

## Monitoring and Alerts

### 1. Health Check Script
```bash
#!/bin/bash
# health-check.sh

SERVICES=("api-backend:8080" "chainpulse:3000" "hermes:5185")

for service in "${SERVICES[@]}"; do
    IFS=':' read -r name port <<< "$service"
    if curl -f -s "http://localhost:$port/health" > /dev/null; then
        echo "✓ $name is healthy"
    else
        echo "✗ $name is down!"
        # Send alert
        curl -X POST https://alerts.example.com/webhook -d "{\"service\":\"$name\",\"status\":\"down\"}"
    fi
done
```

### 2. Log Monitoring
```bash
# Watch for errors:
docker-compose logs -f --tail=100 | grep -i -E "(error|panic|fatal)"

# Monitor stuck packets:
watch -n 60 'docker-compose exec postgres psql -U relayooor_user -d relayooor -c "SELECT COUNT(*) as stuck_packets FROM packets WHERE status = '\''pending'\'' AND created_at < NOW() - INTERVAL '\''1 hour'\'';"'
```

## Quick Reference

### Service URLs
- Frontend: http://localhost
- API: http://localhost:8080
- Chainpulse: http://localhost:3000
- Hermes API: http://localhost:5185
- Metrics: http://localhost:3001/metrics
- PostgreSQL: localhost:5432
- Redis: localhost:6379

### Important Commands
```bash
# Start everything:
./scripts/setup-and-launch.sh

# View logs:
docker-compose logs -f [service-name]

# Restart service:
docker-compose restart [service-name]

# Run tests:
./scripts/run-tests.sh

# Check chain compatibility:
./scripts/check-chain-compatibility.sh [chain-id]

# Database console:
docker-compose exec postgres psql -U relayooor_user -d relayooor
```

### Emergency Contacts
- On-call engineer: (defined in ops runbook)
- Escalation: (defined in ops runbook)
- Status page: https://status.relayooor.com