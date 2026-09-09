// ==========================================
    // TÚNELES P2P & VPN SOCKS5
    // ==========================================
    async function startRDPTunnel() {
      const peer = document.getElementById('rdpPeerID').value.trim() || selectedPeerID;
      const localPort = parseInt(document.getElementById('rdpLocalPort').value, 10) || 13389;
      const status = document.getElementById('rdpTunnelStatus');

      if (!peer) {
        if (typeof showToast === 'function') {
          showToast('Falta Peer', 'Selecciona o ingresa la ID del peer con Windows RDP.', 'warning');
        } else {
          alert('Selecciona o ingresa la ID del peer con Windows RDP.');
        }
        return;
      }

      status.innerText = 'Iniciando túnel P2P...';
      try {
        const res = await fetch('/api/tunnel/start', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ local_port: localPort, peer_id: peer, target_port: 3389 })
        });
        if (res.ok) {
          status.innerHTML = `<span style="color:#10b981;">✅ Activo: Conéctate con RDP a <b>localhost:${localPort}</b></span>`;
          refreshTunnelList();
        } else {
          const err = await res.text();
          status.innerText = 'Error: ' + err;
        }
      } catch (e) {
        status.innerText = 'Error de conexión: ' + e;
      }
    }

    async function startCustomTunnel() {
      const localPort = parseInt(document.getElementById('customLocalPort').value, 10);
      const targetPort = parseInt(document.getElementById('customTargetPort').value, 10);
      const peer = document.getElementById('customPeerID').value.trim() || selectedPeerID;
      const status = document.getElementById('customTunnelStatus');

      if (!peer || !localPort || !targetPort) {
        if (typeof showToast === 'function') {
          showToast('Campos requeridos', 'Por favor completa todos los campos (puerto local, remoto y peer).', 'warning');
        } else {
          alert('Por favor completa todos los campos (puerto local, remoto y peer).');
        }
        return;
      }

      status.innerText = 'Iniciando túnel...';
      try {
        const res = await fetch('/api/tunnel/start', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ local_port: localPort, peer_id: peer, target_port: targetPort })
        });
        if (res.ok) {
          status.innerHTML = `<span style="color:#10b981;">✅ Túnel activo: 127.0.0.1:${localPort} ➔ Remoto:${targetPort}</span>`;
          refreshTunnelList();
        } else {
          const err = await res.text();
          status.innerText = 'Error: ' + err;
        }
      } catch (e) {
        status.innerText = 'Error: ' + e;
      }
    }

    async function toggleSOCKS5VPN() {
      const port = parseInt(document.getElementById('vpnPort').value, 10) || 1080;
      const status = document.getElementById('vpnStatus');

      try {
        const res = await fetch('/api/vpn/start', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ port: port })
        });
        if (res.ok) {
          status.innerHTML = `<span style="color:#10b981;">🛡️ SOCKS5 Proxy activo en <b>127.0.0.1:${port}</b></span>`;
        } else {
          const err = await res.text();
          status.innerText = 'Error activando VPN: ' + err;
        }
      } catch (e) {
        status.innerText = 'Error: ' + e;
      }
    }

    async function refreshTunnelList() {
      try {
        const res = await fetch('/api/tunnel/list');
        if (res.ok) {
          const data = await res.json();
          if (selectedPeerID) {
            const rdpInput = document.getElementById('rdpPeerID');
            if (rdpInput && !rdpInput.value) rdpInput.value = selectedPeerID;
            const customInput = document.getElementById('customPeerID');
            if (customInput && !customInput.value) customInput.value = selectedPeerID;
          }
        }
      } catch (e) {
        console.error(e);
      }
    }

    // Cascade stream test
    async function broadcastTestStream() {
      try {
        await fetch('/api/broadcast-stream', { method: 'POST' });
        if (typeof showToast === 'function') {
          showToast('Streaming en Cascada', 'Frame de telemetría emitido (Fan-out 10)', 'info', 2000);
        }
      } catch (e) {
        console.error(e);
        if (typeof showToast === 'function') {
          showToast('Error', 'No se pudo emitir frame en cascada: ' + e, 'error');
        }
      }
    }
