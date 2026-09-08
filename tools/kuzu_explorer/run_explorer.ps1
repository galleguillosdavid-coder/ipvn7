# Launcher for standalone Kùzu Graph Studio
$ProjectRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
if (-not $ProjectRoot) { $ProjectRoot = (Get-Location).Path }

$ServerScript = Join-Path $PSScriptRoot "server.py"
Write-Host "Iniciando Explorador de Grafos Kùzu en http://localhost:8090 ..." -ForegroundColor Cyan
Start-Process "http://localhost:8090"
python $ServerScript
