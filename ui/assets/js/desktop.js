// ==========================================
    // ESCRITORIO REMOTO WEB (HTML5 + WEBSOCKET)
    // ==========================================
    let desktopWS = null;
    let desktopCanvas = null;
    let desktopCtx = null;
    let isDesktopStreaming = false;
    let frameCount = 0;
    let lastFpsTime = Date.now();
    let allowRemoteControl = true;

    function initDesktop() {
      desktopCanvas = document.getElementById('remoteCanvas');
      if (desktopCanvas) {
        desktopCtx = desktopCanvas.getContext('2d');
        setupDesktopInputListeners();
      }
      updateDesktopLanBanner();
    }

    function toggleDesktopStream() {
      if (isDesktopStreaming) {
        disconnectDesktopWS();
      } else {
        connectDesktopWS();
      }
    }

    function connectDesktopWS() {
      const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
      desktopWS = new WebSocket(proto + '//' + location.host + '/ws/desktop');
      desktopWS.binaryType = 'blob';

      const btn = document.getElementById('btnConnectDesktop');
      const status = document.getElementById('desktopStatus');
      const placeholder = document.getElementById('desktopPlaceholder');

      desktopWS.onopen = () => {
        isDesktopStreaming = true;
        btn.innerText = 'Desconectar';
        btn.style.background = '#ef4444';
        status.innerText = 'Transmitiendo en vivo';
        status.style.color = '#10b981';
        if (placeholder) placeholder.style.display = 'none';
        desktopCanvas.style.display = 'block';
      };

      desktopWS.onmessage = (event) => {
        if (typeof event.data === 'string') {
          const meta = JSON.parse(event.data);
          if (meta.type === 'init') {
            desktopCanvas.width = meta.width;
            desktopCanvas.height = meta.height;
            document.getElementById('desktopRes').innerText = `${meta.width}x${meta.height}`;
          }
        } else if (event.data instanceof Blob) {
          const img = new Image();
          img.onload = () => {
            if (!desktopCanvas.width) {
              desktopCanvas.width = img.width;
              desktopCanvas.height = img.height;
            }
            desktopCtx.drawImage(img, 0, 0, desktopCanvas.width, desktopCanvas.height);
            URL.revokeObjectURL(img.src);

            frameCount++;
            const now = Date.now();
            if (now - lastFpsTime >= 1000) {
              document.getElementById('desktopFps').innerText = frameCount;
              frameCount = 0;
              lastFpsTime = now;
            }
          };
          img.src = URL.createObjectURL(event.data);
        }
      };

      desktopWS.onclose = () => {
        disconnectDesktopWS();
      };
    }

    function disconnectDesktopWS() {
      isDesktopStreaming = false;
      if (desktopWS) {
        desktopWS.close();
        desktopWS = null;
      }
      const btn = document.getElementById('btnConnectDesktop');
      if (btn) {
        btn.innerText = 'Conectar Pantalla';
        btn.style.background = 'var(--accent-gradient)';
      }
      const status = document.getElementById('desktopStatus');
      if (status) {
        status.innerText = 'Desconectado';
        status.style.color = 'var(--text-muted)';
      }
      const placeholder = document.getElementById('desktopPlaceholder');
      if (placeholder) placeholder.style.display = 'block';
      if (desktopCanvas) desktopCanvas.style.display = 'none';
    }

    function updateControlState() {
      allowRemoteControl = document.getElementById('chkAllowControl').checked;
      if (desktopCanvas) {
        if (allowRemoteControl) desktopCanvas.classList.add('controllable');
        else desktopCanvas.classList.remove('controllable');
      }
    }

    function setupDesktopInputListeners() {
      if (!desktopCanvas) return;

      const getNormalizedCoords = (e) => {
        const rect = desktopCanvas.getBoundingClientRect();
        return {
          x: Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width)),
          y: Math.max(0, Math.min(1, (e.clientY - rect.top) / rect.height))
        };
      };

      desktopCanvas.addEventListener('mousemove', (e) => {
        if (!isDesktopStreaming || !allowRemoteControl || !desktopWS) return;
        const coords = getNormalizedCoords(e);
        desktopWS.send(JSON.stringify({ type: 'mousemove', x: coords.x, y: coords.y }));
      });

      desktopCanvas.addEventListener('mousedown', (e) => {
        if (!isDesktopStreaming || !allowRemoteControl || !desktopWS) return;
        e.preventDefault();
        desktopCanvas.focus();
        const coords = getNormalizedCoords(e);
        desktopWS.send(JSON.stringify({ type: 'mousedown', x: coords.x, y: coords.y, button: e.button }));
      });

      desktopCanvas.addEventListener('mouseup', (e) => {
        if (!isDesktopStreaming || !allowRemoteControl || !desktopWS) return;
        e.preventDefault();
        const coords = getNormalizedCoords(e);
        desktopWS.send(JSON.stringify({ type: 'mouseup', x: coords.x, y: coords.y, button: e.button }));
      });

      desktopCanvas.addEventListener('contextmenu', (e) => e.preventDefault());

      desktopCanvas.addEventListener('wheel', (e) => {
        if (!isDesktopStreaming || !allowRemoteControl || !desktopWS) return;
        e.preventDefault();
        desktopWS.send(JSON.stringify({ type: 'wheel', deltaY: e.deltaY }));
      });

      window.addEventListener('keydown', (e) => {
        if (!isDesktopStreaming || !allowRemoteControl || !desktopWS || document.activeElement !== desktopCanvas) return;
        desktopWS.send(JSON.stringify({ type: 'keydown', key: e.key, code: e.code }));
      });

      window.addEventListener('keyup', (e) => {
        if (!isDesktopStreaming || !allowRemoteControl || !desktopWS || document.activeElement !== desktopCanvas) return;
        desktopWS.send(JSON.stringify({ type: 'keyup', key: e.key, code: e.code }));
      });
    }

    function copyDesktopShareUrl() {
      const shareInput = document.getElementById('desktopShareUrl');
      const btn = document.getElementById('btnCopyShareUrl');
      if (!shareInput) return;
      navigator.clipboard.writeText(shareInput.value).then(() => {
        if (btn) {
          const orig = btn.innerText;
          btn.innerText = '✅ ¡Copiado!';
          btn.style.background = '#10b981';
          setTimeout(() => {
            btn.innerText = orig;
            btn.style.background = '';
          }, 2500);
        }
      }).catch(() => {
        alert('Enlace para otro equipo:\n' + shareInput.value);
      });
    }

    function toggleDesktopFullscreen() {
      const vp = document.getElementById('desktopViewport');
      if (!vp) return;
      if (!document.fullscreenElement && !document.webkitFullscreenElement) {
        if (vp.requestFullscreen) {
          vp.requestFullscreen();
        } else if (vp.webkitRequestFullscreen) {
          vp.webkitRequestFullscreen();
        }
      } else {
        if (document.exitFullscreen) {
          document.exitFullscreen();
        } else if (document.webkitExitFullscreen) {
          document.webkitExitFullscreen();
        }
      }
    }

    document.addEventListener('fullscreenchange', handleFullscreenChange);
    document.addEventListener('webkitfullscreenchange', handleFullscreenChange);

    function handleFullscreenChange() {
      const isFs = !!(document.fullscreenElement || document.webkitFullscreenElement);
      const txt = document.getElementById('desktopFsText');
      const ico = document.getElementById('desktopFsIcon');
      if (txt) txt.innerText = isFs ? 'Salir de Pantalla Completa' : 'Pantalla Completa';
      if (ico) ico.innerText = isFs ? '✕' : '⛶';
    }

    let cachedLanInfo = null;
    function updateDesktopLanBanner(info) {
      if (info) cachedLanInfo = info;
      const data = cachedLanInfo;

      const isLocal = location.hostname === 'localhost' || location.hostname === '127.0.0.1' || location.hostname === '::1';
      const warning = document.getElementById('desktopMirrorWarning');
      const roleTitle = document.getElementById('desktopRoleTitle');
      const roleSub = document.getElementById('desktopRoleSubtitle');
      const roleIcon = document.getElementById('desktopRoleIcon');
      const shareInput = document.getElementById('desktopShareUrl');

      let lanHost = location.hostname;
      if (data && data.lan_ip && data.lan_ip !== '127.0.0.1') {
        lanHost = data.lan_ip;
      }
      const port = (data && data.port) ? data.port : (location.port || 8080);
      const targetUrl = `http://${lanHost}:${port}/#tab-desktop`;

      if (shareInput) shareInput.value = targetUrl;

      if (isLocal) {
        if (roleIcon) roleIcon.innerText = '📡';
        if (roleTitle) roleTitle.innerText = 'Modo Emisor: Transmitiendo esta PC';
        if (roleSub) roleSub.innerText = 'Para ver y controlar esta pantalla desde otro computador o celular en tu red local:';
        if (warning) warning.style.display = 'flex';
      } else {
        if (roleIcon) roleIcon.innerText = '👁️';
        if (roleTitle) roleTitle.innerText = 'Modo Receptor: Conectado a PC Remota';
        if (roleSub) roleSub.innerText = `Visualizando la pantalla remota del host (${location.hostname})`;
        if (warning) warning.style.display = 'none';
      }
    }
