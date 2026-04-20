#!/usr/bin/env bash
# List and search AI skills
# Usage: ./list-skills.sh [options]
#   --all                List all skills with full details
#   --scope SCOPE        Filter by scope (root, api, common, infra)
#   --author AUTHOR      Filter by author
#   --search QUERY       Search in name/description
#   --validate           Validate skill integrity and AGENTS.md consistency
#   --json               Output as JSON (for scripts/integration)
#   --help               Show help

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ -f "$SCRIPT_DIR/lib/utils.sh" ]; then
    source "$SCRIPT_DIR/lib/utils.sh"
else
    echo "Error: Could not find lib/utils.sh"
    exit 1
fi
# REPO_ROOT and SKILLS_DIR are provided by utils.sh

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# Options
SHOW_ALL=false
FILTER_SCOPE=""
FILTER_AUTHOR=""
SEARCH_QUERY=""
VALIDATE_ONLY=false
JSON_OUTPUT=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --all)
            SHOW_ALL=true
            shift
            ;;
        --scope)
            FILTER_SCOPE="$2"
            shift 2
            ;;
        --author)
            FILTER_AUTHOR="$2"
            shift 2
            ;;
        --search)
            SEARCH_QUERY="$2"
            shift 2
            ;;
        --validate)
            VALIDATE_ONLY=true
            shift
            ;;
        --json)
            JSON_OUTPUT=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [options]"
            echo ""
            echo "List and search AI skills in the repository."
            echo ""
            echo "Options:"
            echo "  --all           List all skills with full details"
            echo "  --scope SCOPE   Filter by scope (root, api, common, infra)"
            echo "  --author AUTHOR Filter by author"
            echo "  --search QUERY  Search in name/description"
            echo "  --validate      Validate skill integrity and AGENTS.md consistency"
            echo "  --json          Output as JSON (for scripts/integration)"
            echo "  --help          Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0 --all                    # List all skills"
            echo "  $0 --scope api              # List API skills only"
            echo "  $0 --author a.sierra        # List skills by author"
            echo "  $0 --search 'Zod'            # Search for Zod-related skills"
            echo "  $0 --validate               # Validate skills and AGENTS.md"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

# Extract YAML frontmatter field
extract_field() {
    local file="$1"
    local field="$2"
    awk -v field="$field" '
        /^---$/ { in_frontmatter = !in_frontmatter; next }
        in_frontmatter && $1 == field":" {
            sub(/^[^:]+:[[:space:]]*/, "")

            if ($0 != "" && $0 != ">" && $0 != "|") {
                gsub(/^["'\''"]|["'\''"]$/, "")
                print
                exit
            }

            out = ""
            while (getline) {
                if (!in_frontmatter) break
                if (/^[a-z]/ && $0 !~ /^[[:space:]]/) break
                sub(/^[[:space:]]+/, "")
                if (out == "") out = $0
                else out = out " " $0
            }
            print out
            exit
        }
    ' "$file"
}

# Extract nested metadata field
extract_metadata() {
    local file="$1"
    local field="$2"

    awk -v field="$field" '
        function trim(s) {
            sub(/^[[:space:]]+/, "", s)
            sub(/[[:space:]]+$/, "", s)
            return s
        }

        /^---$/ { in_frontmatter = !in_frontmatter; next }

        in_frontmatter && /^metadata:/ { in_metadata = 1; next }
        in_frontmatter && in_metadata && /^[a-z]/ && !/^[[:space:]]/ { in_metadata = 0 }

        in_frontmatter && in_metadata && $1 == field":" {
            sub(/^[^:]+:[[:space:]]*/, "")

            if ($0 != "" && $0 != ">") {
                v = $0
                gsub(/^["'\''"]|["'\''"]$/, "", v)
                print trim(v)
                exit
            }

            out = ""
            while (getline) {
                if (!in_frontmatter) break
                if (!in_metadata) break
                if ($0 ~ /^[a-z]/ && $0 !~ /^[[:space:]]/) break

                line = $0
                if (line ~ /^[[:space:]]*-[[:space:]]*/) {
                    sub(/^[[:space:]]*-[[:space:]]*/, "", line)
                    line = trim(line)
                    gsub(/^["'\''"]|["'\''"]$/, "", line)
                    if (line != "") {
                        if (out == "") out = line
                        else out = out "|" line
                    }
                } else {
                    break
                }
            }

            if (out != "") print out
            exit
        }
    ' "$file"
}

# Process single skill and output as JSON
process_skill_json() {
    local skill_file="$1"
    local skill_name=$(extract_field "$skill_file" "name")
    local description=$(extract_field "$skill_file" "description")
    local author=$(extract_metadata "$skill_file" "author")
    local version=$(extract_metadata "$skill_file" "version")
    local scope=$(extract_metadata "$skill_file" "scope")
    local auto_invoke=$(extract_metadata "$skill_file" "auto_invoke")
    local allowed_tools=$(extract_field "$skill_file" "allowed-tools")

    # Apply filters
    if [ -n "$FILTER_SCOPE" ]; then
        if [[ ! "$scope" =~ $FILTER_SCOPE ]]; then
            return
        fi
    fi

    if [ -n "$FILTER_AUTHOR" ]; then
        if [ "$author" != "$FILTER_AUTHOR" ]; then
            return
        fi
    fi

    if [ -n "$SEARCH_QUERY" ]; then
        local search_lower=$(echo "$SEARCH_QUERY" | tr '[:upper:]' '[:lower:]')
        local name_lower=$(echo "$skill_name" | tr '[:upper:]' '[:lower:]')
        local desc_lower=$(echo "$description" | tr '[:upper:]' '[:lower:]')

        if [[ ! "$name_lower" =~ $search_lower ]] && [[ ! "$desc_lower" =~ $search_lower ]]; then
            return
        fi
    fi

    # Escape quotes for JSON
    description_escaped=$(echo "$description" | sed 's/\\/\\\\/g; s/"/\\"/g')
    author_escaped=$(echo "$author" | sed 's/\\/\\\\/g; s/"/\\"/g')

    # Convert scopes to JSON array
    local scopes_json="["
    IFS=', ' read -ra scopes <<< "$scope"
    for i in "${!scopes[@]}"; do
        [ $i -gt 0 ] && scopes_json+=", "
        scopes_json+="\"${scopes[$i]}\""
    done
    scopes_json+="]"

    # Convert auto_invoke to JSON array
    local auto_invoke_json="["
    if [ -n "$auto_invoke" ]; then
        IFS='|' read -ra invokes <<< "$auto_invoke"
        for i in "${!invokes[@]}"; do
            [ $i -gt 0 ] && auto_invoke_json+=", "
            auto_invoke_json+="\"${invokes[$i]}\""
        done
    fi
    auto_invoke_json+="]"

    # Escape tools for JSON
    local tools_escaped=$(echo "$allowed_tools" | sed 's/\\/\\\\/g; s/"/\\"/g')

    echo "  {"
    echo "    \"name\": \"$skill_name\","
    echo "    \"description\": \"$description_escaped\","
    echo "    \"author\": \"$author_escaped\","
    echo "    \"version\": \"$version\","
    echo "    \"scope\": $scopes_json,"
    echo "    \"auto_invoke\": $auto_invoke_json,"
    echo "    \"tools\": \"$tools_escaped\""
    echo -n "  }"
}

# Process single skill and output as table row
process_skill_table() {
    local skill_file="$1"
    local skill_name=$(extract_field "$skill_file" "name")
    local description=$(extract_field "$skill_file" "description")
    local author=$(extract_metadata "$skill_file" "author")
    local version=$(extract_metadata "$skill_file" "version")
    local scope=$(extract_metadata "$skill_file" "scope")
    local auto_invoke=$(extract_metadata "$skill_file" "auto_invoke")
    local allowed_tools=$(extract_field "$skill_file" "allowed-tools")

    # Apply filters
    if [ -n "$FILTER_SCOPE" ]; then
        if [[ ! "$scope" =~ $FILTER_SCOPE ]]; then
            return
        fi
    fi

    if [ -n "$FILTER_AUTHOR" ]; then
        if [ "$author" != "$FILTER_AUTHOR" ]; then
            return
        fi
    fi

    if [ -n "$SEARCH_QUERY" ]; then
        local search_lower=$(echo "$SEARCH_QUERY" | tr '[:upper:]' '[:lower:]')
        local name_lower=$(echo "$skill_name" | tr '[:upper:]' '[:lower:]')
        local desc_lower=$(echo "$description" | tr '[:upper:]' '[:lower:]')

        if [[ ! "$name_lower" =~ $search_lower ]] && [[ ! "$desc_lower" =~ $search_lower ]]; then
            return
        fi
    fi

    # Truncate description
    local desc_truncated="${description:0:47}"
    if [ ${#description} -gt 47 ]; then
        desc_truncated="${desc_truncated}..."
    fi

    # Format scope
    IFS=', ' read -ra scopes <<< "$scope"
    local scope_display="${scopes[0]}"
    if [ ${#scopes[@]} -gt 1 ]; then
        scope_display="${scope_display}+"
    fi

    # Colorize scope
    local scope_color="$NC"
    case "$scope_display" in
        root*) scope_color="$CYAN" ;;
        api*) scope_color="$BLUE" ;;
        common*) scope_color="$GREEN" ;;
        infra*) scope_color="$YELLOW" ;;
    esac

    printf "%-20s  %-50s  %-12s  ${scope_color}%-8s${NC}\n" "\`$skill_name\`" "$desc_truncated" "$author" "$scope_display"
}

# Process single skill and output full details
process_skill_details() {
    local skill_file="$1"
    local skill_name=$(extract_field "$skill_file" "name")
    local description=$(extract_field "$skill_file" "description")
    local author=$(extract_metadata "$skill_file" "author")
    local version=$(extract_metadata "$skill_file" "version")
    local scope=$(extract_metadata "$skill_file" "scope")
    local auto_invoke=$(extract_metadata "$skill_file" "auto_invoke")
    local allowed_tools=$(extract_field "$skill_file" "allowed-tools")

    # Apply filters
    if [ -n "$FILTER_SCOPE" ]; then
        if [[ ! "$scope" =~ $FILTER_SCOPE ]]; then
            return
        fi
    fi

    if [ -n "$FILTER_AUTHOR" ]; then
        if [ "$author" != "$FILTER_AUTHOR" ]; then
            return
        fi
    fi

    if [ -n "$SEARCH_QUERY" ]; then
        local search_lower=$(echo "$SEARCH_QUERY" | tr '[:upper:]' '[:lower:]')
        local name_lower=$(echo "$skill_name" | tr '[:upper:]' '[:lower:]')
        local desc_lower=$(echo "$description" | tr '[:upper:]' '[:lower:]')

        if [[ ! "$name_lower" =~ $search_lower ]] && [[ ! "$desc_lower" =~ $search_lower ]]; then
            return
        fi
    fi

    echo -e "${CYAN}Skill: $skill_name${NC}"
    echo "  Description: $description"
    echo "  Author: $author"
    echo "  Version: $version"
    echo "  Scope: $scope"
    echo "  Auto-invoke triggers:"
    IFS='|' read -ra invokes <<< "$auto_invoke"
    for invoke in "${invokes[@]}"; do
        echo "    Ã¢â‚¬Â¢ $invoke"
    done
    echo "  Allowed tools: $allowed_tools"
    echo "  Location: $skill_file"
    echo ""
}

# Validate skills
validate_skills() {
    echo -e "${BLUE}Validating Skills...${NC}"
    echo "====================="
    echo ""

    local errors=0
    local warnings=0
    local skill_count=0

    # Check skills directory structure
    find "$SKILLS_DIR" -mindepth 2 -maxdepth 2 -name SKILL.md -print | sort | while IFS= read -r skill_file; do
        [ -f "$skill_file" ] || continue

        local skill_name=$(extract_field "$skill_file" "name")
        local author=$(extract_metadata "$skill_file" "author")
        local version=$(extract_metadata "$skill_file" "version")

        skill_count=$((skill_count + 1))

        # Validate required fields
        if [ -z "$skill_name" ]; then
            echo -e "  ${RED}Ã¢Å“â€” Missing 'name' field${NC} - $(dirname "$skill_file")"
            errors=$((errors + 1))
        fi

        if [ -z "$author" ]; then
            echo -e "  ${YELLOW}Ã¢Å¡Â  Missing 'metadata.author'${NC} - $skill_name"
            warnings=$((warnings + 1))
        fi

        if [ -z "$version" ]; then
            echo -e "  ${YELLOW}Ã¢Å¡Â  Missing 'metadata.version'${NC} - $skill_name"
            warnings=$((warnings + 1))
        fi
    done

    # Check AGENTS.md consistency
    local agents_file="$REPO_ROOT/AGENTS.md"

    if [ -f "$agents_file" ]; then
        echo ""
        echo -e "${BLUE}Checking AGENTS.md consistency...${NC}"

        # Extract skill names from AGENTS.md tables
        local agents_skills=$(grep -oE '\`([a-z0-9-]+)\`' "$agents_file" | grep -oE '[a-z0-9-]+' | sort -u)

        # Find skills listed in AGENTS.md but not in .agent/skills/
        echo "$agents_skills" | while IFS= read -r skill_in_agents; do
            if [ ! -d "$SKILLS_DIR/$skill_in_agents" ]; then
                echo -e "  ${RED}Ã¢Å“â€” '$skill_in_agents' in AGENTS.md but not in .agent/skills/${NC}"
            fi
        done

        # Find skills in .agent/skills/ but not listed in AGENTS.md
        find "$SKILLS_DIR" -mindepth 2 -maxdepth 2 -name SKILL.md -print | sort | while IFS= read -r skill_file; do
            [ -f "$skill_file" ] || continue
            local skill_name=$(extract_field "$skill_file" "name")
            if ! grep -q "\`$skill_name\`" "$agents_file"; then
                echo -e "  ${YELLOW}Ã¢Å¡Â  '$skill_name' exists but not listed in AGENTS.md${NC}"
                warnings=$((warnings + 1))
            fi
        done
    fi

    # Summary
    echo ""
    echo -e "${BLUE}Validation Summary:${NC}"
    echo -e "  ${RED}Errors: $errors${NC}"
    echo -e "  ${YELLOW}Warnings: $warnings${NC}"

    if [ $errors -gt 0 ]; then
        echo ""
        echo -e "${RED}Ã¢ÂÅ’ Validation failed!${NC}"
        return 1
    else
        echo ""
        echo -e "${GREEN}Ã¢Å“â€¦ Validation passed!${NC}"
        return 0
    fi
}

# Count matching skills
count_skills() {
    find "$SKILLS_DIR" -mindepth 2 -maxdepth 2 -name SKILL.md -print | sort | while IFS= read -r skill_file; do
        [ -f "$skill_file" ] || continue

        local skill_name=$(extract_field "$skill_file" "name")
        local description=$(extract_field "$skill_file" "description")
        local author=$(extract_metadata "$skill_file" "author")
        local scope=$(extract_metadata "$skill_file" "scope")

        # Apply filters
        if [ -n "$FILTER_SCOPE" ]; then
            if [[ ! "$scope" =~ $FILTER_SCOPE ]]; then
                continue
            fi
        fi

        if [ -n "$FILTER_AUTHOR" ]; then
            if [ "$author" != "$FILTER_AUTHOR" ]; then
                continue
            fi
        fi

        if [ -n "$SEARCH_QUERY" ]; then
            local search_lower=$(echo "$SEARCH_QUERY" | tr '[:upper:]' '[:lower:]')
            local name_lower=$(echo "$skill_name" | tr '[:upper:]' '[:lower:]')
            local desc_lower=$(echo "$description" | tr '[:upper:]' '[:lower:]')

            if [[ ! "$name_lower" =~ $search_lower ]] && [[ ! "$desc_lower" =~ $search_lower ]]; then
                continue
            fi
        fi

        # Print a marker for each matching skill
        echo "MATCH"
    done
}

# Main execution
if [ "$VALIDATE_ONLY" = true ]; then
    validate_skills
    exit $?
fi

# Output header and process skills
if [ "$JSON_OUTPUT" = true ]; then
    echo "["
    first=true
    find "$SKILLS_DIR" -mindepth 2 -maxdepth 2 -name SKILL.md -print | sort | while IFS= read -r skill_file; do
        [ -f "$skill_file" ] || continue
        if [ "$first" = true ]; then
            first=false
        else
            echo ","
        fi
        process_skill_json "$skill_file"
    done
    echo ""
    echo "]"
else
    # Check if any skills match
    skill_count=$(count_skills | wc -l | tr -d ' ')

    if [ "$skill_count" -eq 0 ]; then
        echo -e "${YELLOW}No skills found matching your criteria.${NC}"
    else
        # Table output
        echo ""
        printf "${BOLD}%-20s  %-50s  %-12s  %-8s${NC}\n" "Skill" "Description" "Author" "Scope"
        printf '%s\n' "$(printf '%*s' 94 '' | tr ' ' '-')"

        find "$SKILLS_DIR" -mindepth 2 -maxdepth 2 -name SKILL.md -print | sort | while IFS= read -r skill_file; do
            [ -f "$skill_file" ] || continue
            process_skill_table "$skill_file"
        done
        echo ""

        if [ "$SHOW_ALL" = true ]; then
            echo -e "${BOLD}Full Details:${NC}"
            echo ""
            find "$SKILLS_DIR" -mindepth 2 -maxdepth 2 -name SKILL.md -print | sort | while IFS= read -r skill_file; do
                [ -f "$skill_file" ] || continue
                process_skill_details "$skill_file"
            done
        fi
    fi
fi

# Show count
if [ "$JSON_OUTPUT" = false ] && [ "$VALIDATE_ONLY" = false ]; then
    echo ""
    echo -e "${BLUE}Total: $skill_count skills${NC}"
    echo -e "${CYAN}Tip: Use --all for full details, --search to filter, --validate to check integrity${NC}"
fi
