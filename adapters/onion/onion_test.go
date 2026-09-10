package onion

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"

	"golang.org/x/crypto/curve25519"
	"ipv7/core"
)

func generateX25519Keypair() ([32]byte, [32]byte) {
	var priv, pub [32]byte
	_, _ = io.ReadFull(rand.Reader, priv[:])
	curve25519.ScalarBaseMult(&pub, &priv)
	return priv, pub
}

func TestSphinxPacketBuildAndUnwrapSingleHop(t *testing.T) {
	destPriv, destPub := generateX25519Keypair()
	destID, _, _ := core.GenerateIdentity()

	circuit := []CircuitHop{
		{
			DID:       destID,
			EncPubKey: destPub[:],
			Endpoint:  "192.168.1.100:7001",
		},
	}

	payload := []byte("SECRET_SOVEREIGN_MESSAGE_DIRECT")
	packet, err := BuildOnionPacket(circuit, payload)
	if err != nil {
		t.Fatalf("Failed to build onion packet: %v", err)
	}

	if len(packet) != OnionPacketSize {
		t.Fatalf("Expected fixed packet size %d, got %d", OnionPacketSize, len(packet))
	}

	// Exit hop unwraps
	instr, err := UnwrapLayer(packet, destPriv)
	if err != nil {
		t.Fatalf("Failed to unwrap layer: %v", err)
	}

	if !instr.IsExit {
		t.Fatalf("Expected exit hop instruction")
	}

	if !bytes.Equal(instr.NextPayload, payload) {
		t.Fatalf("Payload mismatch: expected %s, got %s", payload, instr.NextPayload)
	}
}

func TestSphinxPacketBuildAndUnwrapThreeHops(t *testing.T) {
	privB, pubB := generateX25519Keypair()
	idB, _, _ := core.GenerateIdentity()

	privC, pubC := generateX25519Keypair()
	idC, _, _ := core.GenerateIdentity()

	privD, pubD := generateX25519Keypair()
	idD, _, _ := core.GenerateIdentity()

	circuit := []CircuitHop{
		{
			DID:       idB,
			EncPubKey: pubB[:],
			Endpoint:  "node-b:7002",
		},
		{
			DID:       idC,
			EncPubKey: pubC[:],
			Endpoint:  "node-c:7003",
		},
		{
			DID:       idD,
			EncPubKey: pubD[:],
			Endpoint:  "node-d:7004",
		},
	}

	secretMessage := []byte("CONFIDENTIAL_GLOBAL_INTERNET_DATAGRAM")
	packet, err := BuildOnionPacket(circuit, secretMessage)
	if err != nil {
		t.Fatalf("BuildOnionPacket failed: %v", err)
	}

	// 1. Hop B receives and unwraps Layer 1
	instrB, err := UnwrapLayer(packet, privB)
	if err != nil {
		t.Fatalf("Hop B unwrap failed: %v", err)
	}
	if instrB.IsExit {
		t.Fatalf("Hop B should NOT be exit hop")
	}
	if instrB.NextEndpoint != "node-c:7003" {
		t.Fatalf("Hop B next endpoint mismatch: expected node-c:7003, got %s", instrB.NextEndpoint)
	}

	// 2. Hop C receives next payload and unwraps Layer 2
	instrC, err := UnwrapLayer(instrB.NextPayload, privC)
	if err != nil {
		t.Fatalf("Hop C unwrap failed: %v", err)
	}
	if instrC.IsExit {
		t.Fatalf("Hop C should NOT be exit hop")
	}
	if instrC.NextEndpoint != "node-d:7004" {
		t.Fatalf("Hop C next endpoint mismatch: expected node-d:7004, got %s", instrC.NextEndpoint)
	}

	// 3. Hop D (Exit) receives and unwraps final Layer
	instrD, err := UnwrapLayer(instrC.NextPayload, privD)
	if err != nil {
		t.Fatalf("Hop D unwrap failed: %v", err)
	}
	if !instrD.IsExit {
		t.Fatalf("Hop D MUST be exit hop")
	}
	if !bytes.Equal(instrD.NextPayload, secretMessage) {
		t.Fatalf("Final message corrupted: expected %s, got %s", secretMessage, instrD.NextPayload)
	}

	t.Logf("Success: 3-hop onion circuit successfully traversed (A -> B -> C -> D) without metadata leakage!")
}

func TestSphinxFixedPacketSizePadding(t *testing.T) {
	_, pubD := generateX25519Keypair()
	idD, _, _ := core.GenerateIdentity()

	// 1 hop circuit
	c1 := []CircuitHop{{DID: idD, EncPubKey: pubD[:], Endpoint: "ep1"}}
	pkt1, err := BuildOnionPacket(c1, []byte("short"))
	if err != nil || len(pkt1) != OnionPacketSize {
		t.Fatalf("1-hop size mismatch: %d", len(pkt1))
	}

	// 3 hop circuit
	_, pubB := generateX25519Keypair()
	idB, _, _ := core.GenerateIdentity()
	_, pubC := generateX25519Keypair()
	idC, _, _ := core.GenerateIdentity()
	c3 := []CircuitHop{
		{DID: idB, EncPubKey: pubB[:], Endpoint: "epB"},
		{DID: idC, EncPubKey: pubC[:], Endpoint: "epC"},
		{DID: idD, EncPubKey: pubD[:], Endpoint: "epD"},
	}
	pkt3, err := BuildOnionPacket(c3, []byte("longer payload with variable size"))
	if err != nil || len(pkt3) != OnionPacketSize {
		t.Fatalf("3-hop size mismatch: %d", len(pkt3))
	}

	if len(pkt1) != len(pkt3) {
		t.Fatalf("Onion packets must have identical sizes regardless of hop count: %d vs %d", len(pkt1), len(pkt3))
	}
}

func BenchmarkBuildOnion3Hops(b *testing.B) {
	_, pubB := generateX25519Keypair()
	idB, _, _ := core.GenerateIdentity()
	_, pubC := generateX25519Keypair()
	idC, _, _ := core.GenerateIdentity()
	_, pubD := generateX25519Keypair()
	idD, _, _ := core.GenerateIdentity()

	circuit := []CircuitHop{
		{DID: idB, EncPubKey: pubB[:], Endpoint: "epB"},
		{DID: idC, EncPubKey: pubC[:], Endpoint: "epC"},
		{DID: idD, EncPubKey: pubD[:], Endpoint: "epD"},
	}

	payload := []byte("BENCHMARK_PAYLOAD_DATA")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = BuildOnionPacket(circuit, payload)
	}
}

func BenchmarkUnwrapLayer(b *testing.B) {
	privD, pubD := generateX25519Keypair()
	idD, _, _ := core.GenerateIdentity()

	circuit := []CircuitHop{
		{DID: idD, EncPubKey: pubD[:], Endpoint: "epD"},
	}

	payload := []byte("BENCHMARK_PAYLOAD_DATA")
	pkt, _ := BuildOnionPacket(circuit, payload)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = UnwrapLayer(pkt, privD)
	}
}
