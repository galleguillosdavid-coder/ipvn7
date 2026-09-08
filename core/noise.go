package core

import (
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

// NoiseSession represents an established, forward-secret encrypted session (PFS)
type NoiseSession struct {
	SessionID   [32]byte
	TxKey       []byte // ChaCha20-Poly1305 key for outbound data
	RxKey       []byte // ChaCha20-Poly1305 key for inbound data
	TxNonce     uint64
	RxNonce     uint64
	PeerEd25519 ed25519.PublicKey
	PeerX25519  *ecdh.PublicKey
	CreatedAt   time.Time
}

// NoiseMsg1 is sent by Initiator -> Responder: -> e (ephemeral X25519 public key)
type NoiseMsg1 struct {
	EphemeralPub []byte // 32 bytes
	Timestamp    int64
	Nonce        uint64
}

// NoiseMsg2 is sent by Responder -> Initiator: <- e, ee, s, es
type NoiseMsg2 struct {
	EphemeralPub []byte // 32 bytes (responder ephemeral)
	StaticPub    []byte // 32 bytes (responder static X25519)
	Ed25519Pub   []byte // 32 bytes (responder identity)
	Signature    []byte // Ed25519 signature over (e_init || e_resp || s_resp)
	Tag          []byte // AEAD auth tag
}

// NoiseMsg3 is sent by Initiator -> Responder: -> s, se
type NoiseMsg3 struct {
	StaticPub  []byte // 32 bytes (initiator static X25519)
	Ed25519Pub []byte // 32 bytes (initiator identity)
	Signature  []byte // Ed25519 signature over (e_resp || e_init || s_init)
	Tag        []byte // AEAD auth tag
}

// NoiseHandshakeState holds state during handshake progression
type NoiseHandshakeState struct {
	isInitiator bool
	staticEd    ed25519.PrivateKey
	staticX     *ecdh.PrivateKey
	ephemX      *ecdh.PrivateKey
	peerEd      ed25519.PublicKey
	peerStaticX *ecdh.PublicKey
	peerEphemX  *ecdh.PublicKey
	sharedEE    []byte
}

// NewNoiseHandshake creates a new handshake state for an initiator or responder
func NewNoiseHandshake(isInitiator bool, staticEd ed25519.PrivateKey, staticX *ecdh.PrivateKey) (*NoiseHandshakeState, error) {
	ephemX, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ephemeral x25519: %w", err)
	}

	return &NoiseHandshakeState{
		isInitiator: isInitiator,
		staticEd:    staticEd,
		staticX:     staticX,
		ephemX:      ephemX,
	}, nil
}

// InitiatorStep1 produces Message 1 (Initiator -> Responder)
func (s *NoiseHandshakeState) InitiatorStep1() (*NoiseMsg1, error) {
	if !s.isInitiator {
		return nil, errors.New("only initiator can call Step1")
	}
	return &NoiseMsg1{
		EphemeralPub: s.ephemX.PublicKey().Bytes(),
		Timestamp:    time.Now().UnixNano(),
		Nonce:        GenerateNonce(),
	}, nil
}

// ResponderStep2 processes Message 1 and produces Message 2 (Responder -> Initiator)
func (s *NoiseHandshakeState) ResponderStep2(msg1 *NoiseMsg1) (*NoiseMsg2, error) {
	if s.isInitiator {
		return nil, errors.New("only responder can call Step2")
	}
	peerEphem, err := ecdh.X25519().NewPublicKey(msg1.EphemeralPub)
	if err != nil {
		return nil, fmt.Errorf("invalid peer ephemeral pub: %w", err)
	}
	s.peerEphemX = peerEphem

	// ee = ECDH(ephem_resp, ephem_init)
	ee, err := s.ephemX.ECDH(s.peerEphemX)
	if err != nil {
		return nil, fmt.Errorf("ecdh ee failed: %w", err)
	}
	s.sharedEE = ee

	// es = ECDH(static_resp, ephem_init)
	es, err := s.staticX.ECDH(s.peerEphemX)
	if err != nil {
		return nil, fmt.Errorf("ecdh es failed: %w", err)
	}

	// Sign transcripts: e_init || e_resp || s_resp
	toSign := append(msg1.EphemeralPub, s.ephemX.PublicKey().Bytes()...)
	toSign = append(toSign, s.staticX.PublicKey().Bytes()...)
	sig := ed25519.Sign(s.staticEd, toSign)

	// Derive intermediate key from ee + es for payload encryption/authentication
	kdf := hkdf.New(sha256.New, append(ee, es...), nil, []byte("ipv7-noise-step2"))
	key := make([]byte, chacha20poly1305.KeySize)
	if _, err := io.ReadFull(kdf, key); err != nil {
		return nil, err
	}

	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize()) // 12-byte zero nonce for initial handshake tag
	tag := aead.Seal(nil, nonce, []byte("IPV7_NOISE_PFS_OK"), toSign)

	edPub := s.staticEd.Public().(ed25519.PublicKey)

	return &NoiseMsg2{
		EphemeralPub: s.ephemX.PublicKey().Bytes(),
		StaticPub:    s.staticX.PublicKey().Bytes(),
		Ed25519Pub:   edPub,
		Signature:    sig,
		Tag:          tag,
	}, nil
}

// InitiatorStep3 processes Message 2 and produces Message 3 + final NoiseSession
func (s *NoiseHandshakeState) InitiatorStep3(msg1 *NoiseMsg1, msg2 *NoiseMsg2) (*NoiseMsg3, *NoiseSession, error) {
	if !s.isInitiator {
		return nil, nil, errors.New("only initiator can call Step3")
	}

	// Verify responder Ed25519 signature
	toVerify := append(msg1.EphemeralPub, msg2.EphemeralPub...)
	toVerify = append(toVerify, msg2.StaticPub...)
	if !ed25519.Verify(msg2.Ed25519Pub, toVerify, msg2.Signature) {
		return nil, nil, errors.New("invalid responder ed25519 signature")
	}

	peerEphem, err := ecdh.X25519().NewPublicKey(msg2.EphemeralPub)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid peer ephem: %w", err)
	}
	s.peerEphemX = peerEphem

	peerStatic, err := ecdh.X25519().NewPublicKey(msg2.StaticPub)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid peer static: %w", err)
	}
	s.peerStaticX = peerStatic
	s.peerEd = msg2.Ed25519Pub

	// ee = ECDH(ephem_init, ephem_resp)
	ee, err := s.ephemX.ECDH(s.peerEphemX)
	if err != nil {
		return nil, nil, fmt.Errorf("initiator ecdh ee failed: %w", err)
	}
	s.sharedEE = ee

	// es = ECDH(ephem_init, static_resp)
	es, err := s.ephemX.ECDH(s.peerStaticX)
	if err != nil {
		return nil, nil, fmt.Errorf("initiator ecdh es failed: %w", err)
	}

	// Verify Step2 AEAD Tag
	kdf := hkdf.New(sha256.New, append(ee, es...), nil, []byte("ipv7-noise-step2"))
	key := make([]byte, chacha20poly1305.KeySize)
	if _, err := io.ReadFull(kdf, key); err != nil {
		return nil, nil, err
	}
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := aead.Open(nil, nonce, msg2.Tag, toVerify); err != nil {
		return nil, nil, errors.New("noise step2 aead verification failed")
	}

	// se = ECDH(static_init, ephem_resp)
	se, err := s.staticX.ECDH(s.peerEphemX)
	if err != nil {
		return nil, nil, fmt.Errorf("initiator ecdh se failed: %w", err)
	}

	// Sign Msg3 transcript: e_resp || e_init || s_init
	toSign3 := append(msg2.EphemeralPub, msg1.EphemeralPub...)
	toSign3 = append(toSign3, s.staticX.PublicKey().Bytes()...)
	sig3 := ed25519.Sign(s.staticEd, toSign3)

	// Step3 Tag
	kdf3 := hkdf.New(sha256.New, append(ee, se...), nil, []byte("ipv7-noise-step3"))
	key3 := make([]byte, chacha20poly1305.KeySize)
	if _, err := io.ReadFull(kdf3, key3); err != nil {
		return nil, nil, err
	}
	aead3, _ := chacha20poly1305.New(key3)
	tag3 := aead3.Seal(nil, nonce, []byte("IPV7_NOISE_PFS_FIN"), toSign3)

	edPub := s.staticEd.Public().(ed25519.PublicKey)
	msg3 := &NoiseMsg3{
		StaticPub:  s.staticX.PublicKey().Bytes(),
		Ed25519Pub: edPub,
		Signature:  sig3,
		Tag:        tag3,
	}

	// Final Session Keys derivation: combines ee (PFS) + es + se
	session, err := deriveNoiseSession(ee, es, se, s.peerEd, s.peerStaticX, true)
	if err != nil {
		return nil, nil, err
	}

	return msg3, session, nil
}

// ResponderFinal processes Message 3 and establishes final NoiseSession
func (s *NoiseHandshakeState) ResponderFinal(msg1 *NoiseMsg1, msg2 *NoiseMsg2, msg3 *NoiseMsg3) (*NoiseSession, error) {
	if s.isInitiator {
		return nil, errors.New("only responder can call ResponderFinal")
	}

	// Verify initiator Ed25519 signature
	toVerify3 := append(msg2.EphemeralPub, msg1.EphemeralPub...)
	toVerify3 = append(toVerify3, msg3.StaticPub...)
	if !ed25519.Verify(msg3.Ed25519Pub, toVerify3, msg3.Signature) {
		return nil, errors.New("invalid initiator ed25519 signature in msg3")
	}

	peerStatic, err := ecdh.X25519().NewPublicKey(msg3.StaticPub)
	if err != nil {
		return nil, fmt.Errorf("invalid peer static: %w", err)
	}
	s.peerStaticX = peerStatic
	s.peerEd = msg3.Ed25519Pub

	// se = ECDH(ephem_resp, static_init)
	se, err := s.ephemX.ECDH(s.peerStaticX)
	if err != nil {
		return nil, fmt.Errorf("responder ecdh se failed: %w", err)
	}

	// es = ECDH(static_resp, ephem_init)
	es, err := s.staticX.ECDH(s.peerEphemX)
	if err != nil {
		return nil, fmt.Errorf("responder ecdh es failed: %w", err)
	}

	// Verify Step3 Tag
	kdf3 := hkdf.New(sha256.New, append(s.sharedEE, se...), nil, []byte("ipv7-noise-step3"))
	key3 := make([]byte, chacha20poly1305.KeySize)
	if _, err := io.ReadFull(kdf3, key3); err != nil {
		return nil, err
	}
	aead3, _ := chacha20poly1305.New(key3)
	nonce := make([]byte, aead3.NonceSize())
	if _, err := aead3.Open(nil, nonce, msg3.Tag, toVerify3); err != nil {
		return nil, errors.New("noise step3 aead verification failed")
	}

	// Final Session Keys derivation (responder role)
	return deriveNoiseSession(s.sharedEE, es, se, s.peerEd, s.peerStaticX, false)
}

func deriveNoiseSession(ee, es, se []byte, peerEd ed25519.PublicKey, peerStaticX *ecdh.PublicKey, isInitiator bool) (*NoiseSession, error) {
	// Mixed key material: ee (ephemeral-ephemeral PFS) + es + se
	material := make([]byte, 0, len(ee)+len(es)+len(se))
	material = append(material, ee...)
	material = append(material, es...)
	material = append(material, se...)

	// HKDF derive TxKey (32), RxKey (32), SessionID (32)
	kdf := hkdf.New(sha256.New, material, nil, []byte("ipv7-noise-session-v1"))

	k1 := make([]byte, chacha20poly1305.KeySize)
	k2 := make([]byte, chacha20poly1305.KeySize)
	var sessionID [32]byte

	if _, err := io.ReadFull(kdf, k1); err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(kdf, k2); err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(kdf, sessionID[:]); err != nil {
		return nil, err
	}

	session := &NoiseSession{
		SessionID:   sessionID,
		PeerEd25519: peerEd,
		PeerX25519:  peerStaticX,
		CreatedAt:   time.Now(),
	}

	if isInitiator {
		session.TxKey = k1
		session.RxKey = k2
	} else {
		session.TxKey = k2
		session.RxKey = k1
	}

	return session, nil
}

// Encrypt encrypts a message with the session TxKey and increments TxNonce
func (ns *NoiseSession) Encrypt(plaintext []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(ns.TxKey)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	// Encode 64-bit counter in big-endian into first 8 bytes of nonce
	for i := 0; i < 8; i++ {
		nonce[i] = byte(ns.TxNonce >> (56 - (i * 8)))
	}
	ns.TxNonce++

	ciphertext := aead.Seal(nil, nonce, plaintext, nil)
	// Output: [12 bytes nonce] + [ciphertext + tag]
	out := append(nonce, ciphertext...)
	return out, nil
}

// Decrypt decrypts a message using RxKey and verifies authenticated tag
func (ns *NoiseSession) Decrypt(data []byte) ([]byte, error) {
	minSize := chacha20poly1305.NonceSize + 16
	if len(data) < minSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce := data[:chacha20poly1305.NonceSize]
	ciphertext := data[chacha20poly1305.NonceSize:]

	aead, err := chacha20poly1305.New(ns.RxKey)
	if err != nil {
		return nil, err
	}
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("noise decryption failed: %w", err)
	}
	ns.RxNonce++
	return plaintext, nil
}
