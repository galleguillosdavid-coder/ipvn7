package core

import (
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"
)

// TestProductionBenchmarkSuite runs the comprehensive Cold vs Warm vs Bulk vs PQC matrix
func TestProductionBenchmarkSuite(t *testing.T) {
	const warmIterations = 50000
	const coldIterations = 1000
	payloadSize := 1024 // 1 KB payload

	sampleData := make([]byte, payloadSize)
	_, _ = rand.Read(sampleData)

	// =========================================================================
	// BENCHMARK A: COLD SESSION (Handshake + KeyGen + ECDH + Sign + Verify each op)
	// =========================================================================
	startCold := time.Now()
	for i := 0; i < coldIterations; i++ {
		idA, privA, _ := GenerateIdentity()
		_, privB, _ := GenerateIdentity()
		_ = privB

		seed := make([]byte, 32)
		_, _ = rand.Read(seed)
		privX, pubX, _ := DeriveX25519FromSeed(seed)

		env, err := EncryptE2EE(pubX.Bytes(), sampleData)
		if err != nil {
			t.Fatal(err)
		}
		c := &Container{
			SenderPubKey: idA.Bytes(),
			Payload:      env,
			HopLimit:     12,
		}
		if err := c.Sign(privA); err != nil {
			t.Fatal(err)
		}
		if !c.Verify() {
			t.Fatal("cold verify failed")
		}
		if _, err := DecryptE2EE(privX, c.Payload); err != nil {
			t.Fatal(err)
		}
	}
	durCold := time.Since(startCold)
	coldUsPerOp := float64(durCold.Microseconds()) / float64(coldIterations)
	coldMBps := (float64(payloadSize*coldIterations) / (1024 * 1024)) / durCold.Seconds()
	coldPPS := float64(coldIterations) / durCold.Seconds()

	// =========================================================================
	// BENCHMARK B: WARM SESSION (Established Noise_XX Session with AEAD + Replay)
	// =========================================================================
	// 1. Establish session ONCE via Noise_XX Handshake
	_, privEdA, _ := ed25519.GenerateKey(rand.Reader)
	privXA, _ := ecdh.X25519().GenerateKey(rand.Reader)
	_, privEdB, _ := ed25519.GenerateKey(rand.Reader)
	privXB, _ := ecdh.X25519().GenerateKey(rand.Reader)

	initHs, err := NewNoiseHandshake(true, privEdA, privXA)
	if err != nil {
		t.Fatal(err)
	}
	respHs, err := NewNoiseHandshake(false, privEdB, privXB)
	if err != nil {
		t.Fatal(err)
	}

	msg1, err := initHs.InitiatorStep1()
	if err != nil {
		t.Fatal(err)
	}
	msg2, err := respHs.ResponderStep2(msg1)
	if err != nil {
		t.Fatal(err)
	}
	msg3, sessA, err := initHs.InitiatorStep3(msg1, msg2)
	if err != nil {
		t.Fatal(err)
	}
	sessB, err := respHs.ResponderFinal(msg1, msg2, msg3)
	if err != nil {
		t.Fatal(err)
	}

	// 2. Transmit over warm session
	startWarm := time.Now()
	for i := 0; i < warmIterations; i++ {
		ct, err := sessA.Encrypt(sampleData)
		if err != nil {
			t.Fatal(err)
		}
		pt, err := sessB.Decrypt(ct)
		if err != nil || len(pt) != len(sampleData) {
			t.Fatal("warm decrypt failed")
		}
	}
	durWarm := time.Since(startWarm)
	warmUsPerOp := float64(durWarm.Microseconds()) / float64(warmIterations)
	warmMBps := (float64(payloadSize*warmIterations) / (1024 * 1024)) / durWarm.Seconds()
	warmPPS := float64(warmIterations) / durWarm.Seconds()

	// =========================================================================
	// BENCHMARK C: BULK STREAMING TRANSFER (64 KB Chunks over Warm Session)
	// =========================================================================
	const bulkChunkSize = 64 * 1024 // 64 KB
	const bulkIterations = 5000     // 320 MB total
	bulkData := make([]byte, bulkChunkSize)
	_, _ = rand.Read(bulkData)

	startBulk := time.Now()
	for i := 0; i < bulkIterations; i++ {
		ct, err := sessA.Encrypt(bulkData)
		if err != nil {
			t.Fatal(err)
		}
		pt, err := sessB.Decrypt(ct)
		if err != nil || len(pt) != bulkChunkSize {
			t.Fatal("bulk decrypt failed")
		}
	}
	durBulk := time.Since(startBulk)
	bulkTotalMB := float64(bulkChunkSize*bulkIterations) / (1024 * 1024)
	bulkMBps := bulkTotalMB / durBulk.Seconds()
	bulkGbps := (bulkMBps * 8) / 1000

	// =========================================================================
	// BENCHMARK E: CLASSICAL VS HYBRID PQC COST
	// =========================================================================
	const pqcIterations = 2000

	// Classical Ed25519
	startEdClassic := time.Now()
	for i := 0; i < pqcIterations; i++ {
		sig := ed25519.Sign(privEdA, sampleData)
		if !ed25519.Verify(privEdA.Public().(ed25519.PublicKey), sampleData, sig) {
			t.Fatal("classic verify failed")
		}
	}
	durEdClassic := time.Since(startEdClassic)
	usEdClassic := float64(durEdClassic.Microseconds()) / float64(pqcIterations)

	// Hybrid Ed25519 + ML-DSA
	pqcPriv, pqcPub, _ := GeneratePQCKeyStub()
	hybID := NewHybridIdentity(&Ed25519Identity{PublicKey: privEdA.Public().(ed25519.PublicKey)}, pqcPub, AlgHybridMLDSA)
	startHybSign := time.Now()
	for i := 0; i < pqcIterations; i++ {
		sig, err := SignHybrid(privEdA, pqcPriv, sampleData)
		if err != nil || !VerifyHybrid(hybID, sampleData, sig) {
			t.Fatal("hybrid verify failed")
		}
	}
	durHybSign := time.Since(startHybSign)
	usHybSign := float64(durHybSign.Microseconds()) / float64(pqcIterations)

	// Output Comprehensive Performance Matrix
	t.Logf("\n==========================================================================================")
	t.Logf("                IPV7 PERFORMANCE & PRODUCTION BENCHMARK MATRIX (2026)                    ")
	t.Logf("==========================================================================================")
	t.Logf(" Modo           | Latencia/Op | Throughput MB/s | Throughput Gbps | Paquetes/s (pps)       ")
	t.Logf("----------------+-------------+-----------------+-----------------+-----------------------")
	t.Logf(" Cold Session   | %7.2f µs  | %10.2f MB/s  | %10.4f Gbps | %10.0f pps", coldUsPerOp, coldMBps, (coldMBps*8)/1000, coldPPS)
	t.Logf(" Warm Session   | %7.2f µs  | %10.2f MB/s  | %10.4f Gbps | %10.0f pps", warmUsPerOp, warmMBps, (warmMBps*8)/1000, warmPPS)
	t.Logf(" Bulk (64KB)    | %7.2f µs  | %10.2f MB/s  | %10.4f Gbps | %10.0f pps", float64(durBulk.Microseconds())/float64(bulkIterations), bulkMBps, bulkGbps, float64(bulkIterations)/durBulk.Seconds())
	t.Logf("----------------+-------------+-----------------+-----------------+-----------------------")
	t.Logf(" Aceleración Warm vs Cold  : %.1fx mayor throughput (%.1fx menor latencia)", warmMBps/coldMBps, coldUsPerOp/warmUsPerOp)
	t.Logf("==========================================================================================")
	t.Logf(" CRYPTO OVERHEAD: CLASSICAL VS HYBRID PQC                                                 ")
	t.Logf("------------------------------------------------------------------------------------------")
	t.Logf(" Clásico Ed25519 (Sign + Verify)         : %6.2f µs/op", usEdClassic)
	t.Logf(" Híbrido Ed25519 + ML-DSA (Sign + Verify): %6.2f µs/op (Overhead adicional: %5.2f µs)", usHybSign, usHybSign-usEdClassic)
	t.Logf("==========================================================================================")
}
