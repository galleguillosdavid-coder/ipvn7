# IPv7 Telemetry — Integración con Prometheus OpenMetrics

**Documento**: `docs/telemetry/PROMETHEUS.md`

---

## 1. Endpoints Expuestos por el Nodo

- **`GET /metrics`**: Servidor de métricas en formato estándar texto Prometheus (OpenMetrics v0.0.4).
- **`GET /health`**: Endpoint JSON de diagnóstico de liveness y readiness para balanceadores o monitores externos.

---

## 2. Regla Anti-Explosión de Cardinalidad

Queda **terminantemente prohibido** utilizar como labels de Prometheus:
- Direcciones IP o puertos efímeros (`endpoint="192.168.1.106:54321"`).
- IDs de sesión aleatorios (`session_id="a8f23..."`).
- Marcas de tiempo o UUIDs.

El único label permitido es el identificador lógico del nodo (`node="NODE_CL_HOST"`), asegurando que las métricas permanezcan siempre agregables sin saturar la memoria del servidor de Prometheus ni del nodo.

---

## 3. Ejemplo de Configuración `prometheus.yml`

```yaml
scrape_configs:
  - job_name: 'ipv7-nodes'
    scrape_interval: 15s
    static_configs:
      - targets:
          - '127.0.0.1:9101' # Chile Host
          - '127.0.0.1:9102' # WSL2 Linux
          - '192.168.1.106:9103' # Notebook Físico
          - '127.0.0.1:9104' # USA East Proxy
          - '127.0.0.1:9105' # EU West Proxy
          - '127.0.0.1:9199' # Relay DERP
```
