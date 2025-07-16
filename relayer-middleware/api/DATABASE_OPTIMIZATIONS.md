# Database Optimizations Guide

## Overview

This document describes the database optimizations implemented for handling multiple IBC networks at scale.

## Architecture

### Core Tables

1. **clearing_operations** - Main transaction table
   - Optimized with targeted indexes for wallet queries
   - Soft delete support with `deleted_at`
   - Automatic timestamp tracking

2. **network_stats** - Aggregated metrics per chain
   - Daily rollups for performance
   - Unique constraint on (chain_id, date)
   - Pre-calculated success rates

3. **channel_performance** - Channel-specific metrics
   - Hourly granularity for detailed analysis
   - Congestion scoring (0-100)
   - Success rate tracking

4. **relayer_performance** - Relayer metrics and reputation
   - Daily performance snapshots
   - Reputation scoring system
   - Frontrun tracking

5. **packet_flow** - Real-time packet tracking
   - Time-series optimized with TimescaleDB support
   - Status-based partitioning ready
   - Comprehensive indexing

### Index Strategy

#### Primary Indexes
```sql
-- User queries (most common)
idx_clearing_operations_wallet_created
idx_clearing_operations_wallet_status

-- Time-based queries
idx_packet_flow_status_created
idx_network_stats_chain_date

-- Lookup queries
idx_packet_flow_channel_sequence
idx_relayer_performance_reputation
```

#### Composite Indexes
- Multi-column indexes for complex queries
- Covering indexes to avoid table lookups
- Partial indexes for filtered queries

#### Performance Indexes
- CONCURRENTLY created to avoid locking
- Automatic bloat detection and reindexing
- Usage statistics tracking

### Connection Pooling

Using pgx/v5 for optimal performance:

```go
config := &PoolConfig{
    MaxConnections:    25,
    MinConnections:    5,
    MaxConnLifetime:   time.Hour,
    MaxConnIdleTime:   30 * time.Minute,
    HealthCheckPeriod: time.Minute,
}
```

Features:
- Automatic health checks
- Statement caching
- Slow query logging
- Connection lifecycle management

### Query Optimization

#### Query Builder
Type-safe query construction:
```go
query := NewQueryBuilder("packet_flow").
    Select("*").
    Where("status = $1", "stuck").
    Where("created_at > $2", time.Now().Add(-24*time.Hour)).
    OrderBy("created_at", true).
    Limit(100)
```

#### Batch Operations
Efficient bulk inserts:
```go
builder := NewBatchInsertBuilder("network_stats", 
    "chain_id", "date", "total_packets").
    AddRow("osmosis-1", today, 1000).
    AddRow("cosmoshub-4", today, 500).
    OnConflict("(chain_id, date) DO UPDATE SET total_packets = EXCLUDED.total_packets")
```

#### Query Caching
Simple TTL-based caching for repeated queries:
```go
cache := NewQueryCache(5 * time.Minute)
if data, ok := cache.Get(cacheKey); ok {
    return data
}
```

## Performance Features

### 1. Materialized Views
Pre-computed aggregations updated hourly:
- `mv_user_statistics` - User performance metrics
- `hourly_packet_stats` - Time-bucketed packet data (TimescaleDB)

### 2. Continuous Aggregates
Real-time analytics with TimescaleDB:
```sql
CREATE MATERIALIZED VIEW hourly_packet_stats
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 hour', created_at) AS hour,
    src_chain_id,
    COUNT(*) as packet_count,
    AVG(clear_time) as avg_clear_time
FROM packet_flow
GROUP BY hour, src_chain_id;
```

### 3. Automatic Maintenance
Scheduled with pg_cron:
- Index bloat detection and reindexing
- Statistics updates
- Old data archival
- Materialized view refreshes

### 4. Monitoring Functions
Built-in performance monitoring:
```sql
-- Check index effectiveness
SELECT * FROM check_index_effectiveness();

-- Analyze query performance
SELECT * FROM analyze_query_performance('%clearing%');

-- Calculate network health
SELECT calculate_network_health_score('osmosis-1');
```

## Best Practices

### 1. Query Patterns
- Always use prepared statements
- Leverage indexes with WHERE clause order
- Use LIMIT for large result sets
- Batch similar operations

### 2. Connection Management
- Use connection pooling
- Set appropriate timeouts
- Monitor pool statistics
- Handle connection errors gracefully

### 3. Data Modeling
- Normalize for consistency
- Denormalize for performance (carefully)
- Use appropriate data types
- Plan for partitioning

### 4. Monitoring
- Track slow queries
- Monitor index usage
- Watch for lock contention
- Set up alerts for anomalies

## Migration Guide

### Running Migrations
```bash
# Run all migrations
migrate -path migrations -database $DATABASE_URL up

# Run specific migration
migrate -path migrations -database $DATABASE_URL up 3

# Rollback
migrate -path migrations -database $DATABASE_URL down 1
```

### Adding New Indexes
Always use CONCURRENTLY:
```sql
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_new_index 
ON table_name(column_name);
```

### Schema Changes
1. Add new columns as nullable first
2. Backfill data in batches
3. Add constraints after backfill
4. Update application code
5. Remove old columns later

## Performance Tuning

### PostgreSQL Configuration
Recommended settings for production:
```ini
# Memory
shared_buffers = 25% of RAM
effective_cache_size = 75% of RAM
work_mem = RAM / max_connections / 4

# Connections
max_connections = 100
max_prepared_transactions = max_connections

# Write performance
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100

# Query planning
random_page_cost = 1.1  # For SSD
effective_io_concurrency = 200  # For SSD
```

### Monitoring Queries
```sql
-- Find slow queries
SELECT * FROM pg_stat_statements 
WHERE mean_exec_time > 1000 
ORDER BY mean_exec_time DESC;

-- Check table bloat
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) AS external_size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Monitor locks
SELECT * FROM pg_locks WHERE NOT granted;
```

## Troubleshooting

### Common Issues

1. **Slow Queries**
   - Check EXPLAIN ANALYZE output
   - Verify indexes are being used
   - Update table statistics
   - Consider query rewrite

2. **Connection Exhaustion**
   - Increase pool size carefully
   - Check for connection leaks
   - Monitor long-running transactions
   - Use connection timeout

3. **Lock Contention**
   - Identify blocking queries
   - Use SKIP LOCKED where appropriate
   - Consider advisory locks
   - Batch updates

4. **Disk Space**
   - Monitor WAL growth
   - Set up log rotation
   - Archive old data
   - Use table partitioning

### Emergency Procedures

```sql
-- Kill long-running queries
SELECT pg_terminate_backend(pid) 
FROM pg_stat_activity 
WHERE state = 'active' 
AND query_start < NOW() - INTERVAL '1 hour';

-- Force checkpoint
CHECKPOINT;

-- Vacuum specific table
VACUUM (VERBOSE, ANALYZE) table_name;

-- Reindex table online
REINDEX TABLE CONCURRENTLY table_name;
```

## Future Optimizations

1. **Partitioning**: Implement time-based partitioning for packet_flow
2. **Read Replicas**: Add read replicas for analytics queries
3. **Caching Layer**: Implement Redis for hot data
4. **Column Store**: Consider columnar storage for analytics
5. **Event Streaming**: Add Kafka for real-time processing