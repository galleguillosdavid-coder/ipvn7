package core

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"
)

// TestBenchmarkLayerBottleneckBreakdown executes a comparative analysis of every single layer
// in the IPv7 packet pipeline, providing the exact microsecond cost and % bottleneck breakdown.
func TestBenchmarkLayerBottleneckBreakdown(t *testing.T) {
	const iterations = 2000
	payloadSize := 1024 // 1 KB payload

	// 1. Prepare identities and keys
	idA, privA, _ := GenerateIdentity()
	idB, privB, _ := GenerateIdentity()
	_ = privA
	_ = privB

	seed := make([]byte, 32)
	_, _ = rand.Read(seed)
	privKeyX, pubKeyX, _ := DeriveX25519FromSeed(seed)
	_ = privKeyX

	samplePayload := make([]byte, payloadSize)
	_, _ = rand.Read(samplePayload)

	// Layer 1: CBOR Canonical Serialization & Deserialization
	startCBOR := time.Now()
	for i := 0; i < iterations; i++ {
		c := &Container{
			SenderPubKey: idA.Bytes(),
			SessionID:    "bench-session-123",
			Payload:      samplePayload,
			HopLimit:     12,
		}
		raw, err := c.Marshal()
		if err != nil {
			t.Fatal(err)
		}
		var parsed Container
		if err := parsed.Unmarshal(raw); err != nil {
			t.Fatal(err)
		}
	}
	durCBOR := time.Since(startCBOR)

	// Layer 2: Symmetric E2EE AEAD (ChaCha20-Poly1305 + sync.Pool)
	startAEAD := time.Now()
	for i := 0; i < iterations; i++ {
		env, err := EncryptE2EE(pubKeyX.Bytes(), samplePayload)
		if err != nil {
			t.Fatal(err)
		}
		_, err = DecryptE2EE(privKeyX, env)
		if err != nil {
			t.Fatal(err)
		}
	}
	durAEAD := time.Since(startAEAD)

	// Layer 3: Ed25519 Asymmetric Digital Signature (Sign + Verify)
	startEd25519 := time.Now()
	for i := 0; i < iterations; i++ {
		c := &Container{
			SenderPubKey: idA.Bytes(),
			SessionID:    "bench-session-123",
			Payload:      samplePayload,
			HopLimit:     12,
		}
		if err := c.Sign(privA); err != nil {
			t.Fatal(err)
		}
		if !c.Verify() {
			t.Fatal("verification failed")
		}
	}
	durEd25519 := time.Since(startEd25519)

	// Layer 4: Small-World Greedy XOR Routing Lookup (100 peers)
	table := NewSmallWorldTable(idA, DefaultMaxDegrees, DefaultPeersPerDegree)
	for i := 0; i < 100; i++ {
		peerID, _, _ := GenerateIdentity()
		table.AddPeer(peerID, []string{"192.168.1.1:7001"}, time.Millisecond)
	}
	startRouting := time.Now()
	for i := 0; i < iterations; i++ {
		_ = table.FindClosestPeers(idB, 5)
	}
	durRouting := time.Since(startRouting)

	totalTime := durCBOR + durAEAD + durEd25519 + durRouting

	costCBORUs := float64(durCBOR.Microseconds()) / float64(iterations)
	costAEADUs := float64(durAEAD.Microseconds()) / float64(iterations)
	costEd25519Us := float64(durEd25519.Microseconds()) / float64(iterations)
	costRoutingUs := float64(durRouting.Microseconds()) / float64(iterations)
	totalUs := float64(totalTime.Microseconds()) / float64(iterations)

	pctCBOR := (float64(durCBOR) / float64(totalTime)) * 100
	pctAEAD := (float64(durAEAD) / float64(totalTime)) * 100
	pctEd25519 := (float64(durEd25519) / float64(totalTime)) * 100
	pctRouting := (float64(durRouting) / float64(totalTime)) * 100

	// Throughput calculation: (bytes / second) -> MB/s
	mbProcessed := (float64(payloadSize*iterations) / (1024 * 1024)) / durAEAD.Seconds()
	ppsSustained := float64(iterations) / (float64(totalTime.Microseconds()) / 1e6)

	t.Logf("\n==========================================================================")
	t.Logf("     IPV7 PIPELINE BOTTLENECK & THROUGHPUT PROFILE (Payload: %d B)        ", payloadSize)
	t.Logf("==========================================================================")
	t.Logf(" 1. CBOR (Marshal + Unmarshal) : %6.2f µs/op (%5.1f%% del pipeline)", costCBORUs, pctCBOR)
	t.Logf(" 2. E2EE (ChaCha20-Poly1305)   : %6.2f µs/op (%5.1f%% del pipeline)", costAEADUs, pctAEAD)
	t.Logf(" 3. Firma (Ed25519 Sign+Verify): %6.2f µs/op (%5.1f%% del pipeline)", costEd25519Us, pctEd25519)
	t.Logf(" 4. Routing (Small-World XOR)  : %6.2f µs/op (%5.1f%% del pipeline)", costRoutingUs, pctRouting)
	t.Logf("--------------------------------------------------------------------------")
	t.Logf(" Latencia Total Pipeline (L1->L4): %6.2f µs por paquete", totalUs)
	t.Logf(" Rendimiento Cifrado E2EE        : %6.2f MB/s por core", mbProcessed)
	t.Logf(" Capacidad de Paquetes Teórica   : %6.0f paquetes/segundo por core", ppsSustained)
	t.Logf("==========================================================================")
}

// BenchmarkFullPipelineThroughputMBps measures continuous end-to-end memory throughput
func BenchmarkFullPipelineThroughputMBps(b *testing.B) {
	_, privA, _ := ed25519.GenerateKey(rand.Reader)
	seed := make([]byte, 32)
	_, _ = rand.Read(seed)
	privKeyX, pubKeyX, _ := DeriveX25519FromSeed(seed)

	payload := make([]byte, 1024)
	b.SetBytes(int64(len(payload)))
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Encrypt
		env, err := EncryptE2EE(pubKeyX.Bytes(), payload)
		if err != nil {
			b.Fatal(err)
		}
		// Container
		c := &Container{
			SenderPubKey: privA.Public().(ed25519.PublicKey),
			Payload:      env,
			HopLimit:     12,
		}
		raw, err := c.Marshal()
		if err != nil {
			b.Fatal(err)
		}
		// Decrypt
		var parsed Container
		_ = parsed.Unmarshal(raw)
		_, err = DecryptE2EE(privKeyX, parsed.Payload)
		if err != nil {
			b.Fatal(err)
		}
	}
}
