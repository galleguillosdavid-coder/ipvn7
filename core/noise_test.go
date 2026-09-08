package core

import (
	"bytes"
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"testing"
)

func TestNoiseHandshakeAndSession(t *testing.T) {
	// 1. Setup Static Keys for Peer A (Initiator) and Peer B (Responder)
	pubEdA, privEdA, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey A failed: %v", err)
	}
	privXA, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey X A failed: %v", err)
	}

	pubEdB, privEdB, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey B failed: %v", err)
	}
	privXB, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey X B failed: %v", err)
	}

	// 2. Initialize handshake states
	initHs, err := NewNoiseHandshake(true, privEdA, privXA)
	if err != nil {
		t.Fatalf("NewNoiseHandshake initiator failed: %v", err)
	}
	respHs, err := NewNoiseHandshake(false, privEdB, privXB)
	if err != nil {
		t.Fatalf("NewNoiseHandshake responder failed: %v", err)
	}

	// 3. Step 1: Initiator -> Responder (Msg1)
	msg1, err := initHs.InitiatorStep1()
	if err != nil {
		t.Fatalf("InitiatorStep1 failed: %v", err)
	}
	if len(msg1.EphemeralPub) != 32 {
		t.Fatalf("expected 32-byte ephemeral pubkey, got %d", len(msg1.EphemeralPub))
	}

	// 4. Step 2: Responder -> Initiator (Msg2)
	msg2, err := respHs.ResponderStep2(msg1)
	if err != nil {
		t.Fatalf("ResponderStep2 failed: %v", err)
	}
	if !bytes.Equal(msg2.Ed25519Pub, pubEdB) {
		t.Fatalf("Responder Ed25519 pubkey mismatch")
	}

	// 5. Step 3: Initiator verifies Msg2, produces Msg3 and session
	msg3, initSession, err := initHs.InitiatorStep3(msg1, msg2)
	if err != nil {
		t.Fatalf("InitiatorStep3 failed: %v", err)
	}
	if !bytes.Equal(msg3.Ed25519Pub, pubEdA) {
		t.Fatalf("Initiator Ed25519 pubkey mismatch")
	}

	// 6. Responder verifies Msg3 and establishes session
	respSession, err := respHs.ResponderFinal(msg1, msg2, msg3)
	if err != nil {
		t.Fatalf("ResponderFinal failed: %v", err)
	}

	// 7. Verify cryptographic session consistency
	if initSession.SessionID != respSession.SessionID {
		t.Fatalf("Session IDs do not match: %x vs %x", initSession.SessionID, respSession.SessionID)
	}
	if !bytes.Equal(initSession.TxKey, respSession.RxKey) {
		t.Fatalf("Initiator TxKey does not match Responder RxKey")
	}
	if !bytes.Equal(initSession.RxKey, respSession.TxKey) {
		t.Fatalf("Initiator RxKey does not match Responder TxKey")
	}

	// 8. Test Bidirectional PFS Data Exchange
	// A -> B
	plainA := []byte("Hello sovereign IPv7 mesh from Node A!")
	cipherA, err := initSession.Encrypt(plainA)
	if err != nil {
		t.Fatalf("Encrypt A failed: %v", err)
	}
	decryptedB, err := respSession.Decrypt(cipherA)
	if err != nil {
		t.Fatalf("Decrypt B failed: %v", err)
	}
	if !bytes.Equal(plainA, decryptedB) {
		t.Fatalf("plaintext mismatch A->B: got %s, want %s", decryptedB, plainA)
	}

	// B -> A
	plainB := []byte("Greetings from Node B, forward-secrecy verified!")
	cipherB, err := respSession.Encrypt(plainB)
	if err != nil {
		t.Fatalf("Encrypt B failed: %v", err)
	}
	decryptedA, err := initSession.Decrypt(cipherB)
	if err != nil {
		t.Fatalf("Decrypt A failed: %v", err)
	}
	if !bytes.Equal(plainB, decryptedA) {
		t.Fatalf("plaintext mismatch B->A: got %s, want %s", decryptedA, plainB)
	}
}

func TestNoiseTamperedSignature(t *testing.T) {
	_, privEdA, _ := ed25519.GenerateKey(rand.Reader)
	privXA, _ := ecdh.X25519().GenerateKey(rand.Reader)

	_, privEdB, _ := ed25519.GenerateKey(rand.Reader)
	privXB, _ := ecdh.X25519().GenerateKey(rand.Reader)

	initHs, _ := NewNoiseHandshake(true, privEdA, privXA)
	respHs, _ := NewNoiseHandshake(false, privEdB, privXB)

	msg1, _ := initHs.InitiatorStep1()
	msg2, _ := respHs.ResponderStep2(msg1)

	// Tamper with signature
	msg2.Signature[0] ^= 0xFF

	_, _, err := initHs.InitiatorStep3(msg1, msg2)
	if err == nil {
		t.Fatalf("expected error on tampered signature, got nil")
	}
}
