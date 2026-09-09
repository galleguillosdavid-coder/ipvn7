# Guía para Acceso y Control Remoto del Notebook desde esta PC

Para que yo pueda ejecutar comandos, auditar, compilar y supervisar el nodo directamente en tu notebook (o para que tú lo controles mediante pantalla remota), dispones de las siguientes opciones según lo que prefieras:

---

## 🚀 Opción A: Acceso por Terminal SSH (La mejor para mí como Asistente IA)

Con OpenSSH activo en tu notebook, **yo podré ejecutar comandos directamente en el notebook desde esta terminal**, ver logs en tiempo real, lanzar el nodo y hacer pruebas de red sin que tengas que cambiarte de computador.

### Pasos en el Notebook (Se hace una sola vez):
1. En el notebook, presiona `Inicio`, escribe **PowerShell**, haz **clic derecho** y selecciona **"Ejecutar como administrador"**.
2. Copia y pega este comando y presiona `ENTER`:
   ```powershell
   Add-WindowsCapability -Online -Name OpenSSH.Server~~~~0.0.1.0
   Start-Service sshd
   Set-Service -Name sshd -StartupType 'Automatic'
   New-NetFirewallRule -Name 'OpenSSH-Server-In-TCP' -DisplayName 'OpenSSH Server (sshd)' -Enabled True -Direction Inbound -Protocol TCP -Action Allow -LocalPort 22
   ```
3. ¡Listo! El notebook ya aceptará conexiones remotas.
4. Para conectarme desde esta PC, solo necesitaré saber:
   - El **Usuario** de Windows de tu notebook.
   - La **Contraseña** o PIN.
   - La IP ya la conocemos: `192.168.1.106`.

---

## 🖥️ Opción B: Escritorio Remoto Nativo de Windows (RDP)

Permite ver la pantalla completa del notebook desde esta computadora usando la aplicación oficial de Windows.

> **Nota:** RDP solo viene activado si el notebook tiene **Windows 10/11 Pro** o superior. Si tiene Windows Home, ve a la **Opción C**.

### Pasos en el Notebook:
1. Abre **Configuración** (`Win + I`).
2. Ve a **Sistema** $\rightarrow$ **Escritorio remoto**.
3. Activa la casilla **"Habilitar Escritorio remoto"** y confirma.
4. Anota el nombre de usuario de la cuenta.

### Cómo conectarte desde esta PC:
1. Presiona `Win + R`, escribe:
   ```text
   mstsc
   ```
2. En equipo escribe:
   ```text
   192.168.1.106
   ```
3. Haz clic en **Conectar**, ingresa el usuario y contraseña del notebook, y se abrirá la pantalla de tu notebook en una ventana de esta PC.

---

## 🌐 Opción C: RustDesk (Recomendado si el notebook tiene Windows Home)

**RustDesk** es una aplicación de escritorio remoto libre, rápida y de código abierto (alternativa a TeamViewer/AnyDesk) que funciona en cualquier edición de Windows:

1. En el notebook, entra a: **https://rustdesk.com**
2. Descarga la versión portable para Windows (no requiere instalación compleja).
3. Ábrela: verás un **ID** (de 9 dígitos) y una **Contraseña**.
4. Puedes descargar RustDesk también en esta PC, escribir el ID del notebook y tendrás control total de la pantalla, teclado y mouse.

---

## 📋 Resumen de Datos del Notebook:
- **IP Local en Wi-Fi:** `192.168.1.106`
- **Puerto P2P IPv7:** `7001`
- **Puerto Web UI IPv7:** `8080` (o `8081`)

Elige la opción que prefieras (te sugiero la **Opción A** si quieres que yo controle y pruebe directamente la ejecución mediante terminal, o la **Opción B/C** si quieres controlarlo tú con interfaz gráfica).
