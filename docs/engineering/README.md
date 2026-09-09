# IPv7 Engineer v1 — Sistema de Ingeniería Experimental Asistida por IA

**Misión**: IPv7 ENGINEER v1  
**Filosofía**: La evidencia experimental tiene prioridad sobre cualquier suposición de diseño.  
**Fecha**: 2026-09-09  

---

## 1. El Ciclo Científico Experimental

`ipv7-engineer` no es parte del protocolo de red ni modifica código a ciegas; es un kit de herramientas para gobernar la evolución del protocolo mediante datos empíricos:

```text
OBSERVAR ──► DETECTAR ──► CORRELACIONAR ──► HIPÓTESIS ──► EXPERIMENTO ──► MEDIR ──► BASELINE ──► DECIDIR
```

---

## 2. Los Tres Roles

- **DAVID**: Arquitecto del protocolo, dueño del repositorio y autoridad final de decisión.
- **ANTIGRAVITY**: Operador, ejecutor de automatizaciones, programador y diseñador de pruebas.
- **CHATGPT**: Analista externo, segunda opinión técnica, auditor crítico de sesgos.

---

## 3. Comandos CLI Principales

```powershell
# 1. Registrar o consultar baselines inmutables
go run .\tools\ipv7-engineer baseline

# 2. Generar scorecard multidimensional
go run .\tools\ipv7-engineer scorecard

# 3. Consultar o registrar cuellos de botella e hipótesis descartadas
go run .\tools\ipv7-engineer bottlenecks

# 4. Exportar paquete para análisis de ChatGPT (IPv7_ANALYSIS_REQUEST v1)
go run .\tools\ipv7-engineer request

# 5. Ejecutar experimento automatizado de IP Roaming
go run .\tools\ipv7-engineer roaming
```

---

## 4. Índice de Documentos

- [`EXPERIMENTS.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/EXPERIMENTS.md): Metodología y diseño de experimentos formales.
- [`BASELINES.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/BASELINES.md): Registro de líneas base por hardware y entorno.
- [`BOTTLENECKS.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/BOTTLENECKS.md): Catálogo de cuellos de botella y memoria de fracasos.
- [`KÙZU.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/KÙZU.md): Integración con Kùzu como memoria de red y correlación.
- [`EXTERNAL_ANALYSIS.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/EXTERNAL_ANALYSIS.md): Protocolos `IPv7_ANALYSIS_REQUEST` y `RESPONSE`.
- [`ROAMING.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/ROAMING.md): Resultados del experimento de movilidad de 25 ensayos.
- [`CHAOS.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/CHAOS.md): Pruebas controladas de disrupción.
- [`OPTIMIZATION.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/OPTIMIZATION.md): Reglas de optimización justificada por evidencia.
- [`CANARY_RUNNER_EVALUATION.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/CANARY_RUNNER_EVALUATION.md): Evaluación formal de soak test y watchdog (3.8h, 454 ciclos, 0 fugas).

