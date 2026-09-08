#!/usr/bin/env python3
"""
IPv7 Project Indexer & Kùzu Graph Generator
Analiza el repositorio IPv7, extrae paquetes, archivos, structs, interfaces y funciones Go,
genera la base de datos de grafo en Kùzu (.kuzu_index/) y compila el índice docs/PROJECT_INDEX.md.
"""

import os
import re
import sys
import subprocess
from pathlib import Path

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
INDEX_DIR = PROJECT_ROOT / ".kuzu_index" / "ipv7.db"
DOCS_DIR = PROJECT_ROOT / "docs"
TOOLS_KUZU = PROJECT_ROOT / "tools" / "kuzu"

def find_kuzu_cli():
    if sys.platform == "win32":
        exe = TOOLS_KUZU / "kuzu.exe"
    else:
        exe = TOOLS_KUZU / "kuzu"
    if exe.exists():
        return str(exe)
    return "kuzu"

def parse_go_file(file_path: Path):
    try:
        content = file_path.read_text(encoding="utf-8", errors="ignore")
    except Exception:
        return None

    lines = content.splitlines()
    loc = len(lines)

    # Extract package
    pkg_match = re.search(r"^\s*package\s+([a-zA-Z0-9_]+)", content, re.MULTILINE)
    package_name = pkg_match.group(1) if pkg_match else "unknown"

    # Extract imports
    imports = re.findall(r'\"([^\"]+)\"', re.search(r"import\s*\((.*?)\)", content, re.DOTALL).group(1) if "import (" in content else "")
    single_imports = re.findall(r'^\s*import\s+\"([^\"]+)\"', content, re.MULTILINE)
    all_imports = sorted(set(imports + single_imports))

    # Extract structs & interfaces
    types = re.findall(r"type\s+([A-Z][a-zA-Z0-9_]*)\s+(struct|interface)", content)

    # Extract functions & methods
    funcs = re.findall(r"func\s+(?:\([^\)]+\)\s+)?([A-Z][a-zA-Z0-9_]*)\s*\(", content)

    return {
        "path": str(file_path.relative_to(PROJECT_ROOT)).replace("\\", "/"),
        "name": file_path.name,
        "loc": loc,
        "package": package_name,
        "imports": all_imports,
        "types": types, # [(Name, kind)]
        "funcs": sorted(set(funcs))
    }

def main():
    print("==================================================")
    print("      IPv7 Project Indexer & Kùzu Graph Builder   ")
    print("==================================================")
    print(f"[+] Directorio raíz: {PROJECT_ROOT}")

    go_files = []
    packages = set()

    for root, dirs, files in os.walk(PROJECT_ROOT):
        # Ignore git, bin, tools/kuzu, .kuzu_index
        dirs[:] = [d for d in dirs if d not in [".git", "bin", ".kuzu_index", ".agents", "tools"]]
        for f in files:
            if f.endswith(".go"):
                full_p = Path(root) / f
                parsed = parse_go_file(full_p)
                if parsed:
                    go_files.append(parsed)
                    packages.add(parsed["package"])

    print(f"[+] Archivos Go encontrados: {len(go_files)}")
    print(f"[+] Paquetes Go detectados:  {len(packages)} ({', '.join(sorted(packages))})")

    # 1. Generar docs/PROJECT_INDEX.md
    DOCS_DIR.mkdir(exist_ok=True)
    index_md = DOCS_DIR / "PROJECT_INDEX.md"

    md_lines = [
        "# Índice Integral del Proyecto IPv7",
        "",
        "Este documento cataloga de forma exhaustiva todos los componentes, archivos, tipos, funciones y herramientas del protocolo IPv7. Generado automáticamente por el indexador semántico.",
        "",
        "## Resumen de Paquetes",
        "",
        "| Paquete | Archivos | Líneas de Código (LoC) | Propósito Principal |",
        "|---|---|---|---|"
    ]

    pkg_map = {}
    for item in go_files:
        p = item["package"]
        if p not in pkg_map:
            pkg_map[p] = []
        pkg_map[p].append(item)

    pkg_desc = {
        "core": "Núcleo de protocolo: Identidad Ed25519, Contenedores CBOR, Cifrado E2EE, Mundo Pequeño (12 Grados), Streaming en Cascada y Orquestador de Nodos.",
        "adapters": "Adaptadores de transporte desacoplados: UDP best-effort, QUIC sobre TLS 1.3, NAT Traversal con STUN, Relay seguro tipo DERP y WebRTC.",
        "dht": "Tabla Hash Distribuida Kademlia segura para resolución descentralizada Identity -> Endpoints con firmas digitales.",
        "ui": "Dashboard web SPA interactivo, servidor HTTP/WebSocket, Chat E2EE y monitor de topología en malla.",
        "main": "Puntos de entrada ejecutables: nodo P2P (`cmd/node`) y cliente de chat CLI (`cmd/chat`)."
    }

    for p in sorted(pkg_map.keys()):
        files_count = len(pkg_map[p])
        total_loc = sum(x["loc"] for x in pkg_map[p])
        desc = pkg_desc.get(p, "Módulo de soporte del sistema.")
        md_lines.append(f"| [`{p}`](#paquete-{p}) | {files_count} | {total_loc} | {desc} |")

    md_lines.append("\n---\n")

    for p in sorted(pkg_map.keys()):
        md_lines.append(f"## Paquete `{p}`\n")
        md_lines.append(f"{pkg_desc.get(p, '')}\n")

        for f in sorted(pkg_map[p], key=lambda x: x["name"]):
            rel_link = f"[{f['name']}](file:///{PROJECT_ROOT.as_posix()}/{f['path']})"
            md_lines.append(f"### 📄 {rel_link} ({f['loc']} LoC)")
            if f["types"]:
                md_lines.append("**Tipos y Estructuras:**")
                for t_name, t_kind in f["types"]:
                    md_lines.append(f"- `{t_kind} {t_name}`")
            if f["funcs"]:
                md_lines.append("**Funciones Clave:**")
                for fn in f["funcs"]:
                    md_lines.append(f"- `{fn}()`")
            md_lines.append("")

    # Scripts and Tools section
    md_lines.extend([
        "---",
        "## Herramientas y Scripts de Automatización",
        "",
        f"- [scripts/build_linux.ps1](file:///{PROJECT_ROOT.as_posix()}/scripts/build_linux.ps1): Compilación cruzada para Linux amd64.",
        f"- [scripts/run_wsl.ps1](file:///{PROJECT_ROOT.as_posix()}/scripts/run_wsl.ps1): Supervisor de ejecución del nodo en WSL2.",
        f"- [scripts/wsl_node.sh](file:///{PROJECT_ROOT.as_posix()}/scripts/wsl_node.sh): Runner nativo en bash para Ubuntu WSL2 con logging rotativo.",
        f"- [scripts/dual_node_test.ps1](file:///{PROJECT_ROOT.as_posix()}/scripts/dual_node_test.ps1): Orquestador de pruebas de malla cruzada Windows <-> WSL2.",
        f"- [tools/kuzu/setup_kuzu.ps1](file:///{PROJECT_ROOT.as_posix()}/tools/kuzu/setup_kuzu.ps1): Descarga y aprovisionamiento de Kùzu Graph DB CLI.",
        f"- [tools/indexer/index_project.py](file:///{PROJECT_ROOT.as_posix()}/tools/indexer/index_project.py): Indexador de código hacia Grafo Kùzu.",
        "",
        "## Documentación Técnica Relacionada",
        "",
        f"- [docs/WSL_SUPERVISION.md](file:///{PROJECT_ROOT.as_posix()}/docs/WSL_SUPERVISION.md): Guía operativa de supervisión en WSL2.",
        f"- [docs/KUZU_MESH_GRAPH.md](file:///{PROJECT_ROOT.as_posix()}/docs/KUZU_MESH_GRAPH.md): Modelado en Kùzu y consultas Cypher.",
        f"- [docs/MUNDO_PEQUENO_ROUTING.md](file:///{PROJECT_ROOT.as_posix()}/docs/MUNDO_PEQUENO_ROUTING.md): Enrutamiento logarítmico acotado a 12 grados.",
        f"- [genesis.md](file:///{PROJECT_ROOT.as_posix()}/genesis.md): Plan génesis de arquitectura del proyecto.",
        ""
    ])

    index_md.write_text("\n".join(md_lines), encoding="utf-8")
    print(f"[OK] Generado documento de índice: {index_md}")

    # 2. Generar scripts Cypher para Kùzu
    cypher_commands = []

    # Schema creation
    cypher_commands.extend([
        "CREATE NODE TABLE IF NOT EXISTS Package (name STRING, PRIMARY KEY (name));",
        "CREATE NODE TABLE IF NOT EXISTS File (path STRING, name STRING, loc INT64, PRIMARY KEY (path));",
        "CREATE NODE TABLE IF NOT EXISTS Symbol (id STRING, name STRING, kind STRING, PRIMARY KEY (id));",
        "CREATE REL TABLE IF NOT EXISTS CONTAINS (FROM Package TO File);",
        "CREATE REL TABLE IF NOT EXISTS DEFINES (FROM File TO Symbol);",
        "CREATE REL TABLE IF NOT EXISTS IMPORTS (FROM File TO Package);",
        "CREATE REL TABLE IF NOT EXISTS DEPENDS_ON (FROM Package TO Package);",
        "CREATE NODE TABLE IF NOT EXISTS Peer (id STRING, endpoint STRING, is_local BOOLEAN, PRIMARY KEY (id));",
        "CREATE REL TABLE IF NOT EXISTS CONNECTED_TO (FROM Peer TO Peer, adapter STRING, latency_ms INT64, encrypted BOOLEAN);"
    ])

    # Insert Packages
    for p in packages:
        cypher_commands.append(f"MERGE (p:Package {{name: '{p}'}});")

    # Insert Files and symbols
    for f in go_files:
        f_path = f["path"]
        f_name = f["name"]
        loc = f["loc"]
        cypher_commands.append(f"MERGE (fl:File {{path: '{f_path}', name: '{f_name}', loc: {loc}}});")
        cypher_commands.append(f"MATCH (p:Package {{name: '{f['package']}'}}), (fl:File {{path: '{f_path}'}}) MERGE (p)-[:CONTAINS]->(fl);")

        for t_name, t_kind in f["types"]:
            sym_id = f"{f_path}::{t_name}"
            cypher_commands.append(f"MERGE (s:Symbol {{id: '{sym_id}', name: '{t_name}', kind: '{t_kind}'}});")
            cypher_commands.append(f"MATCH (fl:File {{path: '{f_path}'}}), (s:Symbol {{id: '{sym_id}'}}) MERGE (fl)-[:DEFINES]->(s);")

        for fn in f["funcs"]:
            sym_id = f"{f_path}::{fn}"
            cypher_commands.append(f"MERGE (s:Symbol {{id: '{sym_id}', name: '{fn}', kind: 'func'}});")
            cypher_commands.append(f"MATCH (fl:File {{path: '{f_path}'}}), (s:Symbol {{id: '{sym_id}'}}) MERGE (fl)-[:DEFINES]->(s);")

        # Dependency relations
        for imp in f["imports"]:
            if imp.startswith("ipv7/"):
                imp_pkg = imp.split("/")[-1]
                if imp_pkg in packages:
                    cypher_commands.append(f"MATCH (fl:File {{path: '{f_path}'}}), (p:Package {{name: '{imp_pkg}'}}) MERGE (fl)-[:IMPORTS]->(p);")
                    if f["package"] != imp_pkg:
                        cypher_commands.append(f"MATCH (p1:Package {{name: '{f['package']}'}}), (p2:Package {{name: '{imp_pkg}'}}) MERGE (p1)-[:DEPENDS_ON]->(p2);")

    # Guardar script de Cypher
    cypher_file = PROJECT_ROOT / "tools" / "indexer" / "init_kuzu_graph.cypher"
    cypher_file.write_text("\n".join(cypher_commands) + "\n", encoding="utf-8")
    print(f"[OK] Generado script Cypher: {cypher_file} ({len(cypher_commands)} sentencias)")

    # 3. Inicializar / Actualizar base de datos Kùzu
    kuzu_bin = find_kuzu_cli()
    print(f"[+] Ejecutando Kùzu CLI: {kuzu_bin}")
    INDEX_DIR.parent.mkdir(parents=True, exist_ok=True)

    try:
        # Kùzu CLI reads commands line by line or from stdin
        process = subprocess.run(
            [kuzu_bin, str(INDEX_DIR)],
            input="\n".join(cypher_commands) + "\n:quit\n",
            capture_output=True,
            text=True,
            encoding="utf-8",
            errors="replace",
            timeout=30
        )
        if process.returncode == 0:
            print("[OK] Base de datos Kùzu (.kuzu_index/) creada e indexada exitosamente.")
        else:
            print(f"[!] Kùzu stderr: {process.stderr[:300]}")
    except Exception as e:
        print(f"[!] Nota al ejecutar Kùzu CLI: {e}")

    print("==================================================")
    print("      Indexación completada con éxito.            ")
    print("==================================================")

if __name__ == "__main__":
    main()
