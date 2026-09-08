# Capítulo 4: Casos de Uso Prácticos de IPv7 Paso a Paso

En este capítulo analizamos 7 casos de uso reales de IPv7, con instrucciones exactas, comandos y diagramas de flujo para que cualquier operador pueda ejecutarlos.

---

## Caso 1: Mensajería Instantánea Privada P2P (Chat E2EE)

### Objetivo:
Comunicarte confidencialmente con otra persona sin que ninguna empresa (Google, Meta, etc.) ni proveedor de Internet pueda leer, registrar metadatos ni bloquear tus mensajes.

### Paso a paso:
1. **Lanzar el nodo de mensajería (Alice)**:
   ```powershell
   .\ipv7-chat.exe -port 7001
   ```
   Alice verá su clave pública Ed25519 en consola (ej: `a1b2c3...`).
2. **Lanzar el nodo de mensajería (Bob)**:
   ```powershell
   .\ipv7-chat.exe -port 7002
   ```
   Bob obtiene su clave pública (ej: `d4e5f6...`).
3. **Conectar**:
   Bob escribe en su consola:
   ```text
   /connect 192.168.1.45:7001
   ```
4. **Verificación**:
   Ambos nodos realizan un apretón de manos (*Handshake*) criptográfico automático:
   - Intercambian sus claves Ed25519 y derivan sus claves X25519.
   - Todo mensaje posterior muestra la etiqueta `[SENT [🔒 E2EE]]` y `[FROM ... [🔒 E2EE]]`.

---

## Caso 2: Comunicación a través de NAT Simétrico y Redes Celulares (CGNAT)

### Objetivo:
Conectar dos computadoras donde ambas están detrás de routers domésticos estrictos o redes móviles 4G/5G que impiden conexiones entrantes directas.

### Mecanismo de Resolución:
1. **Fase 1 (STUN Traversal)**:
   Al iniciar el nodo, consulta `stun.l.google.com:19302`.
   - STUN le responde cuál es su **IP pública reflexiva** y su puerto mapeado en Internet.
2. **Fase 2 (Intento Directo QUIC / UDP)**:
   Los nodos intentan hacer "hole punching" (perforación de NAT) enviando paquetes sincronizados.
3. **Fase 3 (Fallback Automático a Relay DERP)**:
   Si ambos routers usan NAT simétrico (puerto aleatorio en cada destino) y el hole punching falla:
   - El nodo conmuta de forma transparente al adaptador de Relay (`adapters/relay_adapter.go`).
   - El servidor Relay simplemente retransmite los paquetes encriptados. **El servidor Relay no puede descifrar nada**, ya que los datos viajan sellados con la clave ChaCha20-Poly1305 de extremo a extremo.

---

## Caso 3: Distribución Masiva en Cascada (Fan-Out 10)

### Objetivo:
Transmitir un flujo continuo de datos (telemetría de sensores, cotizaciones financieras o frames de video) a 1.000 clientes sin saturar el ancho de banda del servidor emisor.

### Arquitectura en Acción:
- El nodo emisor (`CascadeNode`) solo envía los paquetes a 10 nodos de primer nivel.
- Cada uno de los 10 nodos retransmite a 10 nodos hijos de segundo nivel (100 receptores).
- Los de segundo nivel retransmiten al tercer nivel (1.000 receptores).
- **Consumo de ancho de banda del emisor**: Constante e idéntico para 10 personas que para 10.000 personas.

```
                  [ Emisor Raíz ]
                  /     |      \     (Envía solo a 10)
                 v      v       v
             [Hijo 1] [Hijo 2] [Hijo 10]
             /   \     /   \     /   \   (Cada uno retransmite a 10)
            v     v   v     v   v     v
          [100 Receptores en Nivel 2]
```

### Ejecución:
En el Dashboard Web (`http://localhost:8080`), pulsa el botón **"Emitir Frame de Telemetría"**. Verás cómo la carga útil se firma canónicamente en CBOR y se propaga en cascada instantáneamente.

---

## Caso 4: Reenvío de Puertos P2P (Port Forwarding sin abrir puertos en el Router)

### Objetivo:
Acceder al servidor SSH de tu PC de casa (puerto 22) o a una base de datos PostgreSQL (puerto 5432) desde tu laptop de viaje, sin tener IP pública fija ni configurar reglas de Port Forwarding en tu router.

### Configuración:
1. En la máquina de destino (Servidor con SSH en puerto 22):
   - Inicia `ipv7-node.exe`.
   - Anota su clave pública de identidad (ej: `fe44a98...`).
2. En tu laptop cliente:
   - Inicia `ipv7-node.exe`.
   - Abre el Dashboard Web o envía una petición POST a la API:
     ```bash
     curl -X POST http://localhost:8080/api/tunnel/start \
       -H "Content-Type: application/json" \
       -d '{"local_port": 2222, "peer_id": "fe44a98...", "target_port": 22}'
     ```
3. **Conexión**:
   En tu laptop, simplemente te conectas a tu propio puerto local:
   ```bash
   ssh usuario@127.0.0.1 -p 2222
   ```
   Todo el tráfico TCP entre tu cliente SSH y el servidor SSH viajará multiplexado y cifrado a través de los túneles P2P de IPv7.

---

## Caso 5: Control de Escritorio Remoto Web P2P (Sin Software Privativo)

### Objetivo:
Brindar asistencia remota técnica a un familiar o acceder a tu PC de escritorio sin instalar AnyDesk, TeamViewer o pagar licencias de software comercial.

### Funcionamiento:
1. En la máquina a controlar (Host):
   - Iniciar `ipv7-node.exe`.
   - El servicio de escritorio remoto captura la pantalla en búfer nativo a 15 FPS en formato JPEG liviano.
2. En la máquina del operador:
   - Accede a la URL del nodo host en el navegador: `http://<IP_HOST>:8080`.
   - Entra a la sección **"Escritorio Remoto"**.
   - Al mover el cursor o hacer clic sobre el canvas web, los eventos se transmiten por WebSockets (`/ws/desktop`) y el nodo host inyecta los movimientos y clics nativos en el sistema operativo.

---

## Caso 6: Navegación Segura con VPN SOCKS5 en Espacio de Usuario

### Objetivo:
Navegar por Internet desde una red Wi-Fi pública insegura (aeropuerto, hotel) canalizando todo el tráfico web a través de tu nodo hogareño seguro.

### Configuración:
1. En tu nodo de confianza (casa/oficina):
   - Inicia el nodo IPv7.
2. En tu laptop de viaje:
   - En el Dashboard Web, presiona **"Iniciar VPN SOCKS5"** o haz POST a `/api/vpn/start` con `{"port": 1080}`.
   - Configura el proxy de tu navegador (Firefox, Chrome) o del sistema operativo en:
     - **Protocolo**: SOCKS v5
     - **Host**: `127.0.0.1`
     - **Puerto**: `1080`
   - Toda tu navegación web se envía cifrada por la malla IPv7 hasta el nodo de salida antes de salir a Internet.

---

## Caso 7: Prueba de Laboratorio Malla Dual (Windows Host ↔ WSL2 Linux)

### Objetivo:
Validar de forma automatizada y en una sola máquina física que la arquitectura de comunicación cruzada entre dos sistemas operativos distintos (Windows y Linux) funciona a la perfección.

### Ejecución:
Ejecuta el script PowerShell incluido en el repositorio:
```powershell
.\scripts\dual_node_test.ps1
```

### Qué hace este script de forma automática:
1. Compila el binario para Linux en `bin/ipv7-node-linux`.
2. Inicia el **Nodo A** en Windows (Puertos: UDP `7001`, Web `8080`).
3. Inicia el **Nodo B** en WSL2 Ubuntu (Puertos: UDP `7002`, Web `8082`).
4. Conecta el Nodo B al Nodo A ejecutando un apretón de manos criptográfico Ed25519.
5. Abre ambos dashboards en tu navegador para que observes los nodos vinculados en el gráfico de malla.

---

## Caso 8: Operación Automatizada con Asistentes de IA (Model Context Protocol - MCP)

### Objetivo:
Permitir que un asistente de IA (como Claude Desktop, Antigravity o Cursor) opere tu nodo IPv7 de forma autónoma: consultar peers, mandar mensajes cifrados, verificar latencias y abrir túneles usando lenguaje natural.

### Configuración en Claude Desktop o Antigravity:
En el archivo de configuración `claude_desktop_config.json` o en la configuración de herramientas de tu agente:
```json
{
  "mcpServers": {
    "ipv7": {
      "command": "C:\\Users\\Frondabrick\\Desktop\\dvd\\Ipv7\\ipv7-node.exe",
      "args": ["-mcp", "-key", "persistent"]
    }
  }
}
```

### Uso en lenguaje natural:
- *"IA, consulta qué peers están en línea en mi nodo IPv7."* -> La IA invoca `ipv7_list_peers`.
- *"IA, envíale un mensaje cifrado a fe44a98... diciendo 'El servidor está listo'."* -> La IA invoca `ipv7_send_message`.

---

## Caso 9: Identidad Persistente y Mapeo UPnP en Router Hogareño

### Objetivo:
Tener un nodo permanente que conserve su clave pública para siempre, de modo que tus amigos o servidores siempre puedan encontrarte a la misma dirección, mientras el router abre el puerto automáticamente sin tocar su configuración web.

### Puesta en marcha:
```powershell
.\ipv7-node.exe -key persistent -upnp=true
```
- Tu clave Ed25519 se almacena de forma segura en `~/.ipv7/identity.key`.
- El adaptador UPnP negocia con tu router residencial (vía SSDP multicast) la apertura del puerto UDP 7001, logrando latencias óptimas directas sin relays.
