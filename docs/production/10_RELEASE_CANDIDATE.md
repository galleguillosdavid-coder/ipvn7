# IPv7 Release Candidate Checklist & Final Verdict — Fase 15

**Misión**: IPv7 GLOBAL INTERNET PRODUCTION VALIDATION  
**Documento**: `10_RELEASE_CANDIDATE.md`

---

## 1. Matriz de Criterios de Aceptación Release Candidate

| Dominio Operativo | Estado | Evidencia Requerida | Clasificación |
| :--- | :---: | :--- | :--- |
| **LAN Física Inter-Máquina** | `PASS` | Pruebas PC-Notebook (89.7 Mbps, RTT 3.97 ms) | `DEMOSTRADO` |
| **WAN Emulada / Degrada** | `PENDING` | Batería de pérdida (0-70%) y jitter (1-1000ms) | `PENDIENTE` |
| **IPv4 Dual-Stack** | `PASS` | Transporte UDP en IPv4 validado | `DEMOSTRADO` |
| **IPv6 Dual-Stack** | `PENDING` | Validación de socket IPv6 nativo | `PENDIENTE` |
| **NAT Traversal & Hole Punching** | `PASS` | Descubrimiento e intercambio directo LAN/WAN | `OBSERVADO` |
| **Relay DERP Fallback** | `PASS` | Tráfico conmutado a través de relay ciego E2EE | `DEMOSTRADO` |
| **PMTU Dynamic Discovery** | `PASS` | Tráfico acotado a 1280 B sin fragmentación IP | `DEMOSTRADO` |
| **Pérdida & Jitter Tolerancia** | `PASS` | PDR 100% con pacing; tolerancia a desorden | `OBSERVADO` |
| **Resistencia Adversarial** | `PASS` | Replay (100k), Fuzzing, Handshake Flood rechazados | `DEMOSTRADO` |
| **Aislamiento Contención (ADV-01)**| `PASS` | PDR 100% restaurado con sockets dedicados | `DEMOSTRADO` |
| **Persistencia de Identidad DID** | `PASS` | Conservación de llaves Ed25519 tras reinicio | `DEMOSTRADO` |
| **Interoperabilidad Windows/Linux**| `PASS` | Flujos cruzados Windows 11 <-> Ubuntu WSL2 | `DEMOSTRADO` |
| **Estabilidad de Memoria (Leaks)** | `PENDING` | Soak prolongado con `pprof` | `PENDIENTE` |
| **Compatibilidad Upgrade/Rollback** | `PENDING` | Convivencia v0.1 y v0.2 | `PENDIENTE` |

---

## 2. Regla Final de Dictamen

El estado global de "Release Candidate" sólo será concedido cuando todas las casillas contengan evidencia concluyente debidamente documentada y archivada en `docs/production/RESULTS/*.json` y la auditoría final `FINAL_RELEASE_AUDIT.md`.
