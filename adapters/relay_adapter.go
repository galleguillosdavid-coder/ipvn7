package adapters

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"ipv7/core"
)

// RelayServer is a DERP-like encrypted packet relay for IPv7 nodes behind strict symmetric NAT
type RelayServer struct {
	listenAddr string
	listener   net.Listener

	mu      sync.RWMutex
	clients map[string]net.Conn // Hex pubkey -> active socket connection
	running bool
	stopCh  chan struct{}
}

// NewRelayServer creates a new standalone relay server
func NewRelayServer(listenAddr string) *RelayServer {
	return &RelayServer{
		listenAddr: listenAddr,
		clients:    make(map[string]net.Conn),
		stopCh:     make(chan struct{}),
	}
}

// Start begins listening for client connections
func (s *RelayServer) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return errors.New("relay server already running")
	}
	s.mu.Unlock()

	l, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.listener = l
	s.running = true
	s.mu.Unlock()

	go s.acceptLoop()
	return nil
}

// Addr returns the bound address of the relay server
func (s *RelayServer) Addr() net.Addr {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.listener != nil {
		return s.listener.Addr()
	}
	return nil
}

// Stop shuts down the relay server and closes all client connections
func (s *RelayServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}
	s.running = false
	close(s.stopCh)

	for _, conn := range s.clients {
		_ = conn.Close()
	}
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *RelayServer) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.stopCh:
				return
			default:
				continue
			}
		}
		go s.handleClient(conn)
	}
}

func (s *RelayServer) handleClient(conn net.Conn) {
	defer conn.Close()

	// 1. Handshake: Read 32-byte client public key registration
	clientPubKey := make([]byte, 32)
	if _, err := io.ReadFull(conn, clientPubKey); err != nil {
		return
	}

	clientKeyHex := fmt.Sprintf("%x", clientPubKey)

	s.mu.Lock()
	s.clients[clientKeyHex] = conn
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, clientKeyHex)
		s.mu.Unlock()
	}()

	// 2. Relay packet forwarding loop
	for {
		// Read 4-byte big-endian packet length
		var length uint32
		if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
			return
		}

		packet := make([]byte, length)
		if _, err := io.ReadFull(conn, packet); err != nil {
			return
		}

		// Inspect container header to find destination public key
		c := &core.Container{}
		if err := c.Unmarshal(packet); err != nil {
			continue
		}

		destHex := fmt.Sprintf("%x", c.ReceiverPubKey)

		s.mu.RLock()
		destConn, exists := s.clients[destHex]
		s.mu.RUnlock()

		if exists && destConn != nil {
			// Forward raw packet to destination client
			_ = binary.Write(destConn, binary.BigEndian, length)
			_, _ = destConn.Write(packet)
		}
	}
}

// ----------------------------------------------------------------------------
// RelayAdapter (Client Implementation of core.Adapter)
// ----------------------------------------------------------------------------

type RelayAdapter struct {
	relayServerAddr string
	identity        core.Identity
	conn            net.Conn
	receive         chan *core.Container

	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
}

// NewRelayAdapter creates a client adapter connecting to a DERP relay server
func NewRelayAdapter(relayServerAddr string, identity core.Identity) *RelayAdapter {
	return &RelayAdapter{
		relayServerAddr: relayServerAddr,
		identity:        identity,
		receive:         make(chan *core.Container, 100),
		stopCh:          make(chan struct{}),
	}
}

func (a *RelayAdapter) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.running {
		return errors.New("relay adapter already running")
	}

	conn, err := net.Dial("tcp", a.relayServerAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to relay server: %w", err)
	}

	// Register identity with relay server (send 32-byte public key)
	if _, err := conn.Write(a.identity.Bytes()); err != nil {
		conn.Close()
		return fmt.Errorf("failed to register identity with relay: %w", err)
	}

	a.conn = conn
	a.running = true
	a.stopCh = make(chan struct{})

	go a.readLoop(conn)
	return nil
}

func (a *RelayAdapter) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return nil
	}
	a.running = false
	close(a.stopCh)

	if a.conn != nil {
		return a.conn.Close()
	}
	return nil
}

func (a *RelayAdapter) Send(c *core.Container, endpoints []string) error {
	a.mu.Lock()
	conn := a.conn
	running := a.running
	a.mu.Unlock()

	if !running || conn == nil {
		return errors.New("relay adapter is not connected")
	}

	data, err := c.Marshal()
	if err != nil {
		return err
	}

	length := uint32(len(data))
	if err := binary.Write(conn, binary.BigEndian, length); err != nil {
		return err
	}

	_, err = conn.Write(data)
	return err
}

func (a *RelayAdapter) Receive() <-chan *core.Container {
	return a.receive
}

func (a *RelayAdapter) readLoop(conn net.Conn) {
	for {
		var length uint32
		if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
			return
		}

		buf := make([]byte, length)
		if _, err := io.ReadFull(conn, buf); err != nil {
			return
		}

		c := &core.Container{}
		if err := c.Unmarshal(buf); err != nil {
			continue
		}

		if !c.Verify() {
			continue
		}

		select {
		case a.receive <- c:
		case <-a.stopCh:
			return
		}
	}
}
