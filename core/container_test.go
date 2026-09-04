package core

import (
	"bytes"
	"testing"
)

func TestGenerateIdentity(t *testing.T) {
	id, priv, err := GenerateIdentity()
	if err != nil {
		t.Fatalf("Failed to generate identity: %v", err)
	}
	if id == nil || priv == nil {
		t.Fatalf("Identity or private key is nil")
	}
}

func TestContainerSigningAndVerification(t *testing.T) {
	idA, privA, _ := GenerateIdentity()
	idB, _, _ := GenerateIdentity()

	container := &Container{
		SenderPubKey:   idA.Bytes(),
		ReceiverPubKey: idB.Bytes(),
		SessionID:      "session-123",
		Payload:        []byte("hello ipv7"),
	}

	err := container.Sign(privA)
	if err != nil {
		t.Fatalf("Failed to sign container: %v", err)
	}
	if len(container.Signature) == 0 {
		t.Fatalf("Container signature is empty")
	}

	if !container.Verify() {
		t.Fatalf("Container verification failed for valid signature")
	}

	// Tamper payload
	container.Payload = []byte("tampered data")
	if container.Verify() {
		t.Fatalf("Container verification succeeded for tampered payload")
	}
}

func TestContainerSerialization(t *testing.T) {
	id, priv, _ := GenerateIdentity()
	
	c1 := &Container{
		SenderPubKey: id.Bytes(),
		Payload:      []byte("data"),
	}
	c1.Sign(priv)
	
	encoded, err := c1.Marshal()
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}
	
	c2 := &Container{}
	err = c2.Unmarshal(encoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	
	if !bytes.Equal(c1.SenderPubKey, c2.SenderPubKey) {
		t.Fatalf("SenderPubKey mismatch")
	}
	if !bytes.Equal(c1.Signature, c2.Signature) {
		t.Fatalf("Signature mismatch")
	}
	
	if !c2.Verify() {
		t.Fatalf("Unmarshaled container verification failed")
	}
}
