package coinglass

import (
	"context"
	"net/http"
	"testing"
)

func TestFutures_OpenInterestHistory(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("symbol"); got != "BTC" {
			t.Errorf("expected symbol=BTC, got %s", got)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"0","msg":"success","data":[{"openInterest":1.5,"openInterestUsd":100000,"t":123456}]}`))
	})
	defer srv.Close()

	out, err := c.Futures.OpenInterestHistory(context.Background(), &OIHistoryParams{
		Symbol:   "BTC",
		Interval: "1d",
		Limit:    IntPtr(30),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 || out[0].OpenInterestUsd != 100000 {
		t.Errorf("unexpected result: %+v", out)
	}
}

func TestFutures_LiquidationHeatmap(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/futures/liquidation/heatmap/model1" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"0","msg":"success","data":{"data":[{"price":50000,"liquidationUsd":1000}]}}`))
	})
	defer srv.Close()

	out, err := c.Futures.LiquidationHeatmap(context.Background(), 1, &LiquidationHeatmapParams{
		Symbol:   "BTC",
		Interval: "1d",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Model != 1 || len(out.Data) != 1 {
		t.Errorf("unexpected result: %+v", out)
	}
}

func TestFutures_SupportedExchangePairs(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"0","msg":"success","data":{"Binance":[{"exchange":"Binance","symbol":"BTC","pair":"BTCUSDT"}]}}`))
	})
	defer srv.Close()

	out, err := c.Futures.SupportedExchangePairs(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out["Binance"]) != 1 {
		t.Errorf("unexpected result: %+v", out)
	}
}

func TestFutures_WhaleAlert(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"0","msg":"success","data":[{"symbol":"BTC","side":"buy","size":10,"price":50000,"time":123,"link":"x"}]}`))
	})
	defer srv.Close()

	out, err := c.Futures.WhaleAlert(context.Background(), &WhaleAlertParams{Symbol: StringPtr("BTC")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 || out[0].Side != "buy" {
		t.Errorf("unexpected result: %+v", out)
	}
}
