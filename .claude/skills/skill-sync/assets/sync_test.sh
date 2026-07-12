#!/bin/bash
# Unit tests for sync.sh
# Run: ./sync_test.sh
#
# shellcheck disable=SC2317
# Reason: Test functions are discovered and called dynamically via declare -F.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SYNC_SCRIPT="$SCRIPT_DIR/sync.sh"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Test environment
TEST_DIR=""

# =============================================================================
# TEST FRAMEWORK
# =============================================================================

setup_test_env() {
    TEST_DIR=$(mktemp -d)

    # sync.sh's find_repo_root climbs PWD and (as fallback) SCRIPT_DIR until it
    # finds a .agent or .agents directory. The test must create one at
    # $TEST_DIR so REPO_ROOT is anchored there.
    mkdir -p "$TEST_DIR/.agent"

    # Multi-provider skill directories (matches SKILL_PROVIDERS in lib/utils.sh).
    # All skills use scope=root in the default fixture because sync.sh's
    # get_agents_path() maps multiple scopes (root|api|common|infra|database)
    # to the same $REPO_ROOT/AGENTS.md and the final writer wins (the script
    # replaces the existing Auto-invoke section in place). Tests that need
    # multi-scope coverage override specific SKILL.md files locally.
    mkdir -p "$TEST_DIR/.claude/skills/claude-skill"
    mkdir -p "$TEST_DIR/.opencode/skills/opencode-skill"
    mkdir -p "$TEST_DIR/.qwen/skills/qwen-skill"
    mkdir -p "$TEST_DIR/.agents/skills/agents-skill"
    mkdir -p "$TEST_DIR/.agent/skills/no-metadata"
    mkdir -p "$TEST_DIR/.agent/skills/skill-sync/assets"

    write_skill_md() {
        local path="$1" name="$2" scope="$3" action="$4"
        cat > "$path" << EOF
---
name: $name
description: >
  Mock skill for $name testing.
license: Apache-2.0
metadata:
  author: test
  version: "1.0"
  scope: [$scope]
  auto_invoke: "$action"
allowed-tools: Read
---

# $name
EOF
    }

    write_skill_md "$TEST_DIR/.claude/skills/claude-skill/SKILL.md" "claude-skill" "root" "Claude action"
    write_skill_md "$TEST_DIR/.opencode/skills/opencode-skill/SKILL.md" "opencode-skill" "root" "OpenCode action"
    write_skill_md "$TEST_DIR/.qwen/skills/qwen-skill/SKILL.md" "qwen-skill" "root" "Qwen action"
    write_skill_md "$TEST_DIR/.agents/skills/agents-skill/SKILL.md" "agents-skill" "root" "Agents action"

    # Skill missing both scope and auto_invoke -> must be reported in
    # "Skills missing sync metadata" section.
    cat > "$TEST_DIR/.agent/skills/no-metadata/SKILL.md" << 'EOF'
---
name: no-metadata
description: Skill without sync metadata.
license: Apache-2.0
metadata:
  author: test
  version: "1.0"
allowed-tools: Read
---

# No Metadata Skill
EOF

    # Target AGENTS.md (sync.sh writes root|api|common|infra|database to this
    # single file). Include a "Skills Reference" blockquote so the inserter can
    # hook after it.
    cat > "$TEST_DIR/AGENTS.md" << 'EOF'
# Root AGENTS

> **Skills Reference**: For detailed patterns, use these skills:
> - [`agents-skill`](.agents/skills/agents-skill/SKILL.md)

## Project Overview

This is the root agents file.
EOF

    # Copy sync.sh into the test repo at the canonical location the script
    # expects (alongside lib/utils.sh). utils.sh will walk up to find
    # $TEST_DIR/.agent, anchoring REPO_ROOT to $TEST_DIR.
    cp "$SYNC_SCRIPT" "$TEST_DIR/.agent/skills/skill-sync/assets/sync.sh"
    chmod +x "$TEST_DIR/.agent/skills/skill-sync/assets/sync.sh"

    # Copy the utils.sh library next to sync.sh so the script can source it
    # from the standard location. Walk up from SCRIPT_DIR (the mirror's
    # assets directory) to find utils.sh in any of the project's provider
    # skill trees, since the test mirror may live in .claude/, .opencode/,
    # .qwen/, etc.
    UTILS_SRC=""
    if [ -f "$SCRIPT_DIR/../../lib/utils.sh" ]; then
        UTILS_SRC="$SCRIPT_DIR/../../lib/utils.sh"
    elif [ -f "$SCRIPT_DIR/../lib/utils.sh" ]; then
        UTILS_SRC="$SCRIPT_DIR/../lib/utils.sh"
    else
        # Climb up to at most 8 parents, checking each level for utils.sh
        # in any of the project's known provider trees. Mirrors under
        # .claude/, .opencode/, .qwen/ need to climb 4 levels to reach the
        # project root and then look under .agents/skills/lib/. Mirrors
        # under internal/coreskills/ need to climb 4 levels to reach
        # internal/coreskills/ and then look under skills/lib/.
        search_dir="$SCRIPT_DIR"
        for _ in 1 2 3 4 5 6 7 8; do
            search_dir=$(dirname "$search_dir")
            for candidate in \
                "$search_dir/lib/utils.sh" \
                "$search_dir/utils.sh" \
                "$search_dir/.agents/skills/lib/utils.sh" \
                "$search_dir/.agent/skills/lib/utils.sh" \
                "$search_dir/skills/lib/utils.sh" \
                "$search_dir/internal/coreskills/skills/lib/utils.sh"; do
                if [ -f "$candidate" ]; then
                    UTILS_SRC="$candidate"
                    break 2
                fi
            done
        done
    fi
    if [ -n "$UTILS_SRC" ]; then
        mkdir -p "$TEST_DIR/.agent/skills/lib"
        cp "$UTILS_SRC" "$TEST_DIR/.agent/skills/lib/utils.sh"
    else
        echo "ERROR: sync_test.sh could not locate lib/utils.sh from $SCRIPT_DIR" >&2
        return 1
    fi
}

teardown_test_env() {
    if [ -n "$TEST_DIR" ] && [ -d "$TEST_DIR" ]; then
        rm -rf "$TEST_DIR"
    fi
}

# run_sync invokes the test-local copy of sync.sh. The script is executed
# from $TEST_DIR so find_repo_root (PWD-first) anchors REPO_ROOT to $TEST_DIR.
run_sync() {
    (cd "$TEST_DIR" && bash "$TEST_DIR/.agent/skills/skill-sync/assets/sync.sh" "$@" 2>&1)
}

# =============================================================================
# ASSERTIONS
# =============================================================================

assert_equals() {
    local expected="$1" actual="$2" message="$3"
    if [ "$expected" = "$actual" ]; then
        return 0
    fi
    echo -e "${RED}  FAIL: $message${NC}"
    echo "    Expected: $expected"
    echo "    Actual:   $actual"
    return 1
}

assert_contains() {
    local haystack="$1" needle="$2" message="$3"
    if echo "$haystack" | grep -q -F -- "$needle"; then
        return 0
    fi
    echo -e "${RED}  FAIL: $message${NC}"
    echo "    String not found: $needle"
    return 1
}

assert_not_contains() {
    local haystack="$1" needle="$2" message="$3"
    if ! echo "$haystack" | grep -q -F -- "$needle"; then
        return 0
    fi
    echo -e "${RED}  FAIL: $message${NC}"
    echo "    String should not be found: $needle"
    return 1
}

assert_file_contains() {
    local file="$1" needle="$2" message="$3"
    if grep -q -F -- "$needle" "$file" 2>/dev/null; then
        return 0
    fi
    echo -e "${RED}  FAIL: $message${NC}"
    echo "    File: $file"
    echo "    String not found: $needle"
    return 1
}

assert_file_not_contains() {
    local file="$1" needle="$2" message="$3"
    if ! grep -q -F -- "$needle" "$file" 2>/dev/null; then
        return 0
    fi
    echo -e "${RED}  FAIL: $message${NC}"
    echo "    File: $file"
    echo "    String should not be found: $needle"
    return 1
}

# =============================================================================
# TESTS: FLAG PARSING
# =============================================================================

test_flag_help_shows_usage() {
    local output
    output=$(run_sync --help)
    assert_contains "$output" "Usage:" "Help should show usage" && \
    assert_contains "$output" "--dry-run" "Help should mention --dry-run" && \
    assert_contains "$output" "--scope" "Help should mention --scope"
}

test_flag_unknown_reports_error() {
    local output rc=0
    output=$(run_sync --unknown 2>&1) || rc=$?
    [ "$rc" -ne 0 ] || return 1
    assert_contains "$output" "Unknown option" "Should report unknown option"
}

test_flag_dryrun_shows_changes() {
    local output
    output=$(run_sync --dry-run)
    assert_contains "$output" "[DRY RUN]" "Should show dry run marker" && \
    assert_contains "$output" "Would update" "Should say would update"
}

test_flag_dryrun_no_file_changes() {
    run_sync --dry-run > /dev/null
    # AGENTS.md should not have been modified in dry-run: it must not contain
    # the auto-invoke section yet.
    assert_file_not_contains "$TEST_DIR/AGENTS.md" "### Auto-invoke Skills" \
        "AGENTS.md should not be modified in dry run"
}

test_flag_scope_filters_to_matching() {
    local output
    # --scope root should keep root-scope skills and skip others.
    # We rename one skill to a non-root scope to verify filtering.
    cat > "$TEST_DIR/.opencode/skills/opencode-skill/SKILL.md" << 'EOF'
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
EOF

    output=$(run_sync --dry-run --scope root)
    assert_contains "$output" "Processing: root" "Should process root scope" && \
    assert_not_contains "$output" "Processing: ui" "Should not process ui scope"
}

# =============================================================================
# TESTS: METADATA EXTRACTION
# =============================================================================

test_metadata_extracts_scope() {
    local output
    output=$(run_sync --dry-run)
    # All four scoped skills are root-scoped in the default fixture
    assert_contains "$output" "Processing: root" "Should detect root scope"
}

test_metadata_extracts_auto_invoke() {
    local output
    output=$(run_sync --dry-run)
    assert_contains "$output" "Claude action" "Should extract claude-skill auto_invoke" && \
    assert_contains "$output" "OpenCode action" "Should extract opencode-skill auto_invoke" && \
    assert_contains "$output" "Agents action" "Should extract agents-skill auto_invoke"
}

test_metadata_missing_reports_skills() {
    local output
    output=$(run_sync --dry-run)
    assert_contains "$output" "Skills missing sync metadata" "Should report missing metadata section" && \
    assert_contains "$output" "no-metadata" "Should list skill without metadata"
}

test_metadata_skips_without_scope_in_processing() {
    local output processing_lines
    output=$(run_sync --dry-run)
    processing_lines=$(echo "$output" | grep "Processing:")
    assert_not_contains "$processing_lines" "no-metadata" "Should not process skill without scope"
}

# =============================================================================
# TESTS: AUTO-INVOKE GENERATION
# =============================================================================

test_generate_creates_section() {
    run_sync > /dev/null
    assert_file_contains "$TEST_DIR/AGENTS.md" "### Auto-invoke Skills" \
        "Should create Auto-invoke section" && \
    assert_file_contains "$TEST_DIR/AGENTS.md" "| Action | Skill |" \
        "Should create table header"
}

test_generate_includes_all_skills() {
    run_sync > /dev/null
    assert_file_contains "$TEST_DIR/AGENTS.md" "claude-skill" \
        "AGENTS.md should contain claude-skill" && \
    assert_file_contains "$TEST_DIR/AGENTS.md" "opencode-skill" \
        "AGENTS.md should contain opencode-skill" && \
    assert_file_contains "$TEST_DIR/AGENTS.md" "qwen-skill" \
        "AGENTS.md should contain qwen-skill" && \
    assert_file_contains "$TEST_DIR/AGENTS.md" "agents-skill" \
        "AGENTS.md should contain agents-skill"
}

test_generate_uses_action_text() {
    run_sync > /dev/null
    assert_file_contains "$TEST_DIR/AGENTS.md" "Claude action" \
        "Should include claude-skill auto_invoke text" && \
    assert_file_contains "$TEST_DIR/AGENTS.md" "OpenCode action" \
        "Should include opencode-skill auto_invoke text"
}

test_generate_splits_multi_action_auto_invoke_list() {
    # Replace claude-skill with one that has a list of actions.
    cat > "$TEST_DIR/.claude/skills/claude-skill/SKILL.md" << 'EOF'
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
EOF

    run_sync > /dev/null

    # Both actions should produce rows in the table
    assert_file_contains "$TEST_DIR/AGENTS.md" "| Action A | \`claude-skill\` |" \
        "Should create row for Action A" && \
    assert_file_contains "$TEST_DIR/AGENTS.md" "| Action B | \`claude-skill\` |" \
        "Should create row for Action B"
}

test_generate_orders_rows_by_action_then_skill() {
    # Two skills with intentionally out-of-order actions
    cat > "$TEST_DIR/.claude/skills/claude-skill/SKILL.md" << 'EOF'
---
name: claude-skill
description: Mock skill.
license: Apache-2.0
metadata:
  author: test
  version: "1.0"
  scope: [root]
  auto_invoke:
    - "Z action"
    - "A action"
allowed-tools: Read
---
EOF

    cat > "$TEST_DIR/.opencode/skills/opencode-skill/SKILL.md" << 'EOF'
---
name: opencode-skill
description: Second skill.
license: Apache-2.0
metadata:
  author: test
  version: "1.0"
  scope: [root]
  auto_invoke: "A action"
allowed-tools: Read
---
EOF

    run_sync > /dev/null

    # Extract the table region
    local table_segment
    table_segment=$(awk '
        /^\| Action \| Skill \|/ { in_table=1 }
        in_table && /^---$/ { next }
        in_table && /^\|/ { print }
        in_table && !/^\|/ { exit }
    ' "$TEST_DIR/AGENTS.md")

    local first_a first_z
    first_a=$(echo "$table_segment" | awk '/\| A action \|/ { print NR; exit }')
    first_z=$(echo "$table_segment" | awk '/\| Z action \|/ { print NR; exit }')

    [ -n "$first_a" ] && [ -n "$first_z" ] && [ "$first_a" -lt "$first_z" ]
}

# =============================================================================
# TESTS: AGENTS.MD UPDATE
# =============================================================================

test_update_preserves_header() {
    run_sync > /dev/null
    assert_file_contains "$TEST_DIR/AGENTS.md" "# Root AGENTS" \
        "Should preserve original header"
}

test_update_preserves_skills_reference() {
    run_sync > /dev/null
    assert_file_contains "$TEST_DIR/AGENTS.md" "Skills Reference" \
        "Should preserve Skills Reference section"
}

test_update_preserves_content_after() {
    run_sync > /dev/null
    assert_file_contains "$TEST_DIR/AGENTS.md" "## Project Overview" \
        "Should preserve content after the inserted block"
}

test_update_replaces_existing_section() {
    # First run creates the section
    run_sync > /dev/null
    local first_has
    first_has=$(grep -c "### Auto-invoke Skills" "$TEST_DIR/AGENTS.md" || true)

    # Edit the auto_invoke text in the source skill (portable: BSD/GNU sed
    # both accept -i with separate backup arg).
    sed -i.bak 's/Claude action/Modified Claude action/' "$TEST_DIR/.claude/skills/claude-skill/SKILL.md"
    rm -f "$TEST_DIR/.claude/skills/claude-skill/SKILL.md.bak"

    run_sync > /dev/null

    local second_has
    second_has=$(grep -c "### Auto-invoke Skills" "$TEST_DIR/AGENTS.md" || true)

    assert_equals "$first_has" "$second_has" \
        "Second run should not duplicate the section" && \
    assert_file_contains "$TEST_DIR/AGENTS.md" "Modified Claude action" \
        "Should update with new auto_invoke text" && \
    assert_file_not_contains "$TEST_DIR/AGENTS.md" "| Claude action |" \
        "Should remove old auto_invoke row"
}

# =============================================================================
# TESTS: IDEMPOTENCY
# =============================================================================

test_idempotent_multiple_runs() {
    run_sync > /dev/null
    local first_content
    first_content=$(cat "$TEST_DIR/AGENTS.md")

    run_sync > /dev/null
    local second_content
    second_content=$(cat "$TEST_DIR/AGENTS.md")

    assert_equals "$first_content" "$second_content" \
        "Multiple runs should produce identical output"
}

test_idempotent_no_duplicate_sections() {
    run_sync > /dev/null
    run_sync > /dev/null
    run_sync > /dev/null

    local count
    count=$(grep -c "### Auto-invoke Skills" "$TEST_DIR/AGENTS.md")
    assert_equals "1" "$count" "Should have exactly one Auto-invoke section"
}

# =============================================================================
# TESTS: MULTI-SCOPE SKILLS
# =============================================================================

test_multiscope_skill_processes_both_scopes() {
    # Skill with two scopes (root + api) — the script processes each scope
    # but they map to the same $REPO_ROOT/AGENTS.md. The assertion here is
    # that the dry-run output confirms both scopes were processed (covered
    # by the per-scope Processing: line), not that both rows survive the
    # same-file overwrite.
    cat > "$TEST_DIR/.claude/skills/claude-skill/SKILL.md" << 'EOF'
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
EOF

    local output
    output=$(run_sync --dry-run)
    assert_contains "$output" "Processing: root" "Should process root scope" && \
    assert_contains "$output" "Processing: api" "Should process api scope"
}

# =============================================================================
# TEST RUNNER
# =============================================================================

run_all_tests() {
    local test_functions current_section=""

    test_functions=$(declare -F | awk '{print $3}' | grep '^test_' | sort)

    for test_func in $test_functions; do
        local section
        section=$(echo "$test_func" | sed 's/^test_//' | cut -d'_' -f1)
        section="$(echo "${section:0:1}" | tr '[:lower:]' '[:upper:]')${section:1}"

        if [ "$section" != "$current_section" ]; then
            [ -n "$current_section" ] && echo ""
            echo -e "${YELLOW}${section} tests:${NC}"
            current_section="$section"
        fi

        local test_name
        test_name=$(echo "$test_func" | sed 's/^test_//' | tr '_' ' ')

        TESTS_RUN=$((TESTS_RUN + 1))
        echo -n "  $test_name... "

        setup_test_env

        if $test_func; then
            echo -e "${GREEN}PASS${NC}"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi

        teardown_test_env
    done
}

# =============================================================================
# MAIN
# =============================================================================

echo ""
echo "Running sync.sh unit tests"
echo "=============================="
echo ""

run_all_tests

echo ""
echo "=============================="
if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All $TESTS_RUN tests passed!${NC}"
    exit 0
else
    echo -e "${RED}$TESTS_FAILED of $TESTS_RUN tests failed${NC}"
    exit 1
fi
