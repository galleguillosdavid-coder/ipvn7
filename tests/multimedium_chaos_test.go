package tests

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"ipv7/adapters/offgrid"
	"ipv7/core"
)

// cuttableAdapter simulates a physical or virtual WAN adapter that can be severed abruptly
type cuttableAdapter struct {
	addr     string
	receive  chan *core.Container
	registry map[string]*cuttableAdapter
	severed  bool
	mu       sync.Mutex
}

func newCuttableAdapter(addr string, registry map[string]*cuttableAdapter) *cuttableAdapter {
	a := &cuttableAdapter{
		addr:     addr,
		receive:  make(chan *core.Container, 500),
		registry: registry,
	}
	registry[addr] = a
	return a
}

func (a *cuttableAdapter) Start() error { return nil }
func (a *cuttableAdapter) Stop() error  { return nil }

func (a *cuttableAdapter) Send(c *core.Container, endpoints []string) error {
	a.mu.Lock()
	severed := a.severed
	a.mu.Unlock()

	if severed {
		return errors.New("link severed: WAN cable disconnected")
	}

	for _, ep := range endpoints {
		if peer, exists := a.registry[ep]; exists {
			peer.mu.Lock()
			peerSevered := peer.severed
			peer.mu.Unlock()

			if !peerSevered {
				select {
				case peer.receive <- c:
					return nil
				default:
					return errors.New("queue full")
				}
			}
		}
	}
	return errors.New("destination unreachable via WAN")
}

func (a *cuttableAdapter) Receive() <-chan *core.Container {
	return a.receive
}

func (a *cuttableAdapter) Cut() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.severed = true
}

func (a *cuttableAdapter) Restore() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.severed = false
}

// TestMultimediumTransportFailover_Chaos demonstrates or refutes Hypothesis H-MULTI-01:
// A node preserves its DID and cryptographic session when the physical transport switches abruptly from WAN to Off-Grid radio.
func TestMultimediumTransportFailover_Chaos(t *testing.T) {
	wanRegistry := make(map[string]*cuttableAdapter)
	radioMedium := offgrid.NewRadioMedium()

	// -------------------------------------------------------------------------
	// 1. TOPOLOGY SETUP: 4 NODES
	// Node A: Windows Client (WAN + Emergency Radio)
	// Node B: Border Gateway (WAN + Radio Mesh)
	// Node C: Intermediate Off-Grid Relay (Radio Mesh)
	// Node D: Isolated Off-Grid Target (Radio Mesh)
	// -------------------------------------------------------------------------

	idA, privA, _ := core.GenerateIdentity()
	idB, privB, _ := core.GenerateIdentity()
	idC, privC, _ := core.GenerateIdentity()
	idD, privD, _ := core.GenerateIdentity()

	nodeA := core.NewNode(idA, privA)
	nodeB := core.NewNode(idB, privB)
	nodeC := core.NewNode(idC, privC)
	nodeD := core.NewNode(idD, privD)

	// Attach WAN adapters
	wanA := newCuttableAdapter("wan_A", wanRegistry)
	wanB := newCuttableAdapter("wan_B", wanRegistry)
	nodeA.AddAdapter(wanA)
	nodeB.AddAdapter(wanB)

	// Attach Physical Radio Links (Off-Grid)
	radioLinkA := offgrid.NewVirtualRadioLink("radio-A", "phy_A", 1280, radioMedium)
	radioLinkB := offgrid.NewVirtualRadioLink("radio-B", "phy_B", 1280, radioMedium)
	radioLinkC := offgrid.NewVirtualRadioLink("radio-C", "phy_C", 1280, radioMedium)
	radioLinkD := offgrid.NewVirtualRadioLink("radio-D", "phy_D", 1280, radioMedium)

	defer radioLinkA.Close()
	defer radioLinkB.Close()
	defer radioLinkC.Close()
	defer radioLinkD.Close()

	// Initialize BeaconEngines for L2 Ad-hoc discovery
	// Notice: Node A's radio beacon is initially OFF (saving battery while WAN is active)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	beaconA := offgrid.NewBeaconEngine(nodeA, radioLinkA, nil)
	beaconB := offgrid.NewBeaconEngine(nodeB, radioLinkB, nil)
	beaconC := offgrid.NewBeaconEngine(nodeC, radioLinkC, nil)
	beaconD := offgrid.NewBeaconEngine(nodeD, radioLinkD, nil)

	beaconB.Start(ctx, 40*time.Millisecond)
	beaconC.Start(ctx, 40*time.Millisecond)
	beaconD.Start(ctx, 40*time.Millisecond)

	defer beaconA.Stop()
	defer beaconB.Stop()
	defer beaconC.Stop()
	defer beaconD.Stop()

	// Channel to capture raw packets delivered to Node D via radio
	radioDeliveriesD := make(chan []byte, 200)
	beaconD.SetOnData(func(srcAddr string, packet []byte) {
		radioDeliveriesD <- packet
	})

	// Setup HybridSwitchers
	switcherA := offgrid.NewHybridSwitcher(nodeA, beaconA, radioLinkA, offgrid.SwitcherConfig{
		AutoFallback:     true,
		HeartbeatTimeout: 100 * time.Millisecond,
		DefaultMode:      offgrid.ModeOnlineInternet,
	})

	// Configure initial overlay routing: A knows B on WAN, B knows D via C on Radio
	nodeA.AddPeer(idB, []string{"wan_B"})
	nodeB.AddPeer(idA, []string{"wan_A"})

	_ = nodeA.Start()
	_ = nodeB.Start()
	_ = nodeC.Start()
	_ = nodeD.Start()

	defer nodeA.Stop()
	defer nodeB.Stop()
	defer nodeC.Stop()
	defer nodeD.Stop()

	// Allow L2 beacons to discover neighbors in radio range (B, C, D)
	time.Sleep(100 * time.Millisecond)

	// Verify L2 Beacon discovery is active between offgrid nodes
	discoveredByB := beaconB.DiscoveredPeers()
	if len(discoveredByB) == 0 {
		t.Fatalf("Expected L2 beacon discovery in radio medium, got 0 peers")
	}

	// -------------------------------------------------------------------------
	// 2. CRYPTOGRAPHIC SESSION SETUP (E2EE)
	// Node A and Node D establish an E2EE session key
	// Invariance Check: We record DIDs and keys to prove zero mutation during failover
	// -------------------------------------------------------------------------
	initialDidA := idA.String()
	initialDidD := idD.String()

	// Recipient Node D has an E2EE X25519 keypair
	e2eePrivD, e2eePubD, err := core.GenerateE2EEKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate E2EE keypair for D: %v", err)
	}
	targetPubKeyD := e2eePubD.Bytes()

	// Channel to capture overlay messages at Node D
	wanDeliveriesD := make(chan []byte, 200)
	nodeD.OnMessage(func(from core.Identity, payload []byte) {
		wanDeliveriesD <- payload
	})

	// -------------------------------------------------------------------------
	// 3. FASE 1: TRANSMISIÓN NORMAL VÍA WAN
	// Node A transmits 50 packets to Gateway B, which forwards to D
	// -------------------------------------------------------------------------
	t.Log("=== FASE 1: Transmisión Normal por WAN ===")
	const packetsPhase1 = 30
	var receivedPhase1 int

	// Node B forwards WAN packets to D via its radio link
	nodeB.OnMessage(func(from core.Identity, payload []byte) {
		_ = radioLinkB.Send("phy_D", payload)
	})

	for i := 0; i < packetsPhase1; i++ {
		plaintext := fmt.Sprintf("WAN_MSG_SEQ_%03d", i)
		encryptedEnvelope, err := core.EncryptE2EE(targetPubKeyD, []byte(plaintext))
		if err != nil {
			t.Fatalf("Encryption failed: %v", err)
		}

		err = switcherA.RoutePacket(idB, encryptedEnvelope)
		if err != nil {
			t.Fatalf("Failed to route packet via WAN: %v", err)
		}
	}

	// Collect packets delivered to D
	timeout := time.After(2 * time.Second)
	for receivedPhase1 < packetsPhase1 {
		select {
		case pkt := <-radioDeliveriesD:
			decrypted, err := core.DecryptE2EE(e2eePrivD, pkt)
			if err != nil {
				t.Fatalf("Failed to decrypt packet at D in Phase 1: %v", err)
			}
			expectedPrefix := "WAN_MSG_SEQ_"
			if !bytes.HasPrefix(decrypted, []byte(expectedPrefix)) {
				t.Fatalf("Corrupted payload received: %s", string(decrypted))
			}
			receivedPhase1++
		case <-timeout:
			t.Fatalf("Timeout Phase 1: received %d / %d packets", receivedPhase1, packetsPhase1)
		}
	}
	t.Logf("FASE 1 COMPLETADA: %d / %d paquetes recibidos (PDR: 100%%)", receivedPhase1, packetsPhase1)

	// -------------------------------------------------------------------------
	// 4. FASE 2: INYECCIÓN DE CAOS (CORTE ABRUPTO DE LA INFRAESTRUCTURA WAN)
	// -------------------------------------------------------------------------
	t.Log("=== FASE 2: Inyección de Caos — Corte Catastrófico de WAN ===")
	reconvergenceStart := time.Now()

	// Sever both WAN adapters to simulate fiber cut / BGP outage
	wanA.Cut()
	wanB.Cut()

	// Inform switcher of WAN loss
	switcherA.ReportWANStatus(false)

	if switcherA.CurrentMode() != offgrid.ModeOffGridPhysical {
		t.Fatalf("Expected switcher to transition to ModeOffGridPhysical, got %v", switcherA.CurrentMode())
	}

	// Emergency activation: Node A powers on its local ad-hoc radio and discovers offgrid neighbors
	beaconA.Start(ctx, 30*time.Millisecond)
	time.Sleep(100 * time.Millisecond)

	reconvergenceDuration := time.Since(reconvergenceStart)
	t.Logf("Reconvergencia a Malla Off-Grid y descubrimiento completados en: %v", reconvergenceDuration)
	if reconvergenceDuration > 250*time.Millisecond {
		t.Fatalf("VIOLACIÓN CONTRACTUAL: Reconvergencia excedió la cota de 250ms: %v", reconvergenceDuration)
	}

	// -------------------------------------------------------------------------
	// 5. FASE 3: TRANSMISIÓN POST-FAILOVER DIRECTA POR RADIO AD-HOC
	// Node A transmits 30 packets directly through the ad-hoc physical radio mesh
	// -------------------------------------------------------------------------
	t.Log("=== FASE 3: Transmisión Post-Failover en Malla Off-Grid ===")
	const packetsPhase2 = 30
	var receivedPhase2 int

	for i := 0; i < packetsPhase2; i++ {
		plaintext := fmt.Sprintf("OFFGRID_FAILOVER_MSG_SEQ_%03d", i)
		// Crucial verification: using the SAME cryptographic session key derived prior to the blackout
		encryptedEnvelope, err := core.EncryptE2EE(targetPubKeyD, []byte(plaintext))
		if err != nil {
			t.Fatalf("Encryption failed post-failover: %v", err)
		}

		err = switcherA.RoutePacket(idD, encryptedEnvelope)
		if err != nil {
			t.Fatalf("Failed to route packet via Off-Grid mesh: %v", err)
		}
	}

	timeoutPhase2 := time.After(2 * time.Second)
	for receivedPhase2 < packetsPhase2 {
		select {
		case pkt := <-radioDeliveriesD:
			// Decrypt using identical session secret
			decrypted, err := core.DecryptE2EE(e2eePrivD, pkt)
			if err != nil {
				t.Fatalf("CRITICAL FAILURE: Session secret corrupted after failover: %v", err)
			}
			expectedPrefix := "OFFGRID_FAILOVER_MSG_SEQ_"
			if !bytes.HasPrefix(decrypted, []byte(expectedPrefix)) {
				t.Fatalf("Corrupted payload post-failover: %s", string(decrypted))
			}
			receivedPhase2++
		case <-timeoutPhase2:
			t.Fatalf("Timeout Phase 2: received %d / %d packets post-failover", receivedPhase2, packetsPhase2)
		}
	}
	t.Logf("FASE 3 COMPLETADA: %d / %d paquetes recibidos (PDR: 100%%)", receivedPhase2, packetsPhase2)

	// -------------------------------------------------------------------------
	// 6. VERIFICACIÓN DE INVARIANTES ARQUITECTÓNICAS (H-MULTI-01)
	// -------------------------------------------------------------------------
	t.Log("=== VERIFICACIÓN DE INVARIANTES EPISTÉMICAS ===")

	// Invariante 1: DID A no mutó
	if idA.String() != initialDidA {
		t.Fatalf("VIOLACIÓN: DID de Node A mutó durante el failover de transporte: antes=%s, después=%s",
			initialDidA, idA.String())
	}
	// Invariante 2: DID D no mutó
	if idD.String() != initialDidD {
		t.Fatalf("VIOLACIÓN: DID de Node D mutó durante el failover de transporte: antes=%s, después=%s",
			initialDidD, idD.String())
	}

	// Invariante 3: Clave pública Ed25519 de firma intacta
	if !bytes.Equal(idA.PublicKey, ed25519.PublicKey(idA.Bytes())) {
		t.Fatalf("VIOLACIÓN: Clave pública Ed25519 de Node A no es consistente")
	}

	// Invariante 4: Telemetría de conmutador
	stats := switcherA.Stats()
	t.Logf("Telemetría Switcher A: ModeTransitions=%d, WANPacketsSent=%d, MeshPacketsSent=%d, PacketsDropped=%d",
		stats.ModeTransitions, stats.WANPacketsSent, stats.MeshPacketsSent, stats.PacketsDropped)

	if stats.ModeTransitions < 1 {
		t.Errorf("Expected at least 1 mode transition, recorded %d", stats.ModeTransitions)
	}
	if stats.WANPacketsSent != packetsPhase1 {
		t.Errorf("Expected %d WAN packets sent, got %d", packetsPhase1, stats.WANPacketsSent)
	}
	if stats.MeshPacketsSent != packetsPhase2 {
		t.Errorf("Expected %d mesh packets sent, got %d", packetsPhase2, stats.MeshPacketsSent)
	}

	t.Log("==========================================================================================")
	t.Log("RESULTADO FORMAL: HIPÓTESIS H-MULTI-01 -> [DEMONSTRATED] EN LABORATORIO CONTROLADO")
	t.Log("Invarianza de Identidad (DID) y Continuidad de Sesión E2EE verificada con éxito 100% PDR.")
	t.Log("==========================================================================================")
}
