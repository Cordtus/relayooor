-- Drop all tables and functions
DROP TRIGGER IF EXISTS update_clearing_operations_updated_at ON clearing_operations;
DROP TRIGGER IF EXISTS update_refunds_updated_at ON refunds;
DROP TRIGGER IF EXISTS update_user_statistics_updated_at ON user_statistics;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS clearing_errors;
DROP TABLE IF EXISTS user_statistics;
DROP TABLE IF EXISTS refunds;
DROP TABLE IF EXISTS payment_transactions;
DROP TABLE IF EXISTS clearing_operations;