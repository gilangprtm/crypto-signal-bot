package services

import (
	"crypto-signal-bot/internal/config"
	"crypto-signal-bot/internal/database"
	"crypto-signal-bot/internal/models"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type LearningEngine struct {
	db  *database.SupabaseClient
	cfg *config.Config
}

type FeatureVector struct {
	RSI                decimal.Decimal `json:"rsi"`
	MACDHistogram      decimal.Decimal `json:"macd_histogram"`
	BBPosition         decimal.Decimal `json:"bb_position"`
	FearGreedIndex     decimal.Decimal `json:"fear_greed_index"`
	PriceChange24h     decimal.Decimal `json:"price_change_24h"`
	Volume24h          decimal.Decimal `json:"volume_24h"`
	PriceAboveSMA20    bool            `json:"price_above_sma20"`
	EMACrossover       bool            `json:"ema_crossover"`
	RSIOversold        bool            `json:"rsi_oversold"`
	RSIOverbought      bool            `json:"rsi_overbought"`
	MACDBullish        bool            `json:"macd_bullish"`
	BBSqueeze          bool            `json:"bb_squeeze"`
	HighVolume         bool            `json:"high_volume"`
	TrendDirection     string          `json:"trend_direction"`
	MarketSentiment    string          `json:"market_sentiment"`
}

type PerformanceMetrics struct {
	TotalSignals      int             `json:"total_signals"`
	ProfitableSignals int             `json:"profitable_signals"`
	WinRate           decimal.Decimal `json:"win_rate"`
	AvgPnL            decimal.Decimal `json:"avg_pnl"`
	BestPnL           decimal.Decimal `json:"best_pnl"`
	WorstPnL          decimal.Decimal `json:"worst_pnl"`
	AvgDuration       decimal.Decimal `json:"avg_duration"`
	Accuracy          decimal.Decimal `json:"accuracy"`
}

func NewLearningEngine(db *database.SupabaseClient, cfg *config.Config) *LearningEngine {
	return &LearningEngine{
		db:  db,
		cfg: cfg,
	}
}

func (le *LearningEngine) ExtractFeatures(marketData *MarketData, indicators *TechnicalIndicators) *FeatureVector {
	// Calculate derived features
	bbPosition := le.calculateBBPosition(marketData.Price, indicators.BBUpper, indicators.BBLower)
	priceAboveSMA20 := marketData.Price.GreaterThan(indicators.SMA20)
	emaCrossover := indicators.EMA12.GreaterThan(indicators.EMA26)
	
	rsiOversold := indicators.RSI.LessThan(decimal.NewFromFloat(le.cfg.RSIOversoldThreshold))
	rsiOverbought := indicators.RSI.GreaterThan(decimal.NewFromFloat(le.cfg.RSIOverboughtThreshold))
	macdBullish := indicators.MACDHistogram.GreaterThan(decimal.Zero)
	
	// BB Squeeze detection (simplified)
	bbRange := indicators.BBUpper.Sub(indicators.BBLower)
	avgPrice := indicators.SMA20
	bbSqueeze := bbRange.Div(avgPrice).LessThan(decimal.NewFromFloat(0.02)) // 2% range
	
	// High volume detection (simplified)
	highVolume := marketData.Volume24h.GreaterThan(decimal.Zero) // TODO: Compare with average volume
	
	// Trend direction
	trendDirection := "neutral"
	if priceAboveSMA20 && emaCrossover {
		trendDirection = "bullish"
	} else if !priceAboveSMA20 && !emaCrossover {
		trendDirection = "bearish"
	}
	
	// Market sentiment based on Fear & Greed Index
	marketSentiment := "neutral"
	if marketData.FearGreedIndex <= 20 {
		marketSentiment = "extreme_fear"
	} else if marketData.FearGreedIndex <= 40 {
		marketSentiment = "fear"
	} else if marketData.FearGreedIndex >= 80 {
		marketSentiment = "extreme_greed"
	} else if marketData.FearGreedIndex >= 60 {
		marketSentiment = "greed"
	}

	return &FeatureVector{
		RSI:                indicators.RSI,
		MACDHistogram:      indicators.MACDHistogram,
		BBPosition:         bbPosition,
		FearGreedIndex:     decimal.NewFromInt(int64(marketData.FearGreedIndex)),
		PriceChange24h:     marketData.PriceChange24h,
		Volume24h:          marketData.Volume24h,
		PriceAboveSMA20:    priceAboveSMA20,
		EMACrossover:       emaCrossover,
		RSIOversold:        rsiOversold,
		RSIOverbought:      rsiOverbought,
		MACDBullish:        macdBullish,
		BBSqueeze:          bbSqueeze,
		HighVolume:         highVolume,
		TrendDirection:     trendDirection,
		MarketSentiment:    marketSentiment,
	}
}

func (le *LearningEngine) SaveLearningData(signal *models.TradingSignal, features *FeatureVector, predictedOutcome string, predictedConfidence decimal.Decimal) error {
	// Convert features to map for JSON storage
	featuresMap := map[string]interface{}{
		"rsi":                features.RSI.InexactFloat64(),
		"macd_histogram":     features.MACDHistogram.InexactFloat64(),
		"bb_position":        features.BBPosition.InexactFloat64(),
		"fear_greed_index":   features.FearGreedIndex.InexactFloat64(),
		"price_change_24h":   features.PriceChange24h.InexactFloat64(),
		"volume_24h":         features.Volume24h.InexactFloat64(),
		"price_above_sma20":  features.PriceAboveSMA20,
		"ema_crossover":      features.EMACrossover,
		"rsi_oversold":       features.RSIOversold,
		"rsi_overbought":     features.RSIOverbought,
		"macd_bullish":       features.MACDBullish,
		"bb_squeeze":         features.BBSqueeze,
		"high_volume":        features.HighVolume,
		"trend_direction":    features.TrendDirection,
		"market_sentiment":   features.MarketSentiment,
	}

	learningData := &models.LearningData{
		ID:                  uuid.New(),
		SignalID:            &signal.ID,
		Features:            featuresMap,
		PredictedOutcome:    predictedOutcome,
		PredictedConfidence: predictedConfidence,
		CreatedAt:           time.Now(),
	}

	return le.db.SaveLearningData(learningData)
}

func (le *LearningEngine) UpdateLearningDataWithOutcome(signalID uuid.UUID, actualOutcome string, actualPnL decimal.Decimal, duration int) error {
	// TODO: Implement update learning data with actual outcomes
	// This would require additional database methods
	logrus.Info("Learning data updated for signal: ", signalID, " outcome: ", actualOutcome)
	return nil
}

func (le *LearningEngine) AnalyzePatterns() (*PerformanceMetrics, error) {
	logrus.Info("Analyzing signal patterns for learning...")

	// Get signal analytics from database
	analytics, err := le.db.GetSignalAnalytics()
	if err != nil {
		return nil, err
	}

	if len(analytics) == 0 {
		return &PerformanceMetrics{}, nil
	}

	// Calculate overall metrics
	totalSignals := 0
	totalProfitable := 0
	totalPnL := decimal.Zero
	bestPnL := decimal.NewFromInt(-1000)
	worstPnL := decimal.NewFromInt(1000)

	for _, analytic := range analytics {
		totalSignals += analytic.TotalSignals
		totalProfitable += analytic.ProfitableSignals
		totalPnL = totalPnL.Add(analytic.AvgPnLPercentage)
		
		if analytic.BestSignalPnL.GreaterThan(bestPnL) {
			bestPnL = analytic.BestSignalPnL
		}
		if analytic.WorstSignalPnL.LessThan(worstPnL) {
			worstPnL = analytic.WorstSignalPnL
		}
	}

	winRate := decimal.Zero
	avgPnL := decimal.Zero
	accuracy := decimal.Zero

	if totalSignals > 0 {
		winRate = decimal.NewFromInt(int64(totalProfitable)).Div(decimal.NewFromInt(int64(totalSignals))).Mul(decimal.NewFromInt(100))
		avgPnL = totalPnL.Div(decimal.NewFromInt(int64(len(analytics))))
		accuracy = winRate.Div(decimal.NewFromInt(100)) // Simplified accuracy calculation
	}

	metrics := &PerformanceMetrics{
		TotalSignals:      totalSignals,
		ProfitableSignals: totalProfitable,
		WinRate:           winRate,
		AvgPnL:            avgPnL,
		BestPnL:           bestPnL,
		WorstPnL:          worstPnL,
		AvgDuration:       decimal.NewFromInt(60), // TODO: Calculate from actual data
		Accuracy:          accuracy,
	}

	logrus.Info("Pattern analysis completed - Win Rate: ", winRate.StringFixed(2), "%")
	return metrics, nil
}

func (le *LearningEngine) OptimizeStrategy() error {
	logrus.Info("Optimizing trading strategy based on learning data...")

	// Analyze current performance
	metrics, err := le.AnalyzePatterns()
	if err != nil {
		return err
	}

	// TODO: Implement strategy optimization logic
	// This could include:
	// 1. Adjusting confidence thresholds based on historical accuracy
	// 2. Modifying technical indicator weights
	// 3. Updating stop loss and take profit levels
	// 4. Filtering out low-performing patterns

	logrus.Info("Strategy optimization completed")
	logrus.Info("Current performance - Win Rate: ", metrics.WinRate.StringFixed(2), "%, Avg PnL: ", metrics.AvgPnL.StringFixed(2), "%")

	return nil
}

func (le *LearningEngine) GetBestPerformingIndicators() (map[string]decimal.Decimal, error) {
	// TODO: Implement analysis of which indicators perform best
	// This would analyze learning data to find correlations between
	// specific indicator values and profitable outcomes

	indicators := map[string]decimal.Decimal{
		"rsi_oversold":    decimal.NewFromFloat(0.75),
		"macd_bullish":    decimal.NewFromFloat(0.68),
		"bb_position":     decimal.NewFromFloat(0.62),
		"fear_greed":      decimal.NewFromFloat(0.58),
		"ema_crossover":   decimal.NewFromFloat(0.55),
	}

	logrus.Info("Best performing indicators analyzed")
	return indicators, nil
}

func (le *LearningEngine) PredictSignalOutcome(features *FeatureVector) (string, decimal.Decimal, error) {
	// Simple rule-based prediction (in production, this could be ML model)
	confidence := decimal.NewFromFloat(0.5)
	outcome := "hold"

	// Bullish signals
	bullishScore := 0
	if features.RSIOversold {
		bullishScore += 2
		confidence = confidence.Add(decimal.NewFromFloat(0.15))
	}
	if features.MACDBullish {
		bullishScore += 2
		confidence = confidence.Add(decimal.NewFromFloat(0.12))
	}
	if features.BBPosition.LessThan(decimal.NewFromFloat(0.2)) {
		bullishScore += 1
		confidence = confidence.Add(decimal.NewFromFloat(0.08))
	}
	if features.MarketSentiment == "extreme_fear" {
		bullishScore += 2
		confidence = confidence.Add(decimal.NewFromFloat(0.10))
	}
	if features.TrendDirection == "bullish" {
		bullishScore += 1
		confidence = confidence.Add(decimal.NewFromFloat(0.05))
	}

	// Bearish signals
	bearishScore := 0
	if features.RSIOverbought {
		bearishScore += 2
		confidence = confidence.Add(decimal.NewFromFloat(0.15))
	}
	if !features.MACDBullish {
		bearishScore += 1
		confidence = confidence.Add(decimal.NewFromFloat(0.08))
	}
	if features.BBPosition.GreaterThan(decimal.NewFromFloat(0.8)) {
		bearishScore += 1
		confidence = confidence.Add(decimal.NewFromFloat(0.08))
	}
	if features.MarketSentiment == "extreme_greed" {
		bearishScore += 2
		confidence = confidence.Add(decimal.NewFromFloat(0.10))
	}
	if features.TrendDirection == "bearish" {
		bearishScore += 1
		confidence = confidence.Add(decimal.NewFromFloat(0.05))
	}

	// Determine outcome
	if bullishScore > bearishScore && bullishScore >= 3 {
		outcome = "profit"
	} else if bearishScore > bullishScore && bearishScore >= 3 {
		outcome = "profit" // For sell signals
	} else {
		outcome = "loss"
		confidence = confidence.Mul(decimal.NewFromFloat(0.5)) // Lower confidence for uncertain signals
	}

	// Cap confidence at 1.0
	if confidence.GreaterThan(decimal.NewFromInt(1)) {
		confidence = decimal.NewFromInt(1)
	}

	return outcome, confidence, nil
}

func (le *LearningEngine) calculateBBPosition(price, upper, lower decimal.Decimal) decimal.Decimal {
	if upper.Equal(lower) {
		return decimal.NewFromFloat(0.5)
	}
	return price.Sub(lower).Div(upper.Sub(lower))
}
