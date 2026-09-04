package coinglass

import (
	"context"
	"net/http"
	"testing"
)

func TestSpot_CoinsMarkets(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		writeTestResponse(t, w, `{"code":"0","msg":"success","data":[{"symbol":"BTC","price":50000,"priceChange24h":1.2,"volumeUsd24h":1000000}]}`)
	})
	defer srv.Close()

	out, err := c.Spot.CoinsMarkets(context.Background(), &SpotCoinsMarketsParams{Symbol: StringPtr("BTC")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 || out[0].Price != 50000 {
		t.Errorf("unexpected result: %+v", out)
	}
}

func TestSpot_PriceHistory(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		writeTestResponse(t, w, `{"code":"0","msg":"success","data":[{"open":1,"high":2,"low":0.5,"close":1.5,"volume":100,"t":123}]}`)
	})
	defer srv.Close()

	out, err := c.Spot.PriceHistory(context.Background(), &SpotPriceHistoryParams{
		Symbol:   "BTCUSDT",
		Interval: "1h",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 || out[0].Close != 1.5 {
		t.Errorf("unexpected result: %+v", out)
	}
}
