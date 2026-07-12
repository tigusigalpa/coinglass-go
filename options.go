package coinglass

import (
	"context"
	"encoding/json"
)

// OptionsService provides access to all Coinglass options endpoints.
type OptionsService struct {
	client *Client
}

// OptionMaxPain represents the max-pain analysis for an underlying asset.
type OptionMaxPain struct {
	Underlying string          `json:"underlying"`
	Expiry     int64           `json:"expiry"`
	MaxPain    float64         `json:"maxPain"`
	Data       json.RawMessage `json:"data"`
	Raw        json.RawMessage `json:"-"`
}

// OptionParams holds parameters for MaxPain and Info.
type OptionParams struct {
	Underlying string  `url:"underlying"`
	Expiry     *int64  `url:"expiry,omitempty"`
	Interval   *string `url:"interval,omitempty"`
}

// MaxPain returns the option max-pain analysis for the given underlying.
func (s *OptionsService) MaxPain(ctx context.Context, params *OptionParams) (*OptionMaxPain, error) {
	var out OptionMaxPain
	err := s.client.get(ctx, "/api/option/max-pain", buildQuery(params), &out)
	return &out, err
}

// OptionInfo represents general options information for an underlying.
type OptionInfo struct {
	Underlying string          `json:"underlying"`
	Expiry     int64           `json:"expiry"`
	Data       json.RawMessage `json:"data"`
	Raw        json.RawMessage `json:"-"`
}

// Info returns options information for the given underlying.
func (s *OptionsService) Info(ctx context.Context, params *OptionParams) (*OptionInfo, error) {
	var out OptionInfo
	err := s.client.get(ctx, "/api/option/info", buildQuery(params), &out)
	return &out, err
}

// OptionOIPoint represents a single exchange open-interest history point.
type OptionOIPoint struct {
	Exchange  string  `json:"exchange"`
	TotalOI   float64 `json:"totalOI"`
	Timestamp int64   `json:"t"`
}

// OptionHistoryParams holds parameters for options history endpoints.
type OptionHistoryParams struct {
	Interval  string `url:"interval"`
	Limit     *int   `url:"limit,omitempty"`
	StartTime *int64 `url:"startTime,omitempty"`
	EndTime   *int64 `url:"endTime,omitempty"`
}

// ExchangeOIHistory returns options exchange open-interest history.
func (s *OptionsService) ExchangeOIHistory(ctx context.Context, params *OptionHistoryParams) ([]OptionOIPoint, error) {
	var out []OptionOIPoint
	err := s.client.get(ctx, "/api/option/exchange-oi-history", buildQuery(params), &out)
	return out, err
}

// OptionVolPoint represents a single exchange volume history point.
type OptionVolPoint struct {
	Exchange  string  `json:"exchange"`
	TotalVol  float64 `json:"totalVol"`
	Timestamp int64   `json:"t"`
}

// ExchangeVolHistory returns options exchange volume history.
func (s *OptionsService) ExchangeVolHistory(ctx context.Context, params *OptionHistoryParams) ([]OptionVolPoint, error) {
	var out []OptionVolPoint
	err := s.client.get(ctx, "/api/option/exchange-vol-history", buildQuery(params), &out)
	return out, err
}
