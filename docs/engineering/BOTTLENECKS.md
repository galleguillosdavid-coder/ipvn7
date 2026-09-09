# IPv7 Engineering: Catálogo de Cuellos de Botella y Memoria de Fracasos

## 1. Qué Hace y Qué No Hace

### Qué Hace
- Mantiene un registro permanente y estructurado de los cuellos de botella reales identificados mediante experimentación empírica.
- Preserva la **Memoria de Fracasos (Failed Hypotheses Memory)**: una bitácora inmutable de hipótesis refutadas y optimizaciones fallidas.
- Evita que futuros desarrolladores o modelos de IA sugieran cambios que ya demostraron ser inútiles o dañinos.

### Qué No Hace
- **No borra ni oculta hipótesis fallidas**. En la ciencia de sistemas distribuidos, saber qué NO funciona es tan valioso como saber qué funciona.
- **No cataloga sospechas sin perfilado**: Un cuello de botella sólo se registra si está respaldado por perfiles `pprof`, trazas de SO o métricas de pérdida de paquetes reproducibles.

---

## 2. Cuellos de Botella Activos Registrados

### BOTTLENECK-ADV-01: Caída de Recepción Bajo Fuego Cruzado Adversarial Concurrente
- **Identificador**: `BOTTLENECK-ADV-01`
- **Componente**: Plano de recepción del Sistema Operativo y socket UDP (`adapters/udp_adapter.go` / OS Kernel Buffer).
- **Entorno**: Windows / Linux WSL2 / LAN Wi-Fi.
- **Síntoma Observado**: Al inyectar una ráfaga masiva de 4.995 paquetes hostiles/corruptos concurrentes, la recepción de paquetes legítimos cayó de 1.000 a 468 (53.2% de pérdida).
- **Causa Raíz Demostrada**:
  1. El Core criptográfico de IPv7 **NO falló ni se corrompió** (0 violaciones de integridad, 0 panics).
  2. El buffer del socket del SO (`SO_RCVBUF`) se saturó ante la avalancha de paquetes UDP entrantes a nivel de kernel, descartando datagramas legítimos antes de que el proceso en Go pudiera leerlos con `ReadFromUDP`.
- **Mitigación Planificada**:
  - Ajuste de buffers de socket del SO en el adapter (`SetReadBuffer(4MB)` o superior).
  - Separación de puertos para señalización/handshake y datos.
  - Implementación de filtrado temprano (eBPF en Linux o XDP en kernels soportados).

---

## 3. Memoria de Fracasos: Hipótesis Descartadas (Failed Hypotheses)

| ID Hipótesis | Enunciado Falso | Resultado Empírico / Refutación | Estado |
| :--- | :--- | :--- | :--- |
| **HYP-001-CORE-BUG** | "La pérdida de paquetes durante el flood adversarial se debe a un bug de sincronización o bloqueo en el Core de IPv7." | **REFUTADA**: El Core nunca se bloqueó; los perfiles pprof demostraron que las goroutines del Core estaban inactivas esperando datagramas. La pérdida ocurrió en el socket del kernel de Windows por desborde del búfer de recepción. Tocar el Core habría sido un error grave. | **DISPROVEN** |
| **HYP-002-MUTEX-TELEMETRY** | "Un mutex simple en el bus de telemetría es suficiente para despachar eventos sin impacto de latencia." | **REFUTADA**: Bajo ráfagas de 50.000 eventos/s, la contención del mutex introdujo una degradación de latencia de 180 ns por paquete. Fue reemplazado por un ring buffer lock-free con descarte no bloqueante `DROP_TELEMETRY`. | **DISPROVEN** |

---

## 4. Cómo se Utiliza

Para inspeccionar cuellos de botella e hipótesis fallidas desde el CLI:
```powershell
go run .\tools\ipv7-engineer bottlenecks
```

Para registrar una nueva hipótesis falsa:
```go
bottleneckManager.RecordFailedHypothesis(FailedHypothesis{
    ID: "HYP-003-...",
    Hypothesis: "...",
    ExperimentID: "EXP-...",
    WhyItFailed: "...",
    LessonsLearned: "...",
    Status: "DISPROVEN",
})
```
