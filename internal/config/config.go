package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Supabase
	SupabaseURL        string
	SupabaseAnonKey    string
	SupabaseServiceKey string

	// Telegram
	TelegramBotToken string
	TelegramChatID   string

	// WhatsApp
	WhatsAppEnabled bool
	WhatsAppAPIURL  string
	WhatsAppToken   string

	// API Keys
	CoinMarketCapAPIKey string
	CoinGeckoAPIKey     string
	BinanceAPIKey       string
	BinanceSecret       string

	// Bot Settings
	MinConfidenceThreshold   float64
	MaxSignalsPerDay         int
	AnalysisIntervalMinutes  int
	AnalysisIntervalSeconds  int
	StopLossPercentage       float64
	TakeProfit1Percentage    float64
	TakeProfit2Percentage    float64

	// Technical Analysis
	RSIOversoldThreshold    float64
	RSIOverboughtThreshold  float64
	FearGreedMinThreshold   int
	FearGreedMaxThreshold   int

	// Learning
	LearningEnabled  bool
	BacktestEnabled  bool

	// Server
	Port     string
	APIPort  int
	LogLevel string
	Environment string
}

func Load() *Config {
	return &Config{
		// Supabase
		SupabaseURL:        getEnv("SUPABASE_URL", ""),
		SupabaseAnonKey:    getEnv("SUPABASE_ANON_KEY", ""),
		SupabaseServiceKey: getEnv("SUPABASE_SERVICE_KEY", ""),

		// Telegram
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramChatID:   getEnv("TELEGRAM_CHAT_ID", ""),

		// WhatsApp
		WhatsAppEnabled: getEnvBool("WHATSAPP_ENABLED", false),
		WhatsAppAPIURL:  getEnv("WHATSAPP_API_URL", ""),
		WhatsAppToken:   getEnv("WHATSAPP_API_TOKEN", ""),

		// API Keys
		CoinMarketCapAPIKey: getEnv("COINMARKETCAP_API_KEY", ""),
		CoinGeckoAPIKey:     getEnv("COINGECKO_API_KEY", ""),
		BinanceAPIKey:       getEnv("BINANCE_API_KEY", ""),
		BinanceSecret:       getEnv("BINANCE_SECRET_KEY", ""),

		// Bot Settings
		MinConfidenceThreshold:  getEnvFloat("MIN_CONFIDENCE_THRESHOLD", 0.70),
		MaxSignalsPerDay:        getEnvInt("MAX_SIGNALS_PER_DAY", 10),
		AnalysisIntervalMinutes: getEnvInt("ANALYSIS_INTERVAL_MINUTES", 15),
		AnalysisIntervalSeconds: getEnvInt("ANALYSIS_INTERVAL_SECONDS", 900), // 15 minutes
		StopLossPercentage:      getEnvFloat("STOP_LOSS_PERCENTAGE", 5.0),
		TakeProfit1Percentage:   getEnvFloat("TAKE_PROFIT_1_PERCENTAGE", 3.0),
		TakeProfit2Percentage:   getEnvFloat("TAKE_PROFIT_2_PERCENTAGE", 6.0),

		// Technical Analysis
		RSIOversoldThreshold:   getEnvFloat("RSI_OVERSOLD_THRESHOLD", 30),
		RSIOverboughtThreshold: getEnvFloat("RSI_OVERBOUGHT_THRESHOLD", 70),
		FearGreedMinThreshold:  getEnvInt("FEAR_GREED_MIN_THRESHOLD", 20),
		FearGreedMaxThreshold:  getEnvInt("FEAR_GREED_MAX_THRESHOLD", 80),

		// Learning
		LearningEnabled: getEnvBool("LEARNING_ENABLED", true),
		BacktestEnabled: getEnvBool("BACKTEST_ENABLED", true),

		// Server
		Port:        getEnv("PORT", "8080"),
		APIPort:     getEnvInt("API_PORT", 8080),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}

func (c *Config) Validate() error {
	// Add validation logic here
	return nil
}
