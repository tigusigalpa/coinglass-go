package coinglass

import (
	"context"
	"encoding/json"
)

// IndicatorsService provides access to Coinglass market-indicator endpoints.
type IndicatorsService struct {
	client *Client
}

// FearGreedPoint represents a single Fear & Greed Index history point.
type FearGreedPoint struct {
	Value          int    `json:"value"`
	Classification string `json:"classification"`
	Timestamp      int64  `json:"t"`
}

// FearGreedParams holds optional parameters for FearGreedHistory.
type FearGreedParams struct {
	Limit *int `url:"limit,omitempty"`
}

// FearGreedHistory returns the Fear & Greed Index history.
func (s *IndicatorsService) FearGreedHistory(ctx context.Context, params *FearGreedParams) ([]FearGreedPoint, error) {
	var out []FearGreedPoint
	err := s.client.get(ctx, "/api/index/fear-greed-history", buildQuery(params), &out)
	return out, err
}

// RSIItem represents a single RSI list entry.
type RSIItem struct {
	Symbol    string  `json:"symbol"`
	Interval  string  `json:"interval"`
	RSI       float64 `json:"rsi"`
	Timestamp int64   `json:"t"`
}

// RSIListParams holds parameters for RSIList.
type RSIListParams struct {
	Symbol   *string `url:"symbol,omitempty"`
	Interval *string `url:"interval,omitempty"`
	Limit    *int    `url:"limit,omitempty"`
}

// RSIList returns the futures RSI list.
func (s *IndicatorsService) RSIList(ctx context.Context, params *RSIListParams) ([]RSIItem, error) {
	var out []RSIItem
	err := s.client.get(ctx, "/api/futures/rsi/list", buildQuery(params), &out)
	return out, err
}

// BasisPoint represents a single futures basis history point.
type BasisPoint struct {
	Basis     float64 `json:"basis"`
	Timestamp int64   `json:"t"`
}

// BasisHistoryParams holds parameters for BasisHistory.
type BasisHistoryParams struct {
	Symbol    string `url:"symbol"`
	Interval  string `url:"interval"`
	Limit     *int   `url:"limit,omitempty"`
	StartTime *int64 `url:"startTime,omitempty"`
	EndTime   *int64 `url:"endTime,omitempty"`
}

// BasisHistory returns the futures basis history.
func (s *IndicatorsService) BasisHistory(ctx context.Context, params *BasisHistoryParams) ([]BasisPoint, error) {
	var out []BasisPoint
	err := s.client.get(ctx, "/api/futures/basis/history", buildQuery(params), &out)
	return out, err
}

// PremiumPoint represents a single Coinbase premium index point.
type PremiumPoint struct {
	Premium   float64 `json:"premium"`
	Timestamp int64   `json:"t"`
}

// CoinbasePremiumParams holds optional parameters for CoinbasePremium.
type CoinbasePremiumParams struct {
	Limit     *int   `url:"limit,omitempty"`
	StartTime *int64 `url:"startTime,omitempty"`
	EndTime   *int64 `url:"endTime,omitempty"`
}

// CoinbasePremium returns the Coinbase premium index history.
func (s *IndicatorsService) CoinbasePremium(ctx context.Context, params *CoinbasePremiumParams) ([]PremiumPoint, error) {
	var out []PremiumPoint
	err := s.client.get(ctx, "/api/coinbase-premium-index", buildQuery(params), &out)
	return out, err
}

// RainbowChart represents the Bitcoin rainbow chart response.
type RainbowChart struct {
	Data json.RawMessage `json:"data"`
	Raw  json.RawMessage `json:"-"`
}

// BitcoinRainbowChart returns the Bitcoin rainbow chart data.
func (s *IndicatorsService) BitcoinRainbowChart(ctx context.Context) (*RainbowChart, error) {
	var out RainbowChart
	err := s.client.get(ctx, "/api/index/bitcoin/rainbow-chart", nil, &out)
	return &out, err
}

// StockToFlow represents the stock-to-flow model response.
type StockToFlow struct {
	Data json.RawMessage `json:"data"`
	Raw  json.RawMessage `json:"-"`
}

// StockToFlow returns the stock-to-flow model data.
func (s *IndicatorsService) StockToFlow(ctx context.Context) (*StockToFlow, error) {
	var out StockToFlow
	err := s.client.get(ctx, "/api/index/stock-flow", nil, &out)
	return &out, err
}

// StablecoinPoint represents a single stablecoin market-cap history point.
type StablecoinPoint struct {
	MarketCap float64 `json:"marketCap"`
	Timestamp int64   `json:"t"`
}

// StablecoinMarketCapParams holds optional parameters for StablecoinMarketCap.
type StablecoinMarketCapParams struct {
	Limit     *int   `url:"limit,omitempty"`
	StartTime *int64 `url:"startTime,omitempty"`
	EndTime   *int64 `url:"endTime,omitempty"`
}

// StablecoinMarketCap returns the stablecoin market-cap history.
func (s *IndicatorsService) StablecoinMarketCap(ctx context.Context, params *StablecoinMarketCapParams) ([]StablecoinPoint, error) {
	var out []StablecoinPoint
	err := s.client.get(ctx, "/api/index/stableCoin-marketCap-history", buildQuery(params), &out)
	return out, err
}
