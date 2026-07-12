package coinglass

import (
	"context"
	"net/http"
	"testing"
)

func TestETF_BitcoinList(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"0","msg":"success","data":[{"ticker":"IBIT","name":"iShares Bitcoin Trust","holdings":500000}]}`))
	})
	defer srv.Close()

	out, err := c.ETF.BitcoinList(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 || out[0].Ticker != "IBIT" {
		t.Errorf("unexpected result: %+v", out)
	}
}

func TestETF_BitcoinFlowHistory(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("interval"); got != "1w" {
			t.Errorf("expected interval=1w, got %s", got)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"0","msg":"success","data":[{"t":123,"netFlow":1000,"totalInflow":1500,"totalOutflow":500}]}`))
	})
	defer srv.Close()

	out, err := c.ETF.BitcoinFlowHistory(context.Background(), &ETFFlowParams{
		Interval: "1w",
		Limit:    IntPtr(24),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 || out[0].NetFlow != 1000 {
		t.Errorf("unexpected result: %+v", out)
	}
}

func TestETF_GrayscaleHoldings(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"0","msg":"success","data":[{"symbol":"GBTC","asset":"BTC","holdings":300000,"valueUsd":15000000000}]}`))
	})
	defer srv.Close()

	out, err := c.ETF.GrayscaleHoldings(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 || out[0].Symbol != "GBTC" {
		t.Errorf("unexpected result: %+v", out)
	}
}
