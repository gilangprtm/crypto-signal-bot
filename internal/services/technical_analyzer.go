package services

import (
	"crypto-signal-bot/internal/config"
	"math"
	"strconv"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type TechnicalAnalyzer struct {
	cfg *config.Config
}

type TechnicalIndicators struct {
	RSI           decimal.Decimal
	MACDLine      decimal.Decimal
	MACDSignal    decimal.Decimal
	MACDHistogram decimal.Decimal
	BBUpper       decimal.Decimal
	BBMiddle      decimal.Decimal
	BBLower       decimal.Decimal
	SMA20         decimal.Decimal
	EMA12         decimal.Decimal
	EMA26         decimal.Decimal
	Volume        decimal.Decimal
	
	// Additional indicators for decision making
	StochK        decimal.Decimal
	StochD        decimal.Decimal
	Williams      decimal.Decimal
	
	// Price action
	CurrentPrice  decimal.Decimal
	PreviousPrice decimal.Decimal
	HighestHigh   decimal.Decimal
	LowestLow     decimal.Decimal
}

type OHLCV struct {
	Open      decimal.Decimal
	High      decimal.Decimal
	Low       decimal.Decimal
	Close     decimal.Decimal
	Volume    decimal.Decimal
	Timestamp int64
}

func NewTechnicalAnalyzer(cfg *config.Config) *TechnicalAnalyzer {
	return &TechnicalAnalyzer{
		cfg: cfg,
	}
}

func (ta *TechnicalAnalyzer) AnalyzeMarketData(marketData *MarketData) (*TechnicalIndicators, error) {
	logrus.Debug("Analyzing technical indicators for: ", marketData.Symbol)

	// Convert kline data to OHLCV format
	ohlcvData, err := ta.parseKlineData(marketData.KlineData)
	if err != nil {
		return nil, err
	}

	if len(ohlcvData) < 26 {
		logrus.Warn("Insufficient data for technical analysis, need at least 26 periods")
		return nil, err
	}

	indicators := &TechnicalIndicators{
		CurrentPrice: marketData.Price,
		Volume:       marketData.Volume24h,
	}

	// Extract close prices for calculations
	closePrices := make([]decimal.Decimal, len(ohlcvData))
	highPrices := make([]decimal.Decimal, len(ohlcvData))
	lowPrices := make([]decimal.Decimal, len(ohlcvData))
	
	for i, ohlcv := range ohlcvData {
		closePrices[i] = ohlcv.Close
		highPrices[i] = ohlcv.High
		lowPrices[i] = ohlcv.Low
	}

	// Calculate RSI (14 periods)
	indicators.RSI = ta.calculateRSI(closePrices, 14)

	// Calculate MACD (12, 26, 9)
	indicators.EMA12 = ta.calculateEMA(closePrices, 12)
	indicators.EMA26 = ta.calculateEMA(closePrices, 26)
	indicators.MACDLine = indicators.EMA12.Sub(indicators.EMA26)
	
	// Calculate MACD Signal line (9-period EMA of MACD line)
	macdValues := ta.calculateMACDHistory(closePrices, 12, 26)
	indicators.MACDSignal = ta.calculateEMA(macdValues, 9)
	indicators.MACDHistogram = indicators.MACDLine.Sub(indicators.MACDSignal)

	// Calculate Bollinger Bands (20 periods, 2 std dev)
	indicators.SMA20 = ta.calculateSMA(closePrices, 20)
	indicators.BBMiddle = indicators.SMA20
	stdDev := ta.calculateStandardDeviation(closePrices, 20)
	indicators.BBUpper = indicators.BBMiddle.Add(stdDev.Mul(decimal.NewFromInt(2)))
	indicators.BBLower = indicators.BBMiddle.Sub(stdDev.Mul(decimal.NewFromInt(2)))

	// Calculate additional indicators
	indicators.StochK, indicators.StochD = ta.calculateStochastic(highPrices, lowPrices, closePrices, 14, 3)
	indicators.Williams = ta.calculateWilliamsR(highPrices, lowPrices, closePrices, 14)

	// Price action analysis
	if len(closePrices) > 1 {
		indicators.PreviousPrice = closePrices[len(closePrices)-2]
	}
	indicators.HighestHigh = ta.findHighest(highPrices, 20)
	indicators.LowestLow = ta.findLowest(lowPrices, 20)

	logrus.Debug("Technical analysis completed for: ", marketData.Symbol)
	return indicators, nil
}

func (ta *TechnicalAnalyzer) parseKlineData(klineData [][]interface{}) ([]OHLCV, error) {
	var ohlcvData []OHLCV

	for _, kline := range klineData {
		if len(kline) < 6 {
			continue
		}

		// Parse timestamp
		timestamp, _ := kline[0].(float64)
		
		// Parse OHLCV values
		openStr, _ := kline[1].(string)
		highStr, _ := kline[2].(string)
		lowStr, _ := kline[3].(string)
		closeStr, _ := kline[4].(string)
		volumeStr, _ := kline[5].(string)

		open, _ := decimal.NewFromString(openStr)
		high, _ := decimal.NewFromString(highStr)
		low, _ := decimal.NewFromString(lowStr)
		close, _ := decimal.NewFromString(closeStr)
		volume, _ := decimal.NewFromString(volumeStr)

		ohlcv := OHLCV{
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
			Timestamp: int64(timestamp),
		}

		ohlcvData = append(ohlcvData, ohlcv)
	}

	return ohlcvData, nil
}

func (ta *TechnicalAnalyzer) calculateRSI(prices []decimal.Decimal, period int) decimal.Decimal {
	if len(prices) < period+1 {
		return decimal.Zero
	}

	gains := decimal.Zero
	losses := decimal.Zero

	// Calculate initial average gain and loss
	for i := 1; i <= period; i++ {
		change := prices[i].Sub(prices[i-1])
		if change.GreaterThan(decimal.Zero) {
			gains = gains.Add(change)
		} else {
			losses = losses.Add(change.Abs())
		}
	}

	avgGain := gains.Div(decimal.NewFromInt(int64(period)))
	avgLoss := losses.Div(decimal.NewFromInt(int64(period)))

	// Calculate subsequent values using smoothing
	for i := period + 1; i < len(prices); i++ {
		change := prices[i].Sub(prices[i-1])
		
		if change.GreaterThan(decimal.Zero) {
			avgGain = avgGain.Mul(decimal.NewFromInt(int64(period-1))).Add(change).Div(decimal.NewFromInt(int64(period)))
			avgLoss = avgLoss.Mul(decimal.NewFromInt(int64(period-1))).Div(decimal.NewFromInt(int64(period)))
		} else {
			avgGain = avgGain.Mul(decimal.NewFromInt(int64(period-1))).Div(decimal.NewFromInt(int64(period)))
			avgLoss = avgLoss.Mul(decimal.NewFromInt(int64(period-1))).Add(change.Abs()).Div(decimal.NewFromInt(int64(period)))
		}
	}

	if avgLoss.Equal(decimal.Zero) {
		return decimal.NewFromInt(100)
	}

	rs := avgGain.Div(avgLoss)
	rsi := decimal.NewFromInt(100).Sub(decimal.NewFromInt(100).Div(decimal.NewFromInt(1).Add(rs)))

	return rsi
}

func (ta *TechnicalAnalyzer) calculateEMA(prices []decimal.Decimal, period int) decimal.Decimal {
	if len(prices) < period {
		return decimal.Zero
	}

	// Calculate initial SMA
	sum := decimal.Zero
	for i := 0; i < period; i++ {
		sum = sum.Add(prices[i])
	}
	ema := sum.Div(decimal.NewFromInt(int64(period)))

	// Calculate multiplier
	multiplier := decimal.NewFromInt(2).Div(decimal.NewFromInt(int64(period + 1)))

	// Calculate EMA
	for i := period; i < len(prices); i++ {
		ema = prices[i].Sub(ema).Mul(multiplier).Add(ema)
	}

	return ema
}

func (ta *TechnicalAnalyzer) calculateSMA(prices []decimal.Decimal, period int) decimal.Decimal {
	if len(prices) < period {
		return decimal.Zero
	}

	sum := decimal.Zero
	start := len(prices) - period

	for i := start; i < len(prices); i++ {
		sum = sum.Add(prices[i])
	}

	return sum.Div(decimal.NewFromInt(int64(period)))
}

func (ta *TechnicalAnalyzer) calculateMACDHistory(prices []decimal.Decimal, fastPeriod, slowPeriod int) []decimal.Decimal {
	var macdValues []decimal.Decimal

	if len(prices) < slowPeriod {
		return macdValues
	}

	// Calculate EMAs for each point to get MACD history
	for i := slowPeriod - 1; i < len(prices); i++ {
		if i >= fastPeriod-1 {
			subPrices := prices[:i+1]
			ema12 := ta.calculateEMA(subPrices, fastPeriod)
			ema26 := ta.calculateEMA(subPrices, slowPeriod)
			macd := ema12.Sub(ema26)
			macdValues = append(macdValues, macd)
		}
	}

	return macdValues
}

func (ta *TechnicalAnalyzer) calculateStandardDeviation(prices []decimal.Decimal, period int) decimal.Decimal {
	if len(prices) < period {
		return decimal.Zero
	}

	sma := ta.calculateSMA(prices, period)
	start := len(prices) - period

	sumSquaredDiffs := decimal.Zero
	for i := start; i < len(prices); i++ {
		diff := prices[i].Sub(sma)
		sumSquaredDiffs = sumSquaredDiffs.Add(diff.Mul(diff))
	}

	variance := sumSquaredDiffs.Div(decimal.NewFromInt(int64(period)))
	stdDev, _ := decimal.NewFromString(strconv.FormatFloat(math.Sqrt(variance.InexactFloat64()), 'f', 8, 64))

	return stdDev
}

func (ta *TechnicalAnalyzer) calculateStochastic(highs, lows, closes []decimal.Decimal, kPeriod, dPeriod int) (decimal.Decimal, decimal.Decimal) {
	if len(closes) < kPeriod {
		return decimal.Zero, decimal.Zero
	}

	// Calculate %K
	currentClose := closes[len(closes)-1]
	highestHigh := ta.findHighest(highs[len(highs)-kPeriod:], kPeriod)
	lowestLow := ta.findLowest(lows[len(lows)-kPeriod:], kPeriod)

	stochK := decimal.Zero
	if !highestHigh.Equal(lowestLow) {
		stochK = currentClose.Sub(lowestLow).Div(highestHigh.Sub(lowestLow)).Mul(decimal.NewFromInt(100))
	}

	// For %D, we'd need historical %K values, simplified here
	stochD := stochK // Simplified - in practice, this should be SMA of %K

	return stochK, stochD
}

func (ta *TechnicalAnalyzer) calculateWilliamsR(highs, lows, closes []decimal.Decimal, period int) decimal.Decimal {
	if len(closes) < period {
		return decimal.Zero
	}

	currentClose := closes[len(closes)-1]
	highestHigh := ta.findHighest(highs[len(highs)-period:], period)
	lowestLow := ta.findLowest(lows[len(lows)-period:], period)

	if highestHigh.Equal(lowestLow) {
		return decimal.Zero
	}

	williamsR := highestHigh.Sub(currentClose).Div(highestHigh.Sub(lowestLow)).Mul(decimal.NewFromInt(-100))
	return williamsR
}

func (ta *TechnicalAnalyzer) findHighest(prices []decimal.Decimal, period int) decimal.Decimal {
	if len(prices) == 0 {
		return decimal.Zero
	}

	highest := prices[0]
	start := len(prices) - period
	if start < 0 {
		start = 0
	}

	for i := start; i < len(prices); i++ {
		if prices[i].GreaterThan(highest) {
			highest = prices[i]
		}
	}

	return highest
}

func (ta *TechnicalAnalyzer) findLowest(prices []decimal.Decimal, period int) decimal.Decimal {
	if len(prices) == 0 {
		return decimal.Zero
	}

	lowest := prices[0]
	start := len(prices) - period
	if start < 0 {
		start = 0
	}

	for i := start; i < len(prices); i++ {
		if prices[i].LessThan(lowest) {
			lowest = prices[i]
		}
	}

	return lowest
}
