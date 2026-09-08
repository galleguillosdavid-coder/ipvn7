# Kùzu (Kazu) Graph DB: Modelado de Red en Malla y Grafo de Código

El proyecto IPv7 integra **Kùzu Graph Database** (`tools/kuzu/`), una base de datos embebida de grafos orientada a propiedades con soporte de lenguaje de consulta **Cypher** de alto rendimiento.

Kùzu cumple dos propósitos esenciales:
1. **Grafo de Código y Arquitectura**: Indexa paquetes, archivos, estructuras, interfaces y funciones para que tanto el desarrollador como el asistente AI (Antigravity) naveguen y comprendan el sistema de forma relacional instantánea.
2. **Grafo de Topología de Malla P2P**: Almacena instantáneas de la red (peers, latencias, adaptadores activos y estados de cifrado E2EE).

---

## 1. Esquema del Grafo

### Esquema de Código (.kuzu_index/ipv7.db)
```cypher
-- Nodos
(Package {name: STRING})
(File {path: STRING, name: STRING, loc: INT64})
(Symbol {id: STRING, name: STRING, kind: STRING})

-- Relaciones
(Package)-[:CONTAINS]->(File)
(File)-[:DEFINES]->(Symbol)
```

### Esquema de Red P2P en Malla
```cypher
-- Nodos de red
(Peer {id: STRING, endpoint: STRING, is_local: BOOLEAN})

-- Enlaces de conexión
(Peer)-[:CONNECTED_TO {adapter: STRING, latency_ms: INT64, encrypted: BOOLEAN}]->(Peer)
```

---

## 2. Cómo Ejecutar Consultas Cypher

### Desde Windows
```powershell
# Iniciar consola interactiva de Kùzu
.\tools\kuzu\kuzu.exe .kuzu_index\ipv7.db

# O ejecutar consulta directa por tubería
powershell -Command "'MATCH (p:Package) RETURN p.name;' | .\tools\kuzu\kuzu.exe .kuzu_index\ipv7.db"
```

### Desde WSL2 (Linux)
```bash
./tools/kuzu/kuzu .kuzu_index/ipv7.db
```

---

## 3. Consultas Cypher de Ejemplo

### 1. Ver todos los paquetes y la cantidad de archivos que contienen
```cypher
MATCH (p:Package)-[:CONTAINS]->(f:File)
RETURN p.name, count(f) AS total_files
ORDER BY total_files DESC;
```

### 2. Buscar todas las estructuras (`struct`) definidas en el núcleo `core`
```cypher
MATCH (p:Package {name: 'core'})-[:CONTAINS]->(f:File)-[:DEFINES]->(s:Symbol {kind: 'struct'})
RETURN f.name, s.name;
```

### 3. Buscar funciones asociadas a streaming o cascada
```cypher
MATCH (f:File)-[:DEFINES]->(s:Symbol)
WHERE s.name CONTAINS 'Cascade' OR s.name CONTAINS 'Stream'
RETURN f.name, s.name, s.kind;
```

### 4. Consultar la topología de peers conectados y sus latencias
```cypher
MATCH (a:Peer)-[r:CONNECTED_TO]->(b:Peer)
RETURN a.id, r.adapter, r.latency_ms, b.id
ORDER BY r.latency_ms ASC;
```

---

## 4. Exportar la Malla Activa desde la Interfaz Web

En el Dashboard Web de IPv7 (`http://localhost:8080` o `8082`):
1. Haz clic en la pestaña **🕸️ Topología de Malla**.
2. Haz clic en el botón **📊 Exportar a Kùzu Cypher**.
3. Las sentencias `MERGE` de los peers conectados se copiarán automáticamente al portapapeles listas para pegarse en `kuzu.exe`.

---

## 5. Interfaz Gráfica Visual: Kùzu Graph Studio (Redes & Código)

El proyecto incluye un entorno visual interactivo completo para explorar la base de datos de grafos de Kùzu sin necesidad de usar exclusivamente la terminal:

### Opción A: Embebido en el Dashboard del Nodo IPv7
Al iniciar cualquier nodo (Windows o WSL2 Linux):
- Accede a `http://localhost:8080` (o `http://localhost:8082` en WSL2).
- Selecciona la pestaña **📊 Explorador Kùzu (Red & Código)**.
- Alterna entre:
  - **🌐 Redes IPv7**: Visualiza todos los nodos P2P, canales cifrados E2EE, adaptadores de transporte (QUIC, UDP, WebRTC) y latencias registradas en Kùzu.
  - **📁 Código de Repositorio**: Visualiza la arquitectura del repositorio en forma de grafo (Paquetes $\to$ Archivos $\to$ Structs/Interfaces/Funcs con tamaños proporcionales a las LoC).
- **Consola Cypher Integrada**: Escribe consultas Cypher personalizadas o selecciona consultas predefinidas para ejecutar y ver el resultado tanto en el grafo como en JSON.

### Opción B: Servidor Standalone de Exploración (sin requerir nodo P2P)
Si deseas explorar el código y los grafos de Kùzu sin levantar un nodo de red IPv7:
```powershell
.\tools\kuzu_explorer\run_explorer.ps1
```
O ejecutando directamente:
```powershell
python .\tools\kuzu_explorer\server.py
```
Abre tu navegador en **`http://localhost:8090`**.

