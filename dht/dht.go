package dht

import (
	"bytes"
	"crypto/ed25519"
	"errors"
	"fmt"
	"sync"
	"time"

	"ipv7/core"
)

// DHTService implements the decentralized discovery service for IPv7
type DHTService struct {
	identity   *core.Ed25519Identity
	privateKey ed25519.PrivateKey
	table      *core.SmallWorldTable

	// store maps string(PublicKey) -> *Record
	store map[string]*Record
	mu    sync.RWMutex

	// remoteCaller is an optional network dispatcher func(destEp string, req *Message) (*Message, error)
	remoteCaller func(destEp string, req *Message) (*Message, error)
}

// NewDHTService creates a new DHT node
func NewDHTService(id *core.Ed25519Identity, privKey ed25519.PrivateKey) *DHTService {
	return &DHTService{
		identity:   id,
		privateKey: privKey,
		table:      core.NewSmallWorldTable(id, core.DefaultMaxDegrees, core.DefaultPeersPerDegree),
		store:      make(map[string]*Record),
	}
}

// SetRemoteCaller sets the transport handler for RPC calls
func (s *DHTService) SetRemoteCaller(caller func(destEp string, req *Message) (*Message, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.remoteCaller = caller
}

// AddPeer adds a known DHT peer into the routing table
func (s *DHTService) AddPeer(peerID core.Identity, endpoints []string) {
	s.table.AddPeer(peerID, endpoints, 0)
}

// Publish self-announces our identity and endpoints into the DHT
func (s *DHTService) Publish(endpoints []string, ttl time.Duration) (*Record, error) {
	rec := NewRecord(s.identity.PublicKey, endpoints, ttl)
	if err := rec.Sign(s.privateKey); err != nil {
		return nil, fmt.Errorf("failed to sign dht record: %w", err)
	}

	// Store locally
	s.mu.Lock()
	s.store[string(s.identity.Bytes())] = rec
	s.mu.Unlock()

	// Broadcast store request to closest peers
	closest := s.table.FindClosestPeers(s.identity, 3)
	for _, peer := range closest {
		if s.remoteCaller != nil && len(peer.Endpoints) > 0 {
			req := &Message{
				Type:     MsgStore,
				SenderID: s.identity.Bytes(),
				Record:   rec,
			}
			go func(ep string) {
				_, _ = s.remoteCaller(ep, req)
			}(peer.Endpoints[0])
		}
	}

	return rec, nil
}

// Resolve looks up the endpoints for a target public key in the DHT
func (s *DHTService) Resolve(targetPubKey []byte) (*Record, error) {
	// 1. Check local cache
	s.mu.RLock()
	rec, found := s.store[string(targetPubKey)]
	s.mu.RUnlock()

	if found && rec.Verify() {
		return rec, nil
	}

	// 2. Query closest peers
	targetID, err := core.NewIdentityFromBytes(targetPubKey)
	if err != nil {
		return nil, errors.New("invalid target public key")
	}

	closest := s.table.FindClosestPeers(targetID, 3)
	for _, peer := range closest {
		if s.remoteCaller == nil || len(peer.Endpoints) == 0 {
			continue
		}

		req := &Message{
			Type:      MsgFindValue,
			SenderID:  s.identity.Bytes(),
			TargetKey: targetPubKey,
		}

		resp, err := s.remoteCaller(peer.Endpoints[0], req)
		if err == nil && resp != nil && resp.Record != nil {
			if resp.Record.Verify() && bytes.Equal(resp.Record.PublicKey, targetPubKey) {
				// Cache valid record
				s.mu.Lock()
				s.store[string(targetPubKey)] = resp.Record
				s.mu.Unlock()
				return resp.Record, nil
			}
		}
	}

	return nil, fmt.Errorf("peer record not found in dht")
}

// ProcessMessage handles incoming DHT RPC requests and produces responses
func (s *DHTService) ProcessMessage(req *Message) *Message {
	if req == nil {
		return nil
	}

	switch req.Type {
	case MsgPing:
		return &Message{
			Type:     MsgPong,
			SenderID: s.identity.Bytes(),
		}

	case MsgStore:
		if req.Record == nil || !req.Record.Verify() {
			return nil // Reject invalid or forged records
		}

		s.mu.Lock()
		s.store[string(req.Record.PublicKey)] = req.Record
		s.mu.Unlock()

		return &Message{
			Type:     MsgPong,
			SenderID: s.identity.Bytes(),
		}

	case MsgFindValue:
		s.mu.RLock()
		rec, found := s.store[string(req.TargetKey)]
		s.mu.RUnlock()

		if found && rec.Verify() {
			return &Message{
				Type:     MsgValueResponse,
				SenderID: s.identity.Bytes(),
				Record:   rec,
			}
		}

		// Not found locally: return closest contacts for iterative lookup
		targetID, err := core.NewIdentityFromBytes(req.TargetKey)
		if err != nil {
			return nil
		}

		closest := s.table.FindClosestPeers(targetID, 5)
		var peers []PeerInfo
		for _, p := range closest {
			peers = append(peers, PeerInfo{
				PublicKey: p.Identity.Bytes(),
				Endpoints: p.Endpoints,
			})
		}

		return &Message{
			Type:     MsgValueResponse,
			SenderID: s.identity.Bytes(),
			Peers:    peers,
		}

	default:
		return nil
	}
}
