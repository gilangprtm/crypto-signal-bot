package services

import (
	"crypto-signal-bot/internal/config"
	"crypto-signal-bot/internal/database"
	"crypto-signal-bot/internal/models"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type SignalGenerator struct {
	db  *database.SupabaseClient
	cfg *config.Config
}

type SignalDecision struct {
	Action          string
	Confidence      decimal.Decimal
	Reasoning       string
	EntryPrice      decimal.Decimal
	StopLoss        decimal.Decimal
	TakeProfit1     decimal.Decimal
	TakeProfit2     decimal.Decimal
	MarketConditions map[string]interface{}
}

func NewSignalGenerator(db *database.SupabaseClient, cfg *config.Config) *SignalGenerator {
	return &SignalGenerator{
		db:  db,
		cfg: cfg,
	}
}

func (sg *SignalGenerator) GenerateSignal(marketData *MarketData, indicators *TechnicalIndicators, crypto *models.Cryptocurrency) (*models.TradingSignal, error) {
	logrus.Debug("Generating signal for: ", marketData.Symbol)

	// Analyze market conditions and generate decision
	decision := sg.analyzeMarketConditions(marketData, indicators)

	// Check if confidence meets minimum threshold
	minConfidence := decimal.NewFromFloat(sg.cfg.MinConfidenceThreshold)
	if decision.Confidence.LessThan(minConfidence) {
		logrus.Debug("Signal confidence below threshold for ", marketData.Symbol, ": ", decision.Confidence)
		return nil, nil // No signal generated
	}

	// Check daily signal limit
	if sg.hasReachedDailyLimit() {
		logrus.Info("Daily signal limit reached, skipping signal generation")
		return nil, nil
	}

	// Create trading signal
	signal := &models.TradingSignal{
		ID:               uuid.New(),
		CryptoID:         crypto.ID,
		Action:           decision.Action,
		ConfidenceScore:  decision.Confidence,
		EntryPrice:       decision.EntryPrice,
		StopLoss:         &decision.StopLoss,
		TakeProfit1:      &decision.TakeProfit1,
		TakeProfit2:      &decision.TakeProfit2,
		Reasoning:        decision.Reasoning,
		
		// Technical indicators
		RSI:              &indicators.RSI,
		MACDLine:         &indicators.MACDLine,
		MACDSignal:       &indicators.MACDSignal,
		MACDHistogram:    &indicators.MACDHistogram,
		BBUpper:          &indicators.BBUpper,
		BBMiddle:         &indicators.BBMiddle,
		BBLower:          &indicators.BBLower,
		SMA20:            &indicators.SMA20,
		EMA12:            &indicators.EMA12,
		EMA26:            &indicators.EMA26,
		Volume24h:        &marketData.Volume24h,
		PriceChange24h:   &marketData.PriceChange24h,
		
		// Market sentiment
		FearGreedIndex:   &marketData.FearGreedIndex,
		MarketCap:        &marketData.MarketCap,
		
		// Additional context
		MarketConditions: decision.MarketConditions,
		Timeframe:        "15m",
		CreatedAt:        time.Now(),
		Status:           "active",
		
		// Related data
		Crypto:           crypto,
	}

	// Save signal to database
	if err := sg.db.CreateSignal(signal); err != nil {
		logrus.Error("Failed to save signal to database: ", err)
		return nil, err
	}

	logrus.Info("✅ Generated ", decision.Action, " signal for ", marketData.Symbol, " with confidence: ", decision.Confidence)
	return signal, nil
}

func (sg *SignalGenerator) analyzeMarketConditions(marketData *MarketData, indicators *TechnicalIndicators) *SignalDecision {
	var signals []string
	var confidenceFactors []decimal.Decimal
	var reasoning []string

	currentPrice := marketData.Price
	rsi := indicators.RSI
	macdLine := indicators.MACDLine
	macdSignal := indicators.MACDSignal
	macdHistogram := indicators.MACDHistogram
	bbUpper := indicators.BBUpper
	bbLower := indicators.BBLower
	_ = indicators.BBMiddle // Bollinger Bands middle line (not used in current logic)
	fearGreed := decimal.NewFromInt(int64(marketData.FearGreedIndex))

	// RSI Analysis
	rsiOversold := decimal.NewFromFloat(sg.cfg.RSIOversoldThreshold)
	rsiOverbought := decimal.NewFromFloat(sg.cfg.RSIOverboughtThreshold)

	if rsi.LessThan(rsiOversold) {
		signals = append(signals, "BUY")
		confidenceFactors = append(confidenceFactors, decimal.NewFromFloat(0.3))
		reasoning = append(reasoning, fmt.Sprintf("RSI oversold (%.2f)", rsi.InexactFloat64()))
	} else if rsi.GreaterThan(rsiOverbought) {
		signals = append(signals, "SELL")
		confidenceFactors = append(confidenceFactors, decimal.NewFromFloat(0.3))
		reasoning = append(reasoning, fmt.Sprintf("RSI overbought (%.2f)", rsi.InexactFloat64()))
	}

	// MACD Analysis
	if macdLine.GreaterThan(macdSignal) && macdHistogram.GreaterThan(decimal.Zero) {
		signals = append(signals, "BUY")
		confidenceFactors = append(confidenceFactors, decimal.NewFromFloat(0.25))
		reasoning = append(reasoning, "MACD bullish crossover")
	} else if macdLine.LessThan(macdSignal) && macdHistogram.LessThan(decimal.Zero) {
		signals = append(signals, "SELL")
		confidenceFactors = append(confidenceFactors, decimal.NewFromFloat(0.25))
		reasoning = append(reasoning, "MACD bearish crossover")
	}

	// Bollinger Bands Analysis
	if currentPrice.LessThan(bbLower) {
		signals = append(signals, "BUY")
		confidenceFactors = append(confidenceFactors, decimal.NewFromFloat(0.2))
		reasoning = append(reasoning, "Price below lower Bollinger Band")
	} else if currentPrice.GreaterThan(bbUpper) {
		signals = append(signals, "SELL")
		confidenceFactors = append(confidenceFactors, decimal.NewFromFloat(0.2))
		reasoning = append(reasoning, "Price above upper Bollinger Band")
	}

	// Fear & Greed Index Analysis
	fearGreedMin := decimal.NewFromInt(int64(sg.cfg.FearGreedMinThreshold))
	fearGreedMax := decimal.NewFromInt(int64(sg.cfg.FearGreedMaxThreshold))

	if fearGreed.LessThan(fearGreedMin) {
		signals = append(signals, "BUY")
		confidenceFactors = append(confidenceFactors, decimal.NewFromFloat(0.15))
		reasoning = append(reasoning, fmt.Sprintf("Extreme fear in market (%d)", marketData.FearGreedIndex))
	} else if fearGreed.GreaterThan(fearGreedMax) {
		signals = append(signals, "SELL")
		confidenceFactors = append(confidenceFactors, decimal.NewFromFloat(0.15))
		reasoning = append(reasoning, fmt.Sprintf("Extreme greed in market (%d)", marketData.FearGreedIndex))
	}

	// Price Action Analysis
	if currentPrice.GreaterThan(indicators.SMA20) && indicators.EMA12.GreaterThan(indicators.EMA26) {
		signals = append(signals, "BUY")
		confidenceFactors = append(confidenceFactors, decimal.NewFromFloat(0.1))
		reasoning = append(reasoning, "Price above SMA20 with bullish EMA crossover")
	} else if currentPrice.LessThan(indicators.SMA20) && indicators.EMA12.LessThan(indicators.EMA26) {
		signals = append(signals, "SELL")
		confidenceFactors = append(confidenceFactors, decimal.NewFromFloat(0.1))
		reasoning = append(reasoning, "Price below SMA20 with bearish EMA crossover")
	}

	// Determine final signal
	buySignals := 0
	sellSignals := 0
	totalConfidence := decimal.Zero

	for i, signal := range signals {
		if signal == "BUY" {
			buySignals++
		} else if signal == "SELL" {
			sellSignals++
		}
		totalConfidence = totalConfidence.Add(confidenceFactors[i])
	}

	// Decision logic
	var action string
	var confidence decimal.Decimal

	if buySignals > sellSignals {
		action = "BUY"
		confidence = totalConfidence.Mul(decimal.NewFromFloat(float64(buySignals) / float64(len(signals))))
	} else if sellSignals > buySignals {
		action = "SELL"
		confidence = totalConfidence.Mul(decimal.NewFromFloat(float64(sellSignals) / float64(len(signals))))
	} else {
		action = "HOLD"
		confidence = decimal.NewFromFloat(0.1) // Low confidence for hold
	}

	// Calculate price targets
	stopLossPercent := decimal.NewFromFloat(sg.cfg.StopLossPercentage / 100)
	takeProfit1Percent := decimal.NewFromFloat(sg.cfg.TakeProfit1Percentage / 100)
	takeProfit2Percent := decimal.NewFromFloat(sg.cfg.TakeProfit2Percentage / 100)

	var stopLoss, takeProfit1, takeProfit2 decimal.Decimal

	if action == "BUY" {
		stopLoss = currentPrice.Mul(decimal.NewFromInt(1).Sub(stopLossPercent))
		takeProfit1 = currentPrice.Mul(decimal.NewFromInt(1).Add(takeProfit1Percent))
		takeProfit2 = currentPrice.Mul(decimal.NewFromInt(1).Add(takeProfit2Percent))
	} else if action == "SELL" {
		stopLoss = currentPrice.Mul(decimal.NewFromInt(1).Add(stopLossPercent))
		takeProfit1 = currentPrice.Mul(decimal.NewFromInt(1).Sub(takeProfit1Percent))
		takeProfit2 = currentPrice.Mul(decimal.NewFromInt(1).Sub(takeProfit2Percent))
	}

	// Market conditions context
	marketConditions := map[string]interface{}{
		"rsi":                rsi.InexactFloat64(),
		"macd_histogram":     macdHistogram.InexactFloat64(),
		"bb_position":        sg.calculateBBPosition(currentPrice, bbUpper, bbLower),
		"fear_greed_index":   marketData.FearGreedIndex,
		"price_change_24h":   marketData.PriceChange24h.InexactFloat64(),
		"volume_24h":         marketData.Volume24h.InexactFloat64(),
		"buy_signals":        buySignals,
		"sell_signals":       sellSignals,
		"total_signals":      len(signals),
	}

	return &SignalDecision{
		Action:           action,
		Confidence:       confidence,
		Reasoning:        fmt.Sprintf("%s", reasoning),
		EntryPrice:       currentPrice,
		StopLoss:         stopLoss,
		TakeProfit1:      takeProfit1,
		TakeProfit2:      takeProfit2,
		MarketConditions: marketConditions,
	}
}

func (sg *SignalGenerator) calculateBBPosition(price, upper, lower decimal.Decimal) float64 {
	if upper.Equal(lower) {
		return 0.5
	}
	position := price.Sub(lower).Div(upper.Sub(lower))
	return position.InexactFloat64()
}

func (sg *SignalGenerator) hasReachedDailyLimit() bool {
	// TODO: Implement daily signal count check from database
	// For now, return false
	return false
}
