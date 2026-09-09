package telemetry

import (
	"sync/atomic"
	"time"
)

// EventType categorizes significant, low-frequency telemetry events
type EventType string

const (
	EventNodeStart          EventType = "NODE_START"
	EventNodeStop           EventType = "NODE_STOP"
	EventPeerJoined         EventType = "PEER_JOINED"
	EventPeerLeft           EventType = "PEER_LEFT"
	EventHandshakeStarted   EventType = "HANDSHAKE_STARTED"
	EventHandshakeCompleted EventType = "HANDSHAKE_COMPLETED"
	EventHandshakeFailed    EventType = "HANDSHAKE_FAILED"
	EventSessionCreated     EventType = "SESSION_CREATED"
	EventSessionClosed      EventType = "SESSION_CLOSED"
	EventEndpointChanged    EventType = "ENDPOINT_CHANGED"
	EventRouteChanged       EventType = "ROUTE_CHANGED"
	EventPMTUChanged        EventType = "PMTU_CHANGED"
	EventDirectPathLost     EventType = "DIRECT_PATH_LOST"
	EventRelayEntered       EventType = "RELAY_ENTERED"
	EventRelayExited        EventType = "RELAY_EXITED"
	EventBackpressure       EventType = "BACKPRESSURE_DETECTED"
	EventSocketPressure     EventType = "SOCKET_PRESSURE"
	EventAnomalyDetected    EventType = "ANOMALY_DETECTED"
	EventRecoveryCompleted  EventType = "RECOVERY_COMPLETED"
)

// TelemetryEvent is a compact, decoupled event structure
type TelemetryEvent struct {
	Timestamp    int64             `json:"timestamp"` // Unix nano
	NodeID       string            `json:"node_id"`
	Type         EventType         `json:"event_type"`
	PeerID       string            `json:"peer_id,omitempty"`
	SessionID    string            `json:"session_id,omitempty"`
	OldEndpoint  string            `json:"old_endpoint,omitempty"`
	NewEndpoint  string            `json:"new_endpoint,omitempty"`
	Route        string            `json:"route,omitempty"`
	Transport    string            `json:"transport,omitempty"` // "DIRECT" or "RELAY"
	PMTU         int               `json:"pmtu,omitempty"`
	RTTMs        float64           `json:"rtt_ms,omitempty"`
	LossPct      float64           `json:"loss_pct,omitempty"`
	RecoveryMs   float64           `json:"recovery_ms,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// AnomalySeverity categorizes detected issues
type AnomalySeverity string

const (
	SeverityLow      AnomalySeverity = "LOW"
	SeverityMedium   AnomalySeverity = "MEDIUM"
	SeverityHigh     AnomalySeverity = "HIGH"
	SeverityCritical AnomalySeverity = "CRITICAL"
)

// AnomalyEvent defines detected operational deviations
type AnomalyEvent struct {
	AnomalyType   string          `json:"anomaly_type"`
	Severity      AnomalySeverity `json:"severity"`
	ObservedValue float64         `json:"observed_value"`
	BaselineValue float64         `json:"baseline_value"`
	Timestamp     int64           `json:"timestamp"`
	NodeID        string          `json:"node_id"`
	PeerID        string          `json:"peer_id,omitempty"`
	Details       string          `json:"details,omitempty"`
}

// NodeMetrics captures aggregate node health metrics
type NodeMetrics struct {
	StartTime      time.Time
	PeersCount     atomic.Int64
	ActiveSessions atomic.Int64
	ActiveRoutes   atomic.Int64
	DirectSessions atomic.Int64
	RelaySessions  atomic.Int64
	CPUUsagePct    atomic.Uint64 // float64 bits
	MemoryBytes    atomic.Uint64
	HeapBytes      atomic.Uint64
	Goroutines     atomic.Int64
}

// TransportMetrics captures low-overhead aggregate transport metrics
type TransportMetrics struct {
	PacketsRX         atomic.Uint64
	PacketsTX         atomic.Uint64
	BytesRX           atomic.Uint64
	BytesTX           atomic.Uint64
	PacketLossPPM     atomic.Int64 // parts per million (10,000 = 1%)
	CurrentRTTMs      atomic.Uint64 // float64 bits
	CurrentJitterMs   atomic.Uint64 // float64 bits
	CurrentPMTU       atomic.Int64
	SocketQueueDrops  atomic.Uint64
	BackpressureDrops atomic.Uint64
	EndpointChanges   atomic.Uint64
	DirectToRelay     atomic.Uint64
	RelayToDirect     atomic.Uint64
}

// ProtocolMetrics captures protocol-level counters
type ProtocolMetrics struct {
	HandshakesStarted   atomic.Uint64
	HandshakesCompleted atomic.Uint64
	HandshakesFailed    atomic.Uint64
	SessionCreated      atomic.Uint64
	SessionClosed       atomic.Uint64
	ReplayRejected      atomic.Uint64
	MalformedRejected   atomic.Uint64
	SignatureRejected   atomic.Uint64
	AEADRejected        atomic.Uint64
	RouteLookup         atomic.Uint64
	RouteSuccess        atomic.Uint64
	RouteFailure        atomic.Uint64
	RouteChanges        atomic.Uint64
	PMTUProbes          atomic.Uint64
	PMTUChanges         atomic.Uint64
	PeerJoins           atomic.Uint64
	PeerLeaves          atomic.Uint64
}
