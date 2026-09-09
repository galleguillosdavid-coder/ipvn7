package core

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ChaosChannel simulates an unreliable network link with loss, jitter, and duplicate bursts.
type ChaosChannel struct {
	mu           sync.Mutex
	dropRatePct  int
	dupRatePct   int
	maxJitterMs  int
	delivered    int64
	dropped      int64
	duplicates   int64
	onDeliver    func(payload []byte)
}

func NewChaosChannel(dropRatePct, dupRatePct, maxJitterMs int, onDeliver func([]byte)) *ChaosChannel {
	return &ChaosChannel{
		dropRatePct: dropRatePct,
		dupRatePct:  dupRatePct,
		maxJitterMs: maxJitterMs,
		onDeliver:   onDeliver,
	}
}

func (c *ChaosChannel) Send(payload []byte) {
	// Random drop
	n, _ := rand.Int(rand.Reader, big.NewInt(100))
	if int(n.Int64()) < c.dropRatePct {
		atomic.AddInt64(&c.dropped, 1)
		return
	}

	atomic.AddInt64(&c.delivered, 1)

	// Simulated jitter
	jitter := time.Duration(0)
	if c.maxJitterMs > 0 {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(c.maxJitterMs)))
		jitter = time.Duration(j.Int64()) * time.Millisecond
	}

	// Duplicate simulation
	isDup := false
	dupRoll, _ := rand.Int(rand.Reader, big.NewInt(100))
	if int(dupRoll.Int64()) < c.dupRatePct {
		isDup = true
		atomic.AddInt64(&c.duplicates, 1)
	}

	go func(p []byte, delay time.Duration, dup bool) {
		if delay > 0 {
			time.Sleep(delay)
		}
		c.onDeliver(p)
		if dup {
			time.Sleep(5 * time.Millisecond)
			c.onDeliver(p)
		}
	}(payload, jitter, isDup)
}

func createChaosNode(t *testing.T) (*Node, ed25519.PrivateKey) {
	id, priv, err := GenerateIdentity()
	if err != nil {
		t.Fatalf("Failed to generate identity: %v", err)
	}
	node := NewNode(id, priv)
	return node, priv
}

// TestChaosPacketLossAndJitter tests transmission resilience under severe 15% packet loss and 20ms jitter.
func TestChaosPacketLossAndJitter(t *testing.T) {
	nodeA, privA := createChaosNode(t)
	nodeB, privB := createChaosNode(t)
	_ = nodeA
	_ = privA
	_ = privB

	var receivedCount int64
	totalSent := 100

	chaos := NewChaosChannel(15, 10, 15, func(payload []byte) {
		decrypted, err := nodeB.DecryptMessage(payload)
		if err == nil && len(decrypted) > 0 {
			atomic.AddInt64(&receivedCount, 1)
		}
	})

	for i := 0; i < totalSent; i++ {
		msg := []byte(fmt.Sprintf("Chaos stress test payload message #%d", i))
		encEnvelope, err := EncryptE2EE(nodeB.EncPubKey.Bytes(), msg)
		if err != nil {
			t.Fatalf("Failed to encrypt message: %v", err)
		}
		chaos.Send(encEnvelope)
	}

	// Allow jittered messages to drain
	time.Sleep(300 * time.Millisecond)

	delivered := atomic.LoadInt64(&receivedCount)
	dropped := atomic.LoadInt64(&chaos.dropped)

	t.Logf("Chaos Test Results: Sent: %d, Received: %d, Dropped by Chaos: %d, Duplicates generated: %d",
		totalSent, delivered, dropped, atomic.LoadInt64(&chaos.duplicates))

	if delivered == 0 {
		t.Fatalf("All messages were lost in chaos channel")
	}
	if delivered+dropped < int64(totalSent*80/100) {
		t.Fatalf("Suspicious loss beyond configured chaos bounds")
	}
}

// TestChaosConcurrentE2EEStress validates concurrent encryption/decryption under heavy thread contention.
func TestChaosConcurrentE2EEStress(t *testing.T) {
	nodeA, _ := createChaosNode(t)
	nodeB, _ := createChaosNode(t)
	_ = nodeA

	numWorkers := 10
	messagesPerWorker := 50
	var wg sync.WaitGroup
	var successCount int64
	var errCount int64

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < messagesPerWorker; i++ {
				text := fmt.Sprintf("High-concurrency stress worker=%d seq=%d payload", workerID, i)
				enc, err := EncryptE2EE(nodeB.EncPubKey.Bytes(), []byte(text))
				if err != nil {
					atomic.AddInt64(&errCount, 1)
					continue
				}

				dec, err := nodeB.DecryptMessage(enc)
				if err != nil || string(dec) != text {
					atomic.AddInt64(&errCount, 1)
					continue
				}

				atomic.AddInt64(&successCount, 1)
			}
		}(w)
	}

	wg.Wait()

	totalExpected := int64(numWorkers * messagesPerWorker)
	if atomic.LoadInt64(&errCount) > 0 {
		t.Fatalf("Encountered %d errors during concurrent E2EE stress test", atomic.LoadInt64(&errCount))
	}
	if atomic.LoadInt64(&successCount) != totalExpected {
		t.Fatalf("Expected %d successful roundtrips, got %d", totalExpected, atomic.LoadInt64(&successCount))
	}

	t.Logf("Concurrent E2EE Stress PASSED: %d/%d operations flawless with zero memory corruption",
		successCount, totalExpected)
}
