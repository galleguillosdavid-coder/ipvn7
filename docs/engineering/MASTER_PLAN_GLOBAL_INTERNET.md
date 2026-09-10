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
- [ ] Actualización dinámica ante eventos de roaming de peers.

### Checklist de Aceptación (Fase 1)
- [ ] `ping` ICMP exitoso entre PC Windows y máquina WSL2 a través de IPs `10.7.x.x` o `fd07::x`.
- [ ] Conexión SSH y descarga HTTP vía `curl` funcionando transparentemente sobre la interfaz `ipv70`.
- [ ] Throughput medido con `iperf` $\ge 300$ Mbps en enlace local.
- [ ] Cero fugas de memoria tras 100.000 paquetes TCP procesados.
- [ ] Cero líneas modificadas en [`core/`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core).

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
- [ ] Refactorizar el motor de proximidad en [`dht/`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/dht) para implementar 256 k-buckets de tamaño $k=20$.
- [ ] Implementar algoritmo de refresco periódico ($t_{\text{refresh}} = 1\text{ h}$) y reemplazo por tiempo de respuesta (LRU con ping de sondeo).

#### Sub-fase 2.2: Almacenamiento y Búsqueda Descentralizada (`FIND_NODE` / `STORE`)
- [ ] Formato de registro de presencia firmado con clave Ed25519.
- [ ] Algoritmo voraz $\alpha=3$ de búsqueda concurrente con convergencia en $O(\log N)$ saltos.

#### Sub-fase 2.3: Inmunidad Criptográfica Anti-Sybil
- [ ] Validación estricta de firma digital en cada entrada almacenada.
- [ ] Prueba de trabajo ligera (Proof-of-Work) efímera en la cabecera del registro para prevenir inundación masiva de tablas.

#### Sub-fase 2.4: Transición Dual y Desconexión de Firebase
- [ ] Despliegue de nodos semilla (seed nodes) fijos comunitarios.
- [ ] Modo híbrido transitorio: descubrimiento primario por DHT con fallback a Firebase.
- [ ] Retiro formal y desconexión total del cliente Firebase tras certificar 100% de éxito en DHT aislada.

### Checklist de Aceptación (Fase 2)
- [ ] Redescubrimiento de un nodo móvil tras conmutar de IP en $< 2.000$ ms utilizando exclusivamente la DHT.
- [ ] Resistencia demostrada a la caída simultánea del 40% de los nodos de la red sin pérdida de accesibilidad.
- [ ] Rechazo del 100% de intentos de inyección de registros mal firmados o apócrifos.
- [ ] Cero llamadas HTTP hacia `firebaseio.com` durante 24 horas continuas de operación.

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
- [ ] Especificar cabecera de longitud fija con padding determinista (cero pistas por tamaño de paquete).
- [ ] Derivación de claves efímeras mediante intercambio Diffie-Hellman en cascada con las claves X25519 de los nodos del circuito.

#### Sub-fase 3.2: Motor de Retransmisión Ciega (`adapters/onion/`)
- [ ] Procesamiento de capas en tiempo constante $O(1)$ para evitar ataques de canal lateral basados en tiempo de CPU.
- [ ] Registro de hashes efímeros de un solo uso para prevenir ataques de repetición o ramificación en nodos intermedios.

#### Sub-fase 3.3: Integración con los 12 Anillos de Kleinberg
- [ ] Selección de circuitos de 3 o 5 saltos seleccionados aleatoriamente dentro de los anillos logarítmicos de la tabla de Mundo Pequeño.
- [ ] Fallback automático si un nodo intermediario se apaga o degrada.

### Checklist de Aceptación (Fase 3)
- [ ] Demostración de que ningún nodo intermediario puede descifrar el payload ni conocer la identidad del otro extremo.
- [ ] Sobrecarga criptográfica $< 64$ bytes por salto.
- [ ] Latencia adicional por salto $< 4$ ms en red local / $< 20$ ms en WAN.
- [ ] Resiliencia ante terminación intempestiva de un nodo intermedio con re-enrutamiento automático en $< 500$ ms.

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

#### Sub-fase 4.1: Capa de Enlace Abstracta `PhysicalLinkAdapter`
- [ ] Crear interfaz de abstracción para medios de transporte no basados en sockets IP estándar:
  ```go
  type PhysicalLink interface {
      Send(dst []byte, payload []byte) error
      Receive() (src []byte, payload []byte, err error)
      MTU() int
  }
  ```

#### Sub-fase 4.2: Conector Wi-Fi Direct / Ad-Hoc
- [ ] Descubrimiento de balizas de radio locales en capa 2 mediante broadcast de beacon liviano.
- [ ] Enlace directo P2P entre tarjetas inalámbricas de ordenadores y smartphones.

#### Sub-fase 4.3: Conector LoRa de Largo Alcance
- [ ] Compresión de cabeceras CBOR a formato ultra-compacto (< 16 bytes de overhead).
- [ ] Soporte de transferencias asíncronas con tolerancia a retardos de hasta 10 segundos.

#### Sub-fase 4.4: Motor de Conmutación Híbrida Online/Offline
- [ ] Monitorización de salida a Internet WAN.
- [ ] Conmutación suave y transparente: las aplicaciones locales continúan comunicándose por la malla física sin cambiar de DID.

### Checklist de Aceptación (Fase 4)
- [ ] Dos dispositivos aislados físicamente de Internet establecen comunicación cifrada E2EE directa.
- [ ] Transición fluida entre modo Internet y modo Off-Grid sin pérdida de identidad soberana.
- [ ] Mensajería de texto y sincronización de datos completada exitosamente a través del enlace físico ad-hoc.
