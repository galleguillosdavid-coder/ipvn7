# Reglas de Desarrollo y Arquitectura para IPv7

Este conjunto de directrices rige las prácticas de ingeniería, arquitectura y estilo para el desarrollo del protocolo IPv7 dentro del IDE Antigravity.

## 1. Principios Criptográficos e Identidad
- **Identidad Criptográfica**: Cada nodo posee una identidad lógica derivada estrictamente de un par de claves **Ed25519**. No se admiten identificadores arbitrarios o no firmados.
- **Cifrado E2EE**: Toda comunicación confidencial debe utilizar derivación **X25519** mediante Diffie-Hellman (ECDH) y cifrado autenticado **ChaCha20-Poly1305** con clave de sesión efímera.
- **Canonicidad CBOR**: Todos los contenedores de red deben serializarse utilizando codificación CBOR canónica determinista (`fxamacker/cbor/v2`) antes de computar cualquier firma digital.

## 2. Concurrencia y Red en Go
- **Goroutines Seguras**: Toda goroutine de escucha de sockets (UDP, QUIC, Relay, WebRTC) debe respetar contextos (`context.Context`) o canales de parada (`close(stopCh)`) para evitar fugas de memoria.
- **Protección de Estado (`sync.RWMutex`)**: Las tablas de peers, tablas de enrutamiento del Mundo Pequeño (12 anillos) y sockets activos deben protegerse con candados de lectura/escritura granulares.
- **No Bloquear el Hilo Principal de Paquetes**: Las operaciones pesadas (verificación de firmas, cómputo XOR, búsquedas en DHT) deben despacharse a pools o callbacks no bloqueantes.

## 3. Topología de Red y Enrutamiento
- **Mundo Pequeño (12 Grados de Separación)**:
  - Estricto límite de 120 peers en memoria (12 anillos logarítmicos $\times$ 10 peers).
  - Enrutamiento voraz (*Greedy Routing*) por distancia XOR.
  - `HopLimit` máximo de 12 saltos.
- **Streaming en Cascada**:
  - Topología en árbol con fan-out de 10 nodos hijos por defecto.
  - Reemisión aguas abajo automática sin duplicar buffers.
- **Prioridad de Adapters (Shift-Left)**:
  - Intentar QUIC / UDP directo con STUN reflexivo.
  - Caer a Relay (DERP-like) únicamente si el NAT transversal falla (ambos tras NAT simétrico).

## 4. Flujo de Trabajo WSL2 + Windows
- **Compilación Linux**: Usar `.\scripts\build_linux.ps1` para generar binarios en `bin/`.
- **Supervisión en WSL2**: Usar `.\scripts\run_wsl.ps1 -Port <P> -UIPort <UI>` para supervisar la ejecución y consultar los logs en `logs/wsl_node.log`.
- **Pruebas de Malla Cruzada**: Usar `.\scripts\dual_node_test.ps1` para validar la interacción entre host Windows y Linux WSL2.

## 5. Consultas de Grafo con Kùzu CLI
- Para consultar la base de datos de grafo de código o de red, usar:
  ```powershell
  .\tools\kuzu\kuzu.exe .kuzu_index/
  ```
  O en WSL2:
  ```bash
  ./tools/kuzu/kuzu .kuzu_index/
  ```
