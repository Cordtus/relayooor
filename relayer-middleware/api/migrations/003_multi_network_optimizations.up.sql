-- Multi-network database optimizations for scaling
-- This migration adds support for efficient querying across multiple chains

-- Add network-specific partitioning for clearing operations
-- This allows efficient querying by chain while maintaining data locality

-- Create network_stats table for aggregated metrics per chain
CREATE TABLE IF NOT EXISTS network_stats (
    id SERIAL PRIMARY KEY,
    chain_id VARCHAR(50) NOT NULL,
    network_name VARCHAR(100) NOT NULL,
    date DATE NOT NULL,
    total_packets BIGINT DEFAULT 0,
    successful_packets BIGINT DEFAULT 0,
    failed_packets BIGINT DEFAULT 0,
    stuck_packets BIGINT DEFAULT 0,
    total_volume_usd NUMERIC(20,2) DEFAULT 0,
    avg_clear_time_seconds NUMERIC(10,2),
    active_relayers INTEGER DEFAULT 0,
    active_channels INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(chain_id, date)
);

-- Create channel_performance table for tracking channel-specific metrics
CREATE TABLE IF NOT EXISTS channel_performance (
    id SERIAL PRIMARY KEY,
    src_chain_id VARCHAR(50) NOT NULL,
    dst_chain_id VARCHAR(50) NOT NULL,
    src_channel_id VARCHAR(50) NOT NULL,
    dst_channel_id VARCHAR(50) NOT NULL,
    date DATE NOT NULL,
    hour INTEGER NOT NULL CHECK (hour >= 0 AND hour < 24),
    packets_sent INTEGER DEFAULT 0,
    packets_cleared INTEGER DEFAULT 0,
    packets_stuck INTEGER DEFAULT 0,
    avg_clear_time_seconds NUMERIC(10,2),
    congestion_score NUMERIC(5,2) DEFAULT 0 CHECK (congestion_score >= 0 AND congestion_score <= 100),
    success_rate NUMERIC(5,2) DEFAULT 0 CHECK (success_rate >= 0 AND success_rate <= 100),
    volume_usd NUMERIC(20,2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(src_chain_id, dst_chain_id, src_channel_id, dst_channel_id, date, hour)
);

-- Create relayer_performance table for tracking relayer metrics
CREATE TABLE IF NOT EXISTS relayer_performance (
    id SERIAL PRIMARY KEY,
    relayer_address VARCHAR(100) NOT NULL,
    chain_id VARCHAR(50) NOT NULL,
    date DATE NOT NULL,
    packets_relayed INTEGER DEFAULT 0,
    packets_won INTEGER DEFAULT 0,
    packets_frontrun INTEGER DEFAULT 0,
    success_rate NUMERIC(5,2) DEFAULT 0,
    avg_response_time_seconds NUMERIC(10,2),
    total_fees_earned NUMERIC(20,2) DEFAULT 0,
    reputation_score NUMERIC(5,2) DEFAULT 50 CHECK (reputation_score >= 0 AND reputation_score <= 100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(relayer_address, chain_id, date)
);

-- Create packet_flow table for real-time packet tracking
CREATE TABLE IF NOT EXISTS packet_flow (
    id BIGSERIAL PRIMARY KEY,
    packet_id VARCHAR(200) NOT NULL,
    src_chain_id VARCHAR(50) NOT NULL,
    dst_chain_id VARCHAR(50) NOT NULL,
    src_channel_id VARCHAR(50) NOT NULL,
    dst_channel_id VARCHAR(50) NOT NULL,
    sequence BIGINT NOT NULL,
    sender VARCHAR(100) NOT NULL,
    receiver VARCHAR(100) NOT NULL,
    amount NUMERIC(40,0) NOT NULL,
    denom VARCHAR(50) NOT NULL,
    amount_usd NUMERIC(20,2),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    sent_at TIMESTAMP WITH TIME ZONE,
    received_at TIMESTAMP WITH TIME ZONE,
    stuck_at TIMESTAMP WITH TIME ZONE,
    cleared_at TIMESTAMP WITH TIME ZONE,
    relayer_address VARCHAR(100),
    tx_hash_send VARCHAR(100),
    tx_hash_recv VARCHAR(100),
    tx_hash_clear VARCHAR(100),
    attempts INTEGER DEFAULT 0,
    error_message TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    UNIQUE(src_chain_id, src_channel_id, sequence)
);

-- Add chain_id to clearing_operations if not exists
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name='clearing_operations' AND column_name='chain_id') THEN
        ALTER TABLE clearing_operations ADD COLUMN chain_id VARCHAR(50);
    END IF;
END $$;

-- Create indexes for optimal query performance

-- Network stats indexes
CREATE INDEX IF NOT EXISTS idx_network_stats_chain_date 
ON network_stats(chain_id, date DESC);

CREATE INDEX IF NOT EXISTS idx_network_stats_date 
ON network_stats(date DESC);

-- Channel performance indexes
CREATE INDEX IF NOT EXISTS idx_channel_performance_route 
ON channel_performance(src_chain_id, dst_chain_id, date DESC);

CREATE INDEX IF NOT EXISTS idx_channel_performance_channel 
ON channel_performance(src_channel_id, dst_channel_id, date DESC);

CREATE INDEX IF NOT EXISTS idx_channel_performance_congestion 
ON channel_performance(congestion_score DESC, date DESC)
WHERE congestion_score > 50;

-- Relayer performance indexes
CREATE INDEX IF NOT EXISTS idx_relayer_performance_address_date 
ON relayer_performance(relayer_address, date DESC);

CREATE INDEX IF NOT EXISTS idx_relayer_performance_chain_date 
ON relayer_performance(chain_id, date DESC);

CREATE INDEX IF NOT EXISTS idx_relayer_performance_reputation 
ON relayer_performance(reputation_score DESC, date DESC);

-- Packet flow indexes
CREATE INDEX IF NOT EXISTS idx_packet_flow_status_created 
ON packet_flow(status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_packet_flow_sender 
ON packet_flow(sender, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_packet_flow_receiver 
ON packet_flow(receiver, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_packet_flow_stuck 
ON packet_flow(stuck_at DESC)
WHERE status = 'stuck';

CREATE INDEX IF NOT EXISTS idx_packet_flow_relayer 
ON packet_flow(relayer_address, cleared_at DESC)
WHERE relayer_address IS NOT NULL;

-- Composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_packet_flow_route_status 
ON packet_flow(src_chain_id, dst_chain_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_packet_flow_channel_sequence 
ON packet_flow(src_channel_id, sequence);

-- Create hypertable for time-series data if TimescaleDB is available
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'timescaledb') THEN
        -- Convert packet_flow to hypertable
        PERFORM create_hypertable('packet_flow', 'created_at', 
            chunk_time_interval => INTERVAL '1 day',
            if_not_exists => TRUE);
        
        -- Add compression policy
        PERFORM add_compression_policy('packet_flow', INTERVAL '7 days');
        
        -- Create continuous aggregates for real-time analytics
        EXECUTE 'CREATE MATERIALIZED VIEW IF NOT EXISTS hourly_packet_stats
        WITH (timescaledb.continuous) AS
        SELECT 
            time_bucket(''1 hour'', created_at) AS hour,
            src_chain_id,
            dst_chain_id,
            status,
            COUNT(*) as packet_count,
            SUM(amount_usd) as volume_usd,
            AVG(EXTRACT(EPOCH FROM (cleared_at - created_at))) as avg_clear_time
        FROM packet_flow
        GROUP BY hour, src_chain_id, dst_chain_id, status
        WITH NO DATA';
        
        -- Add refresh policy
        PERFORM add_continuous_aggregate_policy('hourly_packet_stats',
            start_offset => INTERVAL '3 hours',
            end_offset => INTERVAL '1 hour',
            schedule_interval => INTERVAL '1 hour');
    END IF;
END $$;

-- Create functions for efficient data aggregation

-- Function to get network health score
CREATE OR REPLACE FUNCTION calculate_network_health_score(p_chain_id VARCHAR)
RETURNS NUMERIC AS $$
DECLARE
    health_score NUMERIC;
    success_rate NUMERIC;
    congestion_level NUMERIC;
    active_channels INTEGER;
BEGIN
    -- Get latest metrics
    SELECT 
        CASE WHEN total_packets > 0 
             THEN (successful_packets::NUMERIC / total_packets) * 100 
             ELSE 0 END,
        CASE WHEN total_packets > 0 
             THEN (stuck_packets::NUMERIC / total_packets) * 100 
             ELSE 0 END,
        active_channels
    INTO success_rate, congestion_level, active_channels
    FROM network_stats
    WHERE chain_id = p_chain_id
    AND date = CURRENT_DATE
    LIMIT 1;
    
    -- Calculate health score (0-100)
    health_score := GREATEST(0, LEAST(100,
        (COALESCE(success_rate, 0) * 0.5) +
        ((100 - COALESCE(congestion_level, 0)) * 0.3) +
        (LEAST(COALESCE(active_channels, 0) * 2, 20))
    ));
    
    RETURN ROUND(health_score, 2);
END;
$$ LANGUAGE plpgsql;

-- Function to identify congested channels
CREATE OR REPLACE FUNCTION identify_congested_channels(
    p_threshold NUMERIC DEFAULT 50,
    p_limit INTEGER DEFAULT 10
)
RETURNS TABLE(
    route VARCHAR,
    channel_pair VARCHAR,
    congestion_score NUMERIC,
    stuck_packets INTEGER,
    avg_wait_time INTERVAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cp.src_chain_id || ' → ' || cp.dst_chain_id as route,
        cp.src_channel_id || ' ↔ ' || cp.dst_channel_id as channel_pair,
        cp.congestion_score,
        cp.packets_stuck,
        INTERVAL '1 second' * cp.avg_clear_time_seconds as avg_wait_time
    FROM channel_performance cp
    WHERE cp.date = CURRENT_DATE
    AND cp.congestion_score > p_threshold
    ORDER BY cp.congestion_score DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- Function to get relayer leaderboard
CREATE OR REPLACE FUNCTION get_relayer_leaderboard(
    p_chain_id VARCHAR DEFAULT NULL,
    p_date DATE DEFAULT CURRENT_DATE,
    p_limit INTEGER DEFAULT 20
)
RETURNS TABLE(
    rank INTEGER,
    relayer_address VARCHAR,
    packets_relayed INTEGER,
    success_rate NUMERIC,
    reputation_score NUMERIC,
    fees_earned NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ROW_NUMBER() OVER (ORDER BY rp.packets_relayed DESC) as rank,
        rp.relayer_address,
        rp.packets_relayed,
        rp.success_rate,
        rp.reputation_score,
        rp.total_fees_earned
    FROM relayer_performance rp
    WHERE rp.date = p_date
    AND (p_chain_id IS NULL OR rp.chain_id = p_chain_id)
    ORDER BY rp.packets_relayed DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- Create optimized views for common queries

-- Real-time stuck packets view
CREATE OR REPLACE VIEW v_stuck_packets AS
SELECT 
    pf.packet_id,
    pf.src_chain_id,
    pf.dst_chain_id,
    pf.src_channel_id || ' → ' || pf.dst_channel_id as channel,
    pf.sequence,
    pf.sender,
    pf.receiver,
    pf.amount,
    pf.denom,
    pf.amount_usd,
    pf.stuck_at,
    CURRENT_TIMESTAMP - pf.stuck_at as stuck_duration,
    pf.attempts,
    pf.error_message
FROM packet_flow pf
WHERE pf.status = 'stuck'
AND pf.stuck_at IS NOT NULL
ORDER BY pf.stuck_at ASC;

-- Network overview dashboard view
CREATE OR REPLACE VIEW v_network_overview AS
SELECT 
    ns.chain_id,
    ns.network_name,
    ns.total_packets,
    ns.successful_packets,
    ns.stuck_packets,
    CASE WHEN ns.total_packets > 0 
         THEN ROUND((ns.successful_packets::NUMERIC / ns.total_packets) * 100, 2)
         ELSE 0 END as success_rate,
    ns.total_volume_usd,
    ns.active_relayers,
    ns.active_channels,
    calculate_network_health_score(ns.chain_id) as health_score
FROM network_stats ns
WHERE ns.date = CURRENT_DATE
ORDER BY ns.total_packets DESC;

-- Grant permissions
GRANT SELECT ON ALL TABLES IN SCHEMA public TO PUBLIC;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO PUBLIC;

-- Create maintenance procedures

-- Procedure to aggregate daily stats
CREATE OR REPLACE PROCEDURE aggregate_daily_stats()
LANGUAGE plpgsql
AS $$
BEGIN
    -- Aggregate network stats
    INSERT INTO network_stats (
        chain_id, network_name, date, total_packets, successful_packets,
        failed_packets, stuck_packets, total_volume_usd, avg_clear_time_seconds,
        active_relayers, active_channels
    )
    SELECT 
        src_chain_id as chain_id,
        src_chain_id as network_name,
        CURRENT_DATE as date,
        COUNT(*) as total_packets,
        COUNT(CASE WHEN status = 'cleared' THEN 1 END) as successful_packets,
        COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_packets,
        COUNT(CASE WHEN status = 'stuck' THEN 1 END) as stuck_packets,
        SUM(amount_usd) as total_volume_usd,
        AVG(EXTRACT(EPOCH FROM (cleared_at - created_at))) as avg_clear_time_seconds,
        COUNT(DISTINCT relayer_address) as active_relayers,
        COUNT(DISTINCT src_channel_id || '-' || dst_channel_id) as active_channels
    FROM packet_flow
    WHERE DATE(created_at) = CURRENT_DATE
    GROUP BY src_chain_id
    ON CONFLICT (chain_id, date) 
    DO UPDATE SET
        total_packets = EXCLUDED.total_packets,
        successful_packets = EXCLUDED.successful_packets,
        failed_packets = EXCLUDED.failed_packets,
        stuck_packets = EXCLUDED.stuck_packets,
        total_volume_usd = EXCLUDED.total_volume_usd,
        avg_clear_time_seconds = EXCLUDED.avg_clear_time_seconds,
        active_relayers = EXCLUDED.active_relayers,
        active_channels = EXCLUDED.active_channels,
        updated_at = CURRENT_TIMESTAMP;
    
    -- Log completion
    RAISE NOTICE 'Daily stats aggregation completed at %', CURRENT_TIMESTAMP;
END;
$$;

-- Schedule daily aggregation if pg_cron is available
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_cron') THEN
        PERFORM cron.schedule('aggregate-daily-stats', '0 1 * * *', 'CALL aggregate_daily_stats()');
    END IF;
END $$;