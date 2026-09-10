package adapters

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// QoSPriority defines the traffic class in the ipvn7 Network Operating System.
type QoSPriority int

const (
	// PriorityControl: Handshakes, DHT pings, Sphinx routing headers (Highest priority).
	PriorityControl QoSPriority = 0
	// PriorityInteractive: SSH, VoIP, terminal interactive sessions.
	PriorityInteractive QoSPriority = 1
	// PriorityBulk: Large file transfers, store-and-forward DAG sync, background telemetry.
	PriorityBulk QoSPriority = 2
)

// TokenBucket implements a single-rate token bucket filter.
type TokenBucket struct {
	Capacity   float64   // Maximum burst capacity (tokens/bytes)
	Rate       float64   // Token replenishment rate (tokens/bytes per second)
	Tokens     float64   // Current available tokens
	LastUpdate time.Time // Timestamp of last replenishment
	mu         sync.Mutex
}

// NewTokenBucket creates an initialized token bucket with full capacity.
func NewTokenBucket(rate float64, capacity float64) *TokenBucket {
	return &TokenBucket{
		Capacity:   capacity,
		Rate:       rate,
		Tokens:     capacity,
		LastUpdate: time.Now(),
	}
}

// Allow checks if the requested cost can be consumed from the bucket.
func (tb *TokenBucket) Allow(cost float64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.LastUpdate).Seconds()
	tb.LastUpdate = now

	// Replenish tokens based on elapsed time
	tb.Tokens += elapsed * tb.Rate
	if tb.Tokens > tb.Capacity {
		tb.Tokens = tb.Capacity
	}

	if tb.Tokens >= cost {
		tb.Tokens -= cost
		return true
	}
	return false
}

// PoWChallenge represents an anti-DDoS proof-of-work challenge issued to rate-exceeded peers.
type PoWChallenge struct {
	ChallengeID string    `json:"challenge_id"`
	TargetDID   string    `json:"target_did"`
	Seed        []byte    `json:"seed"`
	Difficulty  int       `json:"difficulty"` // Required number of leading zero bits (1..32)
	IssuedAt    time.Time `json:"issued_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// HierarchicalQoS coordinates bandwidth management, priority queues, and anti-DDoS defenses.
type HierarchicalQoS struct {
	mu                 sync.RWMutex
	buckets            map[string]*TokenBucket // Key: DID + ":" + Priority
	defaultRates       map[QoSPriority]float64 // Default byte/sec per priority
	defaultCapacities  map[QoSPriority]float64 // Default burst capacity per priority
	activeChallenges   map[string]PoWChallenge // Key: ChallengeID
	challengeTTL       time.Duration
	basePoWDifficulty  int
}

// NewHierarchicalQoS initializes the QoS engine with sensible production defaults.
func NewHierarchicalQoS() *HierarchicalQoS {
	return &HierarchicalQoS{
		buckets: make(map[string]*TokenBucket),
		defaultRates: map[QoSPriority]float64{
			PriorityControl:     100 * 1024,  // 100 KB/s for control plane
			PriorityInteractive: 500 * 1024,  // 500 KB/s for interactive SSH/VoIP
			PriorityBulk:        2048 * 1024, // 2 MB/s for bulk transfers
		},
		defaultCapacities: map[QoSPriority]float64{
			PriorityControl:     50 * 1024,   // 50 KB burst
			PriorityInteractive: 250 * 1024,  // 250 KB burst
			PriorityBulk:        1024 * 1024, // 1 MB burst
		},
		activeChallenges:  make(map[string]PoWChallenge),
		challengeTTL:      60 * time.Second,
		basePoWDifficulty: 10, // 10 leading zero bits (~1024 hashes, fast for test & anti-flood)
	}
}

// getBucket retrieves or lazily allocates a token bucket for a specific peer DID and priority class.
func (q *HierarchicalQoS) getBucket(did string, priority QoSPriority) *TokenBucket {
	key := fmt.Sprintf("%s:%d", did, priority)

	q.mu.RLock()
	tb, exists := q.buckets[key]
	q.mu.RUnlock()

	if exists {
		return tb
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if tb, exists = q.buckets[key]; exists {
		return tb
	}

	rate := q.defaultRates[priority]
	cap := q.defaultCapacities[priority]
	tb = NewTokenBucket(rate, cap)
	q.buckets[key] = tb
	return tb
}

// IngressCheck inspects an arriving packet.
func (q *HierarchicalQoS) IngressCheck(did string, priority QoSPriority, packetSize int) (bool, *PoWChallenge) {
	bucket := q.getBucket(did, priority)

	cost := float64(packetSize)
	if bucket.Allow(cost) {
		return true, nil
	}

	challenge := q.issueChallenge(did)
	return false, challenge
}

// issueChallenge generates a cryptographically random challenge for an offending peer.
func (q *HierarchicalQoS) issueChallenge(did string) *PoWChallenge {
	seed := make([]byte, 16)
	rand.Read(seed)

	challengeID := hex.EncodeToString(seed[:8])
	now := time.Now()

	ch := PoWChallenge{
		ChallengeID: challengeID,
		TargetDID:   did,
		Seed:        seed,
		Difficulty:  q.basePoWDifficulty,
		IssuedAt:    now,
		ExpiresAt:   now.Add(q.challengeTTL),
	}

	q.mu.Lock()
	q.activeChallenges[challengeID] = ch
	q.mu.Unlock()

	return &ch
}

// VerifyAndCredit validates a peer's PoW solution.
func (q *HierarchicalQoS) VerifyAndCredit(challengeID string, nonce uint64) bool {
	q.mu.Lock()
	ch, exists := q.activeChallenges[challengeID]
	if exists {
		delete(q.activeChallenges, challengeID)
	}
	q.mu.Unlock()

	if !exists || time.Now().After(ch.ExpiresAt) {
		return false
	}

	hasher := sha256.New()
	hasher.Write(ch.Seed)
	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, nonce)
	hasher.Write(nonceBytes)
	hash := hasher.Sum(nil)

	if !checkLeadingZeroBits(hash, ch.Difficulty) {
		return false
	}

	bucket := q.getBucket(ch.TargetDID, PriorityInteractive)
	bucket.mu.Lock()
	bucket.Tokens += 100 * 1024
	if bucket.Tokens > bucket.Capacity {
		bucket.Tokens = bucket.Capacity
	}
	bucket.mu.Unlock()

	return true
}

func checkLeadingZeroBits(hash []byte, bits int) bool {
	zeroBytes := bits / 8
	for i := 0; i < zeroBytes; i++ {
		if hash[i] != 0 {
			return false
		}
	}
	remainingBits := bits % 8
	if remainingBits > 0 {
		mask := byte(0xFF << (8 - remainingBits))
		if (hash[zeroBytes] & mask) != 0 {
			return false
		}
	}
	return true
}

func SolveChallenge(ch *PoWChallenge) uint64 {
	var nonce uint64
	nonceBytes := make([]byte, 8)

	for {
		hasher := sha256.New()
		hasher.Write(ch.Seed)
		binary.BigEndian.PutUint64(nonceBytes, nonce)
		hasher.Write(nonceBytes)
		hash := hasher.Sum(nil)

		if checkLeadingZeroBits(hash, ch.Difficulty) {
			return nonce
		}
		nonce++
	}
}
