# IPv7 Telemetry — Especificación de Schema de Eventos

**Documento**: `docs/telemetry/SCHEMA.md`  
**Estándar**: RFC 8949 (CBOR) / JSON Canónico  

---

## 1. Estructura del Evento (`TelemetryEvent`)

```json
{
  "timestamp": 1788950000000000000,
  "node_id": "DID:IPv7:ED25519:731C89...A5DC",
  "event_type": "ENDPOINT_CHANGED",
  "peer_id": "DID:IPv7:ED25519:FCE322...B143",
  "session_id": "SES_8f12a4c9",
  "old_endpoint": "192.168.1.106:7001",
  "new_endpoint": "10.0.0.37:7001",
  "route": "DIRECT_P2P",
  "transport": "DIRECT",
  "pmtu": 1280,
  "rtt_ms": 3.97,
  "loss_pct": 0.0,
  "recovery_ms": 1.10,
  "metadata": {
    "network_if": "Wi-Fi 6",
    "roaming_cause": "SSID_SWITCH"
  }
}
```

---

## 2. Invariantes del Esquema
1. **Identidad Soberana**: `node_id` y `peer_id` son claves públicas Ed25519 / DIDs verificables.
2. **Campos Opcionales Eficientes**: Los campos no utilizados en un evento particular (`peer_id`, `old_endpoint`, etc.) se omiten de la serialización (`omitempty`), minimizando el payload.
3. **Cero Impacto de Serialización**: La emisión de eventos utiliza buffers reutilizables vía `sync.Pool`.
