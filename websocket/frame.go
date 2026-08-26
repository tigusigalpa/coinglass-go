package websocket

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// WebSocket opcodes as defined by RFC 6455 section 11.8.
const (
	opContinuation byte = 0x0
	opText         byte = 0x1
	opBinary       byte = 0x2
	opClose        byte = 0x8
	opPing         byte = 0x9
	opPong         byte = 0xA
)

const maxFramePayload = 16 << 20 // 16 MiB safety limit per assembled message

// frame is a single parsed WebSocket frame header plus its payload.
type frame struct {
	fin     bool
	opcode  byte
	payload []byte
}

// readFrame reads and unmasks (if necessary) a single WebSocket frame from r.
// Coinglass, acting as the server, sends unmasked frames per RFC 6455.
func readFrame(r *bufio.Reader) (frame, error) {
	var head [2]byte
	if _, err := io.ReadFull(r, head[:]); err != nil {
		return frame{}, err
	}

	fin := head[0]&0x80 != 0
	opcode := head[0] & 0x0F
	masked := head[1]&0x80 != 0
	payloadLen := int64(head[1] & 0x7F)

	switch payloadLen {
	case 126:
		var ext [2]byte
		if _, err := io.ReadFull(r, ext[:]); err != nil {
			return frame{}, err
		}
		payloadLen = int64(binary.BigEndian.Uint16(ext[:]))
	case 127:
		var ext [8]byte
		if _, err := io.ReadFull(r, ext[:]); err != nil {
			return frame{}, err
		}
		payloadLen = int64(binary.BigEndian.Uint64(ext[:]))
	}

	if payloadLen < 0 || payloadLen > maxFramePayload {
		return frame{}, fmt.Errorf("coinglass/websocket: frame payload too large: %d", payloadLen)
	}

	var maskKey [4]byte
	if masked {
		if _, err := io.ReadFull(r, maskKey[:]); err != nil {
			return frame{}, err
		}
	}

	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(r, payload); err != nil {
		return frame{}, err
	}

	if masked {
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}

	return frame{fin: fin, opcode: opcode, payload: payload}, nil
}

// readMessage reads a complete (possibly fragmented) message from r,
// returning its final opcode and the concatenated payload.
func readMessage(r *bufio.Reader) (byte, []byte, error) {
	first, err := readFrame(r)
	if err != nil {
		return 0, nil, err
	}

	if first.opcode == opClose {
		return opClose, first.payload, nil
	}

	if first.opcode == opPing || first.opcode == opPong {
		return first.opcode, first.payload, nil
	}

	opcode := first.opcode
	payload := first.payload

	for !first.fin {
		next, err := readFrame(r)
		if err != nil {
			return 0, nil, err
		}
		if next.opcode != opContinuation {
			return 0, nil, errors.New("coinglass/websocket: expected continuation frame")
		}
		if len(payload)+len(next.payload) > maxFramePayload {
			return 0, nil, fmt.Errorf("coinglass/websocket: message payload too large")
		}
		payload = append(payload, next.payload...)
		first.fin = next.fin
	}

	return opcode, payload, nil
}

// writeFrame writes a single, unfragmented, masked frame to w. Per RFC 6455,
// every frame sent by a client MUST be masked with a random 32-bit key.
func writeFrame(w io.Writer, opcode byte, payload []byte) error {
	var header []byte

	b0 := byte(0x80) | opcode // FIN=1

	length := len(payload)
	switch {
	case length <= 125:
		header = append(header, b0, byte(length)|0x80)
	case length <= 0xFFFF:
		header = append(header, b0, 126|0x80)
		var ext [2]byte
		binary.BigEndian.PutUint16(ext[:], uint16(length))
		header = append(header, ext[:]...)
	default:
		header = append(header, b0, 127|0x80)
		var ext [8]byte
		binary.BigEndian.PutUint64(ext[:], uint64(length))
		header = append(header, ext[:]...)
	}

	var maskKey [4]byte
	if _, err := rand.Read(maskKey[:]); err != nil {
		return fmt.Errorf("coinglass/websocket: failed to generate mask key: %w", err)
	}
	header = append(header, maskKey[:]...)

	masked := make([]byte, length)
	for i := 0; i < length; i++ {
		masked[i] = payload[i] ^ maskKey[i%4]
	}

	if _, err := w.Write(header); err != nil {
		return err
	}
	if length > 0 {
		if _, err := w.Write(masked); err != nil {
			return err
		}
	}
	return nil
}
