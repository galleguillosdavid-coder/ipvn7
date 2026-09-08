---
name: ipv7-assistant
description: Asistente especializado para el protocolo P2P descentralizado IPv7. Proporciona comandos de compilación cruzada, supervisión en WSL2, consultas Cypher con Kùzu, arquitectura del Mundo Pequeño y diagnóstico de topología en malla.
---

# IPv7 Developer Assistant Skill

Esta habilidad proporciona a Antigravity y al desarrollador un conjunto rápido de comandos, guías de arquitectura y flujos de trabajo operativos para el protocolo IPv7.

## Comandos Rápidos del Proyecto

| Acción | Comando PowerShell | Comando WSL2 / Bash |
|---|---|---|
| **Ejecutar Pruebas** | `go test -v ./...` | `go test -v ./...` |
| **Compilar Linux** | `.\scripts\build_linux.ps1` | `go build -o bin/ipv7-node-linux ./cmd/node` |
| **Lanzar en WSL2** | `.\scripts\run_wsl.ps1 -Port 7002 -UIPort 8082` | `./scripts/wsl_node.sh -port 7002 -ui 8082` |
| **Prueba Malla Dual** | `.\scripts\dual_node_test.ps1` | N/A (orquestado desde PowerShell) |
| **Reindexar Código** | `python .\tools\indexer\index_project.py` | `python3 ./tools/indexer/index_project.py` |
| **Consultar Kùzu** | `.\tools\kuzu\kuzu.exe .kuzu_index/` | `./tools/kuzu/kuzu .kuzu_index/` |

## Mapa de Paquetes y Responsabilidades

- `core/`:
  - `identity.go`: Claves Ed25519, hashing y derivaciones criptográficas.
  - `container.go`: Formato de paquete IPv7 canónico CBOR y verificación de firmas.
  - `e2ee.go`: Diffie-Hellman X25519 y cifrado simétrico ChaCha20-Poly1305.
  - `smallworld.go`: Tabla de enrutamiento acotada a 120 peers (12 anillos $\times$ 10). Reenvío voraz XOR.
  - `cascade.go`: Distribución de streaming en cascada (árbol con fan-out 10).
  - `node.go`: Orquestador principal que integra adapters, peers y callbacks.
- `adapters/`:
  - `udp_adapter.go`: Transporte UDP para paquetes individuales < MTU.
  - `quic_adapter.go`: Transporte fiable basado en QUIC con TLS 1.3 efímero para archivos y streaming.
  - `relay_adapter.go`: Fallback relay seguro tipo DERP para NAT simétrico.
  - `stun.go`: Descubrimiento de IPs públicas reflexivas con servidores STUN.
  - `webrtc_adapter.go`: Soporte P2P nativo para navegadores y RTCDataChannels.
- `dht/`:
  - `dht.go` / `record.go`: Mapeo distribuido `Identity -> Endpoints` con firmas digitales anti-envenenamiento.
- `ui/`:
  - `server.go`: Servidor HTTP/WebSocket que expone el dashboard web y APIs de topología.
  - `assets/index.html`: Dashboard SPA moderno con Chat E2EE, visualizador de malla en tiempo real y radar.
- `scripts/`:
  - `build_linux.ps1`, `run_wsl.ps1`, `wsl_node.sh`, `dual_node_test.ps1`.
- `tools/`:
  - `kuzu/`: Binarios de Kùzu Graph Database para Windows y Linux.
  - `indexer/`: Indexador semántico del proyecto hacia grafo Kùzu y markdown.

## Consultas Útiles en Kùzu CLI (Cypher)

```cypher
-- Listar todos los paquetes y archivos Go indexados
MATCH (p:Package)-[:CONTAINS]->(f:File) RETURN p.name, f.name;

-- Buscar qué funciones o structs pertenecen al core
MATCH (p:Package {name: 'core'})-[:CONTAINS]->(f:File)-[:DEFINES]->(s:Symbol)
RETURN f.name, s.name, s.kind;

-- Consultar nodos de la malla P2P registrados
MATCH (p:Peer)-[r:CONNECTED_TO]->(m:Peer)
RETURN p.id, r.adapter, r.latency_ms, m.id;
```

## 🧠 Suite Completa de Habilidades para IA y Documentación

Para operaciones autónomas avanzadas, consulta los módulos especializados:
- **Suite Integral de IA**: [skill_ia_ipv7/](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/skill_ia_ipv7/SKILL.md)
  - [System Prompt de IA Especializada](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/skill_ia_ipv7/system_prompt_ia.md)
  - [Referencia Formal de APIs REST & WebSockets](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/skill_ia_ipv7/api_reference_ai.md)
  - [Catálogo de Consultas Cypher para Kùzu](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/skill_ia_ipv7/kuzu_cypher_agent_guide.md)
  - [JSON Schemas para Function Calling / Tools](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/skill_ia_ipv7/function_calling_schemas.json)
  - [SOP de Diagnósticos Automatizados](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/skill_ia_ipv7/workflows_diagnostico_ia.md)
- **Manual Operativo Humano**: [manual_operativo_humano/](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/manual_operativo_humano/README.md)
- **Sugerencias de Código, UX e IA**: [sugerencias_mejora_codigo_ux_ia/](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/sugerencias_mejora_codigo_ux_ia/README.md)

