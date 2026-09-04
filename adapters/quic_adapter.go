package adapters

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"ipv7/core"

	"github.com/quic-go/quic-go"
)

type QUICAdapter struct {
	listenAddr string
	listener   *quic.Listener
	tlsConfig  *tls.Config
	receive    chan *core.Container

	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc
}

func NewQUICAdapter(listenAddr string) (*QUICAdapter, error) {
	tlsConf, err := GenerateTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS config: %w", err)
	}

	return &QUICAdapter{
		listenAddr: listenAddr,
		tlsConfig:  tlsConf,
		receive:    make(chan *core.Container, 100),
	}, nil
}

func (a *QUICAdapter) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.running {
		return errors.New("quic adapter already running")
	}

	listener, err := quic.ListenAddr(a.listenAddr, a.tlsConfig, &quic.Config{
		KeepAlivePeriod: 10 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to listen on QUIC: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	a.listener = listener
	a.cancel = cancel
	a.running = true

	go a.acceptLoop(ctx)
	return nil
}

func (a *QUICAdapter) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return nil
	}

	a.running = false
	if a.cancel != nil {
		a.cancel()
	}
	if a.listener != nil {
		return a.listener.Close()
	}
	return nil
}

func (a *QUICAdapter) LocalAddr() net.Addr {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.listener != nil {
		return a.listener.Addr()
	}
	return nil
}

func (a *QUICAdapter) Send(c *core.Container, endpoints []string) error {
	a.mu.Lock()
	running := a.running
	tlsConf := a.tlsConfig
	a.mu.Unlock()

	if !running {
		return errors.New("quic adapter is not running")
	}

	data, err := c.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal container: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var lastErr error
	for _, ep := range endpoints {
		conn, err := quic.DialAddr(ctx, ep, tlsConf, &quic.Config{})
		if err != nil {
			lastErr = err
			continue
		}

		stream, err := conn.OpenStreamSync(ctx)
		if err != nil {
			conn.CloseWithError(1, "stream error")
			lastErr = err
			continue
		}

		// Write 4-byte big-endian length prefix, then data
		length := uint32(len(data))
		if err := binary.Write(stream, binary.BigEndian, length); err != nil {
			stream.Close()
			conn.CloseWithError(1, "write length error")
			lastErr = err
			continue
		}

		if _, err := stream.Write(data); err != nil {
			stream.Close()
			conn.CloseWithError(1, "write data error")
			lastErr = err
			continue
		}

		_ = stream.Close()
		// Success on this endpoint
		return nil
	}

	return lastErr
}

func (a *QUICAdapter) Receive() <-chan *core.Container {
	return a.receive
}

func (a *QUICAdapter) acceptLoop(ctx context.Context) {
	for {
		conn, err := a.listener.Accept(ctx)
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				continue
			}
		}

		go a.handleConnection(ctx, conn)
	}
}

func (a *QUICAdapter) handleConnection(ctx context.Context, conn *quic.Conn) {
	defer conn.CloseWithError(0, "normal close")

	for {
		stream, err := conn.AcceptStream(ctx)
		if err != nil {
			return
		}

		go func(s *quic.Stream) {
			defer s.Close()

			var length uint32
			if err := binary.Read(s, binary.BigEndian, &length); err != nil {
				return
			}

			// Read container payload (supports large files/chunks)
			buf := make([]byte, length)
			if _, err := io.ReadFull(s, buf); err != nil {
				return
			}

			c := &core.Container{}
			if err := c.Unmarshal(buf); err != nil {
				return
			}

			if !c.Verify() {
				return
			}

			select {
			case a.receive <- c:
			default:
			}
		}(stream)
	}
}
