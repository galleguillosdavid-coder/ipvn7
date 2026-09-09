# IPv7 Security & Adversarial Internet Specification — Fase 6 y 13

**Misión**: IPv7 GLOBAL INTERNET PRODUCTION VALIDATION  
**Documento**: `06_SECURITY.md`

---

## 1. Clasificación del Estado Criptográfico

Para cumplir con el rigor epistémico del proyecto, las garantías criptográficas se dividen estrictamente en cuatro estados:
- **`IMPLEMENTED`**: Código fuente presente y operativo en el repositorio.
- **`TESTED`**: Validado mediante baterías de tests automáticos con vectores de prueba conocidos.
- **`BENCHMARKED`**: Rendimiento medido empíricamente (ciclos/segundo, ns/op, MB/s).
- **`FORMALLY GUARANTEED`**: Propiedades respaldadas por demostraciones matemáticas de seguridad bajo modelos estándar (e.g., Modelo de Oráculo Aleatorio / Estándares NIST).

---

## 2. Batería de Pruebas Adversariales en Capa de Red

1. **UDP Packet Flood**: Saturación de tráfico aleatorio a máxima tasa disponible.
2. **Malformed CBOR Injection**: Payloads no conformes diseñados para explotar parsers.
3. **Oversized Frames (> PMTU)**: Paquetes que exceden la MTU para probar descarte y protección de buffers.
4. **Replay Attack Exhaustivo**: 100.000 paquetes duplicados contra sesión Noise XX activa.
5. **Invalid Signatures & Fake DIDs**: Intentos de suplantación de identidad en el plano de control.
6. **Handshake Flooding**: Inyección de 10.000 intentos de inicio de sesión con claves efímeras falsas para forzar agotamiento de CPU o goroutines.
7. **HopLimit / Sequence Abuse**: Intentos de manipulación de cabeceras IPv7.

---

## 3. Caracterización de Mitigación Arquitectónica (ADV-01)

- **Shared Adapter (Socket Único)**: Medición del cuello de contención en el buffer `SO_RCVBUF` del sistema operativo ante fuego cruzado.
- **Isolated Adapters (Segregación de Puertos)**: Verificación del aislamiento entre el puerto de señalización/descubrimiento público y el puerto dedicado de datos E2EE.
