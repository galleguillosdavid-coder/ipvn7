package dht

import "github.com/fxamacker/cbor/v2"

type MessageType uint8

const (
	MsgPing MessageType = iota + 1
	MsgPong
	MsgStore
	MsgFindValue
	MsgValueResponse
)

// PeerInfo represents contact information for a peer in the DHT
type PeerInfo struct {
	PublicKey []byte   `cbor:"1,keyasint"`
	Endpoints []string `cbor:"2,keyasint"`
}

// Message represents a DHT RPC request or response
type Message struct {
	Type      MessageType `cbor:"1,keyasint"`
	SenderID  []byte      `cbor:"2,keyasint"`
	TargetKey []byte      `cbor:"3,keyasint,omitempty"`
	Record    *Record     `cbor:"4,keyasint,omitempty"`
	Peers     []PeerInfo  `cbor:"5,keyasint,omitempty"`
}

func (m *Message) Marshal() ([]byte, error) {
	return cborEnc.Marshal(m)
}

func (m *Message) Unmarshal(data []byte) error {
	return cbor.Unmarshal(data, m)
}
