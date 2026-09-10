package adapters

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// PathState describes the operational condition of an overlay link.
type PathState string

const (
	PathStateActive   PathState = "ACTIVE"
	PathStateStandby  PathState = "STANDBY"
	PathStateDegraded PathState = "DEGRADED"
	PathStateDead     PathState = "DEAD"
)

// MultipathPolicy determines how packets are dispatched across available paths.
type MultipathPolicy int

const (
	// PolicyActiveStandby: Routes all traffic through primary path; auto-failover on loss.
	PolicyActiveStandby MultipathPolicy = iota
	// PolicyLowestRTT: Dynamically selects the path with the lowest smoothed RTT.
	PolicyLowestRTT
	// PolicyRoundRobin: Balances packets sequentially across all active paths.
	PolicyRoundRobin
	// PolicyDuplication: Replicates packet on all active paths (used for critical handshakes).
	PolicyDuplication
)

// NetworkPath represents a single physical transport channel (e.g. Wi-Fi, 5G, or Ethernet).
type NetworkPath struct {
	ID              string        `json:"id"`
	InterfaceName   string        `json:"interface_name"`   // e.g. "wlan0", "eth0", "wwan0"
	RemoteEndpoint  string        `json:"remote_endpoint"`  // e.g. "198.51.100.1:7001"
	State           PathState     `json:"state"`
	Weight          int           `json:"weight"`           // Priority weight (higher is preferred)
	SmoothedRTT     time.Duration `json:"smoothed_rtt"`
	PacketsSent     uint64        `json:"packets_sent"`
	PacketsReceived uint64        `json:"packets_received"`
	PacketsLost     uint64        `json:"packets_lost"`
	LastSeen        time.Time     `json:"last_seen"`
	mu              sync.RWMutex
}

// UpdateRTT applies exponential moving average (EWMA) smoothing to RTT samples.
func (p *NetworkPath) UpdateRTT(sample time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.LastSeen = time.Now()
	if p.SmoothedRTT == 0 {
		p.SmoothedRTT = sample
	} else {
		// RFC 6298 EWMA: SRTT = (7/8)*SRTT + (1/8)*sample
		p.SmoothedRTT = (7*p.SmoothedRTT + sample) / 8
	}

	if p.State == PathStateDegraded || p.State == PathStateDead {
		p.State = PathStateActive
	}
}

// RecordDelivery records packet transmission statistics.
func (p *NetworkPath) RecordDelivery(success bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if success {
		p.PacketsReceived++
		p.LastSeen = time.Now()
	} else {
		p.PacketsLost++
	}
}

// PacketSenderFunc defines the low-level delivery callback to a physical endpoint.
type PacketSenderFunc func(path *NetworkPath, data []byte) error

// MultipathOverlay coordinates multi-homed path selection, failover and bonding.
type MultipathOverlay struct {
	mu               sync.RWMutex
	paths            map[string]*NetworkPath
	policy           MultipathPolicy
	primaryPathID    string
	heartbeatTimeout time.Duration
	roundRobinIndex  uint64
	sender           PacketSenderFunc
}

// NewMultipathOverlay creates an initialized multipath manager.
func NewMultipathOverlay(policy MultipathPolicy, sender PacketSenderFunc) *MultipathOverlay {
	return &MultipathOverlay{
		paths:            make(map[string]*NetworkPath),
		policy:           policy,
		heartbeatTimeout: 3 * time.Second,
		sender:           sender,
	}
}

// RegisterPath adds a physical path to the overlay pool.
func (m *MultipathOverlay) RegisterPath(path *NetworkPath) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if path.LastSeen.IsZero() {
		path.LastSeen = time.Now()
	}
	m.paths[path.ID] = path

	if m.primaryPathID == "" || path.Weight > 50 {
		m.primaryPathID = path.ID
	}
}

// UnregisterPath removes a path from the overlay pool.
func (m *MultipathOverlay) UnregisterPath(pathID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.paths, pathID)
	if m.primaryPathID == pathID {
		m.primaryPathID = ""
		// Elect next available active path as primary
		for _, p := range m.paths {
			if p.State == PathStateActive {
				m.primaryPathID = p.ID
				break
			}
		}
	}
}

// SetPolicy changes the current routing policy.
func (m *MultipathOverlay) SetPolicy(policy MultipathPolicy) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.policy = policy
}

// SelectPaths determines target paths according to the active policy.
func (m *MultipathOverlay) SelectPaths(forceDuplication bool) ([]*NetworkPath, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	activePaths := make([]*NetworkPath, 0, len(m.paths))
	for _, p := range m.paths {
		p.mu.RLock()
		alive := p.State == PathStateActive || p.State == PathStateStandby
		p.mu.RUnlock()
		if alive {
			activePaths = append(activePaths, p)
		}
	}

	if len(activePaths) == 0 {
		return nil, errors.New("no active multipath channels available")
	}

	if forceDuplication || m.policy == PolicyDuplication {
		return activePaths, nil
	}

	switch m.policy {
	case PolicyActiveStandby:
		primary, exists := m.paths[m.primaryPathID]
		if exists {
			primary.mu.RLock()
			isUp := primary.State == PathStateActive
			primary.mu.RUnlock()
			if isUp {
				return []*NetworkPath{primary}, nil
			}
		}
		// Fallback to first available active path
		return []*NetworkPath{activePaths[0]}, nil

	case PolicyLowestRTT:
		var best *NetworkPath
		var bestRTT time.Duration = time.Hour
		for _, p := range activePaths {
			p.mu.RLock()
			rtt := p.SmoothedRTT
			p.mu.RUnlock()
			if rtt > 0 && rtt < bestRTT {
				bestRTT = rtt
				best = p
			}
		}
		if best != nil {
			return []*NetworkPath{best}, nil
		}
		return []*NetworkPath{activePaths[0]}, nil

	case PolicyRoundRobin:
		idx := atomic.AddUint64(&m.roundRobinIndex, 1) % uint64(len(activePaths))
		return []*NetworkPath{activePaths[idx]}, nil

	default:
		return []*NetworkPath{activePaths[0]}, nil
	}
}

// SendPacket delivers data according to the multipath routing strategy.
func (m *MultipathOverlay) SendPacket(data []byte, forceDuplication bool) (int, error) {
	paths, err := m.SelectPaths(forceDuplication)
	if err != nil {
		return 0, err
	}

	if m.sender == nil {
		return 0, errors.New("no packet sender callback configured")
	}

	var delivered int
	var lastErr error

	for _, path := range paths {
		atomic.AddUint64(&path.PacketsSent, 1)
		if err := m.sender(path, data); err != nil {
			path.RecordDelivery(false)
			lastErr = err
		} else {
			path.RecordDelivery(true)
			delivered++
		}
	}

	if delivered == 0 && lastErr != nil {
		return 0, fmt.Errorf("all multipath transmissions failed: %w", lastErr)
	}

	return delivered, nil
}

// FailoverCheck evaluates path liveness and performs seamless failover if needed.
func (m *MultipathOverlay) FailoverCheck(now time.Time) (failedOver bool, newPrimaryID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	primary, exists := m.paths[m.primaryPathID]
	primaryDown := false

	if !exists {
		primaryDown = true
	} else {
		primary.mu.Lock()
		if now.Sub(primary.LastSeen) > m.heartbeatTimeout {
			primary.State = PathStateDead
			primaryDown = true
		}
		primary.mu.Unlock()
	}

	if primaryDown {
		// Find best replacement path
		var candidate *NetworkPath
		for _, p := range m.paths {
			p.mu.RLock()
			alive := p.State == PathStateActive || p.State == PathStateStandby
			p.mu.RUnlock()

			if alive {
				if candidate == nil || p.Weight > candidate.Weight {
					candidate = p
				}
			}
		}

		if candidate != nil {
			candidate.mu.Lock()
			candidate.State = PathStateActive
			candidate.mu.Unlock()

			m.primaryPathID = candidate.ID
			return true, candidate.ID
		}
	}

	return false, m.primaryPathID
}

// ListPaths returns a snapshot list of registered network paths.
func (m *MultipathOverlay) ListPaths() []*NetworkPath {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*NetworkPath, 0, len(m.paths))
	for _, p := range m.paths {
		result = append(result, p)
	}
	return result
}
