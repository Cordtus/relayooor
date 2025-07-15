-- Performance optimization indexes for clearing operations

-- User query indexes
CREATE INDEX IF NOT EXISTS idx_clearing_operations_wallet_created 
ON clearing_operations(wallet_address, created_at DESC)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_clearing_operations_wallet_status 
ON clearing_operations(wallet_address, clearing_status, created_at DESC)
WHERE deleted_at IS NULL;

-- Token lookup indexes
CREATE INDEX IF NOT EXISTS idx_clearing_operations_token 
ON clearing_operations(clearing_token);

CREATE INDEX IF NOT EXISTS idx_clearing_operations_expires_unused 
ON clearing_operations(created_at) 
WHERE clearing_status = 'pending';

-- Payment verification indexes
CREATE INDEX IF NOT EXISTS idx_clearing_operations_payment_tx 
ON clearing_operations(payment_tx_hash)
WHERE payment_tx_hash IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_clearing_operations_clearing_tx 
ON clearing_operations(clearing_tx_hash)
WHERE clearing_tx_hash IS NOT NULL;

-- Statistics aggregation indexes
CREATE INDEX IF NOT EXISTS idx_clearing_operations_created_date 
ON clearing_operations(DATE(created_at), clearing_status)
WHERE deleted_at IS NULL;

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_clearing_operations_wallet_date_status 
ON clearing_operations(wallet_address, DATE(created_at), clearing_status)
WHERE deleted_at IS NULL;

-- Payment transactions indexes
CREATE INDEX IF NOT EXISTS idx_payment_tx_hash 
ON payment_transactions(tx_hash);

CREATE INDEX IF NOT EXISTS idx_payment_tx_created 
ON payment_transactions(created_at DESC);

-- Refund indexes (if refunds table exists)
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'refunds') THEN
        CREATE INDEX IF NOT EXISTS idx_refunds_status_created 
        ON refunds(status, created_at) 
        WHERE status = 'pending';
        
        CREATE INDEX IF NOT EXISTS idx_refunds_operation 
        ON refunds(clearing_token);
    END IF;
END $$;

-- Create function to analyze query performance
CREATE OR REPLACE FUNCTION analyze_query_performance(query_pattern text)
RETURNS TABLE(
    query_text text,
    calls bigint,
    mean_time_ms numeric,
    max_time_ms numeric,
    total_time_ms numeric
) AS $$
BEGIN
    -- Check if pg_stat_statements exists
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_stat_statements') THEN
        RETURN QUERY
        SELECT 
            s.query::text,
            s.calls,
            round(s.mean_exec_time::numeric, 2) as mean_time_ms,
            round(s.max_exec_time::numeric, 2) as max_time_ms,
            round(s.total_exec_time::numeric, 2) as total_time_ms
        FROM pg_stat_statements s
        WHERE s.query LIKE query_pattern
        ORDER BY s.mean_exec_time DESC
        LIMIT 10;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Create function to check index usage
CREATE OR REPLACE FUNCTION check_index_effectiveness()
RETURNS TABLE(
    table_name text,
    index_name text,
    index_size text,
    index_scans bigint,
    effectiveness text
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        schemaname || '.' || tablename as table_name,
        indexname as index_name,
        pg_size_pretty(pg_relation_size(indexrelid)) as index_size,
        idx_scan as index_scans,
        CASE 
            WHEN idx_scan = 0 THEN 'UNUSED'
            WHEN idx_scan < 100 THEN 'RARELY USED'
            WHEN idx_scan < 1000 THEN 'OCCASIONALLY USED'
            ELSE 'FREQUENTLY USED'
        END as effectiveness
    FROM pg_stat_user_indexes
    WHERE schemaname = 'public'
    ORDER BY idx_scan ASC;
END;
$$ LANGUAGE plpgsql;

-- Create maintenance function
CREATE OR REPLACE FUNCTION maintain_indexes() 
RETURNS void AS $$
DECLARE
    idx RECORD;
    bloat_threshold numeric := 30.0; -- 30% bloat threshold
BEGIN
    -- Update statistics on all tables
    ANALYZE clearing_operations;
    
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'payment_transactions') THEN
        ANALYZE payment_transactions;
    END IF;
    
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'refunds') THEN
        ANALYZE refunds;
    END IF;
    
    -- Reindex bloated indexes
    FOR idx IN 
        SELECT 
            schemaname,
            tablename,
            indexname,
            pg_relation_size(indexrelid) as index_size,
            pg_stat_get_live_tuples(indrelid) as live_tuples
        FROM pg_stat_user_indexes
        JOIN pg_index ON pg_index.indexrelid = pg_stat_user_indexes.indexrelid
        WHERE schemaname = 'public'
        AND pg_relation_size(indexrelid) > 10 * 1024 * 1024 -- > 10MB
    LOOP
        -- Skip if we can't calculate bloat
        CONTINUE WHEN idx.live_tuples = 0;
        
        -- Estimate bloat (very rough estimate)
        IF (idx.index_size::numeric / (idx.live_tuples * 40)) > (1 + bloat_threshold/100) THEN
            -- Use CONCURRENTLY to avoid locking
            BEGIN
                EXECUTE format('REINDEX INDEX CONCURRENTLY %I.%I', 
                    idx.schemaname, idx.indexname);
                RAISE NOTICE 'Reindexed bloated index: %.%', 
                    idx.schemaname, idx.indexname;
            EXCEPTION WHEN OTHERS THEN
                RAISE WARNING 'Failed to reindex %: %', idx.indexname, SQLERRM;
            END;
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Create a view for monitoring query performance
CREATE OR REPLACE VIEW v_clearing_operation_stats AS
SELECT 
    DATE(created_at) as date,
    clearing_status as status,
    COUNT(*) as operation_count,
    AVG(EXTRACT(EPOCH FROM (updated_at - created_at))) as avg_duration_seconds,
    SUM(ARRAY_LENGTH(packet_ids, 1)) as total_packets_cleared,
    SUM(CAST(total_fee AS NUMERIC)) as total_fees
FROM clearing_operations
WHERE deleted_at IS NULL
GROUP BY DATE(created_at), clearing_status;

-- Create a materialized view for user statistics (refreshed periodically)
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_user_statistics AS
SELECT 
    wallet_address,
    COUNT(*) as total_operations,
    COUNT(CASE WHEN clearing_status = 'completed' THEN 1 END) as successful_operations,
    COUNT(CASE WHEN clearing_status = 'failed' THEN 1 END) as failed_operations,
    SUM(ARRAY_LENGTH(packet_ids, 1)) as total_packets_cleared,
    AVG(CASE 
        WHEN clearing_status = 'completed' AND updated_at IS NOT NULL 
        THEN EXTRACT(EPOCH FROM (updated_at - created_at))
        ELSE NULL 
    END) as avg_completion_time,
    SUM(CAST(COALESCE(total_fee, 0) AS NUMERIC)) as total_fees_paid,
    MAX(created_at) as last_operation_at
FROM clearing_operations
WHERE deleted_at IS NULL
GROUP BY wallet_address;

-- Create index on materialized view
CREATE INDEX IF NOT EXISTS idx_mv_user_statistics_wallet 
ON mv_user_statistics(wallet_address);

-- Grant permissions
GRANT SELECT ON v_clearing_operation_stats TO PUBLIC;
GRANT SELECT ON mv_user_statistics TO PUBLIC;

-- Schedule maintenance (if pg_cron is available)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_cron') THEN
        -- Schedule index maintenance weekly
        PERFORM cron.schedule('index-maintenance', '0 3 * * 0', 'SELECT maintain_indexes()');
        
        -- Schedule materialized view refresh hourly
        PERFORM cron.schedule('refresh-user-stats', '0 * * * *', 
            'REFRESH MATERIALIZED VIEW CONCURRENTLY mv_user_statistics');
    END IF;
END $$;