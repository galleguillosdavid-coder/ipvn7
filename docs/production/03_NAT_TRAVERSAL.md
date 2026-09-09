# IPv7 NAT Traversal & Connectivity Matrix — Fase 2

**Misión**: IPv7 GLOBAL INTERNET PRODUCTION VALIDATION  
**Documento**: `03_NAT_TRAVERSAL.md`

---

## 1. Matriz de Combinaciones de Conectividad

| Origen | Destino | Mecanismo Primario | Mecanismo Fallback | Estado Requerido |
| :--- | :--- | :--- | :--- | :--- |
| **Public IPv4** | **Public IPv4** | Direct UDP Socket | DERP Relay | `DIRECT` |
| **Public IPv6** | **Public IPv6** | Direct UDP Socket | DERP Relay | `DIRECT` |
| **Full Cone NAT** | **Public IPv4** | Direct Hole Punching | DERP Relay | `DIRECT` |
| **Restricted Cone** | **Restricted Cone** | STUN Hole Punching | DERP Relay | `DIRECT / RELAY` |
| **Port Restricted** | **Port Restricted** | UDP Hole Punching | DERP Relay | `DIRECT / RELAY` |
| **Symmetric NAT** | **Cualquiera** | DERP Relay | N/A | `RELAY` |
| **CGNAT** | **CGNAT** | DERP Relay | N/A | `RELAY` |
| **IPv4 Endpoint** | **IPv6 Endpoint** | Dual-Stack Relay | N/A | `RELAY / ADAPTER` |
| **Cambio de IP/Puerto** | **Sesión Activa** | Roaming Handshake | DERP Relay | `RECOVERY` |

---

## 2. Requisitos de Ejecución por Caso de Prueba

Cada prueba en la matriz debe ejecutar y registrar:
1. **DISCOVERY**: Resolución de peer y obtención de candidatos de dirección vía DHT / STUN.
2. **HANDSHAKE**: Establecimiento de sesión criptográfica Noise XX.
3. **E2EE**: Canal cifrado con ChaCha20-Poly1305.
4. **DATA**: Flujo de paquetes de datos de prueba con verificación de integridad.
5. **REKEY**: Rotación dinámica de claves simétricas sin interrupción de flujo.
6. **ROAMING**: Cambio forzado de endpoint remoto manteniendo la identidad lógica del DID.
7. **RECOVERY**: Tiempo de restablecimiento tras interrupción temporal.
