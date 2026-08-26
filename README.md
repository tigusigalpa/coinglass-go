# CoinGlass Golang SDK

![CoinGlass Golang SDK](https://i.postimg.cc/D0RXFLYm/coinglass-golang-github.jpg)

> A Go client for the Coinglass API v4, with no external dependencies.

`coinglass-go` is a small, typed Go client for
the [Coinglass API v4](https://docs.coinglass.com/reference/getting-started-with-your-api). I built it to scratch
my own itch while working on liquidation dashboards and funding-rate tooling, and it covers the Futures, Spot,
Options, ETF, and Indicator endpoints. It leans entirely on the standard library, so adding it to your project won't
drag in a tree of transitive dependencies.

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-00ADD8?logo=go)](https://go.dev/)
[![Go Reference](https://pkg.go.dev/badge/github.com/tigusigalpa/coinglass-go.svg)](https://pkg.go.dev/github.com/tigusigalpa/coinglass-go)
[![CI](https://github.com/tigusigalpa/coinglass-go/actions/workflows/ci.yml/badge.svg)](https://github.com/tigusigalpa/coinglass-go/actions/workflows/ci.yml)
[![Codecov](https://codecov.io/gh/tigusigalpa/coinglass-go/graph/badge.svg)](https://codecov.io/gh/tigusigalpa/coinglass-go)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Documentation

- [Project Wiki](https://github.com/tigusigalpa/coinglass-go/wiki) — guides, endpoint notes, and usage details.
- [Go package reference](https://pkg.go.dev/github.com/tigusigalpa/coinglass-go) — API reference generated from source comments.
- [Coinglass API documentation](https://docs.coinglass.com/reference/getting-started-with-your-api) — API keys, plans, rate limits, and upstream endpoint specifications.

## Highlights

- Covers all five endpoint groups: Futures, Spot, Options, ETF, and Indicators.
- Includes a WebSocket client for real-time liquidation, trade, and futures ticker streams.
- No third-party dependencies — just the standard library, for both the REST and WebSocket clients.
- Configured with functional options, the way most Go clients do it.
- Every method takes a `context.Context`, and the client is safe to share across goroutines.
- Retries rate-limited (429) requests with exponential backoff, and honors `Retry-After` when the API sends it.
- Returns typed sentinel errors for `401`, `404`, and `429`, and a detailed `APIError` (with the API's own
  `code`/`msg`) for anything else.
- Each service ships with its own `httptest`-based tests.

## Requirements

| Requirement       | Version                               |
|-------------------|---------------------------------------|
| Go                | 1.21+                                 |
| Dependencies      | none (standard library only)          |
| Coinglass API key | [Get one here](https://coinglass.com) |

## Installation

```bash
go get github.com/tigusigalpa/coinglass-go
```

## Getting started

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    coinglass "github.com/tigusigalpa/coinglass-go"
)

func main() {
    ctx := context.Background()

    client := coinglass.NewClient("YOUR_API_KEY",
        coinglass.WithTimeout(15*time.Second),
        coinglass.WithRetry(3, time.Second), // 3 attempts, 1s initial backoff
    )

    // BTC open interest history (last 30 days, daily)
    oi, err := client.Futures.OpenInterestHistory(ctx, &coinglass.OIHistoryParams{
        Symbol:   "BTC",
        Interval: "1d",
        Limit:    coinglass.IntPtr(30),
    })
    if err != nil {
        log.Fatal(err)
    }
    for _, point := range oi {
        fmt.Printf("OI: %.2f USD at %d\n", point.OpenInterestUsd, point.Timestamp)
    }
}
```

Compile and run it:

```bash
go run main.go
```

## Configuring the client

`NewClient` takes functional options:

```go
client := coinglass.NewClient("YOUR_API_KEY",
    coinglass.WithBaseURL("https://open-api-v4.coinglass.com"), // default, shown for clarity
    coinglass.WithTimeout(15*time.Second),
    coinglass.WithRetry(3, time.Second),
    coinglass.WithHTTPClient(&http.Client{}), // custom transport, proxies, TLS, etc.
)
```

| Option                                                | What it does                                  | When to use it                                      |
|-------------------------------------------------------|-----------------------------------------------|-----------------------------------------------------|
| `WithBaseURL(url string)`                             | Overrides the API base URL.                   | Mocking the API in tests or using a custom gateway. |
| `WithHTTPClient(client *http.Client)`                 | Supplies your own HTTP client.                | Custom TLS, proxies, tracing, or middleware.        |
| `WithTimeout(d time.Duration)`                        | Sets the per-request timeout.                 | Default is 30s; lower it for fast UI endpoints.     |
| `WithRetry(maxAttempts int, baseDelay time.Duration)` | Retries on HTTP 429 with exponential backoff. | Strongly recommended for production workloads.      |

You can also read the API key from the `COINGLASS_API_KEY` environment variable:

```go
client, err := coinglass.NewClientFromEnv(coinglass.WithTimeout(15 * time.Second))
if err != nil {
    log.Fatal(err)
}
```

## Rate limits by plan

| Plan         | Requests/min |
|--------------|--------------|
| Hobbyist     | 30           |
| Startup      | 80           |
| Standard     | 300          |
| Professional | 1200         |

If you enable retries, the client backs off automatically and doubles the wait on each attempt (or uses the
`Retry-After` header when Coinglass sends one):

```go
client := coinglass.NewClient("YOUR_API_KEY",
    coinglass.WithRetry(3, time.Second), // 1s, then 2s, then 4s
)
```

## Full API reference

### Futures — `client.Futures`

| Method                                       | Endpoint                                                   | Description              |
|----------------------------------------------|------------------------------------------------------------|--------------------------|
| `SupportedCoins(ctx)`                        | `GET /futures/supported-coins`                             | Supported futures coins  |
| `SupportedExchangePairs(ctx, params)`        | `GET /api/futures/supported-exchange-pairs`                | Supported exchange pairs |
| `CoinsMarkets(ctx, params)`                  | `GET /api/futures/coins-markets`                           | Futures coin markets     |
| `PairsMarkets(ctx, params)`                  | `GET /api/futures/pairs-markets`                           | Futures pair markets     |
| `PriceChangeList(ctx)`                       | `GET /futures/price-change-list`                           | Price change list        |
| `OpenInterestHistory(ctx, params)`           | `GET /api/futures/openInterest/ohlc-history`               | OI OHLC history          |
| `OpenInterestAggregatedHistory(ctx, params)` | `GET /api/futures/openInterest/ohlc-aggregated-history`    | Aggregated OI OHLC       |
| `OpenInterestExchangeList(ctx, params)`      | `GET /api/futures/openInterest/exchange-list`              | OI by exchange           |
| `FundingRateHistory(ctx, params)`            | `GET /api/futures/fundingRate/ohlc-history`                | Funding rate OHLC        |
| `FundingRateOiWeighted(ctx, params)`         | `GET /api/futures/fundingRate/oi-weight-ohlc-history`      | OI-weighted funding rate |
| `FundingRateExchangeList(ctx, params)`       | `GET /api/futures/fundingRate/exchange-list`               | Funding rate by exchange |
| `FundingRateArbitrage(ctx, params)`          | `GET /api/futures/fundingRate/arbitrage`                   | Funding arbitrage        |
| `LongShortRatioHistory(ctx, params)`         | `GET /api/futures/global-long-short-account-ratio/history` | Global L/S ratio         |
| `TopLongShortRatioHistory(ctx, params)`      | `GET /api/futures/top-long-short-account-ratio/history`    | Top trader L/S ratio     |
| `LiquidationHistory(ctx, params)`            | `GET /api/futures/liquidation/history`                     | Pair liquidation history |
| `LiquidationAggregatedHistory(ctx, params)`  | `GET /api/futures/liquidation/aggregated-history`          | Coin liquidation history |
| `LiquidationCoinList(ctx, params)`           | `GET /api/futures/liquidation/coin-list`                   | Liquidation coin list    |
| `LiquidationHeatmap(ctx, model, params)`     | `GET /api/futures/liquidation/heatmap/model{1,2,3}`        | Liquidation heatmaps     |
| `LiquidationMap(ctx, params)`                | `GET /api/futures/liquidation/map`                         | Liquidation map          |
| `OrderbookHistory(ctx, params)`              | `GET /api/futures/orderbook/history`                       | Orderbook heatmap        |
| `LargeOrders(ctx, params)`                   | `GET /api/futures/orderbook/large-limit-order`             | Large orderbook orders   |
| `TakerBuySellHistory(ctx, params)`           | `GET /api/futures/taker-buy-sell-volume/history`           | Taker buy/sell history   |
| `WhaleAlert(ctx, params)`                    | `GET /api/hyperliquid/whale-alert`                         | Hyperliquid whale alert  |

### Spot — `client.Spot`

| Method                             | Endpoint                                      | Description            |
|------------------------------------|-----------------------------------------------|------------------------|
| `SupportedCoins(ctx)`              | `GET /api/spot/supported-coins`               | Supported coins        |
| `CoinsMarkets(ctx, params)`        | `GET /api/spot/coins-markets`                 | Coins markets          |
| `PairsMarkets(ctx, params)`        | `GET /api/spot/pairs-markets`                 | Pairs markets          |
| `PriceHistory(ctx, params)`        | `GET /api/spot/price/history`                 | Price OHLC history     |
| `OrderbookHistory(ctx, params)`    | `GET /api/spot/orderbook/history`             | Orderbook heatmap      |
| `TakerBuySellHistory(ctx, params)` | `GET /api/spot/taker-buy-sell-volume/history` | Taker buy/sell history |

### Options — `client.Options`

| Method                            | Endpoint                               | Description             |
|-----------------------------------|----------------------------------------|-------------------------|
| `MaxPain(ctx, params)`            | `GET /api/option/max-pain`             | Option max pain         |
| `Info(ctx, params)`               | `GET /api/option/info`                 | Options info            |
| `ExchangeOIHistory(ctx, params)`  | `GET /api/option/exchange-oi-history`  | Exchange OI history     |
| `ExchangeVolHistory(ctx, params)` | `GET /api/option/exchange-vol-history` | Exchange volume history |

### ETF — `client.ETF`

| Method                                 | Endpoint                                  | Description        |
|----------------------------------------|-------------------------------------------|--------------------|
| `BitcoinList(ctx)`                     | `GET /api/etf/bitcoin/list`               | Bitcoin ETF list   |
| `BitcoinFlowHistory(ctx, params)`      | `GET /api/etf/bitcoin/flow-history`       | BTC ETF flows      |
| `BitcoinNetAssetsHistory(ctx, params)` | `GET /api/etf/bitcoin/net-assets/history` | ETF net assets     |
| `EthereumList(ctx)`                    | `GET /api/etf/ethereum/list`              | Ethereum ETF list  |
| `EthereumFlowHistory(ctx, params)`     | `GET /api/etf/ethereum/flow-history`      | ETH ETF flows      |
| `GrayscaleHoldings(ctx)`               | `GET /api/grayscale/holdings-list`        | Grayscale holdings |

### Indicators — `client.Indicators`

| Method                             | Endpoint                                      | Description           |
|------------------------------------|-----------------------------------------------|-----------------------|
| `FearGreedHistory(ctx, params)`    | `GET /api/index/fear-greed-history`           | Fear & Greed index    |
| `RSIList(ctx, params)`             | `GET /api/futures/rsi/list`                   | RSI list              |
| `BasisHistory(ctx, params)`        | `GET /api/futures/basis/history`              | Futures basis         |
| `CoinbasePremium(ctx, params)`     | `GET /api/coinbase-premium-index`             | Coinbase premium      |
| `BitcoinRainbowChart(ctx)`         | `GET /api/index/bitcoin/rainbow-chart`        | BTC rainbow chart     |
| `StockToFlow(ctx)`                 | `GET /api/index/stock-flow`                   | Stock-to-Flow model   |
| `StablecoinMarketCap(ctx, params)` | `GET /api/index/stableCoin-marketCap-history` | Stablecoin market cap |

## Error handling

Any non-2xx response, or a response whose Coinglass envelope `code` isn't zero, comes back as an
`*coinglass.APIError`:

```go
type APIError struct {
    StatusCode int
    Code       string
    Message    string
    RawBody    []byte
}
```

For the common cases, match on the sentinel errors with `errors.Is`; when you need the details, pull out the
`*APIError` with `errors.As`:

```go
import "errors"

oi, err := client.Futures.OpenInterestHistory(ctx, params)
if err != nil {
    switch {
    case errors.Is(err, coinglass.ErrUnauthorized):
        log.Fatal("Invalid API key")
    case errors.Is(err, coinglass.ErrRateLimited):
        log.Println("Rate limited — retries exhausted")
    default:
        var apiErr *coinglass.APIError
        if errors.As(err, &apiErr) {
            log.Printf("API error %d (%s): %s", apiErr.StatusCode, apiErr.Code, apiErr.Message)
        }
    }
}
```

## WebSocket API

Alongside the REST client, `coinglass-go` ships a small WebSocket client under the `websocket` subpackage for
Coinglass's [real-time streams](https://docs.coinglass.com/reference/ws-getting-started) — liquidation orders, spot
and futures trades, and futures ticker snapshots. It's built entirely on the standard library too: the WebSocket
handshake, framing, and masking are implemented directly on top of `net`/`crypto/tls`, so no third-party dependency
is pulled in.

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    coinglass "github.com/tigusigalpa/coinglass-go"
    "github.com/tigusigalpa/coinglass-go/websocket"
)

func main() {
    client := coinglass.NewClient(os.Getenv("COINGLASS_API_KEY"))
    ws := client.WSClient()

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    stream, err := ws.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer stream.Close()

    stream.Subscribe(
        websocket.ChannelLiquidationOrders(),
        websocket.ChannelFuturesTicker("Binance", "BTCUSDT"),
    )

    for {
        select {
        case <-ctx.Done():
            return
        case msg, ok := <-stream.Messages():
            if !ok {
                return
            }
            if msg.Channel == websocket.ChannelLiquidationOrders() {
                orders, _ := websocket.DecodeLiquidationOrders(msg.Data)
                for _, o := range orders {
                    log.Printf("%s %s liquidated %.2f USD", o.Exchange, o.Symbol, o.VolumeUSD)
                }
            }
        case err := <-stream.Errors():
            log.Println("stream error:", err)
        }
    }
}
```

A single connection carries every subscription; `Subscribe`/`Unsubscribe` accept any number of channel names, and
the client sends the `"ping"` heartbeat every 20 seconds that Coinglass expects to keep the socket open.

### Channel helpers

| Helper                                              | Channel                              | Docs                                                                              |
|------------------------------------------------------|---------------------------------------|------------------------------------------------------------------------------------|
| `websocket.ChannelLiquidationOrders()`               | `liquidation_orders`                  | [Liquidation Order](https://docs.coinglass.com/reference/ws-liquidation-order)     |
| `websocket.ChannelSpotTrades(exchange, symbol, minVolumeUSD)`   | `spot_trades@{exchange}_{symbol}@{minVolumeUSD}`    | [Spot Trade Order](https://docs.coinglass.com/reference/websocket_spot_trades)     |
| `websocket.ChannelFuturesTrades(exchange, symbol, minVolumeUSD)` | `futures_trades@{exchange}_{symbol}@{minVolumeUSD}` | [Futures Trade Order](https://docs.coinglass.com/reference/websocket_futures_trades) |
| `websocket.ChannelFuturesTicker(exchange, symbol)`   | `futures_ticker@{exchange}_{symbol}`  | [Futures Ticker Snapshot](https://docs.coinglass.com/reference/websocket_futures_ticker) |

Each channel has a matching decode helper — `DecodeLiquidationOrders`, `DecodeTrades`, and `DecodeFuturesTicker` —
that unmarshals `Message.Data` into typed structs.

## Context and concurrency

Every method takes a context, and a single `Client` is safe to use from multiple goroutines:

```go
var wg sync.WaitGroup
for _, symbol := range []string{"BTC", "ETH", "SOL"} {
    wg.Add(1)
    go func(sym string) {
        defer wg.Done()
        oi, err := client.Futures.OpenInterestHistory(ctx, &coinglass.OIHistoryParams{
            Symbol:   sym,
            Interval: "1d",
        })
        if err != nil {
            log.Printf("%s failed: %v", sym, err)
            return
        }
        log.Printf("%s: %d points", sym, len(oi))
    }(symbol)
}
wg.Wait()
```

## Pointer helpers

Optional parameters are pointer fields, so the client can tell "not set" apart from a real zero value. These helpers
save you a few lines when passing literals:

```go
coinglass.IntPtr(30)
coinglass.StringPtr("BTC")
coinglass.BoolPtr(true)
coinglass.Int64Ptr(1690000000)
coinglass.Float64Ptr(1.5)
```

## Examples

There are a few runnable examples in the [`examples/`](examples) directory:

| Example                                        | What it shows                                                      |
|------------------------------------------------|--------------------------------------------------------------------|
| [`examples/basic`](examples/basic)             | Client setup, Futures/Spot/Options queries, error handling         |
| [`examples/etf`](examples/etf)                 | Bitcoin/Ethereum ETF flows, Grayscale holdings, Fear & Greed Index |
| [`examples/concurrency`](examples/concurrency) | Sharing a single `Client` safely across goroutines                 |
| [`examples/websocket`](examples/websocket)     | Subscribing to liquidation orders, trades, and futures ticker streams |

Run any of them with:

```bash
export COINGLASS_API_KEY=your-api-key
go run ./examples/basic
```

## Running the tests

```bash
go test ./...
```

Race detection requires `CGO_ENABLED=1` and a C toolchain:

```bash
go test -race ./...
```

The same checks run automatically on every push and pull request through GitHub Actions.

## License

MIT © [Igor Sazonov](https://github.com/tigusigalpa)
