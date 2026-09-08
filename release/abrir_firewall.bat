@echo off
title Configurar Firewall de Windows para IPv7
chcp 65001 > nul
cls
echo ==================================================================
echo       CONFIGURACION AUTOMATICA DE FIREWALL PARA IPv7             
echo ==================================================================
echo.
echo Esta herramienta abre los puertos UDP y TCP de IPv7 para
echo permitir que otras PCs en tu red o por Internet se conecten.
echo.
echo Requiere permisos de Administrador.
echo.

netsh advfirewall firewall add rule name="IPv7 P2P Mesh (UDP)" dir=in action=allow protocol=UDP localport=7000-7050 > nul 2>&1
netsh advfirewall firewall add rule name="IPv7 P2P Mesh (UDP Out)" dir=out action=allow protocol=UDP localport=7000-7050 > nul 2>&1
netsh advfirewall firewall add rule name="IPv7 Web Dashboard (TCP)" dir=in action=allow protocol=TCP localport=8080-8095 > nul 2>&1

if %errorlevel% equ 0 (
    echo [OK] Reglas de firewall aplicadas exitosamente.
) else (
    echo [AVISO] Si ves este mensaje, haz clic derecho sobre este archivo
    echo         y selecciona "Ejecutar como administrador".
)

echo.
pause
