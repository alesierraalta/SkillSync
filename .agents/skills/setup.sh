#!/bin/bash
# Setup AI Skills for any project
# Configures AI coding assistants that follow agentskills.io standard:
#   - Claude Code: .claude/skills/ symlink + CLAUDE.md copies
#   - Gemini CLI: .gemini/skills/ symlink + GEMINI.md copies
#   - Codex (OpenAI): .codex/skills/ symlink + AGENTS.md (native)
#   - GitHub Copilot: .github/copilot-instructions.md copy
#   - OpenCode: .opencode/skills/ symlink + OPENCODE.md copies
#
# Usage:
#   ./setup.sh --all        # Configure all AI assistants
#   ./setup.sh --claude     # Configure only Claude Code
#   ./setup.sh --claude --codex  # Configure multiple

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$SCRIPT_DIR/lib/utils.sh" ]; then
    source "$SCRIPT_DIR/lib/utils.sh"
else
    echo "Error: Could not find lib/utils.sh"
    exit 1
fi
# REPO_ROOT is provided by utils.sh
SKILLS_SOURCE="$SCRIPT_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Selection flags
SETUP_CLAUDE=false
SETUP_GEMINI=false
SETUP_CODEX=false
SETUP_COPILOT=false
SETUP_OPENCODE=false
SETUP_LOCAL=false
SETUP_TARGET=""

# =============================================================================
# HELPER FUNCTIONS
# =============================================================================

show_help() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Configure AI coding assistants for your project."
    echo ""
    echo "Options:"
    echo "  --all       Configure all AI assistants"
    echo "  --claude    Configure Claude Code"
    echo "  --gemini    Configure Gemini CLI"
    echo "  --codex     Configure Codex (OpenAI)"
    echo "  --copilot   Configure GitHub Copilot"
    echo "  --opencode  Configure OpenCode"
    echo "  --local     Copy skills physically instead of linking"
    echo "  --target    Override target installation directory (default: current dir)"
    echo "  --help      Show this help message"
    echo ""
    echo "If no options provided, runs in interactive mode."
    echo ""
    echo "Examples:"
    echo "  $0                      # Interactive selection"
    echo "  $0 --all                # All AI assistants"
    echo "  $0 --claude --codex     # Only Claude and Codex"
}

setup_claude() {
    local target="$REPO_ROOT/.claude/skills"

    if [ ! -d "$REPO_ROOT/.claude" ]; then
        mkdir -p "$REPO_ROOT/.claude"
    fi

    if [ -L "$target" ]; then
        rm "$target"
    elif [ -d "$target" ]; then
        mv "$target" "$REPO_ROOT/.claude/skills.backup.$(date +%s)"
    fi

    if [ "$SETUP_LOCAL" = true ]; then
        cp -R "$SKILLS_SOURCE" "$target"
        echo -e "${GREEN}  ✓ Copying .agent/skills -> .claude/skills/${NC}"
    else
        ln -s "$SKILLS_SOURCE" "$target"
        echo -e "${GREEN}  ✓ .claude/skills -> .agent/skills/${NC}"
    fi

    # Copy AGENTS.md to CLAUDE.md
    copy_agents_md "CLAUDE.md"
}

setup_gemini() {
    local target="$REPO_ROOT/.gemini/skills"

    if [ ! -d "$REPO_ROOT/.gemini" ]; then
        mkdir -p "$REPO_ROOT/.gemini"
    fi

    if [ -L "$target" ]; then
        rm "$target"
    elif [ -d "$target" ]; then
        mv "$target" "$REPO_ROOT/.gemini/skills.backup.$(date +%s)"
    fi

    if [ "$SETUP_LOCAL" = true ]; then
        cp -R "$SKILLS_SOURCE" "$target"
        echo -e "${GREEN}  ✓ Copying .agent/skills -> .gemini/skills/${NC}"
    else
        ln -s "$SKILLS_SOURCE" "$target"
        echo -e "${GREEN}  ✓ .gemini/skills -> .agent/skills/${NC}"
    fi

    # Copy AGENTS.md to GEMINI.md
    copy_agents_md "GEMINI.md"
}

setup_codex() {
    local target="$REPO_ROOT/.codex/skills"

    if [ ! -d "$REPO_ROOT/.codex" ]; then
        mkdir -p "$REPO_ROOT/.codex"
    fi

    if [ -L "$target" ]; then
        rm "$target"
    elif [ -d "$target" ]; then
        mv "$target" "$REPO_ROOT/.codex/skills.backup.$(date +%s)"
    fi

    if [ "$SETUP_LOCAL" = true ]; then
        cp -R "$SKILLS_SOURCE" "$target"
        echo -e "${GREEN}  ✓ Copying .agent/skills -> .codex/skills/${NC}"
    else
        ln -s "$SKILLS_SOURCE" "$target"
        echo -e "${GREEN}  ✓ .codex/skills -> .agent/skills/${NC}"
    fi
    echo -e "${GREEN}  ✓ Codex uses AGENTS.md natively${NC}"
}

setup_copilot() {
    if [ -f "$REPO_ROOT/AGENTS.md" ]; then
        mkdir -p "$REPO_ROOT/.github"
        cp "$REPO_ROOT/AGENTS.md" "$REPO_ROOT/.github/copilot-instructions.md"
        echo -e "${GREEN}  âœ“ AGENTS.md -> .github/copilot-instructions.md${NC}"
    fi
}

setup_opencode() {
    local target="$REPO_ROOT/.opencode/skills"

    if [ ! -d "$REPO_ROOT/.opencode" ]; then
        mkdir -p "$REPO_ROOT/.opencode"
    fi

    if [ -L "$target" ]; then
        rm "$target"
    elif [ -d "$target" ]; then
        mv "$target" "$REPO_ROOT/.opencode/skills.backup.$(date +%s)"
    fi

    if [ "$SETUP_LOCAL" = true ]; then
        cp -R "$SKILLS_SOURCE" "$target"
        echo -e "${GREEN}  ✓ Copying .agent/skills -> .opencode/skills/${NC}"
    else
        ln -s "$SKILLS_SOURCE" "$target"
        echo -e "${GREEN}  ✓ .opencode/skills -> .agent/skills/${NC}"
    fi

    # Copy AGENTS.md to OPENCODE.md
    copy_agents_md "OPENCODE.md"
}

copy_agents_md() {
    local target_name="$1"
    local agents_files
    local count=0

    agents_files=$(find "$REPO_ROOT" -name "AGENTS.md" -not -path "*/node_modules/*" -not -path "*/.git/*" 2>/dev/null)

    for agents_file in $agents_files; do
        local agents_dir
        agents_dir=$(dirname "$agents_file")
        cp "$agents_file" "$agents_dir/$target_name"
        count=$((count + 1))
    done

    echo -e "${GREEN}  âœ“ Copied $count AGENTS.md -> $target_name${NC}"
}

# =============================================================================
# PARSE ARGUMENTS
# =============================================================================

while [[ $# -gt 0 ]]; do
    case $1 in
        --all)
            SETUP_CLAUDE=true
            SETUP_GEMINI=true
            SETUP_CODEX=true
            SETUP_COPILOT=true
            SETUP_OPENCODE=true
            shift
            ;;
        --claude)
            SETUP_CLAUDE=true
            shift
            ;;
        --gemini)
            SETUP_GEMINI=true
            shift
            ;;
        --codex)
            SETUP_CODEX=true
            shift
            ;;
        --copilot)
            SETUP_COPILOT=true
            shift
            ;;
        --opencode)
            SETUP_OPENCODE=true
            shift
            ;;
        --local|-l)
            SETUP_LOCAL=true
            shift
            ;;
        --target|-t)
            SETUP_TARGET="$2"
            shift 2
            ;;
        --help|-h)
            show_help
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            show_help
            exit 1
            ;;
    esac
done

# =============================================================================
# MAIN
# =============================================================================

echo "ðŸ¤– AI Skills Setup"
echo "=================="
echo ""

# Determine Target Dir
if [ -z "$SETUP_TARGET" ]; then
    if [[ "$PWD" == */.agent/skills ]]; then
        SETUP_TARGET="$(dirname "$(dirname "$PWD")")"
    else
        SETUP_TARGET="$PWD"
    fi
fi
REPO_ROOT="$SETUP_TARGET"

# Count skills
SKILL_COUNT=$(find "$SKILLS_SOURCE" -maxdepth 2 -name "SKILL.md" | wc -l | tr -d ' ')

if [ "$SKILL_COUNT" -eq 0 ]; then
    echo -e "${RED}No skills found in $SKILLS_SOURCE${NC}"
    exit 1
fi

echo -e "${BLUE}Found $SKILL_COUNT skills to configure${NC}"
echo ""

# Interactive mode if no flags provided
if [ "$SETUP_CLAUDE" = false ] && [ "$SETUP_GEMINI" = false ] && [ "$SETUP_CODEX" = false ] && [ "$SETUP_COPILOT" = false ] && [ "$SETUP_OPENCODE" = false ]; then
    echo -e "${YELLOW}No AI assistants selected via flags. Interactive mode:${NC}"
    echo ""
    
    read -p "Install to current directory? ($SETUP_TARGET) (Y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Nn]$ ]]; then
        read -p "Enter full path to target project directory (e.g. /path/to/project): " CUSTOM_PATH
        if [ -n "$CUSTOM_PATH" ]; then
            SETUP_TARGET="$CUSTOM_PATH"
            REPO_ROOT="$SETUP_TARGET"
            echo -e "${CYAN}Target changed to: $SETUP_TARGET${NC}"
            echo ""
        fi
    fi

    read -p "Install skills locally (Copy) instead of Global (Link)? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then SETUP_LOCAL=true; fi
    
    read -p "Configure Claude Code? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then SETUP_CLAUDE=true; fi
    
    read -p "Configure Gemini CLI? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then SETUP_GEMINI=true; fi
    
    read -p "Configure Codex (OpenAI)? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then SETUP_CODEX=true; fi
    
    read -p "Configure GitHub Copilot? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then SETUP_COPILOT=true; fi
    
    read -p "Configure OpenCode? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then SETUP_OPENCODE=true; fi

    if [ "$SETUP_CLAUDE" = false ] && [ "$SETUP_GEMINI" = false ] && [ "$SETUP_CODEX" = false ] && [ "$SETUP_COPILOT" = false ] && [ "$SETUP_OPENCODE" = false ]; then
        echo -e "${YELLOW}No assistants selected. Exiting.${NC}"
        exit 0
    fi
    echo ""
fi

# Run selected setups
STEP=1
TOTAL=0
[ "$SETUP_CLAUDE" = true ] && TOTAL=$((TOTAL + 1))
[ "$SETUP_GEMINI" = true ] && TOTAL=$((TOTAL + 1))
[ "$SETUP_CODEX" = true ] && TOTAL=$((TOTAL + 1))
[ "$SETUP_COPILOT" = true ] && TOTAL=$((TOTAL + 1))
[ "$SETUP_OPENCODE" = true ] && TOTAL=$((TOTAL + 1))

if [ "$SETUP_CLAUDE" = true ]; then
    echo -e "${YELLOW}[$STEP/$TOTAL] Setting up Claude Code...${NC}"
    setup_claude
    STEP=$((STEP + 1))
fi

if [ "$SETUP_GEMINI" = true ]; then
    echo -e "${YELLOW}[$STEP/$TOTAL] Setting up Gemini CLI...${NC}"
    setup_gemini
    STEP=$((STEP + 1))
fi

if [ "$SETUP_CODEX" = true ]; then
    echo -e "${YELLOW}[$STEP/$TOTAL] Setting up Codex (OpenAI)...${NC}"
    setup_codex
    STEP=$((STEP + 1))
fi

if [ "$SETUP_COPILOT" = true ]; then
    echo -e "${YELLOW}[$STEP/$TOTAL] Setting up GitHub Copilot...${NC}"
    setup_copilot
    STEP=$((STEP + 1))
fi

if [ "$SETUP_OPENCODE" = true ]; then
    echo -e "${YELLOW}[$STEP/$TOTAL] Setting up OpenCode...${NC}"
    setup_opencode
    STEP=$((STEP + 1))
fi

# =============================================================================
# SUMMARY
# =============================================================================
echo ""
echo -e "${GREEN}âœ… Successfully configured $SKILL_COUNT AI skills!${NC}"
echo ""
echo "Configured:"
[ "$SETUP_CLAUDE" = true ] && echo "  â€¢ Claude Code:    .claude/skills/ + CLAUDE.md"
[ "$SETUP_CODEX" = true ] && echo "  â€¢ Codex (OpenAI): .codex/skills/ + AGENTS.md (native)"
[ "$SETUP_GEMINI" = true ] && echo "  â€¢ Gemini CLI:     .gemini/skills/ + GEMINI.md"
[ "$SETUP_COPILOT" = true ] && echo "  â€¢ GitHub Copilot: .github/copilot-instructions.md"
[ "$SETUP_OPENCODE" = true ] && echo "  â€¢ OpenCode:       .opencode/skills/ + OPENCODE.md"
echo ""
echo -e "${BLUE}Note: Restart your AI assistant to load the skills.${NC}"
echo -e "${BLUE}      AGENTS.md is the source of truth - edit it, then re-run this script.${NC}"
