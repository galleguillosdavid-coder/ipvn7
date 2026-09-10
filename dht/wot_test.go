package dht

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"path/filepath"
	"testing"
	"time"
)

func generateTestDID() (string, ed25519.PrivateKey) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	did := "did:ipv7:" + hex.EncodeToString(pub)
	return did, priv
}

func TestWebOfTrustSignatureAndVerification(t *testing.T) {
	didAlice, privAlice := generateTestDID()
	didBob, _ := generateTestDID()

	vouch := TrustVouch{
		IssuerDID:  didAlice,
		SubjectDID: didBob,
		Score:      0.95,
		Tag:        "reliable-relay",
		Comment:    "Ultra-low latency connection",
		Timestamp:  time.Now(),
	}

	// Sign vouch
	vouch.Sign(privAlice)

	// Verify valid signature
	if !vouch.Verify() {
		t.Fatalf("expected valid Ed25519 signature on vouch")
	}

	// Tamper with score -> verification must fail
	vouch.Score = 0.20
	if vouch.Verify() {
		t.Fatalf("tampered vouch should fail verification")
	}
}

func TestWebOfTrustTransitiveCalculation(t *testing.T) {
	didA, privA := generateTestDID()
	didB, privB := generateTestDID()
	didC, _ := generateTestDID()
	didD, _ := generateTestDID()

	wot := NewWebOfTrust(didA, "")

	// A -> B: 0.90
	vouchAB := TrustVouch{
		IssuerDID:  didA,
		SubjectDID: didB,
		Score:      0.90,
		Tag:        "friend",
		Timestamp:  time.Now(),
	}
	vouchAB.Sign(privA)
	_ = wot.AddVouch(vouchAB, true)

	// B -> C: 0.80
	vouchBC := TrustVouch{
		IssuerDID:  didB,
		SubjectDID: didC,
		Score:      0.80,
		Tag:        "workstation",
		Timestamp:  time.Now(),
	}
	vouchBC.Sign(privB)
	_ = wot.AddVouch(vouchBC, true)

	// 1-hop direct check: A -> B
	directAB := wot.CalculateTrust(didA, didB, 3)
	if directAB != 0.90 {
		t.Fatalf("expected direct trust of 0.90, got %f", directAB)
	}

	// 2-hop transitive check: A -> C (decay = 0.85 per hop)
	// Expected: (1.0 * 0.90 * 0.85) * 0.80 * 0.85 = 0.765 * 0.68 = 0.5202
	transitiveAC := wot.CalculateTrust(didA, didC, 3)
	if transitiveAC < 0.51 || transitiveAC > 0.53 {
		t.Fatalf("expected transitive trust ~0.5202, got %f", transitiveAC)
	}

	// Disconnected node D -> should have 0.0 trust
	trustAD := wot.CalculateTrust(didA, didD, 3)
	if trustAD != 0.0 {
		t.Fatalf("expected 0.0 trust for disconnected node D, got %f", trustAD)
	}
}

func TestWebOfTrustSelfVouchRejection(t *testing.T) {
	didA, privA := generateTestDID()
	wot := NewWebOfTrust(didA, "")

	selfVouch := TrustVouch{
		IssuerDID:  didA,
		SubjectDID: didA,
		Score:      1.0,
	}
	selfVouch.Sign(privA)

	err := wot.AddVouch(selfVouch, false)
	if err == nil {
		t.Fatalf("expected error when attempting to vouch for oneself")
	}
}

func TestWebOfTrustKuzuExportAndPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "wot.json")

	didA, privA := generateTestDID()
	didB, _ := generateTestDID()

	wot1 := NewWebOfTrust(didA, path)
	vouch := TrustVouch{
		IssuerDID:  didA,
		SubjectDID: didB,
		Score:      0.85,
		Tag:        "storage-provider",
		Timestamp:  time.Now(),
	}
	vouch.Sign(privA)
	_ = wot1.AddVouch(vouch, true)
	_ = wot1.SaveToFile(path)

	// Test Cypher statement generation
	stmts := wot1.ExportKuzuCypher()
	if len(stmts) != 1 {
		t.Fatalf("expected 1 Cypher statement, got %d", len(stmts))
	}

	// Load in wot2
	wot2 := NewWebOfTrust(didA, path)
	vouches := wot2.ListVouchesFrom(didA)
	if len(vouches) != 1 {
		t.Fatalf("expected 1 loaded vouch in wot2, got %d", len(vouches))
	}
}
