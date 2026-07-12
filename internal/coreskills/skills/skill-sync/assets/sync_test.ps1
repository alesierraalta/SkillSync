$ErrorActionPreference = "Continue" # Don't stop on assertion failure

# Paths
$ScriptDir = $PSScriptRoot
# sync.sh is the script actually shipped by the project; sync_test.sh exercises
# it on Unix/Git-Bash. On Windows we run the same fixture from PowerShell by
# invoking sync.sh through bash so the test reflects real script behavior.
$SyncScript = Join-Path $ScriptDir "sync.sh"

# Colors
function Write-Pass { Write-Host "PASS" -ForegroundColor Green }
function Write-Fail { param($Msg) Write-Host "FAIL: $Msg" -ForegroundColor Red }

# Counters
$global:TestsRun = 0
$global:TestsPassed = 0
$global:TestsFailed = 0

# Test Environment
$global:TestDir = ""

# Locate bash.exe (Git for Windows installs to a well-known path; we also try
# a few alternatives so the harness works on dev machines that put Git
# elsewhere).
function Find-Bash {
    $candidates = @(
        "C:\Program Files\Git\bin\bash.exe",
        "C:\Program Files (x86)\Git\bin\bash.exe",
        "$env:LOCALAPPDATA\Programs\Git\bin\bash.exe"
    )
    foreach ($p in $candidates) {
        if (Test-Path $p) { return $p }
    }
    # Fall back to PATH lookup
    $cmd = Get-Command bash.exe -ErrorAction SilentlyContinue
    if ($cmd) { return $cmd.Source }
    return $null
}

$global:BashExe = Find-Bash
if (-not $global:BashExe) {
    Write-Host "ERROR: bash.exe not found. Install Git for Windows or set PATH so bash is discoverable." -ForegroundColor Red
    exit 1
}

function Setup-TestEnv {
    $global:TestDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
    New-Item -ItemType Directory -Path $global:TestDir -Force | Out-Null

    # sync.sh's find_repo_root climbs PWD until it finds .agent/.agents. The
    # fixture must create one at $TestDir so REPO_ROOT is anchored there.
    New-Item -ItemType Directory -Path (Join-Path $global:TestDir ".agent") -Force | Out-Null

    # Multi-provider skill directories (matches SKILL_PROVIDERS in lib/utils.sh).
    $dirs = @(
        ".claude\skills\claude-skill",
        ".opencode\skills\opencode-skill",
        ".qwen\skills\qwen-skill",
        ".agents\skills\agents-skill",
        ".agent\skills\no-metadata",
        ".agent\skills\skill-sync\assets"
    )
    foreach ($d in $dirs) {
        New-Item -ItemType Directory -Path (Join-Path $global:TestDir $d) -Force | Out-Null
    }

    function Write-Skill([string]$RelPath, [string]$Name, [string]$Scope, [string]$Action) {
        $path = Join-Path $global:TestDir $RelPath
        $content = @"
---
name: $Name
description: >
  Mock skill for $Name testing.
license: Apache-2.0
metadata:
  author: test
  version: "1.0"
  scope: [$Scope]
  auto_invoke: "$Action"
allowed-tools: Read
---

# $Name
"@
        # Write without UTF-8 BOM: bash awk in sync.sh reads the frontmatter
        # with strict line-by-line matching and a BOM breaks the `^---$` test.
        [System.IO.File]::WriteAllText($path, $content, [System.Text.UTF8Encoding]::new($false))
    }

    Write-Skill ".claude\skills\claude-skill\SKILL.md" "claude-skill" "root" "Claude action"
    Write-Skill ".opencode\skills\opencode-skill\SKILL.md" "opencode-skill" "root" "OpenCode action"
    Write-Skill ".qwen\skills\qwen-skill\SKILL.md" "qwen-skill" "root" "Qwen action"
    Write-Skill ".agents\skills\agents-skill\SKILL.md" "agents-skill" "root" "Agents action"

    # Skill missing scope and auto_invoke -- must be reported in
    # "Skills missing sync metadata" section.
    $noMeta = @"
---
name: no-metadata
description: Skill without sync metadata.
license: Apache-2.0
metadata:
  author: test
  version: "1.0"
allowed-tools: Read
---
"@
    [System.IO.File]::WriteAllText((Join-Path $global:TestDir ".agent\skills\no-metadata\SKILL.md"), $noMeta, [System.Text.UTF8Encoding]::new($false))

    # Target AGENTS.md (sync.sh writes root|api|common|infra|database here).
    $agents = @"
# Root AGENTS

> **Skills Reference**: For detailed patterns, use these skills:
> - [`agents-skill`](.agents/skills/agents-skill/SKILL.md)

## Project Overview

This is the root agents file.
"@
    [System.IO.File]::WriteAllText((Join-Path $global:TestDir "AGENTS.md"), $agents, [System.Text.UTF8Encoding]::new($false))

    # Copy sync.sh to the canonical test location.
    Copy-Item $SyncScript (Join-Path $global:TestDir ".agent\skills\skill-sync\assets\sync.sh") -Force

    # Locate utils.sh by walking up from $ScriptDir (the test may live in any
    # of the project's provider mirrors). Look for lib/utils.sh, utils.sh,
    # .agents/skills/lib/utils.sh, .agent/skills/lib/utils.sh, or
    # skills/lib/utils.sh at each ancestor up to 8 levels deep.
    $utilsSrc = $null
    $search = $ScriptDir
    for ($i = 0; $i -lt 8; $i++) {
        $search = Split-Path -Path $search -Parent
        if (-not $search) { break }
        $candidates = @(
            (Join-Path $search "lib\utils.sh"),
            (Join-Path $search "utils.sh"),
            (Join-Path $search ".agents\skills\lib\utils.sh"),
            (Join-Path $search ".agent\skills\lib\utils.sh"),
            (Join-Path $search "skills\lib\utils.sh"),
            (Join-Path $search "internal\coreskills\skills\lib\utils.sh")
        )
        foreach ($c in $candidates) {
            if (Test-Path $c) { $utilsSrc = $c; break }
        }
        if ($utilsSrc) { break }
    }

    if ($utilsSrc) {
        $utilsDest = Join-Path $global:TestDir ".agent\skills\lib\utils.sh"
        New-Item -ItemType Directory -Path (Split-Path $utilsDest) -Force | Out-Null
        Copy-Item $utilsSrc $utilsDest -Force
    } else {
        Write-Host "ERROR: sync_test.ps1 could not locate lib/utils.sh from $ScriptDir" -ForegroundColor Red
    }
}

function Teardown-TestEnv {
    if ($global:TestDir -and (Test-Path $global:TestDir)) {
        Remove-Item $global:TestDir -Recurse -Force -ErrorAction SilentlyContinue
    }
    $global:TestDir = ""
}

function Invoke-Sync {
    # Run sync.sh under bash with the same args we receive. We must cd into
    # $TestDir first so sync.sh's find_repo_root anchors REPO_ROOT there.
    # Output is captured (including stdout+stderr) and returned as a single
    # string. ANSI color codes are preserved -- callers can use Select-String
    # which ignores them as long as the needle text is plain.
    param([string[]]$Arguments)
    $script = Join-Path $global:TestDir ".agent\skills\skill-sync\assets\sync.sh"
    $testDirPosix = $global:TestDir -replace '\\','/'
    $scriptPosix = $script -replace '\\','/'
    $extraArgs = ""
    foreach ($a in $Arguments) {
        if ($extraArgs.Length -gt 0) { $extraArgs += " " }
        $extraArgs += ([System.Management.Automation.Language.CodeGeneration]::EscapeSingleQuotedStringContent($a))
    }
    $cmd = "cd '$testDirPosix' && bash '$scriptPosix' $extraArgs"
    $output = & $global:BashExe -c $cmd 2>&1 | Out-String
    return $output
}

function Run-Test {
    param($Name, $ScriptBlock)
    $global:TestsRun++
    Write-Host -NoNewline "  $Name... "
    Setup-TestEnv
    try {
        & $ScriptBlock
        $global:TestsPassed++
        Write-Pass
    } catch {
        $global:TestsFailed++
        Write-Fail $_.Exception.Message
    }
    Teardown-TestEnv
}

function Assert-Contains {
    param($Haystack, $Needle, $Message)
    if ($Haystack | Select-String -SimpleMatch $Needle -Quiet) { return }
    throw "$Message :: string not found: $Needle"
}

function Assert-NotContains {
    param($Haystack, $Needle, $Message)
    if (-not ($Haystack | Select-String -SimpleMatch $Needle -Quiet)) { return }
    throw "$Message :: string should not be found: $Needle"
}

function Assert-FileContains {
    param($File, $Needle, $Message)
    $content = Get-Content -Path $File -Raw
    if ($content | Select-String -SimpleMatch $Needle -Quiet) { return }
    throw "$Message :: file $File missing: $Needle"
}

function Assert-FileNotContains {
    param($File, $Needle, $Message)
    $content = Get-Content -Path $File -Raw -ErrorAction SilentlyContinue
    if (-not $content) { return }
    if (-not ($content | Select-String -SimpleMatch $Needle -Quiet)) { return }
    throw "$Message :: file $File should not contain: $Needle"
}

function Assert-Equals {
    param($Expected, $Actual, $Message)
    if ($Expected -eq $Actual) { return }
    throw "$Message :: expected '$Expected' got '$Actual'"
}

Write-Host ""
Write-Host "Running sync.sh unit tests (via bash on Windows)"
Write-Host "=============================="

# Flag tests
Run-Test "flag help shows usage" {
    $out = Invoke-Sync @("--help")
    Assert-Contains $out "Usage:" "Help should show usage"
    Assert-Contains $out "--dry-run" "Help should mention --dry-run"
    Assert-Contains $out "--scope" "Help should mention --scope"
}

Run-Test "flag unknown reports error" {
    $out = Invoke-Sync @("--unknown")
    Assert-Contains $out "Unknown option" "Should report unknown option"
}

Run-Test "flag dryrun shows changes" {
    $out = Invoke-Sync @("--dry-run")
    Assert-Contains $out "[DRY RUN]" "Should show dry run marker"
    Assert-Contains $out "Would update" "Should say would update"
}

Run-Test "flag dryrun no file changes" {
    $null = Invoke-Sync @("--dry-run")
    Assert-FileNotContains (Join-Path $global:TestDir "AGENTS.md") "### Auto-invoke Skills" "AGENTS.md should not be modified in dry run"
}

Run-Test "flag scope filters to matching" {
    # Switch opencode-skill to a non-root scope so the filter is observable.
    $path = Join-Path $global:TestDir ".opencode\skills\opencode-skill\SKILL.md"
    $content = @"
---
name: opencode-skill
description: Mock skill under .opencode with a non-root scope.
license: Apache-2.0
metadata:
  author: test
  version: "1.0"
  scope: [ui]
  auto_invoke: "OpenCode action"
allowed-tools: Read
---
"@
    [System.IO.File]::WriteAllText($path, $content, [System.Text.UTF8Encoding]::new($false))

    $out = Invoke-Sync @("--dry-run", "--scope", "root")
    Assert-Contains $out "Processing: root" "Should process root scope"
    Assert-NotContains $out "Processing: ui" "Should not process ui scope"
}

# Metadata tests
Run-Test "metadata extracts scope" {
    $out = Invoke-Sync @("--dry-run")
    Assert-Contains $out "Processing: root" "Should detect root scope"
}

Run-Test "metadata extracts auto invoke" {
    $out = Invoke-Sync @("--dry-run")
    Assert-Contains $out "Claude action" "Should extract claude-skill auto_invoke"
    Assert-Contains $out "OpenCode action" "Should extract opencode-skill auto_invoke"
    Assert-Contains $out "Agents action" "Should extract agents-skill auto_invoke"
}

Run-Test "metadata missing reports skills" {
    $out = Invoke-Sync @("--dry-run")
    Assert-Contains $out "Skills missing sync metadata" "Should report missing metadata section"
    Assert-Contains $out "no-metadata" "Should list skill without metadata"
}

# Generation tests
Run-Test "generate creates section" {
    $null = Invoke-Sync @()
    $agentsPath = Join-Path $global:TestDir "AGENTS.md"
    Assert-FileContains $agentsPath "### Auto-invoke Skills" "Should create Auto-invoke section"
    Assert-FileContains $agentsPath "| Action | Skill |" "Should create table header"
}

Run-Test "generate includes all skills" {
    $null = Invoke-Sync @()
    $agentsPath = Join-Path $global:TestDir "AGENTS.md"
    Assert-FileContains $agentsPath "claude-skill" "AGENTS.md should contain claude-skill"
    Assert-FileContains $agentsPath "opencode-skill" "AGENTS.md should contain opencode-skill"
    Assert-FileContains $agentsPath "qwen-skill" "AGENTS.md should contain qwen-skill"
    Assert-FileContains $agentsPath "agents-skill" "AGENTS.md should contain agents-skill"
}

Run-Test "generate uses action text" {
    $null = Invoke-Sync @()
    $agentsPath = Join-Path $global:TestDir "AGENTS.md"
    Assert-FileContains $agentsPath "Claude action" "Should include claude-skill auto_invoke text"
    Assert-FileContains $agentsPath "OpenCode action" "Should include opencode-skill auto_invoke text"
}

Run-Test "generate splits multi action auto invoke list" {
    $path = Join-Path $global:TestDir ".claude\skills\claude-skill\SKILL.md"
    $content = @"
---
name: claude-skill
description: Mock skill with multi-action auto_invoke list.
license: Apache-2.0
metadata:
  author: test
  version: "1.0"
  scope: [root]
  auto_invoke:
    - "Action B"
    - "Action A"
allowed-tools: Read
---
"@
    [System.IO.File]::WriteAllText($path, $content, [System.Text.UTF8Encoding]::new($false))

    $null = Invoke-Sync @()
    $agentsPath = Join-Path $global:TestDir "AGENTS.md"
    # Use single-quoted here-strings: backticks are literal in single-quoted
    # PowerShell strings (no escape interpretation).
    $rowA = '| Action A | `claude-skill` |'
    $rowB = '| Action B | `claude-skill` |'
    Assert-FileContains $agentsPath $rowA "Should create row for Action A"
    Assert-FileContains $agentsPath $rowB "Should create row for Action B"
}

# Update tests
Run-Test "update preserves header" {
    $null = Invoke-Sync @()
    $agentsPath = Join-Path $global:TestDir "AGENTS.md"
    Assert-FileContains $agentsPath "# Root AGENTS" "Should preserve original header"
}

Run-Test "update preserves skills reference" {
    $null = Invoke-Sync @()
    $agentsPath = Join-Path $global:TestDir "AGENTS.md"
    Assert-FileContains $agentsPath "Skills Reference" "Should preserve Skills Reference section"
}

Run-Test "update preserves content after" {
    $null = Invoke-Sync @()
    $agentsPath = Join-Path $global:TestDir "AGENTS.md"
    Assert-FileContains $agentsPath "## Project Overview" "Should preserve content after the inserted block"
}

# Idempotency tests
Run-Test "idempotent multiple runs" {
    $null = Invoke-Sync @()
    $first = Get-Content (Join-Path $global:TestDir "AGENTS.md") -Raw
    $null = Invoke-Sync @()
    $second = Get-Content (Join-Path $global:TestDir "AGENTS.md") -Raw
    Assert-Equals $first $second "Multiple runs should produce identical output"
}

Run-Test "idempotent no duplicate sections" {
    $null = Invoke-Sync @()
    $null = Invoke-Sync @()
    $null = Invoke-Sync @()
    $count = (Select-String -Path (Join-Path $global:TestDir "AGENTS.md") -Pattern "### Auto-invoke Skills" -SimpleMatch).Count
    Assert-Equals 1 $count "Should have exactly one Auto-invoke section"
}

# Multiscope test
Run-Test "multiscope skill processes both scopes" {
    $path = Join-Path $global:TestDir ".claude\skills\claude-skill\SKILL.md"
    $content = @"
---
name: claude-skill
description: Mock skill with multiple scopes.
license: Apache-2.0
metadata:
  author: test
  version: "1.0"
  scope: [root, api]
  auto_invoke: "Multi-scope action"
allowed-tools: Read
---
"@
    [System.IO.File]::WriteAllText($path, $content, [System.Text.UTF8Encoding]::new($false))

    $out = Invoke-Sync @("--dry-run")
    Assert-Contains $out "Processing: root" "Should process root scope"
    Assert-Contains $out "Processing: api" "Should process api scope"
}

Write-Host ""
Write-Host "=============================="
if ($global:TestsFailed -eq 0) {
    Write-Host "All $global:TestsRun tests passed!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "$($global:TestsFailed) of $($global:TestsRun) tests failed" -ForegroundColor Red
    exit 1
}
