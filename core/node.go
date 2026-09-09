package core

import (
	"bytes"
	"crypto/ecdh"
	"crypto/ed25519"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// MessageHandler is a callback invoked when a valid verified message arrives for this node
type MessageHandler func(from Identity, payload []byte)

// Node represents an active IPv7 network entity
type Node struct {
	Identity   *Ed25519Identity
	privateKey ed25519.PrivateKey

	// X25519 keys for End-to-End Encryption (E2EE)
	EncPrivKey *ecdh.PrivateKey
	EncPubKey  *ecdh.PublicKey

	adapters []Adapter

	// SmallWorld routing table (12 degrees, 10 peers per bucket)
	SmallWorld *SmallWorldTable

	// peerTable maps Identity string (hex pubkey) -> list of known physical endpoints (ip:port)
	peerTable      map[string][]string
	peerEncKeys    map[string][]byte
	localEndpoints []string
	mu             sync.RWMutex

	// Atomic throughput & packet metrics
	totalSent     atomic.Uint64
	totalReceived atomic.Uint64
	bytesSent     atomic.Uint64
	bytesReceived atomic.Uint64

	// handlers holds multiple concurrent message listeners
	handlers []MessageHandler
	running  bool
	stopCh   chan struct{}

	// Pending handshakes and pings mapped by nonce
	pendingHandshakes map[uint64]chan *HandshakePayload
	pendingPings      map[uint64]chan time.Time
	handshakeMu       sync.Mutex

	// Dynamic DID resolver for on-demand identity-to-route lookup
	didResolver func(did string) ([]string, []byte, error)
}

// NewNode initializes a new IPv7 Node with its cryptographic identity and derived E2EE key
func NewNode(identity *Ed25519Identity, privateKey ed25519.PrivateKey) *Node {
	encPriv, encPub, _ := DeriveX25519FromSeed(privateKey.Seed())

	return &Node{
		Identity:          identity,
		privateKey:        privateKey,
		EncPrivKey:        encPriv,
		EncPubKey:         encPub,
		SmallWorld:        NewSmallWorldTable(identity, DefaultMaxDegrees, DefaultPeersPerDegree),
		peerTable:         make(map[string][]string),
		peerEncKeys:       make(map[string][]byte),
		stopCh:            make(chan struct{}),
		handlers:          make([]MessageHandler, 0),
		pendingHandshakes: make(map[uint64]chan *HandshakePayload),
		pendingPings:      make(map[uint64]chan time.Time),
	}
}

// AddAdapter registers a transport adapter (UDP, QUIC, Relay) to this node
func (n *Node) AddAdapter(adapter Adapter) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.adapters = append(n.adapters, adapter)
}

// OnMessage registers a listener callback for incoming payload without overwriting previous listeners
func (n *Node) OnMessage(handler MessageHandler) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.handlers = append(n.handlers, handler)
}

// AddPeer registers known endpoints for a peer's identity and records it in the SmallWorld table
func (n *Node) AddPeer(peerID Identity, endpoints []string) {
	n.AddPeerWithLatency(peerID, endpoints, 0)
}

// AddPeerWithLatency registers a peer with its measured RTT latency
func (n *Node) AddPeerWithLatency(peerID Identity, endpoints []string, latency time.Duration) {
	n.mu.Lock()
	n.peerTable[peerID.String()] = endpoints
	n.mu.Unlock()

	n.SmallWorld.AddPeer(peerID, endpoints, latency)
}

// GetPeerEndpoints returns known endpoints for a peer
func (n *Node) GetPeerEndpoints(peerID Identity) []string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	eps, exists := n.peerTable[peerID.String()]
	if !exists {
		return nil
	}
	result := make([]string, len(eps))
	copy(result, eps)
	return result
}

// SetPeerEncKey caches the 32-byte X25519 encryption public key for a peer identity
func (n *Node) SetPeerEncKey(peerID Identity, pubKey []byte) {
	if len(pubKey) != 32 {
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	n.peerEncKeys[peerID.String()] = append([]byte(nil), pubKey...)
}

// GetPeerEncKey returns the 32-byte X25519 encryption public key for a peer, or nil if unknown
func (n *Node) GetPeerEncKey(peerID Identity) []byte {
	n.mu.RLock()
	defer n.mu.RUnlock()
	key, ok := n.peerEncKeys[peerID.String()]
	if !ok {
		return nil
	}
	return append([]byte(nil), key...)
}

// SetEndpoints sets the node's own discovered reachable endpoints
func (n *Node) SetEndpoints(eps []string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.localEndpoints = make([]string, len(eps))
	copy(n.localEndpoints, eps)
}

// Endpoints returns a copy of the node's reachable endpoints
func (n *Node) Endpoints() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	res := make([]string, len(n.localEndpoints))
	copy(res, n.localEndpoints)
	return res
}

// Stats returns atomic metrics for packets and bytes transferred
func (n *Node) Stats() (sent, recv, bSent, bRecv uint64) {
	return n.totalSent.Load(), n.totalReceived.Load(), n.bytesSent.Load(), n.bytesReceived.Load()
}

// Start boots up all registered adapters and begins listening for containers
func (n *Node) Start() error {
	n.mu.Lock()
	if n.running {
		n.mu.Unlock()
		return errors.New("node already running")
	}
	n.running = true
	n.stopCh = make(chan struct{})
	adapters := make([]Adapter, len(n.adapters))
	copy(adapters, n.adapters)
	n.mu.Unlock()

	for _, a := range adapters {
		if err := a.Start(); err != nil {
			return fmt.Errorf("failed to start adapter: %w", err)
		}
		go n.adapterListenLoop(a)
	}

	return nil
}

// Stop shuts down the node and its adapters
func (n *Node) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.running {
		return nil
	}
	n.running = false
	close(n.stopCh)

	for _, a := range n.adapters {
		_ = a.Stop()
	}
	return nil
}

// SetDIDResolver assigns a dynamic resolver callback to resolve peer DIDs to live endpoints on demand
func (n *Node) SetDIDResolver(resolver func(did string) ([]string, []byte, error)) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.didResolver = resolver
}

// SendMessage sends an authenticated data payload to a target peer identity
func (n *Node) SendMessage(target Identity, payload []byte) error {
	endpoints := n.GetPeerEndpoints(target)

	// If destination not directly known, attempt on-demand DID resolution
	if len(endpoints) == 0 {
		n.mu.RLock()
		resolver := n.didResolver
		n.mu.RUnlock()
		if resolver != nil {
			if eps, encKey, err := resolver(target.String()); err == nil && len(eps) > 0 {
				n.AddPeer(target, eps)
				if len(encKey) == 32 {
					n.SetPeerEncKey(target, encKey)
				}
				endpoints = eps
			}
		}
	}

	// Small-World greedy resolution: if destination not directly connected, find closest next-hop peer
	if len(endpoints) == 0 {
		closest := n.SmallWorld.FindClosestPeers(target, 1)
		if len(closest) > 0 {
			endpoints = closest[0].Endpoints
		}
	}

	if len(endpoints) == 0 {
		return fmt.Errorf("no known route or endpoints for peer %s", target.String())
	}

	// Build IPv7 container with DefaultHopLimit (12 degrees of separation)
	c := &Container{
		SenderPubKey:   n.Identity.Bytes(),
		ReceiverPubKey: target.Bytes(),
		Payload:        payload,
		HopLimit:       DefaultHopLimit,
	}

	// Sign container with node's private key
	if err := c.Sign(n.privateKey); err != nil {
		return fmt.Errorf("failed to sign container: %w", err)
	}

	return n.forwardContainer(c, endpoints)
}

// SendEncryptedMessage encrypts the payload using recipient's X25519 public key before routing
func (n *Node) SendEncryptedMessage(target Identity, recipientX25519Pub []byte, plaintext []byte) error {
	ciphertext, err := EncryptE2EE(recipientX25519Pub, plaintext)
	if err != nil {
		return fmt.Errorf("e2ee encryption failed: %w", err)
	}
	return n.SendMessage(target, ciphertext)
}

// DecryptMessage decrypts an incoming E2EE envelope using node's private X25519 key
func (n *Node) DecryptMessage(envelope []byte) ([]byte, error) {
	if n.EncPrivKey == nil {
		return nil, errors.New("no encryption private key available")
	}
	return DecryptE2EE(n.EncPrivKey, envelope)
}

// Handshake initiates a two-way cryptographic identity exchange with a physical endpoint (ip:port).
// It discovers the peer's authentic Ed25519 public key, X25519 encryption key, and measures real RTT.
func (n *Node) Handshake(endpoint string) (*Ed25519Identity, time.Duration, error) {
	nonce := GenerateNonce()
	respCh := make(chan *HandshakePayload, 1)

	n.handshakeMu.Lock()
	n.pendingHandshakes[nonce] = respCh
	n.handshakeMu.Unlock()

	defer func() {
		n.handshakeMu.Lock()
		delete(n.pendingHandshakes, nonce)
		n.handshakeMu.Unlock()
	}()

	var x25519Bytes []byte
	if n.EncPubKey != nil {
		x25519Bytes = n.EncPubKey.Bytes()
	}

	start := time.Now()
	reqPayload := &HandshakePayload{
		Type:       ControlHandshakeReq,
		Ed25519Pub: n.Identity.Bytes(),
		X25519Pub:  x25519Bytes,
		Endpoints:  n.Endpoints(),
		Timestamp:  start.UnixNano(),
		Nonce:      nonce,
	}

	data, err := EncodeHandshake(reqPayload)
	if err != nil {
		return nil, 0, err
	}

	c := &Container{
		SenderPubKey: n.Identity.Bytes(),
		SessionID:    "handshake",
		Payload:      data,
		HopLimit:     DefaultHopLimit,
	}

	if err := c.Sign(n.privateKey); err != nil {
		return nil, 0, fmt.Errorf("failed to sign handshake container: %w", err)
	}

	if err := n.forwardContainer(c, []string{endpoint}); err != nil {
		return nil, 0, fmt.Errorf("failed to transmit handshake to %s: %w", endpoint, err)
	}

	// Await authenticated response
	select {
	case resp := <-respCh:
		rtt := time.Since(start)
		if rtt <= 0 {
			rtt = time.Millisecond
		}

		peerID, err := NewIdentityFromBytes(resp.Ed25519Pub)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid peer identity in handshake response: %w", err)
		}

		// Register peer with real identity, endpoints and measured RTT
		eps := []string{endpoint}
		for _, ep := range resp.Endpoints {
			if ep != endpoint {
				eps = append(eps, ep)
			}
		}
		n.AddPeerWithLatency(peerID, eps, rtt)
		if len(resp.X25519Pub) == 32 {
			n.SetPeerEncKey(peerID, resp.X25519Pub)
		}
		return peerID, rtt, nil

	case <-time.After(3 * time.Second):
		return nil, 0, fmt.Errorf("handshake timeout with %s", endpoint)
	}
}

// PingPeer actively probes a known peer and returns the measured RTT latency
func (n *Node) PingPeer(targetID Identity) (time.Duration, error) {
	endpoints := n.GetPeerEndpoints(targetID)
	if len(endpoints) == 0 {
		return 0, fmt.Errorf("no known endpoint for %s", targetID.String())
	}

	nonce := GenerateNonce()
	respCh := make(chan time.Time, 1)

	n.handshakeMu.Lock()
	n.pendingPings[nonce] = respCh
	n.handshakeMu.Unlock()

	defer func() {
		n.handshakeMu.Lock()
		delete(n.pendingPings, nonce)
		n.handshakeMu.Unlock()
	}()

	start := time.Now()
	pingPayload := &PingPayload{
		Type:      ControlPing,
		Nonce:     nonce,
		Timestamp: start.UnixNano(),
	}

	data, err := EncodePing(pingPayload)
	if err != nil {
		return 0, err
	}

	c := &Container{
		SenderPubKey:   n.Identity.Bytes(),
		ReceiverPubKey: targetID.Bytes(),
		SessionID:      "ping",
		Payload:        data,
		HopLimit:       DefaultHopLimit,
	}

	if err := c.Sign(n.privateKey); err != nil {
		return 0, err
	}

	if err := n.forwardContainer(c, endpoints); err != nil {
		return 0, err
	}

	select {
	case respTime := <-respCh:
		rtt := respTime.Sub(start)
		if rtt <= 0 {
			rtt = time.Millisecond
		}
		// Update peer latency in SmallWorld table
		n.AddPeerWithLatency(targetID, endpoints, rtt)
		return rtt, nil

	case <-time.After(3 * time.Second):
		return 0, fmt.Errorf("ping timeout for %s", targetID.String())
	}
}

func (n *Node) forwardContainer(c *Container, endpoints []string) error {
	n.mu.RLock()
	adapters := make([]Adapter, len(n.adapters))
	copy(adapters, n.adapters)
	n.mu.RUnlock()

	if len(adapters) == 0 {
		return errors.New("no adapters configured on node")
	}

	var lastErr error
	for _, a := range adapters {
		err := a.Send(c, endpoints)
		if err == nil {
			n.totalSent.Add(1)
			n.bytesSent.Add(uint64(len(c.Payload)))
			return nil
		}
		lastErr = err
	}
	return lastErr
}

func (n *Node) adapterListenLoop(a Adapter) {
	for {
		select {
		case <-n.stopCh:
			return
		case c, ok := <-a.Receive():
			if !ok {
				return
			}
			if c == nil {
				continue
			}

			// Verify end-to-end Ed25519 signature
			if !c.Verify() {
				continue
			}

			// Default or validate HopLimit / TTL: must satisfy <= DefaultHopLimit
			if c.HopLimit == 0 {
				c.HopLimit = DefaultHopLimit
			} else if c.HopLimit > DefaultHopLimit {
				continue
			}

			n.totalReceived.Add(1)
			n.bytesReceived.Add(uint64(len(c.Payload)))

			senderID, err := NewIdentityFromBytes(c.SenderPubKey)
			if err != nil {
				continue
			}

			// Check if packet is a Control Handshake
			if hp, err := DecodeHandshake(c.Payload); err == nil {
				if hp.Type == ControlHandshakeReq {
					// Register incoming peer's announced endpoints immediately
					if len(hp.Endpoints) > 0 {
						n.AddPeerWithLatency(senderID, hp.Endpoints, time.Millisecond)
					}
					if len(hp.X25519Pub) == 32 {
						n.SetPeerEncKey(senderID, hp.X25519Pub)
					}

					// 1. Reply with Handshake Response
					var x25519Bytes []byte
					if n.EncPubKey != nil {
						x25519Bytes = n.EncPubKey.Bytes()
					}
					respPayload := &HandshakePayload{
						Type:       ControlHandshakeResp,
						Ed25519Pub: n.Identity.Bytes(),
						X25519Pub:  x25519Bytes,
						Endpoints:  n.Endpoints(),
						Timestamp:  time.Now().UnixNano(),
						Nonce:      hp.Nonce,
					}
					if respBytes, err := EncodeHandshake(respPayload); err == nil {
						respC := &Container{
							SenderPubKey:   n.Identity.Bytes(),
							ReceiverPubKey: c.SenderPubKey,
							SessionID:      "handshake",
							Payload:        respBytes,
							HopLimit:       DefaultHopLimit,
						}
						if err := respC.Sign(n.privateKey); err == nil {
							// Return response over the peer's endpoints
							eps := n.GetPeerEndpoints(senderID)
							if len(eps) > 0 {
								_ = n.forwardContainer(respC, eps)
							}
						}
					}
					continue
				} else if hp.Type == ControlHandshakeResp {
					if len(hp.Endpoints) > 0 {
						n.AddPeerWithLatency(senderID, hp.Endpoints, time.Millisecond)
					}
					if len(hp.X25519Pub) == 32 {
						n.SetPeerEncKey(senderID, hp.X25519Pub)
					}
					n.handshakeMu.Lock()
					ch, exists := n.pendingHandshakes[hp.Nonce]
					n.handshakeMu.Unlock()
					if exists {
						select {
						case ch <- hp:
						default:
						}
					}
					continue
				}
			}

			// Check if packet is a Control Ping/Pong
			if pp, err := DecodePing(c.Payload); err == nil {
				if pp.Type == ControlPing {
					// Reply with Pong
					pong := &PingPayload{
						Type:      ControlPong,
						Nonce:     pp.Nonce,
						Timestamp: pp.Timestamp,
					}
					if pongBytes, err := EncodePing(pong); err == nil {
						pongC := &Container{
							SenderPubKey:   n.Identity.Bytes(),
							ReceiverPubKey: c.SenderPubKey,
							SessionID:      "ping",
							Payload:        pongBytes,
							HopLimit:       DefaultHopLimit,
						}
						if err := pongC.Sign(n.privateKey); err == nil {
							eps := n.GetPeerEndpoints(senderID)
							if len(eps) > 0 {
								_ = n.forwardContainer(pongC, eps)
							}
						}
					}
					continue
				} else if pp.Type == ControlPong {
					n.handshakeMu.Lock()
					ch, exists := n.pendingPings[pp.Nonce]
					n.handshakeMu.Unlock()
					if exists {
						select {
						case ch <- time.Now():
						default:
						}
					}
					continue
				}
			}

			// 1. Destination is this node (or broadcast): deliver locally to all handlers!
			if len(c.ReceiverPubKey) == 0 || bytes.Equal(c.ReceiverPubKey, n.Identity.Bytes()) {
				n.mu.RLock()
				handlers := make([]MessageHandler, len(n.handlers))
				copy(handlers, n.handlers)
				n.mu.RUnlock()

				for _, h := range handlers {
					if h != nil {
						h(senderID, c.Payload)
					}
				}
				continue
			}

			// 2. Multi-hop Small-World forwarding (packet addressed to another peer)
			if c.HopLimit <= 1 {
				// 12-degree threshold reached or loop detected: drop packet
				continue
			}

			c.HopLimit-- // Decrement hop counter

			destID, err := NewIdentityFromBytes(c.ReceiverPubKey)
			if err != nil {
				continue
			}

			// Look up next hop: direct match or closest peer in SmallWorld table
			nextHopEps := n.GetPeerEndpoints(destID)
			if len(nextHopEps) == 0 {
				closest := n.SmallWorld.FindClosestPeers(destID, 3)
				for _, peer := range closest {
					// Don't bounce backward to sender
					if !bytes.Equal(peer.Identity.Bytes(), c.SenderPubKey) {
						nextHopEps = peer.Endpoints
						break
					}
				}
			}

			if len(nextHopEps) > 0 {
				_ = n.forwardContainer(c, nextHopEps)
			}
		}
	}
}
