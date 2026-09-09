# Capítulo 5: Resolución de Problemas y Preguntas Frecuentes (FAQ)

Este capítulo ofrece una guía de diagnóstico para resolver incidentes comunes en la operación de nodos IPv7.

---

## 1. Guía Rápida de Diagnóstico de Incidentes

| Síntoma observado | Causa más probable | Acción recomendada |
|---|---|---|
| Al iniciar el nodo sale `bind: address already in use` | Otro proceso u otra instancia de IPv7 ya está usando el puerto UDP o HTTP. | El nodo auto-asigna puertos alternativos. Revisa la consola para ver qué puerto asignó (ej. 7003 o 8081). Puedes forzar puertos libres con `-port <N> -ui <M>`. |
| El navegador no abre `http://localhost:8080` | El puerto web fue modificado o el binario se ejecutó con `-ui 0`. | Verifica en la consola la línea `[UI] Dashboard web available at: ...`. Asegúrate de no haber usado `-ui 0`. |
| El handshake con un peer falla con `timeout` | El peer remoto está apagado, el firewall bloquea UDP o el router tiene NAT simétrico sin relay. | 1. Comprueba que el peer remoto esté encendido.<br>2. Verifica permisos en el Firewall de Windows.<br>3. Si están en redes distintas, asegura que ambos alcancen el servidor STUN. |
| El chat muestra mensajes como texto plano en vez de `[🔒 E2EE]` | No se ha obtenido la clave pública X25519 del destinatario. | Utiliza el comando `/connect <ip:port>` en el chat CLI para forzar el apretón de manos criptográfico antes de chatear. |
| WSL2 no arranca el nodo | Falta el paquete Go en WSL o el script de shell no tiene permisos de ejecución. | Ejecuta en WSL2: `chmod +x ./scripts/wsl_node.sh` y recompila con `.\scripts\build_linux.ps1`. |

---

## 2. Inspección de Logs en Vivo

### Logs de Windows:
Los logs del nodo principal se imprimen directamente en la ventana de PowerShell/CMD donde fue ejecutado. Puedes redirigirlos a un archivo con:
```powershell
.\ipv7-node.exe *>&1 | Tee-Object -FilePath .\logs\windows_node.log
```

### Logs de WSL2:
Al usar el script de supervisión, los logs se guardan de forma continua en `logs/wsl_node.log`. Para visualizarlos:
```powershell
Get-Content -Path .\logs\wsl_node.log -Wait -Tail 30
```

---

## 3. Diagnóstico de Topología con la Base de Datos Kùzu (Cypher)

El proyecto incluye el motor de grafos **Kùzu** embebido para auditar la topología de la red y las conexiones entre nodos.

Para abrir la consola interactiva de Kùzu en Windows:
```powershell
.\tools\kuzu\kuzu.exe .kuzu_index/
```

O en WSL2 Linux:
```bash
./tools/kuzu/kuzu .kuzu_index/
```

### Consultas de diagnóstico útiles:

- **Listar todas las conexiones activas en la malla y sus latencias:**
  ```cypher
  MATCH (p:Peer)-[r:CONNECTED_TO]->(m:Peer)
  RETURN p.id, r.adapter, r.latency_ms, m.id;
  ```

- **Encontrar el nodo con mayor grado de conectividad (Hubs):**
  ```cypher
  MATCH (p:Peer)-[r:CONNECTED_TO]->()
  RETURN p.id, count(r) AS conexiones
  ORDER BY conexiones DESC;
  ```

- **Verificar la presencia de tu propio nodo:**
  ```cypher
  MATCH (p:Peer {is_local: true})
  RETURN p.id, p.endpoint;
  ```

---

## 4. Preguntas Frecuentes (FAQ)

### ¿IPv7 necesita que pague por servidores o dominios?
**No.** IPv7 es 100% descentralizado y de código abierto. No requiere servidores en la nube, ni cuentas de usuario, ni suscripciones de pago.

### ¿Puede mi proveedor de Internet (ISP) ver lo que hago en IPv7?
Tu proveedor de Internet solo ve que intercambias paquetes de datos cifrados (UDP o QUIC) con otros puertos de Internet. **No puede ver el contenido**, no puede saber con quién hablas dentro de la red superpuesta ni puede alterar los mensajes gracias a las firmas digitales Ed25519 y al cifrado ChaCha20-Poly1305.

### ¿Se pueden enviar archivos pesados por IPv7?
**Sí.** Gracias al adaptador **QUIC (`adapters/quic_adapter.go`)**, IPv7 soporta transferencia de archivos de gran tamaño (vídeos, ISOs, paquetes masivos de 5MB+) gestionando automáticamente el control de flujo, la retransmisión de paquetes perdidos y la fragmentación sin sobrecargar al usuario.

### ¿Qué pasa si apago mi computadora? ¿Pierdo mi identidad?
En la versión estándar, cada ejecución genera una nueva identidad efímera para máxima privacidad. Si deseas conservar tu clave fija, puedes guardarla en una variable de entorno o archivo de configuración seguro (consulta la carpeta de sugerencias de mejora de código).
