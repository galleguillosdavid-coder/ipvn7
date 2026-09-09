package telemetry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
)

// NodeMetrics encapsulates all operational metrics required for canary observability
type NodeMetrics struct {
	NodeID          string
	StartTime       time.Time
	PeersConnected  atomic.Int64
	RoutesCount     atomic.Int64
	DirectSessions  atomic.Int64
	RelaySessions   atomic.Int64
	HandshakesTotal atomic.Int64
	PacketsSent     atomic.Int64
	PacketsRecv     atomic.Int64
	BytesSent       atomic.Int64
	BytesRecv       atomic.Int64
	AdapterDrops    atomic.Int64
	Reconnections   atomic.Int64
	CurrentPMTU     atomic.Int64
	PacketLossPPM   atomic.Int64 // parts per million (10,000 = 1%)
}

// NewNodeMetrics initializes metrics tracker
func NewNodeMetrics(nodeID string) *NodeMetrics {
	m := &NodeMetrics{
		NodeID:    nodeID,
		StartTime: time.Now(),
	}
	m.CurrentPMTU.Store(1280)
	return m
}

// StartMetricsServer starts HTTP server on given port with /metrics and /health
func StartMetricsServer(port int, m *NodeMetrics) *http.Server {
	mux := http.NewServeMux()

	// Prometheus OpenMetrics format
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)

		uptimeSec := time.Since(m.StartTime).Seconds()
		sentPkts := m.PacketsSent.Load()
		recvPkts := m.PacketsRecv.Load()
		sentBytes := m.BytesSent.Load()
		recvBytes := m.BytesRecv.Load()

		var pps float64 = 0
		var mbps float64 = 0
		if uptimeSec > 0 {
			pps = float64(sentPkts+recvPkts) / uptimeSec
			mbps = (float64(sentBytes+recvBytes) * 8.0) / (uptimeSec * 1000000.0)
		}

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# HELP ipv7_uptime_seconds Total node running time in seconds\n")
		fmt.Fprintf(w, "# TYPE ipv7_uptime_seconds gauge\n")
		fmt.Fprintf(w, "ipv7_uptime_seconds{node=\"%s\"} %.2f\n", m.NodeID, uptimeSec)

		fmt.Fprintf(w, "# HELP ipv7_peers_connected Number of actively authenticated peers\n")
		fmt.Fprintf(w, "# TYPE ipv7_peers_connected gauge\n")
		fmt.Fprintf(w, "ipv7_peers_connected{node=\"%s\"} %d\n", m.NodeID, m.PeersConnected.Load())

		fmt.Fprintf(w, "# HELP ipv7_routes_count Active routing entries in Small-World / Kademlia table\n")
		fmt.Fprintf(w, "# TYPE ipv7_routes_count gauge\n")
		fmt.Fprintf(w, "ipv7_routes_count{node=\"%s\"} %d\n", m.NodeID, m.RoutesCount.Load())

		fmt.Fprintf(w, "# HELP ipv7_direct_sessions Active end-to-end direct transport sessions\n")
		fmt.Fprintf(w, "# TYPE ipv7_direct_sessions gauge\n")
		fmt.Fprintf(w, "ipv7_direct_sessions{node=\"%s\"} %d\n", m.NodeID, m.DirectSessions.Load())

		fmt.Fprintf(w, "# HELP ipv7_relay_sessions Active transport sessions routed through DERP relay\n")
		fmt.Fprintf(w, "# TYPE ipv7_relay_sessions gauge\n")
		fmt.Fprintf(w, "ipv7_relay_sessions{node=\"%s\"} %d\n", m.NodeID, m.RelaySessions.Load())

		fmt.Fprintf(w, "# HELP ipv7_handshakes_total Cumulative authenticated Noise XX handshakes\n")
		fmt.Fprintf(w, "# TYPE ipv7_handshakes_total counter\n")
		fmt.Fprintf(w, "ipv7_handshakes_total{node=\"%s\"} %d\n", m.NodeID, m.HandshakesTotal.Load())

		fmt.Fprintf(w, "# HELP ipv7_packets_sent_total Total packets transmitted\n")
		fmt.Fprintf(w, "# TYPE ipv7_packets_sent_total counter\n")
		fmt.Fprintf(w, "ipv7_packets_sent_total{node=\"%s\"} %d\n", m.NodeID, sentPkts)

		fmt.Fprintf(w, "# HELP ipv7_packets_received_total Total packets received\n")
		fmt.Fprintf(w, "# TYPE ipv7_packets_received_total counter\n")
		fmt.Fprintf(w, "ipv7_packets_received_total{node=\"%s\"} %d\n", m.NodeID, recvPkts)

		fmt.Fprintf(w, "# HELP ipv7_pps_rate Current packets per second rate\n")
		fmt.Fprintf(w, "# TYPE ipv7_pps_rate gauge\n")
		fmt.Fprintf(w, "ipv7_pps_rate{node=\"%s\"} %.2f\n", m.NodeID, pps)

		fmt.Fprintf(w, "# HELP ipv7_throughput_mbps Current megabits per second throughput\n")
		fmt.Fprintf(w, "# TYPE ipv7_throughput_mbps gauge\n")
		fmt.Fprintf(w, "ipv7_throughput_mbps{node=\"%s\"} %.4f\n", m.NodeID, mbps)

		fmt.Fprintf(w, "# HELP ipv7_memory_alloc_bytes Heap memory actively allocated\n")
		fmt.Fprintf(w, "# TYPE ipv7_memory_alloc_bytes gauge\n")
		fmt.Fprintf(w, "ipv7_memory_alloc_bytes{node=\"%s\"} %d\n", m.NodeID, mem.Alloc)

		fmt.Fprintf(w, "# HELP ipv7_memory_sys_bytes Total system memory obtained by Go runtime\n")
		fmt.Fprintf(w, "# TYPE ipv7_memory_sys_bytes gauge\n")
		fmt.Fprintf(w, "ipv7_memory_sys_bytes{node=\"%s\"} %d\n", m.NodeID, mem.Sys)

		fmt.Fprintf(w, "# HELP ipv7_goroutines_count Number of running goroutines\n")
		fmt.Fprintf(w, "# TYPE ipv7_goroutines_count gauge\n")
		fmt.Fprintf(w, "ipv7_goroutines_count{node=\"%s\"} %d\n", m.NodeID, runtime.NumGoroutine())

		fmt.Fprintf(w, "# HELP ipv7_pmtu_bytes Current Path MTU detected\n")
		fmt.Fprintf(w, "# TYPE ipv7_pmtu_bytes gauge\n")
		fmt.Fprintf(w, "ipv7_pmtu_bytes{node=\"%s\"} %d\n", m.NodeID, m.CurrentPMTU.Load())

		fmt.Fprintf(w, "# HELP ipv7_reconnections_total Automatic reconnection events\n")
		fmt.Fprintf(w, "# TYPE ipv7_reconnections_total counter\n")
		fmt.Fprintf(w, "ipv7_reconnections_total{node=\"%s\"} %d\n", m.NodeID, m.Reconnections.Load())

		fmt.Fprintf(w, "# HELP ipv7_adapter_drops_total Drops occurred at adapter buffer level\n")
		fmt.Fprintf(w, "# TYPE ipv7_adapter_drops_total counter\n")
		fmt.Fprintf(w, "ipv7_adapter_drops_total{node=\"%s\"} %d\n", m.NodeID, m.AdapterDrops.Load())
	})

	// JSON Health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)

		status := "HEALTHY"
		if mem.Alloc > 512*1024*1024 {
			status = "DEGRADED_HIGH_MEMORY"
		}

		resp := map[string]interface{}{
			"status":            status,
			"node_id":           m.NodeID,
			"uptime_sec":        time.Since(m.StartTime).Seconds(),
			"peers":             m.PeersConnected.Load(),
			"goroutines":        runtime.NumGoroutine(),
			"heap_mb":           float64(mem.Alloc) / (1024 * 1024),
			"sys_mb":            float64(mem.Sys) / (1024 * 1024),
			"reconnections":     m.Reconnections.Load(),
			"backpressure_mode": "ACTIVE",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		_ = server.ListenAndServe()
	}()

	return server
}
