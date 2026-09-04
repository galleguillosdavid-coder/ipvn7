package dht

import (
	"bytes"
	"testing"
	"time"

	"ipv7/core"
)

func TestRecordSigningAndVerification(t *testing.T) {
	id, priv, _ := core.GenerateIdentity()

	rec := NewRecord(id.PublicKey, []string{"192.168.1.100:8000", "200.50.1.2:8000"}, time.Hour)
	err := rec.Sign(priv)
	if err != nil {
		t.Fatalf("Failed to sign DHT record: %v", err)
	}

	if !rec.Verify() {
		t.Fatalf("Valid DHT record failed verification")
	}

	// Tampered record
	rec.Endpoints = []string{"evil-hacker.com:9999"}
	if rec.Verify() {
		t.Fatalf("Tampered DHT record succeeded verification unexpectedly")
	}
}

func TestDHTPublishAndResolve(t *testing.T) {
	// Node A (Announcer)
	idA, privA, _ := core.GenerateIdentity()
	dhtA := NewDHTService(idA, privA)

	// Node B (Bootstrap / Relay DHT node)
	idB, privB, _ := core.GenerateIdentity()
	dhtB := NewDHTService(idB, privB)

	// Node C (Searcher / Resolver)
	idC, privC, _ := core.GenerateIdentity()
	dhtC := NewDHTService(idC, privC)

	// Network simulation between services
	rpcRouter := func(targetEp string, req *Message) (*Message, error) {
		switch targetEp {
		case "ep-b":
			return dhtB.ProcessMessage(req), nil
		case "ep-a":
			return dhtA.ProcessMessage(req), nil
		default:
			return nil, nil
		}
	}

	dhtA.SetRemoteCaller(rpcRouter)
	dhtC.SetRemoteCaller(rpcRouter)

	// A and C know B as a bootstrap peer
	dhtA.AddPeer(idB, []string{"ep-b"})
	dhtC.AddPeer(idB, []string{"ep-b"})

	// 1. Node A publishes its endpoints to DHT
	realEndpoints := []string{"192.168.1.100:7001", "201.50.2.1:7001"}
	_, err := dhtA.Publish(realEndpoints, time.Hour)
	if err != nil {
		t.Fatalf("Node A failed to publish record: %v", err)
	}

	// Wait a moment for async publish
	time.Sleep(50 * time.Millisecond)

	// 2. Node C resolves Node A's public key by asking Bootstrap Node B
	resolvedRec, err := dhtC.Resolve(idA.PublicKey)
	if err != nil {
		t.Fatalf("Node C failed to resolve Node A: %v", err)
	}

	if !bytes.Equal(resolvedRec.PublicKey, idA.PublicKey) {
		t.Fatalf("Resolved public key mismatch")
	}

	if len(resolvedRec.Endpoints) != 2 || resolvedRec.Endpoints[0] != "192.168.1.100:7001" {
		t.Fatalf("Resolved endpoints mismatch: %v", resolvedRec.Endpoints)
	}

	t.Logf("Success! Node C resolved Node A's endpoints via DHT: %v", resolvedRec.Endpoints)
}

func TestDHTRejectsForgedRecord(t *testing.T) {
	victimID, _, _ := core.GenerateIdentity()
	_, attackerPriv, _ := core.GenerateIdentity()

	// Attacker crafts record pretending to be victim, but signs with attacker's key
	forged := NewRecord(victimID.PublicKey, []string{"attacker-ip:666"}, time.Hour)
	_ = forged.Sign(attackerPriv) // Signature won't match victim's public key

	// Node Bootstrap
	idB, privB, _ := core.GenerateIdentity()
	dhtB := NewDHTService(idB, privB)

	// Try storing forged record
	req := &Message{
		Type:   MsgStore,
		Record: forged,
	}

	resp := dhtB.ProcessMessage(req)
	if resp != nil {
		t.Fatalf("DHT should have rejected forged record, but returned response: %v", resp)
	}

	// Verify nothing was stored
	_, err := dhtB.Resolve(victimID.PublicKey)
	if err == nil {
		t.Fatalf("DHT stored forged record, expected not found error")
	}
}
