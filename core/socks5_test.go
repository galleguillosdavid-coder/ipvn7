package core

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"
)

func TestSOCKS5Proxy(t *testing.T) {
	// 1. Setup target echo server
	echoListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create echo listener: %v", err)
	}
	defer echoListener.Close()
	echoPort := echoListener.Addr().(*net.TCPAddr).Port

	go func() {
		for {
			conn, err := echoListener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				_, _ = io.Copy(c, c)
			}(conn)
		}
	}()

	// 2. Setup SOCKS5 Proxy on dynamic port
	proxy := NewSOCKS5Proxy("127.0.0.1:0", nil)
	if err := proxy.Start(); err != nil {
		t.Fatalf("Failed to start SOCKS5 proxy: %v", err)
	}
	defer proxy.Stop()

	proxyAddr := proxy.Addr().String()

	// 3. Connect client to SOCKS5 Proxy
	client, err := net.DialTimeout("tcp", proxyAddr, 2*time.Second)
	if err != nil {
		t.Fatalf("Failed to connect to SOCKS5 proxy: %v", err)
	}
	defer client.Close()

	// Step A: Client Handshake [VER=5, NMETHODS=1, METHOD=0]
	if _, err := client.Write([]byte{0x05, 0x01, 0x00}); err != nil {
		t.Fatalf("Failed to write handshake: %v", err)
	}
	handshakeResp := make([]byte, 2)
	if _, err := io.ReadFull(client, handshakeResp); err != nil {
		t.Fatalf("Failed to read handshake response: %v", err)
	}
	if handshakeResp[0] != 0x05 || handshakeResp[1] != 0x00 {
		t.Fatalf("Handshake rejected: %v", handshakeResp)
	}

	// Step B: Client Request CONNECT to 127.0.0.1:echoPort
	req := []byte{0x05, 0x01, 0x00, 0x01, 127, 0, 0, 1}
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(echoPort))
	req = append(req, portBytes...)

	if _, err := client.Write(req); err != nil {
		t.Fatalf("Failed to write connect request: %v", err)
	}

	connectResp := make([]byte, 10)
	if _, err := io.ReadFull(client, connectResp); err != nil {
		t.Fatalf("Failed to read connect response: %v", err)
	}
	if connectResp[0] != 0x05 || connectResp[1] != 0x00 {
		t.Fatalf("Connect request failed with code: %d", connectResp[1])
	}

	// Step C: Send data through proxy tunnel to echo server
	testMsg := []byte("SOCKS5 user-space VPN proxy verified on IPv7!")
	if _, err := client.Write(testMsg); err != nil {
		t.Fatalf("Failed to write data: %v", err)
	}

	readBuf := make([]byte, len(testMsg))
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := io.ReadFull(client, readBuf); err != nil {
		t.Fatalf("Failed to read echoed data: %v", err)
	}

	if !bytes.Equal(readBuf, testMsg) {
		t.Fatalf("Echo mismatch: got %s, want %s", string(readBuf), string(testMsg))
	}

	t.Logf("SOCKS5 proxy successfully routed traffic: %s", string(readBuf))
}
