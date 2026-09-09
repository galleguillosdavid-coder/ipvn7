# Índice Integral del Proyecto IPv7

Este documento cataloga de forma exhaustiva todos los componentes, archivos, tipos, funciones y herramientas del protocolo IPv7. Generado automáticamente por el indexador semántico.

## Resumen de Paquetes

| Paquete | Archivos | Líneas de Código (LoC) | Propósito Principal |
|---|---|---|---|
| [`adapters`](#paquete-adapters) | 15 | 1923 | Adaptadores de transporte desacoplados: UDP best-effort, QUIC sobre TLS 1.3, NAT Traversal con STUN, Relay seguro tipo DERP y WebRTC. |
| [`core`](#paquete-core) | 36 | 5368 | Núcleo de protocolo: Identidad Ed25519, Contenedores CBOR, Cifrado E2EE, Mundo Pequeño (12 Grados), Streaming en Cascada y Orquestador de Nodos. |
| [`dht`](#paquete-dht) | 4 | 434 | Tabla Hash Distribuida Kademlia segura para resolución descentralizada Identity -> Endpoints con firmas digitales. |
| [`main`](#paquete-main) | 2 | 472 | Puntos de entrada ejecutables: nodo P2P (`cmd/node`) y cliente de chat CLI (`cmd/chat`). |
| [`ui`](#paquete-ui) | 4 | 1704 | Dashboard web SPA interactivo, servidor HTTP/WebSocket, Chat E2EE y monitor de topología en malla. |

---

## Paquete `adapters`

Adaptadores de transporte desacoplados: UDP best-effort, QUIC sobre TLS 1.3, NAT Traversal con STUN, Relay seguro tipo DERP y WebRTC.

### 📄 [quic_adapter.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/quic_adapter.go) (209 LoC)
**Tipos y Estructuras:**
- `struct QUICAdapter`
**Funciones Clave:**
- `LocalAddr()`
- `NewQUICAdapter()`
- `Receive()`
- `Send()`
- `Start()`
- `Stop()`

### 📄 [quic_adapter_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/quic_adapter_test.go) (107 LoC)
**Funciones Clave:**
- `TestQUICAdapter()`
- `TestQUICAdapterLargePayload()`

### 📄 [relay_adapter.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/relay_adapter.go) (280 LoC)
**Tipos y Estructuras:**
- `struct RelayServer`
- `struct RelayAdapter`
**Funciones Clave:**
- `Addr()`
- `NewRelayAdapter()`
- `NewRelayServer()`
- `Receive()`
- `Send()`
- `Start()`
- `Stop()`

### 📄 [relay_adapter_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/relay_adapter_test.go) (70 LoC)
**Funciones Clave:**
- `TestRelayServerAndAdapter()`

### 📄 [replay_filter.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/replay_filter.go) (87 LoC)
**Tipos y Estructuras:**
- `struct ReplayWindow`
- `struct AntiReplayTable`
**Funciones Clave:**
- `CheckAndSet()`
- `NewAntiReplayTable()`
- `Reset()`

### 📄 [replay_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/replay_test.go) (120 LoC)
**Funciones Clave:**
- `TestAntiReplayTablePerPeerIsolation()`
- `TestReplayWindowDuplicatesRejected()`
- `TestReplayWindowOutOfOrderWithinWindow()`
- `TestReplayWindowSequential()`
- `TestReplayWindowTooOldRejected()`

### 📄 [stun.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/stun.go) (94 LoC)
**Funciones Clave:**
- `DiscoverPublicEndpoint()`
- `GetLocalEndpoints()`

### 📄 [stun_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/stun_test.go) (42 LoC)
**Funciones Clave:**
- `TestDiscoverPublicEndpoint()`
- `TestGetLocalEndpoints()`
- `TestUDPAdapterDiscoverEndpoints()`

### 📄 [tls_helper.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/tls_helper.go) (47 LoC)
**Funciones Clave:**
- `GenerateTLSConfig()`

### 📄 [udp_adapter.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/udp_adapter.go) (238 LoC)
**Tipos y Estructuras:**
- `struct UDPAdapter`
**Funciones Clave:**
- `DiscoverEndpoints()`
- `Endpoints()`
- `NewUDPAdapter()`
- `Receive()`
- `Send()`
- `Start()`
- `Stop()`

### 📄 [udp_adapter_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/udp_adapter_test.go) (79 LoC)
**Funciones Clave:**
- `TestUDPAdapter()`
- `TestUDPAdapterMTU()`

### 📄 [upnp.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/upnp.go) (185 LoC)
**Tipos y Estructuras:**
- `struct UPnPMapper`
**Funciones Clave:**
- `DiscoverAndForward()`
- `NewUPnPMapper()`

### 📄 [upnp_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/upnp_test.go) (19 LoC)
**Funciones Clave:**
- `TestUPnPMapperGracefulTimeout()`

### 📄 [webrtc_adapter.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/webrtc_adapter.go) (249 LoC)
**Tipos y Estructuras:**
- `struct WebRTCAdapter`
**Funciones Clave:**
- `AcceptAnswer()`
- `AcceptOffer()`
- `CreateDataChannel()`
- `CreateOffer()`
- `NewWebRTCAdapter()`
- `Receive()`
- `Send()`
- `Start()`
- `Stop()`

### 📄 [webrtc_adapter_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/webrtc_adapter_test.go) (97 LoC)
**Funciones Clave:**
- `TestWebRTCAdapterCommunication()`

## Paquete `core`

Núcleo de protocolo: Identidad Ed25519, Contenedores CBOR, Cifrado E2EE, Mundo Pequeño (12 Grados), Streaming en Cascada y Orquestador de Nodos.

### 📄 [beacon.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/beacon.go) (236 LoC)
**Tipos y Estructuras:**
- `struct BeaconPacket`
- `struct BeaconService`
**Funciones Clave:**
- `NewBeaconService()`
- `Start()`
- `Stop()`

### 📄 [benchmark_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/benchmark_test.go) (134 LoC)
**Funciones Clave:**
- `BenchmarkContainerSerialization()`
- `BenchmarkContainerSign()`
- `BenchmarkContainerVerify()`
- `BenchmarkE2EEEncryptDecrypt()`
- `BenchmarkSmallWorldLookup()`

### 📄 [cascade.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/cascade.go) (132 LoC)
**Tipos y Estructuras:**
- `struct StreamChunk`
- `struct CascadeNode`
**Funciones Clave:**
- `AddChild()`
- `Broadcast()`
- `ChildrenCount()`
- `NewCascadeNode()`
- `OnChunk()`
- `SetParent()`

### 📄 [cascade_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/cascade_test.go) (110 LoC)
**Funciones Clave:**
- `TestCascadeMaxChildrenEnforcement()`
- `TestCascadeStreamingMultiHop()`

### 📄 [container.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/container.go) (91 LoC)
**Tipos y Estructuras:**
- `struct Container`
**Funciones Clave:**
- `Marshal()`
- `Sign()`
- `Unmarshal()`
- `Verify()`

### 📄 [container_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/container_test.go) (78 LoC)
**Funciones Clave:**
- `TestContainerSerialization()`
- `TestContainerSigningAndVerification()`
- `TestGenerateIdentity()`

### 📄 [discovery_firebase.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/discovery_firebase.go) (306 LoC)
**Tipos y Estructuras:**
- `struct FirebasePeerRecord`
- `struct FirebaseDiscovery`
**Funciones Clave:**
- `Announce()`
- `AnnounceAndDiscover()`
- `Deregister()`
- `FetchPeers()`
- `NewFirebaseDiscovery()`
- `ResolveDID()`
- `SetEndpointRefresher()`
- `Start()`
- `Stop()`

### 📄 [discovery_firebase_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/discovery_firebase_test.go) (67 LoC)
**Funciones Clave:**
- `TestFirebaseDiscoveryAnnounceAndFetch()`

### 📄 [e2ee.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/e2ee.go) (159 LoC)
**Funciones Clave:**
- `DecryptE2EE()`
- `DeriveX25519FromSeed()`
- `EncryptE2EE()`
- `GenerateE2EEKeyPair()`

### 📄 [e2ee_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/e2ee_test.go) (69 LoC)
**Funciones Clave:**
- `TestE2EEDeriveFromSeed()`
- `TestE2EEEncryptionDecryption()`
- `TestE2EETamperRejection()`

### 📄 [handshake.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/handshake.go) (88 LoC)
**Tipos y Estructuras:**
- `struct HandshakePayload`
- `struct PingPayload`
**Funciones Clave:**
- `DecodeHandshake()`
- `DecodePing()`
- `EncodeHandshake()`
- `EncodePing()`
- `GenerateNonce()`
- `IsFresh()`

### 📄 [handshake_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/handshake_test.go) (62 LoC)
**Funciones Clave:**
- `TestHandshakePayloadSerialization()`
- `TestPingPayloadSerialization()`

### 📄 [identity.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/identity.go) (51 LoC)
**Tipos y Estructuras:**
- `struct Ed25519Identity`
**Funciones Clave:**
- `Bytes()`
- `GenerateIdentity()`
- `NewIdentityFromBytes()`
- `NewIdentityFromHex()`
- `String()`
- `Verify()`

### 📄 [interfaces.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/interfaces.go) (34 LoC)
**Tipos y Estructuras:**
- `interface Identity`
- `interface Session`
- `interface Adapter`

### 📄 [keystore.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/keystore.go) (67 LoC)
**Funciones Clave:**
- `DefaultKeyDir()`
- `LoadOrCreatePersistentIdentity()`

### 📄 [keystore_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/keystore_test.go) (39 LoC)
**Funciones Clave:**
- `TestKeystorePersistence()`

### 📄 [logger.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/logger.go) (50 LoC)
**Funciones Clave:**
- `InitLogger()`
- `Log()`

### 📄 [logger_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/logger_test.go) (24 LoC)
**Funciones Clave:**
- `TestStructuredLogger()`

### 📄 [mcp.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/mcp.go) (440 LoC)
**Tipos y Estructuras:**
- `struct MCPRequest`
- `struct MCPResponse`
- `struct MCPError`
- `struct MCPServer`
**Funciones Clave:**
- `Dispatch()`
- `HandleMessage()`
- `NewMCPServer()`
- `ServeStdio()`
- `SetCypherExecutor()`
- `SetVPNProxy()`

### 📄 [mcp_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/mcp_test.go) (144 LoC)
**Funciones Clave:**
- `TestMCPServer()`

### 📄 [node.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/node.go) (612 LoC)
**Tipos y Estructuras:**
- `struct Node`
**Funciones Clave:**
- `AddAdapter()`
- `AddPeer()`
- `AddPeerWithLatency()`
- `DecryptMessage()`
- `Endpoints()`
- `GetPeerEncKey()`
- `GetPeerEndpoints()`
- `Handshake()`
- `NewNode()`
- `OnMessage()`
- `PingPeer()`
- `SendEncryptedMessage()`
- `SendMessage()`
- `SetDIDResolver()`
- `SetEndpoints()`
- `SetPeerEncKey()`
- `Start()`
- `Stats()`
- `Stop()`

### 📄 [node_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/node_test.go) (165 LoC)
**Funciones Clave:**
- `Receive()`
- `Send()`
- `Start()`
- `Stop()`
- `TestNodeHandshakeAndE2EE()`
- `TestNodeP2PCommunication()`

### 📄 [noise.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/noise.go) (365 LoC)
**Tipos y Estructuras:**
- `struct NoiseSession`
- `struct NoiseMsg1`
- `struct NoiseMsg2`
- `struct NoiseMsg3`
- `struct NoiseHandshakeState`
**Funciones Clave:**
- `Decrypt()`
- `Encrypt()`
- `InitiatorStep1()`
- `InitiatorStep3()`
- `NewNoiseHandshake()`
- `ResponderFinal()`
- `ResponderStep2()`

### 📄 [noise_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/noise_test.go) (135 LoC)
**Funciones Clave:**
- `TestNoiseHandshakeAndSession()`
- `TestNoiseTamperedSignature()`

### 📄 [remotedesktop.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/remotedesktop.go) (72 LoC)
**Tipos y Estructuras:**
- `struct RemoteInputEvent`
- `struct RemoteDesktopService`
**Funciones Clave:**
- `GetResolution()`
- `NewRemoteDesktopService()`
- `StartStreaming()`

### 📄 [remotedesktop_other.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/remotedesktop_other.go) (47 LoC)
**Funciones Clave:**
- `CaptureFrame()`
- `InjectInput()`

### 📄 [remotedesktop_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/remotedesktop_test.go) (34 LoC)
**Funciones Clave:**
- `TestRemoteDesktopCapture()`

### 📄 [remotedesktop_windows.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/remotedesktop_windows.go) (284 LoC)
**Funciones Clave:**
- `CaptureFrame()`
- `InjectInput()`

### 📄 [self_healing.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/self_healing.go) (157 LoC)
**Tipos y Estructuras:**
- `struct SelfHealingSupervisor`
**Funciones Clave:**
- `CheckAndHeal()`
- `NewSelfHealingSupervisor()`
- `OnAction()`
- `Start()`
- `Stats()`
- `Stop()`

### 📄 [self_healing_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/self_healing_test.go) (61 LoC)
**Funciones Clave:**
- `TestSelfHealingSupervisor()`

### 📄 [smallworld.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/smallworld.go) (224 LoC)
**Tipos y Estructuras:**
- `struct PeerEntry`
- `struct SmallWorldTable`
**Funciones Clave:**
- `AddPeer()`
- `CalculateDegree()`
- `CompareDistance()`
- `EqualIdentities()`
- `FindClosestPeers()`
- `GetPeer()`
- `LeadingZeroBits()`
- `NewSmallWorldTable()`
- `TotalPeers()`
- `XorDistance()`

### 📄 [smallworld_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/smallworld_test.go) (118 LoC)
**Funciones Clave:**
- `TestSmallWorldHopLimitDrop()`
- `TestSmallWorldMultiHopRelay()`
- `TestSmallWorldTableBoundedCapacity()`

### 📄 [socks5.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/socks5.go) (216 LoC)
**Tipos y Estructuras:**
- `struct SOCKS5Proxy`
**Funciones Clave:**
- `Addr()`
- `NewSOCKS5Proxy()`
- `SetExitPeer()`
- `Start()`
- `Stop()`

### 📄 [socks5_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/socks5_test.go) (97 LoC)
**Funciones Clave:**
- `TestSOCKS5Proxy()`

### 📄 [tunnel.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/tunnel.go) (289 LoC)
**Tipos y Estructuras:**
- `struct TunnelPacket`
- `struct TunnelSession`
- `struct TunnelService`
**Funciones Clave:**
- `ActiveForwarders()`
- `CloseListener()`
- `ForwardPort()`
- `NewTunnelService()`
- `Stop()`

### 📄 [tunnel_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/core/tunnel_test.go) (111 LoC)
**Funciones Clave:**
- `TestP2PTunnelPortForwarding()`

## Paquete `dht`

Tabla Hash Distribuida Kademlia segura para resolución descentralizada Identity -> Endpoints con firmas digitales.

### 📄 [dht.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/dht/dht.go) (188 LoC)
**Tipos y Estructuras:**
- `struct DHTService`
**Funciones Clave:**
- `AddPeer()`
- `NewDHTService()`
- `ProcessMessage()`
- `Publish()`
- `Resolve()`
- `SetRemoteCaller()`

### 📄 [dht_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/dht/dht_test.go) (118 LoC)
**Funciones Clave:**
- `TestDHTPublishAndResolve()`
- `TestDHTRejectsForgedRecord()`
- `TestRecordSigningAndVerification()`

### 📄 [message.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/dht/message.go) (36 LoC)
**Tipos y Estructuras:**
- `struct PeerInfo`
- `struct Message`
**Funciones Clave:**
- `Marshal()`
- `Unmarshal()`

### 📄 [record.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/dht/record.go) (92 LoC)
**Tipos y Estructuras:**
- `struct Record`
**Funciones Clave:**
- `Marshal()`
- `NewRecord()`
- `Sign()`
- `Unmarshal()`
- `Verify()`

## Paquete `main`

Puntos de entrada ejecutables: nodo P2P (`cmd/node`) y cliente de chat CLI (`cmd/chat`).

### 📄 [main.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/cmd/chat/main.go) (193 LoC)

### 📄 [main.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/cmd/node/main.go) (279 LoC)

## Paquete `ui`

Dashboard web SPA interactivo, servidor HTTP/WebSocket, Chat E2EE y monitor de topología en malla.

### 📄 [kuzu.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/ui/kuzu.go) (447 LoC)
**Tipos y Estructuras:**
- `struct GraphNode`
- `struct GraphLink`
- `struct GraphResponse`
- `struct PeerQueryResult`
- `struct CodeQueryResult`
**Funciones Clave:**
- `ExecuteCypher()`
- `FindKuzuBinary()`
- `FindKuzuDB()`
- `SyncPeersToKuzu()`

### 📄 [observability.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/ui/observability.go) (209 LoC)

### 📄 [server.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/ui/server.go) (905 LoC)
**Tipos y Estructuras:**
- `struct Server`
- `struct SendMessageReq`
- `struct PeerDTO`
- `struct BroadcastStreamReq`
- `struct StreamTelemetryFrame`
- `struct PingReq`
- `struct MeshNode`
- `struct MeshLink`
- `struct MeshGraphResponse`
- `struct StartTunnelReq`
- `struct StartVPNReq`
**Funciones Clave:**
- `NewServer()`
- `SetFirebaseDiscovery()`
- `Start()`

### 📄 [server_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/ui/server_test.go) (143 LoC)
**Funciones Clave:**
- `TestKuzuEndpoints()`
- `TestObservabilityEndpoints()`
- `TestUIEndpoints()`

---
## Herramientas y Scripts de Automatización

- [scripts/build_linux.ps1](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/scripts/build_linux.ps1): Compilación cruzada para Linux amd64.
- [scripts/run_wsl.ps1](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/scripts/run_wsl.ps1): Supervisor de ejecución del nodo en WSL2.
- [scripts/wsl_node.sh](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/scripts/wsl_node.sh): Runner nativo en bash para Ubuntu WSL2 con logging rotativo.
- [scripts/dual_node_test.ps1](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/scripts/dual_node_test.ps1): Orquestador de pruebas de malla cruzada Windows <-> WSL2.
- [tools/kuzu/setup_kuzu.ps1](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/tools/kuzu/setup_kuzu.ps1): Descarga y aprovisionamiento de Kùzu Graph DB CLI.
- [tools/indexer/index_project.py](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/tools/indexer/index_project.py): Indexador de código hacia Grafo Kùzu.
- [tools/indexer/audit_kuzu.py](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/tools/indexer/audit_kuzu.py): Auditoría integral y métricas del Grafo Kùzu.

## Documentación Técnica Relacionada

- [docs/README.md](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/README.md): Catálogo y Hub Central de Documentación IPv7.
- [docs/operaciones/06_supervision_wsl2.md](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/06_supervision_wsl2.md): Guía operativa de supervisión en WSL2.
- [docs/arquitectura/03_kuzu_mesh_graph.md](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/arquitectura/03_kuzu_mesh_graph.md): Modelado en Kùzu y consultas Cypher.
- [docs/arquitectura/02_mundo_pequeno_routing.md](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/arquitectura/02_mundo_pequeno_routing.md): Enrutamiento logarítmico acotado a 12 grados.
- [docs/arquitectura/01_genesis.md](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/arquitectura/01_genesis.md): Plan génesis de arquitectura del proyecto.
- [docs/arquitectura/Ruta_de_trabajo_IPv7_P2P_DHT.pdf](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/arquitectura/Ruta_de_trabajo_IPv7_P2P_DHT.pdf): Especificación técnica de arquitectura P2P/DHT.
