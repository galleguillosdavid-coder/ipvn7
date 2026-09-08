package core

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

func TestP2PTunnelPortForwarding(t *testing.T) {
	registry := make(map[string]*mockAdapter)

	// Node A
	idA, privA, _ := GenerateIdentity()
	nodeA := NewNode(idA, privA)
	adapterA := newMockAdapter("addrA", registry)
	nodeA.AddAdapter(adapterA)
	nodeA.SetEndpoints([]string{"addrA"})

	// Node B
	idB, privB, _ := GenerateIdentity()
	nodeB := NewNode(idB, privB)
	adapterB := newMockAdapter("addrB", registry)
	nodeB.AddAdapter(adapterB)
	nodeB.SetEndpoints([]string{"addrB"})

	if err := nodeA.Start(); err != nil {
		t.Fatalf("Failed to start node A: %v", err)
	}
	defer nodeA.Stop()

	if err := nodeB.Start(); err != nil {
		t.Fatalf("Failed to start node B: %v", err)
	}
	defer nodeB.Stop()

	// Perform mutual handshake
	_, _, err := nodeA.Handshake("addrB")
	if err != nil {
		t.Fatalf("Handshake failed: %v", err)
	}

	tunnelA := NewTunnelService(nodeA)
	defer tunnelA.Stop()

	tunnelB := NewTunnelService(nodeB)
	defer tunnelB.Stop()

	// 1. Setup local Echo TCP server on Node B's machine
	echoListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start echo listener: %v", err)
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
				_, _ = io.Copy(c, c) // Echo back everything
			}(conn)
		}
	}()

	// 2. Node A forwards localPort -> Node B -> echoPort
	localListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to get free port: %v", err)
	}
	localPort := localListener.Addr().(*net.TCPAddr).Port
	_ = localListener.Close()

	if err := tunnelA.ForwardPort(localPort, idB, echoPort); err != nil {
		t.Fatalf("Failed to forward port: %v", err)
	}

	// Wait for listener to bind
	time.Sleep(50 * time.Millisecond)

	// 3. Client connects to Node A's local forwarded port
	client, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", localPort), 2*time.Second)
	if err != nil {
		t.Fatalf("Client failed to connect to local forwarded port: %v", err)
	}
	defer client.Close()

	// 4. Send payload and read echo response
	testPayload := []byte("IPv7 encrypted P2P port-forwarding stream test!")
	if _, err := client.Write(testPayload); err != nil {
		t.Fatalf("Failed to write to client: %v", err)
	}

	buf := make([]byte, len(testPayload))
	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, err := io.ReadFull(client, buf); err != nil {
		t.Fatalf("Failed to read echo response: %v", err)
	}

	if !bytes.Equal(buf, testPayload) {
		t.Fatalf("Echo mismatch: got %s, want %s", string(buf), string(testPayload))
	}

	t.Logf("Tunnel test successfully verified roundtrip: %s", string(buf))
}
