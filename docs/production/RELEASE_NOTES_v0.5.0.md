# IPv7 Release Notes - Versión v0.5.0 (Milestone: Multimedium Chaos & Reproducibility Audit)

> **ESTADO DEL PROTOCOLO**: **CONGELADO (Core Freeze Estricto)**.  
> **FECHA**: 2026-09-10  
> **TAG GIT**: `v0.5.0-chaos.audit`  
> **REPOSITORIO GITHUB**: [galleguillosdavid-coder/Ipv7](https://github.com/galleguillosdavid-coder/Ipv7)

---

## 1. Resumen Ejecutivo de la Versión

La versión **v0.5.0** marca la transición formal de IPv7 desde una arquitectura de prototipos en laboratorio hacia una **versión experimental congelada, blindada contra fallos catastróficos y 100% auditable por terceros independientes**:

1. **Protocol Core Freeze Inviolado**: El núcleo algorítmico ([`core/`](../../core)) no sufrió ninguna modificación (**0 líneas modificadas**). Todas las capacidades multimodales residen en adaptadores modulares ([`adapters/`](../../adapters) y [`dht/`](../../dht)).
2. **Independencia de Identidad y Transporte**: Demostrada experimentalmente la capacidad de un nodo de conmutar de Internet WAN a enlaces locales de radio ad-hoc ($H\text{-MULTI-01}$) en $\approx 100$ ms, conservando su DID Ed25519 y su sesión cifrada X25519 sin necesidad de renegociar handshakes.
3. **Desacoplamiento Epistémico en Dos Dimensiones**: Eliminada la ambigüedad conceptual entre certeza formal (`EPISTEMIC_STATUS`) y entorno físico (`ENVIRONMENT`).
4. **Auditoría de Reproducibilidad de Terceros**: Cualquier máquina, auditor humano o modelo de IA independiente puede clonar el repositorio y reproducir cada afirmación técnica mediante un comando determinista local o a través de **GitHub Actions**.
5. **Estabilidad Prolongada (Canary Soak > 9h)**: Más de 196.000 paquetes cursados con 99.43% PDR, memoria heap plana (0.36 MB) y 8 goroutines constantes sin fugas.

---

## 2. Los Cinco Horizontes Integrados y Verificados

| Horizonte | Módulo | Ensayos Principales | Métricas Clave | Estado Epistémico / Entorno |
| :--- | :--- | :--- | :--- | :---: |
| **H1: TUN/TAP Universal** | [`adapters/tun/`](../../adapters/tun) | `TestTunAdapterEndToEnd`, `TestTunAdapterSoak10kPackets` | Derivación ULA IPv6 (`fd07::`) a 12.36 ns/op, 0 allocs. 10k pkts al 100% PDR. | **`DEMONSTRATED`** / `LAB_SIMULATED` |
| **H2: Descentralización DHT** | [`dht/`](../../dht) | `TestPureKademliaTableKBuckets`, `TestAntiSybilProofOfWork` | 256 k-buckets ($k=20$), métrica XOR a 30.97 ns/op, PoW anti-Sybil. | **`DEMONSTRATED`** / `LAB_SIMULATED` |
| **H3: Enrutamiento Sphinx** | [`adapters/onion/`](../../adapters/onion) | `TestSphinxPacketBuildAndUnwrapThreeHops` | Paquetes de longitud fija invariante de 1280 bytes en 3 saltos concéntricos. | **`DEMONSTRATED`** / `LAB_SIMULATED` |
| **H4: Malla Off-Grid** | [`adapters/offgrid/`](../../adapters/offgrid) | `TestHybridSwitcher_FailoverAndRouting` | Balizas L2 `OG7!`, conmutación WAN $\leftrightarrow$ Malla a 1.045 ns/op. | **`DEMONSTRATED`** / `LAB_SIMULATED` |
| **H5: Validación de Caos** | [`tests/`](../../tests) | Batería destructiva completa ($H\text{-MULTI-01}$ a $H\text{-UNIFIED-E2E}$) | Conmutación en 100.35 ms, DoS L2 (888k balizas/s), split-brain healing. | **`DEMONSTRATED`** / `LAB_SIMULATED` |

---

## 3. Batería Destructiva y Certificación de Invariantes

Todas las aserciones han sido formalizadas bajo el contrato de aserciones ([`ASSERTION_CONTRACT.md`](../engineering/ASSERTION_CONTRACT.md)):

```text
┌──────────────────┬──────────────────────────────────────────┬──────────────────┬───────────┬──────────────┐
│ ID Hipótesis     │ Condición Hostil Provocada               │ Estado / Entorno │ Cobertura │ Resultado    │
├──────────────────┼──────────────────────────────────────────┼──────────────────┼───────────┼──────────────┤
│ H-MULTI-01       │ Corte simultáneo y abrupto de WAN        │ DEMONSTRATED     │ COMPLETE  │ 100% PDR     │
│                  │ Conmutación a radio L2 (cota <= 250ms)   │ LAB_SIMULATED    │           │ 100.76 ms    │
│ H-L2-DOS         │ Inundación de 10.000 balizas forjadas    │ DEMONSTRATED     │ COMPLETE  │ 100% PDR leg.│
│                  │ Cota tabla <= 256, heap <= 1024 KB       │ LAB_SIMULATED    │           │ 292 KB heap  │
│ H-FLAPPING       │ 30 ciclos rápidos de corte WAN (60 trans)│ DEMONSTRATED     │ COMPLETE  │ 0 deadlocks  │
│                  │ Cota de fuga goroutines <= 4 (delta = 0) │ LAB_SIMULATED    │           │ Convergente  │
│ H-SPLIT-BRAIN    │ Partición en 2 islas con DHT local       │ DEMONSTRATED     │ PARTIAL   │ Fusión XOR   │
│                  │ Puente local inter-islas (sin bkg DHT)   │ LAB_SIMULATED    │           │ Firmas OK    │
│ H-CONSTRAINED-MTU│ Enlace con MTU físico de 180 bytes       │ DEMONSTRATED     │ PARTIAL   │ 8 fragmentos │
│                  │ Desorden estocástico (sin ARQ L2)        │ LAB_SIMULATED    │           │ SHA256 100%  │
│ H-UNIFIED-E2E    │ TUN -> DHT -> Onion (1280B) -> L2 -> TUN │ DEMONSTRATED     │ COMPLETE  │ 100% PDR     │
│                  │ Pipeline completo de los 4 horizontes    │ LAB_SIMULATED    │           │ E2E íntegro  │
└──────────────────┴──────────────────────────────────────────┴──────────────────┴───────────┴──────────────┘
```

---

## 4. Auditoría de Terceros e Integración Continua (CI)

Cualquier auditor externo puede ejecutar la verificación en un solo paso:
- **En Windows**:
  ```powershell
  powershell -ExecutionPolicy Bypass -File .\scripts\audit_reproducibility.ps1
  ```
- **En Linux / WSL / macOS**:
  ```bash
  ./scripts/audit_reproducibility.sh
  ```
- **GitHub Actions (CI en la Nube)**:
  El workflow [`.github/workflows/reproducibility_audit.yml`](../../.github/workflows/reproducibility_audit.yml) valida automáticamente en runners de `ubuntu-latest` y `windows-latest` cada push y pull request.

---

## 5. Checksums Criptográficos SHA-256 de Binarios Compilados

| Binario | Plataforma / Arquitectura | Hash SHA-256 |
| :--- | :--- | :--- |
| `ipv7-node.exe` | Windows amd64 | `E6FCB7861C6B0A400CF3A1644EC25AED6304F17BDCC8128BF84F320C1A43B498` |
| `bin/ipv7-node-linux` | Linux amd64 | `1489EF070D0765813F992DFA0D4BF4051CB06C19E04D85EC7F1FF8A39DC2785D` |
| `bin/chat-linux` | Linux amd64 | `9D95846349D19228702502BD34D1D63C007EC467A7F9D3E21B3492BA5D0EE7D0` |

---

## 6. Declaración de Congelamiento Experimental

> [!NOTE]
> **NO SE AÑADIRÁ HORIZONTE 6**:  
> Conforme al acuerdo de gobernanza técnica, el proyecto sella este estado como una versión experimental madura. La prioridad operativa futura será la transición hacia hardware físico real (`REAL_HARDWARE`) cuando existan los transceptores de radio pertinentes, manteniendo inviolable la verdad empírica documentada.
