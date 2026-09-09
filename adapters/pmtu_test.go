package adapters

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestPMTUInitialDefaults(t *testing.T) {
	pmtu := NewPMTUDiscovery(200*time.Millisecond, nil)
	ep := "192.168.1.50:7001"

	mtu := pmtu.GetMTU(ep)
	if mtu != DefaultSafeMTU {
		t.Fatalf("Expected default MTU %d, got %d", DefaultSafeMTU, mtu)
	}
}

func TestPMTUConfirmProbeSize(t *testing.T) {
	var notifiedOld, notifiedNew int32
	onChange := func(endpoint string, oldMTU, newMTU int) {
		atomic.StoreInt32(&notifiedOld, int32(oldMTU))
		atomic.StoreInt32(&notifiedNew, int32(newMTU))
	}

	pmtu := NewPMTUDiscovery(200*time.Millisecond, onChange)
	ep := "203.0.113.10:7001"
	pmtu.RegisterEndpoint(ep)

	// Confirm 1472
	pmtu.ConfirmProbeSize(ep, 1472)
	time.Sleep(50 * time.Millisecond)

	if got := pmtu.GetMTU(ep); got != 1472 {
		t.Fatalf("Expected MTU 1472, got %d", got)
	}

	if atomic.LoadInt32(&notifiedNew) != 1472 {
		t.Fatalf("Expected notification of new MTU 1472, got %d", atomic.LoadInt32(&notifiedNew))
	}
}

func TestPMTUBlackHoleStepDown(t *testing.T) {
	pmtu := NewPMTUDiscovery(200*time.Millisecond, nil)
	ep := "198.51.100.20:7001"
	pmtu.RegisterEndpoint(ep)

	// Simulate black-hole dropping 1500 and 1472
	newMTU := pmtu.RecordBlackHoleTimeout(ep, 1500)
	if newMTU != 1472 {
		t.Fatalf("Expected step down to 1472, got %d", newMTU)
	}

	newMTU = pmtu.RecordBlackHoleTimeout(ep, 1472)
	if newMTU != 1400 {
		t.Fatalf("Expected step down to 1400, got %d", newMTU)
	}

	newMTU = pmtu.RecordBlackHoleTimeout(ep, 1400)
	if newMTU != 1280 {
		t.Fatalf("Expected step down to 1280, got %d", newMTU)
	}
}

func TestPMTUDiscoverySequence(t *testing.T) {
	pmtu := NewPMTUDiscovery(50*time.Millisecond, nil)
	ep := "10.0.0.5:7001"

	// Mock sender: passes only for sizes <= 1400
	mockSender := func(ctx context.Context, endpoint string, size int) error {
		if size > 1400 {
			return errors.New("packet dropped silently (black-hole)")
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	discovered, err := pmtu.DiscoverPathMTU(ctx, ep, mockSender)
	if err != nil {
		t.Fatalf("Unexpected discovery error: %v", err)
	}

	if discovered != 1400 {
		t.Fatalf("Expected discovered MTU to be 1400, got %d", discovered)
	}

	if final := pmtu.GetMTU(ep); final != 1400 {
		t.Fatalf("Expected stored MTU to be 1400, got %d", final)
	}
}
