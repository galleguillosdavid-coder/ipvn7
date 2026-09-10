# 📚 Portal Central de Documentación — ipvn7 Network OS

Bienvenido al centro oficial de documentación y arquitectura de **ipvn7**: el **Sistema Operativo de Red Overlay P2P Descentralizado**. Toda la base de conocimiento técnico, experimental, operativo y de observabilidad se encuentra estructurada y condensada en este portal.

---

## 🗺️ Mapa de Navegación Rápida

```text
docs/
├── PLAN_MAESTRO_NOS_IPVN7.md   # [SSOT] Plan Maestro, Checklist Integral y 12 Innovaciones
├── arquitectura/               # Especificación del NOS, Criptografía, Mundo Pequeño y TUN OS
├── engineering/                # Metodología experimental, Baselines inmutables y Laboratorio Vivo
├── telemetry/                  # Observabilidad Lock-Free (<28 ns), Prometheus y Grafos Kùzu
├── production/                 # Despliegue Canario CANARY-01 y Gates de Release Candidate
├── auditorias_y_reportes/      # Batería de 13 informes empíricos (adversarial, caos, ADV-01)
├── operaciones/                # Guías de uso, Dashboard Web, WSL2 Ubuntu y Acceso Remoto
├── skill_ia/                   # Especificaciones agénticas MCP, OpenAPI y Schemas de IA
├── indices/                    # Índice integral de código, paquetes y símbolos
└── historico/                  # Archivo de actas de diseño y borradores preliminares superados
```

---

## 🧭 Documentos Fundacionales

1. [**Plan Maestro y Checklist Integral del NOS (`docs/PLAN_MAESTRO_NOS_IPVN7.md`)**](PLAN_MAESTRO_NOS_IPVN7.md):  
   La **Fuente Única de Verdad (SSOT)** que certifica las capacidades demostradas en los Horizontes 1 a 5 y define el checklist interactivo de las **12 dimensiones de innovación** (Firewall ZTNA eBPF, dDNS Petnames, QoS PoW dinámico, Almacenamiento DAG, Tit-for-Tat, ipvn7-cli, WASM Wazero, Multipath QUIC, Web-of-Trust, emparejamiento QR/BLE, AI Copilot y Zero-Copy).

2. [**Especificación Canónica del Sistema Operativo de Red (`docs/arquitectura/SISTEMA_OPERATIVO_RED_IPVN7.md`)**](arquitectura/SISTEMA_OPERATIVO_RED_IPVN7.md):  
   Diseño de ingeniería de sistemas de las 5 capas (L0 a L4), Core Freeze estricto, ciudadela criptográfica (Noise XX, CBOR determinista, ChaCha20 E2EE, PQC Kyber) y substrato virtual TUN/TAP (`fd07::/64` y `10.7.0.0/16`).

3. [**Guía de Contribución y Código Limpio (`CONTRIBUTING.md`)**](../CONTRIBUTING.md):  
   Política estricta de raíz limpia, flujos de trabajo Windows Host $\leftrightarrow$ WSL2 Ubuntu y estándares de cero alocaciones de memoria en Go.

---

## 📂 Módulos de Conocimiento Especializado

### 🔬 [Ingeniería Experimental y Laboratorio Vivo (`docs/engineering/`)](engineering/README.md)
Metodología de contrastación científica donde la evidencia manda sobre cualquier suposición:
- [**EXPERIMENTS.md**](engineering/EXPERIMENTS.md): Metodología y ciclo de vida de hipótesis.
- [**BASELINES.md**](engineering/BASELINES.md): Líneas base inmutables de latencia y alocaciones.
- [**BOTTLENECKS.md**](engineering/BOTTLENECKS.md): Catálogo de cuellos de botella y Memoria de Fracasos.
- [**ROAMING.md**](engineering/ROAMING.md): Reporte cuantitativo de 25 ensayos de IP Roaming (P50: 1.10 ms, 0% pérdida).

### 📡 [Telemetría Desacoplada Lock-Free (`docs/telemetry/`)](telemetry/README.md)
Observabilidad de ultra bajo impacto sin tocar el camino crítico de paquetes:
- [**SCHEMA.md**](telemetry/SCHEMA.md) & [**METRICS.md**](telemetry/METRICS.md): Esquema canónico de eventos y métricas OpenMetrics.
- [**KUZU_MODEL.md**](telemetry/KUZU_MODEL.md): Esquema relacional de grafos en Kùzu DB.
- [**PERFORMANCE.md**](telemetry/PERFORMANCE.md): Certificación del Ring Buffer (<28 ns/op, 0 allocs/op).

### 🔍 [Auditorías Empíricas y Reportes Adversariales (`docs/auditorias_y_reportes/`)](auditorias_y_reportes/README.md)
Validación contra caos, particiones de red y vectores hostiles:
- **Reporte 11**: Batería adversarial extrema (100.000 replays bloqueados, 0 panics, 0 fugas).
- **Reporte 12**: Caracterización de límites reales y fuego cruzado en sockets.
- **Reporte 13**: Aislamiento causal del finding ADV-01 en el buffer `SO_RCVBUF` del sistema operativo.

### ⚙️ [Operaciones y Puesta en Marcha (`docs/operaciones/`)](operaciones/README.md)
Guías prácticas para desarrolladores y administradores de sistemas:
- **Dashboard Web**: Visualizador en tiempo real con física de partículas y radar de 12 anillos en `http://localhost:8080`.
- **Supervisión WSL2**: Ejecución del nodo en Linux Ubuntu bajo WSL2 con `scripts/run_wsl.ps1`.
- **Malla Dual**: Prueba de interconexión Windows $\leftrightarrow$ WSL2 mediante `scripts/dual_node_test.ps1`.

### 🤖 [Ecosistema de Agentes de IA (`docs/skill_ia/`)](skill_ia/SKILL.md)
- **Servidor MCP Nativo**: Soporte Model Context Protocol sobre stdio en `core/mcp.go`.
- **DeepSeek Worker**: Inferencia multimodal de alta velocidad y razonamiento algorítmico en `tools/deepseek/`.

### 📜 [Archivo Histórico (`docs/historico/`)](historico/)
Preservación de borradores y actas de diseño de etapas previas (`01_adv01_aislamiento_causal`, `02_mision_telemetria_v1`, `03_checklist_historico_posibles`, `04_dialogo_vision_gemini_nos`).
