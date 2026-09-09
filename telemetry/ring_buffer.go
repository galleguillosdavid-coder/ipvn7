package telemetry

import (
	"sync/atomic"
)

// BoundedEventQueue implements a non-blocking, bounded event buffer.
// If the buffer is full, incoming events are immediately dropped and
// counted, guaranteeing that network packet processing is never stalled.
type BoundedEventQueue struct {
	queue        chan TelemetryEvent
	capacity     int
	droppedTotal atomic.Uint64
	enqueuedTotal atomic.Uint64
}

// NewBoundedEventQueue creates a queue with fixed capacity
func NewBoundedEventQueue(capacity int) *BoundedEventQueue {
	if capacity <= 0 {
		capacity = 4096
	}
	return &BoundedEventQueue{
		queue:    make(chan TelemetryEvent, capacity),
		capacity: capacity,
	}
}

// TryPush attempts to enqueue an event without blocking.
// If the queue is at capacity, it drops the event and returns false.
func (q *BoundedEventQueue) TryPush(event TelemetryEvent) bool {
	select {
	case q.queue <- event:
		q.enqueuedTotal.Add(1)
		return true
	default:
		// DROP_TELEMETRY: never block network traffic!
		q.droppedTotal.Add(1)
		return false
	}
}

// PopBatch extracts up to maxBatch events from the queue without blocking
func (q *BoundedEventQueue) PopBatch(maxBatch int) []TelemetryEvent {
	if maxBatch <= 0 {
		maxBatch = 64
	}
	batch := make([]TelemetryEvent, 0, maxBatch)
	for i := 0; i < maxBatch; i++ {
		select {
		case evt := <-q.queue:
			batch = append(batch, evt)
		default:
			return batch
		}
	}
	return batch
}

// DroppedTotal returns the cumulative count of dropped telemetry events
func (q *BoundedEventQueue) DroppedTotal() uint64 {
	return q.droppedTotal.Load()
}

// EnqueuedTotal returns the cumulative count of enqueued telemetry events
func (q *BoundedEventQueue) EnqueuedTotal() uint64 {
	return q.enqueuedTotal.Load()
}

// Len returns current number of events pending in buffer
func (q *BoundedEventQueue) Len() int {
	return len(q.queue)
}

// Capacity returns the maximum configured buffer capacity
func (q *BoundedEventQueue) Capacity() int {
	return q.capacity
}
