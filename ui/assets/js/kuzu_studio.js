// ==========================================
    // KÙZU GRAPH STUDIO (REDES Y CÓDIGO)
    // ==========================================
    let currentKuzuMode = 'network';
    let kuzuNodes = [];
    let kuzuLinks = [];
    let kuzuAnimFrame = null;
    let kuzuDraggedNode = null;
    let kuzuHoverNode = null;
    let kuzuCanvas, kuzuCtx;
    let kuzuZoom = 1.0;
    let kuzuPanX = 0, kuzuPanY = 0;
    let isPanningKuzu = false;
    let kuzuStartPanX = 0, kuzuStartPanY = 0;
    let kuzuDataParticles = [];

    function initKuzuCanvas() {
      kuzuCanvas = document.getElementById('kuzuCanvas');
      if (!kuzuCanvas) return;
      kuzuCtx = kuzuCanvas.getContext('2d');

      const wrap = document.getElementById('kuzuCanvasWrap');
      const rect = wrap.getBoundingClientRect();
      const dpr = window.devicePixelRatio || 1;
      kuzuCanvas.width = rect.width * dpr;
      kuzuCanvas.height = rect.height * dpr;
      kuzuCtx.scale(dpr, dpr);

      // Pan & Drag events
      kuzuCanvas.onmousedown = (e) => {
        const rect = kuzuCanvas.getBoundingClientRect();
        const mx = (e.clientX - rect.left - kuzuPanX) / kuzuZoom;
        const my = (e.clientY - rect.top - kuzuPanY) / kuzuZoom;

        const clicked = findKuzuNodeAt(mx, my);
        if (clicked) {
          kuzuDraggedNode = clicked;
          showKuzuInspector(clicked);
        } else {
          isPanningKuzu = true;
          kuzuStartPanX = e.clientX - kuzuPanX;
          kuzuStartPanY = e.clientY - kuzuPanY;
        }
      };

      window.addEventListener('mousemove', (e) => {
        if (!kuzuCanvas) return;
        const rect = kuzuCanvas.getBoundingClientRect();
        if (kuzuDraggedNode) {
          kuzuDraggedNode.x = (e.clientX - rect.left - kuzuPanX) / kuzuZoom;
          kuzuDraggedNode.y = (e.clientY - rect.top - kuzuPanY) / kuzuZoom;
          kuzuDraggedNode.vx = 0;
          kuzuDraggedNode.vy = 0;
        } else if (isPanningKuzu) {
          kuzuPanX = e.clientX - kuzuStartPanX;
          kuzuPanY = e.clientY - kuzuStartPanY;
        } else {
          const mx = (e.clientX - rect.left - kuzuPanX) / kuzuZoom;
          const my = (e.clientY - rect.top - kuzuPanY) / kuzuZoom;
          kuzuHoverNode = findKuzuNodeAt(mx, my);
        }
      });

      window.addEventListener('mouseup', () => {
        kuzuDraggedNode = null;
        isPanningKuzu = false;
      });

      kuzuCanvas.onwheel = (e) => {
        e.preventDefault();
        const zoomFactor = e.deltaY < 0 ? 1.1 : 0.9;
        kuzuZoom = Math.max(0.3, Math.min(3.0, kuzuZoom * zoomFactor));
      };

      if (!kuzuAnimFrame) {
        kuzuLoop();
      }
    }

    function findKuzuNodeAt(x, y) {
      for (let n of kuzuNodes) {
        const dist = Math.hypot(n.x - x, n.y - y);
        if (dist <= n.radius + 6) return n;
      }
      return null;
    }

    async function setKuzuMode(mode) {
      currentKuzuMode = mode;
      document.getElementById('kuzuModeNetBtn').classList.toggle('active', mode === 'network');
      document.getElementById('kuzuModeCodeBtn').classList.toggle('active', mode === 'code');
      document.getElementById('kuzuInspector').style.display = 'none';

      const t0 = performance.now();
      if (mode === 'network') {
        document.getElementById('cypherQueryInput').value = 'MATCH (p:Peer) OPTIONAL MATCH (p)-[r:CONNECTED_TO]->(m:Peer) RETURN p.id, p.endpoint, p.is_local, r.adapter, r.latency_ms, m.id;';
        await loadKuzuNetworkGraph();
      } else {
        document.getElementById('cypherQueryInput').value = 'MATCH (p:Package)-[:CONTAINS]->(f:File) OPTIONAL MATCH (f)-[:DEFINES]->(s:Symbol) RETURN p.name, f.name, f.path, f.loc, s.name, s.kind LIMIT 150;';
        await loadKuzuCodeGraph();
      }
      const duration = (performance.now() - t0).toFixed(1);
      document.getElementById('kuzuQueryDuration').innerText = `Tiempo de consulta: ${duration} ms`;
    }

    async function loadKuzuNetworkGraph() {
      try {
        document.getElementById('kuzuStatsSummary').innerText = 'Consultando topología de red en Kùzu DB...';
        const res = await fetch('/api/kuzu/network-graph');
        const data = await res.json();
        setupKuzuGraphData(data.nodes || [], data.links || []);
        document.getElementById('kuzuStatsSummary').innerText = `Kùzu Red: ${(data.nodes || []).length} nodos, ${(data.links || []).length} enlaces P2P.`;

        // Network Legend
        document.getElementById('kuzuLegend').innerHTML = `
          <div class="legend-item"><div class="legend-dot" style="background:#00f2fe; box-shadow:0 0 8px #00f2fe;"></div><span>Nodo Local (Host)</span></div>
          <div class="legend-item"><div class="legend-dot" style="background:#8a2be2; box-shadow:0 0 8px #8a2be2;"></div><span>Peer Remoto</span></div>
          <div class="legend-item"><div class="legend-dot" style="background:#10b981; box-shadow:0 0 8px #10b981;"></div><span>Cifrado E2EE ChaCha20</span></div>
        `;
      } catch (e) {
        document.getElementById('kuzuStatsSummary').innerText = 'Error cargando red desde Kùzu: ' + e;
      }
    }

    async function loadKuzuCodeGraph() {
      try {
        document.getElementById('kuzuStatsSummary').innerText = 'Consultando arquitectura de código en Kùzu DB...';
        const res = await fetch('/api/kuzu/code-graph');
        const data = await res.json();
        setupKuzuGraphData(data.nodes || [], data.links || []);
        document.getElementById('kuzuStatsSummary').innerText = `Kùzu Código: ${(data.nodes || []).length} componentes, ${(data.links || []).length} dependencias.`;

        // Code Legend
        document.getElementById('kuzuLegend').innerHTML = `
          <div class="legend-item"><div class="legend-dot" style="background:#6366f1; box-shadow:0 0 8px #6366f1;"></div><span>Paquete (Go Package)</span></div>
          <div class="legend-item"><div class="legend-dot" style="background:#38bdf8; box-shadow:0 0 8px #38bdf8;"></div><span>Archivo (.go)</span></div>
          <div class="legend-item"><div class="legend-dot" style="background:#f43f5e; box-shadow:0 0 8px #f43f5e;"></div><span>Estructura (struct)</span></div>
          <div class="legend-item"><div class="legend-dot" style="background:#fbbf24; box-shadow:0 0 8px #fbbf24;"></div><span>Interfaz (interface)</span></div>
          <div class="legend-item"><div class="legend-dot" style="background:#a78bfa; box-shadow:0 0 8px #a78bfa;"></div><span>Función / Método</span></div>
        `;
      } catch (e) {
        document.getElementById('kuzuStatsSummary').innerText = 'Error cargando código desde Kùzu: ' + e;
      }
    }

    function setupKuzuGraphData(nodes, links) {
      const wrap = document.getElementById('kuzuCanvasWrap');
      const W = wrap ? wrap.clientWidth : 900;
      const H = wrap ? wrap.clientHeight : 560;

      const oldPos = new Map();
      kuzuNodes.forEach(n => oldPos.set(n.id, { x: n.x, y: n.y }));

      kuzuNodes = nodes.map((n, i) => {
        const old = oldPos.get(n.id);
        const angle = (i * 2 * Math.PI) / Math.max(1, nodes.length);
        const radiusDist = n.group === 'package' ? 120 : (n.group === 'file' ? 220 : 300);

        return {
          ...n,
          x: old ? old.x : (W / 2 + Math.cos(angle) * (radiusDist + (Math.random() - 0.5) * 50)),
          y: old ? old.y : (H / 2 + Math.sin(angle) * (radiusDist + (Math.random() - 0.5) * 50)),
          vx: 0,
          vy: 0,
          radius: n.radius || 12,
          color: n.color || '#00f2fe'
        };
      });

      kuzuLinks = links;

      kuzuDataParticles = [];
      kuzuLinks.slice(0, 40).forEach(l => {
        kuzuDataParticles.push({ link: l, progress: Math.random(), speed: 0.005 + Math.random() * 0.007 });
      });
    }

    function resetKuzuSimulation() {
      kuzuZoom = 1.0;
      kuzuPanX = 0;
      kuzuPanY = 0;
      const wrap = document.getElementById('kuzuCanvasWrap');
      const W = wrap ? wrap.clientWidth : 900;
      const H = wrap ? wrap.clientHeight : 560;

      kuzuNodes.forEach((n, i) => {
        const angle = (i * 2 * Math.PI) / Math.max(1, kuzuNodes.length);
        const dist = n.group === 'package' ? 100 : (n.group === 'file' ? 200 : 280);
        n.x = W / 2 + Math.cos(angle) * dist;
        n.y = H / 2 + Math.sin(angle) * dist;
        n.vx = 0;
        n.vy = 0;
      });
    }

    async function syncKuzuNetwork() {
      try {
        document.getElementById('kuzuStatsSummary').innerText = 'Sincronizando peers en vivo con Kùzu...';
        await loadKuzuNetworkGraph();
        alert('¡Malla sincronizada con éxito en Kùzu Graph DB!');
      } catch (e) {
        alert('Error sincronizando: ' + e);
      }
    }

    async function runCustomCypher() {
      const input = document.getElementById('cypherQueryInput');
      const query = input.value.trim();
      if (!query) return;

      const t0 = performance.now();
      try {
        document.getElementById('kuzuStatsSummary').innerText = 'Ejecutando consulta en Kùzu CLI...';
        const res = await fetch('/api/kuzu/query', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ query: query })
        });
        const duration = (performance.now() - t0).toFixed(1);
        document.getElementById('kuzuQueryDuration').innerText = `Tiempo de ejecución: ${duration} ms`;

        const data = await res.json();
        const rawEl = document.getElementById('kuzuRawResults');
        rawEl.innerText = JSON.stringify(data, null, 2);
        rawEl.style.display = 'block';

        if (Array.isArray(data)) {
          document.getElementById('kuzuStatsSummary').innerText = `Resultado: ${data.length} registros obtenidos en ${duration} ms.`;
        } else if (data.error) {
          document.getElementById('kuzuStatsSummary').innerText = `Error Cypher: ${data.error}`;
        }
      } catch (e) {
        document.getElementById('kuzuStatsSummary').innerText = 'Error conectando a Kùzu: ' + e;
      }
    }

    function applyCypherPreset(preset) {
      if (!preset) return;
      document.getElementById('cypherQueryInput').value = preset;
      runCustomCypher();
    }

    function toggleKuzuRawResults() {
      const el = document.getElementById('kuzuRawResults');
      el.style.display = el.style.display === 'block' ? 'none' : 'block';
    }

    function showKuzuInspector(node) {
      const insp = document.getElementById('kuzuInspector');
      const title = document.getElementById('kuzuInspectorTitle');
      const body = document.getElementById('kuzuInspectorBody');

      title.innerText = node.label;
      let html = `<div style="margin-bottom:0.4rem;"><b>Tipo / Grupo:</b> <span style="color:var(--accent-cyan); text-transform:uppercase;">${node.group}</span></div>`;
      html += `<div style="margin-bottom:0.4rem;"><b>ID Kùzu:</b> <span style="font-family:'JetBrains Mono',monospace; font-size:0.75rem;">${node.id}</span></div>`;

      if (node.metadata) {
        for (let k of Object.keys(node.metadata)) {
          html += `<div style="margin-bottom:0.25rem;"><b>${k}:</b> ${node.metadata[k]}</div>`;
        }
      }

      if (node.group === 'peer') {
        html += `<div style="margin-top:0.6rem;"><button class="mesh-btn" style="padding:0.25rem 0.6rem; font-size:0.75rem;" onclick="selectPeer('${node.id}', '${node.metadata && node.metadata.endpoint ? node.metadata.endpoint : ''}'); switchTab('chat');">💬 Chatear con este Peer</button></div>`;
      }

      body.innerHTML = html;
      insp.style.display = 'block';
    }

    function kuzuLoop() {
      kuzuAnimFrame = requestAnimationFrame(kuzuLoop);
      if (!kuzuCtx || !kuzuCanvas) return;

      const wrap = document.getElementById('kuzuCanvasWrap');
      const W = wrap ? wrap.clientWidth : 900;
      const H = wrap ? wrap.clientHeight : 560;

      // 1. Physics Step
      const repDist = currentKuzuMode === 'code' ? 140 : 220;
      for (let i = 0; i < kuzuNodes.length; i++) {
        for (let j = i + 1; j < kuzuNodes.length; j++) {
          const a = kuzuNodes[i];
          const b = kuzuNodes[j];
          let dx = b.x - a.x;
          let dy = b.y - a.y;
          let dist = Math.hypot(dx, dy) || 1;
          if (dist < repDist) {
            let force = (repDist - dist) / dist * 0.4;
            if (a !== kuzuDraggedNode) { a.vx -= dx * force * 0.05; a.vy -= dy * force * 0.05; }
            if (b !== kuzuDraggedNode) { b.vx += dx * force * 0.05; b.vy += dy * force * 0.05; }
          }
        }
      }

      const targetLinkDist = currentKuzuMode === 'code' ? 70 : 130;
      kuzuLinks.forEach(l => {
        const a = kuzuNodes.find(n => n.id === l.source);
        const b = kuzuNodes.find(n => n.id === l.target);
        if (!a || !b) return;
        let dx = b.x - a.x;
        let dy = b.y - a.y;
        let dist = Math.hypot(dx, dy) || 1;
        let force = (dist - targetLinkDist) * 0.015;
        if (a !== kuzuDraggedNode) { a.vx += (dx / dist) * force; a.vy += (dy / dist) * force; }
        if (b !== kuzuDraggedNode) { b.vx -= (dx / dist) * force; b.vy -= (dy / dist) * force; }
      });

      kuzuNodes.forEach(n => {
        if (n !== kuzuDraggedNode) {
          n.vx += (W / 2 - n.x) * 0.004;
          n.vy += (H / 2 - n.y) * 0.004;
          n.vx *= 0.86;
          n.vy *= 0.86;
          n.x += n.vx;
          n.y += n.vy;
        }
      });

      // 2. Clear & apply transform
      kuzuCtx.save();
      kuzuCtx.clearRect(0, 0, W, H);
      kuzuCtx.translate(kuzuPanX, kuzuPanY);
      kuzuCtx.scale(kuzuZoom, kuzuZoom);

      // Draw subtle grid
      kuzuCtx.strokeStyle = 'rgba(255, 255, 255, 0.02)';
      kuzuCtx.lineWidth = 1;
      const gStep = 50;
      for (let x = -W; x < W * 2; x += gStep) {
        kuzuCtx.beginPath(); kuzuCtx.moveTo(x, -H); kuzuCtx.lineTo(x, H * 2); kuzuCtx.stroke();
      }
      for (let y = -H; y < H * 2; y += gStep) {
        kuzuCtx.beginPath(); kuzuCtx.moveTo(-W, y); kuzuCtx.lineTo(W * 2, y); kuzuCtx.stroke();
      }

      // 3. Draw Links
      kuzuLinks.forEach(l => {
        const a = kuzuNodes.find(n => n.id === l.source);
        const b = kuzuNodes.find(n => n.id === l.target);
        if (!a || !b) return;

        kuzuCtx.beginPath();
        kuzuCtx.moveTo(a.x, a.y);
        kuzuCtx.lineTo(b.x, b.y);
        kuzuCtx.strokeStyle = l.color || 'rgba(255, 255, 255, 0.15)';
        kuzuCtx.lineWidth = l.label ? 1.8 : 1.2;
        kuzuCtx.stroke();

        if (l.label && currentKuzuMode === 'network') {
          const mx = (a.x + b.x) / 2;
          const my = (a.y + b.y) / 2;
          kuzuCtx.fillStyle = 'rgba(13, 18, 28, 0.85)';
          kuzuCtx.fillRect(mx - 24, my - 8, 48, 16);
          kuzuCtx.strokeStyle = 'rgba(0, 242, 254, 0.3)';
          kuzuCtx.strokeRect(mx - 24, my - 8, 48, 16);
          kuzuCtx.fillStyle = '#00f2fe';
          kuzuCtx.font = '9px Inter, sans-serif';
          kuzuCtx.textAlign = 'center';
          kuzuCtx.textBaseline = 'middle';
          kuzuCtx.fillText(l.label, mx, my);
        }
      });

      // 4. Data Particles
      kuzuDataParticles.forEach(p => {
        const a = kuzuNodes.find(n => n.id === p.link.source);
        const b = kuzuNodes.find(n => n.id === p.link.target);
        if (!a || !b) return;

        p.progress += p.speed;
        if (p.progress > 1) p.progress = 0;

        const px = a.x + (b.x - a.x) * p.progress;
        const py = a.y + (b.y - a.y) * p.progress;

        kuzuCtx.beginPath();
        kuzuCtx.arc(px, py, 2.5, 0, Math.PI * 2);
        kuzuCtx.fillStyle = '#10b981';
        kuzuCtx.shadowColor = '#10b981';
        kuzuCtx.shadowBlur = 6;
        kuzuCtx.fill();
        kuzuCtx.shadowBlur = 0;
      });

      // 5. Draw Nodes
      kuzuNodes.forEach(n => {
        const isHover = (kuzuHoverNode === n);
        const r = n.radius + (isHover ? 3 : 0);

        kuzuCtx.beginPath();
        kuzuCtx.arc(n.x, n.y, r + 4, 0, Math.PI * 2);
        kuzuCtx.fillStyle = n.color + '22';
        kuzuCtx.fill();

        kuzuCtx.beginPath();
        kuzuCtx.arc(n.x, n.y, r, 0, Math.PI * 2);
        kuzuCtx.fillStyle = n.color;
        kuzuCtx.shadowColor = n.color;
        kuzuCtx.shadowBlur = isHover ? 16 : 8;
        kuzuCtx.fill();
        kuzuCtx.shadowBlur = 0;

        kuzuCtx.strokeStyle = isHover ? '#fff' : 'rgba(255,255,255,0.7)';
        kuzuCtx.lineWidth = isHover ? 2 : 1;
        kuzuCtx.stroke();

        if (n.group === 'package' || n.group === 'local_node' || isHover || kuzuNodes.length < 50) {
          kuzuCtx.fillStyle = '#f3f4f6';
          kuzuCtx.font = (n.group === 'package' ? 'bold 11px' : '9px') + ' Inter, sans-serif';
          kuzuCtx.textAlign = 'center';
          kuzuCtx.textBaseline = 'top';
          kuzuCtx.fillText(n.label, n.x, n.y + r + 4);
        }
      });

      kuzuCtx.restore();
    }
