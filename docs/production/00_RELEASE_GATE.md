# IPv7 Production Release Gate — Fase 0: Release Freeze

**Misión**: IPv7 GLOBAL INTERNET PRODUCTION VALIDATION  
**Tag de Versión**: `IPv7-PRODUCTION-CANDIDATE-0`  
**Fecha de Congelación**: 2026-09-09  
**Estado del Protocol Core**: CONGELADO (CORE FREEZE STRICT - No modifications permitted)

---

## 1. Identificación del Entorno y Build

| Parámetro | Valor Registrado |
| :--- | :--- |
| **Commit Base** | `ef9e94d84528ae547be492631b9c86d3d3040763` |
| **Versión de Go** | `go version go1.26.4 windows/amd64` |
| **Sistema Operativo Local** | Windows 11 Pro 64-bit (10.0.26100) / amd64 |
| **Sistema Operativo Remoto (Notebook)** | Windows 11 Home / amd64 (IP: `192.168.1.106`) |
| **Subentorno Linux** | WSL2 Ubuntu 24.04 LTS (Linux 6.6.x x86_64) |
| **Estado de Tests Unitarios** | PASS 100% (`go test -count=1 ./...` verificado sin fallos) |

---

## 2. Hashes Criptográficos de Binarios de Producción (SHA-256)

```text
731c89a9ad161129a7d114d515409cbde96ca71eb2bc0928303e4518bdd4d5dc  ipv7-node.exe (Windows AMD64)
fce322295cb364d7d6ef0d35b5f90e4da674ec4e35c591decf00080594d5b143  bin/ipv7-node-linux (Linux AMD64)
5d0d40bc665316432ae74dfadae5e2dfd49865644ec570e26f353acf2eed9e83  bin/iperf_ipv7.exe (Windows AMD64)
b613529064d4b27f84259e5d00497b741424f359c2b24cc74e31ff2e7db98a8c  bin/iperf_ipv7_linux (Linux AMD64)
```

---

## 3. Reglas Epistémicas y Criterios de Validación

1. **Clasificación Obligatoria de Resultados**:
   Toda afirmación, métrica o conclusión en los reportes debe estar clasificada rigurosamente bajo una de las cinco categorías:
   - `DEMOSTRADO`: Respaldado por traza experimental exhaustiva e inequívoca.
   - `OBSERVADO`: Dato o comportamiento registrado directamente en el entorno de prueba.
   - `INFERIDO`: Deducción analítica basada en datos empíricos correlacionados.
   - `HIPÓTESIS`: Explicación plausible pendiente de contraste experimental.
   - `NO PROBADO`: Escenario o propiedad aún no sometida a prueba empírica.

2. **Prohibición de Términos Absolutos**:
   Queda estrictamente prohibido el uso de afirmaciones no empíricas tales como *"100% seguro"*, *"irrompible"*, *"irrefutable"*, *"sin límites"* o *"worldwide ready"*.

3. **Criterios de Bloqueo de Release**:
   - Panic o Crash no recuperable en cualquier nodo.
   - Fuga de memoria acumulativa confirmada vía `pprof`.
   - Corrupción de datos en tránsito o colapso criptográfico de E2EE.
   - Pérdida de identidad criptográfica (DID) tras reinicio forzado.
   - Imposibilidad de fallback a Relay DERP ante ruptura de camino directo.
