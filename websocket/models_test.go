package websocket

import (
	"encoding/json"
	"testing"
)

func TestChannelBuilders(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"liquidation_orders", ChannelLiquidationOrders(), "liquidation_orders"},
		{"spot_trades", ChannelSpotTrades("Binance", "BTCUSDT", 10000), "spot_trades@Binance_BTCUSDT@10000"},
		{"futures_trades", ChannelFuturesTrades("Binance", "BTCUSDT", 10000), "futures_trades@Binance_BTCUSDT@10000"},
		{"futures_ticker", ChannelFuturesTicker("Binance", "BTCUSDT"), "futures_ticker@Binance_BTCUSDT"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("got %q, want %q", tc.got, tc.want)
			}
		})
	}
}

func TestDecodeLiquidationOrders(t *testing.T) {
	raw := json.RawMessage(`[{"base_asset":"BTC","exchange":"Binance","price":56738.00,"side":2,"symbol":"BTCUSDT","time":1725416318379,"volume_usd":3858.184}]`)

	orders, err := DecodeLiquidationOrders(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}
	o := orders[0]
	if o.BaseAsset != "BTC" || o.Exchange != "Binance" || o.Symbol != "BTCUSDT" {
		t.Fatalf("unexpected fields: %+v", o)
	}
	if o.Side != LiquidationSideShort {
		t.Fatalf("expected short side, got %v", o.Side)
	}
	if o.Price != 56738.00 {
		t.Fatalf("unexpected price: %v", o.Price)
	}
}

func TestDecodeTrades(t *testing.T) {
	raw := json.RawMessage(`[{"base_asset":"BTC","exchange":"Binance","price":56738.00,"side":2,"symbol":"BTCUSDT","time":1725416318379,"volume_usd":3858.184}]`)

	trades, err := DecodeTrades(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}
	if trades[0].Side != TradeSideBuy {
		t.Fatalf("expected buy side, got %v", trades[0].Side)
	}
}

func TestDecodeFuturesTicker(t *testing.T) {
	raw := json.RawMessage(`[{"exchange":"Binance","symbol":"BTCUSDT","base_asset":"BTC","price":62850.45,"index_price":62847.32,"volume_usd_24h":1256789345.67,"open_interest":5293456789.12,"open_interest_amount":4589.12,"funding_rate":0.0001,"next_funding_time":"2026-04-29 16:00:00","funding_interval_hours":8,"expiry_date":null,"update_time":"2026-04-29 08:20:00"}]`)

	tickers, err := DecodeFuturesTicker(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tickers) != 1 {
		t.Fatalf("expected 1 ticker, got %d", len(tickers))
	}
	tk := tickers[0]
	if tk.Exchange != "Binance" || tk.Symbol != "BTCUSDT" {
		t.Fatalf("unexpected fields: %+v", tk)
	}
	if tk.ExpiryDate != nil {
		t.Fatalf("expected nil expiry date, got %v", *tk.ExpiryDate)
	}
	if tk.FundingIntervalHours != 8 {
		t.Fatalf("unexpected funding interval: %d", tk.FundingIntervalHours)
	}
}

func TestParseMessage(t *testing.T) {
	data := []byte(`{"channel":"liquidation_orders","data":[{"base_asset":"BTC"}]}`)
	msg, ok := parseMessage(data)
	if !ok {
		t.Fatal("expected message to parse")
	}
	if msg.Channel != "liquidation_orders" {
		t.Fatalf("unexpected channel: %s", msg.Channel)
	}

	if _, ok := parseMessage([]byte("pong")); ok {
		t.Fatal("expected non-JSON payload to fail parsing")
	}

	if _, ok := parseMessage([]byte(`{"foo":"bar"}`)); ok {
		t.Fatal("expected message without channel field to be rejected")
	}
}
