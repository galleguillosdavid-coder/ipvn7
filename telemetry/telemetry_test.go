package telemetry

import (
	"context"
	"testing"
	"time"
)

type mockConsumer struct {
	receivedEvents []TelemetryEvent
}

func (m *mockConsumer) ConsumeBatch(events []TelemetryEvent) error {
	m.receivedEvents = append(m.receivedEvents, events...)
	return nil
}

func TestBoundedQueueNonBlockingDrop(t *testing.T) {
	// Queue with tiny capacity = 2
	q := NewBoundedEventQueue(2)

	e1 := TelemetryEvent{NodeID: "A", Type: EventNodeStart}
	e2 := TelemetryEvent{NodeID: "A", Type: EventPeerJoined}
	e3 := TelemetryEvent{NodeID: "A", Type: EventHandshakeStarted}

	if !q.TryPush(e1) {
		t.Fatalf("e1 should be enqueued")
	}
	if !q.TryPush(e2) {
		t.Fatalf("e2 should be enqueued")
	}

	// 3rd push must DROP immediately without blocking
	t0 := time.Now()
	pushed := q.TryPush(e3)
	dur := time.Since(t0)

	if pushed {
		t.Fatalf("e3 should have been dropped")
	}
	if q.DroppedTotal() != 1 {
		t.Fatalf("expected 1 dropped event, got %d", q.DroppedTotal())
	}
	if dur > 1*time.Millisecond {
		t.Fatalf("TryPush blocked! took %v", dur)
	}

	// PopBatch
	batch := q.PopBatch(10)
	if len(batch) != 2 {
		t.Fatalf("expected 2 items popped, got %d", len(batch))
	}
}

func TestTelemetryBusAsyncConsumer(t *testing.T) {
	bus := NewTelemetryBus("TEST_NODE_1", 100)
	consumer := &mockConsumer{}
	bus.RegisterConsumer(consumer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bus.Start(ctx)

	bus.EmitEvent(TelemetryEvent{Type: EventNodeStart})
	bus.EmitEvent(TelemetryEvent{Type: EventPeerJoined, PeerID: "PEER_B"})

	time.Sleep(100 * time.Millisecond)
	bus.Stop()

	if len(consumer.receivedEvents) != 2 {
		t.Fatalf("expected 2 received events by consumer, got %d", len(consumer.receivedEvents))
	}
}

func TestAnomalyEngineDetection(t *testing.T) {
	bus := NewTelemetryBus("TEST_NODE_ANOMALY", 100)
	rules := DefaultAnomalyRule()
	rules.MaxRTTMs = 50.0 // strict threshold

	detector := NewAnomalyEngine(bus, rules)

	// Normal RTT
	bus.RecordRTT(25.0)
	anomalies := detector.CheckAnomalies()
	if len(anomalies) != 0 {
		t.Fatalf("expected 0 anomalies, got %d", len(anomalies))
	}

	// Spike RTT
	bus.RecordRTT(120.0)
	anomalies = detector.CheckAnomalies()
	if len(anomalies) != 1 {
		t.Fatalf("expected 1 anomaly (RTT_SPIKE), got %d", len(anomalies))
	}
	if anomalies[0].AnomalyType != "RTT_SPIKE" {
		t.Fatalf("expected RTT_SPIKE, got %s", anomalies[0].AnomalyType)
	}
}
