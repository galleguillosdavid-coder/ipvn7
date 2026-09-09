# Guía Operativa: Ejecución y Supervisión de IPv7 en WSL2 (Linux)

Esta guía describe cómo compilar, ejecutar, supervisar y depurar nodos IPv7 en un entorno Linux nativo utilizando **WSL2 (Ubuntu)** desde Windows y el IDE Antigravity.

---

## 1. Arquitectura de Ejecución Híbrida (Windows + WSL2)

WSL2 proporciona un kernel Linux virtualizado de alto rendimiento con compartición de sistema de archivos en `/mnt/c/`.

```
+--------------------------------------------------------------------+
|                         WINDOWS HOST                               |
|                                                                    |
|  +---------------------------+       +--------------------------+  |
|  |     Antigravity IDE       |       |       Navegador Web      |  |
|  |  (Supervisión / Edición)  |       |  http://localhost:8080   |  |
|  +-------------+-------------+       |  http://localhost:8082   |  |
|                |                     +------------+-------------+  |
|  +-------------v-------------+                    |                |
|  | scripts/run_wsl.ps1       |                    |                |
|  | scripts/dual_node_test.ps1|                    |                |
|  +-------------+-------------+                    |                |
+----------------|----------------------------------|----------------+
                 | wsl.exe                          |
+----------------v----------------------------------v----------------+
|                         WSL2 (UBUNTU LINUX)                        |
|                                                                    |
|  +---------------------------+       +--------------------------+  |
|  |  scripts/wsl_node.sh      | ----> |  bin/ipv7-node-linux     |  |
|  +---------------------------+       |  (UDP, QUIC, DHT, E2EE)  |  |
|                                      +--------------------------+  |
|                                                   |                |
|                                      +------------v-------------+  |
|                                      |  logs/wsl_node.log       |  |
|                                      +--------------------------+  |
+--------------------------------------------------------------------+
```

---

## 2. Scripts de Automatización Disponibles

### Compilación Cruzada para Linux
Genera los binarios optimizados `bin/ipv7-node-linux` y `bin/chat-linux`:
```powershell
.\scripts\build_linux.ps1
```

### Iniciar y Supervisar un Nodo en WSL2
Lanza el nodo dentro de Ubuntu WSL2, transmitiendo logs a la consola y registrándolos en `logs/wsl_node.log`:
```powershell
# Parámetros por defecto: Puerto P2P 7002, UI 8082
.\scripts\run_wsl.ps1

# Personalizando puertos y conectando a un peer inicial
.\scripts\run_wsl.ps1 -Port 7004 -UIPort 8084 -Peer 127.0.0.1:7001
```

El dashboard web del nodo en Linux queda inmediatamente disponible en Windows en:
**`http://localhost:8082`** (o el puerto configurado).

---

## 3. Prueba de Malla Cruzada Dual (Windows $\leftrightarrow$ WSL2 Linux)

Para validar la comunicación P2P real entre el sistema operativo host (Windows) y el entorno Linux (WSL2):

```powershell
.\scripts\dual_node_test.ps1
```

Este script:
1. Compila los binarios necesarios para Windows y Linux.
2. Inicia el **Nodo A** en Windows (Puerto 7001, UI en `http://localhost:8080`).
3. Inicia el **Nodo B** en WSL2 Linux (Puerto 7002, UI en `http://localhost:8082`, enlazado al Nodo A).
4. Abre ambos dashboards para chatear con cifrado E2EE y observar cómo los nodos se reconocen en la **Topología de Malla**.
5. Al presionar `Ctrl+C`, limpia los procesos de ambos entornos de forma ordenada.

---

## 4. Inspección de Logs y Depuración en Tiempo Real

### Ver logs en vivo desde PowerShell (Windows)
```powershell
Get-Content -Path .\logs\wsl_node.log -Wait -Tail 30
```

### Ver logs en vivo desde WSL2 (Bash)
```bash
tail -f logs/wsl_node.log
```

### Verificar puertos abiertos en WSL2
```bash
wsl -d Ubuntu -e bash -c "ss -tuln"
```
