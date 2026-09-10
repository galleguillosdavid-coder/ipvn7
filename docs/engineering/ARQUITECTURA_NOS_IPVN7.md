# ipvn7: Arquitectura del Sistema Operativo de Red Descentralizado (NOS)
## Especificación Maestra de Capacidades, Virtudes y Plan de Evolución

> **DOCUMENTO RECTOR DE ARQUITECTURA**  
> **Versión**: 1.0.0-PROPOSED  
> **Ámbito**: Repositorio galleguillosdavid-coder/ipvn7  
> **Regla de Continuidad**: Hereda e inmuta la base criptográfica y el Core Freeze de Ipv7@v0.5.0-chaos.audit.

---

## 1. Visión General: De VPN Mesh a Sistema Operativo de Red (NOS)

La transición desde **IPv7** hacia **ipvn7** representa un salto de paradigma: el paso de un conjunto de protocolos de laboratorio a un **Sistema Operativo de Red Descentralizado (NOS)** de grado de producción.

Mientras las soluciones convencionales de malla (Tailscale, Nebula, ZeroTier) actúan como overlays centralizados dependientes de servidores de coordinación propietarios y confianza ciega en el tráfico interno, **ipvn7** se concibe como un sustrato soberano, tolerante a particiones físicas, resistente a la interceptación cuántica y gobernado mediante observabilidad formal asistida por IA.

`	ext
┌──────────────────────────────────────────────────────────────────────────────────┐
│                         ipvn7 NETWORK OPERATING SYSTEM                           │
├──────────────────────────────────────────────────────────────────────────────────┤
│ [L4] GOBERNANZA & CONTROL PLANE                                                  │
│      ipvn7-cli │ API GraphQL / REST │ Web Dashboard SPA │ Servidor MCP (IA)      │
├──────────────────────────────────────────────────────────────────────────────────┤
│ [L3] MEMORIA ESTRUCTURAL & OBSERVABILIDAD                                        │
│      Kùzu Graph Engine │ Prometheus OpenMetrics │ DeepSeek Multimodal Worker     │
├──────────────────────────────────────────────────────────────────────────────────┤
│ [L2] SERVICIOS AUTÓNOMOS DE RED (NUEVAS VIRTUDES NOS)                            │
│      ZTNA Firewall (eBPF/WFP) │ dDNS Petnames │ DAG Store │ QoS Token Bucket     │
├──────────────────────────────────────────────────────────────────────────────────┤
│ [L1] ADAPTADORES MULTIMEDIO & TRANSPORTE                                         │
│      TUN/TAP (fd07::/64) │ Sphinx (1280B) │ Off-Grid (LoRa, BLE, WiFi-D) │ Relay │
├──────────────────────────────────────────────────────────────────────────────────┤
│ [L0] PROTOCOL CORE (STRICT FREEZE)                                               │
│      DID Ed25519 │ Noise XX │ ChaCha20-Poly1305 │ PQC Kyber │ Greedy XOR Routing │
└──────────────────────────────────────────────────────────────────────────────────┘
`

---

## 2. Los Siete Pilares y Nuevas Virtudes Arquitectónicas

A partir del análisis de brechas operativas registrado en [docs/conversa.md](../conversa.md), se formalizan las 7 dimensiones que convierten a ipvn7 en un NOS completo:

### Dimensión 1: Políticas Zero Trust (ZTNA) y Firewall P2P Nativo
- **Problema Abordado**: Al rutear tráfico arbitrario del sistema operativo mediante interfaces virtuales TUN (d07::/64 y 10.7.0.0/16), un peer comprometido podría intentar escaneos o movimiento lateral hacia servicios locales del host anfitrión.
- **Solución Arquitectónica**:
  - Motor de micro-segmentación y filtrado granular por DID, protocolo y puerto virtual.
  - **Política de Denegación por Defecto (Default-Deny)**: Ningún nodo puede abrir un socket hacia otro si no existe una regla explícita aprobada o establecida mediante handshake mutuo.
  - **Aceleración a Nivel de Kernel**: En entornos Linux y WSL2, integración con ganchos **eBPF / XDP** para descarte de datagramas no autorizados en la propia tarjeta de red antes de consumir ciclos de CPU en user-space. En Windows, integración conceptual con **Windows Filtering Platform (WFP)**.
  - **Copiloto de Seguridad IA**: Capacidad de que el agente DeepSeek audite bitácoras de descarte y sugiera reglas ZTNA en lenguaje natural.

### Dimensión 2: Sistema de Nombres Distribuido (dDNS) y Petnames
- **Problema Abordado**: Los identificadores criptográficos puros (did:ipv7:84dc2df4649a93a1...) garantizan unicidad e inviolabilidad matemática, pero son cognitivamente inviables para usuarios y administradores humanos.
- **Solución Arquitectónica**:
  - **Arquitectura Híbrida de Nombres**: Resolución basada en el Triángulo de Zooko equilibrado mediante **Petnames Locales** y registros de nombres anclados al DHT de Kademlia.
  - **Sintaxis Canónica**: Asignación de alias memorizables (ej. lice.ipvn7, gateway-chile.ipvn7) validados criptográficamente contra el DID mediante firmas Ed25519 con tiempo de expiración.
  - **Desacoplamiento del DNS Tradicional**: Resolución 100% interna sin llamadas a servidores raíz de ICANN, inmune a secuestro de DNS (DNS spoofing o poisoning).

### Dimensión 3: Resistencia Sybil, Calidad de Servicio (QoS) y Anti-DDoS
- **Problema Abordado**: En redes públicas y no permisionadas, atacantes pueden crear miles de identidades falsas (ataques Sybil) para saturar nodos de tránsito y agotar búferes UDP.
- **Solución Arquitectónica**:
  - **Token Bucket Jerárquico por Identidad**: Limitador de tasa (*traffic shaping*) acoplado al DID de origen; los nodos vecinos tienen cuotas de ráfaga y ancho de banda sostenido.
  - **Proof of Work (PoW) Adaptativo**: Peaje criptográfico ligero basado en hashing de baja dificultad que se incrementa dinámicamente cuando un nodo entra en contención o bajo sospecha de ataque DoS, forzando al emisor a invertir ciclos de CPU.
  - **Gestión de Prioridades**: Tráfico de señalización crítica (balizas L2, handshakes de enrutamiento) tiene prioridad estricta sobre datagramas de datos masivos.

### Dimensión 4: Almacenamiento de Estado Descentralizado (DAG Store & Forward)
- **Problema Abordado**: El stack actual sobresale en transporte síncrono en tiempo real (streaming, túneles, VoIP, chat E2EE), pero si un nodo receptor está apagado o en una partición de red, el mensaje o archivo se pierde.
- **Solución Arquitectónica**:
  - **Capa Store-and-Forward Basada en DAG**: Estructura de Grafo Acíclico Dirigido con direccionamiento por contenido (**CID**) similar a IPFS.
  - **Entrega Asíncrona Resiliente**: Los mensajes cifrados pueden ser custodiados temporalmente por nodos puente en la malla hasta que el destinatario restablezca contacto topológico.
  - **Sincronización de Configuraciones Off-Grid**: Distribución de políticas de red y llaves públicas de vecinos sin requerir conexión simultánea a Internet.

### Dimensión 5: Economía de Tránsito e Incentivos de Red
- **Problema Abordado**: El enrutamiento cebolla (Sphinx de 3 saltos) y los nodos Relay (DERP) asumen altruismo puro. Para escalar a nivel comunitario global, se requiere reciprocidad sostenible.
- **Solución Arquitectónica**:
  - **Algoritmo Tit-for-Tat de Ancho de Banda**: Inspirado en BitTorrent, los nodos priorizan el tránsito de peers que a su vez retransmiten tráfico de la comunidad.
  - **Contabilidad Criptográfica de Cuotas**: Registro local de créditos y débitos de bytes cursados sin fuga de metadatos mediante atestaciones criptográficas ligeras.

### Dimensión 6: Plano de Control (Control Plane) y CLI para Humanos
- **Problema Abordado**: La infraestructura cuenta con endpoints OpenMetrics y servidores MCP para IAs, pero carece de un instrumental de control intuitivo y directo para administradores de sistemas humanos.
- **Solución Arquitectónica**:
  - **ipvn7-cli**: Herramienta de línea de comandos interactiva escrita en Go para consultar el estado del nodo, añadir vecinos, inspeccionar la tabla de enrutamiento XOR, medir latencias RTT y alterar políticas ZTNA en caliente.
  - **API Local GraphQL / REST**: Endpoint administrativo local protegido por token efímero para conectar dashboards corporativos o consolas de monitoreo existentes.

### Dimensión 7: Extensibilidad WASM (Smart Packets)
- **Problema Abordado**: Probar protocolos experimentales o filtros de paquetes a medida suele requerir recompilar y reiniciar el binario del nodo, violando la continuidad del servicio.
- **Solución Arquitectónica**:
  - **Runtime WebAssembly Embebido (Wazero/Wasmtime)**: Capacidad de cargar módulos .wasm en caliente en el pipeline de adaptadores.
  - **Filtros Dinámicos e Interceptores**: Permite definir transformaciones, métricas personalizadas o inspecciones de seguridad programables en Rust o Go compiladas a Wasm sin tocar una sola línea de core/.

---

## 3. Matriz Epistémica de Evolución de Capacidades

Conforme al rigor epistémico del proyecto, cada dimensión se cataloga honestamente según su grado de madurez:

| Dimensión | Módulos Clave | Estado Epistémico | Hito Asociado |
| :--- | :--- | :---: | :---: |
| **Núcleo Criptográfico & DID** | core/ (Ed25519, Noise, Kyber) | **DEMONSTRATED** (LAB_SIMULATED) | Horizonte 1–5 (Heredado) |
| **Transporte Universal TUN** | dapters/tun/ | **DEMONSTRATED** (LAB_SIMULATED) | Horizonte 1 (Heredado) |
| **Sphinx Onion 1280B** | dapters/onion/ | **DEMONSTRATED** (LAB_SIMULATED) | Horizonte 3 (Heredado) |
| **Telemetría Lock-Free** | 	elemetry/ (<28ns, 0 allocs) | **DEMONSTRATED** (LAB_SIMULATED) | Horizonte 5 (Heredado) |
| **Zero Trust Firewall (ZTNA)** | dapters/firewall/, ebpf/ | **PROPOSED** (Diseño) | Horizonte 7 |
| **Sistema de Nombres (dDNS)** | dapters/names/, dht/ | **IN_LAB** (Prototipo) | Horizonte 7 |
| **QoS & Resistencia Sybil** | dapters/qos/, core/pow.go | **PROPOSED** (Arquitectura) | Horizonte 6 |
| **Almacenamiento DAG Store** | dapters/storage/dag_store.go | **CONCEPTUAL** | Horizonte 8 |
| **Economía de Tránsito** | dapters/accounting/ | **CONCEPTUAL** | Horizonte 8 |
| **Control Plane (ipvn7-cli)** | cmd/ipvn7-cli/ | **PROPOSED** (Diseño) | Horizonte 6 |
| **Extensibilidad Wasm** | dapters/wasm/ | **CONCEPTUAL** | Horizonte 9 |

---

## 4. Gobernanza del Desarrollo y Próximos Pasos Inmediatos

1. **Fase Inmediata (Horizonte 6)**:
   - Despliegue del Soak de 7 días continuos en red heterogénea (Windows + Linux en WSL2).
   - Implementación del primer prototipo de ipvn7-cli para inspección interactiva humana.
   - Creación del módulo base de dapters/firewall/ con reglas de micro-segmentación ZTNA.
2. **Uso Estratégico de IA y Grafos**:
   - **DeepSeek-V4.1-Flash**: Encargado de auditorías visuales y revisiones de código masivo para mantener el consumo de tokens en Antigravity cercano a cero.
   - **Kùzu Graph Engine**: Consolidado como la memoria histórica de topología, relaciones entre DIDs y correlación de incidentes fuera del camino crítico.
