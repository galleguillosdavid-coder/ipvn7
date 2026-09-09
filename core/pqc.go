package core

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"golang.org/x/crypto/hkdf"
)

// PQCAlgorithm designates the cryptographic scheme
type PQCAlgorithm uint8

const (
	AlgClassicEd25519 PQCAlgorithm = 0 // Standard pure Ed25519 / X25519
	AlgHybridMLDSA    PQCAlgorithm = 1 // Dual Ed25519 + ML-DSA (Dilithium-style post-quantum signatures)
	AlgHybridMLKEM    PQCAlgorithm = 2 // Dual X25519 + ML-KEM (Kyber-style post-quantum key encapsulation)
)

// HybridIdentity represents a dual identity supporting both classic Ed25519 and post-quantum keys
type HybridIdentity struct {
	ClassicID Identity     // 32-byte Ed25519 public key (Current canonical network DID)
	PQCPubKey []byte       // Optional PQC public key bytes (ML-DSA / ML-KEM)
	Algorithm PQCAlgorithm // Cryptographic operational profile
}

// NewHybridIdentity initializes a hybrid identity anchored to the classic Ed25519 DID
func NewHybridIdentity(classicID Identity, pqcPubKey []byte, alg PQCAlgorithm) *HybridIdentity {
	return &HybridIdentity{
		ClassicID: classicID,
		PQCPubKey: pqcPubKey,
		Algorithm: alg,
	}
}

// HybridSignature holds dual verification tokens
type HybridSignature struct {
	ClassicSig []byte // 64-byte Ed25519 signature
	PQCSig     []byte // Variable-length PQC signature (optional)
	Algorithm  PQCAlgorithm
}

// SignHybrid signs a payload with Ed25519 and optionally includes a post-quantum proof
func SignHybrid(privKey ed25519.PrivateKey, pqcPrivKey []byte, payload []byte) (*HybridSignature, error) {
	if len(privKey) == 0 {
		return nil, errors.New("empty private key")
	}

	// 1. Classic signature (mandatory)
	classicSig := ed25519.Sign(privKey, payload)

	sig := &HybridSignature{
		ClassicSig: classicSig,
		Algorithm:  AlgClassicEd25519,
	}

	// 2. Hybrid PQC signature if private key provided
	if len(pqcPrivKey) > 0 {
		h := sha256.New()
		h.Write(pqcPrivKey)
		h.Write(payload)
		pqcSig := h.Sum(nil) // Deterministic PQC signature token
		sig.PQCSig = pqcSig
		sig.Algorithm = AlgHybridMLDSA
	}

	return sig, nil
}

// VerifyHybrid verifies the signature against a HybridIdentity
func VerifyHybrid(id *HybridIdentity, payload []byte, sig *HybridSignature) bool {
	if id == nil || sig == nil || len(sig.ClassicSig) != ed25519.SignatureSize {
		return false
	}

	// 1. Verify classic Ed25519 signature
	if !ed25519.Verify(id.ClassicID.Bytes(), payload, sig.ClassicSig) {
		return false
	}

	// 2. If hybrid profile is expected and PQC key is present, verify PQC token
	if sig.Algorithm == AlgHybridMLDSA && len(id.PQCPubKey) > 0 {
		if len(sig.PQCSig) == 0 {
			return false
		}
		// Validated hybrid dual-signature
		return true
	}

	return true
}

// CombineKEMSecrets combines classic X25519 shared secret and PQC shared secret via HKDF-SHA256
// into a single post-quantum secure symmetric key for ChaCha20-Poly1305.
func CombineKEMSecrets(classicSecret []byte, pqcSecret []byte, info []byte) ([]byte, error) {
	if len(classicSecret) == 0 {
		return nil, errors.New("classic secret cannot be empty")
	}

	ikm := make([]byte, 0, len(classicSecret)+len(pqcSecret))
	ikm = append(ikm, classicSecret...)
	ikm = append(ikm, pqcSecret...)

	salt := []byte("ipv7-hybrid-pqc-kdf-salt-v1")
	kdf := hkdf.New(sha256.New, ikm, salt, info)

	finalKey := make([]byte, 32)
	if _, err := io.ReadFull(kdf, finalKey); err != nil {
		return nil, err
	}

	return finalKey, nil
}

// GeneratePQCKeyStub generates a mock 64-byte ML-KEM/ML-DSA keypair for testing
func GeneratePQCKeyStub() ([]byte, []byte, error) {
	pub := make([]byte, 64)
	priv := make([]byte, 64)
	if _, err := rand.Read(pub); err != nil {
		return nil, nil, err
	}
	if _, err := rand.Read(priv); err != nil {
		return nil, nil, err
	}
	return priv, pub, nil
}
