# ipvn7: The Living Network Evolution of IPv7
## Fase de Validación Externa, Resiliencia en Red Real y Hardware Físico

[![Status](https://img.shields.io/badge/Status-Active%20Evolution-blue.svg)](docs/engineering/MASTER_PLAN_IPVN7.md)
[![Inherited Protocol Core](https://img.shields.io/badge/Protocol%20Core-FROZEN%20(from%20IPv7%20v0.5.0)-success.svg)](core/)
[![Historical Baseline](https://img.shields.io/badge/Historical%20Baseline-IPv7%20v0.5.0--chaos.audit-orange.svg)](https://github.com/galleguillosdavid-coder/Ipv7)
[![Master Plan](https://img.shields.io/badge/Master%20Plan-Horizons%206--9-violet.svg)](docs/engineering/MASTER_PLAN_IPVN7.md)

**ipvn7** es el repositorio activo para la siguiente etapa de desarrollo del protocolo IPv7. Mientras que el repositorio histórico [galleguillosdavid-coder/Ipv7](https://github.com/galleguillosdavid-coder/Ipv7) queda **formalmente congelado** tras la certificación de la auditoría de segundo orden en la versión 0.5.0-chaos.audit, **ipvn7** asume la misión de sacar el protocolo del laboratorio controlado y llevarlo al mundo real:

- **Horizonte 6**: Validación Continua de 7 Días y Resiliencia en Red Real (Wi-Fi real, NAT simétrico, roaming, pérdida estocástica).
- **Horizonte 7**: Topología Heterogénea Multi-Host en Internet Público (Windows + Linux + ARM interconectados globalmente).
- **Horizonte 8**: Living Network Graph con Kùzu y Telemetría en Tiempo Real (consultas estructurales Cypher sin tocar el camino crítico).
- **Horizonte 9**: Hardware Físico y Transceptores de Radio Reales (LoRa, BLE, Wi-Fi Direct y SDR).

---

# IPv7: Protocolo de Red Overlay P2P Descentralizada

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue.svg)](https://golang.org)
[![Core Status](https://img.shields.io/badge/Protocol%20Core-CONGELADO%20(Core%20Freeze)-success.svg)](#-protocol-core-congelado)
[![Canary Status](https://img.shields.io/badge/CANARY--01-100%25%20HEALTHY-brightgreen.svg)](#-despliegue-canario-controlado-canary-01)
[![Telemetry Overhead](https://img.shields.io/badge/Telemetry%20Overhead-%3C28%20ns%20%2F%200%20allocs-blueviolet.svg)](#-telemetr%C3%ADa-desacoplada-y-living-network-v1)
[![Independent Reproducibility Audit](https://github.com/galleguillosdavid-coder/Ipv7/actions/workflows/reproducibility_audit.yml/badge.svg)](https://github.com/galleguillosdavid-coder/Ipv7/actions/workflows/reproducibility_audit.yml)
[![License](https://img.shields.io/badge/License-Source--Available%20Audit-orange.svg)](LICENSE.md)

Implementación en **Go** del protocolo **IPv7**: una arquitectura de red overlay de alto rendimiento basada en **identidades criptográficas soberanas (DID Ed25519)**, transporte desacoplado de la dirección IP, enrutamiento en mundo pequeño, telemetría lock-free en tiempo real y gobernanza experimental asistida por IA.

---

## 🏛️ Principios Fundamentales y Filosofía de Diseño

1. **Separación Estricta de Planos**:
   - **Identidad**: Fija, criptográficamente soberana (`DID` derivado de Ed25519).
   - **Endpoint / Transporte**: Efímero, dinámico e intercambiable (UDP, STUN, UPnP, Relay DERP).
   - **Routing**: Autónomo, acotado y escalable (Mundo Pequeño XOR de 12 anillos).
   - **Observabilidad**: Desacoplada del fast-path de paquetes; jamás bloquea el tráfico de red.
2. **Rigor Epistémico**:
   - Clasificación estricta de todo resultado en `DEMONSTRATED`, `OBSERVED`, `INFERRED` o `REFUTED`.
   - Cero optimizaciones prematuras o ad-hoc sin respaldo de perfilado `pprof` o microbenchmarks antes/después.
   - Preservación de la **Memoria de Fracasos** para no repetir hipótesis ya refutadas empíricamente.
3. **Core Freeze Estricto**:
   - El núcleo algorítmico del protocolo ([`core/`](core/)) permanece **congelado e inmutable**. Toda extensión opera a través de adaptadores, telemetría asíncrona o herramientas externas.

---

## 🚀 Arquitectura del Sistema

```text
 ┌──────────────────────────────────────────────────────────────────┐
 │                    CAPA DE APLICACIÓN & UX                       │
 │      Dashboard Web SPA (8080) │ Chat E2EE │ Server MCP JSON-RPC  │
 └─────────────────────────────────┬────────────────────────────────┘
                                   │
 ┌─────────────────────────────────▼────────────────────────────────┐
 │                 PROTOCOL CORE (STRICT FREEZE)                    │
 │  DID Ed25519 │ Noise XX │ ChaCha20-Poly1305 │ Greedy XOR Routing │
 └──────┬──────────────────────────┬─────────────────────────┬──────┘
        │                          │                         │
 ┌──────▼──────┐            ┌──────▼──────┐           ┌──────▼──────┐
 │ TRANSPORTE  │            │ TELEMETRÍA  │           │  ENGINEER   │
 │   UDP Fast  │            │  Lock-Free  │           │ Living Lab  │
 │  STUN/UPnP  │            │ Ring Buffer │           │  Baselines  │
 │ Relay DERP  │            │ OpenMetrics │           │ ChatGPT Bus │
 └─────────────┘            └─────────────┘           └─────────────┘
```

### 1. Criptografía e Identidad Soberana ([`core/`](core/))
- Identidades derivadas de pares de claves **Ed25519** (`DID`).
- Contenedores canónicos deterministas **CBOR** (RFC 8949) firmados digitalmente.
- Handshakes autenticados mutualmente (**Noise Protocol XX**) que miden latencia RTT y verifican frescura temporal para prevenir ataques de replay.
- Cifrado autenticado de extremo a extremo (**E2EE ChaCha20-Poly1305**) con claves derivadas X25519 (ECDH). Ni nodos intermediarios ni servidores relay pueden descifrar los paquetes.

### 2. Transporte Desacoplado y Movilidad Dinámica (IP Roaming)
- **IP Roaming Instantáneo**: Demostrado en 25 ensayos automatizados con **100% de éxito, 0% de pérdida de paquetes y latencia P50 de 1.10 ms** (P99: 2.37 ms). Un dispositivo móvil puede cambiar de Wi-Fi o conmutar de IP sin perder su sesión lógica ni renegociar handshakes pesados.
- **NAT Traversal y UPnP**: Detección automática de IP reflexiva vía STUN (`adapters/stun.go`) y apertura dinámica de puertos residenciales mediante UPnP IGD (`adapters/upnp.go`).
- **Relay DERP Ciego**: Fallback transparente para entornos con cortafuegos simétricos estrictos o CGNAT restrictivo.

### 3. Enrutamiento Mundo Pequeño (12 Grados de Separación)
- Tablas de enrutamiento acotadas a 120 peers (12 anillos logarítmicos $\times$ 10 nodos).
- Cobertura teórica global de hasta **$10^{12}$ nodos** con convergencia voraz $O(\log N)$.
- Límite de saltos inviolable (`HopLimit = 12`) con firmas digitales preservadas de extremo a extremo.

### 4. Telemetría Desacoplada y Living Network v1 ([`telemetry/`](telemetry/))
- **Cero Bloqueo de Red**: Buffer acotado lock-free ([`ring_buffer.go`](telemetry/ring_buffer.go)) con política `DROP_TELEMETRY` O(1) si el búfer se llena, incrementando contadores sin ralentizar el socket UDP.
- **Overhead Despreciable Demostrado en Benchmarks**:
  - `BenchmarkTelemetryOFF`: **0.34 ns/op** (0 B/op, 0 allocs/op)
  - `BenchmarkTelemetryON`: **27.18 ns/op** (0 B/op, 0 allocs/op)
  - Ingesta de CPU inferior al **0.27% de 1 núcleo a 100.000 paquetes/segundo**.
- **Exportadores**:
  - **Prometheus OpenMetrics** (`/metrics` y `/health`) con regla estricta anti-cardinalidad.
  - **Kùzu Graph Engine**: Persistencia asíncrona hacia tablas Cypher y journal append-only (`telemetry/journal/events_journal.jsonl`).
- **Anomaly Engine**: Detección en línea sin ML de RTT spikes, caídas de PMTU, saturación de socket y consumo anómalo de memoria/goroutines.

### 5. Motor Experimental IPv7 ENGINEER v1 ([`tools/ipv7-engineer/`](tools/ipv7-engineer/))
- **Baselines Inmutables**: Registro y comparación automática contra líneas base históricas para impedir regresiones de rendimiento (> 5% latencia o > 0 allocs).
- **Memoria de Fracasos (Failed Hypotheses)**: Catálogo estructurado de cuellos de botella (`BOTTLENECK-ADV-01`) e hipótesis descartadas empíricamente (`HYP-001-CORE-BUG`).
- **Scorecard Multidimensional**: Evaluación en 11 dimensiones individuales exportadas a JSON sin números mágicos arbitrarios.
- **ChatGPT Bridge**: Generador estandarizado del paquete `IPv7_ANALYSIS_REQUEST v1` y receptor de `IPv7_ANALYSIS_RESPONSE v1` para auditoría externa sin sesgos.
- **Laboratorio Vivo**: Automatización del ciclo de investigación y validación de extremo a extremo:
  ```text
  IPv7 -> Telemetry -> Kùzu -> Engineer -> Antigravity -> ChatGPT -> David -> IPv7
  ```

---

## 📊 Estado de Pruebas y Despliegue Canario (CANARY-01)

### 1. Batería Adversarial y Resistencia Extrema (Reportes 11 a 13)
- **Ataque de Replay**: 99.999 / 100.000 paquetes hostiles descartados silenciosamente.
- **Fuzzing y Corrupción**: Rechazo de encabezados malformados y CBOR corrupto con **0 panics y 0 crashes**.
- **SIGKILL + Recuperación**: Reconstrucción de sesión y tablas en < 100 ms.
- **Aislamiento ADV-01**: Demostrado que la degradación bajo fuego cruzado ocurre en el buffer del SO (`SO_RCVBUF`) y no en el Protocol Core de Go.

### 2. Soak Test Prolongado (3.8 Horas Continuas)
- **Duración**: 3 horas, 47 minutos y 5 segundos ininterrumpidos (454 ciclos consecutivos de 30s).
- **Tráfico Cursado**: 158.900 paquetes transmitidos / 157.992 recibidos (**99.43% de entrega**).
- **Estabilidad de Recursos**: Crecimiento de memoria heap de sólo +0.03 MB (**0 fugas de memoria**); concurrencia fija en **8 goroutines** (**0 fugas de goroutines**).
- **Liveness**: 100% HEALTHY; cero reinicios activados por el watchdog.

### 3. Validación Multi-Host Heterogénea
- Conexión P2P directa validada entre **PC Host Windows 11** (`192.168.1.198`) y **Notebook Físico Remoto** (`192.168.1.106`) a través de Wi-Fi real.
- Autodescubrimiento LAN UDP broadcast en **4.78 ms**.
- Huella de memoria en reposo del cliente en Windows: **57.6 MB**.
- Soporte nativo de doble pila IPv4 / IPv6 global.
- **Ruta Viva Celular 4G/5G**: Enlace de campo establecido sobre CGNAT simétrico móvil (`45.232.93.x`) con mensajería E2EE confirmada en 1 ms.

---

## 🌍 Visión y Horizontes: Hacia el Nuevo Internet Mundial

El Internet actual confunde identidad con ubicación, carece de cifrado por defecto y depende de monopolios BGP y DNS raíz centralizados. Para transformar a IPv7 en el nuevo sustrato de interconexión global, se han formalizado **Cuatro Horizontes Estratégicos** gobernados por el método científico experimental ([especificación completa en `docs/arquitectura/04_horizontes_nuevo_internet_mundial.md`](docs/arquitectura/04_horizontes_nuevo_internet_mundial.md)):

1. **🌐 Horizonte 1: Adaptador TUN/TAP Universal (`ipv70`)**:
   - Creación de interfaces virtuales de red en Windows, Linux, Android, iOS y OpenWrt para que cualquier aplicación (navegadores, SSH, streaming, juegos) viaje de forma transparente sobre IPv7 sin modificar su código fuente.
2. **🌐 Horizonte 2: Descentralización Total sin Terceros (P2P Puro & DHT Soberana)**:
   - Eliminación de cualquier dependencia de servicios en la nube (como Firebase), reemplazándolo por una DHT Kademlia global autoinmune contra Sybil y nodos semilla distribuidos.
3. **🌐 Horizonte 3: Enrutamiento Cebolla Multi-Salto (Onion Multi-Hop / Sphinx Routing)**:
   - Enrutamiento indirigible sobre el grafo de Mundo Pequeño de Kleinberg: los nodos intermediarios retransmiten paquetes opacos de longitud fija sin conocer el origen ni el destino final.
4. **🌐 Horizonte 4: Malla Física Fuera de Internet (Mesh Off-Grid / Wi-Fi Direct / LoRa)**:
   - Continuidad operativa en redes locales ad-hoc, Wi-Fi Direct y radioenlaces LoRa, garantizando comunicaciones seguras incluso ante cortes masivos de cables submarinos o apagones de infraestructura nacional.

---

## 🔬 Auditoría de Reproducibilidad Independiente (Terceros e IA)

> *"Trust, but verify with code & math"*. Ninguna afirmación en este repositorio debe aceptarse por fe ni por resumen generativo. Todo hecho técnico afirmado en **Horizonte 5** es reproducible directamente por cualquier máquina, proceso o IA externa mediante código determinista:

| Hipótesis Auditada | Afirmación Técnica | Estado Epistémico | Entorno | Test Reproducible |
| :--- | :--- | :---: | :---: | :--- |
| **$H\text{-MULTI-01}$** | Failover WAN $\to$ Malla con DIDs y clave E2EE invariantes | **`DEMONSTRATED`** | `LAB_SIMULATED` | [`tests/multimedium_chaos_test.go`](tests/multimedium_chaos_test.go) |
| **$H\text{-L2-DOS}$** | Resistencia a inundación de 10k balizas forjadas ($>800\text{k}$ balizas/s) | **`DEMONSTRATED`** | `LAB_SIMULATED` | [`tests/l2_dos_and_flapping_test.go`](tests/l2_dos_and_flapping_test.go) |
| **$H\text{-FLAPPING}$** | 30 ciclos de corte y oscilación rápida WAN sin deadlocks | **`DEMONSTRATED`** | `LAB_SIMULATED` | [`tests/l2_dos_and_flapping_test.go`](tests/l2_dos_and_flapping_test.go) |
| **$H\text{-SPLIT-BRAIN}$** | Fusión autónoma de 2 islas con resolución XOR bilateral | **`DEMONSTRATED`** | `LAB_SIMULATED` | [`tests/split_brain_healing_test.go`](tests/split_brain_healing_test.go) |
| **$H\text{-CONSTRAINED-MTU}$** | Paquete 1280B en canal LoRa 180B con desorden y 0 corrupción | **`DEMONSTRATED`** | `LAB_SIMULATED` | [`tests/constrained_link_fragmentation_test.go`](tests/constrained_link_fragmentation_test.go) |
| **$H\text{-UNIFIED-E2E}$** | Pipeline: TUN (ULA) $\to$ DHT $\to$ Onion (1280B) $\to$ Radio L2 $\to$ TUN | **`DEMONSTRATED`** | `LAB_SIMULATED` | [`tests/e2e_four_horizons_test.go`](tests/e2e_four_horizons_test.go) |

- 📋 **Guía Completa de Verificación**: Consulte [`docs/engineering/REPRODUCIBILITY_AUDIT.md`](docs/engineering/REPRODUCIBILITY_AUDIT.md).
- ⚙️ **Ejecutar Auditoría en un solo comando**:
  - En Windows: `powershell -ExecutionPolicy Bypass -File .\scripts\audit_reproducibility.ps1`
  - En Linux/WSL: `./scripts/audit_reproducibility.sh`

---

## 🛠️ Guía Rápida de Comandos CLI

### 1. Ejecutar Suite Completa de Tests Unitarios y de Integración
```powershell
go test -count=1 ./...
```

### 2. Ejecutar Microbenchmarks de Rendimiento y Alocaciones
```powershell
go test -v -bench=BenchmarkTelemetry -benchmem ./telemetry/...
```

### 3. Comandos de Ingeniería Experimental (`ipv7-engineer`)
```powershell
# Ejecutar ciclo completo del Laboratorio Vivo (Secciones 28 y 29 de docs/3.md)
go run .\tools\ipv7-engineer lab

# Ejecutar batería de IP Roaming (25 ensayos automatizados)
go run .\tools\ipv7-engineer roaming

# Consultar o registrar baselines inmutables
go run .\tools\ipv7-engineer baseline

# Generar scorecard multidimensional
go run .\tools\ipv7-engineer scorecard

# Exportar paquete de análisis para ChatGPT (IPv7_ANALYSIS_REQUEST v1)
go run .\tools\ipv7-engineer request

# Consultar catálogo de cuellos de botella y memoria de fracasos
go run .\tools\ipv7-engineer bottlenecks
```

### 4. Lanzar Nodo Local con Dashboard Web UI
```powershell
.\ipv7-node.exe -port 7001 -ui 8080
```
Abre tu navegador en `http://localhost:8080` para acceder a la visualización de topología de malla con física de partículas, radar de 12 anillos, chat E2EE y métricas.

---

## 📚 Estructura de Documentación Técnica ([`docs/`](docs/))

Toda la documentación técnica está centralizada y clasificada por área:
- [**Centro de Documentación Maestro (`docs/README.md`)**](docs/README.md): Hub principal con índice a todos los módulos.
- [**Ingeniería Experimental y Laboratorio Vivo (`docs/engineering/`)**](docs/engineering/README.md): Metodología de experimentos, baselines inmutables, cuellos de botella, consultas Cypher en Kùzu, puente ChatGPT, reporte de roaming y reglas estrictas de optimización.
- [**Arquitectura de Telemetría (`docs/telemetry/`)**](docs/telemetry/README.md): Esquemas de eventos, métricas Prometheus, modelo de grafos Kùzu, benchmarks de latencia y operaciones.
- [**Producción y Despliegue Canario (`docs/production/`)**](docs/production/11_CANARY_DEPLOYMENT.md): Especificación CANARY-01, matriz de pruebas de conectividad y auditoría final Release Candidate.
- [**Auditorías y Reportes Adversariales (`docs/auditorias_y_reportes/`)**](docs/auditorias_y_reportes/README.md): Informes 01 al 13 que documentan la validación empírica contra fallos, caos y ataques de red.
- [**Operaciones (`docs/operaciones/`)**](docs/operaciones/README.md): Manuales de usuario, instalación, ejecución en WSL2 y acceso remoto.
- [**Skill Especializada para Agentes de IA (`docs/skill_ia/`)**](docs/skill_ia/SKILL.md): Especificaciones MCP, OpenAPI y schemas para modelos de lenguaje.

---

## ⚖️ Términos y Condiciones de Uso (Licencia de Auditoría)

Este proyecto se distribuye bajo una **Licencia de Auditoría e Inspección (Source-Available)**:
- 📖 **Permitido**: Leer, auditar, inspeccionar la arquitectura criptográfica y ejecutar pruebas locales de validación y seguridad.
- 🚫 **Prohibido**: Copiar, redistribuir públicamente, crear forks derivados comerciales o explotar el código comercialmente sin autorización expresa y por escrito del autor.

Para más detalles legales, consulte [`LICENSE.md`](LICENSE.md) y [`TERMS_OF_USE.md`](TERMS_OF_USE.md).  
**Copyright (c) 2026 David Galleguillos (@galleguillosdavid-coder). Todos los derechos reservados.**
