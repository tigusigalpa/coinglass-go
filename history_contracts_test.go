package coinglass

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"testing"
)

const fixtureKind = "doc_schema_example"

func assertHistoryReceipt[T any](t *testing.T, response *HistoryResponse[T], body, documentationURL string) {
	t.Helper()
	if fixtureKind != "doc_schema_example" {
		t.Fatal("fixture must not be presented as a live provider capture")
	}
	if len(response.Receipts) != 1 || string(response.Receipts[0]) != body || sha256.Sum256(response.Envelope) != sha256.Sum256([]byte(body)) {
		t.Fatal("response receipt bytes changed")
	}
	if len(response.ReceiptMetadata) != 1 || string(response.ReceiptMetadata[0].Body) != body || response.ReceiptMetadata[0].CapturedAt.IsZero() {
		t.Fatal("receipt metadata is incomplete")
	}
	if response.Provenance.Source != "CoinGlass" || response.Provenance.DocumentationURL != documentationURL || response.Provenance.Section != "Response Data" {
		t.Fatalf("unexpected provenance: %+v", response.Provenance)
	}
}

func TestAggregatedOpenInterestHistoryContract(t *testing.T) {
	unit := OpenInterestUnitUSD
	start, end := int64(1641522717000), int64(1641609117000)
	body := `{"code":"0","msg":"success","data":[{"time":2644845344000,"open":"2644845344.000","high":"2692643311","low":"2576975597","close":"2608846475","future_field":true}]}`
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/futures/open-interest/aggregated-history" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("symbol") != "BTC" || q.Get("interval") != "1d" || q.Get("unit") != "usd" || q.Get("limit") != "1" || q.Get("start_time") != "1641522717000" || q.Get("end_time") != "1641609117000" {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		if q.Has("startTime") || q.Has("endTime") {
			t.Errorf("legacy camelCase query found: %s", r.URL.RawQuery)
		}
		writeTestResponse(t, w, body)
	})
	defer srv.Close()

	response, err := c.Futures.AggregatedOpenInterestHistory(context.Background(), &AggregatedOpenInterestHistoryParams{
		Symbol: "BTC", Interval: "1d", Unit: &unit, Limit: IntPtr(1), StartTime: &start, EndTime: &end,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Data) != 1 || response.Data[0].Time.Value != 2644845344000 || response.Data[0].Open.Lexeme != "2644845344.000" {
		t.Fatalf("unexpected response: %+v", response.Data)
	}
	if !response.Data[0].Open.Present || response.Data[0].Open.Null || !response.Data[0].Time.Present || response.Data[0].Time.Null {
		t.Fatal("expected present, non-null values")
	}
	if string(response.Data[0].Raw) != `{"time":2644845344000,"open":"2644845344.000","high":"2692643311","low":"2576975597","close":"2608846475","future_field":true}` {
		t.Fatalf("unknown fields were not retained: %s", response.Data[0].Raw)
	}
	assertHistoryReceipt(t, response, body, "https://docs.coinglass.com/reference/oi-ohlc-aggregated-history")
}

func TestOIWeightedFundingHistoryContract(t *testing.T) {
	body := `{"code":"0","msg":"success","data":[{"time":1658880000000,"open":"0.004603","high":"0.009388","low":"-0.005063","close":"0.009229"}]}`
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/futures/funding-rate/oi-weight-history" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("symbol") != "BTC" || q.Get("interval") != "1d" || q.Has("limit") || q.Has("start_time") || q.Has("end_time") {
			t.Errorf("unexpected quarantined query: %s", r.URL.RawQuery)
		}
		writeTestResponse(t, w, body)
	})
	defer srv.Close()

	response, err := c.Futures.OIWeightedFundingHistory(context.Background(), &OIWeightedFundingHistoryParams{Symbol: "BTC", Interval: "1d"})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Data) != 1 || response.Data[0].Time.Value != 1658880000000 || response.Data[0].Low.Lexeme != "-0.005063" {
		t.Fatalf("unexpected response: %+v", response.Data)
	}
	assertHistoryReceipt(t, response, body, "https://docs.coinglass.com/reference/oi-weight-ohlc-history")
}

func TestAggregatedLiquidationHistoryContract(t *testing.T) {
	body := `{"code":"0","msg":"success","data":[{"time":1658966400000,"aggregated_long_liquidation_usd":5916885.14234,"aggregated_short_liquidation_usd":12969583.87632}]}`
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/futures/liquidation/aggregated-history" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("exchange_list") != "Binance,OKX" || q.Get("symbol") != "BTC" || q.Get("interval") != "4h" {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		writeTestResponse(t, w, body)
	})
	defer srv.Close()

	response, err := c.Futures.AggregatedLiquidationHistory(context.Background(), &AggregatedLiquidationHistoryParams{ExchangeList: []string{"Binance", "OKX"}, Symbol: "BTC", Interval: "4h"})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Data) != 1 || response.Data[0].AggregatedLongLiquidationUSD.Lexeme != "5916885.14234" || response.Data[0].AggregatedShortLiquidationUSD.Lexeme != "12969583.87632" {
		t.Fatalf("unexpected response: %+v", response.Data)
	}
	assertHistoryReceipt(t, response, body, "https://docs.coinglass.com/reference/aggregated-liquidation-history")
}

func TestGlobalAccountRatioHistoryContract(t *testing.T) {
	body := `{"code":"0","msg":"success","data":[{"time":1741604400000,"global_account_long_percent":73.88,"global_account_short_percent":26.12,"global_account_long_short_ratio":2.83,"nullable_field":null}]}`
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/futures/global-long-short-account-ratio/history" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("exchange") != "Binance" || q.Get("symbol") != "BTCUSDT" || q.Get("interval") != "4h" {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		writeTestResponse(t, w, body)
	})
	defer srv.Close()

	response, err := c.Futures.GlobalAccountRatioHistory(context.Background(), &GlobalAccountRatioHistoryParams{Exchange: "Binance", Symbol: "BTCUSDT", Interval: "4h"})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Data) != 1 || response.Data[0].GlobalAccountLongPercent.Lexeme != "73.88" || response.Data[0].GlobalAccountShortPercent.Lexeme != "26.12" || response.Data[0].GlobalAccountLongShortRatio.Lexeme != "2.83" {
		t.Fatalf("unexpected response: %+v", response.Data)
	}
	if string(response.Data[0].Raw) != `{"time":1741604400000,"global_account_long_percent":73.88,"global_account_short_percent":26.12,"global_account_long_short_ratio":2.83,"nullable_field":null}` {
		t.Fatalf("unknown fields were not retained: %s", response.Data[0].Raw)
	}
	assertHistoryReceipt(t, response, body, "https://docs.coinglass.com/reference/global-longshort-account-ratio")
}

func TestDecimalPresenceAndNull(t *testing.T) {
	var payload struct {
		Zero    Decimal `json:"zero"`
		Null    Decimal `json:"null"`
		Missing Decimal `json:"missing"`
	}
	if err := json.Unmarshal([]byte(`{"zero":0,"null":null}`), &payload); err != nil {
		t.Fatal(err)
	}
	if !payload.Zero.Present || payload.Zero.Null || payload.Zero.Lexeme != "0" || !payload.Null.Present || !payload.Null.Null || payload.Missing.Present {
		t.Fatalf("presence was not preserved: %+v", payload)
	}
}

func TestHistoryResponseRetainsRetryReceipts(t *testing.T) {
	first := `{"message":"rate limited"}`
	second := `{"code":"0","msg":"success","data":[{"time":1658880000000,"open":"0","high":"0","low":"0","close":"0"}]}`
	attempts := 0
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			writeTestResponse(t, w, first)
			return
		}
		writeTestResponse(t, w, second)
	}, WithRetry(2, 1))
	defer srv.Close()

	response, err := c.Futures.OIWeightedFundingHistory(context.Background(), &OIWeightedFundingHistoryParams{Symbol: "BTC", Interval: "1d"})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Receipts) != 2 || string(response.Receipts[0]) != first || string(response.Receipts[1]) != second {
		t.Fatalf("unexpected retry receipts: %q", response.Receipts)
	}
	if len(response.ReceiptMetadata) != 2 || response.ReceiptMetadata[0].CapturedAt.IsZero() || response.ReceiptMetadata[1].CapturedAt.IsZero() {
		t.Fatalf("retry receipt metadata is incomplete: %+v", response.ReceiptMetadata)
	}
}
