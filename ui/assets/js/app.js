let myIdentity = '';
let selectedPeerID = '';

function escapeHtml(str) {
  if (!str) return '';
  return String(str).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}

function showToast(titleOrMsg, subtitleOrType = 'info', durationOrCustom = 3500) {
  let container = document.getElementById('toast-container');
  if (!container) {
    container = document.createElement('div');
    container.id = 'toast-container';
    document.body.appendChild(container);
  }

  let title = titleOrMsg;
  let subtitle = '';
  let type = 'info';
  let duration = 3500;

  if (typeof subtitleOrType === 'string') {
    if (['info', 'success', 'warning', 'error'].includes(subtitleOrType.toLowerCase())) {
      type = subtitleOrType.toLowerCase();
      if (typeof durationOrCustom === 'number') duration = durationOrCustom;
    } else {
      subtitle = subtitleOrType;
      if (typeof durationOrCustom === 'string') {
        type = (durationOrCustom.startsWith('#') || durationOrCustom.includes('10b981')) ? 'success' : 'info';
      } else if (typeof durationOrCustom === 'number') {
        duration = durationOrCustom;
      }
    }
  }

  const icons = {
    info: 'ℹ️',
    success: '✅',
    warning: '⚠️',
    error: '❌'
  };

  const toast = document.createElement('div');
  toast.className = `toast toast-${type}`;
  toast.innerHTML = `
    <span class="toast-icon">${icons[type] || '✨'}</span>
    <div class="toast-content">
      <div style="font-weight:600;">${escapeHtml(title)}</div>
      ${subtitle ? `<div style="font-size:0.75rem; color:#94a3b8; margin-top:2px;">${escapeHtml(subtitle)}</div>` : ''}
    </div>
  `;

  container.appendChild(toast);

  setTimeout(() => {
    toast.classList.add('toast-hiding');
    setTimeout(() => {
      if (toast.parentNode) toast.parentNode.removeChild(toast);
    }, 250);
  }, duration);
}
window.showToast = showToast;

// Global Keyboard Navigation (Alt+1..7 and Ctrl+K)
window.addEventListener('keydown', (e) => {
  if (e.altKey && e.key >= '1' && e.key <= '7') {
    e.preventDefault();
    const tabMap = {
      '1': 'chat',
      '2': 'desktop',
      '3': 'tunnel',
      '4': 'mesh',
      '5': 'kuzu',
      '6': 'radar',
      '7': 'cascade'
    };
    const tabName = tabMap[e.key];
    if (tabName) {
      switchTab(tabName);
      showToast(`Pestaña: ${tabName.toUpperCase()}`, 'info', 1200);
    }
  } else if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') {
    e.preventDefault();
    const manualInput = document.getElementById('manualPeerID');
    if (manualInput) {
      manualInput.focus();
      manualInput.select();
      showToast('Enfoque rápido en búsqueda DID', 'info', 1500);
    }
  }
});

// Switch tabs
function switchTab(name) {
  document.querySelectorAll('.tab-btn').forEach(b => {
    b.classList.remove('active');
    b.setAttribute('aria-selected', 'false');
  });
  document.querySelectorAll('.tab-pane').forEach(p => p.classList.remove('active'));
  const targetBtn = document.querySelector(`button[onclick*="'${name}'"]`);
  if (targetBtn) {
    targetBtn.classList.add('active');
    targetBtn.setAttribute('aria-selected', 'true');
  } else if (window.event && window.event.target && window.event.target.classList) {
    window.event.target.classList.add('active');
    window.event.target.setAttribute('aria-selected', 'true');
  }
  const pane = document.getElementById('tab-' + name);
  if (pane) pane.classList.add('active');

  if (name === 'radar') renderRadar();
  if (name === 'desktop') initDesktop();
  if (name === 'tunnel') refreshTunnelList();
  if (name === 'mesh') {
    initMeshCanvas();
    fetchMeshTopology();
  }
  if (name === 'kuzu') {
    initKuzuCanvas();
    if (!currentKuzuMode) currentKuzuMode = 'network';
    setKuzuMode(currentKuzuMode);
  }
}

// ==========================================
// CORE APP, PEERS & E2EE CHAT
// ==========================================
// Load initial info
async function loadInfo() {
  try {
    const res = await fetch('/api/info');
    const data = await res.json();
    myIdentity = data.identity;
    document.getElementById('myIdShort').innerText = myIdentity.substring(0, 12) + '...';
    document.getElementById('myIdBadge').onclick = () => {
      navigator.clipboard.writeText(myIdentity);
      showToast('DID copiado al portapapeles', myIdentity.substring(0, 20) + '...', 'success');
    };
    document.getElementById('cascadeChildrenVal').innerText = data.cascade_count + ' / 10';
    if (typeof updateDesktopLanBanner === 'function') {
      updateDesktopLanBanner(data);
    }
    loadPeers();
  } catch (e) {
    console.error('Error loading node info:', e);
  }
}

async function loadPeers() {
  try {
    const res = await fetch('/api/peers');
    const peers = await res.json();
    const peerList = document.getElementById('peerList');
    if (!peers || peers.length === 0) {
      peerList.innerHTML = `
        <div class="peer-empty-state">
          <div class="empty-icon">📡</div>
          <div class="empty-title">Sin pares remotos aún</div>
          <div class="empty-desc">Conéctate a otro nodo con <code>-peer IP:PUERTO</code> o introduce su clave DID arriba.</div>
        </div>
      `;
      return;
    }
        peerList.innerHTML = '';
        peers.forEach(p => {
          const id = p.id || p.ID || '';
          const endpoints = p.endpoints || p.Endpoints || [];
          const latency = p.latency_ms ?? p.LatencyMs ?? 0;
          const degree = p.degree ?? p.Degree ?? 0;

          const div = document.createElement('div');
          div.className = 'peer-item' + (id === selectedPeerID ? ' selected' : '');
          div.onclick = () => selectPeer(id, endpoints[0] || '');
          div.innerHTML = `
            <div class="peer-item-name">
              <span>${id.substring(0, 10)}...</span>
              <span style="color:var(--accent-cyan); font-size:0.75rem;">${latency} ms</span>
            </div>
            <div class="peer-item-sub">Grado: ${degree} | ${endpoints.join(', ') || 'Overlay'}</div>
          `;
          peerList.appendChild(div);
        });
      } catch (e) {
        console.error('Error loading peers:', e);
      }
    }

    async function connectByDID() {
      const input = document.getElementById('manualPeerID');
      const status = document.getElementById('didConnectStatus');
      let did = input.value.trim();
      if (!did) {
        showToast('Atención', 'Por favor ingresa el DID o Clave Pública del destinatario.', 'warning');
        return;
      }
      did = did.replace(/^did:ipv7:/, '');
      if (status) {
        status.style.display = 'block';
        status.style.color = 'var(--accent-cyan)';
        status.innerText = '🔍 Resolviendo rutas WAN por DID...';
      }
      try {
        const res = await fetch(`/api/peers/resolve?did=${encodeURIComponent(did)}`);
        if (res.ok) {
          const data = await res.json();
          if (status) {
            status.style.color = '#10b981';
            status.innerText = `✅ Ruta viva: ${data.endpoint} (RTT: ${data.rtt_ms}ms)`;
          }
          selectPeer(did, data.endpoint);
          showToast('Peer resuelto', `${data.endpoint} (RTT: ${data.rtt_ms}ms)`, 'success');
        } else {
          if (status) {
            status.style.color = '#eab308';
            status.innerText = '⚠️ Aún no visible en WAN. Enrutando por Small-World...';
          }
          showToast('Enrutando por Small-World', 'Peer no visible en WAN directa aún', 'info');
        }
      } catch (e) {
        showToast('Error', 'No se pudo contactar al resolutor DID: ' + e, 'error');
        selectPeer(did, '');
      }
    }

    function selectPeer(id, ep) {
      id = id.replace(/^did:ipv7:/, '').trim();
      selectedPeerID = id;
      document.getElementById('chatTargetTitle').innerText = 'Destino DID: ' + id.substring(0, 16) + '...';
      document.getElementById('chatTargetSub').innerText = 'did:ipv7:' + id;
      document.getElementById('manualPeerID').value = id;
      if (ep) document.getElementById('manualPeerEp').value = ep;
      loadPeers();
    }

    async function sendMessage() {
      const input = document.getElementById('msgInput');
      const text = input.value.trim();
      if (!text) return;

      let recipient = document.getElementById('manualPeerID').value.trim() || selectedPeerID;
      recipient = recipient.replace(/^did:ipv7:/, '').trim();
      const endpoint = document.getElementById('manualPeerEp').value.trim();

      if (!recipient) {
        showToast('Falta destinatario', 'Selecciona un peer de la lista o introduce un DID arriba.', 'warning');
        return;
      }

      // Add to local feed immediately
      appendMessage('Tú', text, true, true);
      input.value = '';

      try {
        const res = await fetch('/api/send', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            recipient_id: recipient,
            endpoint: endpoint,
            message: text,
            encrypted: true
          })
        });
        if (!res.ok) {
          const err = await res.text();
          showToast('Error de transmisión', err, 'error');
        } else {
          showToast('Transmitido', 'Paquete cifrado E2EE enviado', 'success', 2000);
        }
      } catch (e) {
        showToast('Error de conexión', 'No se pudo comunicar con el nodo local: ' + e, 'error');
      }
    }

    function appendMessage(sender, text, isOutgoing, isE2EE) {
      const feed = document.getElementById('messagesFeed');
      const bubble = document.createElement('div');
      bubble.className = 'msg-bubble ' + (isOutgoing ? 'msg-outgoing' : 'msg-incoming');

      let contentHtml = '';
      if (text.startsWith('[ARCHIVO:') && text.includes('data:')) {
        const parts = text.split('\n');
        const header = parts[0];
        const dataUrl = parts.slice(1).join('\n').trim();
        const fileName = (header.match(/\[ARCHIVO:\s*(.*?)\s*\(/) || [])[1] || 'archivo_ipv7';
        if (dataUrl.startsWith('data:image/')) {
          contentHtml = `<div style="font-weight:600;margin-bottom:0.4rem;font-size:0.85rem;">📁 ${escapeHtml(header)}</div>
                         <img src="${dataUrl}" style="max-width:280px; max-height:220px; border-radius:8px; display:block; margin-bottom:0.6rem; border:1px solid rgba(255,255,255,0.1);"/>
                         <a href="${dataUrl}" download="${escapeHtml(fileName)}" class="mesh-btn" style="display:inline-block;padding:0.35rem 0.75rem;font-size:0.75rem;text-decoration:none;">⬇️ Descargar Imagen</a>`;
        } else {
          contentHtml = `<div style="font-weight:600;margin-bottom:0.4rem;font-size:0.85rem;">📁 ${escapeHtml(header)}</div>
                         <a href="${dataUrl}" download="${escapeHtml(fileName)}" class="mesh-btn" style="display:inline-block;padding:0.4rem 0.8rem;font-size:0.75rem;text-decoration:none;background:rgba(0,242,254,0.1);">⬇️ Descargar Archivo</a>`;
        }
      } else {
        contentHtml = `<div>${escapeHtml(text)}</div>`;
      }

      bubble.innerHTML = `
        ${contentHtml}
        <div class="msg-meta">
          <span>${sender}</span>
          ${isE2EE ? '<span style="color:var(--accent-cyan)">🔒 E2EE</span>' : ''}
          <span>${new Date().toLocaleTimeString()}</span>
        </div>
      `;
      feed.appendChild(bubble);
      feed.scrollTop = feed.scrollHeight;
    }

    function handleFileSelected(event) {
      const file = event.target.files[0];
      if (file) {
        uploadAndSendFile(file);
      }
    }

    function uploadAndSendFile(file) {
      if (file.size > 2 * 1024 * 1024) {
        showToast('Archivo grande', 'Para esta demo web el límite por archivo es de 2MB.', 'warning');
        return;
      }
      const reader = new FileReader();
      reader.onload = async function(e) {
        const dataUrl = e.target.result;
        const msg = `[ARCHIVO: ${file.name} (${(file.size/1024).toFixed(1)} KB)]\n${dataUrl}`;
        document.getElementById('msgInput').value = msg;
        await sendMessage();
      };
      reader.readAsDataURL(file);
    }

    function initChatDragDrop() {
      const dropZone = document.querySelector('.chat-window');
      if (!dropZone) return;
      ['dragenter', 'dragover'].forEach(name => {
        dropZone.addEventListener(name, (e) => {
          e.preventDefault();
          e.stopPropagation();
          dropZone.style.border = '2px dashed var(--accent-cyan)';
        });
      });
      ['dragleave', 'drop'].forEach(name => {
        dropZone.addEventListener(name, (e) => {
          e.preventDefault();
          e.stopPropagation();
          dropZone.style.border = '';
        });
      });
      dropZone.addEventListener('drop', (e) => {
        const files = e.dataTransfer.files;
        if (files && files.length > 0) {
          uploadAndSendFile(files[0]);
        }
      });
    }

    

function escapeHtml(str) {
      return str.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
    }

    // Connect WebSocket con Reconexión Automática & Backoff Exponencial
    let mainWS = null;
    let wsReconnectAttempts = 0;
    let wsReconnectTimer = null;

    function updateWSStatus(status, text) {
      const dot = document.getElementById('wsStatusDot');
      const label = document.getElementById('wsStatusText');
      if (!dot || !label) return;

      if (status === 'connected') {
        dot.style.background = '#10b981';
        dot.style.boxShadow = '0 0 10px rgba(16, 185, 129, 0.6)';
        label.innerText = text || 'Malla P2P Conectada';
      } else if (status === 'reconnecting') {
        dot.style.background = '#f59e0b';
        dot.style.boxShadow = '0 0 10px rgba(245, 158, 11, 0.6)';
        label.innerText = text || 'Reconectando...';
      } else {
        dot.style.background = '#ef4444';
        dot.style.boxShadow = '0 0 10px rgba(239, 68, 68, 0.6)';
        label.innerText = text || 'Desconectado';
      }
    }

    function connectWS() {
      if (wsReconnectTimer) {
        clearTimeout(wsReconnectTimer);
        wsReconnectTimer = null;
      }

      const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
      const url = proto + '//' + location.host + '/ws';
      
      try {
        mainWS = new WebSocket(url);
      } catch (err) {
        scheduleReconnect();
        return;
      }

      mainWS.onopen = () => {
        wsReconnectAttempts = 0;
        updateWSStatus('connected', 'Malla P2P Conectada');
        // Re-sincronizar datos automáticamente sin recargar la página (F5)
        loadInfo();
        loadPeers();
      };

      mainWS.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data);
          if (msg.type === 'incoming_message') {
            appendMessage(msg.from.substring(0, 10) + '...', msg.text, false, msg.encrypted);
          } else if (msg.type === 'cascade_chunk') {
            const feed = document.getElementById('cascadeFeed');
            if (feed) {
              const div = document.createElement('div');
              div.style.padding = '0.35rem 0';
              div.style.borderBottom = '1px solid rgba(255,255,255,0.05)';
              div.innerHTML = `<span style="color:var(--accent-cyan)">[Stream ${msg.stream_id}]</span> Frame #${msg.seq} (${msg.size} bytes) a las ${msg.time}`;
              feed.appendChild(div);
              feed.scrollTop = feed.scrollHeight;
            }
          }
        } catch (e) {
          console.error('Error parseando mensaje WS:', e);
        }
      };

      mainWS.onerror = () => {
        updateWSStatus('reconnecting', 'Error en socket P2P');
      };

      mainWS.onclose = () => {
        scheduleReconnect();
      };
    }

    function scheduleReconnect() {
      wsReconnectAttempts++;
      const baseDelay = 1000;
      const maxDelay = 10000;
      const delay = Math.min(baseDelay * Math.pow(1.8, wsReconnectAttempts - 1), maxDelay);
      const seconds = Math.round(delay / 1000);
      
      updateWSStatus('reconnecting', `Reconectando en ${seconds}s...`);

      if (wsReconnectTimer) clearTimeout(wsReconnectTimer);
      wsReconnectTimer = setTimeout(() => {
        connectWS();
      }, delay);
    }

    // ==========================================
    // TOASTS & SSE
    // ==========================================
function showToast(title, msg, color = '#00f2fe') {
      let container = document.getElementById('toastContainer');
      if (!container) {
        container = document.createElement('div');
        container.id = 'toastContainer';
        container.style.cssText = 'position:fixed; top:80px; right:20px; z-index:9999; display:flex; flex-direction:column; gap:10px; pointer-events:none;';
        document.body.appendChild(container);
      }
      const t = document.createElement('div');
      t.style.cssText = `background:rgba(13,18,28,0.92); backdrop-filter:blur(12px); border:1px solid ${color}; border-left:4px solid ${color}; padding:10px 16px; border-radius:8px; color:#f3f4f6; font-size:0.8rem; box-shadow:0 8px 24px rgba(0,0,0,0.5); transform:translateX(100%); transition:all 0.3s ease; max-width:320px; pointer-events:auto;`;
      t.innerHTML = `<b style="color:${color}; display:block; margin-bottom:2px;">${title}</b><span>${msg}</span>`;
      container.appendChild(t);
      setTimeout(() => { t.style.transform = 'translateX(0)'; }, 20);
      setTimeout(() => {
        t.style.transform = 'translateX(120%)';
        setTimeout(() => t.remove(), 350);
      }, 4000);
    }

    function initSSE() {
      if (!window.EventSource) return;
      try {
        const sse = new EventSource('/api/events');
        sse.onmessage = (e) => {
          try {
            const data = JSON.parse(e.data);
            if (data.type === 'self_healing_action') {
              showToast('🩺 Red Auto-Sanada', `${data.action}`, '#10b981');
              if (typeof loadMeshTopology === 'function') loadMeshTopology();
            } else if (data.type === 'incoming_message') {
              showToast('✉️ Mensaje E2EE', `${data.from.substring(0,8)}...: ${data.text}`, '#00f2fe');
            }
          } catch (err) {
            console.warn('SSE Parse error', err);
          }
        };
        sse.onerror = () => {
          // SSE auto-reconnects natively
        };
      } catch (err) {
        console.warn('SSE error:', err);
      }
    }

    window.onload = () => {
      loadInfo();
      connectWS();
      initSSE();
      setInterval(loadPeers, 2000);
      initChatDragDrop();

      const urlParams = new URLSearchParams(window.location.search);
      const tabParam = urlParams.get('tab') || window.location.hash.replace('#', '');
      if (tabParam) {
        setTimeout(() => {
          switchTab(tabParam);
          if (tabParam === 'kuzu') {
            const modeParam = urlParams.get('mode') || 'code';
            setKuzuMode(modeParam);
          }
        }, 150);
      }
    };
