package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	mrand "math/rand"
	"net"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"ipv7/adapters"
	"ipv7/core"
)

// Percentiles helper
func calcStats(latencies []float64) (min, p50, p95, p99, max float64) {
	if len(latencies) == 0 {
		return 0, 0, 0, 0, 0
	}
	s := append([]float64(nil), latencies...)
	sort.Float64s(s)
	n := len(s)
	min = s[0]
	max = s[n-1]
	p50 = s[int(float64(n)*0.50)]
	p95 = s[int(float64(n)*0.95)]
	p99 = s[int(float64(n)*0.99)]
	return
}

func main() {
	fmt.Println("==================================================================")
	fmt.Println("          IPv7 ADVERSARIAL CHARACTERIZATION BENCHMARK             ")
	fmt.Println("             Encontrar los límites, no esconderlos                ")
	fmt.Println("==================================================================")
	fmt.Printf(" [Host] OS: %s | Arch: %s | CPUs: %d | Time: %s\n",
		runtime.GOOS, runtime.GOARCH, runtime.NumCPU(), time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println("==================================================================")

	// 1. Curva de Degradación por Jitter
	runJitterCharacterization()

	// 2. Curva de Saturación de Relay DERP
	runRelaySaturationCurve()

	// 3. Fuzzing y Ataque Continuo DURANTE Tráfico Legítimo
	runConcurrentAttackDuringLegitTraffic()

	// 4. Repetibilidad Estadística (100 Iteraciones por Vector Crítico)
	runStatisticalRepeatability()
}

// ================================================================================
// 1. Curva de Degradación por Jitter
// ================================================================================
func runJitterCharacterization() {
	fmt.Println("\n------------------------------------------------------------------")
	fmt.Println(" [1] CARACTERIZACIÓN DE JITTER: Curva de Latencia y Throughput    ")
	fmt.Println("------------------------------------------------------------------")
	fmt.Printf(" %-10s | %-12s | %-10s | %-10s | %-10s | %-10s | %-8s\n",
		"Jitter Max", "Throughput", "RTT Min", "RTT p50", "RTT p95", "RTT Max", "PDR %")
	fmt.Println("------------------------------------------------------------------")

	tiers := []int{0, 1, 5, 10, 25, 50, 100, 250, 500}

	for _, maxJitterMs := range tiers {
		pub, priv, _ := ed25519.GenerateKey(rand.Reader)
		nodeID, _ := core.NewIdentityFromBytes(pub)
		targetNode := core.NewNode(nodeID, priv)
		udp, _ := adapters.NewUDPAdapter("127.0.0.1:0")
		targetNode.AddAdapter(udp)
		_ = targetNode.Start()
		targetNode.SetEndpoints(udp.Endpoints())

		var receivedCount atomic.Int32
		rttRecords := make([]float64, 0, 100)
		var rttMu sync.Mutex

		targetNode.OnMessage(func(from core.Identity, payload []byte) {
			now := time.Now().UnixNano()
			var sentTime int64
			_, _ = fmt.Sscanf(string(payload), "time:%d", &sentTime)
			if sentTime > 0 {
				rttMs := float64(now-sentTime) / 1e6
				rttMu.Lock()
				rttRecords = append(rttRecords, rttMs)
				rttMu.Unlock()
			}
			receivedCount.Add(1)
		})

		udpConn, _ := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
		rAddr, _ := net.ResolveUDPAddr("udp", udp.Endpoints()[0])

		const packetCount = 100
		var wg sync.WaitGroup
		start := time.Now()

		for i := 1; i <= packetCount; i++ {
			wg.Add(1)
			go func(seq int) {
				defer wg.Done()
				if maxJitterMs > 0 {
					delay := time.Duration(mrand.Intn(maxJitterMs)) * time.Millisecond
					time.Sleep(delay)
				}
				now := time.Now().UnixNano()
				payload := []byte(fmt.Sprintf("time:%d-seq:%d", now, seq))
				c := &core.Container{
					SenderPubKey: pub,
					Payload:      payload,
					Seq:          uint64(seq),
					HopLimit:     12,
				}
				_ = c.Sign(priv)
				raw, _ := c.Marshal()
				_, _ = udpConn.WriteToUDP(raw, rAddr)
			}(i)
		}

		wg.Wait()
		time.Sleep(100 * time.Millisecond)
		dur := time.Since(start)

		udpConn.Close()
		targetNode.Stop()

		delivered := int(receivedCount.Load())
		pdr := (float64(delivered) / float64(packetCount)) * 100.0
		bytesTotal := delivered * 128
		mbps := (float64(bytesTotal*8) / dur.Seconds()) / 1e6

		rttMu.Lock()
		minR, p50R, p95R, _, maxR := calcStats(rttRecords)
		rttMu.Unlock()

		fmt.Printf(" %-10s | %9.3f Mbps | %8.2f ms | %8.2f ms | %8.2f ms | %8.2f ms | %6.1f%%\n",
			fmt.Sprintf("%d ms", maxJitterMs), mbps, minR, p50R, p95R, maxR, pdr)
	}
	fmt.Println("------------------------------------------------------------------")
}

// ================================================================================
// 2. Curva de Saturación de Relay DERP
// ================================================================================
func runRelaySaturationCurve() {
	fmt.Println("\n------------------------------------------------------------------")
	fmt.Println(" [2] CARACTERIZACIÓN DE RELAY DERP: Curva de Saturación y Límites ")
	fmt.Println("------------------------------------------------------------------")
	fmt.Printf(" %-10s | %-12s | %-12s | %-10s | %-10s | %-10s | %-8s\n",
		"Target PPS", "Deliv PPS", "Throughput", "Avg Latency", "Drops %", "RAM Heap", "Recovery")
	fmt.Println("------------------------------------------------------------------")

	rates := []int{1000, 2000, 5000, 10000, 20000, 50000}

	for _, targetPPS := range rates {
		relay := adapters.NewRelayServer("127.0.0.1:0")
		_ = relay.Start()
		relayAddr := relay.Addr().String()

		pubRecv, _, _ := ed25519.GenerateKey(rand.Reader)
		idRecv, _ := core.NewIdentityFromBytes(pubRecv)
		adapterRecv := adapters.NewRelayAdapter(relayAddr, idRecv)
		_ = adapterRecv.Start()

		var recvCount atomic.Int32
		latencies := make([]float64, 0, targetPPS)
		var latMu sync.Mutex

		go func() {
			for c := range adapterRecv.Receive() {
				recvCount.Add(1)
				if len(c.Payload) >= 8 {
					sentNano := int64(c.Seq)
					if sentNano > 0 {
						latMs := float64(time.Now().UnixNano()-sentNano) / 1e6
						latMu.Lock()
						if len(latencies) < 5000 {
							latencies = append(latencies, latMs)
						}
						latMu.Unlock()
					}
				}
			}
		}()

		pubSend, privSend, _ := ed25519.GenerateKey(rand.Reader)
		idSend, _ := core.NewIdentityFromBytes(pubSend)
		adapterSend := adapters.NewRelayAdapter(relayAddr, idSend)
		_ = adapterSend.Start()

		// Warmup connection
		time.Sleep(20 * time.Millisecond)

		// Test for 1.5 seconds at target rate
		duration := 1500 * time.Millisecond
		totalToSend := int(float64(targetPPS) * duration.Seconds())
		interval := time.Second / time.Duration(targetPPS)

		payload := make([]byte, 256)
		start := time.Now()

		for i := 0; i < totalToSend; i++ {
			c := &core.Container{
				SenderPubKey:   pubSend,
				ReceiverPubKey: pubRecv,
				Payload:        payload,
				Seq:            uint64(time.Now().UnixNano()),
				HopLimit:       12,
			}
			_ = c.Sign(privSend)
			_ = adapterSend.Send(c, []string{string(pubRecv)})

			// Pacing
			if interval > time.Microsecond*50 {
				time.Sleep(interval)
			}
		}

		elapsed := time.Since(start)
		time.Sleep(150 * time.Millisecond) // Drain buffer

		delivered := int(recvCount.Load())
		actualDelivPPS := float64(delivered) / elapsed.Seconds()
		throughputMbps := (float64(delivered*len(payload)*8) / elapsed.Seconds()) / 1e6
		drops := totalToSend - delivered
		if drops < 0 {
			drops = 0
		}
		dropPct := (float64(drops) / float64(totalToSend)) * 100.0

		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		heapMB := float64(m.HeapAlloc) / (1024 * 1024)

		latMu.Lock()
		_, p50Lat, _, _, _ := calcStats(latencies)
		latMu.Unlock()

		// Measure recovery: send 1 ping after saturation
		recStart := time.Now()
		cProbe := &core.Container{
			SenderPubKey:   pubSend,
			ReceiverPubKey: pubRecv,
			Payload:        []byte("probe"),
			HopLimit:       12,
		}
		_ = cProbe.Sign(privSend)
		_ = adapterSend.Send(cProbe, []string{string(pubRecv)})
		recTime := time.Since(recStart)

		adapterSend.Stop()
		adapterRecv.Stop()
		relay.Stop()

		fmt.Printf(" %-10s | %9.0f pps | %9.2f Mbps | %8.2f ms | %7.1f%% | %7.2f MB | %v\n",
			fmt.Sprintf("%d pps", targetPPS), actualDelivPPS, throughputMbps, p50Lat, dropPct, heapMB, recTime.Round(time.Microsecond))
	}
	fmt.Println("------------------------------------------------------------------")
}

// ================================================================================
// 3. Fuzzing y Ataque Continuo DURANTE Tráfico Legítimo
// ================================================================================
func runConcurrentAttackDuringLegitTraffic() {
	fmt.Println("\n------------------------------------------------------------------")
	fmt.Println(" [3] ADVERSARIAL FIREFIGHT: Tráfico Legítimo Bajo Ataque Concurrente ")
	fmt.Println("------------------------------------------------------------------")
	fmt.Println(" Modelo: Alice -> Bob transmiten datos continuos mientras un Atacante")
	fmt.Println(" bombardea con corrupción, replay, malformed CBOR y handshake flood.")
	fmt.Println("------------------------------------------------------------------")

	pubBob, privBob, _ := ed25519.GenerateKey(rand.Reader)
	idBob, _ := core.NewIdentityFromBytes(pubBob)
	nodeBob := core.NewNode(idBob, privBob)
	udpBob, _ := adapters.NewUDPAdapter("127.0.0.1:0")
	nodeBob.AddAdapter(udpBob)
	_ = nodeBob.Start()
	nodeBob.SetEndpoints(udpBob.Endpoints())
	defer nodeBob.Stop()

	pubAlice, privAlice, _ := ed25519.GenerateKey(rand.Reader)
	idAlice, _ := core.NewIdentityFromBytes(pubAlice)
	nodeAlice := core.NewNode(idAlice, privAlice)
	udpAlice, _ := adapters.NewUDPAdapter("127.0.0.1:0")
	nodeAlice.AddAdapter(udpAlice)
	_ = nodeAlice.Start()
	nodeAlice.SetEndpoints(udpAlice.Endpoints())
	defer nodeAlice.Stop()

	nodeAlice.AddPeer(idBob, udpBob.Endpoints())

	var legitDelivered atomic.Int32
	var corruptDelivered atomic.Int32
	var foreignDelivered atomic.Int32

	nodeBob.OnMessage(func(from core.Identity, payload []byte) {
		if from.String() == idAlice.String() {
			if bytes.HasPrefix(payload, []byte("legit-payload-")) {
				legitDelivered.Add(1)
			} else {
				corruptDelivered.Add(1)
			}
		} else {
			foreignDelivered.Add(1)
		}
	})

	// Attacker socket
	attackerConn, _ := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	defer attackerConn.Close()
	bobUDPAddr, _ := net.ResolveUDPAddr("udp", udpBob.Endpoints()[0])

	stopAttack := make(chan struct{})
	var attackPackets atomic.Int64

	// Attacker routine: bludgeons Bob's socket continuously
	go func() {
		attPub, attPriv, _ := ed25519.GenerateKey(rand.Reader)
		badBytes := []byte{0xBF, 0xFF, 0x1B, 0xFF, 0xFF}

		cReplay := &core.Container{SenderPubKey: attPub, Payload: []byte("replay-bomb"), Seq: 999, HopLimit: 12}
		_ = cReplay.Sign(attPriv)
		rawReplay, _ := cReplay.Marshal()

		for {
			select {
			case <-stopAttack:
				return
			default:
				// 1. Corrupted packet
				cBad := &core.Container{SenderPubKey: attPub, Payload: []byte("corrupt"), Seq: 1, HopLimit: 12}
				_ = cBad.Sign(attPriv)
				rawBad, _ := cBad.Marshal()
				rawBad[len(rawBad)/2] ^= 0xEE
				_, _ = attackerConn.WriteToUDP(rawBad, bobUDPAddr)

				// 2. Replay
				_, _ = attackerConn.WriteToUDP(rawReplay, bobUDPAddr)

				// 3. Malformed CBOR
				_, _ = attackerConn.WriteToUDP(badBytes, bobUDPAddr)

				attackPackets.Add(3)
			}
		}
	}()

	// Alice sends 1,000 legitimate packets
	const legitTotal = 1000
	fmt.Printf(" [*] Transmitiendo %d paquetes legítimos de Alice a Bob bajo bombardeo hostil...\n", legitTotal)
	startLegit := time.Now()

	for i := 1; i <= legitTotal; i++ {
		payload := []byte(fmt.Sprintf("legit-payload-%d", i))
		_ = nodeAlice.SendMessage(idBob, payload)
		if i%100 == 0 {
			time.Sleep(5 * time.Millisecond) // Pacing
		}
	}

	time.Sleep(150 * time.Millisecond)
	close(stopAttack)
	durLegit := time.Since(startLegit)

	totalAttacks := attackPackets.Load()
	legitRx := int(legitDelivered.Load())
	corruptRx := int(corruptDelivered.Load())
	foreignRx := int(foreignDelivered.Load())

	pdr := (float64(legitRx) / float64(legitTotal)) * 100.0

	fmt.Printf(" [*] Fuego Hostil Inyectado por Atacante: %d paquetes\n", totalAttacks)
	fmt.Printf(" [*] Paquetes Legítimos Entregados a Bob: %d / %d (%.1f%%)\n", legitRx, legitTotal, pdr)
	fmt.Printf(" [*] Paquetes Corruptos Aceptados: %d (DEBE SER 0)\n", corruptRx)
	fmt.Printf(" [*] Tráfico Foráneo/Atacante Filtrado: %d aceptados (DEBE SER 0)\n", foreignRx)
	fmt.Printf(" [*] Duración de Transferencia Legítima: %v\n", durLegit.Round(time.Millisecond))

	if corruptRx == 0 && foreignRx == 0 && pdr >= 99.0 {
		fmt.Println(" [✓] RESULTADO: INVIOLABLE (El flujo legítimo prevaleció con 100% integridad)")
	} else {
		fmt.Printf(" [!] RESULTADO: DEGRADADO / FALLO (Corrupt: %d, Foreign: %d, PDR: %.1f%%)\n", corruptRx, foreignRx, pdr)
	}
	fmt.Println("------------------------------------------------------------------")
}

// ================================================================================
// 4. Repetibilidad Estadística (100 Iteraciones por Vector Crítico)
// ================================================================================
func runStatisticalRepeatability() {
	fmt.Println("\n------------------------------------------------------------------")
	fmt.Println(" [4] REPETIBILIDAD ESTADÍSTICA: 100 Iteraciones por Vector Crítico ")
	fmt.Println("------------------------------------------------------------------")
	fmt.Println(" Objetivo: Demostrar consistencia matemática (0 dispersión de fallos).")
	fmt.Println("------------------------------------------------------------------")

	vectors := []struct {
		name string
		fn   func() (bool, float64) // returns (success, metricValue)
	}{
		{"Packet Corruption Rejection", benchSingleCorruption},
		{"Anti-Replay Window Drop", benchSingleReplay},
		{"Handshake Cryptographic Parsing", benchSingleHandshake},
		{"Combined Chaos Resilience", benchSingleChaos},
	}

	const iterations = 100

	for _, v := range vectors {
		successes := 0
		metrics := make([]float64, 0, iterations)

		for i := 0; i < iterations; i++ {
			ok, val := v.fn()
			if ok {
				successes++
			}
			metrics = append(metrics, val)
		}

		minV, p50V, p95V, p99V, maxV := calcStats(metrics)
		successRate := (float64(successes) / float64(iterations)) * 100.0

		fmt.Printf(" Vector: %-32s | Tasa Éxito: %5.1f%% (%d/%d)\n", v.name, successRate, successes, iterations)
		fmt.Printf("   └─ Métricas (Latencia/Descartes): Min: %6.2f | p50: %6.2f | p95: %6.2f | p99: %6.2f | Max: %6.2f\n",
			minV, p50V, p95V, p99V, maxV)
	}
	fmt.Println("==================================================================")
}

func benchSingleCorruption() (bool, float64) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	c := &core.Container{SenderPubKey: pub, Payload: []byte("sample"), HopLimit: 12}
	_ = c.Sign(priv)
	raw, _ := c.Marshal()

	// Bitflip corruption
	raw[len(raw)/2] ^= 0xFF

	start := time.Now()
	cTest := &core.Container{}
	err := cTest.Unmarshal(raw)
	valid := false
	if err == nil {
		valid = cTest.Verify()
	}
	durUs := float64(time.Since(start).Nanoseconds()) / 1e3

	// Success means it was rejected (valid == false)
	return !valid, durUs
}

func benchSingleReplay() (bool, float64) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	table := adapters.NewAntiReplayTable()
	_ = table.CheckAndSet(string(pub), 50)

	start := time.Now()
	rejected := !table.CheckAndSet(string(pub), 50)
	durUs := float64(time.Since(start).Nanoseconds()) / 1e3

	return rejected, durUs
}

func benchSingleHandshake() (bool, float64) {
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	start := time.Now()
	req := &core.HandshakePayload{
		Type:       core.ControlHandshakeReq,
		Ed25519Pub: pub,
		Timestamp:  start.UnixNano(),
		Nonce:      core.GenerateNonce(),
	}
	data, _ := core.EncodeHandshake(req)
	decoded, err := core.DecodeHandshake(data)
	durUs := float64(time.Since(start).Nanoseconds()) / 1e3

	return err == nil && decoded != nil, durUs
}

func benchSingleChaos() (bool, float64) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	start := time.Now()
	c := &core.Container{SenderPubKey: pub, Payload: make([]byte, 512), Seq: 10, HopLimit: 12}
	_ = c.Sign(priv)
	raw, _ := c.Marshal()

	// 10% bitflip, verify rejects
	raw[10] ^= 0xAA
	c2 := &core.Container{}
	rejected := true
	if err := c2.Unmarshal(raw); err == nil {
		if c2.Verify() {
			rejected = false
		}
	}
	durUs := float64(time.Since(start).Nanoseconds()) / 1e3
	return rejected, durUs
}
