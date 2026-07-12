package websocket

import (
	"time"
)

// zeroTime clears a previously set net.Conn deadline when passed to
// SetDeadline/SetReadDeadline/SetWriteDeadline.
var zeroTime time.Time

// ReadMessage blocks until a complete text or close message has been
// received. Ping/pong control frames are handled transparently: pings are
// answered automatically (though Coinglass uses application-level "ping"/
// "pong" text frames rather than protocol-level control frames) and are
// never returned to the caller.
func (c *rawConn) ReadMessage() (opcode byte, payload []byte, err error) {
	for {
		opcode, payload, err = readMessage(c.br)
		if err != nil {
			return 0, nil, err
		}

		switch opcode {
		case opPing:
			// Best-effort pong per RFC 6455; Coinglass itself does not rely
			// on protocol-level ping/pong, but responding keeps us compliant
			// with any intermediary that does.
			_ = writeFrame(c.conn, opPong, payload)
			continue
		case opPong:
			continue
		default:
			return opcode, payload, nil
		}
	}
}

// WriteText sends a single text message.
func (c *rawConn) WriteText(payload []byte) error {
	return writeFrame(c.conn, opText, payload)
}

// WriteClose sends a close frame with the given status code and reason.
func (c *rawConn) WriteClose(code uint16, reason string) error {
	payload := make([]byte, 2+len(reason))
	payload[0] = byte(code >> 8)
	payload[1] = byte(code)
	copy(payload[2:], reason)
	return writeFrame(c.conn, opClose, payload)
}

// SetReadDeadline sets the deadline for future ReadMessage calls.
func (c *rawConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future WriteText/WriteClose calls.
func (c *rawConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

// Close closes the underlying network connection immediately without
// performing the WebSocket closing handshake.
func (c *rawConn) Close() error {
	return c.conn.Close()
}
