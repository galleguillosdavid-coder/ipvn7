# MISIÓN: IPv7 Telemetry & Living Network v1

## 0. CONTEXTO

Estamos desarrollando IPv7 como una arquitectura de red overlay experimental implementada actualmente en Go.

IPv7 funciona sobre redes existentes y debe mantener separado:

* identidad del nodo;
* endpoint/IP actual;
* transporte;
* routing;
* sesiones;
* observabilidad.

El proyecto ya posee componentes de observabilidad, Prometheus, logs estructurados y utilización de Kùzu para representar información relacionada con el sistema/red.

Estamos actualmente en una fase de **validación de producción/canary**, con el Core congelado.

Por lo tanto:

> **NO debes modificar el comportamiento fundamental del Protocol Core salvo que sea estrictamente necesario y esté demostrado por evidencia.**

La misión de esta tarea es construir una **capa de telemetría extremadamente eficiente y desacoplada del camino crítico**, capaz de permitirnos observar IPv7 en funcionamiento real durante horas o días.

---

# 1. OBJETIVO

Implementar:

## `IPv7 Telemetry v1`

Una capa que permita responder, mediante datos reales:

* ¿Qué está haciendo cada nodo?
* ¿Qué está haciendo cada sesión?
* ¿Qué está haciendo cada ruta?
* ¿Qué ocurre cuando cambia la IP?
* ¿Cuándo aparece pérdida?
* ¿Cuándo aparece jitter?
* ¿Cuándo cambia el PMTU?
* ¿Cuándo se utiliza relay?
* ¿Cuánto cuesta un handshake?
* ¿Dónde aparece backpressure?
* ¿Dónde aparece un cuello de botella?
* ¿Qué ocurre durante churn?
* ¿Qué ocurre durante fallos y recuperación?
* ¿Qué componentes están consumiendo CPU/memoria?
* ¿Qué comportamiento es normal y cuál es anómalo?

La finalidad NO es generar enormes cantidades de logs.

La finalidad es:

> **obtener la máxima información útil con el mínimo costo posible.**

---

# 2. PRINCIPIO FUNDAMENTAL

La observabilidad NO debe convertirse en un nuevo cuello de botella.

La arquitectura debe ser:

```text
                    IPv7
                      │
       ┌──────────────┴──────────────┐
       │                             │
   tráfico normal              Telemetry API
       │                             │
       │                       eventos/métricas
       │                             │
       │                      buffer local
       │                             │
       │                       agregación
       │                             │
       │              ┌──────────────┴──────────────┐
       │              │                             │
       ▼              ▼                             ▼
    network       Prometheus                     Kùzu
```

Nunca:

```text
IPv7 packet
    ↓
Kùzu
    ↓
continue processing
```

Kùzu, Prometheus, logging o cualquier exportador externo NO deben bloquear el procesamiento de paquetes.

Si el sistema de observabilidad desaparece:

> IPv7 debe continuar funcionando.

---

# 3. PRIMERA TAREA: INSPECCIONAR EL REPOSITORIO

Antes de escribir código:

1. Inspecciona toda la arquitectura existente.
2. Identifica:

   * Core;
   * SessionManager;
   * routing;
   * transport;
   * adapters;
   * relay;
   * PMTU;
   * handshake;
   * crypto;
   * discovery;
   * Prometheus;
   * logging;
   * Kùzu;
   * WebSocket/API;
   * self-healing;
   * tests existentes.
3. Identifica cualquier sistema de métricas ya existente.
4. Identifica cualquier exportador Kùzu existente.
5. No dupliques funcionalidades.
6. Reutiliza componentes existentes cuando sea correcto.

Antes de implementar, genera mentalmente un mapa:

```text
COMPONENTE → DATOS DISPONIBLES → MÉTRICA/EVENTO POSIBLE
```

No inventes APIs ni estructuras si ya existe una equivalente.

---

# 4. TELEMETRÍA: CUATRO CATEGORÍAS

Implementa cuatro niveles.

## A. NODE METRICS

Métricas agregadas del nodo:

```text
uptime_seconds
cpu_usage
memory_bytes
heap_bytes
goroutines
peers
active_sessions
active_routes
direct_sessions
relay_sessions
```

---

## B. TRANSPORT METRICS

```text
packets_rx
packets_tx
bytes_rx
bytes_tx

packet_loss
rtt
jitter
pmtu

pps_rx
pps_tx
throughput_rx
throughput_tx

socket_queue_pressure
backpressure_events

endpoint_changes
direct_to_relay
relay_to_direct
```

No calcules jitter mediante procesamiento pesado por paquete.

Utiliza agregación eficiente.

---

## C. PROTOCOL METRICS

```text
handshakes_started
handshakes_completed
handshakes_failed

session_created
session_closed

replay_rejected
malformed_rejected
signature_rejected
aead_rejected

route_lookup
route_success
route_failure
route_change

pmtu_probe
pmtu_change

peer_join
peer_leave
```

---

## D. EVENTS

Los eventos deben ser pocos y significativos.

Implementa inicialmente:

```text
NODE_START
NODE_STOP

PEER_JOINED
PEER_LEFT

HANDSHAKE_STARTED
HANDSHAKE_COMPLETED
HANDSHAKE_FAILED

SESSION_CREATED
SESSION_CLOSED

ENDPOINT_CHANGED

ROUTE_CHANGED

PMTU_CHANGED

DIRECT_PATH_LOST
RELAY_ENTERED
RELAY_EXITED

BACKPRESSURE_DETECTED

SOCKET_PRESSURE
ANOMALY_DETECTED

RECOVERY_COMPLETED
```

No conviertas cada paquete en un evento persistente.

---

# 5. FORMATO DE EVENTO

Crear un formato estructurado, preferentemente reutilizando el sistema de serialización existente.

Conceptualmente:

```text
TelemetryEvent

timestamp
node_id
event_type
peer_id
session_id
old_endpoint
new_endpoint
route
transport
pmtu
rtt
loss
metadata
```

No copies campos innecesarios.

Debe ser compacto.

Los campos opcionales deben realmente ser opcionales.

---

# 6. IDENTIDAD Y PRIVACIDAD

Nunca utilizar la IP como identidad primaria.

El nodo debe identificarse mediante su identidad IPv7/DID existente.

Ejemplo conceptual:

```text
NODE:
DID/NodeID = X

ENDPOINT:
192.168.1.106:7001

```

Después:

```text
NODE:
DID/NodeID = X

ENDPOINT:
10.0.0.37:7001
```

Esto permitirá observar experimentalmente:

> mismo nodo + diferente endpoint.

No registrar información innecesaria.

---

# 7. MÉTRICAS DE MOVILIDAD

Esto es especialmente importante.

Cuando el endpoint cambia:

```text
ENDPOINT_CHANGED
```

registrar:

```text
node_id
peer_id
old_endpoint
new_endpoint
timestamp
recovery_time
packet_loss
rtt_before
rtt_after
pmtu_before
pmtu_after
route_before
route_after
direct_or_relay
```

El objetivo es poder estudiar:

```text
WiFi A
  ↓
WiFi B
  ↓
hotspot
  ↓
router diferente
```

sin perder la relación entre todos los eventos.

---

# 8. MÉTRICAS DE CUELLO DE BOTELLA

Crear métricas que permitan detectar dónde se está acumulando trabajo.

Especialmente:

```text
receive queue
send queue
socket pressure
processing latency
handshake latency
crypto latency
routing latency
serialization latency
relay queue
relay backpressure
```

Cuando sea posible, medir duración mediante histogramas o buckets, no generando un evento por operación.

Queremos poder responder:

> ¿El tiempo se está gastando en Core, crypto, routing, socket, relay o I/O?

---

# 9. PROMETHEUS

Si Prometheus ya existe:

* reutilizarlo;
* no crear un segundo sistema paralelo.

Si faltan métricas, agregarlas.

Preferir:

```text
Counter
Gauge
Histogram
```

según corresponda.

Evitar labels de cardinalidad ilimitada.

MUY IMPORTANTE:

NO utilizar:

```text
IP
session_id
timestamp
random ID
```

como labels de Prometheus si eso puede producir cardinalidad explosiva.

Las métricas deben permanecer agregables.

---

# 10. KÙZU

Kùzu NO debe estar en el camino crítico.

Su función será representar relaciones.

Modelo conceptual:

```text
Node
Peer
Session
Route
Endpoint
Relay
Event
Experiment
```

Relaciones:

```text
Node ──HAS_PEER──> Node

Node ──HAS_SESSION──> Session

Session ──USES_ENDPOINT──> Endpoint

Node ──USES_RELAY──> Relay

Node ──ROUTES_TO──> Node

Node ──GENERATED──> Event

Event ──AFFECTED──> Session

Event ──AFFECTED──> Route
```

NO implementes necesariamente todo esto si el repositorio ya posee un modelo equivalente.

Primero reutiliza.

El objetivo es poder consultar posteriormente cosas como:

```text
¿Qué nodos cambiaron de endpoint durante las últimas 24 horas?

¿Qué rutas cambiaron después de un endpoint change?

¿Qué peers están utilizando relay?

¿Qué nodos presentan mayor churn?

¿Qué rutas presentan mayor RTT?

¿Qué eventos preceden a una pérdida?

¿Dónde aparece backpressure?
```

---

# 11. BUFFERING

La telemetría debe utilizar buffering.

Preferiblemente:

```text
lock-free / low-lock
bounded buffer
batch export
```

según lo que resulte natural para el código existente.

Nunca permitir crecimiento ilimitado.

Si el buffer se llena:

```text
DROP_TELEMETRY
```

y registrar una métrica:

```text
telemetry_dropped_total
```

Pero:

> **Nunca bloquear tráfico IPv7 solamente para preservar una métrica.**

---

# 12. COSTO

Crear un benchmark:

```text
Telemetry OFF
Telemetry ON
Telemetry ON + Prometheus
Telemetry ON + Kùzu
```

Comparar:

```text
CPU
RAM
throughput
pps
latency
allocations
GC
goroutines
```

Objetivo:

> demostrar experimentalmente que la observabilidad no introduce un costo significativo.

No inventar un porcentaje objetivo antes de medir.

Primero establecer baseline.

Después determinar si el overhead es aceptable.

---

# 13. ANOMALÍAS

Crear una capa sencilla de detección.

NO implementar todavía machine learning.

NO implementar IA dentro del Core.

Inicialmente detectar solamente condiciones explícitas:

```text
RTT spike
packet loss spike
PMTU decrease
handshake failure spike
relay backpressure
socket pressure
route failure spike
endpoint change
memory growth
goroutine growth
```

Emitir:

```text
ANOMALY_DETECTED
```

con:

```text
anomaly_type
severity
observed_value
baseline_value
timestamp
node
peer
```

La interpretación avanzada será responsabilidad del sistema externo/IA.

---

# 14. INTEGRACIÓN CON `ipv7-engineer`

Preparar la telemetría para que posteriormente un Skill denominado:

```text
ipv7-engineer
```

pueda consumirla.

No es necesario implementar todavía una IA.

Pero crear una salida estructurada que permita entregar:

```text
EXPERIMENT
ENVIRONMENT
BASELINE
CURRENT
EVENTS
ANOMALIES
BOTTLENECK_CANDIDATES
RECOVERY
```

a otra herramienta.

---

# 15. EXPERIMENTO OBLIGATORIO: IP ROAMING

Crear una prueba automatizada o herramienta que permita registrar:

```text
endpoint A
↓
disconnect
↓
endpoint B
↓
recovery
```

Debe registrar:

```text
old IP
new IP
same NodeID/DID
session state
route state
recovery time
packet loss
RTT
PMTU
direct/relay
```

No asumir que el roaming funciona.

**Medirlo.**

---

# 16. EXPERIMENTO OBLIGATORIO: SOCKET CONTENTION

Reproducir el experimento que previamente mostró:

```text
shared UDP socket
        ↓
hostile traffic
        ↓
legitimate traffic degradation
```

y compararlo contra:

```text
public/discovery socket
+
dedicated transport socket
```

Registrar:

```text
hostile PPS
legitimate PPS
legitimate PDR
socket queue pressure
CPU
memory
drops
```

El objetivo es mantener documentada la evidencia que llevó a separar los adapters/sockets.

---

# 17. NO HACER

NO:

* cambiar criptografía;
* cambiar routing;
* cambiar wire format;
* cambiar identidad;
* cambiar handshake;
* cambiar PMTU;
* cambiar Core;
* agregar blockchain;
* agregar ML;
* agregar complejidad de seguridad;
* agregar mecanismos especulativos;
* agregar features solamente porque "podrían ser útiles".

Esta tarea es:

> **OBSERVABILITY FIRST**

---

# 18. TESTS

Crear tests unitarios para:

```text
event creation
serialization
aggregation
buffer overflow
metric correctness
endpoint change
anomaly detection
Kùzu persistence
Prometheus exposition
```

Crear integración para:

```text
Node → telemetry
Session → telemetry
Route → telemetry
Endpoint change → telemetry
Relay → telemetry
```

Después:

```bash
go test -count=1 ./...
```

Debe continuar pasando.

---

# 19. BENCHMARK

Crear:

```text
benchmark_telemetry_test.go
```

Comparar:

```text
baseline
telemetry enabled
telemetry + Prometheus
telemetry + Kùzu
```

No declarar que algo es "rápido" sin números.

---

# 20. DOCUMENTACIÓN

Crear:

```text
docs/telemetry/
```

con:

```text
README.md
SCHEMA.md
METRICS.md
EVENTS.md
KUZU_MODEL.md
PROMETHEUS.md
PERFORMANCE.md
OPERATIONS.md
```

La documentación debe ser autocontenida.

Debe explicar:

1. qué es IPv7 Telemetry;
2. por qué existe;
3. qué mide;
4. qué NO mide;
5. cómo funciona;
6. cómo consultar los datos;
7. cómo interpretar eventos;
8. cómo detectar cuellos de botella;
9. cómo ejecutar pruebas;
10. cómo verificar el overhead.

---

# 21. REPORTE FINAL

Al terminar, generar:

```text
docs/telemetry/IMPLEMENTATION_REPORT.md
```

con:

```text
IMPLEMENTED
NOT_IMPLEMENTED
REUSED_COMPONENTS
NEW_COMPONENTS
CORE_MODIFICATIONS
PERFORMANCE_OVERHEAD
TEST_RESULTS
BENCHMARK_RESULTS
KÙZU_INTEGRATION
PROMETHEUS_INTEGRATION
KNOWN_LIMITATIONS
NEXT_EXPERIMENTS
```

Clasificar cada conclusión como:

```text
DEMONSTRATED
OBSERVED
INFERRED
HYPOTHESIS
NOT_PROVEN
```

---

# 22. GIT

Antes de modificar:

```text
git status
git branch
git log -5
```

No modificar commits históricos.

No eliminar funcionalidades existentes.

No hacer cambios destructivos.

Si la implementación es correcta:

```text
git diff
go test -count=1 ./...
```

Revisar cuidadosamente el diff.

Si se modificó el Core más de lo estrictamente necesario:

> detenerse y revisar.

No realizar commit si existen regresiones.

---

# 23. CRITERIO DE ÉXITO

La misión NO se considera terminada porque "compila".

Debe demostrar:

```text
[ ] métricas node
[ ] métricas transport
[ ] métricas protocol
[ ] eventos
[ ] endpoint roaming
[ ] relay
[ ] PMTU
[ ] routing
[ ] socket pressure
[ ] backpressure
[ ] Prometheus
[ ] Kùzu
[ ] bounded buffering
[ ] no blocking del Core
[ ] anomaly detection básica
[ ] benchmark de overhead
[ ] tests
[ ] documentación
```

Y especialmente:

> **Debemos poder mirar los datos de IPv7 funcionando durante horas y posteriormente responder dónde estuvo el problema sin tener que adivinar mirando el código.**

---

# 24. PRINCIPIO FINAL

Este proyecto sigue una regla:

> **No optimizar antes de encontrar el cuello de botella.**

Y otra:

> **No implementar una solución para un problema que todavía no hemos demostrado que existe.**

Por lo tanto, si durante esta misión descubres que una funcionalidad propuesta aquí ya existe, reutilízala.

Si descubres que una funcionalidad no es necesaria:

```text
NOT_NEEDED
```

y documenta por qué.

Si descubres un problema nuevo:

```text
OBSERVED
```

no lo conviertas inmediatamente en:

```text
FACT
```

Primero crea un experimento reproducible.

La red debe producir la evidencia.

**Tu trabajo no es hacer IPv7 más complejo.**

Tu trabajo es conseguir que IPv7 nos permita **ver exactamente qué está ocurriendo dentro de la red mientras funciona en el mundo real.**

Ese será el fundamento para el siguiente paso:

```text
OBSERVE
    ↓
FIND BOTTLENECK
    ↓
PROVE CAUSE
    ↓
OPTIMIZE
    ↓
MEASURE AGAIN
```

No optimizar nada que no esté respaldado por medición.


# MISIÓN: IPv7 ENGINEER v1

## PROPÓSITO

Ya existe o está siendo implementada una capa de telemetría para IPv7.

Ahora queremos construir sobre ella un sistema de **ingeniería experimental asistida por IA**.

El objetivo no es que una IA modifique IPv7 indiscriminadamente.

El objetivo es crear un ciclo:

```text
OBSERVAR
   ↓
DETECTAR
   ↓
CORRELACIONAR
   ↓
FORMULAR HIPÓTESIS
   ↓
DISEÑAR EXPERIMENTO
   ↓
EJECUTAR
   ↓
MEDIR
   ↓
COMPARAR CON BASELINE
   ↓
DECIDIR
   ↓
OPTIMIZAR SI ESTÁ JUSTIFICADO
   ↓
REGRESIÓN
   ↓
VOLVER A OBSERVAR
```

La evidencia experimental tiene prioridad sobre cualquier suposición de diseño.

---

# 1. PRIMERA REGLA: INSPECCIONAR ANTES DE PROGRAMAR

No asumas que la misión anterior implementó exactamente lo descrito.

Inspecciona el repositorio actual.

Busca:

* telemetry;
* Prometheus;
* Kùzu;
* logs;
* métricas;
* eventos;
* benchmarks;
* production validation;
* watchdog;
* experiments;
* existing AI/agent tooling;
* scripts;
* APIs;
* CLI;
* documentación.

Reutiliza lo existente.

No dupliques sistemas.

Si ya existe una solución mejor que la propuesta aquí, úsala.

---

# 2. QUÉ ES `ipv7-engineer`

`ipv7-engineer` será un sistema/herramienta de ingeniería, no una parte del protocolo de red.

Debe poder:

1. leer evidencia;
2. identificar anomalías;
3. comparar contra baseline;
4. encontrar correlaciones;
5. formular hipótesis;
6. proponer experimentos;
7. ejecutar experimentos seguros;
8. comparar resultados;
9. detectar regresiones;
10. producir informes;
11. preparar información para análisis externo por otra IA.

Debe mantenerse separado del Protocol Core.

---

# 3. TRES ROLES

El sistema debe reconocer tres roles:

```text
DAVID
Arquitecto / decisión final

ANTIGRAVITY
Ejecutor / programador / operador

CHATGPT
Analista externo / segunda opinión / crítica
```

El software no debe asumir que una IA tiene autoridad absoluta.

La evidencia es la autoridad.

---

# 4. ESTADOS DE CONOCIMIENTO

Toda conclusión debe clasificarse:

```text
DEMONSTRATED
OBSERVED
INFERRED
HYPOTHESIS
NOT_PROVEN
```

Ejemplo:

```text
OBSERVED:
Relay throughput cae después de cierto nivel.

INFERRED:
Existe presión/backpressure en el relay.

HYPOTHESIS:
La cola TCP es el cuello de botella.

NOT_PROVEN:
Que aumentar buffers resolverá el problema.
```

Nunca transformar automáticamente una hipótesis en hecho.

---

# 5. BASELINE

El sistema debe mantener baselines.

Ejemplo:

```text
baseline/
    LAN
    WAN
    RELAY
    ROAMING
    LOSS
    JITTER
    CHURN
    HANDSHAKE
```

Cada baseline debe incluir:

```text
commit
version
OS
CPU
RAM
Go version
configuration
test parameters
timestamp
metrics
results
```

Nunca comparar resultados sin conocer el entorno.

---

# 6. DETECCIÓN DE ANOMALÍAS

Usar primero reglas simples y explicables.

Detectar:

```text
RTT increase
packet loss increase
jitter increase
throughput decrease
PPS decrease
PMTU decrease
handshake failures
route failures
relay backpressure
socket pressure
memory growth
goroutine growth
recovery degradation
endpoint roaming failures
```

No introducir Machine Learning todavía salvo que exista evidencia de que realmente aporta valor.

---

# 7. CORRELACIÓN

Cuando aparezca una anomalía, no concluir inmediatamente.

Buscar relaciones.

Ejemplo:

```text
throughput ↓
CPU estable
RAM estable
RTT ↑
socket queue ↑
```

Resultado:

```text
BOTTLENECK_CANDIDATE:
socket/transport

CONFIDENCE:
medium
```

Otro:

```text
throughput ↓
CPU ↑
crypto latency ↑
socket normal
```

Resultado:

```text
BOTTLENECK_CANDIDATE:
crypto/CPU
```

La herramienta debe intentar localizar el componente responsable antes de proponer modificaciones.

---

# 8. MAPA DE CUELLOS DE BOTELLA

Crear un registro persistente:

```text
bottlenecks/
```

Cada problema:

```text
ID
STATUS
COMPONENT
SYMPTOM
EVIDENCE
CAUSE
CONFIDENCE
EXPERIMENTS
MITIGATION
REGRESSION
```

Estados:

```text
UNKNOWN
SUSPECTED
UNDER_INVESTIGATION
CONFIRMED
MITIGATED
RESOLVED
REOPENED
```

---

# 9. EXPERIMENT ENGINE

Crear una estructura para experimentos.

Ejemplo:

```text
Experiment {
    ID
    Objective
    Hypothesis
    Baseline
    Variables
    Environment
    Procedure
    SafetyLimits
    Metrics
    ExpectedResult
    ActualResult
    Conclusion
}
```

Cada experimento debe ser reproducible.

---

# 10. EXPERIMENTOS CONTROLADOS

Antes de modificar código:

```text
baseline
    ↓
experiment
    ↓
measurement
```

Si se modifica código:

```text
baseline
    ↓
change
    ↓
unit tests
    ↓
integration tests
    ↓
benchmark
    ↓
real experiment
```

Nunca declarar una optimización exitosa únicamente porque el benchmark local mejoró.

---

# 11. OPTIMIZATION ENGINE

Una optimización solamente puede proponerse cuando:

```text
1. existe un problema medido;
2. existe evidencia de su causa;
3. existe una hipótesis de solución;
4. puede reproducirse;
5. existe baseline;
6. existe rollback.
```

Ejemplo:

```text
PROBLEM:
Relay saturates.

EVIDENCE:
Backpressure begins around measured threshold.

HYPOTHESIS:
Resource X limits relay throughput.

EXPERIMENT:
Change only X.

SUCCESS:
Higher throughput without unacceptable latency/loss/memory.

FAIL:
Rollback.
```

---

# 12. NO OPTIMIZAR LO QUE YA FUNCIONA

Si una prueba produce:

```text
PASS
```

y no existe un problema observable:

```text
DECISION = NO_CHANGE
```

Esto es una decisión válida.

El sistema debe registrar:

```text
NOT_OPTIMIZED_BECAUSE:
No demonstrated bottleneck.
```

---

# 13. PROTECCIÓN DEL CORE

Antes de modificar cualquier archivo:

clasificarlo:

```text
CORE
TRANSPORT
ADAPTER
RELAY
OBSERVABILITY
TEST
TOOLING
DOCUMENTATION
```

Si un problema puede resolverse fuera del Core:

> preferir la solución fuera del Core.

No modificar:

* wire format;
* identidad;
* criptografía;
* handshake;
* routing;
* PMTU;
* sesión;

simplemente para resolver un problema que pertenece a otro componente.

---

# 14. EXPERIMENTO DE ROAMING

Convertir el experimento manual de David en un experimento formal.

Escenario:

```text
WiFi A
   ↓
IP A
   ↓
disconnect
   ↓
WiFi B / router B / hotspot
   ↓
IP B
   ↓
recovery
```

Registrar:

```text
same_node_id
old_endpoint
new_endpoint
session
route
PMTU
RTT
packet_loss
recovery_time
direct/relay
```

Repetir muchas veces.

Determinar distribución:

```text
min
p50
p95
p99
max
```

No quedarse solamente con el promedio.

---

# 15. CHAOS ENGINE

Construir pruebas controladas para:

```text
IP change
network disconnect
network reconnect
WiFi change
router change
hotspot change
packet loss
jitter
reordering
duplication
relay failure
peer failure
SIGKILL
restart
churn
```

Cada experimento debe tener límites para no destruir el entorno de producción/canary.

---

# 16. OBSERVACIÓN CONTINUA

El watchdog de 24 horas debe integrarse con este sistema.

Durante el soak:

```text
metrics
events
anomalies
memory
goroutines
sessions
routes
endpoint changes
relay usage
```

Al finalizar generar:

```text
SOAK_REPORT
```

con:

```text
baseline
min
max
mean
p50
p95
p99
growth
anomalies
recoveries
failures
```

Especial atención a:

```text
memory slope
goroutine slope
latency drift
throughput drift
route stability
```

No afirmar ausencia absoluta de memory leaks.

Decir:

```text
NO LEAK-COMPATIBLE GROWTH OBSERVED
UNDER TEST CONDITIONS
```

cuando corresponda.

---

# 17. KÙZU COMO MEMORIA ESTRUCTURAL

Utilizar Kùzu para relacionar:

```text
Experiment
Node
Peer
Session
Route
Endpoint
Relay
Metric
Event
Anomaly
Bottleneck
Commit
```

Esto debe permitir consultas del tipo:

```text
¿Qué ocurrió antes de esta anomalía?

¿Qué rutas cambiaron después de un endpoint change?

¿Qué nodos utilizan relay?

¿Qué nodos presentan mayor churn?

¿Qué cambios de código precedieron una degradación?

¿Qué experimentos demostraron que una hipótesis era falsa?

¿Qué optimizaciones realmente mejoraron throughput?
```

Kùzu no debe bloquear el funcionamiento de IPv7.

---

# 18. EL SISTEMA DEBE RECORDAR LOS FRACASOS

Esto es obligatorio.

Una hipótesis falsa es información valiosa.

Ejemplo:

```text
HYPOTHESIS-014

Theory:
Increasing buffer X will improve relay throughput.

Experiment:
FAILED

Result:
No improvement.

Conclusion:
Hypothesis rejected.
```

No volver a proponer automáticamente la misma solución.

---

# 19. INFORME PARA CHATGPT

Crear un exportador que produzca un paquete compacto para análisis externo.

Formato conceptual:

```text
IPv7_ANALYSIS_REQUEST v1

PROJECT
COMMIT

ENVIRONMENT

OBJECTIVE

BASELINE

CURRENT_RESULT

CHANGES

EVENTS

ANOMALIES

BOTTLENECK_CANDIDATES

HYPOTHESES

PREVIOUS_FAILED_HYPOTHESES

QUESTIONS

REQUESTED_ANALYSIS
```

El paquete debe priorizar evidencia sobre logs irrelevantes.

No enviar gigabytes de logs.

Enviar:

```text
resumen
métricas
anomalías
ventanas relevantes
logs relacionados
comparaciones
```

---

# 20. RESPUESTA DEL ANALISTA EXTERNO

Preparar también un formato compatible con:

```text
IPv7_ANALYSIS_RESPONSE v1
```

con:

```text
CLASSIFICATION
OBSERVATIONS
EVIDENCE
INFERENCES
HYPOTHESES
UNCERTAINTIES
BOTTLENECK
CONFIDENCE
RECOMMENDED_EXPERIMENT
RISK
CORE_CHANGE_REQUIRED
```

La herramienta debe poder almacenar esa respuesta como parte del historial experimental.

---

# 21. REGLA DE DOS IA

Cuando Antigravity y el analista externo estén en desacuerdo:

```text
NO AUTO-MERGE
```

Generar:

```text
DISAGREEMENT
```

y solicitar un experimento que permita resolverlo.

No decidir por autoridad.

Resolver mediante evidencia.

---

# 22. REGRESIÓN AUTOMÁTICA

Cada optimización debe conservar:

```text
before
after
```

y comparar:

```text
throughput
latency
jitter
loss
CPU
RAM
allocations
goroutines
handshakes
routing
relay
```

Una optimización que mejora una métrica pero empeora otra debe marcarse:

```text
TRADEOFF_DETECTED
```

No automáticamente:

```text
SUCCESS
```

---

# 23. SCORECARD

Crear un scorecard experimental:

```text
Performance
Reliability
Recovery
Memory
CPU
Routing
Transport
Relay
Roaming
Security
Observability
```

No convertirlo en un único número mágico.

Mostrar métricas individuales.

---

# 24. SEGURIDAD

El sistema de experimentación NO debe ejecutar automáticamente:

* ataques destructivos;
* cambios de firewall;
* cambios de red irreversibles;
* modificaciones del Core sin aprobación;
* acciones que puedan afectar otros sistemas.

Los experimentos de caos deben estar limitados al entorno autorizado.

---

# 25. PRINCIPIO DE AUTONOMÍA CONTROLADA

Antigravity puede:

```text
OBSERVAR
ANALIZAR
PROPONER
CREAR TEST
EJECUTAR TEST SEGURO
BENCHMARK
COMPARAR
DOCUMENTAR
```

Puede modificar tooling, tests y observabilidad cuando corresponda.

Para cambios sustanciales del Core:

```text
PROPOSE
```

y no:

```text
AUTO_APPLY
```

salvo que el flujo existente del proyecto ya tenga una política explícita de aprobación.

---

# 26. RESULTADO FINAL

Al finalizar la implementación debe existir:

```text
tools/ipv7-engineer/
```

o la ubicación arquitectónicamente correcta determinada después de inspeccionar el repositorio.

Debe incluir:

```text
experiment engine
baseline manager
anomaly detector
bottleneck analyzer
report generator
telemetry reader
Kùzu integration
Prometheus integration
external-analysis exporter
regression analyzer
```

No es obligatorio utilizar exactamente esta estructura si el repositorio existente recomienda otra.

---

# 27. DOCUMENTACIÓN

Crear documentación autocontenida:

```text
docs/engineering/
    README.md
    EXPERIMENTS.md
    BASELINES.md
    BOTTLENECKS.md
    KÙZU.md
    EXTERNAL_ANALYSIS.md
    ROAMING.md
    CHAOS.md
    OPTIMIZATION.md
```

Explicar claramente:

```text
qué hace
qué no hace
cómo se utiliza
cómo se reproducen experimentos
cómo se interpretan resultados
cómo se hace rollback
```

---

# 28. CRITERIO DE ÉXITO

No basta con compilar.

Debe demostrarse al menos:

```text
[ ] leer telemetría
[ ] almacenar experimentos
[ ] crear baseline
[ ] detectar anomalía
[ ] identificar candidato a cuello de botella
[ ] generar hipótesis
[ ] ejecutar experimento controlado
[ ] comparar before/after
[ ] detectar regresión
[ ] registrar hipótesis falsa
[ ] consultar Kùzu
[ ] generar reporte
[ ] generar paquete para analista externo
[ ] analizar roaming
[ ] analizar soak
[ ] preservar Core
[ ] mantener todos los tests existentes PASS
```

---

# 29. TEST FINAL

Ejecutar:

```bash
go test -count=1 ./...
```

y todos los benchmarks relevantes.

Después ejecutar un experimento real.

Preferentemente:

```text
WiFi A
→ IP change
→ WiFi B
→ recovery
→ telemetry
→ Kùzu
→ analysis report
```

El resultado debe demostrar que el sistema puede observar el evento completo.

---

# 30. NO INVENTAR

Si falta infraestructura:

```text
BLOCKED
```

Si falta información:

```text
UNKNOWN
```

Si una hipótesis no se puede probar:

```text
NOT_PROVEN
```

No inventar proveedores cloud, servidores, regiones, resultados, benchmarks ni capacidades.

---

# 31. FILOSOFÍA FINAL

Este sistema debe seguir estas reglas:

> **Medir antes de optimizar.**

> **Encontrar la causa antes de modificar.**

> **Preferir la solución más simple que resuelva el problema demostrado.**

> **No añadir complejidad para problemas hipotéticos.**

> **La telemetría debe observar a IPv7, no convertirse en parte de su cuello de botella.**

> **Las hipótesis se prueban; no se convierten en hechos por repetición.**

> **Una optimización sólo existe si los datos demuestran una mejora.**

> **Una hipótesis falsa también es un resultado exitoso.**

> **El Core permanece pequeño mientras la evidencia no obligue a ampliarlo.**

---

# MISIÓN

No construyas simplemente una herramienta de métricas.

Construye el comienzo de un:

# LABORATORIO VIVO DE IPv7

donde:

```text
IPv7 genera evidencia
        ↓
Telemetry captura evidencia
        ↓
Kùzu relaciona evidencia
        ↓
IPv7 Engineer encuentra patrones
        ↓
Antigravity ejecuta experimentos
        ↓
ChatGPT proporciona análisis independiente
        ↓
David decide
        ↓
IPv7 evoluciona
```

La meta final no es producir más código.

La meta es:

> **hacer que cada línea nueva de código de IPv7 tenga una razón experimental para existir.**

Primero inspecciona el repositorio y la implementación actual. Después implementa la solución que mejor encaje con la arquitectura existente, sin forzar esta especificación literalmente cuando el código existente ya tenga una solución superior.
