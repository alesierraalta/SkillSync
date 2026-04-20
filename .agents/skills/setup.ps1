<#
.SYNOPSIS
    Setup AI Skills for any project.

.DESCRIPTION
    Configures AI coding assistants that follow agentskills.io standard:
      - Claude Code: .claude/skills/ junction + CLAUDE.md copies
      - Gemini CLI: .gemini/skills/ junction + GEMINI.md copies
      - Codex (OpenAI): .codex/skills/ junction + AGENTS.md (native)
      - GitHub Copilot: .github/copilot-instructions.md copy

.PARAMETER All
    Configure all AI assistants

.PARAMETER Claude
    Configure Claude Code

.PARAMETER Gemini
    Configure Gemini CLI

.PARAMETER Codex
    Configure Codex (OpenAI)

.PARAMETER Copilot
    Configure GitHub Copilot

.EXAMPLE
    .\setup.ps1 -All
    Configures all AI assistants

.EXAMPLE
    .\setup.ps1 -Claude -Gemini
    Configures only Claude and Gemini
#>

[CmdletBinding()]
param(
    [switch]$All,
    [switch]$Claude,
    [switch]$Gemini,
    [switch]$Codex,
    [switch]$Copilot,
    [switch]$OpenCode,
    [switch]$Local,
    [string]$TargetDir = "",
    [switch]$Help
)

$ErrorActionPreference = 'Stop'

# Resolve paths
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

# Determine TargetDir where .claude, .gemini, etc. will be created
if ([string]::IsNullOrWhiteSpace($TargetDir)) {
    if ((Get-Location).Path -match "[\\/]\.agent[\\/]skills$") {
        $TargetDir = Split-Path -Parent (Split-Path -Parent (Get-Location).Path)
    }
    else {
        $TargetDir = (Get-Location).Path
    }
}

$RepoRoot = $TargetDir
$SkillsSource = $ScriptDir

# =============================================================================
# HELPER FUNCTIONS
# =============================================================================

function Show-Help {
    Write-Host "Usage: .\setup.ps1 [OPTIONS]" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Configure AI coding assistants for your project."
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -All       Configure all AI assistants"
    Write-Host "  -Claude    Configure Claude Code"
    Write-Host "  -Gemini    Configure Gemini CLI"
    Write-Host "  -Codex     Configure Codex (OpenAI)"
    Write-Host "  -Copilot   Configure GitHub Copilot"
    Write-Host "  -OpenCode  Configure OpenCode"
    Write-Host "  -Local     Copy skills physically instead of linking"
    Write-Host "  -Help      Show this help message"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\setup.ps1 -All              # All AI assistants"
    Write-Host "  .\setup.ps1 -Claude -Codex    # Only Claude and Codex"
}

function Copy-AgentsMd {
    param([string]$TargetName)

    $count = 0
    $agentsFiles = Get-ChildItem -Path $RepoRoot -Filter "AGENTS.md" -Recurse -ErrorAction SilentlyContinue | Where-Object { $_.FullName -notmatch 'node_modules|\.git' }

    foreach ($file in $agentsFiles) {
        $targetPath = Join-Path $file.DirectoryName $TargetName
        Copy-Item $file.FullName $targetPath -Force
        $count++
    }

    Write-Host "  [OK] Copied $count AGENTS.md -> $TargetName" -ForegroundColor Green
}

function New-SkillsJunction {
    param(
        [string]$AssistantDir,
        [string]$AssistantName
    )

    $targetDir = Join-Path $RepoRoot $AssistantDir
    $skillsLink = Join-Path $targetDir "skills"

    # Create assistant directory if needed
    if (-not (Test-Path $targetDir)) {
        New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
    }

    # Remove existing skills link/folder
    if (Test-Path $skillsLink) {
        $item = Get-Item $skillsLink -Force
        # Check if it's a junction/symlink
        if ($item.Attributes -band [System.IO.FileAttributes]::ReparsePoint) {
            cmd /c rmdir "$skillsLink" 2>$null
        }
        else {
            $backupName = "skills.backup." + (Get-Date -Format 'yyyyMMddHHmmss')
            Rename-Item $skillsLink (Join-Path $targetDir $backupName)
        }
    }

    if ($Local) {
        # Copy skills physically 
        Write-Host "  [OK] Copying .agent/skills -> $AssistantDir/skills/" -ForegroundColor Green
        Copy-Item -Path $SkillsSource -Destination $skillsLink -Recurse -Force
    }
    else {
        # Create junction (doesn't require admin privileges)
        cmd /c mklink /J "$skillsLink" "$SkillsSource" | Out-Null
        Write-Host "  [OK] Linking $AssistantDir/skills -> .agent/skills/" -ForegroundColor Green
    }
}

function Initialize-ClaudeAssistant {
    New-SkillsJunction -AssistantDir ".claude" -AssistantName "Claude Code"
    Copy-AgentsMd -TargetName "CLAUDE.md"
}

function Initialize-GeminiAssistant {
    New-SkillsJunction -AssistantDir ".gemini" -AssistantName "Gemini CLI"
    Copy-AgentsMd -TargetName "GEMINI.md"
}

function Initialize-CodexAssistant {
    New-SkillsJunction -AssistantDir ".codex" -AssistantName "Codex"
    Write-Host "  [OK] Codex uses AGENTS.md natively" -ForegroundColor Green
}

function Initialize-CopilotAssistant {
    $agentsMd = Join-Path $RepoRoot "AGENTS.md"
    if (Test-Path $agentsMd) {
        $ghDir = Join-Path $RepoRoot ".github"
        if (-not (Test-Path $ghDir)) {
            New-Item -ItemType Directory -Path $ghDir -Force | Out-Null
        }
        Copy-Item $agentsMd (Join-Path $ghDir "copilot-instructions.md") -Force
        Write-Host "  [OK] AGENTS.md -> .github/copilot-instructions.md" -ForegroundColor Green
    }
    else {
        Write-Host "  [WARN] AGENTS.md not found at repo root" -ForegroundColor Yellow
    }
}

function Initialize-OpenCodeAssistant {
    New-SkillsJunction -AssistantDir ".opencode" -AssistantName "OpenCode"
    Copy-AgentsMd -TargetName "OPENCODE.md"
}

# =============================================================================
# MAIN
# =============================================================================

if ($Help) {
    Show-Help
    exit 0
}

# Handle -All flag
if ($All) {
    $Claude = $true
    $Gemini = $true
    $Codex = $true
    $Copilot = $true
    $OpenCode = $true
}

Write-Host ""
Write-Host "AI Skills Setup" -ForegroundColor Cyan
Write-Host "===============" -ForegroundColor Cyan
Write-Host ""

# Count skills
$skillCount = (Get-ChildItem -Path $SkillsSource -Filter "SKILL.md" -Recurse -ErrorAction SilentlyContinue).Count

if ($skillCount -eq 0) {
    Write-Host "No skills found in $SkillsSource" -ForegroundColor Red
    exit 1
}

Write-Host "Found $skillCount skills to configure" -ForegroundColor Blue
Write-Host ""

# Interactive mode if no flags provided
if (-not ($Claude -or $Gemini -or $Codex -or $Copilot -or $OpenCode)) {
    Write-Host "No AI assistants selected via flags. Interactive mode:" -ForegroundColor Yellow
    Write-Host ""
    
    $customTarget = Read-Host "Install to current directory? ($TargetDir) (Y/n)"
    if ($customTarget -match "^[nN]") {
        $customPath = Read-Host "Enter full path to target project directory (e.g. C:\Path\To\Project)"
        if (-not [string]::IsNullOrWhiteSpace($customPath)) {
            $TargetDir = $customPath
            $RepoRoot = $TargetDir
            Write-Host "Target changed to: $TargetDir" -ForegroundColor Cyan
            Write-Host ""
        }
    }

    $Local = (Read-Host "Install skills locally (Copy) instead of Global (Link)? (y/N)") -match "^[yY]"
    $Claude = (Read-Host "Configure Claude Code? (y/N)") -match "^[yY]"
    $Gemini = (Read-Host "Configure Gemini CLI? (y/N)") -match "^[yY]"
    $Codex = (Read-Host "Configure Codex (OpenAI)? (y/N)") -match "^[yY]"
    $Copilot = (Read-Host "Configure GitHub Copilot? (y/N)") -match "^[yY]"
    $OpenCode = (Read-Host "Configure OpenCode? (y/N)") -match "^[yY]"
    
    if (-not ($Claude -or $Gemini -or $Codex -or $Copilot -or $OpenCode)) {
        Write-Host "No assistants selected. Exiting." -ForegroundColor Yellow
        exit 0
    }
}

# Count total steps
$total = 0
if ($Claude) { $total++ }
if ($Gemini) { $total++ }
if ($Codex) { $total++ }
if ($Copilot) { $total++ }
if ($OpenCode) { $total++ }

$step = 1

# Run selected setups
if ($Claude) {
    Write-Host "[$step/$total] Setting up Claude Code..." -ForegroundColor Yellow
    Initialize-ClaudeAssistant
    $step++
}

if ($Gemini) {
    Write-Host "[$step/$total] Setting up Gemini CLI..." -ForegroundColor Yellow
    Initialize-GeminiAssistant
    $step++
}

if ($Codex) {
    Write-Host "[$step/$total] Setting up Codex (OpenAI)..." -ForegroundColor Yellow
    Initialize-CodexAssistant
    $step++
}

if ($Copilot) {
    Write-Host "[$step/$total] Setting up GitHub Copilot..." -ForegroundColor Yellow
    Initialize-CopilotAssistant
    $step++
}

if ($OpenCode) {
    Write-Host "[$step/$total] Setting up OpenCode..." -ForegroundColor Yellow
    Initialize-OpenCodeAssistant
    $step++
}

# =============================================================================
# SUMMARY
# =============================================================================
Write-Host ""
Write-Host "[SUCCESS] Configured $skillCount AI skills!" -ForegroundColor Green
Write-Host ""
Write-Host "Configured:"
if ($Claude) { Write-Host "  - Claude Code:    .claude/skills/ + CLAUDE.md" }
if ($Codex) { Write-Host "  - Codex (OpenAI): .codex/skills/ + AGENTS.md (native)" }
if ($Gemini) { Write-Host "  - Gemini CLI:     .gemini/skills/ + GEMINI.md" }
if ($Copilot) { Write-Host "  - GitHub Copilot: .github/copilot-instructions.md" }
if ($OpenCode) { Write-Host "  - OpenCode:       .opencode/skills/ + OPENCODE.md" }
Write-Host ""
Write-Host "Note: Restart your AI assistant to load the skills." -ForegroundColor Blue
Write-Host "      AGENTS.md is the source of truth - edit it, then re-run this script." -ForegroundColor Blue
