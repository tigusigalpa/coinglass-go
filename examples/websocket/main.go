// Package main demonstrates the coinglass-go WebSocket client: connecting,
// subscribing to the liquidation orders, spot trades, futures trades, and
// futures ticker channels, and decoding incoming messages.
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
	apiKey := os.Getenv("COINGLASS_API_KEY")
	if apiKey == "" {
		log.Fatal("COINGLASS_API_KEY environment variable is required")
	}

	client := coinglass.NewClient(apiKey)
	ws := client.WSClient()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	stream, err := ws.Connect(ctx)
	if err != nil {
		log.Fatalf("connect failed: %v", err)
	}
	defer stream.Close()

	channels := []string{
		websocket.ChannelLiquidationOrders(),
		websocket.ChannelSpotTrades("Binance", "BTCUSDT", 10000),
		websocket.ChannelFuturesTrades("Binance", "BTCUSDT", 10000),
		websocket.ChannelFuturesTicker("Binance", "BTCUSDT"),
	}

	if err := stream.Subscribe(channels...); err != nil {
		log.Fatalf("subscribe failed: %v", err)
	}

	log.Printf("subscribed to: %v", stream.Subscribed())

	for {
		select {
		case <-ctx.Done():
			log.Println("shutting down")
			return

		case msg, ok := <-stream.Messages():
			if !ok {
				log.Println("stream closed")
				return
			}
			handleMessage(msg)

		case err, ok := <-stream.Errors():
			if !ok {
				continue
			}
			log.Printf("stream error: %v", err)
		}
	}
}

func handleMessage(msg websocket.Message) {
	switch {
	case msg.Channel == websocket.ChannelLiquidationOrders():
		orders, err := websocket.DecodeLiquidationOrders(msg.Data)
		if err != nil {
			log.Printf("decode liquidation orders failed: %v", err)
			return
		}
		for _, o := range orders {
			log.Printf("[liquidation] %s %s price=%.2f volUSD=%.2f side=%v",
				o.Exchange, o.Symbol, o.Price, o.VolumeUSD, o.Side)
		}

	case len(msg.Channel) >= 12 && msg.Channel[:12] == "spot_trades@",
		len(msg.Channel) >= 15 && msg.Channel[:15] == "futures_trades@":
		trades, err := websocket.DecodeTrades(msg.Data)
		if err != nil {
			log.Printf("decode trades failed: %v", err)
			return
		}
		for _, t := range trades {
			log.Printf("[trade] channel=%s %s %s price=%.2f volUSD=%.2f side=%v",
				msg.Channel, t.Exchange, t.Symbol, t.Price, t.VolumeUSD, t.Side)
		}

	case len(msg.Channel) >= 14 && msg.Channel[:14] == "futures_ticker":
		tickers, err := websocket.DecodeFuturesTicker(msg.Data)
		if err != nil {
			log.Printf("decode futures ticker failed: %v", err)
			return
		}
		for _, t := range tickers {
			log.Printf("[ticker] %s %s price=%.2f oi=%.2f fundingRate=%.6f",
				t.Exchange, t.Symbol, t.Price, t.OpenInterest, t.FundingRate)
		}

	default:
		log.Printf("[unhandled] channel=%s data=%s", msg.Channel, string(msg.Data))
	}
}
