---
name: ipv7-ai-ops
description: Comprehensive Autonomous Agent Skill for IPv7 Overlay Network. Encapsulates cryptographic validation, small-world routing heuristics, cascade streaming topology, Kùzu graph DB queries, REST/WS API interaction, and diagnostic playbooks for AI models and LLM agents.
---

# IPv7 AI Operations & Architecture Skill

Esta habilidad dota a cualquier agente de Inteligencia Artificial (Antigravity, Cursor, Claude Code, OpenAI Assistants, LangChain/LlamaIndex agents) de la capacidad de operar, diagnosticar, auditar y programar sobre el ecosistema **IPv7**.

## 🧠 Arquitectura de la Skill

La suite técnica se compone de los siguientes módulos:

1. [**System Prompt & AI Persona**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/skill_ia_ipv7/system_prompt_ia.md): Reglas de comportamiento, tono, axiomas criptográficos y restricciones de seguridad para la IA.
2. [**Especificación Técnica de APIs (REST + WS + CBOR)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/skill_ia_ipv7/api_reference_ai.md): Catálogo completo de endpoints HTTP, sockets WebSocket, formatos binarios CBOR e interfaces de servicio.
3. [**Guía de Consultas Cypher con Kùzu Graph Engine**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/skill_ia_ipv7/kuzu_cypher_agent_guide.md): Cheatsheet de consultas en Cypher para inspeccionar el grafo de código Go y la topología de la malla P2P.
4. [**Schemas de Function Calling para LLMs**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/skill_ia_ipv7/function_calling_schemas.json): Definiciones formales en formato JSON Schema / OpenAI Tools listas para ser invocadas por modelos de lenguaje.
5. [**Workflows Automatizados de Diagnóstico**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/skill_ia_ipv7/workflows_diagnostico_ia.md): Procedimientos operativos estándar (SOP) ejecutables por la IA para resolver anomalías de red de forma autónoma.

---

## ⚡ Reglas Cardinales para Agentes de IA

Cualquier IA que interactúe con el código o nodos de IPv7 DEBE adherirse a estas 5 reglas cardinales:

1. **Invarianza Criptográfica**: Las identidades son SIEMPRE claves Ed25519 de 32 bytes (64 caracteres hex). Nunca inventes identidades aleatorias que no correspondan a una firma válida.
2. **Canonicidad CBOR Estricta**: Antes de firmar cualquier contenedor de datos (`Container`), debe ser serializado de forma canónica y determinista con `fxamacker/cbor/v2`.
3. **Acotamiento del Mundo Pequeño**: La tabla de enrutamiento del nodo NUNCA debe superar los 120 peers (12 anillos $\times$ 10). Los mensajes tienen un límite estricto de `HopLimit = 12`.
4. **Priorización de Transporte Shift-Left**: Favorecer conexiones directas UDP/QUIC con STUN; recurrir al Relay adapter únicamente cuando ambos extremos estén tras NAT simétrico.
5. **No Bloquear Sockets**: Las rutinas de red en Go deben ser concurrentes y controladas por canales `stopCh` o `context.Context`.

---

## 🛠️ Comandos de Interacción Rápida para la IA

| Acción de la IA | Comando de Terminal |
|---|---|
| Validar tests de la suite completa | `go test -v ./...` |
| Test unitario específico de E2EE | `go test -v ./core -run TestE2EESession` |
| Test unitario de túneles P2P | `go test -v ./core -run TestTunnelDataFlow` |
| Reindexar proyecto en base Kùzu | `python .\tools\indexer\index_project.py` |
| Inspeccionar grafo Kùzu en CLI | `.\tools\kuzu\kuzu.exe .kuzu_index/` |
| Consultar estado del nodo vía API | `curl http://localhost:8080/api/info` |
| Obtener lista de peers vía API | `curl http://localhost:8080/api/peers` |
| Enviar ping de latencia a un peer | `curl -X POST http://localhost:8080/api/ping -d '{"target_id":"<HEX>"}'` |
| Arrancar nodo en modo servidor MCP | `.\ipv7-node.exe -mcp -key persistent` |
| Invocar herramienta MCP vía HTTP | `curl -X POST http://localhost:8080/api/mcp -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'` |
| Obtener especificación OpenAPI 3.1 | `curl http://localhost:8080/api/openapi.json` |
| Extraer métricas Prometheus en vivo | `curl http://localhost:8080/metrics` |
