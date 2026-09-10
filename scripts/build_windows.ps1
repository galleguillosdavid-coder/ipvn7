# Build IPv7 binaries for Windows (amd64) targeting bin/ directory
$ErrorActionPreference = "Stop"

$ProjectRoot = Split-Path -Parent $PSScriptRoot
if (-not $ProjectRoot) {
    $ProjectRoot = (Get-Location).Path
}

$BinDir = Join-Path $ProjectRoot "bin"
if (-not (Test-Path $BinDir)) {
    New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
}

Write-Host "=================================================" -ForegroundColor Cyan
Write-Host "  Compilando IPv7 para Windows (amd64)...        " -ForegroundColor Cyan
Write-Host "=================================================" -ForegroundColor Cyan

# 1. Compile Node binary
$NodeOut = Join-Path $BinDir "ipv7-node.exe"
Write-Host "[1/2] Compilando cmd/node -> $NodeOut" -ForegroundColor Yellow
Push-Location $ProjectRoot
try {
    go build -o $NodeOut ./cmd/node
    if ($LASTEXITCODE -ne 0) {
        throw "Fallo la compilacion de cmd/node"
    }
    Write-Host " [OK] ipv7-node.exe generado con exito." -ForegroundColor Green
}
finally {
    Pop-Location
}

# 2. Compile Chat binary
$ChatOut = Join-Path $BinDir "chat.exe"
Write-Host "[2/3] Compilando cmd/chat -> $ChatOut" -ForegroundColor Yellow
Push-Location $ProjectRoot
try {
    go build -ldflags="-s -w" -o $ChatOut ./cmd/chat
    if ($LASTEXITCODE -ne 0) {
        throw "Fallo la compilacion de cmd/chat"
    }
    Write-Host " [OK] chat.exe generado con exito." -ForegroundColor Green
}
finally {
    Pop-Location
}

# 3. Compile ipvn7-cli binary
$CliOut = Join-Path $BinDir "ipvn7-cli.exe"
Write-Host "[3/3] Compilando cmd/ipvn7-cli -> $CliOut" -ForegroundColor Yellow
Push-Location $ProjectRoot
try {
    go build -ldflags="-s -w" -o $CliOut ./cmd/ipvn7-cli
    if ($LASTEXITCODE -ne 0) {
        throw "Fallo la compilacion de cmd/ipvn7-cli"
    }
    Write-Host " [OK] ipvn7-cli.exe generado con exito." -ForegroundColor Green
}
finally {
    Pop-Location
}

Write-Host "=================================================" -ForegroundColor Green
Write-Host "  Binarios Windows listos en: $BinDir" -ForegroundColor Green
Write-Host "=================================================" -ForegroundColor Green
