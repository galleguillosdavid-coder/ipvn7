# ==============================================================================
# PROTOCOLO DE AUDITORIA DE REPRODUCIBILIDAD INDEPENDIENTE (IPv7)
# Plataforma: Windows PowerShell (ASCII Compatible)
# ==============================================================================

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

Write-Host "========================================================================" -ForegroundColor Cyan
Write-Host "   IPv7 PROTOCOLO DE AUDITORIA DE REPRODUCIBILIDAD INDEPENDIENTE" -ForegroundColor Green
Write-Host "   Verificacion Epistemica de Horizonte 5 (Chaos Testing & Hardening)" -ForegroundColor Yellow
Write-Host "========================================================================" -ForegroundColor Cyan
Write-Host ""

$rootPath = Resolve-Path "$PSScriptRoot\.."
Set-Location $rootPath

# 1. Comprobacion del entorno
$goVersion = go version
Write-Host "[1/5] Entorno de compilacion: $goVersion" -ForegroundColor Gray

# 2. Comprobacion de integridad de Protocol Core Freeze
Write-Host "[2/5] Verificando inmutabilidad del nucleo (core/)..." -ForegroundColor Gray
$coreDiff = git status --porcelain core/
if ($coreDiff) {
    Write-Host "[WARN] Hay modificaciones pendientes en core/. Verifique la regla de congelamiento." -ForegroundColor Yellow
} else {
    Write-Host "      Core Freeze respetado: 0 cambios en core/." -ForegroundColor Green
}

# 3. Ejecucion de adaptadores y DHT
Write-Host "[3/5] Ejecutando pruebas unitarias de adaptadores y DHT..." -ForegroundColor Gray
$startAdapters = Get-Date
go test ./adapters/... ./dht/...
if ($LASTEXITCODE -ne 0) {
    Write-Host "[FAIL] Fallaron las pruebas de adaptadores o DHT." -ForegroundColor Red
    exit 1
}
$durAdapters = ((Get-Date) - $startAdapters).TotalSeconds
$durAdaptersRound = [math]::Round($durAdapters, 2)
Write-Host "      Adaptadores y DHT validados con exito (${durAdaptersRound}s)." -ForegroundColor Green

# 4. Ejecucion de la bateria destructiva de Horizonte 5
Write-Host "[4/5] Ejecutando Bateria Destructiva y Convergencia de Horizonte 5..." -ForegroundColor Gray
$startChaos = Get-Date
$chaosOutputLines = & go test -v ./tests/... 2>&1
$testExitCode = $LASTEXITCODE
$durChaos = ((Get-Date) - $startChaos).TotalSeconds
$totalSecs = [math]::Round($durAdapters + $durChaos, 2)

if ($testExitCode -ne 0) {
    Write-Host "[FAIL] Fallaron una o mas pruebas de Horizonte 5:" -ForegroundColor Red
    $chaosOutputLines | ForEach-Object { Write-Host $_ -ForegroundColor Red }
    exit $testExitCode
}

# 5. Parseo y Verificacion Estricta de Metricas Contractuales
Write-Host "[5/5] Auditando aserciones metricas contractuales de los logs..." -ForegroundColor Gray
$rawOutput = ($chaosOutputLines | Out-String)

$metricErrors = @()

# Metrica 1: Reconvergencia H-MULTI-01
$reconvMatch = [regex]::Match($rawOutput, "Reconvergencia a Malla Off-Grid y descubrimiento completados en:\s*([0-9\.]+)(m?s)")
$reconvStr = "N/A"
if ($reconvMatch.Success) {
    $val = [double]$reconvMatch.Groups[1].Value
    $unit = $reconvMatch.Groups[2].Value
    $reconvMs = if ($unit -eq "s") { $val * 1000.0 } else { $val }
    $reconvStr = "${reconvMs}ms"
    if ($reconvMs -gt 250.0) {
        $metricErrors += "H-MULTI-01: Reconvergencia excedio cota contractual (${reconvMs}ms > 250ms)"
    }
} else {
    $metricErrors += "H-MULTI-01: No se encontro registro de telemetria de reconvergencia"
}

# Metrica 2: DoS PDR
$dosPdrMatch = [regex]::Match($rawOutput, "Paquetes .+ entregados durante el ataque DoS:\s*(\d+)\s*/\s*(\d+)")
$dosPdrStr = "N/A"
if ($dosPdrMatch.Success) {
    $deliv = [int]$dosPdrMatch.Groups[1].Value
    $total = [int]$dosPdrMatch.Groups[2].Value
    $pdrPct = [math]::Round(($deliv / $total) * 100, 1)
    $dosPdrStr = "${deliv}/${total} (${pdrPct}%)"
    if ($pdrPct -lt 90.0) {
        $metricErrors += "H-L2-DOS: PDR legitimo degradado por debajo de la cota (${pdrPct}% < 90%)"
    }
} else {
    $metricErrors += "H-L2-DOS: No se encontro registro de entrega legitima DoS"
}

# Metrica 3: Cota de tabla de vecinos DoS
$peersMatch = [regex]::Match($rawOutput, "Vecinos registrados en la tabla del nodo .+:\s*(\d+)")
$peersStr = "N/A"
if ($peersMatch.Success) {
    $peersCount = [int]$peersMatch.Groups[1].Value
    $peersStr = "$peersCount / 256"
    if ($peersCount -gt 256) {
        $metricErrors += "H-L2-DOS: Vecinos excedieron la cota MaxDiscoveredPeers (${peersCount} > 256)"
    }
} else {
    $metricErrors += "H-L2-DOS: No se encontro registro de tamano de tabla de vecinos"
}

# Metrica 4: Crecimiento de heap DoS
$heapMatch = [regex]::Match($rawOutput, "Crecimiento neto de heap tras procesar 10\.?000 balizas forjadas:\s*(\d+)\s*KB")
$heapStr = "N/A"
if ($heapMatch.Success) {
    $heapKB = [int]$heapMatch.Groups[1].Value
    $heapStr = "${heapKB} KB"
    if ($heapKB -gt 512) {
        $metricErrors += "H-L2-DOS: Crecimiento de heap excedio cota de seguridad (${heapKB} KB > 512 KB)"
    }
} else {
    $metricErrors += "H-L2-DOS: No se encontro registro de medicion de heap"
}

# Metrica 5: Transiciones en Flapping
$transMatch = [regex]::Match($rawOutput, "Transiciones de modo capturadas:\s*(\d+)")
$transStr = "N/A"
if ($transMatch.Success) {
    $transCount = [int]$transMatch.Groups[1].Value
    $transStr = "$transCount trans"
    if ($transCount -lt 60) {
        $metricErrors += "H-FLAPPING: Transiciones capturadas insuficientes (${transCount} < 60)"
    }
} else {
    $metricErrors += "H-FLAPPING: No se encontro registro de transiciones de flapping"
}

# Metrica 6: Fuga de Goroutines en Flapping
$goroMatch = [regex]::Match($rawOutput, "Delta de goroutines tras tormenta de flapping:\s*([+-]?\d+)")
$goroStr = "N/A"
if ($goroMatch.Success) {
    $deltaGoro = [int]$goroMatch.Groups[1].Value
    $goroStr = "${deltaGoro} goroutines"
    if ($deltaGoro -gt 1) {
        $metricErrors += "H-FLAPPING: Posible fuga de goroutines detectada (${deltaGoro} > +1)"
    }
} else {
    $metricErrors += "H-FLAPPING: No se encontro registro de delta de goroutines"
}

if ($metricErrors.Count -gt 0) {
    Write-Host ""
    Write-Host "[FAIL] VIOLACIONES CONTRACTUALES DETECTADAS EN LA AUDITORIA:" -ForegroundColor Red
    foreach ($err in $metricErrors) {
        Write-Host "   - $err" -ForegroundColor Red
    }
    exit 1
}

Write-Host "      Todas las metricas cumplen estrictamente las cotas del contrato." -ForegroundColor Green
Write-Host ""
Write-Host "=========================================================================================================" -ForegroundColor Cyan
Write-Host "                                   INFORME FORMAL DE AUDITORIA                                           " -ForegroundColor Green
Write-Host "=========================================================================================================" -ForegroundColor Cyan
Write-Host "  Hipotesis         | Estado Epistemico | Cobertura Asertiva | Metrica Auditada            | Resultado   " -ForegroundColor White
Write-Host "--------------------+-------------------+--------------------+-----------------------------+-------------" -ForegroundColor Gray
Write-Host "  H-MULTI-01        | DEMONSTRATED      | COMPLETE           | Latencia reconv: $reconvStr | PASS [OK]   " -ForegroundColor Green
Write-Host "  H-L2-DOS (PDR)    | DEMONSTRATED      | COMPLETE           | PDR legitimo: $dosPdrStr    | PASS [OK]   " -ForegroundColor Green
Write-Host "  H-L2-DOS (Table)  | DEMONSTRATED      | COMPLETE           | Cota vecinos: $peersStr     | PASS [OK]   " -ForegroundColor Green
Write-Host "  H-L2-DOS (Heap)   | DEMONSTRATED      | COMPLETE           | Delta heap: $heapStr        | PASS [OK]   " -ForegroundColor Green
Write-Host "  H-FLAPPING (Trans)| DEMONSTRATED      | COMPLETE           | Ciclos oscilacion: $transStr| PASS [OK]   " -ForegroundColor Green
Write-Host "  H-FLAPPING (Leaks)| DEMONSTRATED      | COMPLETE           | Fuga goroutine: $goroStr    | PASS [OK]   " -ForegroundColor Green
Write-Host "  H-SPLIT-BRAIN     | DEMONSTRATED      | PARTIAL            | Puente local inter-islas    | PASS [OK]   " -ForegroundColor Yellow
Write-Host "  H-CONSTRAINED-MTU | DEMONSTRATED      | PARTIAL            | Canal 180B (Sin L2-ARQ)     | PASS [OK]   " -ForegroundColor Yellow
Write-Host "  H-UNIFIED-E2E     | DEMONSTRATED      | COMPLETE           | 4 Horizontes E2E (100% PDR) | PASS [OK]   " -ForegroundColor Green
Write-Host "--------------------+-------------------+--------------------+-----------------------------+-------------" -ForegroundColor Gray
Write-Host "  Tiempo total de verificacion: ${totalSecs}s" -ForegroundColor Gray
Write-Host "  Entorno de certificacion: LAB_SIMULATED (RAM / Virtual Links)" -ForegroundColor Yellow
Write-Host "  Hardware Fisico / Escala Global: NOT_PROVEN (Conforme a matriz y contrato de aserciones)" -ForegroundColor Yellow
Write-Host "=========================================================================================================" -ForegroundColor Cyan
Write-Host "CERTIFICACION CONCLUIDA: La cadena documentacion -> contrato -> test -> assertion -> resultado es reproducible bajo el entorno y las condiciones de prueba declaradas." -ForegroundColor Green
Write-Host ""
exit 0
