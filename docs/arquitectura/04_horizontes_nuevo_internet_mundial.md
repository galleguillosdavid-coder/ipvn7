# IPv7 Arquitectura: Los Cuatro Horizontes hacia el Nuevo Internet Mundial

**Misión Estratégica**: Transformar a IPv7 de un sustrato de red superpuesta verificado en laboratorio y producción canaria hacia el estándar de transporte e interconexión soberana para un Internet mundial sin censura, sin intermediarios y con cifrado inviolable por defecto.  
**Autoridad de Decisión**: David Galleguillos (@galleguillosdavid-coder)  
**Operador & Diseñador Experimental**: Antigravity (Google DeepMind)  
**Auditor Externo**: ChatGPT (OpenAI)  
**Fecha de Fijación**: 2026-09-09  
**Estado**: **OBJETIVOS FIJADOS CON PROTOCOLO CIENTÍFICO ESTRICTO**  

---

## 1. El Diagnóstico Fundamental del Internet Actual

El protocolo IPv4 (1981) y su sucesor IPv6 (1998) fueron diseñados bajo supuestos que hoy resultan obsoletos y peligrosos para la soberanía de los usuarios:

1. **Confusión entre Identidad y Ubicación**: La dirección IP determina simultáneamente *quién eres* y *dónde estás conectado*. Si un dispositivo cambia de Wi-Fi a datos celulares o cambia de red, la dirección cambia y toda sesión TCP/IP se destruye.
2. **Carencia Congénita de Seguridad**: El cifrado y la autenticación fueron agregados posteriormente como parches de capa superior (TLS, IPsec, VPNs). Los encabezados, metadatos, nombres de dominio (DNS) y rutas de paquetes siguen expuestos a vigilancia masiva, espionaje de ISP e intercepción estatal.
3. **Dependencia de Infraestructuras Centrales Vulnerables**: El ecosistema depende críticamente de monopolios de telecomunicaciones (BGP), 13 clústeres de servidores raíz de DNS (ICANN) y autoridades de certificación (CAs) sujetas a coerción geopolítica y censura arbitraria.

---

## 2. La Definición y Solución de IPv7

IPv7 resuelve estas fallas en su raíz mediante cuatro pilares demostrados empíricamente:
* **Identidad Soberana Desacoplada**: El identificador global es una clave pública inmutable **Ed25519 (`DID`)**, independiente de cualquier dirección IP física.
* **Cero Caída de Conexión (IP Roaming)**: Los endpoints físicos son efímeros y se actualizan dinámicamente sin interrumpir el túnel criptográfico ni renegociar claves.
* **Cifrado E2EE Universal y Obligatorio**: Cifrado mutuo autenticado mediante Noise Protocol XX y ChaCha20-Poly1305. Ningún datagrama viaja en texto claro.
* **Enrutamiento Mundo Pequeño (Kleinberg)**: Convergencia logarítmica acotada a 12 anillos sin requerir tablas de ruteo gigantescas ni monopolios BGP.

Para expandir esta arquitectura hacia la escala global, se fijan **cuatro horizontes estratégicos gobernados por el método científico experimental**.

---

## 3. Matriz de los Cuatro Horizontes Estratégicos

```text
 ┌─────────────────────────────────────────────────────────────────────────────┐
 │                      EL NUEVO INTERNET MUNDIAL IPv7                         │
 └─────────────────────────────────────────────────────────────────────────────┘
                                       │
      ┌─────────────────┬──────────────┴──────┬──────────────────┐
      ▼                 ▼                     ▼                  ▼
┌─────────────┐   ┌─────────────┐       ┌─────────────┐    ┌─────────────┐
│ HORIZONTE 1 │   │ HORIZONTE 2 │       │ HORIZONTE 3 │    │ HORIZONTE 4 │
│   TUN/TAP   │   │  DHT PURA   │       │ ONION MESH  │    │  OFF-GRID   │
│  Universal  │   │  Soberana   │       │  Multi-Hop  │    │   Físico    │
└─────────────┘   └─────────────┘       └─────────────┘    └─────────────┘
```

---

### 🌐 HORIZONTE 1: El Adaptador TUN/TAP Universal (Cero Modificación de Apps)

#### Objetivo
Permitir que cualquier sistema operativo (Windows, Linux, macOS, Android, iOS, routers OpenWrt) cree una interfaz de red virtual nativa (`ipv70`) para que **todo el tráfico de red de cualquier aplicación existente (navegadores, streaming, juegos, bases de datos, SSH)** fluya de forma transparente a través de la malla cifrada IPv7.

#### Especificaciones Técnicas
- Asignación de rangos IP virtuales (ej. `10.7.0.0/16` o prefijo IPv6 único basado en el hash del DID) mapeados a identidades Ed25519.
- Emulación transparente de capa 3 con encapsulación de datagramas IP en contenedores CBOR firmados de IPv7.
- Manejo determinista de MTU (1280 bytes para evitar fragmentación externa).

#### Protocolo de Verificación y Criterios de Éxito
| Métrica / Prueba | Umbral de Éxito | Comportamiento Inaceptable |
| :--- | :--- | :--- |
| **Throughput Wire-Speed** | $\ge$ 500 Mbps en hardware moderno | $< 50$ Mbps o saturación de CPU $> 30\%$ |
| **Pérdida de Paquetes (PDR)**| $\ge 99.9\%$ en canal estable | Pérdida $> 0.5\%$ en régimen normal |
| **Compatibilidad de Apps** | Navegadores web, cURL, SSH y ping operando sin modificar código | Fallos en reensamblado de paquetes TCP |
| **Preservación del Core** | **0 cambios en `core/`** (implementado como adapter en `adapters/tun/`) | Modificación invasiva al Core |

---

### 🌐 HORIZONTE 2: Eliminación de Terceros (P2P Puro & DHT Soberana sin Firebase)

#### Objetivo
Reemplazar por completo cualquier servicio centralizado o de terceros (como Firebase Realtime Database) mediante una **DHT Kademlia global autoinmune** y una red distribuida de nodos bootstrap comunitarios.

#### Especificaciones Técnicas
- Tablas de enrutamiento DHT basadas en métrica XOR de 256 bits congruente con las claves Ed25519.
- Publicación de registros firmados de endpoints con tiempo de expiración (TTL) y refresco automático.
- Protección estricta contra ataques Sybil y envenenamiento de tablas mediante validación criptográfica de firmas en cada entrada.

#### Protocolo de Verificación y Criterios de Éxito
| Métrica / Prueba | Umbral de Éxito | Comportamiento Inaceptable |
| :--- | :--- | :--- |
| **Latencia de Resolución DID** | $< 1.500$ ms en topología global | Tiempos de espera $> 5.000$ ms |
| **Resistencia a Partición** | Autorecuperación en $< 10$ s tras aislamiento de clúster | Desconexión permanente de peers |
| **Sobrecarga de Red (Overhead)**| $< 2$ KB/s por nodo en mantenimiento de DHT | Ráfagas descontroladas de broadcasting |
| **Inmunidad Sybil** | 100% de registros apócrifos rechazados | Inyección de endpoints falsos en la tabla |

---

### 🌐 HORIZONTE 3: Enrutamiento Cebolla Multi-Salto (Onion Multi-Hop Routing)

#### Objetivo
Implementar un esquema de enrutamiento indirigible y multi-salto basado en el diseño de **paquetes por capas (estilo Sphinx / Onion Routing)** sobre la topología de Mundo Pequeño de Kleinberg.

#### Especificaciones Técnicas
- Si el nodo origen $A$ desea comunicarse con el nodo destino $D$, el datagrama viaja a través de nodos intermedios $B$ y $C$ ($A \to B \to C \to D$).
- Cada nodo intermediario sólo puede pelar su capa criptográfica efímera con su clave privada, conociendo únicamente el salto anterior y el salto siguiente, sin saber quién inició la comunicación ni quién es el destinatario final.
- Paquetes de longitud fija (padding determinista) para neutralizar el análisis de tráfico por longitud o cadencia.

#### Protocolo de Verificación y Criterios de Éxito
| Métrica / Prueba | Umbral de Éxito | Comportamiento Inaceptable |
| :--- | :--- | :--- |
| **Aislamiento de Metadatos** | 0 fuga de identidades de origen/destino a nodos intermediarios | Fuga de DID en encabezados transitarios |
| **Sobrecarga Criptográfica** | $< 64$ bytes de cabecera por cada salto intermedio | Crecimiento no lineal del tamaño del paquete |
| **Latencia por Salto** | $< 5$ ms adicionales por nodo de retransmisión | Encolamiento excesivo o degradación severa |
| **Preservación Antireplay** | Descarte inmediato de paquetes retransmitidos o bifurcados | Replay exitoso a través de nodos puente |

---

### 🌐 HORIZONTE 4: Malla Física Fuera de Internet (Mesh Off-Grid / Wi-Fi Direct / LoRa)

#### Objetivo
Dotar al protocolo IPv7 de la capacidad de operar en **redes físicas completamente desacopladas de Internet** (redes ad-hoc, Wi-Fi Direct, Bluetooth LE, radioenlaces LoRa de largo alcance o redes de fibra oscura locales).

#### Especificaciones Técnicas
- Capa de transporte abstracta capaz de emitir sobre interfaces no-IP o datagramas de baja velocidad con compresión extrema de cabeceras.
- Descubrimiento autónomo de proximidad física en capa 2 sin requerir gateways ni asignación DHCP tradicional.
- Conmutación híbrida transparente: si Internet está disponible, el nodo usa rutas globales; si Internet cae (bloqueo regional, catástrofe natural), el nodo conmuta automáticamente a la malla física local sin interrumpir la identidad soberana.

#### Protocolo de Verificación y Criterios de Éxito
| Métrica / Prueba | Umbral de Éxito | Comportamiento Inaceptable |
| :--- | :--- | :--- |
| **Continuidad de Identidad** | **100% de preservación del DID** en modo Off-Grid | Necesidad de reconfiguración o nuevas credenciales |
| **Descubrimiento Local** | Enlace automático con nodos adyacentes en $< 2.000$ ms | Aislamiento en presencia de pares físicos |
| **Tolerancia a Alto RTT** | Soporte de enlaces con RTT de hasta 10.000 ms (LoRa/Satélite) | Timeouts agresivos que aborten la sesión |
| **Conmutación Híbrida** | Handover automático Online $\leftrightarrow$ Off-Grid en $< 1.000$ ms | Caída forzada del proceso de nodo |

---

## 4. Protocolo Científico de Desarrollo para Cada Horizonte

Para avanzar en cada uno de estos cuatro horizontes con **asertividad matemática y rigor epistémico**, se aplicará estrictamente el protocolo de ingeniería de `docs/3.md`:

```text
       ┌────────────────────────────────────────────────────────┐
       │                CICLO CIENTÍFICO IPv7                   │
       └────────────────────────────────────────────────────────┘
                                   │
 1. HIPÓTESIS FORMAL  ─────────────► Definición precisa y falsable
                                   │
 2. PERFILADO BASELINE ────────────► Captura de rendimiento actual (0 allocs)
                                   │
 3. PROTOTIPO EN ADAPTERS ─────────► Implementación desacoplada (Core congelado)
                                   │
 4. BATERÍA ADVERSARIAL ───────────► Inyección de caos, estrés y fuzzing
                                   │
 5. EVALUACIÓN DE SOAK ────────────► Pruebas de larga duración (cero fugas)
                                   │
 6. REPORTE TRIPARTITO ────────────► Antigravity ejecuta, ChatGPT audita, David decide
```

1. **Invarianza del Core**: Ningún horizonte puede alterar la semántica criptográfica fundamental de [`core/`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/core) salvo demostración empírica incontrovertible.
2. **Medir antes de Optimizar**: Cada nuevo adaptador o mecanismo debe medirse con microbenchmarks (`go test -bench -benchmem`) y verificar que no introduzca contención en los sockets UDP ni sobrecostes en el Garbage Collector.
3. **Memoria de Fracasos Activa**: Cualquier hipótesis técnica que resulte ineficaz o inviable se registrará formalmente en [`docs/engineering/BOTTLENECKS.md`](file:///c:/Users/Frondabrick/Desktop/dvd/Ipv7/docs/engineering/BOTTLENECKS.md) para no volver a gastar recursos en caminos muertos.
4. **Regla de Dos IA y Decisión Soberana**: Las propuestas y análisis generados por Antigravity y auditados por ChatGPT serán sometidos a la revisión y aprobación expresa de **David**.

---

## 5. Criterio de Éxito Global

IPv7 habrá completado estos cuatro horizontes cuando un usuario en cualquier lugar del mundo pueda:
1. Instalar el nodo en su teléfono, ordenador o router doméstico en 1 clic.
2. Abrir cualquier navegador o aplicación y navegar por el ciberespacio soberano sin censura, sin operadoras intermediarias y con criptografía militar por defecto.
3. Mantener su comunicación viva aunque cambie de país, salte de red o enfrente un apagón total de la infraestructura tradicional de telecomunicaciones.
