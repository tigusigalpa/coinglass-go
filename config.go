package coinglass

import (
	"net/http"
	"os"
	"strings"
	"time"
)

// Option configures a Client. Options are applied in the order they are
// passed to NewClient.
type Option func(*Client)

// WithBaseURL overrides the default Coinglass API base URL
// (https://open-api-v4.coinglass.com). This is primarily useful for testing
// against a mock server.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		if url != "" {
			c.baseURL = strings.TrimRight(url, "/")
		}
	}
}

// WithHTTPClient sets a custom *http.Client used to perform requests.
// This allows callers to configure transport-level behaviour such as
// proxies, TLS settings, or custom RoundTrippers.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		if client != nil {
			c.httpClient = client
		}
	}
}

// WithTimeout sets the timeout applied to the underlying HTTP client for
// every request made by the Client.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		if d > 0 {
			c.httpClient.Timeout = d
		}
	}
}

// WithRetry configures automatic retry behaviour for HTTP 429 (rate
// limited) responses. maxAttempts is the total number of attempts
// (including the first one) and baseDelay is the initial backoff delay
// used for exponential backoff (delay doubles on each subsequent retry
// unless a Retry-After header is present, in which case that value takes
// precedence).
func WithRetry(maxAttempts int, baseDelay time.Duration) Option {
	return func(c *Client) {
		if maxAttempts > 0 {
			c.maxAttempts = maxAttempts
		}
		if baseDelay > 0 {
			c.baseDelay = baseDelay
		}
	}
}

// init sets up the real environment lookup once at package load time.
func init() {
	defaultLookupEnv = os.LookupEnv
}
