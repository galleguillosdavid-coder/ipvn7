package telemetry

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// EventConsumer defines a sink for dispatched telemetry event batches
type EventConsumer interface {
	ConsumeBatch(events []TelemetryEvent) error
}

// TelemetryBus orchestrates high-throughput, decoupled telemetry collection
type TelemetryBus struct {
	NodeID          string
	Node            NodeMetrics
	Transport       TransportMetrics
	Protocol        ProtocolMetrics
	eventQueue      *BoundedEventQueue
	consumers       []EventConsumer
	mu              sync.RWMutex
	cancel          context.CancelFunc
	running         atomic.Bool
	flushInterval   time.Duration
	maxBatchSize    int
}

// Global default instance for zero-friction decoupling
var (
	defaultBus   *TelemetryBus
	defaultBusMu sync.RWMutex
)

// InitGlobalBus initializes the global telemetry bus
func InitGlobalBus(nodeID string, queueCapacity int) *TelemetryBus {
	defaultBusMu.Lock()
	defer defaultBusMu.Unlock()

	if defaultBus != nil && defaultBus.running.Load() {
		defaultBus.Stop()
	}

	defaultBus = NewTelemetryBus(nodeID, queueCapacity)
	defaultBus.Start(context.Background())
	return defaultBus
}

// GetGlobalBus retrieves current bus or a dummy fallback
func GetGlobalBus() *TelemetryBus {
	defaultBusMu.RLock()
	defer defaultBusMu.RUnlock()
	if defaultBus == nil {
		return NewTelemetryBus("FALLBACK_NODE", 1024)
	}
	return defaultBus
}

// NewTelemetryBus constructs a new bus instance
func NewTelemetryBus(nodeID string, queueCapacity int) *TelemetryBus {
	bus := &TelemetryBus{
		NodeID:        nodeID,
		eventQueue:    NewBoundedEventQueue(queueCapacity),
		flushInterval: 50 * time.Millisecond,
		maxBatchSize:  128,
	}
	bus.Node.StartTime = time.Now()
	bus.Transport.CurrentPMTU.Store(1280)
	return bus
}

// RegisterConsumer adds a batch sink (e.g. Kùzu or Logger)
func (b *TelemetryBus) RegisterConsumer(consumer EventConsumer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.consumers = append(b.consumers, consumer)
}

// Start launches the background event aggregation loop
func (b *TelemetryBus) Start(ctx context.Context) {
	if b.running.Swap(true) {
		return
	}
	ctx, cancel := context.WithCancel(ctx)
	b.cancel = cancel

	go b.workerLoop(ctx)
}

// Stop terminates the bus background processing
func (b *TelemetryBus) Stop() {
	if !b.running.Swap(false) {
		return
	}
	if b.cancel != nil {
		b.cancel()
	}
}

func (b *TelemetryBus) workerLoop(ctx context.Context) {
	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Final flush on exit
			b.flushAll()
			return
		case <-ticker.C:
			b.flushPending()
		}
	}
}

func (b *TelemetryBus) flushPending() {
	batch := b.eventQueue.PopBatch(b.maxBatchSize)
	if len(batch) == 0 {
		return
	}

	b.mu.RLock()
	consumers := make([]EventConsumer, len(b.consumers))
	copy(consumers, b.consumers)
	b.mu.RUnlock()

	for _, c := range consumers {
		_ = c.ConsumeBatch(batch)
	}
}

func (b *TelemetryBus) flushAll() {
	for b.eventQueue.Len() > 0 {
		batch := b.eventQueue.PopBatch(b.maxBatchSize)
		if len(batch) == 0 {
			break
		}
		b.mu.RLock()
		for _, c := range b.consumers {
			_ = c.ConsumeBatch(batch)
		}
		b.mu.RUnlock()
	}
}

// --------------------------------------------------------------------------------
// High-performance atomic event recording (Zero Core Lock)
// --------------------------------------------------------------------------------

// EmitEvent publishes a significant event into the bounded buffer
func (b *TelemetryBus) EmitEvent(evt TelemetryEvent) bool {
	if evt.Timestamp == 0 {
		evt.Timestamp = time.Now().UnixNano()
	}
	if evt.NodeID == "" {
		evt.NodeID = b.NodeID
	}
	return b.eventQueue.TryPush(evt)
}

// RecordPacketTX records packet and byte transmission
func (b *TelemetryBus) RecordPacketTX(bytes uint64) {
	b.Transport.PacketsTX.Add(1)
	b.Transport.BytesTX.Add(bytes)
}

// RecordPacketRX records packet and byte reception
func (b *TelemetryBus) RecordPacketRX(bytes uint64) {
	b.Transport.PacketsRX.Add(1)
	b.Transport.BytesRX.Add(bytes)
}

// RecordRTT updates exponential moving average RTT in float64 bits
func (b *TelemetryBus) RecordRTT(rttMs float64) {
	bits := math.Float64bits(rttMs)
	b.Transport.CurrentRTTMs.Store(bits)
}

// GetRTT returns the last stored RTT in milliseconds
func (b *TelemetryBus) GetRTT() float64 {
	bits := b.Transport.CurrentRTTMs.Load()
	return math.Float64frombits(bits)
}

// RecordJitter updates the current jitter estimation
func (b *TelemetryBus) RecordJitter(jitterMs float64) {
	bits := math.Float64bits(jitterMs)
	b.Transport.CurrentJitterMs.Store(bits)
}

// RecordSocketDrop records a buffer drop at OS / adapter socket layer
func (b *TelemetryBus) RecordSocketDrop() {
	b.Transport.SocketQueueDrops.Add(1)
}

// RecordBackpressureDrop records an early drop applied by backpressure policy
func (b *TelemetryBus) RecordBackpressureDrop() {
	b.Transport.BackpressureDrops.Add(1)
}

// DroppedEvents returns count of dropped telemetry events
func (b *TelemetryBus) DroppedEvents() uint64 {
	return b.eventQueue.DroppedTotal()
}
