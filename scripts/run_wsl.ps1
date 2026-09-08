param (
    [int]$Port = 7002,
    [int]$UIPort = 8082,
    [string]$Peer = "",
    [string]$Stun = "stun.l.google.com:19302",
    [string]$Distro = "Ubuntu",
    [switch]$Rebuild
)

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent $PSScriptRoot
if (-not $ProjectRoot) {
    $ProjectRoot = (Get-Location).Path
}

$BinPath = Join-Path $ProjectRoot "bin\ipv7-node-linux"

if ($Rebuild -or (-not (Test-Path $BinPath))) {
    Write-Host "[*] Compilando binarios Linux..." -ForegroundColor Cyan
    & (Join-Path $PSScriptRoot "build_linux.ps1")
}

# Convert Windows path to WSL path
$Drive = $ProjectRoot.Substring(0, 1).ToLower()
$SubPath = $ProjectRoot.Substring(2).Replace("\", "/")
$WslProjectRoot = "/mnt/$Drive$SubPath"
$WslScript = "$WslProjectRoot/scripts/wsl_node.sh"

$ArgsList = @("-port", $Port, "-ui", $UIPort, "-stun", $Stun)
if ($Peer -ne "") {
    $ArgsList += @("-peer", $Peer)
}
$ArgsString = $ArgsList -join " "

Write-Host "==================================================================" -ForegroundColor Green
Write-Host "  Supervisando nodo IPv7 en WSL2 ($Distro)" -ForegroundColor Green
Write-Host "  Puerto P2P (UDP/QUIC): $Port / $($Port + 1)" -ForegroundColor Green
Write-Host "  Dashboard Web:         http://localhost:$UIPort" -ForegroundColor Green
Write-Host "==================================================================" -ForegroundColor Green

# Ensure wsl_node.sh has unix line endings
wsl -d $Distro -e bash -c "sed -i 's/\r$//' '$WslScript' 2>/dev/null; chmod +x '$WslScript'"

# Execute in WSL2
wsl -d $Distro -e bash -c "'$WslScript' $ArgsString"
