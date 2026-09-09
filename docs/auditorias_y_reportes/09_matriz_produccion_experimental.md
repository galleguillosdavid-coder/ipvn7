# 09. Matriz de Validación Experimental de Producción (EPV) y Congelamiento del Core

**Fecha de Formalización:** 09 de Septiembre de 2026  
**Fase de Ciclo de Vida:** Transición a *Experimental Production Validation (EPV)*  
**Estado del Protocolo:** **CORE FROZEN (Congelado)**  
**Herramienta de Medición Oficial:** `tools/netbench/iperf_ipv7.go` (`bin/iperf_ipv7.exe` / `bin/iperf_ipv7_linux`)

---

## 1. Declaración Oficial de Congelamiento del Protocol Core

Con los resultados empíricos obtenidos en la *Fase de Perfeccionamiento Full*, el **Core de IPv7 queda técnicamente congelado**.

### Justificación Técnica del Congelamiento
El perfilamiento granular de capas demostró que el Core de IPv7 no representa un cuello de botella arquitectónico:
- **Serialización CBOR:** 1,0 % (~9,7 µs)
- **Enrutamiento Heurístico Small-World (XOR):** 0,6 % (~5,7 µs)
- **Overhead total intrínseco de IPv7:** **1,6 %**
- **Procesamiento Criptográfico (Ed25519 + ChaCha20-Poly1305 + Noise_XX):** **98,4 %**
- **Capacidad interna de transporte en sesión establecida (Warm/Bulk):** **1,16 Gbps por núcleo** (~145,39 MB/s, 87.900 pps)

> [!IMPORTANT]
> **Principio Rector:** A partir de este hito, **no se modificará el Core del protocolo** (formato de contenedor, esquemas CBOR, métricas XOR ni primitivas criptográficas base).  
> Cualquier futura modificación del protocolo deberá estar justificada de forma unívoca por un resultado experimental concreto obtenido en redes físicas de producción.

### Nuevo Flujo de Evolución de la Arquitectura
```
┌─────────────────────────┐
│   IPv7 Core Congelado   │  (Inmutable: CBOR, Small-World, Crypto, Contenedores)
└────────────┬────────────┘
             ▼
┌─────────────────────────┐
│       Adaptadores       │  (UDP, QUIC, DERP Relay, STUN/TURN, PMTU dinámico)
└────────────┬────────────┘
             ▼
┌─────────────────────────┐
│  Herramientas & Sockets │  (iperf_ipv7, Chaos Injector, Telemetría JSON)
└────────────┬────────────┘
             ▼
┌─────────────────────────┐
│  Redes Reales & Cargas  │  (LAN, WAN, NATs simétricos, pruebas de 1h)
└────────────┬────────────┘
             ▼
┌─────────────────────────┐
│    Evidencia Empírica   │  (Decisiones de diseño basadas 100% en datos)
└─────────────────────────┘
```

---

## 2. Rigor Científico: Demostración vs. Alcance Futuro

### Lo que queda formalmente demostrado
| Nivel de Evaluación | Resultado Obtenido | Qué Demuestra Técnicamente |
|---|---|---|
| **Pipeline Interno (Cold)** | ~1,4 ms / op | Coste máximo del establecimiento inicial de identidad criptográfica. |
| **Sesión Establecida (Warm)** | 11,37 µs / op (85,88 MB/s) | Coste marginal mínimo del transporte cifrado una vez establecida la sesión. |
| **Procesamiento Bulk 64 KB** | 1,163 Gbps / core | Capacidad pura de cómputo en memoria sin cuello de botella en el protocolo. |
| **PQC Híbrido Modular** | Overhead +2,0% (+6,29 µs) | Viabilidad de resistencia post-cuántica (ML-DSA / ML-KEM) sin penalizar throughput. |
| **Sockets UDP de SO Real** | **199,79 Mbps sostenidos** | IPv7 opera con éxito sobre la pila de red real del sistema operativo. |
| **Tasa de Transmisión Real** | 24.565 pps | Procesamiento continuo de ráfagas masivas sin ahogamiento de goroutines. |
| **Pérdida de Paquetes** | 0,0 % | En el escenario experimental evaluado con búferes estándar. |
| **Path MTU (PMTU)** | 1.280 B negociado | Adaptabilidad al suelo mínimo estándar IPv6 sin fragmentación IP a nivel kernel. |

### Rigor en la formulación de conclusiones
Para mantener la integridad científica del proyecto, se establece la siguiente directriz de reporte:

* **Afirmación NO admitida (Prematura):**  
  *«IPv7 alcanza 200 Mbps de rendimiento de red.»*
* **Afirmación Rigurosa Aprobada:**  
  *«En el escenario experimental probado, IPv7 sostuvo 199,79 Mbps durante 3 segundos mediante sockets UDP del sistema operativo, transmitiendo 74,93 MB sin pérdida observada.»*

### Variables Pendientes para la Validación Total
Para alcanzar conclusiones sobre el comportamiento WAN global, deben someterse a prueba las siguientes variables del entorno:
1. **Independencia física:** Host A $\rightarrow$ Host B en hardware físicamente desacoplado.
2. **Topología de enlace:** Ethernet cableado, Wi-Fi 6, Enlaces 4G/5G, WAN transcontinental.
3. **Condiciones de canal:** Latencia RTT real (10 ms - 250 ms), jitter fluctuante y pérdida de paquetes natural.
4. **Tráfico cruzado:** Competencia con TCP BBR/Cubic y flujos UDP concurrentes.
5. **Comportamiento NAT:** NAT cónico, NAT restringido por puerto y NAT simétrico.
6. **Duración:** Pruebas de estrés y remojo continuo de 1 minuto, 10 minutos y 60 minutos.
7. **Direct P2P vs Relay:** Comparativa de throughput y latencia vía P2P directo vs DERP Relay intermediado.

---

## 3. Matriz de Validación Experimental de Producción (EPV)

La siguiente matriz define los escenarios de ejecución obligatorios que guiarán la recolección de evidencia experimental:

```
                                     IPv7
                                      │
            ┌─────────────────────────┴─────────────────────────┐
            │                                                   │
         DIRECT                                               RELAY
      (P2P Nativo)                                        (DERP Server)
            │                                                   │
    ┌───────┼───────┐                                   ┌───────┼───────┐
    │       │       │                                   │       │       │
   LAN     WAN     NAT                                 LAN     WAN     NAT
    │       │       │                                   │       │       │
 ┌──┴──┐ ┌──┴──┐ ┌──┴──┐                             ┌──┴──┐ ┌──┴──┐ ┌──┴──┐
 │     │ │     │ │     │                             │     │ │     │ │     │
UDP  QUIC UDP QUIC UDP QUIC                         UDP  QUIC UDP QUIC UDP QUIC
```

### Tabla de Escenarios Experimentales a Ejecutar

| ID Escenario | Topología | Entorno de Red | Transporte | Duraciones Planificadas | Métricas Clave |
|---|---|---|---|---|---|
| `EXP-DIR-LAN-UDP` | Direct P2P | LAN Gigabit / Wi-Fi | UDP Sockets | 1 min, 10 min, 60 min | Mbps, PPS, CPU %, RAM MB, Loss % |
| `EXP-DIR-LAN-QUIC` | Direct P2P | LAN Gigabit / Wi-Fi | QUIC Stream | 1 min, 10 min, 60 min | Mbps, Stream Head-of-line, RTT |
| `EXP-DIR-WAN-UDP` | Direct P2P | WAN Pública (Dual Host) | UDP Sockets | 1 min, 10 min, 60 min | PMTU (1280-1500), Jitter, Pérdida % |
| `EXP-DIR-WAN-QUIC` | Direct P2P | WAN Pública (Dual Host) | QUIC Stream | 1 min, 10 min, 60 min | BBR/Cubic Congestion, Recovery RTT |
| `EXP-DIR-NAT-UDP` | Direct P2P | NAT Simétrico / STUN | UDP Hole-punch | 1 min, 10 min | Éxito de apertura, Keepalive, Mbps |
| `EXP-REL-LAN-UDP` | Relay DERP | LAN (Proxy Relay) | UDP Encapsulado | 1 min, 10 min | Overhead de reenvío, Latencia añadida |
| `EXP-REL-WAN-UDP` | Relay DERP | WAN (Relay en Cloud) | UDP Encapsulado | 1 min, 10 min, 60 min | Penalización Relay vs Direct, Throughput |
| `EXP-REL-WAN-QUIC`| Relay DERP | WAN (Relay en Cloud) | QUIC Multiplex | 1 min, 10 min, 60 min | Pérdida de paquetes inducida en nodo relay |

---

## 4. Registro Estandarizado de Telemetría

Para cada una de las celdas de la matriz, la herramienta `iperf_ipv7` (`-json`) recolectará automáticamente el siguiente vector dimensional de métricas:

$$\text{Vector de Telemetría} = \langle \text{Throughput (Mbps)}, \text{PPS}, \text{RTT (ms)}, \text{Jitter (ms)}, \text{Loss (\%)}, \text{CPU (\%)}, \text{RAM (MB)}, \text{PMTU (B)}, \text{Duración (s)} \rangle$$

### Esquema del Registro JSON
```json
{
  "scenario_id": "EXP-DIR-WAN-UDP",
  "mode": "client",
  "target": "203.0.113.42:9050",
  "duration_sec": 60.0,
  "packet_size_bytes": 1024,
  "bytes_total": 149800000,
  "mb_total": 142.86,
  "throughput_mb_s": 2.38,
  "throughput_mbps": 19.04,
  "packets_total": 146289,
  "pps": 2438.15,
  "packets_lost": 12,
  "loss_percentage": 0.0082,
  "jitter_ms": 1.42,
  "avg_transit_latency_ms": 28.6,
  "alloc_ram_mb": 4.15,
  "sys_ram_mb": 18.25,
  "num_goroutines": 8,
  "pmtu_bytes": 1400
}
```

---

## 5. Validación Empírica en Hardware Físico Real (Notebook $\leftrightarrow$ PC)

**Fecha de Ejecución:** 09 de Septiembre de 2026  
**Topología:** P2P Directo sobre Wi-Fi (`EXP-DIR-LAN-UDP`)  
**Hardware Host A:** PC de Escritorio (`192.168.1.198`, Windows 11 x64, 8 CPUs)  
**Hardware Host B:** Notebook Físico (`192.168.1.106`, Windows 11 x64, Hostname `Dvd`)  
**Mecanismo de Despliegue:** SSH / SFTP nativo automatizado.

---

### Método 1: Saturación de Socket UDP con Encapsulado E2EE (`iperf_ipv7`)

Ambas máquinas ejecutaron la prueba de saturación física bidireccional sobre la interfaz Wi-Fi real del hogar/oficina:

#### A. Flujo Emisor: Notebook $\rightarrow$ PC (Duración: 10 segundos)
```json
{
  "mode": "client",
  "target": "192.168.1.198:9050",
  "duration_sec": 10.00025,
  "packet_size_bytes": 1024,
  "bytes_total": 54084576,
  "mb_total": 51.58,
  "throughput_mb_s": 5.16,
  "throughput_mbps": 41.26,
  "packets_total": 50736,
  "pps": 5073.47,
  "loss_percentage": 0.0,
  "alloc_ram_mb": 1.68,
  "pmtu_bytes": 1280
}
```
*Telemetría en Servidor Receptor (PC):* Ráfagas sostenidas observadas en recepción de **42,39 a 46,35 Mbps** y **5.000 a 5.435 pps**, con solo 1,3 a 3,4 MB de RAM asignada.

#### B. Flujo Emisor: PC $\rightarrow$ Notebook (Duración: 5 segundos)
```json
{
  "mode": "client",
  "target": "192.168.1.106:9055",
  "duration_sec": 5.00789,
  "packet_size_bytes": 1024,
  "bytes_total": 58906094,
  "mb_total": 56.18,
  "throughput_mb_s": 11.22,
  "throughput_mbps": 89.74,
  "packets_total": 55259,
  "pps": 11034.38,
  "loss_percentage": 0.0,
  "alloc_ram_mb": 3.30,
  "pmtu_bytes": 1280
}
```

---

### Método 2: Nodos Completos P2P con Interfaz Web (`ipv7-node.exe`)

Se desplegaron e interconectaron simultáneamente dos instancias completas del protocolo IPv7 con sus respectivos servidores UI web y stacks de descubrimiento:
* **Nodo PC:** `cdb258869c278ebce7ff273a05541f7c1830888a9a32d0021498f17f7b3b4883` en puerto `7001` (UI `:8080`).
* **Nodo Notebook:** `3e99be42d004b3e002a8a56c0cdfbb6b0ae7e8fb95877db1e476473a8e4ecb63` en puerto `7001` (UI `:8080`).

```
[LAN Discovery] Peer detectado en 192.168.1.106:7001! Iniciando handshake automático...
[OK]   ¡Handshake completado con 192.168.1.106:7001! Peer ID: 3e99be42d004b3e0... (RTT: 3.9778ms)
```

**Hallazgos Clave de la Validación Física:**
1. **Latencia RTT Real:** **3,97 ms** en el handshake criptográfico Noise_XX directo sobre Wi-Fi.
2. **Peers Reconocidos Mutuamente:** Ambos nodos reflejan al otro en `/api/peers` con `Degree: 12` (12 anillos del Mundo Pequeño) y canal E2EE asegurado.
3. **Pérdida de Paquetes:** **0,0 %** observado a lo largo de más de 105.000 paquetes cifrados transmitidos.
4. **Throughput y Capacidad del Sistema:** El throughput observado quedó determinado por las condiciones del enlace Wi-Fi durante estas pruebas; no se observó evidencia de saturación del CPU, memoria o del Core IPv7 que limitara el flujo.
5. **Path MTU:** PMTU efectivo de 1.280 bytes, utilizado sin fragmentación en el escenario probado.

---

## 6. Estado Oficial del Proyecto y Veredicto Técnico

Con estos resultados empíricos, la etapa de **Validación Física LAN** queda formalmente **CERRADA**:

```
                    IPv7
                      │
                      ▼
           ┌─────────────────────┐
           │ Validación funcional│
           │       ✓ PASADA      │
           └──────────┬──────────┘
                      ▼
           ┌─────────────────────┐
           │Validación cripto/PFS│
           │       ✓ PASADA      │
           └──────────┬──────────┘
                      ▼
           ┌─────────────────────┐
           │ Benchmark interno   │
           │       ✓ PASADO      │
           └──────────┬──────────┘
                      ▼
           ┌─────────────────────┐
           │ Socket físico real  │
           │       ✓ PASADO      │
           └──────────┬──────────┘
                      ▼
           ┌─────────────────────┐
           │ Dos máquinas reales │
           │       ✓ PASADO      │
           └──────────┬──────────┘
                      ▼
           ┌─────────────────────┐
           │ P2P LAN automático  │
           │       ✓ PASADO      │
           └──────────┬──────────┘
                      ▼
                 NUEVA ETAPA
                      │
                      ▼
          SUSTAINED / WAN / CHAOS
           PRODUCTION VALIDATION
```

### Veredicto Técnico Aprobado:
> *«IPv7 ha demostrado funcionamiento extremo a extremo entre dos hosts físicos independientes sobre una red Wi-Fi real, incluyendo descubrimiento P2P, handshake criptográfico, establecimiento E2EE, transporte UDP sostenido, PMTU efectivo y operación sin pérdida observada en los escenarios ensayados.»*

---

## 7. Cambio de Paradigma: De Perfeccionar a Intentar Romper el Protocolo

A partir de este hito, la pregunta rectora de ingeniería deja de ser:  
*«¿Qué característica podemos agregarle a IPv7?»*  
y pasa a ser:  
*«¿Qué condiciones reales todavía pueden romper IPv7?»*

### Próximos Escenarios de Ataque y Resiliencia (Reporte 10):
1. **Tráfico Sostenido de Remojo:** 30 a 60 minutos ininterrumpidos para monitoreo térmico, buffer bloat y detección de fugas de memoria.
2. **Inyección de Pérdida Artificial:** Escenarios calibrados al 1%, 5%, 10% y 20% de descarte forzado.
3. **Jitter y Reordenamiento Masivo:** Fluctuaciones abruptas de latencia y entrega desordenada.
4. **Mutación de PMTU en Caliente:** Forzado de cambio de MTU a mitad de transferencia para evaluar recuperación sin caídas.
5. **Caída y Recuperación de Nodos:** Desconexión súbita de interfaz de red y reconexión automática en caliente.
6. **Experimento Nuclear: Direct P2P vs. Relay DERP:**
   ```
                    IPv7
                     │
           ┌─────────┴─────────┐
           │                   │
         DIRECT              RELAY
           │                   │
       throughput          throughput
       RTT                 RTT
       jitter              jitter
       loss                loss
       CPU                 CPU
   ```
   *Determinar cuantitativamente el costo real de la abstracción cuando el camino P2P directo no está disponible.*


