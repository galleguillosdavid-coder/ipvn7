package onion

import (
	"crypto/ed25519"
	"errors"

	"ipv7/core"
)

const (
	// DefaultMaxHops maximum circuit hops allowed in onion routing
	DefaultMaxHops = 5
	// OnionPacketSize fixed size of an onion packet to prevent traffic analysis by size
	OnionPacketSize = 1280
	// RoutingHeaderSize fixed size of each per-hop routing instruction
	RoutingHeaderSize = 64
)

var (
	ErrInvalidCircuit   = errors.New("circuit must contain at least one relay node")
	ErrCircuitTooLong   = errors.New("circuit exceeds maximum hop count")
	ErrPayloadTooLarge  = errors.New("payload exceeds maximum onion packet capacity")
	ErrInvalidPadding   = errors.New("invalid onion packet padding")
	ErrReplayDetected   = errors.New("onion packet replay detected")
	ErrUnwrapFailed     = errors.New("failed to unwrap onion layer")
)

// CircuitHop represents an intermediary or destination node in the onion circuit
type CircuitHop struct {
	DID       core.Identity
	EncPubKey []byte // 32-byte X25519 public key
	Endpoint  string // Next physical endpoint
}

// LayerInstruction contains next-hop routing info revealed after unwrapping one layer
type LayerInstruction struct {
	IsExit       bool
	NextEndpoint string
	NextPayload  []byte
}

// OnionRelayNode interface for nodes capable of blind packet peeling and forwarding
type OnionRelayNode interface {
	Identity() core.Identity
	EncPrivateKey() ed25519.PrivateKey
	Forward(nextEp string, packet []byte) error
}
