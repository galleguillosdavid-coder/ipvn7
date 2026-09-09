# Procedimientos de Diagnóstico y Operación Automatizada para Agentes de IA

Este documento contiene los Procedimientos Operativos Estándar (SOP) que un agente de Inteligencia Artificial debe seguir paso a paso cuando se le solicite diagnosticar la red, verificar la salud de los nodos o solucionar anomalías de enrutamiento en IPv7.

---

## SOP 1: Auditoría Integral de Salud del Nodo Local

Cuando el usuario pregunte *"¿El nodo está funcionando bien?"* o *"Verifica el estado del nodo"*, la IA debe ejecutar el siguiente ciclo:

1. **Paso 1: Consultar estado base**:
   - Invocación: `GET http://localhost:8080/api/info`
   - Validación:
     - Comprobar que `identity` sea un hex de 64 caracteres válido.
     - Comprobar que `e2ee_pubkey` no esté vacío.
     - Registrar el número de `peers_count`.

2. **Paso 2: Evaluar la tabla de enrutamiento**:
   - Invocación: `GET http://localhost:8080/api/peers`
   - Validación:
     - Si `peers_count == 0`: Advertir al usuario de que el nodo está aislado en la red y sugerir conectar a un bootstrap peer conocido (`-peer <ip:port>`).
     - Si `peers_count > 0`: Calcular la latencia promedio y listar los peers más cercanos.

3. **Paso 3: Diagnosticar la topología de malla**:
   - Invocación: `GET http://localhost:8080/api/mesh`
   - Verificar si existen enlaces (`links.length > 0`) y constatar si los enlaces son cifrados (`encrypted == true`).

4. **Paso 4: Generar reporte estructurado**:
   - Devolver un resumen conciso con estado de claves, peers activos, latencia mínima/máxima y recomendaciones de acción si aplica.

---

## SOP 2: Diagnóstico de Fallo en Handshake Criptográfico

Si la comunicación con un peer falla o sale error de handshake:

1. **Paso 1: Ping a la dirección física**:
   - Comprobar si la IP y el puerto UDP responden a nivel de red básica.
2. **Paso 2: Verificar versión del binario**:
   - Asegurarse de que ambos nodos estén corriendo versiones compatibles de serialización CBOR canónica.
3. **Paso 3: Validar STUN y NAT Reflexivo**:
   - Comprobar si el peer remoto tiene endpoints públicos descubiertos o si ambos están atrapados tras NAT simétrico.
   - Si ambos tienen NAT simétrico: Sugerir activar el adaptador Relay (`adapters/relay_adapter.go`).

---

## SOP 3: Verificación de Integridad de Código y Pruebas Unitarias

Cuando se solicite *"Verifica que el código de IPv7 compile y pase todos los tests"*:

1. **Paso 1: Ejecutar suite de pruebas en Go**:
   - Comando: `go test -v ./...`
   - Inspeccionar la salida en busca de `FAIL` en cualquier paquete (`core`, `adapters`, `dht`, `ui`).
2. **Paso 2: Validar condiciones de carrera (Race Detector)**:
   - Comando: `go test -race ./core`
   - Asegurar que no existan carreras en `SmallWorldRoutingTable` ni en `TunnelService`.
3. **Paso 3: Reportar cobertura y resultado**:
   - Confirmar si todos los componentes criptográficos, de cascada, de túneles y de interfaz web están en verde.

---

## SOP 4: Extracción de Telemetría Prometheus y Contratos OpenAPI

Cuando se solicite *"Analiza las métricas del nodo"* o *"Inspecciona la API viva"*:

1. **Paso 1: Consultar métricas Prometheus**:
   - Invocación: `GET http://localhost:8080/metrics`
   - Extraer valores:
     - `ipv7_peers_connected_total`: Cantidad de enlaces vivos.
     - `ipv7_cascade_children_count`: Nodos recibiendo difusión.
     - `ipv7_active_tunnels_count`: Túneles TCP mapeados.
     - `ipv7_vpn_running`: Estado del proxy SOCKS5.
2. **Paso 2: Consultar especificación OpenAPI viva**:
   - Invocación: `GET http://localhost:8080/api/openapi.json`
   - Validar versión OpenAPI 3.1.0 y catálogo de endpoints disponibles.
3. **Paso 3: Diagnosticar salud del servidor MCP**:
   - Invocación: `POST http://localhost:8080/api/mcp` con `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`.
   - Confirmar respuesta JSON-RPC válida con la lista de herramientas disponibles.

