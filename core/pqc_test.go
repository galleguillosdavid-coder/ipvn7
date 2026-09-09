package core

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestHybridIdentityAndSigning(t *testing.T) {
	idClassic, priv, err := GenerateIdentity()
	if err != nil {
		t.Fatal(err)
	}

	pqcPriv, pqcPub, err := GeneratePQCKeyStub()
	if err != nil {
		t.Fatal(err)
	}

	id := NewHybridIdentity(idClassic, pqcPub, AlgHybridMLDSA)
	payload := []byte("Post-Quantum hybrid payload token for IPv7 network")

	sig, err := SignHybrid(priv, pqcPriv, payload)
	if err != nil {
		t.Fatalf("Signing failed: %v", err)
	}

	if sig.Algorithm != AlgHybridMLDSA {
		t.Fatalf("Expected AlgHybridMLDSA, got %v", sig.Algorithm)
	}

	if !VerifyHybrid(id, payload, sig) {
		t.Fatalf("Hybrid verification failed")
	}

	// Tampered payload must fail
	tamperedPayload := []byte("Tampered post-quantum payload")
	if VerifyHybrid(id, tamperedPayload, sig) {
		t.Fatalf("Verification should have failed on tampered payload")
	}
}

func TestCombineKEMSecrets(t *testing.T) {
	classicSecret := make([]byte, 32)
	pqcSecret := make([]byte, 32)
	_, _ = rand.Read(classicSecret)
	_, _ = rand.Read(pqcSecret)

	key1, err := CombineKEMSecrets(classicSecret, pqcSecret, []byte("ipv7-session-auth"))
	if err != nil {
		t.Fatalf("Failed to combine secrets: %v", err)
	}

	key2, err := CombineKEMSecrets(classicSecret, pqcSecret, []byte("ipv7-session-auth"))
	if err != nil {
		t.Fatalf("Failed to combine secrets: %v", err)
	}

	if !bytes.Equal(key1, key2) {
		t.Fatalf("Deterministic HKDF combination mismatch")
	}

	if len(key1) != 32 {
		t.Fatalf("Expected 32-byte key, got %d bytes", len(key1))
	}
}
