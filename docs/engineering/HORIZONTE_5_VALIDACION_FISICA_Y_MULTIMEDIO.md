# HORIZONTE 5: VALIDACIÓN FÍSICA Y MULTIMEDIO
## Protocol Hardening, Chaos Testing y Demostración de Invariancia Epistémica

> **ESTADO DEL PROYECTO**: Desarrollo de nuevas características **CONGELADO**. Núcleo de protocolo [`core/`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core) bajo **Core Freeze estricto**.  
> **OBJETIVO PRIMARIO**: Someter la arquitectura modular (`adapters/`, `dht/`) a escenarios hostiles, particiones de red y transiciones físicas para clasificar empíricamente cada afirmación técnica en la escala de certeza rigurosa.

---

## 1. Declaración Epistémica de Certeza

Para evitar la complacencia técnica y separar tajantemente lo *"implementado y probado localmente"* de lo *"demostrado en condiciones reales"*, toda afirmación sobre IPv7 se califica según la siguiente taxonomía:

```text
┌─────────────────┬────────────────────────────────────────────────────────────────────────┐
│ NIVEL           │ CRITERIO OPERATIVO                                                     │
├─────────────────┼────────────────────────────────────────────────────────────────────────┤
│ DEMONSTRATED    │ Verificado empíricamente con mediciones reproducibles y datos reales.  │
│ OBSERVED        │ Observado en simulaciones de laboratorio local bajo condiciones dadas.  │
│ INFERRED        │ Deducido matemáticamente del diseño criptográfico o algorítmico.       │
│ HYPOTHESIS      │ Postulado arquitectónico razonable pero pendiente de instrumentación.   │
│ NOT_PROVEN      │ Sin evidencia experimental de campo; no debe asumirse funcional.       │
└─────────────────┴────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Matriz de Estado Epistémico de los Cuatro Horizontes

| Componente / Propiedad | Afirmación Técnica | Estado Epistémico | Evidencia / Justificación | Limitación Pendiente |
| :--- | :--- | :---: | :--- | :--- |
| **TUN/TAP: Derivación ULA** | `fd07::/64` e `10.7.0.0/16` se derivan deterministicamente del DID Ed25519. | **DEMONSTRATED** | Verificado en microbenchmarks (`12.36 ns/op`, `0 allocs`) y tests unitarios. | Colisión IPv4 en redes mayores a 65k nodos locales (`/16`). |
| **TUN/TAP: Soak Local** | Procesa 10.000 paquetes TCP sintéticos a través de memoria virtual sin pérdidas. | **DEMONSTRATED** | 100% PDR verificado en `TestTunAdapterSoak10kPackets` (1.23s). | Tráfico en RAM mock; no atraviesa el driver NDIS/Wintun del kernel. |
| **TUN/TAP: Driver de Kernel Real** | Creación e inyección de paquetes reales vía `/dev/net/tun` en Linux o Wintun en Windows. | **OBSERVED** | Driver funcional probado manualmente en Linux; NDIS en Windows requiere privilegios admin. | No demostrado en despliegues distribuidos sin privilegios elevados. |
| **DHT: Métrica XOR & Buckets** | Búsqueda y partición métrica en 256 k-buckets con $k=20$. | **DEMONSTRATED** | Verificado en `TestPureKademliaTableKBuckets` y microbenchmarks (`30.97 ns/op`). | Probado con topología sintética en memoria; pendiente escalabilidad a > 10.000 nodos. |
| **DHT: PoW Anti-Sybil** | Requisito computacional previene spam de identidades forjadas. | **DEMONSTRATED** | Verificado en `TestAntiSybilProofOfWork` (dificultad configurable 12..24 bits). | Un atacante con ASICs dedicados podría superar dificultades bajas sin balance dinámico. |
| **DHT: Red Soberana sin Firebase** | Resolución P2P descentralizada directa entre nodos sin servicio en la nube. | **OBSERVED** | Verificado en `TestDHTDiscoveryAdapterSovereign` entre nodos locales. | Asume que los nodos conocen al menos un bootstrap peer inicial alcanzable. |
| **Onion: Cifrado en Cascada** | Generación de secretos compartidos efímeros X25519 en capas concéntricas. | **DEMONSTRATED** | Verificado en `TestSphinxPacketBuildAndUnwrapThreeHops` ($A \to B \to C \to D$). | Costo de CPU de 3 intercambios ECDH (~0.78 ms por paquete construido). |
| **Onion: Tamaño Fijo Invariante** | Paquete exactamente fijado en 1280 bytes en cualquier salto para frustrar análisis de tráfico. | **DEMONSTRATED** | Verificado en `TestSphinxFixedPacketSizePadding` (`len(raw) == 1280`). | Sobrecarga de ancho de banda del ~70% en mensajes cortos (textos de 50 bytes viajan como 1280 B). |
| **Onion: Resistencia a Correlación** | Un adversario global pasivo no puede correlacionar flujos por tamaño ni temporización. | **INFERRED** | El tamaño fijo frustra análisis por longitud; falta inserción de retardo estocástico (chaffing). | No probado contra análisis de correlación temporal con tráfico masivo de Internet. |
| **Off-Grid: Baliza L2 (`OG7!`)** | Descubrimiento local directo en capa 2 mediante broadcast de baliza liviana. | **OBSERVED** | Verificado en `TestBeaconEngine_Discovery` sobre medio simulado `RadioMedium`. | **Crítico**: Debe quedar estrictamente confinado al medio físico local; no propagar al overlay. |
| **Off-Grid: Conmutación Híbrida** | Conmutación automática WAN $\leftrightarrow$ Off-Grid al detectar caída de Internet. | **OBSERVED** | Verificado en `TestHybridSwitcher_FailoverAndRouting` con callbacks asíncronos. | Simulado en laboratorio; pendiente prueba cortando interfaz Ethernet física real. |
| **Off-Grid: Hardware Real (LoRa/Wi-Fi)** | Transmisión de tramas sobre interfaces de radio reales (chips SX1262 o Wi-Fi Direct). | **NOT_PROVEN** | La capa `PhysicalLink` está desacoplada, pero no ha sido enlazada a drivers seriales/SPI reales. | Requiere hardware físico, adaptadores USB-LoRa y pruebas de alcance exterior. |
| **Canary Soak: Estabilidad Continua** | Nodo procesa > 160.000 paquetes durante > 7.5 horas con memoria plana (0.36 MB). | **OBSERVED** | Daemon `task-2137`: 160.300 pkts enviados, 159.384 recibidos (99.43% PDR), 8 goroutines. | **No es prueba matemática de cero fugas**; los 916 paquetes descartados requieren auditoría. |
| **Escala Global de Internet** | IPv7 puede reemplazar la infraestructura de enrutamiento global BGP/IP de Internet. | **NOT_PROVEN** | Es un horizonte estratégico aspiracional; no hay validación más allá de pruebas de laboratorio. | Requiere millones de nodos, tolerancia a particiones continentales y soberanía BGP. |

---

## 3. Precisiones Técnicas Inmediatas

### 3.1. Aislamiento Estricto del Broadcast L2 (`BeaconEngine`)
> [!WARNING]
> **REGLA DE CONFINAMIENTO FÍSICO**:  
> El broadcast utilizado por `BeaconEngine` (`adapters/offgrid/adhoc_mesh.go`) mediante el encabezado `OG7!` es **exclusiva y estrictamente de Capa 2 local** (medio de radio ad-hoc adyacente).  
> **BAJO NINGUNA CIRCUNSTANCIA** se debe permitir que:
> 1. Las balizas físicas L2 se inyecten en el overlay Small-World o en túneles IP globales.
> 2. El broadcast físico sustituya la DHT Kademlia o las consultas voraces de Kleinberg en el nivel de red.
> 3. El nodo emita paquetes broadcast hacia interfaces WAN o Internet tradicional.

### 3.2. Clasificación de Pérdidas del Soak Test (99.43% PDR)
En el ciclo continuo del daemon canario (`task-2137`), de 160.300 paquetes enviados se registraron 159.384 recibidos (**916 paquetes no recibidos**, 0.57% de descarte):
- **Causa primaria**: Descarte de búfer UDP en la pila de red del sistema operativo durante ráfagas de 350 paquetes/segundo cuando el hilo del runner cede tiempo de CPU a la indexación en background.
- **Ausencia de fugas**: La cota de heap plana en 0.36 MB y 8 goroutines constantes es **evidencia empírica sólida de ausencia de acumulación descontrolada**, pero no constituye prueba matemática formal. En redes reales con jitter, MTU variable y congestión, el PDR fluctuará naturalmente.

---

## 4. La Pregunta Central: Invarianza de Identidad Multimedio

```text
                   INTERNET WAN (Fibra / 4G / 5G)
                               │
                          ┌────▼────┐
                          │  Nodo   │
                          │   DID   │ (did:ipv7:02a8f...)
                          └────┬────┘
                               │
                      ┌────────┴────────┐
                      ▼                 ▼
                 [WAN Socket]     [Radio Ad-Hoc]
                 UDP / QUIC       Wi-Fi Direct / LoRa
                      │                 │
                      └────────┬────────┘
                               │
                DID y Sesión Criptográfica INVARIANTES
```

### Hipótesis a Demostrar o Refutar (H-MULTI-01):
> *"Un flujo de datos cifrado de extremo a extremo entre el Nodo $A$ y el Nodo $B$ puede sobrevivir a la desconexión total de la infraestructura WAN conmutando a enlaces de radio física locales ad-hoc, sin que los DIDs cambien, sin renegociar claves maestras y sin reiniciar las aplicaciones de capa superior."*

### Topología de Prueba Mixta (Batería de Experimentos):
1. **Escenario Lineal Mixto**:
   $$A \xrightarrow{\text{Internet (WAN)}} B \xrightarrow{\text{Off-Grid (Radio)}} C \xrightarrow{\text{Off-Grid (Radio)}} D$$
   - Nodo $B$ actúa como puente híbrido frontera.
   - $A$ se comunica con $D$ a través de $B$ y $C$.
2. **Escenario de Black-out y Conmutación**:
   $$A \xrightarrow{\text{WAN}} B \quad \Longrightarrow \quad [\text{Corte WAN}] \quad \Longrightarrow \quad A \xrightarrow{\text{Off-Grid}} C \xrightarrow{\text{Off-Grid}} B$$
   - Identidades $A$ y $B$ deben permanecer idénticas.
   - La tabla de rutas debe actualizar el próximo salto sin destruir la sesión de transporte virtual.

---

## 5. Batería de Pruebas Destructivas (Chaos Engineering)

Para la siguiente etapa de validación, el objetivo no es escribir código nuevo, sino diseñar tests rigurosos que intenten **romper** los supuestos de la arquitectura:

1. **Test de Mutación Física Abrupta**:
   Cortar el socket de red en medio de una transferencia de 10 MB y forzar la entrega por paquetes de 200 bytes en el canal de radio físico. Medir retransmisiones y corrupción.
2. **Test de Envenenamiento de Balizas L2**:
   Inundar el `BeaconEngine` con 10.000 balizas `OG7!` por segundo con DIDs falsificados y medir si satura la CPU o degrada la tabla de vecinos.
3. **Test de Colisión de Red de Kleinberg vs Partición Física**:
   Separar físicamente 5 nodos en dos islas aisladas, permitir que cada isla cree su propia DHT local, volver a unirlas y verificar si la convergencia XOR produce bloqueos mutuos o bucles de enrutamiento.
4. **Test de Desvanecimiento y RSSI Extremo**:
   Simular atenuación de señal de radio (-115 dBm con 40% de pérdida estocástica) y verificar si el conmutador híbrido entra en oscilación rápida (flapping) entre WAN y Off-Grid.

---

## 6. Conclusión de la Fase Actual

- **Código congelado**: Ninguna nueva funcionalidad será introducida hasta que las hipótesis de la Matriz de Certeza hayan sido sometidas a pruebas de rotura.
- **Rumbo**: Avanzar hacia experimentos reproducibles que eleven el nivel de certeza de las afirmaciones clave de `OBSERVED` e `INFERRED` a `DEMONSTRATED`, o documentar honestamente sus límites de fallo.
