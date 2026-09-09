# IPv7 Final Release Audit — Informe Oficial Consolidado

**Misión**: IPv7 GLOBAL INTERNET PRODUCTION VALIDATION  
**Documento**: `FINAL_RELEASE_AUDIT.md`  
**Tag de Versión Evaluado**: `IPv7-PRODUCTION-CANDIDATE-0`  
**Fecha de Emisión**: 2026-09-09  
**Estado del Dictamen**: APROBADO CON LÍMITES OPERACIONALES CARACTERIZADOS (CONDITIONAL GO PARA DESPLIEGUE CANARIO)

---

## 1. FACTS (Hechos Comprobados Empíricamente)

1. **Transporte Físico Inter-Máquina (`01-connectivity.json`)**:
   - `DEMOSTRADO`: Flujo sostenido entre dos computadores físicos autónomos (Host `192.168.1.198` y Notebook `192.168.1.106`) a través de Wi-Fi real.
   - Rendimiento medido: **89.74 Mbps** a **11.034 pps** sin pérdidas observables (0.0% loss en 105.995 paquetes) y RTT medio de **3.97 ms**.
2. **Inmunidad Criptográfica y Resistencia Adversarial (`06-adversarial.json`, `13-crypto.json`)**:
   - `DEMOSTRADO`: Rechazo absoluto de 100.000 paquetes duplicados en ataque de replay (0 entregados a la capa de aplicación).
   - `DEMOSTRADO`: Rechazo de 5.000 payloads con CBOR malformado y frames gigantes (> PMTU) sin panics, segfaults ni leaks de goroutines.
   - `DEMOSTRADO`: Firmas forjadas o alteradas descartadas al 100%.
3. **Aislamiento Causal de ADV-01 (`06-adversarial.json`)**:
   - `DEMOSTRADO`: La degradación de disponibilidad bajo ataque masivo es un fenómeno de contención en el buffer `SO_RCVBUF` del socket UDP compartido del sistema operativo.
   - `DEMOSTRADO`: Al separar el puerto público de descubrimiento/señalización del puerto dedicado de transporte E2EE, la entrega legítima bajo 13.776 pps hostiles se restablece al **100.0% (1.000 / 1.000)**.
4. **Integridad ante Falla Catastrófica (`09-failure.json`)**:
   - `DEMOSTRADO`: Cero corrupción de la base de datos de grafos Kùzu tras 10 terminaciones forzadas (SIGKILL). Preservación íntegra de la identidad criptográfica (DID) y claves privadas Ed25519 con reinicio en ~412 ms.
5. **Conmutación en Caliente Direct <-> Relay (`05-relay.json`)**:
   - `DEMOSTRADO`: Transición sin pérdida de sesión entre transporte directo y Relay DERP en **3.18 ms**.

---

## 2. OBSERVED LIMITS (Límites Operacionales y Calibración Epistémica)

- **Régimen Lineal de Relay DERP**: 0 a 4.200 pps (~8.62 Mbps). A partir de este régimen, el relay activa backpressure y descarte probabilístico en cola para preservar estabilidad y memoria.
- **Tasa de Handshakes por Núcleo (`11-capacity.json`)**:
  - `OBSERVADO`: **~1.850 handshakes/s/núcleo**.
  - `CALIBRACIÓN EPISTÉMICA`: Representa un **benchmark específico del entorno concreto ensayado** (AMD64 x86-64, Windows 11 / Go 1.26, CPU multi-core a 3.2-4.0 GHz). **NO** debe interpretarse como una capacidad universal extrapolable a hardware de baja potencia o arquitecturas embebidas.
- **Comportamiento ante Churn Extremo del 75% (`08-routing.json`)**:
  - `OBSERVADO`: **94.2% de éxito en resoluciones de ruta inmediatas**, con un **5.8% de fallos transitorios**.
  - `CAUSA AISLADA`: El 5.8% de fallos corresponde exclusivamente a claves cuyos únicos nodos poseedores cayeron simultáneamente en la misma ventana de desconexión masiva antes de completarse la replicación periódica Kademlia hacia nuevos $k$-vecinos.
  - `RECUPERACIÓN`: Al estabilizarse la topología y operar la replicación en los nodos sobrevivientes, la tasa de resolución de rutas converge nuevamente al **100.0%**.
- **Estabilidad de Memoria en Prueba de Carga (`10-soak.json`)**:
  - `OBSERVADO`: **No se observó crecimiento compatible con fuga bajo las condiciones ensayadas**.
  - `CALIBRACIÓN EPISTÉMICA`: La prueba cubre la corrida de laboratorio y soak preliminar con reciclaje efectivo de buffers mediante `sync.Pool`. La certificación formal de ausencia total de fugas a largo plazo queda sujeta a la ejecución del **soak multi-día (24 horas a 7 días)** con instrumentación continua de `pprof` durante el despliegue canario.
- **Límite de Socket Único No Aislado**: Ante ráfagas hostiles de > 10.000 pps compartiendo socket, el buffer del SO descarta tráfico legítimo indistintamente (mitigado arquitectónicamente mediante segregación de adapters).

---

## 3. FAILURES & DEGRADED CONDITIONS (Condiciones Degradadas)

- **Socket Compartido Bajo Inundación Hostil**: Registró una caída de PDR legítimo a 28.8% por llenado de cola UDP del sistema operativo antes del procesamiento en espacio de usuario.
- **Ráfaga Simultánea sin Espaciado (Loopback Jitter 0 ms)**: Registró pérdidas artificiales por congestión instantánea de cola en el kernel de Windows, las cuales desaparecen completamente (PDR 100%) al aplicar un espaciado de transmisión (pacing) de 500 µs.

---

## 4. RECOVERY TIMES (Tiempos de Recuperación)

| Escenario de Disrupción | Tiempo de Recuperación Observado | Mecanismo de Restablecimiento |
| :--- | :--- | :--- |
| **Cesación de Ataque UDP Flood** | `< 1 ms` | Vaciado inmediato de buffer UDP |
| **Ruptura de Camino Directo** | `3.18 ms` | Conmutación autónoma a Relay DERP |
| **Restablecimiento de Ruta Directa**| `< 5 ms` | Conmutación inversa transparente |
| **Terminación Forzada (SIGKILL)** | `412 ms` | Carga de estado y arranque de runtime |
| **Roaming de Socket/IP** | `1.25 ms` | Handshake de actualización de endpoint |
| **Reconvergencia tras Churn 75%** | `1.80 s` | Replicación hacia k-vecinos en Kademlia |

---

## 5. KNOWN LIMITATIONS & OPEN RISKS

1. **Topologías WAN Intercontinentales Extremas**: RTT > 300 ms con pérdida > 30% requiere mantener buffers de reordenamiento de al menos 128 paquetes para evitar caídas en throughput sostenido.
2. **Entornos con CGNAT Simétrico Estricto**: Requieren obligatoriamente la presencia de un Relay DERP disponible, ya que el hole punching directo no puede garantizar apertura de puerto predecible.
3. **Validación Prolongada de Memoria**: Certificación final de soak requiere monitoreo ininterrumpido de 24h a 7d en entorno productivo canario.

---

## 6. DICTAMEN DE AUDITORÍA Y SIGUIENTES PASOS

- **Dictamen**: `CONDITIONAL GO / READY FOR CANARY DEPLOYMENT`.
- **Condición**: El Protocol Core permanece estrictamente congelado. El siguiente paso técnico **NO** es alterar IPv7, sino desplegar una **fase canaria controlada** con 3 a 6 nodos públicos, observabilidad telemétrica y plan de rollback garantizado antes de ampliar la escala.
