// Package coinglass provides a Go SDK for the Coinglass API v4
// (https://open-api-v4.coinglass.com). It exposes service structs grouped by
// resource (Futures, Spot, Options, ETF, Indicators) accessible as fields on Client.
package coinglass

import (
	"net/http"
	"time"
)

// DefaultBaseURL is the default Coinglass API v4 base URL used by
// NewClient unless overridden with WithBaseURL.
const DefaultBaseURL = "https://open-api-v4.coinglass.com"

// DefaultTimeout is the default HTTP client timeout used unless
// overridden with WithTimeout.
const DefaultTimeout = 30 * time.Second

// DefaultMaxAttempts is the default number of attempts (including the
// initial request) made for a request before giving up when receiving
// HTTP 429 responses, unless overridden with WithRetry.
const DefaultMaxAttempts = 1

// DefaultBaseDelay is the default initial backoff delay used for retry
// attempts, unless overridden with WithRetry.
const DefaultBaseDelay = 500 * time.Millisecond

// Client is the entry point of the Coinglass Go SDK. It holds the HTTP
// configuration and exposes one service struct per API resource group.
// A Client is safe for concurrent use by multiple goroutines.
type Client struct {
	apiKey      string
	baseURL     string
	httpClient  *http.Client
	maxAttempts int
	baseDelay   time.Duration

	// Futures provides access to the futures endpoints.
	Futures *FuturesService
	// Spot provides access to the spot endpoints.
	Spot *SpotService
	// Options provides access to the options endpoints.
	Options *OptionsService
	// ETF provides access to the ETF endpoints.
	ETF *ETFService
	// Indicators provides access to the indicator endpoints.
	Indicators *IndicatorsService
}

// NewClient creates a new Coinglass API client authenticated with the
// given apiKey. The apiKey is sent in the CG-API-KEY header on every
// request. Behaviour can be customized via Option values such as
// WithBaseURL, WithHTTPClient, WithTimeout, and WithRetry.
//
// Example:
//
//	client := coinglass.NewClient("YOUR_API_KEY",
//	    coinglass.WithTimeout(15*time.Second),
//	    coinglass.WithRetry(3, time.Second),
//	)
func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:      apiKey,
		baseURL:     DefaultBaseURL,
		httpClient:  &http.Client{Timeout: DefaultTimeout},
		maxAttempts: DefaultMaxAttempts,
		baseDelay:   DefaultBaseDelay,
	}

	for _, opt := range opts {
		opt(c)
	}

	c.Futures = &FuturesService{client: c}
	c.Spot = &SpotService{client: c}
	c.Options = &OptionsService{client: c}
	c.ETF = &ETFService{client: c}
	c.Indicators = &IndicatorsService{client: c}

	return c
}

// NewClientFromEnv reads the API key from the COINGLASS_API_KEY
// environment variable and returns a new Coinglass client. Additional
// Option values can be passed to override timeouts, retries, etc.
func NewClientFromEnv(opts ...Option) (*Client, error) {
	apiKey := ""
	if v, ok := lookupEnv("COINGLASS_API_KEY"); ok {
		apiKey = v
	}
	if apiKey == "" {
		return nil, ErrUnauthorized
	}
	return NewClient(apiKey, opts...), nil
}

// lookupEnv wraps os.LookupEnv so it can be swapped in tests if needed.
func lookupEnv(key string) (string, bool) {
	return defaultLookupEnv(key)
}

var defaultLookupEnv = func(key string) (string, bool) {
	return "", false
}
