package coinglass

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Decimal preserves a provider-supplied decimal exactly. Presence, null, and
// zero are distinct: Present is false when the field was absent, Null is true
// for JSON null, and Lexeme retains the exact decimal text otherwise.
type Decimal struct {
	Lexeme  string
	Raw     json.RawMessage
	Present bool
	Null    bool
}

// UnmarshalJSON implements lossless decimal decoding for numeric and quoted
// numeric values without converting them to float64.
func (d *Decimal) UnmarshalJSON(data []byte) error {
	d.Present = true
	d.Raw = append(d.Raw[:0], data...)
	d.Null = string(data) == "null"
	d.Lexeme = ""
	if d.Null {
		return nil
	}
	if len(data) > 0 && data[0] == '"' {
		return json.Unmarshal(data, &d.Lexeme)
	}
	d.Lexeme = string(data)
	return nil
}

// EpochMilliseconds preserves a millisecond epoch timestamp and its original
// JSON representation. It does not accept or infer the legacy t field.
type EpochMilliseconds struct {
	Value   int64
	Raw     json.RawMessage
	Present bool
	Null    bool
}

// UnmarshalJSON implements lossless epoch-millisecond decoding.
func (t *EpochMilliseconds) UnmarshalJSON(data []byte) error {
	t.Present = true
	t.Raw = append(t.Raw[:0], data...)
	t.Null = string(data) == "null"
	t.Value = 0
	if t.Null {
		return nil
	}
	var number json.Number
	if err := json.Unmarshal(data, &number); err != nil {
		return fmt.Errorf("coinglass: invalid epoch milliseconds: %w", err)
	}
	value, err := strconv.ParseInt(number.String(), 10, 64)
	if err != nil {
		return fmt.Errorf("coinglass: invalid epoch milliseconds: %w", err)
	}
	t.Value = value
	return nil
}

// HistoryReceipt is one raw HTTP response body captured during a history
// request. CapturedAt is SDK receipt time, not provider series time.
type HistoryReceipt struct {
	Body       json.RawMessage
	CapturedAt time.Time
}

// HistoryProvenance identifies the primary documentation contract used by a
// typed history method. It does not represent a provider dataset version.
type HistoryProvenance struct {
	Source           string
	DocumentationURL string
	Section          string
}

// HistoryResponse is a lossless outer DTO for a typed history series. Raw is
// the byte-exact data payload, Envelope is the successful API response, and
// Receipts contains every body received for this call (including retries).
// The SDK keeps these in memory only; it does not persist receipts.
type HistoryResponse[T any] struct {
	Data            []T
	Raw             json.RawMessage
	Envelope        json.RawMessage
	Receipts        []json.RawMessage
	ReceiptMetadata []HistoryReceipt
	Provenance      HistoryProvenance
}

// UnmarshalJSON decodes a history data array while retaining its original bytes.
func (r *HistoryResponse[T]) UnmarshalJSON(data []byte) error {
	r.Raw = append(r.Raw[:0], data...)
	return json.Unmarshal(data, &r.Data)
}

func (r *HistoryResponse[T]) setRawEnvelope(data []byte) {
	r.Envelope = append(r.Envelope[:0], data...)
}

func (r *HistoryResponse[T]) addRawReceipt(data []byte) {
	body := append(json.RawMessage(nil), data...)
	r.Receipts = append(r.Receipts, body)
	r.ReceiptMetadata = append(r.ReceiptMetadata, HistoryReceipt{
		Body:       append(json.RawMessage(nil), body...),
		CapturedAt: time.Now().UTC(),
	})
}

func setHistoryProvenance[T any](response *HistoryResponse[T], documentationURL, section string) {
	response.Provenance = HistoryProvenance{
		Source:           "CoinGlass",
		DocumentationURL: documentationURL,
		Section:          section,
	}
}

// OpenInterestUnit is the unit returned by AggregatedOpenInterestHistory.
type OpenInterestUnit string

const (
	// OpenInterestUnitUSD returns open interest denominated in USD.
	OpenInterestUnitUSD OpenInterestUnit = "usd"
	// OpenInterestUnitCoin returns open interest denominated in the requested coin.
	OpenInterestUnitCoin OpenInterestUnit = "coin"
)

// AggregatedOpenInterestHistoryParams defines the verified current-v4 query
// contract. StartTime and EndTime are sent as start_time and end_time.
type AggregatedOpenInterestHistoryParams struct {
	Symbol    string            `url:"symbol"`
	Interval  string            `url:"interval"`
	Unit      *OpenInterestUnit `url:"unit,omitempty"`
	Limit     *int              `url:"limit,omitempty"`
	StartTime *int64            `url:"start_time,omitempty"`
	EndTime   *int64            `url:"end_time,omitempty"`
}

// AggregatedOpenInterestOHLC is an aggregate across exchanges, not an
// exchange-level open-interest snapshot. Raw retains unknown provider fields.
type AggregatedOpenInterestOHLC struct {
	Time  EpochMilliseconds `json:"time"`
	Open  Decimal           `json:"open"`
	High  Decimal           `json:"high"`
	Low   Decimal           `json:"low"`
	Close Decimal           `json:"close"`
	Raw   json.RawMessage   `json:"-"`
}

// UnmarshalJSON retains the complete provider object, including unknown fields.
func (p *AggregatedOpenInterestOHLC) UnmarshalJSON(data []byte) error {
	type plain AggregatedOpenInterestOHLC
	var decoded plain
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*p = AggregatedOpenInterestOHLC(decoded)
	p.Raw = append(p.Raw[:0], data...)
	return nil
}

// AggregatedOpenInterestHistory retrieves the current-v4 aggregated OI OHLC
// contract at /api/futures/open-interest/aggregated-history.
func (s *FuturesService) AggregatedOpenInterestHistory(ctx context.Context, params *AggregatedOpenInterestHistoryParams) (*HistoryResponse[AggregatedOpenInterestOHLC], error) {
	out := new(HistoryResponse[AggregatedOpenInterestOHLC])
	err := s.client.get(ctx, "/api/futures/open-interest/aggregated-history", buildQuery(params), out)
	setHistoryProvenance(out, "https://docs.coinglass.com/reference/oi-ohlc-aggregated-history", "Response Data")
	return out, err
}

// OIWeightedFundingHistoryParams deliberately omits limit, start_time, and
// end_time. CoinGlass documents those fields as seconds while showing
// 13-digit examples; this SDK will not guess a unit before clarification.
type OIWeightedFundingHistoryParams struct {
	Symbol   string `url:"symbol"`
	Interval string `url:"interval"`
}

// OIWeightedFundingOHLC is OI-weighted funding rate data, not a settled rate.
// Raw retains unknown provider fields.
type OIWeightedFundingOHLC struct {
	Time  EpochMilliseconds `json:"time"`
	Open  Decimal           `json:"open"`
	High  Decimal           `json:"high"`
	Low   Decimal           `json:"low"`
	Close Decimal           `json:"close"`
	Raw   json.RawMessage   `json:"-"`
}

// UnmarshalJSON retains the complete provider object, including unknown fields.
func (p *OIWeightedFundingOHLC) UnmarshalJSON(data []byte) error {
	type plain OIWeightedFundingOHLC
	var decoded plain
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*p = OIWeightedFundingOHLC(decoded)
	p.Raw = append(p.Raw[:0], data...)
	return nil
}

// OIWeightedFundingHistory retrieves the verified query subset of the current
// v4 OI-weighted funding OHLC endpoint.
func (s *FuturesService) OIWeightedFundingHistory(ctx context.Context, params *OIWeightedFundingHistoryParams) (*HistoryResponse[OIWeightedFundingOHLC], error) {
	out := new(HistoryResponse[OIWeightedFundingOHLC])
	err := s.client.get(ctx, "/api/futures/funding-rate/oi-weight-history", buildQuery(params), out)
	setHistoryProvenance(out, "https://docs.coinglass.com/reference/oi-weight-ohlc-history", "Response Data")
	return out, err
}

// AggregatedLiquidationHistoryParams defines an aggregate identity that
// includes exchange selection. StartTime and EndTime use snake_case.
type AggregatedLiquidationHistoryParams struct {
	ExchangeList []string `url:"exchange_list"`
	Symbol       string   `url:"symbol"`
	Interval     string   `url:"interval"`
	Limit        *int     `url:"limit,omitempty"`
	StartTime    *int64   `url:"start_time,omitempty"`
	EndTime      *int64   `url:"end_time,omitempty"`
}

// AggregatedLiquidationPoint contains interval totals, not liquidation events.
// Values retain their exact JSON decimal lexemes and Raw keeps unknown fields.
type AggregatedLiquidationPoint struct {
	Time                          EpochMilliseconds `json:"time"`
	AggregatedLongLiquidationUSD  Decimal           `json:"aggregated_long_liquidation_usd"`
	AggregatedShortLiquidationUSD Decimal           `json:"aggregated_short_liquidation_usd"`
	Raw                           json.RawMessage   `json:"-"`
}

// UnmarshalJSON retains the complete provider object, including unknown fields.
func (p *AggregatedLiquidationPoint) UnmarshalJSON(data []byte) error {
	type plain AggregatedLiquidationPoint
	var decoded plain
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*p = AggregatedLiquidationPoint(decoded)
	p.Raw = append(p.Raw[:0], data...)
	return nil
}

// AggregatedLiquidationHistory retrieves aggregate coin liquidations for the
// requested exchange composition.
func (s *FuturesService) AggregatedLiquidationHistory(ctx context.Context, params *AggregatedLiquidationHistoryParams) (*HistoryResponse[AggregatedLiquidationPoint], error) {
	out := new(HistoryResponse[AggregatedLiquidationPoint])
	err := s.client.get(ctx, "/api/futures/liquidation/aggregated-history", buildQuery(params), out)
	setHistoryProvenance(out, "https://docs.coinglass.com/reference/aggregated-liquidation-history", "Response Data")
	return out, err
}

// GlobalAccountRatioHistoryParams preserves the exchange and trading-pair
// scope of the provider's account-ratio analytics contract.
type GlobalAccountRatioHistoryParams struct {
	Exchange  string `url:"exchange"`
	Symbol    string `url:"symbol"`
	Interval  string `url:"interval"`
	Limit     *int   `url:"limit,omitempty"`
	StartTime *int64 `url:"start_time,omitempty"`
	EndTime   *int64 `url:"end_time,omitempty"`
}

// GlobalAccountRatioPoint contains percentage points and the provider's
// long/short ratio. Raw retains unknown provider fields.
type GlobalAccountRatioPoint struct {
	Time                        EpochMilliseconds `json:"time"`
	GlobalAccountLongPercent    Decimal           `json:"global_account_long_percent"`
	GlobalAccountShortPercent   Decimal           `json:"global_account_short_percent"`
	GlobalAccountLongShortRatio Decimal           `json:"global_account_long_short_ratio"`
	Raw                         json.RawMessage   `json:"-"`
}

// UnmarshalJSON retains the complete provider object, including unknown fields.
func (p *GlobalAccountRatioPoint) UnmarshalJSON(data []byte) error {
	type plain GlobalAccountRatioPoint
	var decoded plain
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*p = GlobalAccountRatioPoint(decoded)
	p.Raw = append(p.Raw[:0], data...)
	return nil
}

// GlobalAccountRatioHistory retrieves account-ratio history at the current-v4
// route /api/futures/global-long-short-account-ratio/history.
func (s *FuturesService) GlobalAccountRatioHistory(ctx context.Context, params *GlobalAccountRatioHistoryParams) (*HistoryResponse[GlobalAccountRatioPoint], error) {
	out := new(HistoryResponse[GlobalAccountRatioPoint])
	err := s.client.get(ctx, "/api/futures/global-long-short-account-ratio/history", buildQuery(params), out)
	setHistoryProvenance(out, "https://docs.coinglass.com/reference/global-longshort-account-ratio", "Response Data")
	return out, err
}
