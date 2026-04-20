# AI Skills Global Installer for PowerShell

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProfileFile = $PROFILE

Write-Host ">>> AI Skills Global Installer (PowerShell)" -ForegroundColor Cyan
Write-Host "===========================================" -ForegroundColor Cyan
Write-Host ""

if (!(Test-Path $ProfileFile)) {
    $ProfileDir = Split-Path -Parent $ProfileFile
    if (!(Test-Path $ProfileDir)) {
        New-Item -Path $ProfileDir -ItemType Directory -Force | Out-Null
    }
    Write-Host "Creating PowerShell profile..." -ForegroundColor Yellow
    New-Item -Path $ProfileFile -ItemType File -Force | Out-Null
}

Write-Host "Profile file: $ProfileFile"
Write-Host ""

$Aliases = @(
    @{ Name = "skills-sync"; Target = "$ScriptDir\skill-sync\assets\sync.sh"; Cmd = "bash" },
    @{ Name = "skills-list"; Target = "$ScriptDir\list-skills.ps1"; Cmd = "powershell" },
    @{ Name = "skills-setup"; Target = "$ScriptDir\setup.ps1"; Cmd = "powershell" },
    @{ Name = "skills-sync-platforms"; Target = "$ScriptDir\sync-skills-all-platforms.sh"; Cmd = "bash" }
)

$ProfileContent = ""
if (Test-Path $ProfileFile) {
    $ProfileContent = Get-Content -Path $ProfileFile -Raw
}
$Changed = $false

foreach ($Alias in $Aliases) {
    $Name = $Alias.Name
    $Target = $Alias.Target
    
    if ($Alias.Cmd -eq "powershell") {
        $FunctionDef = "function $Name { & '$Target' @args }"
    }
    else {
        $FunctionDef = "function $Name { bash '$Target' `$args }"
    }

    $RegexSafeName = [regex]::Escape($Name)
    if ($ProfileContent -notmatch "function $RegexSafeName\s*\{") {
        Write-Host "Adding function $Name..."
        Add-Content -Path $ProfileFile -Value "`n$FunctionDef"
        $Changed = $true
    }
    else {
        Write-Host "Function $Name already exists. Consider updating manually if paths changed."
    }
}

Write-Host ""
if ($Changed) {
    Write-Host "[+] Installation complete!" -ForegroundColor Green
    Write-Host "Please restart PowerShell or run: . `$PROFILE" -ForegroundColor Cyan
}
else {
    Write-Host "No changes were needed." -ForegroundColor Yellow
}
