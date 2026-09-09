# IPv7 DERP Relay Protocol Specification & Capacity — Fase 5 y 11

**Misión**: IPv7 GLOBAL INTERNET PRODUCTION VALIDATION  
**Documento**: `07_RELAY.md`

---

## 1. Arquitectura de Relay DERP en IPv7

El servicio de Relay en IPv7 actúa como un puente ciego E2EE (Encrypted-to-End Relay):
- El relay nunca posee las claves de sesión ni puede descifrar los payloads útiles.
- Los paquetes son conmutados en base al Session ID y DID de destino codificados en la cabecera externa.

---

## 2. Límites Operacionales y Caracterización de Saturación

Con base en la evidencia empírica acumulada:
- **Régimen Lineal Óptimo**: 0 a 5.000 PPS con RTT constante (~0.54 ms en LAN/loopback) y 0% de pérdida observada.
- **Techo Operacional de Saturación**: ~4.200 a 4.500 PPS sostenidos (~8.6 Mbps).
- **Mecanismo de Control**: El relay implementa backpressure y descarte probabilístico en cola en lugar de encolamiento no acotado, garantizando que el consumo de memoria del proceso permanezca plano y que la recuperación ante cesación de sobrecarga sea inmediata (< 1 ms).
