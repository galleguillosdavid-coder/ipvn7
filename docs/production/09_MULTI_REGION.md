# IPv7 Multi-Region Crossfire & Capacity Matrix — Fase 7 y 11

**Misión**: IPv7 GLOBAL INTERNET PRODUCTION VALIDATION  
**Documento**: `09_MULTI_REGION.md`

---

## 1. Topología de Crossfire Multi-Región

Para simular una red global distribuida, se despliegan nodos emuladores con perfiles de retardo y pérdida acordes a trayectos intercontinentales reales:
- **Región A (Chile / Sudamérica)**: Nodo central de recepción / evaluación.
- **Región B (USA East)**: RTT base ~120 ms, Jitter ~10 ms, Pérdida 0.5%.
- **Región C (Europa Oeste)**: RTT base ~210 ms, Jitter ~15 ms, Pérdida 1.0%.
- **Región D (Asia Pacífico)**: RTT base ~330 ms, Jitter ~25 ms, Pérdida 1.5%.

---

## 2. Inyección de Fuego Cruzado Concurrente

- Mientras fluye tráfico E2EE legítimo desde las tres regiones remotas hacia el nodo central, se introduce un emisor hostil que satura el puerto con ráfagas no conformes.
- **Métricas de Evaluación**:
  - Supervivencia de flujos legítimos (PDR individual por región).
  - Comportamiento con puerto compartido vs. puertos dedicados por par.
  - Tiempos de restablecimiento tras cesación de ráfagas.
