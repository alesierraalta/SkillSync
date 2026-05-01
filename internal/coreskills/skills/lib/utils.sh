#!/usr/bin/env bash
# Common utilities for AI Skill management

# Find the repository root by looking for the .agent directory
find_repo_root() {
    local curr="$PWD"
    while [ "$curr" != "/" ] && [ "$curr" != "." ]; do
        if [ -d "$curr/.agent" ] || [ -d "$curr/.agents" ]; then
            echo "$curr"
            return 0
        fi
        curr=$(dirname "$curr")
    done
    return 1
}

REPO_ROOT=$(find_repo_root) || {
    # Fallback to script location if PWD is not inside a repo
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    # Try to find .agent or .agents climbing up from script dir
    curr="$SCRIPT_DIR"
    while [ "$curr" != "/" ] && [ "$curr" != "." ]; do
        if [ -d "$curr/.agent" ] || [ -d "$curr/.agents" ]; then
            REPO_ROOT="$curr"
            break
        fi
        curr=$(dirname "$curr")
    done
}

if [ -z "$REPO_ROOT" ]; then
    echo "Error: Could not find .agent or .agents directory in any parent path."
    exit 1
fi

# Multi-provider skill discovery (mirrors Go's DiscoverSkills in internal/discovery/service.go)
SKILL_PROVIDERS=(".claude" ".opencode" ".agents" ".gemini" ".cursor" ".copilot" ".qwen" ".agent")

# Build SKILL_DIRS array - all provider skill directories
SKILL_DIRS=()
for provider in "${SKILL_PROVIDERS[@]}"; do
    provider_path="$REPO_ROOT/$provider/skills"
    if [ -d "$provider_path" ]; then
        SKILL_DIRS+=("$provider_path")
    fi
done

# Legacy: SKILLS_DIR still points to .agents for backward compatibility with existing scripts
# that expect a single SKILLS_DIR variable (sync.sh and list-skills.sh need updating to use find_all_skill_files)
if [ -d "$REPO_ROOT/.agents" ]; then
    SKILLS_DIR="$REPO_ROOT/.agents/skills"
else
    SKILLS_DIR="$REPO_ROOT/.agent/skills"
fi

# find_all_skill_files: walks all provider skill directories, finds SKILL.md at depth 2
find_all_skill_files() {
    for dir in "${SKILL_DIRS[@]}"; do
        if [ -d "$dir" ]; then
            find "$dir" -mindepth 2 -maxdepth 2 -name SKILL.md -print 2>/dev/null
        fi
    done | sort
}
