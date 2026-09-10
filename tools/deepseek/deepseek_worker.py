#!/usr/bin/env python3
"""
DeepSeek Worker & Token Offloader for IPv7
==========================================
Delegates large analysis, code review, documentation generation, and
multimodal visual inspection (UI screenshots, diagrams) to DeepSeek API
to minimize token consumption in the primary agent context.

Supported Models:
- deepseek-flash: Multimodal Vision & high-throughput inference (Default)
- deepseek-v4-pro: Deep textual reasoning & complex algorithmic derivation
"""

import os
import sys
import json
import argparse
import base64
import mimetypes
import urllib.request
import urllib.error
from pathlib import Path

# Force UTF-8 on Windows console
if sys.stdout.encoding != 'utf-8':
    try:
        sys.stdout.reconfigure(encoding='utf-8')
        sys.stderr.reconfigure(encoding='utf-8')
    except AttributeError:
        pass

# Default API Key provided by user
DEFAULT_API_KEY = "sk-115d108618234937b822e73c5581502d"
API_URL = "https://api.deepseek.com/chat/completions"

SYSTEM_PROMPTS = {
    "general": (
        "Eres un asistente de inteligencia artificial de élite especializado en ingeniería "
        "de sistemas de red descentralizados, criptografía P2P y optimización de software de alto rendimiento."
    ),
    "code_review": (
        "Eres un revisor de código senior y auditor de seguridad. Analiza el código adjunto "
        "en busca de: 1) condiciones de carrera (data races), 2) fugas de memoria o goroutines, "
        "3) violaciones de cotas o umbrales de rendimiento, 4) bugs de concurrencia y límites."
    ),
    "audit": (
        "Eres un auditor técnico riguroso. Tu objetivo es auditar la coherencia entre afirmaciones "
        "documentales, contratos matemáticos e implementaciones reales de tests. Sé implacable "
        "con afirmaciones no demostradas o discrepancias en cotas numéricas."
    ),
    "vision_ui": (
        "Eres un experto en diseño de interfaces, UX y observabilidad de telemetría gráfica. "
        "Analiza la captura visual adjunta: identifica anomalías, textos ilegibles, problemas de contraste, "
        "métricas anómalas en dashboards y posibles inconsistencias visuales."
    ),
    "refactor": (
        "Eres un arquitecto de software de sistemas. Propón optimizaciones y refactorizaciones limpias "
        "manteniendo invariantes el rendimiento, la compatibilidad binaria y cero alocaciones de heap innecesarias."
    )
}

def load_api_key(cli_key=None):
    if cli_key:
        return cli_key
    env_key = os.environ.get("DEEPSEEK_API_KEY")
    if env_key:
        return env_key
    env_file = Path(__file__).parent / ".env"
    if env_file.exists():
        for line in env_file.read_text(encoding="utf-8").splitlines():
            line = line.strip()
            if line.startswith("DEEPSEEK_API_KEY="):
                return line.split("=", 1)[1].strip()
    return DEFAULT_API_KEY

def encode_image(image_path):
    path = Path(image_path)
    if not path.exists():
        raise FileNotFoundError(f"Imagen no encontrada: {image_path}")
    mime_type, _ = mimetypes.guess_type(str(path))
    if not mime_type:
        mime_type = "image/png"
    with open(path, "rb") as f:
        encoded = base64.b64encode(f.read()).decode("utf-8")
    return f"data:{mime_type};base64,{encoded}"

def call_deepseek(api_key, model, messages, max_tokens=4096, temperature=0.3):
    headers = {
        "Authorization": f"Bearer {api_key}",
        "Content-Type": "application/json"
    }
    payload = {
        "model": model,
        "messages": messages,
        "max_tokens": max_tokens,
        "temperature": temperature
    }
    req = urllib.request.Request(API_URL, data=json.dumps(payload).encode("utf-8"), headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=120) as resp:
            data = json.loads(resp.read().decode("utf-8"))
            return data
    except urllib.error.HTTPError as e:
        error_body = e.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"HTTP {e.code} de DeepSeek API: {error_body}")
    except Exception as e:
        raise RuntimeError(f"Fallo al conectar con DeepSeek API: {e}")

def main():
    parser = argparse.ArgumentParser(description="DeepSeek Task Offloader & Vision Worker")
    parser.add_argument("--prompt", "-p", help="Instrucción directa de texto")
    parser.add_argument("--prompt-file", help="Ruta a archivo de texto con la instrucción")
    parser.add_argument("--image", "-i", action="append", default=[], help="Ruta a imagen(es) para análisis visual multimodal")
    parser.add_argument("--files", "-f", nargs="*", default=[], help="Archivos de código fuente o documentos a adjuntar")
    parser.add_argument("--model", "-m", default=None, help="Modelo: deepseek-flash (default / vision) o deepseek-v4-pro")
    parser.add_argument("--task", "-t", choices=list(SYSTEM_PROMPTS.keys()), default="general", help="Preset de tarea")
    parser.add_argument("--out", "-o", help="Archivo de destino donde escribir la respuesta")
    parser.add_argument("--api-key", help="Clave API DeepSeek personalizada")
    parser.add_argument("--max-tokens", type=int, default=4096, help="Límite de tokens de salida (default: 4096)")
    parser.add_argument("--show-reasoning", action="store_true", help="Mostrar razonamiento interno si está disponible")

    args = parser.parse_args()

    api_key = load_api_key(args.api_key)

    # Resolve model selection: If images are provided, deepseek-flash MUST be used
    if args.image:
        model = "deepseek-flash"
        if args.model and args.model != "deepseek-flash":
            print(f"[WARN] Se forzó modelo 'deepseek-flash' porque '{args.model}' no soporta visión multimodal.", file=sys.stderr)
    else:
        model = args.model if args.model else "deepseek-flash"

    # Assemble user prompt
    prompt_text = ""
    if args.prompt_file:
        prompt_text = Path(args.prompt_file).read_text(encoding="utf-8")
    elif args.prompt:
        prompt_text = args.prompt
    elif not sys.stdin.isatty():
        prompt_text = sys.stdin.read()
    else:
        prompt_text = "Por favor resume el estado del código y los archivos proporcionados."

    # Attach files if any
    files_payload = []
    for fpath in args.files:
        p = Path(fpath)
        if p.exists() and p.is_file():
            content = p.read_text(encoding="utf-8", errors="replace")
            files_payload.append(f"--- INICIO ARCHIVO: {p.name} ({p.as_posix()}) ---\n{content}\n--- FIN ARCHIVO ---")
        else:
            print(f"[WARN] Archivo no encontrado: {fpath}", file=sys.stderr)

    if files_payload:
        prompt_text = "\n\n".join(files_payload) + "\n\n" + prompt_text

    # Build messages
    system_instruction = SYSTEM_PROMPTS[args.task]
    messages = [{"role": "system", "content": system_instruction}]

    # Build user content (multimodal if images present)
    if args.image:
        user_content = [{"type": "text", "text": prompt_text}]
        for img in args.image:
            b64_url = encode_image(img)
            user_content.append({
                "type": "image_url",
                "image_url": {"url": b64_url}
            })
        messages.append({"role": "user", "content": user_content})
    else:
        messages.append({"role": "user", "content": prompt_text})

    print(f"[*] Enviando tarea a DeepSeek [{model}] (Tarea: {args.task}, Imágenes: {len(args.image)}, Archivos: {len(args.files)})...", file=sys.stderr)

    resp = call_deepseek(api_key, model, messages, max_tokens=args.max_tokens)

    choice = resp["choices"][0]
    msg = choice["message"]
    content = msg.get("content", "")
    reasoning = msg.get("reasoning_content", "")
    usage = resp.get("usage", {})

    print(f"[✓] Respuesta recibida. Tokens: Prompt={usage.get('prompt_tokens')}, Completion={usage.get('completion_tokens')}, Total={usage.get('total_tokens')}", file=sys.stderr)

    final_output = ""
    if args.show_reasoning and reasoning:
        final_output += f"### Razonamiento Interno (DeepSeek):\n{reasoning}\n\n---\n\n"
    final_output += content

    if args.out:
        out_path = Path(args.out)
        out_path.parent.mkdir(parents=True, exist_ok=True)
        out_path.write_text(final_output, encoding="utf-8")
        print(f"[✓] Resultado guardado en: {out_path.resolve()}", file=sys.stderr)
    else:
        print(final_output)

if __name__ == "__main__":
    main()
