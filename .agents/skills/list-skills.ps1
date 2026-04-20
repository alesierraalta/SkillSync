param (
    [switch]$All,
    [string]$Scope,
    [string]$Author,
    [string]$Search,
    [switch]$Validate,
    [switch]$Json
)

$ErrorActionPreference = "Stop"
$ScriptDir = $PSScriptRoot
$RepoRoot = $ScriptDir | Split-Path -Parent | Split-Path -Parent
$SkillsDir = $ScriptDir

# Colors (Emulation)
function Write-Color {
    param ($Text, $Color)
    Write-Host $Text -ForegroundColor $Color -NoNewline
}
function Write-Line { Write-Host "" }

# Helper to extract basic YAML fields using Regex
function Get-YamlField {
    param ($Content, $Field)
    if ($Content -match "(?m)^${Field}:\s*(.*)$") {
        return $Matches[1].Trim().Trim("'" ).Trim('"')
    }
    return $null
}

# Helper to extract metadata fields (nested)
function Get-MetadataField {
    param ($Content, $Field)

    # Extract metadata block
    if ($Content -match "(?ms)^metadata:\s*(.*?)(?:^([a-z].*):|\Z)") {
        $MetadataBlock = $Matches[1]

        # Check for specific field within block
        if ($MetadataBlock -match "(?m)^\s*${Field}:\s*(.*)$") {
            $RawValue = $Matches[1].Trim()

            # Check for inline list [a, b]
            if ($RawValue -match "^[(.*)]$") {
                return $Matches[1].Split(',') | ForEach-Object { $_.Trim().Trim("'" ).Trim('"') }
            }

            # Check for multiline list (if raw value is empty)
            if ([string]::IsNullOrWhiteSpace($RawValue)) {
                # Find where this field starts in the block
                $FieldStart = $MetadataBlock.IndexOf("$Field`:")
                if ($FieldStart -ge 0) {
                    $SubStr = $MetadataBlock.Substring($FieldStart)
                    $Lines = $SubStr -split "\r?\n"
                    $Items = @()
                    foreach ($Line in $Lines | Select-Object -Skip 1) {
                        if ($Line -match "^\s*-\s*(.*)") {
                            $Items += $Matches[1].Trim().Trim("'" ).Trim('"')
                        }
                        elseif ($Line -match "^\S") {
                            # New top level key or unindented
                            break
                        }
                    }
                    if ($Items.Count -gt 0) { return $Items }
                }
            }
            else {
                return $RawValue.Trim("'" ).Trim('"')
            }
        }
    }
    return $null
}

# Get all skill files
$SkillFiles = Get-ChildItem -Path $SkillsDir -Recurse -Filter "SKILL.md"

if ($Validate) {
    Write-Color "Validating Skills..." "Cyan"; Write-Line
    Write-Host "====================="

    $Errors = 0
    $Warnings = 0

    foreach ($File in $SkillFiles) {
        $Content = Get-Content $File.FullName -Raw
        $Name = Get-YamlField $Content "name"
        $Author = Get-MetadataField $Content "author"
        $Version = Get-MetadataField $Content "version"

        if (-not $Name) {
            Write-Color "  X Missing 'name' field" "Red"
            Write-Host " - $($File.Directory.Name)"
            $Errors++
        }

        if (-not $Author) {
            Write-Color "  ! Missing 'metadata.author'" "Yellow"
            Write-Host " - $Name"
            $Warnings++
        }

        if (-not $Version) {
            Write-Color "  ! Missing 'metadata.version'" "Yellow"
            Write-Host " - $Name"
            $Warnings++
        }
    }

    # Check AGENTS.md consistency
    $AgentsFile = Join-Path $RepoRoot "AGENTS.md"
    if (Test-Path $AgentsFile) {
        Write-Line
        Write-Color "Checking AGENTS.md consistency..." "Cyan"; Write-Line

        $AgentsContent = Get-Content $AgentsFile -Raw
        $AgentsSkills = [Regex]::Matches($AgentsContent, '`([a-z0-9-]+)`') | ForEach-Object { $_.Groups[1].Value } | Select-Object -Unique

        # Check skills in AGENTS.md but not on disk
        foreach ($SkillName in $AgentsSkills) {
            if (-not (Test-Path (Join-Path $SkillsDir $SkillName))) {
                Write-Color "  X '$SkillName' in AGENTS.md but not in .agent/skills/" "Red"; Write-Line
            }
        }

        # Check skills on disk but not in AGENTS.md
        foreach ($File in $SkillFiles) {
            $Content = Get-Content $File.FullName -Raw
            $Name = Get-YamlField $Content "name"
            if ($Name -and $AgentsSkills -notcontains $Name) {
                Write-Color "  ! '$Name' exists but not listed in AGENTS.md" "Yellow"; Write-Line
                $Warnings++
            }
        }
    }

    Write-Line
    Write-Color "Validation Summary:" "Cyan"; Write-Line
    Write-Color "  Errors: $Errors" "Red"; Write-Line
    Write-Color "  Warnings: $Warnings" "Yellow"; Write-Line

    if ($Errors -gt 0) { exit 1 } else { exit 0 }
}

$Results = @()

foreach ($File in $SkillFiles) {
    $Content = Get-Content $File.FullName -Raw
    $SkillObj = [PSCustomObject]@{
        Name        = Get-YamlField $Content "name"
        Description = Get-YamlField $Content "description"
        Author      = Get-MetadataField $Content "author"
        Version     = Get-MetadataField $Content "version"
        Scope       = Get-MetadataField $Content "scope"
        AutoInvoke  = Get-MetadataField $Content "auto_invoke"
        Tools       = Get-YamlField $Content "allowed-tools"
        Path        = $File.FullName
    }

    # Filter Logic
    if ($Scope -and ($SkillObj.Scope -notcontains $Scope)) { continue }
    if ($Author -and ($SkillObj.Author -ne $Author)) { continue }
    if ($Search) {
        $S = $Search.ToLower()
        if (-not ($SkillObj.Name.ToLower().Contains($S) -or $SkillObj.Description.ToLower().Contains($S))) { continue }
    }

    $Results += $SkillObj
}

if ($Json) {
    $Results | Select-Object Name, Description, Author, Version, Scope, AutoInvoke, Tools | ConvertTo-Json -Depth 3
}
else {
    if ($Results.Count -eq 0) {
        Write-Color "No skills found matching your criteria." "Yellow"; Write-Line
    }
    else {
        Write-Line
        # Simple table format
        "{0,-20}  {1,-50}  {2,-12}  {3,-8}" -f "Skill", "Description", "Author", "Scope"
        "-" * 100

        foreach ($R in $Results) {
            $ScopeStr = if ($R.Scope -is [array]) { $R.Scope -join "+" } else { "$($R.Scope)" }
            $DescStr = if ($R.Description.Length -gt 47) { $R.Description.Substring(0, 47) + "..." } else { $R.Description }

            "{0,-20}  {1,-50}  {2,-12}  {3,-8}" -f $R.Name, $DescStr, $R.Author, $ScopeStr
        }

        Write-Line

        if ($All) {
            Write-Host "Full Details:" -ForegroundColor Cyan
            foreach ($R in $Results) {
                Write-Line
                Write-Color "Skill: $($R.Name)" "Cyan"; Write-Line
                Write-Host "  Description: $($R.Description)"
                Write-Host "  Author: $($R.Author)"
                Write-Host "  Version: $($R.Version)"
                Write-Host "  Scope: $($R.Scope -join ', ')"
                Write-Host "  Auto-invoke triggers:"
                if ($R.AutoInvoke) {
                    foreach ($Trig in $R.AutoInvoke) { Write-Host "    â€¢ $Trig" }
                }
                Write-Host "  Allowed tools: $($R.Tools)"
                Write-Host "  Location: $($R.Path)"
            }
        }

        Write-Line
        Write-Color "Total: $($Results.Count) skills" "Blue"; Write-Line
    }
}
