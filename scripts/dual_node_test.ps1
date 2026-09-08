# Prueba de Malla Dual: Inicia Nodo A en Windows y Nodo B en WSL2 Linux interconectados
$ErrorActionPreference = "Stop"

$ProjectRoot = Split-Path -Parent $PSScriptRoot
if (-not $ProjectRoot) {
    $ProjectRoot = (Get-Location).Path
}

Write-Host "==================================================================" -ForegroundColor Cyan
Write-Host "       PRUEBA DE MALLA P2P CRUZADA: WINDOWS <-> WSL2 LINUX       " -ForegroundColor Cyan
Write-Host "==================================================================" -ForegroundColor Cyan

# 1. Check or build Windows binary
$WinNode = Join-Path $ProjectRoot "ipv7-node.exe"
if (-not (Test-Path $WinNode)) {
    Write-Host "[*] Compilando ipv7-node.exe para Windows..." -ForegroundColor Yellow
    Push-Location $ProjectRoot
    go build -o ipv7-node.exe ./cmd/node
    Pop-Location
}

# 2. Check or build Linux binary
$LinuxNode = Join-Path $ProjectRoot "bin\ipv7-node-linux"
if (-not (Test-Path $LinuxNode)) {
    Write-Host "[*] Compilando binarios Linux para WSL2..." -ForegroundColor Yellow
    & (Join-Path $PSScriptRoot "build_linux.ps1")
}

Write-Host "[1/2] Iniciando Nodo A en Windows (Puerto 7001, UI http://localhost:8080)..." -ForegroundColor Green
$WinProc = Start-Process -FilePath $WinNode -ArgumentList "-port", "7001", "-ui", "8080" -PassThru

Start-Sleep -Seconds 2

Write-Host "[2/2] Iniciando Nodo B en WSL2 Linux (Puerto 7002, UI http://localhost:8082, conectando a 127.0.0.1:7001)..." -ForegroundColor Green

$Drive = $ProjectRoot.Substring(0, 1).ToLower()
$SubPath = $ProjectRoot.Substring(2).Replace("\", "/")
$WslProjectRoot = "/mnt/$Drive$SubPath"
$WslScript = "$WslProjectRoot/scripts/wsl_node.sh"

wsl -d Ubuntu -e bash -c "sed -i 's/\r$//' '$WslScript' 2>/dev/null; chmod +x '$WslScript'"

Write-Host ""
Write-Host "Ambos nodos estan corriendo:" -ForegroundColor Cyan
Write-Host "  - NODO A (Windows):    http://localhost:8080" -ForegroundColor White
Write-Host "  - NODO B (WSL2 Linux): http://localhost:8082" -ForegroundColor White
Write-Host ""
Write-Host "Presiona Ctrl+C en esta ventana para detener el supervisor WSL2." -ForegroundColor Yellow
Write-Host "==================================================================" -ForegroundColor Cyan

try {
    # Run WSL node in foreground
    wsl -d Ubuntu -e bash -c "'$WslScript' -port 7002 -ui 8082 -peer 127.0.0.1:7001"
}
finally {
    Write-Host "`n[*] Deteniendo Nodo A de Windows..." -ForegroundColor Yellow
    if ($WinProc -and -not $WinProc.HasExited) {
        Stop-Process -Id $WinProc.Id -Force -ErrorAction SilentlyContinue
    }
    Write-Host "[OK] Nodos finalizados." -ForegroundColor Green
}
