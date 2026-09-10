package dht

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestPetnameStoreBasic(t *testing.T) {
	store := NewPetnameStore("", false)

	aliceDID := "did:ipv7:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	bobDID := "did:ipv7:fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210"

	// Add Alice
	err := store.Set("alice", aliceDID, "Primary developer workstation")
	if err != nil {
		t.Fatalf("failed to add alice: %v", err)
	}

	// Add Bob with .ipv7 suffix -> should normalize cleanly
	err = store.Set("bob.ipv7", bobDID, "Remote cloud relay node")
	if err != nil {
		t.Fatalf("failed to add bob: %v", err)
	}

	// Resolve alice
	resolvedAlice, ok := store.Resolve("alice")
	if !ok || resolvedAlice != aliceDID {
		t.Fatalf("expected alice DID to resolve, got %s (ok=%v)", resolvedAlice, ok)
	}

	// Resolve bob using plain or .ipv7
	resolvedBob, ok := store.Resolve("bob")
	if !ok || resolvedBob != bobDID {
		t.Fatalf("expected bob to resolve from 'bob', got %s", resolvedBob)
	}
	resolvedBobSuffix, ok := store.Resolve("bob.ipv7")
	if !ok || resolvedBobSuffix != bobDID {
		t.Fatalf("expected bob to resolve from 'bob.ipv7', got %s", resolvedBobSuffix)
	}

	// Reverse lookup
	name, ok := store.ReverseLookup(aliceDID)
	if !ok || name != "alice" {
		t.Fatalf("expected reverse lookup of alice DID to be 'alice', got '%s'", name)
	}

	// List entries
	list := store.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 entries in list, got %d", len(list))
	}

	// Delete
	deleted := store.Delete("alice")
	if !deleted {
		t.Fatalf("expected alice to be deleted")
	}
	_, ok = store.Resolve("alice")
	if ok {
		t.Fatalf("expected alice to be missing after deletion")
	}
}

func TestPetnameValidation(t *testing.T) {
	store := NewPetnameStore("", false)
	validDID := "did:ipv7:11112222333344445555666677778888"

	// Too short
	if err := store.Set("a", validDID, ""); err == nil {
		t.Fatalf("expected error for 1-char petname")
	}

	// Invalid characters (spaces, symbols)
	if err := store.Set("alice node!", validDID, ""); err == nil {
		t.Fatalf("expected error for invalid characters")
	}

	// Bad DID format
	if err := store.Set("server", "192.168.1.1", ""); err == nil {
		t.Fatalf("expected error for non-DID format")
	}
}

func TestPetnameVirtualIPAndHosts(t *testing.T) {
	store := NewPetnameStore("", false)
	did := "did:ipv7:abcdef1234567890abcdef1234567890"

	_ = store.Set("mynas", did, "Home storage")

	ipv4 := DeriveVirtualIPv4(did)
	if !strings.HasPrefix(ipv4, "10.7.") {
		t.Fatalf("expected IPv4 to start with 10.7., got %s", ipv4)
	}

	ipv6 := DeriveVirtualIPv6(did)
	if !strings.HasPrefix(ipv6, "fd07::") {
		t.Fatalf("expected IPv6 to start with fd07::, got %s", ipv6)
	}

	hosts := store.ExportHostsFile(false)
	if !strings.Contains(hosts, "mynas.ipv7 mynas") {
		t.Fatalf("expected hosts export to contain 'mynas.ipv7 mynas', got:\n%s", hosts)
	}
}

func TestPetnamePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "contacts.json")

	store1 := NewPetnameStore(filePath, true)
	did := "did:ipv7:persistent_alias_001"
	_ = store1.Set("workstation", did, "Office PC")

	// Store2 loads from same path
	store2 := NewPetnameStore(filePath, false)
	resolved, ok := store2.Resolve("workstation")
	if !ok || resolved != did {
		t.Fatalf("expected store2 to load persisted entry, got resolved=%s, ok=%v", resolved, ok)
	}
}

func TestPetnamePoW(t *testing.T) {
	name := "coolservice"
	ownerDID := "did:ipv7:99887766554433221100"
	targetZeros := 2 // light PoW for fast test

	nonce, hash := CalculateNamePoW(name, ownerDID, targetZeros)
	if !strings.HasPrefix(hash, "00") {
		t.Fatalf("expected hash to start with '00', got %s", hash)
	}

	valid := VerifyNamePoW(name, ownerDID, nonce, targetZeros)
	if !valid {
		t.Fatalf("expected PoW verification to succeed")
	}

	// Tampered name must fail
	tampered := VerifyNamePoW("coolservice2", ownerDID, nonce, targetZeros)
	if tampered {
		t.Fatalf("expected tampered PoW verification to fail")
	}
}
