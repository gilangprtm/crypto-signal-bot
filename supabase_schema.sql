-- Personal Crypto Signal Bot - Supabase Schema
-- Optimized for learning and analytics

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Cryptocurrencies table
CREATE TABLE cryptocurrencies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    symbol VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    coingecko_id VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Trading signals table (enhanced for learning)
CREATE TABLE trading_signals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    crypto_id UUID NOT NULL REFERENCES cryptocurrencies(id),
    action VARCHAR(10) NOT NULL CHECK (action IN ('BUY', 'SELL', 'HOLD')),
    confidence_score DECIMAL(5,4) NOT NULL CHECK (confidence_score >= 0 AND confidence_score <= 1),
    entry_price DECIMAL(20,8) NOT NULL,
    stop_loss DECIMAL(20,8),
    take_profit_1 DECIMAL(20,8),
    take_profit_2 DECIMAL(20,8),
    reasoning TEXT,
    
    -- Technical indicators at signal time
    rsi DECIMAL(5,2),
    macd_line DECIMAL(10,6),
    macd_signal DECIMAL(10,6),
    macd_histogram DECIMAL(10,6),
    bb_upper DECIMAL(20,8),
    bb_middle DECIMAL(20,8),
    bb_lower DECIMAL(20,8),
    sma_20 DECIMAL(20,8),
    ema_12 DECIMAL(20,8),
    ema_26 DECIMAL(20,8),
    volume_24h DECIMAL(30,8),
    price_change_24h DECIMAL(10,4),
    
    -- Market sentiment at signal time
    fear_greed_index INTEGER,
    market_cap DECIMAL(30,8),
    
    -- Additional context for learning
    market_conditions JSONB DEFAULT '{}',
    timeframe VARCHAR(10) DEFAULT '15m',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'expired', 'triggered', 'cancelled'))
);

-- Signal performance tracking (enhanced for learning)
CREATE TABLE signal_performance (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    signal_id UUID NOT NULL REFERENCES trading_signals(id),
    entry_price DECIMAL(20,8) NOT NULL,
    exit_price DECIMAL(20,8),
    highest_price DECIMAL(20,8), -- Track max profit potential
    lowest_price DECIMAL(20,8),  -- Track max loss potential
    pnl_percentage DECIMAL(10,4),
    entry_time TIMESTAMP WITH TIME ZONE NOT NULL,
    exit_time TIMESTAMP WITH TIME ZONE,
    outcome VARCHAR(20) CHECK (outcome IN ('profit', 'loss', 'breakeven', 'pending')),
    duration_minutes INTEGER,
    
    -- Learning metrics
    hit_stop_loss BOOLEAN DEFAULT false,
    hit_take_profit_1 BOOLEAN DEFAULT false,
    hit_take_profit_2 BOOLEAN DEFAULT false,
    max_profit_percentage DECIMAL(10,4),
    max_loss_percentage DECIMAL(10,4),
    
    -- Exit reason for learning
    exit_reason VARCHAR(50) -- 'manual', 'stop_loss', 'take_profit', 'timeout', 'market_change'
);

-- Market data snapshots (for learning patterns)
CREATE TABLE market_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    crypto_id UUID NOT NULL REFERENCES cryptocurrencies(id),
    price DECIMAL(20,8) NOT NULL,
    volume_24h DECIMAL(30,8),
    market_cap DECIMAL(30,8),
    price_change_1h DECIMAL(10,4),
    price_change_24h DECIMAL(10,4),
    price_change_7d DECIMAL(10,4),
    
    -- Technical indicators
    rsi DECIMAL(5,2),
    macd_line DECIMAL(10,6),
    macd_signal DECIMAL(10,6),
    macd_histogram DECIMAL(10,6),
    bb_upper DECIMAL(20,8),
    bb_middle DECIMAL(20,8),
    bb_lower DECIMAL(20,8),
    sma_20 DECIMAL(20,8),
    ema_12 DECIMAL(20,8),
    ema_26 DECIMAL(20,8),
    
    -- Market sentiment
    fear_greed_index INTEGER,
    
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Learning data for AI improvement
CREATE TABLE learning_data (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    signal_id UUID REFERENCES trading_signals(id),
    
    -- Feature vector at signal time
    features JSONB NOT NULL, -- All technical indicators, market conditions
    
    -- Actual outcomes
    actual_outcome VARCHAR(20), -- 'profit', 'loss', 'breakeven'
    actual_pnl_percentage DECIMAL(10,4),
    actual_duration_minutes INTEGER,
    
    -- Predicted outcomes (for model evaluation)
    predicted_outcome VARCHAR(20),
    predicted_confidence DECIMAL(5,4),
    
    -- Model performance metrics
    prediction_accuracy DECIMAL(5,4),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Personal notifications log
CREATE TABLE notification_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    signal_id UUID REFERENCES trading_signals(id),
    channel_type VARCHAR(20) NOT NULL, -- 'telegram', 'whatsapp'
    channel_id VARCHAR(255) NOT NULL, -- Your personal chat ID
    message_text TEXT NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    delivery_status VARCHAR(20) DEFAULT 'sent' CHECK (delivery_status IN ('sent', 'failed', 'delivered'))
);

-- System logs for debugging
CREATE TABLE system_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    level VARCHAR(10) NOT NULL, -- 'info', 'warning', 'error', 'debug'
    component VARCHAR(50) NOT NULL, -- 'data_collector', 'analyzer', 'signal_generator', 'notifier'
    message TEXT NOT NULL,
    context JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Bot settings
CREATE TABLE bot_settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_signals_crypto_id ON trading_signals(crypto_id);
CREATE INDEX idx_signals_created_at ON trading_signals(created_at);
CREATE INDEX idx_signals_confidence ON trading_signals(confidence_score);
CREATE INDEX idx_signals_action ON trading_signals(action);
CREATE INDEX idx_performance_signal_id ON signal_performance(signal_id);
CREATE INDEX idx_performance_outcome ON signal_performance(outcome);
CREATE INDEX idx_performance_entry_time ON signal_performance(entry_time);
CREATE INDEX idx_market_crypto_timestamp ON market_snapshots(crypto_id, timestamp);
CREATE INDEX idx_learning_signal_id ON learning_data(signal_id);
CREATE INDEX idx_learning_created_at ON learning_data(created_at);
CREATE INDEX idx_notifications_signal_id ON notification_logs(signal_id);
CREATE INDEX idx_logs_level_created ON system_logs(level, created_at);
CREATE INDEX idx_logs_component ON system_logs(component);

-- Insert default cryptocurrencies
INSERT INTO cryptocurrencies (symbol, name, coingecko_id) VALUES
('BTC', 'Bitcoin', 'bitcoin'),
('ETH', 'Ethereum', 'ethereum'),
('BNB', 'Binance Coin', 'binancecoin'),
('ADA', 'Cardano', 'cardano'),
('SOL', 'Solana', 'solana'),
('DOT', 'Polkadot', 'polkadot'),
('MATIC', 'Polygon', 'matic-network'),
('AVAX', 'Avalanche', 'avalanche-2'),
('LINK', 'Chainlink', 'chainlink'),
('ATOM', 'Cosmos', 'cosmos');

-- Insert default bot settings
INSERT INTO bot_settings (key, value, description) VALUES
('min_confidence_threshold', '0.70', 'Minimum confidence score to generate signal'),
('max_signals_per_day', '10', 'Maximum number of signals per day'),
('analysis_interval_minutes', '15', 'How often to analyze markets (minutes)'),
('telegram_bot_token', '', 'Telegram bot token for notifications'),
('telegram_chat_id', '', 'Your personal Telegram chat ID'),
('whatsapp_enabled', 'false', 'Enable WhatsApp notifications'),
('stop_loss_percentage', '5.0', 'Default stop loss percentage'),
('take_profit_1_percentage', '3.0', 'First take profit target percentage'),
('take_profit_2_percentage', '6.0', 'Second take profit target percentage'),
('rsi_oversold_threshold', '30', 'RSI oversold threshold'),
('rsi_overbought_threshold', '70', 'RSI overbought threshold'),
('fear_greed_min_threshold', '20', 'Minimum fear & greed index for signals'),
('fear_greed_max_threshold', '80', 'Maximum fear & greed index for signals'),
('learning_enabled', 'true', 'Enable machine learning features'),
('backtest_enabled', 'true', 'Enable backtesting of strategies');

-- Create views for analytics
CREATE VIEW signal_analytics AS
SELECT 
    c.symbol,
    COUNT(*) as total_signals,
    COUNT(CASE WHEN sp.outcome = 'profit' THEN 1 END) as profitable_signals,
    COUNT(CASE WHEN sp.outcome = 'loss' THEN 1 END) as loss_signals,
    ROUND(
        COUNT(CASE WHEN sp.outcome = 'profit' THEN 1 END) * 100.0 / 
        NULLIF(COUNT(CASE WHEN sp.outcome IN ('profit', 'loss') THEN 1 END), 0), 
        2
    ) as win_rate_percentage,
    ROUND(AVG(sp.pnl_percentage), 2) as avg_pnl_percentage,
    ROUND(MAX(sp.pnl_percentage), 2) as best_signal_pnl,
    ROUND(MIN(sp.pnl_percentage), 2) as worst_signal_pnl,
    ROUND(AVG(ts.confidence_score), 4) as avg_confidence
FROM trading_signals ts
JOIN cryptocurrencies c ON ts.crypto_id = c.id
LEFT JOIN signal_performance sp ON ts.id = sp.signal_id
WHERE ts.created_at >= NOW() - INTERVAL '30 days'
GROUP BY c.symbol
ORDER BY win_rate_percentage DESC;

-- Create view for learning insights
CREATE VIEW learning_insights AS
SELECT 
    DATE_TRUNC('day', created_at) as date,
    COUNT(*) as signals_generated,
    AVG(prediction_accuracy) as avg_accuracy,
    COUNT(CASE WHEN actual_outcome = 'profit' THEN 1 END) as actual_profits,
    COUNT(CASE WHEN predicted_outcome = 'profit' THEN 1 END) as predicted_profits
FROM learning_data
WHERE created_at >= NOW() - INTERVAL '90 days'
GROUP BY DATE_TRUNC('day', created_at)
ORDER BY date DESC;
