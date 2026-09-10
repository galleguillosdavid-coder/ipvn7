# IPv7 → ipvn7: Análisis Integral y Hoja de Ruta hacia el Laboratorio Vivo

> **Documento Fundacional de Transición**  
> Fecha conceptual: post-`v0.5.0-chaos.audit`  
> Naturaleza: arquitectónico, epistémico y operativo.  
> Regla rectora: *La evidencia experimental tiene prioridad sobre cualquier suposición de diseño.*

---

## PARTE I — ANÁLISIS SINTÉTICO INTEGRAL DEL ESTADO ACTUAL

### 1.1 Mapa de Capacidades Demostradas

Tras cinco horizontes y una auditoría de segundo orden, IPv7 ha alcanzado un estado que pocas redes experimentales logran: **una cadena metodológica cerrada y auditable por terceros**. La síntesis rigurosa es la siguiente:

| Dimensión | Estado Real | Estado Epistémico | Naturaleza del Logro |
| :--- | :--- | :---: | :--- |
| **Identidad** | Ed25519 como DID primario, IP como endpoint secundario | `DEMONSTRATED` | Separación limpia identidad/transporte |
| **Transporte** | UDP fast-path + TUN/TAP universal (`fd07::/64`, 10.7.0.0/16) | `DEMONSTRATED` (`LAB_SIMULATED`) | Cualquier app transpone |
| **Descubrimiento** | DHT Kademlia pura 256 bits, PoW anti-Sybil, sin Firebase | `DEMONSTRATED` (`LAB_SIMULATED`) | Soberanía sin terceros |
| **Privacidad** | Sphinx 3 saltos, 1280 B fijos, ECDH cascada X25519 | `DEMONSTRATED` (`LAB_SIMULATED`) | Aislamiento multi-salto |
| **Resiliencia** | Conmutador híbrido WAN↔Off-Grid (`OG7!` L2) | `DEMONSTRATED` (`LAB_SIMULATED`) | Independencia de medio |
| **Ingeniería** | `core/` sellado (0 líneas modificadas en 5 horizontes) | `DEMONSTRATED` | Frontera inviolable intacta |
| **Reproducibilidad** | Scripts + GitHub Actions + checksums SHA-256 | `DEMONSTRATED` | Auditoría de terceros cerrada |

### 1.2 Deuda Epistémica Residual

Toda la batería destructiva de Horizonte 5 —`H-MULTI-01`, `H-L2-DOS`, `H-FLAPPING`, `H-SPLIT-BRAIN`, `H-CONSTRAINED-MTU`, `H-UNIFIED-E2E`— está sellada bajo **`LAB_SIMULATED`**. Esta palabra no es cosmética:

- **`LAB_SIMULATED`** significa: *in-process, sin pérdida física, sin NAT real, sin jitter WiFi, sin duty cycle, sin clock drift, sin reordering producido por conmutadores reales.*  
- **La distancia entre `LAB_SIMULATED` y `REAL_NETWORK`/`REAL_HARDWARE` es la deuda epistémica fundamental del proyecto.**

La auditoría ya lo declaró con precisión quirúrgica:

> *"La cadena documentación → contrato → test → assertion → resultado es reproducible **bajo el entorno y las condiciones de prueba declaradas**."*

Ese *bajo* es la frontera que los Horizontes 6–9 deben cruzar.

### 1.3 Los Tres Invariantes que Sobreviven

Independientemente del entorno, tres propiedades han demostrado ser **estructurales** y no accidentales:

```text
┌─────────────────────────────────────────────────────────────┐
│  INVARIANTE I   —   SEPARACIÓN IDENTIDAD / ENDPOINT / MEDIO │
│  El DID no muta cuando cambia IP, router, WiFi o radio.     │
├─────────────────────────────────────────────────────────────┤
│  INVARIANTE II  —   ADAPTADORES, NO NÚCLEO                  │
│  Toda capacidad nueva vive fuera de core/. Toda.            │
├─────────────────────────────────────────────────────────────┤
│  INVARIANTE III —   MEDIR ANTES DE PROGRAMAR                │
│  Cero optimización sin baseline. Cero hipótesis sin test.   │
└─────────────────────────────────────────────────────────────┘
```

### 1.4 El Salto Pendiente

```text
                      DEUDA EPISTÉMICA ABIERTA
  ┌──────────────────────────────────────────────────────────┐
  │  LAB_SIMULATED  ─────►  REAL_NETWORK  ─────►  FIELD      │
  │    DEMOSTRADO            OBSERVABLE          OBSERVADO   │
  │    hoy                   H6-H7                H7+        │
  └──────────────────────────────────────────────────────────┘
```

Todo lo demás —Kùzu, telemetría, ingeniería asistida— son **medios para reducir esa deuda**, no fines.

---

## PARTE II — RACIONAL DEL NUEVO REPOSITORIO `ipvn7`

### 2.1 Por qué un nuevo repositorio y no una rama

| Razón | Justificación Ingenieril |
| :--- | :--- |
| **Sello histórico** | `Ipv7@v0.5.0-chaos.audit` debe permanecer inmutable como artefacto auditable y reproducible bit a bit. |
| **Cambio de paradigma** | De "prototipo en laboratorio" a "red viva operando en Internet público y hardware físico". |
| **Divergencia de dependencias** | Hardware real (LoRa, BLE, Wi-Fi Direct), Kùzu en caliente, telemetría continua y orquestación multi-host requieren árbol de dependencias distinto. |
| **Separación epistémica** | Las aserciones de `Ipv7` son `LAB_SIMULATED`. Las de `ipvn7` aspiran a `REAL_NETWORK`/`REAL_HARDWARE`. No deben mezclarse en un mismo corpus. |

### 2.2 Contrato de Continuidad Epistémica

`ipvn7` **hereda sin negociación** los tres invariantes y las siguientes cláusulas vinculantes:

1. **`core/` permanece sellado byte a byte** como en `Ipv7@v0.5.0`. Cualquier cambio en `core/` requiere hipótesis falsada demostrada + aprobación de David.
2. **Todo artefacto de `Ipv7` es importable como caja negra auditable**: adaptadores TUN, DHT, Sphinx, off-grid y contratos de aserción viajan intactos.
3. **Ninguna afirmación de `ipvn7` puede heredar el estado `DEMONSTRATED` de `Ipv7`**: deberá **re-demostrarse** en el nuevo entorno o degradarse explícitamente a `OBSERVED`/`INFERRED`.
4. **Regla de no-duplicación**: si algo ya existe en `Ipv7`, se reutiliza; si no encaja, se justifica en `BOTTLENECKS.md`.

### 2.3 Nomenklatura: por qué `ipvn7`

- `Ipv7` = protocolo experimental en laboratorio.  
- `ipvn7` = **N**ext-**N**etwork: el mismo protocolo, **vivo**, en campo y hardware real.  
- El cambio de nombre evita contaminación semántica de las aserciones históricas.

---

## PARTE III — ARQUITECTURA DEL REPOSITORIO `ipvn7`

### 3.1 Estructura de Directorios

```text
ipvn7/
├── core/                          # FROZEN — importado tal cual de Ipv7@v0.5.0
│                                  #   (verificado con checksum SHA-256 en CI)
├── adapters/
│   ├── tun/                       # heredado
│   ├── onion/                     # heredado
│   └── offgrid/                   # heredado + expansión H9
│       ├── virtual/               # (legacy LAB_SIMULATED)
│       ├── wifi_direct/           # H9
│       ├── ble/                   # H9
│       ├── lora/                  # H9
│       └── sdr/                   # H9 — experimental
├── dht/                           # heredado + observabilidad reforzada
├── telemetry/                     # H8 — fundación
│   ├── events/                    # estructuras canónicas
│   ├── metrics/                   # counters/gauges/histograms
│   ├── buffer/                    # bounded, lock-low
│   ├── aggregator/
│   └── exporters/
│       ├── prometheus/
│       └── kuzu/                  # feed estructurado (no crítico)
├── engineer/                      # H8 — ipv7-engineer
│   ├── experiments/
│   ├── baselines/
│   ├── anomalies/
│   ├── bottlenecks/
│  ├── regression/
│   └── reports/
├── tools/
│   ├── soak-runner/               # H6 — 7 días continuos
│   ├── chaos-injector/            # H6/H7 — seguro y limitado
│   ├── topology-orchestrator/     # H7 — multi-host heterogéneo
│   └── hardware-bench/            # H9 — LoRa/BLE/WiFiDirect
├── experiments/                   # corpus reproducible (fuente de verdad)
│   ├── exp-0001-roaming/
│   ├── exp-0002-socket-contention/
│   └── ...
├── docs/
│   ├── engineering/               # EXPERIMENTS, BASELINES, BOTTLENECKS, ...
│   ├── telemetry/
│   └── horizons/                  # un dossier por horizonte
└── .github/workflows/             # CI reproducibilidad + integración hardware
```

### 3.2 Capas de Responsabilidad

```text
┌─────────────────────────────────────────────────────────────────┐
│  L4  DECISIÓN Y GOBERNANZA                                      │
│      David (autoridad) · ChatGPT (auditor externo)              │
├─────────────────────────────────────────────────────────────────┤
│  L3  INGENIERÍA ASISTIDA  (engineer/)                           │
│      Hipótesis · Experimentos · Regresión · Informes            │
├─────────────────────────────────────────────────────────────────┤
│  L2  OBSERVABILIDAD  (telemetry/)                               │
│      Eventos · Métricas · Buffer acotado · Exportadores         │
├─────────────────────────────────────────────────────────────────┤
│  L1  ADAPTADORES  (adapters/, dht/)                             │
│      TUN · Sphinx · Off-grid · LoRa · BLE · Wi-Fi Direct        │
├─────────────────────────────────────────────────────────────────┤
│  L0  CORE CONGELADO  (core/)                                    │
│      Identidad · Crypto · Wire format · Sesión · Handshake      │
└─────────────────────────────────────────────────────────────────┘
              ↓ No cruza hacia arriba en el crítico.
```

**Regla de flujo:** los paquetes fluyen por `L0↔L1`. La observabilidad (`L2`) **observa en paralelo**, nunca en serie. El ingeniero (`L3`) **lee** telemetría; nunca escribe al camino crítico.

### 3.3 Fronteras Inviolables (reafirmadas)

```text
┌───────────────────────────────────────────────────────────────┐
│  JAMÁS en el camino crítico:                                  │
│    · Kùzu                                                     │
│    · Prometheus (scrape passive, no push blocking)            │
│    · logging estructurado síncrono                            │
│    · serialización a disco                                    │
│    · cualquier I/O de red externa                             │
└───────────────────────────────────────────────────────────────┘
```

---

## PARTE IV — HORIZONTE 6: VALIDACIÓN CONTINUA 7 DÍAS Y RESILIENCIA EN RED REAL

### 6.0 Pregunta Central

> ¿IPv7 sobrevive **7 días continuos** de operación en una red real (Wi-Fi doméstica, NAT, roaming entre routers, ISP público) sin fugas de memoria, sin deriva de latencia, sin degradación de PDR, sin crecimiento de goroutines, conservando el DID?

### 6.1 Transición Epistémica

```text
LAB_SIMULATED  ─────────────►  REAL_NETWORK
{DEMONSTRATED}                 {OBSERVED → DEMONSTRATED}
```

**Nada se hereda silenciosamente.** Cada test de `Ipv7@v0.5.0` se **re-ejecuta** en `REAL_NETWORK` y se re-clasifica.

### 6.2 Arquitectura del Horizonte 6

```text
┌──────────────────────────────────────────────────────────────┐
│                   SOAK RUNNER (7 DÍAS)                       │
│  ┌────────────────────┐   ┌─────────────────────────────┐    │
│  │ Traffic Generator  │   │ Chaos Injector (BOUNDED)    │    │
│  │  (sintético+real)  │   │  · IP change cada N horas   │    │
│  └─────────┬──────────┘   │  · WAN blip controlado      │    │
│            │              │  · WiFi re-assoc             │    │
│            ▼              └──────────────┬──────────────┘    │
│  ┌─────────────────────────┐             │                   │
│  │  IPv7 NODE (real)       │◄────────────┘                   │
│  └───────────┬─────────────┘                                 │
│              │ Telemetry (bounded)                           │
│              ▼                                               │
│  ┌─────────────────────────┐                                 │
│  │  Telemetry Buffer       │──► Prometheus (passive)         │
│  │  (bounded, 0 drop al    │──► Kùzu (off-critical, batch)   │
│  │   camino crítico)       │──► SOAK_REPORT (JSON+dash)      │
│  └─────────────────────────┘                                 │
└──────────────────────────────────────────────────────────────┘
```

### 6.3 Sub-fases y Etapas

#### Sub-fase 6.1: Telemetría Mínima Viable
- [ ] Implementar `telemetry/events/` con `TelemetryEvent` canónico (campos opcionales genuinamente opcionales).
- [ ] Buffer acotado `lock-low` con política `DROP_TELEMETRY` y contador `telemetry_dropped_total`.
- [ ] Verificar **cero bloqueo** del path crítico mediante `BenchmarkTelemetryPathCritical` (objetivo: Δlatencia p99 < umbral medido, no supuesto).

#### Sub-fase 6.2: Soak Runner de 7 Días
- [ ] Loop de carga sintética estable (pps constante + ráfagas).
- [ ] Registro por ventana temporal: PDR, RTT (min/p50/p95/p99/max), jitter, PMTU, goroutines, heap.
- [ ] Cálculo de **slope** (regresión lineal sobre ventanas) para heap y goroutines.
- [ ] Rotación de logs por tamaño y tiempo (nunca por evento).
- [ ] Checkpoint cada 6 h: snapshot de estado reanudable.

#### Sub-fase 6.3: Resiliencia en Red Real
- [ ] **IP Roaming WiFi A → WiFi B** — registrar `ENDPOINT_CHANGED` con `old_endpoint`, `new_endpoint`, `recovery_time`, `RTT_before`, `RTT_after`, `PMTU_before`, `PMTU_after`, `route_before`, `route_after`, `direct_or_relay`.
- [ ] **Roaming hotspot → router doméstico → ISP distinto**.
- [ ] **Pérdida WiFi estocástica** (no simulada: real vía apagado/encendido AP controlado).
- [ ] **Traversal NAT real** (NAT simétrico, NAT cónico, CG-NAT).
- [ ] **Relay de respaldo** activado/desactivado por corte de puerto.

#### Sub-fase 6.4: Detección Temprana de Derivas
- [ ] Detector de `MEMORY_GROWTH_SUSPECTED` (slope positivo sostenido > umbral calibrado).
- [ ] Detector de `GOROUTINE_GROWTH_SUSPECTED`.
- [ ] Detector de `LATENCY_DRIFT` (p99 retardado respecto a baseline).
- [ ] Detector de `PDR_DEGRADATION` (umbral configurable, justificado por baseline).

### 6.4 Batería de Validación (Horizonte 6)

| ID | Hipótesis | Entorno | Cota Contractual | Aserción |
| :--- | :--- | :---: | :--- | :--- |
| `H6-SOAK-7D` | Heap estable tras 7 días | `REAL_NETWORK` | slope ≤ ε (calibrado) | reporte `NO LEAK-COMPATIBLE GROWTH OBSERVED UNDER TEST CONDITIONS` |
| `H6-GOROUTINE-STEADY` | Goroutines planas tras 7 días | `REAL_NETWORK` | Δ ≤ 2 sostenido | `runtime.NumGoroutine()` en ventana final |
| `H6-PDR-STEADY` | PDR ≥ objetivo tras 7 días | `REAL_NETWORK` | ≥ 99% (a confirmar) | conteo real de paquetes |
| `H6-ROAMING-REAL` | DID invariante en roaming real | `REAL_NETWORK` | bytes.Equal | aserción de identidad |
| `H6-NAT-TRAVERSAL` | Sesión E2EE sobrevive NAT simétrico | `REAL_NETWORK` | err == nil | decrypt post-NAT |
| `H6-TELEMETRY-NONBLOCKING` | Telemetría no degrada path | `REAL_NETWORK` | Δp99 < umbral medido | benchmark A/B |

### 6.5 Cláusulas de Limitación Explícitas

1. **NO se afirma** ausencia absoluta de leaks. Se afirma `NO LEAK-COMPATIBLE GROWTH OBSERVED` bajo condiciones declaradas.
2. **NO se afirma** universalidad del PDR. Cada entorno tiene su propio baseline.
3. **NO se afirma** que los detectores de deriva no tengan falsos positivos; se mide la tasa `FP_rate` y se documenta.

---

## PARTE V — HORIZONTE 7: TOPOLOGÍA HETEROGÉNEA MULTI-HOST EN INTERNET PÚBLICO

### 7.0 Pregunta Central

> ¿La topología de mundo pequeño tipo Kleinberg que `Ipv7` modela se sostiene empíricamente cuando se despliega sobre N hosts heterogéneos (Windows, Linux, ARM, macOS) repartidos geográficamente, con ISPs, NATs y stacks de red distintos?

### 7.1 Transición Epistémica

```text
REAL_NETWORK (single-host)  ─────────►  REAL_NETWORK (multi-host geográfico)
{OBSERVED}                              {OBSERVED → DEMONSTRATED}
```

### 7.2 Arquitectura de Orquestación

```text
                       ┌──────────────────────────────┐
                       │  ORQUESTADOR DE TOPOLOGÍA    │
                       │  (declarativo YAML)          │
                       └──────────────┬───────────────┘
                                      │
        ┌────────────┬────────────┬───┴────────┬────────────┐
        ▼            ▼            ▼            ▼            ▼
   ┌─────────┐ ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐
   │ Win x64 │ │ Linux   │  │ ARM64   │  │ macOS   │  │ Linux   │
   │ ISP A   │ │ AMD64   │  │ RPi4    │  │ ISP D   │  │ VPS E   │
   │ NAT-C   │ │ ISP B   │  │ 4G      │  │ NAT-P   │  │ IPv6    │
   └─────────┘ └─────────┘  └─────────┘  └─────────┘  └─────────┘
        │            │            │            │            │
        └────────────┴───── DHT + Sphinx ─────┴────────────┘
                          Malla IPv7 viva
```

### 7.3 Sub-fases y Etapas

#### Sub-fase 7.1: Orquestador Declarativo Multi-Host
- [ ] Especificación YAML: `{host, os, arch, isp, nat_type, geography, role}`.
- [ ] Provisioning reproducible (por SSH o agentes ya existentes; **no inventar cloud ficticio**).
- [ ] Si falta infraestructura real: marcar `BLOCKED` y no simular resultados.

#### Sub-fase 7.2: Convergencia DHT a Escala
- [ ] Medir latencia de lookup DHT: min/p50/p95/p99/max.
- [ ] Medir convergencia tras `peer_join` masivo (partición de tabla).
- [ ] Test de `eclipse attack` benigno: cuántos IDs maliciosos puede absorber la tabla antes de degradar lookup.
- [ ] Verificar inviolabilidad de firmas Ed25519 en **cada entrada DHT** recibida de red real.

#### Sub-fase 7.3: Diversidad de Caminos Sphinx
- [ ] Construir circuitos de 3 saltos **con restricción geográfica/ISP** (diversidad real, no solo XOR).
- [ ] Medir entropía efectiva de la selección de saltos vs. topología teórica.
- [ ] Verificar 1280 B fijos en tránsito real (MTU y path MTU efectivo).
- [ ] Test `Sphinx under loss`: con pérdida real, ¿qué % de circuitos completan?

#### Sub-fase 7.4: Resistencia a Churn Real
- [ ] Ciclos de join/leave controlados y medidos (no simulados).
- [ ] Cobertura DHT post-churn: ¿la búsqueda sigue resolviendo tras perder 30% de nodos?
- [ ] Reconstrucción de circuitos tras fallo de hop intermedio.

#### Sub-fase 7.5: Comportamiento Cross-ISP y Cross-NAT
- [ ] Matriz de compatibilidad NAT (cónico-simétrico, simétrico-simétrico, CG-NAT).
- [ ] Fallback a relay cuando traversal directo falla.
- [ ] Medir **costo real** de relay: latencia, throughput, jitter adicional.

### 7.4 Batería de Validación (Horizonte 7)

| ID | Hipótesis | Métrica | Cota | Estado Esperado |
| :--- | :--- | :--- | :--- | :---: |
| `H7-DHT-CONVERGENCE` | DHT converge en N hosts reales | tiempo de lookup p99 | umbral medido | `OBSERVED` |
| `H7-SPHINX-DIVERSITY` | Circuitos atraviesan ISPs distintos | % circuitos diversos | ≥ 90% | `OBSERVED` |
| `H7-CHURN-TOLERANCE` | DHT resuelve con 30% churn | PDR de lookups | ≥ 95% | `OBSERVED` |
| `H7-NAT-MATRIX` | Combinaciones NAT documentadas | tasa de éxito directo | documentada | `OBSERVED` |
| `H7-CROSS-OS` | Windows↔Linux↔ARM interoperan | 100% E2E | == 100% | `OBSERVED` |
| `H7-GEO-LATENCY` | Latencia geográfica documentada | p50, p95, p99 | no cota, dato | `OBSERVED` |

### 7.5 Riesgos y Cláusulas

1. **NO se afirma** comportamiento universal en todos los ISPs del mundo; se documenta la matriz real probada.
2. **NO se afirma** resistencia a actores estatales o DPI avanzado: fuera de alcance declarado.
3. **NO se propondrán optimizaciones** sin evidencia de bottleneck específico en este horizonte.

---

## PARTE VI — HORIZONTE 8: LIVING NETWORK GRAPH CON KÙZU Y TELEMETRÍA EN TIEMPO REAL

### 8.0 Pregunta Central

> ¿Podemos consultar estructuralmente lo que ocurre **mientras** ocurre, respondiendo preguntas como *"¿qué rutas cambiaron después del endpoint change del nodo X en las últimas 24 h?"* sin que Kùzu interfiera con la red?

### 8.1 Transición Epistémica

```text
Observación puntual (H6/H7)      ─────►      Memoria estructural persistente
{telemetría lineal}                          {grafo consultable}
```

El grafo es la **memoria de largo plazo** del laboratorio vivo.

### 8.2 Arquitectura de Living Graph

```text
                    IPv7 NODE (camino crítico)
                            │
                  eventos / métricas (bounded)
                            ▼
                  ┌───────────────────┐
                  │ Telem. Buffer     │   ← lock-low, acotado
                  └─────────┬─────────┘
                            │ batch (N eventos / T seg)
                            ▼
                  ┌───────────────────┐
                  │ Kùzu Ingestor     │   ← NUNCA en crítico
                  │ (single writer)   │
                  └─────────┬─────────┘
                            ▼
                  ┌───────────────────┐
                  │ Kùzu Graph DB     │
                  │ Node/Peer/Session │
                  │ Route/Endpoint/   │
                  │ Relay/Event/      │
                  │ Experiment/       │
                  │ Anomaly/Bottleneck│
                  └─────────┬─────────┘
                            ▼
                  ┌───────────────────┐
                  │ ipv7-engineer     │
                  │ · Correlación     │
                  │ · Hipótesis       │
                  │ · Informes        │
                  └───────────────────┘
```

### 8.3 Modelo de Grafo Canónico

```cypher
(:Node      {did, first_seen, last_seen})
(:Peer      {did})
(:Session   {id, transport, encrypted, created_at, closed_at})
(:Route     {id, hops, direct_or_relay, created_at, valid_until})
(:Endpoint  {addr, ip_family, asn, nat_type})
(:Relay     {did, region})
(:Event     {type, timestamp, node_id, peer_id