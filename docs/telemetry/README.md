# IPv7 Telemetry & Living Network v1 — Arquitectura de Observabilidad

**Misión**: IPv7 Telemetry & Living Network v1  
**Estado del Protocol Core**: CONGELADO (CORE FREEZE STRICT - Cero bloqueos en camino crítico)  
**Fecha**: 2026-09-09  

---

## 1. ¿Qué es IPv7 Telemetry?

IPv7 Telemetry es una capa de observabilidad desacoplada de ultra bajo overhead (~25 ns por operación, 0 bytes asignados) diseñada para monitorear el protocolo IPv7 en condiciones reales de red durante horas o días ininterrumpidos.

### Principio Fundamental:
> **"La observabilidad nunca debe convertirse en un nuevo cuello de botella. Jamás bloquear el procesamiento o transmisión de paquetes para registrar una métrica o persistir en Kùzu."**

```text
                    IPv7 Network Wire
                            │
       ┌────────────────────┴────────────────────┐
       │                                         │
   Tráfico Normal                          Telemetry API
   (Sockets E2EE)                         (Eventos & Contadores)
       │                                         │
       ▼                                  Bounded Event Buffer
     Socket Wire                        (Descarte O(1) si lleno)
                                                 │
                                          Batch Aggregator
                                                 │
                             ┌───────────────────┴───────────────────┐
                             ▼                                       ▼
                    Prometheus Exporter                      Kùzu Graph Exporter
                    (OpenMetrics HTTP)                       (Memoria Relacional)
```

---

## 2. Los Cuatro Niveles de Telemetría

1. **Node Metrics**: Métricas agregadas de ciclo de vida del nodo (uptime, goroutines, memoria heap/sys, peers, sesiones directas y relay).
2. **Transport Metrics**: Conteo acumulativo de paquetes, bytes, tasa PPS, throughput Mbps, RTT medido, jitter agregado, PMTU seguro (1280 B) y drops de buffer de socket.
3. **Protocol Metrics**: Handshakes Noise XX completados/fallidos, sesiones creadas/destruidas, y descartes de seguridad (replay RFC 6479, firmas Ed25519 inválidas, paquetes corruptos).
4. **Events Significativos**: Eventos discretos de baja frecuencia (`NODE_START`, `PEER_JOINED`, `ENDPOINT_CHANGED`, `ROUTE_CHANGED`, `ANOMALY_DETECTED`, `RECOVERY_COMPLETED`).

---

## 3. Estructura de Documentación

- [`SCHEMA.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/telemetry/SCHEMA.md): Formato JSON y CBOR estructurado de eventos.
- [`METRICS.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/telemetry/METRICS.md): Catálogo completo de métricas de nodo, transporte y protocolo.
- [`EVENTS.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/telemetry/EVENTS.md): Taxonomía de eventos operacionales y ciclo de vida.
- [`KUZU_MODEL.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/telemetry/KUZU_MODEL.md): Esquema de nodos y relaciones en la base de datos de grafos Kùzu.
- [`PROMETHEUS.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/telemetry/PROMETHEUS.md): Exposición OpenMetrics y endpoints `/metrics` y `/health`.
- [`PERFORMANCE.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/telemetry/PERFORMANCE.md): Resultados formales del benchmark de overhead (25 ns/op, 0 allocs/op).
- [`OPERATIONS.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/telemetry/OPERATIONS.md): Guía operativa de monitoreo y alertas.
- [`IMPLEMENTATION_REPORT.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/telemetry/IMPLEMENTATION_REPORT.md): Reporte final clasificado con rigor epistémico.
