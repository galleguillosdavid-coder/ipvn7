package adapters

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"
)

func TestPairingPayload_SignAndVerify(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate keypair: %v", err)
	}

	endpoints := []string{"192.168.1.100:7001", "203.0.113.15:7001"}
	payload, err := NewPairingPayload("Alice Workstation", priv, endpoints, "10.7.1.10", "fd07::1:10", 10*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create pairing payload: %v", err)
	}

	// 1. Verify fresh payload
	valid, err := payload.Verify()
	if err != nil || !valid {
		t.Fatalf("Expected valid payload, got valid=%v, err=%v", valid, err)
	}

	// 2. Tampering test: modify endpoints -> verification must fail
	payload.Endpoints = append(payload.Endpoints, "10.0.0.99:9999")
	valid, err = payload.Verify()
	if valid {
		t.Fatalf("Tampered payload should have failed verification")
	}

	// 3. Expiry test
	expiredPayload, _ := NewPairingPayload("Bob Expired", priv, endpoints, "", "", -1*time.Minute)
	valid, err = expiredPayload.Verify()
	if valid || err == nil {
		t.Fatalf("Expired payload should return error and valid=false")
	}
	_ = pub
}

func TestPairing_ComputeSAS_SymmetryAndFormat(t *testing.T) {
	pubA, _, _ := ed25519.GenerateKey(rand.Reader)
	pubB, _, _ := ed25519.GenerateKey(rand.Reader)
	salt := []byte("ipv7_secure_salt_oob_2026")

	// Alice computes SAS
	sasA := ComputeSAS(pubA, pubB, salt)

	// Bob computes SAS with inverted parameter order
	sasB := ComputeSAS(pubB, pubA, salt)

	// Must be identical regardless of caller perspective
	if sasA.Numeric != sasB.Numeric {
		t.Fatalf("SAS numeric mismatch: A=%s, B=%s", sasA.Numeric, sasB.Numeric)
	}
	if len(sasA.Emojis) != 4 || len(sasB.Emojis) != 4 {
		t.Fatalf("Expected 4 emojis in SAS, got A=%d, B=%d", len(sasA.Emojis), len(sasB.Emojis))
	}
	for i := 0; i < 4; i++ {
		if sasA.Emojis[i] != sasB.Emojis[i] {
			t.Fatalf("SAS emoji mismatch at index %d: A=%s, B=%s", i, sasA.Emojis[i], sasB.Emojis[i])
		}
	}

	// Format check: XXX-XXX
	if len(sasA.Numeric) != 7 || sasA.Numeric[3] != '-' {
		t.Fatalf("Invalid SAS format, expected XXX-XXX, got: %s", sasA.Numeric)
	}
}

func TestPairing_URMultiFrameReassembly(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	payload, err := NewPairingPayload(
		"Cluster Node Alpha",
		priv,
		[]string{"10.0.1.50:7001", "172.16.0.2:7001", "203.0.113.88:7001"},
		"10.7.2.1",
		"fd07::2:1",
		15*time.Minute,
	)
	if err != nil {
		t.Fatalf("Failed to create payload: %v", err)
	}

	// Force small chunk size to generate multiple frames (e.g. 40 bytes)
	frames, err := EncodeMultiFrame(payload, 40)
	if err != nil {
		t.Fatalf("Failed to encode multi-frame UR: %v", err)
	}

	if len(frames) < 3 {
		t.Fatalf("Expected at least 3 chunks with 40-byte limit, got %d", len(frames))
	}

	reassembler := NewURReassembler()

	// Feed in reverse order to test order independence
	var completedPayload *PairingPayload
	for i := len(frames) - 1; i >= 0; i-- {
		p, complete, err := reassembler.Feed(frames[i])
		if err != nil {
			t.Fatalf("Error feeding frame %d: %v", i, err)
		}
		if complete {
			completedPayload = p
		}
	}

	if completedPayload == nil {
		t.Fatalf("Expected reassembler to complete after receiving all frames")
	}

	if completedPayload.DID != payload.DID {
		t.Fatalf("Reassembled DID mismatch: expected %s, got %s", payload.DID, completedPayload.DID)
	}

	if completedPayload.DeviceName != payload.DeviceName {
		t.Fatalf("Device name mismatch: expected %s, got %s", payload.DeviceName, completedPayload.DeviceName)
	}

	valid, err := completedPayload.Verify()
	if err != nil || !valid {
		t.Fatalf("Reassembled payload signature failed verification: %v", err)
	}
}
