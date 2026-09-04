package adapters

import (
	"bytes"
	"crypto/rand"
	"testing"
	"time"

	"ipv7/core"
)

func TestQUICAdapter(t *testing.T) {
	adapterA, err := NewQUICAdapter("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create QUIC adapter A: %v", err)
	}

	adapterB, err := NewQUICAdapter("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create QUIC adapter B: %v", err)
	}

	if err := adapterA.Start(); err != nil {
		t.Fatalf("Failed to start adapter A: %v", err)
	}
	defer adapterA.Stop()

	if err := adapterB.Start(); err != nil {
		t.Fatalf("Failed to start adapter B: %v", err)
	}
	defer adapterB.Stop()

	idA, privA, _ := core.GenerateIdentity()
	c := &core.Container{
		SenderPubKey: idA.Bytes(),
		Payload:      []byte("hello via reliable quic"),
	}
	if err := c.Sign(privA); err != nil {
		t.Fatalf("Failed to sign container: %v", err)
	}

	addrB := adapterB.LocalAddr().String()
	if err := adapterA.Send(c, []string{addrB}); err != nil {
		t.Fatalf("Failed to send container via QUIC: %v", err)
	}

	select {
	case received := <-adapterB.Receive():
		if string(received.Payload) != "hello via reliable quic" {
			t.Fatalf("Payload mismatch, got: %s", string(received.Payload))
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("Timed out waiting for QUIC container")
	}
}

func TestQUICAdapterLargePayload(t *testing.T) {
	adapterA, err := NewQUICAdapter("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create QUIC adapter A: %v", err)
	}

	adapterB, err := NewQUICAdapter("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create QUIC adapter B: %v", err)
	}

	if err := adapterA.Start(); err != nil {
		t.Fatalf("Failed to start adapter A: %v", err)
	}
	defer adapterA.Stop()

	if err := adapterB.Start(); err != nil {
		t.Fatalf("Failed to start adapter B: %v", err)
	}
	defer adapterB.Stop()

	idA, privA, _ := core.GenerateIdentity()

	// 5 MB payload - far exceeding normal UDP MTU
	largePayload := make([]byte, 5*1024*1024)
	if _, err := rand.Read(largePayload); err != nil {
		t.Fatalf("Failed to generate random data: %v", err)
	}

	c := &core.Container{
		SenderPubKey: idA.Bytes(),
		Payload:      largePayload,
	}
	if err := c.Sign(privA); err != nil {
		t.Fatalf("Failed to sign large container: %v", err)
	}

	addrB := adapterB.LocalAddr().String()
	if err := adapterA.Send(c, []string{addrB}); err != nil {
		t.Fatalf("Failed to send large container via QUIC: %v", err)
	}

	select {
	case received := <-adapterB.Receive():
		if !bytes.Equal(received.Payload, largePayload) {
			t.Fatalf("Large payload content mismatch")
		}
	case <-time.After(10 * time.Second):
		t.Fatalf("Timed out waiting for large QUIC container")
	}
}
