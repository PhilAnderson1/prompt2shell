$ErrorActionPreference = "Stop"

$principal = New-Object Security.Principal.WindowsPrincipal(
    [Security.Principal.WindowsIdentity]::GetCurrent()
)
if (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    Write-Error "Run this script from PowerShell as Administrator."
    exit 1
}

$sourceDirectory = $PSScriptRoot
$binarySource = Join-Path $sourceDirectory "p2s-windows-amd64.exe"
$configSource = Join-Path $sourceDirectory "prompt2shell.conf"
$installDirectory = Join-Path $env:ProgramFiles "prompt2shell"
$binaryDestination = Join-Path $installDirectory "p2s.exe"
$configDirectory = Join-Path $env:ProgramData "prompt2shell"
$configDestination = Join-Path $configDirectory "prompt2shell.conf"

if (-not (Test-Path -LiteralPath $binarySource -PathType Leaf)) {
    Write-Error "Missing $binarySource. Place the installer beside p2s-windows-amd64.exe."
    exit 1
}

New-Item -ItemType Directory -Force -Path $installDirectory | Out-Null
Copy-Item -LiteralPath $binarySource -Destination $binaryDestination -Force

if (-not (Test-Path -LiteralPath $configDestination -PathType Leaf)) {
    if (-not (Test-Path -LiteralPath $configSource -PathType Leaf)) {
        Write-Error "Missing $configSource. Place the example configuration beside this installer."
        exit 1
    }
    New-Item -ItemType Directory -Force -Path $configDirectory | Out-Null
    Copy-Item -LiteralPath $configSource -Destination $configDestination
    Write-Host "Installed configuration: $configDestination"
    Write-Host "Edit it and replace the placeholder API key before running p2s."
} else {
    Write-Host "Kept existing configuration: $configDestination"
}

$machinePath = [Environment]::GetEnvironmentVariable("Path", "Machine")
$pathEntries = @($machinePath -split ";" | Where-Object { $_ })
$alreadyPresent = $pathEntries | Where-Object {
    $_.TrimEnd("\") -ieq $installDirectory.TrimEnd("\")
}
if (-not $alreadyPresent) {
    $newPath = ($pathEntries + $installDirectory) -join ";"
    [Environment]::SetEnvironmentVariable("Path", $newPath, "Machine")
    Write-Host "Added to system PATH: $installDirectory"
} else {
    Write-Host "Already on system PATH: $installDirectory"
}

Write-Host "Installed executable: $binaryDestination"
Write-Host "Open a new PowerShell window, then run: p2s print the current directory"
