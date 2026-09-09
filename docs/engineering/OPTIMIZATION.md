# IPv7 Engineering: Reglas de Optimización Justificada por Evidencia

## 1. Qué Hace y Qué No Hace

### Qué Hace
- Establece la disciplina de ingeniería obligatoria antes de autorizar cualquier refactorización o cambio de código con fines de rendimiento.
- Exige evidencia empírica cuantitativa (perfilado `pprof`, microbenchmarks antes/después, baselines inmutables) antes de tocar el Core.
- Aplica la regla estricta de **0 allocs/op en el fast-path de red**.

### Qué No Hace
- **No permite optimizaciones prematuras**. "No añadas complejidad para resolver problemas hipotéticos".
- **No autoriza cambios en `core/` sin la aprobación expresa de David**.
- **No considera una optimización válida si sólo mejora microbenchmarks sintéticos** degradando la legibilidad, mantenibilidad o seguridad.

---

## 2. Los Diez Mandamientos de Optimización de IPv7

1. **Medir antes de optimizar**: Sin un perfil `pprof` (CPU y memoria) que señale una línea de código específica como cuello de botella, la optimización está prohibida.
2. **Fast-path libre de alocaciones en el heap**: El procesamiento de paquetes en tránsito (cifrado, descifrado, routing y telemetría) debe mantener `0 B/op` y `0 allocs/op`.
3. **Buffers reutilizables con `sync.Pool`**: Los buffers de datagramas y estructuras de eventos deben reciclarse para evitar la presión sobre el Garbage Collector de Go.
4. **La telemetría nunca bloquea la red**: Cualquier sistema de métricas o logging debe tener un coste despreciable (< 30 ns/op) y descarte no bloqueante `DROP_TELEMETRY`.
5. **Preferir algoritmos simples**: Un algoritmo O(N) con memoria contigua en caché suele ser más rápido que estructuras complejas de grafos/árboles para conjuntos pequeños de peers.
6. **No inventar criptografía ni modificar primitivas probadas**: Usar `crypto/ed25519` y `crypto/cipher` estándar de Go; no implementar algoritmos artesanales.
7. **Perfilado en condiciones realistas**: Medir con concurrencia real, jitter de red y pérdidas, no sólo en local loopback aislado.
8. **Validación cruzada**: Comparar la métrica optimizada contra el `BASELINE` inmutable. Si la mejora es menor al 5%, el código añadido no justifica la complejidad.
9. **Registrar fracasos**: Si una optimización prometedora no arrojó mejoras, registrarla en `BOTTLENECKS.md` y `RecordFailedHypothesis`.
10. **Aprobación de David**: Antigravity propone con benchmarks; David decide y autoriza la integración al Core.

---

## 3. Protocolo de Profilado Oficial con pprof

Para perfilar el rendimiento de un nodo IPv7 en ejecución:
```powershell
# Capturar perfil de CPU durante 30 segundos
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

# Capturar perfil de memoria heap
go tool pprof http://localhost:8080/debug/pprof/heap

# Capturar contención de goroutines y mutex
go tool pprof http://localhost:8080/debug/pprof/goroutine
go tool pprof http://localhost:8080/debug/pprof/mutex
```

---

## 4. Criterio de Rollback de Optimizaciones

Toda optimización introducida será revertida inmediatamente (`git checkout`) si:
- Introduce cualquier alocación de heap en la función de recepción/envío de paquetes.
- Incrementa el consumo de CPU en reposo.
- Provoca panics o fugas de goroutines en pruebas de soak prolongadas (e.g. > 1 hora).
