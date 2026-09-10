package adapters

import (
	"testing"
	"time"
)

func TestHierarchicalQoS_IngressAndBurst(t *testing.T) {
	qos := NewHierarchicalQoS()
	peerDID := "did:ipv7:alice_test_node_01"

	// 1. First packet within burst capacity should be allowed immediately
	admitted, ch := qos.IngressCheck(peerDID, PriorityInteractive, 1024)
	if !admitted || ch != nil {
		t.Fatalf("Expected first 1KB packet to be admitted, got admitted=%v, ch=%v", admitted, ch)
	}

	// 2. Consume the rest of the interactive burst capacity (250 KB - 1024 B)
	remaining := int(qos.defaultCapacities[PriorityInteractive]) - 1024
	admitted, _ = qos.IngressCheck(peerDID, PriorityInteractive, remaining)
	if !admitted {
		t.Fatalf("Expected remaining burst capacity to be admitted")
	}

	// 3. Next packet should now exceed capacity, drop, and issue an anti-DDoS PoW challenge
	admitted, ch = qos.IngressCheck(peerDID, PriorityInteractive, 1024)
	if admitted {
		t.Fatalf("Expected rate limit to trigger after exhausting capacity")
	}
	if ch == nil {
		t.Fatalf("Expected a PoWChallenge to be issued upon drop")
	}
	if ch.TargetDID != peerDID {
		t.Fatalf("Challenge target DID mismatch: expected %s, got %s", peerDID, ch.TargetDID)
	}
}

func TestHierarchicalQoS_PoWChallengeResolution(t *testing.T) {
	qos := NewHierarchicalQoS()
	peerDID := "did:ipv7:bob_challenger"

	// Drain interactive bucket completely
	cap := int(qos.defaultCapacities[PriorityInteractive])
	qos.IngressCheck(peerDID, PriorityInteractive, cap)

	// Attempt packet -> should drop and get challenge
	admitted, ch := qos.IngressCheck(peerDID, PriorityInteractive, 1024)
	if admitted || ch == nil {
		t.Fatalf("Failed to trigger PoW challenge: admitted=%v, ch=%v", admitted, ch)
	}

	// Solve challenge
	nonce := SolveChallenge(ch)

	// Verify solution and credit bonus tokens
	success := qos.VerifyAndCredit(ch.ChallengeID, nonce)
	if !success {
		t.Fatalf("Failed to verify valid PoW solution with nonce=%d", nonce)
	}

	// Peer should now have received bonus tokens (100 KB) and packet admitted
	admitted, _ = qos.IngressCheck(peerDID, PriorityInteractive, 1024)
	if !admitted {
		t.Fatalf("Expected peer to be admitted after solving PoW challenge")
	}
}

func TestHierarchicalQoS_ReplenishOverTime(t *testing.T) {
	// Create token bucket with 10 KB/s rate and 10 KB capacity
	tb := NewTokenBucket(10*1024, 10*1024)

	// Drain bucket completely
	if !tb.Allow(10 * 1024) {
		t.Fatalf("Failed to drain token bucket")
	}
	if tb.Allow(1024) {
		t.Fatalf("Bucket should be empty")
	}

	// Wait 150ms -> should replenish approx 1.5 KB
	time.Sleep(150 * time.Millisecond)
	if !tb.Allow(1024) {
		t.Fatalf("Bucket should have replenished at least 1 KB")
	}
}
