# IPv7 Canary Deployment Specification — CANARY-01

**Misión**: IPv7 CANARY-01: MULTI-REGION CONTROLLED DEPLOYMENT  
**Tag Base**: `IPv7-PRODUCTION-CANDIDATE-0`  
**Estado del Protocol Core**: CONGELADO (CORE FREEZE STRICT - Cero modificaciones permitidas)  
**Fecha de Publicación**: 2026-09-09  

---

## 1. Justificación y Objetivos de la Etapa Canaria

Habiendo alcanzado el dictamen formal de **`CONDITIONAL GO`** en la validación de laboratorio, LAN física y caracterización adversarial (Reportes 1 a 13 y Fases 1 a 15), el proyecto avanza a la etapa de despliegue real controlado:

```text
[Laboratorio / LAN] ✅ ──► [Adversarial & Resiliencia] ✅ ──► [CANARY CONTROLADO] ──► [Beta Pública] ──► [Producción Mundial]
```

### Objetivos Principales:
1. Validar la estabilidad del protocolo durante una operación continua de **24 horas a 7 días** (Soak prolongado).
2. Probar la interacción heterogénea entre nodos multi-región:
   - **Nodo Chile (Host Windows 11)**: `192.168.1.198` (Métricas: `http://127.0.0.1:9101/metrics`)
   - **Nodo Notebook Físico (Wi-Fi)**: `192.168.1.106` (Métricas: `http://192.168.1.106:9103/metrics`)
   - **Nodo Linux Ubuntu (WSL2)**: `127.0.0.1:7002` (Métricas: `http://127.0.0.1:9102/metrics`)
   - **Nodo Emulado USA East (WAN Proxy)**: `127.0.0.1:7004` (Métricas: `http://127.0.0.1:9104/metrics`)
   - **Nodo Emulado EU West (WAN Proxy)**: `127.0.0.1:7005` (Métricas: `http://127.0.0.1:9105/metrics`)
   - **Relay DERP Central Dedicado**: `192.168.1.198:7099` (Métricas: `http://127.0.0.1:9199/metrics`)
3. Garantizar que ningún nodo colapse ante sobrecarga mediante **Backpressure Estricto** en lugar de crash.

---

## 2. Límites Operativos por Nodo (Backpressure vs. Crash)

Cada nodo canario tiene configurados límites duros para proteger el socket y la memoria del sistema operativo:

| Parámetro | Límite Máximo | Comportamiento al Alcanzar el Límite |
| :--- | :--- | :--- |
| **`MAX_PEERS`** | 1.000 peers | Rechazo de nuevas solicitudes de autenticación |
| **`MAX_HANDSHAKES`** | 2.000 handshakes/s | Pacing y descarte con aviso al emisor |
| **`MAX_SESSIONS`** | 5.000 sesiones activas | Descarte de nuevas aperturas de túnel E2EE |
| **`MAX_PPS`** | 20.000 pps | Rate-limiting en capa de adapter |
| **`MAX_RELAY_PPS`** | 4.200 pps | Descarte probabilístico en cola DERP (`Early Drop`) |
| **`MAX_MEMORY_MB`** | 512 MB Heap | Forzado de ciclo GC e inhibición de buffers temporales |

---

## 3. Telemetría y Métricas Expuestas (Prometheus OpenMetrics)

Cada nodo expone un servidor HTTP local en su puerto asignado (`9101` a `9105`, `9199`):
- **`/metrics`**: Formato estándar de Prometheus para ingesta en Grafana / Prometheus Server.
- **`/health`**: JSON para verificaciones de liveness y readiness por parte de orquestadores y balanceadores.

### Métricas Clave Monitorizadas:
- `ipv7_uptime_seconds`: Tiempo de actividad continuo.
- `ipv7_peers_connected`: Número de pares autenticados Noise XX activos.
- `ipv7_routes_count`: Rutas activas en la tabla Kademlia / Mundo Pequeño.
- `ipv7_direct_sessions` vs. `ipv7_relay_sessions`: Proporción de sesiones directas vs intermediadas.
- `ipv7_pps_rate` y `ipv7_throughput_mbps`: Rendimiento de transporte en tiempo real.
- `ipv7_memory_alloc_bytes` y `ipv7_goroutines_count`: Monitoreo continuo de fugas de memoria y goroutines.
- `ipv7_adapter_drops_total`: Diagnóstico de saturación en buffer de socket UDP.

---

## 4. Watchdog y Auto-Recuperación (`canary/watchdog/`)

- El supervisor `NodeSupervisor` relanza el proceso del nodo ante cualquier salida intempestiva o congelamiento.
- Al reiniciar, la persistencia en base de datos Kùzu y las claves privadas Ed25519 garantizan que el nodo conserve su DID original y restablezca sus sesiones en menos de 500 ms.

---

## 5. Orquestador CANARY-01 (`canary/runner/`)

Para ejecutar la prueba de despliegue canario:
```powershell
go run canary/runner/canary_runner.go -duration=24h -interval=30s
```
El runner simula y supervisa los eventos de churn, conmutación de relay y tráfico sostenido, registrando las métricas sin alterar una sola línea del Protocol Core.
