# IPv7 Telemetry — Catálogo de Métricas

**Documento**: `docs/telemetry/METRICS.md`

---

## 1. Categoría A: Node Metrics (Métricas de Nodo)

| Métrica | Tipo | Descripción |
| :--- | :--- | :--- |
| `ipv7_uptime_seconds` | Gauge | Tiempo de actividad continuo del runtime del nodo en segundos |
| `ipv7_peers` | Gauge | Número de pares remotos con identidad Noise XX verificada |
| `ipv7_sessions_active` | Gauge | Total de túneles E2EE actualmente establecidos |
| `ipv7_direct_sessions` | Gauge | Sesiones operando en camino directo punto a punto UDP |
| `ipv7_relay_sessions` | Gauge | Sesiones operando a través de DERP Relay ciego |
| `ipv7_memory_alloc_bytes` | Gauge | Memoria activa en el heap de Go runtime |
| `ipv7_memory_sys_bytes` | Gauge | Memoria virtual total asignada por el sistema operativo |
| `ipv7_goroutines` | Gauge | Número de goroutines activas en el proceso |

---

## 2. Categoría B: Transport Metrics (Métricas de Transporte)

| Métrica | Tipo | Descripción |
| :--- | :--- | :--- |
| `ipv7_packets_tx_total` | Counter | Total de paquetes IPv7 transmitidos |
| `ipv7_packets_rx_total` | Counter | Total de paquetes IPv7 recibidos y verificados |
| `ipv7_bytes_tx_total` | Counter | Volumen de bytes útiles transmitidos |
| `ipv7_bytes_rx_total` | Counter | Volumen de bytes útiles recibidos |
| `ipv7_pps_rate` | Gauge | Tasa de paquetes por segundo calculada en ventana móvil |
| `ipv7_throughput_mbps` | Gauge | Rendimiento medido en megabits por segundo |
| `ipv7_rtt_ms` | Gauge | Latencia de ida y vuelta (RTT) estimada en milisegundos |
| `ipv7_jitter_ms` | Gauge | Dispersión estadística de retardo de entrega |
| `ipv7_pmtu_bytes` | Gauge | Path MTU activo verificado (por defecto 1280 B) |
| `ipv7_socket_queue_drops_total` | Counter | Paquetes descartados en la cola de socket del SO por contención |
| `ipv7_backpressure_drops_total` | Counter | Paquetes descartados por política adaptativa antes de saturar |
| `ipv7_endpoint_changes_total` | Counter | Número de eventos de roaming (cambio de IP/puerto) completados |

---

## 3. Categoría C: Protocol Metrics (Métricas de Protocolo)

| Métrica | Tipo | Descripción |
| :--- | :--- | :--- |
| `ipv7_handshakes_completed_total` | Counter | Handshakes de autenticación mutua Ed25519/X25519 exitosos |
| `ipv7_handshakes_failed_total` | Counter | Intentos de handshake fallidos o expirados por timeout |
| `ipv7_replay_rejected_total` | Counter | Paquetes duplicados descartados por ventana anti-replay RFC 6479 |
| `ipv7_malformed_rejected_total` | Counter | Paquetes descartados por formato CBOR no conforme o tamaño > PMTU |
| `ipv7_telemetry_dropped_total` | Counter | Eventos de telemetría descartados para priorizar el tráfico de red |
