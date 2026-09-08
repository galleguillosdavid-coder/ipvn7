# Propuestas de Mejora de Arquitectura y Rendimiento del Código

Este documento describe optimizaciones profundas de ingeniería de software para el núcleo (`core/`) y adaptadores (`adapters/`) de IPv7, con el fin de maximizar el throughput, reducir la latencia y garantizar seguridad criptográfica a largo plazo.

---

## 1. Multiplexación Nativa sobre QUIC para Túneles P2P

### Diagnóstico Actual:
El módulo de túneles (`core/tunnel.go`) implementa una capa de multiplexación manual con números de secuencia `StreamID`, comandos `TunnelCmdOpen/Data/Close` y serialización CBOR sobre UDP/QUIC.
- **Problema**: La serialización y deserialización CBOR en cada fragmento de datos TCP añade sobrecarga de CPU y un retardo de procesamiento innecesario cuando se transfieren flujos masivos de datos (ej. descargas HTTP o streams de video).

### Solución Propuesta:
- Aprovechar los streams bidireccionales nativos que ya provee la librería `quic-go` (`quic.Connection.OpenStreamSync()`).
- Cada túnel TCP crea un stream QUIC nativo independiente.
- **Ventajas**:
  - Zero-copy I/O: Uso directo de `io.Copy(quicStream, tcpConn)` sin framing intermedio.
  - Control de congestión y retransmisión por stream: Si un paquete de un túnel se pierde, los demás túneles no se bloquean (eliminación total del bloqueo de cabecera de línea o *Head-of-Line Blocking*).
  - Incremento del rendimiento de transferencia de datos en un **300% a 500%**.

---

## 2. Migración del Handshake al Framework Noise Protocol (Noise_XX)

### Diagnóstico Actual:
El apretón de manos actual (`core/handshake.go`) utiliza un intercambio simple de claves públicas Ed25519 y X25519 con verificación de firma y cálculo de RTT.
- **Limitación**: No implementa secreto hacia adelante perfecto (*Perfect Forward Secrecy - PFS*). Si la clave privada a largo plazo del nodo fuera comprometida en el futuro, las sesiones pasadas grabadas por un adversario pasivo podrían descifrarse.

### Solución Propuesta:
- Adoptar el framework formal **Noise Protocol (Noise_XX)** (el mismo estándar criptográfico utilizado por WireGuard y Lightning Network).
- **Flujo Noise_XX (3 mensajes)**:
  1. `-> e`: El iniciador envía clave efímera.
  2. `<- e, ee, s, es`: El receptor responde con su clave efímera, realiza ECDH efímero-efímero, envía su clave estática cifrada y realiza ECDH efímero-estático.
  3. `-> s, se`: El iniciador envía su clave estática cifrada y completa el intercambio.
- **Garantías**:
  - Autenticación mutua probada formalmente.
  - Perfect Forward Secrecy: Cada sesión utiliza claves simétricas que se destruyen inmediatamente tras el cierre de la conexión.
  - Resistencia contra ataques de repetición (*Replay Attacks*).

---

## 3. Aceleración por Hardware H.264 / VP8 para el Escritorio Remoto Web

### Diagnóstico Actual:
El servicio de escritorio remoto (`core/remotedesktop.go` y `remotedesktop_windows.go`) captura la pantalla mediante GDI y comprime cada fotograma completo como imagen JPEG independiente enviada por WebSocket.
- **Limitación**:
  - Consumo de ancho de banda elevado: A 15 FPS en 1080p, enviar fotogramas JPEG completos consume entre 5 y 15 Mbps de red.
  - No aprovecha las diferencias entre fotogramas (la mayor parte de la pantalla permanece estática cuando el usuario lee o escribe).

### Solución Propuesta:
- Implementar captura con **Desktop Duplication API (DXGI)** en Windows en lugar de BitBlt de GDI (captura directa desde la memoria de la GPU con latencia sub-milisegundo).
- Codificar las diferencias de fotogramas utilizando el estándar WebRTC nativo que ya incluye el proyecto (`adapters/webrtc_adapter.go`) con códec **H.264** o **VP8**.
- **Resultado**:
  - Reducción del uso de ancho de banda en un **80% a 90%** (de 10 Mbps a menos de 1.5 Mbps).
  - Aumento de la fluidez de 15 FPS a **30 o 60 FPS**.

---

## 4. Persistencia Segura de Identidad en Disco (`keystore.go`) [✅ IMPLEMENTADO]

### Estado Actual:
Implementado en [`core/keystore.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/keystore.go) con la función `LoadOrCreatePersistentIdentity`.
- Guarda y recarga automáticamente la semilla Ed25519 en `~/.ipv7/identity.key` (con permisos `0600`).
- Activado mediante el flag `-key persistent` o especificando un archivo personalizado con `-key <ruta>`.
- Permite que el nodo mantenga siempre la misma identidad pública y reputación tras sucesivos reinicios.
- Próxima microfase: almacenamiento de la tabla de peers conocidos en base de datos embebida ligera para reconexión autónoma inmediata.
