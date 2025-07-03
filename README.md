# 🤖 Personal Crypto Signal Bot with Interactive Telegram Menu

A sophisticated cryptocurrency trading signal bot built with Go, featuring **Interactive Telegram Menu System**, AI-powered learning capabilities and comprehensive technical analysis.

## 🎯 **NEW: Interactive Telegram Menu System**

✅ **Menu Commands:**

- `/start` - Welcome message dengan tombol interaktif
- `/menu` - Menu utama dengan navigation buttons
- `/status` - Real-time bot status
- `/coins` - Daftar cryptocurrency yang dipantau
- `/performance` - Laporan performa trading
- `/help` - Bantuan lengkap

✅ **Interactive Features:**

- 📊 Real-time bot monitoring
- 🔍 Manual analysis trigger
- 💰 Dynamic coin management (add/remove)
- 📈 Performance tracking
- 🧠 AI learning statistics
- ⚙️ Settings configuration

## ✨ Features

### 📊 **Technical Analysis**

- **RSI (Relative Strength Index)** - Overbought/oversold detection
- **MACD (Moving Average Convergence Divergence)** - Trend momentum analysis
- **Bollinger Bands** - Price volatility and support/resistance levels
- **Moving Averages (SMA/EMA)** - Trend direction analysis
- **Stochastic & Williams %R** - Additional momentum indicators

### 🎯 **Signal Generation**

- **Multi-factor Analysis** - Combines all technical indicators
- **Confidence Scoring** - Weighted signal strength calculation (0-1)
- **Risk Management** - Automatic stop loss & take profit calculation
- **Market Sentiment** - Fear & Greed Index integration

### 🧠 **AI Learning Engine**

- **Pattern Recognition** - Identifies profitable signal patterns
- **Performance Analytics** - Tracks win rate and PnL
- **Strategy Optimization** - Continuous improvement algorithms
- **Feature Extraction** - Converts market data to ML features

### 📱 **Notifications**

- **Telegram Integration** - Rich formatted signal messages
- **WhatsApp Support** - Business API integration ready
- **Real-time Alerts** - Instant signal notifications
- **Daily Summaries** - Performance reports

### 🌐 **Monitoring & Control**

- **REST API** - Complete monitoring interface
- **Real-time Dashboard** - Bot status and analytics
- **Manual Controls** - Start/stop, manual analysis
- **Health Monitoring** - System status checks

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- CoinMarketCap API Key (free tier: 10,000 calls/month)
- Supabase account (free tier available)
- Telegram Bot Token

### 1. Clone & Setup

```bash
git clone <repository-url>
cd crypto-signal-bot
go mod tidy
```

### 2. Database Setup

1. Create a Supabase project at [supabase.com](https://supabase.com)
2. Run the SQL schema from `supabase_schema.sql`
3. Get your Supabase URL and keys

### 3. Configuration

```bash
cp .env.example .env
# Edit .env with your configuration
```

Required environment variables:

```env
# CoinMarketCap API (Primary data source)
COINMARKETCAP_API_KEY=your-cmc-api-key

# Supabase Database
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_KEY=your-service-key

# Telegram Notifications
TELEGRAM_BOT_TOKEN=your-telegram-bot-token
TELEGRAM_CHAT_ID=your-chat-id
```

### 4. Run the Bot

```bash
go run main.go
```

## 📋 Configuration Options

### Bot Settings

- `MIN_CONFIDENCE_THRESHOLD` - Minimum signal confidence (0.0-1.0)
- `MAX_SIGNALS_PER_DAY` - Maximum signals per day
- `ANALYSIS_INTERVAL_MINUTES` - Analysis frequency
- `STOP_LOSS_PERCENTAGE` - Default stop loss %
- `TAKE_PROFIT_1_PERCENTAGE` - First take profit %
- `TAKE_PROFIT_2_PERCENTAGE` - Second take profit %

### Technical Analysis

- `RSI_OVERSOLD_THRESHOLD` - RSI oversold level (default: 30)
- `RSI_OVERBOUGHT_THRESHOLD` - RSI overbought level (default: 70)
- `FEAR_GREED_MIN_THRESHOLD` - Fear threshold (default: 20)
- `FEAR_GREED_MAX_THRESHOLD` - Greed threshold (default: 80)

## 🔧 API Endpoints

### Health & Status

- `GET /api/v1/health` - System health check
- `GET /api/v1/bot/status` - Bot status and metrics

### Control

- `POST /api/v1/bot/start` - Start the bot
- `POST /api/v1/bot/stop` - Stop the bot
- `POST /api/v1/bot/analyze` - Run manual analysis

### Analytics

- `GET /api/v1/signals` - Recent trading signals
- `GET /api/v1/signals/analytics` - Signal performance analytics
- `GET /api/v1/performance/metrics` - Performance metrics
- `GET /api/v1/performance/learning` - Learning insights

### Scheduler

- `GET /api/v1/scheduler/status` - Scheduler status
- `POST /api/v1/scheduler/jobs/{job}/run` - Run specific job

## 📊 Database Schema

The bot uses Supabase (PostgreSQL) with the following main tables:

- `cryptocurrencies` - Supported cryptocurrencies
- `trading_signals` - Generated trading signals
- `signal_performance` - Signal outcome tracking
- `market_snapshots` - Historical market data
- `learning_data` - AI learning dataset
- `notification_logs` - Notification history

## 🤖 How It Works

1. **Data Collection** - Fetches real-time data from CoinMarketCap (primary) with Binance fallback
2. **Technical Analysis** - Calculates multiple technical indicators
3. **Signal Generation** - Analyzes market conditions and generates signals
4. **Risk Management** - Calculates stop loss and take profit levels
5. **Notification** - Sends formatted signals to Telegram/WhatsApp
6. **Learning** - Tracks outcomes and improves strategy over time

### 📡 **Data Sources**

- **Primary**: CoinMarketCap API (free tier: 10,000 calls/month)
- **Fallback**: Binance Public API (for kline data)
- **Sentiment**: Fear & Greed Index
- **Technical**: Real-time OHLCV data for indicators

## 📈 Signal Format

```
🚀 BUY SIGNAL - BTC/USDT
💰 Entry: $45,250.00
🛑 Stop Loss: $42,987.50 (-5.0%)
🎯 Take Profit 1: $46,607.50 (+3.0%)
🎯 Take Profit 2: $47,965.00 (+6.0%)
📊 Confidence: 85%

📈 Technical Analysis:
• RSI: 25.4 (Oversold)
• MACD: Bullish crossover
• Bollinger: Price near lower band
• Fear & Greed: 15 (Extreme Fear)

⚡ Market Conditions: Strong buy signals across multiple indicators
```

## 🚀 Production Deployment

### Quick Deploy Options:

#### 🚂 Railway (Recommended)

```bash
# 1. Push to GitHub
git add .
git commit -m "Deploy crypto signal bot"
git push origin main

# 2. Deploy to Railway
# - Go to railway.app
# - Connect GitHub repo
# - Add environment variables
# - Auto-deploy with Dockerfile
```

#### 🎨 Render Alternative

```bash
# 1. Connect to render.com
# 2. Use render.yaml configuration
# 3. Set environment variables
# 4. Deploy automatically
```

### Environment Variables for Production:

```env
COINMARKETCAP_API_KEY=your_cmc_api_key
SUPABASE_URL=your_supabase_url
SUPABASE_SERVICE_KEY=your_service_key
TELEGRAM_BOT_TOKEN=your_bot_token
TELEGRAM_CHAT_ID=your_chat_id
PORT=8080
LOG_LEVEL=info
```

### Post-Deployment Verification:

- ✅ Health check: `https://your-app.domain/health`
- ✅ Telegram test: Send `/start` to your bot
- ✅ Interactive menu: Send `/menu` and test buttons
- ✅ Monitor logs for any errors

📖 **Detailed deployment guide**: See [DEPLOYMENT.md](DEPLOYMENT.md)

## 🔒 Security Notes

- Store sensitive keys in environment variables
- Use Supabase Row Level Security (RLS)
- Never commit `.env` files to version control
- Consider using a VPS for 24/7 operation

## 📝 License

This project is for personal use only. Please ensure compliance with your local regulations regarding automated trading.

## 🤝 Contributing

This is a personal project, but suggestions and improvements are welcome!

## ⚠️ Disclaimer

This bot is for educational and personal use only. Cryptocurrency trading involves significant risk. Always do your own research and never invest more than you can afford to lose.
