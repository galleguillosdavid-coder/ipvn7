#!/usr/bin/env python3
"""
Standalone Kùzu Graph Studio Web Server for IPv7
Permite explorar interactivamente en http://localhost:8090 las redes P2P y la arquitectura
del código almacenadas en la base de datos de grafos Kùzu (.kuzu_index/ipv7.db).
"""

import http.server
import socketserver
import json
import subprocess
import sys
from pathlib import Path

PROJECT_ROOT = Path(__file__).resolve().parent.parent.parent
INDEX_DB = PROJECT_ROOT / ".kuzu_index" / "ipv7.db"
TOOLS_KUZU = PROJECT_ROOT / "tools" / "kuzu"

PORT = 8090

def find_kuzu_cli():
    if sys.platform == "win32":
        exe = TOOLS_KUZU / "kuzu.exe"
    else:
        exe = TOOLS_KUZU / "kuzu"
    if exe.exists():
        return str(exe)
    return "kuzu"

def run_cypher(query: str):
    kuzu_bin = find_kuzu_cli()
    proc = subprocess.run(
        [kuzu_bin, str(INDEX_DB), "-m", "json", "-s", "-b"],
        input=query + "\n",
        capture_output=True,
        text=True,
        encoding="utf-8",
        errors="replace"
    )
    if proc.returncode != 0:
        raise RuntimeError(proc.stderr.strip() or "Error desconocido en Kùzu")
    
    out = proc.stdout.strip()
    if out.startswith("["):
        # Take last JSON array if multiple queries
        last_open = out.rfind("[")
        if last_open >= 0:
            out = out[last_open:]
        return json.loads(out)
    return []

class KuzuExplorerHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/" or self.path.startswith("/index.html"):
            # Serve the embedded index.html from ui/assets
            html_path = PROJECT_ROOT / "ui" / "assets" / "index.html"
            content = html_path.read_bytes()
            self.send_response(200)
            self.send_header("Content-Type", "text/html; charset=utf-8")
            self.send_header("Content-Length", str(len(content)))
            self.end_headers()
            self.wfile.write(content)
            return

        if self.path == "/api/kuzu/code-graph":
            try:
                rows = run_cypher("MATCH (p:Package)-[:CONTAINS]->(f:File) OPTIONAL MATCH (f)-[:DEFINES]->(s:Symbol) RETURN p.name, f.name, f.path, f.loc, s.name, s.kind LIMIT 150;")
                nodes_map = {}
                links = []
                pkg_colors = {"core": "#00f2fe", "adapters": "#8a2be2", "dht": "#10b981", "ui": "#f59e0b", "main": "#ec4899"}
                
                for r in rows:
                    p_name = r.get("p.name")
                    f_name = r.get("f.name")
                    f_path = r.get("f.path")
                    f_loc = r.get("f.loc", 0)
                    s_name = r.get("s.name")
                    s_kind = r.get("s.kind")

                    if not p_name or not f_name:
                        continue

                    pkg_id = f"pkg:{p_name}"
                    if pkg_id not in nodes_map:
                        nodes_map[pkg_id] = {
                            "id": pkg_id,
                            "label": f"pkg: {p_name}",
                            "group": "package",
                            "color": pkg_colors.get(p_name, "#6366f1"),
                            "radius": 24,
                            "metadata": {"package": p_name}
                        }

                    file_id = f"file:{f_path}"
                    if file_id not in nodes_map:
                        nodes_map[file_id] = {
                            "id": file_id,
                            "label": f_name,
                            "group": "file",
                            "color": "#38bdf8",
                            "radius": 14,
                            "metadata": {"package": p_name, "file": f_name, "path": f_path, "loc": f_loc}
                        }
                        links.append({"source": pkg_id, "target": file_id, "label": "contains", "color": "rgba(255,255,255,0.2)"})

                    if s_name:
                        s_id = f"sym:{f_path}::{s_name}"
                        if s_id not in nodes_map:
                            s_color = "#f43f5e" if s_kind == "struct" else ("#fbbf24" if s_kind == "interface" else "#a78bfa")
                            nodes_map[s_id] = {
                                "id": s_id,
                                "label": f"{s_name} ({s_kind})",
                                "group": "symbol",
                                "color": s_color,
                                "radius": 8,
                                "metadata": {"name": s_name, "kind": s_kind, "file": f_path}
                            }
                            links.append({"source": file_id, "target": s_id, "label": "defines", "color": "rgba(255,255,255,0.1)"})

                data = {"nodes": list(nodes_map.values()), "links": links}
                self.send_json(200, data)
            except Exception as e:
                self.send_json(500, {"error": str(e)})
            return

        if self.path == "/api/kuzu/network-graph":
            try:
                rows = run_cypher("MATCH (p:Peer) OPTIONAL MATCH (p)-[r:CONNECTED_TO]->(m:Peer) RETURN p.id, p.endpoint, p.is_local, r.adapter, r.latency_ms, r.encrypted, m.id;")
                nodes_map = {}
                links = []
                for r in rows:
                    p_id = r.get("p.id")
                    if p_id and p_id not in nodes_map:
                        is_loc = r.get("p.is_local", False)
                        nodes_map[p_id] = {
                            "id": p_id,
                            "label": ("Nodo Local (" + p_id[:10] + "...)") if is_loc else (p_id[:10] + "..."),
                            "group": "local_node" if is_loc else "peer",
                            "color": "#00f2fe" if is_loc else "#8a2be2",
                            "radius": 20 if is_loc else 14,
                            "metadata": {"id": p_id, "endpoint": r.get("p.endpoint"), "is_local": is_loc}
                        }
                    m_id = r.get("m.id")
                    if m_id:
                        adapter = r.get("r.adapter") or "QUIC/UDP"
                        lat = r.get("r.latency_ms") or 1
                        links.append({
                            "source": p_id,
                            "target": m_id,
                            "label": f"{adapter} ({lat} ms)",
                            "color": "rgba(0, 242, 254, 0.5)",
                            "metadata": {"adapter": adapter, "latency_ms": lat}
                        })
                self.send_json(200, {"nodes": list(nodes_map.values()), "links": links})
            except Exception as e:
                self.send_json(500, {"error": str(e)})
            return

        if self.path == "/api/info":
            self.send_json(200, {"identity": "KUZU_EXPLORER_STANDALONE", "peers_count": 0, "cascade_count": 0})
            return

        if self.path == "/api/peers" or self.path == "/api/mesh":
            self.send_json(200, [])
            return

        self.send_response(404)
        self.end_headers()

    def do_POST(self):
        if self.path == "/api/kuzu/query":
            content_length = int(self.headers.get("Content-Length", 0))
            body = self.rfile.read(content_length).decode("utf-8")
            try:
                payload = json.loads(body)
                query = payload.get("query", "")
                rows = run_cypher(query)
                self.send_json(200, rows)
            except Exception as e:
                self.send_json(500, {"error": str(e)})
            return

        self.send_response(404)
        self.end_headers()

    def send_json(self, status, obj):
        data = json.dumps(obj).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

def main():
    print("==================================================================")
    print("          KÙZU GRAPH STUDIO STANDALONE - PROTOCOLO IPv7           ")
    print("==================================================================")
    print(f"[+] Base de datos Kùzu: {INDEX_DB}")
    print(f"[+] Servidor activo en:  http://localhost:{PORT}")
    print("==================================================================")

    socketserver.TCPServer.allow_reuse_address = True
    with socketserver.TCPServer(("", PORT), KuzuExplorerHandler) as httpd:
        try:
            httpd.serve_forever()
        except KeyboardInterrupt:
            print("\n[OK] Servidor Kùzu Explorer detenido.")

if __name__ == "__main__":
    main()
