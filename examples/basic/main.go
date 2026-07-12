// Package main demonstrates basic usage of the coinglass-go SDK: client
// initialization, Futures/Spot/Options queries, and error handling.
package main

import (
	"context"
	"errors"
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

	client := coinglass.NewClient(apiKey,
		coinglass.WithTimeout(15*time.Second),
		coinglass.WithRetry(3, time.Second),
	)

	ctx := context.Background()

	fmt.Println("=== coinglass-go - Basic Usage Examples ===")

	// 1. Futures: BTC open interest history (last 30 days, daily).
	fmt.Println("\n1. Futures Open Interest History:")
	oi, err := client.Futures.OpenInterestHistory(ctx, &coinglass.OIHistoryParams{
		Symbol:   "BTC",
		Interval: "1d",
		Limit:    coinglass.IntPtr(30),
	})
	if err != nil {
		log.Printf("   error: %v\n", err)
	} else {
		for _, point := range oi {
			fmt.Printf("   OI: %.2f USD at %d\n", point.OpenInterestUsd, point.Timestamp)
		}
	}

	// 2. Futures: funding rate by exchange.
	fmt.Println("\n2. Funding Rate by Exchange:")
	funding, err := client.Futures.FundingRateExchangeList(ctx, &coinglass.FundingRateExchangeListParams{
		Symbol: "BTC",
	})
	if err != nil {
		log.Printf("   error: %v\n", err)
	} else {
		for _, f := range funding {
			fmt.Printf("   %s: %.4f%%\n", f.Exchange, f.FundingRate*100)
		}
	}

	// 3. Spot: coin markets.
	fmt.Println("\n3. Spot Coin Markets:")
	markets, err := client.Spot.CoinsMarkets(ctx, &coinglass.SpotCoinsMarketsParams{
		Symbol: coinglass.StringPtr("BTC"),
	})
	if err != nil {
		log.Printf("   error: %v\n", err)
	} else {
		for _, m := range markets {
			fmt.Printf("   %s: $%.2f\n", m.Symbol, m.Price)
		}
	}

	// 4. Options: max pain.
	fmt.Println("\n4. Options Max Pain:")
	maxPain, err := client.Options.MaxPain(ctx, &coinglass.OptionParams{Underlying: "BTC"})
	if err != nil {
		log.Printf("   error: %v\n", err)
	} else {
		fmt.Printf("   Max pain: $%.2f\n", maxPain.MaxPain)
	}

	// 5. Error handling.
	fmt.Println("\n5. Error Handling:")
	_, err = client.Futures.OpenInterestHistory(ctx, &coinglass.OIHistoryParams{
		Symbol:   "INVALIDCOIN",
		Interval: "1d",
	})
	if err != nil {
		switch {
		case errors.Is(err, coinglass.ErrUnauthorized):
			fmt.Println("   Invalid API key")
		case errors.Is(err, coinglass.ErrRateLimited):
			fmt.Println("   Rate limited — retries exhausted")
		case errors.Is(err, coinglass.ErrNotFound):
			fmt.Println("   Resource not found")
		default:
			var apiErr *coinglass.APIError
			if errors.As(err, &apiErr) {
				fmt.Printf("   API error %d (%s): %s\n", apiErr.StatusCode, apiErr.Code, apiErr.Message)
			} else {
				fmt.Printf("   Error: %v\n", err)
			}
		}
	}

	fmt.Println("\n=== Examples Complete ===")
}
