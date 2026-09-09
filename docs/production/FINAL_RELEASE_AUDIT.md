# IPv7 Final Release Audit — Informe Oficial Consolidado

**Misión**: IPv7 GLOBAL INTERNET PRODUCTION VALIDATION  
**Documento**: `FINAL_RELEASE_AUDIT.md`  
**Tag de Versión Evaluado**: `IPv7-PRODUCTION-CANDIDATE-0`  
**Fecha de Emisión**: 2026-09-09  
**Estado del Dictamen**: APROBADO CON LÍMITES OPERACIONALES CARACTERIZADOS

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

## 2. OBSERVED LIMITS (Límites Operacionales Observados)

- **Régimen Lineal de Relay DERP**: 0 a 4.200 pps (~8.62 Mbps). A partir de este régimen, el relay activa backpressure y descarte probabilístico en cola para preservar estabilidad y memoria.
- **Tasa Máxima de Handshakes**: ~1.850 handshakes por segundo por núcleo CPU bajo curva de intercambio Noise XX (X25519 + ChaCha20-Poly1305).
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

---

## 5. KNOWN LIMITATIONS & OPEN RISKS

1. **Topologías WAN Intercontinentales Extremas**: RTT > 300 ms con pérdida > 30% requiere mantener buffers de reordenamiento de al menos 128 paquetes para evitar caídas en throughput sostenido.
2. **Entornos con CGNAT Simétrico Estricto**: Requieren obligatoriamente la presencia de un Relay DERP disponible, ya que el hole punching directo no puede garantizar apertura de puerto predecible.

---

## 6. RECOMMENDATIONS FOR PRODUCTION ROLLOUT

1. **Despliegue de Adapters Segregados por Defecto**: Configurar el nodo en producción para escuchar señalización en un puerto público y establecer sockets de transporte dedicados por túnel activo de alta prioridad.
2. **Monitoreo Continuo de Colas UDP**: Supervisar `netstat -su` y contadores de drops en el sistema operativo para alertar saturación de buffers `SO_RCVBUF`.
3. **Adopción Gradual de Criptografía Híbrida**: Mantener activa la capa de crypto-agility con Kyber-768/ML-KEM para salvaguardar el tráfico frente a futuros riesgos criptoanalíticos cuánticos.
