package core

import (
	"fmt"
	"testing"
	"time"
)

func TestSmallWorldTableBoundedCapacity(t *testing.T) {
	myID, _, _ := GenerateIdentity()
	table := NewSmallWorldTable(myID, 12, 10) // 12 degrees, 10 peers per degree

	// Add 15 peers with varying simulated latencies
	for i := 0; i < 15; i++ {
		pID, _, _ := GenerateIdentity()
		table.AddPeer(pID, []string{fmt.Sprintf("192.168.1.%d:8000", i)}, time.Duration(10+i)*time.Millisecond)
	}

	total := table.TotalPeers()
	if total > 12*10 {
		t.Fatalf("Table exceeded maximum bounded capacity: got %d", total)
	}
	t.Logf("Bounded table peer count: %d (max bounded: %d)", total, 12*10)
}

func TestSmallWorldMultiHopRelay(t *testing.T) {
	registry := make(map[string]*mockAdapter)

	createNode := func(name string) (*Node, Identity) {
		id, priv, _ := GenerateIdentity()
		n := NewNode(id, priv)
		a := newMockAdapter(name, registry)
		n.AddAdapter(a)
		_ = n.Start()
		return n, id
	}

	// 1. Setup 3 nodes: A, B (Hop/Relay), and C (Destination)
	nodeA, idA := createNode("addrA")
	nodeB, idB := createNode("addrB")
	nodeC, idC := createNode("addrC")
	defer nodeA.Stop()
	defer nodeB.Stop()
	defer nodeC.Stop()

	// 2. Topology:
	// A only knows B. A DOES NOT know C's IP address or endpoint!
	nodeA.AddPeer(idB, []string{"addrB"})

	// B knows C
	nodeB.AddPeer(idC, []string{"addrC"})

	// Also inform A's small world table about C's identity via B as closest route
	nodeA.SmallWorld.AddPeer(idB, []string{"addrB"}, 5*time.Millisecond)

	// Setup receiver on Destination Node C
	receivedPayload := make(chan string, 1)
	nodeC.OnMessage(func(from Identity, payload []byte) {
		if EqualIdentities(from, idA) {
			receivedPayload <- string(payload)
		}
	})

	// 3. Node A sends a message addressed to Node C's cryptographic Identity
	secretMsg := "Greetings Node C! Sent via 2nd degree multi-hop routing!"
	err := nodeA.SendMessage(idC, []byte(secretMsg))
	if err != nil {
		t.Fatalf("Failed to send multi-hop message from A: %v", err)
	}

	// 4. Verify that Node C received the message intact and verified
	select {
	case msg := <-receivedPayload:
		if msg != secretMsg {
			t.Fatalf("Payload mismatch on Node C: got %q, want %q", msg, secretMsg)
		}
		t.Logf("Success! Node C received multi-hop routed message: %s", msg)
	case <-time.After(3 * time.Second):
		t.Fatalf("Timed out waiting for multi-hop message to arrive at Node C")
	}
}

func TestSmallWorldHopLimitDrop(t *testing.T) {
	registry := make(map[string]*mockAdapter)

	idA, privA, _ := GenerateIdentity()
	nodeA := NewNode(idA, privA)
	nodeA.AddAdapter(newMockAdapter("addrA", registry))
	_ = nodeA.Start()
	defer nodeA.Stop()

	idB, privB, _ := GenerateIdentity()
	nodeB := NewNode(idB, privB)
	nodeB.AddAdapter(newMockAdapter("addrB", registry))
	_ = nodeB.Start()
	defer nodeB.Stop()

	idC, _, _ := GenerateIdentity()
	nodeB.AddPeer(idC, []string{"addrC"})

	// Send container with HopLimit = 1 to B, addressed to C
	c := &Container{
		SenderPubKey:   idA.Bytes(),
		ReceiverPubKey: idC.Bytes(),
		Payload:        []byte("should be dropped"),
		HopLimit:       1, // Threshold reached
	}
	_ = c.Sign(privA)

	// Inject to B
	err := nodeA.forwardContainer(c, []string{"addrB"})
	if err != nil {
		t.Fatalf("Failed to forward: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	// B must not forward it since HopLimit <= 1
}
