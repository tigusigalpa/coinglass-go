package websocket

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientOptionsAndConnect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cg-api-key") != "key" {
			t.Errorf("missing API key")
		}
		h, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("response writer cannot be hijacked")
		}
		conn, rw, err := h.Hijack()
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		key := r.Header.Get("Sec-WebSocket-Key")
		_, _ = fmt.Fprintf(rw, "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: %s\r\n\r\n", computeAcceptKey(key))
		_ = rw.Flush()
		_, _ = readFrame(bufio.NewReader(conn))
	}))
	defer server.Close()

	baseURL := "ws" + strings.TrimPrefix(server.URL, "http")
	c := NewClient("key", WithBaseURL(baseURL), WithHandshakeTimeout(time.Second), WithPingInterval(time.Hour))
	if c.baseURL != baseURL || c.handshakeTimeout != time.Second || c.pingInterval != time.Hour {
		t.Fatal("client options were not applied")
	}
	s, err := c.Connect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestConnectRejectsInvalidURL(t *testing.T) {
	_, err := NewClient("", WithBaseURL("://bad")).Connect(context.Background())
	if err == nil {
		t.Fatal("expected invalid URL error")
	}
	if _, err := dial(context.Background(), "http://example.com", nil); err == nil {
		t.Fatal("expected unsupported scheme error")
	}
}

func TestRawConnOperations(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()
	raw := &rawConn{conn: client, br: bufio.NewReader(client)}

	done := make(chan error, 1)
	go func() {
		fr, err := readFrame(bufio.NewReader(server))
		if err == nil && (fr.opcode != opText || string(fr.payload) != "hello") {
			err = fmt.Errorf("unexpected frame: %#v", fr)
		}
		done <- err
	}()
	if err := raw.WriteText([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if err := raw.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := raw.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := raw.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRawConnReadMessageSkipsControlFrames(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()
	raw := &rawConn{conn: client, br: bufio.NewReader(client)}

	go func() {
		_, _ = server.Write([]byte{0x80 | opPing, 1, 'x', 0x80 | opText, 2, 'o', 'k'})
		_, _ = readFrame(bufio.NewReader(server))
	}()
	opcode, payload, err := raw.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	if opcode != opText || string(payload) != "ok" {
		t.Fatalf("got opcode=%d payload=%q", opcode, payload)
	}
}

func TestStreamSubscriptionsAndErrors(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()
	s := newStream(&Client{pingInterval: time.Hour}, &rawConn{conn: client, br: bufio.NewReader(client)})

	frames := make(chan frame, 2)
	go func() {
		br := bufio.NewReader(server)
		for i := 0; i < 2; i++ {
			fr, err := readFrame(br)
			if err != nil {
				return
			}
			frames <- fr
		}
	}()
	if err := s.Subscribe("one", "two"); err != nil {
		t.Fatal(err)
	}
	if got := s.Subscribed(); len(got) != 2 {
		t.Fatalf("subscriptions: %v", got)
	}
	if err := s.Unsubscribe("one"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if fr := <-frames; fr.opcode != opText {
			t.Fatalf("unexpected opcode %d", fr.opcode)
		}
	}
	s.pushError(fmt.Errorf("test error"))
	if err := <-s.Errors(); err == nil {
		t.Fatal("expected pushed error")
	}
	_ = server.Close()
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestWriteFrameExtendedPayload(t *testing.T) {
	for _, payload := range [][]byte{make([]byte, 126), make([]byte, 65536)} {
		var buf bytes.Buffer
		if err := writeFrame(&buf, opText, payload); err != nil {
			t.Fatal(err)
		}
		if _, err := readFrame(bufio.NewReader(&buf)); err != nil {
			t.Fatal(err)
		}
	}
}
