# Script para habilitar OpenSSH Server en el Notebook
# Ejecutar como Administrador en Windows PowerShell

Write-Host "Habilitando OpenSSH Server en Windows..." -ForegroundColor Cyan

$capability = Get-WindowsCapability -Online | Where-Object Name -like 'OpenSSH.Server*'
if ($capability.State -ne 'Installed') {
    Add-WindowsCapability -Online -Name OpenSSH.Server~~~~0.0.1.0
}

Start-Service sshd
Set-Service -Name sshd -StartupType 'Automatic'

# Abrir puerto 22 en Firewall
New-NetFirewallRule -Name 'OpenSSH-Server-In-TCP' -DisplayName 'OpenSSH Server (sshd)' -Enabled True -Direction Inbound -Protocol TCP -Action Allow -LocalPort 22 -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "[OK] OpenSSH Server activado con exito!" -ForegroundColor Green
Write-Host "Ahora este notebook puede recibir comandos remotos desde la otra PC." -ForegroundColor Yellow
pause
