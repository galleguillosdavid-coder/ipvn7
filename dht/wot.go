package dht

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// TrustVouch represents a cryptographically signed endorsement from one peer to another.
type TrustVouch struct {
	IssuerDID  string    `json:"issuer_did"`  // Signing peer
	SubjectDID string    `json:"subject_did"` // Endorsed peer
	Score      float64   `json:"score"`       // 0.0 (untrusted) to 1.0 (fully trusted)
	Tag        string    `json:"tag"`         // e.g. "reliable-relay", "stable-node", "friend"
	Comment    string    `json:"comment,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	Signature  string    `json:"signature"`   // Hex-encoded Ed25519 signature
}

// PayloadToSign produces the deterministic payload for signing a vouch.
func (v *TrustVouch) PayloadToSign() []byte {
	str := fmt.Sprintf("%s:%s:%.4f:%s:%d",
		v.IssuerDID, v.SubjectDID, v.Score, v.Tag, v.Timestamp.Unix())
	h := sha256.Sum256([]byte(str))
	return h[:]
}

// Sign signs the vouch statement with the issuer's private key.
func (v *TrustVouch) Sign(privKey ed25519.PrivateKey) {
	digest := v.PayloadToSign()
	sig := ed25519.Sign(privKey, digest)
	v.Signature = hex.EncodeToString(sig)
}

// Verify checks the cryptographic validity of the vouch signature.
func (v *TrustVouch) Verify() bool {
	if len(v.Signature) == 0 {
		return false
	}
	sigBytes, err := hex.DecodeString(v.Signature)
	if err != nil {
		return false
	}

	// Extract public key from IssuerDID ("did:ipv7:<hex32>" or raw hex)
	hexKey := strings.TrimPrefix(v.IssuerDID, "did:ipv7:")
	pubKeyBytes, err := hex.DecodeString(hexKey)
	if err != nil || len(pubKeyBytes) != ed25519.PublicKeySize {
		return false
	}

	digest := v.PayloadToSign()
	return ed25519.Verify(ed25519.PublicKey(pubKeyBytes), digest, sigBytes)
}

// WebOfTrust manages a decentralized graph of trust endorsements without any blockchain.
type WebOfTrust struct {
	mu          sync.RWMutex
	localDID    string
	vouches     map[string]map[string]TrustVouch // issuer -> subject -> vouch
	storagePath string
}

// NewWebOfTrust initializes a new Web of Trust engine.
func NewWebOfTrust(localDID string, storagePath string) *WebOfTrust {
	wot := &WebOfTrust{
		localDID:    localDID,
		vouches:     make(map[string]map[string]TrustVouch),
		storagePath: storagePath,
	}
	if storagePath != "" {
		_ = wot.LoadFromFile(storagePath)
	}
	return wot
}

// AddVouch registers a vouch statement. If requireSig is true, verifies the Ed25519 signature.
func (w *WebOfTrust) AddVouch(v TrustVouch, requireSig bool) error {
	if v.Score < 0.0 || v.Score > 1.0 {
		return fmt.Errorf("trust score must be between 0.0 and 1.0, got %f", v.Score)
	}
	if v.IssuerDID == v.SubjectDID {
		return fmt.Errorf("cannot vouch for oneself")
	}
	if v.Timestamp.IsZero() {
		v.Timestamp = time.Now()
	}

	if requireSig && !v.Verify() {
		return fmt.Errorf("invalid cryptographic signature on trust vouch from %s", v.IssuerDID)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.vouches[v.IssuerDID]; !exists {
		w.vouches[v.IssuerDID] = make(map[string]TrustVouch)
	}
	w.vouches[v.IssuerDID][v.SubjectDID] = v

	if w.storagePath != "" {
		_ = w.saveLocked()
	}
	return nil
}

// CalculateTrust computes transitive trust score from a starting root DID to target DID.
// Uses a decaying breadth-first traversal (decay factor 0.85 per hop, maxHops <= 3).
func (w *WebOfTrust) CalculateTrust(fromDID string, targetDID string, maxHops int) float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if fromDID == targetDID {
		return 1.0
	}
	if maxHops <= 0 {
		maxHops = 3
	}

	// 1-hop direct check
	if fromVouches, ok := w.vouches[fromDID]; ok {
		if direct, found := fromVouches[targetDID]; found {
			return direct.Score
		}
	}

	// BFS for transitive trust
	type queueItem struct {
		did   string
		score float64
		hop   int
	}

	visited := make(map[string]bool)
	visited[fromDID] = true
	queue := []queueItem{{did: fromDID, score: 1.0, hop: 0}}
	maxDiscoveredTrust := 0.0

	const decay = 0.85

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr.hop >= maxHops {
			continue
		}

		issuerVouches, ok := w.vouches[curr.did]
		if !ok {
			continue
		}

		for subject, vouch := range issuerVouches {
			transitiveScore := curr.score * vouch.Score * decay

			if subject == targetDID {
				if transitiveScore > maxDiscoveredTrust {
					maxDiscoveredTrust = transitiveScore
				}
				continue
			}

			if !visited[subject] && transitiveScore > 0.1 {
				visited[subject] = true
				queue = append(queue, queueItem{
					did:   subject,
					score: transitiveScore,
					hop:   curr.hop + 1,
				})
			}
		}
	}

	return maxDiscoveredTrust
}

// ListVouchesFrom returns all endorsements issued by a specific DID.
func (w *WebOfTrust) ListVouchesFrom(issuerDID string) []TrustVouch {
	w.mu.RLock()
	defer w.mu.RUnlock()

	out := make([]TrustVouch, 0)
	if m, ok := w.vouches[issuerDID]; ok {
		for _, v := range m {
			out = append(out, v)
		}
	}
	return out
}

// ListVouchesFor returns all endorsements received by a subject DID.
func (w *WebOfTrust) ListVouchesFor(subjectDID string) []TrustVouch {
	w.mu.RLock()
	defer w.mu.RUnlock()

	out := make([]TrustVouch, 0)
	for _, subjectMap := range w.vouches {
		if v, ok := subjectMap[subjectDID]; ok {
			out = append(out, v)
		}
	}
	return out
}

// ExportKuzuCypher generates Cypher graph statements to visualize the Web of Trust in Kùzu.
func (w *WebOfTrust) ExportKuzuCypher() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var statements []string
	for issuer, subjectMap := range w.vouches {
		for subject, v := range subjectMap {
			stmt := fmt.Sprintf("MATCH (p1:Peer {id: '%s'}), (p2:Peer {id: '%s'}) CREATE (p1)-[:VOUCHES {score: %.4f, tag: '%s'}]->(p2);",
				issuer, subject, v.Score, v.Tag)
			statements = append(statements, stmt)
		}
	}
	return statements
}

// SaveToFile exports the Web of Trust graph to JSON.
func (w *WebOfTrust) SaveToFile(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.storagePath = path
	return w.saveLocked()
}

func (w *WebOfTrust) saveLocked() error {
	flat := make([]TrustVouch, 0)
	for _, subjectMap := range w.vouches {
		for _, v := range subjectMap {
			flat = append(flat, v)
		}
	}

	data, err := json.MarshalIndent(flat, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(w.storagePath, data, 0600)
}

// LoadFromFile imports the Web of Trust graph from JSON.
func (w *WebOfTrust) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var flat []TrustVouch
	if err := json.Unmarshal(data, &flat); err != nil {
		return err
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	w.vouches = make(map[string]map[string]TrustVouch)
	for _, v := range flat {
		if _, ok := w.vouches[v.IssuerDID]; !ok {
			w.vouches[v.IssuerDID] = make(map[string]TrustVouch)
		}
		w.vouches[v.IssuerDID][v.SubjectDID] = v
	}
	w.storagePath = path
	return nil
}
