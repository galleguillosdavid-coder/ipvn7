package core

import (
	"crypto/ed25519"
	"github.com/fxamacker/cbor/v2"
)

// Container represents the atomic unit of data in IPv7
type Container struct {
	// Sender's public key (Identity)
	SenderPubKey []byte `cbor:"1,keyasint"`
	
	// Receiver's public key (Identity), can be empty for broadcast/DHT
	ReceiverPubKey []byte `cbor:"2,keyasint,omitempty"`
	
	// Session ID (if applicable)
	SessionID string `cbor:"3,keyasint,omitempty"`
	
	// Data payload (e.g. chunks or control messages)
	Payload []byte `cbor:"4,keyasint"`
	
	// Signature of the container (excluding the signature and hoplimit fields)
	Signature []byte `cbor:"5,keyasint,omitempty"`

	// HopLimit restricts propagation depth (default: 12 degrees of separation)
	HopLimit uint8 `cbor:"6,keyasint,omitempty"`
}

// cborEnc is a canonical CBOR encoder
var cborEnc cbor.EncMode

func init() {
	var err error
	cborEnc, err = cbor.CanonicalEncOptions().EncMode()
	if err != nil {
		panic(err)
	}
}

// Sign encodes the container (without signature and hoplimit) and signs it using the provided private key
func (c *Container) Sign(privKey ed25519.PrivateKey) error {
	// Temporarily clear signature and hoplimit for deterministic serialization
	c.Signature = nil
	hop := c.HopLimit
	c.HopLimit = 0
	defer func() { c.HopLimit = hop }()

	data, err := cborEnc.Marshal(c)
	if err != nil {
		return err
	}
	
	c.Signature = ed25519.Sign(privKey, data)
	return nil
}

// Verify checks if the container's signature is valid using the SenderPubKey
func (c *Container) Verify() bool {
	if len(c.SenderPubKey) != ed25519.PublicKeySize || len(c.Signature) != ed25519.SignatureSize {
		return false
	}
	
	sig := c.Signature
	c.Signature = nil
	hop := c.HopLimit
	c.HopLimit = 0
	defer func() {
		c.Signature = sig
		c.HopLimit = hop
	}()
	
	data, err := cborEnc.Marshal(c)
	if err != nil {
		return false
	}
	
	return ed25519.Verify(c.SenderPubKey, data, sig)
}

// Marshal returns the CBOR encoded container
func (c *Container) Marshal() ([]byte, error) {
	return cborEnc.Marshal(c)
}

// Unmarshal parses CBOR data into the container
func (c *Container) Unmarshal(data []byte) error {
	return cbor.Unmarshal(data, c)
}
