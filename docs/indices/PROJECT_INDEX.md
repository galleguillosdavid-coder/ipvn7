# Índice Integral del Proyecto IPv7

Este documento cataloga de forma exhaustiva todos los componentes, archivos, tipos, funciones y herramientas del protocolo IPv7. Generado automáticamente por el indexador semántico.

## Resumen de Paquetes

| Paquete | Archivos | Líneas de Código (LoC) | Propósito Principal |
|---|---|---|---|
| [`adapters`](#paquete-adapters) | 18 | 2298 | Adaptadores de transporte desacoplados: UDP best-effort, QUIC sobre TLS 1.3, NAT Traversal con STUN, Relay seguro tipo DERP y WebRTC. |
| [`core`](#paquete-core) | 41 | 6081 | Núcleo de protocolo: Identidad Ed25519, Contenedores CBOR, Cifrado E2EE, Mundo Pequeño (12 Grados), Streaming en Cascada y Orquestador de Nodos. |
| [`dht`](#paquete-dht) | 7 | 955 | Tabla Hash Distribuida Kademlia segura para resolución descentralizada Identity -> Endpoints con firmas digitales. |
| [`main`](#paquete-main) | 3 | 587 | Puntos de entrada ejecutables: nodo P2P (`cmd/node`) y cliente de chat CLI (`cmd/chat`). |
| [`offgrid`](#paquete-offgrid) | 5 | 961 | Módulo de soporte del sistema. |
| [`onion`](#paquete-onion) | 4 | 561 | Módulo de soporte del sistema. |
| [`telemetry`](#paquete-telemetry) | 9 | 1221 | Módulo de soporte del sistema. |
| [`tests`](#paquete-tests) | 5 | 1153 | Módulo de soporte del sistema. |
| [`tun`](#paquete-tun) | 5 | 859 | Módulo de soporte del sistema. |
| [`ui`](#paquete-ui) | 4 | 1728 | Dashboard web SPA interactivo, servidor HTTP/WebSocket, Chat E2EE y monitor de topología en malla. |
| [`watchdog`](#paquete-watchdog) | 1 | 87 | Módulo de soporte del sistema. |

---

## Paquete `adapters`

Adaptadores de transporte desacoplados: UDP best-effort, QUIC sobre TLS 1.3, NAT Traversal con STUN, Relay seguro tipo DERP y WebRTC.

### 📄 [pmtu.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/pmtu.go) (215 LoC)
**Tipos y Estructuras:**
- `struct PMTUState`
- `struct PMTUDiscovery`
**Funciones Clave:**
- `ConfirmProbeSize()`
- `DiscoverPathMTU()`
- `GetMTU()`
- `NewPMTUDiscovery()`
- `NextProbeSize()`
- `RecordBlackHoleTimeout()`
- `RegisterEndpoint()`

### 📄 [pmtu_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/pmtu_test.go) (94 LoC)
**Funciones Clave:**
- `TestPMTUBlackHoleStepDown()`
- `TestPMTUConfirmProbeSize()`
- `TestPMTUDiscoverySequence()`
- `TestPMTUInitialDefaults()`

### 📄 [quic_adapter.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/quic_adapter.go) (209 LoC)
**Tipos y Estructuras:**
- `struct QUICAdapter`
**Funciones Clave:**
- `LocalAddr()`
- `NewQUICAdapter()`
- `Receive()`
- `Send()`
- `Start()`
- `Stop()`

### 📄 [quic_adapter_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/quic_adapter_test.go) (107 LoC)
**Funciones Clave:**
- `TestQUICAdapter()`
- `TestQUICAdapterLargePayload()`

### 📄 [relay_adapter.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/relay_adapter.go) (280 LoC)
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

### 📄 [relay_adapter_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/relay_adapter_test.go) (70 LoC)
**Funciones Clave:**
- `TestRelayServerAndAdapter()`

### 📄 [replay_filter.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/replay_filter.go) (87 LoC)
**Tipos y Estructuras:**
- `struct ReplayWindow`
- `struct AntiReplayTable`
**Funciones Clave:**
- `CheckAndSet()`
- `NewAntiReplayTable()`
- `Reset()`

### 📄 [replay_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/replay_test.go) (120 LoC)
**Funciones Clave:**
- `TestAntiReplayTablePerPeerIsolation()`
- `TestReplayWindowDuplicatesRejected()`
- `TestReplayWindowOutOfOrderWithinWindow()`
- `TestReplayWindowSequential()`
- `TestReplayWindowTooOldRejected()`

### 📄 [stun.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/stun.go) (94 LoC)
**Funciones Clave:**
- `DiscoverPublicEndpoint()`
- `GetLocalEndpoints()`

### 📄 [stun_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/stun_test.go) (42 LoC)
**Funciones Clave:**
- `TestDiscoverPublicEndpoint()`
- `TestGetLocalEndpoints()`
- `TestUDPAdapterDiscoverEndpoints()`

### 📄 [tls_helper.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/tls_helper.go) (47 LoC)
**Funciones Clave:**
- `GenerateTLSConfig()`

### 📄 [tls_helper_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/tls_helper_test.go) (42 LoC)
**Funciones Clave:**
- `TestGenerateTLSConfig()`

### 📄 [udp_adapter.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/udp_adapter.go) (262 LoC)
**Tipos y Estructuras:**
- `struct UDPAdapter`
**Funciones Clave:**
- `DiscoverEndpoints()`
- `Endpoints()`
- `LocalAddr()`
- `NewUDPAdapter()`
- `Receive()`
- `Send()`
- `Start()`
- `Stop()`

### 📄 [udp_adapter_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/udp_adapter_test.go) (79 LoC)
**Funciones Clave:**
- `TestUDPAdapter()`
- `TestUDPAdapterMTU()`

### 📄 [upnp.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/upnp.go) (185 LoC)
**Tipos y Estructuras:**
- `struct UPnPMapper`
**Funciones Clave:**
- `DiscoverAndForward()`
- `NewUPnPMapper()`

### 📄 [upnp_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/upnp_test.go) (19 LoC)
**Funciones Clave:**
- `TestUPnPMapperGracefulTimeout()`

### 📄 [webrtc_adapter.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/webrtc_adapter.go) (249 LoC)
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

### 📄 [webrtc_adapter_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/webrtc_adapter_test.go) (97 LoC)
**Funciones Clave:**
- `TestWebRTCAdapterCommunication()`

## Paquete `core`

Núcleo de protocolo: Identidad Ed25519, Contenedores CBOR, Cifrado E2EE, Mundo Pequeño (12 Grados), Streaming en Cascada y Orquestador de Nodos.

### 📄 [beacon.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/beacon.go) (236 LoC)
**Tipos y Estructuras:**
- `struct BeaconPacket`
- `struct BeaconService`
**Funciones Clave:**
- `NewBeaconService()`
- `Start()`
- `Stop()`

### 📄 [benchmark_pipeline_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/benchmark_pipeline_test.go) (161 LoC)
**Funciones Clave:**
- `BenchmarkFullPipelineThroughputMBps()`
- `TestBenchmarkLayerBottleneckBreakdown()`

### 📄 [benchmark_production_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/benchmark_production_test.go) (179 LoC)
**Funciones Clave:**
- `TestProductionBenchmarkSuite()`

### 📄 [benchmark_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/benchmark_test.go) (134 LoC)
**Funciones Clave:**
- `BenchmarkContainerSerialization()`
- `BenchmarkContainerSign()`
- `BenchmarkContainerVerify()`
- `BenchmarkE2EEEncryptDecrypt()`
- `BenchmarkSmallWorldLookup()`

### 📄 [cascade.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/cascade.go) (132 LoC)
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

### 📄 [cascade_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/cascade_test.go) (110 LoC)
**Funciones Clave:**
- `TestCascadeMaxChildrenEnforcement()`
- `TestCascadeStreamingMultiHop()`

### 📄 [chaos_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/chaos_test.go) (172 LoC)
**Tipos y Estructuras:**
- `struct ChaosChannel`
**Funciones Clave:**
- `NewChaosChannel()`
- `Send()`
- `TestChaosConcurrentE2EEStress()`
- `TestChaosPacketLossAndJitter()`

### 📄 [container.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/container.go) (91 LoC)
**Tipos y Estructuras:**
- `struct Container`
**Funciones Clave:**
- `Marshal()`
- `Sign()`
- `Unmarshal()`
- `Verify()`

### 📄 [container_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/container_test.go) (78 LoC)
**Funciones Clave:**
- `TestContainerSerialization()`
- `TestContainerSigningAndVerification()`
- `TestGenerateIdentity()`

### 📄 [discovery_firebase.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/discovery_firebase.go) (306 LoC)
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

### 📄 [discovery_firebase_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/discovery_firebase_test.go) (67 LoC)
**Funciones Clave:**
- `TestFirebaseDiscoveryAnnounceAndFetch()`

### 📄 [e2ee.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/e2ee.go) (159 LoC)
**Funciones Clave:**
- `DecryptE2EE()`
- `DeriveX25519FromSeed()`
- `EncryptE2EE()`
- `GenerateE2EEKeyPair()`

### 📄 [e2ee_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/e2ee_test.go) (69 LoC)
**Funciones Clave:**
- `TestE2EEDeriveFromSeed()`
- `TestE2EEEncryptionDecryption()`
- `TestE2EETamperRejection()`

### 📄 [handshake.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/handshake.go) (88 LoC)
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

### 📄 [handshake_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/handshake_test.go) (62 LoC)
**Funciones Clave:**
- `TestHandshakePayloadSerialization()`
- `TestPingPayloadSerialization()`

### 📄 [identity.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/identity.go) (51 LoC)
**Tipos y Estructuras:**
- `struct Ed25519Identity`
**Funciones Clave:**
- `Bytes()`
- `GenerateIdentity()`
- `NewIdentityFromBytes()`
- `NewIdentityFromHex()`
- `String()`
- `Verify()`

### 📄 [interfaces.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/interfaces.go) (34 LoC)
**Tipos y Estructuras:**
- `interface Identity`
- `interface Session`
- `interface Adapter`

### 📄 [keystore.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/keystore.go) (67 LoC)
**Funciones Clave:**
- `DefaultKeyDir()`
- `LoadOrCreatePersistentIdentity()`

### 📄 [keystore_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/keystore_test.go) (39 LoC)
**Funciones Clave:**
- `TestKeystorePersistence()`

### 📄 [logger.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/logger.go) (50 LoC)
**Funciones Clave:**
- `InitLogger()`
- `Log()`

### 📄 [logger_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/logger_test.go) (24 LoC)
**Funciones Clave:**
- `TestStructuredLogger()`

### 📄 [mcp.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/mcp.go) (440 LoC)
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

### 📄 [mcp_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/mcp_test.go) (144 LoC)
**Funciones Clave:**
- `TestMCPServer()`

### 📄 [node.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/node.go) (619 LoC)
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

### 📄 [node_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/node_test.go) (165 LoC)
**Funciones Clave:**
- `Receive()`
- `Send()`
- `Start()`
- `Stop()`
- `TestNodeHandshakeAndE2EE()`
- `TestNodeP2PCommunication()`

### 📄 [noise.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/noise.go) (365 LoC)
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

### 📄 [noise_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/noise_test.go) (135 LoC)
**Funciones Clave:**
- `TestNoiseHandshakeAndSession()`
- `TestNoiseTamperedSignature()`

### 📄 [pqc.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/pqc.go) (128 LoC)
**Tipos y Estructuras:**
- `struct HybridIdentity`
- `struct HybridSignature`
**Funciones Clave:**
- `CombineKEMSecrets()`
- `GeneratePQCKeyStub()`
- `NewHybridIdentity()`
- `SignHybrid()`
- `VerifyHybrid()`

### 📄 [pqc_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/pqc_test.go) (66 LoC)
**Funciones Clave:**
- `TestCombineKEMSecrets()`
- `TestHybridIdentityAndSigning()`

### 📄 [remotedesktop.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/remotedesktop.go) (72 LoC)
**Tipos y Estructuras:**
- `struct RemoteInputEvent`
- `struct RemoteDesktopService`
**Funciones Clave:**
- `GetResolution()`
- `NewRemoteDesktopService()`
- `StartStreaming()`

### 📄 [remotedesktop_other.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/remotedesktop_other.go) (47 LoC)
**Funciones Clave:**
- `CaptureFrame()`
- `InjectInput()`

### 📄 [remotedesktop_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/remotedesktop_test.go) (34 LoC)
**Funciones Clave:**
- `TestRemoteDesktopCapture()`

### 📄 [remotedesktop_windows.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/remotedesktop_windows.go) (284 LoC)
**Funciones Clave:**
- `CaptureFrame()`
- `InjectInput()`

### 📄 [self_healing.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/self_healing.go) (157 LoC)
**Tipos y Estructuras:**
- `struct SelfHealingSupervisor`
**Funciones Clave:**
- `CheckAndHeal()`
- `NewSelfHealingSupervisor()`
- `OnAction()`
- `Start()`
- `Stats()`
- `Stop()`

### 📄 [self_healing_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/self_healing_test.go) (61 LoC)
**Funciones Clave:**
- `TestSelfHealingSupervisor()`

### 📄 [smallworld.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/smallworld.go) (224 LoC)
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

### 📄 [smallworld_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/smallworld_test.go) (118 LoC)
**Funciones Clave:**
- `TestSmallWorldHopLimitDrop()`
- `TestSmallWorldMultiHopRelay()`
- `TestSmallWorldTableBoundedCapacity()`

### 📄 [socks5.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/socks5.go) (216 LoC)
**Tipos y Estructuras:**
- `struct SOCKS5Proxy`
**Funciones Clave:**
- `Addr()`
- `NewSOCKS5Proxy()`
- `SetExitPeer()`
- `Start()`
- `Stop()`

### 📄 [socks5_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/socks5_test.go) (97 LoC)
**Funciones Clave:**
- `TestSOCKS5Proxy()`

### 📄 [tunnel.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/tunnel.go) (289 LoC)
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

### 📄 [tunnel_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/core/tunnel_test.go) (111 LoC)
**Funciones Clave:**
- `TestP2PTunnelPortForwarding()`

## Paquete `dht`

Tabla Hash Distribuida Kademlia segura para resolución descentralizada Identity -> Endpoints con firmas digitales.

### 📄 [dht.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/dht/dht.go) (188 LoC)
**Tipos y Estructuras:**
- `struct DHTService`
**Funciones Clave:**
- `AddPeer()`
- `NewDHTService()`
- `ProcessMessage()`
- `Publish()`
- `Resolve()`
- `SetRemoteCaller()`

### 📄 [dht_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/dht/dht_test.go) (118 LoC)
**Funciones Clave:**
- `TestDHTPublishAndResolve()`
- `TestDHTRejectsForgedRecord()`
- `TestRecordSigningAndVerification()`

### 📄 [discovery_dht.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/dht/discovery_dht.go) (102 LoC)
**Tipos y Estructuras:**
- `struct DHTDiscoveryAdapter`
**Funciones Clave:**
- `NewDHTDiscoveryAdapter()`
- `Start()`
- `Stop()`

### 📄 [kademlia_pure.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/dht/kademlia_pure.go) (229 LoC)
**Tipos y Estructuras:**
- `struct Contact`
- `struct KBucket`
- `struct PureKademliaTable`
**Funciones Clave:**
- `AddContact()`
- `BucketIndex()`
- `ComputePoW()`
- `FindClosest()`
- `NewPureKademliaTable()`
- `TotalContacts()`
- `VerifyPoW()`
- `XOR()`

### 📄 [kademlia_pure_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/dht/kademlia_pure_test.go) (190 LoC)
**Funciones Clave:**
- `BenchmarkKademliaFindClosest()`
- `BenchmarkXORDistance()`
- `TestAntiSybilProofOfWork()`
- `TestDHTDiscoveryAdapterSovereign()`
- `TestPureKademliaTableKBuckets()`
- `TestXORDistanceMetric()`

### 📄 [message.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/dht/message.go) (36 LoC)
**Tipos y Estructuras:**
- `struct PeerInfo`
- `struct Message`
**Funciones Clave:**
- `Marshal()`
- `Unmarshal()`

### 📄 [record.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/dht/record.go) (92 LoC)
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

### 📄 [canary_runner.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/canary/runner/canary_runner.go) (115 LoC)

### 📄 [main.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/cmd/chat/main.go) (193 LoC)

### 📄 [main.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/cmd/node/main.go) (279 LoC)

## Paquete `offgrid`



### 📄 [adhoc_mesh.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/offgrid/adhoc_mesh.go) (289 LoC)
**Tipos y Estructuras:**
- `struct VirtualRadioLink`
- `struct RadioMedium`
- `struct BeaconEngine`
**Funciones Clave:**
- `Broadcast()`
- `Close()`
- `DiscoveredPeers()`
- `MTU()`
- `Name()`
- `NewBeaconEngine()`
- `NewRadioMedium()`
- `NewVirtualRadioLink()`
- `Receive()`
- `Register()`
- `Send()`
- `SetOnData()`
- `Start()`
- `Stop()`
- `Transmit()`
- `Unregister()`

### 📄 [fragmentation.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/offgrid/fragmentation.go) (136 LoC)
**Tipos y Estructuras:**
- `struct Reassembler`
**Funciones Clave:**
- `FragmentPacket()`
- `NewReassembler()`
- `ProcessChunk()`

### 📄 [hybrid_switcher.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/offgrid/hybrid_switcher.go) (248 LoC)
**Tipos y Estructuras:**
- `struct SwitcherConfig`
- `struct SwitcherStats`
- `struct HybridSwitcher`
**Funciones Clave:**
- `CurrentMode()`
- `IsWANOnline()`
- `NewHybridSwitcher()`
- `ReportWANStatus()`
- `RoutePacket()`
- `SetMode()`
- `SetOnModeChange()`
- `StartWatchdog()`
- `Stats()`
- `Stop()`

### 📄 [offgrid_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/offgrid/offgrid_test.go) (234 LoC)
**Funciones Clave:**
- `BenchmarkHybridSwitcher_RouteDirectPhysical()`
- `TestBeaconEngine_Discovery()`
- `TestHybridSwitcher_FailoverAndRouting()`
- `TestVirtualRadioLink_Broadcast()`

### 📄 [types.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/offgrid/types.go) (54 LoC)
**Tipos y Estructuras:**
- `interface PhysicalLink`
- `struct PhysicalPeer`
**Funciones Clave:**
- `String()`

## Paquete `onion`



### 📄 [circuit.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/onion/circuit.go) (131 LoC)
**Tipos y Estructuras:**
- `struct OnionRouter`
**Funciones Clave:**
- `BuildCircuit()`
- `NewOnionRouter()`
- `ProcessInboundPacket()`
- `SetExitHandler()`

### 📄 [onion_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/onion/onion_test.go) (200 LoC)
**Funciones Clave:**
- `BenchmarkBuildOnion3Hops()`
- `BenchmarkUnwrapLayer()`
- `TestSphinxFixedPacketSizePadding()`
- `TestSphinxPacketBuildAndUnwrapSingleHop()`
- `TestSphinxPacketBuildAndUnwrapThreeHops()`

### 📄 [sphinx.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/onion/sphinx.go) (183 LoC)
**Funciones Clave:**
- `BuildOnionPacket()`
- `UnwrapLayer()`

### 📄 [types.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/onion/types.go) (47 LoC)
**Tipos y Estructuras:**
- `struct CircuitHop`
- `struct LayerInstruction`
- `interface OnionRelayNode`

## Paquete `telemetry`



### 📄 [anomaly_engine.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/telemetry/anomaly_engine.go) (160 LoC)
**Tipos y Estructuras:**
- `struct AnomalyRule`
- `struct AnomalyEngine`
**Funciones Clave:**
- `CheckAnomalies()`
- `DefaultAnomalyRule()`
- `NewAnomalyEngine()`

### 📄 [benchmark_telemetry_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/telemetry/benchmark_telemetry_test.go) (63 LoC)
**Funciones Clave:**
- `BenchmarkTelemetryOFF()`
- `BenchmarkTelemetryON()`
- `BenchmarkTelemetryON_Prometheus()`

### 📄 [bus.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/telemetry/bus.go) (206 LoC)
**Tipos y Estructuras:**
- `interface EventConsumer`
- `struct TelemetryBus`
**Funciones Clave:**
- `DroppedEvents()`
- `EmitEvent()`
- `GetGlobalBus()`
- `GetRTT()`
- `InitGlobalBus()`
- `NewTelemetryBus()`
- `RecordBackpressureDrop()`
- `RecordJitter()`
- `RecordPacketRX()`
- `RecordPacketTX()`
- `RecordRTT()`
- `RecordSocketDrop()`
- `RegisterConsumer()`
- `Start()`
- `Stop()`

### 📄 [exporter.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/canary/telemetry/exporter.go) (164 LoC)
**Tipos y Estructuras:**
- `struct NodeMetrics`
**Funciones Clave:**
- `NewNodeMetrics()`
- `StartMetricsServer()`

### 📄 [kuzu_exporter.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/telemetry/kuzu_exporter.go) (152 LoC)
**Tipos y Estructuras:**
- `struct KuzuExporter`
**Funciones Clave:**
- `Close()`
- `ConsumeBatch()`
- `NewKuzuExporter()`

### 📄 [prometheus_exporter.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/telemetry/prometheus_exporter.go) (178 LoC)
**Tipos y Estructuras:**
- `struct PrometheusExporter`
**Funciones Clave:**
- `NewPrometheusExporter()`
- `Start()`
- `Stop()`

### 📄 [ring_buffer.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/telemetry/ring_buffer.go) (77 LoC)
**Tipos y Estructuras:**
- `struct BoundedEventQueue`
**Funciones Clave:**
- `Capacity()`
- `DroppedTotal()`
- `EnqueuedTotal()`
- `Len()`
- `NewBoundedEventQueue()`
- `PopBatch()`
- `TryPush()`

### 📄 [telemetry_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/telemetry/telemetry_test.go) (98 LoC)
**Funciones Clave:**
- `ConsumeBatch()`
- `TestAnomalyEngineDetection()`
- `TestBoundedQueueNonBlockingDrop()`
- `TestTelemetryBusAsyncConsumer()`

### 📄 [types.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/telemetry/types.go) (123 LoC)
**Tipos y Estructuras:**
- `struct TelemetryEvent`
- `struct AnomalyEvent`
- `struct NodeMetrics`
- `struct TransportMetrics`
- `struct ProtocolMetrics`

## Paquete `tests`



### 📄 [constrained_link_fragmentation_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/tests/constrained_link_fragmentation_test.go) (125 LoC)
**Funciones Clave:**
- `TestConstrainedLink_FragmentationAndReassemblyUnderChaos()`
- `TestConstrainedLink_StrictMTUEnforcement()`

### 📄 [e2e_four_horizons_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/tests/e2e_four_horizons_test.go) (225 LoC)
**Funciones Clave:**
- `TestFourHorizons_UnifiedEndToEnd()`

### 📄 [l2_dos_and_flapping_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/tests/l2_dos_and_flapping_test.go) (239 LoC)
**Funciones Clave:**
- `TestBeaconEngine_PoisonAndFloodResistance()`
- `TestHybridSwitcher_FlappingResistance()`

### 📄 [multimedium_chaos_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/tests/multimedium_chaos_test.go) (355 LoC)
**Funciones Clave:**
- `Cut()`
- `Receive()`
- `Restore()`
- `Send()`
- `Start()`
- `Stop()`
- `TestMultimediumTransportFailover_Chaos()`

### 📄 [split_brain_healing_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/tests/split_brain_healing_test.go) (209 LoC)
**Funciones Clave:**
- `TestSplitBrainAndHealing_MeshConvergence()`

## Paquete `tun`



### 📄 [adapter.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/tun/adapter.go) (206 LoC)
**Tipos y Estructuras:**
- `struct TunAdapter`
**Funciones Clave:**
- `DeliverInbound()`
- `Drops()`
- `HandlePeerRoamed()`
- `NewTunAdapter()`
- `PacketsIn()`
- `PacketsOut()`
- `ResolveAndRegister()`
- `Routes()`
- `Start()`
- `Stop()`

### 📄 [addressing.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/tun/addressing.go) (39 LoC)
**Funciones Clave:**
- `DeriveIPv6FromDID()`
- `GenerateVirtualIPv4()`

### 📄 [device_mock.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/tun/device_mock.go) (112 LoC)
**Tipos y Estructuras:**
- `struct MemoryTunDevice`
**Funciones Clave:**
- `Close()`
- `InjectPacket()`
- `MTU()`
- `Name()`
- `NewMemoryTunDevice()`
- `Read()`
- `ReceiveOutbound()`
- `Write()`

### 📄 [tun_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/tun/tun_test.go) (421 LoC)
**Funciones Clave:**
- `BenchmarkExtractDestinationIP()`
- `BenchmarkExtractDestinationIPKeyZeroAlloc()`
- `BenchmarkVirtualRouteLookup()`
- `BenchmarkVirtualRouteLookupZeroAlloc()`
- `Receive()`
- `Send()`
- `Start()`
- `Stop()`
- `TestAddressingDerivation()`
- `TestExtractDestinationIP()`
- `TestTunAdapterEndToEnd()`
- `TestTunAdapterRoamingPreservation()`
- `TestTunAdapterSoak10kPackets()`
- `TestVirtualRouteTable()`

### 📄 [types.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/adapters/tun/types.go) (81 LoC)
**Tipos y Estructuras:**
- `interface TunDevice`
- `struct VirtualRouteTable`
**Funciones Clave:**
- `LookupDID()`
- `LookupDIDKey()`
- `LookupIP()`
- `NewVirtualRouteTable()`
- `Register()`
- `ToIPKey()`

## Paquete `ui`

Dashboard web SPA interactivo, servidor HTTP/WebSocket, Chat E2EE y monitor de topología en malla.

### 📄 [kuzu.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/ui/kuzu.go) (447 LoC)
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

### 📄 [observability.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/ui/observability.go) (209 LoC)

### 📄 [server.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/ui/server.go) (929 LoC)
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

### 📄 [server_test.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/ui/server_test.go) (143 LoC)
**Funciones Clave:**
- `TestKuzuEndpoints()`
- `TestObservabilityEndpoints()`
- `TestUIEndpoints()`

## Paquete `watchdog`



### 📄 [watchdog.go](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/canary/watchdog/watchdog.go) (87 LoC)
**Tipos y Estructuras:**
- `struct NodeSupervisor`
**Funciones Clave:**
- `NewNodeSupervisor()`
- `Start()`
- `Stop()`

---
## Herramientas y Scripts de Automatización

- [scripts/build_linux.ps1](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/scripts/build_linux.ps1): Compilación cruzada para Linux amd64.
- [scripts/run_wsl.ps1](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/scripts/run_wsl.ps1): Supervisor de ejecución del nodo en WSL2.
- [scripts/wsl_node.sh](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/scripts/wsl_node.sh): Runner nativo en bash para Ubuntu WSL2 con logging rotativo.
- [scripts/dual_node_test.ps1](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/scripts/dual_node_test.ps1): Orquestador de pruebas de malla cruzada Windows <-> WSL2.
- [tools/kuzu/setup_kuzu.ps1](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/tools/kuzu/setup_kuzu.ps1): Descarga y aprovisionamiento de Kùzu Graph DB CLI.
- [tools/indexer/index_project.py](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/tools/indexer/index_project.py): Indexador de código hacia Grafo Kùzu.
- [tools/indexer/audit_kuzu.py](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/tools/indexer/audit_kuzu.py): Auditoría integral y métricas del Grafo Kùzu.

## Documentación Técnica Relacionada

- [docs/README.md](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/docs/README.md): Catálogo y Hub Central de Documentación IPv7.
- [docs/operaciones/06_supervision_wsl2.md](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/docs/operaciones/06_supervision_wsl2.md): Guía operativa de supervisión en WSL2.
- [docs/arquitectura/03_kuzu_mesh_graph.md](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/docs/arquitectura/03_kuzu_mesh_graph.md): Modelado en Kùzu y consultas Cypher.
- [docs/arquitectura/02_mundo_pequeno_routing.md](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/docs/arquitectura/02_mundo_pequeno_routing.md): Enrutamiento logarítmico acotado a 12 grados.
- [docs/arquitectura/01_genesis.md](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/docs/arquitectura/01_genesis.md): Plan génesis de arquitectura del proyecto.
- [docs/arquitectura/Ruta_de_trabajo_IPv7_P2P_DHT.pdf](file:///C:/Users/Frondabrick/Desktop/dvd/ipvn7/docs/arquitectura/Ruta_de_trabajo_IPv7_P2P_DHT.pdf): Especificación técnica de arquitectura P2P/DHT.
