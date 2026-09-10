package adapters

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// PairingPayload represents an out-of-band enrollment offer or profile.
type PairingPayload struct {
	Version      int      `json:"v"`
	DeviceName   string   `json:"name"`
	DID          string   `json:"did"`
	PublicKeyHex string   `json:"pk"`
	Endpoints    []string `json:"endpoints"`
	VirtualIPv4  string   `json:"vip4,omitempty"`
	VirtualIPv6  string   `json:"vip6,omitempty"`
	Nonce        string   `json:"nonce"`
	IssuedAt     int64    `json:"iat"`
	ExpiresAt    int64    `json:"exp"`
	Signature    string   `json:"sig,omitempty"`
}

// CanonicalBytes returns the deterministic byte slice to be signed or verified.
func (p *PairingPayload) CanonicalBytes() []byte {
	// Recreate deterministic string omitting the signature
	return []byte(fmt.Sprintf("%d:%s:%s:%s:%s:%s:%s:%s:%d:%d",
		p.Version,
		p.DeviceName,
		p.DID,
		p.PublicKeyHex,
		strings.Join(p.Endpoints, ","),
		p.VirtualIPv4,
		p.VirtualIPv6,
		p.Nonce,
		p.IssuedAt,
		p.ExpiresAt,
	))
}

// Sign signs the pairing payload with the sovereign Ed25519 private key.
func (p *PairingPayload) Sign(privKey ed25519.PrivateKey) error {
	if len(privKey) != ed25519.PrivateKeySize {
		return errors.New("invalid private key size for Ed25519")
	}
	canonical := p.CanonicalBytes()
	sig := ed25519.Sign(privKey, canonical)
	p.Signature = hex.EncodeToString(sig)
	return nil
}

// Verify validates that the signature matches the public key and that the payload has not expired.
func (p *PairingPayload) Verify() (bool, error) {
	if p.ExpiresAt > 0 && time.Now().Unix() > p.ExpiresAt {
		return false, errors.New("pairing payload has expired")
	}

	sigBytes, err := hex.DecodeString(p.Signature)
	if err != nil {
		return false, fmt.Errorf("invalid signature hex: %w", err)
	}

	cleanHex := strings.TrimPrefix(p.PublicKeyHex, "did:ipv7:")
	pubBytes, err := hex.DecodeString(cleanHex)
	if err != nil {
		return false, fmt.Errorf("invalid public key hex: %w", err)
	}

	if len(pubBytes) != ed25519.PublicKeySize {
		return false, errors.New("public key is not 32 bytes")
	}

	canonical := p.CanonicalBytes()
	valid := ed25519.Verify(pubBytes, canonical, sigBytes)
	return valid, nil
}

// NewPairingPayload constructs and signs a new pairing payload.
func NewPairingPayload(deviceName string, privKey ed25519.PrivateKey, endpoints []string, vip4, vip6 string, ttl time.Duration) (*PairingPayload, error) {
	pubKey := privKey.Public().(ed25519.PublicKey)
	pkHex := hex.EncodeToString(pubKey)
	did := fmt.Sprintf("did:ipv7:%s", pkHex)

	nonceBytes := make([]byte, 8)
	if _, err := rand.Read(nonceBytes); err != nil {
		return nil, err
	}

	now := time.Now()
	payload := &PairingPayload{
		Version:      1,
		DeviceName:   deviceName,
		DID:          did,
		PublicKeyHex: pkHex,
		Endpoints:    endpoints,
		VirtualIPv4:  vip4,
		VirtualIPv6:  vip6,
		Nonce:        hex.EncodeToString(nonceBytes),
		IssuedAt:     now.Unix(),
		ExpiresAt:    now.Add(ttl).Unix(),
	}

	if err := payload.Sign(privKey); err != nil {
		return nil, err
	}

	return payload, nil
}

// SASResult contains the 6-digit numeric and human-readable emoji verification representations.
type SASResult struct {
	Numeric string   // e.g. "724-819"
	Emojis  []string // e.g. ["🛡️", "🚀", "⚡", "🦊"]
}

var sasEmojiTable = []string{
	"🛡️", "🚀", "⚡", "🦊", "🌊", "🔑", "🦅", "💎",
	"🌲", "🪐", "🔥", "🔮", "🛸", " Anchor", "🎯", "🌟",
}

// ComputeSAS derives a deterministic 6-digit Short Authentication String and emoji sequence.
// Both peers will see identical strings if and only if there is no Man-in-the-Middle (MitM) attacker.
func ComputeSAS(alicePubKey, bobPubKey []byte, salt []byte) SASResult {
	hasher := sha256.New()
	hasher.Write(salt)

	// Lexicographically order keys so both Alice and Bob generate identical SAS regardless of who initiated
	if strings.Compare(hex.EncodeToString(alicePubKey), hex.EncodeToString(bobPubKey)) < 0 {
		hasher.Write(alicePubKey)
		hasher.Write(bobPubKey)
	} else {
		hasher.Write(bobPubKey)
		hasher.Write(alicePubKey)
	}

	digest := hasher.Sum(nil)

	// Derive 6-digit number (100000..999999)
	numVal := binary.BigEndian.Uint32(digest[:4]) % 900000 + 100000
	numeric := fmt.Sprintf("%03d-%03d", numVal/1000, numVal%1000)

	// Derive 4 emojis from next 4 bytes
	emojis := make([]string, 4)
	for i := 0; i < 4; i++ {
		idx := int(digest[4+i]) % len(sasEmojiTable)
		emojis[i] = sasEmojiTable[idx]
	}

	return SASResult{
		Numeric: numeric,
		Emojis:  emojis,
	}
}

// EncodeMultiFrame encodes a pairing payload into Uniform Resource (UR) multi-frame animated strings.
// Format: "ur:ipv7-pair/<seq>-<total>/<base64_url_chunk>"
func EncodeMultiFrame(payload *PairingPayload, maxChunkSize int) ([]string, error) {
	if maxChunkSize <= 0 {
		maxChunkSize = 80 // Ideal for high-speed camera scanning
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	encoded := base64.RawURLEncoding.EncodeToString(data)
	totalChunks := (len(encoded) + maxChunkSize - 1) / maxChunkSize
	if totalChunks == 0 {
		totalChunks = 1
	}

	frames := make([]string, totalChunks)
	for i := 0; i < totalChunks; i++ {
		start := i * maxChunkSize
		end := start + maxChunkSize
		if end > len(encoded) {
			end = len(encoded)
		}
		frames[i] = fmt.Sprintf("ur:ipv7-pair/%d-%d/%s", i+1, totalChunks, encoded[start:end])
	}

	return frames, nil
}

// URReassembler statefully collects multi-frame QR streams and reassembles the payload.
type URReassembler struct {
	mu          sync.Mutex
	totalFrames int
	received    map[int]string
}

// NewURReassembler initializes a new multi-frame reassembler.
func NewURReassembler() *URReassembler {
	return &URReassembler{
		received: make(map[int]string),
	}
}

// Feed ingests a single UR frame. Returns (*PairingPayload, complete=true, nil) when all frames are collected.
func (r *URReassembler) Feed(frame string) (*PairingPayload, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	frame = strings.TrimSpace(frame)
	if !strings.HasPrefix(frame, "ur:ipv7-pair/") {
		// If it's a raw JSON payload directly, parse it immediately
		var rawPayload PairingPayload
		if err := json.Unmarshal([]byte(frame), &rawPayload); err == nil {
			if ok, vErr := rawPayload.Verify(); ok {
				return &rawPayload, true, nil
			} else {
				return nil, false, fmt.Errorf("signature verification failed: %w", vErr)
			}
		}
		return nil, false, errors.New("invalid UR prefix, expected 'ur:ipv7-pair/'")
	}

	parts := strings.SplitN(strings.TrimPrefix(frame, "ur:ipv7-pair/"), "/", 2)
	if len(parts) != 2 {
		return nil, false, errors.New("malformed UR frame syntax")
	}

	seqParts := strings.Split(parts[0], "-")
	if len(seqParts) != 2 {
		return nil, false, errors.New("malformed frame sequence notation")
	}

	seq, err := strconv.Atoi(seqParts[0])
	if err != nil {
		return nil, false, err
	}
	total, err := strconv.Atoi(seqParts[1])
	if err != nil {
		return nil, false, err
	}

	if r.totalFrames == 0 {
		r.totalFrames = total
	} else if r.totalFrames != total {
		return nil, false, errors.New("mismatched total frame count in UR stream")
	}

	r.received[seq] = parts[1]

	if len(r.received) == r.totalFrames {
		// Reassemble in sequence
		var sb strings.Builder
		for i := 1; i <= r.totalFrames; i++ {
			chunk, ok := r.received[i]
			if !ok {
				return nil, false, fmt.Errorf("missing chunk %d", i)
			}
			sb.WriteString(chunk)
		}

		rawBytes, err := base64.RawURLEncoding.DecodeString(sb.String())
		if err != nil {
			return nil, false, fmt.Errorf("failed to decode base64 UR payload: %w", err)
		}

		var payload PairingPayload
		if err := json.Unmarshal(rawBytes, &payload); err != nil {
			return nil, false, fmt.Errorf("failed to parse JSON payload: %w", err)
		}

		valid, err := payload.Verify()
		if err != nil || !valid {
			return nil, false, fmt.Errorf("failed to verify reassembled payload signature: %v", err)
		}

		return &payload, true, nil
	}

	return nil, false, nil
}
