# Build IPv7 binaries for Linux (amd64) targeting WSL2 / Linux environments
$ErrorActionPreference = "Stop"

$ProjectRoot = Split-Path -Parent $PSScriptRoot
if (-not $ProjectRoot) {
    $ProjectRoot = (Get-Location).Path
}

$BinDir = Join-Path $ProjectRoot "bin"
if (-not (Test-Path $BinDir)) {
    New-Item -ItemType Directory -Path $BinDir | Out-Null
}

Write-Host "=================================================" -ForegroundColor Cyan
Write-Host "  Compilando IPv7 para Linux (amd64 / WSL2)...  " -ForegroundColor Cyan
Write-Host "=================================================" -ForegroundColor Cyan

$env:GOOS = "linux"
$env:GOARCH = "amd64"

# 1. Compile Node binary
$NodeOut = Join-Path $BinDir "ipv7-node-linux"
Write-Host "[1/2] Compilando cmd/node -> $NodeOut" -ForegroundColor Yellow
Push-Location $ProjectRoot
try {
    go build -ldflags="-s -w" -o $NodeOut ./cmd/node
    if ($LASTEXITCODE -ne 0) {
        throw "Fallo la compilacion de cmd/node"
    }
    Write-Host " [OK] ipv7-node-linux generado con exito." -ForegroundColor Green
}
finally {
    Pop-Location
}

# 2. Compile Chat binary
$ChatOut = Join-Path $BinDir "chat-linux"
Write-Host "[2/2] Compilando cmd/chat -> $ChatOut" -ForegroundColor Yellow
Push-Location $ProjectRoot
try {
    go build -ldflags="-s -w" -o $ChatOut ./cmd/chat
    if ($LASTEXITCODE -ne 0) {
        throw "Fallo la compilacion de cmd/chat"
    }
    Write-Host " [OK] chat-linux generado con exito." -ForegroundColor Green
}
finally {
    Pop-Location
}

# Reset env
$env:GOOS = ""
$env:GOARCH = ""

Write-Host "=================================================" -ForegroundColor Green
Write-Host "  Binarios Linux listos en: $BinDir" -ForegroundColor Green
Write-Host "=================================================" -ForegroundColor Green
