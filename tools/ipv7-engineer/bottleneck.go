package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BottleneckStatus reflects the lifecycle of a detected architectural bottleneck
type BottleneckStatus string

const (
	StatusUnknown            BottleneckStatus = "UNKNOWN"
	StatusSuspected          BottleneckStatus = "SUSPECTED"
	StatusUnderInvestigation BottleneckStatus = "UNDER_INVESTIGATION"
	StatusConfirmed          BottleneckStatus = "CONFIRMED"
	StatusMitigated          BottleneckStatus = "MITIGATED"
	StatusResolved           BottleneckStatus = "RESOLVED"
	StatusReopened           BottleneckStatus = "REOPENED"
)

// BottleneckRecord encapsulates a documented performance bottleneck
type BottleneckRecord struct {
	ID          string           `json:"id"`
	Status      BottleneckStatus `json:"status"`
	Component   string           `json:"component"` // "SOCKET", "CRYPTO", "ROUTING", "RELAY", "MEM"
	Symptom     string           `json:"symptom"`
	Evidence    string           `json:"evidence"`
	Cause       string           `json:"cause"`
	Confidence  string           `json:"confidence"` // "LOW", "MEDIUM", "HIGH"
	Experiments []string         `json:"experiments"`
	Mitigation  string           `json:"mitigation"`
	LastUpdated string           `json:"last_updated"`
}

// FailedHypothesis preserves false theories so the system never re-proposes them blindly
type FailedHypothesis struct {
	ID         string `json:"id"`
	Theory     string `json:"theory"`
	Experiment string `json:"experiment"`
	Result     string `json:"result"`
	Conclusion string `json:"conclusion"`
	Timestamp  string `json:"timestamp"`
}

// BottleneckManager handles bottleneck catalog and failed hypothesis memory
type BottleneckManager struct {
	dir string
}

// NewBottleneckManager initializes manager with disk persistence
func NewBottleneckManager(dir string) *BottleneckManager {
	if dir == "" {
		dir = filepath.Join("tools", "ipv7-engineer", "bottlenecks")
	}
	_ = os.MkdirAll(dir, 0755)
	return &BottleneckManager{dir: dir}
}

// RegisterBottleneck records or updates a bottleneck definition
func (bm *BottleneckManager) RegisterBottleneck(b BottleneckRecord) error {
	b.LastUpdated = time.Now().UTC().Format(time.RFC3339)
	filePath := filepath.Join(bm.dir, fmt.Sprintf("%s.json", b.ID))
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// RecordFailedHypothesis records a falsified theory to prevent repeated regressions
func (bm *BottleneckManager) RecordFailedHypothesis(h FailedHypothesis) error {
	h.Timestamp = time.Now().UTC().Format(time.RFC3339)
	filePath := filepath.Join(bm.dir, fmt.Sprintf("FAILED_HYPOTHESIS_%s.json", h.ID))
	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// ListBottlenecks returns all persisted bottleneck records
func (bm *BottleneckManager) ListBottlenecks() ([]BottleneckRecord, error) {
	files, err := os.ReadDir(bm.dir)
	if err != nil {
		return nil, err
	}
	var res []BottleneckRecord
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".json" && !isFailedHypothesis(f.Name()) {
			b, err := os.ReadFile(filepath.Join(bm.dir, f.Name()))
			if err == nil {
				var r BottleneckRecord
				if json.Unmarshal(b, &r) == nil {
					res = append(res, r)
				}
			}
		}
	}
	return res, nil
}

func isFailedHypothesis(name string) bool {
	return len(name) >= 17 && name[:17] == "FAILED_HYPOTHESIS"
}
