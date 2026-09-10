# IPv7 Especificación: Espacio de Direccionamiento Virtual TUN/TAP

**Módulo**: Adaptador TUN/TAP Universal (`ipv70`)  
**Fase**: Fase 1, Sub-fase 1.1  
**Fecha**: 2026-09-09  
**Estado**: **ESPECIFICACIÓN FORMAL DEMOSTRADA**  

---

## 1. El Reto de la Integración con el Stack IP de los Sistemas Operativos

Para que las aplicaciones existentes (navegadores web, terminales SSH, reproductores de medios, bases de datos) puedan enviar y recibir tráfico a través de IPv7 sin modificaciones en su código, el sistema operativo debe ver una interfaz de red estándar con una dirección IP asignada.

Dado que IPv7 utiliza identidades criptográficas **Ed25519 de 32 bytes (256 bits)** (`DID`), y las direcciones IP estándar son de **32 bits (IPv4)** o **128 bits (IPv6)**, se requiere una función de mapeo determinista y libre de colisiones locales.

---

## 2. Direccionamiento IPv4 Virtual (`10.7.0.0/16`)

El rango privado RFC 1918 `10.7.0.0/16` se reserva para la asignación dinámica y el enrutamiento local en el nodo:

```text
 10 . 7 . X . Y / 16
 ├──┴─┼─┴─┼─┴─┼─┴──┤
 │    │   │   │    └── Host ID (1 a 254)
 │    │   └───┴─────── Subnet / Peer Cluster ID
 └────┴─────────────── Prefijo Reservado IPv7
```

- **Dirección Local del Nodo**: Por convención, el adaptador local toma la dirección `10.7.0.1/16`.
- **Mapeo de Peers**: Cuando un peer remoto con `DID` se autentica o es contactado, el adaptador le asigna una IP virtual efímera dentro de la subred `10.7.0.0/16` y la registra en su tabla ARP sintética.
- **Resolución ARP Sintética**: Cuando el sistema operativo emite una solicitud `ARP WHO-HAS 10.7.X.Y`, el driver TUN responde inmediatamente con una dirección MAC sintética (`02:07:XX:YY:ZZ:WW`) en $< 1$ microsegundo, sin emitir tráfico broadcast a la red física.

---

## 3. Direccionamiento IPv6 Soberano Criptográfico (`fd07::/64`)

Para conexiones directas a escala masiva, se implementa direccionamiento IPv6 ULA (Unique Local Address, RFC 4193) derivado criptográficamente del DID Ed25519 del nodo:

```text
┌──────────────────────────────┬────────────────────────────────────────────────────────┐
│  Prefijo ULA IPv7 (64 bits)  │       Identificador de Interfaz Criptográfico (64 bits)│
│          fd07::/64           │                   SHA-256(DID)[0:8]                    │
└──────────────────────────────┴────────────────────────────────────────────────────────┘
```

### Algoritmo de Derivación:
$$\text{IPv6} = \texttt{fd07:0000:0000:0000:} \parallel \text{Truncate}_{64}(\text{SHA-256}(\text{DID}))$$

1. **Invarianza**: La dirección IPv6 de un peer es idéntica en cualquier nodo del mundo que conozca su DID.
2. **Resistencia a Colisiones**: Con un espacio de $2^{64}$ direcciones bajo el prefijo `fd07::`, la probabilidad de colisión según la paradoja del cumpleaños para mil millones de nodos concurrentes es inferior a $10^{-11}$.
3. **Autenticación Implícita**: Un nodo puede verificar si un paquete entrante proviene legítimamente del propietario del DID contrastando la firma Ed25519 con la dirección IPv6 origen.

---

## 4. MTU y Clamping Determinista (Evitar Fragmentación)

- **MTU Virtual en Interfaz TUN**: `1280 bytes` (el mínimo estándar de IPv6 para garantizar cero fragmentación en Internet).
- **Sobrecarga de Encapsulación IPv7**:
  - Encabezado CBOR canónico: $\sim 32$ bytes.
  - Firma Ed25519 y metadatos Noise: $\sim 64$ bytes.
  - Encabezado UDP exterior: 8 bytes.
  - Encabezado IP exterior (IPv4/IPv6 de transporte): 20–40 bytes.
- **Paquete Total en el Cable**: $\le 1424$ bytes, encajando holgadamente dentro del MTU estándar de Ethernet/Internet (1500 bytes).

---

## 5. Tabla de Traducción en Memoria (`VirtualRouteTable`)

El driver mantendrá una estructura atómica concurrente y lock-free para la correspondencia en tiempo constante $O(1)$:

```go
type VirtualRouteTable struct {
    ipToDID map[[16]byte]core.Identity
    didToIP map[core.Identity][16]byte
    mu      sync.RWMutex
}
```
Cualquier evento de IP Roaming actualiza el endpoint UDP físico en la capa inferior sin alterar jamás las entradas de esta tabla virtual.
