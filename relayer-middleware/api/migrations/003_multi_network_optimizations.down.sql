-- Drop multi-network optimization tables and indexes
DROP FUNCTION IF EXISTS analyze_query_performance;
DROP FUNCTION IF EXISTS check_index_effectiveness;
DROP FUNCTION IF EXISTS calculate_network_health_score;
DROP FUNCTION IF EXISTS get_relayer_reliability_score;
DROP MATERIALIZED VIEW IF EXISTS mv_user_statistics;
DROP TABLE IF EXISTS packet_flow;
DROP TABLE IF EXISTS relayer_performance;
DROP TABLE IF EXISTS channel_performance;
DROP TABLE IF EXISTS network_stats;