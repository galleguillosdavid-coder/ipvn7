package adapters

import (
	"errors"
	"fmt"
	"net"
	"sync"
	
	"ipv7/core"
)

const MaxUDPSize = 1400 // Safe MTU limit to avoid fragmentation

type UDPAdapter struct {
	addr      *net.UDPAddr
	conn      *net.UDPConn
	receive   chan *core.Container
	endpoints []string
	
	addrCache map[string]*net.UDPAddr
	cacheMu   sync.RWMutex
	
	antiReplay *AntiReplayTable

	mu        sync.Mutex
	running   bool
}

var udpBufPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 65535)
		return &b
	},
}

func NewUDPAdapter(listenAddr string) (*UDPAdapter, error) {
	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return nil, err
	}
	
	return &UDPAdapter{
		addr:       addr,
		receive:    make(chan *core.Container, 4096),
		addrCache:  make(map[string]*net.UDPAddr),
		antiReplay: NewAntiReplayTable(),
	}, nil
}

func (a *UDPAdapter) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if a.running {
		return errors.New("adapter already running")
	}
	
	conn, err := net.ListenUDP("udp", a.addr)
	if err != nil {
		return err
	}
	
	// Mitigación Finding ADV-01: Ampliar buffers de socket del SO para absorción de ráfagas
	_ = conn.SetReadBuffer(4 * 1024 * 1024)  // 4MB kernel receive buffer
	_ = conn.SetWriteBuffer(4 * 1024 * 1024) // 4MB kernel send buffer

	a.conn = conn
	a.running = true
	if len(a.endpoints) == 0 {
		a.endpoints = []string{conn.LocalAddr().String()}
	}
	
	go a.listenLoop()
	return nil
}

func (a *UDPAdapter) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if !a.running {
		return nil
	}
	
	a.running = false
	if a.conn != nil {
		return a.conn.Close()
	}
	return nil
}

func (a *UDPAdapter) resolveCached(ep string) (*net.UDPAddr, error) {
	a.cacheMu.RLock()
	addr, exists := a.addrCache[ep]
	a.cacheMu.RUnlock()
	if exists {
		return addr, nil
	}

	resolved, err := net.ResolveUDPAddr("udp", ep)
	if err != nil {
		return nil, err
	}

	a.cacheMu.Lock()
	a.addrCache[ep] = resolved
	a.cacheMu.Unlock()
	return resolved, nil
}

func (a *UDPAdapter) Send(c *core.Container, endpoints []string) error {
	a.mu.Lock()
	conn := a.conn
	running := a.running
	a.mu.Unlock()
	
	if !running || conn == nil {
		return errors.New("adapter is not running")
	}
	
	data, err := c.Marshal()
	if err != nil {
		return err
	}
	
	if len(data) > MaxUDPSize {
		return fmt.Errorf("container size %d exceeds MTU limit %d", len(data), MaxUDPSize)
	}
	
	var lastErr error
	sentCount := 0
	for _, ep := range endpoints {
		addr, err := a.resolveCached(ep)
		if err != nil {
			lastErr = err
			continue
		}
		
		_, err = conn.WriteToUDP(data, addr)
		if err == nil {
			sentCount++
		} else {
			lastErr = err
		}
	}
	
	if sentCount > 0 {
		return nil
	}
	return lastErr
}

func (a *UDPAdapter) Receive() <-chan *core.Container {
	return a.receive
}

// Endpoints returns the known local and external reachable endpoints for this adapter
func (a *UDPAdapter) Endpoints() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	eps := make([]string, len(a.endpoints))
	copy(eps, a.endpoints)
	return eps
}

// LocalAddr returns the local bound address string
func (a *UDPAdapter) LocalAddr() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.conn != nil {
		return a.conn.LocalAddr().String()
	}
	return ""
}

// DiscoverEndpoints queries local interfaces and STUN to populate reachable endpoints
func (a *UDPAdapter) DiscoverEndpoints(stunServer string) error {
	a.mu.Lock()
	conn := a.conn
	a.mu.Unlock()

	if conn == nil {
		return errors.New("adapter must be started to discover endpoints")
	}

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return errors.New("cannot determine local udp port")
	}

	var endpoints []string

	// Local LAN endpoints
	localEps, err := GetLocalEndpoints(localAddr.Port)
	if err == nil {
		endpoints = append(endpoints, localEps...)
	}

	// External reflexive endpoint via STUN
	publicEp, err := DiscoverPublicEndpoint(stunServer)
	if err == nil && publicEp != "" {
		endpoints = append(endpoints, publicEp)
	}

	a.mu.Lock()
	a.endpoints = endpoints
	a.mu.Unlock()

	return nil
}

func (a *UDPAdapter) listenLoop() {
	bufPtr := udpBufPool.Get().(*[]byte)
	buf := *bufPtr
	defer udpBufPool.Put(bufPtr)
	
	for {
		a.mu.Lock()
		running := a.running
		conn := a.conn
		a.mu.Unlock()
		
		if !running || conn == nil {
			break
		}
		
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue // usually means connection closed
		}
		
		c := &core.Container{}
		err = c.Unmarshal(buf[:n])
		if err != nil {
			continue // ignore malformed data
		}
		
		if !c.Verify() {
			continue // ignore invalid signatures
		}

		// Default or validate HopLimit / TTL (bounded by DefaultHopLimit)
		if c.HopLimit == 0 {
			c.HopLimit = core.DefaultHopLimit
		} else if c.HopLimit > core.DefaultHopLimit {
			continue // drop packets exceeding Small-World bound
		}
		
		// Anti-replay filter: verify sequence number is not duplicated or stale
		if a.antiReplay != nil && c.Seq > 0 && len(c.SenderPubKey) > 0 {
			if !a.antiReplay.CheckAndSet(string(c.SenderPubKey), c.Seq) {
				continue // drop replayed packet
			}
		}
		
		select {
		case a.receive <- c:
		default:
			// channel full, drop packet (best effort)
		}
	}
}
