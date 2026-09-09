package telemetry

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"runtime"
	"time"
)

// PrometheusExporter serves OpenMetrics and JSON health endpoints
type PrometheusExporter struct {
	bus    *TelemetryBus
	server *http.Server
}

// NewPrometheusExporter creates a new exporter attached to a TelemetryBus
func NewPrometheusExporter(bus *TelemetryBus) *PrometheusExporter {
	return &PrometheusExporter{bus: bus}
}

// Start launches the HTTP server on specified port (e.g. 9100)
func (e *PrometheusExporter) Start(port int) error {
	mux := http.NewServeMux()

	// 1. Prometheus text format (/metrics)
	mux.HandleFunc("/metrics", e.handleMetrics)

	// 2. Health & liveness JSON endpoint (/health)
	mux.HandleFunc("/health", e.handleHealth)

	e.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		_ = e.server.ListenAndServe()
	}()

	return nil
}

// Stop shuts down the exporter HTTP server
func (e *PrometheusExporter) Stop() {
	if e.server != nil {
		_ = e.server.Close()
	}
}

func (e *PrometheusExporter) handleMetrics(w http.ResponseWriter, r *http.Request) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	uptimeSec := time.Since(e.bus.Node.StartTime).Seconds()
	nodeID := e.bus.NodeID

	txPkts := e.bus.Transport.PacketsTX.Load()
	rxPkts := e.bus.Transport.PacketsRX.Load()
	txBytes := e.bus.Transport.BytesTX.Load()
	rxBytes := e.bus.Transport.BytesRX.Load()

	var pps float64 = 0
	var mbps float64 = 0
	if uptimeSec > 0 {
		pps = float64(txPkts+rxPkts) / uptimeSec
		mbps = (float64(txBytes+rxBytes) * 8.0) / (uptimeSec * 1000000.0)
	}

	rttMs := math.Float64frombits(e.bus.Transport.CurrentRTTMs.Load())
	jitterMs := math.Float64frombits(e.bus.Transport.CurrentJitterMs.Load())

	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	// --- A. Node Metrics ---
	fmt.Fprintf(w, "# HELP ipv7_uptime_seconds Node running time in seconds\n# TYPE ipv7_uptime_seconds gauge\n")
	fmt.Fprintf(w, "ipv7_uptime_seconds{node=\"%s\"} %.2f\n", nodeID, uptimeSec)

	fmt.Fprintf(w, "# HELP ipv7_peers Active authenticated peer count\n# TYPE ipv7_peers gauge\n")
	fmt.Fprintf(w, "ipv7_peers{node=\"%s\"} %d\n", nodeID, e.bus.Node.PeersCount.Load())

	fmt.Fprintf(w, "# HELP ipv7_sessions_active Total active E2EE sessions\n# TYPE ipv7_sessions_active gauge\n")
	fmt.Fprintf(w, "ipv7_sessions_active{node=\"%s\"} %d\n", nodeID, e.bus.Node.ActiveSessions.Load())

	fmt.Fprintf(w, "# HELP ipv7_direct_sessions Active direct transport sessions\n# TYPE ipv7_direct_sessions gauge\n")
	fmt.Fprintf(w, "ipv7_direct_sessions{node=\"%s\"} %d\n", nodeID, e.bus.Node.DirectSessions.Load())

	fmt.Fprintf(w, "# HELP ipv7_relay_sessions Active DERP relay sessions\n# TYPE ipv7_relay_sessions gauge\n")
	fmt.Fprintf(w, "ipv7_relay_sessions{node=\"%s\"} %d\n", nodeID, e.bus.Node.RelaySessions.Load())

	fmt.Fprintf(w, "# HELP ipv7_memory_alloc_bytes Heap memory allocated\n# TYPE ipv7_memory_alloc_bytes gauge\n")
	fmt.Fprintf(w, "ipv7_memory_alloc_bytes{node=\"%s\"} %d\n", nodeID, mem.Alloc)

	fmt.Fprintf(w, "# HELP ipv7_memory_sys_bytes Total system memory\n# TYPE ipv7_memory_sys_bytes gauge\n")
	fmt.Fprintf(w, "ipv7_memory_sys_bytes{node=\"%s\"} %d\n", nodeID, mem.Sys)

	fmt.Fprintf(w, "# HELP ipv7_goroutines Number of running goroutines\n# TYPE ipv7_goroutines gauge\n")
	fmt.Fprintf(w, "ipv7_goroutines{node=\"%s\"} %d\n", nodeID, runtime.NumGoroutine())

	// --- B. Transport Metrics ---
	fmt.Fprintf(w, "# HELP ipv7_packets_tx_total Total packets transmitted\n# TYPE ipv7_packets_tx_total counter\n")
	fmt.Fprintf(w, "ipv7_packets_tx_total{node=\"%s\"} %d\n", nodeID, txPkts)

	fmt.Fprintf(w, "# HELP ipv7_packets_rx_total Total packets received\n# TYPE ipv7_packets_rx_total counter\n")
	fmt.Fprintf(w, "ipv7_packets_rx_total{node=\"%s\"} %d\n", nodeID, rxPkts)

	fmt.Fprintf(w, "# HELP ipv7_bytes_tx_total Total bytes transmitted\n# TYPE ipv7_bytes_tx_total counter\n")
	fmt.Fprintf(w, "ipv7_bytes_tx_total{node=\"%s\"} %d\n", nodeID, txBytes)

	fmt.Fprintf(w, "# HELP ipv7_bytes_rx_total Total bytes received\n# TYPE ipv7_bytes_rx_total counter\n")
	fmt.Fprintf(w, "ipv7_bytes_rx_total{node=\"%s\"} %d\n", nodeID, rxBytes)

	fmt.Fprintf(w, "# HELP ipv7_pps_rate Current aggregate PPS rate\n# TYPE ipv7_pps_rate gauge\n")
	fmt.Fprintf(w, "ipv7_pps_rate{node=\"%s\"} %.2f\n", nodeID, pps)

	fmt.Fprintf(w, "# HELP ipv7_throughput_mbps Current throughput in Mbps\n# TYPE ipv7_throughput_mbps gauge\n")
	fmt.Fprintf(w, "ipv7_throughput_mbps{node=\"%s\"} %.4f\n", nodeID, mbps)

	fmt.Fprintf(w, "# HELP ipv7_rtt_ms Current smoothed round-trip time\n# TYPE ipv7_rtt_ms gauge\n")
	fmt.Fprintf(w, "ipv7_rtt_ms{node=\"%s\"} %.3f\n", nodeID, rttMs)

	fmt.Fprintf(w, "# HELP ipv7_jitter_ms Current jitter estimation\n# TYPE ipv7_jitter_ms gauge\n")
	fmt.Fprintf(w, "ipv7_jitter_ms{node=\"%s\"} %.3f\n", nodeID, jitterMs)

	fmt.Fprintf(w, "# HELP ipv7_pmtu_bytes Current detected Path MTU\n# TYPE ipv7_pmtu_bytes gauge\n")
	fmt.Fprintf(w, "ipv7_pmtu_bytes{node=\"%s\"} %d\n", nodeID, e.bus.Transport.CurrentPMTU.Load())

	fmt.Fprintf(w, "# HELP ipv7_socket_queue_drops_total Drops occurred in OS socket receive queue\n# TYPE ipv7_socket_queue_drops_total counter\n")
	fmt.Fprintf(w, "ipv7_socket_queue_drops_total{node=\"%s\"} %d\n", nodeID, e.bus.Transport.SocketQueueDrops.Load())

	fmt.Fprintf(w, "# HELP ipv7_backpressure_drops_total Drops applied by adaptive backpressure policy\n# TYPE ipv7_backpressure_drops_total counter\n")
	fmt.Fprintf(w, "ipv7_backpressure_drops_total{node=\"%s\"} %d\n", nodeID, e.bus.Transport.BackpressureDrops.Load())

	fmt.Fprintf(w, "# HELP ipv7_endpoint_changes_total Roaming events where endpoint changed\n# TYPE ipv7_endpoint_changes_total counter\n")
	fmt.Fprintf(w, "ipv7_endpoint_changes_total{node=\"%s\"} %d\n", nodeID, e.bus.Transport.EndpointChanges.Load())

	// --- C. Protocol Metrics ---
	fmt.Fprintf(w, "# HELP ipv7_handshakes_completed_total Authenticated Noise XX handshakes\n# TYPE ipv7_handshakes_completed_total counter\n")
	fmt.Fprintf(w, "ipv7_handshakes_completed_total{node=\"%s\"} %d\n", nodeID, e.bus.Protocol.HandshakesCompleted.Load())

	fmt.Fprintf(w, "# HELP ipv7_handshakes_failed_total Failed or timed-out handshakes\n# TYPE ipv7_handshakes_failed_total counter\n")
	fmt.Fprintf(w, "ipv7_handshakes_failed_total{node=\"%s\"} %d\n", nodeID, e.bus.Protocol.HandshakesFailed.Load())

	fmt.Fprintf(w, "# HELP ipv7_replay_rejected_total Duplicate packets rejected by sliding window\n# TYPE ipv7_replay_rejected_total counter\n")
	fmt.Fprintf(w, "ipv7_replay_rejected_total{node=\"%s\"} %d\n", nodeID, e.bus.Protocol.ReplayRejected.Load())

	fmt.Fprintf(w, "# HELP ipv7_malformed_rejected_total Non-conforming or fuzz frames dropped\n# TYPE ipv7_malformed_rejected_total counter\n")
	fmt.Fprintf(w, "ipv7_malformed_rejected_total{node=\"%s\"} %d\n", nodeID, e.bus.Protocol.MalformedRejected.Load())

	// --- D. Telemetry Internal Self-Health ---
	fmt.Fprintf(w, "# HELP ipv7_telemetry_dropped_total Events dropped due to full telemetry buffer (never blocks core)\n# TYPE ipv7_telemetry_dropped_total counter\n")
	fmt.Fprintf(w, "ipv7_telemetry_dropped_total{node=\"%s\"} %d\n", nodeID, e.bus.DroppedEvents())
}

func (e *PrometheusExporter) handleHealth(w http.ResponseWriter, r *http.Request) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	status := "HEALTHY"
	if mem.Alloc > 512*1024*1024 {
		status = "DEGRADED_HIGH_MEMORY"
	}

	resp := map[string]interface{}{
		"status":                 status,
		"node_id":                e.bus.NodeID,
		"uptime_sec":             time.Since(e.bus.Node.StartTime).Seconds(),
		"peers":                  e.bus.Node.PeersCount.Load(),
		"goroutines":             runtime.NumGoroutine(),
		"heap_mb":                float64(mem.Alloc) / (1024 * 1024),
		"telemetry_buffer_len":   e.bus.eventQueue.Len(),
		"telemetry_dropped_cnt":  e.bus.DroppedEvents(),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
