package adapters

import (
	"testing"
)

func TestReplayWindowSequential(t *testing.T) {
	w := &ReplayWindow{}

	// Sequential packets 1 to 100 must all be accepted
	for i := uint64(1); i <= 100; i++ {
		if !w.CheckAndSet(i) {
			t.Fatalf("Expected packet %d to be accepted", i)
		}
	}
}

func TestReplayWindowDuplicatesRejected(t *testing.T) {
	w := &ReplayWindow{}

	if !w.CheckAndSet(10) {
		t.Fatalf("Expected seq 10 to be accepted")
	}

	// Immediate replay of 10 must be rejected
	if w.CheckAndSet(10) {
		t.Fatalf("Expected replay of seq 10 to be rejected")
	}

	// Advance to 15
	if !w.CheckAndSet(15) {
		t.Fatalf("Expected seq 15 to be accepted")
	}

	// Replay 10 again must be rejected
	if w.CheckAndSet(10) {
		t.Fatalf("Expected replay of seq 10 to be rejected")
	}

	// Replay 15 must be rejected
	if w.CheckAndSet(15) {
		t.Fatalf("Expected replay of seq 15 to be rejected")
	}
}

func TestReplayWindowOutOfOrderWithinWindow(t *testing.T) {
	w := &ReplayWindow{}

	// Packet 20 arrives first
	if !w.CheckAndSet(20) {
		t.Fatalf("Expected seq 20 to be accepted")
	}

	// Packets arrive out of order: 18, 15, 19, 16
	outOfOrder := []uint64{18, 15, 19, 16}
	for _, seq := range outOfOrder {
		if !w.CheckAndSet(seq) {
			t.Fatalf("Expected out-of-order seq %d within window to be accepted", seq)
		}
	}

	// Re-sending any of them must now be rejected
	for _, seq := range outOfOrder {
		if w.CheckAndSet(seq) {
			t.Fatalf("Expected duplicate seq %d to be rejected", seq)
		}
	}
}

func TestReplayWindowTooOldRejected(t *testing.T) {
	w := &ReplayWindow{}

	// Max seq advances to 100
	if !w.CheckAndSet(100) {
		t.Fatalf("Expected seq 100 to be accepted")
	}

	// Seq 35 is 65 positions behind (WindowSize = 64); must be rejected
	if w.CheckAndSet(35) {
		t.Fatalf("Expected seq 35 (out of window) to be rejected")
	}

	// Seq 36 is exactly 64 positions behind; must be rejected
	if w.CheckAndSet(36) {
		t.Fatalf("Expected seq 36 to be rejected")
	}

	// Seq 37 is 63 positions behind (within window); must be accepted first time
	if !w.CheckAndSet(37) {
		t.Fatalf("Expected seq 37 (inside window) to be accepted")
	}

	// Second time seq 37 must be rejected
	if w.CheckAndSet(37) {
		t.Fatalf("Expected duplicate seq 37 to be rejected")
	}
}

func TestAntiReplayTablePerPeerIsolation(t *testing.T) {
	table := NewAntiReplayTable()

	peerA := "peerA_identity_hash_hex"
	peerB := "peerB_identity_hash_hex"

	// Both peers can send packet seq 5 without interfering
	if !table.CheckAndSet(peerA, 5) {
		t.Fatalf("Expected peer A seq 5 to be accepted")
	}
	if !table.CheckAndSet(peerB, 5) {
		t.Fatalf("Expected peer B seq 5 to be accepted")
	}

	// Replay for peer A must be rejected, while peer B can send seq 6
	if table.CheckAndSet(peerA, 5) {
		t.Fatalf("Expected duplicate peer A seq 5 to be rejected")
	}
	if !table.CheckAndSet(peerB, 6) {
		t.Fatalf("Expected peer B seq 6 to be accepted")
	}
}
