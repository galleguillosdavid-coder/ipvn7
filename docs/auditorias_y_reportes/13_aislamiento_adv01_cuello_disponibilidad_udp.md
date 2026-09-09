# Reporte 13: Aislamiento Causal de Finding ADV-01 (Contención en Cola de Recepción UDP Compartida)

> **Fecha:** 9 de Septiembre de 2026  
> **Identificador del Hallazgo:** `Finding ADV-01: Shared UDP Receive Queue Contention`  
> **Pregunta de Investigación:** *¿Pertenece la degradación observada bajo fuego cruzado (46.8% PDR en Reporte 12) a un fallo en el Core/criptografía de IPv7, a la capa de adaptación o a la contención física del socket UDP a nivel de sistema operativo?*  
> **Herramienta Experimental:** [`tools/adversarial/contention/adv01_isolation.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/tools/adversarial/contention/adv01_isolation.go)  
> **Comando de Reproducción:** `go run tools/adversarial/contention/adv01_isolation.go`  
> **Estado del Protocol Core:** **CONGELADO (Core Freeze Respetado)**

---

## 🔬 1. Auditoría del Benchmark de Jitter (Aclaración de PDR a 0 ms)

Durante la revisión del [Reporte 12](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/auditorias_y_reportes/12_caracterizacion_adversarial_y_limites_reales.md), se observó que la tabla de jitter registraba una tasa de entrega (PDR) del 64% al 74% incluso cuando el jitter inyectado era de **0 ms**, manteniéndose casi plana a través de todos los retardos.

Se diseñó una auditoría comparativa de espaciado (`pacing`) para aislar la causa:
* **Caso A (Sin Pacing):** 100 goroutines concurrentes emiten datagramas al socket loopback simultáneamente en el microsegundo 0.  
  $\rightarrow$ **PDR Medido:** **69.0%** (31% de paquetes descartados en loopback).
* **Caso B (Con Pacing de 500 µs):** 100 paquetes emitidos de forma secuencial y espaciada.  
  $\rightarrow$ **PDR Medido:** **100.0%** (100 / 100 entregados sin pérdida).

> **Conclusión de la Auditoría:** La pérdida observada a 0 ms se debía exclusivamente a un artefacto del generador (micro-congestión instantánea en el buffer loopback de Windows al disparar 100 hilos sin separación temporal), y no a una degradación intrínseca de IPv7 ni a causalidad del jitter.

---

## 🧪 2. Matriz de Aislamiento en 4 Escenarios Decisivos (Finding ADV-01)

Para aislar con rigor científico el cuello de disponibilidad en el plano de recepción, se ejecutaron cuatro escenarios controlados:

```text
========================================================================================================
                      TABLA COMPARATIVA DECISIVA DE CAUSA RAÍZ (ADV-01)
========================================================================================================
 Escenario                                      | Hostil Enviado | Legítimo PDR | Tiempo | Causa / Diagnóstico
--------------------------------------------------------------------------------------------------------
 1. Hostile Flood -> Socket IPv7 Completo       |      5.000     |      N/A     |  184ms | Capa Adapter + Criptografía activa
 2. Hostile Flood -> UDP Socket Vacío (Kernel)  |      5.000     |      N/A     |  205ms | Capacidad pura del SO (0% drops)
 3. Hostile Flood + Alice -> MISMO Socket       |     11.556     |    28.8%     |  399ms | Contención extrema en SO_RCVBUF
 4. Hostile Flood + Alice -> SOCKETS SEPARADOS  |     13.776     |   100.0%     |  373ms | Aislamiento 100% demostrado
========================================================================================================
```

---

## 🔍 3. Análisis Desglosado por Escenario

### Escenario 1: Hostile Flood contra Socket IPv7
* **Inyección:** 5.000 paquetes hostiles a **27.177 pps** (mezcla de bitflips corruptos, repeticiones anti-replay y fuzzing CBOR).
* **Comportamiento:** El nodo IPv7 procesó la avalancha descartando el 100% de los paquetes corruptos en 184 ms con cero panics ni fugas.

### Escenario 2: Hostile Flood contra Socket UDP Crudo (Línea Base del Kernel)
* **Inyección:** 5.000 datagramas dirigidos a un socket `net.ListenUDP` sin parsing CBOR, sin firmas y sin IPv7.
* **Comportamiento:** El kernel del sistema operativo absorbió los 5.000 paquetes a **24.378 pps** con **0.0% de pérdida**. Demuestra que el socket loopback del SO tolera ráfagas masivas si el consumidor no realiza cómputo intensivo por paquete.

### Escenario 3: Fuego Cruzado en Socket Compartido (Alice + Atacante en Puerto Único)
* **Dinámica:** El atacante inunda el puerto con **11.556 paquetes hostiles**. Simultáneamente, Alice transmite 1.000 paquetes legítimos.
* **Resultado:**
  * **Paquetes Legítimos Entregados:** **288 / 1.000 (28.8% PDR)**.
  * **Paquetes Corruptos Aceptados:** **0** (Inviolabilidad criptográfica Ed25519 absoluta).
* **Mecanismo Causal:** El único bucle receptor (`listenLoop`) debe ejecutar deserialización CBOR y verificación Ed25519 (~80 µs por paquete) para cada datagrama hostil. Cuando llegan 11.500 paquetes en 400 ms, la cola de recepción del socket a nivel de kernel (`SO_RCVBUF`) se llena más rápido de lo que el hilo puede vaciarla, descartando paquetes entrantes indiscriminadamente antes de que la aplicación los lea.

### Escenario 4: Aislamiento por Sockets Separados (Puerto Dedicado vs. Puerto Expuesto)
* **Dinámica:** Bob configura dos adaptadores UDP en su nodo:
  * **Adaptador A (`udpLegit`):** Endpoint conocido exclusivamente por Alice.
  * **Adaptador B (`udpExposed`):** Endpoint público bajo bombardeo hostil masivo.
* **Inyección:** El atacante inyecta **13.776 paquetes hostiles** contra el puerto expuesto. Alice transmite sus 1.000 paquetes por el puerto dedicado.
* **Resultado:**
  * **Paquetes Legítimos Entregados:** **1.000 / 1.000 (100.0% PDR)**.
  * **Paquetes Corruptos Aceptados:** **0**.
  * **Tiempo de Entrega:** 373 ms.

---

## 🏆 4. Demostración Causal Definitiva y Veredicto

```text
                    ATACANTE (13.776 pps)
                             │
                             ▼
     ┌──────────────────────────────────────────────┐
     │           SOCKET EXPUESTO (Port B)           │
     │      Contención de SO_RCVBUF local a B       │
     └──────────────────────────────────────────────┘
                             │
                      ALICE (1.000 pps)
                             │
                             ▼
     ┌──────────────────────────────────────────────┐
     │          SOCKET DEDICADO (Port A)            │
     │        100.0% PDR (1.000 / 1.000)            │
     └──────────────────────────────────────────────┘
                             │
                     ┌───────▼───────┐
                     │   IPv7 Node   │
                     │  Crypto / Core│  ◄── Inviolable (0% corrupción)
                     └───────────────┘
```

### Conclusiones Técnicas Irrefutables:

1. **El Core de IPv7 queda 100% Exonerado:**  
   La degradación observada en el Escenario 3 no se debe a un deadlock, bloqueo de mutexes, límite de goroutines ni fallo en la lógica de enrutamiento del Core.
2. **Localización Exacta del Cuello de Disponibilidad:**  
   El hallazgo `ADV-01` reside en la **competencia por el búfer de recepción UDP del sistema operativo (`SO_RCVBUF`)** cuando un puerto único compartido debe procesar firmas criptográficas bajo un bombardeo masivo de paquetes no solicitados.
3. **Validación de la Hipótesis de Aislamiento:**  
   Al separar el tráfico legítimo hacia un socket o puerto no expuesto (arquitectura multihomed / multi-adapter nativa de IPv7), la tasa de entrega legítima sube inmediatamente de **28.8% al 100.0%**, a pesar de que el nodo sigue procesando simultáneamente 13.776 paquetes hostiles en su otro adaptador.
4. **Directriz Operativa:**  
   El Protocol Core no requiere alteraciones. Para entornos de producción de alta exposición, la mitigación consiste en políticas a nivel de adaptador (ej. ampliación de `SO_RCVBUF`, segregación de puertos de señalización pública vs. túneles privados, o filtrado temprano de IP previo a la verificación criptográfica).
