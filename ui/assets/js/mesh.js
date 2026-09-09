// ==========================================
    // INTERACTIVE MESH GRAPH (TOPOLOGÍA EN MALLA)
    // ==========================================
    let meshNodes = [];
    let meshLinks = [];
    let meshAnimFrame = null;
    let draggedNode = null;
    let inspectedNode = null;
    let meshCanvas, meshCtx;
    let dataParticles = [];

    function initMeshCanvas() {
      meshCanvas = document.getElementById('meshCanvas');
      if (!meshCanvas) return;
      meshCtx = meshCanvas.getContext('2d');

      const wrap = document.getElementById('meshCanvasWrap');
      const rect = wrap.getBoundingClientRect();
      const dpr = window.devicePixelRatio || 1;
      meshCanvas.width = rect.width * dpr;
      meshCanvas.height = rect.height * dpr;
      meshCtx.scale(dpr, dpr);

      // Event listeners for dragging & inspecting
      meshCanvas.onmousedown = (e) => {
        const rect = meshCanvas.getBoundingClientRect();
        const mx = e.clientX - rect.left;
        const my = e.clientY - rect.top;
        const clicked = findNodeAt(mx, my);
        if (clicked) {
          draggedNode = clicked;
          showNodeInspector(clicked);
        } else {
          document.getElementById('meshInfoCard').style.display = 'none';
        }
      };

      window.onmousemove = (e) => {
        if (!draggedNode || !meshCanvas) return;
        const rect = meshCanvas.getBoundingClientRect();
        draggedNode.x = e.clientX - rect.left;
        draggedNode.y = e.clientY - rect.top;
        draggedNode.vx = 0;
        draggedNode.vy = 0;
      };

      window.onmouseup = () => {
        draggedNode = null;
      };

      if (!meshAnimFrame) {
        meshLoop();
      }
    }

    function findNodeAt(x, y) {
      for (let n of meshNodes) {
        const dist = Math.hypot(n.x - x, n.y - y);
        if (dist <= n.radius + 6) return n;
      }
      return null;
    }

    async function fetchMeshTopology() {
      try {
        const res = await fetch('/api/mesh');
        const data = await res.json();
        const wrap = document.getElementById('meshCanvasWrap');
        const W = wrap ? wrap.clientWidth : 800;
        const H = wrap ? wrap.clientHeight : 500;

        const oldPositions = new Map();
        meshNodes.forEach(n => oldPositions.set(n.id, { x: n.x, y: n.y }));

        // Map nodes
        meshNodes = data.nodes.map((n, i) => {
          const old = oldPositions.get(n.id);
          const isLoc = n.is_local;
          const x = old ? old.x : (isLoc ? W / 2 : W / 2 + (Math.random() - 0.5) * 300);
          const y = old ? old.y : (isLoc ? H / 2 : H / 2 + (Math.random() - 0.5) * 200);

          return {
            ...n,
            x: x,
            y: y,
            vx: 0,
            vy: 0,
            radius: isLoc ? 18 : 13,
            color: isLoc ? '#00f2fe' : '#8a2be2'
          };
        });

        meshLinks = data.links;

        // Seed data particles along links
        if (dataParticles.length < meshLinks.length * 2) {
          dataParticles = [];
          meshLinks.forEach(l => {
            dataParticles.push({ link: l, progress: Math.random(), speed: 0.008 + Math.random() * 0.008 });
          });
        }
      } catch (e) {
        console.error('Error loading mesh topology:', e);
      }
    }

    function resetMeshSimulation() {
      const wrap = document.getElementById('meshCanvasWrap');
      const W = wrap ? wrap.clientWidth : 800;
      const H = wrap ? wrap.clientHeight : 500;
      meshNodes.forEach((n, idx) => {
        if (n.is_local) {
          n.x = W / 2; n.y = H / 2;
        } else {
          const angle = (idx * (2 * Math.PI / Math.max(1, meshNodes.length - 1)));
          n.x = W / 2 + Math.cos(angle) * 160;
          n.y = H / 2 + Math.sin(angle) * 140;
        }
        n.vx = 0; n.vy = 0;
      });
    }

    function showNodeInspector(node) {
      inspectedNode = node;
      const card = document.getElementById('meshInfoCard');
      const title = document.getElementById('meshCardTitle');
      const content = document.getElementById('meshCardContent');

      title.innerText = node.is_local ? '📍 NODO LOCAL (Este Host)' : '🌐 PEER REMOTO';
      content.innerHTML = `
        <div style="margin-bottom:0.25rem;"><b>ID:</b> <span style="font-family:'JetBrains Mono',monospace; font-size:0.75rem;">${node.id}</span></div>
        <div style="margin-bottom:0.25rem;"><b>Endpoints:</b> ${node.endpoints.join(', ') || 'Descubierto en Overlay'}</div>
        <div style="margin-bottom:0.25rem;"><b>Grado Kleinberg:</b> ${node.degree}</div>
        <div><b>Seguridad:</b> ${node.e2ee ? '<span style="color:#10b981">🔒 Cifrado E2EE ChaCha20-Poly1305</span>' : 'Firma CBOR'}</div>
      `;
      card.style.display = 'block';
    }

    function selectPeerFromMesh() {
      if (!inspectedNode) return;
      if (inspectedNode.is_local) {
        if (typeof showToast === 'function') showToast('Nodo Local', 'Este es tu propio nodo.', 'info');
        else alert('Este es tu propio nodo.');
        return;
      }
      selectPeer(inspectedNode.id, inspectedNode.endpoints[0] || '');
      document.querySelector('.tab-btn:first-child').click();
    }

    function copyMeshNodeID() {
      if (!inspectedNode) return;
      navigator.clipboard.writeText(inspectedNode.id);
      if (typeof showToast === 'function') showToast('DID Copiado', inspectedNode.id.substring(0, 20) + '...', 'success');
      else alert('Clave pública copiada:\n' + inspectedNode.id);
    }

    async function exportMeshToKuzu() {
      try {
        const res = await fetch('/api/mesh/export-cypher');
        const cypher = await res.text();
        await navigator.clipboard.writeText(cypher);
        if (typeof showToast === 'function') {
          showToast('Cypher Copiado', 'Consultas Cypher copiadas al portapapeles para Kùzu CLI.', 'success');
        } else {
          alert('¡Consultas Cypher copiadas al portapapeles!\n\nPuedes ejecutarlas en Kùzu CLI con:\n.\\tools\\kuzu\\kuzu.exe .kuzu_index/ipv7.db\n\n' + cypher);
        }
      } catch (e) {
        if (typeof showToast === 'function') showToast('Error', 'Error exportando a Kùzu: ' + e, 'error');
        else alert('Error exportando a Kùzu: ' + e);
      }
    }

    function meshLoop() {
      meshAnimFrame = requestAnimationFrame(meshLoop);
      if (!meshCtx || !meshCanvas) return;

      const wrap = document.getElementById('meshCanvasWrap');
      const W = wrap ? wrap.clientWidth : 800;
      const H = wrap ? wrap.clientHeight : 500;

      // 1. Force simulation step
      // Repulsion between nodes
      for (let i = 0; i < meshNodes.length; i++) {
        for (let j = i + 1; j < meshNodes.length; j++) {
          const a = meshNodes[i];
          const b = meshNodes[j];
          let dx = b.x - a.x;
          let dy = b.y - a.y;
          let dist = Math.hypot(dx, dy) || 1;
          if (dist < 260) {
            let force = (260 - dist) / dist * 0.45;
            if (a !== draggedNode) { a.vx -= dx * force * 0.05; a.vy -= dy * force * 0.05; }
            if (b !== draggedNode) { b.vx += dx * force * 0.05; b.vy += dy * force * 0.05; }
          }
        }
      }

      // Spring attraction along links
      meshLinks.forEach(l => {
        const a = meshNodes.find(n => n.id === l.source);
        const b = meshNodes.find(n => n.id === l.target);
        if (!a || !b) return;
        let dx = b.x - a.x;
        let dy = b.y - a.y;
        let dist = Math.hypot(dx, dy) || 1;
        let targetDist = 150;
        let force = (dist - targetDist) * 0.02;
        if (a !== draggedNode) { a.vx += (dx / dist) * force; a.vy += (dy / dist) * force; }
        if (b !== draggedNode) { b.vx -= (dx / dist) * force; b.vy -= (dy / dist) * force; }
      });

      // Centering force & damping
      meshNodes.forEach(n => {
        if (n !== draggedNode) {
          n.vx += (W / 2 - n.x) * 0.008;
          n.vy += (H / 2 - n.y) * 0.008;
          n.vx *= 0.88;
          n.vy *= 0.88;
          n.x += n.vx;
          n.y += n.vy;
        }
      });

      // 2. Clear canvas
      meshCtx.clearRect(0, 0, W, H);

      // Draw background grid lines subtle
      meshCtx.strokeStyle = 'rgba(255, 255, 255, 0.025)';
      meshCtx.lineWidth = 1;
      const step = 40;
      for (let x = 0; x < W; x += step) {
        meshCtx.beginPath(); meshCtx.moveTo(x, 0); meshCtx.lineTo(x, H); meshCtx.stroke();
      }
      for (let y = 0; y < H; y += step) {
        meshCtx.beginPath(); meshCtx.moveTo(0, y); meshCtx.lineTo(W, y); meshCtx.stroke();
      }

      // 2.5. Concentric Small-World distance rings around local node
      const localNode = meshNodes.find(n => n.is_local);
      if (localNode) {
        for (let r = 1; r <= 3; r++) {
          const radius = r * 115;
          const alpha = 0.04 - r * 0.008 + Math.sin(Date.now() * 0.002 + r) * 0.015;
          meshCtx.beginPath();
          meshCtx.arc(localNode.x, localNode.y, radius, 0, Math.PI * 2);
          meshCtx.strokeStyle = `rgba(0, 242, 254, ${Math.max(0.01, alpha)})`;
          meshCtx.lineWidth = 1;
          meshCtx.setLineDash([4, 6]);
          meshCtx.stroke();
          meshCtx.setLineDash([]);
        }
      }

      // 3. Draw links
      meshLinks.forEach(l => {
        const a = meshNodes.find(n => n.id === l.source);
        const b = meshNodes.find(n => n.id === l.target);
        if (!a || !b) return;

        // Line
        meshCtx.beginPath();
        meshCtx.moveTo(a.x, a.y);
        meshCtx.lineTo(b.x, b.y);
        const grad = meshCtx.createLinearGradient(a.x, a.y, b.x, b.y);
        grad.addColorStop(0, 'rgba(0, 242, 254, 0.4)');
        grad.addColorStop(1, 'rgba(138, 43, 226, 0.4)');
        meshCtx.strokeStyle = grad;
        meshCtx.lineWidth = 2;
        meshCtx.stroke();

        // Latency badge in midpoint
        const mx = (a.x + b.x) / 2;
        const my = (a.y + b.y) / 2;
        meshCtx.fillStyle = 'rgba(13, 18, 28, 0.85)';
        meshCtx.fillRect(mx - 22, my - 9, 44, 18);
        meshCtx.strokeStyle = 'rgba(0, 242, 254, 0.3)';
        meshCtx.strokeRect(mx - 22, my - 9, 44, 18);
        meshCtx.fillStyle = '#00f2fe';
        meshCtx.font = '10px Inter, sans-serif';
        meshCtx.textAlign = 'center';
        meshCtx.textBaseline = 'middle';
        meshCtx.fillText(l.latency_ms + ' ms', mx, my);
      });

      // 4. Draw data packet particles
      dataParticles.forEach(p => {
        const a = meshNodes.find(n => n.id === p.link.source);
        const b = meshNodes.find(n => n.id === p.link.target);
        if (!a || !b) return;

        p.progress += p.speed;
        if (p.progress > 1) p.progress = 0;

        const px = a.x + (b.x - a.x) * p.progress;
        const py = a.y + (b.y - a.y) * p.progress;

        meshCtx.beginPath();
        meshCtx.arc(px, py, 3, 0, Math.PI * 2);
        meshCtx.fillStyle = '#10b981';
        meshCtx.shadowColor = '#10b981';
        meshCtx.shadowBlur = 8;
        meshCtx.fill();
        meshCtx.shadowBlur = 0;
      });

      // 5. Draw nodes
      meshNodes.forEach(n => {
        // Outer glow
        const pulse = Math.sin(Date.now() * 0.004 + (n.is_local ? 0 : 2)) * 3;
        meshCtx.beginPath();
        meshCtx.arc(n.x, n.y, n.radius + 5 + pulse, 0, Math.PI * 2);
        meshCtx.fillStyle = n.is_local ? 'rgba(0, 242, 254, 0.15)' : 'rgba(138, 43, 226, 0.15)';
        meshCtx.fill();

        // Core circle
        meshCtx.beginPath();
        meshCtx.arc(n.x, n.y, n.radius, 0, Math.PI * 2);
        meshCtx.fillStyle = n.color;
        meshCtx.shadowColor = n.color;
        meshCtx.shadowBlur = 12;
        meshCtx.fill();
        meshCtx.shadowBlur = 0;

        // Border
        meshCtx.strokeStyle = '#fff';
        meshCtx.lineWidth = 1.5;
        meshCtx.stroke();

        // Label
        meshCtx.fillStyle = '#f3f4f6';
        meshCtx.font = 'bold 11px Inter, sans-serif';
        meshCtx.textAlign = 'center';
        meshCtx.textBaseline = 'top';
        meshCtx.fillText(n.short_id, n.x, n.y + n.radius + 6);

        meshCtx.fillStyle = 'rgba(255,255,255,0.5)';
        meshCtx.font = '9px Inter, sans-serif';
        meshCtx.fillText(n.label, n.x, n.y + n.radius + 20);
      });
    }

    // ==========================================
    // RADAR MUNDO PEQUEÑO (12 GRADOS)
    // ==========================================
    function renderRadar() {
      const svg = document.getElementById('radarSvg');
      svg.innerHTML = '';
      const cx = 250, cy = 250, maxR = 220;

      // Draw 12 rings
      for (let i = 1; i <= 12; i++) {
        const r = (maxR / 12) * i;
        const circle = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
        circle.setAttribute('cx', cx);
        circle.setAttribute('cy', cy);
        circle.setAttribute('r', r);
        circle.setAttribute('class', 'radar-ring' + (i % 3 === 0 ? ' accent' : ''));
        svg.appendChild(circle);
      }

      // Center Node
      const center = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
      center.setAttribute('cx', cx);
      center.setAttribute('cy', cy);
      center.setAttribute('r', 8);
      center.setAttribute('fill', '#10b981');
      svg.appendChild(center);

      // Fetch peers to plot on rings
      fetch('/api/peers').then(r => r.json()).then(peers => {
        if (!peers) return;
        peers.forEach((p, idx) => {
          const id = p.id || p.ID || '';
          const degVal = p.degree ?? p.Degree ?? 1;
          const latVal = p.latency_ms ?? p.LatencyMs ?? 0;
          const deg = Math.max(1, Math.min(12, degVal));
          const r = (maxR / 12) * deg;
          const angle = (idx * (360 / Math.max(1, peers.length))) * (Math.PI / 180);
          const px = cx + r * Math.cos(angle);
          const py = cy + r * Math.sin(angle);

          const dot = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
          dot.setAttribute('cx', px);
          dot.setAttribute('cy', py);
          dot.setAttribute('r', 6);
          dot.setAttribute('fill', '#00f2fe');
          dot.setAttribute('cursor', 'pointer');
          dot.onclick = () => {
            if (typeof showToast === 'function') {
              showToast(`Anillo ${degVal}`, `Peer: ${id.substring(0, 14)}... | Latencia: ${latVal} ms`, 'info');
            } else {
              alert('Peer ID: ' + id + '\nGrado: ' + degVal + '\nLatencia: ' + latVal + ' ms');
            }
          };
          svg.appendChild(dot);
        });
      });
    }
