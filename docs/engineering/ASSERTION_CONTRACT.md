# Contrato Formal de Aserciones e Invariantes de Horizonte 5
## Especificación Vinculante entre Documentación, Código y Tests

> **ESTADO**: **CONTRATO VINCULANTE (Fase 2 de la Auditoría ChatGPT)**.  
> **PROPÓSITO**: Definir la relación estricta `HIPÓTESIS -> INVARIANTE -> MÉTRICA -> COTA -> ASSERTION` para garantizar que ningún test pueda pasar por falsos positivos ni por degradación silenciosa de umbrales en el futuro.

---

## 1. Principio del Contrato

En IPv7, **ninguna métrica cuantitativa mencionada en la documentación puede quedar sin una aserción fatal (`t.Fatalf`) en el código de prueba**. 

Si un test permite un margen más permisivo que el afirmado en la documentación (por ejemplo, permitir 80% cuando se promete 96%), la afirmación documental se considera **no demostrada formalmente** hasta que el test endurezca su aserción o la documentación ajuste su afirmación a la cota real.

---

## 2. Tabla Contractual de Aserciones de Horizonte 5

```text
┌──────────────────┬──────────────────────────────────────────┬─────────────────────────────┬─────────────────┬─────────────────────────────────────────┐
│ HIPÓTESIS        │ INVARIANTE A DEMOSTRAR                   │ MÉTRICA EVALUADA            │ COTA CONTRACTUAL│ ASERCIÓN FORMAL (GO TEST)               │
├──────────────────┼──────────────────────────────────────────┼─────────────────────────────┼─────────────────┼─────────────────────────────────────────┤
│ H-MULTI-01       │ Invarianza de DID en corte WAN           │ bytes.Equal(idA, idA_post)  │ true            │ if !bytes.Equal(...) { t.Fatalf(...) }  │
│ H-MULTI-01       │ Continuidad de sesión E2EE sin handshake │ core.DecryptE2EE(privD, pkt)│ err == nil      │ if err != nil { t.Fatalf(...) }         │
│ H-MULTI-01       │ PDR post-blackout en radio ad-hoc        │ received / sent             │ == 1.0 (100%)   │ if received != sent { t.Fatalf(...) }   │
│ H-MULTI-01       │ Cota de tiempo de reconvergencia L2      │ time.Since(cutStart)        │ < 250 ms        │ if dur > 250*time.Millisecond { t.Fatal}│
├──────────────────┼──────────────────────────────────────────┼─────────────────────────────┼─────────────────┼─────────────────────────────────────────┤
│ H-L2-DOS         │ No colapso ante 10k balizas forjadas     │ panic / crash               │ 0 panics        │ Completitud normal del test             │
│ H-L2-DOS         │ Disponibilidad de tráfico legítimo       │ deliveredLegit / sentLegit  │ >= 0.90 (90%)   │ if delivered < uint64(sent*0.90) {t.Fail}│
│ H-L2-DOS         │ Capacidad máxima acotada de tabla L2     │ len(beacon.DiscoveredPeers) │ <= 256 peers    │ if len(peers) > 256 { t.Fatalf(...) }   │
│ H-L2-DOS         │ Crecimiento de memoria heap acotado      │ deltaHeapKB                 │ <= 512 KB       │ if deltaHeapKB > 512 { t.Fatalf(...) }  │
├──────────────────┼──────────────────────────────────────────┼─────────────────────────────┼─────────────────┼─────────────────────────────────────────┤
│ H-FLAPPING       │ Resistencia a 30 ciclos de oscilación    │ recordedTransitions         │ >= 60 trans     │ if trans < 60 { t.Fatalf(...) }         │
│ H-FLAPPING       │ Convergencia determinista de modo        │ switcher.CurrentMode()      │ OffGrid / Hybrid│ if mode != expected { t.Fatalf(...) }   │
│ H-FLAPPING       │ Ausencia de fugas de goroutines          │ deltaGoroutines (post-settle│ <= 1 goroutine  │ if deltaG > 1 { t.Fatalf(...) }         │
├──────────────────┼──────────────────────────────────────────┼─────────────────────────────┼─────────────────┼─────────────────────────────────────────┤
│ H-SPLIT-BRAIN    │ Resolución transversal bilateral         │ foundB1 && foundA1          │ true            │ if !foundB1 || !foundA1 { t.Fatalf(...) }│
│ H-SPLIT-BRAIN    │ Inviolabilidad de firmas DHT Ed25519     │ rec.Verify()                │ true            │ if !rec.Verify() { t.Fatalf(...) }      │
│ H-SPLIT-BRAIN    │ Métrica XOR decreciente                  │ XOR(A1, B1) > XOR(Bridge, B)│ Monotónica      │ if distBridge >= distOrig { t.Fatalf }  │
├──────────────────┼──────────────────────────────────────────┼─────────────────────────────┼─────────────────┼─────────────────────────────────────────┤
│ H-CONSTRAINED-MTU│ Rechazo estricto de paquetes > MTU       │ link.Send(largePacket)      │ ErrExceedsMTU   │ if err != ErrExceedsMTU { t.Fatalf }    │
│ H-CONSTRAINED-MTU│ Número exacto de fragmentos L2 (180B)    │ len(fragments)              │ == 8 fragmentos │ if len(frags) != 8 { t.Fatalf(...) }    │
│ H-CONSTRAINED-MTU│ Integridad criptográfica bajo desorden   │ SHA256(reconstructed)       │ == SHA256(orig) │ if shaRec != shaOrig { t.Fatalf(...) }  │
├──────────────────┼──────────────────────────────────────────┼─────────────────────────────┼─────────────────┼─────────────────────────────────────────┤
│ H-UNIFIED-E2E    │ Derivación ULA IPv6 de DID Ed25519       │ DeriveIPv6FromDID(id)       │ Prefijo fd07::  │ if !bytes.HasPrefix(...) { t.Fatalf }   │
│ H-UNIFIED-E2E    │ Longitud fija Sphinx 3 saltos concéntricos len(onionPacket)         │ == 1280 bytes   │ if len != 1280 { t.Fatalf(...) }        │
│ H-UNIFIED-E2E    │ Entrega íntegra de TUN A a TUN B (E2E)   │ SHA256(deliveredTunB)       │ == SHA256(orig) │ if shaDeliv != shaOrig { t.Fatalf }     │
└──────────────────┴──────────────────────────────────────────┴─────────────────────────────┴─────────────────┴─────────────────────────────────────────┘
```

---

## 3. Cláusulas de Limitación Explícita (Lo que NO se Afirma)

1. **Reconvergencia L2**: No se afirma que la reconvergencia física sea un valor absoluto de $100.35$ ms en hardware real; se afirma que en laboratorio controlado la topología ad-hoc converge en menos de $250$ ms.
2. **Propagación Inter-Islas en Split-Brain**: La sanación demostrada corresponde al **intercambio explícito de tablas de enrutamiento a través del puente**. No se afirma la existencia de un protocolo de descubrimiento autónomo background `FIND_NODE` que opere sin conocimiento de la existencia del puente.
3. **Canales Estrechos (MTU = 180B)**: El motor de fragmentación y reensamblaje maneja desorden y jitter, pero **carece de capa de retransmisión selectiva L2 (ARQ)**. La pérdida física de un fragmento causará descarte por timeout.
4. **Crecimiento de Heap en DoS ($H\text{-L2-DOS}$)**: `DEMONSTRATED`: Bajo este test y entorno (`LAB_SIMULATED`), el crecimiento neto de `HeapAlloc` tras 10.000 balizas forjadas permanece $\le 512\text{ KB}$ (observado $\approx 293\text{ KB}$). No se afirma que $512\text{ KB}$ sea una propiedad universal e invariable de IPv7 ante cualquier carga o arquitectura en hardware físico, dado que `HeapAlloc` depende del runtime de Go y del entorno de ejecución.
5. **Concurrencia en Flapping ($H\text{-FLAPPING}$)**: `DEMONSTRATED`: No se observó crecimiento residual de goroutines superior a 1 ($\Delta \le 1$, observado $\Delta = +0$) tras una tormenta de 30 oscilaciones de conmutador en laboratorio. `runtime.NumGoroutine()` actúa como alarma de regresión determinista en laboratorio; no constituye una prueba matemática absoluta de ausencia de fugas en cualquier condición arbitraria.

---

## 4. Dictamen Epistémico Vinculante

> [!IMPORTANT]
> **FÓRMULA CANÓNICA DE REPRODUCIBILIDAD (Aprobada por Auditoría)**:  
> *"La cadena `documentación → contrato → test → assertion → resultado` es reproducible bajo el entorno y las condiciones de prueba declaradas."*

