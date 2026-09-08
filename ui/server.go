package ui

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
	"unicode/utf8"

	"ipv7/core"

	"github.com/fxamacker/cbor/v2"
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
		} else if !utf8.Valid(payload) {
			text = fmt.Sprintf("[Mensaje binario no legible: %d bytes]", len(payload))
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
	mux.HandleFunc("/api/mesh", s.handleMesh)
	mux.HandleFunc("/api/mesh/export-cypher", s.handleMeshCypher)
	mux.HandleFunc("/api/kuzu/query", s.handleKuzuQuery)
	mux.HandleFunc("/api/kuzu/network-graph", s.handleKuzuNetworkGraph)
	mux.HandleFunc("/api/kuzu/code-graph", s.handleKuzuCodeGraph)
	mux.HandleFunc("/api/send", s.handleSend)
	mux.HandleFunc("/api/ping", s.handlePing)
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
		ID           string   `json:"id"`
		Endpoints    []string `json:"endpoints"`
		Degree       int      `json:"degree"`
		LatencyMs    int64    `json:"latency_ms"`
		IDCap        string   `json:"ID"`
		EndpointsCap []string `json:"Endpoints"`
		DegreeCap    int      `json:"Degree"`
		LatencyMsCap int64    `json:"LatencyMs"`
	}

	var list []PeerDTO
	for _, p := range peers {
		lat := p.Latency.Milliseconds()
		if lat <= 0 {
			lat = 1
		}
		list = append(list, PeerDTO{
			ID:           p.Identity.String(),
			Endpoints:    p.Endpoints,
			Degree:       p.Degree,
			LatencyMs:    lat,
			IDCap:        p.Identity.String(),
			EndpointsCap: p.Endpoints,
			DegreeCap:    p.Degree,
			LatencyMsCap: lat,
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
		// Lookup recipient's authentic X25519 key
		recipEncKey := s.node.GetPeerEncKey(targetID)
		if len(recipEncKey) != 32 {
			// If not yet cached, attempt handshake with peer to acquire authentic X25519 key
			eps := s.node.GetPeerEndpoints(targetID)
			if len(eps) == 0 && req.Endpoint != "" {
				eps = []string{req.Endpoint}
			}
			if len(eps) > 0 {
				_, _, _ = s.node.Handshake(eps[0])
				recipEncKey = s.node.GetPeerEncKey(targetID)
			}
		}

		if len(recipEncKey) == 32 {
			err = s.node.SendEncryptedMessage(targetID, recipEncKey, payload)
		} else {
			// Fallback: send authenticated plaintext if peer X25519 key is unavailable
			err = s.node.SendMessage(targetID, payload)
		}
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

type BroadcastStreamReq struct {
	StreamID string `json:"stream_id"`
	Payload  string `json:"payload"`
}

type StreamTelemetryFrame struct {
	Timestamp int64  `json:"timestamp" cbor:"1,keyasint"`
	Origin    string `json:"origin" cbor:"2,keyasint"`
	Sequence  uint64 `json:"seq" cbor:"3,keyasint"`
	SHA256    string `json:"sha256" cbor:"4,keyasint"`
	Payload   []byte `json:"payload" cbor:"5,keyasint"`
}

func (s *Server) handleBroadcastStream(w http.ResponseWriter, r *http.Request) {
	if s.cascade == nil {
		http.Error(w, "Cascade node not enabled", http.StatusBadRequest)
		return
	}

	streamID := "live-ui-stream"
	var rawData []byte

	if r.Method == http.MethodPost && r.Body != nil {
		var req BroadcastStreamReq
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.Payload != "" {
			if req.StreamID != "" {
				streamID = req.StreamID
			}
			rawData = []byte(req.Payload)
		}
	}

	if len(rawData) == 0 {
		rawData = []byte(fmt.Sprintf("telemetry_epoch_%d", time.Now().UnixNano()))
	}

	hash := sha256.Sum256(rawData)
	seq := uint64(time.Now().UnixNano())

	frame := StreamTelemetryFrame{
		Timestamp: time.Now().Unix(),
		Origin:    s.node.Identity.String(),
		Sequence:  seq,
		SHA256:    hex.EncodeToString(hash[:]),
		Payload:   rawData,
	}

	encoded, err := cbor.Marshal(frame)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = s.cascade.Broadcast(streamID, seq, encoded)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "broadcasted",
		"stream_id": streamID,
		"seq":       seq,
		"sha256":    frame.SHA256,
		"bytes":     len(encoded),
	})
}

type PingReq struct {
	TargetID string `json:"target_id"`
}

func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PingReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pubBytes, err := hex.DecodeString(req.TargetID)
	if err != nil || len(pubBytes) != 32 {
		http.Error(w, "Invalid target identity (must be 32-byte hex)", http.StatusBadRequest)
		return
	}

	targetID, err := core.NewIdentityFromBytes(pubBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rtt, err := s.node.PingPeer(targetID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusGatewayTimeout)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"target":  req.TargetID,
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"target":  req.TargetID,
		"rtt_ms":  rtt.Milliseconds(),
		"success": true,
	})
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

type MeshNode struct {
	ID        string   `json:"id"`
	ShortID   string   `json:"short_id"`
	Label     string   `json:"label"`
	IsLocal   bool     `json:"is_local"`
	Endpoints []string `json:"endpoints"`
	Degree    int      `json:"degree"`
	E2EE      bool     `json:"e2ee"`
}

type MeshLink struct {
	Source    string `json:"source"`
	Target    string `json:"target"`
	Adapter   string `json:"adapter"`
	LatencyMs int64  `json:"latency_ms"`
	Encrypted bool   `json:"encrypted"`
}

type MeshGraphResponse struct {
	LocalID string     `json:"local_id"`
	Nodes   []MeshNode `json:"nodes"`
	Links   []MeshLink `json:"links"`
}

func (s *Server) handleMesh(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	peers := s.node.SmallWorld.FindClosestPeers(s.node.Identity, 100)

	localIDStr := s.node.Identity.String()
	shortLocal := localIDStr
	if len(shortLocal) > 8 {
		shortLocal = shortLocal[:8] + "..."
	}

	nodes := []MeshNode{
		{
			ID:        localIDStr,
			ShortID:   shortLocal,
			Label:     "Nodo Local",
			IsLocal:   true,
			Endpoints: []string{fmt.Sprintf("127.0.0.1:%d", s.port)},
			Degree:    s.node.SmallWorld.CalculateDegree(s.node.Identity),
			E2EE:      s.node.EncPubKey != nil,
		},
	}

	var links []MeshLink
	for _, p := range peers {
		pIDStr := p.Identity.String()
		shortPeer := pIDStr
		if len(shortPeer) > 8 {
			shortPeer = shortPeer[:8] + "..."
		}

		nodes = append(nodes, MeshNode{
			ID:        pIDStr,
			ShortID:   shortPeer,
			Label:     fmt.Sprintf("Peer (D=%d)", p.Degree),
			IsLocal:   false,
			Endpoints: p.Endpoints,
			Degree:    p.Degree,
			E2EE:      true,
		})

		lat := p.Latency.Milliseconds()
		if lat <= 0 {
			lat = 1
		}

		links = append(links, MeshLink{
			Source:    localIDStr,
			Target:    pIDStr,
			Adapter:   "QUIC/UDP",
			LatencyMs: lat,
			Encrypted: true,
		})
	}

	_ = json.NewEncoder(w).Encode(MeshGraphResponse{
		LocalID: localIDStr,
		Nodes:   nodes,
		Links:   links,
	})
}

func (s *Server) handleMeshCypher(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	peers := s.node.SmallWorld.FindClosestPeers(s.node.Identity, 100)
	localIDStr := s.node.Identity.String()

	var cypher string
	cypher += fmt.Sprintf("MERGE (p:Peer {id: '%s', endpoint: '127.0.0.1:%d', is_local: true});\n", localIDStr, s.port)

	for _, p := range peers {
		pIDStr := p.Identity.String()
		ep := "unknown"
		if len(p.Endpoints) > 0 {
			ep = p.Endpoints[0]
		}
		lat := p.Latency.Milliseconds()
		if lat <= 0 {
			lat = 1
		}

		cypher += fmt.Sprintf("MERGE (p:Peer {id: '%s', endpoint: '%s', is_local: false});\n", pIDStr, ep)
		cypher += fmt.Sprintf("MATCH (a:Peer {id: '%s'}), (b:Peer {id: '%s'}) MERGE (a)-[:CONNECTED_TO {adapter: 'QUIC/UDP', latency_ms: %d, encrypted: true}]->(b);\n", localIDStr, pIDStr, lat)
	}

	_, _ = w.Write([]byte(cypher))
}
