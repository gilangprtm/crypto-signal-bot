-- FRESH DATABASE SCHEMA FOR CRYPTO SIGNAL BOT
-- This will drop all existing tables and create new ones from scratch
-- WARNING: This will delete all existing data!
-- Run this entire script in Supabase SQL Editor

-- =====================================================
-- 1. DROP ALL EXISTING TABLES (CASCADE to handle dependencies)
-- =====================================================

DROP TABLE IF EXISTS learning_data CASCADE;
DROP TABLE IF EXISTS signal_performance CASCADE;
DROP TABLE IF EXISTS notification_logs CASCADE;
DROP TABLE IF EXISTS system_logs CASCADE;
DROP TABLE IF EXISTS bot_settings CASCADE;
DROP TABLE IF EXISTS trading_signals CASCADE;
DROP TABLE IF EXISTS market_snapshots CASCADE;
DROP TABLE IF EXISTS cryptocurrencies CASCADE;

-- =====================================================
-- 2. CREATE FRESH SCHEMA
-- =====================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create cryptocurrencies table
CREATE TABLE cryptocurrencies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    symbol VARCHAR(10) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    cmc_id INTEGER UNIQUE,
    slug VARCHAR(100),
    contract_address TEXT,
    platform VARCHAR(50),
    coingecko_id VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create market_snapshots table
CREATE TABLE market_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cryptocurrency_id UUID NOT NULL REFERENCES cryptocurrencies(id),
    price DECIMAL(20,8) NOT NULL,
    volume_24h DECIMAL(20,2),
    market_cap DECIMAL(20,2),
    price_change_1h DECIMAL(10,4),
    price_change_24h DECIMAL(10,4),
    price_change_7d DECIMAL(10,4),
    rsi DECIMAL(5,2),
    macd_line DECIMAL(15,8),
    macd_signal DECIMAL(15,8),
    macd_histogram DECIMAL(15,8),
    bb_upper DECIMAL(20,8),
    bb_middle DECIMAL(20,8),
    bb_lower DECIMAL(20,8),
    sma_20 DECIMAL(20,8),
    ema_12 DECIMAL(20,8),
    ema_26 DECIMAL(20,8),
    fear_greed_index INTEGER,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create trading_signals table
CREATE TABLE trading_signals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cryptocurrency_id UUID NOT NULL REFERENCES cryptocurrencies(id),
    action VARCHAR(10) NOT NULL CHECK (action IN ('BUY', 'SELL', 'HOLD')),
    confidence_score DECIMAL(3,2) NOT NULL CHECK (confidence_score >= 0 AND confidence_score <= 1),
    entry_price DECIMAL(20,8) NOT NULL,
    stop_loss DECIMAL(20,8),
    take_profit_1 DECIMAL(20,8),
    take_profit_2 DECIMAL(20,8),
    reasoning TEXT,
    timeframe VARCHAR(10) DEFAULT '15m',
    rsi DECIMAL(5,2),
    macd_line DECIMAL(15,8),
    macd_signal DECIMAL(15,8),
    macd_histogram DECIMAL(15,8),
    bb_upper DECIMAL(20,8),
    bb_middle DECIMAL(20,8),
    bb_lower DECIMAL(20,8),
    sma_20 DECIMAL(20,8),
    ema_12 DECIMAL(20,8),
    ema_26 DECIMAL(20,8),
    volume_24h DECIMAL(20,2),
    price_change_24h DECIMAL(10,4),
    fear_greed_index INTEGER,
    market_cap DECIMAL(20,2),
    market_conditions JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'triggered', 'expired', 'cancelled')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    triggered_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ
);

-- Create signal_performance table
CREATE TABLE signal_performance (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    signal_id UUID NOT NULL REFERENCES trading_signals(id),
    entry_price DECIMAL(20,8) NOT NULL,
    exit_price DECIMAL(20,8),
    profit_loss_percentage DECIMAL(10,4),
    profit_loss_amount DECIMAL(20,8),
    execution_type VARCHAR(20) CHECK (execution_type IN ('take_profit_1', 'take_profit_2', 'stop_loss', 'manual', 'expired')),
    duration_hours INTEGER,
    entry_time TIMESTAMPTZ NOT NULL,
    exit_time TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create learning_data table
CREATE TABLE learning_data (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    signal_id UUID NOT NULL REFERENCES trading_signals(id),
    features JSONB NOT NULL,
    actual_outcome VARCHAR(20) NOT NULL CHECK (actual_outcome IN ('profit', 'loss', 'neutral')),
    profit_loss_percentage DECIMAL(10,4),
    model_version VARCHAR(20),
    confidence_score DECIMAL(3,2),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create notification_logs table
CREATE TABLE notification_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    notification_type VARCHAR(20) NOT NULL CHECK (notification_type IN ('signal', 'performance', 'daily_summary', 'error')),
    recipient VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    signal_id UUID REFERENCES trading_signals(id),
    cryptocurrency_id UUID REFERENCES cryptocurrencies(id),
    status VARCHAR(20) DEFAULT 'sent' CHECK (status IN ('sent', 'failed', 'pending')),
    error_message TEXT,
    sent_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create system_logs table
CREATE TABLE system_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    log_level VARCHAR(10) NOT NULL CHECK (log_level IN ('debug', 'info', 'warning', 'error', 'fatal')),
    component VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    context JSONB DEFAULT '{}',
    error_stack TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create bot_settings table
CREATE TABLE bot_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    setting_key VARCHAR(100) NOT NULL UNIQUE,
    setting_value TEXT NOT NULL,
    description TEXT,
    data_type VARCHAR(20) DEFAULT 'string' CHECK (data_type IN ('string', 'integer', 'float', 'boolean', 'json')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- =====================================================
-- 3. CREATE INDEXES FOR PERFORMANCE
-- =====================================================

CREATE INDEX idx_market_snapshots_crypto_timestamp ON market_snapshots(cryptocurrency_id, timestamp);
CREATE INDEX idx_market_snapshots_timestamp ON market_snapshots(timestamp);
CREATE INDEX idx_trading_signals_crypto_status ON trading_signals(cryptocurrency_id, status);
CREATE INDEX idx_trading_signals_created_at ON trading_signals(created_at);
CREATE INDEX idx_signal_performance_signal_id ON signal_performance(signal_id);
CREATE INDEX idx_notification_logs_type_created ON notification_logs(notification_type, created_at);
CREATE INDEX idx_system_logs_level_created ON system_logs(log_level, created_at);
CREATE INDEX idx_cryptocurrencies_symbol ON cryptocurrencies(symbol);
CREATE INDEX idx_cryptocurrencies_cmc_id ON cryptocurrencies(cmc_id) WHERE cmc_id IS NOT NULL;
CREATE INDEX idx_bot_settings_key ON bot_settings(setting_key);

-- =====================================================
-- 4. INSERT INITIAL DATA
-- =====================================================

-- Insert cryptocurrencies
INSERT INTO cryptocurrencies (symbol, name, cmc_id, slug, is_active) VALUES
('BTC', 'Bitcoin', 1, 'bitcoin', true),
('ETH', 'Ethereum', 1027, 'ethereum', true),
('BNB', 'BNB', 1839, 'bnb', true),
('SOL', 'Solana', 5426, 'solana', true),
('ADA', 'Cardano', 2010, 'cardano', true),
('DOT', 'Polkadot', 6636, 'polkadot', true),
('MATIC', 'Polygon', 3890, 'polygon', true),
('AVAX', 'Avalanche', 5805, 'avalanche', true),
('LINK', 'Chainlink', 1975, 'chainlink', true),
('ATOM', 'Cosmos', 3794, 'cosmos', true);

-- Insert bot settings
INSERT INTO bot_settings (setting_key, setting_value, description, data_type) VALUES
('min_confidence_threshold', '0.70', 'Minimum confidence score to generate signal', 'float'),
('max_signals_per_day', '10', 'Maximum number of signals per day', 'integer'),
('analysis_interval_minutes', '15', 'How often to analyze markets (minutes)', 'integer'),
('rsi_oversold_threshold', '30', 'RSI oversold threshold', 'float'),
('rsi_overbought_threshold', '70', 'RSI overbought threshold', 'float'),
('stop_loss_percentage', '5.0', 'Default stop loss percentage', 'float'),
('take_profit_1_percentage', '3.0', 'First take profit target percentage', 'float'),
('take_profit_2_percentage', '6.0', 'Second take profit target percentage', 'float'),
('fear_greed_min_threshold', '20', 'Minimum fear & greed index for signals', 'integer'),
('fear_greed_max_threshold', '80', 'Maximum fear & greed index for signals', 'integer'),
('learning_enabled', 'true', 'Enable machine learning features', 'boolean'),
('backtest_enabled', 'true', 'Enable backtesting of strategies', 'boolean');

COMMIT;

-- =====================================================
-- 5. VERIFICATION QUERIES
-- =====================================================

-- Check table structure
SELECT 
    table_name,
    COUNT(*) as column_count
FROM information_schema.columns 
WHERE table_schema = 'public' 
AND table_name IN ('cryptocurrencies', 'market_snapshots', 'trading_signals', 'signal_performance', 'learning_data', 'notification_logs', 'system_logs', 'bot_settings')
GROUP BY table_name
ORDER BY table_name;

-- Check cryptocurrencies data
SELECT symbol, name, cmc_id, slug, is_active FROM cryptocurrencies ORDER BY symbol;

-- Check bot settings
SELECT setting_key, setting_value, data_type FROM bot_settings ORDER BY setting_key;

-- Show table sizes
SELECT 
    schemaname,
    tablename,
    attname,
    n_distinct,
    correlation
FROM pg_stats 
WHERE schemaname = 'public' 
AND tablename IN ('cryptocurrencies', 'market_snapshots', 'trading_signals', 'signal_performance', 'learning_data', 'notification_logs', 'system_logs', 'bot_settings')
ORDER BY tablename, attname;
