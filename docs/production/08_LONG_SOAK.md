# IPv7 Long Soak Test & Memory Leak Protocol — Fase 10

**Misión**: IPv7 GLOBAL INTERNET PRODUCTION VALIDATION  
**Documento**: `08_LONG_SOAK.md`

---

## 1. Objetivos del Test de Larga Duración (Soak)

Demostrar la estabilidad del runtime en Go y el gestor de memoria frente a:
- Tráfico continuo sostenido.
- Inyección periódica de handshakes y churn de peers.
- Rotaciones periódicas de claves de sesión (rekey).
- Ataques de fuzzing y ráfagas hostiles intermitentes.

---

## 2. Metodología de Monitoreo de Memoria

1. **Muestreo por Intervalo**:
   - Cada 60 segundos se capturan estadísticas de `runtime.ReadMemStats`:
     - `Alloc`: Bytes de heap activos.
     - `TotalAlloc`: Bytes acumulativos asignados.
     - `Sys`: Memoria virtual total obtenida del SO.
     - `NumGC`: Número de ciclos de recolección de basura ejecutados.
     - `NumGoroutine`: Conteo de goroutines vivas.
2. **Perfiles Criptográficos y de Runtime**:
   - Exportación de perfiles `pprof` (`heap`, `goroutine`, `profile`) para auditoría forense.
3. **Criterio de Aprobación**:
   - La curva de `Alloc` debe converger asintóticamente o estabilizarse dentro de un rango acotado tras el calentamiento inicial. Queda terminantemente prohibida cualquier tendencia monótonamente creciente no acotada (Memory Leak).
