package core

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	DefaultFirebaseInterval = 30 * time.Second
	FirebaseStaleTimeout    = 300 // 5 minutes in seconds
)

// FirebasePeerRecord represents an active peer advertised on Firebase Realtime DB
type FirebasePeerRecord struct {
	ID        string   `json:"id"`
	Endpoints []string `json:"endpoints"`
	EncKey    string   `json:"enc_key,omitempty"`
	Port      int      `json:"port"`
	LastSeen  int64    `json:"last_seen"`
	Version   string   `json:"version"`
}

// FirebaseDiscovery manages zero-configuration WAN/Internet peer rendezvous via Firebase
type FirebaseDiscovery struct {
	baseURL    string
	node       *Node
	port       int
	httpClient *http.Client
	interval   time.Duration
	ctx        context.Context
	cancel     context.CancelFunc
	discovered map[string]time.Time
	refresher  func() []string
	mu         sync.Mutex
	running    bool
}

// NewFirebaseDiscovery creates a new discovery service using the provided Firebase RTDB URL
func NewFirebaseDiscovery(baseURL string, node *Node, port int) *FirebaseDiscovery {
	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &FirebaseDiscovery{
		baseURL: baseURL,
		node:    node,
		port:    port,
		httpClient: &http.Client{
			Timeout: 8 * time.Second,
		},
		interval:   DefaultFirebaseInterval,
		ctx:        ctx,
		cancel:     cancel,
		discovered: make(map[string]time.Time),
	}
}

// Start launches periodic registration and discovery loops
func (f *FirebaseDiscovery) Start() {
	f.mu.Lock()
	if f.running {
		f.mu.Unlock()
		return
	}
	f.running = true
	f.mu.Unlock()

	// Initial announce and discovery synchronously or quick goroutine
	go func() {
		f.AnnounceAndDiscover()

		ticker := time.NewTicker(f.interval)
		defer ticker.Stop()

		for {
			select {
			case <-f.ctx.Done():
				return
			case <-ticker.C:
				f.AnnounceAndDiscover()
			}
		}
	}()
}

// Stop unregisters this node from Firebase and terminates background polling
func (f *FirebaseDiscovery) Stop() {
	f.mu.Lock()
	if !f.running {
		f.mu.Unlock()
		return
	}
	f.running = false
	f.mu.Unlock()

	f.cancel()
	f.Deregister()
}

// AnnounceAndDiscover registers local endpoints and fetches remote peers
func (f *FirebaseDiscovery) AnnounceAndDiscover() {
	// 1. Announce self
	if err := f.Announce(); err != nil {
		fmt.Printf("[Firebase Rendezvous] Error anunciando nodo: %v\n", err)
	}

	// 2. Discover remote peers
	peers, err := f.FetchPeers()
	if err != nil {
		fmt.Printf("[Firebase Rendezvous] Error consultando peers: %v\n", err)
		return
	}

	now := time.Now().Unix()
	localID := f.node.Identity.String()

	for _, p := range peers {
		// Ignore self
		if p.ID == localID || p.ID == "" {
			continue
		}

		// Ignore stale peers (older than 5 minutes)
		if now-p.LastSeen > FirebaseStaleTimeout {
			continue
		}

		// Try candidates
		for _, ep := range p.Endpoints {
			f.mu.Lock()
			lastSeen, exists := f.discovered[ep]
			if exists && time.Since(lastSeen) < 1*time.Minute {
				f.mu.Unlock()
				continue
			}
			f.discovered[ep] = time.Now()
			f.mu.Unlock()

			// Initiate automatic handshake
			go func(endpoint string, targetID string) {
				fmt.Printf("[Firebase Rendezvous] Peer remoto detectado: %s (%s). Conectando...\n", endpoint, targetID[:16]+"...")
				peerID, rtt, err := f.node.Handshake(endpoint)
				if err != nil {
					// Pending or firewalled NAT probe
					return
				}
				fmt.Printf("[OK]   ¡Conexión P2P establecida vía Firebase con %s! RTT: %v\n", peerID.String()[:16]+"...", rtt)
			}(ep, p.ID)
		}
	}
}

// SetEndpointRefresher registers a callback that re-probes local interfaces and STUN to handle network changes
func (f *FirebaseDiscovery) SetEndpointRefresher(fn func() []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.refresher = fn
}

// Announce registers or updates the local node in Firebase Realtime Database
func (f *FirebaseDiscovery) Announce() error {
	f.mu.Lock()
	refresher := f.refresher
	f.mu.Unlock()

	if refresher != nil {
		if fresh := refresher(); len(fresh) > 0 {
			f.node.SetEndpoints(fresh)
		}
	}

	encKeyHex := ""
	if f.node.EncPubKey != nil {
		encKeyHex = hex.EncodeToString(f.node.EncPubKey.Bytes())
	}

	endpoints := f.node.Endpoints()
	if len(endpoints) == 0 {
		return nil // No endpoints known yet
	}

	record := FirebasePeerRecord{
		ID:        f.node.Identity.String(),
		Endpoints: endpoints,
		EncKey:    encKeyHex,
		Port:      f.port,
		LastSeen:  time.Now().Unix(),
		Version:   "1.0.0",
	}

	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/peers/%s.json", f.baseURL, f.node.Identity.String())
	req, err := http.NewRequestWithContext(f.ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("http status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// FetchPeers retrieves all registered peers from Firebase Realtime DB
func (f *FirebaseDiscovery) FetchPeers() ([]FirebasePeerRecord, error) {
	url := fmt.Sprintf("%s/peers.json", f.baseURL)
	req, err := http.NewRequestWithContext(f.ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("http status %d: %s", resp.StatusCode, string(body))
	}

	var peerMap map[string]FirebasePeerRecord
	if err := json.NewDecoder(resp.Body).Decode(&peerMap); err != nil {
		// Could be null if database is empty
		return nil, nil
	}

	list := make([]FirebasePeerRecord, 0, len(peerMap))
	for _, rec := range peerMap {
		list = append(list, rec)
	}

	return list, nil
}

// Deregister deletes the local node entry from Firebase Realtime DB
func (f *FirebaseDiscovery) Deregister() {
	url := fmt.Sprintf("%s/peers/%s.json", f.baseURL, f.node.Identity.String())
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return
	}

	resp, err := f.httpClient.Do(req)
	if err == nil && resp != nil {
		_ = resp.Body.Close()
	}
}

// ResolveDID queries Firebase Realtime DB directly for a specific peer identity or DID
func (f *FirebaseDiscovery) ResolveDID(targetID string) (*FirebasePeerRecord, error) {
	targetID = strings.TrimPrefix(targetID, "did:ipv7:")
	targetID = strings.TrimSpace(targetID)

	url := fmt.Sprintf("%s/peers/%s.json", f.baseURL, targetID)
	req, err := http.NewRequestWithContext(f.ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http error %d", resp.StatusCode)
	}

	var rec FirebasePeerRecord
	if err := json.NewDecoder(resp.Body).Decode(&rec); err != nil {
		return nil, err
	}
	if rec.ID == "" {
		return nil, fmt.Errorf("peer %s record is empty or offline", targetID)
	}
	return &rec, nil
}
