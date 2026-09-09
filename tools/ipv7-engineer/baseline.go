package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// BaselineRecord stores an immutable benchmark snapshot with strict environment context
type BaselineRecord struct {
	ID             string                 `json:"id"`
	Category       string                 `json:"category"` // "LAN", "WAN", "RELAY", "ROAMING", "CHURN"
	Commit         string                 `json:"commit"`
	Version        string                 `json:"version"`
	OS             string                 `json:"os"`
	Arch           string                 `json:"arch"`
	CPU            string                 `json:"cpu"`
	RAMTotalMB     uint64                 `json:"ram_total_mb"`
	GoVersion      string                 `json:"go_version"`
	Configuration  map[string]interface{} `json:"configuration"`
	TestParameters map[string]interface{} `json:"test_parameters"`
	Timestamp      string                 `json:"timestamp"`
	Metrics        map[string]float64     `json:"metrics"`
	Notes          string                 `json:"notes"`
}

// BaselineManager manages baseline records in disk directory
type BaselineManager struct {
	baseDir string
}

// NewBaselineManager initializes manager pointing to baseline storage directory
func NewBaselineManager(baseDir string) *BaselineManager {
	if baseDir == "" {
		baseDir = filepath.Join("tools", "ipv7-engineer", "baseline")
	}
	_ = os.MkdirAll(baseDir, 0755)
	return &BaselineManager{baseDir: baseDir}
}

// SaveBaseline persists a baseline record into JSON
func (bm *BaselineManager) SaveBaseline(rec BaselineRecord) error {
	if rec.ID == "" {
		rec.ID = fmt.Sprintf("%s_%s_%d", rec.Category, rec.Commit[:7], time.Now().Unix())
	}
	rec.Timestamp = time.Now().UTC().Format(time.RFC3339)
	rec.OS = runtime.GOOS
	rec.Arch = runtime.GOARCH
	rec.GoVersion = runtime.Version()

	filePath := filepath.Join(bm.baseDir, fmt.Sprintf("%s.json", rec.ID))
	b, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, b, 0644)
}

// LoadBaseline loads a specific baseline by ID or Category
func (bm *BaselineManager) LoadBaseline(id string) (*BaselineRecord, error) {
	filePath := filepath.Join(bm.baseDir, fmt.Sprintf("%s.json", id))
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var rec BaselineRecord
	err = json.Unmarshal(b, &rec)
	return &rec, err
}

// CompareResult compares current metrics against baseline
type CompareResult struct {
	Metric      string  `json:"metric"`
	BaselineVal float64 `json:"baseline_val"`
	CurrentVal  float64 `json:"current_val"`
	DeltaPct    float64 `json:"delta_pct"`
	Regression  bool    `json:"regression"`
}

// CompareAgainstBaseline evaluates current metrics against a reference baseline
func (bm *BaselineManager) CompareAgainstBaseline(rec *BaselineRecord, current map[string]float64) []CompareResult {
	var diffs []CompareResult
	for k, baseVal := range rec.Metrics {
		if curVal, ok := current[k]; ok {
			delta := 0.0
			if baseVal != 0 {
				delta = ((curVal - baseVal) / baseVal) * 100.0
			}
			// General heuristic: throughput decrease or latency/loss increase is regression
			reg := false
			if (k == "throughput_mbps" || k == "pps") && delta < -10.0 {
				reg = true
			} else if (k == "rtt_ms" || k == "loss_pct" || k == "heap_mb") && delta > 15.0 {
				reg = true
			}

			diffs = append(diffs, CompareResult{
				Metric:      k,
				BaselineVal: baseVal,
				CurrentVal:  curVal,
				DeltaPct:    delta,
				Regression:  reg,
			})
		}
	}
	return diffs
}
