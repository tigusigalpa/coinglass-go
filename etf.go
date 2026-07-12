package coinglass

import (
	"context"
)

// ETFService provides access to all Coinglass ETF endpoints.
type ETFService struct {
	client *Client
}

// ETFItem represents a single ETF listing entry.
type ETFItem struct {
	Ticker   string  `json:"ticker"`
	Name     string  `json:"name"`
	Holdings float64 `json:"holdings"`
}

// BitcoinList returns the list of Bitcoin ETFs.
func (s *ETFService) BitcoinList(ctx context.Context) ([]ETFItem, error) {
	var out []ETFItem
	err := s.client.get(ctx, "/api/etf/bitcoin/list", nil, &out)
	return out, err
}

// ETFFlowPoint represents a single ETF flow history point.
type ETFFlowPoint struct {
	Timestamp    int64   `json:"t"`
	NetFlow      float64 `json:"netFlow"`
	TotalInflow  float64 `json:"totalInflow"`
	TotalOutflow float64 `json:"totalOutflow"`
}

// ETFFlowParams holds parameters for ETF flow history endpoints.
type ETFFlowParams struct {
	Interval string `url:"interval"`
	Limit    *int   `url:"limit,omitempty"`
}

// BitcoinFlowHistory returns Bitcoin ETF flow history.
func (s *ETFService) BitcoinFlowHistory(ctx context.Context, params *ETFFlowParams) ([]ETFFlowPoint, error) {
	var out []ETFFlowPoint
	err := s.client.get(ctx, "/api/etf/bitcoin/flow-history", buildQuery(params), &out)
	return out, err
}

// ETFNetAssetsPoint represents a single ETF net-assets history point.
type ETFNetAssetsPoint struct {
	Timestamp int64   `json:"t"`
	NetAssets float64 `json:"netAssets"`
}

// ETFNetAssetsParams holds parameters for BitcoinNetAssetsHistory.
type ETFNetAssetsParams struct {
	Interval string `url:"interval"`
	Limit    *int   `url:"limit,omitempty"`
}

// BitcoinNetAssetsHistory returns Bitcoin ETF net assets history.
func (s *ETFService) BitcoinNetAssetsHistory(ctx context.Context, params *ETFNetAssetsParams) ([]ETFNetAssetsPoint, error) {
	var out []ETFNetAssetsPoint
	err := s.client.get(ctx, "/api/etf/bitcoin/net-assets/history", buildQuery(params), &out)
	return out, err
}

// EthereumList returns the list of Ethereum ETFs.
func (s *ETFService) EthereumList(ctx context.Context) ([]ETFItem, error) {
	var out []ETFItem
	err := s.client.get(ctx, "/api/etf/ethereum/list", nil, &out)
	return out, err
}

// EthereumFlowHistory returns Ethereum ETF flow history.
func (s *ETFService) EthereumFlowHistory(ctx context.Context, params *ETFFlowParams) ([]ETFFlowPoint, error) {
	var out []ETFFlowPoint
	err := s.client.get(ctx, "/api/etf/ethereum/flow-history", buildQuery(params), &out)
	return out, err
}

// GrayscaleHolding represents a single Grayscale holdings entry.
type GrayscaleHolding struct {
	Symbol   string  `json:"symbol"`
	Asset    string  `json:"asset"`
	Holdings float64 `json:"holdings"`
	ValueUsd float64 `json:"valueUsd"`
}

// GrayscaleHoldings returns the Grayscale holdings list.
func (s *ETFService) GrayscaleHoldings(ctx context.Context) ([]GrayscaleHolding, error) {
	var out []GrayscaleHolding
	err := s.client.get(ctx, "/api/grayscale/holdings-list", nil, &out)
	return out, err
}
