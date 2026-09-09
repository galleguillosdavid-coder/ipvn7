# Guía de Consultas Cypher con Kùzu Graph Engine para Agentes de IA

El proyecto IPv7 cuenta con una base de datos de grafos embebida **Kùzu** (`.kuzu_index/`). Esta guía provee las consultas estructuradas en Cypher para que los agentes de IA puedan auditar el código fuente del proyecto y la topología de la malla de red P2P.

---

## 1. Esquema del Grafo en Kùzu

### Grafo de Código:
- **Nodos**:
  - `(p:Package)`: Paquetes Go (`core`, `adapters`, `dht`, `ui`, `cmd`).
  - `(f:File)`: Archivos fuente de código (`node.go`, `e2ee.go`, etc.).
  - `(s:Symbol)`: Símbolos declarados (structs, interfaces, funciones, constantes).
- **Relaciones**:
  - `(Package)-[:CONTAINS]->(File)`
  - `(File)-[:DEFINES]->(Symbol)`
  - `(File)-[:IMPORTS]->(File)`

### Grafo de Red (Mesh Topology):
- **Nodos**:
  - `(p:Peer)`: Atributos: `id` (string), `endpoint` (string), `is_local` (bool).
- **Relaciones**:
  - `(p1:Peer)-[:CONNECTED_TO {adapter: string, latency_ms: int64, encrypted: bool}]->(p2:Peer)`

---

## 2. Invocación de Consultas

La IA puede invocar Kùzu mediante CLI o a través de la API REST del nodo:
- **Vía CLI (PowerShell)**:
  ```powershell
  echo "MATCH (p:Peer) RETURN p.id;" | .\tools\kuzu\kuzu.exe .kuzu_index/
  ```
- **Vía API REST**:
  ```bash
  curl -X POST http://localhost:8080/api/kuzu/query \
    -H "Content-Type: application/json" \
    -d '{"query": "MATCH (p:Peer) RETURN p.id, p.is_local;"}'
  ```

---

## 3. Catálogo de Consultas Útiles para Agentes de IA

### A. Auditoría de Red P2P

1. **Obtener la lista de todos los peers conectados al nodo local:**
   ```cypher
   MATCH (local:Peer {is_local: true})-[c:CONNECTED_TO]->(remote:Peer)
   RETURN remote.id AS peer_id, remote.endpoint AS ip_port, c.latency_ms AS rtt, c.adapter AS transporte;
   ```

2. **Identificar enlaces de alta latencia (> 100 ms) para optimización:**
   ```cypher
   MATCH (a:Peer)-[c:CONNECTED_TO]->(b:Peer)
   WHERE c.latency_ms > 100
   RETURN a.id, b.id, c.latency_ms, c.adapter
   ORDER BY c.latency_ms DESC;
   ```

3. **Comprobar si existe enlace directo o ruta entre dos peers específicos:**
   ```cypher
   MATCH (a:Peer {id: $idA})-[c:CONNECTED_TO*1..3]->(b:Peer {id: $idB})
   RETURN count(c) > 0 AS existe_ruta;
   ```

### B. Inspección y Navegación del Código Go

1. **Listar todos los archivos pertenecientes a un paquete:**
   ```cypher
   MATCH (p:Package {name: 'core'})-[:CONTAINS]->(f:File)
   RETURN f.name AS archivo
   ORDER BY f.name;
   ```

2. **Buscar la definición de un struct o función específica en todo el proyecto:**
   ```cypher
   MATCH (f:File)-[:DEFINES]->(s:Symbol)
   WHERE s.name = 'SmallWorldRoutingTable' OR s.name = 'CascadeNode'
   RETURN f.name AS archivo, s.name AS simbolo, s.kind AS tipo;
   ```

3. **Encontrar todas las interfaces declaradas en el core:**
   ```cypher
   MATCH (p:Package {name: 'core'})-[:CONTAINS]->(f:File)-[:DEFINES]->(s:Symbol {kind: 'interface'})
   RETURN f.name, s.name;
   ```

4. **Calcular la densidad y métricas globales del repositorio:**
   ```cypher
   MATCH (f:File) WITH count(f) AS total_archivos
   MATCH (s:Symbol)
   RETURN total_archivos, count(s) AS total_simbolos;
   ```
