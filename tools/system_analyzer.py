#!/usr/bin/env python3
import json
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
INDEX_DB = ROOT / ".kuzu_index" / "ipv7.db"
KUZU_EXE = ROOT / "tools" / "kuzu" / "kuzu.exe"

def run_cypher(query):
    p = subprocess.run([str(KUZU_EXE), str(INDEX_DB), "-m", "json", "-s", "-b"],
                       input=query + "\n", capture_output=True, text=True, encoding="utf-8")
    if p.returncode == 0 and p.stdout.strip().startswith("["):
        out = p.stdout.strip()
        last_bracket = out.rfind("[")
        return json.loads(out[last_bracket:])
    return []

def analyze():
    print("=== ANÁLISIS DE SISTEMAS CON KÙZU GRAPH ===")
    
    # 1. Archivos con mayor LoC
    print("\n[1] Archivos más extensos (densidad de lógica):")
    top_files = run_cypher("MATCH (p:Package)-[:CONTAINS]->(f:File) RETURN f.name, p.name, f.loc ORDER BY f.loc DESC LIMIT 6;")
    for f in top_files:
        print(f"  - {f['f.name']} ({f['p.name']}): {f['f.loc']} LoC")

    # 2. Structs clave en el sistema
    print("\n[2] Structs centrales del sistema:")
    structs = run_cypher("MATCH (f:File)-[:DEFINES]->(s:Symbol) WHERE s.kind = 'struct' RETURN s.name, f.name ORDER BY s.name;")
    for s in structs:
        print(f"  - Struct: {s['s.name']:<25} en {s['f.name']}")

    # 3. Dependencias entre paquetes
    print("\n[3] Dependencias inter-paquetes:")
    deps = run_cypher("MATCH (p1:Package)-[:DEPENDS_ON]->(p2:Package) RETURN p1.name, p2.name;")
    for d in deps:
        print(f"  - {d['p1.name']} -> {d['p2.name']}")

if __name__ == "__main__":
    analyze()
