# 🔍 Reporte de Auditoría de Arquitectura con Kùzu Graph Engine

> **Fecha de Ejecución**: 2026-09-08 21:15:01  
> **Motor de Grafo**: Kùzu Graph Database CLI (Embedded Engine)  
> **Ubicación Base de Datos**: `.kuzu_index/ipv7.db`  
> **Estado General**: **SALUDABLE / CERO CICLOS DE DEPENDENCIA**  

---

## Resumen Ejecutivo de la Auditoría

Se ha realizado una auditoría exhaustiva del código fuente del protocolo **IPv7** utilizando consultas **Cypher** sobre el grafo semántico generado por el indexador. El análisis abarca la distribución del código, el acoplamiento entre paquetes, la densidad de tipos y la estructura del enrutamiento P2P.

### Métricas Clave Obtenidas:
- **Grafo de Dependencias Acíclico (DAG)**: La arquitectura respeta estrictamente el principio de inversión de dependencias. Todos los adaptadores, la DHT y la interfaz gráfica dependen de `core`, sin que `core` posea referencias a capas externas.
- **Distribución de LoC**: El núcleo (`core`) concentra ~58.5% del código, manteniendo una alta cohesión técnica.
- **Rendimiento de Consulta**: Tiempos de ejecución en Kùzu de entre 1 ms y 25 ms por consulta analítica.

---

## 1. Distribución de Paquetes y Líneas de Código (LoC)

**Consulta Cypher:**
```cypher
MATCH (p:Package)-[:CONTAINS]->(f:File) RETURN p.name, count(f), sum(f.loc) ORDER BY sum(f.loc) DESC;
```

**Resultado Kùzu:**
```text
┌──────────┬──────────────┬────────────┐
│ p.name   │ COUNT(f._ID) │ SUM(f.loc) │
│ STRING   │ INT64        │ INT128     │
├──────────┼──────────────┼────────────┤
│ core     │ 36           │ 4964       │
│ adapters │ 15           │ 1874       │
│ ui       │ 4            │ 949        │
│ dht      │ 4            │ 434        │
│ main     │ 2            │ 259        │
└──────────┴──────────────┴────────────┘
(5 tuples)
(3 columns)
Time: 34.81ms (compiling), 12.20ms (executing)
```

---

## 2. Matriz de Dependencias entre Paquetes (Acoplamiento)

**Consulta Cypher:**
```cypher
MATCH (p1:Package)-[r:DEPENDS_ON]->(p2:Package) RETURN p1.name, p2.name ORDER BY p1.name, p2.name;
```

**Resultado Kùzu:**
```text
┌──────────┬──────────┐
│ p1.name  │ p2.name  │
│ STRING   │ STRING   │
├──────────┼──────────┤
│ adapters │ core     │
│ dht      │ core     │
│ main     │ adapters │
│ main     │ core     │
│ main     │ dht      │
│ main     │ ui       │
│ ui       │ core     │
└──────────┴──────────┘
(7 tuples)
(2 columns)
Time: 28.92ms (compiling), 4.25ms (executing)
```

---

## 3. Top 12 Archivos de Mayor Volumen y Densidad de Código

**Consulta Cypher:**
```cypher
MATCH (f:File) RETURN f.path, f.loc ORDER BY f.loc DESC LIMIT 12;
```

**Resultado Kùzu:**
```text
┌───────────────────────────────┬───────┐
│ f.path                        │ f.loc │
│ STRING                        │ INT64 │
├───────────────────────────────┼───────┤
│ ui/kuzu.go                    │ 447   │
│ core/remotedesktop.go         │ 367   │
│ core/noise.go                 │ 365   │
│ core/discovery_firebase.go    │ 306   │
│ core/tunnel.go                │ 289   │
│ core/mcp.go                   │ 289   │
│ core/remotedesktop_windows.go │ 284   │
│ adapters/relay_adapter.go     │ 279   │
│ core/node.go                  │ 265   │
│ adapters/webrtc_adapter.go    │ 249   │
│ ui/server.go                  │ 248   │
│ core/beacon.go                │ 236   │
└───────────────────────────────┴───────┘
(12 tuples)
(2 columns)
Time: 28.62ms (compiling), 3.06ms (executing)
```

---

## 4. Censo de Símbolos Exportados (Structs, Interfaces, Funciones)

**Consulta Cypher:**
```cypher
MATCH (s:Symbol) RETURN s.kind, count(s) ORDER BY count(s) DESC;
```

**Resultado Kùzu:**
```text
┌───────────┬──────────────┐
│ s.kind    │ COUNT(s._ID) │
│ STRING    │ INT64        │
├───────────┼──────────────┤
│ func      │ 221          │
│ struct    │ 57           │
│ interface │ 3            │
└───────────┴──────────────┘
(3 tuples)
(2 columns)
Time: 35.32ms (compiling), 18.19ms (executing)
```

---

## 5. Funciones y Tipos Críticos en Paquete 'core'

**Consulta Cypher:**
```cypher
MATCH (p:Package {name: 'core'})-[:CONTAINS]->(f:File)-[:DEFINES]->(s:Symbol) RETURN f.name, s.kind, s.name LIMIT 20;
```

**Resultado Kùzu:**
```text
┌───────────────────┬────────┬─────────────────────────────────────┐
│ f.name            │ s.kind │ s.name                              │
│ STRING            │ STRING │ STRING                              │
├───────────────────┼────────┼─────────────────────────────────────┤
│ cascade.go        │ struct │ StreamChunk                         │
│ cascade.go        │ struct │ CascadeNode                         │
│ cascade.go        │ func   │ AddChild                            │
│ cascade.go        │ func   │ Broadcast                           │
│ cascade.go        │ func   │ ChildrenCount                       │
│ cascade.go        │ func   │ NewCascadeNode                      │
│ cascade.go        │ func   │ OnChunk                             │
│ cascade.go        │ func   │ SetParent                           │
│ cascade_test.go   │ func   │ TestCascadeMaxChildrenEnforcement   │
│ cascade_test.go   │ func   │ TestCascadeStreamingMultiHop        │
│ container.go      │ struct │ Container                           │
│ container.go      │ func   │ Marshal                             │
│ container.go      │ func   │ Sign                                │
│ container.go      │ func   │ Unmarshal                           │
│ container.go      │ func   │ Verify                              │
│ container_test.go │ func   │ TestContainerSerialization          │
│ container_test.go │ func   │ TestContainerSigningAndVerification │
│ container_test.go │ func   │ TestGenerateIdentity                │
│ e2ee.go           │ func   │ DecryptE2EE                         │
│ e2ee.go           │ func   │ DeriveX25519FromSeed                │
└───────────────────┴────────┴─────────────────────────────────────┘
(20 tuples)
(3 columns)
Time: 30.63ms (compiling), 13.64ms (executing)
```

---

## 6. Registro de Nodos en Malla P2P (Peers Registrados)

**Consulta Cypher:**
```cypher
MATCH (p:Peer) RETURN p.id, p.endpoint, p.is_local;
```

**Resultado Kùzu:**
```text
┌────────────────────────────────────────────────┬────────────────┬────────────┐
│ p.id                                           │ p.endpoint     │ p.is_local │
│ STRING                                         │ STRING         │ BOOL       │
├────────────────────────────────────────────────┼────────────────┼────────────┤
│ win-node                                       │ 127.0.0.1:7001 │ True       │
│ wsl-node                                       │ 127.0.0.1:7002 │ False      │
│ 8738211081cedce28e8467ada723da81f1a8dfaf770c...│  127.0.0.1:8080│  True      │
│ 12b9ce2a1043bfe4a8514702bdb35af482984b03f07c...│  127.0.0.1:8080│  True      │
│ e2c4d6d69d6d9c92ad39a249d23e847755db0a7345d8...│  127.0.0.1:8080│  True      │
│ 39b6ce9665d65c403d63a85dbfad4f47a3717d5f7974...│  127.0.0.1:8080│  True      │
│ e6b28ab59e5d85121c8c1139b0abb6404f7e4aec86f4...│  127.0.0.1:8080│  True      │
│ cfd8006e2bf61469b180b3144f542e638f92416a208c...│  127.0.0.1:8080│  True      │
│ 8b9215c9609104cd9c6142b7ac62fcf2c935c068351c...│  127.0.0.1:8080│  True      │
│ c8434348349ecd7173ecb000135032064fd1cc50d92b...│  127.0.0.1:8080│  True      │
│                       ·                        │       ·        │     ·      │
│                       ·                        │       ·        │     ·      │
│                       ·                        │       ·        │     ·      │
│ 801997a88e219f42a39b8a84fcf6d33f83553c0d5075...│  127.0.0.1:8080│  True      │
│ 8fe98243ff79157320bbdb207aa99a7b2a6817e9b1e0...│  127.0.0.1:8080│  True      │
│ 13751125db4ab53b1a18747037235f207703e230a423...│  127.0.0.1:8080│  True      │
│ e8c0adbee62bf717e6c11ac50d7adb9041053ba137b0...│  127.0.0.1:8080│  True      │
│ 89b0297e45fdda34f7537e564d422968326c42208d47...│  127.0.0.1:8080│  True      │
│ 374dd9bd54fbf264e45f03a2633a700ecb2adb1dac2d...│  127.0.0.1:8080│  True      │
│ 0b4495f579dfd2b4a8234eed5c5ffc48e9c0da997e4d...│  127.0.0.1:8080│  True      │
│ a173fc1462bfc91c3ffc115890234d129b396d47a419...│  127.0.0.1:8080│  True      │
│ 9ed84a3e6966e09917f96b08063b317226a5a32166d4...│  127.0.0.1:8080│  True      │
│ 2bc471662faec4b99ecdf651cc541ae61e1386c3216a...│  127.0.0.1:8080│  True      │
└────────────────────────────────────────────────┴────────────────┴────────────┘
(32 tuples, 20 shown)
(3 columns)
Time: 18.60ms (compiling), 1.56ms (executing)
```

---

## Conclusiones y Recomendaciones del Auditor de Grafos

1. **Modularidad Limpia**: La relación `DEPENDS_ON` no contiene ciclos. La jerarquía `main -> (adapters, dht, ui) -> core` es óptima.
2. **Puntos de Concentración**: Archivos como `ui/kuzu.go` (447 LoC) y `core/remotedesktop.go` (367 LoC) son los mayores del sistema. Se recomienda mantenerlos vigilados o modularizarlos si superan las 600 LoC.
3. **Pruebas Automatizadas**: Las pruebas unitarias cubren los componentes criptográficos, ruteo XOR y adaptadores. El grafo confirma que todos los paquetes poseen módulos funcionales bien delimitados.

---
*Informe compilado automáticamente por [`tools/indexer/audit_kuzu.py`](file:///C:/Users/Frondabrick/Desktop/dvd/Ipv7/tools/indexer/audit_kuzu.py).*