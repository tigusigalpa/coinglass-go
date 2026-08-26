package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrStreamClosed is returned by Subscribe/Unsubscribe/Send when called on a
// Stream that has already been closed.
var ErrStreamClosed = errors.New("coinglass/websocket: stream is closed")

// Stream represents a single, live WebSocket connection to the Coinglass
// server. It manages the application-level ping heartbeat, dispatches
// incoming messages, and tracks subscribed channels so subscription state is
// easy to inspect. A Stream is safe for concurrent use by multiple
// goroutines, but it is not safe to call Connect again on the same Stream;
// create a new one via Client.Connect instead.
type Stream struct {
	client *Client
	conn   *rawConn

	ctx    context.Context
	cancel context.CancelFunc

	msgCh chan Message
	errCh chan error

	sendMu sync.Mutex

	subMu sync.RWMutex
	subs  map[string]struct{}

	closedMu sync.Mutex
	closed   bool

	loops sync.WaitGroup
}

func newStream(client *Client, conn *rawConn) *Stream {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Stream{
		client: client,
		conn:   conn,
		ctx:    ctx,
		cancel: cancel,
		msgCh:  make(chan Message, 256),
		errCh:  make(chan error, 16),
		subs:   make(map[string]struct{}),
	}

	s.loops.Add(1)
	go s.readLoop()
	if client.pingInterval > 0 {
		s.loops.Add(1)
		go s.pingLoop()
	}
	go s.closeChannels()

	return s
}

// Messages returns the channel of incoming WebSocket messages. The channel is
// closed once the stream is closed and all buffered messages drained.
func (s *Stream) Messages() <-chan Message {
	return s.msgCh
}

// Errors returns the channel of connection and parse errors encountered while
// the stream is running. The channel is closed once the stream is closed.
func (s *Stream) Errors() <-chan error {
	return s.errCh
}

// Subscribed returns the list of channel names currently subscribed to on
// this stream.
func (s *Stream) Subscribed() []string {
	s.subMu.RLock()
	defer s.subMu.RUnlock()

	out := make([]string, 0, len(s.subs))
	for ch := range s.subs {
		out = append(out, ch)
	}
	return out
}

// Subscribe sends a subscribe request for the given channel names, e.g.
// "liquidation_orders" or "futures_ticker@Binance_BTCUSDT". Multiple channels
// may be subscribed in a single call.
func (s *Stream) Subscribe(channels ...string) error {
	if len(channels) == 0 {
		return nil
	}

	if err := s.send(map[string]any{
		"method":   "subscribe",
		"channels": channels,
	}); err != nil {
		return err
	}

	s.subMu.Lock()
	for _, ch := range channels {
		s.subs[ch] = struct{}{}
	}
	s.subMu.Unlock()

	return nil
}

// Unsubscribe sends an unsubscribe request for the given channel names.
func (s *Stream) Unsubscribe(channels ...string) error {
	if len(channels) == 0 {
		return nil
	}

	if err := s.send(map[string]any{
		"method":   "unsubscribe",
		"channels": channels,
	}); err != nil {
		return err
	}

	s.subMu.Lock()
	for _, ch := range channels {
		delete(s.subs, ch)
	}
	s.subMu.Unlock()

	return nil
}

// Close closes the underlying WebSocket connection and stops the read and
// ping loops. It is safe to call Close multiple times.
func (s *Stream) Close() error {
	s.closedMu.Lock()
	if s.closed {
		s.closedMu.Unlock()
		return nil
	}
	s.closed = true
	s.closedMu.Unlock()

	s.cancel()

	s.sendMu.Lock()
	_ = s.conn.WriteClose(1000, "")
	s.sendMu.Unlock()

	return s.conn.Close()
}

// send marshals v to JSON and writes it as a single text frame.
func (s *Stream) send(v any) error {
	s.closedMu.Lock()
	closed := s.closed
	s.closedMu.Unlock()
	if closed {
		return ErrStreamClosed
	}

	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("coinglass/websocket: failed to marshal message: %w", err)
	}

	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	return s.conn.WriteText(b)
}

// pushError sends a non-fatal error to the Errors() channel, dropping it if
// the channel is full so the read loop never blocks on a slow consumer.
func (s *Stream) pushError(err error) {
	select {
	case s.errCh <- err:
	default:
	}
}

// readLoop reads incoming frames and dispatches parsed messages until the
// stream is closed or a read error occurs.
func (s *Stream) readLoop() {
	defer s.loops.Done()
	defer s.Close()

	for {
		opcode, data, err := s.conn.ReadMessage()
		if err != nil {
			if s.ctx.Err() == nil {
				s.pushError(fmt.Errorf("coinglass/websocket: read error: %w", err))
			}
			return
		}

		if opcode == opClose {
			return
		}

		// Coinglass replies to the client's "ping" text message with a plain
		// "pong" text message; swallow it rather than surfacing it as data.
		if string(data) == "pong" {
			continue
		}

		msg, ok := parseMessage(data)
		if !ok {
			continue
		}

		select {
		case s.msgCh <- msg:
		case <-s.ctx.Done():
			return
		}
	}
}

// pingLoop periodically sends the application-level "ping" text message
// Coinglass expects to keep the connection alive.
func (s *Stream) pingLoop() {
	defer s.loops.Done()
	ticker := time.NewTicker(s.client.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.sendMu.Lock()
			err := s.conn.WriteText([]byte("ping"))
			s.sendMu.Unlock()
			if err != nil {
				s.pushError(fmt.Errorf("coinglass/websocket: ping failed: %w", err))
			}
		}
	}
}

// closeChannels closes the output channels only after every producer has
// stopped. This prevents a concurrent heartbeat from sending on a closed
// error channel while the read loop is shutting down.
func (s *Stream) closeChannels() {
	s.loops.Wait()
	close(s.msgCh)
	close(s.errCh)
}

// Message is a single decoded WebSocket message pushed by a Coinglass
// channel. Data holds the raw JSON "data" array/object for the channel;
// use the Decode* helpers (e.g. DecodeLiquidationOrders) to unmarshal it into
// typed values.
type Message struct {
	// Channel is the channel name the message was pushed for, e.g.
	// "liquidation_orders" or "futures_ticker@Binance_BTCUSDT".
	Channel string `json:"channel"`
	// Data is the raw "data" payload, typically a JSON array of records.
	Data json.RawMessage `json:"data"`
}

// parseMessage attempts to interpret a raw text frame as a Coinglass channel
// message. Frames that are not JSON objects with a "channel" field (e.g. a
// bare "pong") are ignored.
func parseMessage(data []byte) (Message, bool) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return Message{}, false
	}
	if msg.Channel == "" {
		return Message{}, false
	}
	return msg, true
}
