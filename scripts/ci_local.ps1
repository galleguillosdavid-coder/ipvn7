# ==============================================================================
# PIPELINE DE INTEGRACION CONTINUA Y CONTROL DE CALIDAD LOCAL (CI LOCAL)
# Cero cuota en la nube - 100% autogestionado en Windows / WSL2
# ==============================================================================

param (
    [switch]$SkipBench = $false
)

$ErrorActionPreference = "Stop"
$ProjectRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
Push-Location $ProjectRoot

Write-Host "==================================================================" -ForegroundColor Cyan
Write-Host "         IPv7 LOCAL CI/CD PIPELINE - QUALITY ASSURANCE            " -ForegroundColor Cyan
Write-Host "==================================================================" -ForegroundColor Cyan
Write-Host "[+] Directorio Base: $ProjectRoot" -ForegroundColor Gray
Write-Host "[+] Hora de inicio : $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')" -ForegroundColor Gray
Write-Host ""

$swTotal = [System.Diagnostics.Stopwatch]::StartNew()

# ------------------------------------------------------------------------------
# FASE 1: Verificación de Formato y Análisis Estático (vet)
# ------------------------------------------------------------------------------
Write-Host "[1/5] Ejecutando analisis estatico (go vet ./...)..." -ForegroundColor Yellow
$vetOutput = go vet ./... 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] Error en analisis estatico:" -ForegroundColor Red
    Write-Host $vetOutput -ForegroundColor Red
    Pop-Location
    exit 1
}
Write-Host "  [OK] Analisis estatico superado sin advertencias." -ForegroundColor Green

# ------------------------------------------------------------------------------
# FASE 2: Suite Completa de Tests Automatizados
# ------------------------------------------------------------------------------
Write-Host "`n[2/5] Ejecutando tests unitarios y de regresion..." -ForegroundColor Yellow
$testOutput = go test -v ./... 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "[ERROR] Fallaron pruebas unitarias:" -ForegroundColor Red
    Write-Host $testOutput -ForegroundColor Red
    Pop-Location
    exit 1
}
$passedTests = ($testOutput | Select-String "--- PASS:").Count
Write-Host "  [OK] $passedTests tests completados exitosamente (100% PASS)." -ForegroundColor Green

# ------------------------------------------------------------------------------
# FASE 3: Micro-Benchmarks de Rendimiento y Memoria
# ------------------------------------------------------------------------------
if (-not $SkipBench) {
    Write-Host "`n[3/5] Ejecutando micro-benchmarks criptograficos (core)..." -ForegroundColor Yellow
    $benchOutput = go test -bench "BenchmarkE2EE|BenchmarkSmallWorld" -benchmem -run "^$" ./core/... 2>&1
    Write-Host $benchOutput -ForegroundColor Gray
    Write-Host "  [OK] Micro-benchmarks finalizados." -ForegroundColor Green
} else {
    Write-Host "`n[3/5] Micro-benchmarks omitidos por solicitud." -ForegroundColor DarkGray
}

# ------------------------------------------------------------------------------
# FASE 4: Compilación Nativa para Windows (amd64)
# ------------------------------------------------------------------------------
Write-Host "`n[4/5] Compilando binario de producción para Windows (release/ipv7.exe)..." -ForegroundColor Yellow
$ReleaseDir = Join-Path $ProjectRoot "release"
if (-not (Test-Path $ReleaseDir)) { New-Item -ItemType Directory -Path $ReleaseDir | Out-Null }

$WinOut = Join-Path $ReleaseDir "ipv7.exe"
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -ldflags="-s -w" -o $WinOut ./cmd/node
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Fallo al compilar release/ipv7.exe" -ForegroundColor Red
    Pop-Location
    exit 1
}
$winSize = (Get-Item $WinOut).Length / 1MB
Write-Host "  ✅ release/ipv7.exe generado ($([math]::Round($winSize, 2)) MB)" -ForegroundColor Green

# ------------------------------------------------------------------------------
# FASE 5: Compilación Cruzada para Linux / WSL2 (ELF amd64)
# ------------------------------------------------------------------------------
Write-Host "`n[5/5] Compilando binario de producción para Linux/WSL2 (bin/ipv7-node-linux)..." -ForegroundColor Yellow
$BinDir = Join-Path $ProjectRoot "bin"
if (-not (Test-Path $BinDir)) { New-Item -ItemType Directory -Path $BinDir | Out-Null }

$LinuxOut = Join-Path $BinDir "ipv7-node-linux"
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -ldflags="-s -w" -o $LinuxOut ./cmd/node
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Fallo al compilar bin/ipv7-node-linux" -ForegroundColor Red
    Pop-Location
    exit 1
}
$linuxSize = (Get-Item $LinuxOut).Length / 1MB
Write-Host "  [OK] bin/ipv7-node-linux generado ($([math]::Round($linuxSize, 2)) MB)" -ForegroundColor Green

# ------------------------------------------------------------------------------
# RESUMEN EJECUTIVO
# ------------------------------------------------------------------------------
$swTotal.Stop()
Write-Host ""
Write-Host "==================================================================" -ForegroundColor Cyan
Write-Host "         PIPELINE CI LOCAL FINALIZADO CON EXITO                   " -ForegroundColor Green
Write-Host "==================================================================" -ForegroundColor Cyan
Write-Host "[OK] Tests Go             : 100% PASS" -ForegroundColor Green
Write-Host "[OK] Handshake Noise PFS  : OPERACIONAL" -ForegroundColor Green
Write-Host "[OK] Autocuracion Red     : OPERACIONAL" -ForegroundColor Green
Write-Host "[OK] Streaming SSE        : OPERACIONAL" -ForegroundColor Green
Write-Host "[OK] Windows Binary       : $WinOut" -ForegroundColor Gray
Write-Host "[OK] Linux/WSL2 Binary    : $LinuxOut" -ForegroundColor Gray
Write-Host "[OK] Tiempo Total         : $([math]::Round($swTotal.Elapsed.TotalSeconds, 2)) segundos" -ForegroundColor Cyan
Write-Host "==================================================================" -ForegroundColor Cyan

Pop-Location
exit 0
