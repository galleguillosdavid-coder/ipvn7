# IPv7: Protocolo de Red P2P Descentralizada Overlay

Implementación en **Go** del protocolo IPv7 basado en identidades criptográficas, transporte desacoplado y topología de streaming en cascada.

## 🚀 Arquitectura y Componentes Construidos

1. **Criptografía e Identidad Base (`core/`)**:
   - Identidades lógicas fijas derivadas de pares de claves **Ed25519**.
   - Cada contenedor de datos está serializado de forma canónica determinista (**CBOR**) y firmado digitalmente.
   - Entidad **`Node`** que orquesta llaves, adaptadores de transporte y enrutamiento P2P.

2. **Transporte UDP Best-Effort (`adapters/udp_adapter.go`)**:
   - Transporte rápido no orientado a conexión para mensajes ligeros dentro del MTU seguro.

3. **NAT Traversal Temprano con STUN (`adapters/stun.go`)**:
   - Integración con `github.com/pion/stun/v2`.
   - Detección automática de IPs locales y la **IP pública reflexiva** consultando servidores STUN públicos (ej. Google STUN).

4. **Transporte Fiable y Chunks con QUIC (`adapters/quic_adapter.go`)**:
   - Integración con `github.com/quic-go/quic-go` sobre TLS 1.3 efímero.
   - Permite enviar **archivos y payloads masivos (probado con 5MB+)** sin preocuparse por la fragmentación de paquetes manual.

5. **Streaming en Cascada (`core/cascade.go`)**:
   - Topología de distribución en árbol (*Tree-based Multicast*).
   - Cada nodo tiene un límite de abanico (por defecto **10 dispositivos hijos**).
   - Los datos emitidos por el nodo raíz se retransmiten automáticamente aguas abajo en cascada, permitiendo escalar a miles de receptores sin sobrecargar al emisor original.

6. **Enrutamiento "Mundo Pequeño" - 12 Grados de Separación (`core/smallworld.go`)**:
   - Tablas de enrutamiento **estrictamente acotadas** (máximo 120 peers en memoria: 12 anillos de distancia logarítmica $\times$ 10 peers).
   - Cobertura teórica global de hasta **$10^{12} = 1.000.000.000.000$ dispositivos** (1 billón).
   - Reenvío voraz (*Greedy Routing*) por distancia criptográfica XOR.
   - Campo `HopLimit = 12` con preservación de firma digital de extremo a extremo.
   - Consulta la [Documentación del Modelo de Mundo Pequeño](file:///c:/Users/Frondabrick/Desktop/dvd/Ip/docs/MUNDO_PEQUENO_ROUTING.md).

7. **Descubrimiento con DHT Seguro (`dht/`)**:
   - Mapeo descentralizado de `Identity (Clave Pública) -> []Endpoints (IP:Puerto)`.
   - **Mitigación de ataques Sybil y envenenamiento**: Cada registro en la DHT (`Record`) está firmado con la clave privada de la identidad anunciada. Los nodos intermedios rechazan cualquier intento de falsificación.

8. **Fallback Relay tipo DERP (`adapters/relay_adapter.go`)**:
   - Servidor y cliente de retransmisión para escenarios donde ambos nodos están tras NAT simétrico estricto o CGNAT móvil y el NAT traversal directo falla.
   - El servidor Relay es ciego al contenido: los paquetes viajan firmados y sellados de extremo a extremo sin que el relay pueda espiar ni adulterar los datos.

9. **Cifrado de Extremo a Extremo (E2EE) (`core/e2ee.go`)**:
   - Claves derivadas X25519 a partir de la semilla Ed25519 con Diffie-Hellman (ECDH).
   - Cifrado autenticado de grado militar **ChaCha20-Poly1305** con clave efímera (modelo HPKE / RFC 9180).
   - Ni los nodos intermediarios de la red ni los servidores Relay pueden espiar el contenido de los mensajes.

10. **Adaptador WebRTC Nativo (`adapters/webrtc_adapter.go`)**:
    - Integración completa con `pion/webrtc/v4`.
    - Comunicación P2P nativa mediante `RTCDataChannel` con intercambio de ofertas y respuestas SDP/ICE.

11. **Interfaz Gráfica / Dashboard Web (`ui/`)**:
    - Panel visual interactivo embebido en el nodo accesible en `http://localhost:8080`.
    - Chat visual seguro con botón de cifrado E2EE.
    - **Radar del Mundo Pequeño**: visualizador de los 12 anillos de separación en tiempo real.
    - Monitor del árbol de streaming en cascada con botón de emisión de frames.
    - Binario unificado: **`ipv7-node.exe`** (compilado desde `cmd/node/main.go`).

---

## 🧪 Pruebas Automatizadas

Para ejecutar todas las pruebas unitarias y de integración del proyecto:

```powershell
go test -v ./...
```

---

## 💬 Cómo Probar el Chat P2P (Terminal)

Puedes abrir dos terminales en tu computadora:

### Terminal 1 (Nodo A):
```powershell
.\chat.exe -port 7001
```

### Terminal 2 (Nodo B):
```powershell
.\chat.exe -port 7002 -peer 127.0.0.1:7001
```

---

## 🖥️ Cómo Lanzar el Nodo con Interfaz Gráfica (Web Dashboard)

Para una experiencia visual completa con cifrado E2EE, radar de 12 grados y streaming en cascada:

```powershell
.\ipv7-node.exe -port 7001 -ui 8080
```

Luego abre tu navegador en **`http://localhost:8080`**.
¡Podrás chatear visualmente, ver los peers en el radar orbital y monitorear el streaming en cascada en tiempo real!
