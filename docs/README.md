# 📚 Hub Central de Documentación IPv7

Bienvenido al centro oficial de documentación del protocolo descentralizado **IPv7**. Toda la base de conocimiento técnico, arquitectónico, operativo, experimental y de observabilidad se encuentra estructurada y clasificada en este directorio [`docs/`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs).

---

## 🗺️ Mapa de Navegación por Módulos

```text
docs/
├── engineering/            # Sistema de Ingeniería Experimental, Laboratorio Vivo y Baselines
├── telemetry/              # Capa de observabilidad desacoplada, Prometheus y Kùzu Journal
├── production/             # Despliegue canario CANARY-01 y auditorías de Release Candidate
├── auditorias_y_reportes/  # Informes empíricos 01 al 13 (adversarial, caos, ADV-01)
├── arquitectura/           # Fundamentos criptográficos, ruteo Small-World y diseño de red
├── operaciones/            # Manuales operativos, instalación, WSL2 y acceso remoto
├── skill_ia/               # Especificación agéntica MCP, OpenAPI y schemas de IA
└── indices/                # Catálogo de símbolos, paquetes y grafo de código
```

---

## 1. 🔬 Ingeniería Experimental y Laboratorio Vivo (`docs/engineering/`)

Gobierna la evolución del protocolo mediante evidencia empírica inmutable y auditoría externa independiente:
- [**README de Ingeniería**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/README.md): Filosofía, ciclo científico experimental y comandos del CLI `ipv7-engineer`.
- [**LIVING_LAB_REPORT.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/LIVING_LAB_REPORT.md): Informe de ejecución del Laboratorio Vivo (cumplimiento 17/17 de la Sección 28 de `docs/3.md`).
- [**EXPERIMENTS.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/EXPERIMENTS.md): Metodología estandarizada, ciclo de vida de hipótesis y registro histórico (`EXP-ROAM-01`, `EXP-SOAK-01`, `EXP-PHYSICAL-NB-01`).
- [**BASELINES.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/BASELINES.md): Registro inmutable de líneas base de rendimiento, umbrales de regresión y consumo de recursos.
- [**BOTTLENECKS.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/BOTTLENECKS.md): Catálogo de cuellos de botella reales (`BOTTLENECK-ADV-01`) y Memoria de Fracasos (`HYP-001-CORE-BUG`).
- [**KÙZU.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/KÙZU.md): Integración con Kùzu Graph Engine, DDL Cypher y resiliencia con journal append-only.
- [**EXTERNAL_ANALYSIS.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/EXTERNAL_ANALYSIS.md): Especificación de contratos `IPv7_ANALYSIS_REQUEST v1` y `RESPONSE v1` para ChatGPT.
- [**ROAMING.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/ROAMING.md): Reporte cuantitativo de 25 ensayos de IP Roaming (P50: 1.10 ms, 0% pérdida).
- [**CHAOS.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/CHAOS.md): Límites de seguridad y marco para inyección controlada de fallos.
- [**OPTIMIZATION.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/OPTIMIZATION.md): Reglas estrictas de optimización justificada por perfilado pprof (0 allocs/op en fast-path).
- [**CANARY_RUNNER_EVALUATION.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/CANARY_RUNNER_EVALUATION.md): Evaluación formal del soak test de 3.8 horas y supervisor watchdog.

---

## 2. 📡 Telemetría Desacoplada (`docs/telemetry/`)

Arquitectura de observabilidad de ultra bajo costo (< 28 ns/op) sin bloqueo del Protocol Core:
- [**README de Telemetría**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/telemetry/README.md): Principio fundamental de desacoplamiento y buffer lock-free.
- [**SCHEMA.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/telemetry/SCHEMA.md): Definición canónica de eventos de red y anomalías.
- [**METRICS.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/telemetry/METRICS.md): Métricas atómicas (Node, Transport, Protocol) y su significado operacional.
- [**EVENTS.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/telemetry/EVENTS.md): Catálogo exhaustivo de eventos del ciclo de vida del nodo y sesiones.
- [**KUZU_MODEL.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/telemetry/KUZU_MODEL.md): Modelo topológico relacional de grafos en Kùzu.
- [**PROMETHEUS.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/telemetry/PROMETHEUS.md): Guía de integración OpenMetrics Prometheus y prevención de explosión de cardinalidad.
- [**PERFORMANCE.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/telemetry/PERFORMANCE.md): Resultados de microbenchmarks formales de CPU y memoria.
- [**OPERATIONS.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/telemetry/OPERATIONS.md): Monitoreo de producción, políticas de descarte y scraping.
- [**IMPLEMENTATION_REPORT.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/telemetry/IMPLEMENTATION_REPORT.md): Certificación del cumplimiento de la Misión 1 de `docs/3.md`.

---

## 3. 🚀 Producción y Despliegue Canario (`docs/production/`)

Especificaciones de despliegue controlado, gates de calidad y resultados de Release Candidate:
- [**11_CANARY_DEPLOYMENT.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/production/11_CANARY_DEPLOYMENT.md): Arquitectura del despliegue CANARY-01, límites de backpressure, watchdog y resultados consolidados de soak test y nodo físico.
- [**00_RELEASE_GATE.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/production/00_RELEASE_GATE.md): Criterios formales de aprobación de release.
- [**01_GLOBAL_TEST_PLAN.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/production/01_GLOBAL_TEST_PLAN.md): Plan de prueba global de 14 fases automatizadas.
- [**FINAL_RELEASE_AUDIT.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/production/FINAL_RELEASE_AUDIT.md): Dictamen formal de auditoría (`CONDITIONAL GO` y Core Freeze).
- [**RESULTS/ (Fases 01 a 14)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/production/RESULTS): 14 datasets JSON empíricos de conectividad, NAT, wan-chaos, PMTU, relay, adversarial, crossfire, routing, failure, soak, capacity, platform, crypto y upgrade.

---

## 4. 🔍 Auditorías y Reportes Empíricos (`docs/auditorias_y_reportes/`)

Informes de caracterización adversarial, estrés de red y aislamiento causal:
- [**Reportes 01 al 06**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/auditorias_y_reportes/README.md): Auditorías unitarias, benchmarks y verificación topológica de Kùzu.
- [**Reporte 09: Matriz de Validación Experimental de Producción (EPV)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/auditorias_y_reportes/09_matriz_produccion_experimental.md): Establecimiento del Protocol Core Freeze y rigor empírico.
- [**Reporte 10: Validación de Resiliencia y Caos Físico**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/auditorias_y_reportes/10_validacion_resiliencia_chaos_y_wan.md): Pruebas de estrés físico en Wi-Fi real (pérdida, jitter, SIGKILL y comparación Direct vs. Relay).
- [**Reporte 11: Batería Adversarial Extrema y Fallo Seguro**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/auditorias_y_reportes/11_bateria_adversarial_y_resistencia_extrema.md): 18 vectores de ataque: 100.000 replays bloqueados, 0 panics, 0 crashes y 100% autorecuperación.
- [**Reporte 12: Caracterización de Límites Reales y Fuego Cruzado**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/auditorias_y_reportes/12_caracterizacion_adversarial_y_limites_reales.md): Techo del relay (~8.6 Mbps), backpressure y degradación de recepción bajo fuego cruzado 5:1.
- [**Reporte 13: Aislamiento Causal de Finding ADV-01**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/auditorias_y_reportes/13_aislamiento_adv01_cuello_disponibilidad_udp.md): Demostración matemática y experimental de que la contención ocurre en `SO_RCVBUF` del sistema operativo y no en el Protocol Core de IPv7.

---

## 5. 🏛️ Arquitectura y Fundamentos (`docs/arquitectura/`)

- [**01. Plan Génesis de Arquitectura**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/arquitectura/01_genesis.md): Modelo de capas, identidades Ed25519 y serialización CBOR.
- [**02. Enrutamiento en Mundo Pequeño (12 Anillos)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/arquitectura/02_mundo_pequeno_routing.md): Algoritmo voraz XOR acotado a 120 peers con convergencia logarítmica.
- [**03. Modelado de Grafo en Kùzu**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/arquitectura/03_kuzu_mesh_graph.md): Esquema de datos de nodos y aristas para Kùzu Graph Engine.
- [**04. Los Cuatro Horizontes hacia el Nuevo Internet Mundial**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/arquitectura/04_horizontes_nuevo_internet_mundial.md): Hoja de ruta estratégica (TUN/TAP universal, DHT pura soberana, Onion Routing multi-salto y Malla física off-grid).

---

## 6. ⚙️ Operaciones y Puesta en Marcha (`docs/operaciones/`)

- [**01. Conceptos y Fundamentos para Humanos**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/01_conceptos_y_fundamentos_para_humanos.md): La clave pública como dirección de red soberana.
- [**02. Instalación y Puesta en Marcha**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/02_instalacion_puesta_en_marcha.md): Compilación nativa, flags CLI y variables de entorno.
- [**03. Guía del Dashboard Web Interactivo**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/03_guia_dashboard_web.md): Uso de la interfaz SPA en puerto 8080 con física de partículas.
- [**06. Supervisión en WSL2 (Linux)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/06_supervision_wsl2.md): Ejecución y monitoreo de nodos en entornos Linux/WSL2.
- [**07. Acceso Remoto de Dispositivos Móviles y Notebooks**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/07_acceso_remoto_notebook.md): Conectividad Wi-Fi heterogénea y diagnóstico SSH.

---

## 7. 🤖 Suite para Agentes de IA (`docs/skill_ia/`)

- [**SKILL.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/skill_ia/SKILL.md): Manifiesto del asistente autónomo para IPv7.
- [**Servidor MCP Nativo**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/mcp.go): Servidor JSON-RPC 2.0 integrado compatible con Claude Desktop y Antigravity.
- [**Consultas Cypher para Kùzu**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/skill_ia/kuzu_cypher_agent_guide.md): Guía de consultas topológicas para agentes.
