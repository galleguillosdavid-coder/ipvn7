# Capítulo 1: Conceptos y Fundamentos de IPv7 para Humanos

## 1. ¿Por qué existe IPv7?

El modelo de Internet tradicional se diseñó en la década de 1970 sobre supuestos que ya no se cumplen:
1. **Direcciones IP fijas y públicas**: Hoy la inmensa mayoría de dispositivos están atrapados tras NAT domésticos, firewalls corporativos o redes móviles (CGNAT). No tienen IP pública ni se pueden comunicar directamente.
2. **Confianza ciega en la red de capa 3**: Las direcciones IP numéricas no ofrecen ninguna garantía de autenticidad; cualquiera puede falsificar paquetes o interceptar tráfico sin cifrar.
3. **Dependencia de servidores centrales y monopolios**: Para conectar dos dispositivos normalmente dependes de servidores intermedios en la nube de terceros (Zoom, WhatsApp, TeamViewer, AWS).

**IPv7** no sustituye los cables físicos ni los routers de tu proveedor de fibra. **IPv7 es una Red Superpuesta Descentralizada (Overlay Network)** que opera en el espacio de usuario (Capa 7 de aplicación). Construye una malla global peer-to-peer donde cada dispositivo es soberano.

```
+------------------------------------------------------------------------+
|                          IPv7 Overlay Network                          |
|   (Identidades Criptográficas Ed25519, Cifrado E2EE, Mundo Pequeño)    |
+------------------------------------------------------------------------+
|      Adaptadores de Transporte: UDP  |  QUIC (TLS 1.3)  |  WebRTC      |
+------------------------------------------------------------------------+
|       Red Física / Internet Tradicional: Fibra, Wi-Fi, 4G/5G, NAT       |
+------------------------------------------------------------------------+
```

---

## 2. Direcciones IP vs Identidades Criptográficas

En IPv4 tu dirección cambia si te mueves de tu casa al trabajo o si cambias de Wi-Fi a datos móviles. En IPv7:

- **Tu dirección lógica ES tu Clave Pública Ed25519**: Una cadena única e inmutable de 32 bytes (64 caracteres hexadecimales).
  - *Ejemplo:* `4a8f9c2d1e0b5a6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c`
- **Inmutabilidad de Ubicación**: Da igual si tu dispositivo está en Santiago de Chile, Tokio o conectado al Wi-Fi de una cafetería: tu dirección IPv7 siempre es la misma.
- **Autenticación Criptográfica Automática**: Todo paquete que envías es firmado digitalmente con tu clave privada. Nadie en el planeta puede suplantar tu identidad ni falsificar un mensaje a tu nombre.

---

## 3. Seguridad Matemática de Extremo a Extremo (E2EE)

La privacidad en IPv7 no es una promesa contractual; es una garantía matemática basada en dos primitivas modernas:

1. **Diffie-Hellman sobre Curva Elíptica X25519 (ECDH)**:
   - A partir de la clave de identidad de cada nodo, se deriva matemáticamente una clave de cifrado asimétrica X25519.
   - Dos nodos pueden calcular un secreto compartido efímero instantáneamente sin haber hablado previamente por canales seguros.
2. **Cifrado Autenticado ChaCha20-Poly1305 (AEAD)**:
   - Estándar de grado militar (RFC 8439) adoptado por WireGuard y TLS 1.3.
   - Proporciona tanto confidencialidad (nadie puede leer el contenido) como integridad autenticada (cualquier manipulación de 1 solo bit hace que el mensaje sea descartado).
   - **Nodos intermedios ciegos**: Si tu mensaje viaja por 5 nodos intermediarios o a través de un servidor Relay de emergencia, ninguno de ellos puede ver el contenido.

---

## 4. Enrutamiento del "Mundo Pequeño" (12 Grados de Separación)

¿Cómo encuentra tu nodo a otro nodo entre millones sin saturar su memoria?

Inspirado en la teoría sociológica de los "seis grados de separación" y las redes de Small-World (Kleinberg / Kademlia):
- **Estructura en 12 Anillos Logarítmicos**: Tu nodo divide el universo criptográfico en 12 niveles de distancia matemática (métrica XOR).
- **Máximo 10 contactos por anillo**: Tu nodo solo guarda como máximo **120 peers** en memoria viva.
- **Escala planetaria**: Con solo 120 contactos por nodo y un límite de 12 saltos (`HopLimit = 12`), la red puede enrutar datos de forma voraz (*Greedy Routing*) a través de hasta **$10^{12}$ dispositivos (1 billón de nodos)** en todo el mundo.

```
       [ Tu Nodo Local ]
         /      |      \
     Anillo 1 Anillo 6 Anillo 12
     (Vecinos  (Media   (Extremo
      Locales) Distancia) Global)
         \      |      /
     [ Destino Final en <= 12 Saltos ]
```

---

## 5. Streaming en Cascada (Fan-Out 10)

Cuando un nodo quiere emitir datos pesados (audio, video, telemetría o archivos en vivo) a cientos de personas, no transmite 100 veces el mismo flujo porque agotaría su ancho de banda de subida.

- El nodo emisor envía el flujo únicamente a sus **10 hijos directos**.
- Cada hijo retransmite automáticamente a sus **10 hijos**, y así sucesivamente.
- **Progresión geométrica**: 
  - Nivel 1: 10 receptores.
  - Nivel 2: 100 receptores.
  - Nivel 3: 1.000 receptores.
  - Nivel 4: 10.000 receptores.
- El costo de red se distribuye solidariamente entre todos los miembros de la cascada.
