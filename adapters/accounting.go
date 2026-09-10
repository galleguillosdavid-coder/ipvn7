package adapters

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// TransitTier defines the prioritization level based on peer cooperation (Tit-for-Tat).
type TransitTier string

const (
	TierPriority   TransitTier = "PRIORITY"    // Highly cooperative peers (Ratio >= 1.2)
	TierNormal     TransitTier = "NORMAL"      // Fair peers (0.8 <= Ratio < 1.2)
	TierBestEffort TransitTier = "BEST_EFFORT" // Under-contributing peers (0.3 <= Ratio < 0.8)
	TierThrottled  TransitTier = "THROTTLED"   // Free-riders / Parasitic peers (Ratio < 0.3)
)

// TransitAccount tracks bandwidth reciprocity and credits for a sovereign DID.
type TransitAccount struct {
	DID              string      `json:"did"`
	BytesRelayedFor  uint64      `json:"bytes_relayed_for"`  // Data we routed on their behalf
	BytesRelayedBy   uint64      `json:"bytes_relayed_by"`   // Data they routed on our behalf
	CreditBalance    int64       `json:"credit_balance"`     // RelayedBy - RelayedFor (positive = surplus)
	LastActivity     time.Time   `json:"last_activity"`
	Tier             TransitTier `json:"tier"`
	ReciprocityRatio float64     `json:"reciprocity_ratio"`
}

// AccountingEngine manages decentralized transit reciprocity, preventing free-riders
// across IPv7 DERP relay nodes and cascade streaming trees.
type AccountingEngine struct {
	mu            sync.RWMutex
	accounts      map[string]*TransitAccount
	storagePath   string
	minVolumeGate uint64 // Minimum bytes exchanged before applying throttling (grace volume)
}

// NewAccountingEngine creates a new transit accounting engine.
func NewAccountingEngine(storagePath string, graceVolume uint64) *AccountingEngine {
	if graceVolume == 0 {
		graceVolume = 10 * 1024 * 1024 // 10 MB default grace volume for new peers
	}
	engine := &AccountingEngine{
		accounts:      make(map[string]*TransitAccount),
		storagePath:   storagePath,
		minVolumeGate: graceVolume,
	}
	if storagePath != "" {
		_ = engine.LoadFromFile(storagePath)
	}
	return engine
}

// RecordTransit updates bytes forwarded on behalf of or received from a peer DID.
func (e *AccountingEngine) RecordTransit(peerDID string, bytesForPeer uint64, bytesByPeer uint64) TransitTier {
	e.mu.Lock()
	defer e.mu.Unlock()

	acc, exists := e.accounts[peerDID]
	if !exists {
		acc = &TransitAccount{
			DID:          peerDID,
			Tier:         TierNormal,
			LastActivity: time.Now(),
		}
		e.accounts[peerDID] = acc
	}

	acc.BytesRelayedFor += bytesForPeer
	acc.BytesRelayedBy += bytesByPeer
	acc.CreditBalance = int64(acc.BytesRelayedBy) - int64(acc.BytesRelayedFor)
	acc.LastActivity = time.Now()

	// Compute reciprocity ratio
	totalConsumed := acc.BytesRelayedFor
	totalContributed := acc.BytesRelayedBy

	if totalConsumed == 0 {
		if totalContributed > 0 {
			acc.ReciprocityRatio = 999.0 // Infinite surplus
			acc.Tier = TierPriority
		} else {
			acc.ReciprocityRatio = 1.0
			acc.Tier = TierNormal
		}
	} else {
		acc.ReciprocityRatio = float64(totalContributed) / float64(totalConsumed)

		// Apply grace volume logic
		if totalConsumed < e.minVolumeGate {
			// In grace period: don't throttle new nodes prematurely
			if acc.ReciprocityRatio >= 1.2 {
				acc.Tier = TierPriority
			} else {
				acc.Tier = TierNormal
			}
		} else {
			// Strict Tit-for-Tat enforcement
			if acc.ReciprocityRatio >= 1.2 {
				acc.Tier = TierPriority
			} else if acc.ReciprocityRatio >= 0.8 {
				acc.Tier = TierNormal
			} else if acc.ReciprocityRatio >= 0.3 {
				acc.Tier = TierBestEffort
			} else {
				acc.Tier = TierThrottled
			}
		}
	}

	return acc.Tier
}

// GetAccount retrieves the accounting record and tier for a given DID.
func (e *AccountingEngine) GetAccount(peerDID string) (TransitAccount, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	acc, exists := e.accounts[peerDID]
	if !exists {
		return TransitAccount{DID: peerDID, Tier: TierNormal, ReciprocityRatio: 1.0}, false
	}
	return *acc, true
}

// GetTier returns the active QoS tier for a peer.
func (e *AccountingEngine) GetTier(peerDID string) TransitTier {
	e.mu.RLock()
	defer e.mu.RUnlock()

	acc, exists := e.accounts[peerDID]
	if !exists {
		return TierNormal
	}
	return acc.Tier
}

// ListAccounts returns all tracked peer accounts.
func (e *AccountingEngine) ListAccounts() []TransitAccount {
	e.mu.RLock()
	defer e.mu.RUnlock()

	out := make([]TransitAccount, 0, len(e.accounts))
	for _, acc := range e.accounts {
		out = append(out, *acc)
	}
	return out
}

// SaveToFile exports the transit accounting state to a JSON file.
func (e *AccountingEngine) SaveToFile(path string) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	data, err := json.MarshalIndent(e.accounts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// LoadFromFile imports the transit accounting state from a JSON file.
func (e *AccountingEngine) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var raw map[string]*TransitAccount
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	e.accounts = raw
	e.storagePath = path
	return nil
}

// FormatTraffic converts byte counts into human-readable strings (KB, MB, GB).
func FormatTraffic(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
