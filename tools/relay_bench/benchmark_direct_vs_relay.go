package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"ipv7/adapters"
	"ipv7/core"
)

type ComparisonResult struct {
	Mode           string  `json:"mode"`
	DurationSec    float64 `json:"duration_sec"`
	PacketSizeBytes int    `json:"packet_size_bytes"`
	BytesTotal     int64   `json:"bytes_total"`
	ThroughputMbps float64 `json:"throughput_mbps"`
	ThroughputMBs  float64 `json:"throughput_mb_s"`
	PacketsTotal   int64   `json:"packets_total"`
	PPS            float64 `json:"pps"`
	AvgLatencyMs   float64 `json:"avg_latency_ms"`
	AllocRAMMB     float64 `json:"alloc_ram_mb"`
}

func main() {
	durationSec := flag.Int("duration", 5, "Duration per test in seconds")
	packetSize := flag.Int("size", 1024, "Payload size in bytes")
	targetRelay := flag.String("relay", "", "External relay address (empty for local embedded relay)")
	jsonOut := flag.Bool("json", false, "Output results in JSON format")
	flag.Parse()

	if !*jsonOut {
		fmt.Println("==================================================================")
		fmt.Println("       IPv7 NUCLEAR BENCHMARK: DIRECT P2P vs. DERP RELAY          ")
		fmt.Println("==================================================================")
		fmt.Printf(" [Config] Duration: %d s | Size: %d B | CPUs: %d\n", *durationSec, *packetSize, runtime.NumCPU())
		fmt.Println("==================================================================")
	}

	dur := time.Duration(*durationSec) * time.Second

	// 1. Direct P2P Benchmark
	if !*jsonOut {
		fmt.Println("\n[*] Evaluando Canal 1: DIRECT P2P (UDP Socket a Socket)...")
	}
	directRes := runDirectBenchmark(*packetSize, dur)

	// 2. DERP Relay Benchmark
	if !*jsonOut {
		fmt.Println("\n[*] Evaluando Canal 2: RELAY DERP (Proxy Intermediado Ciego)...")
	}
	relayRes := runRelayBenchmark(*packetSize, dur, *targetRelay)

	if *jsonOut {
		out := map[string]ComparisonResult{
			"direct_p2p": directRes,
			"derp_relay": relayRes,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(out)
		return
	}

	fmt.Println("\n==================================================================")
	fmt.Println("        COMPARATIVA CUANTITATIVA: DIRECT P2P vs. RELAY DERP       ")
	fmt.Println("==================================================================")
	fmt.Printf(" Métrica                | DIRECT P2P         | RELAY DERP (Fallback)\n")
	fmt.Println("------------------------------------------------------------------")
	fmt.Printf(" Throughput Sostenido   | %7.2f Mbps        | %7.2f Mbps\n", directRes.ThroughputMbps, relayRes.ThroughputMbps)
	fmt.Printf(" Tasa de Paquetes       | %7.0f pps         | %7.0f pps\n", directRes.PPS, relayRes.PPS)
	fmt.Printf(" Latencia Tránsito      | %7.3f ms          | %7.3f ms\n", directRes.AvgLatencyMs, relayRes.AvgLatencyMs)
	fmt.Printf(" Memoria Asignada (RAM) | %7.2f MB          | %7.2f MB\n", directRes.AllocRAMMB, relayRes.AllocRAMMB)
	fmt.Printf(" Total Transferido      | %7.2f MB          | %7.2f MB\n", float64(directRes.BytesTotal)/(1024*1024), float64(relayRes.BytesTotal)/(1024*1024))
	fmt.Println("------------------------------------------------------------------")

	penaltyMbps := ((directRes.ThroughputMbps - relayRes.ThroughputMbps) / directRes.ThroughputMbps) * 100
	latencyMultiplier := relayRes.AvgLatencyMs / directRes.AvgLatencyMs

	fmt.Printf(" >> Penalización de Throughput en Relay: %.1f%%\n", penaltyMbps)
	fmt.Printf(" >> Multiplicador de Latencia en Relay : %.2fx\n", latencyMultiplier)
	fmt.Println("==================================================================")
}

func runDirectBenchmark(size int, dur time.Duration) ComparisonResult {
	// Server
	connSrv, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer connSrv.Close()
	srvAddr := connSrv.LocalAddr().String()

	var bytesRecv, pktsRecv, totalTransitNano, transitSamples int64
	stopSrv := make(chan struct{})

	go func() {
		buf := make([]byte, 65535)
		for {
			select {
			case <-stopSrv:
				return
			default:
				_ = connSrv.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
				n, _, err := connSrv.ReadFrom(buf)
				if err == nil && n >= 16 {
					now := time.Now().UnixNano()
					atomic.AddInt64(&bytesRecv, int64(n))
					atomic.AddInt64(&pktsRecv, 1)

					c := &core.Container{}
					if err := c.Unmarshal(buf[:n]); err == nil && len(c.Payload) >= 16 {
						sendNano := int64(binary.BigEndian.Uint64(c.Payload[8:16]))
						d := now - sendNano
						if d >= 0 {
							atomic.AddInt64(&totalTransitNano, d)
							atomic.AddInt64(&transitSamples, 1)
						}
					}
				}
			}
		}
	}()

	// Client
	connCli, err := net.Dial("udp", srvAddr)
	if err != nil {
		panic(err)
	}
	defer connCli.Close()

	idA, privA, _ := core.GenerateIdentity()
	_ = privA

	payload := make([]byte, size)
	_, _ = rand.Read(payload)

	c := &core.Container{
		SenderPubKey: idA.Bytes(),
		Payload:      payload,
		HopLimit:     12,
	}

	start := time.Now()
	deadline := start.Add(dur)
	var seq uint64

	for time.Now().Before(deadline) {
		seq++
		binary.BigEndian.PutUint64(c.Payload[0:8], seq)
		binary.BigEndian.PutUint64(c.Payload[8:16], uint64(time.Now().UnixNano()))
		data, _ := c.Marshal()
		_, _ = connCli.Write(data)
	}

	time.Sleep(100 * time.Millisecond)
	close(stopSrv)

	elapsed := time.Since(start)
	totalBytes := atomic.LoadInt64(&bytesRecv)
	totalPkts := atomic.LoadInt64(&pktsRecv)
	mbps := (float64(totalBytes) * 8) / (elapsed.Seconds() * 1e6)

	var avgLat float64
	samples := atomic.LoadInt64(&transitSamples)
	if samples > 0 {
		avgLat = float64(atomic.LoadInt64(&totalTransitNano)) / float64(samples) / 1e6
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return ComparisonResult{
		Mode:           "DIRECT_P2P",
		DurationSec:    elapsed.Seconds(),
		PacketSizeBytes: size,
		BytesTotal:     totalBytes,
		ThroughputMbps: mbps,
		ThroughputMBs:  float64(totalBytes) / (1024 * 1024) / elapsed.Seconds(),
		PacketsTotal:   totalPkts,
		PPS:            float64(totalPkts) / elapsed.Seconds(),
		AvgLatencyMs:   avgLat,
		AllocRAMMB:     float64(m.Alloc) / (1024 * 1024),
	}
}

func runRelayBenchmark(size int, dur time.Duration, extRelay string) ComparisonResult {
	relayAddr := extRelay
	var server *adapters.RelayServer

	if relayAddr == "" {
		server = adapters.NewRelayServer("127.0.0.1:0")
		if err := server.Start(); err != nil {
			panic(err)
		}
		defer server.Stop()
		relayAddr = server.Addr().String()
	}

	idA, privA, _ := core.GenerateIdentity()
	idB, privB, _ := core.GenerateIdentity()
	_ = privB

	adapterA := adapters.NewRelayAdapter(relayAddr, idA)
	if err := adapterA.Start(); err != nil {
		panic(err)
	}
	defer adapterA.Stop()

	adapterB := adapters.NewRelayAdapter(relayAddr, idB)
	if err := adapterB.Start(); err != nil {
		panic(err)
	}
	defer adapterB.Stop()

	time.Sleep(50 * time.Millisecond)

	var bytesRecv, pktsRecv, totalTransitNano, transitSamples int64
	stopRecv := make(chan struct{})

	go func() {
		for {
			select {
			case <-stopRecv:
				return
			case c := <-adapterB.Receive():
				now := time.Now().UnixNano()
				atomic.AddInt64(&bytesRecv, int64(len(c.Payload)+100))
				atomic.AddInt64(&pktsRecv, 1)

				if len(c.Payload) >= 16 {
					sendNano := int64(binary.BigEndian.Uint64(c.Payload[8:16]))
					d := now - sendNano
					if d >= 0 {
						atomic.AddInt64(&totalTransitNano, d)
						atomic.AddInt64(&transitSamples, 1)
					}
				}
			}
		}
	}()

	payload := make([]byte, size)
	_, _ = rand.Read(payload)

	c := &core.Container{
		SenderPubKey:   idA.Bytes(),
		ReceiverPubKey: idB.Bytes(),
		Payload:        payload,
		HopLimit:       12,
	}

	start := time.Now()
	deadline := start.Add(dur)
	var seq uint64

	for time.Now().Before(deadline) {
		seq++
		binary.BigEndian.PutUint64(c.Payload[0:8], seq)
		binary.BigEndian.PutUint64(c.Payload[8:16], uint64(time.Now().UnixNano()))
		_ = c.Sign(privA)
		_ = adapterA.Send(c, nil)
	}

	time.Sleep(100 * time.Millisecond)
	close(stopRecv)

	elapsed := time.Since(start)
	totalBytes := atomic.LoadInt64(&bytesRecv)
	totalPkts := atomic.LoadInt64(&pktsRecv)
	mbps := (float64(totalBytes) * 8) / (elapsed.Seconds() * 1e6)

	var avgLat float64
	samples := atomic.LoadInt64(&transitSamples)
	if samples > 0 {
		avgLat = float64(atomic.LoadInt64(&totalTransitNano)) / float64(samples) / 1e6
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return ComparisonResult{
		Mode:           "RELAY_DERP",
		DurationSec:    elapsed.Seconds(),
		PacketSizeBytes: size,
		BytesTotal:     totalBytes,
		ThroughputMbps: mbps,
		ThroughputMBs:  float64(totalBytes) / (1024 * 1024) / elapsed.Seconds(),
		PacketsTotal:   totalPkts,
		PPS:            float64(totalPkts) / elapsed.Seconds(),
		AvgLatencyMs:   avgLat,
		AllocRAMMB:     float64(m.Alloc) / (1024 * 1024),
	}
}
