package services

import (
	"crypto-signal-bot/internal/config"
	"crypto-signal-bot/internal/models"
	"crypto-signal-bot/internal/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type CoinMarketCapService struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// Use existing structs from data_collector.go to avoid duplication

func NewCoinMarketCapService(cfg *config.Config) *CoinMarketCapService {
	return &CoinMarketCapService{
		apiKey:  cfg.CoinMarketCapAPIKey,
		baseURL: "https://pro-api.coinmarketcap.com/v1",
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *CoinMarketCapService) GetCryptocurrencyBySymbol(symbol string) (*models.Cryptocurrency, error) {
	url := fmt.Sprintf("%s/cryptocurrency/quotes/latest?symbol=%s", c.baseURL, strings.ToUpper(symbol))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-CMC_PRO_API_KEY", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var cmcResp CMCQuoteResponse
	if err := json.Unmarshal(body, &cmcResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if cmcResp.Status.ErrorCode != 0 {
		return nil, fmt.Errorf("CMC API error: %s", cmcResp.Status.ErrorMessage)
	}

	// Get the first (and should be only) result
	for _, coinData := range cmcResp.Data {
		crypto := &models.Cryptocurrency{
			Symbol:   coinData.Symbol,
			Name:     coinData.Name,
			CmcID:    utils.IntPtr(coinData.ID),
			Slug:     utils.StringPtr(coinData.Slug),
			IsActive: coinData.IsActive == 1,
		}

		return crypto, nil
	}

	return nil, fmt.Errorf("no data found for symbol %s", symbol)
}

func (c *CoinMarketCapService) GetMarketData(symbol string) (*models.MarketSnapshot, error) {
	url := fmt.Sprintf("%s/cryptocurrency/quotes/latest?symbol=%s", c.baseURL, strings.ToUpper(symbol))
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-CMC_PRO_API_KEY", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var cmcResp CMCQuoteResponse
	if err := json.Unmarshal(body, &cmcResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if cmcResp.Status.ErrorCode != 0 {
		return nil, fmt.Errorf("CMC API error: %s", cmcResp.Status.ErrorMessage)
	}

	// Get the first (and should be only) result
	for _, coinData := range cmcResp.Data {
		usdQuote := coinData.Quote["USD"]

		// Create a simple market snapshot with basic data
		snapshot := &models.MarketSnapshot{
			Price:          decimal.NewFromFloat(usdQuote.Price),
			Volume24h:      decimal.NewFromFloat(usdQuote.Volume24h),
			MarketCap:      decimal.NewFromFloat(usdQuote.MarketCap),
			PriceChange1h:  decimal.NewFromFloat(usdQuote.PercentChange1h),
			PriceChange24h: decimal.NewFromFloat(usdQuote.PercentChange24h),
			PriceChange7d:  decimal.NewFromFloat(usdQuote.PercentChange7d),
			Timestamp:      time.Now(),
		}

		logrus.Debugf("CMC Market data for %s: Price=$%.2f, Volume24h=$%.0f, MarketCap=$%.0f",
			symbol, usdQuote.Price, usdQuote.Volume24h, usdQuote.MarketCap)

		return snapshot, nil
	}

	return nil, fmt.Errorf("no market data found for symbol %s", symbol)
}

// GetTopCryptocurrencies retrieves top cryptocurrencies by market cap
func (c *CoinMarketCapService) GetTopCryptocurrencies(limit int) ([]*models.Cryptocurrency, error) {
	url := fmt.Sprintf("%s/cryptocurrency/listings/latest?limit=%d", c.baseURL, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-CMC_PRO_API_KEY", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response struct {
		Status struct {
			ErrorCode    int    `json:"error_code"`
			ErrorMessage string `json:"error_message"`
		} `json:"status"`
		Data []CMCCurrency `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if response.Status.ErrorCode != 0 {
		return nil, fmt.Errorf("CMC API error: %s", response.Status.ErrorMessage)
	}

	var cryptos []*models.Cryptocurrency
	for _, coinData := range response.Data {
		crypto := &models.Cryptocurrency{
			Symbol:   coinData.Symbol,
			Name:     coinData.Name,
			CmcID:    utils.IntPtr(coinData.ID),
			Slug:     utils.StringPtr(coinData.Slug),
			IsActive: coinData.IsActive == 1,
		}

		cryptos = append(cryptos, crypto)
	}

	logrus.Infof("Retrieved %d top cryptocurrencies from CoinMarketCap", len(cryptos))
	return cryptos, nil
}
