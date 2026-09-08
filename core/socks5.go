package core

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
)

// SOCKS5 protocol constants (RFC 1928)
const (
	socksVersion5    = 0x05
	socksAuthNone    = 0x00
	socksCmdConnect  = 0x01
	socksAtypIPv4    = 0x01
	socksAtypDomain  = 0x03
	socksAtypIPv6    = 0x04
	socksSuccess     = 0x00
	socksCmdFail     = 0x01
	socksNetUnreach  = 0x03
	socksHostUnreach = 0x04
)

// SOCKS5Proxy provides a universal user-space VPN proxy over IPv7
type SOCKS5Proxy struct {
	listenAddr string
	listener   net.Listener
	exitPeer   Identity
	tunnel     *TunnelService
	running    atomic.Bool
	stopCh     chan struct{}
	mu         sync.Mutex
}

// NewSOCKS5Proxy creates a SOCKS5 proxy server
func NewSOCKS5Proxy(listenAddr string, tunnel *TunnelService) *SOCKS5Proxy {
	if listenAddr == "" {
		listenAddr = "127.0.0.1:1080"
	}
	return &SOCKS5Proxy{
		listenAddr: listenAddr,
		tunnel:     tunnel,
		stopCh:     make(chan struct{}),
	}
}

// SetExitPeer configures an IPv7 peer as the exit gateway for all proxy traffic
func (s *SOCKS5Proxy) SetExitPeer(peer Identity) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.exitPeer = peer
}

// Start begins listening for SOCKS5 connections
func (s *SOCKS5Proxy) Start() error {
	l, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return fmt.Errorf("socks5 failed to listen on %s: %w", s.listenAddr, err)
	}

	s.listener = l
	s.running.Store(true)

	go s.acceptLoop()
	return nil
}

// Addr returns the actual listening address
func (s *SOCKS5Proxy) Addr() net.Addr {
	if s.listener != nil {
		return s.listener.Addr()
	}
	return nil
}

func (s *SOCKS5Proxy) acceptLoop() {
	defer s.listener.Close()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.stopCh:
				return
			default:
				return
			}
		}

		go s.handleClient(conn)
	}
}

func (s *SOCKS5Proxy) handleClient(conn net.Conn) {
	defer conn.Close()

	// 1. Negotiation Handshake
	// Client: [VER, NMETHODS, METHODS...]
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return
	}

	if header[0] != socksVersion5 {
		return
	}

	numMethods := int(header[1])
	methods := make([]byte, numMethods)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return
	}

	// Reply: [VER, METHOD] (0x00 = NO AUTHENTICATION REQUIRED)
	if _, err := conn.Write([]byte{socksVersion5, socksAuthNone}); err != nil {
		return
	}

	// 2. Request Details
	// Client: [VER, CMD, RSV, ATYP, DST.ADDR, DST.PORT]
	reqHeader := make([]byte, 4)
	if _, err := io.ReadFull(conn, reqHeader); err != nil {
		return
	}

	if reqHeader[0] != socksVersion5 || reqHeader[1] != socksCmdConnect {
		_, _ = conn.Write([]byte{socksVersion5, socksCmdFail, 0x00, socksAtypIPv4, 0, 0, 0, 0, 0, 0})
		return
	}

	var targetHost string
	switch reqHeader[3] {
	case socksAtypIPv4:
		ipv4 := make([]byte, 4)
		if _, err := io.ReadFull(conn, ipv4); err != nil {
			return
		}
		targetHost = net.IP(ipv4).String()

	case socksAtypDomain:
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			return
		}
		domainLen := int(lenBuf[0])
		domainBuf := make([]byte, domainLen)
		if _, err := io.ReadFull(conn, domainBuf); err != nil {
			return
		}
		targetHost = string(domainBuf)

	case socksAtypIPv6:
		ipv6 := make([]byte, 16)
		if _, err := io.ReadFull(conn, ipv6); err != nil {
			return
		}
		targetHost = net.IP(ipv6).String()

	default:
		_, _ = conn.Write([]byte{socksVersion5, socksCmdFail, 0x00, socksAtypIPv4, 0, 0, 0, 0, 0, 0})
		return
	}

	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBuf); err != nil {
		return
	}
	targetPort := binary.BigEndian.Uint16(portBuf)
	targetAddr := net.JoinHostPort(targetHost, strconv.Itoa(int(targetPort)))

	// 3. Connect to destination
	targetConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		_, _ = conn.Write([]byte{socksVersion5, socksHostUnreach, 0x00, socksAtypIPv4, 0, 0, 0, 0, 0, 0})
		return
	}
	defer targetConn.Close()

	// 4. Send Success Reply
	bndAddr := targetConn.LocalAddr().(*net.TCPAddr)
	reply := []byte{socksVersion5, socksSuccess, 0x00, socksAtypIPv4}
	reply = append(reply, bndAddr.IP.To4()...)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(bndAddr.Port))
	reply = append(reply, portBytes...)

	if _, err := conn.Write(reply); err != nil {
		return
	}

	// 5. Pipe Bidirectionally
	errCh := make(chan error, 2)
	go func() {
		_, err := io.Copy(targetConn, conn)
		errCh <- err
	}()
	go func() {
		_, err := io.Copy(conn, targetConn)
		errCh <- err
	}()

	<-errCh
}

// Stop terminates the SOCKS5 proxy
func (s *SOCKS5Proxy) Stop() {
	if s.running.Swap(false) {
		close(s.stopCh)
		if s.listener != nil {
			_ = s.listener.Close()
		}
	}
}
