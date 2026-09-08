package core

import (
	"testing"
	"time"
)

func TestHandshakePayloadSerialization(t *testing.T) {
	hp := &HandshakePayload{
		Type:       ControlHandshakeReq,
		Ed25519Pub: []byte("01234567890123456789012345678901"),
		X25519Pub:  []byte("abcdefabcdefabcdefabcdefabcdefab"),
		Endpoints:  []string{"127.0.0.1:7001", "192.168.1.5:7001"},
		Timestamp:  time.Now().UnixNano(),
		Nonce:      GenerateNonce(),
	}

	data, err := EncodeHandshake(hp)
	if err != nil {
		t.Fatalf("Failed to encode handshake: %v", err)
	}

	decoded, err := DecodeHandshake(data)
	if err != nil {
		t.Fatalf("Failed to decode handshake: %v", err)
	}

	if decoded.Type != hp.Type {
		t.Errorf("Expected type %d, got %d", hp.Type, decoded.Type)
	}
	if decoded.Nonce != hp.Nonce {
		t.Errorf("Expected nonce %d, got %d", hp.Nonce, decoded.Nonce)
	}
	if !decoded.IsFresh(10 * time.Second) {
		t.Errorf("Expected handshake payload to be fresh")
	}
}

func TestPingPayloadSerialization(t *testing.T) {
	pp := &PingPayload{
		Type:      ControlPing,
		Nonce:     123456789,
		Timestamp: time.Now().UnixNano(),
	}

	data, err := EncodePing(pp)
	if err != nil {
		t.Fatalf("Failed to encode ping: %v", err)
	}

	decoded, err := DecodePing(data)
	if err != nil {
		t.Fatalf("Failed to decode ping: %v", err)
	}

	if decoded.Type != ControlPing {
		t.Errorf("Expected type %d, got %d", ControlPing, decoded.Type)
	}
	if decoded.Nonce != 123456789 {
		t.Errorf("Expected nonce 123456789, got %d", decoded.Nonce)
	}
}
