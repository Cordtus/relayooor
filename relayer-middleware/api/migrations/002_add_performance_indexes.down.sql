-- Drop performance indexes
DROP INDEX IF EXISTS idx_clearing_operations_wallet_created;
DROP INDEX IF EXISTS idx_clearing_operations_wallet_status;
DROP INDEX IF EXISTS idx_payment_tx_created;
DROP INDEX IF EXISTS idx_refunds_token_status;