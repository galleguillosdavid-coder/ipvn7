# Capítulo 3: Guía del Dashboard Web Interactivo de IPv7

El Dashboard Web de IPv7 es una interfaz gráfica moderna de una sola página (SPA) que se ejecuta localmente dentro de cada nodo en `http://localhost:8080`. Proporciona observabilidad total de la red, herramientas criptográficas y utilidades de comunicación.

---

## 1. Encabezado y Estado de Identidad

En la parte superior del panel encontrarás:
- **Tu Identidad Criptográfica (Ed25519 Public Key)**: Representa tu dirección inmutable en la red. Puedes hacer clic en el botón de copiar para compartirla con otros usuarios.
- **Clave de Cifrado X25519 (E2EE Key)**: Clave pública asimétrica derivada utilizada para cifrar de extremo a extremo los mensajes directos.
- **Contador de Peers Activos**: Muestra cuántos nodos están conectados y registrados en la tabla de enrutamiento del Mundo Pequeño.
- **Contador de Suscriptores en Cascada**: Nodos hijos que reciben retransmisión de streaming de tu nodo.

---

## 2. Visualizador de Topología de Malla (Mesh Graph)

Un lienzo interactivo con física de partículas que renderiza en tiempo real la topología de la red P2P:
- **Nodo Central Azul/Dorado**: Representa tu propio nodo local.
- **Nodos Periféricos**: Representan peers descubiertos y conectados.
  - El tamaño del nodo indica su **grado de conectividad** (número de enlaces).
- **Líneas de Enlace (Links)**:
  - Color verde: Enlace de baja latencia (< 30 ms).
  - Color ámbar/rojo: Enlace transcontinental o con mayor latencia.
  - Etiqueta de protocolo: Muestra si el enlace viaja sobre `QUIC` o `UDP`.
- **Botón "Exportar Cypher Kùzu"**:
  - Genera al instante el script Cypher de la topología para auditar la red en la base de datos de grafos Kùzu.

---

## 3. Radar del Mundo Pequeño (12 Anillos de Separación)

Una representación concéntrica visual única de la arquitectura de enrutamiento de IPv7:
- **Centro**: Tu nodo.
- **12 Anillos Concéntricos**: Cada anillo representa un orden de magnitud de distancia criptográfica XOR respecto a tu clave pública.
  - **Anillo 1 (Cercano)**: Peers matemáticamente vecinos en el espacio de claves.
  - **Anillos 6-12 (Lejanos)**: Peers geográficamente o matemáticamente dispersos que garantizan atajos de enrutamiento global (*long-range contacts*).
- **Indicador de Salud**: Te permite verificar de un vistazo que tu nodo tiene cobertura equilibrada en múltiples anillos para garantizar que cualquier mensaje llegue en menos de 12 saltos.

---

## 4. Consola de Mensajería y Chat E2EE

Permite entablar conversaciones directas con cualquier peer de la red sin pasar por ningún servidor central:
- **Campo de Destinatario**: Introduce la clave pública Ed25519 del destinatario o su dirección física (`IP:Puerto`).
- **Interruptor de Cifrado E2EE (🔒)**:
  - Al estar activado, el mensaje se sella con ChaCha20-Poly1305 antes de salir del adaptador de red.
  - La etiqueta `[🔒 E2EE]` confirma en pantalla que el mensaje fue descifrado exitosamente con tu clave privada.
- **Historial en Vivo**: Los mensajes recibidos se actualizan instantáneamente a través de WebSockets sin recargar la página.
- **Transferencia de Archivos Drag & Drop**:
  - Puedes arrastrar cualquier archivo (imágenes, documentos, zips de hasta 2MB en demo web) directamente sobre la ventana del chat.
  - Alternativamente, pulsa el botón del clip (📎) para abrir el explorador de archivos.
  - Las imágenes se muestran con vista previa integrada directamente en la burbuja de chat y cualquier archivo cuenta con botón de descarga inmediata (`⬇️ Descargar`).
  - Todo el contenido de los archivos viaja sellado con cifrado E2EE.

---

## 5. Observabilidad Viva: OpenAPI 3.1 y Métricas Prometheus

En la barra superior de navegación dispones de dos accesos directos de observabilidad:
- **Botón `📜 OpenAPI`**: Abre la especificación dinámica en `http://localhost:8080/api/openapi.json`, permitiendo auditar todos los contratos y esquemas REST.
- **Botón `📈 Métricas`**: Abre el endpoint `http://localhost:8080/metrics` en formato estándar de Prometheus, reportando contadores de peers, estado de túneles y actividad del nodo en tiempo real.

---

## 5. Monitor de Streaming en Cascada (Fan-Out 10)

Diseñado para probar la distribución masiva de datos en árbol:
- **Botón "Emitir Frame de Telemetría"**: Genera un paquete CBOR firmado canónicamente con un hash SHA-256 de carga útil y lo difunde a los 10 nodos hijos de la cascada.
- **Panel de Flujo**: Muestra los identificadores de stream, secuencias numéricas y tamaño de bytes recibidos por retransmisión descendente.

---

## 6. Centro de Túneles P2P y VPN SOCKS5

Permite convertir a IPv7 en una red privada virtual de grado empresarial:
- **Túnel P2P (Port Forwarding)**:
  - Formulario para asociar un puerto local (`Local Port`), la clave del peer remoto (`Peer ID`) y el puerto de destino (`Target Port`).
  - Ejemplo: `Local Port: 2222` -> `Peer Remoto` -> `Target Port: 22 (SSH)`.
- **Botón "Iniciar VPN SOCKS5"**:
  - Levanta un proxy local en el puerto `1080` (o el que configures) que canaliza todo el tráfico web a través de la malla encriptada.

---

## 7. Pantalla de Escritorio Remoto Web (Web Remote Desktop)

Permite visualizar y controlar otra PC remotamente a través de un canvas web:
- Conexión fluida a 15 FPS utilizando streaming JPEG por WebSockets bidireccionales.
- Inyección de eventos de ratón (movimiento, clic izquierdo, medio, derecho y rueda de scroll).
- Inyección de teclas del teclado en tiempo real.
- **Sin instalación de drivers externos**: Funciona nativamente con las APIs del sistema operativo.
