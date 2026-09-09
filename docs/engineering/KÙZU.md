# IPv7 Engineering: Integración con Kùzu Graph Engine

## 1. Qué Hace y Qué No Hace

### Qué Hace
- Utiliza la base de datos de grafos embebida **Kùzu** como motor de memoria topológica y correlación de eventos en IPv7.
- Modela nodos, peers, sesiones, rutas lógicas, eventos de telemetría y anomalías como nodos y relaciones en un grafo distribuido.
- Permite ejecutar consultas analíticas complejas en lenguaje **Cypher** para rastrear causas raíz de cuellos de botella, rutas degradadas y conmutaciones de roaming.
- Opera de forma **asíncrona** y desacoplada mediante un buffer lock-free y un journal transaccional append-only (`events_journal.jsonl`).

### Qué No Hace
- **NUNCA bloquea el camino crítico de red de paquetes**. Un paquete UDP nunca espera una transacción o consulta Cypher en Kùzu.
- Si el motor Kùzu no está disponible o se detiene, la red IPv7 continúa procesando paquetes al 100% de velocidad sin inmutarse (`fail-silent / fail-safe`).

---

## 2. Esquema Relacional en Grafos (Kùzu Schema)

```text
 (Node:IPv7Node) ──[:ESTABLISHED]──► (Session:SessionNode)
       │                                     │
       │                                     ├──[:USES_ROUTE]──► (Route:RouteNode)
       │                                     │
       ▼                                     ▼
[:PRODUCED_EVENT]                    [:EXPERIENCED_ANOMALY]
       │                                     │
       ▼                                     ▼
(Event:TelemetryEvent)                (Anomaly:AnomalyEvent)
```

### Definiciones DDL (Cypher)

```cypher
CREATE NODE TABLE IPv7Node(
    did STRING,
    current_endpoint STRING,
    os STRING,
    PRIMARY KEY (did)
);

CREATE NODE TABLE SessionNode(
    session_id STRING,
    peer_did STRING,
    created_at INT64,
    PRIMARY KEY (session_id)
);

CREATE NODE TABLE TelemetryEvent(
    id INT64,
    node_did STRING,
    category STRING,
    event_type STRING,
    value DOUBLE,
    timestamp INT64,
    PRIMARY KEY (id)
);

CREATE NODE TABLE AnomalyEvent(
    anomaly_id STRING,
    node_did STRING,
    category STRING,
    description STRING,
    timestamp INT64,
    PRIMARY KEY (anomaly_id)
);

CREATE REL TABLE SESSIONS(FROM IPv7Node TO SessionNode);
CREATE REL TABLE EVENTS(FROM IPv7Node TO TelemetryEvent);
CREATE REL TABLE ANOMALIES(FROM IPv7Node TO AnomalyEvent);
```

---

## 3. Consultas Cypher Analíticas de Diagnóstico

### Rastrear Rutas con Pérdida Anómala (> 5%)
```cypher
MATCH (n:IPv7Node)-[:PRODUCED_EVENT]->(e:TelemetryEvent)
WHERE e.event_type = 'packet_loss' AND e.value > 0.05
RETURN n.did, n.current_endpoint, e.value, e.timestamp
ORDER BY e.timestamp DESC
LIMIT 20;
```

### Correlacionar Roaming con Latencia RTT P99
```cypher
MATCH (n:IPv7Node)-[:PRODUCED_EVENT]->(e1:TelemetryEvent),
      (n)-[:PRODUCED_EVENT]->(e2:TelemetryEvent)
WHERE e1.event_type = 'roaming_switch'
  AND e2.event_type = 'rtt_ms'
  AND abs(e2.timestamp - e1.timestamp) < 5000000000 // ventana de 5 segundos
RETURN n.did, e1.timestamp, e2.value AS rtt_after_roam;
```

---

## 4. Resiliencia y Fallback: Journal Append-Only

Para garantizar que ningún evento de telemetría se pierda incluso si el proceso se termina inesperadamente o el motor Kùzu experimenta bloqueo de I/O, el `KuzuExporter` escribe inmediatamente cada evento a un journal append-only:
```text
telemetry/journal/events_journal.jsonl
```
En el arranque siguiente, la herramienta sincroniza automáticamente el contenido del journal hacia Kùzu en segundo plano.
