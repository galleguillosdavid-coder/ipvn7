package core

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"time"

	"github.com/fxamacker/cbor/v2"
)

// Protocol packet control types
const (
	ControlHandshakeReq  uint8 = 0x01
	ControlHandshakeResp uint8 = 0x02
	ControlPing          uint8 = 0x03
	ControlPong          uint8 = 0x04
	ControlStreamChunk   uint8 = 0x05
	ControlData          uint8 = 0x06
)

// HandshakePayload represents a cryptographic identity exchange
type HandshakePayload struct {
	Type       uint8    `cbor:"1,keyasint"`
	Ed25519Pub []byte   `cbor:"2,keyasint"`
	X25519Pub  []byte   `cbor:"3,keyasint,omitempty"`
	Endpoints  []string `cbor:"4,keyasint,omitempty"`
	Timestamp  int64    `cbor:"5,keyasint"`
	Nonce      uint64   `cbor:"6,keyasint"`
}

// PingPayload represents an active RTT latency probe
type PingPayload struct {
	Type      uint8  `cbor:"1,keyasint"`
	Nonce     uint64 `cbor:"2,keyasint"`
	Timestamp int64  `cbor:"3,keyasint"`
}

// EncodeHandshake serializes a HandshakePayload using canonical CBOR
func EncodeHandshake(hp *HandshakePayload) ([]byte, error) {
	return cborEnc.Marshal(hp)
}

// DecodeHandshake deserializes a HandshakePayload
func DecodeHandshake(data []byte) (*HandshakePayload, error) {
	hp := &HandshakePayload{}
	if err := cbor.Unmarshal(data, hp); err != nil {
		return nil, err
	}
	if hp.Type != ControlHandshakeReq && hp.Type != ControlHandshakeResp {
		return nil, errors.New("not a valid handshake payload")
	}
	return hp, nil
}

// EncodePing serializes a PingPayload
func EncodePing(pp *PingPayload) ([]byte, error) {
	return cborEnc.Marshal(pp)
}

// DecodePing deserializes a PingPayload
func DecodePing(data []byte) (*PingPayload, error) {
	pp := &PingPayload{}
	if err := cbor.Unmarshal(data, pp); err != nil {
		return nil, err
	}
	if pp.Type != ControlPing && pp.Type != ControlPong {
		return nil, errors.New("not a valid ping/pong payload")
	}
	return pp, nil
}

// GenerateNonce creates a secure pseudo-random 64-bit unsigned integer
func GenerateNonce() uint64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return binary.BigEndian.Uint64(b[:])
}

// IsFresh checks that a handshake timestamp is within acceptable drift window (e.g. 5 minutes)
func (hp *HandshakePayload) IsFresh(maxDrift time.Duration) bool {
	now := time.Now().UnixNano()
	diff := time.Duration(now - hp.Timestamp)
	if diff < 0 {
		diff = -diff
	}
	return diff <= maxDrift
}
