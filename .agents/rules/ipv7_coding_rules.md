# Reglas de Ingeniería de Sistemas y Arquitectura para ipvn7

Este documento establece las directrices obligatorias de ingeniería de sistemas, arquitectura de red, concurrencia en Go y gobernanza operativa para el desarrollo del **Sistema Operativo de Red ipvn7** dentro del IDE Antigravity y entornos de ejecución asociados (Windows + WSL2 Ubuntu).

---

## 1. Principio Fundamental de Capas y Core Freeze (L0 a L4)

El sistema se estructura en 5 capas estrictamente desacopladas:

```text
┌─────────────────────────────────────────────────────────────────┐
│ L4: APLICACIONES & GOBERNANZA                                   │
│     SPA Web (ui/), CLI Chat (cmd/chat), Auditoría y David       │
├─────────────────────────────────────────────────────────────────┤
│ L3: INGENIERÍA ASISTIDA & IA                                    │
│     MCP Server (core/mcp.go), DeepSeek Worker, Living Lab       │
├─────────────────────────────────────────────────────────────────┤
│ L2: OBSERVABILIDAD DESACOPLADA (telemetry/)                     │
│     Ring Buffer Lock-Free, Prometheus OpenMetrics, Kùzu Export  │
├─────────────────────────────────────────────────────────────────┤
│ L1: ADAPTADORES & SUBSTRATO DE RED (adapters/, dht/)            │
│     TUN OS (fd07::/64), Sphinx Onion, Off-Grid Mesh, STUN/UPnP │
├─────────────────────────────────────────────────────────────────┤
│ L0: PROTOCOL CORE (core/ — CONGELADO / STRICT FREEZE)           │
│     DID Ed25519, Contenedores CBOR, Noise XX, E2EE ChaCha20     │
└─────────────────────────────────────────────────────────────────┘
```

- **Invariante de Core Freeze**: El directorio `core/` permanece **inmutable byte a byte**. Ningún PR o modificación puede alterar algoritmos de `core/` a menos que una hipótesis formal falsada sea aprobada por la autoridad del proyecto.
- **Toda nueva capacidad vive en adaptadores**: Nuevos medios físicos (LoRa, BLE, SDR), virtualización o protocolos de transporte deben implementarse como adaptadores en `adapters/`.

---

## 2. Identidad Soberana (DID) y Criptografía

- **DID Criptográfico Soberano**: Cada nodo posee un identificador único derivado estrictamente de su clave pública **Ed25519** (`did:ipv7:<hex_pubkey>`).
- **Desacoplamiento Total de IP**: La identidad lógica es 100% independiente de direcciones IP, puertos UDP o interfaces de red. La conmutación de IP (roaming Wi-Fi / WAN) conserva intacta la sesión lógica.
- **Canonicidad CBOR (RFC 8949)**: Todo contenedor de red debe serializarse con codificador determinista (`fxamacker/cbor/v2`) previo a computar o verificar firmas Ed25519.
- **Cifrado E2EE Autenticado**: Tráfico confidencial cifrado de extremo a extremo mediante **X25519 ECDH + ChaCha20-Poly1305**. Los nodos intermediarios o relays jamás acceden al texto plano.
- **Handshake Autenticado**: Protocolo **Noise XX** con verificación temporal contra ataques de repetición.
- **Blindaje Post-Cuántica (PQC)**: El subsistema `core/pqc.go` implementa Kyber ML-KEM para resistencia ante ordenadores cuánticos.
- **Filtro Anti-Replay**: Todo adaptador de recepción debe pasar paquetes por `adapters/replay_filter.go` con ventana deslizante y Bloom filter.

---

## 3. Concurrencia en Go y Fast-Path de Red (Cero Alocaciones)

- **Cero Bloqueos en el Critical Path**: Las operaciones de I/O pesadas (escritura a disco, indexación Kùzu, llamadas HTTP/REST) **JAMÁS** deben ejecutarse sincrónicamente en el bucle de recepción UDP o TUN.
- **Ring Buffer Lock-Free**: La telemetría debe despacharse exclusivamente a través del buffer acotado (`telemetry/ring_buffer.go`). Si el buffer se satura, aplica la política `DROP_TELEMETRY` O(1) incrementando contadores sin ralentizar el tráfico.
- **Contextos y Goroutines**: Toda goroutine de socket o procesamiento de fondo debe aceptar `context.Context` o escuchar canales `stopCh` para garantizar una parada limpia con 0 fugas de goroutines.
- **Control de Memoria**: Reutilizar buffers mediante `sync.Pool` en caminos de alto throughput. Prohibido alocar estructuras temporales dentro de bucles de paquetes.

---

## 4. Política Estricta de Raíz Limpia (Clean Repository Root)

- **Prohibido ensuciar la raíz**: Ningún binario (`*.exe`, binarios ELF), script temporal (`*.bat`, `*.sh`), archivo de log (`*.log`), volcado de memoria o caché (`__pycache__`) debe residir en la raíz del repositorio.
- **Ubicación de artefactos**:
  - Binarios compilados $\rightarrow$ `bin/` (o empaquetados en `release/`).
  - Scripts de automatización y launchers $\rightarrow$ `scripts/`.
  - Herramientas y analizadores $\rightarrow$ `tools/`.
  - Bases de datos y grafos temporales $\rightarrow$ `.kuzu_index/` (ignorado en git).
  - Documentación técnica $\rightarrow$ `docs/`.

---

## 5. Ecosistema de Herramientas y Multi-Plataforma

### 5.1 WSL2 Ubuntu + Windows Host
- El desarrollo aprovecha la dualidad:
  - Host Windows: IDE Antigravity, navegador web en `http://localhost:8080`, compilación cruzada con Go.
  - WSL2 Ubuntu: Entorno Linux real para sockets nativos, aislamiento de namespaces de red y pruebas canarias.
- **Comandos canónicos**:
  - Compilar Windows: `.\scripts\build_windows.ps1`
  - Iniciar Nodo Windows: `.\scripts\run_node.bat`
  - Compilar Linux: `.\scripts\build_linux.ps1`
  - Iniciar en WSL2: `.\scripts\run_wsl.ps1 -Port 7002 -UIPort 8082`
  - Prueba de Malla Cruzada: `.\scripts\dual_node_test.ps1`

### 5.2 Base de Datos de Grafo Kùzu (openCypher)
- La topología de la red y el AST de código están modelados como un grafo de conocimiento.
- **Actualizar índice**: `python .\tools\indexer\index_project.py`
- **Consultar grafo**: `.\tools\kuzu\kuzu.exe .kuzu_index/ipv7.db` (en Windows) o `./tools/kuzu/kuzu .kuzu_index/ipv7.db` (en WSL2).
- **Auditoría arquitectónica**: `python .\tools\indexer\audit_kuzu.py`

### 5.3 Delegación a DeepSeek API Worker (Ahorro de Tokens)
- Para evitar sobrecargar la ventana de contexto de Antigravity en auditorías masivas de código, derivaciones matemáticas o inspección visual de UI, se DEBE delegar la tarea pesada al DeepSeek Worker:
  - **Revisión de código masivo**:
    `python tools/deepseek/deepseek_worker.py -f archivo1.go archivo2.go -t code_review -o scratch/review.md`
  - **Inspección visual de capturas / UI**:
    `python tools/deepseek/deepseek_worker.py -i captura.png -t vision_ui`
  - **Razonamiento profundo algorítmico**:
    `python tools/deepseek/deepseek_worker.py -m deepseek-v4-pro -p "Derivación formal del enrutamiento XOR"`

### 5.4 Model Context Protocol (MCP)
- `core/mcp.go` expone el nodo como un servidor MCP nativo sobre `stdio` mediante la bandera `-mcp`. Permite a asistentes de IA consultar la topología, túneles y telemetría vía JSON-RPC.

---

## 6. Rigor Epistémico en Resultados
Todo reporte experimental, log o benchmark debe clasificarse inequívocamente:
1. `DEMONSTRATED`: Respaldado por tests empíricos reproducibles con scripts y cotas numéricas.
2. `OBSERVED`: Visto en ejecuciones vivas pero dependiente de condiciones estocásticas de red.
3. `INFERRED`: Hipótesis lógica pendiente de ensayo destructivo o benchmark formal.
4. `REFUTED`: Hipótesis descartada formalmente (debe registrarse en la Memoria de Fracasos).
