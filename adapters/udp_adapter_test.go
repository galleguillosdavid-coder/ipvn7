package adapters

import (
	"testing"
	"time"
	
	"ipv7/core"
)

func TestUDPAdapter(t *testing.T) {
	// Create two adapters
	adapterA, err := NewUDPAdapter("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create adapter A: %v", err)
	}
	
	adapterB, err := NewUDPAdapter("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create adapter B: %v", err)
	}
	
	err = adapterA.Start()
	if err != nil {
		t.Fatalf("Failed to start adapter A: %v", err)
	}
	defer adapterA.Stop()
	
	err = adapterB.Start()
	if err != nil {
		t.Fatalf("Failed to start adapter B: %v", err)
	}
	defer adapterB.Stop()
	
	// Create identity and container
	idA, privA, _ := core.GenerateIdentity()
	c := &core.Container{
		SenderPubKey: idA.Bytes(),
		Payload:      []byte("hello via udp"),
	}
	c.Sign(privA)
	
	// Send from A to B
	addrB := adapterB.conn.LocalAddr().String()
	err = adapterA.Send(c, []string{addrB})
	if err != nil {
		t.Fatalf("Failed to send container: %v", err)
	}
	
	// Wait to receive on B
	select {
	case received := <-adapterB.Receive():
		if string(received.Payload) != "hello via udp" {
			t.Fatalf("Received payload mismatch, got: %s", string(received.Payload))
		}
	case <-time.After(time.Second):
		t.Fatalf("Timeout waiting for container")
	}
}

func TestUDPAdapterMTU(t *testing.T) {
	adapter, _ := NewUDPAdapter("127.0.0.1:0")
	adapter.Start()
	defer adapter.Stop()
	
	id, priv, _ := core.GenerateIdentity()
	
	// Payload too large
	largePayload := make([]byte, MaxUDPSize+100)
	c := &core.Container{
		SenderPubKey: id.Bytes(),
		Payload:      largePayload,
	}
	c.Sign(priv)
	
	err := adapter.Send(c, []string{"127.0.0.1:8080"})
	if err == nil {
		t.Fatalf("Expected MTU error for large container, got nil")
	}
}
