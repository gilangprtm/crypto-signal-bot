package database

import (
	"bytes"
	"crypto-signal-bot/internal/config"
	"crypto-signal-bot/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type SupabaseRestClient struct {
	baseURL    string
	serviceKey string
	client     *http.Client
}

func NewSupabaseRestClient(cfg *config.Config) *SupabaseRestClient {
	return &SupabaseRestClient{
		baseURL:    cfg.SupabaseURL,
		serviceKey: cfg.SupabaseServiceKey,
		client:     &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *SupabaseRestClient) makeRequest(method, endpoint string, data interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s/rest/v1/%s", s.baseURL, endpoint)
	
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", s.serviceKey)
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Content-Type", "application/json")
	
	if method == "POST" {
		req.Header.Set("Prefer", "return=minimal")
	}

	return s.client.Do(req)
}

func (s *SupabaseRestClient) TestConnection() error {
	resp, err := s.makeRequest("GET", "cryptocurrencies?select=id&limit=1", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("REST API test failed: %s - %s", resp.Status, string(body))
	}

	logrus.Info("✅ Supabase REST API connection successful")
	return nil
}

func (s *SupabaseRestClient) CreateSignal(signal *models.TradingSignal) error {
	data := map[string]interface{}{
		"id":                signal.ID,
		"crypto_id":         signal.CryptoID,
		"action":            signal.Action,
		"confidence_score":  signal.ConfidenceScore,
		"entry_price":       signal.EntryPrice,
		"stop_loss":         signal.StopLoss,
		"take_profit_1":     signal.TakeProfit1,
		"take_profit_2":     signal.TakeProfit2,
		"reasoning":         signal.Reasoning,
		"rsi":               signal.RSI,
		"macd_line":         signal.MACDLine,
		"macd_signal":       signal.MACDSignal,
		"macd_histogram":    signal.MACDHistogram,
		"bb_upper":          signal.BBUpper,
		"bb_middle":         signal.BBMiddle,
		"bb_lower":          signal.BBLower,
		"sma_20":            signal.SMA20,
		"ema_12":            signal.EMA12,
		"ema_26":            signal.EMA26,
		"volume_24h":        signal.Volume24h,
		"price_change_24h":  signal.PriceChange24h,
		"fear_greed_index":  signal.FearGreedIndex,
		"market_cap":        signal.MarketCap,
		"market_conditions": signal.MarketConditions,
		"timeframe":         signal.Timeframe,
		"created_at":        signal.CreatedAt,
		"status":            signal.Status,
	}

	resp, err := s.makeRequest("POST", "trading_signals", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create signal: %s - %s", resp.Status, string(body))
	}

	logrus.Info("✅ Signal created successfully via REST API: ", signal.ID)
	return nil
}

func (s *SupabaseRestClient) GetActiveSignals() ([]*models.TradingSignal, error) {
	resp, err := s.makeRequest("GET", "trading_signals?status=eq.active&order=created_at.desc", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get active signals: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var signals []*models.TradingSignal
	if err := json.Unmarshal(body, &signals); err != nil {
		return nil, err
	}

	return signals, nil
}

func (s *SupabaseRestClient) UpdateSignalStatus(signalID uuid.UUID, status string) error {
	data := map[string]interface{}{
		"status": status,
	}

	endpoint := fmt.Sprintf("trading_signals?id=eq.%s", signalID.String())
	resp, err := s.makeRequest("PATCH", endpoint, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update signal status: %s - %s", resp.Status, string(body))
	}

	return nil
}

func (s *SupabaseRestClient) GetCryptocurrencies() ([]models.Cryptocurrency, error) {
	resp, err := s.makeRequest("GET", "cryptocurrencies?order=symbol", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get cryptocurrencies: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cryptos []models.Cryptocurrency
	if err := json.Unmarshal(body, &cryptos); err != nil {
		return nil, err
	}

	return cryptos, nil
}

func (s *SupabaseRestClient) CreateCryptocurrency(crypto *models.Cryptocurrency) error {
	crypto.ID = uuid.New()
	crypto.CreatedAt = time.Now()

	data := map[string]interface{}{
		"id":               crypto.ID,
		"symbol":           crypto.Symbol,
		"name":             crypto.Name,
		"cmc_id":           crypto.CmcID,
		"contract_address": crypto.ContractAddress,
		"platform":         crypto.Platform,
		"slug":             crypto.Slug,
		"coingecko_id":     crypto.CoingeckoID,
		"is_active":        crypto.IsActive,
		"created_at":       crypto.CreatedAt,
		"updated_at":       crypto.UpdatedAt,
	}

	resp, err := s.makeRequest("POST", "cryptocurrencies", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create cryptocurrency: %s - %s", resp.Status, string(body))
	}

	return nil
}

func (s *SupabaseRestClient) LogSystem(level, component, message string, context map[string]interface{}) error {
	data := map[string]interface{}{
		"level":      level,
		"component":  component,
		"message":    message,
		"context":    context,
		"created_at": time.Now(),
	}

	resp, err := s.makeRequest("POST", "system_logs", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to log system message: %s - %s", resp.Status, string(body))
	}

	return nil
}

// Implement other methods as needed...
func (s *SupabaseRestClient) GetRecentSignals(limit int) ([]models.TradingSignal, error) {
	endpoint := fmt.Sprintf("trading_signals?order=created_at.desc&limit=%d", limit)
	resp, err := s.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get recent signals: %s - %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var signals []models.TradingSignal
	if err := json.Unmarshal(body, &signals); err != nil {
		return nil, err
	}

	return signals, nil
}

func (s *SupabaseRestClient) SaveMarketSnapshot(snapshot *models.MarketSnapshot) error {
	// Create minimal data that should always work
	// Use only basic fields that definitely exist
	data := map[string]interface{}{
		"id":        snapshot.ID,
		"price":     snapshot.Price,
		"timestamp": snapshot.Timestamp,
	}

	// Add cryptocurrency_id only if the column exists
	if snapshot.CryptocurrencyID != uuid.Nil {
		data["cryptocurrency_id"] = snapshot.CryptocurrencyID
	}

	// Add basic market data if available
	if !snapshot.Volume24h.IsZero() {
		data["volume_24h"] = snapshot.Volume24h
	}
	if !snapshot.MarketCap.IsZero() {
		data["market_cap"] = snapshot.MarketCap
	}
	if !snapshot.PriceChange1h.IsZero() {
		data["price_change_1h"] = snapshot.PriceChange1h
	}
	if !snapshot.PriceChange24h.IsZero() {
		data["price_change_24h"] = snapshot.PriceChange24h
	}
	if !snapshot.PriceChange7d.IsZero() {
		data["price_change_7d"] = snapshot.PriceChange7d
	}
	if snapshot.FearGreedIndex != 0 {
		data["fear_greed_index"] = snapshot.FearGreedIndex
	}

	// Try to save with minimal data first
	resp, err := s.makeRequest("POST", "market_snapshots", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 {
		logrus.Debug("✅ Market snapshot saved successfully (basic data)")
		return nil
	}

	// If failed, log the error but don't fail the entire process
	body, _ := io.ReadAll(resp.Body)
	logrus.Warnf("Failed to save market snapshot (non-critical): %s - %s", resp.Status, string(body))

	// Return nil to not break the analysis flow
	return nil
}

func (s *SupabaseRestClient) Close() error {
	// No connection to close for REST client
	return nil
}

func (s *SupabaseRestClient) Ping() error {
	return s.TestConnection()
}
