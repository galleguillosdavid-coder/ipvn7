package adapters

import (
	"testing"
	"time"

	"ipv7/core"
)

func TestRelayServerAndAdapter(t *testing.T) {
	// 1. Start Relay Server on ephemeral TCP port
	server := NewRelayServer("127.0.0.1:0")
	if err := server.Start(); err != nil {
		t.Fatalf("Failed to start relay server: %v", err)
	}
	defer server.Stop()

	serverAddr := server.Addr().String()

	// 2. Setup Node A and Node B identities
	idA, privA, _ := core.GenerateIdentity()
	idB, privB, _ := core.GenerateIdentity()
	_ = privB

	// 3. Connect RelayAdapters for Node A and Node B
	adapterA := NewRelayAdapter(serverAddr, idA)
	if err := adapterA.Start(); err != nil {
		t.Fatalf("Failed to start relay adapter A: %v", err)
	}
	defer adapterA.Stop()

	adapterB := NewRelayAdapter(serverAddr, idB)
	if err := adapterB.Start(); err != nil {
		t.Fatalf("Failed to start relay adapter B: %v", err)
	}
	defer adapterB.Stop()

	// Wait for registration
	time.Sleep(50 * time.Millisecond)

	// 4. Create and sign Container from A addressed to B
	payload := []byte("Hello Node B via DERP-like Fallback Relay!")
	c := &core.Container{
		SenderPubKey:   idA.Bytes(),
		ReceiverPubKey: idB.Bytes(),
		Payload:        payload,
	}
	if err := c.Sign(privA); err != nil {
		t.Fatalf("Failed to sign container: %v", err)
	}

	// 5. Send container through relay adapter
	if err := adapterA.Send(c, nil); err != nil {
		t.Fatalf("Failed to send container via relay: %v", err)
	}

	// 6. Verify delivery on Node B's relay adapter
	select {
	case received := <-adapterB.Receive():
		if string(received.Payload) != string(payload) {
			t.Fatalf("Payload mismatch: got %q, want %q", string(received.Payload), string(payload))
		}
		if !received.Verify() {
			t.Fatalf("Received container failed cryptographic signature verification")
		}
		t.Logf("Success! Container relayed and verified through DERP server: %s", string(received.Payload))
	case <-time.After(3 * time.Second):
		t.Fatalf("Timed out waiting for relayed container on Node B")
	}
}
