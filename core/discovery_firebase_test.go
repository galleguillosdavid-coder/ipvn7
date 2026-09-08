package core

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFirebaseDiscoveryAnnounceAndFetch(t *testing.T) {
	storedPeers := make(map[string]FirebasePeerRecord)

	// Mock Firebase Realtime Database HTTP Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPut {
			var rec FirebasePeerRecord
			if err := json.NewDecoder(r.Body).Decode(&rec); err == nil {
				storedPeers[rec.ID] = rec
			}
			_ = json.NewEncoder(w).Encode(rec)
			return
		}
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(storedPeers)
			return
		}
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("null"))
			return
		}
	}))
	defer server.Close()

	// Create test node
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	node := NewNode(&Ed25519Identity{PublicKey: pub}, priv)
	node.SetEndpoints([]string{"192.168.1.100:7001", "200.1.2.3:7001"})

	discovery := NewFirebaseDiscovery(server.URL, node, 7001)

	// 1. Announce
	if err := discovery.Announce(); err != nil {
		t.Fatalf("Announce failed: %v", err)
	}

	// 2. Fetch peers
	peers, err := discovery.FetchPeers()
	if err != nil {
		t.Fatalf("FetchPeers failed: %v", err)
	}
	if len(peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(peers))
	}
	if peers[0].ID != node.Identity.String() {
		t.Fatalf("peer ID mismatch: got %s, want %s", peers[0].ID, node.Identity.String())
	}
	if len(peers[0].Endpoints) != 2 {
		t.Fatalf("expected 2 endpoints, got %d", len(peers[0].Endpoints))
	}

	// 3. Deregister
	discovery.Deregister()
}
