@echo off
:: ================================================================
::   IPv7 Node Launcher — DID Persistente + UPnP
::   Mantiene el mismo DID entre reinicios usando -key persistent
::   Mapea el puerto en el router automáticamente con -upnp=true
::   El navegador se abre automáticamente en http://localhost:8080
:: ================================================================
title IPv7 P2P Node

echo ==================================================================
echo              IPv7 NEXT-GEN NODE — DID PERSISTENTE
echo ==================================================================
echo  Tu identidad DID se conserva en: %USERPROFILE%\.ipv7\identity.key
echo  UPnP: mapeo automatico de puertos en router habilitado
echo  Dashboard web: http://localhost:8080
echo ==================================================================
echo.

:: Verificar que el ejecutable existe
if not exist "%~dp0ipv7-node.exe" (
    echo [ERROR] No se encontro ipv7-node.exe en %~dp0
    echo         Compila primero con: go build -o ipv7-node.exe .\cmd\node\
    pause
    exit /b 1
)

:: Lanzar nodo con identidad persistente + UPnP (segun Caso 9 de la documentacion)
"%~dp0ipv7-node.exe" -key persistent -upnp=true

echo.
echo [IPv7] Nodo detenido.
pause
