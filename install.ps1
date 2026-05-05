# ClawSec Installer for Windows
# Usage: iwr -useb https://raw.githubusercontent.com/clawsec/clawsec/main/install.ps1 | iex
# Or: .\install.ps1 [-Version v0.1.0] [-InstallDir "$env:LOCALAPPDATA\ClawSec"]

param(
    [string]$Version = "latest",
    [string]$InstallDir = "$env:LOCALAPPDATA\ClawSec",
    [switch]$NoPath,
    [switch]$Force
)

$ErrorActionPreference = "Stop"
$AppName = "clawsec"
$Repo = "clawsec/clawsec"

function Write-Info($msg) { Write-Host "[INFO] $msg" -ForegroundColor Cyan }
function Write-Ok($msg) { Write-Host "[OK]   $msg" -ForegroundColor Green }
function Write-Warn($msg) { Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Write-Err($msg) { Write-Host "[ERR]  $msg" -ForegroundColor Red }

function Get-LatestVersion {
    try {
        $resp = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -TimeoutSec 10
        return $resp.tag_name
    } catch {
        return $null
    }
}

function Download-Binary($version, $outPath) {
    $arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
    $asset = "${AppName}_${version}_windows_${arch}.zip"
    $url = "https://github.com/$Repo/releases/download/$version/$asset"

    Write-Info "Downloading $asset ..."
    $tmp = [System.IO.Path]::GetTempFileName() + ".zip"
    try {
        Invoke-WebRequest -Uri $url -OutFile $tmp -TimeoutSec 120
        Expand-Archive -Path $tmp -DestinationPath $outPath -Force
        Write-Ok "Extracted to $outPath"
    } finally {
        if (Test-Path $tmp) { Remove-Item $tmp -Force }
    }
}

function Install-FromSource($outPath) {
    Write-Info "Building from source ..."
    $go = Get-Command go -ErrorAction SilentlyContinue
    if (-not $go) {
        throw "Go is not installed. Please install Go 1.22+ from https://go.dev/dl/"
    }
    $ver = (& go version) -replace '.*go(\d+\.\d+).*','$1'
    Write-Info "Go version: $ver"

    if (-not (Test-Path "$PSScriptRoot\cmd\clawsec")) {
        throw "Source code not found. Run this script from the clawsec repository root."
    }

    $cwd = if ($PSScriptRoot) { $PSScriptRoot } else { Get-Location }
    Push-Location $cwd
    try {
        go build -ldflags "-s -w" -o "$outPath\clawsec.exe" .\cmd\clawsec
        if ($LASTEXITCODE -ne 0) { throw "Build failed" }
    } finally {
        Pop-Location
    }
    Write-Ok "Built successfully"
}

function Add-ToUserPath($dir) {
    $current = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($current -notlike "*$dir*") {
        [Environment]::SetEnvironmentVariable("Path", "$current;$dir", "User")
        Write-Ok "Added $dir to user PATH (restart terminal to apply)"
    } else {
        Write-Info "$dir already in PATH"
    }
}

function Initialize-Config($configDir) {
    $cfgFile = Join-Path $configDir "config.yaml"
    if (-not (Test-Path $cfgFile)) {
        $defaultCfg = @"
# ClawSec Configuration File
# https://github.com/clawsec/clawsec

output_format: text
timeout: 5
threads: 50
rate_limit: 150

# AI settings
ai:
  enabled: false
  endpoint: ""
  model: "claude-sonnet-4-20250514"
  api_key: ""

# Product configurations (uncomment and configure as needed)
# safeline:
#   url: "https://safeline.example.com"
#   api_key: "your-api-key"
# xray:
#   url: "https://xray.example.com"
#   api_key: "your-api-key"
"@
        New-Item -ItemType Directory -Force -Path $configDir | Out-Null
        Set-Content -Path $cfgFile -Value $defaultCfg -Encoding UTF8
        Write-Ok "Created default config: $cfgFile"
    }
}

# ============ Main ============

Write-Host ""
Write-Host "=============================================" -ForegroundColor Green
Write-Host "  ClawSec Installer" -ForegroundColor Green
Write-Host "  AI-Native Offensive Security Platform" -ForegroundColor Green
Write-Host "=============================================" -ForegroundColor Green
Write-Host ""

# Check existing installation
$exePath = Join-Path $InstallDir "clawsec.exe"
if ((Test-Path $exePath) -and (-not $Force)) {
    $existing = & $exePath version 2>$null
    Write-Warn "Already installed: $existing"
    $resp = Read-Host "Reinstall? [y/N]"
    if ($resp -notmatch '^[Yy]') {
        Write-Info "Installation cancelled."
        exit 0
    }
}

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

# Determine version
if ($Version -eq "latest") {
    $Version = Get-LatestVersion
    if (-not $Version) {
        Write-Warn "Could not fetch latest version from GitHub"
        $Version = "v0.1.0-alpha"
    }
}
Write-Info "Installing version: $Version"

# Install
$sourceMode = $false
try {
    Download-Binary $Version $InstallDir
} catch {
    Write-Warn "Download failed ($($_.Exception.Message))"
    Write-Info "Falling back to source build..."
    Install-FromSource $InstallDir
    $sourceMode = $true
}

# Verify
if (Test-Path $exePath) {
    $verOut = & $exePath version 2>$null
    Write-Ok "Installed: $verOut"
} else {
    Write-Err "Installation failed - clawsec.exe not found in $InstallDir"
    exit 1
}

# PATH
if (-not $NoPath) {
    Add-ToUserPath $InstallDir
}

# Config
$configDir = Join-Path $env:USERPROFILE ".clawsec"
Initialize-Config $configDir

Write-Host ""
Write-Ok "Installation complete!"
Write-Info "Binary location: $exePath"
Write-Info "Config directory: $configDir"
if (-not $NoPath) {
    Write-Info "Restart your terminal or run: `$env:Path = [Environment]::GetEnvironmentVariable('Path', 'User')"
}
Write-Host ""
Write-Host "Quick start:" -ForegroundColor Green
Write-Host "  clawsec scan port -t 127.0.0.1 -p top100"
Write-Host "  clawsec crawl dir -t http://target.com --ext"
Write-Host "  clawsec poc run -u http://target.com --severity critical,high"
Write-Host ""
