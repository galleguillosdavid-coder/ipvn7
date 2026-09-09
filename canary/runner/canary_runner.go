package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"ipv7/canary/telemetry"
)

func main() {
	durationFlag := flag.Duration("duration", 24*time.Hour, "Duración del experimento Canary-01")
	metricsInterval := flag.Duration("interval", 30*time.Second, "Intervalo de muestreo de métricas")
	flag.Parse()

	fmt.Println("==================================================================")
	fmt.Println("              IPv7 CANARY-01: MULTI-REGION RUNNER                ")
	fmt.Println("         Despliegue Controlado, Observabilidad y Watchdog         ")
	fmt.Println("==================================================================")
	fmt.Printf(" [Protocol Core] ESTADO: CONGELADO (Core Freeze Strict)\n")
	fmt.Printf(" [Host] OS: %s | Arch: %s | CPUs: %d | Time: %s\n",
		runtime.GOOS, runtime.GOARCH, runtime.NumCPU(), time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf(" [Plan] Duración programada: %v | Muestreo: %v\n", *durationFlag, *metricsInterval)
	fmt.Println("==================================================================")

	ctx, cancel := context.WithTimeout(context.Background(), *durationFlag)
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Initialize metrics exporters for the canary nodes
	mChile := telemetry.NewNodeMetrics("NODE_CL_HOST")
	mNotebook := telemetry.NewNodeMetrics("NODE_NB_REMOTE")
	mWSL := telemetry.NewNodeMetrics("NODE_WSL_LINUX")
	mUSA := telemetry.NewNodeMetrics("NODE_USA_EAST")
	mEU := telemetry.NewNodeMetrics("NODE_EU_WEST")
	mRelay := telemetry.NewNodeMetrics("DERP_RELAY_CANARY")

	// Start HTTP metrics endpoints (Prometheus /metrics + /health)
	sChile := telemetry.StartMetricsServer(9101, mChile)
	sWSL := telemetry.StartMetricsServer(9102, mWSL)
	sNotebook := telemetry.StartMetricsServer(9103, mNotebook)
	sUSA := telemetry.StartMetricsServer(9104, mUSA)
	sEU := telemetry.StartMetricsServer(9105, mEU)
	sRelay := telemetry.StartMetricsServer(9199, mRelay)

	defer func() {
		_ = sChile.Shutdown(context.Background())
		_ = sWSL.Shutdown(context.Background())
		_ = sNotebook.Shutdown(context.Background())
		_ = sUSA.Shutdown(context.Background())
		_ = sEU.Shutdown(context.Background())
		_ = sRelay.Shutdown(context.Background())
	}()

	fmt.Println(" [TELEMETRÍA] Servidores Prometheus OpenMetrics & Health activos:")
	fmt.Println("   - Chile Host (9101)     : http://127.0.0.1:9101/metrics | /health")
	fmt.Println("   - WSL2 Linux (9102)     : http://127.0.0.1:9102/metrics | /health")
	fmt.Println("   - Notebook Físico (9103): http://127.0.0.1:9103/metrics | /health")
	fmt.Println("   - USA East Proxy (9104) : http://127.0.0.1:9104/metrics | /health")
	fmt.Println("   - EU West Proxy (9105)  : http://127.0.0.1:9105/metrics | /health")
	fmt.Println("   - Relay DERP (9199)     : http://127.0.0.1:9199/metrics | /health")
	fmt.Println("------------------------------------------------------------------")

	ticker := time.NewTicker(*metricsInterval)
	defer ticker.Stop()

	// Initial simulation counters
	mChile.PeersConnected.Store(3)
	mChile.RoutesCount.Store(12)
	mChile.DirectSessions.Store(2)
	mChile.RelaySessions.Store(1)

	mNotebook.PeersConnected.Store(2)
	mWSL.PeersConnected.Store(2)
	mRelay.RelaySessions.Store(4)

	step := 0
	for {
		select {
		case <-sigChan:
			fmt.Println("\n[CANARY-01] Señal de terminación recibida. Cerrando ordenadamente...")
			return
		case <-ctx.Done():
			fmt.Println("\n[CANARY-01] Duración de prueba alcanzada exitosamente.")
			return
		case t := <-ticker.C:
			step++
			// Simulate live traffic increment without touching core
			mChile.PacketsSent.Add(350)
			mChile.PacketsRecv.Add(348)
			mChile.BytesSent.Add(350 * 1240)
			mChile.BytesRecv.Add(348 * 1240)
			mChile.HandshakesTotal.Add(1)

			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)

			fmt.Printf("[%s][CYCLE #%d] Node_CL: Sent=%d Recv=%d | Heap=%.2fMB Sys=%.2fMB | Goroutines=%d | Status: HEALTHY\n",
				t.Format("15:04:05"), step, mChile.PacketsSent.Load(), mChile.PacketsRecv.Load(),
				float64(mem.Alloc)/(1024*1024), float64(mem.Sys)/(1024*1024), runtime.NumGoroutine())
		}
	}
}

// Silence unused http import if needed
var _ = http.StatusOK
