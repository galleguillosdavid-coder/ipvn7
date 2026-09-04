package dht

import (
	"crypto/ed25519"
	"errors"
	"time"

	"github.com/fxamacker/cbor/v2"
)

var cborEnc cbor.EncMode

func init() {
	var err error
	cborEnc, err = cbor.CanonicalEncOptions().EncMode()
	if err != nil {
		panic(err)
	}
}

// Record represents a self-authenticating entry in the IPv7 DHT
// mapping a cryptographic public key to its physical network endpoints.
type Record struct {
	PublicKey []byte   `cbor:"1,keyasint"`
	Endpoints []string `cbor:"2,keyasint"`
	Timestamp int64    `cbor:"3,keyasint"`
	TTL       int64    `cbor:"4,keyasint"`
	Signature []byte   `cbor:"5,keyasint,omitempty"`
}

// NewRecord creates a new DHT record for an identity
func NewRecord(pubKey ed25519.PublicKey, endpoints []string, ttl time.Duration) *Record {
	return &Record{
		PublicKey: pubKey,
		Endpoints: endpoints,
		Timestamp: time.Now().Unix(),
		TTL:       int64(ttl.Seconds()),
	}
}

// Sign hashes the record fields deterministically and signs with the identity's private key.
// Only the legitimate owner of the private key can publish their endpoints!
func (r *Record) Sign(privKey ed25519.PrivateKey) error {
	r.Signature = nil
	data, err := cborEnc.Marshal(r)
	if err != nil {
		return err
	}
	r.Signature = ed25519.Sign(privKey, data)
	return nil
}

// Verify ensures the record has not been tampered with and was signed by the PublicKey advertised.
// This completely mitigates Sybil endpoint-spoofing in the DHT!
func (r *Record) Verify() bool {
	if len(r.PublicKey) != ed25519.PublicKeySize || len(r.Signature) != ed25519.SignatureSize {
		return false
	}

	// Check expiration
	now := time.Now().Unix()
	if r.TTL > 0 && now > r.Timestamp+r.TTL {
		return false // Record expired
	}

	sig := r.Signature
	r.Signature = nil
	defer func() { r.Signature = sig }()

	data, err := cborEnc.Marshal(r)
	if err != nil {
		return false
	}

	return ed25519.Verify(r.PublicKey, data, sig)
}

// Marshal serializes the record into deterministic CBOR
func (r *Record) Marshal() ([]byte, error) {
	return cborEnc.Marshal(r)
}

// Unmarshal parses CBOR bytes into a Record
func (r *Record) Unmarshal(data []byte) error {
	if err := cbor.Unmarshal(data, r); err != nil {
		return err
	}
	if !r.Verify() {
		return errors.New("dht record verification failed: invalid signature or expired")
	}
	return nil
}
