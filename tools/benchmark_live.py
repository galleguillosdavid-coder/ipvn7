#!/usr/bin/env python3
import time
import json
import urllib.request
import urllib.error
import statistics
import sys

WIN_NODE_API = "http://127.0.0.1:8080"
WSL_NODE_API = "http://127.0.0.1:8082"

def get_json(url):
    req = urllib.request.Request(url, headers={"User-Agent": "IPv7-Benchmark/1.0"})
    with urllib.request.urlopen(req, timeout=5) as resp:
        return json.loads(resp.read().decode("utf-8"))

def post_json(url, data):
    payload = json.dumps(data).encode("utf-8")
    req = urllib.request.Request(url, data=payload, headers={
        "Content-Type": "application/json",
        "User-Agent": "IPv7-Benchmark/1.0"
    })
    start = time.perf_counter()
    with urllib.request.urlopen(req, timeout=5) as resp:
        elapsed = time.perf_counter() - start
        body = json.loads(resp.read().decode("utf-8"))
        return body, elapsed

def main():
    print("==================================================================")
    print("      IPv7 LIVE P2P BENCHMARK & STRESS TEST (Windows <-> WSL2)    ")
    print("==================================================================")

    # 1. Health check & Identity verification
    print("\n[+] Verificando estado de nodos...")
    try:
        win_info = get_json(f"{WIN_NODE_API}/api/info")
        print(f" [OK] Nodo Windows (8080) ID: {win_info['identity'][:16]}... Peers: {win_info['peers_count']}")
    except Exception as e:
        print(f" [ERROR] Nodo Windows no disponible en {WIN_NODE_API}: {e}")
        sys.exit(1)

    try:
        wsl_info = get_json(f"{WSL_NODE_API}/api/info")
        print(f" [OK] Nodo WSL2 Linux (8082) ID: {wsl_info['identity'][:16]}... Peers: {wsl_info['peers_count']}")
    except Exception as e:
        print(f" [ERROR] Nodo WSL2 Linux no disponible en {WSL_NODE_API}: {e}")
        sys.exit(1)

    win_id = win_info["identity"]
    wsl_id = wsl_info["identity"]

    # 2. Benchmark de Ping / RTT P2P
    print("\n[+] Benchmark 1: Latencia RTT P2P Activa (Ping/Pong Ed25519)")
    ping_samples = 50
    ping_rtts = []
    ping_errors = 0

    for i in range(ping_samples):
        try:
            res, _ = post_json(f"{WIN_NODE_API}/api/ping", {"target_id": wsl_id})
            if res.get("success"):
                ping_rtts.append(res.get("rtt_ms", 1))
            else:
                ping_errors += 1
        except Exception:
            ping_errors += 1
        time.sleep(0.02)

    if ping_rtts:
        p50 = statistics.median(ping_rtts)
        p90 = sorted(ping_rtts)[int(len(ping_rtts) * 0.90)]
        p99 = sorted(ping_rtts)[int(len(ping_rtts) * 0.99)]
        mean_rtt = statistics.mean(ping_rtts)
        min_rtt = min(ping_rtts)
        max_rtt = max(ping_rtts)
        print(f"    - Muestras exitosas: {len(ping_rtts)}/{ping_samples} (Errores: {ping_errors})")
        print(f"    - RTT Min/Prom/Max : {min_rtt:.1f} ms / {mean_rtt:.1f} ms / {max_rtt:.1f} ms")
        print(f"    - Percentiles      : P50 = {p50:.1f} ms | P90 = {p90:.1f} ms | P99 = {p99:.1f} ms")
    else:
        print("    [!] No se obtuvieron mediciones de ping exitosas.")

    # 3. Benchmark de Rendimiento Mensajería E2EE (Throughput)
    print("\n[+] Benchmark 2: Inyección de Carga E2EE ChaCha20-Poly1305 (Windows -> WSL2)")
    msg_count = 200
    msg_times = []
    msg_errors = 0
    total_bytes = 0

    t_start = time.perf_counter()
    for i in range(msg_count):
        payload_text = f"BENCH_PACKET_SEQ_{i:05d}_DATA_{'X'*128}"
        total_bytes += len(payload_text)
        try:
            res, elapsed = post_json(f"{WIN_NODE_API}/api/send", {
                "recipient_id": wsl_id,
                "message": payload_text,
                "encrypted": True
            })
            msg_times.append(elapsed * 1000.0)
        except Exception as e:
            msg_errors += 1

    total_duration = time.perf_counter() - t_start
    throughput_msg_sec = len(msg_times) / total_duration if total_duration > 0 else 0
    throughput_kb_sec = (total_bytes / 1024.0) / total_duration if total_duration > 0 else 0

    print(f"    - Mensajes transmitidos: {len(msg_times)}/{msg_count} (Errores: {msg_errors})")
    print(f"    - Duración total        : {total_duration:.2f} s")
    print(f"    - Throughput Mensajes   : {throughput_msg_sec:.1f} msg/segundo")
    print(f"    - Ancho de Banda E2EE   : {throughput_kb_sec:.1f} KB/segundo")
    if msg_times:
        print(f"    - Latencia Inyección   : Min = {min(msg_times):.2f} ms | Media = {statistics.mean(msg_times):.2f} ms | Max = {max(msg_times):.2f} ms")

    # 4. Benchmark de Streaming Cascada CBOR + SHA-256
    print("\n[+] Benchmark 3: Telemetría y Streaming en Cascada (CBOR + SHA-256)")
    stream_count = 100
    stream_times = []
    stream_errors = 0

    t_stream_start = time.perf_counter()
    for i in range(stream_count):
        try:
            res, elapsed = post_json(f"{WIN_NODE_API}/api/broadcast-stream", {
                "stream_id": "bench-stream-01",
                "payload": f"TELEMETRY_FRAME_{i:04d}_SAMPLES"
            })
            stream_times.append(elapsed * 1000.0)
        except Exception:
            stream_errors += 1

    stream_duration = time.perf_counter() - t_stream_start
    stream_fps = len(stream_times) / stream_duration if stream_duration > 0 else 0
    print(f"    - Frames transmitidos   : {len(stream_times)}/{stream_count} (Errores: {stream_errors})")
    print(f"    - Frames por segundo    : {stream_fps:.1f} frames/segundo")

    print("\n==================================================================")
    print("                 BENCHMARK COMPLETADO EXITOSAMENTE                ")
    print("==================================================================")

if __name__ == "__main__":
    main()
