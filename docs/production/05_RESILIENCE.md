# IPv7 Resilience & Catastrophic Recovery — Fase 5, 8 y 9

**Misión**: IPv7 GLOBAL INTERNET PRODUCTION VALIDATION  
**Documento**: `05_RESILIENCE.md`

---

## 1. Conmutación en Caliente Direct <-> Relay

El objetivo es comprobar la conmutación transparente:
```text
[DIRECT PATH] ───(Ruptura / Firewall)───► [RELAY DERP] ───(Restauración)───► [DIRECT PATH]
```
- **Condición Requerida**: Cero pérdida de identidad DID, cero deadlock, y recuperación automática de ruta sin intervención de capa de aplicación.

---

## 2. Escenarios de Falla Catastrófica

1. **Terminación Abrupta (SIGKILL)**:
   - Matar el proceso `ipv7-node` mediante señal incondicional durante flujo continuo de datos.
   - Reiniciar el proceso: Verificar que la base de datos Kùzu y las claves privadas Ed25519 no se corrompen y que el nodo reanuda su DID original.
2. **Reinicio de Máquina**:
   - Reanudación de servicios de red tras reboot del sistema operativo (Windows / Linux).
3. **Pérdida y Reaparición de Peers (Network Churn)**:
   - Churn del 10%, 25%, 50% y 75% en una topología de 12 a 100 nodos.
   - Medir convergencia de la tabla de enrutamiento y tasa de éxito en lookup Kademlia.
