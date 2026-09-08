package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// handleOpenAPI returns a dynamic OpenAPI 3.1 schema of the IPv7 node HTTP APIs
func (s *Server) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	schema := map[string]interface{}{
		"openapi": "3.1.0",
		"info": map[string]interface{}{
			"title":       "IPv7 Node API",
			"version":     "1.0.0",
			"description": "API REST para consulta de estado, topología de red, mensajería y servicios P2P de IPv7",
		},
		"paths": map[string]interface{}{
			"/api/info": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Obtener información del nodo local",
					"description": "Retorna la clave pública Ed25519, clave E2EE X25519 y contadores del nodo.",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Información del nodo",
						},
					},
				},
			},
			"/api/peers": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Listar peers de la red",
					"description": "Retorna la lista de peers conocidos y sus latencias.",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Lista de peers",
						},
					},
				},
			},
			"/api/mesh": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Obtener grafo de malla completo",
					"description": "Retorna nodos y aristas de la red para visualización.",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Grafo de malla",
						},
					},
				},
			},
			"/api/send": map[string]interface{}{
				"post": map[string]interface{}{
					"summary":     "Enviar mensaje P2P",
					"description": "Transmite un mensaje cifrado o plano a un peer.",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Mensaje enviado",
						},
					},
				},
			},
			"/api/ping": map[string]interface{}{
				"post": map[string]interface{}{
					"summary":     "Ping criptográfico a un peer",
					"description": "Mide la latencia RTT hacia el peer especificado.",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Resultado de latencia",
						},
					},
				},
			},
			"/api/tunnel/start": map[string]interface{}{
				"post": map[string]interface{}{
					"summary":     "Crear túnel de reenvío de puertos",
					"description": "Mapea un puerto TCP local a un puerto remoto en un peer.",
				},
			},
			"/api/vpn/start": map[string]interface{}{
				"post": map[string]interface{}{
					"summary":     "Iniciar proxy VPN SOCKS5",
					"description": "Levanta el proxy universal en espacio de usuario.",
				},
			},
			"/api/mcp": map[string]interface{}{
				"post": map[string]interface{}{
					"summary":     "Model Context Protocol JSON-RPC Endpoint",
					"description": "Procesa llamadas a herramientas y consultas de IA mediante protocolo MCP.",
				},
			},
			"/api/events": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Server-Sent Events (SSE) Stream",
					"description": "Flujo unidireccional continuo de eventos en tiempo real para agentes IA y dashboards.",
				},
			},
			"/api/self-healing/status": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Estado de Autocuración de Malla",
					"description": "Retorna métricas de auditoría Small-World, anillos vacíos y acciones de reparación.",
				},
			},
			"/metrics": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Métricas en formato Prometheus",
					"description": "Retorna contadores de red y salud del nodo para observabilidad.",
				},
			},
		},
	}

	_ = json.NewEncoder(w).Encode(schema)
}

// handleMetrics returns Prometheus-compatible metrics text
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	peersCount := s.node.SmallWorld.TotalPeers()
	cascadeCount := 0
	if s.cascade != nil {
		cascadeCount = s.cascade.ChildrenCount()
	}

	activeTunnels := 0
	if s.tunnel != nil {
		activeTunnels = len(s.tunnel.ActiveForwarders())
	}

	vpnRunning := 0
	s.mu.Lock()
	if s.vpnProxy != nil {
		vpnRunning = 1
	}
	s.mu.Unlock()

	var healActions uint64
	var degradedRings int
	if s.supervisor != nil {
		healActions, degradedRings, _ = s.supervisor.Stats()
	}

	ts := time.Now().Unix()

	out := fmt.Sprintf(
		"# HELP ipv7_peers_connected_total Total connected peers in small-world table\n"+
			"# TYPE ipv7_peers_connected_total gauge\n"+
			"ipv7_peers_connected_total %d %d\n\n"+
			"# HELP ipv7_cascade_children_count Active children in cascade streaming tree\n"+
			"# TYPE ipv7_cascade_children_count gauge\n"+
			"ipv7_cascade_children_count %d %d\n\n"+
			"# HELP ipv7_active_tunnels_count Number of active P2P port forwarders\n"+
			"# TYPE ipv7_active_tunnels_count gauge\n"+
			"ipv7_active_tunnels_count %d %d\n\n"+
			"# HELP ipv7_vpn_running Status of local SOCKS5 proxy (1 running, 0 stopped)\n"+
			"# TYPE ipv7_vpn_running gauge\n"+
			"ipv7_vpn_running %d %d\n\n"+
			"# HELP ipv7_self_healing_actions_total Total autonomous topology repairs executed\n"+
			"# TYPE ipv7_self_healing_actions_total counter\n"+
			"ipv7_self_healing_actions_total %d %d\n\n"+
			"# HELP ipv7_degraded_rings_count Number of unpopulated small-world distance rings\n"+
			"# TYPE ipv7_degraded_rings_count gauge\n"+
			"ipv7_degraded_rings_count %d %d\n\n"+
			"# HELP ipv7_node_uptime_seconds Process uptime\n"+
			"# TYPE ipv7_node_uptime_seconds counter\n"+
			"ipv7_node_uptime_seconds %d\n",
		peersCount, ts,
		cascadeCount, ts,
		activeTunnels, ts,
		vpnRunning, ts,
		healActions, ts,
		degradedRings, ts,
		ts,
	)

	_, _ = w.Write([]byte(out))
}

// handleMCP processes JSON-RPC MCP requests via HTTP POST
func (s *Server) handleMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if s.mcpServer != nil {
		resp := s.mcpServer.HandleMessage(body)
		if resp != nil {
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Error(w, "MCP server not initialized", http.StatusInternalServerError)
}
