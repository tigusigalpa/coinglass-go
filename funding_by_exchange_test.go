package coinglass

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"errors"
	"net/http"
	"testing"
)

//go:embed testdata/funding_rate_exchange_list_doc_schema_example.json
var fundingRateExchangeListFixture []byte

func TestFundingRateExchangeListV4Contract(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/futures/funding-rate/exchange-list" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.RawQuery != "" {
			t.Errorf("unexpected undocumented query: %s", r.URL.RawQuery)
		}
		writeTestResponse(t, w, string(fundingRateExchangeListFixture))
	})
	defer srv.Close()

	response, err := c.Futures.FundingRateExchangeListV4(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	markets := response.Snapshot()
	if len(markets) != 1 || markets[0].Symbol != "BTC" {
		t.Fatalf("unexpected markets: %+v", markets)
	}
	if len(markets[0].StablecoinMarginList) != 2 || len(markets[0].TokenMarginList) != 1 {
		t.Fatalf("margin modes were not preserved: %+v", markets[0])
	}
	if markets[0].StablecoinMarginList[0].Exchange != "Binance" || markets[0].StablecoinMarginList[1].Exchange != "OKX" || markets[0].TokenMarginList[0].Exchange != "Binance" {
		t.Fatalf("exchange order was not preserved: %+v", markets[0])
	}
	if markets[0].StablecoinMarginList[1].FundingRate.Lexeme != "0.00736901950628" || markets[0].TokenMarginList[0].FundingRate.Lexeme != "-0.001829" {
		t.Fatalf("funding-rate lexemes changed: %+v", markets[0])
	}
	if markets[0].StablecoinMarginList[0].FundingRateInterval.Lexeme != "8" || markets[0].StablecoinMarginList[0].NextFundingTime.Value != 1745222400000 {
		t.Fatalf("native interval or next funding time changed: %+v", markets[0].StablecoinMarginList[0])
	}
	if got := response.Provenance(); got.Source != "CoinGlass" || got.DocumentationURL != fundingRateExchangeListDocumentationURL || got.Section != "Response Data" {
		t.Fatalf("unexpected provenance: %+v", got)
	}
	if sha256.Sum256(response.RawEnvelope()) != sha256.Sum256(fundingRateExchangeListFixture) {
		t.Fatal("provider envelope was not retained byte-for-byte")
	}
	if string(response.RawData()) != `[{"symbol":"BTC","stablecoin_margin_list":[{"exchange":"Binance","funding_rate_interval":8,"funding_rate":0.007343,"next_funding_time":1745222400000},{"exchange":"OKX","funding_rate_interval":8,"funding_rate":0.00736901950628,"next_funding_time":1745222400000}],"token_margin_list":[{"exchange":"Binance","funding_rate_interval":8,"funding_rate":-0.001829,"next_funding_time":1745222400000}]}]` {
		t.Fatal("provider data payload was not retained byte-for-byte")
	}
	if receipts := response.Receipts(); len(receipts) != 1 || string(receipts[0].Body) != string(fundingRateExchangeListFixture) || receipts[0].CapturedAt.IsZero() {
		t.Fatalf("unexpected receipts: %+v", receipts)
	}

	markets[0].Symbol = "mutated"
	markets[0].StablecoinMarginList[0].FundingRate.Lexeme = "mutated"
	markets[0].Raw[0] = 'x'
	if next := response.Snapshot(); next[0].Symbol != "BTC" || next[0].StablecoinMarginList[0].FundingRate.Lexeme != "0.007343" || next[0].Raw[0] != '{' {
		t.Fatal("Snapshot must return a defensive copy")
	}
	envelope := response.RawEnvelope()
	envelope[0] = 'x'
	if response.RawEnvelope()[0] != '{' {
		t.Fatal("RawEnvelope must return a defensive copy")
	}
	rawData := response.RawData()
	rawData[0] = 'x'
	if response.RawData()[0] != '[' {
		t.Fatal("RawData must return a defensive copy")
	}
	receipts := response.Receipts()
	receipts[0].Body[0] = 'x'
	if response.Receipts()[0].Body[0] != '{' {
		t.Fatal("Receipts must return defensive copies")
	}
}

func TestFundingRateExchangeListV4SyntheticNumericAndPresence(t *testing.T) {
	const synthetic = `{"code":"0","msg":"success","data":[{"symbol":"BTC","unknown_parent":{"preserve":true},"stablecoin_margin_list":[{"exchange":"Zero","funding_rate_interval":0,"funding_rate":0,"next_funding_time":0,"unknown_row":"value"},{"exchange":"Large","funding_rate_interval":8,"funding_rate":9007199254740993,"next_funding_time":1745222400000},{"exchange":"Huge","funding_rate_interval":8,"funding_rate":1e400,"next_funding_time":1745222400000},{"exchange":"Tiny","funding_rate_interval":8,"funding_rate":1e-1000,"next_funding_time":1745222400000}],"token_margin_list":[{"exchange":null,"funding_rate_interval":null,"funding_rate":null,"next_funding_time":null,"unknown_token_row":true},{}]},{"symbol":"EMPTY","stablecoin_margin_list":[],"token_margin_list":[]}],"unknown_envelope":"preserve"}`
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeTestResponse(t, w, synthetic)
	})
	defer srv.Close()

	response, err := c.Futures.FundingRateExchangeListV4(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	markets := response.Snapshot()
	if len(markets) != 2 || markets[1].Symbol != "EMPTY" || len(markets[1].StablecoinMarginList) != 0 || len(markets[1].TokenMarginList) != 0 {
		t.Fatalf("empty margin lists were not preserved: %+v", markets)
	}
	rows := markets[0].StablecoinMarginList
	if rows[0].FundingRate.Lexeme != "0" || !rows[0].FundingRate.Present || rows[0].FundingRate.Null || rows[0].NextFundingTime.Value != 0 {
		t.Fatalf("zero was not preserved: %+v", rows[0])
	}
	if rows[1].FundingRate.Lexeme != "9007199254740993" || rows[2].FundingRate.Lexeme != "1e400" || rows[3].FundingRate.Lexeme != "1e-1000" {
		t.Fatalf("numeric lexemes changed: %+v", rows)
	}
	if string(rows[0].Raw) == "" || string(markets[0].TokenMarginList[0].Raw) == "" || string(markets[0].Raw) == "" || string(response.RawEnvelope()) != synthetic {
		t.Fatal("unknown fields or raw envelope were not retained")
	}
	if string(markets[0].Raw) != `{"symbol":"BTC","unknown_parent":{"preserve":true},"stablecoin_margin_list":[{"exchange":"Zero","funding_rate_interval":0,"funding_rate":0,"next_funding_time":0,"unknown_row":"value"},{"exchange":"Large","funding_rate_interval":8,"funding_rate":9007199254740993,"next_funding_time":1745222400000},{"exchange":"Huge","funding_rate_interval":8,"funding_rate":1e400,"next_funding_time":1745222400000},{"exchange":"Tiny","funding_rate_interval":8,"funding_rate":1e-1000,"next_funding_time":1745222400000}],"token_margin_list":[{"exchange":null,"funding_rate_interval":null,"funding_rate":null,"next_funding_time":null,"unknown_token_row":true},{}]}` || string(rows[0].Raw) == "" {
		t.Fatal("unknown parent or row fields were not retained byte-for-byte")
	}
	nullRow := markets[0].TokenMarginList[0]
	missingRow := markets[0].TokenMarginList[1]
	if !nullRow.FundingRate.Present || !nullRow.FundingRate.Null || !nullRow.FundingRateInterval.Null || !nullRow.NextFundingTime.Null || missingRow.FundingRate.Present || missingRow.FundingRateInterval.Present || missingRow.NextFundingTime.Present {
		t.Fatalf("absent and null values were not distinguished: null=%+v missing=%+v", nullRow, missingRow)
	}
}

func TestFundingRateExchangeListV4AdmissionValidation(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{
			name: "synthetic legacy flat row",
			body: `{"code":"0","msg":"success","data":[{"exchange":"Binance","fundingRate":0.01,"t":123}]}`,
		},
		{
			name: "synthetic null parent",
			body: `{"code":"0","msg":"success","data":[null]}`,
		},
		{
			name: "synthetic mixed current and legacy parents",
			body: `{"code":"0","msg":"success","data":[{"symbol":"BTC","stablecoin_margin_list":[]},{"exchange":"Binance","fundingRate":0.01,"t":123}]}`,
		},
		{
			name: "synthetic missing data",
			body: `{"code":"0","msg":"success"}`,
		},
		{
			name: "synthetic missing code",
			body: `{"msg":"success","data":[]}`,
		},
		{
			name: "synthetic missing msg",
			body: `{"code":"0","data":[]}`,
		},
		{
			name: "synthetic null code",
			body: `{"code":null,"msg":"success","data":[]}`,
		},
		{
			name: "synthetic malformed code",
			body: `{"code":0,"msg":"success","data":[]}`,
		},
		{
			name: "synthetic null msg",
			body: `{"code":"0","msg":null,"data":[]}`,
		},
		{
			name: "synthetic malformed msg",
			body: `{"code":"0","msg":true,"data":[]}`,
		},
		{
			name: "synthetic null data",
			body: `{"code":"0","msg":"success","data":null}`,
		},
		{
			name: "synthetic object data",
			body: `{"code":"0","msg":"success","data":{"symbol":"BTC"}}`,
		},
		{
			name: "synthetic scalar data",
			body: `{"code":"0","msg":"success","data":42}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			attempts := 0
			c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				attempts++
				if r.Method != http.MethodGet || r.URL.Path != "/api/futures/funding-rate/exchange-list" || r.URL.RawQuery != "" {
					t.Errorf("unexpected request: %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
				}
				writeTestResponse(t, w, tc.body)
			}, WithRetry(3, 0))
			defer srv.Close()

			response, err := c.Futures.FundingRateExchangeListV4(context.Background())
			if err == nil {
				t.Fatal("expected admission error")
			}
			if attempts != 1 {
				t.Fatalf("admission failure made %d requests, want 1", attempts)
			}
			assertFundingRateExchangeFailureEvidence(t, response, tc.body)
			if markets := response.Snapshot(); len(markets) != 0 {
				t.Fatalf("invalid response admitted typed markets: %+v", markets)
			}
		})
	}
}

func TestFundingRateExchangeListV4AcceptsEmptyCurrentV4Results(t *testing.T) {
	cases := []struct {
		name string
		body string
		want int
	}{
		{
			name: "empty data array",
			body: `{"code":"0","msg":"success","data":[]}`,
		},
		{
			name: "documented parent with empty margin lists",
			body: `{"code":"0","msg":"success","data":[{"symbol":"BTC","stablecoin_margin_list":[],"token_margin_list":[]}]}`,
			want: 1,
		},
		{
			name: "one documented margin list",
			body: `{"code":"0","msg":"success","data":[{"stablecoin_margin_list":[]}]}`,
			want: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				writeTestResponse(t, w, tc.body)
			})
			defer srv.Close()

			response, err := c.Futures.FundingRateExchangeListV4(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if markets := response.Snapshot(); len(markets) != tc.want {
				t.Fatalf("got %d markets, want %d", len(markets), tc.want)
			}
		})
	}
}

func TestFundingRateExchangeListV4RejectsUnsupportedAndProviderErrorResponses(t *testing.T) {
	t.Run("unsupported data shape", func(t *testing.T) {
		c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			writeTestResponse(t, w, `{"code":"0","msg":"success","data":{"symbol":"BTC"}}`)
		})
		defer srv.Close()
		if _, err := c.Futures.FundingRateExchangeListV4(context.Background()); err == nil {
			t.Fatal("expected error for unsupported data shape")
		}
	})

	t.Run("provider error envelope", func(t *testing.T) {
		const body = `{"code":"30001","msg":"invalid parameter","data":null}`
		c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			writeTestResponse(t, w, body)
		})
		defer srv.Close()
		response, err := c.Futures.FundingRateExchangeListV4(context.Background())
		var apiErr *APIError
		if !errors.As(err, &apiErr) || apiErr.Code != "30001" {
			t.Fatalf("expected provider API error, got %v", err)
		}
		assertFundingRateExchangeFailureEvidence(t, response, body)
	})
}

func assertFundingRateExchangeFailureEvidence(t *testing.T, response *FundingRateExchangeListV4Response, body string) {
	t.Helper()
	if response == nil {
		t.Fatal("expected response with retained failure evidence")
	}
	if string(response.RawEnvelope()) != body {
		t.Fatalf("raw envelope was not retained: %q", response.RawEnvelope())
	}
	receipts := response.Receipts()
	if len(receipts) != 1 || string(receipts[0].Body) != body || receipts[0].CapturedAt.IsZero() {
		t.Fatalf("receipt was not retained: %+v", receipts)
	}
}

func TestFundingRateExchangeListLegacyCompatibility(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/futures/fundingRate/exchange-list" {
			t.Errorf("unexpected legacy path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("symbol") != "BTC" || r.URL.Query().Get("interval") != "1h" {
			t.Errorf("unexpected legacy query: %s", r.URL.RawQuery)
		}
		writeTestResponse(t, w, `{"code":"0","msg":"success","data":[{"exchange":"Binance","fundingRate":0.01,"t":123}]}`)
	})
	defer srv.Close()

	rows, err := c.Futures.FundingRateExchangeList(context.Background(), &FundingRateExchangeListParams{Symbol: "BTC", Interval: "1h"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Exchange != "Binance" || rows[0].FundingRate != 0.01 || rows[0].Timestamp != 123 {
		t.Fatalf("legacy contract changed: %+v", rows)
	}
}
