# IPv7 WAN Performance & Chaos Matrix — Fase 3 y 4

**Misión**: IPv7 GLOBAL INTERNET PRODUCTION VALIDATION  
**Documento**: `04_WAN_PERFORMANCE.md`

---

## 1. Parámetros de Inyección de Caos WAN

La caracterización de degradación no busca forzar un resultado "PASS" artificial, sino documentar los límites de usabilidad del protocolo (`OBSERVADO` y `DEMOSTRADO`):

### 1.1 Pérdida Inducida (Packet Loss)
- Niveles: `[0%, 1%, 5%, 10%, 20%, 30%, 50%, 70%]`
- Métricas: Packet Delivery Ratio (PDR), Throughput sostenido (Mbps), RTT p50/p95/p99.

### 1.2 Jitter Inducido
- Niveles: `[1ms, 5ms, 10ms, 25ms, 50ms, 100ms, 250ms, 500ms, 1000ms]`
- Métricas: Dispersión de RTT, buffers de reordenamiento, estabilidad de handshake.

### 1.3 Reordenamiento (Packet Reordering)
- Niveles: `[1%, 5%, 10%, 25%, 50%]`
- Métricas: Eficiencia de ventana anti-replay y tasa de drops válidos desfasados.

### 1.4 Duplicación
- Niveles: `[1%, 5%, 10%, 25%, 50%]`
- Métricas: Tasa de descarte anti-replay en Core sin consumo innecesario de CPU.

### 1.5 Pérdida en Ráfaga (Burst Loss)
- Niveles: `[10, 50, 100, 250, 500, 1000 paquetes continuos]`
- Métricas: Tiempo de detección de desvanecimiento de enlace y recuperación de sesión.

---

## 2. Matriz de PMTU y Path MTU Discovery

- Tamaños probados: `[1500, 1472, 1400, 1360, 1280, 1200, 1100, 1000, 900, 576]` bytes.
- Escenarios de estrés:
  - **PMTU Black-Hole**: Filtrado total de respuestas ICMP Type 3 Code 4. El protocolo debe aplicar probe dinámico y fallback automático al PMTU seguro de 1280 B.
  - **Cambio de MTU en Vuelo**: Reducción de MTU durante transferencia continua sin provocar panic, deadlock ni desincronización de canal E2EE.
