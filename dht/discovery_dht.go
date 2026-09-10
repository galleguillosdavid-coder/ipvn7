package dht

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"ipv7/core"
)

// DHTDiscoveryAdapter connects the pure Kademlia DHT service into the Node discovery ecosystem
type DHTDiscoveryAdapter struct {
	dhtService *DHTService
	node       *core.Node
	running    bool
	stopCh     chan struct{}
	mu         sync.RWMutex
}

// NewDHTDiscoveryAdapter initializes the sovereign DHT discovery adapter
func NewDHTDiscoveryAdapter(dht *DHTService, node *core.Node) *DHTDiscoveryAdapter {
	return &DHTDiscoveryAdapter{
		dhtService: dht,
		node:       node,
		stopCh:     make(chan struct{}),
	}
}

// Start begins periodic autonomous announcement of node endpoints to the DHT
func (a *DHTDiscoveryAdapter) Start(ctx context.Context, announceInterval time.Duration) error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return fmt.Errorf("dht discovery already running")
	}
	a.running = true
	a.mu.Unlock()

	// Register on-demand DID resolver on Node
	a.node.SetDIDResolver(func(didHex string) ([]string, []byte, error) {
		cleanDID := didHex
		if len(cleanDID) > 8 && cleanDID[:8] == "did:ipv7:" {
			cleanDID = cleanDID[8:]
		}
		pubBytes, err := hex.DecodeString(cleanDID)
		if err != nil || len(pubBytes) != 32 {
			return nil, nil, fmt.Errorf("invalid did hex")
		}

		rec, err := a.dhtService.Resolve(pubBytes)
		if err != nil {
			return nil, nil, err
		}

		return rec.Endpoints, nil, nil
	})

	go a.announcementLoop(ctx, announceInterval)
	return nil
}

// Stop halts the announcement loop
func (a *DHTDiscoveryAdapter) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.running {
		a.running = false
		close(a.stopCh)
	}
}

// announcementLoop periodically publishes reachable endpoints to the DHT
func (a *DHTDiscoveryAdapter) announcementLoop(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial immediate announce
	a.publishSelf()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stopCh:
			return
		case <-ticker.C:
			a.publishSelf()
		}
	}
}

func (a *DHTDiscoveryAdapter) publishSelf() {
	eps := a.node.Endpoints()
	if len(eps) > 0 {
		_, _ = a.dhtService.Publish(eps, 2*time.Hour)
	}
}
