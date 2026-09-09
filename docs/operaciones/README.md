# Manual de Operaciones para Humanos: Ecosistema IPv7

¡Bienvenido al **Manual de Operaciones de IPv7**! Este compendio está redactado para operadores, ingenieros y usuarios que desean comprender, desplegar, utilizar y resolver incidentes con el protocolo de red descentralizada P2P **IPv7**.

---

## 📚 Mapa de Navegación del Manual

Este manual está dividido en 5 módulos progresivos:

1. [**01. Conceptos y Fundamentos para Humanos**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/01_conceptos_y_fundamentos_para_humanos.md)
   - ¿Qué es IPv7 y por qué surge?
   - Diferencias fundamentales entre IPv4/IPv6 (capa 3) e IPv7 (capa 7 superpuesta).
   - Identidades criptográficas Ed25519 vs IPs tradicionales.
   - Seguridad matemática: Cifrado E2EE ChaCha20-Poly1305 y X25519.
   - Modelo de enrutamiento "Mundo Pequeño" (12 grados de separación).

2. [**02. Instalación y Puesta en Marcha**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/02_instalacion_puesta_en_marcha.md)
   - Requisitos del sistema (Windows 10/11, Linux nativo, WSL2 Ubuntu).
   - Compilación desde código fuente y uso de binarios listos (`ipv7-node.exe`, `ipv7-chat.exe`).
   - Flags y opciones de línea de comandos (`-port`, `-ui`, `-peer`, `-stun`, `-open`).
   - Gestión inteligente de puertos y permisos de Firewall.
   - Orquestación en WSL2 con scripts automatizados.

3. [**03. Guía del Dashboard Web Interactivo**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/03_guia_dashboard_web.md)
   - Acceso al panel (`http://localhost:8080`).
   - Panel de Estado e Identidad criptográfica.
   - Visualizador de Topología de Malla (Mesh Graph) interactivo con partículas.
   - Radar del Mundo Pequeño (12 anillos logarítmicos de distancia XOR).
   - Consola de Mensajería Instantánea E2EE en tiempo real.
   - Monitor de Cascada de Streaming (Fan-out 10).
   - Centro de Túneles y VPN SOCKS5.
   - Pantalla de Escritorio Remoto Web P2P.

4. [**04. Casos de Uso Detallados Paso a Paso**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/04_casos_de_uso_detallados.md)
   - **Caso 1:** Mensajería instantánea privada P2P sin servidores centrales.
   - **Caso 2:** Traspaso de NAT simétrico y redes celulares (CGNAT) con STUN y Relay.
   - **Caso 3:** Distribución de telemetría y datos en Cascada a miles de receptores.
   - **Caso 4:** Reenvío de puertos P2P (Port Forwarding) para SSH y Web local sin abrir puertos en el router.
   - **Caso 5:** Soporte y Control de Escritorio Remoto Web P2P sin software privativo (reemplazo AnyDesk/TeamViewer).
   - **Caso 6:** Navegación anónima/segura mediante VPN SOCKS5 en espacio de usuario.
   - **Caso 7:** Prueba de laboratorio en Malla Dual (Windows Host ↔ WSL2 Linux).
   - **Caso 8:** Operación automatizada con Asistentes de IA mediante Servidor MCP (`-mcp`).
   - **Caso 9:** Identidad fija persistente con Keystore (`-key persistent`) y mapeo UPnP.

5. [**05. Resolución de Problemas y Preguntas Frecuentes (FAQ)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/05_resolucion_problemas_y_faq.md)
   - Diagnóstico rápido ante fallos de conexión.
   - ¿Qué hacer si STUN no descubre mi IP pública?
   - Resolución de conflictos de puertos ocupados.
   - Auditoría de logs de ejecución (`logs/wsl_node.log`).
   - Inspección con base de datos de grafos Kùzu CLI.
   - Preguntas frecuentes operativas y de seguridad.

---

## ⚡ Comandos Rápidos de Supervivencia

| Tarea Operativa | Comando en Windows (PowerShell) | Comando en Linux / WSL2 |
|---|---|---|
| Iniciar nodo estándar con UI web | `.\ipv7-node.exe` | `./bin/ipv7-node-linux` |
| Iniciar con identidad fija persistente | `.\ipv7-node.exe -key persistent` | `./bin/ipv7-node-linux -key persistent` |
| Iniciar como servidor MCP para IA | `.\ipv7-node.exe -mcp -key persistent` | `./bin/ipv7-node-linux -mcp -key persistent` |
| Iniciar en puertos específicos | `.\ipv7-node.exe -port 7001 -ui 8080` | `./bin/ipv7-node-linux -port 7002 -ui 8082` |
| Conectar a un nodo inicial conocido | `.\ipv7-node.exe -peer 192.168.1.50:7001` | `./bin/ipv7-node-linux -peer 10.0.0.5:7001` |
| Iniciar chat por terminal CLI | `.\ipv7-chat.exe -peer 127.0.0.1:7001` | `./bin/chat-linux -peer 127.0.0.1:7001` |
| Ejecutar batería de tests | `go test -v ./...` | `go test -v ./...` |
| Lanzar nodo dual de prueba | `.\scripts\dual_node_test.ps1` | N/A (invocar desde PowerShell) |

---

> [!TIP]
> Si es tu primera vez utilizando IPv7, comienza por el [Capítulo 01](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/01_conceptos_y_fundamentos_para_humanos.md) para entender cómo tus claves criptográficas reemplazan a las direcciones IP numéricas.
