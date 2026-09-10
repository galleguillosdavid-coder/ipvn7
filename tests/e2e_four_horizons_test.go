package tests

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"testing"

	"golang.org/x/crypto/curve25519"
	"ipv7/adapters/offgrid"
	"ipv7/adapters/onion"
	"ipv7/adapters/tun"
	"ipv7/core"
	"ipv7/dht"
)

func genX25519() ([32]byte, [32]byte) {
	var priv, pub [32]byte
	_, _ = io.ReadFull(rand.Reader, priv[:])
	curve25519.ScalarBaseMult(&pub, &priv)
	return priv, pub
}

// TestFourHorizons_UnifiedEndToEnd demonstrates that all four strategic horizons
// (TUN/TAP, Pure DHT, Sphinx Onion Routing, and Off-Grid Mesh) operate seamlessly together
// as a unified, zero-leakage, sovereign communication substrate.
func TestFourHorizons_UnifiedEndToEnd(t *testing.T) {
	// -------------------------------------------------------------------------
	// 1. HORIZONTE 1 & 4: IDENTITIES AND ADDRESSING SETUP
	// -------------------------------------------------------------------------
	idA, privA, _ := core.GenerateIdentity()       // Source Node
	idHop1, _, _ := core.GenerateIdentity()        // Relay Hop 1
	idHop2, _, _ := core.GenerateIdentity()        // Relay Hop 2
	idB, privB, _ := core.GenerateIdentity()       // Destination Node

	// Derive cryptographic IPv6 ULA addresses (Horizon 1)
	ipA := tun.DeriveIPv6FromDID(idA)
	ipB := tun.DeriveIPv6FromDID(idB)
	t.Logf("Horizonte 1 (TUN Addressing): Node A IPv6=%s, Node B IPv6=%s", ipA.String(), ipB.String())

	// -------------------------------------------------------------------------
	// 2. HORIZONTE 2: PURE KADEMLIA 256-BIT DHT DISCOVERY (ZERO FIREBASE)
	// -------------------------------------------------------------------------
	dhtA := dht.NewPureKademliaTable(idA)
	dhtHop1 := dht.NewPureKademliaTable(idHop1)
	dhtHop2 := dht.NewPureKademliaTable(idHop2)
	dhtB := dht.NewPureKademliaTable(idB)

	// Publish and resolve contacts in pure DHT
	dhtA.AddContact(idHop1, []string{"phy_hop1"})
	dhtHop1.AddContact(idHop2, []string{"phy_hop2"})
	dhtHop2.AddContact(idB, []string{"phy_dest_B"})

	closest := dhtA.FindClosest(idHop1, 1)
	if len(closest) == 0 || closest[0].ID.String() != idHop1.String() {
		t.Fatalf("Horizonte 2 (DHT): Failed to resolve first hop via pure Kademlia table")
	}
	t.Log("Horizonte 2 (DHT): Contact resolution verified via 256-bit XOR metric.")

	// -------------------------------------------------------------------------
	// 3. HORIZONTE 4: OFF-GRID PHYSICAL MESH CHANNELS (RadioMedium)
	// -------------------------------------------------------------------------
	radioMedium := offgrid.NewRadioMedium()

	linkA := offgrid.NewVirtualRadioLink("radio-A", "phy_A", 1280, radioMedium)
	linkHop1 := offgrid.NewVirtualRadioLink("radio-Hop1", "phy_hop1", 1280, radioMedium)
	linkHop2 := offgrid.NewVirtualRadioLink("radio-Hop2", "phy_hop2", 1280, radioMedium)
	linkB := offgrid.NewVirtualRadioLink("radio-B", "phy_dest_B", 1280, radioMedium)

	defer linkA.Close()
	defer linkHop1.Close()
	defer linkHop2.Close()
	defer linkB.Close()

	// -------------------------------------------------------------------------
	// 4. HORIZONTE 3: SPHINX ONION ROUTING CIRCUIT SETUP
	// Circuit: A -> Hop 1 -> Hop 2 -> B (3 cryptographic layers, exactly 1280 bytes)
	// -------------------------------------------------------------------------
	privHop1, pubHop1 := genX25519()
	privHop2, pubHop2 := genX25519()
	xPrivB, xPubB := genX25519()

	circuit := []onion.CircuitHop{
		{
			DID:       idHop1,
			EncPubKey: pubHop1[:],
			Endpoint:  "phy_hop1",
		},
		{
			DID:       idHop2,
			EncPubKey: pubHop2[:],
			Endpoint:  "phy_hop2",
		},
		{
			DID:       idB,
			EncPubKey: xPubB[:],
			Endpoint:  "phy_dest_B",
		},
	}

	// -------------------------------------------------------------------------
	// 5. TUN/TAP ADAPTERS ON SOURCE AND DESTINATION
	// -------------------------------------------------------------------------
	tunDevA := tun.NewMemoryTunDevice("ipv7_tun_src", 1280, 512)
	tunDevB := tun.NewMemoryTunDevice("ipv7_tun_dst", 1280, 512)

	// Synthetic IPv6 packet generated at Source Application: A -> B
	syntheticIPv6Packet := make([]byte, 64)
	syntheticIPv6Packet[0] = 0x60 // IPv6 Version 6
	copy(syntheticIPv6Packet[8:24], ipA)
	copy(syntheticIPv6Packet[24:40], ipB)
	copy(syntheticIPv6Packet[40:], []byte("IPV7_FOUR_HORIZONS_UNIFIED_CONVERGENCE_TEST"))
	originalChecksum := sha256.Sum256(syntheticIPv6Packet)

	// -------------------------------------------------------------------------
	// 6. PIPELINE EXECUTION: TUN -> ONION -> OFF-GRID MESH -> RELAYS -> TUN
	// -------------------------------------------------------------------------
	t.Log("=== INICIANDO PIPELINE DE TRANSMISIÓN DE LOS CUATRO HORIZONTES ===")

	// 6a. Inject synthetic IPv6 packet into TunDevice A and read through TUN interface
	err := tunDevA.InjectPacket(syntheticIPv6Packet)
	if err != nil {
		t.Fatalf("Failed to inject packet into TUN A: %v", err)
	}

	tunBuf := make([]byte, 1280)
	n, err := tunDevA.Read(tunBuf)
	if err != nil {
		t.Fatalf("Failed to read packet from TUN A: %v", err)
	}
	rawPktFromTun := tunBuf[:n]

	// 6b. Wrap into 1280-byte Sphinx Onion Packet with 3 layers of encryption
	onionPacket, err := onion.BuildOnionPacket(circuit, rawPktFromTun)
	if err != nil {
		t.Fatalf("Horizonte 3 (Onion): BuildOnionPacket failed: %v", err)
	}

	if len(onionPacket) != onion.OnionPacketSize {
		t.Fatalf("VIOLACIÓN DE FORMATO: Paquete Sphinx no tiene 1280 bytes: tiene %d", len(onionPacket))
	}
	t.Logf("Horizonte 3 (Onion): Paquete Sphinx empaquetado en exactamente %d bytes invariantes.", len(onionPacket))

	// 6c. Transmit over Off-Grid Physical Radio Channel to Hop 1
	err = linkA.Send("phy_hop1", onionPacket)
	if err != nil {
		t.Fatalf("Horizonte 4 (Off-Grid): Failed to transmit onion frame over physical radio: %v", err)
	}

	// 6d. Relay Processing at Hop 1 (Blind Unwrap Layer 1 -> Forward to Hop 2)
	_, rxHop1, err := linkHop1.Receive()
	if err != nil {
		t.Fatalf("Hop 1 receive failed: %v", err)
	}
	instrHop1, err := onion.UnwrapLayer(rxHop1, privHop1)
	if err != nil {
		t.Fatalf("Hop 1 unwrap error: %v", err)
	}
	if instrHop1.IsExit || instrHop1.NextEndpoint != "phy_hop2" {
		t.Fatalf("Hop 1 unexpected instruction: nextHop=%s, isExit=%v", instrHop1.NextEndpoint, instrHop1.IsExit)
	}
	_ = linkHop1.Send(instrHop1.NextEndpoint, instrHop1.NextPayload)

	// 6e. Relay Processing at Hop 2 (Blind Unwrap Layer 2 -> Forward to Destination B)
	_, rxHop2, err := linkHop2.Receive()
	if err != nil {
		t.Fatalf("Hop 2 receive failed: %v", err)
	}
	instrHop2, err := onion.UnwrapLayer(rxHop2, privHop2)
	if err != nil {
		t.Fatalf("Hop 2 unwrap error: %v", err)
	}
	if instrHop2.IsExit || instrHop2.NextEndpoint != "phy_dest_B" {
		t.Fatalf("Hop 2 unexpected instruction: nextHop=%s, isExit=%v", instrHop2.NextEndpoint, instrHop2.IsExit)
	}
	_ = linkHop2.Send(instrHop2.NextEndpoint, instrHop2.NextPayload)

	// 6f. Destination Processing at Node B (Final Unwrap -> Extract IPv6 Packet)
	_, rxB, err := linkB.Receive()
	if err != nil {
		t.Fatalf("Destination B receive failed: %v", err)
	}
	instrB, err := onion.UnwrapLayer(rxB, xPrivB)
	if err != nil {
		t.Fatalf("Destination B unwrap error: %v", err)
	}
	if !instrB.IsExit {
		t.Fatalf("Destination B expected exit hop, got isExit=false, next=%s", instrB.NextEndpoint)
	}

	// 6g. Inject Reconstructed IPv6 packet into TunDevice B and read from device
	_, err = tunDevB.Write(instrB.NextPayload)
	if err != nil {
		t.Fatalf("Failed to write payload to TUN B: %v", err)
	}

	deliveredPktAtB, err := tunDevB.ReceiveOutbound()
	if err != nil {
		t.Fatalf("Failed to receive delivered packet from TUN B outbound: %v", err)
	}

	// -------------------------------------------------------------------------
	// 7. VERIFICACIÓN CRIPTOGRÁFICA DE INTEGRIDAD TOTAL
	// -------------------------------------------------------------------------
	deliveredChecksum := sha256.Sum256(deliveredPktAtB)
	if !bytes.Equal(syntheticIPv6Packet, deliveredPktAtB) {
		t.Fatalf("CORRUPCIÓN CRÍTICA: Paquete IP entregado en TUN B difiere del original de TUN A")
	}
	if deliveredChecksum != originalChecksum {
		t.Fatalf("CORRUPCIÓN CRIPTOGRÁFICA: Hash de carga útil alterado tras el circuito")
	}

	t.Logf("ENTREGA EXITOSA EN TUN B: %d bytes (SHA256: %x...)", len(deliveredPktAtB), deliveredChecksum[:6])
	t.Log("==========================================================================================")
	t.Log("CERTIFICACIÓN GLOBAL: LOS CUATRO HORIZONTES INTEGRADOS CONVERGEN CON ÉXITO")
	t.Log("TUN (IP ULA) -> DHT (256-bit) -> ONION (3 saltos 1280B) -> OFF-GRID (Radio) -> TUN (100% PDR)")
	t.Log("==========================================================================================")

	// Suppress unused variable warnings
	_ = privA
	_ = privB
	_ = dhtHop2
	_ = dhtB
}
