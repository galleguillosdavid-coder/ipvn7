# Capítulo 2: Instalación y Puesta en Marcha de IPv7

Este capítulo detalla cómo preparar el entorno de ejecución, compilar los binarios, configurar puertos y ejecutar nodos IPv7 en sistemas Windows y Linux/WSL2.

---

## 1. Requisitos Previos

- **Sistema Operativo**:
  - Windows 10 u 11 (64-bit).
  - Distribuciones Linux (Ubuntu 20.04+, Debian 11+, Arch, Fedora).
  - WSL2 (Windows Subsystem for Linux) con Ubuntu instalado.
- **Go (Opcional si usas binarios precompilados)**:
  - Go versión `1.21` o superior para compilar desde el código fuente.
- **Navegador Web Moderno**:
  - Edge, Chrome, Brave o Firefox para el Dashboard Web.

---

## 2. Puesta en Marcha Inmediata (Windows)

Si ya cuentas con los ejecutables en la raíz del proyecto, puedes iniciar un nodo con un solo clic:

1. **Vía Launcher por lotes**:
   - Haz doble clic en el archivo [INICIAR_AQUI.bat](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/INICIAR_AQUI.bat).
   - Este script arranca el nodo en segundo plano y abre tu navegador en `http://localhost:8080`.

2. **Vía PowerShell o CMD**:
   ```powershell
   # Desde la carpeta raíz del proyecto:
   .\ipv7-node.exe
   ```

   **Salida esperada en consola:**
   ```text
   ==================================================================
                IPv7 NEXT-GEN NODE (E2EE + P2P MESH + UI)            
   ==================================================================
   [ID]   Ed25519 Public Key : 8a73f...
   [E2EE] X25519 Encrypt Key : b4190...
   [...]  Discovering endpoints via STUN...
   [OK]   Reachable Endpoints:
          - 192.168.1.45:7001
          - 181.42.10.88:45231
   [UI]   Dashboard web available at: http://127.0.0.1:8080
   >> Node running! Open http://localhost:8080 in your browser.
   >> Press Ctrl+C to terminate.
   ==================================================================
   ```

---

## 3. Opciones y Parámetros de Línea de Comandos

El ejecutable `ipv7-node.exe` (o binario Linux `ipv7-node-linux`) admite múltiples banderas configurables:

| Parámetro | Valor por Defecto | Descripción |
|---|---|---|
| `-port <número>` | `7001` | Puerto UDP base. El adaptador QUIC utiliza automáticamente `port + 1`. |
| `-ui <número>` | `8080` | Puerto TCP para el servidor web y WebSockets. Poner `0` para desactivar la UI. |
| `-peer <ip:port>` | `""` | Dirección física de un nodo conocido para unirse de inmediato a la red. |
| `-key <ruta>` | `""` | Ruta al archivo de clave persistente Ed25519 o `'persistent'` para usar `~/.ipv7/identity.key`. |
| `-mcp` | `false` | Ejecuta el nodo como servidor Model Context Protocol (MCP) sobre stdio para agentes de IA. |
| `-log-json` | `false` | Emite logs estructurados en formato JSON (estándar `log/slog`). |
| `-upnp` | `true` | Intenta abrir automáticamente el puerto UDP en el router local mediante UPnP IGD. |
| `-stun <host:port>` | `stun.l.google.com:19302` | Servidor STUN utilizado para NAT Traversal e IP pública reflexiva. |
| `-open <bool>` | `true` | Abre automáticamente la pestaña en el navegador predeterminado. |

### Ejemplos Prácticos:

- **Arrancar con identidad fija persistente:**
  ```powershell
  .\ipv7-node.exe -key persistent
  ```
  *(La clave se conserva en `~/.ipv7/identity.key`; nunca cambia entre reinicios).*

- **Ejecutar como servidor MCP para asistentes de IA (Claude / Cursor / Antigravity):**
  ```powershell
  .\ipv7-node.exe -mcp
  ```

- **Lanzar en modo producción con logs estructurados JSON:**
  ```powershell
  .\ipv7-node.exe -log-json -ui 8080
  ```

- **Lanzar un segundo nodo en la misma PC sin colisión de puertos:**
  ```powershell
  .\ipv7-node.exe -port 7005 -ui 8085
  ```
  *(Nota: El nodo cuenta con detección automática de colisión: si el puerto 7001 u 8080 está en uso, avanza dinámicamente al siguiente disponible).*

- **Conectar inmediatamente a otro nodo de la oficina:**
  ```powershell
  .\ipv7-node.exe -peer 192.168.1.50:7001
  ```

- **Ejecutar modo servidor silencioso (Headless, sin abrir navegador ni UI):**
  ```powershell
  .\ipv7-node.exe -ui 0 -open=false
  ```

---

## 4. Compilación Cruzada (Windows a Linux)

Para compilar las versiones actualizadas para Linux / WSL2:

```powershell
# Ejecutar script automatizado de compilación:
.\scripts\build_linux.ps1
```

O manualmente con comandos de Go:
```powershell
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o bin/ipv7-node-linux ./cmd/node
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o ipv7-node.exe ./cmd/node
```

---

## 5. Ejecución y Supervisión en WSL2 (Ubuntu)

Para correr un nodo dentro de WSL2 y monitorearlo en vivo desde Windows:

1. **Lanzar en WSL2 con logs automáticos:**
   ```powershell
   .\scripts\run_wsl.ps1 -Port 7002 -UIPort 8082
   ```
2. **Acceder a la interfaz web de WSL2 desde Windows:**
   - Abre `http://localhost:8082` en cualquier navegador en Windows.
3. **Ver logs de WSL2 en tiempo real:**
   ```powershell
   Get-Content -Path .\logs\wsl_node.log -Wait
   ```

---

## 6. Configuración de Firewall de Windows

Al arrancar por primera vez, Windows Defender Firewall mostrará una alerta emergente solicitando permisos de red:
- Marca la casilla **Redes privadas (red doméstica o del trabajo)**.
- Haz clic en **Permitir acceso**.
- Si no otorgas este permiso, el nodo solo podrá comunicarse con procesos locales (`127.0.0.1`), pero los peers en tu red local o Internet no podrán alcanzar tu puerto UDP.
