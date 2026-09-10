package offgrid

import (
	"bytes"
	"context"
	"sync"
	"time"

	"ipv7/core"
)

var BeaconMagic = [4]byte{0x4F, 0x47, 0x37, 0x21} // OG7! (OffGrid IPv7)

const MaxDiscoveredPeers = 256

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

// Unregister detaches a virtual radio link from the medium
func (m *RadioMedium) Unregister(addr string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.nodes, addr)
}

// Broadcast broadcasts a frame to all physical radios in range safely
func (m *RadioMedium) Broadcast(srcAddr string, data []byte) {
	m.mu.RLock()
	// Copy slice of links to avoid holding lock during deliver
	links := make([]*VirtualRadioLink, 0, len(m.nodes))
	for addr, link := range m.nodes {
		if addr != srcAddr {
			links = append(links, link)
		}
	}
	m.mu.RUnlock()

	for _, link := range links {
		buf := make([]byte, len(data))
		copy(buf, data)
		link.deliver(radioFrame{srcAddr: srcAddr, data: buf})
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

func (l *VirtualRadioLink) deliver(frame radioFrame) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return
	}
	select {
	case l.rxQueue <- frame:
	default:
	}
}

func (l *VirtualRadioLink) Send(targetAddr string, packet []byte) error {
	l.mu.Lock()
	closed := l.closed
	l.mu.Unlock()

	if closed {
		return ErrLinkClosed
	}
	if l.mtu > 0 && len(packet) > l.mtu {
		return ErrPacketExceedsMTU
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
		if l.medium != nil {
			l.medium.Unregister(l.addr)
		}
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
				if _, exists := b.peers[did.String()]; !exists && len(b.peers) >= MaxDiscoveredPeers {
					// Evict oldest seen peer to guarantee O(1) memory bound
					var oldestKey string
					var oldestTime time.Time
					first := true
					for k, v := range b.peers {
						if first || v.LastSeen.Before(oldestTime) {
							oldestTime = v.LastSeen
							oldestKey = k
							first = false
						}
					}
					if oldestKey != "" {
						delete(b.peers, oldestKey)
					}
				}
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
