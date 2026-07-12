package coinglass

import (
	"context"
	"net/http"
	"testing"
)

func TestIndicators_FearGreedHistory(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"0","msg":"success","data":[{"value":72,"classification":"Greed","t":123}]}`))
	})
	defer srv.Close()

	out, err := c.Indicators.FearGreedHistory(context.Background(), &FearGreedParams{Limit: IntPtr(7)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 || out[0].Classification != "Greed" {
		t.Errorf("unexpected result: %+v", out)
	}
}

func TestIndicators_RSIList(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"0","msg":"success","data":[{"symbol":"BTC","interval":"1h","rsi":65.4,"t":123}]}`))
	})
	defer srv.Close()

	out, err := c.Indicators.RSIList(context.Background(), &RSIListParams{Symbol: StringPtr("BTC")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 || out[0].RSI != 65.4 {
		t.Errorf("unexpected result: %+v", out)
	}
}

func TestIndicators_BitcoinRainbowChart(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"0","msg":"success","data":{"data":[1,2,3]}}`))
	})
	defer srv.Close()

	out, err := c.Indicators.BitcoinRainbowChart(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil result")
	}
}
