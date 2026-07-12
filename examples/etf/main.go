// Package main demonstrates ETF and market Indicators usage of the
// coinglass-go SDK: Bitcoin/Ethereum ETF flows, Grayscale holdings, the
// Fear & Greed Index, and the Bitcoin rainbow chart.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	coinglass "github.com/tigusigalpa/coinglass-go"
)

func main() {
	apiKey := os.Getenv("COINGLASS_API_KEY")
	if apiKey == "" {
		log.Fatal("COINGLASS_API_KEY environment variable is required")
	}

	client := coinglass.NewClient(apiKey, coinglass.WithTimeout(15*time.Second))
	ctx := context.Background()

	fmt.Println("=== coinglass-go - ETF & Indicators Examples ===")

	// 1. Bitcoin ETF flow history.
	fmt.Println("\n1. Bitcoin ETF Flow History (weekly, last 24 points):")
	flows, err := client.ETF.BitcoinFlowHistory(ctx, &coinglass.ETFFlowParams{
		Interval: "1w",
		Limit:    coinglass.IntPtr(24),
	})
	if err != nil {
		log.Printf("   error: %v\n", err)
	} else {
		for _, f := range flows {
			fmt.Printf("   t=%d net flow: $%.2f\n", f.Timestamp, f.NetFlow)
		}
	}

	// 2. Bitcoin ETF list.
	fmt.Println("\n2. Bitcoin ETF List:")
	etfs, err := client.ETF.BitcoinList(ctx)
	if err != nil {
		log.Printf("   error: %v\n", err)
	} else {
		for _, e := range etfs {
			fmt.Printf("   %s (%s): %.2f holdings\n", e.Ticker, e.Name, e.Holdings)
		}
	}

	// 3. Grayscale holdings.
	fmt.Println("\n3. Grayscale Holdings:")
	holdings, err := client.ETF.GrayscaleHoldings(ctx)
	if err != nil {
		log.Printf("   error: %v\n", err)
	} else {
		for _, h := range holdings {
			fmt.Printf("   %s (%s): %.2f\n", h.Symbol, h.Asset, h.Holdings)
		}
	}

	// 4. Fear & Greed Index.
	fmt.Println("\n4. Fear & Greed Index (last 7 days):")
	fg, err := client.Indicators.FearGreedHistory(ctx, &coinglass.FearGreedParams{
		Limit: coinglass.IntPtr(7),
	})
	if err != nil {
		log.Printf("   error: %v\n", err)
	} else if len(fg) > 0 {
		fmt.Printf("   Latest: %d (%s)\n", fg[0].Value, fg[0].Classification)
	}

	// 5. Bitcoin rainbow chart.
	fmt.Println("\n5. Bitcoin Rainbow Chart:")
	rainbow, err := client.Indicators.BitcoinRainbowChart(ctx)
	if err != nil {
		log.Printf("   error: %v\n", err)
	} else {
		fmt.Printf("   Received %d bytes of chart data\n", len(rainbow.Data))
	}

	fmt.Println("\n=== Examples Complete ===")
}
