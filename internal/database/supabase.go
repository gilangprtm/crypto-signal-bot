package database

import (
	"crypto-signal-bot/internal/config"
	"crypto-signal-bot/internal/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type SupabaseClient struct {
	db        *sql.DB
	restClient *SupabaseRestClient
	cfg       *config.Config
	useRest   bool
}

func NewSupabaseClient(cfg *config.Config) (*SupabaseClient, error) {
	// Initialize REST client as fallback
	restClient := NewSupabaseRestClient(cfg)

	// Try direct database connection first
	logrus.Info("Attempting direct database connection...")

	// Get database connection details from environment variables
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	// Set defaults if not provided
	if dbHost == "" {
		// Fallback to extracting from Supabase URL
		projectID := extractProjectID(cfg.SupabaseURL)
		if projectID == "" {
			logrus.Warn("Invalid Supabase URL and no DB_HOST provided, using REST API only")
			return &SupabaseClient{
				restClient: restClient,
				cfg:        cfg,
				useRest:    true,
			}, nil
		}
		dbHost = fmt.Sprintf("db.%s.supabase.co", projectID)
	}
	if dbPort == "" {
		dbPort = "5432"
	}
	if dbUser == "" {
		dbUser = "postgres"
	}
	if dbPassword == "" {
		// Fallback to service key
		dbPassword = cfg.SupabaseServiceKey
	}
	if dbName == "" {
		dbName = "postgres"
	}
	if dbSSLMode == "" {
		dbSSLMode = "require"
	}

	logrus.Debugf("Connecting to database: host=%s port=%s user=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbName, dbSSLMode)

	// Build connection string for PostgreSQL
	connStr := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s password=%s connect_timeout=10",
		dbHost, dbPort, dbUser, dbName, dbSSLMode, dbPassword)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logrus.Warnf("Failed to open database connection: %v, falling back to REST API", err)
		return &SupabaseClient{
			restClient: restClient,
			cfg:        cfg,
			useRest:    true,
		}, nil
	}

	// Set connection pool settings for production stability
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		logrus.Warnf("Failed to ping database: %v, falling back to REST API", err)
		db.Close()

		// Test REST API connection
		if err := restClient.TestConnection(); err != nil {
			return nil, fmt.Errorf("both database and REST API connections failed: %w", err)
		}

		return &SupabaseClient{
			restClient: restClient,
			cfg:        cfg,
			useRest:    true,
		}, nil
	}

	logrus.Info("✅ Successfully connected to Supabase database via direct connection")

	return &SupabaseClient{
		db:         db,
		restClient: restClient,
		cfg:        cfg,
		useRest:    false,
	}, nil
}

func (s *SupabaseClient) Close() error {
	if s.useRest {
		return s.restClient.Close()
	}
	return s.db.Close()
}

func (s *SupabaseClient) Ping() error {
	if s.useRest {
		return s.restClient.Ping()
	}
	return s.db.Ping()
}

// Signal operations
func (s *SupabaseClient) CreateSignal(signal *models.TradingSignal) error {
	if s.useRest {
		return s.restClient.CreateSignal(signal)
	}
	query := `
		INSERT INTO trading_signals (
			id, crypto_id, action, confidence_score, entry_price, stop_loss,
			take_profit_1, take_profit_2, reasoning, rsi, macd_line, macd_signal,
			macd_histogram, bb_upper, bb_middle, bb_lower, sma_20, ema_12, ema_26,
			volume_24h, price_change_24h, fear_greed_index, market_cap,
			market_conditions, timeframe, created_at, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27
		)`

	marketConditionsJSON, _ := json.Marshal(signal.MarketConditions)

	_, err := s.db.Exec(query,
		signal.ID, signal.CryptoID, signal.Action, signal.ConfidenceScore,
		signal.EntryPrice, signal.StopLoss, signal.TakeProfit1, signal.TakeProfit2,
		signal.Reasoning, signal.RSI, signal.MACDLine, signal.MACDSignal,
		signal.MACDHistogram, signal.BBUpper, signal.BBMiddle, signal.BBLower,
		signal.SMA20, signal.EMA12, signal.EMA26, signal.Volume24h,
		signal.PriceChange24h, signal.FearGreedIndex, signal.MarketCap,
		marketConditionsJSON, signal.Timeframe, signal.CreatedAt, signal.Status,
	)

	if err != nil {
		logrus.Error("Failed to create signal: ", err)
		return err
	}

	logrus.Info("✅ Signal created successfully: ", signal.ID)
	return nil
}

func (s *SupabaseClient) GetActiveSignals() ([]*models.TradingSignal, error) {
	if s.useRest {
		return s.restClient.GetActiveSignals()
	}
	query := `
		SELECT id, crypto_id, action, confidence_score, entry_price, stop_loss,
			   take_profit_1, take_profit_2, reasoning, created_at, status
		FROM trading_signals 
		WHERE status = 'active' 
		ORDER BY created_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var signals []*models.TradingSignal
	for rows.Next() {
		signal := &models.TradingSignal{}
		err := rows.Scan(
			&signal.ID, &signal.CryptoID, &signal.Action, &signal.ConfidenceScore,
			&signal.EntryPrice, &signal.StopLoss, &signal.TakeProfit1,
			&signal.TakeProfit2, &signal.Reasoning, &signal.CreatedAt, &signal.Status,
		)
		if err != nil {
			logrus.Error("Failed to scan signal: ", err)
			continue
		}
		signals = append(signals, signal)
	}

	return signals, nil
}

func (s *SupabaseClient) UpdateSignalStatus(signalID uuid.UUID, status string) error {
	if s.useRest {
		return s.restClient.UpdateSignalStatus(signalID, status)
	}
	query := `UPDATE trading_signals SET status = $1 WHERE id = $2`
	_, err := s.db.Exec(query, status, signalID)
	return err
}

// Performance tracking
func (s *SupabaseClient) CreatePerformanceRecord(perf *models.SignalPerformance) error {
	query := `
		INSERT INTO signal_performance (
			id, signal_id, entry_price, exit_price, highest_price, lowest_price,
			pnl_percentage, entry_time, exit_time, outcome, duration_minutes,
			hit_stop_loss, hit_take_profit_1, hit_take_profit_2,
			max_profit_percentage, max_loss_percentage, exit_reason
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)`

	_, err := s.db.Exec(query,
		perf.ID, perf.SignalID, perf.EntryPrice, perf.ExitPrice,
		perf.HighestPrice, perf.LowestPrice, perf.PnLPercentage,
		perf.EntryTime, perf.ExitTime, perf.Outcome, perf.DurationMinutes,
		perf.HitStopLoss, perf.HitTakeProfit1, perf.HitTakeProfit2,
		perf.MaxProfitPercentage, perf.MaxLossPercentage, perf.ExitReason,
	)

	return err
}

// Market data
func (s *SupabaseClient) SaveMarketSnapshot(snapshot *models.MarketSnapshot) error {
	if s.useRest {
		return s.restClient.SaveMarketSnapshot(snapshot)
	}
	query := `
		INSERT INTO market_snapshots (
			id, crypto_id, price, volume_24h, market_cap, price_change_1h,
			price_change_24h, price_change_7d, rsi, macd_line, macd_signal,
			macd_histogram, bb_upper, bb_middle, bb_lower, sma_20, ema_12,
			ema_26, fear_greed_index, timestamp
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)`

	_, err := s.db.Exec(query,
		snapshot.ID, snapshot.CryptocurrencyID, snapshot.Price, snapshot.Volume24h,
		snapshot.MarketCap, snapshot.PriceChange1h, snapshot.PriceChange24h,
		snapshot.PriceChange7d, snapshot.RSI, snapshot.MACDLine, snapshot.MACDSignal,
		snapshot.MACDHistogram, snapshot.BBUpper, snapshot.BBMiddle, snapshot.BBLower,
		snapshot.SMA20, snapshot.EMA12, snapshot.EMA26, snapshot.FearGreedIndex,
		snapshot.Timestamp,
	)

	return err
}

// Learning data
func (s *SupabaseClient) SaveLearningData(data *models.LearningData) error {
	query := `
		INSERT INTO learning_data (
			id, signal_id, features, actual_outcome, actual_pnl_percentage,
			actual_duration_minutes, predicted_outcome, predicted_confidence,
			prediction_accuracy, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)`

	featuresJSON, _ := json.Marshal(data.Features)

	_, err := s.db.Exec(query,
		data.ID, data.SignalID, featuresJSON, data.ActualOutcome,
		data.ActualPnLPercentage, data.ActualDurationMinutes,
		data.PredictedOutcome, data.PredictedConfidence,
		data.PredictionAccuracy, data.CreatedAt,
	)

	return err
}

// Analytics
func (s *SupabaseClient) GetSignalAnalytics() ([]*models.SignalAnalytics, error) {
	query := `SELECT * FROM signal_analytics ORDER BY win_rate_percentage DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analytics []*models.SignalAnalytics
	for rows.Next() {
		analytic := &models.SignalAnalytics{}
		err := rows.Scan(
			&analytic.Symbol, &analytic.TotalSignals, &analytic.ProfitableSignals,
			&analytic.LossSignals, &analytic.WinRatePercentage, &analytic.AvgPnLPercentage,
			&analytic.BestSignalPnL, &analytic.WorstSignalPnL, &analytic.AvgConfidence,
		)
		if err != nil {
			continue
		}
		analytics = append(analytics, analytic)
	}

	return analytics, nil
}

// Utility functions
func (s *SupabaseClient) GetCryptoBySymbol(symbol string) (*models.Cryptocurrency, error) {
	query := `SELECT id, symbol, name, coingecko_id FROM cryptocurrencies WHERE symbol = $1`

	crypto := &models.Cryptocurrency{}
	err := s.db.QueryRow(query, symbol).Scan(
		&crypto.ID, &crypto.Symbol, &crypto.Name, &crypto.CoingeckoID,
	)

	if err != nil {
		return nil, err
	}

	return crypto, nil
}

func (s *SupabaseClient) LogSystem(level, component, message string, context map[string]interface{}) error {
	query := `
		INSERT INTO system_logs (level, component, message, context, created_at)
		VALUES ($1, $2, $3, $4, $5)`

	contextJSON, _ := json.Marshal(context)

	_, err := s.db.Exec(query, level, component, message, contextJSON, time.Now())
	return err
}

// GetCryptocurrencies retrieves all cryptocurrencies from database
func (s *SupabaseClient) GetCryptocurrencies() ([]models.Cryptocurrency, error) {
	if s.useRest {
		return s.restClient.GetCryptocurrencies()
	}
	query := `SELECT id, symbol, name, cmc_id, contract_address, platform, slug, coingecko_id, is_active, created_at, updated_at FROM cryptocurrencies ORDER BY symbol`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query cryptocurrencies: %w", err)
	}
	defer rows.Close()

	var cryptos []models.Cryptocurrency
	for rows.Next() {
		var crypto models.Cryptocurrency
		err := rows.Scan(
			&crypto.ID,
			&crypto.Symbol,
			&crypto.Name,
			&crypto.CmcID,
			&crypto.ContractAddress,
			&crypto.Platform,
			&crypto.Slug,
			&crypto.CoingeckoID,
			&crypto.IsActive,
			&crypto.CreatedAt,
			&crypto.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cryptocurrency: %w", err)
		}
		cryptos = append(cryptos, crypto)
	}

	return cryptos, nil
}

// CreateCryptocurrency creates a new cryptocurrency record
func (s *SupabaseClient) CreateCryptocurrency(crypto *models.Cryptocurrency) error {
	if s.useRest {
		return s.restClient.CreateCryptocurrency(crypto)
	}
	query := `
		INSERT INTO cryptocurrencies (id, symbol, name, cmc_id, contract_address, platform, slug, coingecko_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	crypto.ID = uuid.New()
	crypto.CreatedAt = time.Now()
	now := time.Now()
	crypto.UpdatedAt = &now

	_, err := s.db.Exec(query,
		crypto.ID,
		crypto.Symbol,
		crypto.Name,
		crypto.CmcID,
		crypto.ContractAddress,
		crypto.Platform,
		crypto.Slug,
		crypto.CoingeckoID,
		crypto.IsActive,
		crypto.CreatedAt,
		crypto.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create cryptocurrency: %w", err)
	}

	return nil
}

// GetRecentSignals retrieves recent trading signals
func (s *SupabaseClient) GetRecentSignals(limit int) ([]models.TradingSignal, error) {
	if s.useRest {
		return s.restClient.GetRecentSignals(limit)
	}
	query := `
		SELECT id, crypto_id, action, confidence_score, entry_price, stop_loss,
		       take_profit_1, take_profit_2, market_conditions, created_at
		FROM trading_signals
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent signals: %w", err)
	}
	defer rows.Close()

	var signals []models.TradingSignal
	for rows.Next() {
		var signal models.TradingSignal
		var marketConditionsJSON []byte

		err := rows.Scan(
			&signal.ID,
			&signal.CryptoID,
			&signal.Action,
			&signal.ConfidenceScore,
			&signal.EntryPrice,
			&signal.StopLoss,
			&signal.TakeProfit1,
			&signal.TakeProfit2,
			&marketConditionsJSON,
			&signal.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan signal: %w", err)
		}

		// Parse market conditions JSON
		if len(marketConditionsJSON) > 0 {
			if err := json.Unmarshal(marketConditionsJSON, &signal.MarketConditions); err != nil {
				logrus.Warn("Failed to parse market conditions: ", err)
			}
		}

		signals = append(signals, signal)
	}

	return signals, nil
}

// GetSignalByID retrieves a specific trading signal by ID
func (s *SupabaseClient) GetSignalByID(id string) (*models.TradingSignal, error) {
	query := `
		SELECT id, crypto_id, action, confidence_score, entry_price, stop_loss,
		       take_profit_1, take_profit_2, market_conditions, created_at
		FROM trading_signals
		WHERE id = $1
	`

	var signal models.TradingSignal
	var marketConditionsJSON []byte

	err := s.db.QueryRow(query, id).Scan(
		&signal.ID,
		&signal.CryptoID,
		&signal.Action,
		&signal.ConfidenceScore,
		&signal.EntryPrice,
		&signal.StopLoss,
		&signal.TakeProfit1,
		&signal.TakeProfit2,
		&marketConditionsJSON,
		&signal.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get signal: %w", err)
	}

	// Parse market conditions JSON
	if len(marketConditionsJSON) > 0 {
		if err := json.Unmarshal(marketConditionsJSON, &signal.MarketConditions); err != nil {
			logrus.Warn("Failed to parse market conditions: ", err)
		}
	}

	return &signal, nil
}

// GetLearningInsights retrieves learning insights from analytics view
func (s *SupabaseClient) GetLearningInsights() (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total_learning_records,
			AVG(CASE WHEN outcome = 'WIN' THEN 1.0 ELSE 0.0 END) as win_rate,
			COUNT(CASE WHEN outcome = 'WIN' THEN 1 END) as total_wins,
			COUNT(CASE WHEN outcome = 'LOSS' THEN 1 END) as total_losses
		FROM learning_data
		WHERE created_at >= NOW() - INTERVAL '30 days'
	`

	var totalRecords, totalWins, totalLosses int
	var winRate float64

	err := s.db.QueryRow(query).Scan(&totalRecords, &winRate, &totalWins, &totalLosses)
	if err != nil {
		return nil, fmt.Errorf("failed to get learning insights: %w", err)
	}

	insights := map[string]interface{}{
		"total_learning_records": totalRecords,
		"win_rate":              winRate * 100, // Convert to percentage
		"total_wins":            totalWins,
		"total_losses":          totalLosses,
		"period":                "30 days",
	}

	return insights, nil
}

// TestConnection tests the database connection
func (s *SupabaseClient) TestConnection() error {
	return s.Ping()
}

// Helper function to extract project ID from Supabase URL
func extractProjectID(url string) string {
	// Extract project ID from URL like https://syojcjdcpufgyojnxhqa.supabase.co
	if len(url) > 8 && url[:8] == "https://" {
		url = url[8:] // Remove https://
	}

	// Find .supabase.co and extract project ID
	if idx := len(url); idx > 12 && url[idx-12:] == ".supabase.co" {
		return url[:idx-12] // Remove .supabase.co
	}

	// If no .supabase.co found, try to extract from the beginning
	if dotIndex := len(url); dotIndex > 0 {
		for i, char := range url {
			if char == '.' {
				return url[:i]
			}
		}
	}

	return ""
}
