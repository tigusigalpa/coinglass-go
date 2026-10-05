package coinglass

import (
	"context"
	"encoding/json"
	"time"
)

const fundingRateExchangeListDocumentationURL = "https://docs.coinglass.com/reference/fr-exchange-list"

// FundingRateExchangeListV4Response is the lossless current-v4 response for
// FundingRateExchangeListV4. Snapshot and raw accessors return defensive
// copies, so callers cannot alter the retained provider receipt.
type FundingRateExchangeListV4Response struct {
	markets    []FundingRateExchangeMarket
	raw        json.RawMessage
	envelope   json.RawMessage
	receipts   []HistoryReceipt
	provenance HistoryProvenance
}

// Snapshot returns a defensive copy of the markets in their provider order.
func (r *FundingRateExchangeListV4Response) Snapshot() []FundingRateExchangeMarket {
	out := make([]FundingRateExchangeMarket, len(r.markets))
	for i := range r.markets {
		out[i] = r.markets[i].clone()
	}
	return out
}

// RawData returns a byte-identical copy of the provider's data payload.
func (r *FundingRateExchangeListV4Response) RawData() json.RawMessage {
	return append(json.RawMessage(nil), r.raw...)
}

// RawEnvelope returns a byte-identical copy of the successful provider envelope.
func (r *FundingRateExchangeListV4Response) RawEnvelope() json.RawMessage {
	return append(json.RawMessage(nil), r.envelope...)
}

// Receipts returns defensive copies of all in-memory HTTP receipts for this call.
func (r *FundingRateExchangeListV4Response) Receipts() []HistoryReceipt {
	out := make([]HistoryReceipt, len(r.receipts))
	for i := range r.receipts {
		out[i] = HistoryReceipt{
			Body:       append(json.RawMessage(nil), r.receipts[i].Body...),
			CapturedAt: r.receipts[i].CapturedAt,
		}
	}
	return out
}

// Provenance returns the primary source documentation for this contract.
func (r *FundingRateExchangeListV4Response) Provenance() HistoryProvenance {
	return r.provenance
}

// UnmarshalJSON retains the original data bytes and decodes only the documented
// current-v4 array shape.
func (r *FundingRateExchangeListV4Response) UnmarshalJSON(data []byte) error {
	var markets []FundingRateExchangeMarket
	if err := json.Unmarshal(data, &markets); err != nil {
		return err
	}
	r.raw = append(r.raw[:0], data...)
	r.markets = markets
	return nil
}

func (r *FundingRateExchangeListV4Response) setRawEnvelope(data []byte) {
	r.envelope = append(r.envelope[:0], data...)
}

func (r *FundingRateExchangeListV4Response) addRawReceipt(data []byte) {
	r.receipts = append(r.receipts, HistoryReceipt{
		Body:       append(json.RawMessage(nil), data...),
		CapturedAt: time.Now().UTC(),
	})
}

// FundingRateExchangeMarket groups the two documented margin-mode lists for
// one symbol. Raw retains unknown parent fields.
type FundingRateExchangeMarket struct {
	Symbol               string                    `json:"symbol"`
	StablecoinMarginList []FundingRateExchangeRate `json:"stablecoin_margin_list"`
	TokenMarginList      []FundingRateExchangeRate `json:"token_margin_list"`
	Raw                  json.RawMessage           `json:"-"`
}

// UnmarshalJSON retains the complete provider parent object, including fields
// the SDK does not currently model.
func (m *FundingRateExchangeMarket) UnmarshalJSON(data []byte) error {
	type plain FundingRateExchangeMarket
	var decoded plain
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*m = FundingRateExchangeMarket(decoded)
	m.Raw = append(m.Raw[:0], data...)
	return nil
}

func (m FundingRateExchangeMarket) clone() FundingRateExchangeMarket {
	m.Raw = append(json.RawMessage(nil), m.Raw...)
	m.StablecoinMarginList = cloneFundingRateExchangeRates(m.StablecoinMarginList)
	m.TokenMarginList = cloneFundingRateExchangeRates(m.TokenMarginList)
	return m
}

func cloneFundingRateExchangeRates(in []FundingRateExchangeRate) []FundingRateExchangeRate {
	if in == nil {
		return nil
	}
	out := make([]FundingRateExchangeRate, len(in))
	for i := range in {
		out[i] = in[i].clone()
	}
	return out
}

// FundingRateExchangeRate is one exchange and margin-mode funding-rate row.
// FundingRate preserves its exact JSON decimal lexeme. FundingRateInterval is
// expressed in hours and NextFundingTime is a Unix timestamp in milliseconds.
type FundingRateExchangeRate struct {
	Exchange            string            `json:"exchange"`
	FundingRateInterval Decimal           `json:"funding_rate_interval"`
	FundingRate         Decimal           `json:"funding_rate"`
	NextFundingTime     EpochMilliseconds `json:"next_funding_time"`
	Raw                 json.RawMessage   `json:"-"`
}

// UnmarshalJSON retains the complete provider row, including unknown fields.
func (r *FundingRateExchangeRate) UnmarshalJSON(data []byte) error {
	type plain FundingRateExchangeRate
	var decoded plain
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*r = FundingRateExchangeRate(decoded)
	r.Raw = append(r.Raw[:0], data...)
	return nil
}

func (r FundingRateExchangeRate) clone() FundingRateExchangeRate {
	r.Raw = append(json.RawMessage(nil), r.Raw...)
	r.FundingRateInterval.Raw = append(json.RawMessage(nil), r.FundingRateInterval.Raw...)
	r.FundingRate.Raw = append(json.RawMessage(nil), r.FundingRate.Raw...)
	r.NextFundingTime.Raw = append(json.RawMessage(nil), r.NextFundingTime.Raw...)
	return r
}

// FundingRateExchangeListV4 retrieves the documented current-v4 funding rate
// exchange list. The official contract does not establish query parameters, so
// this method sends none.
func (s *FuturesService) FundingRateExchangeListV4(ctx context.Context) (*FundingRateExchangeListV4Response, error) {
	out := new(FundingRateExchangeListV4Response)
	err := s.client.get(ctx, "/api/futures/funding-rate/exchange-list", nil, out)
	out.provenance = HistoryProvenance{
		Source:           "CoinGlass",
		DocumentationURL: fundingRateExchangeListDocumentationURL,
		Section:          "Response Data",
	}
	return out, err
}
