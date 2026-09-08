package core

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"testing"
	"time"
)

// BenchmarkContainerSign measures Ed25519 signing performance per container
func BenchmarkContainerSign(b *testing.B) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		b.Fatal(err)
	}

	payload := make([]byte, 512)
	c := &Container{
		SenderPubKey: priv.Public().(ed25519.PublicKey),
		SessionID:    "bench-session",
		Payload:      payload,
		HopLimit:     12,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := c.Sign(priv); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkContainerVerify measures Ed25519 verification performance per container
func BenchmarkContainerVerify(b *testing.B) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		b.Fatal(err)
	}

	payload := make([]byte, 512)
	c := &Container{
		SenderPubKey: priv.Public().(ed25519.PublicKey),
		SessionID:    "bench-session",
		Payload:      payload,
		HopLimit:     12,
	}
	if err := c.Sign(priv); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if !c.Verify() {
			b.Fatal("verification failed")
		}
	}
}

// BenchmarkE2EEEncryptDecrypt measures full roundtrip X25519 + ChaCha20-Poly1305 throughput
func BenchmarkE2EEEncryptDecrypt(b *testing.B) {
	seed := make([]byte, 32)
	_, _ = rand.Read(seed)
	privKey, pubKey, err := DeriveX25519FromSeed(seed)
	if err != nil {
		b.Fatal(err)
	}

	plaintext := []byte("Benchmark message payload for E2EE ChaCha20-Poly1305 encryption in IPv7 mesh node!")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		envelope, err := EncryptE2EE(pubKey.Bytes(), plaintext)
		if err != nil {
			b.Fatal(err)
		}
		decrypted, err := DecryptE2EE(privKey, envelope)
		if err != nil || len(decrypted) != len(plaintext) {
			b.Fatal("decryption failed or length mismatch")
		}
	}
}

// BenchmarkSmallWorldLookup measures Kleinberg XOR distance calculation and closest peer lookup
func BenchmarkSmallWorldLookup(b *testing.B) {
	localID, _, _ := GenerateIdentity()
	table := NewSmallWorldTable(localID, DefaultMaxDegrees, DefaultPeersPerDegree)

	// Populate table with 100 peers
	for i := 0; i < 100; i++ {
		id, _, _ := GenerateIdentity()
		table.AddPeer(id, []string{fmt.Sprintf("192.168.1.%d:7001", (i%200)+1)}, time.Duration(i)*time.Millisecond)
	}

	targetID, _, _ := GenerateIdentity()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		peers := table.FindClosestPeers(targetID, 5)
		if len(peers) == 0 {
			b.Fatal("no peers found")
		}
	}
}

// BenchmarkContainerSerialization measures canonical CBOR encoding/decoding performance
func BenchmarkContainerSerialization(b *testing.B) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	c := &Container{
		SenderPubKey:   priv.Public().(ed25519.PublicKey),
		ReceiverPubKey: priv.Public().(ed25519.PublicKey),
		SessionID:      "serialization-bench",
		Payload:        make([]byte, 1024),
		HopLimit:       DefaultHopLimit,
	}
	_ = c.Sign(priv)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		data, err := c.Marshal()
		if err != nil {
			b.Fatal(err)
		}
		var decoded Container
		if err := decoded.Unmarshal(data); err != nil {
			b.Fatal(err)
		}
	}
}
