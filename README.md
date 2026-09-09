# IPv7: Protocolo de Red P2P Descentralizada Overlay

Implementación en **Go** del protocolo IPv7 basado en identidades criptográficas, transporte desacoplado y topología de streaming en cascada.

## 🚀 Arquitectura y Componentes Construidos

1. **Criptografía e Identidad Base (`core/`)**:
   - Identidades lógicas fijas derivadas de pares de claves **Ed25519**.
   - Cada contenedor de datos está serializado de forma canónica determinista (**CBOR**) y firmado digitalmente.
   - Entidad **`Node`** que orquesta llaves, adaptadores de transporte y enrutamiento P2P.

2. **Transporte UDP Best-Effort (`adapters/udp_adapter.go`)**:
   - Transporte rápido no orientado a conexión para mensajes ligeros dentro del MTU seguro.

3. **NAT Traversal Temprano con STUN (`adapters/stun.go`)**:
   - Integración con `github.com/pion/stun/v2`.
   - Detección automática de IPs locales y la **IP pública reflexiva** consultando servidores STUN públicos (ej. Google STUN).

4. **Transporte Fiable y Chunks con QUIC (`adapters/quic_adapter.go`)**:
   - Integración con `github.com/quic-go/quic-go` sobre TLS 1.3 efímero.
   - Permite enviar **archivos y payloads masivos (probado con 5MB+)** sin preocuparse por la fragmentación de paquetes manual.

5. **Streaming en Cascada (`core/cascade.go`)**:
   - Topología de distribución en árbol (*Tree-based Multicast*).
   - Cada nodo tiene un límite de abanico (por defecto **10 dispositivos hijos**).
   - Los datos emitidos por el nodo raíz se retransmiten automáticamente aguas abajo en cascada, permitiendo escalar a miles de receptores sin sobrecargar al emisor original.

6. **Enrutamiento "Mundo Pequeño" - 12 Grados de Separación (`core/smallworld.go`)**:
   - Tablas de enrutamiento **estrictamente acotadas** (máximo 120 peers en memoria: 12 anillos de distancia logarítmica $\times$ 10 peers).
   - Cobertura teórica global de hasta **$10^{12} = 1.000.000.000.000$ dispositivos** (1 billón).
   - Reenvío voraz (*Greedy Routing*) por distancia criptográfica XOR.
   - Campo `HopLimit = 12` con preservación de firma digital de extremo a extremo.
   - Consulta la [Documentación del Modelo de Mundo Pequeño](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/arquitectura/02_mundo_pequeno_routing.md).

7. **Descubrimiento con DHT Seguro (`dht/`)**:
   - Mapeo descentralizado de `Identity (Clave Pública) -> []Endpoints (IP:Puerto)`.
   - **Mitigación de ataques Sybil y envenenamiento**: Cada registro en la DHT (`Record`) está firmado con la clave privada de la identidad anunciada. Los nodos intermedios rechazan cualquier intento de falsificación.

8. **Fallback Relay tipo DERP (`adapters/relay_adapter.go`)**:
   - Servidor y cliente de retransmisión para escenarios donde ambos nodos están tras NAT simétrico estricto o CGNAT móvil y el NAT traversal directo falla.
   - El servidor Relay es ciego al contenido: los paquetes viajan firmados y sellados de extremo a extremo sin que el relay pueda espiar ni adulterar los datos.

9. **Cifrado de Extremo a Extremo (E2EE) (`core/e2ee.go`)**:
   - Claves derivadas X25519 a partir de la semilla Ed25519 con Diffie-Hellman (ECDH).
   - Cifrado autenticado de grado militar **ChaCha20-Poly1305** con clave efímera (modelo HPKE / RFC 9180).
   - Ni los nodos intermediarios de la red ni los servidores Relay pueden espiar el contenido de los mensajes.

10. **Adaptador WebRTC Nativo (`adapters/webrtc_adapter.go`)**:
    - Integración completa con `pion/webrtc/v4`.
    - Comunicación P2P nativa mediante `RTCDataChannel` con intercambio de ofertas y respuestas SDP/ICE.

11. **Interfaz Gráfica / Dashboard Web (`ui/`)**:
    - Panel visual interactivo embebido en el nodo accesible en `http://localhost:8080`.
    - Chat visual seguro con botón de cifrado E2EE y **transferencia de archivos Drag & Drop**.
    - **Topología de Red en Malla (Mesh Graph)**: Visualizador interactivo de grafo con física de partículas, mostrando nodos, adapters de transporte, latencias y exportación directa a Kùzu Cypher.
    - **Radar del Mundo Pequeño**: Visualizador de los 12 anillos de separación logarítmica en tiempo real.
    - Monitor del árbol de streaming en cascada con botón de emisión de frames.
    - Acceso directo a especificación viva **OpenAPI 3.1** (`/api/openapi.json`) y métricas **Prometheus** (`/metrics`).
    - Binario unificado: **`ipv7-node.exe`** (Windows) y **`bin/ipv7-node-linux`** (Linux/WSL2).

12. **Servidor Model Context Protocol (MCP) Nativo (`core/mcp.go`)**:
    - Servidor JSON-RPC 2.0 integrado compatible con la especificación de Anthropic.
    - Permite a Claude Desktop, Antigravity o Cursor controlar el nodo directamente mediante `stdio` (flag `-mcp`) o vía HTTP POST en `/api/mcp`.

13. **Keystore Persistente Ed25519 (`core/keystore.go`)**:
    - Almacenamiento seguro y recarga de identidades fijas en `~/.ipv7/identity.key` mediante flag `-key persistent`.

14. **Apertura Dinámica de Puertos UPnP IGD (`adapters/upnp.go`)**:
    - Descubrimiento SSDP multicast y mapeo automático del puerto UDP en routers residenciales para reducir latencia y evitar el paso por relays.

---

## 📚 Documentación Integral y Recursos

Toda la documentación técnica, operativa y de arquitectura está centralizada en [`docs/`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/README.md):

- [**Centro de Documentación Unificado (`docs/README.md`)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/README.md): Hub principal con acceso directo a todos los módulos.
- [**Arquitectura y Protocolo (`docs/arquitectura/`)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/arquitectura/01_genesis.md): Plan Génesis, enrutamiento en Mundo Pequeño (12 anillos) y modelado de grafos Kùzu.
- [**Manual Operativo para Humanos (`docs/operaciones/`)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/README.md): Guías de usuario, instalación, supervisión en WSL2, acceso remoto y casos prácticos.
- [**Auditorías y Reportes de Pruebas (`docs/auditorias_y_reportes/`)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/auditorias_y_reportes/README.md): Auditorías unitarias, benchmarks, pruebas de malla en vivo y auditoría de grafo Kùzu.
- [**Sugerencias de Arquitectura, UX y Código (`docs/sugerencias_mejora/`)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/sugerencias_mejora/README.md): Roadmap técnico, estado evolutivo y prototipos de código.
- [**Skill Especializada para Agentes de IA (`docs/skill_ia/`)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/skill_ia/SKILL.md): System prompt, OpenAPI, schemas de Function Calling y consultas Cypher para Kùzu.
- [**Índice de Código (`docs/indices/`)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/indices/PROJECT_INDEX.md): Catálogo exhaustivo de todos los paquetes, archivos y funciones Go.

---

## 🧪 Pruebas Automatizadas

Para ejecutar todas las pruebas unitarias y de integración del proyecto:

```powershell
go test -v ./...
```

---

## 🐧 Ejecución y Supervisión en WSL2 (Linux)

IPv7 puede compilarse y ejecutarse en **WSL2 (Ubuntu)** con supervisión en tiempo real desde Windows:

1. **Compilar binarios Linux:**
   ```powershell
   .\scripts\build_linux.ps1
   ```

2. **Iniciar nodo en WSL2 con logs automáticos en `logs/wsl_node.log`:**
   ```powershell
   .\scripts\run_wsl.ps1 -Port 7002 -UIPort 8082
   ```
   Abre tu navegador en Windows en **`http://localhost:8082`**.

3. **Guía operativa detallada:** Consulta [docs/operaciones/06_supervision_wsl2.md](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/06_supervision_wsl2.md).

---

## 🔀 Prueba de Malla Cruzada (Windows ↔ WSL2 Linux)

Para validar la red P2P real entre el host Windows y el entorno Linux:

```powershell
.\scripts\dual_node_test.ps1
```
- **Nodo A (Windows):** `http://localhost:8080`
- **Nodo B (WSL2 Linux):** `http://localhost:8082` (conectado automáticamente a Nodo A)

¡Observa ambos nodos en la pestaña **🕸️ Topología de Malla** reconociéndose, intercambiando mensajes E2EE y frames en cascada!

---

## 📊 Base de Datos de Grafo Kùzu (Kazu) e Interfaz Visual

El proyecto integra **Kùzu Graph DB** (`tools/kuzu/`) para indexar la arquitectura del código y la topología de la red:

- **🖥️ Interfaz Gráfica Visual (Kùzu Studio):**
  - **Embebido en el Nodo:** Abre cualquier nodo (`http://localhost:8080` o `http://localhost:8082` en WSL2) y ve a la pestaña **📊 Explorador Kùzu (Red & Código)** para alternar entre la vista de red P2P y la arquitectura del código, con consola Cypher interactiva.
  - **Servidor Standalone (sin requerir nodo P2P):**
    ```powershell
    .\tools\kuzu_explorer\run_explorer.ps1
    ```
    Disponible en `http://localhost:8090`.

- **Reindexar el proyecto:**
  ```powershell
  python .\tools\indexer\index_project.py
  ```
- **Consultar el grafo con Cypher CLI:**
  ```powershell
  .\tools\kuzu\kuzu.exe .kuzu_index\ipv7.db
  ```
- **Índice completo del repositorio:** Consulta [docs/indices/PROJECT_INDEX.md](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/indices/PROJECT_INDEX.md).
- **Manual de consultas Cypher:** Consulta [docs/arquitectura/03_kuzu_mesh_graph.md](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/arquitectura/03_kuzu_mesh_graph.md).

---

## 💬 Cómo Probar el Chat P2P (Terminal)

Puedes abrir dos terminales en tu computadora:

### Terminal 1 (Nodo A):
```powershell
.\chat.exe -port 7001
```

### Terminal 2 (Nodo B):
```powershell
.\chat.exe -port 7002 -peer 127.0.0.1:7001
```

---

## 🖥️ Cómo Lanzar el Nodo con Interfaz Gráfica (Web Dashboard)

Para una experiencia visual completa con cifrado E2EE, visualizador de malla, radar de 12 grados y streaming en cascada:

```powershell
.\ipv7-node.exe -port 7001 -ui 8080
```

Luego abre tu navegador en **`http://localhost:8080`**.

---

## 🛡️ Auditoría de Código con Kùzu y Protocolo Criptográfico Real

Se eliminaron completamente todas las claves públicas dummy y valores cableados mediante el motor de grafo **Kùzu**:

- **Herramienta de Auditoría de Grafo:**
  ```powershell
  python .\tools\auditor.py
  ```
  Analiza archivos, dependencias y llamadas estructurales indexadas en Kùzu para garantizar cero buffers en blanco, stubs o claves dummy en producción.

- **Handshake Autenticado Ed25519 (`core/handshake.go`):**
  Al conectar con un peer (`-peer ip:port`), el nodo no asume identidades en cero ni claves estáticas. Envía un `ControlHandshakeReq` firmado con un nonce criptográfico y timestamp de frescura. El peer responde con `ControlHandshakeResp` firmado con su propia clave Ed25519. Al verificarse mutuamente, se registra el peer con su identidad real y su latencia RTT medida dinámicamente en la tabla de Kleinberg y en la DHT.

- **Telemetría y Streaming Verificable:**
  Los frames en cascada utilizan codificación CBOR estructurada (`StreamTelemetryFrame`) con sumas de comprobación SHA-256 e identidades criptográficas de origen en lugar de cadenas de texto estáticas.

---

## ⚖️ Términos y Condiciones de Uso (Licencia de Auditoría)

Este proyecto está protegido bajo la **Licencia de Auditoría e Inspección (Source-Available)**:
- 📖 **Permitido**: Leer, auditar, inspeccionar la criptografía y ejecutar pruebas locales de validación de seguridad.
- 🚫 **Prohibido**: Copiar, redistribuir públicamente, crear forks comerciales o explotar el código con fines comerciales sin autorización explícita por escrito del autor.

Para más detalles legales, consulte el archivo **[LICENSE.md](LICENSE.md)** o **[TERMS_OF_USE.md](TERMS_OF_USE.md)**.  
**Copyright (c) 2026 David Galleguillos (@galleguillosdavid-coder). Todos los derechos reservados.**

