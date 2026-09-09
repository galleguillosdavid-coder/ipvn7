# 📚 Hub Central de Documentación IPv7

Bienvenido al centro oficial de documentación del protocolo descentralizado **IPv7**. Toda la base de conocimiento técnico, arquitectónico, operativo y agéntico del sistema se encuentra centralizada en este directorio [`docs/`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs) organizada en subdirectorios temáticos.

---

## 🗺️ Mapa de Navegación por Módulos

```text
docs/
├── arquitectura/           # Fundamentos criptográficos, ruteo y diseño de red
├── operaciones/            # Manuales de instalación, dashboards y supervisión
├── auditorias_y_reportes/  # Informes de tests, pruebas de malla y auditoría Kùzu
├── sugerencias_mejora/     # Propuestas de evolución, UX, IA y snippets de código
├── skill_ia/               # Especificación para agentes autónomos LLM / MCP
└── indices/                # Catálogo de símbolos, paquetes y grafo de código
```

---

## 1. 🏛️ Arquitectura y Fundamentos (`docs/arquitectura/`)
Especificaciones esenciales sobre la topología del protocolo, criptografía y enrutamiento:
- [**01. Plan Génesis de Arquitectura**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/arquitectura/01_genesis.md): Visión, modelo de capas, axiomas del protocolo y diseño de paquetes CBOR.
- [**02. Enrutamiento en Mundo Pequeño (12 Anillos)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/arquitectura/02_mundo_pequeno_routing.md): Algoritmo voraz XOR, acotamiento a 120 peers y convergencia en $O(\log N)$.
- [**03. Modelado de Grafo en Kùzu**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/arquitectura/03_kuzu_mesh_graph.md): Esquema de datos de nodos y aristas en Kùzu Graph Engine.
- [**Ruta de Trabajo P2P / DHT (PDF)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/arquitectura/Ruta_de_trabajo_IPv7_P2P_DHT.pdf): Especificación técnica original de la arquitectura de transporte y DHT.

---

## 2. ⚙️ Operaciones y Puesta en Marcha (`docs/operaciones/`)
Guías prácticas para administradores de nodos y usuarios finales:
- [**README de Operaciones**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/README.md): Resumen de la suite operativa para operadores.
- [**01. Conceptos y Fundamentos para Humanos**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/01_conceptos_y_fundamentos_para_humanos.md): Analogías y principios clave (criptografía como dirección).
- [**02. Instalación y Puesta en Marcha**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/02_instalacion_puesta_en_marcha.md): Compilación nativa, flags CLI y variables de entorno.
- [**03. Guía del Dashboard Web Interactivo**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/03_guia_dashboard_web.md): Uso de la interfaz SPA, radar de nodos y monitor de enlaces.
- [**04. Casos de Uso Detallados Paso a Paso**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/04_casos_de_uso_detallados.md): Transferencia de archivos, túneles TCP/UDP y streaming.
- [**05. Resolución de Problemas y FAQ**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/05_resolucion_problemas_y_faq.md): Diagnóstico de conectividad, NAT simétrico y relays.
- [**06. Guía Operativa de Supervisión en WSL2**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/06_supervision_wsl2.md): Ejecución y monitoreo de nodos en entornos Linux/WSL2.
- [**07. Acceso Remoto de Notebooks**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/operaciones/07_acceso_remoto_notebook.md): Configuración de conexiones móviles y enlaces dinámicos.

---

## 3. 🔍 Auditorías y Reportes (`docs/auditorias_y_reportes/`)
Evaluación rigurosa del rendimiento, cobertura de código y topología en malla:
- [**README de Auditorías**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/auditorias_y_reportes/README.md): Resumen ejecutivo de calidad y benchmarking.
- [**01. Auditoría de Pruebas Unitarias y Benchmarks**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/auditorias_y_reportes/01_auditoria_pruebas_unitarias_y_benchmarks.md): Tiempos de ejecución criptográficos y rendimiento de paquetes.
- [**02. Pruebas en Vivo de Malla y Servicios**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/auditorias_y_reportes/02_pruebas_en_vivo_red_malla_y_servicios.md): Conectividad dual Windows <-> WSL2 y telemetría.
- [**03. Auditoría del Grafo Kùzu y Topología**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/auditorias_y_reportes/03_auditoria_grafo_kuzu_y_topologia.md): Verificación de índices y reglas Small-World.
- [**04. Catálogo de Oportunidades de Mejora**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/auditorias_y_reportes/04_catalogo_oportunidades_de_mejora.md): Áreas de optimización identificadas.
- [**05. Plan de Acción y Roadmap Evolutivo**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/auditorias_y_reportes/05_plan_de_accion_y_roadmap_evolutivo.md): Fases de desarrollo y matriz de impacto.
- [**06. Auditoría Integral Kùzu Graph 2026**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/auditorias_y_reportes/06_auditoria_kuzu_completa_2026.md): Auditoría completa automatizada mediante consultas Cypher en `.kuzu_index/`.
- [**07. Análisis Crítico del Checklist ChatGPT (posibles.md)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/auditorias_y_reportes/07_analisis_checklist_posibles.md): Contraste riguroso entre el checklist conceptual y la implementación real en Go.
- [**08. Fase de Perfeccionamiento Full (5 Objetivos)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/auditorias_y_reportes/08_fase_perfeccionamiento_completa.md): Resultados empíricos de cuellos de botella, PMTU dinámico, caos, PQC híbrido y benchmarks.
- [**09. Matriz de Validación Experimental de Producción (EPV)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/auditorias_y_reportes/09_matriz_produccion_experimental.md): Congelamiento oficial del Core, rigor empírico, matriz Direct/Relay x LAN/WAN/NAT x UDP/QUIC y telemetría estandarizada.
- [**10. Validación de Resiliencia, Caos y Direct vs. Relay**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/auditorias_y_reportes/10_validacion_resiliencia_chaos_y_wan.md): Batería de estrés físico (pérdida 1-20%, jitter 50ms, caída de proceso SIGKILL y cuantificación nuclear Direct 199.9 Mbps vs Relay 34.5 Mbps).
- [**11. Batería de Validación Adversarial Extrema y Resistencia en Fallo**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/auditorias_y_reportes/11_bateria_adversarial_y_resistencia_extrema.md): 18 vectores de ciberataque, caos, fuzzing y saturación sin alterar el Core. 100.000 replays bloqueados, 0 panics, 0 crashes, 0 security fails y 100% auto-recuperación.

---

## 4. 💡 Sugerencias de Mejora y Evolución (`docs/sugerencias_mejora/`)
Propuestas de ingeniería, diseño y prototipos:
- [**README de Sugerencias**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/sugerencias_mejora/README.md): Roadmap técnico y mejoras propuestas.
- [**01. Mejoras de Experiencia de Usuario (UX)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/sugerencias_mejora/01_mejoras_experiencia_usuario_ux.md): Notificaciones de escritorio, audio/video y dashboard responsivo.
- [**02. Mejoras para Inteligencia Artificial (AI)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/sugerencias_mejora/02_mejoras_inteligencia_artificial_ai.md): Integración agéntica y optimizaciones para LLMs.
- [**03. Mejoras de Arquitectura y Rendimiento**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/sugerencias_mejora/03_mejoras_arquitectura_y_rendimiento.md): Noise Protocol, compresión de paquetes y multipath.
- [**04. Prototipos y Snippets de Código en Go**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/sugerencias_mejora/04_prototipos_y_snippets_codigo.md): Implementaciones de referencia para futuras fases.

---

## 5. 🤖 Suite para Agentes de IA (`docs/skill_ia/`)
Recursos de integración para LLMs, servidores MCP y agentes autónomos:
- [**SKILL.md de IA**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/skill_ia/SKILL.md): Manifiesto del asistente autónomo para IPv7.
- [**System Prompt de IA**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/skill_ia/system_prompt_ia.md): Reglas de comportamiento y contexto de red.
- [**Referencia Formal de APIs REST & WebSockets**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/skill_ia/api_reference_ai.md): Especificación de contratos JSON y CBOR.
- [**Catálogo de Consultas Cypher para Kùzu**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/skill_ia/kuzu_cypher_agent_guide.md): Consultas predefinidas para inspección topológica.
- [**Schemas de Function Calling**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/skill_ia/function_calling_schemas.json): JSON Schemas para OpenAI / Anthropic / Gemini Tools.
- [**Workflows Automatizados de Diagnóstico (SOP)**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/skill_ia/workflows_diagnostico_ia.md): Procedimientos guiados para autorecuperación.

---

## 6. 📊 Índices Semánticos del Código (`docs/indices/`)
- [**PROJECT_INDEX.md**](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/indices/PROJECT_INDEX.md): Catálogo exhaustivo de todos los paquetes, archivos, structs, interfaces y funciones exportadas del proyecto generado automáticamente.
