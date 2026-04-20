#!/usr/bin/env bash
# Sync skills from .agent/skills/ to all AI platform directories
# Usage: ./sync-skills-all-platforms.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$SCRIPT_DIR/lib/utils.sh" ]; then
    source "$SCRIPT_DIR/lib/utils.sh"
else
    echo "Error: Could not find lib/utils.sh"
    exit 1
fi

echo "ÃƒÆ’Ã‚Â°Ãƒâ€¦Ã‚Â¸ÃƒÂ¢Ã¢â€šÂ¬Ã‚ÂÃƒÂ¢Ã¢â€šÂ¬Ã…Â¾ Synchronizing skills across all AI platform directories..."
echo ""

# Source directory
SOURCE="$REPO_ROOT/.agent/skills"

# Target directories
TARGETS=(
    "$REPO_ROOT/.gemini/skills"
    "$REPO_ROOT/.codex/skills"
    "$REPO_ROOT/.claude/skills"
)

# Validate source exists
if [ ! -d "$SOURCE" ]; then
    echo "ÃƒÆ’Ã‚Â¢Ãƒâ€šÃ‚ÂÃƒâ€¦Ã¢â‚¬â„¢ Error: Source directory not found: $SOURCE"
    exit 1
fi

# Count source skills
SKILL_COUNT=$(find "$SOURCE" -mindepth 2 -maxdepth 2 -name "SKILL.md" | wc -l)
echo "ÃƒÆ’Ã‚Â°Ãƒâ€¦Ã‚Â¸ÃƒÂ¢Ã¢â€šÂ¬Ã…â€œÃƒâ€šÃ‚Â¦ Source: $SOURCE ($SKILL_COUNT skills)"
echo ""

# Sync to each target
for TARGET in "${TARGETS[@]}"; do
    # Create target if it doesn't exist
    mkdir -p "$TARGET"

    # Remove old content
    rm -rf "${TARGET:?}"/*

    # Copy all skills
    cp -r "$SOURCE"/* "$TARGET"/

    # Verify
    TARGET_SKILL_COUNT=$(find "$TARGET" -mindepth 2 -maxdepth 2 -name "SKILL.md" 2>/dev/null | wc -l || echo "0")

    if [ "$TARGET_SKILL_COUNT" -eq "$SKILL_COUNT" ]; then
        echo "ÃƒÆ’Ã‚Â¢Ãƒâ€¦Ã¢â‚¬Å“ÃƒÂ¢Ã¢â€šÂ¬Ã‚Â¦ Synced to: $TARGET ($TARGET_SKILL_COUNT skills)"
    else
        echo "ÃƒÆ’Ã‚Â¢Ãƒâ€¦Ã‚Â¡Ãƒâ€šÃ‚Â ÃƒÆ’Ã‚Â¯Ãƒâ€šÃ‚Â¸Ãƒâ€šÃ‚Â  Warning: Skill count mismatch in $TARGET (expected: $SKILL_COUNT, found: $TARGET_SKILL_COUNT)"
    fi
done

echo ""
echo "ÃƒÆ’Ã‚Â°Ãƒâ€¦Ã‚Â¸Ãƒâ€¦Ã‚Â½ÃƒÂ¢Ã¢â€šÂ¬Ã‚Â° Synchronization complete!"
echo ""
echo "ÃƒÆ’Ã‚Â°Ãƒâ€¦Ã‚Â¸ÃƒÂ¢Ã¢â€šÂ¬Ã¢â€žÂ¢Ãƒâ€šÃ‚Â¡ Next steps:"
echo "   1. Run validation: .agent/skills/list-skills.sh --validate"
echo "   2. Review changes: git status"
echo "   3. Commit if needed: git add .agent .gemini .codex .claude && git commit -m 'chore: sync skills'"
