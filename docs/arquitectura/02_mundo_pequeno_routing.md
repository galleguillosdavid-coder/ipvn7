# Modelo de Enrutamiento "Mundo Pequeño" (Small-World Routing) en IPv7

## 1. Fundamento Teórico: De los 6 a los 12 Grados de Separación

El concepto de **"Mundo Pequeño"** (*Small-World Experiment*), introducido originalmente por Stanley Milgram (1967) y formalizado matemáticamente por Duncan Watts, Steven Strogatz (1998) y Jon Kleinberg (2000), demostró que en redes complejas, dos entidades arbitrarias están conectadas por cadenas asombrosamente cortas de intermediarios.

En **IPv7**, transpolamos esta lógica al direccionamiento de dispositivos electrónicos en el planeta:
- En lugar de obligar a los routers a memorizar el mapa completo de Internet (como hace BGP con más de 950.000 rutas en memoria TCAM), cada nodo en IPv7 solo mantiene un **conjunto diminuto y acotado de conexiones estratégicas**.
- Con un factor de abanico (*fan-out*) de apenas **$k = 10$ dispositivos por nivel**:
  $$\text{Nodos alcanzables en } d \text{ saltos} = k^d = 10^d$$
  - En 1 salto: $10$ nodos
  - En 3 saltos: $1.000$ nodos
  - En 6 saltos: $1.000.000$ de nodos (1 millón)
  - En 9 saltos: $1.000.000.000$ de nodos (mil millones)
  - **En 12 saltos: $1.000.000.000.000$ de nodos (1 billón de dispositivos)**

Un límite de **12 grados de separación** cubre holgadamente todos los dispositivos electrónicos, sensores IoT, teléfonos y servidores existentes en la Tierra.

---

## 2. Arquitectura de la Tabla Acotada (`SmallWorldTable`)

Implementada en [`core/smallworld.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ip/core/smallworld.go):

```
+-------------------------------------------------------------+
|                      Nodo IPv7 Local                        |
|        Identidad Criptográfica: Clave Pública Ed25519       |
+-------------------------------------------------------------+
                              |
    +-------------------------+--------------------------+
    |                                                    |
[Anillo 0 - Directos]                              [Anillo 1..12 - Atajos]
Max 10 vecinos LAN / baja latencia                 Max 10 peers por grado logarítmico
```

### Características de la Tabla Acotada:
1. **Consumo de Memoria Minúsculo**:
   - $12 \text{ anillos} \times 10 \text{ contactos} = 120 \text{ entradas máx}$.
   - Cada entrada almacena clave pública (32 bytes), IP:puerto y latencia.
   - **Consumo total en RAM: ~10 Kilobytes**. Esto permite que un smartwatch o microcontrolador ESP32 funcione como un router IPv7 completo.
2. **Reemplazo por Latencia (Optimización Física)**:
   - Cuando un anillo de distancia se llena ($10$ contactos), si se descubre un nuevo peer para ese anillo con menor latencia (RTT), se expulsa al contacto más lento. Esto asegura que los atajos de larga distancia viajen por las rutas físicas más rápidas.

---

## 3. Métrica de Distancia y Enrutamiento Voraz (Greedy Routing)

Para determinar a qué grado de separación pertenece un nodo y cómo enviar paquetes hacia un destino desconocido:

1. **Métrica Criptográfica XOR**:
   $$\Delta(A, B) = \text{PublicKey}_A \oplus \text{PublicKey}_B$$
2. **Cálculo del Grado**:
   Se cuentan los ceros a la izquierda (*leading zeros*) del resultado XOR de 256 bits y se proyectan proporcionalmente en el rango $[0..12]$.
3. **Reenvío Voraz (Greedy Forwarding)**:
   - Cuando el Nodo $A$ desea enviar un mensaje a $C$, si no tiene una conexión IP directa con $C$, consulta su `SmallWorldTable`.
   - Selecciona el contacto $B$ cuya clave pública tenga la **menor distancia matemática a $C$**.
   - Envía el contenedor a $B$.
   - $B$ repite el proceso hasta que el paquete llega a su destino final.

---

## 4. Prevención de Bucles y Preservación de Firma Digital

Uno de los mayores retos al reenviar paquetes multi-salto es evitar bucles infinitos sin romper la firma digital del emisor original.

### Solución en IPv7 (`HopLimit = 12`):
En [`core/container.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ip/core/container.go):
```go
type Container struct {
    SenderPubKey   []byte `cbor:"1,keyasint"`
    ReceiverPubKey []byte `cbor:"2,keyasint,omitempty"`
    SessionID      string `cbor:"3,keyasint,omitempty"`
    Payload        []byte `cbor:"4,keyasint"`
    Signature      []byte `cbor:"5,keyasint,omitempty"`
    HopLimit       uint8  `cbor:"6,keyasint,omitempty"` // Máximo 12 grados
}
```

1. **Invariante Criptográfico**:
   Durante `.Sign()` y `.Verify()`, el campo `HopLimit` es temporalmente excluido del cálculo del hash canónico CBOR.
2. **Decremento en Tránsito**:
   - Cada nodo intermediario valida la firma del remitente original.
   - Si `HopLimit <= 1`, el paquete se descarta (evita tormentas de paquetes o bucles).
   - Si `HopLimit > 1`, se decrementa `HopLimit--` y se reenvía al siguiente salto.
   - **Resultado**: El receptor final recibe el paquete y valida exitosamente la firma del remitente original sin importar cuántos saltos haya dado.

---

## 5. Tabla Comparativa: IPv4/BGP vs. IPv7 Mundo Pequeño

| Característica | Internet Actual (IPv4/IPv6 BGP) | IPv7 (Mundo Pequeño) |
| :--- | :--- | :--- |
| **Tamaño de Tabla** | > 950.000 rutas (crece sin límite) | Acotada a **120 peers** máx. |
| **Memoria requerida** | Gigabytes en hardware especializado (TCAM) | **~10 KB** en RAM estándar |
| **Identidad** | Atada a la IP física (cambia con el Wi-Fi) | Criptográfica fija (**Ed25519**) |
| **Límite de Saltos** | TTL de 64 o 255 | Cota de **12 grados de separación** |
| **Hardware para Router** | Routers industriales dedicados | Cualquier PC, móvil o dispositivo IoT |
| **Enrutamiento sin IP Destino** | Imposible | **Posible mediante salto voraz por identidad** |

---

## 6. Pruebas Automatizadas Verificadas

En [`core/smallworld_test.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ip/core/smallworld_test.go):
- `TestSmallWorldTableBoundedCapacity`: Confirma que la tabla no excede el límite de capacidad acotada y descarta peers lentos.
- `TestSmallWorldMultiHopRelay`: Demuestra que el Nodo A puede enviar un mensaje al Nodo C a través del Nodo B sin que A conozca jamás la IP física de C.
- `TestSmallWorldHopLimitDrop`: Confirma que al agotarse el contador de 12 saltos, los paquetes en bucle son eliminados de la red de forma segura.
