package offgrid

import (
	"context"
	"testing"
	"time"

	"ipv7/core"
)

func TestVirtualRadioLink_Broadcast(t *testing.T) {
	medium := NewRadioMedium()

	link1 := NewVirtualRadioLink("wifi-direct-1", "02:00:00:00:00:01", 1280, medium)
	link2 := NewVirtualRadioLink("wifi-direct-2", "02:00:00:00:00:02", 1280, medium)
	defer link1.Close()
	defer link2.Close()

	payload := []byte("HELLO_AD_HOC_RADIO")
	err := link1.Send("BROADCAST", payload)
	if err != nil {
		t.Fatalf("link1.Send error: %v", err)
	}

	src, received, err := link2.Receive()
	if err != nil {
		t.Fatalf("link2.Receive error: %v", err)
	}

	if src != "02:00:00:00:00:01" {
		t.Errorf("expected src 02:00:00:00:00:01, got %s", src)
	}
	if string(received) != string(payload) {
		t.Errorf("expected payload %s, got %s", payload, received)
	}
}

func TestBeaconEngine_Discovery(t *testing.T) {
	id1, _, _ := core.GenerateIdentity()
	id2, _, _ := core.GenerateIdentity()

	node1 := &core.Node{Identity: id1}
	node2 := &core.Node{Identity: id2}

	medium := NewRadioMedium()
	link1 := NewVirtualRadioLink("lora-1", "lora-node-1", 1280, medium)
	link2 := NewVirtualRadioLink("lora-2", "lora-node-2", 1280, medium)
	defer link1.Close()
	defer link2.Close()

	discovered := make(chan PhysicalPeer, 2)
	b1 := NewBeaconEngine(node1, link1, nil)
	b2 := NewBeaconEngine(node2, link2, func(peer PhysicalPeer) {
		discovered <- peer
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b1.Start(ctx, 50*time.Millisecond)
	b2.Start(ctx, 50*time.Millisecond)
	defer b1.Stop()
	defer b2.Stop()

	select {
	case peer := <-discovered:
		if peer.DID.String() != id1.String() {
			t.Fatalf("expected discovered peer DID %s, got %s", id1.String(), peer.DID.String())
		}
		if peer.LocalAddr != "lora-node-1" {
			t.Fatalf("expected local addr lora-node-1, got %s", peer.LocalAddr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for beacon discovery")
	}

	peers := b2.DiscoveredPeers()
	if len(peers) != 1 {
		t.Fatalf("expected 1 discovered peer in table, got %d", len(peers))
	}
}

func TestHybridSwitcher_FailoverAndRouting(t *testing.T) {
	id1, _, _ := core.GenerateIdentity()
	id2, _, _ := core.GenerateIdentity()

	node1 := &core.Node{Identity: id1}
	node2 := &core.Node{Identity: id2}

	medium := NewRadioMedium()
	link1 := NewVirtualRadioLink("radio-1", "phy-1", 1280, medium)
	link2 := NewVirtualRadioLink("radio-2", "phy-2", 1280, medium)
	defer link1.Close()
	defer link2.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b1 := NewBeaconEngine(node1, link1, nil)
	b2 := NewBeaconEngine(node2, link2, nil)
	b1.Start(ctx, 30*time.Millisecond)
	b2.Start(ctx, 30*time.Millisecond)
	defer b1.Stop()
	defer b2.Stop()

	// Wait for discovery
	time.Sleep(100 * time.Millisecond)

	modeChangeCh := make(chan LinkMode, 4)
	switcher := NewHybridSwitcher(node1, b1, link1, SwitcherConfig{
		AutoFallback:     true,
		HeartbeatTimeout: 200 * time.Millisecond,
		DefaultMode:      ModeOnlineInternet,
	})
	switcher.SetOnModeChange(func(oldMode, newMode LinkMode) {
		modeChangeCh <- newMode
	})

	if switcher.CurrentMode() != ModeOnlineInternet {
		t.Fatalf("expected initial ModeOnlineInternet, got %v", switcher.CurrentMode())
	}

	dataReceivedCh := make(chan []byte, 4)
	b2.SetOnData(func(srcAddr string, p []byte) {
		dataReceivedCh <- p
	})

	// 1. Send packet while peer is discovered via radio:
	// Since destination is in direct physical range, switcher transmits it over physical link
	packet := []byte("PING_OVER_OFFGRID")
	err := switcher.RoutePacket(id2, packet)
	if err != nil {
		t.Fatalf("RoutePacket error: %v", err)
	}

	select {
	case rxData := <-dataReceivedCh:
		if string(rxData) != string(packet) {
			t.Fatalf("expected %s, got %s", packet, rxData)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for physical packet reception")
	}

	stats := switcher.Stats()
	if stats.MeshPacketsSent != 1 {
		t.Errorf("expected 1 mesh packet sent, got %d", stats.MeshPacketsSent)
	}

	// 2. Simulate WAN loss (blackout)
	switcher.ReportWANStatus(false)

	select {
	case newMode := <-modeChangeCh:
		if newMode != ModeOffGridPhysical {
			t.Fatalf("expected ModeOffGridPhysical, got %v", newMode)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for mode transition callback")
	}

	if switcher.CurrentMode() != ModeOffGridPhysical {
		t.Fatalf("expected ModeOffGridPhysical after WAN loss, got %v", switcher.CurrentMode())
	}

	// In off-grid mode, trying to reach an unknown peer (not in radio range) fails cleanly
	unknownID, _, _ := core.GenerateIdentity()
	err = switcher.RoutePacket(unknownID, []byte("CANNOT_REACH"))
	if err != ErrPeerUnreachable {
		t.Fatalf("expected ErrPeerUnreachable, got %v", err)
	}

	// 3. Restore WAN connectivity
	switcher.ReportWANStatus(true)
	select {
	case newMode := <-modeChangeCh:
		if newMode != ModeHybrid {
			t.Fatalf("expected ModeHybrid after WAN restored with physical link, got %v", newMode)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for mode transition callback to Hybrid")
	}

	finalStats := switcher.Stats()
	if finalStats.ModeTransitions != 2 {
		t.Errorf("expected 2 mode transitions, got %d", finalStats.ModeTransitions)
	}
	if finalStats.PacketsDropped != 1 {
		t.Errorf("expected 1 dropped packet, got %d", finalStats.PacketsDropped)
	}
}

func BenchmarkHybridSwitcher_RouteDirectPhysical(b *testing.B) {
	id1, _, _ := core.GenerateIdentity()
	id2, _, _ := core.GenerateIdentity()
	node1 := &core.Node{Identity: id1}

	medium := NewRadioMedium()
	link1 := NewVirtualRadioLink("radio-1", "phy-1", 1280, medium)
	link2 := NewVirtualRadioLink("radio-2", "phy-2", 1280, medium)
	defer link1.Close()
	defer link2.Close()

	// Drain link2 in background
	go func() {
		for {
			_, _, err := link2.Receive()
			if err != nil {
				return
			}
		}
	}()

	b1 := NewBeaconEngine(node1, link1, nil)
	b1.peers[id2.String()] = PhysicalPeer{
		DID:       id2,
		LocalAddr: "phy-2",
		RSSI:      -50,
		LastSeen:  time.Now(),
	}

	switcher := NewHybridSwitcher(node1, b1, link1, SwitcherConfig{
		AutoFallback: false,
		DefaultMode:  ModeHybrid,
	})

	packet := []byte("DATA_BENCHMARK_PAYLOAD")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = switcher.RoutePacket(id2, packet)
	}
}
