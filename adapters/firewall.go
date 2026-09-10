package adapters

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// FirewallAction defines the decision for a network packet.
type FirewallAction string

const (
	ActionAllow FirewallAction = "ALLOW"
	ActionDeny  FirewallAction = "DENY"
)

// FirewallRule specifies an access control rule based on sovereign DID and port.
type FirewallRule struct {
	ID          string         `json:"id"`
	SourceDID   string         `json:"source_did"`   // "*" for any DID, or specific "did:ipv7:..."
	Protocol    string         `json:"protocol"`     // "tcp", "udp", or "*"
	DestPort    uint16         `json:"dest_port"`    // 0 for any port
	Action      FirewallAction `json:"action"`       // ALLOW or DENY
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	ExpiresAt   *time.Time     `json:"expires_at,omitempty"`
}

// ZTNAFirewall provides zero-trust packet micro-segmentation and access control
// for the IPv7 Network Operating System (NOS).
type ZTNAFirewall struct {
	mu            sync.RWMutex
	defaultAction FirewallAction
	rules         map[string]FirewallRule // key: rule.ID
	didIndex      map[string][]string     // source_did -> slice of rule IDs
	rateLimits    map[string]*tokenBucket // per-DID rate limiter
	rateLimitMu   sync.Mutex
	blockedDrops  uint64
	allowedPasses uint64
}

type tokenBucket struct {
	tokens         float64
	maxTokens      float64
	refillRate     float64 // tokens per second
	lastRefillTime time.Time
}

// NewZTNAFirewall creates a new zero-trust firewall with the given default action (typically ActionDeny).
func NewZTNAFirewall(defaultAction FirewallAction) *ZTNAFirewall {
	return &ZTNAFirewall{
		defaultAction: defaultAction,
		rules:         make(map[string]FirewallRule),
		didIndex:      make(map[string][]string),
		rateLimits:    make(map[string]*tokenBucket),
	}
}

// SetDefaultAction changes the fallback action when no rules match.
func (f *ZTNAFirewall) SetDefaultAction(action FirewallAction) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.defaultAction = action
}

// AddRule registers a new access control rule.
func (f *ZTNAFirewall) AddRule(rule FirewallRule) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if rule.ID == "" {
		rule.ID = fmt.Sprintf("rule_%d", time.Now().UnixNano())
	}
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now()
	}
	if rule.Action != ActionAllow && rule.Action != ActionDeny {
		return fmt.Errorf("invalid firewall action: %s", rule.Action)
	}

	normDID := strings.TrimSpace(rule.SourceDID)
	rule.SourceDID = normDID
	rule.Protocol = strings.ToLower(strings.TrimSpace(rule.Protocol))

	f.rules[rule.ID] = rule
	f.didIndex[normDID] = append(f.didIndex[normDID], rule.ID)
	return nil
}

// RemoveRule removes a rule by its unique ID.
func (f *ZTNAFirewall) RemoveRule(ruleID string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	rule, exists := f.rules[ruleID]
	if !exists {
		return false
	}

	delete(f.rules, ruleID)
	// Remove from didIndex
	if ids, ok := f.didIndex[rule.SourceDID]; ok {
		var updated []string
		for _, id := range ids {
			if id != ruleID {
				updated = append(updated, id)
			}
		}
		if len(updated) == 0 {
			delete(f.didIndex, rule.SourceDID)
		} else {
			f.didIndex[rule.SourceDID] = updated
		}
	}
	return true
}

// Inspect evaluates whether incoming traffic from sourceDID to targetPort is permitted.
func (f *ZTNAFirewall) Inspect(sourceDID string, protocol string, targetPort uint16) (bool, FirewallAction, string) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	now := time.Now()
	proto := strings.ToLower(protocol)

	// Check explicit rules for this specific DID, and then wildcard "*"
	candidates := append([]string{}, f.didIndex[sourceDID]...)
	candidates = append(candidates, f.didIndex["*"]...)

	for _, ruleID := range candidates {
		rule, ok := f.rules[ruleID]
		if !ok {
			continue
		}

		// Check expiration
		if rule.ExpiresAt != nil && now.After(*rule.ExpiresAt) {
			continue
		}

		// Check protocol match
		if rule.Protocol != "*" && rule.Protocol != proto {
			continue
		}

		// Check port match
		if rule.DestPort != 0 && rule.DestPort != targetPort {
			continue
		}

		// Matched rule!
		if rule.Action == ActionAllow {
			f.mu.RUnlock()
			f.mu.Lock()
			f.allowedPasses++
			f.mu.Unlock()
			f.mu.RLock()
			return true, ActionAllow, fmt.Sprintf("matched rule %s (%s)", rule.ID, rule.Description)
		}
		f.mu.RUnlock()
		f.mu.Lock()
		f.blockedDrops++
		f.mu.Unlock()
		f.mu.RLock()
		return false, ActionDeny, fmt.Sprintf("matched rule %s (%s)", rule.ID, rule.Description)
	}

	// Fallback to default action
	if f.defaultAction == ActionAllow {
		f.mu.RUnlock()
		f.mu.Lock()
		f.allowedPasses++
		f.mu.Unlock()
		f.mu.RLock()
		return true, ActionAllow, "default policy (ALLOW)"
	}

	f.mu.RUnlock()
	f.mu.Lock()
	f.blockedDrops++
	f.mu.Unlock()
	f.mu.RLock()
	return false, ActionDeny, "default policy (DENY)"
}

// AllowRateLimited checks per-DID Token Bucket rate limiting.
func (f *ZTNAFirewall) AllowRateLimited(sourceDID string, maxBurst float64, refillRate float64) bool {
	f.rateLimitMu.Lock()
	defer f.rateLimitMu.Unlock()

	now := time.Now()
	tb, exists := f.rateLimits[sourceDID]
	if !exists {
		tb = &tokenBucket{
			tokens:         maxBurst - 1,
			maxTokens:      maxBurst,
			refillRate:     refillRate,
			lastRefillTime: now,
		}
		f.rateLimits[sourceDID] = tb
		return true
	}

	// Refill tokens
	elapsed := now.Sub(tb.lastRefillTime).Seconds()
	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}
	tb.lastRefillTime = now

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}
	return false
}

// Stats returns the total allowed passes and blocked drops.
func (f *ZTNAFirewall) Stats() (allowed uint64, dropped uint64, totalRules int) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.allowedPasses, f.blockedDrops, len(f.rules)
}

// ListRules returns a copy of all current firewall rules.
func (f *ZTNAFirewall) ListRules() []FirewallRule {
	f.mu.RLock()
	defer f.mu.RUnlock()
	out := make([]FirewallRule, 0, len(f.rules))
	for _, r := range f.rules {
		out = append(out, r)
	}
	return out
}

// SaveToFile exports the current firewall ruleset to a JSON file.
func (f *ZTNAFirewall) SaveToFile(path string) error {
	f.mu.RLock()
	rules := f.ListRules()
	f.mu.RUnlock()

	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// LoadFromFile imports firewall rules from a JSON file.
func (f *ZTNAFirewall) LoadFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	var rules []FirewallRule
	if err := json.Unmarshal(bytes, &rules); err != nil {
		return err
	}

	for _, r := range rules {
		_ = f.AddRule(r)
	}
	return nil
}
