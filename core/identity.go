package core

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
)

// Ed25519Identity implements the Identity interface
type Ed25519Identity struct {
	PublicKey ed25519.PublicKey
}

// GenerateIdentity creates a new Ed25519 keypair and returns the Identity and PrivateKey
func GenerateIdentity() (*Ed25519Identity, ed25519.PrivateKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return &Ed25519Identity{PublicKey: pub}, priv, nil
}

// NewIdentityFromBytes creates an Identity from a raw public key
func NewIdentityFromBytes(pubKey []byte) (*Ed25519Identity, error) {
	if len(pubKey) != ed25519.PublicKeySize {
		return nil, errors.New("invalid public key size")
	}
	return &Ed25519Identity{PublicKey: pubKey}, nil
}

func (id *Ed25519Identity) String() string {
	return hex.EncodeToString(id.PublicKey)
}

func (id *Ed25519Identity) Bytes() []byte {
	return id.PublicKey
}

func (id *Ed25519Identity) Verify(data, signature []byte) bool {
	return ed25519.Verify(id.PublicKey, data, signature)
}
