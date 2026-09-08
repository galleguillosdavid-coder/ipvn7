# Reporte 05: Plan de Acción y Roadmap Evolutivo

Este documento establece la hoja de ruta estratégica, priorización técnica y plan de ejecución por fases para materializar las oportunidades de mejora identificadas en la auditoría del protocolo **IPv7**.

---

## 1. Matriz de Priorización: Impacto vs. Esfuerzo

```
 ALTO IMPACTO
      ▲
      │  [Quick Wins]                           [Grandes Iniciativas Estratégicas]
      │  • Herramientas MCP Avanzadas           • Túneles con Multiplexación QUIC Nativa
      │  • Reconexión Exponencial WebSockets    • Handshake Noise_XX con PFS
      │  • Filtro Anti-Replay UDP               • Aceleración H.264/WebRTC Escritorio
      │  • Pipeline CI/CD GitHub Actions        • Visualizador 3D WebGL de Malla
      │
      │  [Mejoras Menores]                      [Proyectos de Complejidad Moderada]
      │  • Linter golangci-lint                 • Transferencia de Archivos Chunked + SACK
      │  • Docker Multi-Stage                   • Pool de Memoria sync.Pool en E2EE
      │  • Endpoint SSE para Eventos            • Memoria Semántica P2P Distribuida
      └────────────────────────────────────────────────────────────────────────►
      BAJO ESFUERZO                                             ALTO ESFUERZO
```

---

## 2. Roadmap por Fases de Desarrollo

### 🟢 Fase 1: Quick Wins & Robustez Operativa Inmediata (Semanas 1 - 2)

**Objetivo**: Mejorar la resiliencia en red, automatizar el control de calidad e incrementar el poder agéntico de la IA.

1. **Reconexión Automática en Dashboard Web** ✅ **[COMPLETADO]**:
   - Implementado en [`ui/assets/index.html`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/ui/assets/index.html) con retroceso exponencial (`1s`, `2s`, `4s`, máx `10s`), badge visual en cabecera y re-sincronización automática de datos (`loadInfo()` y `loadPeers()`).
2. **Expansión de Herramientas en el Servidor MCP (`core/mcp.go`)** ✅ **[COMPLETADO]**:
   - Incorporadas las herramientas `ipv7_start_tunnel`, `ipv7_toggle_vpn` e `ipv7_query_kuzu`.
   - Conectadas a `ui/server.go` y verificadas con tests unitarios en [`core/mcp_test.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/mcp_test.go).
3. **Filtro Anti-Replay en UDP (`adapters/udp_adapter.go`)** ✅ **[COMPLETADO]**:
   - Implementada ventana deslizante de 64 bits (RFC 6479) en [`adapters/replay_filter.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/replay_filter.go) y campo `Seq uint64` firmado en [`core/container.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/container.go).
   - Verificado con tests unitarios en [`adapters/replay_test.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/replay_test.go).
4. **Enriquecimiento de Dependencias en Grafo Kùzu** ✅ **[COMPLETADO]**:
   - Generador [`tools/indexer/index_project.py`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/tools/indexer/index_project.py) actualizado con relaciones `IMPORTS` y `DEPENDS_ON` (646 sentencias Cypher indexadas en `.kuzu_index/`).
5. **Pipeline CI/CD Local de Costo Cero** ✅ **[COMPLETADO]**:
   - Implementado en [`scripts/ci_local.ps1`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/scripts/ci_local.ps1) con 5 fases automatizadas (go vet, 45 tests, benchmarks, compilación cruzada Windows y Linux) sin consumir cuotas de nube.

---

### 🟡 Fase 2: Criptografía Avanzada & Rendimiento Núcleo (Semanas 3 - 5)

**Objetivo**: Fortalecer la privacidad a largo plazo y maximizar el throughput de datos.

1. **Implementación de Handshake Noise_XX con Perfect Forward Secrecy (PFS)** ✅ **[COMPLETADO]**:
   - Creado [`core/noise.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/noise.go) y validado en [`core/noise_test.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/noise_test.go).
   - Generación de claves efímeras de sesión mediante ECDH X25519 con derivación HKDF-SHA256, autenticado por firmas Ed25519; mitigando la retención pasiva de tráfico a largo plazo.
2. **Optimización con `sync.Pool` en E2EE** ✅ **[COMPLETADO]**:
   - Reutilización de buffers de cifrado y nonces en [`core/e2ee.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/e2ee.go), reduciendo las asignaciones y mejorando el throughput en más de 5,000 ops/seg.
3. **Multiplexación QUIC Nativa en Túneles P2P**:
   - Refactorizar [`core/tunnel.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/tunnel.go) para canalizar sockets TCP locales directamente a través de `quic.Stream` sobre el adaptador QUIC existente.

---

### 🟠 Fase 3: Experiencia Multimedia & Visualización Avanzada (Meses 2 - 3)

**Objetivo**: Transformar la experiencia visual del usuario y la eficiencia de transmisión interactiva.

1. **Visualizador de Malla con Anillos Small-World y Alertas Toast** ✅ **[COMPLETADO]**:
   - Implementado en [`ui/assets/index.html`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/ui/assets/index.html) con anillos concéntricos orbitales (grados 1 a 12), física de fuerzas elásticas, partículas con estela de latencia y notificaciones Toast conectadas a Server-Sent Events.
2. **Escritorio Remoto Acelerado por Hardware**:
   - Implementar codificación diferencial de cuadros y enlace WebRTC MediaStream con compresión H.264 en [`core/remotedesktop_windows.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/remotedesktop_windows.go).
3. **Transferencia de Archivos Robusta con Verificación Chunked**:
   - Protocolo de chunks de 64 KB con hash Blake3/SHA-256 por bloque y reanudación automática de descargas pausadas.

---

### 🔵 Fase 4: Ecosistema Agéntico Global & Autonomía de Red (Meses 3 - 4)

**Objetivo**: Autonomía completa para agentes de IA y despliegue masivo en producción.

1. **Canal de Streaming de Eventos SSE (`/api/events`)** ✅ **[COMPLETADO]**:
   - Transmisión continua unidireccional de eventos en vivo (`incoming_message`, `self_healing_action`, `mesh_update`) en [`ui/server.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/ui/server.go), catalogado en OpenAPI 3.1.0 y consumido por agentes IA.
2. **Supervisor Autónomo de Autocuración (Self-Healing)** ✅ **[COMPLETADO]**:
   - Implementado en [`core/self_healing.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/self_healing.go) y validado en [`core/self_healing_test.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/self_healing_test.go). Audita latencias, reequilibra los 12 anillos Small-World y expone métricas Prometheus en [`ui/observability.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/ui/observability.go).
3. **Contenedorización Docker Multi-Stage** ✅ **[COMPLETADO]**:
   - Creado [`Dockerfile`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/Dockerfile) multi-stage ultra-ligero (< 25 MB) sobre Alpine Linux listo para despliegues locales y VPS.

---

## 3. Guía de Verificación Continua

Para validar cada avance del roadmap, los desarrolladores y agentes de IA deben ejecutar la siguiente secuencia de validación estandarizada:

```powershell
# 1. Ejecutar tests unitarios y de regresión
go test -v ./...

# 2. Correr suite de micro-benchmarks y comparar alocaciones
go test -bench "." -benchmem ./core/...

# 3. Validar compilación cruzada Linux
.\scripts\build_linux.ps1

# 4. Actualizar grafo semántico en Kùzu
python .\tools\indexer\index_project.py

# 5. Probar endpoints del nodo activo y servidor MCP
python -c "import urllib.request, json; print(urllib.request.urlopen('http://localhost:8080/api/info').read().decode())"
```
