package coinglass

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

// RawEnvelope returns a byte-identical copy of the provider envelope. It is
// retained even when the response fails this method's admission checks.
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
// current-v4 array shape. A market must be a non-null object with at least one
// documented margin-list member. That is the smallest current-v4 discriminator:
// the documentation's example does not establish symbol or both lists as
// globally required, but a legacy flat rate row has neither margin-list member.
func (r *FundingRateExchangeListV4Response) UnmarshalJSON(data []byte) error {
	r.raw = append(r.raw[:0], data...)
	r.markets = nil

	var rawMarkets []json.RawMessage
	if err := json.Unmarshal(data, &rawMarkets); err != nil {
		return err
	}

	markets := make([]FundingRateExchangeMarket, 0, len(rawMarkets))
	for i, rawMarket := range rawMarkets {
		if err := validateFundingRateExchangeMarket(rawMarket); err != nil {
			return fmt.Errorf("market %d: %w", i, err)
		}

		var market FundingRateExchangeMarket
		if err := json.Unmarshal(rawMarket, &market); err != nil {
			return fmt.Errorf("market %d: %w", i, err)
		}
		markets = append(markets, market)
	}

	r.markets = markets
	return nil
}

func validateFundingRateExchangeMarket(data json.RawMessage) error {
	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) {
		return fmt.Errorf("must be a non-null current-v4 market object")
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil || fields == nil {
		if err != nil {
			return fmt.Errorf("must be a current-v4 market object: %w", err)
		}
		return fmt.Errorf("must be a non-null current-v4 market object")
	}
	if _, ok := fields["stablecoin_margin_list"]; ok {
		return nil
	}
	if _, ok := fields["token_margin_list"]; ok {
		return nil
	}
	return fmt.Errorf("does not contain a current-v4 margin list")
}

// decodeResponseEnvelope strictly admits this current-v4 response without
// changing legacy methods that use the shared envelope decoder. RawEnvelope
// and Receipts retain the exact body on every admission failure.
func (r *FundingRateExchangeListV4Response) decodeResponseEnvelope(body []byte) error {
	r.setRawEnvelope(body)
	r.raw = nil
	r.markets = nil

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil || fields == nil {
		if err != nil {
			return fmt.Errorf("coinglass: failed to decode response envelope: %w", err)
		}
		return fmt.Errorf("coinglass: response envelope must be an object")
	}

	code, err := requiredFundingEnvelopeString(fields, "code")
	if err != nil {
		return err
	}
	if code != "0" {
		var message string
		_ = json.Unmarshal(fields["msg"], &message)
		return &APIError{
			StatusCode: 200,
			Code:       code,
			Message:    message,
			RawBody:    body,
		}
	}

	if _, err := requiredFundingEnvelopeString(fields, "msg"); err != nil {
		return err
	}
	data, ok := fields["data"]
	if !ok || bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return fmt.Errorf("coinglass: funding-by-exchange response requires a non-null array data field")
	}
	r.raw = append(r.raw[:0], data...)

	var dataShape []json.RawMessage
	if err := json.Unmarshal(data, &dataShape); err != nil {
		return fmt.Errorf("coinglass: funding-by-exchange response data must be an array: %w", err)
	}
	if err := json.Unmarshal(data, r); err != nil {
		return fmt.Errorf("coinglass: failed to decode response data: %w", err)
	}
	return nil
}

func requiredFundingEnvelopeString(fields map[string]json.RawMessage, name string) (string, error) {
	raw, ok := fields[name]
	if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return "", fmt.Errorf("coinglass: funding-by-exchange response requires a non-null string %q field", name)
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", fmt.Errorf("coinglass: funding-by-exchange response requires a string %q field: %w", name, err)
	}
	return value, nil
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
