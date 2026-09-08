package core

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	DefaultBeaconPort    = 7000
	BeaconBroadcastAddr  = "255.255.255.255"
	BeaconInterval       = 3 * time.Second
)

// BeaconPacket contains metadata broadcasted across the local subnet
type BeaconPacket struct {
	Identity string   `json:"id"`
	Port     int      `json:"port"`
	Endpoints []string `json:"endpoints"`
}

// BeaconService handles zero-configuration automatic discovery in local LAN / Wi-Fi
type BeaconService struct {
	node       *Node
	listenPort int
	targetPort int
	conn       *net.UDPConn
	stopCh     chan struct{}
	discovered map[string]time.Time
	mu         sync.Mutex
	running    bool
}

// NewBeaconService creates a new LAN auto-discovery service
func NewBeaconService(node *Node, listenPort, targetPort int) *BeaconService {
	if listenPort <= 0 {
		listenPort = DefaultBeaconPort
	}
	if targetPort <= 0 {
		targetPort = 7001
	}
	return &BeaconService{
		node:       node,
		listenPort: listenPort,
		targetPort: targetPort,
		stopCh:     make(chan struct{}),
		discovered: make(map[string]time.Time),
	}
}

// Start initiates listening for broadcast beacons and sending periodic announcements
func (b *BeaconService) Start() error {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return nil
	}

	// Try to listen on UDP broadcast port; fallback to any port if busy
	laddr := &net.UDPAddr{Port: b.listenPort}
	conn, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		// Fallback to ephemeral port for listening if 7000 is occupied
		conn, err = net.ListenUDP("udp4", &net.UDPAddr{Port: 0})
		if err != nil {
			b.mu.Unlock()
			return fmt.Errorf("failed to bind beacon udp socket: %w", err)
		}
	}

	b.conn = conn
	b.running = true
	b.mu.Unlock()

	go b.listenLoop()
	go b.broadcastLoop()

	return nil
}

// Stop terminates the beacon service
func (b *BeaconService) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.running {
		return
	}
	b.running = false
	close(b.stopCh)
	if b.conn != nil {
		_ = b.conn.Close()
	}
}

func (b *BeaconService) listenLoop() {
	buf := make([]byte, 2048)
	for {
		select {
		case <-b.stopCh:
			return
		default:
		}

		n, remoteAddr, err := b.conn.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-b.stopCh:
				return
			default:
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}

		var pkt BeaconPacket
		if err := json.Unmarshal(buf[:n], &pkt); err != nil {
			continue
		}

		// Ignore own broadcasts
		if pkt.Identity == b.node.Identity.String() {
			continue
		}

		// Candidate endpoint: remote IP + advertised node port
		candidateEp := fmt.Sprintf("%s:%d", remoteAddr.IP.String(), pkt.Port)

		b.mu.Lock()
		lastSeen, exists := b.discovered[candidateEp]
		// Re-handshake only once every 30 seconds to avoid spamming
		if exists && time.Since(lastSeen) < 30*time.Second {
			b.mu.Unlock()
			continue
		}
		b.discovered[candidateEp] = time.Now()
		b.mu.Unlock()

		// Automatically initiate cryptographic handshake with newly discovered LAN peer!
		go func(ep string) {
			fmt.Printf("[LAN Discovery] Peer detectado en %s! Iniciando handshake automático...\n", ep)
			_, _, _ = b.node.Handshake(ep)
		}(candidateEp)
	}
}

func (b *BeaconService) broadcastLoop() {
	ticker := time.NewTicker(BeaconInterval)
	defer ticker.Stop()

	for {
		select {
		case <-b.stopCh:
			return
		case <-ticker.C:
			pkt := BeaconPacket{
				Identity:  b.node.Identity.String(),
				Port:      b.targetPort,
				Endpoints: b.node.Endpoints(),
			}

			data, err := json.Marshal(pkt)
			if err != nil {
				continue
			}

			// Send to generic broadcast AND each active interface's directed broadcast (e.g. 192.168.1.255)
			targets := getBroadcastAddresses(b.listenPort)
			for _, target := range targets {
				_, _ = b.conn.WriteToUDP(data, target)
			}
		}
	}
}

func getBroadcastAddresses(port int) []*net.UDPAddr {
	var addrs []*net.UDPAddr
	addrs = append(addrs, &net.UDPAddr{IP: net.IPv4(255, 255, 255, 255), Port: port})

	ifaces, err := net.Interfaces()
	if err != nil {
		return addrs
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrsList, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrsList {
			ipNet, ok := a.(*net.IPNet)
			if !ok || ipNet.IP.To4() == nil {
				continue
			}
			ip4 := ipNet.IP.To4()
			mask := ipNet.Mask
			if len(mask) != 4 {
				continue
			}
			// Compute directed broadcast address: IP | ^Mask
			bcast := net.IPv4(
				ip4[0]|^mask[0],
				ip4[1]|^mask[1],
				ip4[2]|^mask[2],
				ip4[3]|^mask[3],
			)
			addrs = append(addrs, &net.UDPAddr{IP: bcast, Port: port})
		}
	}
	return addrs
}
