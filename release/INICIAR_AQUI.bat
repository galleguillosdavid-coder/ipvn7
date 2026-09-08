@echo off
title IPv7 - Nodo Descentralizado P2P
chcp 65001 > nul
cls
echo ==================================================================
echo                  INICIANDO NODO IPv7 P2P                          
echo ==================================================================
echo.
echo [1] Verificando entorno Windows...
echo [2] Auto-descubrimiento en red local (Wi-Fi/LAN) activado.
echo [3] Iniciando dashboard web y abriendo navegador...
echo.
echo Presiona Ctrl+C en esta ventana para detener el nodo.
echo ==================================================================
echo.

if exist ipv7.exe (
    ipv7.exe -open=true
) else (
    if exist release\ipv7.exe (
        cd release
        ipv7.exe -open=true
    ) else (
        echo [ERROR] No se encontro ipv7.exe.
        pause
    )
)
pause
