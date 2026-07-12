// Package websocket provides a client for the Coinglass real-time WebSocket
// API (https://docs.coinglass.com/reference/ws-getting-started). It is built
// entirely on the Go standard library — the SDK's "no third-party
// dependencies" guarantee also applies to the WebSocket client — and
// implements just enough of RFC 6455 (client-side handshake, masked text
// frames, and orderly close) to talk to the Coinglass server, which only
// ever sends unmasked text frames.
package websocket

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// DefaultBaseURL is the Coinglass WebSocket endpoint used by NewClient unless
// overridden with WithBaseURL.
const DefaultBaseURL = "wss://open-ws.coinglass.com/ws-api"

// DefaultHandshakeTimeout is the default timeout for the initial TCP/TLS
// connection and WebSocket handshake.
const DefaultHandshakeTimeout = 10 * time.Second

// DefaultPingInterval is how often the SDK sends the application-level
// "ping" text message recommended by the Coinglass docs to keep the
// connection alive.
const DefaultPingInterval = 20 * time.Second

// Client configures how connections to the Coinglass WebSocket API are
// established. It holds no connection state itself; call Connect to obtain a
// *Stream, which manages a single connection and its subscriptions. A Client
// is safe for concurrent use by multiple goroutines.
type Client struct {
	apiKey           string
	baseURL          string
	handshakeTimeout time.Duration
	pingInterval     time.Duration
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the default Coinglass WebSocket URL
// (wss://open-ws.coinglass.com/ws-api). This is primarily useful for testing
// against a mock server.
func WithBaseURL(rawURL string) Option {
	return func(c *Client) {
		if rawURL != "" {
			c.baseURL = rawURL
		}
	}
}

// WithHandshakeTimeout sets the timeout for the initial TCP/TLS connection
// and WebSocket handshake performed by Connect.
func WithHandshakeTimeout(d time.Duration) Option {
	return func(c *Client) {
		if d > 0 {
			c.handshakeTimeout = d
		}
	}
}

// WithPingInterval overrides how often the client sends the "ping" text
// message that keeps the connection alive. Coinglass recommends 20 seconds.
func WithPingInterval(d time.Duration) Option {
	return func(c *Client) {
		if d > 0 {
			c.pingInterval = d
		}
	}
}

// NewClient creates a WebSocket client authenticated with the given API key.
// The API key is sent as the cg-api-key query parameter on every connection,
// as required by the Coinglass WebSocket API.
func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:           apiKey,
		baseURL:          DefaultBaseURL,
		handshakeTimeout: DefaultHandshakeTimeout,
		pingInterval:     DefaultPingInterval,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Connect opens a single WebSocket connection to the Coinglass server and
// returns a *Stream for subscribing to channels and consuming messages. The
// returned Stream owns the connection until Close is called.
func (c *Client) Connect(ctx context.Context) (*Stream, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("coinglass/websocket: invalid base URL: %w", err)
	}

	if c.apiKey != "" {
		q := u.Query()
		q.Set("cg-api-key", c.apiKey)
		u.RawQuery = q.Encode()
	}

	dialCtx := ctx
	if c.handshakeTimeout > 0 {
		var cancel context.CancelFunc
		dialCtx, cancel = context.WithTimeout(ctx, c.handshakeTimeout)
		defer cancel()
	}

	conn, err := dial(dialCtx, u.String(), http.Header{})
	if err != nil {
		return nil, err
	}

	return newStream(c, conn), nil
}
