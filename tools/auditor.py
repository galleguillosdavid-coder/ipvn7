#!/usr/bin/env python3
import os
import re
import subprocess
import json
from pathlib import Path

PROJECT_ROOT = Path(__file__).resolve().parent.parent
INDEX_DB = PROJECT_ROOT / ".kuzu_index" / "ipv7.db"
KUZU_EXE = PROJECT_ROOT / "tools" / "kuzu" / "kuzu.exe"

def run_cypher(query):
    p = subprocess.run([str(KUZU_EXE), str(INDEX_DB), "-m", "json", "-s", "-b"],
                       input=query + "\n", capture_output=True, text=True, encoding="utf-8")
    if p.returncode == 0 and p.stdout.strip().startswith("["):
        out = p.stdout.strip()
        last_bracket = out.rfind("[")
        return json.loads(out[last_bracket:])
    return []

def audit():
    print("==================================================================")
    print("       AUDITORÍA DE CÓDIGO CON KÙZU GRAPH DATABASE (IPv7)        ")
    print("==================================================================")

    # 1. Query Kùzu for packages and files
    files = run_cypher("MATCH (p:Package)-[:CONTAINS]->(f:File) RETURN p.name, f.name, f.path, f.loc ORDER BY p.name, f.loc DESC;")
    print(f"[+] Total de archivos Go en base de datos Kùzu: {len(files)}")

    findings = []

    # Patterns to look for
    patterns = [
        (r'dummyKey', 'Clave pública dummy (32 bytes de ceros) en lugar de handshake de identidad real'),
        (r'make\(\[\]byte,\s*32\)', 'Buffer estático de 32 bytes en blanco'),
        (r'time\.Sleep\(', 'Sleep síncrono arbitrario en código de producción'),
        (r'fmt\.Sprintf\(\"STREAM_FRAME_%d\"', 'Payload de streaming hardcodeado como string'),
        (r'Degree:\s*0', 'Grado Kleinberg forzado a 0 sin calcular distancia XOR real'),
        (r'LatencyMs:\s*1', 'Latencia hardcodeada a 1ms estática'),
        (r'127\.0\.0\.1:7001', 'Endpoint hardcodeado en producción'),
        (r'adapters\.DefaultSTUNServer', 'STUN por defecto'),
    ]

    for item in files:
        fpath = PROJECT_ROOT / item["f.path"]
        if not fpath.exists():
            continue
        content = fpath.read_text(encoding="utf-8", errors="ignore")
        lines = content.splitlines()

        # Skip test files from strict hardcoding audits, but check them too
        is_test = fpath.name.endswith("_test.go")

        for idx, line in enumerate(lines, 1):
            for pat, desc in patterns:
                if re.search(pat, line):
                    findings.append({
                        "file": item["f.path"],
                        "line": idx,
                        "content": line.strip(),
                        "issue": desc,
                        "is_test": is_test
                    })

    print(f"\n[!] Hallazgos detectados: {len(findings)}")
    prod_findings = [f for f in findings if not f["is_test"]]
    print(f"[!] Hallazgos en código de producción (no tests): {len(prod_findings)}\n")

    for f in prod_findings:
        print(f" -> [{f['file']}:{f['line']}] {f['issue']}")
        print(f"    Código: {f['content']}\n")

if __name__ == "__main__":
    audit()
