# Referencia de APIs del Nodo IPv7 para Agentes de IA

Esta referencia describe los contratos de comunicación que expone el servidor local embebido del nodo IPv7 (`ui/server.go`). La IA puede consultar estos endpoints HTTP REST o conectarse mediante WebSockets para monitorear y controlar el nodo.

---

## 1. Endpoints de Estado y Topología

### `GET /api/info`
Retorna la identidad criptográfica y contadores clave del nodo local.

- **Respuesta (JSON)**:
  ```json
  {
    "identity": "8a73f1b2c3d4...",
    "e2ee_pubkey": "b4190c2e1f...",
    "peers_count": 5,
    "cascade_count": 2
  }
  ```

### `GET /api/peers`
Retorna la lista de peers conocidos en la tabla del Mundo Pequeño.

- **Respuesta (JSON Array)**:
  ```json
  [
    {
      "id": "fe44a98b1c2d...",
      "endpoints": ["192.168.1.50:7001", "181.42.10.88:45231"],
      "degree": 3,
      "latency_ms": 12
    }
  ]
  ```

### `GET /api/mesh`
Retorna el grafo completo de la malla local (nodos y aristas) en formato compatible con librerías de grafos.

- **Respuesta (JSON)**:
  ```json
  {
    "local_id": "8a73f1b2c3d4...",
    "nodes": [
      {
        "id": "8a73f1b2c3d4...",
        "short_id": "8a73f1b2...",
        "label": "Nodo Local",
        "is_local": true,
        "endpoints": ["127.0.0.1:8080"],
        "degree": 4,
        "e2ee": true
      }
    ],
    "links": [
      {
        "source": "8a73f1b2c3d4...",
        "target": "fe44a98b1c2d...",
        "adapter": "QUIC/UDP",
        "latency_ms": 14,
        "encrypted": true
      }
    ]
  }
  ```

### `GET /api/mesh/export-cypher`
Genera sentencias Cypher en texto plano (`MERGE (p:Peer...)`) para importar la topología actual directamente a Kùzu o Neo4j.

---

## 2. Endpoints de Operación y Mensajería

### `POST /api/send`
Envía un mensaje a un peer, con soporte opcional de cifrado ChaCha20-Poly1305.

- **Cuerpo de Solicitud (JSON)**:
  ```json
  {
    "recipient_id": "fe44a98b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f",
    "message": "Hola desde agente IA",
    "encrypted": true,
    "endpoint": "192.168.1.50:7001"
  }
  ```
- **Respuesta (JSON)**:
  ```json
  { "status": "sent" }
  ```

### `POST /api/ping`
Mide la latencia RTT (Round Trip Time) hacia un peer conocido.

- **Cuerpo de Solicitud (JSON)**:
  ```json
  { "target_id": "fe44a98b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f" }
  ```
- **Respuesta (JSON)**:
  ```json
  { "target": "fe44a98b...", "rtt_ms": 15, "success": true }
  ```

### `POST /api/broadcast-stream`
Difunde una carga útil de datos a través de la cascada de streaming (Fan-out 10).

- **Cuerpo de Solicitud (JSON)**:
  ```json
  { "stream_id": "telemetry-ai", "payload": "sensor_batch_data" }
  ```

---

## 3. Endpoints de Túneles y VPN

### `POST /api/tunnel/start`
Inicia un reenvío de puertos P2P seguro hacia un peer remoto.

- **Cuerpo de Solicitud (JSON)**:
  ```json
  {
    "local_port": 2222,
    "peer_id": "fe44a98b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f",
    "target_port": 22
  }
  ```
- **Respuesta (JSON)**:
  ```json
  { "status": "active", "local_port": 2222, "target_port": 22, "peer_id": "fe44a98b..." }
  ```

### `GET /api/tunnel/list`
Devuelve los puertos locales actualmente asociados a túneles activos.

- **Respuesta (JSON)**:
  ```json
  { "active_ports": [2222, 5432] }
  ```

### `POST /api/vpn/start`
Levanta el servidor proxy local SOCKS5 universal.

- **Cuerpo de Solicitud (JSON)**:
  ```json
  { "port": 1080 }
  ```

---

## 4. Endpoints de WebSockets

### `ws://localhost:8080/ws`
Canal de eventos en tiempo real:
- **`incoming_message`**: Notificación inmediata de recepción de mensaje (con indicador `encrypted: bool`).
- **`cascade_chunk`**: Notificación de recepción de fragmentos de telemetría o streaming en cascada.

### `ws://localhost:8080/ws/desktop`
Canal de streaming de escritorio remoto:
- Recibe frames JPEG en formato binario (`websocket.BinaryMessage`) a 15 FPS.
- Permite inyectar eventos de entrada JSON (`RemoteInputEvent`): `mousemove`, `mousedown`, `mouseup`, `wheel`, `keydown`, `keyup`.
