package offgrid

import (
	"bytes"
	"context"
	"sync"
	"time"

	"ipv7/core"
)

var BeaconMagic = [4]byte{0x4F, 0x47, 0x37, 0x21} // OG7! (OffGrid IPv7)

// VirtualRadioLink simulates ad-hoc physical radio channels (Wi-Fi Direct, Bluetooth LE, LoRa)
type VirtualRadioLink struct {
	name      string
	addr      string
	mtu       int
	medium    *RadioMedium
	rxQueue   chan radioFrame
	closed    bool
	mu        sync.Mutex
}

type radioFrame struct {
	srcAddr string
	data    []byte
}

// RadioMedium simulates shared airwaves between ad-hoc nodes
type RadioMedium struct {
	nodes map[string]*VirtualRadioLink
	mu    sync.RWMutex
}

// NewRadioMedium creates a shared airwave simulator
func NewRadioMedium() *RadioMedium {
	return &RadioMedium{
		nodes: make(map[string]*VirtualRadioLink),
	}
}

// Register attaches a virtual radio link to the shared medium
func (m *RadioMedium) Register(link *VirtualRadioLink) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nodes[link.addr] = link
}

// Broadcast broadcasts a frame to all physical radios in range
func (m *RadioMedium) Broadcast(srcAddr string, data []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for addr, link := range m.nodes {
		if addr != srcAddr {
			buf := make([]byte, len(data))
			copy(buf, data)
			select {
			case link.rxQueue <- radioFrame{srcAddr: srcAddr, data: buf}:
			default:
			}
		}
	}
}

// NewVirtualRadioLink creates a new ad-hoc physical radio link
func NewVirtualRadioLink(name, addr string, mtu int, medium *RadioMedium) *VirtualRadioLink {
	if mtu <= 0 {
		mtu = 1280
	}
	link := &VirtualRadioLink{
		name:    name,
		addr:    addr,
		mtu:     mtu,
		medium:  medium,
		rxQueue: make(chan radioFrame, 1024),
	}
	if medium != nil {
		medium.Register(link)
	}
	return link
}

func (l *VirtualRadioLink) Name() string { return l.name }
func (l *VirtualRadioLink) MTU() int     { return l.mtu }

func (l *VirtualRadioLink) Send(targetAddr string, packet []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return ErrLinkClosed
	}
	if l.medium != nil {
		l.medium.Broadcast(l.addr, packet)
	}
	return nil
}

func (l *VirtualRadioLink) Receive() (string, []byte, error) {
	frame, ok := <-l.rxQueue
	if !ok {
		return "", nil, ErrLinkClosed
	}
	return frame.srcAddr, frame.data, nil
}

func (l *VirtualRadioLink) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.closed {
		l.closed = true
		close(l.rxQueue)
	}
	return nil
}

// BeaconEngine broadcasts and listens for local physical layer-2 beacons
type BeaconEngine struct {
	node      *core.Node
	link      PhysicalLink
	peers     map[string]PhysicalPeer
	mu        sync.RWMutex
	onPeer    func(peer PhysicalPeer)
	onData    func(srcAddr string, packet []byte)
	cancel    context.CancelFunc
}

// NewBeaconEngine creates a beacon announcer/listener on top of a physical link
func NewBeaconEngine(node *core.Node, link PhysicalLink, onPeer func(peer PhysicalPeer)) *BeaconEngine {
	return &BeaconEngine{
		node:   node,
		link:   link,
		peers:  make(map[string]PhysicalPeer),
		onPeer: onPeer,
	}
}

// SetOnData registers a handler for non-beacon data packets received on the physical link
func (b *BeaconEngine) SetOnData(cb func(srcAddr string, packet []byte)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onData = cb
}

// Start begins periodic local beacons and background reception
func (b *BeaconEngine) Start(ctx context.Context, beaconInterval time.Duration) {
	ctx, cancel := context.WithCancel(ctx)
	b.cancel = cancel

	if beaconInterval <= 0 {
		beaconInterval = 2 * time.Second
	}

	go b.txLoop(ctx, beaconInterval)
	go b.rxLoop(ctx)
}

func (b *BeaconEngine) Stop() {
	if b.cancel != nil {
		b.cancel()
	}
}

func (b *BeaconEngine) txLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Beacon payload: [BeaconMagic (4B)] [DID (32B)]
	var beaconBuf [36]byte
	copy(beaconBuf[:4], BeaconMagic[:])
	copy(beaconBuf[4:], b.node.Identity.Bytes())

	_ = b.link.Send("BROADCAST", beaconBuf[:])

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = b.link.Send("BROADCAST", beaconBuf[:])
		}
	}
}

func (b *BeaconEngine) rxLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		srcAddr, packet, err := b.link.Receive()
		if err != nil {
			return
		}

		if len(packet) == 36 && bytes.Equal(packet[:4], BeaconMagic[:]) {
			did, err := core.NewIdentityFromBytes(packet[4:36])
			if err == nil && !bytes.Equal(did.Bytes(), b.node.Identity.Bytes()) {
				peer := PhysicalPeer{
					DID:       did,
					LocalAddr: srcAddr,
					RSSI:      -45, // Simulado
					LastSeen:  time.Now(),
				}
				b.mu.Lock()
				b.peers[did.String()] = peer
				b.mu.Unlock()

				if b.onPeer != nil {
					b.onPeer(peer)
				}
			}
		} else {
			b.mu.RLock()
			cb := b.onData
			b.mu.RUnlock()
			if cb != nil {
				cb(srcAddr, packet)
			}
		}
	}
}

// DiscoveredPeers returns currently active physical peers
func (b *BeaconEngine) DiscoveredPeers() []PhysicalPeer {
	b.mu.RLock()
	defer b.mu.RUnlock()
	res := make([]PhysicalPeer, 0, len(b.peers))
	for _, p := range b.peers {
		res = append(res, p)
	}
	return res
}
