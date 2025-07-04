package services

import (
	"crypto-signal-bot/internal/config"
	"crypto-signal-bot/internal/database"
	"crypto-signal-bot/internal/models"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type BotService struct {
	db                  *database.SupabaseClient
	cfg                 *config.Config
	dataCollector       *DataCollector
	technicalAnalyzer   *TechnicalAnalyzer
	signalGenerator     *SignalGenerator
	notificationService *NotificationService
	learningEngine      *LearningEngine
	
	// Runtime state
	isRunning           bool
	lastAnalysisTime    time.Time
	totalSignalsToday   int
	cryptoList          []*models.Cryptocurrency
}

func NewBotService(db *database.SupabaseClient, cfg *config.Config) *BotService {
	bs := &BotService{
		db:                  db,
		cfg:                 cfg,
		dataCollector:       NewDataCollector(cfg),
		technicalAnalyzer:   NewTechnicalAnalyzer(cfg),
		signalGenerator:     NewSignalGenerator(db, cfg),
		notificationService: NewNotificationService(cfg),
		learningEngine:      NewLearningEngine(db, cfg),
		isRunning:           false,
		cryptoList:          []*models.Cryptocurrency{},
	}

	// Set bot service reference for notification service
	bs.notificationService.SetBotService(bs)

	return bs
}

func (bs *BotService) Start() error {
	logrus.Info("🚀 Starting Crypto Signal Bot...")

	// Initialize cryptocurrency list
	if err := bs.initializeCryptoList(); err != nil {
		return err
	}

	// Test connections
	if err := bs.testConnections(); err != nil {
		logrus.Warn("Some connections failed during startup: ", err)
	}

	// Start Telegram bot with interactive menu
	if err := bs.notificationService.StartTelegramBot(); err != nil {
		logrus.Warn("Failed to start Telegram bot with menu: ", err)
	}

	// Send startup notification
	bs.notificationService.SendSystemNotification("info", "🤖 Crypto Signal Bot started successfully!\n\nGunakan /menu untuk mengakses fitur interaktif.")

	bs.isRunning = true
	logrus.Info("✅ Crypto Signal Bot is now running")

	return nil
}

func (bs *BotService) Stop() error {
	logrus.Info("🛑 Stopping Crypto Signal Bot...")

	bs.isRunning = false

	// Send shutdown notification
	bs.notificationService.SendSystemNotification("info", "🤖 Crypto Signal Bot stopped")

	logrus.Info("✅ Crypto Signal Bot stopped successfully")
	return nil
}

func (bs *BotService) RunAnalysis() error {
	if !bs.isRunning {
		return nil
	}

	logrus.Info("🔍 Running market analysis...")
	bs.lastAnalysisTime = time.Now()

	// Check daily signal limit
	if bs.totalSignalsToday >= bs.cfg.MaxSignalsPerDay {
		logrus.Info("Daily signal limit reached, skipping analysis")
		return nil
	}

	signalsGenerated := 0

	// Analyze each cryptocurrency
	for _, crypto := range bs.cryptoList {
		if err := bs.analyzeCryptocurrency(crypto); err != nil {
			logrus.Error("Failed to analyze ", crypto.Symbol, ": ", err)
			continue
		}
		
		// Rate limiting between analyses
		time.Sleep(time.Duration(bs.cfg.AnalysisIntervalSeconds) * time.Second / time.Duration(len(bs.cryptoList)))
	}

	// Update performance tracking
	if err := bs.updatePerformanceTracking(); err != nil {
		logrus.Error("Failed to update performance tracking: ", err)
	}

	// Run learning optimization (daily)
	if bs.shouldRunLearningOptimization() {
		if err := bs.learningEngine.OptimizeStrategy(); err != nil {
			logrus.Error("Failed to run learning optimization: ", err)
		}
	}

	logrus.Info("✅ Market analysis completed. Signals generated: ", signalsGenerated)
	return nil
}

func (bs *BotService) analyzeCryptocurrency(crypto *models.Cryptocurrency) error {
	logrus.Debug("Analyzing cryptocurrency: ", crypto.Symbol)

	// Collect market data
	marketData, err := bs.dataCollector.GetMarketData(crypto.Symbol)
	if err != nil {
		return err
	}

	// Perform technical analysis
	indicators, err := bs.technicalAnalyzer.AnalyzeMarketData(marketData)
	if err != nil {
		return err
	}

	// Save market snapshot
	if err := bs.saveMarketSnapshot(crypto, marketData, indicators); err != nil {
		logrus.Error("Failed to save market snapshot: ", err)
	}

	// Extract features for learning
	features := bs.learningEngine.ExtractFeatures(marketData, indicators)

	// Predict signal outcome using learning engine
	predictedOutcome, predictedConfidence, err := bs.learningEngine.PredictSignalOutcome(features)
	if err != nil {
		logrus.Error("Failed to predict signal outcome: ", err)
	}

	// Generate trading signal
	signal, err := bs.signalGenerator.GenerateSignal(marketData, indicators, crypto)
	if err != nil {
		return err
	}

	// If signal was generated
	if signal != nil {
		// Save learning data
		if err := bs.learningEngine.SaveLearningData(signal, features, predictedOutcome, predictedConfidence); err != nil {
			logrus.Error("Failed to save learning data: ", err)
		}

		// Send notification
		if err := bs.notificationService.SendSignalNotification(signal); err != nil {
			logrus.Error("Failed to send signal notification: ", err)
		}

		bs.totalSignalsToday++
		logrus.Info("✅ Signal generated and sent for ", crypto.Symbol)
	}

	return nil
}

func (bs *BotService) saveMarketSnapshot(crypto *models.Cryptocurrency, marketData *MarketData, indicators *TechnicalIndicators) error {
	// Skip saving if database is not available
	if bs.db == nil {
		logrus.Debug("Database not available, skipping market snapshot save")
		return nil
	}

	snapshot := &models.MarketSnapshot{
		ID:                 uuid.New(),
		CryptocurrencyID:   crypto.ID,
		Price:              marketData.Price,
		Volume24h:          marketData.Volume24h,
		MarketCap:          marketData.MarketCap,
		PriceChange1h:      marketData.PriceChange1h,
		PriceChange24h:     marketData.PriceChange24h,
		PriceChange7d:      marketData.PriceChange7d,
		FearGreedIndex:     marketData.FearGreedIndex,
		Timestamp:          marketData.Timestamp,
		Crypto:             crypto,
	}

	// Add technical indicators only if available
	if indicators != nil {
		snapshot.RSI = indicators.RSI
		snapshot.MACDLine = indicators.MACDLine
		snapshot.MACDSignal = indicators.MACDSignal
		snapshot.MACDHistogram = indicators.MACDHistogram
		snapshot.BBUpper = indicators.BBUpper
		snapshot.BBMiddle = indicators.BBMiddle
		snapshot.BBLower = indicators.BBLower
		snapshot.SMA20 = indicators.SMA20
		snapshot.EMA12 = indicators.EMA12
		snapshot.EMA26 = indicators.EMA26
	}

	return bs.db.SaveMarketSnapshot(snapshot)
}

func (bs *BotService) updatePerformanceTracking() error {
	// TODO: Implement performance tracking update
	// This would check active signals and update their performance
	// based on current market prices
	logrus.Debug("Updating performance tracking...")
	return nil
}

func (bs *BotService) shouldRunLearningOptimization() bool {
	// Run learning optimization once per day
	now := time.Now()
	return now.Hour() == 0 && now.Minute() < 30 // Run between 00:00-00:30
}

func (bs *BotService) initializeCryptoList() error {
	logrus.Info("Initializing cryptocurrency list...")

	// Default cryptocurrencies to monitor
	defaultCryptos := []struct {
		Symbol      string
		Name        string
		CoingeckoID string
	}{
		{"BTC", "Bitcoin", "bitcoin"},
		{"ETH", "Ethereum", "ethereum"},
		{"BNB", "Binance Coin", "binancecoin"},
		{"ADA", "Cardano", "cardano"},
		{"SOL", "Solana", "solana"},
		{"DOT", "Polkadot", "polkadot"},
		{"MATIC", "Polygon", "matic-network"},
		{"AVAX", "Avalanche", "avalanche-2"},
		{"LINK", "Chainlink", "chainlink"},
		{"ATOM", "Cosmos", "cosmos"},
	}

	// If database is not available, use default list
	if bs.db == nil {
		logrus.Warn("Database not available, using default cryptocurrency list")
		for _, defaultCrypto := range defaultCryptos {
			newCrypto := &models.Cryptocurrency{
				ID:        uuid.New(),
				Symbol:    defaultCrypto.Symbol,
				Name:      defaultCrypto.Name,
				IsActive:  true,
				CreatedAt: time.Now(),
			}
			bs.cryptoList = append(bs.cryptoList, newCrypto)
		}
		logrus.Infof("✅ Initialized %d cryptocurrencies (offline mode)", len(bs.cryptoList))
		return nil
	}

	// Get existing cryptocurrencies from database
	existingCryptos, err := bs.db.GetCryptocurrencies()
	if err != nil {
		logrus.Warnf("Failed to get cryptocurrencies from database: %v, using defaults", err)
		// Fallback to default list
		for _, defaultCrypto := range defaultCryptos {
			newCrypto := &models.Cryptocurrency{
				ID:        uuid.New(),
				Symbol:    defaultCrypto.Symbol,
				Name:      defaultCrypto.Name,
				IsActive:  true,
				CreatedAt: time.Now(),
			}
			bs.cryptoList = append(bs.cryptoList, newCrypto)
		}
		logrus.Infof("✅ Initialized %d cryptocurrencies (fallback mode)", len(bs.cryptoList))
		return nil
	}

	// Create map of existing symbols
	existingMap := make(map[string]*models.Cryptocurrency)
	for _, crypto := range existingCryptos {
		existingMap[crypto.Symbol] = &crypto
	}

	// Add missing cryptocurrencies
	for _, defaultCrypto := range defaultCryptos {
		if existing, exists := existingMap[defaultCrypto.Symbol]; exists {
			bs.cryptoList = append(bs.cryptoList, existing)
		} else {
			// Create new cryptocurrency
			newCrypto := &models.Cryptocurrency{
				ID:        uuid.New(),
				Symbol:    defaultCrypto.Symbol,
				Name:      defaultCrypto.Name,
				IsActive:  true,
				CreatedAt: time.Now(),
			}

			if err := bs.db.CreateCryptocurrency(newCrypto); err != nil {
				logrus.Error("Failed to create cryptocurrency ", defaultCrypto.Symbol, ": ", err)
				continue
			}

			bs.cryptoList = append(bs.cryptoList, newCrypto)
			logrus.Info("Added new cryptocurrency: ", defaultCrypto.Symbol)
		}
	}

	logrus.Info("✅ Cryptocurrency list initialized with ", len(bs.cryptoList), " coins")
	return nil
}

func (bs *BotService) testConnections() error {
	logrus.Info("Testing connections...")

	// Test Telegram connection
	if err := bs.notificationService.TestConnection(); err != nil {
		logrus.Error("Telegram connection test failed: ", err)
		return err
	}

	// Test database connection (only if available)
	if bs.db != nil {
		if err := bs.db.TestConnection(); err != nil {
			logrus.Warn("Database connection test failed (continuing without database): ", err)
			// Don't return error, continue without database
		} else {
			logrus.Info("✅ Database connection test passed")
		}
	} else {
		logrus.Warn("⚠️ Database not available, skipping database test")
	}

	// Test data collector (get BTC data)
	if _, err := bs.dataCollector.GetMarketData("BTC"); err != nil {
		logrus.Error("Data collector test failed: ", err)
		return err
	}

	logrus.Info("✅ Essential connections tested successfully")
	return nil
}

func (bs *BotService) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"is_running":           bs.isRunning,
		"last_analysis_time":   bs.lastAnalysisTime,
		"total_signals_today":  bs.totalSignalsToday,
		"monitored_cryptos":    len(bs.cryptoList),
		"max_signals_per_day":  bs.cfg.MaxSignalsPerDay,
	}
}

func (bs *BotService) SendDailySummary() error {
	analytics, err := bs.db.GetSignalAnalytics()
	if err != nil {
		return err
	}

	return bs.notificationService.SendDailySummary(analytics)
}

func (bs *BotService) GetPerformanceMetrics() (*PerformanceMetrics, error) {
	return bs.learningEngine.AnalyzePatterns()
}
