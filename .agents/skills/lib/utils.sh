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

if [ -d "$REPO_ROOT/.agents" ]; then
    SKILLS_DIR="$REPO_ROOT/.agents/skills"
else
    SKILLS_DIR="$REPO_ROOT/.agent/skills"
fi
