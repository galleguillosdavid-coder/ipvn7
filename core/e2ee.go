package core

import (
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"sync"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

var (
	// keyPool reutiliza buffers de 32 bytes para derivaciones simétricas HKDF
	keyPool = sync.Pool{
		New: func() interface{} {
			b := make([]byte, chacha20poly1305.KeySize)
			return &b
		},
	}
	// noncePool reutiliza buffers de 12 bytes para nonces ChaCha20-Poly1305
	noncePool = sync.Pool{
		New: func() interface{} {
			b := make([]byte, chacha20poly1305.NonceSize)
			return &b
		},
	}
)

// GenerateE2EEKeyPair creates a Curve25519 (X25519) keypair for End-to-End Encryption
func GenerateE2EEKeyPair() (*ecdh.PrivateKey, *ecdh.PublicKey, error) {
	priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return priv, priv.PublicKey(), nil
}

// DeriveX25519FromSeed creates a deterministic X25519 key from a 32-byte seed (e.g. Ed25519 private seed)
func DeriveX25519FromSeed(seed []byte) (*ecdh.PrivateKey, *ecdh.PublicKey, error) {
	if len(seed) < 32 {
		return nil, nil, errors.New("seed must be at least 32 bytes")
	}
	// Hash seed with SHA256 to produce a valid 32-byte X25519 scalar
	hash := sha256.Sum256(seed[:32])
	priv, err := ecdh.X25519().NewPrivateKey(hash[:])
	if err != nil {
		return nil, nil, err
	}
	return priv, priv.PublicKey(), nil
}

// EncryptE2EE encrypts plaintext for a recipient's X25519 public key using ephemeral ECDH + ChaCha20-Poly1305 AEAD.
// Format: [32-byte ephemeral pubkey] + [12-byte nonce] + [ciphertext + 16-byte Poly1305 tag]
func EncryptE2EE(recipientPubKey []byte, plaintext []byte) ([]byte, error) {
	recipPub, err := ecdh.X25519().NewPublicKey(recipientPubKey)
	if err != nil {
		return nil, fmt.Errorf("invalid recipient public key: %w", err)
	}

	// 1. Generate ephemeral sender keypair
	ephemPriv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral key: %w", err)
	}

	// 2. Diffie-Hellman Key Exchange (ECDH)
	sharedSecret, err := ephemPriv.ECDH(recipPub)
	if err != nil {
		return nil, fmt.Errorf("ecdh computation failed: %w", err)
	}

	// 3. Derive 32-byte symmetric key via HKDF-SHA256 (reusing pooled buffer)
	keyPtr := keyPool.Get().(*[]byte)
	defer keyPool.Put(keyPtr)
	symmetricKey := *keyPtr

	hkdfReader := hkdf.New(sha256.New, sharedSecret, nil, []byte("ipv7-e2ee-chacha20poly1305"))
	if _, err := io.ReadFull(hkdfReader, symmetricKey); err != nil {
		return nil, fmt.Errorf("hkdf key derivation failed: %w", err)
	}

	// 4. Initialize ChaCha20-Poly1305 AEAD cipher
	aead, err := chacha20poly1305.New(symmetricKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create aead cipher: %w", err)
	}

	// 5. Generate unique 12-byte nonce (reusing pooled buffer)
	noncePtr := noncePool.Get().(*[]byte)
	defer noncePool.Put(noncePtr)
	nonce := *noncePtr

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 6. Seal ciphertext
	ciphertext := aead.Seal(nil, nonce, plaintext, nil)

	// 7. Assemble final envelope: EphemeralPubKey (32) + Nonce (12) + Ciphertext
	ephemPubBytes := ephemPriv.PublicKey().Bytes()
	result := make([]byte, 0, len(ephemPubBytes)+len(nonce)+len(ciphertext))
	result = append(result, ephemPubBytes...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return result, nil
}

// DecryptE2EE decrypts an envelope produced by EncryptE2EE using the recipient's X25519 private key.
func DecryptE2EE(recipientPrivKey *ecdh.PrivateKey, envelope []byte) ([]byte, error) {
	minSize := 32 + chacha20poly1305.NonceSize + 16 // EphemPub(32) + Nonce(12) + Tag(16)
	if len(envelope) < minSize {
		return nil, errors.New("e2ee envelope too short")
	}

	// Extract components
	ephemPubBytes := envelope[:32]
	nonce := envelope[32 : 32+chacha20poly1305.NonceSize]
	ciphertext := envelope[32+chacha20poly1305.NonceSize:]

	ephemPub, err := ecdh.X25519().NewPublicKey(ephemPubBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid ephemeral public key in envelope: %w", err)
	}

	// Diffie-Hellman Key Exchange (ECDH)
	sharedSecret, err := recipientPrivKey.ECDH(ephemPub)
	if err != nil {
		return nil, fmt.Errorf("ecdh computation failed: %w", err)
	}

	// Derive symmetric key via HKDF (reusing pooled buffer)
	keyPtr := keyPool.Get().(*[]byte)
	defer keyPool.Put(keyPtr)
	symmetricKey := *keyPtr

	hkdfReader := hkdf.New(sha256.New, sharedSecret, nil, []byte("ipv7-e2ee-chacha20poly1305"))
	if _, err := io.ReadFull(hkdfReader, symmetricKey); err != nil {
		return nil, fmt.Errorf("hkdf key derivation failed: %w", err)
	}

	// Decrypt
	aead, err := chacha20poly1305.New(symmetricKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create aead cipher: %w", err)
	}

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (bad key or corrupted/tampered payload): %w", err)
	}

	return plaintext, nil
}
