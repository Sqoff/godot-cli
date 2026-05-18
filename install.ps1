$ErrorActionPreference = "Stop"
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$repo       = "Sqoff/godot-cli"
$installDir = "$env:LOCALAPPDATA\godot-cli"
$exe        = "$installDir\godot-cli.exe"
$pluginDir  = "$installDir\plugin"

# --- CLI binary ---
New-Item -ItemType Directory -Force -Path $installDir | Out-Null

$url = "https://github.com/$repo/releases/latest/download/godot-cli-windows-amd64.exe"
Write-Host "Downloading godot-cli for windows/amd64..."
Invoke-WebRequest -Uri $url -OutFile $exe -UseBasicParsing

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$installDir;$userPath", "User")
    $env:Path = "$installDir;$env:Path"
    Write-Host "Added $installDir to PATH (restart shell to apply)"
}

Write-Host "Installed godot-cli to $exe"

# --- Verify ---
Write-Host ""
& $exe --help

# --- Next steps ---
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Restart terminal (PATH update takes effect)"
Write-Host "  2. cd into your Godot project, then run:"
Write-Host "     godot-cli plugin install"
Write-Host "  3. In Godot: Project -> Project Settings -> Plugins -> Enable GodotCLI"
