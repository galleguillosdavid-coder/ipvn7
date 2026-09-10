# ==============================================================================
# DEEPSEEK TASK DELEGATOR & TOKEN SAVER (POWERSHELL WRAPPER)
# ==============================================================================
param(
    [Parameter(Position=0, Mandatory=$false)]
    [string]$Prompt,

    [Parameter(Mandatory=$false)]
    [string[]]$Files,

    [Parameter(Mandatory=$false)]
    [string[]]$Images,

    [Parameter(Mandatory=$false)]
    [string]$Task = "general",

    [Parameter(Mandatory=$false)]
    [string]$Model = "deepseek-flash",

    [Parameter(Mandatory=$false)]
    [string]$OutFile,

    [Parameter(Mandatory=$false)]
    [switch]$Reasoning
)

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$rootDir = Resolve-Path "$scriptDir\.."
$workerPath = "$rootDir\tools\deepseek\deepseek_worker.py"

$cmdArgs = @("$workerPath")

if ($Prompt) {
    $cmdArgs += @("-p", $Prompt)
}

if ($Files -and $Files.Count -gt 0) {
    $cmdArgs += "-f"
    foreach ($f in $Files) {
        $cmdArgs += (Resolve-Path $f).Path
    }
}

if ($Images -and $Images.Count -gt 0) {
    foreach ($img in $Images) {
        $cmdArgs += @("-i", (Resolve-Path $img).Path)
    }
}

if ($Task) {
    $cmdArgs += @("-t", $Task)
}

if ($Model) {
    $cmdArgs += @("-m", $Model)
}

if ($OutFile) {
    $cmdArgs += @("-o", $OutFile)
}

if ($Reasoning) {
    $cmdArgs += "--show-reasoning"
}

python @cmdArgs
