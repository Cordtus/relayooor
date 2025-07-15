-- Initial schema for Relayooor packet clearing platform

-- Create clearing_operations table
CREATE TABLE IF NOT EXISTS clearing_operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_address VARCHAR(255) NOT NULL,
    clearing_token VARCHAR(255) UNIQUE NOT NULL,
    packet_ids TEXT[] NOT NULL,
    total_fee BIGINT NOT NULL,
    payment_tx_hash VARCHAR(255),
    payment_verified BOOLEAN DEFAULT FALSE,
    clearing_status VARCHAR(50) DEFAULT 'pending',
    clearing_tx_hash VARCHAR(255),
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Create payment_transactions table
CREATE TABLE IF NOT EXISTS payment_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clearing_token VARCHAR(255) NOT NULL,
    tx_hash VARCHAR(255) UNIQUE NOT NULL,
    sender_address VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL,
    memo TEXT NOT NULL,
    verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (clearing_token) REFERENCES clearing_operations(clearing_token)
);

-- Create refunds table
CREATE TABLE IF NOT EXISTS refunds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clearing_token VARCHAR(255) NOT NULL,
    reason VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL,
    tx_hash VARCHAR(255),
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (clearing_token) REFERENCES clearing_operations(clearing_token)
);

-- Create user_statistics table
CREATE TABLE IF NOT EXISTS user_statistics (
    wallet_address VARCHAR(255) PRIMARY KEY,
    total_packets_cleared INTEGER DEFAULT 0,
    total_amount_cleared BIGINT DEFAULT 0,
    success_rate DECIMAL(5,2) DEFAULT 0,
    last_clearing_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create clearing_errors table for debugging
CREATE TABLE IF NOT EXISTS clearing_errors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    clearing_token VARCHAR(255),
    error_type VARCHAR(100) NOT NULL,
    error_message TEXT NOT NULL,
    packet_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (clearing_token) REFERENCES clearing_operations(clearing_token)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_clearing_operations_wallet ON clearing_operations(wallet_address);
CREATE INDEX IF NOT EXISTS idx_clearing_operations_status ON clearing_operations(clearing_status);
CREATE INDEX IF NOT EXISTS idx_clearing_operations_created ON clearing_operations(created_at);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_sender ON payment_transactions(sender_address);
CREATE INDEX IF NOT EXISTS idx_refunds_status ON refunds(status);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_clearing_operations_updated_at BEFORE UPDATE ON clearing_operations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_refunds_updated_at BEFORE UPDATE ON refunds
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_statistics_updated_at BEFORE UPDATE ON user_statistics
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();