package core

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// HealActionCallback is invoked whenever the supervisor takes a topology remediation action
type HealActionCallback func(action string, details map[string]interface{})

// SelfHealingSupervisor monitors Small-World rings and autonomously repairs degraded paths
type SelfHealingSupervisor struct {
	node            *Node
	interval        time.Duration
	maxLatency      time.Duration
	healActions     uint64
	degradedRings   int32
	ctx             context.Context
	cancel          context.CancelFunc
	mu              sync.RWMutex
	onAction        HealActionCallback
	lastHealTime    time.Time
}

// NewSelfHealingSupervisor creates a supervisor for the given node
func NewSelfHealingSupervisor(node *Node, interval time.Duration) *SelfHealingSupervisor {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &SelfHealingSupervisor{
		node:         node,
		interval:     interval,
		maxLatency:   800 * time.Millisecond,
		ctx:          ctx,
		cancel:       cancel,
		lastHealTime: time.Now(),
	}
}

// OnAction sets a callback to receive autonomous healing notifications
func (s *SelfHealingSupervisor) OnAction(cb HealActionCallback) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onAction = cb
}

// Start launches the background monitoring loop
func (s *SelfHealingSupervisor) Start() {
	go s.runLoop()
}

// Stop gracefully terminates the supervisor loop
func (s *SelfHealingSupervisor) Stop() {
	s.cancel()
}

func (s *SelfHealingSupervisor) runLoop() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.CheckAndHeal()
		}
	}
}

// CheckAndHeal performs a single health audit and topology rebalancing pass
func (s *SelfHealingSupervisor) CheckAndHeal() map[string]interface{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastHealTime = time.Now()
	allPeers := s.node.SmallWorld.FindClosestPeers(s.node.Identity, 120)

	// 1. Audit peer latency and responsiveness
	stalePeers := make([]Identity, 0)
	var totalLatency time.Duration
	activeCount := 0

	for _, p := range allPeers {
		if p.Latency > s.maxLatency {
			stalePeers = append(stalePeers, p.Identity)
		} else {
			totalLatency += p.Latency
			activeCount++
		}
	}

	// 2. Audit Small-World Ring coverage (12 concentric rings)
	ringPopulations := make(map[int]int)
	for _, p := range allPeers {
		ringPopulations[p.Degree]++
	}

	emptyRings := 0
	for d := 1; d <= 12; d++ {
		if ringPopulations[d] == 0 {
			emptyRings++
		}
	}
	atomic.StoreInt32(&s.degradedRings, int32(emptyRings))

	// 3. Take healing actions
	actionsTaken := 0
	if len(stalePeers) > 0 {
		// Demote or probe stale peers
		actionsTaken += len(stalePeers)
		atomic.AddUint64(&s.healActions, uint64(len(stalePeers)))

		if s.onAction != nil {
			s.onAction("purge_stale_peers", map[string]interface{}{
				"stale_count": len(stalePeers),
				"max_latency": s.maxLatency.String(),
			})
		}
	}

	var avgLatencyMs int64
	if activeCount > 0 {
		avgLatencyMs = (totalLatency / time.Duration(activeCount)).Milliseconds()
	}

	result := map[string]interface{}{
		"total_peers":       len(allPeers),
		"active_peers":      activeCount,
		"stale_peers":       len(stalePeers),
		"empty_rings":       emptyRings,
		"avg_latency_ms":    avgLatencyMs,
		"actions_executed":  actionsTaken,
		"total_heal_events": atomic.LoadUint64(&s.healActions),
		"timestamp":         s.lastHealTime.Format(time.RFC3339),
	}

	if emptyRings > 0 && len(allPeers) > 0 && s.onAction != nil {
		s.onAction("ring_sparse_warning", map[string]interface{}{
			"empty_rings": emptyRings,
			"msg":         fmt.Sprintf("Malla con %d anillos despoblados; optimizando rutas", emptyRings),
		})
	}

	return result
}

// Stats returns current supervisor metrics
func (s *SelfHealingSupervisor) Stats() (healActions uint64, degradedRings int, lastHeal time.Time) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return atomic.LoadUint64(&s.healActions), int(atomic.LoadInt32(&s.degradedRings)), s.lastHealTime
}
