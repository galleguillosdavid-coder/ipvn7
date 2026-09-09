# IPv7 Telemetry — Modelo Relacional en Kùzu Graph Database

**Documento**: `docs/telemetry/KUZU_MODEL.md`  
**Motor**: Kùzu Graph Engine (`tools/kuzu/kuzu.exe`)

---

## 1. Esquema de Tablas de Nodos y Relaciones (DDL Cypher)

```cypher
// 1. Tablas de Nodos
CREATE NODE TABLE IF NOT EXISTS TelemetryNode(node_id STRING, PRIMARY KEY (node_id));
CREATE NODE TABLE IF NOT EXISTS TelemetryPeer(peer_id STRING, PRIMARY KEY (peer_id));
CREATE NODE TABLE IF NOT EXISTS TelemetrySession(session_id STRING, transport STRING, PRIMARY KEY (session_id));
CREATE NODE TABLE IF NOT EXISTS TelemetryEndpoint(addr STRING, PRIMARY KEY (addr));
CREATE NODE TABLE IF NOT EXISTS TelemetryEvent(event_id STRING, event_type STRING, timestamp INT64, rtt_ms DOUBLE, loss_pct DOUBLE, PRIMARY KEY (event_id));

// 2. Tablas de Relaciones
CREATE REL TABLE IF NOT EXISTS HAS_PEER(FROM TelemetryNode TO TelemetryPeer);
CREATE REL TABLE IF NOT EXISTS HAS_SESSION(FROM TelemetryNode TO TelemetrySession);
CREATE REL TABLE IF NOT EXISTS USES_ENDPOINT(FROM TelemetrySession TO TelemetryEndpoint);
CREATE REL TABLE IF NOT EXISTS GENERATED(FROM TelemetryNode TO TelemetryEvent);
CREATE REL TABLE IF NOT EXISTS AFFECTED(FROM TelemetryEvent TO TelemetrySession);
```

---

## 2. Consultas Analíticas Clave en Cypher

### ¿Qué nodos cambiaron de endpoint durante las últimas 24 horas?
```cypher
MATCH (n:TelemetryNode)-[:GENERATED]->(e:TelemetryEvent)
WHERE e.event_type = 'ENDPOINT_CHANGED'
RETURN n.node_id, e.timestamp, e.event_id
ORDER BY e.timestamp DESC;
```

### ¿Qué sesiones E2EE operaron a través de Relay DERP?
```cypher
MATCH (s:TelemetrySession)
WHERE s.transport = 'RELAY'
RETURN s.session_id, s.transport;
```

### ¿Qué eventos precedieron a una anomalía o caída de ruta?
```cypher
MATCH (n:TelemetryNode)-[:GENERATED]->(e:TelemetryEvent)-[:AFFECTED]->(s:TelemetrySession)
WHERE e.event_type = 'ANOMALY_DETECTED' OR e.event_type = 'DIRECT_PATH_LOST'
RETURN n.node_id, e.event_type, e.rtt_ms, s.session_id;
```
