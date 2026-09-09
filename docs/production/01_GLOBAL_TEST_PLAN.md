# IPv7 Global Test Plan — Fase 1 a 15

**Alcance**: Validación empírica integral del protocolo IPv7 bajo condiciones heterogéneas de red local, inter-máquina, emulación WAN y exposición adversarial en capas de transporte.

---

## Estructura de Fases

### Fase 0 — Release Freeze
- Congelación de código fuente del Protocol Core.
- Etiquetado Git `IPv7-PRODUCTION-CANDIDATE-0`.
- Certificación de hash SHA-256 para todos los artefactos compilados.

### Fase 1 — Global Testbed Setup
- Disposición de topología de nodos:
  - Nodo Primario Windows x86_64 (`192.168.1.198`).
  - Nodo Remoto Notebook Windows x86_64 (`192.168.1.106`).
  - Nodo Linux WSL2 Ubuntu (`127.0.0.1` / vEthernet virtual).
  - Nodos Virtualizados Multi-Región (emuladores de latencia/pérdida WAN intercontinental: Chile, USA East, Europa, Asia).

### Fase 2 — Connectivity & NAT Traversal Matrix
- Verificación cruzada de pares:
  - Direct P2P UDP en LAN física sin intermediarios.
  - Relay DERP fallback ante restricción de firewall/NAT simétrico.
  - Soportes Dual-Stack IPv4 / IPv6 y Cross-Stack (IPv4 <-> IPv6 vía Adapter/Relay).
  - Preservación de sesión lógica y Handshake Noise XX bajo Roaming de IP/Puerto.

### Fase 3 — WAN Chaos Matrix
- Pruebas sistemáticas bajo perfil degradado:
  - Pérdida: 0%, 1%, 5%, 10%, 20%, 30%, 50%, 70%.
  - Jitter: 1, 5, 10, 25, 50, 100, 250, 500, 1000 ms.
  - Reordenamiento: 1%, 5%, 10%, 25%, 50%.
  - Duplicación: 1%, 5%, 10%, 25%, 50%.
  - Burst Loss: 10, 50, 100, 250, 500, 1000 paquetes.

### Fase 4 — PMTU Matrix
- Búsqueda y adaptación dinámica de MTU (1500, 1472, 1400, 1360, 1280, 1200, 1100, 1000, 900, 576 bytes).
- Respuesta ante PMTU Black-Hole (bloqueo total de ICMP Destination Unreachable / Packet Too Big).

### Fase 5 — Direct vs Relay Seamless Transition
- Benchmark comparativo de rendimiento (RTT, Throughput, Jitter, CPU, RAM).
- Simulación en caliente de pérdida de ruta directa (`A <-> B Direct -> Ruptura -> A <-> DERP <-> B -> Restauración -> A <-> B Direct`) sin pérdida de estado criptográfico ni desconexión en capa de aplicación.

### Fase 6 — Adversarial Internet Test & Socket Isolation
- Batería de ataques combinados externos: UDP Flood, Malformed CBOR, Truncated Frames, Fuzzing, Replay de 100.000 paquetes, Handshake Flood, Invalid DID.
- Evaluación comparativa: **Shared Adapter** vs. **Isolated Adapters**.

### Fase 7 — Cross-Region Crossfire
- Bombardeo concurrente multi-nodo mientras fluyen sesiones legítimas entre regiones virtuales. Medición de saturación, drops de socket vs. drops de core.

### Fase 8 — Routing Scale & Churn
- Topología Kademlia/Mundo Pequeño con churn dinámico (10%, 25%, 50%, 75% de nodos offline). Medición de convergencia de rutas y tasa de éxito en lookup.

### Fase 9 — Catastrophic Failure & Recovery
- Envío masivo de SIGKILL, reinicio de procesos y máquinas, cambios intempestivos de interfaz de red. Verificación de no-corrupción de base de datos Kùzu y persistencia de DID.

### Fase 10 — Long Soak Test & Memory Profiling
- Operación ininterrumpida con inyección periódica de eventos caóticos. Muestreo de `runtime.MemStats` y perfiles `pprof` (heap y goroutines).

### Fase 11 — Capacity Matrix
- Determinación de techos máximos sostenibles de PPS, Mbps, Handshakes por segundo y consumo de CPU por núcleo.

### Fase 12 — Platform Matrix
- Verificación cruzada Windows 11 x86_64, Linux Ubuntu x86_64 y compatibilidad arquitectónica.

### Fase 13 — Cryptographic Validation
- Auditoría empírica de confidencialidad, autenticidad, Perfect Forward Secrecy (PFS), anti-replay tras reinicio, rotación de claves y desacoplamiento de identidad DID.

### Fase 14 — Upgrade & Rollback
- Transición en caliente entre versiones de binarios asegurando compatibilidad hacia atrás y preservación de estado.

### Fase 15 — Release Candidate Final Gate
- Compilación de resultados en JSON (`docs/production/RESULTS/`) y dictamen en `FINAL_RELEASE_AUDIT.md`.
