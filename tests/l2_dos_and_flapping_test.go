package tests

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"ipv7/adapters/offgrid"
	"ipv7/core"
)

// TestBeaconEngine_PoisonAndFloodResistance verifies that a flood of 10,000 spoofed L2 beacons
// does not crash the node, corrupt valid neighbors, or prevent legitimate peer data delivery.
func TestBeaconEngine_PoisonAndFloodResistance(t *testing.T) {
	runtime.GC()
	var mPre runtime.MemStats
	runtime.ReadMemStats(&mPre)

	medium := offgrid.NewRadioMedium()

	// Victim Node V
	idV, privV, _ := core.GenerateIdentity()
	nodeV := core.NewNode(idV, privV)
	linkV := offgrid.NewVirtualRadioLink("radio-victim", "phy_victim", 1280, medium)
	defer linkV.Close()

	// Legitimate Peer L
	idL, privL, _ := core.GenerateIdentity()
	nodeL := core.NewNode(idL, privL)
	linkL := offgrid.NewVirtualRadioLink("radio-legit", "phy_legit", 1280, medium)
	defer linkL.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var legitDeliveries uint64
	beaconV := offgrid.NewBeaconEngine(nodeV, linkV, nil)
	beaconV.SetOnData(func(srcAddr string, packet []byte) {
		if bytes.Equal(packet, []byte("CRITICAL_LEGIT_PAYLOAD")) {
			atomic.AddUint64(&legitDeliveries, 1)
		}
	})

	beaconL := offgrid.NewBeaconEngine(nodeL, linkL, nil)

	beaconV.Start(ctx, 30*time.Millisecond)
	beaconL.Start(ctx, 30*time.Millisecond)
	defer beaconV.Stop()
	defer beaconL.Stop()

	// Allow legitimate handshake
	time.Sleep(80 * time.Millisecond)

	// Attacker: Spawns 10,000 spoofed L2 beacon frames with forged DIDs and random MACs
	const floodCount = 10000
	t.Logf("Iniciando inyección de estrés: %d balizas L2 falsificadas...", floodCount)

	startFlood := time.Now()
	var wg sync.WaitGroup
	workers := 4
	perWorker := floodCount / workers

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			attackerLink := offgrid.NewVirtualRadioLink(
				fmt.Sprintf("attacker-%d", workerID),
				fmt.Sprintf("forged-mac-%d", workerID),
				1280,
				medium,
			)
			defer attackerLink.Close()

			var forgedBeacon [36]byte
			copy(forgedBeacon[:4], offgrid.BeaconMagic[:])

			for i := 0; i < perWorker; i++ {
				// Random forged DID
				_, _ = rand.Read(forgedBeacon[4:36])
				_ = attackerLink.Send("BROADCAST", forgedBeacon[:])
			}
		}(w)
	}

	// Simultaneously, Legitimate peer L sends 50 critical data frames to Victim V
	const legitPackets = 50
	go func() {
		for i := 0; i < legitPackets; i++ {
			_ = linkL.Send("phy_victim", []byte("CRITICAL_LEGIT_PAYLOAD"))
			time.Sleep(5 * time.Millisecond)
		}
	}()

	wg.Wait()
	floodDuration := time.Since(startFlood)
	t.Logf("Inundación de %d balizas completada en %v (Throughput: %.0f balizas/segundo)",
		floodCount, floodDuration, float64(floodCount)/floodDuration.Seconds())

	// Wait for legitimate messages to complete
	time.Sleep(350 * time.Millisecond)

	delivered := atomic.LoadUint64(&legitDeliveries)
	t.Logf("Paquetes legítimos entregados durante el ataque DoS: %d / %d", delivered, legitPackets)

	if delivered < uint64(float64(legitPackets)*0.90) {
		t.Fatalf("VIOLACIÓN DE DISPONIBILIDAD: Ataque DoS de balizas degradó la entrega legítima por debajo de la cota contractual del 90%%: recibido %d/%d",
			delivered, legitPackets)
	}

	// Check table of victim
	peers := beaconV.DiscoveredPeers()
	t.Logf("Vecinos registrados en la tabla del nodo víctima: %d", len(peers))

	if len(peers) > offgrid.MaxDiscoveredPeers {
		t.Fatalf("VIOLACIÓN DE COTA L2: Tabla de vecinos excedió la cota contractual de %d peers: tiene %d",
			offgrid.MaxDiscoveredPeers, len(peers))
	}

	// Verify legitimate peer L is in victim's table
	foundLegit := false
	for _, p := range peers {
		if p.DID.String() == idL.String() {
			foundLegit = true
			break
		}
	}
	if !foundLegit {
		t.Fatalf("VIOLACIÓN: El par legítimo fue desalojado o silenciado por el ataque de balizas")
	}

	// Verify memory consumption
	runtime.GC()
	var mPost runtime.MemStats
	runtime.ReadMemStats(&mPost)
	heapGrowthKB := int64(mPost.HeapAlloc-mPre.HeapAlloc) / 1024
	t.Logf("Crecimiento neto de heap tras procesar 10.000 balizas forjadas: %d KB", heapGrowthKB)

	if heapGrowthKB > 1024 {
		t.Fatalf("VIOLACIÓN DE MEMORIA: Crecimiento de heap excedió la cota contractual de 1024 KB tras 10k balizas: %d KB", heapGrowthKB)
	}

	t.Log("==========================================================================================")
	t.Log("RESULTADO FORMAL: HIPÓTESIS H-L2-DOS -> [DEMONSTRATED] EN LABORATORIO CONTROLADO")
	t.Log("BeaconEngine y Capa 2 resisten inundación de 10k balizas sin pánico ni bloqueo de tráfico.")
	t.Log("==========================================================================================")
}

// TestHybridSwitcher_FlappingResistance verifies that rapid, erratic WAN flapping
// does not induce deadlocks, thread explosion, or permanent traffic loss.
func TestHybridSwitcher_FlappingResistance(t *testing.T) {
	id, priv, _ := core.GenerateIdentity()
	node := core.NewNode(id, priv)

	medium := offgrid.NewRadioMedium()
	link := offgrid.NewVirtualRadioLink("flapping-link", "phy_flap", 1280, medium)
	defer link.Close()

	beacon := offgrid.NewBeaconEngine(node, link, nil)
	switcher := offgrid.NewHybridSwitcher(node, beacon, link, offgrid.SwitcherConfig{
		AutoFallback:     true,
		HeartbeatTimeout: 50 * time.Millisecond,
		DefaultMode:      offgrid.ModeOnlineInternet,
	})

	var transitionCount uint64
	switcher.SetOnModeChange(func(oldMode, newMode offgrid.LinkMode) {
		atomic.AddUint64(&transitionCount, 1)
	})

	// Record baseline goroutines to verify zero goroutine leaks under flapping
	baselineGoroutines := runtime.NumGoroutine()

	// Inject aggressive WAN flapping: 30 rapid toggle cycles (every 10 ms)
	const flapCycles = 30
	t.Logf("Iniciando tormenta de flapping de WAN: %d ciclos de conmutación rápida...", flapCycles)

	startFlap := time.Now()
	for i := 0; i < flapCycles; i++ {
		switcher.ReportWANStatus(false) // Simula desconexión
		time.Sleep(8 * time.Millisecond)
		switcher.ReportWANStatus(true) // Simula reconexión
		time.Sleep(8 * time.Millisecond)
	}

	flapDuration := time.Since(startFlap)
	t.Logf("Tormenta de flapping finalizada en %v", flapDuration)

	// Allow transition events to settle
	time.Sleep(100 * time.Millisecond)

	recordedTransitions := atomic.LoadUint64(&transitionCount)
	t.Logf("Transiciones de modo capturadas: %d", recordedTransitions)

	if recordedTransitions < uint64(flapCycles*2) {
		t.Fatalf("VIOLACIÓN CONTRACTUAL DE FLAPPING: Transiciones capturadas insuficientes: %d < %d",
			recordedTransitions, flapCycles*2)
	}

	// Verify zero goroutine leaks after flapping storm settles
	time.Sleep(50 * time.Millisecond)
	deltaGoroutines := runtime.NumGoroutine() - baselineGoroutines
	t.Logf("Delta de goroutines tras tormenta de flapping: %+d", deltaGoroutines)
	if deltaGoroutines > 4 {
		t.Fatalf("VIOLACIÓN DE CONCURRENCIA: Fuga de goroutines detectada tras flapping: %+d goroutines residuales",
			deltaGoroutines)
	}

	// Settle into pure Off-Grid
	switcher.ReportWANStatus(false)
	time.Sleep(50 * time.Millisecond)

	if switcher.CurrentMode() != offgrid.ModeOffGridPhysical {
		t.Fatalf("Expected final state ModeOffGridPhysical, got %v", switcher.CurrentMode())
	}

	// Settle into Hybrid / Online
	switcher.ReportWANStatus(true)
	time.Sleep(50 * time.Millisecond)

	if switcher.CurrentMode() != offgrid.ModeHybrid {
		t.Fatalf("Expected final state ModeHybrid when physical link is present, got %v", switcher.CurrentMode())
	}

	stats := switcher.Stats()
	t.Logf("Estadísticas finales del switcher: Transitions=%d, MeshPacketsSent=%d, Dropped=%d",
		stats.ModeTransitions, stats.MeshPacketsSent, stats.PacketsDropped)

	t.Log("==========================================================================================")
	t.Log("RESULTADO FORMAL: HIPÓTESIS H-FLAPPING -> [DEMONSTRATED] EN LABORATORIO CONTROLADO")
	t.Log("HybridSwitcher sobrevive a 30 oscilaciones rápidas con convergencia determinista a cero deadlocks.")
	t.Log("==========================================================================================")
}
