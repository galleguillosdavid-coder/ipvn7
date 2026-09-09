package telemetry

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// AnomalyRule defines threshold configuration for anomaly detection
type AnomalyRule struct {
	MaxRTTMs         float64
	MaxLossPct       float64
	MinPMTU          int
	MaxGoroutines    int
	MaxHeapMB        float64
	MaxSocketDrops   uint64
	MaxConsecFailures int
}

// DefaultAnomalyRule returns calibrated default operational thresholds
func DefaultAnomalyRule() AnomalyRule {
	return AnomalyRule{
		MaxRTTMs:          150.0,
		MaxLossPct:        5.0,
		MinPMTU:           1200,
		MaxGoroutines:     200,
		MaxHeapMB:         256.0,
		MaxSocketDrops:    50,
		MaxConsecFailures: 3,
	}
}

// AnomalyEngine continuously monitors metrics to emit structured anomaly alerts
type AnomalyEngine struct {
	bus             *TelemetryBus
	rules           AnomalyRule
	mu              sync.Mutex
	lastCheckTime   time.Time
	prevSocketDrops uint64
	prevHeapAlloc   uint64
}

// NewAnomalyEngine initializes the rule-based anomaly detector
func NewAnomalyEngine(bus *TelemetryBus, rules AnomalyRule) *AnomalyEngine {
	return &AnomalyEngine{
		bus:           bus,
		rules:         rules,
		lastCheckTime: time.Now(),
	}
}

// CheckAnomalies executes a single evaluation pass and emits events if thresholds are exceeded
func (a *AnomalyEngine) CheckAnomalies() []AnomalyEvent {
	a.mu.Lock()
	defer a.mu.Unlock()

	var detected []AnomalyEvent
	now := time.Now().UnixNano()
	nodeID := a.bus.NodeID

	// 1. Check RTT Spike
	currentRTT := a.bus.GetRTT()
	if currentRTT > a.rules.MaxRTTMs {
		evt := AnomalyEvent{
			AnomalyType:   "RTT_SPIKE",
			Severity:      SeverityMedium,
			ObservedValue: currentRTT,
			BaselineValue: a.rules.MaxRTTMs,
			Timestamp:     now,
			NodeID:        nodeID,
			Details:       fmt.Sprintf("RTT %.2f ms exceeds threshold %.2f ms", currentRTT, a.rules.MaxRTTMs),
		}
		detected = append(detected, evt)
		a.emitAnomaly(evt)
	}

	// 2. Check PMTU decrease
	pmtu := int(a.bus.Transport.CurrentPMTU.Load())
	if pmtu < a.rules.MinPMTU && pmtu > 0 {
		evt := AnomalyEvent{
			AnomalyType:   "PMTU_DECREASE",
			Severity:      SeverityHigh,
			ObservedValue: float64(pmtu),
			BaselineValue: float64(a.rules.MinPMTU),
			Timestamp:     now,
			NodeID:        nodeID,
			Details:       fmt.Sprintf("Detected PMTU %d B below safe threshold %d B", pmtu, a.rules.MinPMTU),
		}
		detected = append(detected, evt)
		a.emitAnomaly(evt)
	}

	// 3. Check Socket Queue Pressure / Drops
	currentSocketDrops := a.bus.Transport.SocketQueueDrops.Load()
	dropDelta := currentSocketDrops - a.prevSocketDrops
	a.prevSocketDrops = currentSocketDrops
	if dropDelta > a.rules.MaxSocketDrops {
		evt := AnomalyEvent{
			AnomalyType:   "SOCKET_PRESSURE",
			Severity:      SeverityCritical,
			ObservedValue: float64(dropDelta),
			BaselineValue: float64(a.rules.MaxSocketDrops),
			Timestamp:     now,
			NodeID:        nodeID,
			Details:       fmt.Sprintf("Detected %d drops in socket buffer during period", dropDelta),
		}
		detected = append(detected, evt)
		a.emitAnomaly(evt)
	}

	// 4. Check Goroutine growth
	numGoroutines := runtime.NumGoroutine()
	if numGoroutines > a.rules.MaxGoroutines {
		evt := AnomalyEvent{
			AnomalyType:   "GOROUTINE_GROWTH",
			Severity:      SeverityHigh,
			ObservedValue: float64(numGoroutines),
			BaselineValue: float64(a.rules.MaxGoroutines),
			Timestamp:     now,
			NodeID:        nodeID,
			Details:       fmt.Sprintf("Goroutine count (%d) exceeds safe threshold (%d)", numGoroutines, a.rules.MaxGoroutines),
		}
		detected = append(detected, evt)
		a.emitAnomaly(evt)
	}

	// 5. Check Heap Memory Growth
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	heapMB := float64(mem.Alloc) / (1024 * 1024)
	if heapMB > a.rules.MaxHeapMB {
		evt := AnomalyEvent{
			AnomalyType:   "MEMORY_GROWTH",
			Severity:      SeverityHigh,
			ObservedValue: heapMB,
			BaselineValue: a.rules.MaxHeapMB,
			Timestamp:     now,
			NodeID:        nodeID,
			Details:       fmt.Sprintf("Heap allocation %.2f MB exceeds threshold %.2f MB", heapMB, a.rules.MaxHeapMB),
		}
		detected = append(detected, evt)
		a.emitAnomaly(evt)
	}

	return detected
}

func (a *AnomalyEngine) emitAnomaly(anomaly AnomalyEvent) {
	a.bus.EmitEvent(TelemetryEvent{
		Timestamp: anomaly.Timestamp,
		NodeID:    anomaly.NodeID,
		Type:      EventAnomalyDetected,
		Metadata: map[string]string{
			"anomaly_type": anomaly.AnomalyType,
			"severity":     string(anomaly.Severity),
			"details":      anomaly.Details,
		},
	})
}
