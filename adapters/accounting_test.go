package adapters

import (
	"path/filepath"
	"testing"
)

func TestAccountingEngineTitForTat(t *testing.T) {
	// Grace volume: 1 MB for testing
	engine := NewAccountingEngine("", 1024*1024)

	generousDID := "did:ipv7:generous_peer"
	parasiteDID := "did:ipv7:parasite_peer"

	// 1. Generous peer: relays 5 MB for us, consumes only 1 MB
	tierGen := engine.RecordTransit(generousDID, 1024*1024, 5*1024*1024)
	if tierGen != TierPriority {
		t.Fatalf("expected generous peer to get TierPriority, got %s", tierGen)
	}

	accGen, _ := engine.GetAccount(generousDID)
	if accGen.CreditBalance <= 0 || accGen.ReciprocityRatio < 1.2 {
		t.Fatalf("expected positive credit and high ratio, got balance=%d, ratio=%.2f", accGen.CreditBalance, accGen.ReciprocityRatio)
	}

	// 2. Parasitic peer in grace volume: consumes 500 KB, gives 0 -> should remain Normal in grace period
	tierGrace := engine.RecordTransit(parasiteDID, 500*1024, 0)
	if tierGrace != TierNormal {
		t.Fatalf("expected peer in grace period to be TierNormal, got %s", tierGrace)
	}

	// 3. Parasitic peer past grace volume: consumes 20 MB total, gives 0 -> should drop to THROTTLED
	tierThrottled := engine.RecordTransit(parasiteDID, 20*1024*1024, 0)
	if tierThrottled != TierThrottled {
		t.Fatalf("expected free-rider to be THROTTLED, got %s", tierThrottled)
	}

	accPara, _ := engine.GetAccount(parasiteDID)
	if accPara.CreditBalance >= 0 || accPara.ReciprocityRatio > 0.1 {
		t.Fatalf("expected heavy negative balance and low ratio, got balance=%d, ratio=%.2f", accPara.CreditBalance, accPara.ReciprocityRatio)
	}
}

func TestAccountingEnginePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "transit_accounting.json")

	engine1 := NewAccountingEngine(path, 1024)
	did := "did:ipv7:saved_peer"
	engine1.RecordTransit(did, 1000, 2000)

	if err := engine1.SaveToFile(path); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	engine2 := NewAccountingEngine(path, 1024)
	acc, ok := engine2.GetAccount(did)
	if !ok {
		t.Fatalf("expected peer account to be loaded")
	}
	if acc.BytesRelayedBy != 2000 || acc.BytesRelayedFor != 1000 {
		t.Fatalf("mismatched bytes: by=%d, for=%d", acc.BytesRelayedBy, acc.BytesRelayedFor)
	}
}

func TestFormatTraffic(t *testing.T) {
	cases := []struct {
		bytes    uint64
		expected string
	}{
		{500, "500 B"},
		{2048, "2.00 KB"},
		{10 * 1024 * 1024, "10.00 MB"},
		{5 * 1024 * 1024 * 1024, "5.00 GB"},
	}

	for _, c := range cases {
		out := FormatTraffic(c.bytes)
		if out != c.expected {
			t.Fatalf("expected %s for %d bytes, got %s", c.expected, c.bytes, out)
		}
	}
}
