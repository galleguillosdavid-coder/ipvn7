package tun

import (
	"errors"
	"sync"
)

// MemoryTunDevice represents an in-memory virtual TUN device for testing and simulation
type MemoryTunDevice struct {
	name    string
	mtu     int
	inbound chan []byte
	outbound chan []byte
	closed  bool
	mu      sync.Mutex
}

// NewMemoryTunDevice creates a new virtual TUN device backed by non-blocking ring channels
func NewMemoryTunDevice(name string, mtu int, queueSize int) *MemoryTunDevice {
	if mtu <= 0 {
		mtu = DefaultTunMTU
	}
	if queueSize <= 0 {
		queueSize = 1024
	}
	return &MemoryTunDevice{
		name:     name,
		mtu:      mtu,
		inbound:  make(chan []byte, queueSize),
		outbound: make(chan []byte, queueSize),
	}
}

// Name returns the device name
func (d *MemoryTunDevice) Name() string {
	return d.name
}

// MTU returns the configured MTU
func (d *MemoryTunDevice) MTU() int {
	return d.mtu
}

// Read reads a packet injected into the TUN device
func (d *MemoryTunDevice) Read(p []byte) (n int, err error) {
	pkt, ok := <-d.inbound
	if !ok {
		return 0, errors.New("device closed")
	}
	n = copy(p, pkt)
	return n, nil
}

// Write writes a packet from the OS stack to the TUN device outbound channel
func (d *MemoryTunDevice) Write(p []byte) (n int, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return 0, errors.New("device closed")
	}

	buf := make([]byte, len(p))
	copy(buf, p)

	select {
	case d.outbound <- buf:
		return len(p), nil
	default:
		// Drop non-blocking on ring full to prevent deadlock
		return len(p), nil
	}
}

// InjectPacket simulates the OS network stack writing an IP packet to the TUN device
func (d *MemoryTunDevice) InjectPacket(p []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return errors.New("device closed")
	}

	buf := make([]byte, len(p))
	copy(buf, p)

	select {
	case d.inbound <- buf:
		return nil
	default:
		return errors.New("tun buffer full")
	}
}

// ReceiveOutbound receives a packet emitted by the TUN adapter to the virtual network
func (d *MemoryTunDevice) ReceiveOutbound() ([]byte, error) {
	pkt, ok := <-d.outbound
	if !ok {
		return nil, errors.New("device closed")
	}
	return pkt, nil
}

// Close closes the virtual device
func (d *MemoryTunDevice) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.closed {
		d.closed = true
		close(d.inbound)
		close(d.outbound)
	}
	return nil
}
