#!/usr/bin/env bash
# ==============================================================================
# DEEPSEEK TASK DELEGATOR & TOKEN SAVER (POSIX BASH WRAPPER)
# ==============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
WORKER_PATH="$ROOT_DIR/tools/deepseek/deepseek_worker.py"

PYTHON_BIN="python3"
if ! command -v python3 &> /dev/null; then
    if command -v python &> /dev/null; then
        PYTHON_BIN="python"
    fi
fi

exec "$PYTHON_BIN" "$WORKER_PATH" "$@"
