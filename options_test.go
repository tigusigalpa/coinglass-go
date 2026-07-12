package coinglass

import (
	"context"
	"net/http"
	"testing"
)

func TestOptions_MaxPain(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"0","msg":"success","data":{"underlying":"BTC","expiry":123,"maxPain":50000}}`))
	})
	defer srv.Close()

	out, err := c.Options.MaxPain(context.Background(), &OptionParams{Underlying: "BTC"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.MaxPain != 50000 {
		t.Errorf("unexpected result: %+v", out)
	}
}

func TestOptions_ExchangeOIHistory(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"0","msg":"success","data":[{"exchange":"Deribit","totalOI":1000,"t":123}]}`))
	})
	defer srv.Close()

	out, err := c.Options.ExchangeOIHistory(context.Background(), &OptionHistoryParams{Interval: "1d"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 || out[0].Exchange != "Deribit" {
		t.Errorf("unexpected result: %+v", out)
	}
}
