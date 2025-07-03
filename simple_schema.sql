-- Simplified CryptoSignal Database Schema (SQLite)

-- Cryptocurrencies table
CREATE TABLE cryptocurrencies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    is_active BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Trading signals table
CREATE TABLE trading_signals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    crypto_id INTEGER NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('BUY', 'SELL', 'HOLD')),
    confidence REAL NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    entry_price REAL NOT NULL,
    stop_loss REAL,
    take_profit_1 REAL,
    take_profit_2 REAL,
    reasoning TEXT,
    rsi REAL,
    macd_signal TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    status TEXT DEFAULT 'active' CHECK (status IN ('active', 'expired', 'triggered')),
    FOREIGN KEY (crypto_id) REFERENCES cryptocurrencies (id)
);

-- Telegram groups table
CREATE TABLE telegram_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chat_id TEXT UNIQUE NOT NULL,
    group_name TEXT NOT NULL,
    group_type TEXT DEFAULT 'free' CHECK (group_type IN ('personal', 'free', 'premium')),
    is_active BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Signal distributions table (simple tracking)
CREATE TABLE signal_distributions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    signal_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    success BOOLEAN DEFAULT 1,
    FOREIGN KEY (signal_id) REFERENCES trading_signals (id),
    FOREIGN KEY (group_id) REFERENCES telegram_groups (id)
);

-- Performance tracking (simplified)
CREATE TABLE signal_performance (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    signal_id INTEGER NOT NULL,
    entry_price REAL NOT NULL,
    exit_price REAL,
    pnl_percentage REAL,
    outcome TEXT CHECK (outcome IN ('profit', 'loss', 'pending')),
    closed_at DATETIME,
    FOREIGN KEY (signal_id) REFERENCES trading_signals (id)
);

-- System settings
CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX idx_signals_crypto_id ON trading_signals(crypto_id);
CREATE INDEX idx_signals_created_at ON trading_signals(created_at);
CREATE INDEX idx_distributions_signal_id ON signal_distributions(signal_id);
CREATE INDEX idx_performance_signal_id ON signal_performance(signal_id);

-- Insert default cryptocurrencies
INSERT INTO cryptocurrencies (symbol, name) VALUES
('BTC', 'Bitcoin'),
('ETH', 'Ethereum'),
('BNB', 'Binance Coin'),
('ADA', 'Cardano'),
('SOL', 'Solana'),
('DOT', 'Polkadot'),
('MATIC', 'Polygon'),
('AVAX', 'Avalanche');

-- Insert default settings
INSERT INTO settings (key, value) VALUES
('min_confidence', '0.7'),
('max_signals_per_day', '10'),
('analysis_interval_minutes', '15'),
('telegram_bot_token', ''),
('personal_chat_id', ''),
('system_status', 'active');

-- Insert personal telegram group
INSERT INTO telegram_groups (chat_id, group_name, group_type) VALUES
('YOUR_PERSONAL_CHAT_ID', 'Personal Signals', 'personal');
