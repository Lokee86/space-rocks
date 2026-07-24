[CmdletBinding()]
param(
    [string]$InstallDir = (Join-Path $env:LOCALAPPDATA "Programs\Space Rocks"),
    [switch]$NoStartMenuShortcut,
    [switch]$DesktopShortcut
)

$ErrorActionPreference = "Stop"

$sourceDir = [IO.Path]::GetFullPath((Split-Path -Parent $MyInvocation.MyCommand.Path)).TrimEnd('\', '/')
$installDir = [IO.Path]::GetFullPath($InstallDir).TrimEnd('\', '/')
$separator = [IO.Path]::DirectorySeparatorChar

foreach ($required in @("SpaceRocks.exe", "SpaceRocks.pck", "space-rocks-server.exe", "space-rocks-credential-helper.exe")) {
    if (-not (Test-Path -LiteralPath (Join-Path $sourceDir $required) -PathType Leaf)) {
        throw "$required was not found beside install.ps1. Run the installer from an extracted Space Rocks release package."
    }
}

if (
    $installDir.StartsWith($sourceDir + $separator, [StringComparison]::OrdinalIgnoreCase) -or
    $sourceDir.StartsWith($installDir + $separator, [StringComparison]::OrdinalIgnoreCase)
) {
    throw "The installation directory and extracted release package cannot contain one another."
}

if (-not [StringComparer]::OrdinalIgnoreCase.Equals($sourceDir, $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    Get-ChildItem -LiteralPath $sourceDir -Force | Copy-Item -Destination $installDir -Recurse -Force
}

$targetExecutable = Join-Path $installDir "SpaceRocks.exe"

function New-SpaceRocksShortcut([string]$ShortcutPath) {
    $parent = Split-Path -Parent $ShortcutPath
    New-Item -ItemType Directory -Path $parent -Force | Out-Null
    $shell = New-Object -ComObject WScript.Shell
    $shortcut = $shell.CreateShortcut($ShortcutPath)
    $shortcut.TargetPath = $targetExecutable
    $shortcut.WorkingDirectory = $installDir
    $shortcut.IconLocation = "$targetExecutable,0"
    $shortcut.Save()
}

if (-not $NoStartMenuShortcut) {
    $startMenu = Join-Path $env:APPDATA "Microsoft\Windows\Start Menu\Programs\Space Rocks.lnk"
    New-SpaceRocksShortcut $startMenu
}

if ($DesktopShortcut) {
    New-SpaceRocksShortcut (Join-Path ([Environment]::GetFolderPath("Desktop")) "Space Rocks.lnk")
}

Write-Host "Space Rocks installed to $installDir"
if (-not $NoStartMenuShortcut) {
    Write-Host "A Start Menu shortcut was created."
}
if ($DesktopShortcut) {
    Write-Host "A desktop shortcut was created."
}
