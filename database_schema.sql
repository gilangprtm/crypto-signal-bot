-- CryptoSignal AI Database Schema
-- PostgreSQL 15+

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    telegram_id VARCHAR(50) UNIQUE NOT NULL,
    username VARCHAR(100),
    email VARCHAR(255),
    subscription_tier VARCHAR(20) DEFAULT 'free' CHECK (subscription_tier IN ('free', 'premium', 'vip')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true,
    preferences JSONB DEFAULT '{}',
    
    -- Indexes
    INDEX idx_users_telegram_id (telegram_id),
    INDEX idx_users_subscription_tier (subscription_tier),
    INDEX idx_users_created_at (created_at)
);

-- Cryptocurrencies table
CREATE TABLE cryptocurrencies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    symbol VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    coingecko_id VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Indexes
    INDEX idx_crypto_symbol (symbol),
    INDEX idx_crypto_active (is_active)
);

-- Trading signals table
CREATE TABLE trading_signals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    crypto_id UUID NOT NULL REFERENCES cryptocurrencies(id),
    action VARCHAR(10) NOT NULL CHECK (action IN ('BUY', 'SELL', 'HOLD')),
    confidence_score DECIMAL(5,4) NOT NULL CHECK (confidence_score >= 0 AND confidence_score <= 1),
    entry_price DECIMAL(20,8) NOT NULL,
    stop_loss DECIMAL(20,8),
    take_profit_levels JSONB DEFAULT '[]',
    reasoning TEXT,
    technical_indicators JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'expired', 'triggered')),
    
    -- Indexes
    INDEX idx_signals_crypto_id (crypto_id),
    INDEX idx_signals_action (action),
    INDEX idx_signals_confidence (confidence_score),
    INDEX idx_signals_created_at (created_at),
    INDEX idx_signals_status (status)
);

-- Signal distributions table
CREATE TABLE signal_distributions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    signal_id UUID NOT NULL REFERENCES trading_signals(id),
    user_id UUID NOT NULL REFERENCES users(id),
    channel_type VARCHAR(20) NOT NULL CHECK (channel_type IN ('telegram', 'discord', 'email', 'webhook')),
    channel_id VARCHAR(255) NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    delivery_status VARCHAR(20) DEFAULT 'pending' CHECK (delivery_status IN ('pending', 'sent', 'failed', 'delivered')),
    metadata JSONB DEFAULT '{}',
    
    -- Indexes
    INDEX idx_distributions_signal_id (signal_id),
    INDEX idx_distributions_user_id (user_id),
    INDEX idx_distributions_channel_type (channel_type),
    INDEX idx_distributions_sent_at (sent_at),
    INDEX idx_distributions_status (delivery_status)
);

-- User feedback table
CREATE TABLE user_feedback (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id),
    signal_id UUID NOT NULL REFERENCES trading_signals(id),
    feedback_type VARCHAR(20) NOT NULL CHECK (feedback_type IN ('rating', 'comment', 'result')),
    rating DECIMAL(3,2) CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Indexes
    INDEX idx_feedback_user_id (user_id),
    INDEX idx_feedback_signal_id (signal_id),
    INDEX idx_feedback_type (feedback_type),
    INDEX idx_feedback_created_at (created_at)
);

-- Market data table (time series)
CREATE TABLE market_data (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    crypto_id UUID NOT NULL REFERENCES cryptocurrencies(id),
    price DECIMAL(20,8) NOT NULL,
    volume DECIMAL(30,8),
    market_cap DECIMAL(30,8),
    ohlcv JSONB DEFAULT '{}',
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Indexes
    INDEX idx_market_data_crypto_id (crypto_id),
    INDEX idx_market_data_timestamp (timestamp),
    INDEX idx_market_data_crypto_timestamp (crypto_id, timestamp)
);

-- Performance metrics table
CREATE TABLE performance_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    signal_id UUID NOT NULL REFERENCES trading_signals(id),
    entry_price DECIMAL(20,8) NOT NULL,
    exit_price DECIMAL(20,8),
    pnl_percentage DECIMAL(10,4),
    entry_time TIMESTAMP WITH TIME ZONE NOT NULL,
    exit_time TIMESTAMP WITH TIME ZONE,
    outcome VARCHAR(20) CHECK (outcome IN ('profit', 'loss', 'breakeven', 'pending')),
    
    -- Indexes
    INDEX idx_performance_signal_id (signal_id),
    INDEX idx_performance_outcome (outcome),
    INDEX idx_performance_entry_time (entry_time)
);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to users table
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Insert sample cryptocurrencies
INSERT INTO cryptocurrencies (symbol, name, coingecko_id) VALUES
('BTC', 'Bitcoin', 'bitcoin'),
('ETH', 'Ethereum', 'ethereum'),
('ADA', 'Cardano', 'cardano'),
('SOL', 'Solana', 'solana'),
('DOT', 'Polkadot', 'polkadot'),
('LINK', 'Chainlink', 'chainlink'),
('MATIC', 'Polygon', 'matic-network'),
('AVAX', 'Avalanche', 'avalanche-2'),
('ATOM', 'Cosmos', 'cosmos'),
('NEAR', 'NEAR Protocol', 'near');
