package offgrid

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"ipv7/core"
)

var (
	ErrNoRouteAvailable = errors.New("no route available in current link mode")
	ErrPeerUnreachable  = errors.New("peer unreachable via physical mesh or internet")
)

// SwitcherConfig configures the HybridSwitcher behavior
type SwitcherConfig struct {
	AutoFallback     bool          // Automatically switch to off-grid on WAN loss
	HeartbeatTimeout time.Duration // Time without WAN confirmation before declaring off-grid
	DefaultMode      LinkMode
}

// SwitcherStats records runtime traffic and mode transitions
type SwitcherStats struct {
	ModeTransitions uint64
	WANPacketsSent  uint64
	MeshPacketsSent uint64
	PacketsDropped  uint64
}

// HybridSwitcher coordinates traffic between global WAN and local ad-hoc physical radio
type HybridSwitcher struct {
	node         *core.Node
	beaconEngine *BeaconEngine
	physicalLink PhysicalLink
	config       SwitcherConfig

	currentMode LinkMode
	wanOnline   bool
	lastWANPing time.Time

	mu          sync.RWMutex
	onModeChange func(oldMode, newMode LinkMode)

	stats SwitcherStats
	cancel context.CancelFunc
}

// NewHybridSwitcher creates an online/offgrid hybrid orchestrator
func NewHybridSwitcher(
	node *core.Node,
	beaconEngine *BeaconEngine,
	physicalLink PhysicalLink,
	cfg SwitcherConfig,
) *HybridSwitcher {
	if cfg.HeartbeatTimeout <= 0 {
		cfg.HeartbeatTimeout = 5 * time.Second
	}
	if cfg.DefaultMode == 0 {
		cfg.DefaultMode = ModeOnlineInternet
	}

	return &HybridSwitcher{
		node:         node,
		beaconEngine: beaconEngine,
		physicalLink: physicalLink,
		config:       cfg,
		currentMode:  cfg.DefaultMode,
		wanOnline:    cfg.DefaultMode == ModeOnlineInternet || cfg.DefaultMode == ModeHybrid,
		lastWANPing:  time.Now(),
	}
}

// SetOnModeChange registers a listener for topology/mode shifts
func (s *HybridSwitcher) SetOnModeChange(cb func(oldMode, newMode LinkMode)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onModeChange = cb
}

// CurrentMode returns active operational mode
func (s *HybridSwitcher) CurrentMode() LinkMode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentMode
}

// IsWANOnline returns whether the node perceives global internet reachability
func (s *HybridSwitcher) IsWANOnline() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.wanOnline
}

// ReportWANStatus manually or periodically updates WAN reachability
func (s *HybridSwitcher) ReportWANStatus(online bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.wanOnline = online
	if online {
		s.lastWANPing = time.Now()
	}

	if s.config.AutoFallback {
		var targetMode LinkMode
		if online {
			if s.physicalLink != nil {
				targetMode = ModeHybrid
			} else {
				targetMode = ModeOnlineInternet
			}
		} else {
			targetMode = ModeOffGridPhysical
		}

		if targetMode != s.currentMode {
			old := s.currentMode
			s.currentMode = targetMode
			s.stats.ModeTransitions++
			if s.onModeChange != nil {
				go s.onModeChange(old, targetMode)
			}
		}
	}
}

// SetMode explicitly forces a specific operational mode
func (s *HybridSwitcher) SetMode(mode LinkMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.currentMode != mode {
		old := s.currentMode
		s.currentMode = mode
		s.stats.ModeTransitions++
		if s.onModeChange != nil {
			go s.onModeChange(old, mode)
		}
	}
}

// StartWatchdog runs periodic background checks on link health
func (s *HybridSwitcher) StartWatchdog(ctx context.Context, checkInterval time.Duration) {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	if checkInterval <= 0 {
		checkInterval = 1 * time.Second
	}

	go func() {
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.mu.Lock()
				if s.wanOnline && time.Since(s.lastWANPing) > s.config.HeartbeatTimeout {
					s.wanOnline = false
					if s.config.AutoFallback && s.currentMode != ModeOffGridPhysical {
						old := s.currentMode
						s.currentMode = ModeOffGridPhysical
						s.stats.ModeTransitions++
						if s.onModeChange != nil {
							cb := s.onModeChange
							go cb(old, ModeOffGridPhysical)
						}
					}
				}
				s.mu.Unlock()
			}
		}
	}()
}

// Stop stops the watchdog
func (s *HybridSwitcher) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

// RoutePacket intelligently sends data through physical radio or Internet WAN depending on mode and peer proximity
func (s *HybridSwitcher) RoutePacket(destDID core.Identity, packet []byte) error {
	s.mu.RLock()
	mode := s.currentMode
	s.mu.RUnlock()

	// 1. Check if peer is locally reachable on ad-hoc physical radio
	if s.beaconEngine != nil {
		peers := s.beaconEngine.DiscoveredPeers()
		for _, p := range peers {
			if p.DID.String() == destDID.String() {
				// Transmit directly via PhysicalLink
				if s.physicalLink != nil {
					err := s.physicalLink.Send(p.LocalAddr, packet)
					if err == nil {
						atomic.AddUint64(&s.stats.MeshPacketsSent, 1)
						return nil
					}
				}
			}
		}
	}

	// 2. If peer is not within direct physical radio range:
	switch mode {
	case ModeOffGridPhysical:
		// Cannot route via WAN because we are in strict off-grid isolation
		atomic.AddUint64(&s.stats.PacketsDropped, 1)
		return ErrPeerUnreachable

	case ModeOnlineInternet, ModeHybrid:
		// Peer must be contacted through IPv7 Node overlay network (WAN)
		if s.node != nil {
			err := s.node.SendMessage(destDID, packet)
			if err != nil {
				atomic.AddUint64(&s.stats.PacketsDropped, 1)
				return err
			}
			atomic.AddUint64(&s.stats.WANPacketsSent, 1)
			return nil
		}
		atomic.AddUint64(&s.stats.PacketsDropped, 1)
		return ErrNoRouteAvailable

	default:
		atomic.AddUint64(&s.stats.PacketsDropped, 1)
		return ErrNoRouteAvailable
	}
}

// Stats returns a snapshot of switcher operations
func (s *HybridSwitcher) Stats() SwitcherStats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return SwitcherStats{
		ModeTransitions: s.stats.ModeTransitions,
		WANPacketsSent:  atomic.LoadUint64(&s.stats.WANPacketsSent),
		MeshPacketsSent: atomic.LoadUint64(&s.stats.MeshPacketsSent),
		PacketsDropped:  atomic.LoadUint64(&s.stats.PacketsDropped),
	}
}
