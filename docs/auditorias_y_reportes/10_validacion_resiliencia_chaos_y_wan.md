# 10. Validación de Resiliencia en Producción: Caos, Estrés y Direct vs. Relay

**Fecha de Inicio:** 09 de Septiembre de 2026  
**Etapa:** *Sustained / WAN / Chaos Production Validation*  
**Objetivo de Ingeniería:** Someter el protocolo a condiciones adversas extremas para determinar sus límites reales de fallo.  
**Pregunta Rectora:** *«¿Qué condiciones reales todavía pueden romper IPv7?»*  
**Entorno Físico:** Host A (PC `192.168.1.198`) $\leftrightarrow$ Host B (Notebook `192.168.1.106`) sobre red Wi-Fi real.

---

## Índice de Fases de Resiliencia

1. [Fase 1: Inyección de Caos Físico (Pérdida 1%, 5%, 10%, 20% y Jitter)](#fase-1-inyección-de-caos-físico)
2. [Fase 2: Experimento Nuclear: Direct P2P vs. Relay DERP](#fase-2-experimento-nuclear-direct-p2p-vs-relay-derp)
3. [Fase 3: Resiliencia de Sesión y Ruptura Dinámica](#fase-3-resiliencia-de-sesión-y-ruptura-dinámica)
4. [Fase 4: Soak Test Sostenido (Prueba de Remojo)](#fase-4-soak-test-sostenido)

---

## Fase 1: Inyección de Caos Físico (Pérdida 1%, 5%, 10%, 20% y Jitter 50ms)

**Fecha de Ejecución:** 09 de Septiembre de 2026  
**Entorno:** Tráfico físico continuo emitido desde el Notebook (`192.168.1.106`) hacia la PC (`192.168.1.198:9050`) sobre enlace Wi-Fi real.  
**Duración por Escenario:** 10 segundos ininterrumpidos por prueba.  
**Carga Criptográfica:** Encapsulado `core.Container` con firma Ed25519, timestamps en nanosegundos y números de secuencia `uint64`.

### Resultados Empíricos Tabulados

| ID Escenario | Condición de Caos Inducida | Throughput Neto | Tasa de Paquetes (PPS) | Pérdida Observada | Consumo RAM | Path MTU | Estado |
|---|---|---|---|---|---|---|---|
| `EXP-CHAOS-00` | Línea Base (0% Pérdida, 0ms Jitter) | **42,11 Mbps** | **5.178 pps** | 0,00 % | 3,23 MB | 1.280 B | **PASS** |
| `EXP-CHAOS-01` | Pérdida Leve (1% Pérdida Inducida) | **40,88 Mbps** | **5.027 pps** | 1,02 % | 1,50 MB | 1.280 B | **PASS** |
| `EXP-CHAOS-05` | Pérdida Media (5% Pérdida Inducida) | **38,59 Mbps** | **4.745 pps** | 4,83 % | 1,90 MB | 1.280 B | **PASS** |
| `EXP-CHAOS-10` | Pérdida Alta (10% Pérdida Inducida) | **38,00 Mbps** | **4.672 pps** | 10,28 % | 1,08 MB | 1.280 B | **PASS** |
| `EXP-CHAOS-20` | Pérdida Severa (20% Pérdida Inducida) | **35,57 Mbps** | **4.374 pps** | 19,82 % | 1,12 MB | 1.280 B | **PASS** |
| `EXP-CHAOS-JIT` | Jitter Extremo (50ms Jitter Inducido) | **0,34 Mbps** | **42 pps** | 0,00 % | 0,77 MB | 1.280 B | **PASS** |

### Análisis Técnico del Comportamiento ante el Caos
1. **Degradación Proporcional sin Colapso:** Con un 20% de paquetes deliberadamente descartados en vuelo, el throughput disminuyó suavemente de 42,11 Mbps a 35,57 Mbps (-15,5%), demostrando que el receptor no entra en bloqueo (`deadlock`), no sufre ahogamiento de buffers ni intenta reensambles infinitos.
2. **Estabilidad de Memoria Inmutable:** El consumo de memoria RAM se mantuvo constante entre 1,08 MB y 3,23 MB a lo largo de las 6 pruebas consecutivas. No hubo fugas (`memory leaks`) ni acumulación de buffers huérfanos.
3. **Resistencia a Jitter Severo (50 ms):** Con pausas pseudoaleatorias de hasta 50 ms por paquete, el flujo se ralentizó de forma esperada (42 pps), pero sin perder un solo paquete y manteniendo el PMTU efectivo en 1.280 B.

---

## Fase 2: Experimento Nuclear: Direct P2P vs. Relay DERP

*(En preparación)*

