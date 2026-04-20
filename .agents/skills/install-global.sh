#!/usr/bin/env bash
# Global Installer for AI Skill Tools

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SKILLS_BIN_DIR="$SCRIPT_DIR"

echo "ðŸ¤– AI Skills Global Installer"
echo "============================="
echo ""

DETECTED_SHELL=$(basename "$SHELL")
PROFILE_FILE=""

case "$DETECTED_SHELL" in
    bash)
        PROFILE_FILE="$HOME/.bashrc"
        [ -f "$HOME/.bash_profile" ] && PROFILE_FILE="$HOME/.bash_profile"
        ;;
    zsh)
        PROFILE_FILE="$HOME/.zshrc"
        ;;
    *)
        echo "Unsupported shell: $DETECTED_SHELL"
        echo "Please add $SKILLS_BIN_DIR to your PATH manually."
        exit 1
        ;;
esac

echo "Detected shell: $DETECTED_SHELL"
echo "Profile file: $PROFILE_FILE"
echo ""

# Create aliases
declare -A ALIASES
ALIASES["skills-sync"]="$SKILLS_BIN_DIR/skill-sync/assets/sync.sh"
ALIASES["skills-list"]="$SKILLS_BIN_DIR/list-skills.sh"
ALIASES["skills-setup"]="$SKILLS_BIN_DIR/setup.sh"
ALIASES["skills-sync-platforms"]="$SKILLS_BIN_DIR/sync-skills-all-platforms.sh"

for cmd in "${!ALIASES[@]}"; do
    target="${ALIASES[$cmd]}"
    if ! grep -q "alias $cmd=" "$PROFILE_FILE"; then
        echo "Adding alias $cmd..."
        echo "alias $cmd='$target'" >> "$PROFILE_FILE"
    else
        echo "Alias $cmd already exists in $PROFILE_FILE. Updating..."
        sed -i "s|alias $cmd=.*|alias $cmd='$target'|g" "$PROFILE_FILE"
    fi
done

echo ""
echo "âœ… Installation complete!"
echo "Please run: source $PROFILE_FILE"
echo "Then you can use: skills-sync, skills-list, skills-setup, skills-sync-platforms from any project root."
