package core

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// TunnelCmd represents the action of a tunnel packet
type TunnelCmd uint8

const (
	TunnelCmdOpen  TunnelCmd = 1
	TunnelCmdData  TunnelCmd = 2
	TunnelCmdClose TunnelCmd = 3
)

// TunnelPacket carries stream multiplexing control and data
type TunnelPacket struct {
	Cmd        TunnelCmd `cbor:"1,keyasint"`
	StreamID   uint64    `cbor:"2,keyasint"`
	TargetPort uint16    `cbor:"3,keyasint,omitempty"`
	Data       []byte    `cbor:"4,keyasint,omitempty"`
}

// TunnelSession represents an active forwarded connection
type TunnelSession struct {
	StreamID   uint64
	RemotePeer Identity
	Conn       net.Conn
	closed     atomic.Bool
	closeCh    chan struct{}
}

// TunnelService manages P2P port forwarding over IPv7 encrypted transport
type TunnelService struct {
	node      *Node
	sessions  map[uint64]*TunnelSession
	listeners map[int]net.Listener
	mu        sync.RWMutex
	stopCh    chan struct{}
}

// NewTunnelService creates a tunnel service attached to an IPv7 node
func NewTunnelService(node *Node) *TunnelService {
	ts := &TunnelService{
		node:      node,
		sessions:  make(map[uint64]*TunnelSession),
		listeners: make(map[int]net.Listener),
		stopCh:    make(chan struct{}),
	}

	// Register message handler on node for tunnel session packets
	node.OnMessage(func(from Identity, payload []byte) {
		ts.handleIncomingPacket(from, payload)
	})

	return ts
}

// ForwardPort opens a local port listener that forwards all incoming TCP connections
// to targetPort on the remotePeer over IPv7 encrypted containers.
func (ts *TunnelService) ForwardPort(localPort int, remotePeer Identity, targetPort int) error {
	ts.mu.Lock()
	if _, exists := ts.listeners[localPort]; exists {
		ts.mu.Unlock()
		return fmt.Errorf("local port %d is already being forwarded", localPort)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		ts.mu.Unlock()
		return fmt.Errorf("failed to listen on 127.0.0.1:%d: %w", localPort, err)
	}
	ts.listeners[localPort] = listener
	ts.mu.Unlock()

	go ts.acceptLoop(listener, remotePeer, targetPort)
	return nil
}

// CloseListener terminates a local port forwarder
func (ts *TunnelService) CloseListener(localPort int) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	listener, exists := ts.listeners[localPort]
	if !exists {
		return errors.New("listener not found")
	}
	delete(ts.listeners, localPort)
	return listener.Close()
}

func (ts *TunnelService) acceptLoop(listener net.Listener, remotePeer Identity, targetPort int) {
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ts.stopCh:
				return
			default:
				return
			}
		}

		streamID := rand.Uint64()
		sess := &TunnelSession{
			StreamID:   streamID,
			RemotePeer: remotePeer,
			Conn:       conn,
			closeCh:    make(chan struct{}),
		}

		ts.mu.Lock()
		ts.sessions[streamID] = sess
		ts.mu.Unlock()

		// 1. Send TunnelOpen packet to remote peer
		openPkt := &TunnelPacket{
			Cmd:        TunnelCmdOpen,
			StreamID:   streamID,
			TargetPort: uint16(targetPort),
		}
		_ = ts.sendTunnelPacket(remotePeer, openPkt)

		// 2. Read from local connection and pipe to IPv7 peer
		go ts.pipeLocalToRemote(sess)
	}
}

func (ts *TunnelService) pipeLocalToRemote(sess *TunnelSession) {
	defer ts.closeSession(sess.StreamID)

	buf := make([]byte, 16384) // 16 KB chunk size
	for {
		n, err := sess.Conn.Read(buf)
		if n > 0 {
			pkt := &TunnelPacket{
				Cmd:      TunnelCmdData,
				StreamID: sess.StreamID,
				Data:     append([]byte(nil), buf[:n]...),
			}
			if sendErr := ts.sendTunnelPacket(sess.RemotePeer, pkt); sendErr != nil {
				break
			}
		}
		if err != nil {
			break
		}
	}
}

func (ts *TunnelService) handleIncomingPacket(from Identity, payload []byte) {
	// Attempt decryption if payload is encrypted
	data := payload
	if decrypted, err := ts.node.DecryptMessage(payload); err == nil {
		data = decrypted
	}

	pkt, err := decodeTunnelPacket(data)
	if err != nil {
		return
	}

	switch pkt.Cmd {
	case TunnelCmdOpen:
		// Remote peer wants to connect to a local port
		targetAddr := fmt.Sprintf("127.0.0.1:%d", pkt.TargetPort)
		remoteConn, err := net.DialTimeout("tcp", targetAddr, 3*time.Second)
		if err != nil {
			// Notify remote peer of failure
			closePkt := &TunnelPacket{Cmd: TunnelCmdClose, StreamID: pkt.StreamID}
			_ = ts.sendTunnelPacket(from, closePkt)
			return
		}

		sess := &TunnelSession{
			StreamID:   pkt.StreamID,
			RemotePeer: from,
			Conn:       remoteConn,
			closeCh:    make(chan struct{}),
		}

		ts.mu.Lock()
		ts.sessions[pkt.StreamID] = sess
		ts.mu.Unlock()

		go ts.pipeLocalToRemote(sess)

	case TunnelCmdData:
		ts.mu.RLock()
		sess, exists := ts.sessions[pkt.StreamID]
		ts.mu.RUnlock()
		if exists && !sess.closed.Load() && len(pkt.Data) > 0 {
			_, _ = sess.Conn.Write(pkt.Data)
		}

	case TunnelCmdClose:
		ts.closeSession(pkt.StreamID)
	}
}

func (ts *TunnelService) closeSession(streamID uint64) {
	ts.mu.Lock()
	sess, exists := ts.sessions[streamID]
	if exists {
		delete(ts.sessions, streamID)
	}
	ts.mu.Unlock()

	if exists && !sess.closed.Swap(true) {
		close(sess.closeCh)
		if sess.Conn != nil {
			_ = sess.Conn.Close()
		}
		closePkt := &TunnelPacket{Cmd: TunnelCmdClose, StreamID: streamID}
		_ = ts.sendTunnelPacket(sess.RemotePeer, closePkt)
	}
}

func (ts *TunnelService) sendTunnelPacket(target Identity, pkt *TunnelPacket) error {
	bytes := encodeTunnelPacket(pkt)
	// Try E2EE encryption if peer key is known
	recipEncKey := ts.node.GetPeerEncKey(target)
	if len(recipEncKey) == 32 {
		return ts.node.SendEncryptedMessage(target, recipEncKey, bytes)
	}
	return ts.node.SendMessage(target, bytes)
}

// ActiveForwarders returns list of active local port forwarding listeners
func (ts *TunnelService) ActiveForwarders() []int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	ports := make([]int, 0, len(ts.listeners))
	for p := range ts.listeners {
		ports = append(ports, p)
	}
	return ports
}

// Stop shuts down all listeners and active sessions
func (ts *TunnelService) Stop() {
	close(ts.stopCh)
	ts.mu.Lock()
	defer ts.mu.Unlock()

	for _, l := range ts.listeners {
		_ = l.Close()
	}
	for _, sess := range ts.sessions {
		if !sess.closed.Swap(true) {
			_ = sess.Conn.Close()
		}
	}
}

// Binary wire format for TunnelPacket: [1-byte Cmd][8-byte StreamID][2-byte TargetPort][Data...]
func encodeTunnelPacket(pkt *TunnelPacket) []byte {
	buf := make([]byte, 11+len(pkt.Data))
	buf[0] = byte(pkt.Cmd)
	binary.BigEndian.PutUint64(buf[1:9], pkt.StreamID)
	binary.BigEndian.PutUint16(buf[9:11], pkt.TargetPort)
	copy(buf[11:], pkt.Data)
	return buf
}

func decodeTunnelPacket(data []byte) (*TunnelPacket, error) {
	if len(data) < 11 {
		return nil, io.ErrUnexpectedEOF
	}
	pkt := &TunnelPacket{
		Cmd:        TunnelCmd(data[0]),
		StreamID:   binary.BigEndian.Uint64(data[1:9]),
		TargetPort: binary.BigEndian.Uint16(data[9:11]),
	}
	if len(data) > 11 {
		pkt.Data = append([]byte(nil), data[11:]...)
	}
	return pkt, nil
}
