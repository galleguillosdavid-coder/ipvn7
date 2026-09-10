---
name: deepseek-worker
description: Delegador de tareas pesadas a la API de DeepSeek (DeepSeek-V4.1-Flash y DeepSeek-v4-pro) con visión multimodal para ahorrar tokens en Antigravity. Úsalo para análisis masivos de código, auditorías de documentación, revisión visual de capturas de pantalla/UI y generación de reportes extensos.
---

# DeepSeek Worker Skill (Token Saver & Multimodal Vision)

Esta habilidad permite a **Antigravity** delegar el trabajo cognitivo pesado (análisis de miles de líneas de código, auditorías formales exhaustivas, inspección visual de capturas de pantalla/UI) a la API de **DeepSeek**, evitando consumir los tokens de contexto de la sesión principal.

## Modelos Disponibles

1. **`deepseek-flash` (DeepSeek-V4.1-Flash - Modelo por defecto)**:
   - Capacidad nativa de **Visión Multimodal** (imágenes, capturas de UI, diagramas).
   - Altísima velocidad de inferencia y alto throughput.
   - Ideal para: revisión de UI, análisis de capturas, síntesis de código grande, auditorías.
2. **`deepseek-v4-pro`**:
   - Razonamiento textual profundo (chain-of-thought interno).
   - Ideal para: derivaciones matemáticas, deducción lógica de concurrencia y algoritmos complejos sin imágenes.

## Ubicación de Herramientas

- Motor Python: `tools/deepseek/deepseek_worker.py`
- Wrapper PowerShell: `scripts/ask_deepseek.ps1`
- Wrapper Bash (WSL2/Linux): `scripts/ask_deepseek.sh`
- Configuración de API Key: `tools/deepseek/.env` (archivo protegido e ignorado en Git)

---

## Modos de Uso desde Antigravity

### 1. Análisis Visual de Capturas de Pantalla / UI
Para auditar una pantalla o imagen sin gastar tokens de visión en Antigravity:

```powershell
python tools/deepseek/deepseek_worker.py -i "ruta/a/captura.png" -p "Analiza problemas de contraste, bugs visuales y métricas anómalas" -t vision_ui
```

### 2. Revisión de Código Masivo (Ahorro del 90% de Tokens)
Para que DeepSeek lea archivos grandes sin cargarlos al contexto de la conversación:

```powershell
python tools/deepseek/deepseek_worker.py -f archivo1.go archivo2.go -p "Busca data races, fugas de memoria o cuellos de botella" -t code_review -o docs/engineering/deepseek_review.md
```

### 3. Auditoría de Documentos y Contratos
```powershell
python tools/deepseek/deepseek_worker.py -f docs/production/RELEASE_NOTES_v0.5.0.md docs/engineering/ASSERTION_CONTRACT.md -p "Audita la coherencia de cotas y números entre ambos documentos" -t audit
```

### 4. Razonamiento Algorítmico Puro con `deepseek-v4-pro`
```powershell
python tools/deepseek/deepseek_worker.py -m deepseek-v4-pro -p "Explica la derivación formal del fan-out 10 en streaming en cascada" --show-reasoning
```

---

## Flujo de Trabajo Recomendado para Antigravity

1. **Recibir tarea extensa del usuario**: En lugar de leer 10 archivos a la vez con `view_file`.
2. **Ejecutar comando**:
   `python tools/deepseek/deepseek_worker.py -f <archivos> -p "<instrucción>" -o scratch/analisis.md`
3. **Leer solo el resultado sintetizado**:
   `view_file scratch/analisis.md`
4. **Responder al usuario**: Respuestas rápidas, precisas y con un consumo insignificante de tokens en la sesión.
