package adapters

import (
	"sync"
	"testing"
	"time"
)

func TestMultipath_ActiveStandbyFailover(t *testing.T) {
	var sentPaths []string
	var mu sync.Mutex

	mockSender := func(path *NetworkPath, data []byte) error {
		mu.Lock()
		sentPaths = append(sentPaths, path.ID)
		mu.Unlock()
		return nil
	}

	overlay := NewMultipathOverlay(PolicyActiveStandby, mockSender)

	wifi := &NetworkPath{
		ID:             "path-wifi-0",
		InterfaceName:  "wlan0",
		RemoteEndpoint: "192.168.1.1:7001",
		State:          PathStateActive,
		Weight:         100, // Preferred primary
		LastSeen:       time.Now(),
	}

	cellular := &NetworkPath{
		ID:             "path-lte-1",
		InterfaceName:  "wwan0",
		RemoteEndpoint: "10.15.2.1:7001",
		State:          PathStateStandby,
		Weight:         50,
		LastSeen:       time.Now(),
	}

	overlay.RegisterPath(wifi)
	overlay.RegisterPath(cellular)

	// 1. Send initial packet -> should route to primary (wifi)
	delivered, err := overlay.SendPacket([]byte("packet_01"), false)
	if err != nil || delivered != 1 {
		t.Fatalf("Failed to deliver packet 1: delivered=%d, err=%v", delivered, err)
	}

	mu.Lock()
	if len(sentPaths) != 1 || sentPaths[0] != "path-wifi-0" {
		t.Fatalf("Expected packet routed to path-wifi-0, got: %v", sentPaths)
	}
	mu.Unlock()

	// 2. Simulate Wi-Fi drop: last seen exceeds heartbeat timeout (3 seconds)
	now := time.Now().Add(5 * time.Second)
	failedOver, newPrimary := overlay.FailoverCheck(now)
	if !failedOver || newPrimary != "path-lte-1" {
		t.Fatalf("Expected failover to path-lte-1, got failedOver=%v, newPrimary=%s", failedOver, newPrimary)
	}

	// 3. Next packet should now be transmitted over cellular path seamlessly without error
	delivered, err = overlay.SendPacket([]byte("packet_02"), false)
	if err != nil || delivered != 1 {
		t.Fatalf("Failed to deliver packet 2 post-failover: delivered=%d, err=%v", delivered, err)
	}

	mu.Lock()
	if len(sentPaths) != 2 || sentPaths[1] != "path-lte-1" {
		t.Fatalf("Expected packet 2 routed to path-lte-1, got: %v", sentPaths)
	}
	mu.Unlock()
}

func TestMultipath_LowestRTT(t *testing.T) {
	mockSender := func(path *NetworkPath, data []byte) error {
		return nil
	}

	overlay := NewMultipathOverlay(PolicyLowestRTT, mockSender)

	p1 := &NetworkPath{
		ID:          "p1",
		State:       PathStateActive,
		SmoothedRTT: 50 * time.Millisecond,
		LastSeen:    time.Now(),
	}
	p2 := &NetworkPath{
		ID:          "p2",
		State:       PathStateActive,
		SmoothedRTT: 15 * time.Millisecond, // Lowest latency
		LastSeen:    time.Now(),
	}
	p3 := &NetworkPath{
		ID:          "p3",
		State:       PathStateActive,
		SmoothedRTT: 80 * time.Millisecond,
		LastSeen:    time.Now(),
	}

	overlay.RegisterPath(p1)
	overlay.RegisterPath(p2)
	overlay.RegisterPath(p3)

	selected, err := overlay.SelectPaths(false)
	if err != nil || len(selected) != 1 {
		t.Fatalf("Expected 1 path selected, got %v, err=%v", selected, err)
	}

	if selected[0].ID != "p2" {
		t.Fatalf("Expected lowest RTT path p2, got %s", selected[0].ID)
	}
}

func TestMultipath_PacketDuplication(t *testing.T) {
	var count int
	var mu sync.Mutex

	mockSender := func(path *NetworkPath, data []byte) error {
		mu.Lock()
		count++
		mu.Unlock()
		return nil
	}

	overlay := NewMultipathOverlay(PolicyDuplication, mockSender)

	overlay.RegisterPath(&NetworkPath{ID: "pathA", State: PathStateActive, LastSeen: time.Now()})
	overlay.RegisterPath(&NetworkPath{ID: "pathB", State: PathStateActive, LastSeen: time.Now()})
	overlay.RegisterPath(&NetworkPath{ID: "pathC", State: PathStateActive, LastSeen: time.Now()})

	delivered, err := overlay.SendPacket([]byte("critical_handshake"), true)
	if err != nil {
		t.Fatalf("SendPacket failed: %v", err)
	}

	if delivered != 3 {
		t.Fatalf("Expected packet duplicated across all 3 active paths, got %d", delivered)
	}

	mu.Lock()
	if count != 3 {
		t.Fatalf("Expected mock sender invoked 3 times, got %d", count)
	}
	mu.Unlock()
}
