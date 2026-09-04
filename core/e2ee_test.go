package core

import (
	"bytes"
	"testing"
)

func TestE2EEEncryptionDecryption(t *testing.T) {
	// Generate Bob's keypair
	bobPriv, bobPub, err := GenerateE2EEKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate Bob keypair: %v", err)
	}

	secretMessage := []byte("Top secret IPv7 encrypted payload: E2EE is working!")

	// Alice encrypts for Bob using Bob's public key
	envelope, err := EncryptE2EE(bobPub.Bytes(), secretMessage)
	if err != nil {
		t.Fatalf("Failed to encrypt message: %v", err)
	}

	if bytes.Equal(envelope, secretMessage) {
		t.Fatalf("Envelope is identical to plaintext, expected ciphertext")
	}

	// Bob decrypts with his private key
	decrypted, err := DecryptE2EE(bobPriv, envelope)
	if err != nil {
		t.Fatalf("Bob failed to decrypt: %v", err)
	}

	if !bytes.Equal(decrypted, secretMessage) {
		t.Fatalf("Decrypted message does not match original plaintext: got %s, want %s", string(decrypted), string(secretMessage))
	}
	t.Logf("Decrypted successfully: %s", string(decrypted))
}

func TestE2EETamperRejection(t *testing.T) {
	bobPriv, bobPub, _ := GenerateE2EEKeyPair()
	secretMessage := []byte("Sensitive data")

	envelope, _ := EncryptE2EE(bobPub.Bytes(), secretMessage)

	// Tamper with one byte in the ciphertext portion
	envelope[len(envelope)-1] ^= 0xFF

	// Decryption must fail with authentication error
	_, err := DecryptE2EE(bobPriv, envelope)
	if err == nil {
		t.Fatalf("Expected decryption error for tampered envelope, got nil")
	}
	t.Logf("Correctly rejected tampered envelope: %v", err)
}

func TestE2EEDeriveFromSeed(t *testing.T) {
	seed := []byte("01234567890123456789012345678901") // 32 bytes
	priv1, pub1, err := DeriveX25519FromSeed(seed)
	if err != nil {
		t.Fatalf("Failed to derive from seed: %v", err)
	}

	priv2, pub2, _ := DeriveX25519FromSeed(seed)
	if !bytes.Equal(pub1.Bytes(), pub2.Bytes()) {
		t.Fatalf("Derived public keys from identical seed are not deterministic")
	}
	_ = priv1
	_ = priv2
}
