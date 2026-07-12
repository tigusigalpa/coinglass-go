package websocket

import (
	"encoding/json"
	"fmt"
)

// ChannelLiquidationOrders returns the channel name for the real-time
// liquidation orders stream.
// See https://docs.coinglass.com/reference/ws-liquidation-order.
func ChannelLiquidationOrders() string {
	return "liquidation_orders"
}

// ChannelSpotTrades returns the channel name for the real-time spot trades
// stream for the given exchange/symbol pair, filtered to trades worth at
// least minVolumeUSD.
// See https://docs.coinglass.com/reference/websocket_spot_trades.
func ChannelSpotTrades(exchange, symbol string, minVolumeUSD int64) string {
	return fmt.Sprintf("spot_trades@%s_%s@%d", exchange, symbol, minVolumeUSD)
}

// ChannelFuturesTrades returns the channel name for the real-time futures
// trades stream for the given exchange/symbol pair, filtered to trades worth
// at least minVolumeUSD.
// See https://docs.coinglass.com/reference/websocket_futures_trades.
func ChannelFuturesTrades(exchange, symbol string, minVolumeUSD int64) string {
	return fmt.Sprintf("futures_trades@%s_%s@%d", exchange, symbol, minVolumeUSD)
}

// ChannelFuturesTicker returns the channel name for the real-time futures
// ticker snapshot stream for the given exchange/symbol pair.
// See https://docs.coinglass.com/reference/websocket_futures_ticker.
func ChannelFuturesTicker(exchange, symbol string) string {
	return fmt.Sprintf("futures_ticker@%s_%s", exchange, symbol)
}

// LiquidationSide identifies which side of a position was force-liquidated.
type LiquidationSide int

const (
	// LiquidationSideLong indicates a long position was liquidated.
	LiquidationSideLong LiquidationSide = 1
	// LiquidationSideShort indicates a short position was liquidated.
	LiquidationSideShort LiquidationSide = 2
)

// LiquidationOrder is a single record pushed on the liquidation_orders
// channel.
type LiquidationOrder struct {
	BaseAsset string          `json:"base_asset"`
	Exchange  string          `json:"exchange"`
	Price     float64         `json:"price"`
	Side      LiquidationSide `json:"side"`
	Symbol    string          `json:"symbol"`
	Time      int64           `json:"time"`
	VolumeUSD float64         `json:"volume_usd"`
}

// TradeSide identifies which side of the book a trade executed on.
type TradeSide int

const (
	// TradeSideSell indicates a sell (ask-side) execution.
	TradeSideSell TradeSide = 1
	// TradeSideBuy indicates a buy (bid-side) execution.
	TradeSideBuy TradeSide = 2
)

// Trade is a single record pushed on the spot_trades or futures_trades
// channels.
type Trade struct {
	BaseAsset string    `json:"base_asset"`
	Exchange  string    `json:"exchange"`
	Price     float64   `json:"price"`
	Side      TradeSide `json:"side"`
	Symbol    string    `json:"symbol"`
	Time      int64     `json:"time"`
	VolumeUSD float64   `json:"volume_usd"`
}

// FuturesTicker is a single record pushed on the futures_ticker channel.
type FuturesTicker struct {
	Exchange             string  `json:"exchange"`
	Symbol               string  `json:"symbol"`
	BaseAsset            string  `json:"base_asset"`
	Price                float64 `json:"price"`
	IndexPrice           float64 `json:"index_price"`
	VolumeUSD24h         float64 `json:"volume_usd_24h"`
	OpenInterest         float64 `json:"open_interest"`
	OpenInterestAmount   float64 `json:"open_interest_amount"`
	FundingRate          float64 `json:"funding_rate"`
	NextFundingTime      string  `json:"next_funding_time"`
	FundingIntervalHours int     `json:"funding_interval_hours"`
	ExpiryDate           *string `json:"expiry_date"`
	UpdateTime           string  `json:"update_time"`
}

// DecodeLiquidationOrders unmarshals the Data payload of a Message received
// on the liquidation_orders channel.
func DecodeLiquidationOrders(data json.RawMessage) ([]LiquidationOrder, error) {
	var out []LiquidationOrder
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("coinglass/websocket: failed to decode liquidation orders: %w", err)
	}
	return out, nil
}

// DecodeTrades unmarshals the Data payload of a Message received on a
// spot_trades or futures_trades channel.
func DecodeTrades(data json.RawMessage) ([]Trade, error) {
	var out []Trade
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("coinglass/websocket: failed to decode trades: %w", err)
	}
	return out, nil
}

// DecodeFuturesTicker unmarshals the Data payload of a Message received on a
// futures_ticker channel.
func DecodeFuturesTicker(data json.RawMessage) ([]FuturesTicker, error) {
	var out []FuturesTicker
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("coinglass/websocket: failed to decode futures ticker: %w", err)
	}
	return out, nil
}
