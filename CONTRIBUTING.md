# Guía de Contribución y Gobernanza — ipvn7 Network OS

¡Bienvenido al repositorio **ipvn7**! Este proyecto implementa la evolución del protocolo IPv7 como un **Sistema Operativo de Red Overlay P2P Descentralizado**. Para mantener la máxima calidad de ingeniería, reproducibilidad e higiene del código, todos los colaboradores y agentes de IA deben adherirse a las siguientes normas.

---

## 🏛️ Invariantes de Arquitectura y Capas

1. **Protocol Core Congelado (`core/`)**:
   El núcleo algorítmico heredado de `v0.5.0` permanece bajo **Core Freeze estricto**. No se modifican archivos en `core/` salvo refutación experimental formal y aprobación de la autoridad técnica del proyecto.
2. **Extensibilidad mediante Adaptadores (`adapters/`)**:
   Toda nueva capacidad de transporte (radio LoRa, Bluetooth Mesh, Wi-Fi Direct, túneles virtuales) debe encapsularse en `adapters/` respetando la interfaz de comunicación canónica.
3. **Observabilidad Desacoplada (`telemetry/`)**:
   La telemetría opera a través de un Ring Buffer lock-free. Queda estrictamente prohibido introducir llamadas bloqueantes o I/O síncrono en el fast-path de recepción de paquetes.

---

## 🧹 Política Estricta de Raíz Limpia

La raíz del repositorio debe permanecer **100% limpia**:
- **Prohibido**: No colocar archivos ejecutables (`*.exe`, binarios ELF), archivos temporales (`*.bat`, `*.sh`), volcados de logs (`*.log`), carpetas de build ad-hoc o volcados de memoria en la raíz.
- **Ubicación de binarios**: Todos los binarios deben compilarse hacia `bin/` (o empaquetarse en `release/`).
- **Ubicación de scripts**: Todos los scripts de lanzamiento, compilación o pruebas residen en `scripts/`.
- **Ubicación de herramientas**: Analizadores, indexadores y workers residen en `tools/`.

---

## 🛠️ Comandos Canónicos de Trabajo

| Tarea | Entorno Windows (PowerShell) | Entorno Linux / WSL2 (Bash) |
| :--- | :--- | :--- |
| **Compilar Binarios** | `.\scripts\build_windows.ps1` | `.\scripts\build_linux.ps1` |
| **Iniciar Nodo** | `.\scripts\run_node.bat` | `./scripts/wsl_node.sh -port 7002 -ui 8082` |
| **Prueba Malla Dual** | `.\scripts\dual_node_test.ps1` | N/A (orquestado desde PowerShell) |
| **Ejecutar Pruebas** | `go test -v ./...` | `go test -v ./...` |
| **Actualizar Grafo Kùzu** | `python .\tools\indexer\index_project.py` | `python3 ./tools/indexer/index_project.py` |
| **Consultar Grafo Kùzu** | `.\tools\kuzu\kuzu.exe .kuzu_index/ipv7.db` | `./tools/kuzu/kuzu .kuzu_index/ipv7.db` |
| **Delegar a DeepSeek** | `.\scripts\ask_deepseek.ps1 -p "<tarea>"` | `./scripts/ask_deepseek.sh -p "<tarea>"` |

---

## 🤖 Protocolo de Ahorro de Tokens (DeepSeek Worker)

Para tareas cognitivas de gran volumen (auditoría de cientos de líneas de código, derivaciones matemáticas, análisis de capturas de interfaz de usuario), utiliza el worker de DeepSeek para delegar la carga computacional y ahorrar tokens en la sesión de Antigravity:

```powershell
python tools/deepseek/deepseek_worker.py -f archivo1.go archivo2.go -t code_review -o scratch/review.md
```

---

## 🧪 Rigor Epistémico

Al reportar resultados o documentar experimentos, clasifica cada afirmación según su grado de certeza empírica:
- `DEMONSTRATED`: Aserción validada por tests automatizados con código reproducible y cotas cuantitativas.
- `OBSERVED`: Fenómeno detectado en redes reales bajo condiciones no estrictamente reproducibles.
- `INFERRED`: Hipótesis teórica deducida de la arquitectura pendiente de validación empírica.
- `REFUTED`: Hipótesis descartada (debe consignarse en la Memoria de Fracasos para evitar reincidencia).
