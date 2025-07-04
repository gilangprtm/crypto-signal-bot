package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Cryptocurrency represents a cryptocurrency
type Cryptocurrency struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	Symbol          string     `json:"symbol" db:"symbol"`
	Name            string     `json:"name" db:"name"`
	CmcID           *int       `json:"cmc_id" db:"cmc_id"`                     // CoinMarketCap ID
	ContractAddress *string    `json:"contract_address" db:"contract_address"` // Smart contract address
	Platform        *string    `json:"platform" db:"platform"`                // Blockchain platform (e.g., "ethereum", "binance-smart-chain")
	Slug            *string    `json:"slug" db:"slug"`                         // CoinMarketCap slug
	CoingeckoID     *string    `json:"coingecko_id" db:"coingecko_id"`        // Keep for backward compatibility
	IsActive        bool       `json:"is_active" db:"is_active"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at" db:"updated_at"`
}

// TradingSignal represents a trading signal
type TradingSignal struct {
	ID               uuid.UUID              `json:"id" db:"id"`
	CryptoID         uuid.UUID              `json:"crypto_id" db:"crypto_id"`
	Action           string                 `json:"action" db:"action"` // BUY, SELL, HOLD
	ConfidenceScore  decimal.Decimal        `json:"confidence_score" db:"confidence_score"`
	EntryPrice       decimal.Decimal        `json:"entry_price" db:"entry_price"`
	StopLoss         *decimal.Decimal       `json:"stop_loss" db:"stop_loss"`
	TakeProfit1      *decimal.Decimal       `json:"take_profit_1" db:"take_profit_1"`
	TakeProfit2      *decimal.Decimal       `json:"take_profit_2" db:"take_profit_2"`
	Reasoning        string                 `json:"reasoning" db:"reasoning"`
	
	// Technical indicators at signal time
	RSI              *decimal.Decimal       `json:"rsi" db:"rsi"`
	MACDLine         *decimal.Decimal       `json:"macd_line" db:"macd_line"`
	MACDSignal       *decimal.Decimal       `json:"macd_signal" db:"macd_signal"`
	MACDHistogram    *decimal.Decimal       `json:"macd_histogram" db:"macd_histogram"`
	BBUpper          *decimal.Decimal       `json:"bb_upper" db:"bb_upper"`
	BBMiddle         *decimal.Decimal       `json:"bb_middle" db:"bb_middle"`
	BBLower          *decimal.Decimal       `json:"bb_lower" db:"bb_lower"`
	SMA20            *decimal.Decimal       `json:"sma_20" db:"sma_20"`
	EMA12            *decimal.Decimal       `json:"ema_12" db:"ema_12"`
	EMA26            *decimal.Decimal       `json:"ema_26" db:"ema_26"`
	Volume24h        *decimal.Decimal       `json:"volume_24h" db:"volume_24h"`
	PriceChange24h   *decimal.Decimal       `json:"price_change_24h" db:"price_change_24h"`
	
	// Market sentiment
	FearGreedIndex   *int                   `json:"fear_greed_index" db:"fear_greed_index"`
	MarketCap        *decimal.Decimal       `json:"market_cap" db:"market_cap"`
	
	// Additional context
	MarketConditions map[string]interface{} `json:"market_conditions" db:"market_conditions"`
	Timeframe        string                 `json:"timeframe" db:"timeframe"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	Status           string                 `json:"status" db:"status"` // active, expired, triggered, cancelled
	
	// Related data (not stored in DB)
	Crypto           *Cryptocurrency        `json:"crypto,omitempty"`
}

// SignalPerformance tracks the performance of a trading signal
type SignalPerformance struct {
	ID                   uuid.UUID        `json:"id" db:"id"`
	SignalID             uuid.UUID        `json:"signal_id" db:"signal_id"`
	EntryPrice           decimal.Decimal  `json:"entry_price" db:"entry_price"`
	ExitPrice            *decimal.Decimal `json:"exit_price" db:"exit_price"`
	HighestPrice         *decimal.Decimal `json:"highest_price" db:"highest_price"`
	LowestPrice          *decimal.Decimal `json:"lowest_price" db:"lowest_price"`
	PnLPercentage        *decimal.Decimal `json:"pnl_percentage" db:"pnl_percentage"`
	EntryTime            time.Time        `json:"entry_time" db:"entry_time"`
	ExitTime             *time.Time       `json:"exit_time" db:"exit_time"`
	Outcome              string           `json:"outcome" db:"outcome"` // profit, loss, breakeven, pending
	DurationMinutes      *int             `json:"duration_minutes" db:"duration_minutes"`
	
	// Learning metrics
	HitStopLoss          bool             `json:"hit_stop_loss" db:"hit_stop_loss"`
	HitTakeProfit1       bool             `json:"hit_take_profit_1" db:"hit_take_profit_1"`
	HitTakeProfit2       bool             `json:"hit_take_profit_2" db:"hit_take_profit_2"`
	MaxProfitPercentage  *decimal.Decimal `json:"max_profit_percentage" db:"max_profit_percentage"`
	MaxLossPercentage    *decimal.Decimal `json:"max_loss_percentage" db:"max_loss_percentage"`
	ExitReason           string           `json:"exit_reason" db:"exit_reason"`
	
	// Related data
	Signal               *TradingSignal   `json:"signal,omitempty"`
}

// MarketSnapshot represents market data at a specific time
type MarketSnapshot struct {
	ID               uuid.UUID       `json:"id" db:"id"`
	CryptoID         uuid.UUID       `json:"crypto_id" db:"crypto_id"`
	Price            decimal.Decimal `json:"price" db:"price"`
	Volume24h        decimal.Decimal `json:"volume_24h" db:"volume_24h"`
	MarketCap        decimal.Decimal `json:"market_cap" db:"market_cap"`
	PriceChange1h    decimal.Decimal `json:"price_change_1h" db:"price_change_1h"`
	PriceChange24h   decimal.Decimal `json:"price_change_24h" db:"price_change_24h"`
	PriceChange7d    decimal.Decimal `json:"price_change_7d" db:"price_change_7d"`
	
	// Technical indicators
	RSI              decimal.Decimal `json:"rsi" db:"rsi"`
	MACDLine         decimal.Decimal `json:"macd_line" db:"macd_line"`
	MACDSignal       decimal.Decimal `json:"macd_signal" db:"macd_signal"`
	MACDHistogram    decimal.Decimal `json:"macd_histogram" db:"macd_histogram"`
	BBUpper          decimal.Decimal `json:"bb_upper" db:"bb_upper"`
	BBMiddle         decimal.Decimal `json:"bb_middle" db:"bb_middle"`
	BBLower          decimal.Decimal `json:"bb_lower" db:"bb_lower"`
	SMA20            decimal.Decimal `json:"sma_20" db:"sma_20"`
	EMA12            decimal.Decimal `json:"ema_12" db:"ema_12"`
	EMA26            decimal.Decimal `json:"ema_26" db:"ema_26"`
	
	// Market sentiment
	FearGreedIndex   int             `json:"fear_greed_index" db:"fear_greed_index"`
	
	Timestamp        time.Time       `json:"timestamp" db:"timestamp"`
	
	// Related data
	Crypto           *Cryptocurrency `json:"crypto,omitempty"`
}

// LearningData represents data for machine learning
type LearningData struct {
	ID                      uuid.UUID              `json:"id" db:"id"`
	SignalID                *uuid.UUID             `json:"signal_id" db:"signal_id"`
	Features                map[string]interface{} `json:"features" db:"features"`
	ActualOutcome           string                 `json:"actual_outcome" db:"actual_outcome"`
	ActualPnLPercentage     decimal.Decimal        `json:"actual_pnl_percentage" db:"actual_pnl_percentage"`
	ActualDurationMinutes   int                    `json:"actual_duration_minutes" db:"actual_duration_minutes"`
	PredictedOutcome        string                 `json:"predicted_outcome" db:"predicted_outcome"`
	PredictedConfidence     decimal.Decimal        `json:"predicted_confidence" db:"predicted_confidence"`
	PredictionAccuracy      decimal.Decimal        `json:"prediction_accuracy" db:"prediction_accuracy"`
	CreatedAt               time.Time              `json:"created_at" db:"created_at"`
}

// NotificationLog represents a notification sent to user
type NotificationLog struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	SignalID       *uuid.UUID `json:"signal_id" db:"signal_id"`
	ChannelType    string     `json:"channel_type" db:"channel_type"` // telegram, whatsapp
	ChannelID      string     `json:"channel_id" db:"channel_id"`
	MessageText    string     `json:"message_text" db:"message_text"`
	SentAt         time.Time  `json:"sent_at" db:"sent_at"`
	DeliveryStatus string     `json:"delivery_status" db:"delivery_status"` // sent, failed, delivered
}

// SystemLog represents system logs
type SystemLog struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	Level     string                 `json:"level" db:"level"`
	Component string                 `json:"component" db:"component"`
	Message   string                 `json:"message" db:"message"`
	Context   map[string]interface{} `json:"context" db:"context"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// BotSetting represents bot configuration
type BotSetting struct {
	Key         string    `json:"key" db:"key"`
	Value       string    `json:"value" db:"value"`
	Description string    `json:"description" db:"description"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Analytics models
type SignalAnalytics struct {
	Symbol              string          `json:"symbol" db:"symbol"`
	TotalSignals        int             `json:"total_signals" db:"total_signals"`
	ProfitableSignals   int             `json:"profitable_signals" db:"profitable_signals"`
	LossSignals         int             `json:"loss_signals" db:"loss_signals"`
	WinRatePercentage   decimal.Decimal `json:"win_rate_percentage" db:"win_rate_percentage"`
	AvgPnLPercentage    decimal.Decimal `json:"avg_pnl_percentage" db:"avg_pnl_percentage"`
	BestSignalPnL       decimal.Decimal `json:"best_signal_pnl" db:"best_signal_pnl"`
	WorstSignalPnL      decimal.Decimal `json:"worst_signal_pnl" db:"worst_signal_pnl"`
	AvgConfidence       decimal.Decimal `json:"avg_confidence" db:"avg_confidence"`
}

type LearningInsight struct {
	Date              time.Time `json:"date" db:"date"`
	SignalsGenerated  int       `json:"signals_generated" db:"signals_generated"`
	AvgAccuracy       decimal.Decimal `json:"avg_accuracy" db:"avg_accuracy"`
	ActualProfits     int       `json:"actual_profits" db:"actual_profits"`
	PredictedProfits  int       `json:"predicted_profits" db:"predicted_profits"`
}

// API Response models
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}


