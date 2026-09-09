# IPv7 Engineering: Gestión de Baselines Inmutables

## 1. Qué Hace y Qué No Hace

### Qué Hace
- Almacena mediciones de rendimiento de referencia inmutables, ancladas a un commit de Git y a un perfil de hardware específico.
- Permite detectar de forma automática regresiones de rendimiento en microbenchmarks y pruebas de sistema.
- Registra throughput, latencias percentiles (P50, P95, P99), tasa de handshakes por core, consumo de memoria heap y alocaciones por operación.

### Qué No Hace
- **No sobreescribe baselines históricos**. Cada línea base genera un archivo JSON con timestamp y commit hash (`tools/ipv7-engineer/baseline/LAN_WIFI_<hash>_<timestamp>.json`).
- **No asume valores de laboratorio como universales**. Un baseline en LAN Wi-Fi no es comparable con un baseline WAN o Loopback.

---

## 2. Estructura de un Baseline

```json
{
  "id": "LAN_WIFI_8d5aa93_20260909_130000",
  "commit_hash": "8d5aa93...",
  "environment": "LAN_WIFI_INTEL_I7",
  "created_at": "2026-09-09T13:00:00Z",
  "metrics": {
    "handshake_rate_per_core": 1850.0,
    "crypto_seal_ns_per_op": 542.0,
    "crypto_open_ns_per_op": 618.0,
    "packet_pipeline_allocs": 0,
    "packet_pipeline_bytes_op": 0,
    "relay_saturation_pps": 4200.0,
    "relay_saturation_mbps": 8.6,
    "telemetry_overhead_ns": 27.18,
    "roaming_switch_p50_ms": 1.10,
    "roaming_switch_p99_ms": 2.37
  },
  "parameters": {
    "go_version": "go1.23.0",
    "os": "windows/amd64",
    "cpu": "Intel Core i7"
  }
}
```

---

## 3. Cómo se Utiliza

### Consultar Baselines Existentes
```powershell
go run .\tools\ipv7-engineer baseline
```
Salida esperada:
```text
=== BASES DE RENDIMIENTO INMUTABLES (BASELINES) ===
ID: LAN_WIFI_8d5aa93_20260909_130000 | Commit: 8d5aa93... | Env: LAN_WIFI_INTEL_I7
  - Handshake Rate/Core: 1850.00 /s
  - Crypto Seal: 542.00 ns/op
  - Crypto Open: 618.00 ns/op
  - Relay Saturation: 4200.00 PPS (8.60 Mbps)
  - Telemetry Overhead: 27.18 ns/op
  - Roaming Switch P50: 1.10 ms
```

### Crear un Nuevo Baseline
Se crea programáticamente llamando a `baselineManager.SaveBaseline(b)` al concluir un ciclo completo de benchmarks oficiales tras un release o cambio significativo aprobado por David.

---

## 4. Detección de Regresiones

La herramienta `baseline.go` compara cualquier métrica actual contra el baseline inmutable más reciente del mismo entorno:

| Métrica | Umbral de Alerta de Regresión |
| :--- | :--- |
| **Latencia Crypto / Packet** | > +5% de degradación |
| **Alocaciones de Heap** | > 0 allocs/op en el fast-path (Tolerancia Cero) |
| **Throughput / PPS** | > -10% de reducción de capacidad |
| **Tasa de Handshakes** | > -10% de caída |
| **Overhead Telemetría** | > 50 ns/op |

Si se sobrepasa cualquiera de estos umbrales, el pipeline de CI/CD o el ingeniero debe suspender la integración hasta subsanar la causa raíz.

---

## 5. Líneas Base de Recursos y Estabilidad (Soak & Multi-Host)

| Parámetro / Entorno | Valor de Referencia Inmutable | Fuente / Validación |
| :--- | :--- | :--- |
| **Crecimiento de Heap en Soak (3.8h)** | **<= 0.01 MB / hora** (0.03 MB en 3.8h) | `EXP-SOAK-01` (`CANARY_RUNNER_EVALUATION.md`) |
| **Goroutines en Steady-State** | **8 goroutines** (concurrencia constante) | `canary_runner.go` (454 ciclos continuos) |
| **Memoria SO Windows (Host B Notebook)**| **57.6 MB** (`ipv7-node.exe` en Win 11 Home) | `EXP-PHYSICAL-NB-01` (`192.168.1.106`) |
| **Latencia de Handshake LAN Wi-Fi** | **4.78 ms - 5.05 ms** | Noise XX sobre 802.11 ac/ax |
| **Latencia RTT Transporte Wi-Fi** | **0.51 ms** | ICMP / Direct UDP |

