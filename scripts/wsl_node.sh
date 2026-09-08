#!/usr/bin/env bash
# Runner y supervisor del nodo IPv7 dentro de WSL2 (Ubuntu / Linux)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
BIN_PATH="${PROJECT_ROOT}/bin/ipv7-node-linux"
LOG_DIR="${PROJECT_ROOT}/logs"
LOG_FILE="${LOG_DIR}/wsl_node.log"

mkdir -p "${LOG_DIR}"

if [ ! -f "${BIN_PATH}" ]; then
    echo "[!] Binario no encontrado en: ${BIN_PATH}"
    echo "[!] Compilalo primero ejecutando build_linux.ps1 o go build en el directorio raíz."
    exit 1
fi

chmod +x "${BIN_PATH}"

echo "=================================================================="
echo "    Iniciando IPv7 Node en WSL2 (Linux: $(uname -s) $(uname -m))  "
echo "=================================================================="
echo "[+] Directorio: ${PROJECT_ROOT}"
echo "[+] Binario:    ${BIN_PATH}"
echo "[+] Log file:   ${LOG_FILE}"
echo "[+] Argumentos: $@"
echo "=================================================================="

# Ejecutar y duplicar la salida tanto en terminal como en log con marca de tiempo
"${BIN_PATH}" "$@" 2>&1 | tee -a "${LOG_FILE}"
