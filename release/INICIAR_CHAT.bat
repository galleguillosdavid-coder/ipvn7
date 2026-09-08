@echo off
title IPv7 - Chat P2P Consola
chcp 65001 > nul
cls
echo ==================================================================
echo                  IPv7 TERMINAL CHAT P2P                          
echo ==================================================================
echo.
set /p PEER="Ingresa la IP:PUERTO del otro nodo (o presiona ENTER para iniciar solo): "
echo.

if "%PEER%"=="" (
    chat.exe
) else (
    chat.exe -peer %PEER%
)
pause
