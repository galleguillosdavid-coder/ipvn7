package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"ipv7/adapters"
	"ipv7/core"
	"ipv7/dht"
	"ipv7/ui"
)

func main() {
	port := flag.Int("port", 7001, "UDP and QUIC listening port")
	uiPort := flag.Int("ui", 8080, "Web GUI dashboard port (0 to disable)")
	peerAddr := flag.String("peer", "", "Optional initial peer address to connect to (ip:port)")
	stunServer := flag.String("stun", adapters.DefaultSTUNServer, "STUN server for public NAT traversal")
	flag.Parse()

	// 1. Generate or load node cryptographic identity
	id, priv, err := core.GenerateIdentity()
	if err != nil {
		fmt.Printf("Error generating identity: %v\n", err)
		return
	}

	node := core.NewNode(id, priv)

	// 2. Setup UDP Adapter
	listenAddr := fmt.Sprintf("0.0.0.0:%d", *port)
	udpAdapter, err := adapters.NewUDPAdapter(listenAddr)
	if err != nil {
		fmt.Printf("Error creating UDP adapter: %v\n", err)
		return
	}
	node.AddAdapter(udpAdapter)

	// 3. Setup QUIC Adapter (port+1 to avoid UDP socket collision)
	quicListenAddr := fmt.Sprintf("0.0.0.0:%d", *port+1)
	quicAdapter, err := adapters.NewQUICAdapter(quicListenAddr)
	if err == nil {
		node.AddAdapter(quicAdapter)
		fmt.Printf("[QUIC] Listening on %s\n", quicListenAddr)
	}

	// 4. Setup Streaming en Cascada (Fan-out 10)
	cascade := core.NewCascadeNode(node, 10)

	// 5. Setup DHT Service
	_ = dht.NewDHTService(id, priv)

	// 6. Start Node
	if err := node.Start(); err != nil {
		fmt.Printf("Error starting IPv7 node: %v\n", err)
		return
	}
	defer node.Stop()

	// 7. Discover endpoints (Local LAN + STUN NAT)
	fmt.Println("==================================================================")
	fmt.Println("             IPv7 NEXT-GEN NODE (E2EE + P2P MESH + UI)            ")
	fmt.Println("==================================================================")
	fmt.Printf("[ID]   Ed25519 Public Key : %s\n", id.String())
	if node.EncPubKey != nil {
		fmt.Printf("[E2EE] X25519 Encrypt Key : %x\n", node.EncPubKey.Bytes())
	}
	fmt.Println("[...]  Discovering endpoints via STUN...")

	_ = udpAdapter.DiscoverEndpoints(*stunServer)
	endpoints := udpAdapter.Endpoints()
	fmt.Println("[OK]   Reachable Endpoints:")
	for _, ep := range endpoints {
		fmt.Printf("       - %s\n", ep)
	}

	// Connect to initial peer if given
	if *peerAddr != "" {
		dummyKey := make([]byte, 32)
		peerID, _ := core.NewIdentityFromBytes(dummyKey)
		node.AddPeer(peerID, []string{*peerAddr})
		fmt.Printf("[+]    Linked initial peer: %s\n", *peerAddr)
	}

	// 8. Start Web GUI Dashboard
	if *uiPort > 0 {
		uiServer := ui.NewServer(node, cascade, *uiPort)
		if err := uiServer.Start(); err != nil {
			fmt.Printf("Warning: failed to start Web UI: %v\n", err)
		}
	}

	fmt.Println("==================================================================")
	fmt.Printf(">> Node running! Open http://localhost:%d in your browser.\n", *uiPort)
	fmt.Println(">> Press Ctrl+C to terminate.")
	fmt.Println("==================================================================")

	// Wait for termination signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\nShutting down IPv7 node gracefully...")
}
