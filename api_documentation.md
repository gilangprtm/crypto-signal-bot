# CryptoSignal AI - API Documentation

## Base URL
```
Production: https://api.cryptosignal-ai.com/v1
Development: http://localhost:8080/v1
```

## Authentication
All API requests require authentication using Bearer token:
```
Authorization: Bearer <your-api-token>
```

## AI Engine Endpoints

### 1. Generate Trading Signal
**POST** `/signals/generate`

Generate a trading signal for a specific cryptocurrency.

**Request Body:**
```json
{
  "symbol": "BTC",
  "timeframe": "1h",
  "analysis_type": "full"
}
```

**Response:**
```json
{
  "signal_id": "uuid",
  "symbol": "BTC",
  "action": "BUY",
  "confidence": 0.85,
  "entry_price": 45000.00,
  "stop_loss": 43000.00,
  "take_profit": [46000.00, 47000.00, 48000.00],
  "reasoning": "Strong bullish momentum with RSI oversold recovery",
  "technical_indicators": {
    "rsi": 35.2,
    "macd": {
      "macd": 120.5,
      "signal": 115.2,
      "histogram": 5.3
    },
    "bollinger_bands": {
      "upper": 46500.00,
      "middle": 45000.00,
      "lower": 43500.00
    }
  },
  "timestamp": "2025-01-03T10:30:00Z"
}
```

### 2. Get Market Analysis
**GET** `/analysis/{symbol}`

Get comprehensive market analysis for a cryptocurrency.

**Parameters:**
- `symbol` (required): Cryptocurrency symbol (e.g., BTC, ETH)
- `timeframe` (optional): Analysis timeframe (1m, 5m, 15m, 1h, 4h, 1d)

**Response:**
```json
{
  "symbol": "BTC",
  "current_price": 45000.00,
  "price_change_24h": 2.5,
  "volume_24h": 25000000000,
  "market_cap": 850000000000,
  "technical_analysis": {
    "trend": "bullish",
    "support_levels": [43000, 42000, 41000],
    "resistance_levels": [46000, 47000, 48000],
    "indicators": {
      "rsi": 65.2,
      "macd": "bullish_crossover",
      "moving_averages": {
        "sma_20": 44500.00,
        "sma_50": 43800.00,
        "ema_12": 44800.00,
        "ema_26": 44200.00
      }
    }
  },
  "sentiment_analysis": {
    "overall_sentiment": "positive",
    "news_sentiment": 0.7,
    "social_sentiment": 0.6,
    "fear_greed_index": 65
  },
  "timestamp": "2025-01-03T10:30:00Z"
}
```

### 3. Get Signal History
**GET** `/signals/history`

Retrieve historical trading signals.

**Query Parameters:**
- `symbol` (optional): Filter by cryptocurrency symbol
- `action` (optional): Filter by action (BUY, SELL, HOLD)
- `from_date` (optional): Start date (ISO 8601)
- `to_date` (optional): End date (ISO 8601)
- `limit` (optional): Number of results (default: 50, max: 200)
- `offset` (optional): Pagination offset

**Response:**
```json
{
  "signals": [
    {
      "signal_id": "uuid",
      "symbol": "BTC",
      "action": "BUY",
      "confidence": 0.85,
      "entry_price": 45000.00,
      "stop_loss": 43000.00,
      "take_profit": [46000.00, 47000.00],
      "created_at": "2025-01-03T10:30:00Z",
      "status": "active",
      "performance": {
        "current_pnl": 2.5,
        "max_pnl": 3.2,
        "min_pnl": -0.8
      }
    }
  ],
  "pagination": {
    "total": 150,
    "limit": 50,
    "offset": 0,
    "has_next": true
  }
}
```

## n8n Integration Endpoints

### 4. Trigger Signal Distribution
**POST** `/n8n/trigger/signal-distribution`

Trigger n8n workflow for signal distribution.

**Request Body:**
```json
{
  "signal_id": "uuid",
  "distribution_channels": ["telegram", "discord", "email"],
  "user_filters": {
    "subscription_tier": ["premium", "vip"],
    "preferences": {
      "symbols": ["BTC", "ETH"],
      "min_confidence": 0.8
    }
  }
}
```

### 5. User Management
**POST** `/users/register`

Register a new user.

**Request Body:**
```json
{
  "telegram_id": "123456789",
  "username": "crypto_trader",
  "email": "user@example.com",
  "subscription_tier": "free",
  "preferences": {
    "symbols": ["BTC", "ETH", "ADA"],
    "min_confidence": 0.7,
    "notifications": {
      "telegram": true,
      "email": false
    }
  }
}
```

**GET** `/users/{user_id}`

Get user information.

**PUT** `/users/{user_id}`

Update user preferences.

### 6. Performance Metrics
**GET** `/performance/signals`

Get signal performance metrics.

**Query Parameters:**
- `period` (optional): Time period (7d, 30d, 90d, 1y)
- `symbol` (optional): Filter by symbol

**Response:**
```json
{
  "period": "30d",
  "total_signals": 45,
  "profitable_signals": 32,
  "win_rate": 71.1,
  "average_return": 3.2,
  "best_signal": {
    "signal_id": "uuid",
    "symbol": "ETH",
    "return": 15.8
  },
  "worst_signal": {
    "signal_id": "uuid",
    "symbol": "ADA",
    "return": -5.2
  },
  "by_symbol": {
    "BTC": {
      "signals": 15,
      "win_rate": 73.3,
      "avg_return": 2.8
    },
    "ETH": {
      "signals": 12,
      "win_rate": 75.0,
      "avg_return": 4.1
    }
  }
}
```

## Error Responses

All endpoints return standardized error responses:

```json
{
  "error": {
    "code": "INVALID_SYMBOL",
    "message": "The provided cryptocurrency symbol is not supported",
    "details": {
      "symbol": "INVALID",
      "supported_symbols": ["BTC", "ETH", "ADA", "SOL"]
    }
  },
  "timestamp": "2025-01-03T10:30:00Z"
}
```

## Rate Limiting

- **Free Tier**: 100 requests/hour
- **Premium Tier**: 1000 requests/hour  
- **VIP Tier**: 5000 requests/hour

Rate limit headers are included in all responses:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1641196800
```

## Webhooks

### Signal Generated Webhook
When a new signal is generated, a webhook can be sent to your endpoint:

**POST** `{your_webhook_url}`

```json
{
  "event": "signal.generated",
  "signal": {
    "signal_id": "uuid",
    "symbol": "BTC",
    "action": "BUY",
    "confidence": 0.85,
    "entry_price": 45000.00,
    "timestamp": "2025-01-03T10:30:00Z"
  }
}
```
