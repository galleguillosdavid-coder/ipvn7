# IPv7 Engineering: Protocolo de Experimentación Formal

## 1. Qué Hace y Qué No Hace

### Qué Hace
- Estandariza el ciclo de vida de todo experimento técnico sobre la red overlay IPv7.
- Exige hipótesis falsables antes de tocar una sola línea de código o configuración.
- Registra variables independientes, dependientes y de control.
- Garantiza reproducibilidad mediante semillas pseudoaleatorias y parámetros explícitos.
- Establece criterios claros de `ACCEPT` vs `REVERT` (Rollback) basados en métricas duras.

### Qué No Hace
- **No permite "optimizar a ciegas"** basándose en intuiciones o modas de programación.
- **No modifica el Core (`core/`)** durante la fase experimental. Las pruebas se ejecutan mediante adaptadores, arneses de test o herramientas auxiliares en `tools/`.
- **No acepta afirmaciones sin datos numéricos** ni gráficos/trailers de telemetría cuantificables.

---

## 2. Ciclo de Vida de un Experimento

```text
[ OBSERVAR ] ───► Telemetría detecta una anomalía o punto de saturación.
      │
      ▼
[ HIPÓTESIS ] ──► Formulación falsable: "Si cambiamos X en Y, la métrica Z mejorará un W% sin degradar K".
      │
      ▼
[ DISEÑO ] ─────► Parámetros: N muestras, duración, entorno hardware, red controlada.
      │
      ▼
[ EJECUCIÓN ] ──► Arneses automatizados con captura de telemetría Prometheus/Kùzu.
      │
      ▼
[ EVALUACIÓN ] ─► Comparación estadística contra el BASELINE inmutable (P50, P95, P99).
      │
      ├─── Si mejora sin regresiones ────► PROPOSE TO DAVID (Aprobación del arquitecto).
      └─── Si falla o no hay mejora ────► REGISTRAR EN MEMORIA DE FRACASOS (Rollback inmediato).
```

---

## 3. Plantilla Estándar de Experimento

Cada experimento debe registrarse bajo el siguiente formato:

```markdown
# EXP-XXX: [Nombre Descriptivo]

- **Fecha**: YYYY-MM-DD
- **Entorno**: [LAN Wi-Fi | Local Loopback | WAN CGNAT]
- **Commit Baseline**: [git hash]
- **Hipótesis**: [Enunciado falsable]
- **Variables Controladas**: [MTU, tasa PPS, número de peers, CPU core affinity]
- **Variable Independiente**: [Lo que se modifica]
- **Variables Dependientes**: [Throughput, Latencia RTT P99, Alocaciones B/op, Packet Loss]
- **Muestreo**: [N iteraciones, duración por iteración]
- **Criterio de Aceptación**: [e.g., Latencia P99 <= 2.0ms con 0% pérdida]
- **Protocolo de Rollback**: [Comando exacto de reversión git/configuración]
```

---

## 4. Cómo se Utiliza y Reproduce un Experimento

### Ejecución de un experimento registrado
Para reproducir la batería de roaming, por ejemplo:
```powershell
go run .\tools\ipv7-engineer roaming
```

Para reproducir la batería adversarial completa:
```powershell
go test -v -count=1 ./tools/adversarial/...
```

Para un benchmark de telemetría de bajo nivel:
```powershell
go test -v -bench=BenchmarkTelemetry -benchmem ./telemetry/...
```

---

## 5. Interpretación de Resultados

Los resultados experimentales se clasifican según el rigor epistémico:
- **DEMONSTRATED**: Probado matemáticamente o con repetibilidad 100/100 en banco de pruebas.
- **OBSERVED**: Medido experimentalmente en al menos N >= 10 ejecuciones en hardware real.
- **INFERRED**: Deducción lógica coherente derivada de observaciones comprobadas.
- **HYPOTHESIS**: Suposición plausible que aún no cuenta con prueba empírica suficiente.
- **NOT_PROVEN / REFUTED**: Descartado por los datos experimentales (queda registrado en la Memoria de Fracasos).

---

## 6. Procedimiento Obligatorio de Rollback

Si durante o tras la ejecución de un experimento se detecta:
1. Regresión en latencia P95/P99 > 10%.
2. Caída en throughput > 5%.
3. Aumento en alocaciones de memoria heap (`allocs/op > 0` en el fast-path).
4. Cualquier drop de paquetes en condiciones nominales.

Se debe ejecutar inmediatamente:
```powershell
git checkout -- <archivos_modificados>
```
Y registrar la hipótesis fallida en `tools/ipv7-engineer/bottlenecks/` mediante `RecordFailedHypothesis`.
