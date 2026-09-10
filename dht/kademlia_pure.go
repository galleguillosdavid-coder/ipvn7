package dht

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"math/bits"
	"sort"
	"sync"
	"time"

	"ipv7/core"
)

const (
	// KBucketSize standard Kademlia bucket capacity (k = 20)
	KBucketSize = 20
	// AlphaConcurrency concurrent query factor (alpha = 3)
	AlphaConcurrency = 3
	// KeyBits length of identity keys in bits (256 bits for Ed25519 / SHA-256)
	KeyBits = 256
)

// Contact represents a known node in the Kademlia routing tree
type Contact struct {
	ID        core.Identity
	Endpoints []string
	LastSeen  time.Time
}

// KBucket holds up to k contacts sharing a common prefix distance
type KBucket struct {
	contacts []Contact
	mu       sync.RWMutex
}

// PureKademliaTable implements a full 256-bucket Kademlia routing table
type PureKademliaTable struct {
	localID core.Identity
	buckets [KeyBits]*KBucket
	mu      sync.RWMutex
}

// NewPureKademliaTable initializes a 256-bit Kademlia table
func NewPureKademliaTable(localID core.Identity) *PureKademliaTable {
	t := &PureKademliaTable{
		localID: localID,
	}
	for i := 0; i < KeyBits; i++ {
		t.buckets[i] = &KBucket{
			contacts: make([]Contact, 0, KBucketSize),
		}
	}
	return t
}

// XOR calculates bitwise XOR distance between two 32-byte slices
func XOR(a, b []byte) [32]byte {
	var res [32]byte
	for i := 0; i < 32 && i < len(a) && i < len(b); i++ {
		res[i] = a[i] ^ b[i]
	}
	return res
}

// BucketIndex calculates the leading zeros bucket index (0 to 255)
func (t *PureKademliaTable) BucketIndex(targetID core.Identity) int {
	dist := XOR(t.localID.Bytes(), targetID.Bytes())
	for byteIdx, b := range dist {
		if b != 0 {
			lz := bits.LeadingZeros8(b)
			idx := byteIdx*8 + lz
			if idx >= KeyBits {
				return KeyBits - 1
			}
			return idx
		}
	}
	return KeyBits - 1
}

// AddContact inserts or updates a peer contact into the appropriate k-bucket
func (t *PureKademliaTable) AddContact(id core.Identity, endpoints []string) {
	if bytes.Equal(t.localID.Bytes(), id.Bytes()) {
		return // Do not store self
	}

	bIdx := t.BucketIndex(id)
	bucket := t.buckets[bIdx]

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// 1. Check if contact already exists: move to tail (most recently seen)
	for i, c := range bucket.contacts {
		if bytes.Equal(c.ID.Bytes(), id.Bytes()) {
			bucket.contacts[i].LastSeen = time.Now()
			if len(endpoints) > 0 {
				bucket.contacts[i].Endpoints = endpoints
			}
			// Move to end (LRU)
			elem := bucket.contacts[i]
			bucket.contacts = append(append(bucket.contacts[:i], bucket.contacts[i+1:]...), elem)
			return
		}
	}

	// 2. If bucket has room, append
	if len(bucket.contacts) < KBucketSize {
		bucket.contacts = append(bucket.contacts, Contact{
			ID:        id,
			Endpoints: endpoints,
			LastSeen:  time.Now(),
		})
		return
	}

	// 3. Bucket full: LRU eviction check (in pure Kademlia, ping head; if alive, drop new, else replace)
	// For high-churn P2P, we replace the oldest contact if last seen > 1 hour ago
	if time.Since(bucket.contacts[0].LastSeen) > time.Hour {
		copy(bucket.contacts[0:], bucket.contacts[1:])
		bucket.contacts[KBucketSize-1] = Contact{
			ID:        id,
			Endpoints: endpoints,
			LastSeen:  time.Now(),
		}
	}
}

// FindClosest returns the count closest contacts to targetID sorted by XOR distance
func (t *PureKademliaTable) FindClosest(targetID core.Identity, count int) []Contact {
	if count <= 0 {
		count = KBucketSize
	}

	type contactDistance struct {
		contact  Contact
		distance [32]byte
	}

	var all []contactDistance

	for _, b := range t.buckets {
		b.mu.RLock()
		for _, c := range b.contacts {
			dist := XOR(c.ID.Bytes(), targetID.Bytes())
			all = append(all, contactDistance{
				contact:  c,
				distance: dist,
			})
		}
		b.mu.RUnlock()
	}

	// Sort by XOR distance
	sort.Slice(all, func(i, j int) bool {
		return bytes.Compare(all[i].distance[:], all[j].distance[:]) < 0
	})

	limit := count
	if len(all) < limit {
		limit = len(all)
	}

	result := make([]Contact, limit)
	for i := 0; i < limit; i++ {
		result[i] = all[i].contact
	}
	return result
}

// TotalContacts returns the total number of unique peers tracked in all k-buckets
func (t *PureKademliaTable) TotalContacts() int {
	total := 0
	for _, b := range t.buckets {
		b.mu.RLock()
		total += len(b.contacts)
		b.mu.RUnlock()
	}
	return total
}

// ComputePoW calculates a light proof-of-work (difficulty bits) for anti-Sybil record publication
func ComputePoW(payload []byte, difficulty int) uint64 {
	var nonce uint64
	for {
		h := sha256.New()
		h.Write(payload)
		var nBuf [8]byte
		binary.LittleEndian.PutUint64(nBuf[:], nonce)
		h.Write(nBuf[:])
		sum := h.Sum(nil)

		// Check leading zero bits
		zeros := 0
		for _, b := range sum {
			lz := bits.LeadingZeros8(b)
			zeros += lz
			if lz < 8 {
				break
			}
		}

		if zeros >= difficulty {
			return nonce
		}
		nonce++
	}
}

// VerifyPoW verifies that payload + nonce yields the requested number of leading zero bits
func VerifyPoW(payload []byte, nonce uint64, difficulty int) bool {
	h := sha256.New()
	h.Write(payload)
	var nBuf [8]byte
	binary.LittleEndian.PutUint64(nBuf[:], nonce)
	h.Write(nBuf[:])
	sum := h.Sum(nil)

	zeros := 0
	for _, b := range sum {
		lz := bits.LeadingZeros8(b)
		zeros += lz
		if lz < 8 {
			break
		}
	}
	return zeros >= difficulty
}
