package websocket

import (
	"bufio"
	"bytes"
	"testing"
)

// TestWriteFrameThenReadFrame verifies that a frame written by writeFrame
// (client -> server, masked) can be read back by readFrame after simulating
// the masking a real server would have already removed. Since readFrame
// expects the masked bit as sent by the actual peer, we instead validate the
// wire format directly: FIN+opcode byte, masked length byte, and that the
// payload round-trips once unmasked using the same key extracted from the
// header.
func TestWriteFrameRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	payload := []byte("hello coinglass")

	if err := writeFrame(&buf, opText, payload); err != nil {
		t.Fatalf("writeFrame failed: %v", err)
	}

	data := buf.Bytes()
	if len(data) < 2 {
		t.Fatalf("frame too short: %d bytes", len(data))
	}

	if data[0] != 0x80|opText {
		t.Fatalf("unexpected first byte: %x", data[0])
	}
	if data[1]&0x80 == 0 {
		t.Fatalf("expected masked bit to be set")
	}

	length := int(data[1] & 0x7F)
	if length != len(payload) {
		t.Fatalf("unexpected payload length: got %d want %d", length, len(payload))
	}

	maskKey := data[2:6]
	masked := data[6 : 6+length]
	unmasked := make([]byte, length)
	for i := range unmasked {
		unmasked[i] = masked[i] ^ maskKey[i%4]
	}

	if string(unmasked) != string(payload) {
		t.Fatalf("payload mismatch: got %q want %q", unmasked, payload)
	}
}

// TestReadFrameUnmasked verifies readFrame can parse an unmasked server frame
// (as sent by Coinglass) including the extended 16-bit length form.
func TestReadFrameUnmasked(t *testing.T) {
	payload := make([]byte, 200) // forces the 126 extended-length form
	for i := range payload {
		payload[i] = byte('a' + i%26)
	}

	var buf bytes.Buffer
	buf.WriteByte(0x80 | opText) // FIN=1, opcode=text
	buf.WriteByte(126)           // unmasked, extended length follows
	buf.WriteByte(byte(len(payload) >> 8))
	buf.WriteByte(byte(len(payload)))
	buf.Write(payload)

	fr, err := readFrame(bufio.NewReader(&buf))
	if err != nil {
		t.Fatalf("readFrame failed: %v", err)
	}
	if !fr.fin {
		t.Fatal("expected FIN bit set")
	}
	if fr.opcode != opText {
		t.Fatalf("unexpected opcode: %x", fr.opcode)
	}
	if string(fr.payload) != string(payload) {
		t.Fatalf("payload mismatch")
	}
}

// TestReadMessageFragmented verifies readMessage reassembles a message split
// across a text frame and a continuation frame.
func TestReadMessageFragmented(t *testing.T) {
	var buf bytes.Buffer

	// First fragment: FIN=0, opcode=text, payload="hel"
	buf.WriteByte(0x00 | opText)
	buf.WriteByte(3)
	buf.WriteString("hel")

	// Final fragment: FIN=1, opcode=continuation, payload="lo"
	buf.WriteByte(0x80 | opContinuation)
	buf.WriteByte(2)
	buf.WriteString("lo")

	opcode, payload, err := readMessage(bufio.NewReader(&buf))
	if err != nil {
		t.Fatalf("readMessage failed: %v", err)
	}
	if opcode != opText {
		t.Fatalf("unexpected opcode: %x", opcode)
	}
	if string(payload) != "hello" {
		t.Fatalf("unexpected payload: %q", payload)
	}
}

func TestComputeAcceptKey(t *testing.T) {
	// Example from RFC 6455 section 1.3.
	key := "dGhlIHNhbXBsZSBub25jZQ=="
	want := "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="

	if got := computeAcceptKey(key); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestHeaderHasToken(t *testing.T) {
	if !headerHasToken("keep-alive, Upgrade", "upgrade") {
		t.Fatal("expected Upgrade token to be found")
	}
	if headerHasToken("keep-alive", "upgrade") {
		t.Fatal("did not expect Upgrade token to be found")
	}
}
