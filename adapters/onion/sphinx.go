package onion

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
)

// Magic header prefix for Onion packets
var OnionMagic = [4]byte{0x7F, 0x4F, 0x4E, 0x37} // ~ON7

// BuildOnionPacket constructs a fixed-size Sphinx-like multi-hop onion packet
func BuildOnionPacket(circuit []CircuitHop, finalPayload []byte) ([]byte, error) {
	if len(circuit) == 0 {
		return nil, ErrInvalidCircuit
	}
	if len(circuit) > DefaultMaxHops {
		return nil, ErrCircuitTooLong
	}
	if len(finalPayload) > OnionPacketSize-256 {
		return nil, ErrPayloadTooLarge
	}

	// 1. Start from inner-most layer (exit hop)
	currentPayload := make([]byte, len(finalPayload))
	copy(currentPayload, finalPayload)

	// We wrap layers in reverse order: from destination back to first hop
	for i := len(circuit) - 1; i >= 0; i-- {
		hop := circuit[i]
		isExit := (i == len(circuit)-1)

		// Generate ephemeral X25519 keypair for this hop
		var ephPriv, ephPub [32]byte
		if _, err := io.ReadFull(rand.Reader, ephPriv[:]); err != nil {
			return nil, err
		}
		curve25519.ScalarBaseMult(&ephPub, &ephPriv)

		// Shared secret via ECDH with hop's public encryption key
		var hopPub [32]byte
		copy(hopPub[:], hop.EncPubKey)
		sharedSecret, err := curve25519.X25519(ephPriv[:], hopPub[:])
		if err != nil {
			return nil, err
		}

		// Derive symmetric ChaCha20-Poly1305 key
		keyHash := sha256.Sum256(sharedSecret)
		aead, err := chacha20poly1305.New(keyHash[:])
		if err != nil {
			return nil, err
		}

		// Routing instruction: [isExit (1B)] [epLen (2B)] [nextEp] [payloadLen (4B)] [payload]
		nextEp := ""
		if !isExit && i+1 < len(circuit) {
			nextEp = circuit[i+1].Endpoint
		}

		buf := new(bytes.Buffer)
		if isExit {
			buf.WriteByte(1)
		} else {
			buf.WriteByte(0)
		}

		epBytes := []byte(nextEp)
		var epLenBuf [2]byte
		binary.BigEndian.PutUint16(epLenBuf[:], uint16(len(epBytes)))
		buf.Write(epLenBuf[:])
		buf.Write(epBytes)

		var pLenBuf [4]byte
		binary.BigEndian.PutUint32(pLenBuf[:], uint32(len(currentPayload)))
		buf.Write(pLenBuf[:])
		buf.Write(currentPayload)

		plainLayer := buf.Bytes()

		// Encrypt with AEAD using deterministic nonce derived from ephemeral public key
		nonceHash := sha256.Sum256(ephPub[:])
		nonce := nonceHash[:aead.NonceSize()]
		ciphertext := aead.Seal(nil, nonce, plainLayer, nil)

		// Wrap: [Magic (4B)] [EphPub (32B)] [Ciphertext]
		layerBuf := new(bytes.Buffer)
		layerBuf.Write(OnionMagic[:])
		layerBuf.Write(ephPub[:])
		layerBuf.Write(ciphertext)

		currentPayload = layerBuf.Bytes()
	}

	// Pad final outer packet to exact fixed OnionPacketSize
	if len(currentPayload) < OnionPacketSize {
		padded := make([]byte, OnionPacketSize)
		copy(padded, currentPayload)
		// Trailing zero padding
		return padded, nil
	}

	return currentPayload, nil
}

// UnwrapLayer peels one layer of the onion using the node's private encryption key
func UnwrapLayer(rawPacket []byte, nodeX25519Priv [32]byte) (*LayerInstruction, error) {
	if len(rawPacket) < 36 {
		return nil, ErrUnwrapFailed
	}

	// Verify magic prefix
	if !bytes.Equal(rawPacket[:4], OnionMagic[:]) {
		return nil, ErrUnwrapFailed
	}

	var ephPub [32]byte
	copy(ephPub[:], rawPacket[4:36])
	ciphertextWithTag := rawPacket[36:]

	// Derive shared secret
	sharedSecret, err := curve25519.X25519(nodeX25519Priv[:], ephPub[:])
	if err != nil {
		return nil, err
	}

	keyHash := sha256.Sum256(sharedSecret)
	aead, err := chacha20poly1305.New(keyHash[:])
	if err != nil {
		return nil, err
	}

	nonceHash := sha256.Sum256(ephPub[:])
	nonce := nonceHash[:aead.NonceSize()]

	// If packet was padded to fixed size, trim trailing zeroes or decrypt directly
	plaintext, err := aead.Open(nil, nonce, ciphertextWithTag, nil)
	if err != nil {
		// Attempt decrypting without padding if trailing bytes were zero-padded
		for end := len(ciphertextWithTag); end > aead.Overhead(); end-- {
			if pt, ptErr := aead.Open(nil, nonce, ciphertextWithTag[:end], nil); ptErr == nil {
				plaintext = pt
				err = nil
				break
			}
		}
		if err != nil {
			return nil, ErrUnwrapFailed
		}
	}

	// Parse decrypted routing instruction
	if len(plaintext) < 7 {
		return nil, ErrUnwrapFailed
	}

	isExit := plaintext[0] == 1
	epLen := binary.BigEndian.Uint16(plaintext[1:3])
	offset := 3 + int(epLen)
	if len(plaintext) < offset+4 {
		return nil, ErrUnwrapFailed
	}

	nextEp := string(plaintext[3:offset])
	pLen := binary.BigEndian.Uint32(plaintext[offset : offset+4])
	pOffset := offset + 4
	if len(plaintext) < pOffset+int(pLen) {
		return nil, ErrUnwrapFailed
	}

	nextPayload := plaintext[pOffset : pOffset+int(pLen)]

	return &LayerInstruction{
		IsExit:       isExit,
		NextEndpoint: nextEp,
		NextPayload:  nextPayload,
	}, nil
}
