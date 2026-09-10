# ipvn7 NOS — Propuesta Técnica de Alta Ingeniería
## Network Operating System Descentralizado con Plano de Datos Programable

---

## 0. Resumen Ejecutivo y Principios de Diseño

**ipvn7 NOS** es un sistema operativo de red híbrido (kernel-space + user-space) orientado a topologías parcialmente conectadas, con identidad criptográfica soberana, plano de datos programable en eBPF/XDP y plano de extensión en WebAssembly. El diseño sigue cinco axiomas:

| Axioma | Implicación técnica |
|---|---|
| **Zero-Trust por defecto** | Ninguna sesión se establece sin verificación postura-identidad |
| **Graceful degradation** | Cada subsistema opera en partición, mesh parcial o WAN |
| **Idempotencia de estado** | Todo estado es reconstruible desde DAG + WoT |
| **Determinismo presupuestario** | CPU, ancho de banda y almacenamiento bajo cuotas explícitas |
| **Observabilidad sin fuga** | Métricas agregables sin exponer topología interna |

**Stack base objetivo:** Linux ≥ 6.6 (BTF, CO-RE, XDP multi-buffer), WSL2 kernel ≥ 6.6.36.1, Rust 1.79+, Zig para código sin asignador en fast-path, Wazero 1.7+ (runtime WASM sin cgo).

---

## 1. Firewall ZTNA con eBPF/XDP (Linux + WSL2)

### 1.1 Arquitectura

```
┌─────────────────────────────────────────────────────────────┐
│                     ipvn7-ztna (user-space)                 │
│  ┌──────────────┐  ┌──────────────┐  ┌───────────────────┐  │
│  │ Policy Engine│  │ SPIFFE ID    │  │ Posture Attestor  │  │
│  │ (Rego/CEL)   │←→│ Resolver     │←→│ (TPM/IMA/Boot)    │  │
│  └──────┬───────┘  └──────┬───────┘  └─────────┬─────────┘  │
│         │ mmap            │ bpf_map_update    │             │
└─────────┼─────────────────┼───────────────────┼─────────────┘
          ▼                 ▼                   ▼
┌─────────────────────────────────────────────────────────────┐
│              BPF MAP LAYER (pinned, BTF-typed)              │
│  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌──────────┐  │
│  │sess_table  │ │policy_lpm  │ │rate_limit  │ │flow_ctx  │  │
│  │LRU hash    │ │LPM_TRIE    │ │PERCPU_HASH │ │ringbuf   │  │
│  └────────────┘ └────────────┘ └────────────┘ └──────────┘  │
└─────────────────────────────────────────────────────────────┘
          ▲                 ▲                   ▲
          │                 │                   │
┌─────────┴─────────────────┴───────────────────┴─────────────┐
│  XDP HOOK (native | generic | offload)                      │
│    └→ tc/egress (clsact) → cgroup/skb → sockops             │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 Especificación técnica

**Clasificación de paquetes en XDP (pre-stack):**
- Parseo L2/L3/L4 con verificación de longitud estricta (evitar *variable-length header* overflow).
- Uso de `bpf_xdp_adjust_head` sólo tras validar checksums.
- `XDP_PASS` → stack normal solo si `sess_table` contiene `SPIFFE-ID` válido.
- `XDP_DROP` por defecto (default-deny), `XDP_TX` para reflect/ratelimit.

**Maps clave (BTF-native):**
```c
struct sess_key { __u32 saddr; __u32 daddr; __u16 sport; __u16 dport; __u8 proto; };
struct sess_val {
    __u64 spiffe_lo, spiffe_hi;   // identidad verificada
    __u32 ttl_ms;                 // caducidad perezosa
    __u16 posture_score;          // 0-1000
    __u8  state;                  // NEW|EST|DENY
    __u8  _pad;
};
```
- `sess_table`: `BPF_MAP_TYPE_LRU_HASH` (evicción bajo presión).
- `policy_lpm`: `BPF_MAP_TYPE_LPM_TRIE` con prefijos `/32`, `/48`, `/128`.
- `rate_limit`: `BPF_MAP_TYPE_PERCPU_HASH` + token bucket inline (ver §3).
- Flow events: `BPF_MAP_TYPE_RINGBUF` (evita pérdida por out-of-order con `bpf_ringbuf_reserve`).

**Tail calls (máx. 32 niveles):**
`xdp_parse → xdp_ztna_check → xdp_qos → xdp_tx`

**WSL2 — consideraciones críticas:**

| Aspecto | Restricción | Mitigación ipvn7 |
|---|---|---|
| Kernel | Sólo mirror/bridged si `.wslconfig` lo permite | `networkingMode=mirrored` + `ipvn7-ztna` en modo supervisor |
| XDP native | No siempre disponible en `hv_netvsc` | Fallback automático a `generic` (SKB-mode) con warning en CLI |
| BTF/CO-RE | Requiere kernel con BTF | Detección en arranque: `bpftool btf dump file /sys/kernel/btf/vmlinux` |
| `CAP_BPF` | Restringido en distros WSL | Deploy via `ipvn7-agent` con systemd en distro `systemd=true` |
| Offload NIC | Ausente | Deshabilitar `XDP_FLAGS_HW_MODE`, sólo `XDP_FLAGS_SKB_MODE` |

Se debe instrumentar `ipvn7-cli ztna diagnose` que detecte modo (native/generic/offload) y avise al operador.

### 1.3 Checklist ZTNA

- [ ] Módulo eBPF firmado (BLAKE3 + firma Ed25519) y cargado con `BPF_F_TEST_STATE_FREQ` en CI.
- [ ] Verificación de verifier: ejecutar con `BPF_F_STRICT_ALIGNMENT` en kernel 6.6+.
- [ ] CO-RE habilitado: `vmlinux.h` generado por `bpftool btf dump`.
- [ ] Mapas `pinned` en `/sys/fs/bpf/ipvn7/` con permisos 0640 root:ipvn7.
- [ ] Fallback `generic` probado en WSL2 con NIC Hyper-V.
- [ ] Fuzz de parser PCAP con `libFuzzer` ≥ 72h sin crashes.
- [ ] Tasa de paquetes objetivo: ≥ 8 Mpps/vCPU native, ≥ 1.2 Mpps/vCPU generic.
- [ ] Métricas exportadas: `xdp_packets_total`, `xdp_drops_total{reason}`, `ztna_posture_avg`.
- [ ] Rotación de mapas sin downtime vía `BPF_MAP_FREEZE` + doble buffer.
- [ ] Documentado modelo de amenaza (`docs/THREAT_MODEL_ZTNA.md`).

---

## 2. Sistema de Nombres dDNS con Petnames

### 2.1 Modelo de resolución (tríada de Zooko resuelta)

ipvn7 resuelve la tríada **Global + Seguro + Legible** mediante **estratificación en 3 capas**:

| Capa | Función | Estructura |
|---|---|---|
| **L0 — Crypto-ID** | Identidad canónica, inmutable | `ipvn7:<BLAKE3-256-pubkey>` (base32 lowercase) |
| **L1 — dDNS** | Nombres globales firmados en DHT | `alice.ipvn7` → registro firmado |
| **L2 — Petnames** | Nombres locales, no compartidos | `mi-nas`, `casa`, `impresora-2piso` |

**dDNS global** se resuelve sobre una **Kademlia DHT modificada** con:
- Registros firmados por la clave del dueño (`name → (pubkey, ttl, seq)`).
- **Squatting resistance**: precio PoW por registro (`work = 2^bits` con `bits = f(longitud_nombre)`).
- **Revocación** vía `seq` monotónico; último seq gana ante conflicto.
- Cachés TTL acotados a 5 min, negative caching 30 s.

**Petnames** son estrictamente locales, derivados de **BIP-39 mnemónico** (English-only) + sufijo de desambiguación:
```
petname = mnemonic_2_words(seed=local_secret || fingerprint) + "." + counter
```

### 2.2 Flujo de resolución

```
resolver("mi-nas") 
  → L2 petname store (local, SQLite WAL) 
  → hit: devuelve Crypto-ID
  → miss: consulta dDNS distribuida (L1)
       → miss: consulta DHT (L0 → registro firmado)
            → verifica firma con pubkey embebida
            → materializa petname sugerido: "mi-nas-3" (evita colisión)
```

### 2.3 Anti-spoofing y privacidad

- **DNSSEC-análogo**: firma obligatoria con Ed25519.
- **Privacy querying**: queries resueltas por 3 hops aleatorios (onion-lite) para no revelar interés.
- **No enumeration**: DHT no expone listado; sólo resolución por hash exacto.
- **Per-user salt** para petnames locales: nunca salen del nodo.

### 2.4 Checklist dDNS

- [ ] Especificación de wire format (`ipvn7-name@v1`, Protobuf/Canonical CBOR).
- [ ] Suites de test contra typosquatting (Levenshtein ≤ 2 bloqueado en registro).
- [ ] Bench: 10k resoluciones/s en cache L2, 800/s DHT round-trip (LAN).
- [ ] Petname generator determinista: misma semilla → mismo petname (test vectors).
- [ ] CLI: `ipvn7-cli name resolve <petname>`, `ipvn7-cli name sign <fqdn> --key file`.
- [ ] Rotación de claves sin perder nombre (registro `key-roll` firmado por clave vieja y nueva).
- [ ] Capa de compatibilidad: `ipvn7-cli name export --format hosts` para apps legacy.
- [ ] Auditoría: registro de cada resolución local en `~/.ipvn7/logs/ddns.jsonl` con rotación.

---

## 3. QoS, Token Bucket y PoW Dinámico Anti-DDoS

### 3.1 Modelo de QoS

Colas **HTB** (jerárquicas) + `fq_codel` como leaf:

```
qdisc root htb (10 Gbps)
├── class 1:10 realtime     (VOIP/gaming)       rate 200M ceil 1G   prio 0
├── class 1:20 interactive  (SSH/control)       rate 500M ceil 2G   prio 1
├── class 1:30 bulk         (transferencias)    rate 5G   ceil 8G   prio 2
└── class 1:40 scavenger    (P2P background)    rate 1G   ceil 3G   prio 3
```

Clasificación:
- DSCP (IPv4) / Traffic Class (IPv6) + `SO_MARK` + eBPF `tc` classifier.
- Reinyección a `fq_codel` con `quantum=1514`, `target=5ms`, `interval=100ms`.

**SLA por sesión** (declarativo, aplicado por ZTNA):
```yaml
session_sla:
  ingress: { rate: 100Mbps, burst: 2MB }
  egress:  { rate: 100Mbps, burst: 2MB }
  rtt_max: 150ms
  jitter_max: 30ms
  loss_max: 0.5%
```
Si el SLA se viola, `ipvn7-cli qos reclassify` degrada a la clase siguiente.

### 3.2 Token Bucket (doble cubo, precisión sub-ms)

Implementación en eBPF (fast path) y user-space (control):

```c
struct tb_state { __u64 tokens; __u64 last_refill_ns; __u64 rate; __u64 burst; };
static __always_inline bool tb_take(struct tb_state *s, __u64 now, __u64 cost) {
    __u64 delta = now - s->last_refill_ns;
    __u64 add   = (delta * s->rate) / 1_000_000_000ULL;
    s->tokens   = min_u64(s->tokens + add, s->burst);
    s->last_refill_ns = now;
    if (s->tokens >= cost) { s->tokens -= cost; return true; }
    return false;
}
```
- **Aritmética**: enteros sin división flotante (verificado por verifier).
- **Clock source**: `bpf_ktime_get_ns()` (monotónico, evita saltos NTP).
- **Sincronización**: `PERCPU` para evitar contention; agregación en `libbpf` stats.

### 3.3 PoW Dinámico Anti-DDoS

**Modelo de desafío-respuesta por conexión, con dificultad adaptativa:**

```
D(t) = clamp( base_bits + log2(load) - log2(reputation), [0, 28] )
```

- `load`: loadavg del nodo (1m) + backlog de accept queue.
- `reputation`: score WoT (§9), rango `[-100, 1000]`, mapeo a reducción de bits.
- Algoritmo: **Equihash-lite** (memoria 32 MiB, ASIC-resistant razonable) o **Argon2id-lite** para móviles.
- Verificación del resultado en eBPF (una sola verificación por sesión; el cómputo lo hace el cliente).

**Protocolo de handshake PoW:**
```
1. Cliente → CHALLENGE_REQ
2. Servidor → CHALLENGE { nonce, D, algorithm, expiry }
3. Cliente computa sol; envía PROOF { nonce, sol }
4. Servidor verifica en XDP (mapa pending_challenges)
5. Éxito → emite session token firmado (HMAC-BLAKE3)
```

**Adaptación de dificultad (control loop):**

```
error        = target_utilization - current_utilization
integral    += error * dt
derivative   = (error - prev_error) / dt
D_next       = D + Kp*error + Ki*integral + Kd*derivative
```
PID con anti-windup, `Kp=2.0, Ki=0.05, Kd=0.5`. Actualización cada 1s.

### 3.4 Checklist QoS / Token Bucket / PoW

- [ ] HTB validado con `netem` (delay/jitter/loss) contra SLAs.
- [ ] Benchmark: qdisc drop/backlog ≤ 0.01% en carga sostenida 1h.
- [ ] Token bucket verificado bajo `-O3` y sin FP (disassembly check).
- [ ] PoW: tiempos observ