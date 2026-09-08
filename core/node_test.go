package core

import (
	"sync"
	"testing"
	"time"
)

// mockAdapter is an in-memory test adapter simulating network transfer
type mockAdapter struct {
	addr    string
	receive chan *Container
	peers   map[string]*mockAdapter
	mu      sync.Mutex
}

func newMockAdapter(addr string, registry map[string]*mockAdapter) *mockAdapter {
	a := &mockAdapter{
		addr:    addr,
		receive: make(chan *Container, 10),
		peers:   registry,
	}
	registry[addr] = a
	return a
}

func (m *mockAdapter) Start() error { return nil }
func (m *mockAdapter) Stop() error  { return nil }
func (m *mockAdapter) Send(c *Container, endpoints []string) error {
	for _, ep := range endpoints {
		if peer, exists := m.peers[ep]; exists {
			peer.receive <- c
			return nil
		}
	}
	return nil
}
func (m *mockAdapter) Receive() <-chan *Container { return m.receive }

func TestNodeP2PCommunication(t *testing.T) {
	registry := make(map[string]*mockAdapter)

	// Node A
	idA, privA, _ := GenerateIdentity()
	nodeA := NewNode(idA, privA)
	adapterA := newMockAdapter("addrA", registry)
	nodeA.AddAdapter(adapterA)

	// Node B
	idB, privB, _ := GenerateIdentity()
	nodeB := NewNode(idB, privB)
	adapterB := newMockAdapter("addrB", registry)
	nodeB.AddAdapter(adapterB)

	// Register B's endpoint in A's peer table
	nodeA.AddPeer(idB, []string{"addrB"})

	// Setup message receiver on Node B
	receivedMsg := make(chan string, 1)
	nodeB.OnMessage(func(from Identity, payload []byte) {
		if from.String() == idA.String() {
			receivedMsg <- string(payload)
		}
	})

	if err := nodeA.Start(); err != nil {
		t.Fatalf("Failed to start node A: %v", err)
	}
	defer nodeA.Stop()

	if err := nodeB.Start(); err != nil {
		t.Fatalf("Failed to start node B: %v", err)
	}
	defer nodeB.Stop()

	// Send message from A to B
	err := nodeA.SendMessage(idB, []byte("Hello Node B from Node A via IPv7!"))
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	select {
	case msg := <-receivedMsg:
		if msg != "Hello Node B from Node A via IPv7!" {
			t.Fatalf("Received unexpected message: %s", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Timed out waiting for message on Node B")
	}
}

func TestNodeHandshakeAndE2EE(t *testing.T) {
	registry := make(map[string]*mockAdapter)

	// Node A
	idA, privA, _ := GenerateIdentity()
	nodeA := NewNode(idA, privA)
	adapterA := newMockAdapter("addrA", registry)
	nodeA.AddAdapter(adapterA)
	nodeA.SetEndpoints([]string{"addrA"})

	// Node B
	idB, privB, _ := GenerateIdentity()
	nodeB := NewNode(idB, privB)
	adapterB := newMockAdapter("addrB", registry)
	nodeB.AddAdapter(adapterB)
	nodeB.SetEndpoints([]string{"addrB"})

	if err := nodeA.Start(); err != nil {
		t.Fatalf("Failed to start node A: %v", err)
	}
	defer nodeA.Stop()

	if err := nodeB.Start(); err != nil {
		t.Fatalf("Failed to start node B: %v", err)
	}
	defer nodeB.Stop()

	// 1. Node A initiates handshake with Node B's endpoint
	discoveredID, rtt, err := nodeA.Handshake("addrB")
	if err != nil {
		t.Fatalf("Handshake failed: %v", err)
	}
	if discoveredID.String() != idB.String() {
		t.Fatalf("Discovered ID mismatch: got %s, want %s", discoveredID.String(), idB.String())
	}
	t.Logf("Handshake verified! RTT: %v", rtt)

	// Verify both nodes learned each other's X25519 keys
	keyForB := nodeA.GetPeerEncKey(idB)
	if len(keyForB) != 32 {
		t.Fatalf("Node A did not store Node B's X25519 key")
	}
	keyForA := nodeB.GetPeerEncKey(idA)
	if len(keyForA) != 32 {
		t.Fatalf("Node B did not store Node A's X25519 key")
	}

	// 2. Node B listens for encrypted message and decrypts it
	decryptedChan := make(chan string, 1)
	nodeB.OnMessage(func(from Identity, payload []byte) {
		decrypted, decErr := nodeB.DecryptMessage(payload)
		if decErr != nil {
			t.Errorf("Node B failed to decrypt: %v", decErr)
			return
		}
		decryptedChan <- string(decrypted)
	})

	// 3. Node A sends E2EE encrypted message to Node B
	secretText := "¡Hola notebook desde IPv7 E2EE seguro!"
	if err := nodeA.SendEncryptedMessage(idB, keyForB, []byte(secretText)); err != nil {
		t.Fatalf("Failed to send encrypted message: %v", err)
	}

	select {
	case result := <-decryptedChan:
		if result != secretText {
			t.Fatalf("Decrypted text mismatch: got %s, want %s", result, secretText)
		}
		t.Logf("E2EE successfully decrypted on Node B: %s", result)
	case <-time.After(2 * time.Second):
		t.Fatalf("Timed out waiting for decrypted message on Node B")
	}
}
