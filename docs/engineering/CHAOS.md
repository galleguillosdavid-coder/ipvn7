# IPv7 Engineering: Batería de Caos y Pruebas Adversariales

## 1. Qué Hace y Qué No Hace

### Qué Hace
- Somete a la implementación de IPv7 a condiciones hostiles extremas:
  - Corrupción bit a bit de paquetes (1–20%).
  - Truncamiento y fuzzing de longitudes inválidas.
  - Replay masivo de paquetes antiguos (1 a 100.000 copias).
  - Paquetes CBOR inválidos / encabezados malformados.
  - Inyección de jitter variable (1 a 500 ms) y ráfagas de pérdida (burst loss).
  - Flood adversarial concurrente de paquetes maliciosos mezclados con tráfico legítimo.
  - Terminación violenta de procesos (SIGKILL) y recuperación inmediata de peers.
- Registra el comportamiento exacto ante fallos para certificar que el sistema falla de manera segura (`fail-safe`), sin panics ni desbordes de memoria.

### Qué No Hace
- **No ejecuta ataques fuera del entorno de prueba autorizado ni contra hosts externos**.
- **No altera reglas de firewall de forma destructiva o irreversible**.

---

## 2. Resumen de Resultados de la Batería Adversarial (Reportes 11 y 12)

| Prueba Adversarial | Intensidad Ensayada | Comportamiento Observado | Clasificación |
| :--- | :--- | :--- | :--- |
| **Ataque de Replay** | 100.000 paquetes duplicados | 1 mensaje válido entregado, 99.999 descartados | **PASS** (Criptográficamente Seguro) |
| **Fuzzing de Entrada** | Longitudes aleatorias, CBOR corrupto | Rechazados silenciosamente, 0 panics | **PASS** (Robusto) |
| **Handshake Flood** | 1.500 intentos concurrentes | Rechazo por rate limit; goroutines estabilizadas | **PASS** (Sin fugas) |
| **SIGKILL + Reanudar**| Kill inmediato del proceso | Reconstrucción de sesión automática en < 100 ms | **PASS** (Resiliente) |
| **Firmas Inválidas** | Paquetes firmados con claves apócrifas | 100% de descarte en el punto de entrada | **PASS** (Integridad Total) |
| **Caos Combinado** | Pérdida + Jitter + Corrupción simultánea | Sesión degradada pero recuperada sin crash | **DEGRADED (Esperado)** |
| **Fuego Cruzado** | 4.995 paquetes hostiles concurrentes | 468/1.000 legítimos procesados | **BOTTLENECK-ADV-01** (Saturación Kernel) |

---

## 3. Principio de Falla Segura (Fail-Safe Principle)

Bajo estrés extremo o saturación, IPv7 debe garantizar:
1. **Invariante Criptográfica Inviolable**: Ningún paquete corrupto, mal firmado o rejugado puede ser entregado a la capa de aplicación.
2. **Estabilidad de Goroutines y Memoria**: La tasa de goroutines y heap debe mantenerse acotada; la saturación debe resolverse mediante descarte de paquetes (`backpressure/drop`), nunca acumulando buffers infinitos.
3. **Persistencia de Identidad**: Un peer que cae y reinicia debe retomar la comunicación sin perder su DID Ed25519.

---

## 4. Cómo Reproducir las Pruebas de Caos

Para ejecutar la batería adversarial completa:
```powershell
go test -v -count=1 ./tools/adversarial/...
```
Para ejecutar pruebas de soak prolongadas:
```powershell
go test -v -timeout=30m -run=TestSoak ./tools/adversarial/...
```
