package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"net"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"ipv7/adapters"
	"ipv7/core"
)

type BenchmarkStats struct {
	Mode            string  `json:"mode"`
	Target          string  `json:"target,omitempty"`
	ListenPort      int     `json:"port,omitempty"`
	DurationSec     float64 `json:"duration_sec"`
	PacketSize      int     `json:"packet_size_bytes"`
	BytesTotal      int64   `json:"bytes_total"`
	MBTotal         float64 `json:"mb_total"`
	ThroughputMBs   float64 `json:"throughput_mb_s"`
	ThroughputMbps  float64 `json:"throughput_mbps"`
	PacketsTotal    int64   `json:"packets_total"`
	PPS             float64 `json:"pps"`
	PacketsLost     int64   `json:"packets_lost"`
	LossPercentage  float64 `json:"loss_percentage"`
	JitterMs        float64 `json:"jitter_ms"`
	AvgTransitMs    float64 `json:"avg_transit_latency_ms"`
	AllocRAMMB      float64 `json:"alloc_ram_mb"`
	SysRAMMB        float64 `json:"sys_ram_mb"`
	NumGoroutine    int     `json:"num_goroutines"`
	PathMTU         int     `json:"pmtu_bytes"`
}

func main() {
	mode := flag.String("mode", "server", "Run mode: 'server' or 'client'")
	target := flag.String("target", "127.0.0.1:9050", "Target server endpoint (for client mode)")
	listenPort := flag.Int("port", 9050, "UDP listening port (for server mode)")
	durationSec := flag.Int("duration", 10, "Benchmark duration in seconds")
	packetSize := flag.Int("size", 1024, "Packet payload size in bytes (min 16)")
	jsonOut := flag.Bool("json", false, "Output results in JSON format")
	flag.Parse()

	if *packetSize < 16 {
		*packetSize = 16
	}

	if !*jsonOut {
		fmt.Println("==================================================================")
		fmt.Println("       IPv7 EXPERIMENTAL PRODUCTION BENCHMARK (IPERF-IPV7)        ")
		fmt.Println("==================================================================")
		fmt.Printf(" [Config] Mode: %s | Packet Size: %d B | Duration: %d s\n", *mode, *packetSize, *durationSec)
		fmt.Printf(" [System] OS: %s | Arch: %s | CPUs: %d\n", runtime.GOOS, runtime.GOARCH, runtime.NumCPU())
		fmt.Println("==================================================================")
	}

	if *mode == "server" {
		runServer(*listenPort, *jsonOut)
	} else {
		runClient(*target, *packetSize, time.Duration(*durationSec)*time.Second, *jsonOut)
	}
}

func runServer(port int, jsonOut bool) {
	listenAddr := fmt.Sprintf("0.0.0.0:%d", port)
	conn, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Server Error] Failed to bind %s: %v\n", listenAddr, err)
		os.Exit(1)
	}
	defer conn.Close()

	if !jsonOut {
		fmt.Printf("[Server] Escuchando tráfico IPv7 UDP real en %s...\n", listenAddr)
	}

	var bytesReceived int64
	var packetsReceived int64
	var minSeq uint64 = math.MaxUint64
	var maxSeq uint64 = 0
	var totalTransitNano int64
	var transitSamples int64

	// RFC 3550 Jitter variables
	var prevTransit float64
	var jitter float64
	var hasPrev bool

	// Reporter ticker
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()

	go func() {
		var lastBytes int64
		var lastPackets int64
		for range ticker.C {
			currBytes := atomic.LoadInt64(&bytesReceived)
			currPackets := atomic.LoadInt64(&packetsReceived)

			deltaBytes := currBytes - lastBytes
			deltaPackets := currPackets - lastPackets

			mbps := (float64(deltaBytes) * 8) / 1e6
			mbPerSec := float64(deltaBytes) / (1024 * 1024)

			if deltaPackets > 0 && !jsonOut {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				fmt.Printf("[Live Stats] %6.2f MB/s | %7.2f Mbps | %6d pps | RAM: %.1f MB | Total: %.2f MB\n",
					mbPerSec, mbps, deltaPackets, float64(m.Alloc)/(1024*1024), float64(currBytes)/(1024*1024))
			}

			lastBytes = currBytes
			lastPackets = currPackets
		}
	}()

	buf := make([]byte, 65535)
	for {
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			break
		}

		nowNano := time.Now().UnixNano()
		atomic.AddInt64(&bytesReceived, int64(n))
		atomic.AddInt64(&packetsReceived, 1)

		// Check if payload contains benchmark header (min 16 bytes: uint64 seq + int64 timestamp)
		// Try unmarshaling Container to read payload
		c := &core.Container{}
		if err := c.Unmarshal(buf[:n]); err == nil && len(c.Payload) >= 16 {
			seq := binary.BigEndian.Uint64(c.Payload[0:8])
			sendNano := int64(binary.BigEndian.Uint64(c.Payload[8:16]))

			if seq < minSeq {
				minSeq = seq
			}
			if seq > maxSeq {
				maxSeq = seq
			}

			transitMs := float64(nowNano-sendNano) / 1e6
			if transitMs >= 0 {
				atomic.AddInt64(&totalTransitNano, nowNano-sendNano)
				atomic.AddInt64(&transitSamples, 1)

				// RFC 3550 Jitter: D = transit - prevTransit; J = J + (|D| - J) / 16
				if hasPrev {
					d := math.Abs(transitMs - prevTransit)
					jitter += (d - jitter) / 16.0
				} else {
					hasPrev = true
				}
				prevTransit = transitMs
			}
		}
	}

	_ = startTime
}

func runClient(target string, size int, duration time.Duration, jsonOut bool) {
	conn, err := net.Dial("udp", target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Client Error] Failed to dial %s: %v\n", target, err)
		os.Exit(1)
	}
	defer conn.Close()

	idA, privA, _ := core.GenerateIdentity()
	_ = privA

	payload := make([]byte, size)
	_, _ = rand.Read(payload)

	c := &core.Container{
		SenderPubKey: idA.Bytes(),
		Payload:      payload,
		HopLimit:     12,
	}

	pmtu := adapters.NewPMTUDiscovery(200*time.Millisecond, nil)
	pmtu.RegisterEndpoint(target)

	if !jsonOut {
		fmt.Printf("[Client] Transmitiendo a saturación máxima hacia %s durante %v...\n", target, duration)
	}

	var bytesSent int64
	var packetsSent int64
	var seq uint64

	startTime := time.Now()
	deadline := startTime.Add(duration)

	for time.Now().Before(deadline) {
		seq++
		binary.BigEndian.PutUint64(c.Payload[0:8], seq)
		binary.BigEndian.PutUint64(c.Payload[8:16], uint64(time.Now().UnixNano()))

		wireData, err := c.Marshal()
		if err != nil {
			break
		}

		n, err := conn.Write(wireData)
		if err == nil {
			atomic.AddInt64(&bytesSent, int64(n))
			atomic.AddInt64(&packetsSent, 1)
		}
	}

	elapsed := time.Since(startTime)
	totalMB := float64(atomic.LoadInt64(&bytesSent)) / (1024 * 1024)
	mbps := (totalMB * 8) / elapsed.Seconds()
	mbPerSec := totalMB / elapsed.Seconds()
	pps := float64(atomic.LoadInt64(&packetsSent)) / elapsed.Seconds()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := BenchmarkStats{
		Mode:           "client",
		Target:         target,
		DurationSec:    elapsed.Seconds(),
		PacketSize:     size,
		BytesTotal:     atomic.LoadInt64(&bytesSent),
		MBTotal:        totalMB,
		ThroughputMBs:  mbPerSec,
		ThroughputMbps: mbps,
		PacketsTotal:   atomic.LoadInt64(&packetsSent),
		PPS:            pps,
		AllocRAMMB:     float64(m.Alloc) / (1024 * 1024),
		SysRAMMB:       float64(m.Sys) / (1024 * 1024),
		NumGoroutine:   runtime.NumGoroutine(),
		PathMTU:        pmtu.GetMTU(target),
	}

	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(stats)
		return
	}

	fmt.Println("\n==================================================================")
	fmt.Println("               RESULTADOS FINALES DE TRANSMISIÓN                  ")
	fmt.Println("==================================================================")
	fmt.Printf(" Datos Transmitidos  : %.2f MB en %.2f segundos\n", stats.MBTotal, stats.DurationSec)
	fmt.Printf(" Throughput Sostenido: %.2f MB/s (%.2f Mbps / %.4f Gbps)\n", stats.ThroughputMBs, stats.ThroughputMbps, stats.ThroughputMbps/1000)
	fmt.Printf(" Tasa de Paquetes    : %.0f paquetes/segundo (pps)\n", stats.PPS)
	fmt.Printf(" Path MTU Negociado  : %d bytes\n", stats.PathMTU)
	fmt.Printf(" Memoria Asignada    : %.2f MB (Sys: %.2f MB)\n", stats.AllocRAMMB, stats.SysRAMMB)
	fmt.Printf(" Goroutines Activas  : %d\n", stats.NumGoroutine)
	fmt.Println("==================================================================")
}

