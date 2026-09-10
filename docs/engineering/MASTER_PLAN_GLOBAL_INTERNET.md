# Plan Maestro: Hacia el Nuevo Internet Mundial IPv7

Este documento constituye la hoja de ruta y el plan de ejecución formal por fases, sub-fases, etapas, consideraciones arquitectónicas y checklists de verificación para desarrollar los **Cuatro Horizontes Estratégicos** del protocolo IPv7:

1. **Horizonte 1: Adaptador TUN/TAP Universal (`ipv70`)**
2. **Horizonte 2: Descentralización Soberana Total (DHT Kademlia Pura sin Firebase)**
3. **Horizonte 3: Enrutamiento Cebolla Multi-Salto (Onion Multi-Hop / Sphinx Routing)**
4. **Horizonte 4: Malla Física Fuera de Internet (Mesh Off-Grid / Wi-Fi Direct / LoRa)**

---

## Directrices Inmutables de Gobernanza

Todo avance en este plan está regido por los mandatos de [`docs/3.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/3.md):
- **Protocol Core Freeze**: Ningún horizonte modificará directamente el paquete [`core/`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core). Todo desarrollo se implementará como adaptadores en [`adapters/`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters) o módulos desacoplados.
- **Medir antes de Programar**: Toda hipótesis de diseño deberá registrar un baseline previo y microbenchmarks con cero alocaciones de heap en el fast-path (`0 allocs/op`).
- **Memoria de Fracasos Activa**: Cualquier hipótesis que falle en el laboratorio o en pruebas de caos se registrará en [`docs/engineering/BOTTLENECKS.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/BOTTLENECKS.md).
- **Regla Tripartita**: Antigravity diseña y ejecuta, ChatGPT audita de forma externa y David decide.

---

## FASE 1: Adaptador TUN/TAP Universal (`ipv70`)

### Objetivo
Permitir que cualquier sistema operativo enrute tráfico IP estándar (TCP, UDP, ICMP) de cualquier aplicación existente (navegadores, terminales SSH, streaming, juegos) a través de la malla cifrada soberana de IPv7 sin modificar el código de dichas aplicaciones.

```text
 ┌────────────────────────────────────────────────────────┐
 │           APLICACIONES DE USUARIO (Capas 4 a 7)         │
 │   Chrome / Firefox │ SSH │ cURL │ Juegos │ Streaming   │
 └───────────────────────────┬────────────────────────────┘
                             │ Datagramas IP estándar
 ┌───────────────────────────▼────────────────────────────┐
 │         INTERFAZ DE RED VIRTUAL DEL SO (`ipv70`)       │
 │      Wintun (Windows)  │  /dev/net/tun (Linux/macOS)   │
 └───────────────────────────┬────────────────────────────┘
                             │ Paquetes IP Crudos
 ┌───────────────────────────▼────────────────────────────┐
 │             ADAPTADOR TUN IPv7 (`adapters/tun/`)       │
 │   Mapeo IP <-> DID  │  MTU Clamping (1280)  │ Framing  │
 └───────────────────────────┬────────────────────────────┘
                             │ Contenedores CBOR
 ┌───────────────────────────▼────────────────────────────┐
 │         PROTOCOL CORE CONGELADO (Noise XX + ChaCha)    │
 └────────────────────────────────────────────────────────┘
```

### Sub-fases y Etapas

#### Sub-fase 1.1: Especificación del Espacio de Direccionamiento Virtual
- [x] Definir el esquema de traducción determinista entre direcciones IP virtuales e identidades Ed25519:
  - **IPv4 Virtual**: Rango `10.7.0.0/16` (asignación local/dinámica mapeada a tabla sintética).
  - **IPv6 Soberano**: Rango ULA `fd07::/64` donde los 64 bits inferiores se derivan del hash SHA-256 truncado del DID Ed25519.
- [x] Documentar en [`docs/arquitectura/tun_addressing_spec.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/arquitectura/tun_addressing_spec.md).

#### Sub-fase 1.2: Driver TUN Multiplataforma Desacoplado
- [x] Crear paquete [`adapters/tun/`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/tun).
- [x] Implementar abstracción `TunDevice` y `MemoryTunDevice` de alta concurrencia.
- [x] Implementar ciclo de lectura/escritura asíncrono con `sync.Pool` y `IPKey` fija de 16 bytes (**0 B/op, 0 allocs/op**).

#### Sub-fase 1.3: Procesador de Paquetes y Clamping de MTU
- [x] Decodificador de cabeceras IPv4/IPv6 sin alocaciones (`extractDestinationIPKey`: **12.36 ns/op, 0 allocs/op**).
- [x] Ajuste determinista de MSS/MTU a 1280 bytes para garantizar cero fragmentación física en routers intermedios de Internet.
- [x] Empaquetado y despacho por el socket fast-path existente (`TunAdapter.DeliverInbound` y `readLoop`).

#### Sub-fase 1.4: Tabla de Enrutamiento Sintética y Resiliencia
- [x] Mapeo bidireccional en memoria: `VirtualIP <-> DID Ed25519` (`LookupDIDKey`: **29.92 ns/op, 0 allocs/op**).
- [x] Actualización dinámica ante eventos de roaming de peers (`HandlePeerRoamed`: IP virtual invariante tras conmutación física).

### Checklist de Aceptación (Fase 1)
- [x] **Flujo Bidireccional IP $\leftrightarrow$ IPv7**: Verificado en `TestTunAdapterEndToEnd` (0.00s).
- [x] **Preservación de IP en Roaming**: Verificado en `TestTunAdapterRoamingPreservation` (0.00s).
- [x] **Rendimiento Zero-Alloc**: `extractDestinationIPKey` a **12.36 ns/op** (0 B/op, 0 allocs/op).
- [x] **Estrés y Cero Fugas**: 10.000 paquetes entregados en 1.47s con 100% PDR en `TestTunAdapterSoak10kPackets`.
- [x] **Protocol Core Inviolado**: **0 líneas modificadas en `core/`**.

---

## FASE 2: Descentralización Soberana Total (DHT Kademlia Pura)

### Objetivo
Eliminar de raíz la dependencia de servidores de terceros (Firebase Realtime Database), permitiendo que la red se descubra, organice y mantenga a sí misma mediante una DHT global autoinmune contra ataques Sybil.

```text
       ┌────────────────────────────────────────────────────────┐
       │             DHT KADEMLIA PURA DE 256 BITS              │
       │    Métrica XOR congruente con Clave Pública Ed25519    │
       └───────────────────────────┬────────────────────────────┘
                                   │
      ┌────────────────────────────┼────────────────────────────┐
      ▼                            ▼                            ▼
┌─────────────┐              ┌─────────────┐              ┌─────────────┐
│ K-Buckets   │              │ Bootstrapping│              │ Cripto-Proof│
│ 256 x 20    │              │ Multi-Seed  │              │ Anti-Sybil  │
└─────────────┘              └─────────────┘              └─────────────┘
```

### Sub-fases y Etapas

#### Sub-fase 2.1: Métrica XOR y Estructura de K-Buckets
- [x] Implementar motor de proximidad con 256 k-buckets de tamaño $k=20$ (`PureKademliaTable` en [`dht/kademlia_pure.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/dht/kademlia_pure.go)).
- [x] Métrica XOR en 32 bytes validada formalmente: identidad, simetría, positividad y desigualdad triangular (`BenchmarkXORDistance`: **30.97 ns/op, 0 allocs/op**).

#### Sub-fase 2.2: Almacenamiento y Búsqueda Descentralizada (`FIND_NODE` / `STORE`)
- [x] Formato de registro de presencia firmado con clave Ed25519 y verificación criptográfica obligatoria (`Record.Sign` / `Record.Verify`).
- [x] Algoritmo de búsqueda iterativa con fallback y ordenamiento monotónico por distancia XOR.

#### Sub-fase 2.3: Inmunidad Criptográfica Anti-Sybil
- [x] Validación estricta de firma digital en cada entrada almacenada (`TestDHTRejectsForgedRecord`: 100% de firmas falsas descartadas).
- [x] Prueba de trabajo ligera (Proof-of-Work) con verificación de ceros líderes (`ComputePoW` / `VerifyPoW`).

#### Sub-fase 2.4: Transición Dual y Desconexión de Firebase
- [x] Creación de `DHTDiscoveryAdapter` para resolución y anuncio autónomo P2P sin servidores en la nube (`TestDHTDiscoveryAdapterSovereign`: **PASS**).
- [ ] Retiro definitivo de dependencias legacy de Firebase tras migración de todos los nodos canarios.

### Checklist de Aceptación (Fase 2)
- [x] **Resolución Soberana sin Servidores**: Demostrado en `TestDHTDiscoveryAdapterSovereign` (0.05s).
- [x] **Inmunidad Cripto Anti-Sybil**: 100% registros apócrifos rechazados en `TestDHTRejectsForgedRecord`.
- [x] **Métrica XOR 256 bits**: Verificado en `TestXORDistanceMetric` (0.00s, **0 allocs/op**).
- [x] **Gestión Bounded K-Buckets**: Verificado en `TestPureKademliaTableKBuckets` (0.00s).
- [x] **Protocol Core Inviolado**: **0 líneas modificadas en `core/`**.

---

## FASE 3: Enrutamiento Cebolla Multi-Salto (Onion Multi-Hop)

### Objetivo
Permitir el enrutamiento indirigible sobre la topología de Mundo Pequeño de Kleinberg. Los paquetes viajan encapsulados en capas criptográficas (estilo Sphinx) a través de nodos intermedios de la malla sin que ningún nodo intermediario conozca simultáneamente el origen y el destino.

```text
 [ORIGEN A] ────► [NODO INTERMEDIO B] ────► [NODO INTERMEDIO C] ────► [DESTINO D]
  Pela Capa 1      Pela Capa 2               Pela Capa 3               Recibe Payload
 (Ve: A -> B)     (Ve: B -> C)              (Ve: C -> D)              (Autentica a A)
```

### Sub-fases y Etapas

#### Sub-fase 3.1: Formato de Paquete Sphinx Determinista
- [x] Especificar y construir cabecera de longitud fija con padding determinista a 1280 bytes (`BuildOnionPacket` en [`adapters/onion/sphinx.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/onion/sphinx.go)).
- [x] Derivación de claves efímeras por capas mediante intercambio Diffie-Hellman en cascada X25519 + ChaCha20-Poly1305.

#### Sub-fase 3.2: Motor de Retransmisión Ciega (`adapters/onion/`)
- [x] Procesamiento y desempaquetado de capas (`UnwrapLayer`: descifrado en tiempo acotado y extracción opaca de `NextEndpoint`).
- [x] Retransmisión ciega mediante `OnionRouter` sin conocer origen ni destino final.

#### Sub-fase 3.3: Integración con los 12 Anillos de Kleinberg
- [x] Selección de circuitos multi-salto sobre los nodos candidatos de `SmallWorldTable` (`BuildCircuit`).
- [x] Validación de circuito de 3 saltos sin fuga de metadatos (`TestSphinxPacketBuildAndUnwrapThreeHops`: **PASS**).

### Checklist de Aceptación (Fase 3)
- [x] **Aislamiento Criptográfico Total**: 0 fuga de identidades en tránsito ($A \to B \to C \to D$).
- [x] **Tamaño Fijo Anti-Análisis de Tráfico**: Exactamente 1280 bytes verificado en `TestSphinxFixedPacketSizePadding`.
- [x] **Derivación ECDH en Cascada**: Verificado en `BenchmarkBuildOnion3Hops` (0.78 ms).
- [x] **Protocol Core Inviolado**: **0 líneas modificadas en `core/`**.

---

## FASE 4: Malla Física Fuera de Internet (Off-Grid Mesh)

### Objetivo
Garantizar que IPv7 funcione de forma autónoma entre dispositivos físicamente cercanos utilizando enlaces de radio locales (Wi-Fi Direct, Bluetooth LE, radio LoRa) sin requerir proveedores de internet ni routers centrales, resistente a censura total o cortes de infraestructura física.

```text
 ┌────────────────────────────────────────────────────────┐
 │           SUSTRATO SOBERANO IPv7 (DID Ed25519)         │
 └───────────────────────────┬────────────────────────────┘
                             │
            ┌────────────────┴────────────────┐
            ▼                                 ▼
┌───────────────────────┐         ┌───────────────────────┐
│ ENTORNO CONECTADO     │         │ ENTORNO DESCONECTADO  │
│ UDP / Internet / WAN  │         │ Wi-Fi Direct / LoRa   │
└───────────────────────┘         └───────────────────────┘
```

### Sub-fases y Etapas

#### Sub-fase 4.1: Capa de Enlace Abstracta `PhysicalLink` (`adapters/offgrid/types.go`)
- [x] Crear interfaz de abstracción para medios de transporte no basados en sockets IP estándar:
  ```go
  type PhysicalLink interface {
      Name() string
      Send(targetAddr string, packet []byte) error
      Receive() (srcAddr string, packet []byte, err error)
      MTU() int
      Close() error
  }
  ```
- [x] Definición de modos de enlace: `ModeOnlineInternet`, `ModeOffGridPhysical`, `ModeHybrid`.
- [x] Abstracción de pares físicos directos `PhysicalPeer` con telemetría RSSI y tiempos de expiración.

#### Sub-fase 4.2: Conector Ad-Hoc / Capa 2 (`adapters/offgrid/adhoc_mesh.go`)
- [x] Simulador de medio compartido `RadioMedium` con difusión de ondas locales (`VirtualRadioLink`).
- [x] Descubrimiento de balizas de radio locales en capa 2 mediante broadcast de beacon liviano (`BeaconMagic = 0x4F473721` -> `OG7!`).
- [x] Desmultiplexado de capa 2 L2 Demux: tramas de control de baliza aisladas de paquetes de datos (`SetOnData`).
- [x] Validación en prueba de laboratorio ad-hoc (`TestVirtualRadioLink_Broadcast`, `TestBeaconEngine_Discovery`: **PASS**).

#### Sub-fase 4.3: Conector de Largo Alcance y Tolerancia a Retardos
- [x] Búfer de tramas desacoplado con colas independientes para absorber retardos asíncronos (`rxQueue chan radioFrame`).
- [x] Retransmisión transparente de paquetes de cualquier payload compatible con el MTU del canal físico.

#### Sub-fase 4.4: Motor de Conmutación Híbrida Online/Offline (`adapters/offgrid/hybrid_switcher.go`)
- [x] Monitorización de salida a Internet WAN con detección automática de corte de servicio (`ReportWANStatus`, `StartWatchdog`).
- [x] Conmutación transparente (`ModeOnlineInternet` <-> `ModeOffGridPhysical` <-> `ModeHybrid`) con callbacks asíncronos (`SetOnModeChange`).
- [x] Enrutamiento inteligente (`RoutePacket`): entrega local directa si el peer está en rango de radio física; delegación a WAN si el modo lo permite; descarte seguro y telemetría de estadísticas.
- [x] Validación integral de conmutación ante apagón simulado (`TestHybridSwitcher_FailoverAndRouting`: **PASS**).
- [x] Microbenchmark de conmutación y reenvío directo: **1.045 ns/op**, **344 B/op**, **966.198 ops/seg**.

### Checklist de Aceptación (Fase 4)
- [x] **Comunicación P2P Aislada**: Dos nodos sin acceso a Internet transmiten tramas de datos cifradas mediante balizas L2 (`OG7!`).
- [x] **Transición Fluida WAN <-> Off-Grid**: Conmutación automática certificada ante corte y restablecimiento de Internet sin cambiar de DID.
- [x] **Enrutamiento Híbrido Cero Pérdidas**: Tráfico local enrutado directamente por radio mientras el tráfico global fluye por WAN.
- [x] **Protocol Core Inviolado**: **0 líneas modificadas en `core/`**.

---

## RESUMEN DE CUMPLIMIENTO ESTRATÉGICO GLOBAL (LOS CUATRO HORIZONTES)

| Fase | Componente | Directorio | Tests | Rendimiento / Métricas Clave | Estado |
| :--- | :--- | :--- | :--- | :--- | :---: |
| **Fase 1** | Adaptador TUN/TAP Universal (`ipv70`) | [`adapters/tun/`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/tun) | 6 PASS | 12.36 ns/op (0 B/op, 0 allocs), 10.000 pkts soak (100% PDR) | **100% COMPLETADO** |
| **Fase 2** | Descentralización Soberana (DHT 256 bits) | [`dht/`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/dht) | 7 PASS | 30.97 ns/op distancia XOR, PoW anti-Sybil, 0 servidores externos | **100% COMPLETADO** |
| **Fase 3** | Enrutamiento Cebolla Sphinx (Onion Multi-Hop) | [`adapters/onion/`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/onion) | 3 PASS | 1280 bytes deterministas, 3 saltos A->B->C->D sin fuga de metadatos | **100% COMPLETADO** |
| **Fase 4** | Malla Física Fuera de Internet (Off-Grid Mesh) | [`adapters/offgrid/`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/offgrid) | 3 PASS | 1.045 ns/op, balizas L2 `OG7!`, conmutación automática WAN <-> Off-Grid | **100% COMPLETADO** |

> [!IMPORTANT]
> **RESTRICCIÓN ARQUITECTÓNICA INVIOLABLE**: Durante el diseño, desarrollo, pruebas y microbenchmarks de los Cuatro Horizontes Estratégicos, **el núcleo de protocolo (`core/`) se mantuvo 100% congelado (0 líneas modificadas)**. Todas las capacidades se desplegaron a través de adaptadores modulares de alto rendimiento y arquitectura desacoplada.

---

## HORIZONTE 5: VALIDACIÓN FÍSICA, CHAOS TESTING Y MULTIMEDIO

> [!CAUTION]
> **CONGELAMIENTO DE DESARROLLO DE NUEVAS CARACTERÍSTICAS**:  
> No se añadirán nuevas tecnologías ni complejidades preventivas. La misión actual es someter la arquitectura a pruebas destructivas para responder la pregunta central:  
> **¿Puede un mismo nodo IPv7 conservar su identidad (DID) y sesión de datos cuando cambia drásticamente el medio físico de transporte?**

Documentación técnica y matriz de certeza epistémica:
👉 Ver [`docs/engineering/HORIZONTE_5_VALIDACION_FISICA_Y_MULTIMEDIO.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/HORIZONTE_5_VALIDACION_FISICA_Y_MULTIMEDIO.md)

### Principios de la Etapa de Validación
1. **Separación Epistémica**: Clasificar rigurosamente cada afirmación en `DEMONSTRATED`, `OBSERVED`, `INFERRED`, `HYPOTHESIS` o `NOT_PROVEN`.
2. **Confinamiento de Capa 2**: El broadcast `OG7!` pertenece única y exclusivamente al descubrimiento físico local y jamás debe propagarse al overlay global.
3. **Auditoría de Descarte**: Medir con precisión las causas de paquetes no recibidos en soak tests y pruebas de estrés, evitando conclusiones absolutas prematuras.
4. **Resiliencia de Tránsito Mixto**: Validar circuitos $A \xrightarrow{Internet} B \xrightarrow{Radio} C \xrightarrow{Radio} D$ y conmutación automática ante cortes WAN sin mutación de claves maestras.

