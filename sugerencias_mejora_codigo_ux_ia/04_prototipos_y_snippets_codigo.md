# Prototipos y Snippets de Código en Go

Este documento contiene implementaciones de referencia y prototipos en lenguaje Go listos para ser adaptados o integrados en el código de IPv7.

---

## Prototipo 1: Servidor MCP (Model Context Protocol) Nativo [✅ Implementado en core/mcp.go]

Este diseño fue implementado e integrado formalmente en [`core/mcp.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/mcp.go) y en el endpoint HTTP `/api/mcp`:

```go
package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"ipv7/core"
	"os"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type MCPServer struct {
	node *core.Node
}

func NewMCPServer(node *core.Node) *MCPServer {
	return &MCPServer{node: node}
}

func (s *MCPServer) ServeStdio() {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		var req JSONRPCRequest
		if err := decoder.Decode(&req); err != nil {
			if err == io.EOF {
				return
			}
			continue
		}

		resp := s.handleRequest(&req)
		_ = encoder.Encode(resp)
	}
}

func (s *MCPServer) handleRequest(req *JSONRPCRequest) *JSONRPCResponse {
	switch req.Method {
	case "tools/list":
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"tools": []map[string]interface{}{
					{
						"name":        "ipv7_get_info",
						"description": "Retorna identidad criptográfica y estado del nodo local",
						"inputSchema": map[string]interface{}{"type": "object"},
					},
					{
						"name":        "ipv7_send_message",
						"description": "Envía un mensaje P2P cifrado o plano",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"recipient_id": map[string]string{"type": "string"},
								"message":      map[string]string{"type": "string"},
							},
							"required": []string{"recipient_id", "message"},
						},
					},
				},
			},
		}

	case "tools/call":
		var callParams struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		_ = json.Unmarshal(req.Params, &callParams)

		if callParams.Name == "ipv7_get_info" {
			return &JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"identity":    s.node.Identity.String(),
					"peers_count": s.node.SmallWorld.TotalPeers(),
				},
			}
		}

		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]string{"status": "executed"},
		}

	default:
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   map[string]interface{}{"code": -32601, "message": "Method not found"},
		}
	}
}
```

---

## Prototipo 2: Logging Estructurado con `log/slog` Nativo [✅ Implementado en core/logger.go]

Implementado e integrado en [`core/logger.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/logger.go) con el flag `-log-json`:

```go
package logging

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func InitLogger(asJSON bool, debugLevel bool) {
	var level slog.Level = slog.LevelInfo
	if debugLevel {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if asJSON {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}

// Ejemplo de uso en el ciclo de vida del nodo:
func ExampleLogUsage(peerID string, latencyMs int64) {
	slog.Info("peer_connected",
		slog.String("peer_id", peerID),
		slog.Int64("rtt_ms", latencyMs),
		slog.String("subsystem", "small_world"),
	)
}
```

---

## Prototipo 3: Mapeo Automático de Puertos con UPnP [✅ Implementado en adapters/upnp.go]

Implementado en [`adapters/upnp.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/upnp.go) e integrado con flag `-upnp` en el nodo principal:

```go
package adapters

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/huin/goupnp/dcps/internetgateway2"
)

// RequestUPnPPortMapping solicita al router local abrir el puerto UDP especificado
func RequestUPnPPortMapping(port uint16, desc string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	clients, _, err := internetgateway2.NewWANIPConnection1Clients()
	if err != nil || len(clients) == 0 {
		return fmt.Errorf("no se encontraron clientes UPnP WAN IP en la red local")
	}

	client := clients[0]
	// Obtener IP local en la interfaz
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return err
	}
	defer conn.Close()
	localIP := conn.LocalAddr().(*net.UDPAddr).IP.String()

	// Añadir mapeo de puerto en el router: Puerto Externo -> IP Local:Puerto Interno
	err = client.AddPortMappingCtx(ctx,
		"",           // Nueva IP remota (cualquiera)
		port,         // Puerto externo
		"UDP",        // Protocolo
		port,         // Puerto interno
		localIP,      // Cliente interno
		true,         // Habilitado
		desc,         // Descripción del servicio ("IPv7 Node")
		3600,         // Duración del arrendamiento (en segundos)
	)
	return err
}
```

---

## Prototipo 4: Persistencia y Almacenamiento de Claves en Disco (`keystore.go`) [✅ Implementado en core/keystore.go]

Implementado en [`core/keystore.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/keystore.go) y utilizable mediante `-key persistent`:

```go
package core

import (
	"crypto/ed25519"
	"encoding/hex"
	"os"
	"path/filepath"
)

// LoadOrCreatePersistentIdentity busca o genera una clave privada en el directorio de usuario
func LoadOrCreatePersistentIdentity(customPath string) (Identity, ed25519.PrivateKey, error) {
	if customPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, nil, err
		}
		dir := filepath.Join(home, ".ipv7")
		_ = os.MkdirAll(dir, 0700)
		customPath = filepath.Join(dir, "identity.key")
	}

	// 1. Si el archivo existe, cargar la clave existente
	if data, err := os.ReadFile(customPath); err == nil {
		seed, hexErr := hex.DecodeString(string(data))
		if hexErr == nil && len(seed) == ed25519.SeedSize {
			privKey := ed25519.NewKeyFromSeed(seed)
			pubKey := privKey.Public().(ed25519.PublicKey)
			id, _ := NewIdentityFromBytes(pubKey)
			return id, privKey, nil
		}
	}

	// 2. Si no existe o está corrupto, generar nueva clave y persistir
	id, privKey, err := GenerateIdentity()
	if err != nil {
		return nil, nil, err
	}

	seed := privKey.Seed()
	encoded := hex.EncodeToString(seed)
	_ = os.WriteFile(customPath, []byte(encoded), 0600)

	return id, privKey, nil
}
```
