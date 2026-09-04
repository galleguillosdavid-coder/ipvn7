package core

import (
	"sync"
	"testing"
	"time"
)

func TestCascadeStreamingMultiHop(t *testing.T) {
	registry := make(map[string]*mockAdapter)

	createNode := func(name string) (*Node, *CascadeNode) {
		id, priv, _ := GenerateIdentity()
		n := NewNode(id, priv)
		a := newMockAdapter(name, registry)
		n.AddAdapter(a)
		_ = n.Start()
		cn := NewCascadeNode(n, 10) // Max 10 children per tier
		return n, cn
	}

	// 1. Root broadcaster
	_, rootCascade := createNode("root")

	// 2. Tier 1 (2 children nodes)
	nodeChild1, cascadeChild1 := createNode("child1")
	nodeChild2, cascadeChild2 := createNode("child2")
	_ = cascadeChild2

	_ = rootCascade.AddChild(nodeChild1.Identity, "child1")
	_ = rootCascade.AddChild(nodeChild2.Identity, "child2")

	// 3. Tier 2 (Grandchildren nodes attached to Child 1)
	nodeGC1, cascadeGC1 := createNode("gc1")
	nodeGC2, cascadeGC2 := createNode("gc2")

	_ = cascadeChild1.AddChild(nodeGC1.Identity, "gc1")
	_ = cascadeChild1.AddChild(nodeGC2.Identity, "gc2")

	// Track received chunks across grandchildren
	var wg sync.WaitGroup
	wg.Add(2)

	receivedGC1 := make(chan string, 1)
	cascadeGC1.OnChunk(func(chunk *StreamChunk) {
		receivedGC1 <- string(chunk.Payload)
		wg.Done()
	})

	receivedGC2 := make(chan string, 1)
	cascadeGC2.OnChunk(func(chunk *StreamChunk) {
		receivedGC2 <- string(chunk.Payload)
		wg.Done()
	})

	// 4. Root broadcasts a stream chunk (e.g. Video Keyframe)
	streamData := []byte("STREAM_CHUNK_VIDEO_FRAME_001")
	err := rootCascade.Broadcast("live-channel-1", 1, streamData)
	if err != nil {
		t.Fatalf("Failed to broadcast chunk: %v", err)
	}

	// 5. Wait for chunk to cascade through the tree
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
		msg1 := <-receivedGC1
		msg2 := <-receivedGC2
		if msg1 != string(streamData) || msg2 != string(streamData) {
			t.Fatalf("Stream payload mismatch in grandchildren: %s, %s", msg1, msg2)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("Timed out waiting for cascade stream to reach grandchildren")
	}
}

func TestCascadeMaxChildrenEnforcement(t *testing.T) {
	registry := make(map[string]*mockAdapter)

	idRoot, privRoot, _ := GenerateIdentity()
	nodeRoot := NewNode(idRoot, privRoot)
	nodeRoot.AddAdapter(newMockAdapter("root", registry))
	_ = nodeRoot.Start()

	// Cascade with limit of 2 children
	cascade := NewCascadeNode(nodeRoot, 2)

	id1, _, _ := GenerateIdentity()
	id2, _, _ := GenerateIdentity()
	id3, _, _ := GenerateIdentity()

	if err := cascade.AddChild(id1, "addr1"); err != nil {
		t.Fatalf("Failed to add child 1: %v", err)
	}
	if err := cascade.AddChild(id2, "addr2"); err != nil {
		t.Fatalf("Failed to add child 2: %v", err)
	}

	// Third child should fail due to max capacity
	err := cascade.AddChild(id3, "addr3")
	if err == nil {
		t.Fatalf("Expected error adding 3rd child when limit is 2, got nil")
	}
}
