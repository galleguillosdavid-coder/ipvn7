package telemetry

import (
	"context"
	"testing"
)

func BenchmarkTelemetryOFF(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Pure baseline simulated packet processing
		_ = i * 1280
	}
}

func BenchmarkTelemetryON(b *testing.B) {
	bus := NewTelemetryBus("BENCH_NODE", 4096)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bus.Start(ctx)
	defer bus.Stop()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bus.RecordPacketTX(1280)
		bus.RecordPacketRX(1280)
		if i%100 == 0 {
			bus.EmitEvent(TelemetryEvent{
				Type:  EventEndpointChanged,
				RTTMs: 4.5,
			})
		}
	}
}

func BenchmarkTelemetryON_Prometheus(b *testing.B) {
	bus := NewTelemetryBus("BENCH_NODE_PROM", 4096)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bus.Start(ctx)
	defer bus.Stop()

	exporter := NewPrometheusExporter(bus)
	_ = exporter.Start(9299)
	defer exporter.Stop()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bus.RecordPacketTX(1280)
		bus.RecordPacketRX(1280)
		if i%100 == 0 {
			bus.EmitEvent(TelemetryEvent{
				Type:  EventEndpointChanged,
				RTTMs: 4.5,
			})
		}
	}
}
