package ui

import (
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"ipv7/core"

	"github.com/gorilla/websocket"
)

//go:embed assets/index.html
var indexHTML []byte

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Server struct {
	node      *core.Node
	cascade   *core.CascadeNode
	port      int
	wsClients map[*websocket.Conn]bool
	mu        sync.Mutex
}

type SendMessageReq struct {
	RecipientID string `json:"recipient_id"`
	Message     string `json:"message"`
	Encrypted   bool   `json:"encrypted"`
	Endpoint    string `json:"endpoint,omitempty"`
}

func NewServer(node *core.Node, cascade *core.CascadeNode, port int) *Server {
	if port <= 0 {
		port = 8080
	}
	s := &Server{
		node:      node,
		cascade:   cascade,
		port:      port,
		wsClients: make(map[*websocket.Conn]bool),
	}

	// Register message listener on node to broadcast to UI clients
	node.OnMessage(func(from core.Identity, payload []byte) {
		text := string(payload)
		wasEncrypted := false

		// Attempt transparent E2EE decryption
		if decrypted, err := node.DecryptMessage(payload); err == nil {
			text = string(decrypted)
			wasEncrypted = true
		}

		s.broadcastWS(map[string]interface{}{
			"type":      "incoming_message",
			"from":      from.String(),
			"text":      text,
			"encrypted": wasEncrypted,
			"time":      time.Now().Format("15:04:05"),
		})
	})

	if cascade != nil {
		cascade.OnChunk(func(chunk *core.StreamChunk) {
			s.broadcastWS(map[string]interface{}{
				"type":      "cascade_chunk",
				"stream_id": chunk.StreamID,
				"seq":       chunk.Sequence,
				"size":      len(chunk.Payload),
				"time":      time.Now().Format("15:04:05"),
			})
		})
	}

	return s
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(indexHTML)
	})

	mux.HandleFunc("/api/info", s.handleInfo)
	mux.HandleFunc("/api/peers", s.handlePeers)
	mux.HandleFunc("/api/send", s.handleSend)
	mux.HandleFunc("/api/broadcast-stream", s.handleBroadcastStream)
	mux.HandleFunc("/ws", s.handleWS)

	addr := fmt.Sprintf("127.0.0.1:%d", s.port)
	fmt.Printf("[UI]  Dashboard web available at: http://%s\n", addr)

	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil && err != http.ErrServerClosed {
			fmt.Printf("[UI Error] Failed to listen on %s: %v\n", addr, err)
		}
	}()
	return nil
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var encKeyHex string
	if s.node.EncPubKey != nil {
		encKeyHex = hex.EncodeToString(s.node.EncPubKey.Bytes())
	}

	info := map[string]interface{}{
		"identity":      s.node.Identity.String(),
		"e2ee_pubkey":   encKeyHex,
		"peers_count":   s.node.SmallWorld.TotalPeers(),
		"cascade_count": 0,
	}
	if s.cascade != nil {
		info["cascade_count"] = s.cascade.ChildrenCount()
	}
	_ = json.NewEncoder(w).Encode(info)
}

func (s *Server) handlePeers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	peers := s.node.SmallWorld.FindClosestPeers(s.node.Identity, 50)
	type PeerDTO struct {
		ID        string   `json:"id"`
		Endpoints []string `json:"endpoints"`
		Degree    int      `json:"degree"`
		LatencyMs int64    `json:"latency_ms"`
	}

	var list []PeerDTO
	for _, p := range peers {
		list = append(list, PeerDTO{
			ID:        p.Identity.String(),
			Endpoints: p.Endpoints,
			Degree:    p.Degree,
			LatencyMs: p.Latency.Milliseconds(),
		})
	}
	_ = json.NewEncoder(w).Encode(list)
}

func (s *Server) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SendMessageReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pubBytes, err := hex.DecodeString(req.RecipientID)
	if err != nil || len(pubBytes) != 32 {
		http.Error(w, "Invalid recipient identity (must be 32-byte hex)", http.StatusBadRequest)
		return
	}

	targetID, err := core.NewIdentityFromBytes(pubBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Endpoint != "" {
		s.node.AddPeer(targetID, []string{req.Endpoint})
	}

	payload := []byte(req.Message)
	if req.Encrypted {
		// Derive or use recipient's X25519 key
		_, recipX25519, _ := core.DeriveX25519FromSeed(pubBytes)
		err = s.node.SendEncryptedMessage(targetID, recipX25519.Bytes(), payload)
	} else {
		err = s.node.SendMessage(targetID, payload)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
}

func (s *Server) handleBroadcastStream(w http.ResponseWriter, r *http.Request) {
	if s.cascade == nil {
		http.Error(w, "Cascade node not enabled", http.StatusBadRequest)
		return
	}

	data := []byte(fmt.Sprintf("STREAM_FRAME_%d", time.Now().UnixNano()))
	err := s.cascade.Broadcast("live-ui-stream", uint64(time.Now().Unix()), data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "broadcasted"})
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	s.mu.Lock()
	s.wsClients[conn] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.wsClients, conn)
		s.mu.Unlock()
		conn.Close()
	}()

	for {
		if _, _, err := conn.NextReader(); err != nil {
			break
		}
	}
}

func (s *Server) broadcastWS(msg interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	for client := range s.wsClients {
		_ = client.WriteMessage(websocket.TextMessage, data)
	}
}
