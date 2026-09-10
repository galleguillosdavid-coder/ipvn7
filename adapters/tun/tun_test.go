package tun

import (
	"bytes"
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"ipv7/core"
)

// mockAdapter is an in-memory test adapter simulating network transfer
type mockAdapter struct {
	addr    string
	receive chan *core.Container
	peers   map[string]*mockAdapter
	mu      sync.Mutex
}

func newMockAdapter(addr string, registry map[string]*mockAdapter) *mockAdapter {
	a := &mockAdapter{
		addr:    addr,
		receive: make(chan *core.Container, 100),
		peers:   registry,
	}
	registry[addr] = a
	return a
}

func (m *mockAdapter) Start() error { return nil }
func (m *mockAdapter) Stop() error  { return nil }
func (m *mockAdapter) Send(c *core.Container, endpoints []string) error {
	for _, ep := range endpoints {
		if peer, exists := m.peers[ep]; exists {
			peer.receive <- c
			return nil
		}
	}
	return nil
}
func (m *mockAdapter) Receive() <-chan *core.Container { return m.receive }

func TestAddressingDerivation(t *testing.T) {
	id, _, err := core.GenerateIdentity()
	if err != nil {
		t.Fatalf("Failed to generate identity: %v", err)
	}

	ipv6 := DeriveIPv6FromDID(id)
	if len(ipv6) != 16 {
		t.Fatalf("Expected 16-byte IPv6, got %d bytes", len(ipv6))
	}

	// Verify prefix fd07::
	if ipv6[0] != 0xfd || ipv6[1] != 0x07 {
		t.Fatalf("Expected ULA prefix fd07::, got %x%x", ipv6[0], ipv6[1])
	}

	// Determinism: same DID must always yield identical IPv6
	ipv6Again := DeriveIPv6FromDID(id)
	if !ipv6.Equal(ipv6Again) {
		t.Fatalf("Expected deterministic IPv6 derivation, got %s vs %s", ipv6, ipv6Again)
	}

	ipv4 := GenerateVirtualIPv4(42)
	if !ipv4.Equal(net.IPv4(10, 7, 0, 42)) {
		t.Fatalf("Expected 10.7.0.42, got %s", ipv4)
	}
}

func TestVirtualRouteTable(t *testing.T) {
	id, _, _ := core.GenerateIdentity()
	tbl := NewVirtualRouteTable()

	testIP := net.IPv4(10, 7, 1, 50)
	tbl.Register(testIP, id)

	did, found := tbl.LookupDID(testIP)
	if !found || did != id {
		t.Fatalf("Expected to find node identity %s, got %s (found: %v)", id, did, found)
	}

	ip, found := tbl.LookupIP(id)
	if !found || !ip.Equal(testIP) {
		t.Fatalf("Expected IP %s, got %s (found: %v)", testIP, ip, found)
	}

	// Unknown IP
	_, found = tbl.LookupDID(net.IPv4(10, 7, 99, 99))
	if found {
		t.Fatalf("Expected unknown IP to not be found")
	}
}

func TestExtractDestinationIP(t *testing.T) {
	// Synthesize valid IPv4 header (20 bytes minimum)
	ipv4Header := make([]byte, 20)
	ipv4Header[0] = 0x45 // Version 4, IHL 5
	// Source IP: 10.7.0.1 (bytes 12-15)
	copy(ipv4Header[12:16], []byte{10, 7, 0, 1})
	// Dest IP: 10.7.0.2 (bytes 16-19)
	copy(ipv4Header[16:20], []byte{10, 7, 0, 2})

	dst := extractDestinationIP(ipv4Header)
	if dst == nil || !dst.Equal(net.IPv4(10, 7, 0, 2)) {
		t.Fatalf("Expected destination 10.7.0.2, got %v", dst)
	}

	// Synthesize valid IPv6 header (40 bytes)
	ipv6Header := make([]byte, 40)
	ipv6Header[0] = 0x60 // Version 6
	// Dest IPv6: fd07::1 at offset 24-39
	copy(ipv6Header[24:40], net.ParseIP("fd07::1"))

	dst6 := extractDestinationIP(ipv6Header)
	if dst6 == nil || !dst6.Equal(net.ParseIP("fd07::1")) {
		t.Fatalf("Expected destination fd07::1, got %v", dst6)
	}
}

func TestTunAdapterEndToEnd(t *testing.T) {
	registry := make(map[string]*mockAdapter)

	// Node A
	idA, privA, _ := core.GenerateIdentity()
	nodeA := core.NewNode(idA, privA)
	adapterNetA := newMockAdapter("addrA", registry)
	nodeA.AddAdapter(adapterNetA)

	// Node B
	idB, privB, _ := core.GenerateIdentity()
	nodeB := core.NewNode(idB, privB)
	adapterNetB := newMockAdapter("addrB", registry)
	nodeB.AddAdapter(adapterNetB)

	// Interconnect nodes A and B
	nodeA.AddPeer(idB, []string{"addrB"})
	nodeB.AddPeer(idA, []string{"addrA"})

	if err := nodeA.Start(); err != nil {
		t.Fatalf("Failed to start node A: %v", err)
	}
	defer nodeA.Stop()

	if err := nodeB.Start(); err != nil {
		t.Fatalf("Failed to start node B: %v", err)
	}
	defer nodeB.Stop()

	tunDevA := NewMemoryTunDevice("ipv7_test0", 1280, 1024)
	adapterA := NewTunAdapter(tunDevA, nodeA, nil)

	ipB := net.IPv4(10, 7, 0, 2)
	adapterA.Routes().Register(ipB, idB)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := adapterA.Start(ctx); err != nil {
		t.Fatalf("Failed to start adapter A: %v", err)
	}
	defer func() { _ = adapterA.Stop() }()

	received := make(chan []byte, 1)
	nodeB.OnMessage(func(from core.Identity, payload []byte) {
		received <- payload
	})

	// Inject synthetic IPv4 packet targeting IP B into TunDevice A
	syntheticPkt := make([]byte, 28)
	syntheticPkt[0] = 0x45
	copy(syntheticPkt[16:20], []byte{10, 7, 0, 2})
	copy(syntheticPkt[20:], []byte("PING_DATA"))

	if err := tunDevA.InjectPacket(syntheticPkt); err != nil {
		t.Fatalf("Failed to inject packet: %v", err)
	}

	select {
	case payload := <-received:
		if !bytes.Equal(payload, syntheticPkt) {
			t.Fatalf("Payload mismatch: got %v, expected %v", payload, syntheticPkt)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Timed out waiting for packet to arrive over IPv7 mesh")
	}

	// Test Inbound delivery to virtual interface
	responsePkt := []byte("PONG_DATA")
	if err := adapterA.DeliverInbound(responsePkt); err != nil {
		t.Fatalf("DeliverInbound failed: %v", err)
	}

	outPkt, err := tunDevA.ReceiveOutbound()
	if err != nil || !bytes.Equal(outPkt, responsePkt) {
		t.Fatalf("Inbound packet delivery mismatch: %v, err: %v", outPkt, err)
	}

	if adapterA.PacketsIn() != 1 || adapterA.PacketsOut() != 1 {
		t.Fatalf("Counters mismatch: in=%d out=%d", adapterA.PacketsIn(), adapterA.PacketsOut())
	}
}

func BenchmarkExtractDestinationIP(b *testing.B) {
	ipv4Header := make([]byte, 20)
	ipv4Header[0] = 0x45
	copy(ipv4Header[16:20], []byte{10, 7, 0, 2})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dst := extractDestinationIP(ipv4Header)
		if dst == nil {
			b.Fatal("nil destination")
		}
	}
}

func BenchmarkExtractDestinationIPKeyZeroAlloc(b *testing.B) {
	ipv4Header := make([]byte, 20)
	ipv4Header[0] = 0x45
	copy(ipv4Header[16:20], []byte{10, 7, 0, 2})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		key, ok := extractDestinationIPKey(ipv4Header)
		if !ok || key[15] != 2 {
			b.Fatal("invalid destination key")
		}
	}
}

func BenchmarkVirtualRouteLookup(b *testing.B) {
	id, _, _ := core.GenerateIdentity()
	tbl := NewVirtualRouteTable()
	testIP := net.IPv4(10, 7, 0, 2)
	tbl.Register(testIP, id)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, found := tbl.LookupDID(testIP)
		if !found {
			b.Fatal("not found")
		}
	}
}

func BenchmarkVirtualRouteLookupZeroAlloc(b *testing.B) {
	id, _, _ := core.GenerateIdentity()
	tbl := NewVirtualRouteTable()
	testIP := net.IPv4(10, 7, 0, 2)
	tbl.Register(testIP, id)
	key := ToIPKey(testIP)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, found := tbl.LookupDIDKey(key)
		if !found {
			b.Fatal("not found")
		}
	}
}

func TestTunAdapterRoamingPreservation(t *testing.T) {
	registry := make(map[string]*mockAdapter)

	// Node A
	idA, privA, _ := core.GenerateIdentity()
	nodeA := core.NewNode(idA, privA)
	adapterNetA := newMockAdapter("addrA", registry)
	nodeA.AddAdapter(adapterNetA)

	// Node B
	idB, privB, _ := core.GenerateIdentity()
	nodeB := core.NewNode(idB, privB)
	adapterNetB := newMockAdapter("addrB_Initial", registry)
	nodeB.AddAdapter(adapterNetB)

	nodeA.AddPeer(idB, []string{"addrB_Initial"})
	nodeB.AddPeer(idA, []string{"addrA"})

	_ = nodeA.Start()
	defer nodeA.Stop()
	_ = nodeB.Start()
	defer nodeB.Stop()

	tunDevA := NewMemoryTunDevice("ipv7_test_roam", 1280, 1024)
	adapterA := NewTunAdapter(tunDevA, nodeA, nil)

	// Register virtual IP for B
	virtIP := net.IPv4(10, 7, 0, 88)
	adapterA.Routes().Register(virtIP, idB)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = adapterA.Start(ctx)
	defer func() { _ = adapterA.Stop() }()

	received := make(chan []byte, 10)
	nodeB.OnMessage(func(from core.Identity, payload []byte) {
		received <- payload
	})

	// 1. Send packet prior to roaming
	pkt1 := make([]byte, 28)
	pkt1[0] = 0x45
	copy(pkt1[16:20], []byte{10, 7, 0, 88})
	copy(pkt1[20:], []byte("BEFORE_ROAM"))
	_ = tunDevA.InjectPacket(pkt1)

	select {
	case p := <-received:
		if !bytes.Equal(p, pkt1) {
			t.Fatalf("Pre-roam payload mismatch")
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Pre-roam timeout")
	}

	// 2. Peer B roams to a new network interface "addrB_Roamed_Cellular"
	registry["addrB_Roamed_Cellular"] = adapterNetB

	// Inform TunAdapter of physical roaming
	adapterA.HandlePeerRoamed(idB, []string{"addrB_Roamed_Cellular"})

	// Verify virtual IP remains exactly identical
	storedIP, found := adapterA.Routes().LookupIP(idB)
	if !found || !storedIP.Equal(virtIP) {
		t.Fatalf("Virtual IP altered after roaming: expected %s, got %s", virtIP, storedIP)
	}

	// 3. Send packet post-roaming targeting same virtual IP
	pkt2 := make([]byte, 28)
	pkt2[0] = 0x45
	copy(pkt2[16:20], []byte{10, 7, 0, 88})
	copy(pkt2[20:], []byte("AFTER_ROAM"))
	_ = tunDevA.InjectPacket(pkt2)

	select {
	case p := <-received:
		if !bytes.Equal(p, pkt2) {
			t.Fatalf("Post-roam payload mismatch: %s", string(p))
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Post-roam timeout: packet lost after roaming")
	}
}

func TestTunAdapterSoak10kPackets(t *testing.T) {
	registry := make(map[string]*mockAdapter)

	idA, privA, _ := core.GenerateIdentity()
	nodeA := core.NewNode(idA, privA)
	adapterNetA := newMockAdapter("addrA_Soak", registry)
	nodeA.AddAdapter(adapterNetA)

	idB, privB, _ := core.GenerateIdentity()
	nodeB := core.NewNode(idB, privB)
	adapterNetB := newMockAdapter("addrB_Soak", registry)
	nodeB.AddAdapter(adapterNetB)

	nodeA.AddPeer(idB, []string{"addrB_Soak"})
	nodeB.AddPeer(idA, []string{"addrA_Soak"})

	_ = nodeA.Start()
	defer nodeA.Stop()
	_ = nodeB.Start()
	defer nodeB.Stop()

	tunDevA := NewMemoryTunDevice("ipv7_soak", 1280, 4096)
	adapterA := NewTunAdapter(tunDevA, nodeA, nil)

	virtIP := net.IPv4(10, 7, 5, 1)
	adapterA.Routes().Register(virtIP, idB)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = adapterA.Start(ctx)
	defer func() { _ = adapterA.Stop() }()

	var receivedCount int
	var mu sync.Mutex
	done := make(chan struct{})

	const totalPackets = 10000

	nodeB.OnMessage(func(from core.Identity, payload []byte) {
		mu.Lock()
		receivedCount++
		if receivedCount == totalPackets {
			close(done)
		}
		mu.Unlock()
	})

	syntheticPkt := make([]byte, 28)
	syntheticPkt[0] = 0x45
	copy(syntheticPkt[16:20], []byte{10, 7, 5, 1})
	copy(syntheticPkt[20:], []byte("SOAK"))

	for i := 0; i < totalPackets; i++ {
		for err := tunDevA.InjectPacket(syntheticPkt); err != nil; err = tunDevA.InjectPacket(syntheticPkt) {
			time.Sleep(10 * time.Microsecond)
		}
	}

	select {
	case <-done:
		// 100% PDR achieved
	case <-time.After(10 * time.Second):
		mu.Lock()
		count := receivedCount
		mu.Unlock()
		t.Fatalf("Soak test timed out: delivered %d / %d packets", count, totalPackets)
	}
}
