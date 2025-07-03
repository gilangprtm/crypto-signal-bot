package services

import (
	"crypto-signal-bot/internal/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type DataCollector struct {
	cfg        *config.Config
	httpClient *http.Client
}

type BinanceKlineData struct {
	Symbol   string `json:"symbol"`
	Interval string `json:"interval"`
	Data     [][]interface{} `json:"data"`
}

type BinanceTicker struct {
	Symbol             string `json:"symbol"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	WeightedAvgPrice   string `json:"weightedAvgPrice"`
	PrevClosePrice     string `json:"prevClosePrice"`
	LastPrice          string `json:"lastPrice"`
	LastQty            string `json:"lastQty"`
	BidPrice           string `json:"bidPrice"`
	AskPrice           string `json:"askPrice"`
	OpenPrice          string `json:"openPrice"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
	OpenTime           int64  `json:"openTime"`
	CloseTime          int64  `json:"closeTime"`
	Count              int    `json:"count"`
}

// CoinMarketCap API structures
type CMCQuoteResponse struct {
	Status struct {
		Timestamp    string `json:"timestamp"`
		ErrorCode    int    `json:"error_code"`
		ErrorMessage string `json:"error_message"`
		Elapsed      int    `json:"elapsed"`
		CreditCount  int    `json:"credit_count"`
	} `json:"status"`
	Data map[string]CMCCurrency `json:"data"`
}

type CMCCurrency struct {
	ID                int                    `json:"id"`
	Name              string                 `json:"name"`
	Symbol            string                 `json:"symbol"`
	Slug              string                 `json:"slug"`
	IsActive          int                    `json:"is_active"`
	IsFiat            int                    `json:"is_fiat"`
	CirculatingSupply float64                `json:"circulating_supply"`
	TotalSupply       float64                `json:"total_supply"`
	MaxSupply         float64                `json:"max_supply"`
	DateAdded         string                 `json:"date_added"`
	NumMarketPairs    int                    `json:"num_market_pairs"`
	CMCRank           int                    `json:"cmc_rank"`
	LastUpdated       string                 `json:"last_updated"`
	Quote             map[string]CMCQuoteUSD `json:"quote"`
}

type CMCQuoteUSD struct {
	Price                 float64 `json:"price"`
	Volume24h             float64 `json:"volume_24h"`
	VolumeChange24h       float64 `json:"volume_change_24h"`
	PercentChange1h       float64 `json:"percent_change_1h"`
	PercentChange24h      float64 `json:"percent_change_24h"`
	PercentChange7d       float64 `json:"percent_change_7d"`
	PercentChange30d      float64 `json:"percent_change_30d"`
	PercentChange60d      float64 `json:"percent_change_60d"`
	PercentChange90d      float64 `json:"percent_change_90d"`
	MarketCap             float64 `json:"market_cap"`
	MarketCapDominance    float64 `json:"market_cap_dominance"`
	FullyDilutedMarketCap float64 `json:"fully_diluted_market_cap"`
	LastUpdated           string  `json:"last_updated"`
}

type CoinGeckoPrice struct {
	ID                string  `json:"id"`
	Symbol            string  `json:"symbol"`
	Name              string  `json:"name"`
	CurrentPrice      float64 `json:"current_price"`
	MarketCap         float64 `json:"market_cap"`
	MarketCapRank     int     `json:"market_cap_rank"`
	TotalVolume       float64 `json:"total_volume"`
	PriceChange24h    float64 `json:"price_change_24h"`
	PriceChangePercent24h float64 `json:"price_change_percentage_24h"`
	PriceChangePercent1h  float64 `json:"price_change_percentage_1h_in_currency"`
	PriceChangePercent7d  float64 `json:"price_change_percentage_7d_in_currency"`
}

type FearGreedIndex struct {
	Name      string `json:"name"`
	Data      []struct {
		Value         string `json:"value"`
		ValueClassification string `json:"value_classification"`
		Timestamp     string `json:"timestamp"`
		TimeUntilUpdate string `json:"time_until_update"`
	} `json:"data"`
}

type MarketData struct {
	Symbol           string
	Price            decimal.Decimal
	Volume24h        decimal.Decimal
	MarketCap        decimal.Decimal
	PriceChange1h    decimal.Decimal
	PriceChange24h   decimal.Decimal
	PriceChange7d    decimal.Decimal
	FearGreedIndex   int
	KlineData        [][]interface{} // OHLCV data for technical analysis
	Timestamp        time.Time
}

func NewDataCollector(cfg *config.Config) *DataCollector {
	return &DataCollector{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (dc *DataCollector) GetMarketData(symbol string) (*MarketData, error) {
	logrus.Debug("Fetching market data for: ", symbol)

	// Primary: Get price data from CoinMarketCap (free tier)
	cmcData, err := dc.getCMCData(symbol)
	if err != nil {
		logrus.Warn("Failed to get CoinMarketCap data: ", err)
		// Fallback to Binance if available
		binanceData, binanceErr := dc.getBinanceData(symbol)
		if binanceErr != nil {
			logrus.Error("Failed to get both CMC and Binance data: ", binanceErr)
			return nil, fmt.Errorf("no market data available: CMC error: %v, Binance error: %v", err, binanceErr)
		}
		logrus.Info("Using Binance data as fallback")
		return dc.processMarketDataFromBinance(symbol, binanceData)
	}

	// Get additional market data from CoinGecko (optional)
	coinGeckoData, err := dc.getCoinGeckoData(symbol)
	if err != nil {
		logrus.Warn("Failed to get CoinGecko data, using CMC only: ", err)
		// Continue with CMC data only
	}

	// Get Fear & Greed Index
	fearGreedIndex, err := dc.getFearGreedIndex()
	if err != nil {
		logrus.Warn("Failed to get Fear & Greed Index: ", err)
		fearGreedIndex = 50 // Default neutral value
	}

	// Try to get kline data for technical analysis (fallback to Binance if CMC doesn't provide)
	klineData, err := dc.getBinanceKlines(symbol, "15m", 100)
	if err != nil {
		logrus.Warn("Failed to get kline data from Binance: ", err)
		// For now, we'll continue without kline data
		// In production, you might want to use alternative sources
		klineData = [][]interface{}{}
	}

	// Create market data from CMC
	marketData := &MarketData{
		Symbol:         symbol,
		FearGreedIndex: fearGreedIndex,
		KlineData:      klineData,
		Timestamp:      time.Now(),
	}

	// Parse CMC data
	if usdQuote, exists := cmcData.Quote["USD"]; exists {
		marketData.Price = decimal.NewFromFloat(usdQuote.Price)
		marketData.Volume24h = decimal.NewFromFloat(usdQuote.Volume24h)
		marketData.MarketCap = decimal.NewFromFloat(usdQuote.MarketCap)
		marketData.PriceChange1h = decimal.NewFromFloat(usdQuote.PercentChange1h)
		marketData.PriceChange24h = decimal.NewFromFloat(usdQuote.PercentChange24h)
		marketData.PriceChange7d = decimal.NewFromFloat(usdQuote.PercentChange7d)
	}

	// Parse CoinGecko data (if available)
	if coinGeckoData != nil {
		marketData.MarketCap = decimal.NewFromFloat(coinGeckoData.MarketCap)
		marketData.PriceChange1h = decimal.NewFromFloat(coinGeckoData.PriceChangePercent1h)
		marketData.PriceChange7d = decimal.NewFromFloat(coinGeckoData.PriceChangePercent7d)
		
		// Use CoinGecko price if more accurate
		if coinGeckoData.CurrentPrice > 0 {
			marketData.Price = decimal.NewFromFloat(coinGeckoData.CurrentPrice)
		}
	}

	logrus.Debug("Market data collected successfully for: ", symbol)
	return marketData, nil
}

func (dc *DataCollector) getBinanceData(symbol string) (*BinanceTicker, error) {
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/24hr?symbol=%sUSDT", symbol)
	
	resp, err := dc.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("binance API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ticker BinanceTicker
	if err := json.Unmarshal(body, &ticker); err != nil {
		return nil, err
	}

	return &ticker, nil
}

func (dc *DataCollector) getCoinGeckoData(symbol string) (*CoinGeckoPrice, error) {
	// Map common symbols to CoinGecko IDs
	coinGeckoIDs := map[string]string{
		"BTC":   "bitcoin",
		"ETH":   "ethereum",
		"BNB":   "binancecoin",
		"ADA":   "cardano",
		"SOL":   "solana",
		"DOT":   "polkadot",
		"MATIC": "matic-network",
		"AVAX":  "avalanche-2",
		"LINK":  "chainlink",
		"ATOM":  "cosmos",
	}

	coinID, exists := coinGeckoIDs[symbol]
	if !exists {
		return nil, fmt.Errorf("unsupported symbol for CoinGecko: %s", symbol)
	}

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&ids=%s&order=market_cap_desc&per_page=1&page=1&sparkline=false&price_change_percentage=1h,24h,7d", coinID)
	
	if dc.cfg.CoinGeckoAPIKey != "" {
		url += "&x_cg_demo_api_key=" + dc.cfg.CoinGeckoAPIKey
	}

	resp, err := dc.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("coingecko API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var prices []CoinGeckoPrice
	if err := json.Unmarshal(body, &prices); err != nil {
		return nil, err
	}

	if len(prices) == 0 {
		return nil, fmt.Errorf("no data found for symbol: %s", symbol)
	}

	return &prices[0], nil
}

func (dc *DataCollector) getFearGreedIndex() (int, error) {
	url := "https://api.alternative.me/fng/"
	
	resp, err := dc.httpClient.Get(url)
	if err != nil {
		return 50, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 50, fmt.Errorf("fear & greed API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 50, err
	}

	var fgi FearGreedIndex
	if err := json.Unmarshal(body, &fgi); err != nil {
		return 50, err
	}

	if len(fgi.Data) == 0 {
		return 50, fmt.Errorf("no fear & greed data available")
	}

	value, err := strconv.Atoi(fgi.Data[0].Value)
	if err != nil {
		return 50, err
	}

	return value, nil
}

func (dc *DataCollector) getBinanceKlines(symbol, interval string, limit int) ([][]interface{}, error) {
	url := fmt.Sprintf("https://api.binance.com/api/v3/klines?symbol=%sUSDT&interval=%s&limit=%d", symbol, interval, limit)
	
	resp, err := dc.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("binance klines API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var klines [][]interface{}
	if err := json.Unmarshal(body, &klines); err != nil {
		return nil, err
	}

	return klines, nil
}

func (dc *DataCollector) GetMultipleMarketData(symbols []string) (map[string]*MarketData, error) {
	logrus.Info("Fetching market data for multiple symbols: ", symbols)
	
	results := make(map[string]*MarketData)
	
	for _, symbol := range symbols {
		data, err := dc.GetMarketData(symbol)
		if err != nil {
			logrus.Error("Failed to get market data for ", symbol, ": ", err)
			continue
		}
		results[symbol] = data
		
		// Rate limiting - be nice to APIs
		time.Sleep(100 * time.Millisecond)
	}
	
	logrus.Info("Successfully collected market data for ", len(results), " symbols")
	return results, nil
}

// getCMCData fetches cryptocurrency data from CoinMarketCap API
func (dc *DataCollector) getCMCData(symbol string) (*CMCCurrency, error) {
	if dc.cfg.CoinMarketCapAPIKey == "" {
		return nil, fmt.Errorf("CoinMarketCap API key not configured")
	}

	// CMC API endpoint for quotes
	url := fmt.Sprintf("https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol=%s&convert=USD", symbol)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add required headers
	req.Header.Set("X-CMC_PRO_API_KEY", dc.cfg.CoinMarketCapAPIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := dc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("CMC API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var cmcResponse CMCQuoteResponse
	if err := json.Unmarshal(body, &cmcResponse); err != nil {
		return nil, fmt.Errorf("failed to parse CMC response: %w", err)
	}

	// Check for API errors
	if cmcResponse.Status.ErrorCode != 0 {
		return nil, fmt.Errorf("CMC API error: %s", cmcResponse.Status.ErrorMessage)
	}

	// Get currency data
	currency, exists := cmcResponse.Data[symbol]
	if !exists {
		return nil, fmt.Errorf("symbol %s not found in CMC response", symbol)
	}

	return &currency, nil
}

// processMarketDataFromBinance processes market data when using Binance as fallback
func (dc *DataCollector) processMarketDataFromBinance(symbol string, binanceData *BinanceTicker) (*MarketData, error) {
	// Get Fear & Greed Index
	fearGreedIndex, err := dc.getFearGreedIndex()
	if err != nil {
		logrus.Warn("Failed to get Fear & Greed Index: ", err)
		fearGreedIndex = 50 // Default neutral value
	}

	// Get kline data for technical analysis
	klineData, err := dc.getBinanceKlines(symbol, "15m", 100)
	if err != nil {
		logrus.Error("Failed to get kline data: ", err)
		return nil, err
	}

	// Create market data
	marketData := &MarketData{
		Symbol:         symbol,
		FearGreedIndex: fearGreedIndex,
		KlineData:      klineData,
		Timestamp:      time.Now(),
	}

	// Parse Binance data
	if price, err := decimal.NewFromString(binanceData.LastPrice); err == nil {
		marketData.Price = price
	}
	if volume, err := decimal.NewFromString(binanceData.Volume); err == nil {
		marketData.Volume24h = volume
	}
	if change, err := decimal.NewFromString(binanceData.PriceChangePercent); err == nil {
		marketData.PriceChange24h = change
	}

	return marketData, nil
}
