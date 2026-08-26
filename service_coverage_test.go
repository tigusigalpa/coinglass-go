package coinglass

import (
	"context"
	"net/http"
	"testing"
)

// TestServiceEndpoints exercises every public endpoint wrapper. Individual
// response-shape tests live beside their services; this table ensures every
// wrapper sends its request through the shared transport.
func TestServiceEndpoints(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"0","msg":"success","data":null}`))
	})
	defer srv.Close()

	ctx := context.Background()
	limit := IntPtr(1)
	interval := "1h"
	start := Int64Ptr(1)
	end := Int64Ptr(2)
	exchange := StringPtr("Binance")
	symbol := StringPtr("BTC")

	tests := []struct {
		name string
		call func() error
	}{
		{"futures supported coins", func() error { _, err := c.Futures.SupportedCoins(ctx); return err }},
		{"futures exchange pairs", func() error {
			_, err := c.Futures.SupportedExchangePairs(ctx, &SupportedExchangePairsParams{Exchange: exchange})
			return err
		}},
		{"futures coins markets", func() error {
			_, err := c.Futures.CoinsMarkets(ctx, &CoinsMarketsParams{Symbol: symbol, Exchanges: []string{"Binance"}, Limit: limit})
			return err
		}},
		{"futures pairs markets", func() error {
			_, err := c.Futures.PairsMarkets(ctx, &PairsMarketsParams{Symbol: symbol, Exchange: exchange, Limit: limit})
			return err
		}},
		{"futures price changes", func() error { _, err := c.Futures.PriceChangeList(ctx); return err }},
		{"futures oi aggregate", func() error {
			_, err := c.Futures.OpenInterestAggregatedHistory(ctx, &OIHistoryParams{Symbol: "BTC", Interval: interval, Limit: limit, StartTime: start, EndTime: end})
			return err
		}},
		{"futures oi exchanges", func() error {
			_, err := c.Futures.OpenInterestExchangeList(ctx, &OIExchangeListParams{Symbol: "BTC", Interval: interval, Exchange: exchange})
			return err
		}},
		{"futures funding history", func() error {
			_, err := c.Futures.FundingRateHistory(ctx, &FundingRateHistoryParams{Symbol: "BTC", Interval: interval})
			return err
		}},
		{"futures funding exchanges", func() error {
			_, err := c.Futures.FundingRateExchangeList(ctx, &FundingRateExchangeListParams{Symbol: "BTC", Interval: interval, Exchange: exchange})
			return err
		}},
		{"futures weighted funding", func() error {
			_, err := c.Futures.FundingRateOiWeighted(ctx, &FundingRateHistoryParams{Symbol: "BTC", Interval: interval})
			return err
		}},
		{"futures arbitrage", func() error {
			_, err := c.Futures.FundingRateArbitrage(ctx, &FundingRateArbitrageParams{Symbol: symbol, Interval: &interval, Limit: limit})
			return err
		}},
		{"futures long short", func() error {
			_, err := c.Futures.LongShortRatioHistory(ctx, &LongShortRatioParams{Symbol: "BTC", Interval: interval, Exchange: exchange})
			return err
		}},
		{"futures top long short", func() error {
			_, err := c.Futures.TopLongShortRatioHistory(ctx, &LongShortRatioParams{Symbol: "BTC", Interval: interval, Exchange: exchange})
			return err
		}},
		{"futures liquidations", func() error {
			_, err := c.Futures.LiquidationHistory(ctx, &LiquidationHistoryParams{Symbol: "BTC", Pair: "BTCUSDT", Interval: interval})
			return err
		}},
		{"futures aggregate liquidations", func() error {
			_, err := c.Futures.LiquidationAggregatedHistory(ctx, &LiquidationAggregatedHistoryParams{Symbol: "BTC", Interval: interval})
			return err
		}},
		{"futures liquidation coins", func() error {
			_, err := c.Futures.LiquidationCoinList(ctx, &LiquidationCoinListParams{Symbol: symbol, Limit: limit})
			return err
		}},
		{"futures liquidation map", func() error {
			_, err := c.Futures.LiquidationMap(ctx, &LiquidationMapParams{Symbol: "BTC", Pair: "BTCUSDT", Interval: interval})
			return err
		}},
		{"futures orderbook", func() error {
			_, err := c.Futures.OrderbookHistory(ctx, &OrderbookHistoryParams{Symbol: "BTC", Exchange: "Binance", Interval: interval})
			return err
		}},
		{"futures large orders", func() error {
			_, err := c.Futures.LargeOrders(ctx, &LargeOrdersParams{Symbol: "BTC", Exchange: "Binance", Interval: interval})
			return err
		}},
		{"futures taker volume", func() error {
			_, err := c.Futures.TakerBuySellHistory(ctx, &TakerBuySellHistoryParams{Symbol: "BTC", Exchange: "Binance", Interval: interval})
			return err
		}},
		{"spot supported coins", func() error { _, err := c.Spot.SupportedCoins(ctx); return err }},
		{"spot pairs", func() error {
			_, err := c.Spot.PairsMarkets(ctx, &SpotPairsMarketsParams{Symbol: symbol, Exchange: exchange, Limit: limit})
			return err
		}},
		{"spot orderbook", func() error {
			_, err := c.Spot.OrderbookHistory(ctx, &SpotOrderbookHistoryParams{Symbol: "BTC", Exchange: "Binance", Interval: interval})
			return err
		}},
		{"spot taker volume", func() error {
			_, err := c.Spot.TakerBuySellHistory(ctx, &SpotTakerBuySellHistoryParams{Symbol: "BTC", Exchange: "Binance", Interval: interval})
			return err
		}},
		{"option info", func() error { _, err := c.Options.Info(ctx, &OptionParams{Underlying: "BTC"}); return err }},
		{"option volume", func() error {
			_, err := c.Options.ExchangeVolHistory(ctx, &OptionHistoryParams{Interval: interval})
			return err
		}},
		{"indicator basis", func() error {
			_, err := c.Indicators.BasisHistory(ctx, &BasisHistoryParams{Symbol: "BTC", Interval: interval})
			return err
		}},
		{"indicator premium", func() error {
			_, err := c.Indicators.CoinbasePremium(ctx, &CoinbasePremiumParams{Limit: limit})
			return err
		}},
		{"indicator stock to flow", func() error { _, err := c.Indicators.StockToFlow(ctx); return err }},
		{"indicator stablecoin market cap", func() error {
			_, err := c.Indicators.StablecoinMarketCap(ctx, &StablecoinMarketCapParams{Limit: limit})
			return err
		}},
		{"etf bitcoin assets", func() error {
			_, err := c.ETF.BitcoinNetAssetsHistory(ctx, &ETFNetAssetsParams{Interval: interval, Limit: limit})
			return err
		}},
		{"etf ethereum list", func() error { _, err := c.ETF.EthereumList(ctx); return err }},
		{"etf ethereum flow", func() error {
			_, err := c.ETF.EthereumFlowHistory(ctx, &ETFFlowParams{Interval: interval, Limit: limit})
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
