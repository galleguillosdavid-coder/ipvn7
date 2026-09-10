@echo off
:: ================================================================
::   IPv7 Node Launcher — DID Persistente + UPnP (scripts/run_node.bat)
::   Mantiene el mismo DID entre reinicios usando -key persistent
::   Mapea el puerto en el router automáticamente con -upnp=true
::   El navegador se abre automáticamente en http://localhost:8080
:: ================================================================
title IPv7 P2P Network OS Node

set "SCRIPT_DIR=%~dp0"
set "ROOT_DIR=%SCRIPT_DIR%.."
set "BIN_EXE=%ROOT_DIR%\bin\ipv7-node.exe"

echo ==================================================================
echo              IPv7 NEXT-GEN NODE — DID PERSISTENTE
echo ==================================================================
echo  Identidad DID conservada en: %USERPROFILE%\.ipv7\identity.key
echo  UPnP: mapeo automatico de puertos habilitado
echo  Dashboard web: http://localhost:8080
echo  Directorio binario: %BIN_EXE%
echo ==================================================================
echo.

:: Verificar si el binario existe; si no, compilarlo limpiamente a bin/
if not exist "%BIN_EXE%" (
    echo [*] No se encontro ipv7-node.exe en bin/. Compilando ahora...
    if not exist "%ROOT_DIR%\bin" mkdir "%ROOT_DIR%\bin"
    pushd "%ROOT_DIR%"
    go build -o "%BIN_EXE%" .\cmd\node
    if errorlevel 1 (
        echo [ERROR] Fallo la compilacion automatica de cmd/node.
        popd
        pause
        exit /b 1
    )
    popd
    echo [OK] Binario compilado exitosamente.
    echo.
)

:: Lanzar nodo con identidad persistente + UPnP
"%BIN_EXE%" -key persistent -upnp=true

echo.
echo [IPv7] Nodo detenido.
pause
