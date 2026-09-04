package core

import (
	"fmt"
	"sync"

	"github.com/fxamacker/cbor/v2"
)

// StreamChunk represents a chunk of media or file data propagating down the cascade
type StreamChunk struct {
	StreamID  string `cbor:"1,keyasint"`
	Sequence  uint64 `cbor:"2,keyasint"`
	Origin    []byte `cbor:"3,keyasint"` // Originator's public key
	Payload   []byte `cbor:"4,keyasint"`
	Signature []byte `cbor:"5,keyasint,omitempty"`
}

// CascadeNode manages tree-based multicast distribution for streaming
type CascadeNode struct {
	node        *Node
	maxChildren int

	mu          sync.RWMutex
	children    map[string]Identity // Hex -> Identity
	parentID    Identity
	parentEp    string

	onChunk func(chunk *StreamChunk)
}

// NewCascadeNode creates a tree forwarder wrapping an existing IPv7 node
func NewCascadeNode(node *Node, maxChildren int) *CascadeNode {
	if maxChildren <= 0 {
		maxChildren = 10 // Default cascade fan-out of 10 peers
	}
	cn := &CascadeNode{
		node:        node,
		maxChildren: maxChildren,
		children:    make(map[string]Identity),
	}

	// Register message handler on node to process cascade stream chunks
	node.OnMessage(func(from Identity, payload []byte) {
		chunk := &StreamChunk{}
		if err := cbor.Unmarshal(payload, chunk); err != nil {
			return
		}

		// Notify local subscriber
		cn.mu.RLock()
		handler := cn.onChunk
		cn.mu.RUnlock()

		if handler != nil {
			handler(chunk)
		}

		// Forward down the cascade to children (amplification)
		cn.forwardToChildren(payload)
	})

	return cn
}

// SetParent registers the upstream feed provider
func (cn *CascadeNode) SetParent(parent Identity, endpoint string) {
	cn.mu.Lock()
	defer cn.mu.Unlock()
	cn.parentID = parent
	cn.parentEp = endpoint
	cn.node.AddPeer(parent, []string{endpoint})
}

// AddChild attaches a downstream peer to receive the stream
func (cn *CascadeNode) AddChild(child Identity, endpoint string) error {
	cn.mu.Lock()
	defer cn.mu.Unlock()

	if len(cn.children) >= cn.maxChildren {
		return fmt.Errorf("cascade node reached max capacity (%d children)", cn.maxChildren)
	}

	cn.children[child.String()] = child
	cn.node.AddPeer(child, []string{endpoint})
	return nil
}

// ChildrenCount returns current number of attached children
func (cn *CascadeNode) ChildrenCount() int {
	cn.mu.RLock()
	defer cn.mu.RUnlock()
	return len(cn.children)
}

// OnChunk registers a listener for incoming stream chunks
func (cn *CascadeNode) OnChunk(handler func(chunk *StreamChunk)) {
	cn.mu.Lock()
	defer cn.mu.Unlock()
	cn.onChunk = handler
}

// Broadcast sends a new stream chunk origin from this node down to all children
func (cn *CascadeNode) Broadcast(streamID string, seq uint64, data []byte) error {
	chunk := &StreamChunk{
		StreamID: streamID,
		Sequence: seq,
		Origin:   cn.node.Identity.Bytes(),
		Payload:  data,
	}

	encoded, err := cborEnc.Marshal(chunk)
	if err != nil {
		return fmt.Errorf("failed to encode stream chunk: %w", err)
	}

	cn.forwardToChildren(encoded)
	return nil
}

func (cn *CascadeNode) forwardToChildren(encodedChunk []byte) {
	cn.mu.RLock()
	targets := make([]Identity, 0, len(cn.children))
	for _, child := range cn.children {
		targets = append(targets, child)
	}
	cn.mu.RUnlock()

	for _, target := range targets {
		_ = cn.node.SendMessage(target, encodedChunk)
	}
}
