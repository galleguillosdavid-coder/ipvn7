package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"ipv7/adapters"
	"ipv7/core"
	"ipv7/dht"
	"ipv7/ui"
)

func openBrowser(url string) {
	if runtime.GOOS == "windows" {
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	} else if runtime.GOOS == "linux" {
		_ = exec.Command("xdg-open", url).Start()
	} else if runtime.GOOS == "darwin" {
		_ = exec.Command("open", url).Start()
	}
}

func main() {
	port := flag.Int("port", 7001, "UDP and QUIC listening port")
	uiPort := flag.Int("ui", 8080, "Web GUI dashboard port (0 to disable)")
	peerAddr := flag.String("peer", "", "Optional initial peer address to connect to (ip:port)")
	stunServer := flag.String("stun", adapters.DefaultSTUNServer, "STUN server for public NAT traversal")
	autoOpen := flag.Bool("open", true, "Automatically open web dashboard in default browser")
	flag.Parse()

	// Robust Port Auto-selection: fallback to next available port if port is busy
	actualPort := *port
	for i := 0; i < 20; i++ {
		conn, testErr := net.ListenUDP("udp4", &net.UDPAddr{Port: actualPort})
		if testErr == nil {
			_ = conn.Close()
			break
		}
		actualPort += 2
	}
	if actualPort != *port {
		fmt.Printf("[PORT] Puerto %d ocupado; usando puerto disponible %d\n", *port, actualPort)
		*port = actualPort
	}

	actualUIPort := *uiPort
	if *uiPort > 0 {
		for i := 0; i < 20; i++ {
			l, testErr := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", actualUIPort))
			if testErr == nil {
				_ = l.Close()
				break
			}
			actualUIPort++
		}
		if actualUIPort != *uiPort {
			fmt.Printf("[UI]   Puerto Web %d ocupado; usando puerto disponible %d\n", *uiPort, actualUIPort)
			*uiPort = actualUIPort
		}
	}

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
	dhtService := dht.NewDHTService(id, priv)
	node.OnMessage(func(from core.Identity, payload []byte) {
		var dhtMsg dht.Message
		if err := dhtMsg.Unmarshal(payload); err == nil && dhtMsg.Type >= dht.MsgPing && dhtMsg.Type <= dht.MsgValueResponse {
			if reply := dhtService.ProcessMessage(&dhtMsg); reply != nil {
				if encoded, err := reply.Marshal(); err == nil {
					_ = node.SendMessage(from, encoded)
				}
			}
		}
	})

	// 5.1 LAN Auto-Discovery Beacon (Zero-configuration in local Wi-Fi/LAN)
	beacon := core.NewBeaconService(node, core.DefaultBeaconPort, *port)
	if err := beacon.Start(); err != nil {
		fmt.Printf("[LAN Discovery] Info: beacon socket fallback: %v\n", err)
	} else {
		defer beacon.Stop()
		fmt.Println("[LAN Discovery] Auto-descubrimiento en red local (Wi-Fi/LAN) activo.")
	}

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
	node.SetEndpoints(endpoints)
	fmt.Println("[OK]   Reachable Endpoints:")
	for _, ep := range endpoints {
		fmt.Printf("       - %s\n", ep)
	}
	if len(endpoints) > 0 {
		_, _ = dhtService.Publish(endpoints, 1*time.Hour)
	}

	// Connect to initial peer via cryptographic handshake if given
	if *peerAddr != "" {
		fmt.Printf("[...]  Initiating cryptographic handshake with %s...\n", *peerAddr)
		peerID, rtt, err := node.Handshake(*peerAddr)
		if err != nil {
			fmt.Printf("[!]    Handshake with %s pending/failed: %v (retrying in background)\n", *peerAddr, err)
			go func(addr string) {
				ticker := time.NewTicker(2 * time.Second)
				defer ticker.Stop()
				attempts := 0
				for range ticker.C {
					attempts++
					if id, rtt, err := node.Handshake(addr); err == nil {
						dhtService.AddPeer(id, []string{addr})
						fmt.Printf("[OK]   Handshake verified with %s! Peer ID: %s (RTT: %v)\n", addr, id.String()[:16]+"...", rtt)
						return
					}
					if attempts >= 5 {
						return
					}
				}
			}(*peerAddr)
		} else {
			dhtService.AddPeer(peerID, []string{*peerAddr})
			fmt.Printf("[OK]   Handshake verified! Peer ID: %s (RTT: %v, Degree: %d)\n",
				peerID.String()[:16]+"...", rtt, node.SmallWorld.CalculateDegree(peerID))
		}
	}

	// 8. Start Web GUI Dashboard
	if *uiPort > 0 {
		uiServer := ui.NewServer(node, cascade, *uiPort)
		if err := uiServer.Start(); err != nil {
			fmt.Printf("Warning: failed to start Web UI: %v\n", err)
		} else if *autoOpen {
			go func() {
				time.Sleep(600 * time.Millisecond)
				openBrowser(fmt.Sprintf("http://localhost:%d", *uiPort))
			}()
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
