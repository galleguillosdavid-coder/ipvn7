package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AnalysisRequest represents the standardized IPv7_ANALYSIS_REQUEST v1 package
type AnalysisRequest struct {
	Header                   string                 `json:"header"` // "IPv7_ANALYSIS_REQUEST v1"
	Project                  string                 `json:"project"`
	Commit                   string                 `json:"commit"`
	Environment              map[string]interface{} `json:"environment"`
	Objective                string                 `json:"objective"`
	Baseline                 map[string]float64     `json:"baseline"`
	CurrentResult            map[string]float64     `json:"current_result"`
	RecentEvents             []string               `json:"recent_events"`
	Anomalies                []string               `json:"anomalies"`
	BottleneckCandidates     []string               `json:"bottleneck_candidates"`
	Hypotheses               []string               `json:"hypotheses"`
	PreviousFailedHypotheses []string               `json:"previous_failed_hypotheses"`
	Questions                []string               `json:"questions"`
	RequestedAnalysis        string                 `json:"requested_analysis"`
	Timestamp                string                 `json:"timestamp"`
}

// AnalysisResponse represents the standardized IPv7_ANALYSIS_RESPONSE v1 feedback
type AnalysisResponse struct {
	Header                string   `json:"header"` // "IPv7_ANALYSIS_RESPONSE v1"
	Classification        string   `json:"classification"` // "DEMONSTRATED", "OBSERVED", etc.
	Observations          []string `json:"observations"`
	Evidence              []string `json:"evidence"`
	Inferences            []string `json:"inferences"`
	Hypotheses            []string `json:"hypotheses"`
	Uncertainties         []string `json:"uncertainties"`
	BottleneckIdentified  string   `json:"bottleneck_identified"`
	Confidence            string   `json:"confidence"` // "LOW", "MEDIUM", "HIGH"
	RecommendedExperiment string   `json:"recommended_experiment"`
	Risk                  string   `json:"risk"`
	CoreChangeRequired    bool     `json:"core_change_required"`
	Evaluator             string   `json:"evaluator"` // e.g. "ChatGPT" or "David"
	Timestamp             string   `json:"timestamp"`
}

// GenerateAnalysisPackage creates a ready-to-share markdown / JSON bundle for external AI analysis
func GenerateAnalysisPackage(req AnalysisRequest, outPath string) error {
	req.Header = "IPv7_ANALYSIS_REQUEST v1"
	req.Project = "IPv7 Decentralized Overlay Protocol"
	req.Timestamp = time.Now().UTC().Format(time.RFC3339)

	_ = os.MkdirAll(filepath.Dir(outPath), 0755)

	b, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, b, 0644)
}

// ParseAnalysisResponse parses an incoming external review response
func ParseAnalysisResponse(filePath string) (*AnalysisResponse, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var resp AnalysisResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, fmt.Errorf("invalid analysis response format: %w", err)
	}
	return &resp, nil
}
