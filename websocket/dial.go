package websocket

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// websocketGUID is the magic value defined by RFC 6455 used to compute the
// Sec-WebSocket-Accept header from the client's Sec-WebSocket-Key.
const websocketGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// rawConn is a minimal RFC 6455 client-side WebSocket connection built purely
// on top of the standard library (net, crypto/tls, bufio). Coinglass uses
// plain text frames for both JSON payloads and the "ping"/"pong" heartbeat,
// so this implementation only needs to support text messages and orderly
// close handling.
type rawConn struct {
	conn net.Conn
	br   *bufio.Reader
}

// dial performs the TCP/TLS connection and WebSocket handshake against
// rawURL (which must have the ws:// or wss:// scheme) using ctx for the
// initial connection deadline.
func dial(ctx context.Context, rawURL string, headers http.Header) (*rawConn, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("coinglass/websocket: invalid URL: %w", err)
	}

	var tlsEnabled bool
	switch u.Scheme {
	case "wss":
		tlsEnabled = true
	case "ws":
		tlsEnabled = false
	default:
		return nil, fmt.Errorf("coinglass/websocket: unsupported scheme %q", u.Scheme)
	}

	host := u.Host
	if !strings.Contains(host, ":") {
		if tlsEnabled {
			host += ":443"
		} else {
			host += ":80"
		}
	}

	var netConn net.Conn
	dialer := &net.Dialer{}
	if tlsEnabled {
		tlsDialer := &tls.Dialer{
			NetDialer: dialer,
			Config:    &tls.Config{ServerName: u.Hostname()},
		}
		netConn, err = tlsDialer.DialContext(ctx, "tcp", host)
	} else {
		netConn, err = dialer.DialContext(ctx, "tcp", host)
	}
	if err != nil {
		return nil, fmt.Errorf("coinglass/websocket: dial failed: %w", err)
	}

	if deadline, ok := ctx.Deadline(); ok {
		_ = netConn.SetDeadline(deadline)
	}

	key, err := randomWebSocketKey()
	if err != nil {
		netConn.Close()
		return nil, err
	}

	requestPath := u.EscapedPath()
	if requestPath == "" {
		requestPath = "/"
	}
	if u.RawQuery != "" {
		requestPath += "?" + u.RawQuery
	}

	var sb strings.Builder
	sb.WriteString("GET " + requestPath + " HTTP/1.1\r\n")
	sb.WriteString("Host: " + u.Host + "\r\n")
	sb.WriteString("Upgrade: websocket\r\n")
	sb.WriteString("Connection: Upgrade\r\n")
	sb.WriteString("Sec-WebSocket-Key: " + key + "\r\n")
	sb.WriteString("Sec-WebSocket-Version: 13\r\n")
	for name, values := range headers {
		for _, v := range values {
			sb.WriteString(name + ": " + v + "\r\n")
		}
	}
	sb.WriteString("\r\n")

	if _, err := netConn.Write([]byte(sb.String())); err != nil {
		netConn.Close()
		return nil, fmt.Errorf("coinglass/websocket: failed to send handshake: %w", err)
	}

	br := bufio.NewReader(netConn)
	resp, err := http.ReadResponse(br, &http.Request{Method: "GET"})
	if err != nil {
		netConn.Close()
		return nil, fmt.Errorf("coinglass/websocket: failed to read handshake response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		netConn.Close()
		return nil, fmt.Errorf("coinglass/websocket: unexpected handshake status: %s", resp.Status)
	}
	if !strings.EqualFold(resp.Header.Get("Upgrade"), "websocket") || !headerHasToken(resp.Header.Get("Connection"), "upgrade") {
		netConn.Close()
		return nil, fmt.Errorf("coinglass/websocket: invalid upgrade response headers")
	}

	expectedAccept := computeAcceptKey(key)
	if resp.Header.Get("Sec-WebSocket-Accept") != expectedAccept {
		netConn.Close()
		return nil, fmt.Errorf("coinglass/websocket: invalid Sec-WebSocket-Accept header")
	}

	// Clear the handshake deadline; per-operation deadlines are set by the caller.
	_ = netConn.SetDeadline(zeroTime)

	return &rawConn{conn: netConn, br: br}, nil
}

func headerHasToken(value, token string) bool {
	for _, part := range strings.Split(value, ",") {
		if strings.EqualFold(strings.TrimSpace(part), token) {
			return true
		}
	}
	return false
}

func randomWebSocketKey() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("coinglass/websocket: failed to generate key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(raw[:]), nil
}

func computeAcceptKey(key string) string {
	h := sha1.New()
	h.Write([]byte(key + websocketGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
