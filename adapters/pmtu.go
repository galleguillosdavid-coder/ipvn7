package adapters

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Standard probe sizes in descending order (IPv4/IPv6 safe sizes)
var DefaultMTUCandidates = []int{
	1500, // Standard Ethernet MTU
	1472, // 1500 - 20 (IP) - 8 (UDP)
	1400, // Common overlay/VPN tunnel MTU
	1280, // Minimum IPv6 MTU (Default safe floor)
	1200, // Restricted CGNAT / DSL floor
	1000, // Ultra-restricted / Satellite link
}

const (
	MinSafeMTU     = 1000
	DefaultSafeMTU = 1280
	MaxEthernetMTU = 1500
)

// PMTUState tracks Path MTU discovery state for a remote endpoint.
type PMTUState struct {
	CurrentMTU    int
	CandidateIdx  int
	LastProbed    time.Time
	BlackHoleHits int
	Confirmed     bool
}

// PMTUDiscovery coordinates dynamic Path MTU probes and silent black-hole recovery.
type PMTUDiscovery struct {
	mu           sync.RWMutex
	peerMTU      map[string]*PMTUState
	candidates   []int
	probeTimeout time.Duration
	onMTUChange  func(endpoint string, oldMTU, newMTU int)
}

// NewPMTUDiscovery creates a dynamic PMTU discovery coordinator.
func NewPMTUDiscovery(timeout time.Duration, onChange func(endpoint string, oldMTU, newMTU int)) *PMTUDiscovery {
	if timeout <= 0 {
		timeout = 500 * time.Millisecond
	}
	return &PMTUDiscovery{
		peerMTU:      make(map[string]*PMTUState),
		candidates:   DefaultMTUCandidates,
		probeTimeout: timeout,
		onMTUChange:  onChange,
	}
}

// GetMTU returns the current established MTU for an endpoint, defaulting to 1280.
func (p *PMTUDiscovery) GetMTU(endpoint string) int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	state, ok := p.peerMTU[endpoint]
	if !ok || state.CurrentMTU <= 0 {
		return DefaultSafeMTU
	}
	return state.CurrentMTU
}

// RegisterEndpoint initializes tracking for a new destination endpoint.
func (p *PMTUDiscovery) RegisterEndpoint(endpoint string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.peerMTU[endpoint]; !ok {
		p.peerMTU[endpoint] = &PMTUState{
			CurrentMTU:    DefaultSafeMTU,
			CandidateIdx:  3, // Index 3 is 1280
			LastProbed:    time.Now(),
			BlackHoleHits: 0,
			Confirmed:     false,
		}
	}
}

// NextProbeSize calculates the next candidate size to test.
// If seeking upward, it tries higher values (1400, 1472, 1500).
func (p *PMTUDiscovery) NextProbeSize(endpoint string) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	state, ok := p.peerMTU[endpoint]
	if !ok {
		p.peerMTU[endpoint] = &PMTUState{
			CurrentMTU:   DefaultSafeMTU,
			CandidateIdx: 3,
			LastProbed:   time.Now(),
		}
		state = p.peerMTU[endpoint]
	}

	state.LastProbed = time.Now()

	// If not confirmed, try highest candidate downward
	if state.CandidateIdx > 0 && state.BlackHoleHits == 0 {
		return p.candidates[state.CandidateIdx-1], nil
	}

	return state.CurrentMTU, nil
}

// ConfirmProbeSize records that a packet of probeSize was successfully acknowledged.
func (p *PMTUDiscovery) ConfirmProbeSize(endpoint string, probeSize int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	state, ok := p.peerMTU[endpoint]
	if !ok {
		state = &PMTUState{CurrentMTU: probeSize, Confirmed: true}
		p.peerMTU[endpoint] = state
		return
	}

	oldMTU := state.CurrentMTU
	if probeSize > state.CurrentMTU {
		state.CurrentMTU = probeSize
		state.Confirmed = true
		state.BlackHoleHits = 0

		// Update candidate index
		for i, cand := range p.candidates {
			if cand == probeSize {
				state.CandidateIdx = i
				break
			}
		}

		if p.onMTUChange != nil {
			go p.onMTUChange(endpoint, oldMTU, probeSize)
		}
	}
}

// RecordBlackHoleTimeout signals that packets sent at current/probed size were dropped
// without ICMP unreachable messages (silent black-hole detection).
func (p *PMTUDiscovery) RecordBlackHoleTimeout(endpoint string, failedSize int) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	state, ok := p.peerMTU[endpoint]
	if !ok {
		state = &PMTUState{CurrentMTU: DefaultSafeMTU}
		p.peerMTU[endpoint] = state
	}

	state.BlackHoleHits++
	oldMTU := state.CurrentMTU

	// Step down to next smaller candidate
	newMTU := DefaultSafeMTU
	for _, cand := range p.candidates {
		if cand < failedSize {
			newMTU = cand
			break
		}
	}

	if newMTU < MinSafeMTU {
		newMTU = MinSafeMTU
	}

	state.CurrentMTU = newMTU
	state.Confirmed = false

	if p.onMTUChange != nil && oldMTU != newMTU {
		go p.onMTUChange(endpoint, oldMTU, newMTU)
	}

	return newMTU
}

// ProbeSender defines the function signature to emit a probe of exact byte length.
type ProbeSender func(ctx context.Context, endpoint string, size int) error

// DiscoverPathMTU runs an active binary probe sequence to discover the exact path MTU.
func (p *PMTUDiscovery) DiscoverPathMTU(ctx context.Context, endpoint string, sender ProbeSender) (int, error) {
	if sender == nil {
		return DefaultSafeMTU, errors.New("nil probe sender provided")
	}

	p.RegisterEndpoint(endpoint)

	// Probe downwards from highest (1500 -> 1472 -> 1400 -> 1280)
	for _, candidate := range p.candidates {
		select {
		case <-ctx.Done():
			return p.GetMTU(endpoint), ctx.Err()
		default:
		}

		probeCtx, cancel := context.WithTimeout(ctx, p.probeTimeout)
		err := sender(probeCtx, endpoint, candidate)
		cancel()

		if err == nil {
			// Succeeded at this size
			p.ConfirmProbeSize(endpoint, candidate)
			return candidate, nil
		}

		// Black-hole hit at this candidate
		p.RecordBlackHoleTimeout(endpoint, candidate)
	}

	return MinSafeMTU, nil
}
