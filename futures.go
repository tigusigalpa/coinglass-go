package coinglass

import (
	"context"
	"encoding/json"
	"strconv"
)

// FuturesService provides access to all Coinglass futures endpoints.
type FuturesService struct {
	client *Client
}

// SupportedCoins returns the list of coins supported by Coinglass futures.
func (s *FuturesService) SupportedCoins(ctx context.Context) ([]string, error) {
	var out []string
	err := s.client.get(ctx, "/futures/supported-coins", nil, &out)
	return out, err
}

// ExchangePair describes a single futures exchange + symbol pair.
type ExchangePair struct {
	Exchange string `json:"exchange"`
	Symbol   string `json:"symbol"`
	Pair     string `json:"pair"`
}

// SupportedExchangePairsParams holds the optional parameters for SupportedExchangePairs.
type SupportedExchangePairsParams struct {
	Exchange *string `url:"exchange,omitempty"`
}

// SupportedExchangePairs returns the supported futures exchange pairs,
// optionally filtered by a single exchange.
func (s *FuturesService) SupportedExchangePairs(ctx context.Context, params *SupportedExchangePairsParams) (map[string][]ExchangePair, error) {
	var out map[string][]ExchangePair
	err := s.client.get(ctx, "/api/futures/supported-exchange-pairs", buildQuery(params), &out)
	return out, err
}

// CoinMarket represents a single futures coin market snapshot.
type CoinMarket struct {
	Symbol            string  `json:"symbol"`
	Price             float64 `json:"price"`
	PriceChange1h     float64 `json:"priceChange1h"`
	PriceChange24h    float64 `json:"priceChange24h"`
	VolumeUsd24h      float64 `json:"volumeUsd24h"`
	OpenInterestUsd   float64 `json:"openInterestUsd"`
	FundingRate       float64 `json:"fundingRate"`
	TurnoverNumber24h float64 `json:"turnoverNumber24h"`
}

// CoinsMarketsParams holds the optional parameters for CoinsMarkets.
type CoinsMarketsParams struct {
	Symbol    *string  `url:"symbol,omitempty"`
	Exchanges []string `url:"exchanges,omitempty"`
	Limit     *int     `url:"limit,omitempty"`
}

// CoinsMarkets returns futures coin markets.
func (s *FuturesService) CoinsMarkets(ctx context.Context, params *CoinsMarketsParams) ([]CoinMarket, error) {
	var out []CoinMarket
	err := s.client.get(ctx, "/api/futures/coins-markets", buildQuery(params), &out)
	return out, err
}

// PairMarket represents a single futures trading pair market snapshot.
type PairMarket struct {
	Exchange        string  `json:"exchange"`
	Symbol          string  `json:"symbol"`
	Pair            string  `json:"pair"`
	Price           float64 `json:"price"`
	PriceChange24h  float64 `json:"priceChange24h"`
	VolumeUsd24h    float64 `json:"volumeUsd24h"`
	OpenInterestUsd float64 `json:"openInterestUsd"`
}

// PairsMarketsParams holds the optional parameters for PairsMarkets.
type PairsMarketsParams struct {
	Symbol   *string `url:"symbol,omitempty"`
	Exchange *string `url:"exchange,omitempty"`
	Limit    *int    `url:"limit,omitempty"`
}

// PairsMarkets returns futures pair markets.
func (s *FuturesService) PairsMarkets(ctx context.Context, params *PairsMarketsParams) ([]PairMarket, error) {
	var out []PairMarket
	err := s.client.get(ctx, "/api/futures/pairs-markets", buildQuery(params), &out)
	return out, err
}

// PriceChangeItem represents a single entry in the price change list.
type PriceChangeItem struct {
	Symbol    string  `json:"symbol"`
	Change1h  float64 `json:"change1h"`
	Change24h float64 `json:"change24h"`
	Change7d  float64 `json:"change7d"`
}

// PriceChangeList returns the futures price change list.
func (s *FuturesService) PriceChangeList(ctx context.Context) ([]PriceChangeItem, error) {
	var out []PriceChangeItem
	err := s.client.get(ctx, "/futures/price-change-list", nil, &out)
	return out, err
}

// OIHistoryPoint represents a single open interest OHLC point.
type OIHistoryPoint struct {
	OpenInterest    float64 `json:"openInterest"`
	OpenInterestUsd float64 `json:"openInterestUsd"`
	Timestamp       int64   `json:"t"`
}

// OIHistoryParams holds parameters for open-interest history endpoints.
type OIHistoryParams struct {
	Symbol    string `url:"symbol"`
	Interval  string `url:"interval"`
	Limit     *int   `url:"limit,omitempty"`
	StartTime *int64 `url:"startTime,omitempty"`
	EndTime   *int64 `url:"endTime,omitempty"`
}

// OpenInterestHistory returns OHLC open-interest history for a symbol and interval.
func (s *FuturesService) OpenInterestHistory(ctx context.Context, params *OIHistoryParams) ([]OIHistoryPoint, error) {
	var out []OIHistoryPoint
	err := s.client.get(ctx, "/api/futures/openInterest/ohlc-history", buildQuery(params), &out)
	return out, err
}

// OpenInterestAggregatedHistory returns aggregated OHLC open-interest history.
func (s *FuturesService) OpenInterestAggregatedHistory(ctx context.Context, params *OIHistoryParams) ([]OIHistoryPoint, error) {
	var out []OIHistoryPoint
	err := s.client.get(ctx, "/api/futures/openInterest/ohlc-aggregated-history", buildQuery(params), &out)
	return out, err
}

// OIExchangeItem represents open interest broken down by exchange.
type OIExchangeItem struct {
	Exchange        string  `json:"exchange"`
	OpenInterest    float64 `json:"openInterest"`
	OpenInterestUsd float64 `json:"openInterestUsd"`
	Timestamp       int64   `json:"t"`
}

// OIExchangeListParams holds parameters for OpenInterestExchangeList.
type OIExchangeListParams struct {
	Symbol    string  `url:"symbol"`
	Interval  string  `url:"interval"`
	Limit     *int    `url:"limit,omitempty"`
	StartTime *int64  `url:"startTime,omitempty"`
	EndTime   *int64  `url:"endTime,omitempty"`
	Exchange  *string `url:"exchange,omitempty"`
}

// OpenInterestExchangeList returns open-interest history grouped by exchange.
func (s *FuturesService) OpenInterestExchangeList(ctx context.Context, params *OIExchangeListParams) ([]OIExchangeItem, error) {
	var out []OIExchangeItem
	err := s.client.get(ctx, "/api/futures/openInterest/exchange-list", buildQuery(params), &out)
	return out, err
}

// FundingRatePoint represents a single funding-rate OHLC point.
type FundingRatePoint struct {
	FundingRate              float64 `json:"fundingRate"`
	FundingRateAnnualPercent float64 `json:"fundingRateAnnualPercent"`
	Timestamp                int64   `json:"t"`
}

// FundingRateHistoryParams holds parameters for FundingRateHistory.
type FundingRateHistoryParams struct {
	Symbol    string `url:"symbol"`
	Interval  string `url:"interval"`
	Limit     *int   `url:"limit,omitempty"`
	StartTime *int64 `url:"startTime,omitempty"`
	EndTime   *int64 `url:"endTime,omitempty"`
}

// FundingRateHistory returns funding-rate OHLC history.
func (s *FuturesService) FundingRateHistory(ctx context.Context, params *FundingRateHistoryParams) ([]FundingRatePoint, error) {
	var out []FundingRatePoint
	err := s.client.get(ctx, "/api/futures/fundingRate/ohlc-history", buildQuery(params), &out)
	return out, err
}

// FundingRateExchange represents funding-rate data for a specific exchange.
type FundingRateExchange struct {
	Exchange    string  `json:"exchange"`
	FundingRate float64 `json:"fundingRate"`
	Timestamp   int64   `json:"t"`
}

// FundingRateExchangeListParams holds parameters for FundingRateExchangeList.
type FundingRateExchangeListParams struct {
	Symbol    string  `url:"symbol"`
	Interval  string  `url:"interval"`
	Limit     *int    `url:"limit,omitempty"`
	StartTime *int64  `url:"startTime,omitempty"`
	EndTime   *int64  `url:"endTime,omitempty"`
	Exchange  *string `url:"exchange,omitempty"`
}

// FundingRateExchangeList returns funding-rate history grouped by exchange.
func (s *FuturesService) FundingRateExchangeList(ctx context.Context, params *FundingRateExchangeListParams) ([]FundingRateExchange, error) {
	var out []FundingRateExchange
	err := s.client.get(ctx, "/api/futures/fundingRate/exchange-list", buildQuery(params), &out)
	return out, err
}

// FundingRateOiWeighted returns OI-weighted funding-rate OHLC history.
func (s *FuturesService) FundingRateOiWeighted(ctx context.Context, params *FundingRateHistoryParams) ([]FundingRatePoint, error) {
	var out []FundingRatePoint
	err := s.client.get(ctx, "/api/futures/fundingRate/oi-weight-ohlc-history", buildQuery(params), &out)
	return out, err
}

// ArbitrageItem represents a single funding-rate arbitrage opportunity.
type ArbitrageItem struct {
	Symbol      string  `json:"symbol"`
	Exchange    string  `json:"exchange"`
	FundingRate float64 `json:"fundingRate"`
	Spread      float64 `json:"spread"`
}

// FundingRateArbitrageParams holds optional parameters for FundingRateArbitrage.
type FundingRateArbitrageParams struct {
	Symbol   *string `url:"symbol,omitempty"`
	Interval *string `url:"interval,omitempty"`
	Limit    *int    `url:"limit,omitempty"`
}

// FundingRateArbitrage returns funding-rate arbitrage opportunities.
func (s *FuturesService) FundingRateArbitrage(ctx context.Context, params *FundingRateArbitrageParams) ([]ArbitrageItem, error) {
	var out []ArbitrageItem
	err := s.client.get(ctx, "/api/futures/fundingRate/arbitrage", buildQuery(params), &out)
	return out, err
}

// LongShortPoint represents a single long/short account ratio point.
type LongShortPoint struct {
	LongAccount  float64 `json:"longAccount"`
	ShortAccount float64 `json:"shortAccount"`
	LongRatio    float64 `json:"longRatio"`
	ShortRatio   float64 `json:"shortRatio"`
	Timestamp    int64   `json:"t"`
}

// LongShortRatioParams holds parameters for long/short ratio history endpoints.
type LongShortRatioParams struct {
	Symbol    string  `url:"symbol"`
	Interval  string  `url:"interval"`
	Limit     *int    `url:"limit,omitempty"`
	StartTime *int64  `url:"startTime,omitempty"`
	EndTime   *int64  `url:"endTime,omitempty"`
	Exchange  *string `url:"exchange,omitempty"`
}

// LongShortRatioHistory returns the global long/short account ratio history.
func (s *FuturesService) LongShortRatioHistory(ctx context.Context, params *LongShortRatioParams) ([]LongShortPoint, error) {
	var out []LongShortPoint
	err := s.client.get(ctx, "/api/futures/global-long-short-account-ratio/history", buildQuery(params), &out)
	return out, err
}

// TopLongShortRatioHistory returns the top trader long/short account ratio history.
func (s *FuturesService) TopLongShortRatioHistory(ctx context.Context, params *LongShortRatioParams) ([]LongShortPoint, error) {
	var out []LongShortPoint
	err := s.client.get(ctx, "/api/futures/top-long-short-account-ratio/history", buildQuery(params), &out)
	return out, err
}

// LiquidationPoint represents a single liquidation history point.
type LiquidationPoint struct {
	BuyQty     float64 `json:"buyQty"`
	SellQty    float64 `json:"sellQty"`
	BuyAmount  float64 `json:"buyAmount"`
	SellAmount float64 `json:"sellAmount"`
	Timestamp  int64   `json:"t"`
}

// LiquidationHistoryParams holds parameters for LiquidationHistory.
type LiquidationHistoryParams struct {
	Symbol    string `url:"symbol"`
	Pair      string `url:"pair"`
	Interval  string `url:"interval"`
	Limit     *int   `url:"limit,omitempty"`
	StartTime *int64 `url:"startTime,omitempty"`
	EndTime   *int64 `url:"endTime,omitempty"`
}

// LiquidationHistory returns pair liquidation history.
func (s *FuturesService) LiquidationHistory(ctx context.Context, params *LiquidationHistoryParams) ([]LiquidationPoint, error) {
	var out []LiquidationPoint
	err := s.client.get(ctx, "/api/futures/liquidation/history", buildQuery(params), &out)
	return out, err
}

// LiquidationAggregatedHistoryParams holds parameters for LiquidationAggregatedHistory.
type LiquidationAggregatedHistoryParams struct {
	Symbol    string `url:"symbol"`
	Interval  string `url:"interval"`
	Limit     *int   `url:"limit,omitempty"`
	StartTime *int64 `url:"startTime,omitempty"`
	EndTime   *int64 `url:"endTime,omitempty"`
}

// LiquidationAggregatedHistory returns aggregated coin liquidation history.
func (s *FuturesService) LiquidationAggregatedHistory(ctx context.Context, params *LiquidationAggregatedHistoryParams) ([]LiquidationPoint, error) {
	var out []LiquidationPoint
	err := s.client.get(ctx, "/api/futures/liquidation/aggregated-history", buildQuery(params), &out)
	return out, err
}

// LiquidationCoin represents a liquidation coin list entry.
type LiquidationCoin struct {
	Symbol         string  `json:"symbol"`
	Exchange       string  `json:"exchange"`
	LiquidationUsd float64 `json:"liquidationUsd"`
	Timestamp      int64   `json:"t"`
}

// LiquidationCoinListParams holds optional parameters for LiquidationCoinList.
type LiquidationCoinListParams struct {
	Symbol *string `url:"symbol,omitempty"`
	Limit  *int    `url:"limit,omitempty"`
}

// LiquidationCoinList returns the liquidation coin list.
func (s *FuturesService) LiquidationCoinList(ctx context.Context, params *LiquidationCoinListParams) ([]LiquidationCoin, error) {
	var out []LiquidationCoin
	err := s.client.get(ctx, "/api/futures/liquidation/coin-list", buildQuery(params), &out)
	return out, err
}

// LiquidationHeatmapParams holds parameters for LiquidationHeatmap.
type LiquidationHeatmapParams struct {
	Symbol   string `url:"symbol"`
	Interval string `url:"interval"`
	Limit    *int   `url:"limit,omitempty"`
}

// LiquidationHeatmapPoint is one bucket in a liquidation heatmap.
type LiquidationHeatmapPoint struct {
	Price          float64 `json:"price"`
	LiquidationUsd float64 `json:"liquidationUsd"`
}

// LiquidationHeatmap is the modelled liquidation heatmap response.
type LiquidationHeatmap struct {
	Model int                       `json:"model"`
	Data  []LiquidationHeatmapPoint `json:"data"`
	Raw   json.RawMessage           `json:"-"`
}

// LiquidationHeatmap returns a liquidation heatmap for the requested model (1, 2 or 3).
func (s *FuturesService) LiquidationHeatmap(ctx context.Context, model int, params *LiquidationHeatmapParams) (*LiquidationHeatmap, error) {
	var out LiquidationHeatmap
	path := "/api/futures/liquidation/heatmap/model" + strconv.Itoa(model)
	err := s.client.get(ctx, path, buildQuery(params), &out)
	if err != nil {
		return nil, err
	}
	out.Model = model
	return &out, nil
}

// LiquidationMapParams holds parameters for LiquidationMap.
type LiquidationMapParams struct {
	Symbol   string `url:"symbol"`
	Pair     string `url:"pair"`
	Interval string `url:"interval"`
	Limit    *int   `url:"limit,omitempty"`
}

// LiquidationMap represents the liquidation map response.
type LiquidationMap struct {
	Symbol string          `json:"symbol"`
	Data   json.RawMessage `json:"data"`
	Raw    json.RawMessage `json:"-"`
}

// LiquidationMap returns the liquidation map for a symbol/pair.
func (s *FuturesService) LiquidationMap(ctx context.Context, params *LiquidationMapParams) (*LiquidationMap, error) {
	var out LiquidationMap
	err := s.client.get(ctx, "/api/futures/liquidation/map", buildQuery(params), &out)
	return &out, err
}

// OrderbookPoint represents a single orderbook heatmap point.
type OrderbookPoint struct {
	Price     float64 `json:"price"`
	BidQty    float64 `json:"bidQty"`
	AskQty    float64 `json:"askQty"`
	BidAmount float64 `json:"bidAmount"`
	AskAmount float64 `json:"askAmount"`
	Timestamp int64   `json:"t"`
}

// OrderbookHistoryParams holds parameters for OrderbookHistory.
type OrderbookHistoryParams struct {
	Symbol    string `url:"symbol"`
	Exchange  string `url:"exchange"`
	Interval  string `url:"interval"`
	Limit     *int   `url:"limit,omitempty"`
	StartTime *int64 `url:"startTime,omitempty"`
	EndTime   *int64 `url:"endTime,omitempty"`
}

// OrderbookHistory returns orderbook heatmap history.
func (s *FuturesService) OrderbookHistory(ctx context.Context, params *OrderbookHistoryParams) ([]OrderbookPoint, error) {
	var out []OrderbookPoint
	err := s.client.get(ctx, "/api/futures/orderbook/history", buildQuery(params), &out)
	return out, err
}

// LargeOrder represents a large limit order in the orderbook.
type LargeOrder struct {
	Symbol    string  `json:"symbol"`
	Exchange  string  `json:"exchange"`
	Side      string  `json:"side"`
	Price     float64 `json:"price"`
	Size      float64 `json:"size"`
	ValueUsd  float64 `json:"valueUsd"`
	Timestamp int64   `json:"t"`
}

// LargeOrdersParams holds parameters for LargeOrders.
type LargeOrdersParams struct {
	Symbol   string `url:"symbol"`
	Exchange string `url:"exchange"`
	Interval string `url:"interval"`
	Limit    *int   `url:"limit,omitempty"`
}

// LargeOrders returns large limit orders from the orderbook.
func (s *FuturesService) LargeOrders(ctx context.Context, params *LargeOrdersParams) ([]LargeOrder, error) {
	var out []LargeOrder
	err := s.client.get(ctx, "/api/futures/orderbook/large-limit-order", buildQuery(params), &out)
	return out, err
}

// TakerBuySellPoint represents a single taker buy/sell volume point.
type TakerBuySellPoint struct {
	BuyVolume  float64 `json:"buyVolume"`
	SellVolume float64 `json:"sellVolume"`
	Timestamp  int64   `json:"t"`
}

// TakerBuySellHistoryParams holds parameters for TakerBuySellHistory.
type TakerBuySellHistoryParams struct {
	Symbol   string `url:"symbol"`
	Exchange string `url:"exchange"`
	Interval string `url:"interval"`
	Limit    *int   `url:"limit,omitempty"`
}

// TakerBuySellHistory returns futures taker buy/sell volume history.
func (s *FuturesService) TakerBuySellHistory(ctx context.Context, params *TakerBuySellHistoryParams) ([]TakerBuySellPoint, error) {
	var out []TakerBuySellPoint
	err := s.client.get(ctx, "/api/futures/taker-buy-sell-volume/history", buildQuery(params), &out)
	return out, err
}

// WhaleAlert represents a Hyperliquid whale alert entry.
type WhaleAlert struct {
	Symbol string  `json:"symbol"`
	Side   string  `json:"side"`
	Size   float64 `json:"size"`
	Price  float64 `json:"price"`
	Time   int64   `json:"time"`
	Link   string  `json:"link"`
}

// WhaleAlertParams holds optional parameters for WhaleAlert.
type WhaleAlertParams struct {
	Symbol   *string `url:"symbol,omitempty"`
	Interval *string `url:"interval,omitempty"`
	Limit    *int    `url:"limit,omitempty"`
}

// WhaleAlert returns the Hyperliquid whale alert feed.
func (s *FuturesService) WhaleAlert(ctx context.Context, params *WhaleAlertParams) ([]WhaleAlert, error) {
	var out []WhaleAlert
	err := s.client.get(ctx, "/api/hyperliquid/whale-alert", buildQuery(params), &out)
	return out, err
}
