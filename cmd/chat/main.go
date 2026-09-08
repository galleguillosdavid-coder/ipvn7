package main

import (
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"ipv7/adapters"
	"ipv7/core"
)

func main() {
	port := flag.Int("port", 0, "Local UDP/QUIC port to listen on (0 for random)")
	peerAddr := flag.String("peer", "", "Peer physical address (ip:port) to send messages to")
	peerKeyHex := flag.String("peer-key", "", "Peer Ed25519 public key in hex (optional)")
	stunServer := flag.String("stun", adapters.DefaultSTUNServer, "STUN server for public NAT discovery")
	flag.Parse()

	// 1. Generate local Identity
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

	// 3. Start Node
	if err := node.Start(); err != nil {
		fmt.Printf("Error starting node: %v\n", err)
		return
	}
	defer node.Stop()

	// 4. Discover endpoints (Local LAN + STUN NAT Reflexive)
	fmt.Println("==================================================")
	fmt.Println("             IPv7 P2P MESSENGER NODE             ")
	fmt.Println("==================================================")
	fmt.Printf("[ID]  My Identity (Ed25519): %s\n", id.String())
	fmt.Println("[...] Discovering local and STUN public endpoints...")

	_ = udpAdapter.DiscoverEndpoints(*stunServer)
	endpoints := udpAdapter.Endpoints()
	node.SetEndpoints(endpoints)
	fmt.Println("[OK]  Reachable Endpoints:")
	for _, ep := range endpoints {
		fmt.Printf("      - %s\n", ep)
	}
	fmt.Println("==================================================")
	fmt.Println("Commands:")
	fmt.Println("  /connect <ip:port> [optional_hex_pubkey]")
	fmt.Println("  Type any message and hit ENTER to send.")
	fmt.Println("  /exit to quit.")
	fmt.Println("==================================================")

	var currentPeerID core.Identity

	// Register peer if provided via flags
	if *peerAddr != "" {
		if *peerKeyHex != "" {
			pubBytes, err := hex.DecodeString(*peerKeyHex)
			if err == nil {
				currentPeerID, _ = core.NewIdentityFromBytes(pubBytes)
			}
		}
		if currentPeerID == nil {
			fmt.Printf("[...] Handshake criptográfico con %s...\n", *peerAddr)
			discoveredID, rtt, err := node.Handshake(*peerAddr)
			if err == nil {
				currentPeerID = discoveredID
				fmt.Printf("[+] Handshake verificado con %s! (ID: %s, RTT: %v)\n", *peerAddr, currentPeerID.String()[:12]+"...", rtt)
			} else {
				fmt.Printf("[!] Handshake inicial pendiente con %s: %v\n", *peerAddr, err)
			}
		} else {
			node.AddPeer(currentPeerID, []string{*peerAddr})
			fmt.Printf("[+] Linked peer: %s at %s\n", currentPeerID.String()[:12]+"...", *peerAddr)
		}
	}

	// 5. Incoming message handler
	node.OnMessage(func(from core.Identity, payload []byte) {
		ts := time.Now().Format("15:04:05")
		shortID := from.String()
		if len(shortID) > 12 {
			shortID = shortID[:12] + "..."
		}

		text := string(payload)
		encTag := ""
		if decrypted, err := node.DecryptMessage(payload); err == nil {
			text = string(decrypted)
			encTag = " [🔒 E2EE]"
		} else if !utf8.Valid(payload) {
			text = fmt.Sprintf("[Mensaje binario cifrado no descifrable: %d bytes]", len(payload))
		}

		fmt.Printf("\n[%s] [FROM %s%s]: %s\n> ", ts, shortID, encTag, text)

		// Auto-register peer for easy reply
		if currentPeerID == nil {
			currentPeerID = from
		}
	})

	// 6. Interactive console loop
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			fmt.Print("> ")
			continue
		}

		if line == "/exit" || line == "/quit" {
			break
		}

		if strings.HasPrefix(line, "/connect ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				addr := parts[1]
				var peerKey core.Identity
				if len(parts) >= 3 {
					keyBytes, err := hex.DecodeString(parts[2])
					if err == nil {
						peerKey, _ = core.NewIdentityFromBytes(keyBytes)
					}
				}
				if peerKey == nil {
					fmt.Printf("[...] Handshake con %s...\n", addr)
					discoveredID, rtt, err := node.Handshake(addr)
					if err == nil {
						peerKey = discoveredID
						fmt.Printf("[+] Autenticado con %s! (ID: %s, RTT: %v)\n> ", addr, peerKey.String()[:12]+"...", rtt)
					} else {
						fmt.Printf("[!] Handshake falló con %s: %v\n> ", addr, err)
						continue
					}
				} else {
					node.AddPeer(peerKey, []string{addr})
					fmt.Printf("[+] Conectado a %s (%s)\n> ", addr, peerKey.String()[:12]+"...")
				}
				currentPeerID = peerKey
			}
			continue
		}

		if currentPeerID == nil {
			fmt.Println("[!] No peer connected. Use: /connect <ip:port> to set a recipient.")
			fmt.Print("> ")
			continue
		}

		// Send message with automatic E2EE if peer key is known
		recipEncKey := node.GetPeerEncKey(currentPeerID)
		var err error
		isEnc := false
		if len(recipEncKey) == 32 {
			err = node.SendEncryptedMessage(currentPeerID, recipEncKey, []byte(line))
			isEnc = true
		} else {
			err = node.SendMessage(currentPeerID, []byte(line))
		}

		if err != nil {
			fmt.Printf("[Error] Failed to send: %v\n", err)
		} else {
			ts := time.Now().Format("15:04:05")
			encTag := ""
			if isEnc {
				encTag = " [🔒 E2EE]"
			}
			fmt.Printf("[%s] [SENT%s]: %s\n", ts, encTag, line)
		}
		fmt.Print("> ")
	}
}
