---
name: sequential-thinking
description: >
  Disciplined use of the sequential-thinking tool for complex planning, analysis, revisions, branching, verification, and knowing when NOT to use it.
  Trigger: complex planning, multi-step analysis, architectural revisions, or when explicit step-by-step thinking is required.
metadata:
  version: "1.0"
  scope: [root]
  auto_invoke: "Complex planning or multi-step analysis"
---

# Sequential Thinking

## When to Use

Use this skill when:
- Breaking down complex problems into steps.
- Planning and design with room for revision.
- Analysis that might need course correction.
- Problems where the full scope might not be clear initially.
- Tasks that need to maintain context over multiple steps.

---

## Critical Patterns

### Pattern 1: Mapping the Solution
Always start with an initial estimate of needed thoughts, but be ready to adjust. Use the first few thoughts to define the problem and constraints.

### Pattern 2: Branching and Backtracking
Use branching when exploring alternative approaches. If a path seems suboptimal, revise previous thoughts or backtrack to a better branching point.

### Pattern 3: Hypothesis Verification
Generate a solution hypothesis only after sufficient analysis. Verify the hypothesis against all previous thought steps before concluding.

### Pattern 4: Disciplined Termination
Only set `nextThoughtNeeded` to false when a satisfactory and verified answer is reached. Avoid ending early if uncertainty remains.

### Pattern 5: Avoiding Overkill
DO NOT use this tool for trivial tasks, simple file edits, or direct command executions where the path is obvious.

---

## Decision Tree

```text
Task is complex and multi-step? → Use sequential-thinking
Need to explore alternatives? → Use branching
Uncertain about initial plan? → Use revisions
Task is trivial or direct? → SKIP tool, execute directly
```

---

## Code Examples

### Example 1: Planning a Refactor

```json
{
  "thought": "I need to move the auth logic to a separate package. First, I'll identify all dependencies...",
  "thoughtNumber": 1,
  "totalThoughts": 5,
  "nextThoughtNeeded": true
}
```

---

## Commands

```bash
# No specific shell commands, but use the tool via the MCP interface.
```
