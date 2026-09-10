package onion

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"ipv7/core"
)

// ForwardFunc abstracts blind packet forwarding over UDP/QUIC to a next hop endpoint
type ForwardFunc func(endpoint string, packet []byte) error

// OnionRouter manages circuit creation and anonymous multi-hop forwarding
type OnionRouter struct {
	node          *core.Node
	encPriv       [32]byte
	forwarder     ForwardFunc
	replayCache   map[[32]byte]time.Time
	cacheMu       sync.Mutex
	onExitPayload func(payload []byte)
}

// NewOnionRouter initializes an OnionRouter for a given node
func NewOnionRouter(node *core.Node, encPriv [32]byte, forwarder ForwardFunc) *OnionRouter {
	r := &OnionRouter{
		node:        node,
		encPriv:     encPriv,
		forwarder:   forwarder,
		replayCache: make(map[[32]byte]time.Time),
	}
	// Background cache cleanup
	go r.cleanupReplayCache()
	return r
}

// SetExitHandler sets callback for when this node is the exit/destination of an onion packet
func (r *OnionRouter) SetExitHandler(handler func(payload []byte)) {
	r.onExitPayload = handler
}

// BuildCircuit selects a random multi-hop circuit through known SmallWorld peers
func (r *OnionRouter) BuildCircuit(destDID core.Identity, destEncKey []byte, destEndpoint string, hopsCount int) ([]CircuitHop, error) {
	if hopsCount < 1 {
		hopsCount = 1
	}
	if hopsCount > DefaultMaxHops {
		hopsCount = DefaultMaxHops
	}

	var circuit []CircuitHop

	// Query closest peers from SmallWorld table
	knownPeers := r.node.SmallWorld.FindClosestPeers(destDID, 20)

	// Filter viable relay peers (must have known endpoint and encKey)
	var candidates []CircuitHop
	for _, p := range knownPeers {
		if len(p.Endpoints) > 0 {
			encKey := r.node.GetPeerEncKey(p.Identity)
			if len(encKey) == 32 {
				candidates = append(candidates, CircuitHop{
					DID:       p.Identity,
					EncPubKey: encKey,
					Endpoint:  p.Endpoints[0],
				})
			}
		}
	}

	// Select intermediate hops randomly
	intermediateCount := hopsCount - 1
	if len(candidates) > 0 && intermediateCount > 0 {
		perm := rand.Perm(len(candidates))
		for i := 0; i < intermediateCount && i < len(perm); i++ {
			circuit = append(circuit, candidates[perm[i]])
		}
	}

	// Final exit hop
	circuit = append(circuit, CircuitHop{
		DID:       destDID,
		EncPubKey: destEncKey,
		Endpoint:  destEndpoint,
	})

	return circuit, nil
}

// ProcessInboundPacket handles an incoming onion packet: peels one layer and forwards or delivers
func (r *OnionRouter) ProcessInboundPacket(packet []byte) error {
	instr, err := UnwrapLayer(packet, r.encPriv)
	if err != nil {
		return err
	}

	if instr.IsExit {
		// We are the final destination! Deliver to upper application layer
		if r.onExitPayload != nil {
			r.onExitPayload(instr.NextPayload)
		}
		return nil
	}

	// Intermediary relay hop: forward blindly to next endpoint without knowing origin or ultimate destination
	if instr.NextEndpoint == "" {
		return fmt.Errorf("empty next endpoint for relay hop")
	}

	if r.forwarder != nil {
		return r.forwarder(instr.NextEndpoint, instr.NextPayload)
	}

	return nil
}

func (r *OnionRouter) cleanupReplayCache() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		r.cacheMu.Lock()
		now := time.Now()
		for k, t := range r.replayCache {
			if now.Sub(t) > 30*time.Minute {
				delete(r.replayCache, k)
			}
		}
		r.cacheMu.Unlock()
	}
}
