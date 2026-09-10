# Plan Maestro y Checklist Integral: Sistema Operativo de Red ipvn7 (NOS)

> **Documento Rector y Fuente Única de Verdad (SSOT)**  
> **Versión**: `v2.0-NOS-EVOLUTION` · **Fecha**: Septiembre 2026  
> **Estado**: Planificación y Hoja de Ruta Activa · **Autoridad**: Ingeniería de Sistemas y Arquitectura

---

## 🏛️ PARTE I — ESTADO BASE CERTIFICADO (HORIZONTES 1 A 5)

Antes de proyectar las nuevas capacidades, el estado base del protocolo IPv7 se encuentra **100% auditado, probado y certificado** en la versión congelada `v0.5.0-chaos.audit`.

### Checklist de Capacidades Demostradas (`DEMONSTRATED`)

- [x] **Identidad Criptográfica Soberana (DID)**: Pares de claves Ed25519 (`did:ipv7:<pubkey>`) 100% desacopladas de IP, puertos y hardware.
- [x] **Almacén Seguro de Claves**: Keystore persistente en `%USERPROFILE%\.ipv7\identity.key` (o `~/.ipv7/identity.key`) con permisos restringidos 0600.
- [x] **Wire Format Canónico Determinista**: Serialización CBOR (RFC 8949) mediante `fxamacker/cbor/v2` con verificación estricta de firmas digitales.
- [x] **Handshake Autenticado Noise XX**: Intercambio de claves en 3 vías con secreto perfecto hacia adelante (*PFS*) y ocultación de identidades.
- [x] **Cifrado E2EE Autenticado**: ChaCha20-Poly1305 con claves simétricas efímeras derivadas mediante ECDH sobre X25519.
- [x] **Filtro Anti-Replay**: Ventana deslizante de 128 paquetes con Bloom filter para descarte en $O(1)$ de paquetes duplicados o retrasados.
- [x] **Criptografía Híbrida Post-Cuántica (PQC)**: Módulo Kyber / ML-KEM en `core/pqc.go` para encapsulación resistente a computación cuántica.
- [x] **Substrato de Red OS (TUN/TAP)**: Driver virtual en `adapters/tun/` con prefijos `fd07::/64` (IPv6) y `10.7.0.0/16` (IPv4) para transposición transparente de apps.
- [x] **Enrutamiento de Mundo Pequeño (12 Anillos)**: Tabla acotada a **120 peers en memoria** (12 anillos $\times$ 10 nodos), reenvío voraz métrico XOR y `HopLimit = 12`.
- [x] **Streaming en Cascada**: Topología en árbol con fan-out de 10 nodos para distribución masiva de flujos de datos.
- [x] **DHT Kademlia Pura de 256 bits**: Almacén distribuido descentralizado con Proof-of-Work (PoW) y firmas anti-envenenamiento en `dht/`.
- [x] **Auto-Descubrimiento en Red Local (LAN)**: Baliza de broadcast UDP en `core/beacon.go` para detección instantánea sin configuración en Wi-Fi.
- [x] **Movilidad y Roaming IP Instantáneo**: 25 ensayos automatizados con **100% de entrega, 0% de pérdida de paquetes y latencia P50 de 1.10 ms**.
- [x] **NAT Traversal & UPnP IGD**: Detección de IP pública vía STUN RFC 5389, mapeo automático en routers residenciales UPnP y fallback a Relay DERP ciego.
- [x] **Observabilidad Lock-Free de Ultra-Bajo Costo**: Ring Buffer desacoplado (<28 ns/op, 0 alocaciones de heap), Prometheus OpenMetrics y Anomaly Engine.
- [x] **Base de Datos de Grafo Kùzu**: 1.324 sentencias Cypher indexadas en `.kuzu_index/ipv7.db` modelando AST, paquetes y topología en malla.
- [x] **Servidor Nativo MCP**: Protocolo Model Context Protocol sobre stdio (`-mcp`) para control agéntico por asistentes de IA.
- [x] **Worker de IA DeepSeek**: Delegador de tareas cognitivas pesadas y visión multimodal (`deepseek_worker.py`) para ahorro de tokens.
- [x] **Orquestación Multi-Plataforma**: Ejecución cruzada validada entre Windows Host y Ubuntu en WSL2 (`scripts/dual_node_test.ps1`).

---

## 🧭 PARTE II — MANIFIESTO DEL SISTEMA OPERATIVO DE RED (NOS)

ipvn7 trasciende la noción de una simple "VPN en malla" para convertirse en un **Sistema Operativo de Red Autónomo y Programable**:

```text
┌────────────────────────────────────────────────────────────────────────────────────────┐
│  L4: APLICACIONES, UX Y GOBERNANZA                                                     │
│      · Web Dashboard SPA (8080)        · Chat Seguro E2EE CLI (cmd/chat)               │
│      · Monitor de Radar de Malla       · Gobernanza de Living Lab y Auditoría Externa  │
├────────────────────────────────────────────────────────────────────────────────────────┤
│  L3: INGENIERÍA ASISTIDA, GRAFOS E INTELIGENCIA ARTIFICIAL                             │
│      · Servidor Nativo MCP (core/mcp.go) sobre JSON-RPC stdio                          │
│      · DeepSeek Worker (Visión Multimodal + Deducción Algorítmica con 0 gasto tokens)  │
│      · Consultas Estructurales openCypher sobre Base de Datos Kùzu (.kuzu_index/)      │
├────────────────────────────────────────────────────────────────────────────────────────┤
│  L2: OBSERVABILIDAD DESACOPLADA Y TELEMETRÍA LOCK-FREE                                 │
│      · Ring Buffer Lock-Free (<28 ns overhead, 0 alocaciones de heap)                  │
│      · Exportador OpenMetrics / Prometheus (/metrics, /health con regla anti-cardinal) │
│      · Motor de Detección de Anomalías (RTT spikes, degradación PMTU, contención SO)  │
│      · Exportador Asíncrono de Topología de Peers a Grafo Kùzu                         │
├────────────────────────────────────────────────────────────────────────────────────────┤
│  L1: ADAPTADORES, SUBSTRATO DE KERNEL Y ENRUTAMIENTO DISTRIBUIDO                       │
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

## ⚡ PARTE III — CHECKLIST MAESTRO DE 12 INNOVACIONES DE PRÓXIMA GENERACIÓN

Las siguientes 12 dimensiones expanden la arquitectura para conferirle a `ipvn7` máxima eficiencia, interoperabilidad, resiliencia y experiencia de usuario.

---

### 1. 🛡️ Firewall Zero Trust (ZTNA) & eBPF/XDP en Kernel
Micro-segmentación nativa y descarte de paquetes maliciosos a nivel de kernel antes de alcanzar el espacio de usuario.
- [x] **Motor ZTNA Nativo (`adapters/firewall.go`)**: Micro-segmentación por DID soberano, protocolo y puerto destino.
- [x] **Política Default-Deny**: Tráfico no explícitamente autorizado se descarta sin procesar.
- [x] **Rate Limiting Per-DID**: Token Bucket elástico contra inundación y abuso de ancho de banda.
- [x] **Persistencia y Tests (`adapters/firewall_test.go`)**: Serialización JSON de reglas y 100% de tests unitarios aprobados.
- [ ] Implementación de hooks XDP en modo `native` y fallback a modo `generic` (SKB-mode) para WSL2.
- [ ] Parser seguro de cabeceras CBOR/IPv7 verificado con `libFuzzer` (> 72h sin crashes).
- [ ] Métricas de kernel exportadas a telemetría: `xdp_packets_total`, `xdp_drops_total{reason}`.

---

### 2. 📇 Sistema de Nombres Distribuido (dDNS) & Petnames
Resolución de nombres legible por humanos que resuelve el Triángulo de Zooko sin servidores centralizados.
- [x] **L2 Local (Petnames) (`dht/petnames.go`)**: Libreta de direcciones local persistente en `%USERPROFILE%\.ipv7\contacts.json`.
- [x] **Mapeo Determinista**: Resolución bidireccional `alice.ipv7` $\leftrightarrow$ `did:ipv7:...` con normalización automática.
- [x] **Derivación de IPs Virtuales**: Cálculo determinista de IPv4 (`10.7.x.x`) e IPv6 ULA (`fd07::/64`) para cada DID.
- [x] **Exportador `/etc/hosts`**: Generación de sintaxis estándar para que cualquier aplicación del SO use nombres `.ipv7`.
- [x] **PoW Criptográfico Anti-Squatting**: Algoritmo de desafío SHA-256 para reclamo de nombres y tests en `dht/petnames_test.go`.
- [ ] L1 Global (dDNS en DHT): Propagación de registros `RecordName` en la DHT Kademlia.
- [ ] Caché jerárquico con TTL dinámico y negative caching para mitigar bombardeo de consultas.

---

### 3. 🚦 QoS Jerárquico, Token Bucket & PoW Dinámico Anti-DDoS
Control de congestión de tráfico y barrera criptográfica elástica contra ataques de denegación de servicio.
- [x] **Algoritmo Token Bucket Multi-Nivel (`adapters/qos.go`)**: Clases de tráfico prioritarias (`PriorityControl`, `PriorityInteractive`, `PriorityBulk`) con reposición en tiempo real.
- [x] **PoW Dinámico Adaptativo**: Motor de emisión de desafíos de prueba de trabajo SHA-256 ante saturación de ancho de banda.
- [x] **Verificación y Bonificación Asimétrica**: Desafío computacional asimétrico: verificación en < 1 µs y acreditación de ancho de banda tras resolución exitosa.
- [x] **Suite de Pruebas Unitarias (`adapters/qos_test.go`)**: 100% de tests aprobados para burst, agotamiento, desafío y reposición temporal.
- [ ] Colas HTB + `fq_codel` en espacio de kernel para WSL2.

---

### 4. 📦 Almacenamiento DAG & Mensajería Store-and-Forward
Persistencia descentralizada asíncrona tolerante a desconexión prolongada y particiones de red (DTN).
- [ ] Direccionamiento por contenido (CID) basado en hashes BLAKE3 inmutables en `adapters/storage/`.
- [ ] Búfer circular persistente en disco para retención temporal de paquetes huérfanos.
- [ ] Algoritmo de enrutamiento epidémico (*Epidemic Routing*) para sincronización oportunista off-grid (Horizonte 9).
- [ ] Sincronización de configuraciones y blobs de estado entre nodos mediante grafos acíclicos dirigidos (DAG).

---

### 5. ⚖️ Economía de Tránsito y Reciprocidad Tit-for-Tat
Protocolo de incentivos algorítmicos para asegurar la sostenibilidad del ancho de banda en nodos Relay (DERP).
- [x] **Medidor de Balance de Tránsito (`adapters/accounting.go`)**: Contabilidad estricta de bytes TX vs RX por peer DID soberano.
- [x] **Regla de Reciprocidad Tit-for-Tat**: Clasificación automática en Tiers de Calidad de Servicio (`PRIORITY`, `NORMAL`, `BEST_EFFORT`, `THROTTLED`) para penalizar peers parásitos (*free-riders*).
- [x] **Volumen de Gracia Inicial (*Grace Volume*)**: Margen de 10 MB para permitir el bootstrap de nuevos nodos en la malla.
- [x] **Persistencia JSON y Pruebas (`adapters/accounting_test.go`)**: Almacén persistente en `%USERPROFILE%\.ipv7\transit_accounting.json` y 100% PASS.
- [x] **Integración en CLI (`cmd/ipvn7-cli/`)**: Comandos `accounting list`, `accounting status` y `accounting record`.

---

### 6. 🎛️ Plano de Control Humano Interactivo (`ipvn7-cli` & GraphQL)
Interfaz de control ergonómica para administradores humanos complementando la interfaz de IA (MCP).
- [x] **CLI Interactiva Dedicada (`cmd/ipvn7-cli/`)**: Comandos operativos `status`, `petname`, `firewall`, `accounting`, `wot`, `kuzu`, `version`.
- [x] **Gestión de Nombres y ZTNA**: Adición, eliminación, resolución y auditoría de reglas y alias directamente en terminal.
- [x] **Compilación Nativa y Cruzada**: Binarios generados en `bin/ipvn7-cli.exe` (Windows) y `bin/ipvn7-cli-linux` (WSL2/Linux).
- [ ] API local GraphQL / REST para integración con herramientas externas de administración.
- [ ] Visualización en consola con tablas ANSI, radar en texto plano y diagnóstico en vivo de enlaces.

---

### 7. 🧩 Smart Packets Extensibles con WebAssembly (Wazero)
Plataforma de plugins en sandbox de memoria para ejecutar filtros y transformaciones sin recompilar el núcleo.
- [ ] Integración del motor WebAssembly **Wazero** (100% Go puro, 0 CGO, compatible con Windows y Linux).
- [ ] Hook de interceptación de paquetes: ejecución de binarios `.wasm` para compresión, filtrado DLP o inspección de payloads.
- [ ] Aislamiento estricto de memoria: límites duros de tiempo de ejecución (timeout de 1 ms) y memoria (máximo 4 MB por plugin).

---

### 8. 🌐 Multipath QUIC Overlay (Conmutación Wi-Fi + Celular)
Transmisión simultánea y agregación de enlaces sobre interfaces físicas heterogéneas.
- [ ] Detección automática de interfaces disponibles (Wi-Fi, Ethernet, 4G/5G).
- [ ] Sub-flujos QUIC multipath coordinados: distribución de paquetes basada en RTT y pérdida de cada enlace.
- [ ] Conmutación suave sin caída de conexión (*zero-downtime failover*) ante desconexión física de Wi-Fi.

---

### 9. 🕸️ Web-of-Trust (WoT) & Reputación P2P
Red de confianza descentralizada sin cadenas de bloques para asignación de reputación entre nodos.
- [x] **Emisión de Avales Criptográficos (`dht/wot.go`)**: Estructuras `TrustVouch` firmadas digitalmente con pares de claves Ed25519.
- [x] **Cálculo de Confianza Transitiva BFS**: Algoritmo de exploración en anchura con atenuación exponencial ($0.85$ por salto hasta 3 saltos).
- [x] **Exportador a Grafo Kùzu**: Generación de sentencias openCypher (`CREATE (:Peer)-[:VOUCHES_FOR]->(:Peer)`) para análisis topológico.
- [x] **Persistencia JSON y Pruebas (`dht/wot_test.go`)**: Almacén en `%USERPROFILE%\.ipv7\wot.json` y 100% de tests unitarios aprobados.
- [x] **Comandos CLI (`cmd/ipvn7-cli/`)**: Subcomandos `wot list`, `wot vouch`, `wot score` y `wot export-kuzu`.

---

### 10. 📲 Emparejamiento Out-of-Band (QR Animado & BLE)
Bootstrap instantáneo de conexiones seguras entre dispositivos móviles y de escritorio.
- [ ] Generación de códigos QR animados en el dashboard web codificando DID, endpoints y clave Noise efímera.
- [ ] Emparejamiento por proximidad mediante balizas Bluetooth Low Energy (BLE).
- [ ] Confirmación SAS (Short Authentication String) de 6 dígitos para protección contra ataques Man-in-the-Middle (MitM).

---

### 11. 🧠 Copiloto Autónomo de Malla (AI Mesh Copilot)
Supervisión proactiva y auto-recuperación de la red asistida por la API de DeepSeek.
- [ ] Ingesta continua de anomalías detectadas por el `anomaly_engine.go`.
- [ ] Generación autónoma de diagnósticos y propuestas de reglas de firewall ZTNA delegadas al `deepseek-worker`.
- [ ] Detección de particiones de red y reconfiguración autónoma de árboles de streaming en cascada.

---

### 12. ⚡ Arquitectura Zero-Copy de Ultra-Alto Throughput
Pipeline de procesamiento de paquetes optimizado para enlaces de 10 Gbps en Linux y WSL2.
- [ ] Reutilización agresiva de búferes de memoria mediante pools sincronizados (`sync.Pool`).
- [ ] Buffers circulares mapeados en memoria (`mmap`) para paso directo de datos entre TUN y el socket UDP.
- [ ] Eliminación de alocaciones intermedias en serialización CBOR en el camino crítico.

---

## 🚀 PARTE IV — HOJA DE RUTA EJECUTIVA (HORIZONTES 6 A 9)

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│ HORIZONTE 6: RESILIENCIA EN RED REAL (7 DÍAS CONTINUOS)                     │
│ · Soak test prolongado sin fugas de memoria ni goroutines                   │
│ · Resistencia a fluctuaciones reales de ISP doméstico y Wi-Fi               │
├─────────────────────────────────────────────────────────────────────────────┤
│ HORIZONTE 7: TOPOLOGÍA HETEROGÉNEA MULTI-HOST & ZTNA                        │
│ · Despliegue de Firewall ZTNA con eBPF/XDP en Linux/WSL2                    │
│ · Sistema dDNS con libreta de Petnames local                                │
│ · Conexión multi-host Windows + Linux + ARM interconectados globalmente     │
├─────────────────────────────────────────────────────────────────────────────┤
│ HORIZONTE 8: LIVING NETWORK GRAPH, WASM & REPUTACIÓN                        │
│ · Plugins Smart Packets con WebAssembly (Wazero)                            │
│ · Red de confianza Web-of-Trust (WoT) y economía Tit-for-Tat               │
│ · Control Plane humano ipvn7-cli y agregación Multipath QUIC                │
├─────────────────────────────────────────────────────────────────────────────┤
│ HORIZONTE 9: HARDWARE FÍSICO Y COMUNICACIONES OFF-GRID                      │
│ · Transceptores LoRa, BLE Mesh y Wi-Fi Direct con persistencia DAG          │
│ · Emparejamiento por QR animado y continuidad operativa sin Internet       │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 📚 ESTRUCTURA DOCUMENTAL CONDENSADA Y CONSOLIDADA

Para garantizar la máxima legibilidad e higiene del repositorio, la documentación se organiza bajo el siguiente esquema unificado:

- [**Portal Central de Documentación (`docs/README.md`)**](README.md): Hub único de navegación.
- [**Plan Maestro y Checklist del NOS (`docs/PLAN_MAESTRO_NOS_IPVN7.md`)**](PLAN_MAESTRO_NOS_IPVN7.md): Este documento (SSOT).
- [**Especificación de Arquitectura del NOS (`docs/arquitectura/SISTEMA_OPERATIVO_RED_IPVN7.md`)**](arquitectura/SISTEMA_OPERATIVO_RED_IPVN7.md): Documento técnico de diseño.
- [**Guía de Contribución (`CONTRIBUTING.md`)**](../CONTRIBUTING.md): Directrices de código limpio y no ensuciar la raíz.
- [**Archivo Histórico (`docs/historico/`)**](historico/): Repositorio de borradores preliminares y actas de diseño histórico (01 al 04).
