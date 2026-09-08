package adapters

import (
	"sync"
)

const WindowSize = 64

// ReplayWindow implements an RFC 6479 / IPsec-style 64-bit sliding window filter
type ReplayWindow struct {
	maxSeq uint64
	bitmap uint64
}

// CheckAndSet returns true if the sequence number is valid and not replayed.
// If valid, it updates the window state atomically.
func (w *ReplayWindow) CheckAndSet(seq uint64) bool {
	// First packet initialized
	if w.maxSeq == 0 && w.bitmap == 0 {
		w.maxSeq = seq
		w.bitmap = 1
		return true
	}

	// Packet is newer than the highest seen
	if seq > w.maxSeq {
		diff := seq - w.maxSeq
		if diff < WindowSize {
			w.bitmap = (w.bitmap << diff) | 1
		} else {
			// Jump forward beyond window size
			w.bitmap = 1
		}
		w.maxSeq = seq
		return true
	}

	// Packet is older than or equal to highest seen
	diff := w.maxSeq - seq
	if diff >= WindowSize {
		// Too old, falls outside window
		return false
	}

	mask := uint64(1) << diff
	if (w.bitmap & mask) != 0 {
		// Already seen packet (duplicate / replay)
		return false
	}

	// Within window and not seen before; mark as seen
	w.bitmap |= mask
	return true
}

// AntiReplayTable manages sliding windows per peer sender identity
type AntiReplayTable struct {
	mu      sync.Mutex
	windows map[string]*ReplayWindow
}

// NewAntiReplayTable creates an initialized AntiReplayTable
func NewAntiReplayTable() *AntiReplayTable {
	return &AntiReplayTable{
		windows: make(map[string]*ReplayWindow),
	}
}

// CheckAndSet verifies if the given sequence number from peerID is valid and marks it
func (t *AntiReplayTable) CheckAndSet(peerID string, seq uint64) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	w, exists := t.windows[peerID]
	if !exists {
		w = &ReplayWindow{}
		t.windows[peerID] = w
	}
	return w.CheckAndSet(seq)
}

// Reset clears state for a specific peer (e.g. on new cryptographic session)
func (t *AntiReplayTable) Reset(peerID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.windows, peerID)
}
