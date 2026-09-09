#!/usr/bin/env python3
"""
Modularizador de Frontend para IPv7
Separa ui/assets/index.html (2,578 líneas) en:
- ui/assets/css/styles.css
- ui/assets/js/kuzu_studio.js
- ui/assets/js/mesh.js
- ui/assets/js/desktop.js
- ui/assets/js/tunnels.js
- ui/assets/js/app.js
- ui/assets/index.html (HTML semántico limpio)
"""

from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
ASSETS = ROOT / "ui" / "assets"
INDEX_HTML = ASSETS / "index.html"

CSS_DIR = ASSETS / "css"
JS_DIR = ASSETS / "js"

CSS_DIR.mkdir(parents=True, exist_ok=True)
JS_DIR.mkdir(parents=True, exist_ok=True)

content = INDEX_HTML.read_text(encoding="utf-8")

# 1. Extraer CSS
css_start = content.find("<style>") + len("<style>")
css_end = content.find("</style>")
css_code = content[css_start:css_end].strip()
(CSS_DIR / "styles.css").write_text(css_code + "\n", encoding="utf-8")
print(f"[OK] CSS extraído: ui/assets/css/styles.css ({len(css_code.splitlines())} líneas)")

# 2. Extraer JS y Segmentar
script_start = content.find("<script>") + len("<script>")
script_end = content.rfind("</script>")
script_all = content[script_start:script_end].strip()

# Delimitadores conocidos
kuzu_marker = "// ==========================================\n    // KÙZU GRAPH STUDIO"
if kuzu_marker not in script_all:
    # Buscar con fallback
    kuzu_marker = [line for line in script_all.splitlines() if "KÙZU GRAPH STUDIO" in line or "KZU GRAPH STUDIO" in line][0]
    kuzu_idx = script_all.find(kuzu_marker)
    # retroceder al anterior // ====
    kuzu_start = script_all.rfind("// ====", 0, kuzu_idx)
else:
    kuzu_start = script_all.find(kuzu_marker)

mesh_marker = [line for line in script_all.splitlines() if "INTERACTIVE MESH GRAPH" in line][0]
mesh_start = script_all.rfind("// ====", 0, script_all.find(mesh_marker))

desktop_marker = [line for line in script_all.splitlines() if "ESCRITORIO REMOTO WEB" in line][0]
desktop_start = script_all.rfind("// ====", 0, script_all.find(desktop_marker))

tunnel_marker = [line for line in script_all.splitlines() if "TÚNELES P2P" in line or "TNELES P2P" in line][0]
tunnel_start = script_all.rfind("// ====", 0, script_all.find(tunnel_marker))

# Segmentación:
# 1. Kùzu Studio JS:
kuzu_js = script_all[kuzu_start:mesh_start].strip()
(JS_DIR / "kuzu_studio.js").write_text(kuzu_js + "\n", encoding="utf-8")
print(f"[OK] JS extraído: ui/assets/js/kuzu_studio.js ({len(kuzu_js.splitlines())} líneas)")

# 2. Mesh & Radar JS:
# En la sección mesh, extraemos el visualizador de malla
mesh_full_sec = script_all[mesh_start:desktop_start].strip()
info_marker = "// Load initial info"
info_idx = mesh_full_sec.find(info_marker)
mesh_canvas_part = mesh_full_sec[:info_idx].strip()
chat_and_ws_part = mesh_full_sec[info_idx:].strip()

# En chat_and_ws_part, renderRadar está presente. Lo extraemos para mesh.js
radar_start = chat_and_ws_part.find("function renderRadar()")
radar_end = chat_and_ws_part.find("function escapeHtml(str)")
radar_code = chat_and_ws_part[radar_start:radar_end].strip()

mesh_js = mesh_canvas_part + "\n\n    // ==========================================\n    // RADAR MUNDO PEQUEÑO (12 GRADOS)\n    // ==========================================\n    " + radar_code
(JS_DIR / "mesh.js").write_text(mesh_js + "\n", encoding="utf-8")
print(f"[OK] JS extraído: ui/assets/js/mesh.js ({len(mesh_js.splitlines())} líneas)")

# 3. Desktop JS:
desktop_js = script_all[desktop_start:tunnel_start].strip()
(JS_DIR / "desktop.js").write_text(desktop_js + "\n", encoding="utf-8")
print(f"[OK] JS extraído: ui/assets/js/desktop.js ({len(desktop_js.splitlines())} líneas)")

# 4. Tunnels JS:
toast_marker = "function showToast"
toast_idx = script_all.find(toast_marker, tunnel_start)
tunnel_js = script_all[tunnel_start:toast_idx].strip()

# Agregar broadcastTestStream a tunnels.js si se desea, o mantener
broadcast_code = ""
if "async function broadcastTestStream()" in chat_and_ws_part:
    b_start = chat_and_ws_part.find("async function broadcastTestStream()")
    b_end = chat_and_ws_part.find("function renderRadar()")
    broadcast_code = chat_and_ws_part[b_start:b_end].strip()
    tunnel_js += "\n\n    // Cascade stream test\n    " + broadcast_code

(JS_DIR / "tunnels.js").write_text(tunnel_js + "\n", encoding="utf-8")
print(f"[OK] JS extraído: ui/assets/js/tunnels.js ({len(tunnel_js.splitlines())} líneas)")

# 5. App & Chat JS (Core state, tabs, websocket, info, chat, dragdrop, sse, toasts, onload)
app_top = script_all[:kuzu_start].strip()
chat_clean = chat_and_ws_part[:chat_and_ws_part.find("async function broadcastTestStream()")] + "\n\n" + chat_and_ws_part[radar_end:]
app_bottom = script_all[toast_idx:].strip()

app_js = app_top + "\n\n    // ==========================================\n    // CORE APP, PEERS & E2EE CHAT\n    // ==========================================\n" + chat_clean + "\n\n    // ==========================================\n    // TOASTS & SSE\n    // ==========================================\n" + app_bottom
(JS_DIR / "app.js").write_text(app_js + "\n", encoding="utf-8")
print(f"[OK] JS extraído: ui/assets/js/app.js ({len(app_js.splitlines())} líneas)")

# 3. Generar nuevo index.html modular
head_part = content[:content.find("<style>")].strip()
body_part = content[content.find("<body>"):content.find("</main>") + len("</main>")].strip()

new_html = f"""{head_part}
  <link rel="stylesheet" href="/css/styles.css">
</head>
{body_part}

  <!-- MODULOS JAVASCRIPT MODULARES (IPV7 WEB SUITE) -->
  <script src="/js/app.js"></script>
  <script src="/js/kuzu_studio.js"></script>
  <script src="/js/mesh.js"></script>
  <script src="/js/desktop.js"></script>
  <script src="/js/tunnels.js"></script>
</body>
</html>
"""

INDEX_HTML.write_text(new_html, encoding="utf-8")
print(f"[OK] Generado ui/assets/index.html modular ({len(new_html.splitlines())} líneas)")
