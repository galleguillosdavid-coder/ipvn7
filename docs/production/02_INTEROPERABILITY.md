# IPv7 Interoperability Specification — Fase 2 y 12

**Misión**: IPv7 GLOBAL INTERNET PRODUCTION VALIDATION  
**Documento**: `02_INTEROPERABILITY.md`  
**Estado**: REGLAS DE VALIDACIÓN MULTIPLATAFORMA Y COMPATIBILIDAD

---

## 1. Matriz de Plataformas y Arquitecturas

| Sistema Operativo | Arquitectura | Kernel / Build | Rol de Prueba |
| :--- | :--- | :--- | :--- |
| **Windows 11** | `amd64` (x86_64) | 10.0.26100 | Nodo Host Primario / Nodo Remoto Notebook |
| **Linux (Ubuntu)** | `amd64` (x86_64) | WSL2 Linux 6.6.x | Nodo Linux Nativo / Receptor y Emisor Inter-OS |
| **Linux (Simulación)** | `arm64` (cross-build) | Binario compilado cross | Validación de compilación e invariantes |

---

## 2. Invariantes de Wire Format y Serialización

1. **CBOR Determinista**: Todos los paquetes del plano de control y datos en IPv7 deben respetar el formato canónico RFC 8949 (CBOR) con codificación de longitud definida.
2. **Endianness de Red**: Todos los enteros en la cabecera IPv7 (Version, HopLimit, Length, SequenceNumber, Timestamp) se codifican en Big-Endian (Network Byte Order).
3. **Cero-Copia y Alineación**: Las estructuras internas del Adapter Layer no deben depender de la alineación de memoria del compilador de un sistema operativo particular (struct packing seguro).
4. **Interoperabilidad Inter-OS Verificada**:
   - Windows <-> Linux (WSL2 / Nativo)
   - Windows <-> Windows (PC Principal <-> Notebook Físico)
   - Linux <-> Linux (Contenedores / Multi-proceso)

---

## 3. Compatibilidad de Versiones (Upgrade & Rollback)

- **Compatibilidad Hacia Atrás**: Un nodo v0.1 debe ser capaz de procesar handshakes y paquetes de datos emitidos por un nodo v0.2 si la cabecera mantiene compatibilidad de protocolo.
- **Rollback Seguro**: Al degradar el binario a una versión anterior, la base de datos de identidad criptográfica y enrutamiento (Kùzu / KV Store) debe permanecer íntegra sin requerir migraciones destructivas.
