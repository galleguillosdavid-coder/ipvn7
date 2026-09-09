# Reporte 04: Catálogo Exhaustivo de Oportunidades de Mejora

Este documento recopila y analiza en profundidad las oportunidades de mejora identificadas durante la auditoría técnica y ejecución de pruebas sobre el ecosistema **IPv7**, clasificadas en 5 dimensiones estratégicas.

---

## 1. Arquitectura de Red y Rendimiento

### 1.1. Multiplexación QUIC Nativa en Túneles P2P
- **Diagnóstico Actual**: La funcionalidad de reenvío de puertos TCP ([`core/tunnel.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/tunnel.go)) encapsula los bytes leídos de conexiones locales en paquetes discretos CBOR (`Container`) transmitidos sobre UDP o llamadas `node.SendMessage()`.
- **Problema**: En conexiones TCP con alto flujo (RDP, transferencias masivas de archivos o HTTP/2), la pérdida aleatoria de paquetes UDP causa retransmisiones redundantes y Head-of-Line Blocking a nivel de aplicación.
- **Solución Propuesta**:
  - Utilizar flujos bidireccionales nativos de QUIC (`quic.Stream`) ya provistos por el adaptador QUIC ([`adapters/quic_adapter.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/quic_adapter.go)).
  - Cada nuevo socket TCP local mapeado abre un `quic.Stream` directo cifrado con control de congestión BBR integrado y multiplexación zero-copy.

### 1.2. Optimización de Memoria Zero-Copy y `sync.Pool`
- **Diagnóstico Actual**: En los micro-benchmarks, `BenchmarkE2EEEncryptDecrypt` generó **53 alocaciones** por operación (3,632 bytes), mientras que la serialización CBOR generó 7 alocaciones (2,587 bytes).
- **Solución Propuesta**:
  - Implementar un pool de buffers reutilizables con `sync.Pool` para payloads de encriptación (`nonce`, `ciphertext` y buffers de compresión CBOR).
  - Reducir las alocaciones por operación a menos de 5, disminuyendo la latencia en un ~35% y eliminando pausas de Garbage Collector en nodos de alta concurrencia.

### 1.3. Aceleración H.264 / WebRTC para el Escritorio Remoto
- **Diagnóstico Actual**: El módulo de captura ([`core/remotedesktop_windows.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/remotedesktop_windows.go)) comprime pantallas completas a 1080p como imágenes JPEG enviadas secuencialmente vía WebSocket binario.
- **Problema**: A 15-30 FPS, esto consume entre 5 y 15 Mbps de ancho de banda y genera alta utilización de CPU por compresión JPEG software.
- **Solución Propuesta**:
  - Integrar codificación diferencial por bloques (enviar solo los rectángulos de la pantalla que han cambiado).
  - Alternativamente, enlazar con la API de Desktop Duplication de DirectX 11 (DXGI) y codificador por hardware H.264/NVENC/QuickSync para transmitir directamente sobre `adapters/webrtc_adapter.go` a < 2 Mbps y 60 FPS fluidos.

---

## 2. Seguridad Criptográfica Avanzada

### 2.1. Adopción del Framework Noise (Patrón `Noise_XX`) con PFS
- **Diagnóstico Actual**: El cifrado E2EE actual deriva la clave simétrica ChaCha20-Poly1305 mediante un intercambio Diffie-Hellman estático X25519 derivado de las claves de identidad.
- **Riesgo**: Si la clave privada de identidad de un nodo se viera comprometida en el futuro, un adversario que haya grabado tráfico histórico cifrado podría desencriptar pasivamente todas las conversaciones pasadas (falta de Perfect Forward Secrecy).
- **Solución Propuesta**:
  - Implementar el protocolo de handshake `Noise_XX` (utilizado por WireGuard y Signal).
  - En cada sesión P2P, ambos extremos generan pares de claves efímeras $e_{local}, e_{remote}$ que se combinan con las claves estáticas $s_{local}, s_{remote}$. Al terminar la sesión o tras $2^{16}$ mensajes, las claves de sesión se destruyen irremediablemente.

### 2.2. Ventana Deslizante Anti-Replay para UDP ✅ [IMPLEMENTADO]
- **Estado**: ✅ **COMPLETADO Y AUDITADO**.
- **Implementación**:
  - Estructura `ReplayWindow` y `AntiReplayTable` con ventana deslizante de 64 bits (RFC 6479) implementada en [`adapters/replay_filter.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/replay_filter.go).
  - Integrada en [`adapters/udp_adapter.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/udp_adapter.go) tras la verificación de firma Ed25519.
  - Paquetes duplicados o fuera de ventana son descartados en $O(1)$ sin alocaciones adicionales.
  - Verificado en [`adapters/replay_test.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/replay_test.go) con 100% de éxito.

### 2.3. Rate Limiting Basado en Token Bucket por Identidad
- **Diagnóstico Actual**: No existe limitación de tasa por peer en la capa de transporte UDP. Un nodo infectado o malicioso podría inundar al nodo local con peticiones de ping o payloads pesados.
- **Solución Propuesta**:
  - Incorporar un limitador de tasa Token Bucket (`golang.org/x/time/rate`) asignado dinámicamente por hash de clave Ed25519 con una cuota máxima (ej. 100 paquetes/seg por peer normal, con burst de 200).

---

## 3. Experiencia de Usuario (UX) e Interfaz Web

### 3.1. Reconexión Automática con Backoff Exponencial en Dashboard Web ✅ [IMPLEMENTADO]
- **Estado**: ✅ **COMPLETADO Y AUDITADO**.
- **Implementación**:
  - Función `connectWS()` refactorizada en [`ui/assets/index.html`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/ui/assets/index.html) con retroceso exponencial (`1s`, `2s`, `4s`, máx `10s`).
  - Badge visual dinámico en la cabecera (`wsStatusPill`, `wsStatusDot`, `wsStatusText`) que informa al usuario en tiempo real con cuenta regresiva.
  - Al reestablecer el socket, se invocan automáticamente `loadInfo()` y `loadPeers()`, restaurando el estado sin requerir recargar la página (`F5`).

### 3.2. Visualización 3D Interactiva de la Malla P2P con Force-Directed Graph
- **Diagnóstico Actual**: El visualizador de malla actual utiliza Canvas 2D con un radar estático.
- **Solución Propuesta**:
  - Incorporar visualizador 3D interactivo basado en WebGL / Three.js (o `force-graph`) donde los nodos se posicionen según su latencia real (RTT) y grado de anillo del Mundo Pequeño, permitiendo rotar, hacer zoom y pulsar sobre cualquier nodo para abrir un túnel o chat directo.

### 3.3. Transferencia de Archivos Chunked con Reanudación
- **Diagnóstico Actual**: La transferencia de archivos actual envía bloques de datos simples sobre WebSockets o mensajes P2P.
- **Solución Propuesta**:
  - Implementar protocolo chunked con cálculo de hash Blake3/SHA-256 por bloque de 64 KB, confirmaciones selectivas (SACK) y reanudación automática de descargas interrumpidas.

---

## 4. Inteligencia Artificial y Servidor MCP

### 4.1. Ampliación del Catálogo de Herramientas MCP ✅ [IMPLEMENTADO]
- **Estado**: ✅ **COMPLETADO Y AUDITADO**.
- **Implementación**:
  - Expandido en [`core/mcp.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/mcp.go) y conectado con [`ui/server.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/ui/server.go):
    1. `ipv7_start_tunnel` (alias `ipv7_create_tunnel`): Inicia un túnel TCP cifrado P2P hacia cualquier peer remoto.
    2. `ipv7_toggle_vpn` (alias `ipv7_start_vpn`): Inicia o detiene el proxy VPN SOCKS5 en espacio de usuario.
    3. `ipv7_query_kuzu`: Ejecuta consultas Cypher dinámicas sobre la base de datos de grafos Kùzu (`.kuzu_index/`).
  - Verificado en [`core/mcp_test.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/mcp_test.go).

### 4.2. Streaming de Eventos SSE (Server-Sent Events) para Agentes
- **Solución Propuesta**:
  - Exponer `/api/events` vía HTTP Server-Sent Events (SSE) para que agentes LLM reciban notificaciones en tiempo real cuando un nuevo peer se conecte, caiga un enlace o se detecte una anomalía de red sin necesidad de polling repetitivo.

---

## 5. DevOps, Calidad de Código y CI/CD

### 5.1. Pipeline Automatizado de GitHub Actions
- **Solución Propuesta**:
  - Crear `.github/workflows/ci.yml` con:
    - Validación de compilación en matriz triple (Windows, Ubuntu Linux, macOS).
    - Ejecución de suite de tests con detector de condiciones de carrera: `go test -race ./...`.
    - Linter de código estricto con `golangci-lint` (detección de variables no usadas, chequeo de errores no controlados y formateo canónico).

### 5.2. Empaquetado en Contenedor Docker Multi-Stage
- **Solución Propuesta**:
  - Añadir un `Dockerfile` optimizado que compile el binario con Go estático (`CGO_ENABLED=0`) y empaquete el runtime en una imagen mínima basada en Alpine o Scratch (< 25 MB), facilitando el despliegue de nodos bootstrap en nubes como AWS, GCP, Hetzner o Fly.io.
