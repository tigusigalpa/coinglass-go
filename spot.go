package coinglass

import (
	"context"
)

// SpotService provides access to all Coinglass spot market endpoints.
type SpotService struct {
	client *Client
}

// SupportedCoins returns the list of coins supported by Coinglass spot markets.
func (s *SpotService) SupportedCoins(ctx context.Context) ([]string, error) {
	var out []string
	err := s.client.get(ctx, "/api/spot/supported-coins", nil, &out)
	return out, err
}

// SpotCoinMarket represents a single spot coin market snapshot.
type SpotCoinMarket struct {
	Symbol         string  `json:"symbol"`
	Price          float64 `json:"price"`
	PriceChange24h float64 `json:"priceChange24h"`
	VolumeUsd24h   float64 `json:"volumeUsd24h"`
}

// SpotCoinsMarketsParams holds optional parameters for CoinsMarkets.
type SpotCoinsMarketsParams struct {
	Symbol   *string `url:"symbol,omitempty"`
	Exchange *string `url:"exchange,omitempty"`
	Limit    *int    `url:"limit,omitempty"`
}

// CoinsMarkets returns spot coin markets.
func (s *SpotService) CoinsMarkets(ctx context.Context, params *SpotCoinsMarketsParams) ([]SpotCoinMarket, error) {
	var out []SpotCoinMarket
	err := s.client.get(ctx, "/api/spot/coins-markets", buildQuery(params), &out)
	return out, err
}

// SpotPairMarket represents a single spot pair market snapshot.
type SpotPairMarket struct {
	Exchange       string  `json:"exchange"`
	Symbol         string  `json:"symbol"`
	Pair           string  `json:"pair"`
	Price          float64 `json:"price"`
	PriceChange24h float64 `json:"priceChange24h"`
	VolumeUsd24h   float64 `json:"volumeUsd24h"`
}

// SpotPairsMarketsParams holds optional parameters for PairsMarkets.
type SpotPairsMarketsParams struct {
	Symbol   *string `url:"symbol,omitempty"`
	Exchange *string `url:"exchange,omitempty"`
	Limit    *int    `url:"limit,omitempty"`
}

// PairsMarkets returns spot pair markets.
func (s *SpotService) PairsMarkets(ctx context.Context, params *SpotPairsMarketsParams) ([]SpotPairMarket, error) {
	var out []SpotPairMarket
	err := s.client.get(ctx, "/api/spot/pairs-markets", buildQuery(params), &out)
	return out, err
}

// PricePoint represents a single OHLC price history point.
type PricePoint struct {
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
	Timestamp int64   `json:"t"`
}

// SpotPriceHistoryParams holds parameters for PriceHistory.
type SpotPriceHistoryParams struct {
	Symbol    string `url:"symbol"`
	Interval  string `url:"interval"`
	Limit     *int   `url:"limit,omitempty"`
	StartTime *int64 `url:"startTime,omitempty"`
	EndTime   *int64 `url:"endTime,omitempty"`
}

// PriceHistory returns spot price OHLC history.
func (s *SpotService) PriceHistory(ctx context.Context, params *SpotPriceHistoryParams) ([]PricePoint, error) {
	var out []PricePoint
	err := s.client.get(ctx, "/api/spot/price/history", buildQuery(params), &out)
	return out, err
}

// SpotOrderbookHistoryParams holds parameters for Spot OrderbookHistory.
type SpotOrderbookHistoryParams struct {
	Symbol    string `url:"symbol"`
	Exchange  string `url:"exchange"`
	Interval  string `url:"interval"`
	Limit     *int   `url:"limit,omitempty"`
	StartTime *int64 `url:"startTime,omitempty"`
	EndTime   *int64 `url:"endTime,omitempty"`
}

// OrderbookHistory returns the spot orderbook heatmap history.
func (s *SpotService) OrderbookHistory(ctx context.Context, params *SpotOrderbookHistoryParams) ([]OrderbookPoint, error) {
	var out []OrderbookPoint
	err := s.client.get(ctx, "/api/spot/orderbook/history", buildQuery(params), &out)
	return out, err
}

// SpotTakerBuySellHistoryParams holds parameters for Spot TakerBuySellHistory.
type SpotTakerBuySellHistoryParams struct {
	Symbol   string `url:"symbol"`
	Exchange string `url:"exchange"`
	Interval string `url:"interval"`
	Limit    *int   `url:"limit,omitempty"`
}

// TakerBuySellHistory returns spot taker buy/sell volume history.
func (s *SpotService) TakerBuySellHistory(ctx context.Context, params *SpotTakerBuySellHistoryParams) ([]TakerBuySellPoint, error) {
	var out []TakerBuySellPoint
	err := s.client.get(ctx, "/api/spot/taker-buy-sell-volume/history", buildQuery(params), &out)
	return out, err
}
