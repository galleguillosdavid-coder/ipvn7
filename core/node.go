package core

import (
	"bytes"
	"crypto/ecdh"
	"crypto/ed25519"
	"errors"
	"fmt"
	"sync"
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
	peerTable map[string][]string
	mu        sync.RWMutex

	handler MessageHandler
	running bool
	stopCh  chan struct{}
}

// NewNode initializes a new IPv7 Node with its cryptographic identity and derived E2EE key
func NewNode(identity *Ed25519Identity, privateKey ed25519.PrivateKey) *Node {
	encPriv, encPub, _ := DeriveX25519FromSeed(privateKey.Seed())

	return &Node{
		Identity:   identity,
		privateKey: privateKey,
		EncPrivKey: encPriv,
		EncPubKey:  encPub,
		SmallWorld: NewSmallWorldTable(identity, DefaultMaxDegrees, DefaultPeersPerDegree),
		peerTable:  make(map[string][]string),
		stopCh:     make(chan struct{}),
	}
}

// AddAdapter registers a transport adapter (UDP, QUIC, Relay) to this node
func (n *Node) AddAdapter(adapter Adapter) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.adapters = append(n.adapters, adapter)
}

// OnMessage registers the listener callback for incoming payload
func (n *Node) OnMessage(handler MessageHandler) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.handler = handler
}

// AddPeer registers known endpoints for a peer's identity and records it in the SmallWorld table
func (n *Node) AddPeer(peerID Identity, endpoints []string) {
	n.mu.Lock()
	n.peerTable[peerID.String()] = endpoints
	n.mu.Unlock()

	n.SmallWorld.AddPeer(peerID, endpoints, 0)
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

// SendMessage sends an authenticated data payload to a target peer identity
func (n *Node) SendMessage(target Identity, payload []byte) error {
	endpoints := n.GetPeerEndpoints(target)
	
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

			// Verify end-to-end signature
			if !c.Verify() {
				continue
			}

			// 1. Destination is this node (or broadcast): deliver locally!
			if len(c.ReceiverPubKey) == 0 || bytes.Equal(c.ReceiverPubKey, n.Identity.Bytes()) {
				senderID, err := NewIdentityFromBytes(c.SenderPubKey)
				if err != nil {
					continue
				}

				n.mu.RLock()
				handler := n.handler
				n.mu.RUnlock()

				if handler != nil {
					handler(senderID, c.Payload)
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
