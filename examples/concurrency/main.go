// Package main demonstrates that a coinglass-go Client is safe to share
// across goroutines, fetching open-interest history for multiple symbols
// concurrently.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
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
	symbols := []string{"BTC", "ETH", "SOL"}

	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make(map[string]int)

	for _, symbol := range symbols {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()

			oi, err := client.Futures.OpenInterestHistory(ctx, &coinglass.OIHistoryParams{
				Symbol:   sym,
				Interval: "1d",
				Limit:    coinglass.IntPtr(30),
			})
			if err != nil {
				log.Printf("%s: error: %v\n", sym, err)
				return
			}

			mu.Lock()
			results[sym] = len(oi)
			mu.Unlock()
		}(symbol)
	}

	wg.Wait()

	fmt.Println("=== coinglass-go - Concurrency Example ===")
	for _, symbol := range symbols {
		fmt.Printf("%s: %d open-interest points fetched\n", symbol, results[symbol])
	}
}
