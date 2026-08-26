package coinglass

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// newTestClient spins up an httptest.Server using handler and returns a
// *Client configured to talk to it, along with the server for cleanup.
func newTestClient(t *testing.T, handler http.HandlerFunc, opts ...Option) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	allOpts := append([]Option{WithBaseURL(srv.URL)}, opts...)
	c := NewClient("test-api-key", allOpts...)
	return c, srv
}

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient("key")
	if c.baseURL != DefaultBaseURL {
		t.Errorf("expected baseURL %s, got %s", DefaultBaseURL, c.baseURL)
	}
	if c.httpClient.Timeout != DefaultTimeout {
		t.Errorf("expected timeout %s, got %s", DefaultTimeout, c.httpClient.Timeout)
	}
	if c.maxAttempts != DefaultMaxAttempts {
		t.Errorf("expected maxAttempts %d, got %d", DefaultMaxAttempts, c.maxAttempts)
	}
	if c.Futures == nil || c.Spot == nil || c.Options == nil || c.ETF == nil || c.Indicators == nil {
		t.Error("expected all services to be initialized")
	}
}

func TestNewClient_Options(t *testing.T) {
	customClient := &http.Client{}
	c := NewClient("key",
		WithBaseURL("https://example.com"),
		WithHTTPClient(customClient),
		WithTimeout(5*time.Second),
		WithRetry(3, 100*time.Millisecond),
	)
	if c.baseURL != "https://example.com" {
		t.Errorf("expected custom baseURL, got %s", c.baseURL)
	}
	if c.httpClient != customClient {
		t.Error("expected custom http client to be used")
	}
	if c.httpClient.Timeout != 5*time.Second {
		t.Errorf("expected timeout 5s, got %s", c.httpClient.Timeout)
	}
	if c.maxAttempts != 3 {
		t.Errorf("expected maxAttempts 3, got %d", c.maxAttempts)
	}
	if c.baseDelay != 100*time.Millisecond {
		t.Errorf("expected baseDelay 100ms, got %s", c.baseDelay)
	}
}

func TestWithBaseURL_TrimsTrailingSlashes(t *testing.T) {
	c := NewClient("key", WithBaseURL("https://example.com/api///"))
	if c.baseURL != "https://example.com/api" {
		t.Errorf("expected normalized baseURL, got %s", c.baseURL)
	}
}

func TestGet_Retry429(t *testing.T) {
	var attempts int
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"message":"rate limited"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"0","msg":"success","data":["BTC","ETH"]}`))
	}, WithRetry(5, 10*time.Millisecond))
	defer srv.Close()

	coins, err := c.Futures.SupportedCoins(context.Background())
	if err != nil {
		t.Fatalf("expected no error after retries, got %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
	if len(coins) != 2 || coins[0] != "BTC" {
		t.Errorf("unexpected coins: %v", coins)
	}
}

func TestGet_RateLimitedExhausted(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"message":"rate limited"}`))
	}, WithRetry(2, 5*time.Millisecond))
	defer srv.Close()

	_, err := c.Futures.SupportedCoins(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("expected ErrRateLimited, got %v", err)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Errorf("expected *APIError, got %T", err)
	} else if apiErr.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", apiErr.StatusCode)
	}
}

func TestGet_Unauthorized(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"invalid api key"}`))
	})
	defer srv.Close()

	_, err := c.Futures.SupportedCoins(context.Background())
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestGet_NotFound(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	})
	defer srv.Close()

	_, err := c.Futures.SupportedCoins(context.Background())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGet_ContextCancellation(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"0","msg":"success","data":[]}`))
	})
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := c.Futures.SupportedCoins(ctx)
	if err == nil {
		t.Fatal("expected context deadline error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestGet_AuthorizationHeader(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("CG-API-KEY"); got != "test-api-key" {
			t.Errorf("expected CG-API-KEY header 'test-api-key', got %q", got)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"0","msg":"success","data":[]}`))
	})
	defer srv.Close()

	if _, err := c.Futures.SupportedCoins(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGet_NonZeroEnvelopeCode(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"30001","msg":"invalid parameter","data":null}`))
	})
	defer srv.Close()

	_, err := c.Futures.SupportedCoins(context.Background())
	if err == nil {
		t.Fatal("expected error for non-zero envelope code, got nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Code != "30001" {
		t.Errorf("expected code 30001, got %s", apiErr.Code)
	}
}

func TestAPIError_Error(t *testing.T) {
	err := &APIError{StatusCode: 500, Code: "1", Message: "server error"}
	want := "coinglass: api error: status=500 code=1 message=server error"
	if err.Error() != want {
		t.Errorf("expected %q, got %q", want, err.Error())
	}
}

func TestClient_ConcurrentUse(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":"0","msg":"success","data":["BTC"]}`))
	})
	defer srv.Close()

	var wg sync.WaitGroup
	errCh := make(chan error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := c.Futures.SupportedCoins(context.Background())
			if err != nil {
				errCh <- err
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Errorf("unexpected concurrent error: %v", err)
	}
}
