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
5. **Pipeline CI/CD con GitHub Actions**:
   - Configurar `.github/workflows/ci.yml` con compilación multiplataforma y ejecución de tests continuos.

---

### 🟡 Fase 2: Criptografía Avanzada & Rendimiento Núcleo (Semanas 3 - 5)

**Objetivo**: Fortalecer la privacidad a largo plazo y maximizar el throughput de datos.

1. **Multiplexación QUIC Nativa en Túneles P2P**:
   - Refactorizar [`core/tunnel.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/tunnel.go) para canalizar sockets TCP locales directamente a través de `quic.Stream` sobre el adaptador QUIC existente ([`adapters/quic_adapter.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/quic_adapter.go)).
   - Eliminar el empaquetado manual de fragmentos TCP sobre UDP, logrando rendimiento Gigabit nativo.
2. **Implementación de Handshake Noise_XX con Perfect Forward Secrecy (PFS)**:
   - Crear [`core/noise.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/noise.go) utilizando la especificación Noise Protocol Framework.
   - Generación de claves de cifrado efímeras por sesión; la clave estática Ed25519 solo se utiliza para firmar y autenticar el handshake inicial.
3. **Optimización con `sync.Pool` en E2EE**:
   - Reutilización de buffers de cifrado ChaCha20-Poly1305 en [`core/e2ee.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/e2ee.go) para reducir las 53 alocaciones por operación medidas en los benchmarks a < 5.

---

### 🟠 Fase 3: Experiencia Multimedia & Visualización Avanzada (Meses 2 - 3)

**Objetivo**: Transformar la experiencia visual del usuario y la eficiencia de transmisión interactiva.

1. **Visualizador de Malla 3D con WebGL / Force-Directed Graph**:
   - Reemplazar el canvas estático en el Dashboard Web por una visualización tridimensional interactiva que refleje con fidelidad los anillos del Mundo Pequeño y las latencias RTT en tiempo real.
2. **Escritorio Remoto Acelerado por Hardware**:
   - Implementar codificación diferencial de cuadros y enlace WebRTC MediaStream con compresión H.264 en [`core/remotedesktop_windows.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/remotedesktop_windows.go) para reducir el consumo de ancho de banda de 15 Mbps a < 2 Mbps.
3. **Transferencia de Archivos Robusta con Verificación Chunked**:
   - Protocolo de chunks de 64 KB con hash Blake3/SHA-256 por bloque y reanudación automática de descargas pausadas.

---

### 🔵 Fase 4: Ecosistema Agéntico Global & Autonomía de Red (Meses 3 - 4)

**Objetivo**: Autonomía completa para agentes de IA y despliegue masivo en producción.

1. **Canal de Streaming de Eventos SSE (`/api/events`)**:
   - Publicación de eventos en vivo hacia agentes de IA conectados vía HTTP para diagnóstico proactivo sin necesidad de polling.
2. **Agente Autónomo de Autocuración (Self-Healing Mesh Agent)**:
   - Rutina residente que analiza la tabla de enrutamiento cada 60 segundos; si detecta nodos aislados o latencias anómalas, consulta STUN/DHT y reconfigura los enlaces automáticamente.
3. **Contenedorización Docker Multi-Stage & Nodos Bootstrap Globales**:
   - `Dockerfile` ligero (< 20 MB) para desplegar nodos bootstrap de IPv7 en cualquier proveedor cloud en cuestión de segundos.

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
