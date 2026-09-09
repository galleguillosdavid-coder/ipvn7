# IPv7 Engineering: Protocolo Experimental de Roaming Celular WAN (4G/5G)

**ID de Experimento**: `EXP-WAN-CELLULAR-01`  
**Objetivo**: Evaluar la resiliencia y el comportamiento del protocolo IPv7 durante la transición de red fija (Wi-Fi residencial) hacia red móvil celular (4G/5G bajo CGNAT simétrico estricto) entre dos hosts físicos autónomos.  
**Estado**: **PROTOCOLO OPERATIVO ESTABLECIDO**  

---

## 1. Topología del Ensayo

```text
 ┌───────────────────────────┐                 ┌───────────────────────────┐
 │       HOST A (PC FIJA)    │                 │   HOST B (NOTEBOOK MÓVIL) │
 │     IP: 192.168.1.198     │                 │   DID: 2581f063c103ed29   │
 │   Enlace: Fibra / Wi-Fi   │                 │   Inicial: Wi-Fi LAN      │
 │  DID: f791c67085cf5135... │                 │   Transición: Celular 4G  │
 └─────────────┬─────────────┘                 └─────────────┬─────────────┘
               │                                             │
               │               (Internet WAN)                │
               └──────────────► Firebase STUN ◄──────────────┘
                                      │
                               (Fallback DERP)
                                      │
                            192.168.1.198:7099
```

---

## 2. Variables del Experimento

- **Variable Independiente**: Conmutación de interfaz de red en Host B (desconexión de Wi-Fi y conexión inmediata a hotspot 4G/5G de smartphone).
- **Variables de Control**:
  - Par de claves Ed25519 fijo en Host B (`~/.ipv7/identity.json`).
  - Buffer de socket `SO_RCVBUF` fijado en 4MB.
  - Intervalo de beacon de señalización: 10 segundos.
- **Variables Dependientes**:
  - Tiempo de detección de cambio de IP / endpoint por Host A.
  - Latencia de conmutación (ms) hasta la primera entrega cifrada E2EE tras el salto.
  - Modo de transporte resultante: `DIRECT UDP` (vía STUN reflexivo) o `RELAY DERP` (si el CGNAT del operador móvil bloquea datagramas entrantes no solicitados).
  - Tasa de pérdida de paquetes durante el salto.

---

## 3. Procedimiento de Ejecución Paso a Paso

1. **En Host A (PC Windows Fija)**:
   ```powershell
   .\ipv7-node.exe -port 7001 -ui 8080 -firebase
   ```
2. **En Host B (Notebook Remoto por SSH `192.168.1.106`)**:
   - Conectar a red Wi-Fi LAN y verificar handshake inicial (RTT ~4.78 ms).
   - Iniciar emisión de flujo de datos continuo hacia Host A.
3. **Inyección de Evento de Roaming**:
   - Desconectar Wi-Fi en el Notebook y activar interfaz celular 4G/5G.
   - El socket del notebook adquiere una IP privada de operador (ej. `10.x.x.x` o `100.64.x.x` CGNAT).
4. **Captura de Telemetría**:
   - El bus de telemetría emitirá `EventEndpointChanged` y registrará la conmutación en Kùzu journal.
   - Si la perforación directa falla, se activará `EventRelayEntered` hacia el relay DERP en menos de 500 ms sin abortar la aplicación.

---

## 4. Criterios de Aceptación (Success Criteria)

| Métrica | Umbral de Éxito | Comportamiento Inaceptable (Fail) |
| :--- | :--- | :--- |
| **Preservación DID** | **100% ID idéntica** | Regeneración de claves o cambio de DID |
| **Tiempo de Recuperación** | **< 3.000 ms** (incluyendo handover del módem) | Desconexión permanente (> 10s) |
| **Integridad Criptográfica**| **0 violaciones de firma / 0 panics** | Decodificación errónea de CBOR |
| **Pérdida en Régimen Estable**| **0.0%** tras conmutar | Pérdida sostenida > 2% |
