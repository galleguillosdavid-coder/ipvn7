package adapters

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestZTNAFirewallDefaultDeny(t *testing.T) {
	fw := NewZTNAFirewall(ActionDeny)

	allowed, action, _ := fw.Inspect("did:ipv7:attacker123", "tcp", 8080)
	if allowed || action != ActionDeny {
		t.Fatalf("expected packet to be DENIED by default, got action=%s, allowed=%v", action, allowed)
	}

	statsAllowed, statsDropped, _ := fw.Stats()
	if statsAllowed != 0 || statsDropped != 1 {
		t.Fatalf("expected 0 allowed and 1 dropped, got allowed=%d, dropped=%d", statsAllowed, statsDropped)
	}
}

func TestZTNAFirewallExplicitAllow(t *testing.T) {
	fw := NewZTNAFirewall(ActionDeny)

	trustedDID := "did:ipv7:trusted_peer_001"
	err := fw.AddRule(FirewallRule{
		ID:          "rule_allow_ssh",
		SourceDID:   trustedDID,
		Protocol:    "tcp",
		DestPort:    22,
		Action:      ActionAllow,
		Description: "Allow trusted peer to access SSH port 22",
	})
	if err != nil {
		t.Fatalf("failed to add rule: %v", err)
	}

	// Allowed check
	allowed, action, reason := fw.Inspect(trustedDID, "tcp", 22)
	if !allowed || action != ActionAllow {
		t.Fatalf("expected trusted peer SSH to be allowed, got action=%s (reason: %s)", action, reason)
	}

	// Denied check: different port
	allowedWrongPort, _, _ := fw.Inspect(trustedDID, "tcp", 8080)
	if allowedWrongPort {
		t.Fatalf("expected trusted peer port 8080 to be denied, but was allowed")
	}

	// Denied check: untrusted peer on SSH port
	allowedUntrusted, _, _ := fw.Inspect("did:ipv7:unknown_peer", "tcp", 22)
	if allowedUntrusted {
		t.Fatalf("expected untrusted peer SSH to be denied, but was allowed")
	}
}

func TestZTNAFirewallWildcardAndExpiry(t *testing.T) {
	fw := NewZTNAFirewall(ActionDeny)

	// Add an expired rule
	past := time.Now().Add(-1 * time.Hour)
	err := fw.AddRule(FirewallRule{
		ID:          "expired_rule",
		SourceDID:   "*",
		Protocol:    "*",
		DestPort:    53,
		Action:      ActionAllow,
		ExpiresAt:   &past,
		Description: "Expired DNS rule",
	})
	if err != nil {
		t.Fatalf("failed to add rule: %v", err)
	}

	// Should NOT match expired rule, falling back to default deny
	allowed, action, _ := fw.Inspect("did:ipv7:any", "udp", 53)
	if allowed || action != ActionDeny {
		t.Fatalf("expected expired rule to be ignored, resulting in DENY")
	}

	// Add an active wildcard rule for ICMP/ping
	future := time.Now().Add(1 * time.Hour)
	_ = fw.AddRule(FirewallRule{
		ID:          "active_wildcard",
		SourceDID:   "*",
		Protocol:    "icmp",
		DestPort:    0,
		Action:      ActionAllow,
		ExpiresAt:   &future,
		Description: "Allow ping from anywhere",
	})

	allowedPing, actionPing, _ := fw.Inspect("did:ipv7:random_node", "icmp", 0)
	if !allowedPing || actionPing != ActionAllow {
		t.Fatalf("expected ping to be allowed by wildcard rule, got action=%s", actionPing)
	}
}

func TestZTNAFirewallRemoveRule(t *testing.T) {
	fw := NewZTNAFirewall(ActionDeny)
	targetDID := "did:ipv7:peer_to_revoke"

	_ = fw.AddRule(FirewallRule{
		ID:        "revocable_rule",
		SourceDID: targetDID,
		Protocol:  "*",
		DestPort:  80,
		Action:    ActionAllow,
	})

	// Inspect before removal -> ALLOW
	allowed, _, _ := fw.Inspect(targetDID, "tcp", 80)
	if !allowed {
		t.Fatalf("expected allowed before removal")
	}

	// Remove rule
	removed := fw.RemoveRule("revocable_rule")
	if !removed {
		t.Fatalf("expected rule to be removed")
	}

	// Inspect after removal -> DENY
	allowedAfter, _, _ := fw.Inspect(targetDID, "tcp", 80)
	if allowedAfter {
		t.Fatalf("expected denied after rule removal")
	}
}

func TestZTNAFirewallTokenBucketRateLimit(t *testing.T) {
	fw := NewZTNAFirewall(ActionAllow)
	did := "did:ipv7:burst_tester"

	// Allow burst of 3 tokens, refill 1 token/sec
	burst := 3.0
	refillRate := 1.0

	// 1st request -> pass
	if !fw.AllowRateLimited(did, burst, refillRate) {
		t.Fatalf("1st request should pass")
	}
	// 2nd request -> pass
	if !fw.AllowRateLimited(did, burst, refillRate) {
		t.Fatalf("2nd request should pass")
	}
	// 3rd request -> pass
	if !fw.AllowRateLimited(did, burst, refillRate) {
		t.Fatalf("3rd request should pass")
	}
	// 4th immediate request -> should be rate limited!
	if fw.AllowRateLimited(did, burst, refillRate) {
		t.Fatalf("4th immediate request should be rate limited (blocked)")
	}
}

func TestZTNAFirewallPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "firewall_rules.json")

	fw1 := NewZTNAFirewall(ActionDeny)
	_ = fw1.AddRule(FirewallRule{
		ID:          "rule_persisted",
		SourceDID:   "did:ipv7:persisted_node",
		Protocol:    "udp",
		DestPort:    9000,
		Action:      ActionAllow,
		Description: "Persisted node rule",
	})

	err := fw1.SaveToFile(filePath)
	if err != nil {
		t.Fatalf("failed to save rules: %v", err)
	}

	fw2 := NewZTNAFirewall(ActionDeny)
	err = fw2.LoadFromFile(filePath)
	if err != nil {
		t.Fatalf("failed to load rules: %v", err)
	}

	allowed, _, _ := fw2.Inspect("did:ipv7:persisted_node", "udp", 9000)
	if !allowed {
		t.Fatalf("expected persisted rule to be loaded and active")
	}

	// Clean up
	_ = os.Remove(filePath)
}
