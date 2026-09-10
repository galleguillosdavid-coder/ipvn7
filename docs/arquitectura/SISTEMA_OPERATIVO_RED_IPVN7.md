# Especificación de Ingeniería de Sistemas: Sistema Operativo de Red ipvn7

> **Documento Canónico de Arquitectura del Sistema Operativo de Red (NOS)**  
> **Estado**: Auditado y Consolidado · Versión: `ipvn7-living-nos-v1.0`  
> **Autoridad de Diseño**: Ingeniería de Sistemas de Red Descentralizada · Ecosistema Antigravity

---

## 1. Definición y Paradigma del Sistema Operativo de Red

**ipvn7** no es una simple aplicación cliente-servidor ni una VPN tradicional; es un **Sistema Operativo de Red Overlay (Network Operating System - NOS)** diseñado para operar de forma totalmente soberana, descentralizada y resiliente sobre cualquier infraestructura física existente o inexistente (Internet público, intranets corporativas, redes de radio off-grid LoRa/BLE, y conmutación ad-hoc).

### Axiomas Fundamentales
1. **La Identidad es Soberana y Criptográfica**: Ninguna autoridad central, registro DNS, dirección IP o proveedor de servicios asigna la identidad de un nodo. El par de claves **Ed25519** define de forma inmutable al nodo (`did:ipv7:<pubkey>`).
2. **Desacoplamiento Estricto Identidad $\leftrightarrow$ Ubicación**: Una dirección IP o puerto UDP es únicamente un localizador de transporte efímero y mutable. Un nodo puede conmutar de Wi-Fi, cambiar de proveedor celular o apagar sus interfaces de red sin perder sus sesiones lógicas, túneles ni estado criptográfico.
3. **El Núcleo Algorítmico es Sagrado (Core Freeze)**: Toda la lógica fundacional de identidad, wire format, sesiones y criptografía se mantiene congelada e inmutable (`core/`). Las nuevas capacidades crecen hacia afuera como **Adaptadores** (`adapters/`) o **Servicios de Observabilidad e IA** (`telemetry/`, `tools/`).
4. **La Observabilidad Jamás Afecta al Fast-Path**: La telemetría se extrae en paralelo mediante estructuras de datos *lock-free* acotadas. Si la telemetría satura su capacidad, se descarta a sí misma en $O(1)$ sin introducir jamás contención en los sockets de paquetes.

---

## 2. Modelo de Capas y Arquitectura del Sistema (L0 a L4)

```text
┌────────────────────────────────────────────────────────────────────────────────────────┐
│  L4: CAPA DE APLICACIONES, UX Y GOBERNANZA                                             │
│      · Web Dashboard SPA (8080)        · Chat Seguro E2EE CLI (cmd/chat)               │
│      · Monitor de Radar de Malla       · Gobernanza de Living Lab y Auditoría Externa  │
├────────────────────────────────────────────────────────────────────────────────────────┤
│  L3: CAPA DE INGENIERÍA ASISTIDA, GRAFOS E INTELIGENCIA ARTIFICIAL                     │
│      · Servidor Nativo MCP (core/mcp.go) sobre JSON-RPC stdio                          │
│      · DeepSeek Worker (Visión Multimodal + Deducción Algorítmica con 0 gasto tokens)  │
│      · Consultas Estructurales openCypher sobre Base de Datos Kùzu (.kuzu_index/)      │
├────────────────────────────────────────────────────────────────────────────────────────┤
│  L2: CAPA DE OBSERVABILIDAD DESACOPLADA Y TELEMETRÍA LOCK-FREE                         │
│      · Ring Buffer Lock-Free (<28 ns overhead, 0 alocaciones de heap)                  │
│      · Exportador OpenMetrics / Prometheus (/metrics, /health con regla anti-cardinal) │
│      · Motor de Detección de Anomalías (RTT spikes, degradación PMTU, contención SO)  │
│      · Exportador Asíncrono de Topología de Peers a Grafo Kùzu                         │
├────────────────────────────────────────────────────────────────────────────────────────┤
│  L1: CAPA DE ADAPTADORES, SUBSTRATO DE KERNEL Y ENRUTAMIENTO DISTRIBUIDO               │
│      · Driver de Red TUN/TAP OS (Prefijos virtuales fd07::/64 y 10.7.0.0/16)           │
│      · Enrutamiento de Cebolla Sphinx (3 saltos, paquetes fijos de 1280B, ECDH cascada)│
│      · Conmutador Híbrido Off-Grid (LoRa, BLE, Wi-Fi Direct, ad-hoc con fragmentador)  │
│      · Proxy SOCKS5 Embebido y Tunelización L3 P2P Directa                             │
│      · Streaming P2P de Escritorio Remoto de Baja Latencia                             │
│      · Tabla de Mundo Pequeño (12 Anillos logarítmicos × 10 peers = 120 peers max)     │
│      · Streaming en Cascada con Árbol de Fan-out 10                                    │
│      · DHT Kademlia Pura de 256 bits (Anti-Sybil PoW, firmas criptográficas)           │
│      · Transportes: UDP Fast-Path, QUIC (TLS 1.3), STUN RFC 5389, UPnP IGD, Relay DERP │
├────────────────────────────────────────────────────────────────────────────────────────┤
│  L0: PROTOCOL CORE (STRICT FREEZE — INMUTABLE)                                         │
│      · Identidad Descentralizada Soberana (DID Ed25519) + Almacén Persistente Seguro    │
│      · Contenedor de Red Canónico Determinista CBOR (RFC 8949) con Firma Digital       │
│      · Handshake Autenticado Mutuamente Noise XX con Forward Secrecy                   │
│      · Cifrado E2EE Autenticado ChaCha20-Poly1305 con Derivación X25519 (ECDH)         │
│      · Módulo Híbrido Criptográfico Post-Cuántica (PQC Kyber / ML-KEM)                 │
│      · Filtro Anti-Replay con Ventana Deslizante y Bitmask                             │
└────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Desglose Exhaustivo de Virtudes Tecnológicas

### 3.1 Identidad Soberana (DID) y Keystore Seguro
- **Identificador DID**: `did:ipv7:<hex_public_key_32bytes>`.
- **Independencia Total**: Generado localmente a partir de entropía criptográfica del sistema operativo (`crypto/rand`). Jamás depende de asignaciones IP ni de registros centrales.
- **Persistencia Restringida**: Almacenado en `%USERPROFILE%\.ipv7\identity.key` (o `~/.ipv7/identity.key` en Linux) con permisos estrictos de sistema (0600 en UNIX).
- **Invariante de Movilidad (IP Roaming)**: Demostrado con **100% de éxito y 0% de pérdida de paquetes** en conmutaciones dinámicas de interfaz y red Wi-Fi $\leftrightarrow$ Celular.

### 3.2 La Ciudadela Criptográfica
- **Determinismo CBOR (RFC 8949)**: Mediante `fxamacker/cbor/v2`, la serialización canónica asegura que los campos de los contenedores se ordenen de manera idéntica bit a bit antes de calcular la firma Ed25519.
- **Handshake Noise XX**: Intercambio de claves en 3 vías que garantiza autenticación mutua, secreto perfecto hacia adelante (*PFS*) y ocultación de identidades ante observadores pasivos.
- **Cifrado E2EE Autenticado**: ChaCha20-Poly1305 con claves efímeras derivadas mediante Diffie-Hellman sobre curva Curve25519 (X25519). Los nodos de retransmisión (*relays*) solo ven contenedores opacos cifrados.
- **Resistencia Post-Cuántica (PQC)**: `core/pqc.go` introduce encapsulación de clave basada en retículos (**Kyber / ML-KEM**), protegiendo las sesiones contra la recolección hostil actual con vistas al descifrado futuro mediante computación cuántica ("harvest now, decrypt later").
- **Protección Anti-Replay**: `adapters/replay_filter.go` descarta instantáneamente cualquier paquete duplicado o fuera de la ventana temporal de frescura.
- **Enrutamiento Sphinx Onion**: Paquetes de longitud fija de 1280 bytes que atraviesan 3 nodos de la malla con cifrado concéntrico ECDH, impidiendo la correlación de tráfico por tamaño o metadatos de capa física.

### 3.3 Substrato de Red a Nivel de Sistema Operativo (TUN/TAP)
- **Controlador Virtual OS**: El adaptador `adapters/tun/` crea una interfaz de red virtual en el Kernel del host.
- **Espacio de Direccionamiento**:
  - IPv6: Prefijo ULA asignado `fd07::/64`.
  - IPv4: Subred virtual asignada `10.7.0.0/16`.
- **Transparencia Total**: Cualquier binario del sistema (curl, navegadores, bases de datos, SSH) puede comunicarse a través del Sistema Operativo de Red `ipvn7` simplemente apuntando a las direcciones virtuales del overlay.

### 3.4 Enrutamiento de Mundo Pequeño y Streaming en Cascada
- **Mundo Pequeño Acotado**: La tabla de enrutamiento implementa el modelo de Kleinberg dividida en **12 anillos logarítmicos concéntricos con un máximo estricto de 10 peers por anillo**.
- **Cota en Memoria**: Máximo absoluto de **120 peers en memoria**, permitiendo que nodos en hardware modesto o microcontroladores mantengan una topología acotada pero con capacidad de alcanzar hasta **$10^{12}$ nodos** mediante reenvío voraz XOR en no más de 12 saltos (`HopLimit = 12`).
- **Cascade Streaming**: Distribución en árbol con **fan-out 10** que permite la entrega masiva de streaming de audio, video o archivos pesados con retransmisión aguas abajo sin sobrecargar el ancho de banda del emisor raíz.

### 3.5 Servicios Integrados del Sistema Operativo de Red
- **Proxy SOCKS5 Nativo** (`core/socks5.go`): Expone un puerto SOCKS5 local para enrutar tráfico TCP estándar a través del overlay cifrado.
- **Tunelización L3 P2P** (`core/tunnel.go`): Enlace túnel punto a punto para interconectar subredes corporativas o privadas.
- **Escritorio Remoto P2P** (`core/remotedesktop.go`, `core/remotedesktop_windows.go`): Captura y codificación de pantalla en tiempo real transmitida por el canal de baja latencia de IPv7 con control de teclado/ratón seguro.

### 3.6 Observabilidad Desacoplada y Grafo de Red en Kùzu
- **Ring Buffer Lock-Free** (`telemetry/ring_buffer.go`): Búfer circular sin contención de locks. Overhead medido en benchmark: **27.18 ns por evento con 0 alocaciones de heap**.
- **Regla Anti-Cardinalidad en Prometheus**: Monitoreo pasivo (`/metrics`) con contadores atómicos y métricas de salud que protegen la memoria contra explosión de series temporales.
- **Base de Datos de Grafo Kùzu**: Persistencia de la topología P2P y del AST del proyecto (`.kuzu_index/ipv7.db`) accesible mediante lenguaje de consulta estándar **openCypher**.
- **Motor de Detección de Anomalías** (`telemetry/anomaly_engine.go`): Detección automática en caliente de RTT anómalo, degradación de PMTU y saturación de búferes de recepción del socket del sistema operativo (aislamiento del cuello de botella ADV-01).

---

## 4. Orquestación del Ecosistema y Herramientas

### 4.1 Entorno Dual: Windows Host + Ubuntu en WSL2
El sistema está diseñado para aprovechar la sinergia de ambos sistemas operativos:
- **Windows Host**: Ejecuta la interfaz gráfica, el dashboard web interactivo (`http://localhost:8080`) y el IDE Antigravity.
- **Ubuntu WSL2**: Proporciona el entorno de sockets Linux de alto rendimiento, simulación de namespaces de red y pruebas de resistencia cruzada.

**Flujo Operativo de Scripts**:
1. Compilar Windows: `.\scripts\build_windows.ps1` $\rightarrow$ genera `bin/ipv7-node.exe` y `bin/chat.exe`.
2. Lanzar Nodo Windows: `.\scripts\run_node.bat` $\rightarrow$ auto-compila si es necesario y arranca con `-key persistent -upnp=true`.
3. Compilar Linux: `.\scripts\build_linux.ps1` $\rightarrow$ genera `bin/ipv7-node-linux` y `bin/chat-linux`.
4. Lanzar Nodo en WSL2: `.\scripts\run_wsl.ps1 -Port 7002 -UIPort 8082`.
5. Validación de Malla Dual: `.\scripts\dual_node_test.ps1` $\rightarrow$ conecta automáticamente el nodo de Windows con el nodo de WSL2.

### 4.2 Inteligencia de Grafo: Kùzu CLI y Consultas Cypher
Kùzu almacena el mapa topológico del protocolo y el índice del código:
- **Binarios**: `tools/kuzu/kuzu.exe` (Windows) y `tools/kuzu/kuzu` (Linux/WSL2).
- **Indexación**: `python tools/indexer/index_project.py`
- **Auditoría**: `python tools/indexer/audit_kuzu.py`

**Ejemplo de Consultas Cypher Útiles**:
```cypher
-- Explorar todos los structs y funciones del núcleo de red
MATCH (p:Package {name: 'core'})-[:CONTAINS]->(f:File)-[:DEFINES]->(s:Symbol)
RETURN f.name, s.name, s.kind;

-- Auditar acoplamiento entre adaptadores y el núcleo congelado
MATCH (p1:Package)-[:DEPENDS_ON]->(p2:Package)
WHERE p2.name = 'core'
RETURN p1.name, p2.name;

-- Inspeccionar peers activos registrados en la topología
MATCH (p:Peer)-[r:CONNECTED_TO]->(m:Peer)
RETURN p.id, r.adapter, r.latency_ms, m.id;
```

### 4.3 Delegación a la API de DeepSeek (Ahorro de Tokens y Visión Multimodal)
Para mantener la eficiencia de la sesión en Antigravity y evitar la saturación de la ventana de contexto:
- **Herramienta**: `tools/deepseek/deepseek_worker.py` (con wrappers `scripts/ask_deepseek.ps1` y `scripts/ask_deepseek.sh`).
- **Modelos**:
  - `deepseek-flash` (por defecto): Inferencia multimodal de alta velocidad. Procesa capturas de pantalla de la UI, diagramas y revisión masiva de archivos de código.
  - `deepseek-v4-pro`: Razonamiento profundo algorítmico (*chain-of-thought*) para análisis matemático de cotas y pruebas de concurrencia.

### 4.4 Model Context Protocol (MCP) Nativo
El flag `-mcp` transforma el ejecutable del nodo en un servidor nativo **MCP** sobre `stdio`. Cualquier agente de IA compatible con MCP puede:
- Invocar herramientas (`discover_endpoints`, `query_topology`, `create_tunnel`, `read_telemetry`).
- Supervisar el estado de la red sin necesidad de scraping de pantalla o APIs REST ad-hoc.

---

## 5. Política de Higiene y Mantenimiento del Repositorio

1. **Raíz Inmaculada**:
   - La raíz contiene exclusivamente metadatos del proyecto (`README.md`, `LICENSE.md`, `TERMS_OF_USE.md`, `CONTRIBUTING.md`, `Dockerfile`, `go.mod`, `go.sum`, `.gitignore`).
   - Todos los binarios se dirigen a `bin/` o `release/`.
   - Todos los scripts de lanzamiento y mantenimiento residen en `scripts/`.
2. **Cero Regresiones en Benchmarks**:
   - Ninguna modificación de adaptadores puede degradar la latencia base en más de un 5% ni introducir alocaciones de memoria en el bucle crítico de paquetes.
3. **Reproducibilidad Absoluta**:
   - Todo resultado publicado en `docs/` debe ser reproducible mediante los scripts en `scripts/audit_reproducibility.ps1` y `scripts/audit_reproducibility.sh`.
