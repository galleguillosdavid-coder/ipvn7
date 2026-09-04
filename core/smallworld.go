package core

import (
	"bytes"
	"crypto/subtle"
	"math/bits"
	"sort"
	"sync"
	"time"
)

// Small-World constants based on 12 degrees of separation and 10 peers per degree
const (
	DefaultMaxDegrees      = 12 // Maximum degrees of separation across the planet
	DefaultPeersPerDegree  = 10 // Maximum peers kept per degree bucket (fanout factor = 10)
	DefaultHopLimit  uint8 = 12 // Initial TTL / HopLimit for all routed containers
)

// PeerEntry represents a known peer in the small-world routing table
type PeerEntry struct {
	Identity  Identity
	Endpoints []string
	Degree    int       // 0 = direct local neighbor, 1..12 = logarithmic distance ring
	Latency   time.Duration
	LastSeen  time.Time
}

// SmallWorldTable implements a Kleinberg Small-World bounded routing table
type SmallWorldTable struct {
	localID        Identity
	maxDegrees     int
	peersPerDegree int

	// buckets holds peers partitioned into distance rings (0 to maxDegrees)
	buckets [][]PeerEntry
	mu      sync.RWMutex
}

// NewSmallWorldTable creates a bounded routing table for a node
func NewSmallWorldTable(localID Identity, maxDegrees, peersPerDegree int) *SmallWorldTable {
	if maxDegrees <= 0 {
		maxDegrees = DefaultMaxDegrees
	}
	if peersPerDegree <= 0 {
		peersPerDegree = DefaultPeersPerDegree
	}

	table := &SmallWorldTable{
		localID:        localID,
		maxDegrees:     maxDegrees,
		peersPerDegree: peersPerDegree,
		buckets:        make([][]PeerEntry, maxDegrees+1),
	}

	for i := range table.buckets {
		table.buckets[i] = make([]PeerEntry, 0, peersPerDegree)
	}

	return table
}

// AddPeer inserts or updates a peer in the appropriate degree ring
func (t *SmallWorldTable) AddPeer(peer Identity, endpoints []string, latency time.Duration) {
	if bytes.Equal(peer.Bytes(), t.localID.Bytes()) {
		return // Do not add self
	}

	degree := t.CalculateDegree(peer)

	t.mu.Lock()
	defer t.mu.Unlock()

	bucket := t.buckets[degree]

	// Check if peer already exists in this bucket; if so, update
	for i, entry := range bucket {
		if bytes.Equal(entry.Identity.Bytes(), peer.Bytes()) {
			bucket[i].Endpoints = endpoints
			bucket[i].Latency = latency
			bucket[i].LastSeen = time.Now()
			return
		}
	}

	// If bucket has capacity (< peersPerDegree), append
	if len(bucket) < t.peersPerDegree {
		t.buckets[degree] = append(bucket, PeerEntry{
			Identity:  peer,
			Endpoints: endpoints,
			Degree:    degree,
			Latency:   latency,
			LastSeen:  time.Now(),
		})
		return
	}

	// Bucket is full: replace peer with highest latency if this peer is faster
	worstIdx := -1
	var worstLatency time.Duration
	for i, entry := range bucket {
		if worstIdx == -1 || entry.Latency > worstLatency {
			worstIdx = i
			worstLatency = entry.Latency
		}
	}

	if latency < worstLatency {
		bucket[worstIdx] = PeerEntry{
			Identity:  peer,
			Endpoints: endpoints,
			Degree:    degree,
			Latency:   latency,
			LastSeen:  time.Now(),
		}
	}
}

// GetPeer looks up a peer directly by identity
func (t *SmallWorldTable) GetPeer(peer Identity) (*PeerEntry, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	degree := t.CalculateDegree(peer)
	for _, entry := range t.buckets[degree] {
		if bytes.Equal(entry.Identity.Bytes(), peer.Bytes()) {
			copyEntry := entry
			return &copyEntry, true
		}
	}
	return nil, false
}

// TotalPeers returns the total number of entries currently stored in the table
func (t *SmallWorldTable) TotalPeers() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	total := 0
	for _, bucket := range t.buckets {
		total += len(bucket)
	}
	return total
}

// CalculateDegree maps the XOR distance between localID and targetID into [0..maxDegrees]
func (t *SmallWorldTable) CalculateDegree(target Identity) int {
	dist := XorDistance(t.localID.Bytes(), target.Bytes())
	leadingZeros := LeadingZeroBits(dist)

	// In a 256-bit key space (Ed25519), map [0..256] leading zeros into [0..maxDegrees]
	degree := t.maxDegrees - (leadingZeros * t.maxDegrees / 256)
	if degree < 0 {
		degree = 0
	}
	if degree > t.maxDegrees {
		degree = t.maxDegrees
	}
	return degree
}

// FindClosestPeers finds the k peers closest in XOR distance to targetID (Greedy Routing)
func (t *SmallWorldTable) FindClosestPeers(target Identity, count int) []PeerEntry {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var allPeers []PeerEntry
	for _, bucket := range t.buckets {
		allPeers = append(allPeers, bucket...)
	}

	targetBytes := target.Bytes()

	// Sort peers by XOR distance to target
	sort.Slice(allPeers, func(i, j int) bool {
		distI := XorDistance(allPeers[i].Identity.Bytes(), targetBytes)
		distJ := XorDistance(allPeers[j].Identity.Bytes(), targetBytes)
		return CompareDistance(distI, distJ) < 0
	})

	if len(allPeers) > count {
		allPeers = allPeers[:count]
	}
	return allPeers
}

// XorDistance computes bitwise XOR between two byte slices of equal length
func XorDistance(a, b []byte) []byte {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	res := make([]byte, n)
	for i := 0; i < n; i++ {
		res[i] = a[i] ^ b[i]
	}
	return res
}

// LeadingZeroBits returns the number of consecutive zero bits at the beginning of a slice
func LeadingZeroBits(data []byte) int {
	zeros := 0
	for _, b := range data {
		if b == 0 {
			zeros += 8
		} else {
			zeros += bits.LeadingZeros8(b)
			break
		}
	}
	return zeros
}

// CompareDistance compares two XOR distance byte slices (-1 if a < b, 1 if a > b, 0 if equal)
func CompareDistance(a, b []byte) int {
	return bytes.Compare(a, b)
}

// EqualIdentities checks if two identities are identical in constant time
func EqualIdentities(a, b Identity) bool {
	if a == nil || b == nil {
		return false
	}
	return subtle.ConstantTimeCompare(a.Bytes(), b.Bytes()) == 1
}
