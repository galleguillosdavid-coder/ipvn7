# Sugerencias de Arquitectura, UX e IA para IPv7

Este directorio contiene propuestas de ingeniería de software, mejoras de experiencia de usuario (UX) y adaptaciones de inteligencia artificial (AI-readiness) para evolucionar el protocolo **IPv7** hacia un estándar de nivel de producción masiva.

---

## 🧭 Índice de Propuestas y Módulos

1. [**01. Mejoras de Experiencia de Usuario (UX)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/sugerencias_mejora_codigo_ux_ia/01_mejoras_experiencia_usuario_ux.md)
   - Dashboard Web Reactivo con grafos 3D y Canvas interactivo.
   - Transferencia de archivos Drag & Drop con progreso visual.
   - Gestor visual de túneles P2P y selector de nodo de salida VPN.
   - Terminal TUI interactiva (Text User Interface) con `charmbracelet/bubbletea`.
   - Bandeja del sistema (Systray) para ejecución en segundo plano.
   - Protocolos automáticos de apertura de puertos (UPnP IGD / NAT-PMP).

2. [**02. Mejoras para Inteligencias Artificiales y Agentes Autónomos (AI)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/sugerencias_mejora_codigo_ux_ia/02_mejoras_inteligencia_artificial_ai.md)
   - Servidor **Model Context Protocol (MCP)** nativo embebido en Go.
   - Especificación OpenAPI 3.1 viva (`/api/openapi.json`) y Swagger UI.
   - Logs estructurados con `log/slog` nativo de Go en formato JSON.
   - Métricas de telemetría Prometheus (`/metrics`) para observabilidad por IA.
   - Optimizador de topología Small-World autónomo guiado por IA con Kùzu.

3. [**03. Mejoras de Arquitectura y Rendimiento del Código**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/sugerencias_mejora_codigo_ux_ia/03_mejoras_arquitectura_y_rendimiento.md)
   - Multiplexación QUIC nativa (`quic.Stream`) para túneles P2P de alto ancho de banda.
   - Protocolo de Handshake Noise (Noise_XX) con Perfect Forward Secrecy (PFS).
   - Aceleración por hardware H.264 / VP8 para el Escritorio Remoto Web.
   - Persistencia de identidad y pares en base de datos ligera embebida (BuntDB / Pebble).

4. [**04. Prototipos y Snippets de Código en Go**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/sugerencias_mejora_codigo_ux_ia/04_prototipos_y_snippets_codigo.md)
   - Prototipo 1: Servidor MCP integrado en el binario del nodo.
   - Prototipo 2: Logging estructurado JSON con `log/slog`.
   - Prototipo 3: Soporte UPnP con `huin/goupnp`.
   - Prototipo 4: Almacenamiento seguro de claves con AES-GCM + archivo de configuración.

---

## 📊 Matriz de Priorización y Estado de Implementación

| Iniciativa | Área | Impacto | Esfuerzo | Estado Actual | Ubicación en Código |
|---|---|---|---|---|---|
| **Servidor MCP Nativo** | IA | Muy Alto | Bajo | ✅ **IMPLEMENTADO** | [`core/mcp.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/mcp.go) / `-mcp` / `/api/mcp` |
| **Logs estructurados con `slog`** | IA / DevOps | Alto | Muy Bajo | ✅ **IMPLEMENTADO** | [`core/logger.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/logger.go) / `-log-json` |
| **OpenAPI 3.1 & Prometheus** | IA / DevOps | Alto | Muy Bajo | ✅ **IMPLEMENTADO** | [`ui/observability.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/ui/observability.go) / `/api/openapi.json` / `/metrics` |
| **Apertura de puertos UPnP IGD** | UX | Muy Alto | Medio | ✅ **IMPLEMENTADO** | [`adapters/upnp.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/adapters/upnp.go) / `-upnp` |
| **Persistencia de Identidad en disco** | UX / Core | Alto | Bajo | ✅ **IMPLEMENTADO** | [`core/keystore.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/keystore.go) / `-key persistent` |
| **Transferencia Drag & Drop en Web UI** | UX | Alto | Medio | ✅ **IMPLEMENTADO** | [`ui/assets/index.html`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/ui/assets/index.html) |
| **Multiplexación QUIC Nativa en Túneles** | Rendimiento | Muy Alto | Alto | ⏳ En Roadmap | [`core/tunnel.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/tunnel.go) |
| **Códec H.264 para Escritorio Remoto** | Rendimiento | Alto | Alto | ⏳ En Roadmap | [`core/remotedesktop_windows.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/remotedesktop_windows.go) |
| **Handshake Noise_XX con PFS** | Seguridad | Muy Alto | Alto | ⏳ En Roadmap | [`core/handshake.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core/handshake.go) |
| **TUI interactiva (Bubbletea)** | UX | Medio | Medio | ⏳ En Roadmap | `cmd/chat/` |
| **Systray en segundo plano** | UX | Medio | Bajo | ⏳ En Roadmap | `cmd/node/` |

