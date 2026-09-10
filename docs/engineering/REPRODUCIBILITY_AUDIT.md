# PROTOCOLO DE AUDITORÍA DE REPRODUCIBILIDAD INDEPENDIENTE
## Guía de Verificación Epistémica de Terceros para el Proyecto IPv7

> **PROPÓSITO DE ESTE DOCUMENTO**:  
> Permitir que cualquier auditor técnico independiente, máquina de CI/CD, desarrollador externo o modelo de IA pueda clonar el repositorio de IPv7 ([galleguillosdavid-coder/Ipv7](https://github.com/galleguillosdavid-coder/Ipv7)), ejecutar la batería de pruebas y certificar empíricamente la veracidad de cada afirmación técnica **sin confiar en testimonios verbales ni resúmenes generativos**.

---

## 1. Principio Fundamental: Verificación vs. Aserción

En el diseño de protocolos de red descentralizados, **la aserción no constituye prueba**. Una arquitectura no es segura, resiliente ni multimodal porque su documentación lo declare, sino porque sus invariantes matemáticas y de estado se sostienen bajo condiciones de estrés destructivo reproducible.

Este repositorio está estructurado para que cualquier tercero pueda ejecutar:

```bash
go test -v ./tests/...
```

y obtener en su propia terminal la validación bit a bit de los resultados descritos en el **Horizonte 5**.

---

## 2. Taxonomía Epistémica en Dos Dimensiones

Todo resultado registrado en la documentación de IPv7 se descompone en **dos variables ortogonales**:

### 2.1. Certeza Formal (`EPISTEMIC_STATUS`)
- **`DEMONSTRATED`**: Demostrado experimentalmente con pruebas deterministas automatizadas bajo las condiciones controladas del ensayo.
- **`OBSERVED`**: Observado durante ejecuciones empíricas (p.ej. telemetría de soak continuo), sujeto a variables estocásticas del sistema operativo.
- **`INFERRED`**: Derivado de propiedades matemáticas o de criptografía estándar (ej. resistencia de X25519 o SHA256), sin ensayo de fuerza bruta local.
- **`HYPOTHESIS`**: Postulado de diseño en espera de instrumentación o medición.
- **`NOT_PROVEN`**: Sin evidencia suficiente; explícitamente excluido de afirmaciones de completitud.

### 2.2. Entorno Operativo (`ENVIRONMENT`)
- **`LAB_SIMULATED`**: Ensayos en memoria virtual, canales sintéticos en RAM (`RadioMedium`, sockets locales virtuales).
- **`LAB_REAL_NETWORK`**: Comunicaciones a través de la pila de red real del sistema operativo (loopback UDP/TCP, sockets Windows 11 ↔ WSL2 Linux).
- **`REAL_HARDWARE`**: Transceptores físicos reales de radio (chips LoRa SX1262, adaptadores Wi-Fi Direct, puertos serie/SPI). **Actualmente clasificado como `NOT_PROVEN`**.
- **`WAN`**: Red pública de Internet (atravesamiento NAT, STUN, BGP). **Actualmente clasificado como `NOT_PROVEN`**.
- **`FIELD`**: Pruebas de campo exteriores con nodos móviles. **Actualmente clasificado como `NOT_PROVEN`**.

---

## 3. Matriz de Mapeo: Afirmación Documentada $\leftrightarrow$ Test Reproducible

A continuación se detalla la correspondencia unívoca entre las afirmaciones del proyecto y los archivos de código ejecutable:

| ID de Hipótesis | Afirmación Técnica Auditada | Archivo de Prueba | Función de Test | Invariante Verificada |
| :--- | :--- | :--- | :--- | :--- |
| **$H\text{-MULTI-01}$** | **Invarianza de Identidad & Continuidad E2EE ante Corte WAN** | [`tests/multimedium_chaos_test.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/tests/multimedium_chaos_test.go) | `TestMultimediumTransportFailover_Chaos` | Corte simultáneo de WAN; reconvergencia $\approx 100$ ms a radio L2; PDR 100%; $DID_A$ y $DID_D$ idénticos bit a bit; clave de sesión X25519 conservada sin re-handshake. |
| **$H\text{-L2-DOS}$** | **Resistencia a Inundación de Balizas L2 (`OG7!`)** | [`tests/l2_dos_and_flapping_test.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/tests/l2_dos_and_flapping_test.go) | `TestBeaconEngine_PoisonAndFloodResistance` | Inyección de 10.000 balizas forjadas a $>800\text{k}$ balizas/s; entrega de paquetes legítimos $\ge 90\%$; tabla L2 acotada a 256 vecinos con LRU ($O(1)$ mem); crecimiento de heap $< 500\text{ KB}$. |
| **$H\text{-FLAPPING}$** | **Resiliencia ante Oscilación Rápida WAN/Malla** | [`tests/l2_dos_and_flapping_test.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/tests/l2_dos_and_flapping_test.go) | `TestHybridSwitcher_FlappingResistance` | 30 ciclos agresivos de corte y reconexión WAN (60 transiciones de modo); cero deadlocks de goroutines; convergencia determinista a cero bloqueos. |
| **$H\text{-SPLIT-BRAIN}$** | **Fusión de Particiones y Convergencia DHT Autónoma** | [`tests/split_brain_healing_test.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/tests/split_brain_healing_test.go) | `TestSplitBrainAndHealing_MeshConvergence` | Dos islas aisladas operando de forma autónoma; enlace puente inter-islas; resolución transversal bilateral $A_1 \leftrightarrow B_1$; métrica XOR decreciente monotónicamente; firmas Ed25519 intactas. |
| **$H\text{-CONSTRAINED-MTU}$** | **Fragmentación L2 y Reensamblado Bajo Caos** | [`tests/constrained_link_fragmentation_test.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/tests/constrained_link_fragmentation_test.go) | `TestConstrainedLink_...` | Enlace con MTU físico de 180B; rechazo estricto con `ErrPacketExceedsMTU`; fragmentación transparente de paquete de 1280B en 8 partes; entrega desordenada estocástica; SHA256 final idéntico bit a bit (0 corrupción). |
| **$H\text{-UNIFIED-E2E}$** | **Pipeline Unificado de los Cuatro Horizontes** | [`tests/e2e_four_horizons_test.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/tests/e2e_four_horizons_test.go) | `TestFourHorizons_UnifiedEndToEnd` | TUN (IPv6 ULA `fd07::`) $\to$ DHT (Búsqueda Kademlia 256 bits) $\to$ Onion Sphinx (1280B fijo en 3 saltos) $\to$ Off-Grid L2 Radio $\to$ TUN Destino. Integridad criptográfica SHA256 y 100% PDR. |

---

## 4. Procedimiento de Auditoría Paso a Paso

### Paso 1: Clonar el Repositorio
```bash
git clone https://github.com/galleguillosdavid-coder/Ipv7.git
cd Ipv7
```

### Paso 2: Verificar la Regla Inviolable de Core Freeze
El núcleo del protocolo (`core/`) debe mantenerse inmutable respecto a las extensiones modulares (`adapters/` y `dht/`):
```bash
git log -n 5 -- core/
```
*(Debe comprobarse que no existen alteraciones arbitrarias en las primitivas fundamentales).*

### Paso 3: Ejecución de la Suite Completa de Reproducibilidad
Ejecutar la suite de validación cruzada:
```bash
go test -v ./tests/...
```

O para verificar el repositorio completo:
```bash
go test ./...
```

**Criterio de Aceptación para el Auditor**:
- Salida final: `PASS` en todos los paquetes (`ipv7/adapters`, `ipv7/adapters/offgrid`, `ipv7/adapters/onion`, `ipv7/adapters/tun`, `ipv7/core`, `ipv7/dht`, `ipv7/tests`, `ipv7/ui`).
- Ningún test debe requerir conexión a Internet pública ni servicios externos en la nube (como Firebase o servidores STUN propietarios).
- Ningún test debe requerir permisos de superusuario (`root`/administrador), ya que todos los módulos de red admiten abstracciones de laboratorio controladas.

---

## 5. Scripts Canónicos de Auditoría Automatizada

Para facilitar la ejecución de un solo comando en cualquier plataforma, el repositorio proporciona scripts dedicados que compilan, ejecutan y formatean el informe tabular de auditoría:

### En Windows (PowerShell):
```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\audit_reproducibility.ps1
```

### En Linux / macOS / WSL (Bash):
```bash
chmod +x ./scripts/audit_reproducibility.sh
./scripts/audit_reproducibility.sh
```

---

## 6. Auditoría del Soak Test Canario de Larga Duración

El repositorio incluye un arnés de prueba de estabilidad continua de 24 horas ubicado en [`canary/runner/canary_runner.go`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/canary/runner/canary_runner.go).

### Verificación del Soak en Ejecución:
El auditor puede reproducir una prueba de estabilidad canaria de 5 o 10 minutos directamente:
```bash
go run canary/runner/canary_runner.go -duration=5m -interval=10s
```
O para pruebas continuas de 24 horas:
```bash
go run canary/runner/canary_runner.go -duration=24h -interval=60s
```

### Criterios de Evaluación del Soak:
1. **Packet Delivery Ratio (PDR)**: Debe mantenerse en $\ge 99.0\%$. En las ejecuciones de referencia se registra $99.43\%$ PDR tras $>180.000$ paquetes transmitidos.
2. **Explicación de Paquetes Descartados**: El $\approx 0.57\%$ de descartes no debe ser ocultado ni enmascarado como 100% artificial. Ocurre por saturación transitoria de búfer UDP en el kernel del host durante ráfagas de alta frecuencia, lo cual es normal en redes no congestionadas sin control de flujo TCP.
3. **Consumo de Memoria Heap**: Debe mantenerse plano (meseta en $\approx 0.36\text{ MB} \pm 0.05\text{ MB}$), certificando empíricamente la ausencia de fugas de memoria en los bucles de transporte.
4. **Recuento de Goroutines**: Debe mantenerse constante ($\le 12$ goroutines activas), confirmando que las transiciones de conexión no dejan hilos huérfanos.

---

## 7. Conclusión para el Auditor Externo

Si todos los pasos anteriores se ejecutan con éxito en su entorno:

1. Las propiedades de **multimedio, invariancia de identidad y failover** descritas en Horizonte 5 quedan formalmente demostradas en `LAB_SIMULATED`.
2. Las propiedades de **red física real (`REAL_HARDWARE`) y escala global (`WAN`)** permanecen clasificadas honestamente como **`NOT_PROVEN`** hasta que existan transceptores de hardware físico y pruebas de campo documentadas con este mismo nivel de rigor reproducible.
