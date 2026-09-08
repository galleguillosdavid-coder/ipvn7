Plan de Implementación: Proyecto IPv7
Este documento detalla la ruta de trabajo técnica para construir IPv7 como una red P2P descentralizada overlay. Incorpora las mejoras arquitectónicas propuestas (Shift-Left Security, QUIC temprano, NAT Traversal priorizado) y se enfoca en reutilizar bibliotecas maduras para reducir la complejidad y el código a mantener.

TIP

Recomendación de Lenguaje: Se recomienda encarecidamente utilizar Go (Golang) o Rust para este proyecto. Go tiene un ecosistema inigualable para redes P2P gracias a libp2p, quic-go y el ecosistema de Tailscale (para implementación de DERP/Relay).

User Review Required
IMPORTANT

Decisión de Stack Tecnológico: Este plan asume el uso de Go y bibliotecas del ecosistema libp2p para acelerar el desarrollo. Si prefieres otro lenguaje (ej. Rust, C++, TypeScript), el diseño general se mantiene, pero las bibliotecas recomendadas cambiarán. ¿Estás de acuerdo en basar el MVP en Go?

Open Questions
WARNING

¿Qué tipo de aplicaciones consumirán IPv7 inicialmente? (Ej. mensajería, transferencia de archivos, streaming de video). Esto afectará si priorizamos el rendimiento (UDP) o la fiabilidad (QUIC).
¿Existe alguna restricción de entorno? (Ej. debe correr en navegadores web, dispositivos IoT de bajos recursos, móviles). Si requiere navegadores, WebRTC/WebSockets debe subir en la prioridad.
Proposed Changes
Fase 0 — Contrato Mínimo e Identidad (Criptografía Base)
El objetivo es definir las interfaces abstractas (Core) y establecer que la Identidad es criptográfica desde el inicio.

 Definir interfaces del Core (Container, Identity, Session, Adapter).
 Implementar la Identidad basada en claves públicas (Ej. Ed25519).
Biblioteca sugerida: crypto/ed25519 (Estándar de Go).
 Definir el formato del Container (Unidad binaria de IPv7).
 Implementar codificación determinista estricta.
Biblioteca sugerida: github.com/fxamacker/cbor/v2 (Configurado en modo canónico estricto).
 Añadir firmas digitales al Container para asegurar que el remitente es el dueño de la Identidad.
 Escribir tests unitarios para las interfaces, codificación y verificación de firmas.
Fase 1 — Transporte Base UDP (Best-Effort)
Implementar el primer Adapter utilizando UDP simple. No manejaremos retransmisiones aquí, solo envío de datagramas que caben en el MTU.

 Crear estructura UDPAdapter implementando la interfaz Adapter.
 Implementar listener UDP (net.ListenUDP en Go).
 Implementar envío y recepción de Containers sobre sockets UDP.
 Validar tamaño contra MTU (descartar si es mayor sin fragmentar).
 Escribir tests de Loopback y envío/recepción local.
Fase 2 — NAT Traversal Temprano (STUN)
Antes de intentar P2P real, necesitamos descubrir nuestras IPs públicas y puertos para atravesar el NAT.

 Integrar un cliente STUN para descubrir los endpoints externos (IP:Puerto).
Biblioteca sugerida: github.com/pion/stun.
 Modificar UDPAdapter para que reporte tanto los endpoints locales como los públicos (descubiertos por STUN).
 Pruebas automatizadas conectando dos nodos detrás de NATs permisivos.
Fase 3 — Fiabilidad y Chunks con QUIC
En lugar de reinventar el reensamblaje y retransmisión de chunks sobre UDP, utilizamos QUIC como Adapter principal para comunicaciones fiables.

 Crear estructura QUICAdapter implementando la interfaz Adapter.
 Integrar biblioteca QUIC.
Biblioteca sugerida: github.com/quic-go/quic-go.
 Mapear el concepto de Session de IPv7 a conexiones/streams de QUIC.
 QUIC manejará automáticamente la división de datos grandes (chunks), congestión y retransmisiones.
 Realizar pruebas de envío de archivos grandes (ej. 20MB) sobre QUICAdapter.
Fase 4 — P2P Básico y Sesiones
Crear el gestor que mantiene el estado de con quién estamos conectados y a través de qué Adapters.

 Crear el SessionManager dentro del Core.
 Implementar lógica de reconexión y selección de Adapter (Ej. intentar QUIC primero, caer a UDP si no requiere fiabilidad).
 Implementar una tabla de ruteo local (peers conocidos y sus endpoints).
 Testear conexión directa entre dos nodos usando sus DIDs (Identidades) y endpoints conocidos manualmente.
Fase 5 — Descubrimiento con DHT Seguro
Usar una DHT para encontrar los endpoints (IP:Puerto) de un nodo basándonos únicamente en su Identidad (Clave Pública).

 Integrar una DHT Kademlia.
Biblioteca sugerida: github.com/libp2p/go-libp2p-kad-dht.
 Implementar mitigación Sybil básica: Requerir que el NodeID de la DHT sea un hash válido derivado de la clave pública (Ed25519) del nodo.
 Crear el servicio DHTResolver que traduce Identity -> []Endpoints.
 Pruebas: Iniciar un nodo Bootstrap, conectar Nodos A y B a él, y que A encuentre a B pidiendo a la DHT.
Fase 6 — Fallback Relay (DERP-like)
Para los casos donde el NAT es muy estricto (CGNAT, Firewalls simétricos) y STUN/QUIC fallan, se necesita un Relay.

 Crear estructura RelayAdapter.
 Desplegar o integrar un servidor Relay genérico.
Biblioteca/Inspiración: El servidor y cliente DERP de Tailscale (tailscale.com/derp) es el estándar de oro actual para esto.
 Implementar la lógica de fallback en el Core: Directo (QUIC/UDP) -> Si falla después de X timeout -> Conectar vía Relay.
 Pruebas de conectividad forzando el fallo de las conexiones directas.
Fase 7 — Routing Multi-Salto (Opcional para MVP extendido)
Si el nodo destino no está en conexión directa ni en relay, enrutar a través de otros nodos de la red P2P.

 Implementar métricas de latencia y disponibilidad por enlace.
 Diseñar algoritmo de routing (ej. ruteo cebolla simple o basado en métricas).
 Testear red topológica de 3 nodos (A -> B -> C) donde A y C no se ven.
Fase 8 — Benchmark y Optimización
 Escribir suite de benchmarks para UDPAdapter y QUICAdapter (Throughput y latencia).
 Medir consumo de CPU/Memoria al procesar firmas digitales en cada paquete.
 Optimizar recolección de basura (GC) y asignación de memoria para los Containers.
Verification Plan
Automated Tests
go test -v ./core/...: Pruebas de criptografía, firmas y serialización canónica.
go test -v ./adapters/udp/...: Envío de paquetes de prueba < MTU.
go test -v ./adapters/quic/...: Envío de un archivo generado de 20MB y verificación de checksum SHA256 en destino.
Manual Verification
Compilar el binario del nodo (MVP).
Desplegar un nodo en la nube (AWS/DigitalOcean) actuando como Bootstrap DHT/STUN.
Iniciar el Nodo A en una red Wi-Fi local y el Nodo B en una conexión de datos móvil (4G/5G).
Ejecutar el comando para que el Nodo A envíe un archivo de 20MB al Nodo B usando únicamente el identificador criptográfico del Nodo B.
Verificar que el archivo se transfiere correctamente sorteando los NATs (vía conexión directa QUIC/STUN o cayendo al Relay si es necesario).