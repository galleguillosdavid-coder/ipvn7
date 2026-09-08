package core

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
)

// MCP JSON-RPC 2.0 structs
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCPServer exposes IPv7 node operations via Model Context Protocol (JSON-RPC 2.0)
type MCPServer struct {
	node           *Node
	tunnel         *TunnelService
	vpn            *SOCKS5Proxy
	cypherExecutor func(string) (string, error)
	mu             sync.Mutex
}

// NewMCPServer initializes a new MCP server attached to an IPv7 Node
func NewMCPServer(node *Node, tunnel *TunnelService) *MCPServer {
	return &MCPServer{
		node:   node,
		tunnel: tunnel,
	}
}

// SetCypherExecutor registers the callback to run Cypher queries on Kùzu
func (s *MCPServer) SetCypherExecutor(fn func(string) (string, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cypherExecutor = fn
}

// SetVPNProxy assigns an existing SOCKS5 VPN instance
func (s *MCPServer) SetVPNProxy(vpn *SOCKS5Proxy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vpn = vpn
}

// HandleMessage processes a single JSON-RPC request and returns the response
func (s *MCPServer) HandleMessage(raw []byte) *MCPResponse {
	var req MCPRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return &MCPResponse{
			JSONRPC: "2.0",
			Error:   &MCPError{Code: -32700, Message: "Parse error"},
		}
	}
	return s.Dispatch(&req)
}

// Dispatch executes the appropriate RPC method
func (s *MCPServer) Dispatch(req *MCPRequest) *MCPResponse {
	switch req.Method {
	case "initialize":
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]interface{}{
					"tools": map[string]bool{"listChanged": false},
				},
				"serverInfo": map[string]string{
					"name":    "ipv7-node-mcp",
					"version": "1.0.0",
				},
			},
		}

	case "notifications/initialized":
		return nil

	case "ping":
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]interface{}{},
		}

	case "tools/list":
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"tools": []map[string]interface{}{
					{
						"name":        "ipv7_get_info",
						"description": "Obtiene la identidad criptográfica y estado operativo del nodo local IPv7.",
						"inputSchema": map[string]interface{}{
							"type":       "object",
							"properties": map[string]interface{}{},
						},
					},
					{
						"name":        "ipv7_list_peers",
						"description": "Lista los peers conocidos en la tabla de enrutamiento del Mundo Pequeño.",
						"inputSchema": map[string]interface{}{
							"type":       "object",
							"properties": map[string]interface{}{},
						},
					},
					{
						"name":        "ipv7_send_message",
						"description": "Envía un mensaje de texto plano o cifrado (E2EE) a un peer por su clave Ed25519.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"recipient_id": map[string]string{"type": "string", "description": "Clave pública Ed25519 en hex (64 chars)"},
								"message":      map[string]string{"type": "string", "description": "Mensaje de texto"},
								"encrypted":    map[string]string{"type": "boolean", "description": "Cifrar con ChaCha20-Poly1305"},
							},
							"required": []string{"recipient_id", "message"},
						},
					},
					{
						"name":        "ipv7_ping_peer",
						"description": "Mide el RTT en milisegundos hacia un peer.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"target_id": map[string]string{"type": "string", "description": "Clave hex del peer objetivo"},
							},
							"required": []string{"target_id"},
						},
					},
					{
						"name":        "ipv7_start_tunnel",
						"description": "Crea un túnel P2P cifrado mapeando un puerto local a un servicio remoto en un peer.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"local_port":  map[string]string{"type": "integer", "description": "Puerto TCP local a abrir (ej: 13389, 8088)"},
								"peer_id":     map[string]string{"type": "string", "description": "Clave pública Ed25519 en hex (64 caracteres)"},
								"target_port": map[string]string{"type": "integer", "description": "Puerto TCP destino en el peer remoto (ej: 3389, 80)"},
							},
							"required": []string{"local_port", "peer_id", "target_port"},
						},
					},
					{
						"name":        "ipv7_toggle_vpn",
						"description": "Inicia o detiene el proxy VPN SOCKS5 en espacio de usuario.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"action": map[string]string{"type": "string", "description": "'start' para iniciar o 'stop' para detener"},
								"port":   map[string]string{"type": "integer", "description": "Puerto local SOCKS5 (por defecto 1080)"},
							},
						},
					},
					{
						"name":        "ipv7_query_kuzu",
						"description": "Ejecuta una consulta Cypher sobre la base de datos de grafos Kùzu (.kuzu_index/) para consultar topología o arquitectura de código.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"query": map[string]string{"type": "string", "description": "Sentencia Cypher a ejecutar (ej: 'MATCH (p:Peer) RETURN p;')"},
							},
							"required": []string{"query"},
						},
					},
				},
			},
		}

	case "tools/call":
		var params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return &MCPResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error:   &MCPError{Code: -32602, Message: "Invalid params"},
			}
		}

		result, err := s.executeTool(params.Name, params.Arguments)
		if err != nil {
			return &MCPResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"content": []map[string]string{
						{"type": "text", "text": fmt.Sprintf("Error: %v", err)},
					},
					"isError": true,
				},
			}
		}

		resBytes, _ := json.Marshal(result)
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"content": []map[string]string{
					{"type": "text", "text": string(resBytes)},
				},
			},
		}

	default:
		return &MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &MCPError{Code: -32601, Message: fmt.Sprintf("Method '%s' not found", req.Method)},
		}
	}
}

func (s *MCPServer) executeTool(name string, args map[string]interface{}) (interface{}, error) {
	switch name {
	case "ipv7_get_info":
		var e2eePub string
		if s.node.EncPubKey != nil {
			e2eePub = hex.EncodeToString(s.node.EncPubKey.Bytes())
		}
		return map[string]interface{}{
			"identity":    s.node.Identity.String(),
			"e2ee_pubkey": e2eePub,
			"peers_count": s.node.SmallWorld.TotalPeers(),
			"endpoints":   s.node.Endpoints(),
		}, nil

	case "ipv7_list_peers":
		peers := s.node.SmallWorld.FindClosestPeers(s.node.Identity, 50)
		type peerSummary struct {
			ID        string   `json:"id"`
			Endpoints []string `json:"endpoints"`
			LatencyMs int64    `json:"latency_ms"`
			Degree    int      `json:"degree"`
		}
		var list []peerSummary
		for _, p := range peers {
			list = append(list, peerSummary{
				ID:        p.Identity.String(),
				Endpoints: p.Endpoints,
				LatencyMs: p.Latency.Milliseconds(),
				Degree:    p.Degree,
			})
		}
		return list, nil

	case "ipv7_send_message":
		recipHex, _ := args["recipient_id"].(string)
		msgText, _ := args["message"].(string)
		isEnc, _ := args["encrypted"].(bool)

		recipBytes, err := hex.DecodeString(recipHex)
		if err != nil || len(recipBytes) != 32 {
			return nil, fmt.Errorf("invalid recipient hex identity")
		}
		targetID, err := NewIdentityFromBytes(recipBytes)
		if err != nil {
			return nil, err
		}

		if isEnc {
			recipEncKey := s.node.GetPeerEncKey(targetID)
			if len(recipEncKey) == 32 {
				err = s.node.SendEncryptedMessage(targetID, recipEncKey, []byte(msgText))
			} else {
				err = s.node.SendMessage(targetID, []byte(msgText))
			}
		} else {
			err = s.node.SendMessage(targetID, []byte(msgText))
		}
		if err != nil {
			return nil, err
		}
		return map[string]string{"status": "delivered"}, nil

	case "ipv7_ping_peer":
		targetHex, _ := args["target_id"].(string)
		targetBytes, err := hex.DecodeString(targetHex)
		if err != nil || len(targetBytes) != 32 {
			return nil, fmt.Errorf("invalid target hex identity")
		}
		targetID, err := NewIdentityFromBytes(targetBytes)
		if err != nil {
			return nil, err
		}
		rtt, err := s.node.PingPeer(targetID)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"target": targetHex,
			"rtt_ms": rtt.Milliseconds(),
		}, nil

	case "ipv7_start_tunnel", "ipv7_create_tunnel":
		if s.tunnel == nil {
			return nil, fmt.Errorf("tunnel service not available on this node")
		}
		localPortFloat, ok1 := args["local_port"].(float64)
		peerHex, ok2 := args["peer_id"].(string)
		targetPortFloat, ok3 := args["target_port"].(float64)
		if !ok1 || !ok2 || !ok3 {
			return nil, fmt.Errorf("local_port, peer_id, and target_port are required")
		}
		localPort := int(localPortFloat)
		targetPort := int(targetPortFloat)

		peerBytes, err := hex.DecodeString(peerHex)
		if err != nil || len(peerBytes) != 32 {
			return nil, fmt.Errorf("invalid peer_id hex identity")
		}
		peerID, err := NewIdentityFromBytes(peerBytes)
		if err != nil {
			return nil, err
		}

		if err := s.tunnel.ForwardPort(localPort, peerID, targetPort); err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"status":      "active",
			"local_port":  localPort,
			"peer_id":     peerHex,
			"target_port": targetPort,
			"address":     fmt.Sprintf("127.0.0.1:%d", localPort),
		}, nil

	case "ipv7_toggle_vpn", "ipv7_start_vpn":
		action, _ := args["action"].(string)
		if action == "" {
			action = "start"
		}
		port := 1080
		if portFloat, ok := args["port"].(float64); ok && portFloat > 0 {
			port = int(portFloat)
		}

		s.mu.Lock()
		defer s.mu.Unlock()

		if action == "stop" {
			if s.vpn != nil {
				s.vpn.Stop()
				s.vpn = nil
				return map[string]string{"status": "stopped"}, nil
			}
			return map[string]string{"status": "already_stopped"}, nil
		}

		// Start VPN
		if s.vpn != nil {
			return map[string]interface{}{
				"status":  "running",
				"address": s.vpn.listenAddr,
			}, nil
		}
		listenAddr := fmt.Sprintf("127.0.0.1:%d", port)
		vpn := NewSOCKS5Proxy(listenAddr, s.tunnel)
		if err := vpn.Start(); err != nil {
			return nil, fmt.Errorf("failed to start SOCKS5 proxy: %w", err)
		}
		s.vpn = vpn
		return map[string]interface{}{
			"status":  "running",
			"address": listenAddr,
		}, nil

	case "ipv7_query_kuzu":
		query, ok := args["query"].(string)
		if !ok || strings.TrimSpace(query) == "" {
			return nil, fmt.Errorf("query string is required")
		}

		s.mu.Lock()
		executor := s.cypherExecutor
		s.mu.Unlock()

		if executor != nil {
			out, err := executor(query)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"query":  query,
				"result": out,
			}, nil
		}

		return map[string]interface{}{
			"query":  query,
			"notice": "Kuzu executor not registered in standalone mode",
		}, nil

	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

// ServeStdio runs the MCP JSON-RPC event loop over standard I/O
func (s *MCPServer) ServeStdio(in io.Reader, out io.Writer) {
	dec := json.NewDecoder(in)
	enc := json.NewEncoder(out)

	for {
		var req MCPRequest
		if err := dec.Decode(&req); err != nil {
			if err == io.EOF {
				return
			}
			continue
		}

		resp := s.Dispatch(&req)
		if resp != nil {
			_ = enc.Encode(resp)
		}
	}
}
