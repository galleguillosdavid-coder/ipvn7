package core

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"
)

func TestSelfHealingSupervisor(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	id := &Ed25519Identity{PublicKey: pub}
	node := NewNode(id, priv)
	supervisor := NewSelfHealingSupervisor(node, 100*time.Millisecond)

	var actionsCaptured []string
	supervisor.OnAction(func(action string, details map[string]interface{}) {
		actionsCaptured = append(actionsCaptured, action)
	})

	// 1. Initial check on empty table
	res := supervisor.CheckAndHeal()
	if res["total_peers"].(int) != 0 {
		t.Fatalf("expected 0 peers, got %v", res["total_peers"])
	}

	// 2. Add peer with acceptable latency
	pubPeer1, _, _ := ed25519.GenerateKey(rand.Reader)
	id1 := &Ed25519Identity{PublicKey: pubPeer1}
	node.SmallWorld.AddPeer(id1, []string{"192.168.1.50:7890"}, 15*time.Millisecond)

	// 3. Add peer with degraded latency (> 800ms)
	pubPeer2, _, _ := ed25519.GenerateKey(rand.Reader)
	id2 := &Ed25519Identity{PublicKey: pubPeer2}
	node.SmallWorld.AddPeer(id2, []string{"10.0.0.99:7890"}, 1200*time.Millisecond)

	res2 := supervisor.CheckAndHeal()
	if res2["stale_peers"].(int) != 1 {
		t.Fatalf("expected 1 stale peer, got %v", res2["stale_peers"])
	}

	foundPurge := false
	for _, a := range actionsCaptured {
		if a == "purge_stale_peers" {
			foundPurge = true
			break
		}
	}
	if !foundPurge {
		t.Fatalf("expected action purge_stale_peers in %v", actionsCaptured)
	}

	actions, _, _ := supervisor.Stats()
	if actions == 0 {
		t.Fatalf("expected heal actions > 0, got %d", actions)
	}
}
