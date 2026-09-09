# IPv7 Telemetry — Manual de Operaciones y Diagnóstico

**Documento**: `docs/telemetry/OPERATIONS.md`

---

## 1. Verificación Rápida de Estado

### Consultar estado de salud (Health Check):
```powershell
curl http://127.0.0.1:9101/health
```
Respuesta esperada:
```json
{
  "backpressure_mode": "ACTIVE",
  "goroutines": 10,
  "heap_mb": 0.31,
  "node_id": "NODE_CL_HOST",
  "peers": 3,
  "status": "HEALTHY",
  "telemetry_dropped_cnt": 0
}
```

---

## 2. Diagnóstico de Cuellos de Botella en Tiempo Real

1. **¿El problema es de socket UDP o del Core?**
   - Monitorear `ipv7_socket_queue_drops_total`: Si aumenta mientras `ipv7_cpu_percent` es bajo, el cuello está en el buffer `SO_RCVBUF` del sistema operativo (solución: verificar que el adapter use socket dedicado, ver ADV-01).
2. **¿El Relay DERP está saturado?**
   - Monitorear `ipv7_backpressure_drops_total`: Si el relay supera los 4.200 PPS, comenzará a aplicar descarte probabilístico temprano para proteger la memoria.
3. **¿Hubo evento de Roaming de IP?**
   - Consultar `ipv7_endpoint_changes_total`: Muestra los cambios de interfaz de red o Wi-Fi completados con éxito sin romper la identidad DID.
