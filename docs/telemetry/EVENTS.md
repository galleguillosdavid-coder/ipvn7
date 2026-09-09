# IPv7 Telemetry — Catálogo de Eventos Operacionales

**Documento**: `docs/telemetry/EVENTS.md`

---

## 1. Eventos de Ciclo de Vida y Sesión

- **`NODE_START`**: Inicialización del runtime y enlace de adapters.
- **`NODE_STOP`**: Apagado ordenado del nodo y cierre de sockets.
- **`PEER_JOINED`**: Nuevo peer autenticado mutuamente mediante handshake Ed25519.
- **`PEER_LEFT`**: Peer desconectado por timeout o señal explícita de cierre.
- **`SESSION_CREATED`**: Apertura de un nuevo canal cifrado E2EE (ChaCha20-Poly1305).
- **`SESSION_CLOSED`**: Destrucción de canal y borrado seguro de claves efímeras.

---

## 2. Eventos de Red, Movilidad y Ruta

- **`ENDPOINT_CHANGED`**: Detección de cambio de interfaz física, Wi-Fi o NAT (IP Roaming).
- **`ROUTE_CHANGED`**: Actualización de la tabla Kademlia ante convergencia topológica.
- **`PMTU_CHANGED`**: Ajuste del tamaño máximo de paquete seguro tras sondeo.
- **`DIRECT_PATH_LOST`**: Pérdida de conectividad UDP directa (e.g. bloqueo por firewall).
- **`RELAY_ENTERED`**: Conmutación automática y transparente al DERP Relay.
- **`RELAY_EXITED`**: Restauración del camino UDP directo y desconexión del relay.

---

## 3. Eventos de Presión y Diagnóstico

- **`BACKPRESSURE_DETECTED`**: Activación de descarte temprano ante llegada al techo de tasa.
- **`SOCKET_PRESSURE`**: Detección de contención en la cola del buffer `SO_RCVBUF` del sistema operativo.
- **`ANOMALY_DETECTED`**: Disparo de regla de alerta (spikes de RTT, pérdidas o memoria).
- **`RECOVERY_COMPLETED`**: Restablecimiento exitoso de estado tras una disrupción.
