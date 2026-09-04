package adapters

import (
	"testing"
)

func TestGetLocalEndpoints(t *testing.T) {
	endpoints, err := GetLocalEndpoints(9000)
	if err != nil {
		t.Fatalf("Failed to get local endpoints: %v", err)
	}
	t.Logf("Discovered local endpoints: %v", endpoints)
}

func TestDiscoverPublicEndpoint(t *testing.T) {
	// Query Google's public STUN server
	ep, err := DiscoverPublicEndpoint("stun.l.google.com:19302")
	if err != nil {
		t.Logf("STUN discovery failed (possibly offline or firewall): %v", err)
		return
	}
	if ep == "" {
		t.Fatalf("Discovered empty public endpoint")
	}
	t.Logf("Discovered public reflexive endpoint: %s", ep)
}

func TestUDPAdapterDiscoverEndpoints(t *testing.T) {
	adapter, err := NewUDPAdapter("127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}
	if err := adapter.Start(); err != nil {
		t.Fatalf("Failed to start adapter: %v", err)
	}
	defer adapter.Stop()

	// Discover endpoints (local + STUN fallback if reachable)
	_ = adapter.DiscoverEndpoints("stun.l.google.com:19302")
	eps := adapter.Endpoints()
	t.Logf("Adapter known endpoints: %v", eps)
}
