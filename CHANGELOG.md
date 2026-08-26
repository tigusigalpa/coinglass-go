# Changelog

All notable changes to this project will be documented in this file.

## [v1.1.0] - 2026-07-12

### Added

- **WebSocket API support** for Coinglass's real-time streams, under the new `websocket` subpackage:
  - `websocket.Client` / `Client.Connect` — connects to `wss://open-ws.coinglass.com/ws-api`, authenticated with your
    API key. Accessible from the main SDK via `client.WSClient()`.
  - `Stream.Subscribe` / `Stream.Unsubscribe` — subscribe to any number of channels over a single connection, with
    the `"ping"`/`"pong"` heartbeat (every 20s) handled automatically.
  - `Stream.Messages()` / `Stream.Errors()` — channels for consuming incoming data and connection/parse errors.
  - Channel helpers and typed decoders for every documented stream:
    - `ChannelLiquidationOrders()` + `DecodeLiquidationOrders` —
      [Liquidation Order](https://docs.coinglass.com/reference/ws-liquidation-order)
    - `ChannelSpotTrades(exchange, symbol, minVolumeUSD)` + `DecodeTrades` —
      [Spot Trade Order](https://docs.coinglass.com/reference/websocket_spot_trades)
    - `ChannelFuturesTrades(exchange, symbol, minVolumeUSD)` + `DecodeTrades` —
      [Futures Trade Order](https://docs.coinglass.com/reference/websocket_futures_trades)
    - `ChannelFuturesTicker(exchange, symbol)` + `DecodeFuturesTicker` —
      [Futures Ticker Snapshot](https://docs.coinglass.com/reference/websocket_futures_ticker)
  - The WebSocket client is implemented entirely on the standard library (`net`, `crypto/tls`, `bufio`) — no
    third-party dependency (e.g. `gorilla/websocket`) was introduced, keeping the SDK dependency-free.
  - New runnable example: [`examples/websocket`](examples/websocket).
  - Unit tests covering frame encoding/decoding, the WebSocket handshake accept-key computation, channel builders,
    and message decoding (`websocket/frame_test.go`, `websocket/models_test.go`).

### Documentation

- Added a "WebSocket API" section to the README with a usage example and a channel-helper reference table.
- Updated the Highlights section and Examples table to mention WebSocket support.

## [v1.0.0] - Initial release

### Added

- Go SDK for the Coinglass API v4, covering Futures, Spot, Options, ETF, and Indicator endpoints.
- Functional-option configuration (`WithBaseURL`, `WithHTTPClient`, `WithTimeout`, `WithRetry`).
- Automatic retry with exponential backoff for HTTP 429 responses, honoring `Retry-After`.
- Typed sentinel errors (`ErrUnauthorized`, `ErrNotFound`, `ErrRateLimited`) and a detailed `APIError`.
- No third-party dependencies — standard library only.
