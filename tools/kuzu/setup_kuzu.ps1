# Setup script to download Kùzu CLI for Windows and WSL2 Linux
$ErrorActionPreference = "Stop"

$toolsDir = Split-Path -Parent $MyInvocation.MyCommand.Path
if (-not $toolsDir) {
    $toolsDir = "c:\Users\Frondabrick\Desktop\dvd\Ipv7\tools\kuzu"
}

if (-not (Test-Path $toolsDir)) {
    New-Item -ItemType Directory -Path $toolsDir -Force | Out-Null
}

$winUrl = "https://github.com/kuzudb/kuzu/releases/download/v0.11.3/kuzu_cli-windows-x86_64.zip"
$winZip = Join-Path $toolsDir "kuzu_cli-windows-x86_64.zip"
$winExe = Join-Path $toolsDir "kuzu.exe"

if (-not (Test-Path $winExe)) {
    Write-Host "Descargando Kùzu CLI para Windows..." -ForegroundColor Cyan
    Invoke-WebRequest -Uri $winUrl -OutFile $winZip -UseBasicParsing
    Write-Host "Extrayendo Kùzu CLI Windows..." -ForegroundColor Cyan
    Expand-Archive -Path $winZip -DestinationPath $toolsDir -Force
    Remove-Item $winZip -Force -ErrorAction SilentlyContinue
}
Write-Host " [OK] Kùzu CLI Windows listo en: $winExe" -ForegroundColor Green

# Linux CLI for WSL2
$linUrl = "https://github.com/kuzudb/kuzu/releases/download/v0.11.3/kuzu_cli-linux-x86_64.tar.gz"
$linTar = Join-Path $toolsDir "kuzu_cli-linux-x86_64.tar.gz"
$linExe = Join-Path $toolsDir "kuzu"

if (-not (Test-Path $linExe)) {
    Write-Host "Descargando Kùzu CLI para Linux (WSL2)..." -ForegroundColor Cyan
    Invoke-WebRequest -Uri $linUrl -OutFile $linTar -UseBasicParsing
    Write-Host "Extrayendo Kùzu CLI Linux..." -ForegroundColor Cyan
    $Drive = $toolsDir.Substring(0, 1).ToLower()
    $SubPath = $toolsDir.Substring(2).Replace("\", "/")
    $WslToolsDir = "/mnt/$Drive$SubPath"
    wsl -d Ubuntu -e bash -c "tar -xzf '$WslToolsDir/kuzu_cli-linux-x86_64.tar.gz' -C '$WslToolsDir/' && chmod +x '$WslToolsDir/kuzu'"
    Remove-Item $linTar -Force -ErrorAction SilentlyContinue
}
Write-Host " [OK] Kùzu CLI Linux listo en: $linExe" -ForegroundColor Green
