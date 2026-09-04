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
